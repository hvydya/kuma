package bootstrap_test

import (
	"context"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	dp_server_cfg "github.com/kumahq/kuma/pkg/config/dp-server"
	bootstrap_config "github.com/kumahq/kuma/pkg/config/xds/bootstrap"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	dp_server "github.com/kumahq/kuma/pkg/dp-server"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test"
	test_metrics "github.com/kumahq/kuma/pkg/test/metrics"
	"github.com/kumahq/kuma/pkg/xds/bootstrap"
)

var _ = Describe("Bootstrap Server", func() {

	var stop chan struct{}
	var resManager manager.ResourceManager
	var config *bootstrap_config.BootstrapParamsConfig
	var baseUrl string
	var metrics core_metrics.Metrics

	httpClient := &http.Client{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}

	BeforeEach(func() {
		resManager = manager.NewResourceManager(memory.NewStore())
		config = bootstrap_config.DefaultBootstrapParamsConfig()
		config.XdsHost = "localhost"
		config.XdsPort = 5678

		port, err := test.GetFreePort()
		baseUrl = "https://localhost:" + strconv.Itoa(port)
		Expect(err).ToNot(HaveOccurred())
		metrics, err = core_metrics.NewMetrics("Standalone")
		Expect(err).ToNot(HaveOccurred())

		dpServerCfg := dp_server_cfg.DpServerConfig{
			Port:        port,
			TlsCertFile: filepath.Join("..", "..", "..", "test", "certs", "server-cert.pem"),
			TlsKeyFile:  filepath.Join("..", "..", "..", "test", "certs", "server-key.pem"),
		}
		dpServer := dp_server.NewDpServer(dpServerCfg, metrics)

		generator, err := bootstrap.NewDefaultBootstrapGenerator(resManager, config, filepath.Join("..", "..", "..", "test", "certs", "server-cert.pem"), true)
		Expect(err).ToNot(HaveOccurred())
		bootstrapHandler := bootstrap.BootstrapHandler{
			Generator: generator,
		}
		dpServer.HTTPMux().HandleFunc("/bootstrap", bootstrapHandler.Handle)

		stop = make(chan struct{})
		go func() {
			defer GinkgoRecover()
			err := dpServer.Start(stop)
			Expect(err).ToNot(HaveOccurred())
		}()
		Eventually(func() bool {
			resp, err := httpClient.Get(baseUrl)
			if err != nil {
				return false
			}
			Expect(resp.Body.Close()).To(Succeed())
			return true
		}).Should(BeTrue())
	}, 5)

	AfterEach(func() {
		close(stop)
	})

	BeforeEach(func() {
		err := resManager.Create(context.Background(), &mesh.MeshResource{}, store.CreateByKey(model.DefaultMesh, model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
		core.Now = func() time.Time {
			now, _ := time.Parse(time.RFC3339, "2018-07-17T16:05:36.995+00:00")
			return now
		}
	})

	type testCase struct {
		dataplaneName      string
		body               string
		expectedConfigFile string
	}
	DescribeTable("should return configuration",
		func(given testCase) {
			// given
			res := mesh.DataplaneResource{
				Spec: mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "8.8.8.8",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Port:        443,
								ServicePort: 8443,
								Tags: map[string]string{
									"kuma.io/service": "backend",
								},
							},
						},
					},
				},
			}
			err := resManager.Create(context.Background(), &res, store.CreateByKey(given.dataplaneName, "default"))
			Expect(err).ToNot(HaveOccurred())

			// when
			resp, err := httpClient.Post(baseUrl+"/bootstrap", "application/json", strings.NewReader(given.body))

			// then
			Expect(err).ToNot(HaveOccurred())
			received, err := ioutil.ReadAll(resp.Body)
			Expect(resp.Body.Close()).To(Succeed())
			Expect(err).ToNot(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			expected, err := ioutil.ReadFile(filepath.Join("testdata", given.expectedConfigFile))
			Expect(err).ToNot(HaveOccurred())

			Expect(received).To(MatchYAML(expected))
		},
		Entry("minimal data provided (universal)", testCase{
			dataplaneName:      "dp-1",
			body:               `{ "mesh": "default", "name": "dp-1", "dataplaneTokenPath": "/tmp/token" }`,
			expectedConfigFile: "bootstrap.universal.golden.yaml",
		}),
		Entry("minimal data provided (k8s)", testCase{
			dataplaneName:      "dp-1.default",
			body:               `{ "mesh": "default", "name": "dp-1.default", "dataplaneTokenPath": "/tmp/token" }`,
			expectedConfigFile: "bootstrap.k8s.golden.yaml",
		}),
		Entry("full data provided", testCase{
			dataplaneName:      "dp-1.default",
			body:               `{ "mesh": "default", "name": "dp-1.default", "adminPort": 1234, "dataplaneTokenPath": "/tmp/token" }`,
			expectedConfigFile: "bootstrap.overridden.golden.yaml",
		}),
	)

	It("should return 404 for unknown dataplane", func() {
		// when
		json := `
		{
			"mesh": "default",
			"name": "dp-1.default",
			"dataplaneTokenPath": "/tmp/token"
		}
		`

		resp, err := httpClient.Post(baseUrl+"/bootstrap", "application/json", strings.NewReader(json))
		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Body.Close()).To(Succeed())
		Expect(resp.StatusCode).To(Equal(404))
	})

	It("should return 422 for the lack of the dataplane token", func() {
		// given
		res := mesh.DataplaneResource{
			Spec: mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "8.8.8.8",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port:        443,
							ServicePort: 8443,
							Tags: map[string]string{
								"kuma.io/service": "backend",
							},
						},
					},
				},
			},
		}
		err := resManager.Create(context.Background(), &res, store.CreateByKey("dp-1", "default"))
		Expect(err).ToNot(HaveOccurred())

		// when
		json := `
		{
			"mesh": "default",
			"name": "dp-1"
		}
		`

		resp, err := httpClient.Post(baseUrl+"/bootstrap", "application/json", strings.NewReader(json))
		// then
		Expect(err).ToNot(HaveOccurred())
		bytes, err := ioutil.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Body.Close()).To(Succeed())
		Expect(resp.StatusCode).To(Equal(422))
		Expect(string(bytes)).To(Equal("Dataplane Token is required. Generate token using 'kumactl generate dataplane-token > /path/file' and provide it via --dataplane-token-file=/path/file argument to Kuma DP"))

	})

	It("should publish metrics", func() {
		// given
		res := mesh.DataplaneResource{
			Spec: mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "8.8.8.8",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port:        443,
							ServicePort: 8443,
							Tags: map[string]string{
								"kuma.io/service": "backend",
							},
						},
					},
				},
			},
		}
		err := resManager.Create(context.Background(), &res, store.CreateByKey("dp-1", "default"))
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = httpClient.Post(baseUrl+"/bootstrap", "application/json", strings.NewReader(`{ "mesh": "default", "name": "dp-1" }`))

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(test_metrics.FindMetric(metrics, "dp_server_http_request_duration_seconds", "handler", "/bootstrap")).ToNot(BeNil())
		Expect(test_metrics.FindMetric(metrics, "dp_server_http_requests_inflight", "handler", "/bootstrap")).ToNot(BeNil())
		Expect(test_metrics.FindMetric(metrics, "dp_server_http_response_size_bytes", "handler", "/bootstrap")).ToNot(BeNil())
	})
})
