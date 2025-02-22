package framework

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/kumahq/kuma/pkg/tls"

	"github.com/go-errors/errors"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type InstallFunc func(cluster Cluster) error

func YamlK8s(yaml string) InstallFunc {
	return func(cluster Cluster) error {
		_, err := retry.DoWithRetryE(cluster.GetTesting(), "install yaml resource", DefaultRetries, DefaultTimeout,
			func() (s string, err error) {
				return "", k8s.KubectlApplyFromStringE(cluster.GetTesting(), cluster.GetKubectlOptions(), yaml)
			})
		return err
	}
}

func YamlUniversal(yaml string) InstallFunc {
	return func(cluster Cluster) error {
		_, err := retry.DoWithRetryE(cluster.GetTesting(), "install yaml resource", DefaultRetries, DefaultTimeout,
			func() (s string, err error) {
				kumactl := cluster.GetKumactlOptions()
				return "", kumactl.KumactlApplyFromString(yaml)
			})
		return err
	}
}

func YamlPathK8s(path string) InstallFunc {
	return func(cluster Cluster) error {
		_, err := retry.DoWithRetryE(cluster.GetTesting(), "install yaml resource by path", DefaultRetries, DefaultTimeout,
			func() (s string, err error) {
				return "", k8s.KubectlApplyE(cluster.GetTesting(), cluster.GetKubectlOptions(), path)
			})
		return err
	}
}

func Kuma(mode string, fs ...DeployOptionsFunc) InstallFunc {
	return func(cluster Cluster) error {
		err := cluster.DeployKuma(mode, fs...)
		return err
	}
}

func KumaDNS() InstallFunc {
	return func(cluster Cluster) error {
		err := cluster.InjectDNS(KumaNamespace)
		return err
	}
}

func WaitService(namespace, service string) InstallFunc {
	return func(c Cluster) error {
		k8s.WaitUntilServiceAvailable(c.GetTesting(), c.GetKubectlOptions(namespace), service, 10, 3*time.Second)
		return nil
	}
}

func WaitNumPods(num int, app string) InstallFunc {
	return func(c Cluster) error {
		k8s.WaitUntilNumPodsCreated(c.GetTesting(), c.GetKubectlOptions(),
			kube_meta.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", app),
			}, num, DefaultRetries, DefaultTimeout)
		return nil
	}
}

func WaitPodsAvailable(namespace, app string) InstallFunc {
	return func(c Cluster) error {
		pods, err := k8s.ListPodsE(c.GetTesting(), c.GetKubectlOptions(namespace),
			kube_meta.ListOptions{LabelSelector: fmt.Sprintf("app=%s", app)})
		if err != nil {
			return err
		}
		for _, p := range pods {
			err := k8s.WaitUntilPodAvailableE(c.GetTesting(), c.GetKubectlOptions(namespace), p.GetName(), DefaultRetries, DefaultTimeout)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WaitPodsNotAvailable(namespace, app string) InstallFunc {
	return func(c Cluster) error {
		pods, err := k8s.ListPodsE(c.GetTesting(), c.GetKubectlOptions(namespace),
			kube_meta.ListOptions{LabelSelector: fmt.Sprintf("app=%s", app)})
		if err != nil {
			return err
		}

		for _, p := range pods {
			_, _ = retry.DoWithRetryE(
				c.GetTesting(),
				"Wait pod deletion",
				DefaultRetries,
				DefaultTimeout,
				func() (string, error) {
					pod, err := k8s.GetPodE(c.GetTesting(), c.GetKubectlOptions(namespace), p.GetName())
					if err == nil {
						return "", err
					}
					if !k8s.IsPodAvailable(pod) {
						return "Pod is not available", nil
					}
					return "", errors.Errorf("Pod is still available")
				},
			)
		}
		return nil
	}
}

func EchoServerK8s() InstallFunc {
	const name = "echo-server"
	return Combine(
		YamlPathK8s(filepath.Join("testdata", fmt.Sprintf("%s.yaml", name))),
		WaitService(TestNamespace, name),
		WaitNumPods(1, name),
		WaitPodsAvailable(TestNamespace, name),
	)
}

func EchoServerUniversal(id, token string, fs ...DeployOptionsFunc) InstallFunc {
	return func(cluster Cluster) error {
		fs = append(fs, WithAppname(AppModeEchoServer), WithId(id), WithToken(token))
		return cluster.DeployApp(fs...)
	}
}

func IngressUniversal(token string) InstallFunc {
	return func(cluster Cluster) error {
		uniCluster := cluster.(*UniversalCluster)
		app, err := NewUniversalApp(cluster.GetTesting(), uniCluster.name, AppIngress, true, []string{}, []string{})
		if err != nil {
			return err
		}
		err = app.mainApp.Start()
		if err != nil {
			return err
		}
		uniCluster.apps[AppIngress] = app

		publicAddress := uniCluster.apps[AppIngress].ip
		dpyaml := fmt.Sprintf(IngressDataplane, publicAddress, kdsPort, kdsPort)
		return uniCluster.CreateDP(app, "ingress", app.ip, dpyaml, token)
	}
}

func DemoClientK8s() InstallFunc {
	const name = "demo-client"
	return Combine(
		YamlPathK8s(filepath.Join("testdata", fmt.Sprintf("%s.yaml", name))),
		WaitService(TestNamespace, name),
		WaitNumPods(1, name),
		WaitPodsAvailable(TestNamespace, name),
	)
}

func DemoClientUniversal(token string) InstallFunc {
	return func(cluster Cluster) error {
		return cluster.DeployApp(WithAppname(AppModeDemoClient), WithToken(token))
	}
}

func Combine(fs ...InstallFunc) InstallFunc {
	return func(cluster Cluster) error {
		for _, f := range fs {
			if err := f(cluster); err != nil {
				return err
			}
		}
		return nil
	}
}

func Namespace(name string) InstallFunc {
	return func(cluster Cluster) error {
		return k8s.CreateNamespaceE(cluster.GetTesting(), cluster.GetKubectlOptions(), name)
	}
}

type ClusterSetup struct {
	installFuncs []InstallFunc
}

func NewClusterSetup() *ClusterSetup {
	return &ClusterSetup{}
}

func (cs *ClusterSetup) Install(fn InstallFunc) *ClusterSetup {
	cs.installFuncs = append(cs.installFuncs, fn)
	return cs
}

func (cs *ClusterSetup) Setup(cluster Cluster) error {
	return Combine(cs.installFuncs...)(cluster)
}

func CreateCertsForIP(ip string) (cert, key string, err error) {
	keyPair, err := tls.NewSelfSignedCert("kuma", tls.ServerCertType, "localhost", ip)
	if err != nil {
		return "", "", err
	}

	return string(keyPair.CertPEM), string(keyPair.KeyPEM), nil
}
