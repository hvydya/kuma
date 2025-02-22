package get_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	gomega_types "github.com/onsi/gomega/types"
	"github.com/spf13/cobra"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/app/kumactl/cmd"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	config_proto "github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	memory_resources "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("kumactl get meshes", func() {

	sampleMeshes := []*mesh.MeshResource{
		{
			Spec: mesh_proto.Mesh{
				Mtls: &mesh_proto.Mesh_Mtls{
					EnabledBackend: "builtin-1",
					Backends: []*mesh_proto.CertificateAuthorityBackend{
						{
							Name: "builtin-1",
							Type: "builtin",
						},
						{
							Name: "builtin-2",
							Type: "builtin",
						},
					},
				},
				Metrics: &mesh_proto.Metrics{
					EnabledBackend: "prometheus-1",
					Backends: []*mesh_proto.MetricsBackend{
						{
							Name: "prometheus-1",
							Type: mesh_proto.MetricsPrometheusType,
							Conf: util_proto.MustToStruct(&mesh_proto.PrometheusMetricsBackendConfig{
								Port: 1234,
								Path: "/non-standard-path",
							}),
						},
						{
							Name: "prometheus-2",
							Type: mesh_proto.MetricsPrometheusType,
							Conf: util_proto.MustToStruct(&mesh_proto.PrometheusMetricsBackendConfig{
								Port: 1235,
								Path: "/non-standard-path",
							}),
						},
					},
				},
				Logging: &mesh_proto.Logging{
					Backends: []*mesh_proto.LoggingBackend{
						{
							Name: "logstash",
							Type: mesh_proto.LoggingTcpType,
							Conf: util_proto.MustToStruct(&mesh_proto.TcpLoggingBackendConfig{
								Address: "127.0.0.1:5000",
							}),
						},
						{
							Name: "file",
							Type: mesh_proto.LoggingFileType,
							Conf: util_proto.MustToStruct(&mesh_proto.FileLoggingBackendConfig{
								Path: "/tmp/service.log",
							}),
						},
					},
				},
				Tracing: &mesh_proto.Tracing{
					Backends: []*mesh_proto.TracingBackend{
						{
							Name: "zipkin-us",
							Type: mesh_proto.TracingZipkinType,
							Conf: util_proto.MustToStruct(&mesh_proto.ZipkinTracingBackendConfig{
								Url: "http://zipkin.us:8080/v1/spans",
							}),
						},
						{
							Name: "zipkin-eu",
							Type: mesh_proto.TracingZipkinType,
							Conf: util_proto.MustToStruct(&mesh_proto.ZipkinTracingBackendConfig{
								Url: "http://zipkin.eu:8080/v1/spans",
							}),
						},
					},
				},
				Routing: &mesh_proto.Routing{
					LocalityAwareLoadBalancing: true,
				},
			},
			Meta: &test_model.ResourceMeta{
				Name: "mesh1",
			},
		},
		{
			Spec: mesh_proto.Mesh{
				Metrics: &mesh_proto.Metrics{
					Backends: []*mesh_proto.MetricsBackend{},
				},
				Logging: &mesh_proto.Logging{
					Backends: []*mesh_proto.LoggingBackend{},
				},
				Tracing: &mesh_proto.Tracing{
					Backends: []*mesh_proto.TracingBackend{},
				},
			},
			Meta: &test_model.ResourceMeta{
				Name: "mesh2",
			},
		},
	}

	Describe("GetMeshesCmd", func() {

		var rootCtx *kumactl_cmd.RootContext
		var rootCmd *cobra.Command
		var buf *bytes.Buffer
		var store core_store.ResourceStore
		rootTime, _ := time.Parse(time.RFC3339, "2008-04-27T16:05:36.995Z")
		BeforeEach(func() {
			// setup
			rootCtx = &kumactl_cmd.RootContext{
				Runtime: kumactl_cmd.RootRuntime{
					Now: func() time.Time { return rootTime },
					NewResourceStore: func(*config_proto.ControlPlaneCoordinates_ApiServer) (core_store.ResourceStore, error) {
						return store, nil
					},
				},
			}

			store = core_store.NewPaginationStore(memory_resources.NewStore())

			for _, ds := range sampleMeshes {
				key := core_model.ResourceKey{
					Mesh: ds.Meta.GetMesh(),
					Name: ds.Meta.GetName(),
				}
				err := store.Create(context.Background(), ds, core_store.CreateBy(key))
				Expect(err).ToNot(HaveOccurred())
			}

			rootCmd = cmd.NewRootCmd(rootCtx)
			buf = &bytes.Buffer{}
			rootCmd.SetOut(buf)
		})

		type testCase struct {
			outputFormat string
			goldenFile   string
			pagination   string
			matcher      func(interface{}) gomega_types.GomegaMatcher
		}

		DescribeTable("kumactl get meshes -o table|json|yaml",
			func(given testCase) {
				// given
				rootCmd.SetArgs(append([]string{
					"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
					"get", "meshes"}, given.outputFormat, given.pagination))

				// when
				err := rootCmd.Execute()
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				expected, err := ioutil.ReadFile(filepath.Join("testdata", given.goldenFile))
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(buf.String()).To(given.matcher(expected))
			},
			Entry("should support Table output by default", testCase{
				outputFormat: "",
				goldenFile:   "get-meshes.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support Table output explicitly", testCase{
				outputFormat: "-otable",
				goldenFile:   "get-meshes.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support pagination", testCase{
				outputFormat: "-otable",
				pagination:   "--size=1",
				goldenFile:   "get-meshes.pagination.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support JSON output", testCase{
				outputFormat: "-ojson",
				goldenFile:   "get-meshes.golden.json",
				matcher:      MatchJSON,
			}),
			Entry("should support YAML output", testCase{
				outputFormat: "-oyaml",
				goldenFile:   "get-meshes.golden.yaml",
				matcher:      MatchYAML,
			}),
		)
	})

})
