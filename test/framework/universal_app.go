package framework

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/pkg/errors"

	"github.com/gruntwork-io/terratest/modules/docker"
	"github.com/gruntwork-io/terratest/modules/testing"

	util_net "github.com/kumahq/kuma/pkg/util/net"
)

type AppMode string

const (
	AppModeCP              = "kuma-cp"
	AppIngress             = "ingress"
	AppModeEchoServer      = "echo-server"
	AppModeHttpsEchoServer = "https-echo-server"
	sshPort                = "22"

	IngressDataplane = `
type: Dataplane
mesh: default
name: dp-ingress
networking:
  address: {{ address }}
  ingress:
    publicAddress: %s
    publicPort: %d
  inbound:
  - port: %d
    tags:
      kuma.io/service: ingress
`
	EchoServerDataplane = `
type: Dataplane
mesh: default
name: {{ name }}
networking:
  address:  {{ address }}
  inbound:
  - port: %s
    servicePort: %s
    tags:
      kuma.io/service: echo-server_kuma-test_svc_%s
      kuma.io/protocol: http
`
	AppModeDemoClient   = "demo-client"
	DemoClientDataplane = `
type: Dataplane
mesh: default
name: {{ name }}
networking:
  address: {{ address }}
  inbound:
  - port: %s
    servicePort: %s
    tags:
      kuma.io/service: demo-client
  outbound:
  - port: 4000
    tags:
      kuma.io/service: echo-server_kuma-test_svc_%s
  - port: 4001
    tags:
      kuma.io/service: echo-server_kuma-test_svc_%s
  - port: 5000
    tags:
      kuma.io/service: external-service
`
)

var defaultDockerOptions = docker.RunOptions{
	Command: nil,
	Detach:  true,
	//Entrypoint:           "",
	EnvironmentVariables: nil,
	Init:                 false,
	//Name:                 "",
	Privileged: false,
	Remove:     true,
	Tty:        false,
	//User:                 "",
	Volumes:      nil,
	OtherOptions: []string{},
	Logger:       nil,
}

type UniversalApp struct {
	t            testing.TestingT
	mainApp      *SshApp
	mainAppEnv   []string
	mainAppArgs  []string
	dpApp        *SshApp
	ports        map[string]string
	lastUsedPort uint32
	container    string
	ip           string
	verbose      bool
}

func NewUniversalApp(t testing.TestingT, clusterName string, mode AppMode, verbose bool, env []string, args []string) (*UniversalApp, error) {
	app := &UniversalApp{
		t:            t,
		ports:        map[string]string{},
		lastUsedPort: 10204,
		verbose:      verbose,
	}

	app.allocatePublicPortsFor("22")

	if mode == AppModeCP {
		app.allocatePublicPortsFor("5678", "5680", "5681", "5682", "5685")
	}

	opts := defaultDockerOptions
	opts.OtherOptions = append(opts.OtherOptions, "--name", clusterName+"_"+string(mode))
	opts.OtherOptions = append(opts.OtherOptions, "--network", "kind")
	opts.OtherOptions = append(opts.OtherOptions, app.publishPortsForDocker()...)
	container, err := docker.RunAndGetIDE(t, KumaUniversalImage, &opts)
	if err != nil {
		return nil, err
	}

	app.container = container

	retry.DoWithRetry(app.t, "get IP "+app.container, DefaultRetries, DefaultTimeout,
		func() (string, error) {
			app.ip, err = app.getIP()
			if err != nil {
				return "Unable to get Container IP", err
			}
			return "Success", nil
		})

	fmt.Printf("Node IP %s\n", app.ip)

	app.CreateMainApp(env, args)

	return app, nil
}

func (s *UniversalApp) allocatePublicPortsFor(ports ...string) {
	for _, port := range ports {
		pubPortUInt32, err := util_net.PickTCPPort("", s.lastUsedPort+1, 11204)
		if err != nil {
			panic(err)
		}
		s.ports[port] = strconv.Itoa(int(pubPortUInt32))
		s.lastUsedPort = pubPortUInt32
	}
}

func (s *UniversalApp) publishPortsForDocker() (args []string) {
	for port, pubPort := range s.ports {
		args = append(args, "--publish="+pubPort+":"+port)
	}
	return
}

func (s *UniversalApp) Stop() error {
	out, err := docker.StopE(s.t, []string{s.container}, &docker.StopOptions{Time: 1})
	if err != nil {
		return errors.Wrapf(err, "Returned %s", out)
	}

	retry.DoWithRetry(s.t, "stop "+s.container, DefaultRetries, DefaultTimeout,
		func() (string, error) {
			_, err := docker.StopE(s.t, []string{s.container}, &docker.StopOptions{Time: 1})
			if err == nil {
				return "Container still running", errors.Errorf("Container still running")
			}
			return "Container stopped", nil
		})

	return nil
}

func (s *UniversalApp) ReStart() error {
	if err := s.mainApp.cmd.Process.Kill(); err != nil {
		return err
	}
	if _, err := s.mainApp.cmd.Process.Wait(); err != nil {
		return err
	}

	s.CreateMainApp(s.mainAppEnv, s.mainAppArgs)

	if err := s.mainApp.Start(); err != nil {
		return err
	}
	return nil
}

func (s *UniversalApp) CreateMainApp(env []string, args []string) {
	s.mainAppEnv = env
	s.mainAppArgs = args
	s.mainApp = NewSshApp(s.verbose, s.ports[sshPort], env, args)
}

func (s *UniversalApp) CreateDP(token, cpAddress, appname, ip, dpyaml string) {
	// and echo it to the Application Node
	err := NewSshApp(s.verbose, s.ports[sshPort], []string{}, []string{"printf ", "\"" + token + "\"", ">", "/kuma/token-" + appname}).Run()
	if err != nil {
		panic(err)
	}

	err = NewSshApp(s.verbose, s.ports[sshPort], []string{}, []string{"printf ", "\"" + dpyaml + "\"", ">", "/kuma/dpyaml-" + appname}).Run()
	if err != nil {
		panic(err)
	}
	// run the DP
	s.dpApp = NewSshApp(s.verbose, s.ports[sshPort], []string{}, []string{"kuma-dp", "run",
		"--cp-address=" + cpAddress,
		"--dataplane-token-file=/kuma/token-" + appname,
		"--dataplane-file=/kuma/dpyaml-" + appname,
		"--dataplane-var", "name=dp-" + appname,
		"--dataplane-var", "address=" + ip,
		"--binary-path", "/usr/local/bin/envoy"})
}

func (s *UniversalApp) getIP() (string, error) {
	cmd := SshCmd(s.ports[sshPort], []string{}, []string{"getent", "ahosts", s.container[:12]})
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		return "invalid", errors.Wrapf(err, "getent failed with %s", string(bytes))
	}
	split := strings.Split(string(bytes), " ")
	return split[0], nil
}

type SshApp struct {
	cmd    *exec.Cmd
	stdout bytes.Buffer
	stderr bytes.Buffer
	port   string
}

func NewSshApp(verbose bool, port string, env []string, args []string) *SshApp {
	app := &SshApp{
		port: port,
	}
	app.cmd = app.SshCmd(env, args)

	outWriters := []io.Writer{&app.stdout}
	errWriters := []io.Writer{&app.stderr}
	if verbose {
		outWriters = append(outWriters, os.Stdout)
		errWriters = append(errWriters, os.Stderr)
	}
	app.cmd.Stdout = io.MultiWriter(outWriters...)
	app.cmd.Stderr = io.MultiWriter(errWriters...)
	return app
}

func (s *SshApp) Run() error {
	fmt.Printf("Running %v\n", s.cmd)
	return s.cmd.Run()
}

func (s *SshApp) Start() error {
	fmt.Printf("Starting %v\n", s.cmd)
	return s.cmd.Start()
}

func (s *SshApp) Stop() error {
	if err := s.cmd.Process.Kill(); err != nil {
		return err
	}
	if _, err := s.cmd.Process.Wait(); err != nil {
		return err
	}
	return nil
}

func (s *SshApp) Wait() error {
	return s.cmd.Wait()
}

func (s *SshApp) Out() string {
	return s.stdout.String()
}

func (s *SshApp) Err() string {
	return s.stderr.String()
}

func (s *SshApp) SshCmd(env []string, args []string) *exec.Cmd {
	return SshCmd(s.port, env, args)
}

func SshCmd(port string, env []string, args []string) *exec.Cmd {
	sshArgs := append([]string{
		"-q", "-tt",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"root@localhost", "-p", port}, env...)
	sshArgs = append(sshArgs, args...)

	cmd := exec.Command("ssh", sshArgs...)
	return cmd
}
