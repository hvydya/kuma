package generator_test

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"time"

	envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"

	"github.com/kumahq/kuma/pkg/xds/envoy/endpoints"

	"github.com/golang/protobuf/ptypes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	model "github.com/kumahq/kuma/pkg/core/xds"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

type dummyCLACache struct {
	outboundTargets model.EndpointMap
}

func (d *dummyCLACache) GetCLA(ctx context.Context, meshName, service string) (*envoy_api_v2.ClusterLoadAssignment, error) {
	return endpoints.CreateClusterLoadAssignment(service, d.outboundTargets[service]), nil
}

var _ = Describe("ProxyTemplateProfileSource", func() {

	type testCase struct {
		mesh            string
		dataplane       string
		profile         string
		envoyConfigFile string
	}

	DescribeTable("Generate Envoy xDS resources",
		func(given testCase) {
			// setup
			gen := &generator.ProxyTemplateProfileSource{
				ProfileName: given.profile,
			}

			// given
			ctx := xds_context.Context{
				ConnectionInfo: xds_context.ConnectionInfo{
					Authority: "kuma-system:5677",
				},
				ControlPlane: &xds_context.ControlPlaneContext{
					SdsTlsCert: []byte("12345"),
				},
				Mesh: xds_context.MeshContext{
					Resource: &mesh_core.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "demo",
						},
					},
				},
			}

			Expect(util_proto.FromYAML([]byte(given.mesh), &ctx.Mesh.Resource.Spec)).To(Succeed())

			dataplane := mesh_proto.Dataplane{}
			Expect(util_proto.FromYAML([]byte(given.dataplane), &dataplane)).To(Succeed())

			outboundTargets := model.EndpointMap{
				"db": []model.Endpoint{
					{
						Target: "192.168.0.3",
						Port:   5432,
						Tags:   map[string]string{"kuma.io/service": "db", "role": "master"},
						Weight: 1,
					},
				},
				"elastic": []model.Endpoint{
					{
						Target: "192.168.0.4",
						Port:   9200,
						Tags:   map[string]string{"kuma.io/service": "elastic"},
						Weight: 1,
					},
				},
			}

			proxy := &model.Proxy{
				Id: model.ProxyId{Name: "demo.backend-01"},
				Dataplane: &mesh_core.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Name:    "backend-01",
						Mesh:    "demo",
						Version: "1",
					},
					Spec: dataplane,
				},
				TrafficRoutes: model.RouteMap{
					mesh_proto.OutboundInterface{
						DataplaneIP:   "127.0.0.1",
						DataplanePort: 54321,
					}: &mesh_core.TrafficRouteResource{
						Spec: mesh_proto.TrafficRoute{
							Conf: &mesh_proto.TrafficRoute_Conf{
								Split: []*mesh_proto.TrafficRoute_Split{
									{
										Weight:      100,
										Destination: mesh_proto.MatchService("db"),
									},
								},
							},
						},
					},
					mesh_proto.OutboundInterface{
						DataplaneIP:   "127.0.0.1",
						DataplanePort: 59200,
					}: &mesh_core.TrafficRouteResource{
						Spec: mesh_proto.TrafficRoute{
							Conf: &mesh_proto.TrafficRoute_Conf{
								Split: []*mesh_proto.TrafficRoute_Split{
									{
										Weight:      100,
										Destination: mesh_proto.MatchService("elastic"),
									},
								},
							},
						},
					},
				},
				OutboundTargets: outboundTargets,
				HealthChecks: model.HealthCheckMap{
					"elastic": &mesh_core.HealthCheckResource{
						Spec: mesh_proto.HealthCheck{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"kuma.io/service": "*"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"kuma.io/service": "elastic"}},
							},
							Conf: &mesh_proto.HealthCheck_Conf{
								Interval:           ptypes.DurationProto(5 * time.Second),
								Timeout:            ptypes.DurationProto(4 * time.Second),
								UnhealthyThreshold: 3,
								HealthyThreshold:   2,
							},
						},
					},
				},
				Metadata: &model.DataplaneMetadata{
					AdminPort: 9902,
				},
				CLACache: &dummyCLACache{outboundTargets: outboundTargets},
			}

			// when
			rs, err := gen.Generate(ctx, proxy)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			resp, err := rs.List().ToDeltaDiscoveryResponse()
			// then
			Expect(err).ToNot(HaveOccurred())
			// when
			actual, err := util_proto.ToYAML(resp)
			// then
			Expect(err).ToNot(HaveOccurred())

			expected, err := ioutil.ReadFile(filepath.Join("testdata", "profile-source", given.envoyConfigFile))
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(expected))
		},
		Entry("should support pre-defined `default-proxy` profile; transparent_proxying=false", testCase{
			mesh: `
            mtls:
              enabledBackend: builtin
              backends:
              - type: builtin
                name: builtin
`,
			dataplane: `
            networking:
              address: 192.168.0.1
              inbound:
                - port: 80
                  servicePort: 8080
                  tags:
                    kuma.io/service: backend
              outbound:
              - port: 54321
                service: db
              - port: 59200
                service: elastic
`,
			profile:         mesh_core.ProfileDefaultProxy,
			envoyConfigFile: "1-envoy-config.golden.yaml",
		}),
		Entry("should support pre-defined `default-proxy` profile; transparent_proxying=true", testCase{
			mesh: `
            mtls:
              enabledBackend: builtin
              backends:
              - type: builtin
                name: builtin
`,
			dataplane: `
            networking:
              address: 192.168.0.1
              inbound:
                - port: 80
                  servicePort: 8080
                  tags:
                    kuma.io/service: backend
              outbound:
              - port: 54321
                service: db
              - port: 59200
                service: elastic
              transparentProxying:
                redirectPortOutbound: 15001
                redirectPortInbound: 15006
`,
			profile:         mesh_core.ProfileDefaultProxy,
			envoyConfigFile: "2-envoy-config.golden.yaml",
		}),
		Entry("should support pre-defined `default-proxy` profile; transparent_proxying=false; prometheus_metrics=true", testCase{
			mesh: `
            mtls:
              enabledBackend: builtin
              backends:
              - type: builtin
                name: builtin
            metrics:
              enabledBackend: prometheus-1
              backends:
              - name: prometheus-1
                type: prometheus
                conf:
                  port: 1234
                  path: /non-standard-path
                  skipMTLS: false
`,
			dataplane: `
            networking:
              address: 192.168.0.1
              inbound:
                - port: 80
                  servicePort: 8080
                  tags:
                    kuma.io/service: backend
                    kuma.io/protocol: http
              outbound:
              - port: 54321
                service: db
              - port: 59200
                service: elastic
`,
			profile:         mesh_core.ProfileDefaultProxy,
			envoyConfigFile: "3-envoy-config.golden.yaml",
		}),
		Entry("should support pre-defined `default-proxy` profile; transparent_proxying=true; prometheus_metrics=true", testCase{
			mesh: `
            mtls:
              enabledBackend: builtin
              backends:
              - type: builtin
                name: builtin
            metrics:
              enabledBackend: prometheus-1
              backends:
              - name: prometheus-1
                type: prometheus
                conf:
                  port: 1234
                  path: /non-standard-path
                  skipMTLS: false
`,
			dataplane: `
            networking:
              address: 192.168.0.1
              inbound:
                - port: 80
                  servicePort: 8080
                  tags:
                    kuma.io/service: backend
                    kuma.io/protocol: http
              outbound:
              - port: 54321
                service: db
              - port: 59200
                service: elastic
              transparentProxying:
                redirectPortOutbound: 15001
                redirectPortInbound: 15006
`,
			profile:         mesh_core.ProfileDefaultProxy,
			envoyConfigFile: "4-envoy-config.golden.yaml",
		}),
	)
})
