package tls_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/xds/envoy/tls"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("CreateDownstreamTlsContext()", func() {

	Context("when mTLS is disabled on a given Mesh", func() {

		It("should return `nil`", func() {
			// given
			ctx := xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: &mesh_core.MeshResource{},
				},
			}
			metadata := &core_xds.DataplaneMetadata{}

			// when
			snippet, err := tls.CreateDownstreamTlsContext(ctx, metadata)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(snippet).To(BeNil())
		})
	})

	Context("when mTLS is enabled on a given Mesh", func() {

		type testCase struct {
			metadata *core_xds.DataplaneMetadata
			expected string
		}

		DescribeTable("should generate proper Envoy config",
			func(given testCase) {
				// given
				ctx := xds_context.Context{
					ConnectionInfo: xds_context.ConnectionInfo{
						Authority: "kuma-control-plane:5677",
					},
					ControlPlane: &xds_context.ControlPlaneContext{
						SdsTlsCert: []byte("CERTIFICATE"),
					},
					Mesh: xds_context.MeshContext{
						Resource: &mesh_core.MeshResource{
							Meta: &test_model.ResourceMeta{
								Name: "default",
							},
							Spec: mesh_proto.Mesh{
								Mtls: &mesh_proto.Mesh_Mtls{
									EnabledBackend: "builtin",
									Backends: []*mesh_proto.CertificateAuthorityBackend{
										{
											Name: "builtin",
											Type: "builtin",
										},
									},
								},
							},
						},
					},
				}

				// when
				snippet, err := tls.CreateDownstreamTlsContext(ctx, given.metadata)
				// then
				Expect(err).ToNot(HaveOccurred())
				// when
				actual, err := util_proto.ToYAML(snippet)
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("metadata is `nil`", testCase{
				metadata: nil,
				expected: `
                commonTlsContext:
                  combinedValidationContext:
                    defaultValidationContext:
                      matchSubjectAltNames:
                      - prefix: spiffe://default/
                    validationContextSdsSecretConfig:
                      name: mesh_ca
                      sdsConfig:
                        apiConfigSource:
                          apiType: GRPC
                          grpcServices:
                          - googleGrpc:
                              channelCredentials:
                                sslCredentials:
                                  rootCerts:
                                    inlineBytes: Q0VSVElGSUNBVEU=
                              statPrefix: sds_mesh_ca
                              targetUri: kuma-control-plane:5677
                  tlsCertificateSdsSecretConfigs:
                  - name: identity_cert
                    sdsConfig:
                      apiConfigSource:
                        apiType: GRPC
                        grpcServices:
                        - googleGrpc:
                            channelCredentials:
                              sslCredentials:
                                rootCerts:
                                  inlineBytes: Q0VSVElGSUNBVEU=
                            statPrefix: sds_identity_cert
                            targetUri: kuma-control-plane:5677
                requireClientCertificate: true
`,
			}),
			Entry("dataplane without a token", testCase{
				metadata: &core_xds.DataplaneMetadata{},
				expected: `
                commonTlsContext:
                  combinedValidationContext:
                    defaultValidationContext:
                      matchSubjectAltNames:
                      - prefix: spiffe://default/
                    validationContextSdsSecretConfig:
                      name: mesh_ca
                      sdsConfig:
                        apiConfigSource:
                          apiType: GRPC
                          grpcServices:
                          - googleGrpc:
                              channelCredentials:
                                sslCredentials:
                                  rootCerts:
                                    inlineBytes: Q0VSVElGSUNBVEU=
                              statPrefix: sds_mesh_ca
                              targetUri: kuma-control-plane:5677
                  tlsCertificateSdsSecretConfigs:
                  - name: identity_cert
                    sdsConfig:
                      apiConfigSource:
                        apiType: GRPC
                        grpcServices:
                        - googleGrpc:
                            channelCredentials:
                              sslCredentials:
                                rootCerts:
                                  inlineBytes: Q0VSVElGSUNBVEU=
                            statPrefix: sds_identity_cert
                            targetUri: kuma-control-plane:5677
                requireClientCertificate: true
`,
			}),
			Entry("dataplane with a token", testCase{

				metadata: &core_xds.DataplaneMetadata{
					DataplaneTokenPath: "/path/to/token",
				},
				expected: `
                commonTlsContext:
                  combinedValidationContext:
                    defaultValidationContext:
                      matchSubjectAltNames:
                      - prefix: spiffe://default/
                    validationContextSdsSecretConfig:
                      name: mesh_ca
                      sdsConfig:
                        apiConfigSource:
                          apiType: GRPC
                          grpcServices:
                          - googleGrpc:
                              callCredentials:
                              - fromPlugin:
                                  name: envoy.grpc_credentials.file_based_metadata
                                  typedConfig:
                                    '@type': type.googleapis.com/envoy.config.grpc_credential.v2alpha.FileBasedMetadataConfig
                                    secretData:
                                      filename: /path/to/token
                              channelCredentials:
                                sslCredentials:
                                  rootCerts:
                                    inlineBytes: Q0VSVElGSUNBVEU=
                              credentialsFactoryName: envoy.grpc_credentials.file_based_metadata
                              statPrefix: sds_mesh_ca
                              targetUri: kuma-control-plane:5677
                  tlsCertificateSdsSecretConfigs:
                  - name: identity_cert
                    sdsConfig:
                      apiConfigSource:
                        apiType: GRPC
                        grpcServices:
                        - googleGrpc:
                            callCredentials:
                            - fromPlugin:
                                name: envoy.grpc_credentials.file_based_metadata
                                typedConfig:
                                  '@type': type.googleapis.com/envoy.config.grpc_credential.v2alpha.FileBasedMetadataConfig
                                  secretData:
                                    filename: /path/to/token
                            channelCredentials:
                              sslCredentials:
                                rootCerts:
                                  inlineBytes: Q0VSVElGSUNBVEU=
                            credentialsFactoryName: envoy.grpc_credentials.file_based_metadata
                            statPrefix: sds_identity_cert
                            targetUri: kuma-control-plane:5677
                requireClientCertificate: true
`,
			}),
		)
	})
})

var _ = Describe("CreateUpstreamTlsContext()", func() {

	Context("when mTLS is disabled on a given Mesh", func() {

		It("should return `nil`", func() {
			// given
			ctx := xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: &mesh_core.MeshResource{},
				},
			}
			metadata := &core_xds.DataplaneMetadata{}

			// when
			snippet, err := tls.CreateUpstreamTlsContext(ctx, metadata, "backend", "backend")
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(snippet).To(BeNil())
		})
	})

	Context("when mTLS is enabled on a given Mesh", func() {

		type testCase struct {
			metadata        *core_xds.DataplaneMetadata
			upstreamService string
			expected        string
		}

		DescribeTable("should generate proper Envoy config",
			func(given testCase) {
				// given
				ctx := xds_context.Context{
					ConnectionInfo: xds_context.ConnectionInfo{
						Authority: "kuma-control-plane:5677",
					},
					ControlPlane: &xds_context.ControlPlaneContext{
						SdsTlsCert: []byte("CERTIFICATE"),
					},
					Mesh: xds_context.MeshContext{
						Resource: &mesh_core.MeshResource{
							Meta: &test_model.ResourceMeta{
								Name: "default",
							},
							Spec: mesh_proto.Mesh{
								Mtls: &mesh_proto.Mesh_Mtls{
									EnabledBackend: "builtin",
									Backends: []*mesh_proto.CertificateAuthorityBackend{
										{
											Name: "builtin",
											Type: "builtin",
										},
									},
								},
							},
						},
					},
				}

				// when
				snippet, err := tls.CreateUpstreamTlsContext(ctx, given.metadata, given.upstreamService, "")
				// then
				Expect(err).ToNot(HaveOccurred())
				// when
				actual, err := util_proto.ToYAML(snippet)
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("metadata is `nil`", testCase{
				metadata:        nil,
				upstreamService: "backend",
				expected: `
                commonTlsContext:
                  combinedValidationContext:
                    defaultValidationContext:
                      matchSubjectAltNames:
                      - exact: spiffe://default/backend
                    validationContextSdsSecretConfig:
                      name: mesh_ca
                      sdsConfig:
                        apiConfigSource:
                          apiType: GRPC
                          grpcServices:
                          - googleGrpc:
                              channelCredentials:
                                sslCredentials:
                                  rootCerts:
                                    inlineBytes: Q0VSVElGSUNBVEU=
                              statPrefix: sds_mesh_ca
                              targetUri: kuma-control-plane:5677
                  tlsCertificateSdsSecretConfigs:
                  - name: identity_cert
                    sdsConfig:
                      apiConfigSource:
                        apiType: GRPC
                        grpcServices:
                        - googleGrpc:
                            channelCredentials:
                              sslCredentials:
                                rootCerts:
                                  inlineBytes: Q0VSVElGSUNBVEU=
                            statPrefix: sds_identity_cert
                            targetUri: kuma-control-plane:5677
`,
			}),
			Entry("dataplane without a token", testCase{
				metadata:        &core_xds.DataplaneMetadata{},
				upstreamService: "backend",
				expected: `
                commonTlsContext:
                  combinedValidationContext:
                    defaultValidationContext:
                      matchSubjectAltNames:
                      - exact: spiffe://default/backend
                    validationContextSdsSecretConfig:
                      name: mesh_ca
                      sdsConfig:
                        apiConfigSource:
                          apiType: GRPC
                          grpcServices:
                          - googleGrpc:
                              channelCredentials:
                                sslCredentials:
                                  rootCerts:
                                    inlineBytes: Q0VSVElGSUNBVEU=
                              statPrefix: sds_mesh_ca
                              targetUri: kuma-control-plane:5677
                  tlsCertificateSdsSecretConfigs:
                  - name: identity_cert
                    sdsConfig:
                      apiConfigSource:
                        apiType: GRPC
                        grpcServices:
                        - googleGrpc:
                            channelCredentials:
                              sslCredentials:
                                rootCerts:
                                  inlineBytes: Q0VSVElGSUNBVEU=
                            statPrefix: sds_identity_cert
                            targetUri: kuma-control-plane:5677
`,
			}),
			Entry("dataplane with a token", testCase{
				upstreamService: "backend",
				metadata: &core_xds.DataplaneMetadata{
					DataplaneTokenPath: "/path/to/token",
				},
				expected: `
                commonTlsContext:
                  combinedValidationContext:
                    defaultValidationContext:
                      matchSubjectAltNames:
                      - exact: spiffe://default/backend
                    validationContextSdsSecretConfig:
                      name: mesh_ca
                      sdsConfig:
                        apiConfigSource:
                          apiType: GRPC
                          grpcServices:
                          - googleGrpc:
                              callCredentials:
                              - fromPlugin:
                                  name: envoy.grpc_credentials.file_based_metadata
                                  typedConfig:
                                    '@type': type.googleapis.com/envoy.config.grpc_credential.v2alpha.FileBasedMetadataConfig
                                    secretData:
                                      filename: /path/to/token
                              channelCredentials:
                                sslCredentials:
                                  rootCerts:
                                    inlineBytes: Q0VSVElGSUNBVEU=
                              credentialsFactoryName: envoy.grpc_credentials.file_based_metadata
                              statPrefix: sds_mesh_ca
                              targetUri: kuma-control-plane:5677
                  tlsCertificateSdsSecretConfigs:
                  - name: identity_cert
                    sdsConfig:
                      apiConfigSource:
                        apiType: GRPC
                        grpcServices:
                        - googleGrpc:
                            callCredentials:
                            - fromPlugin:
                                name: envoy.grpc_credentials.file_based_metadata
                                typedConfig:
                                  '@type': type.googleapis.com/envoy.config.grpc_credential.v2alpha.FileBasedMetadataConfig
                                  secretData:
                                    filename: /path/to/token
                            channelCredentials:
                              sslCredentials:
                                rootCerts:
                                  inlineBytes: Q0VSVElGSUNBVEU=
                            credentialsFactoryName: envoy.grpc_credentials.file_based_metadata
                            statPrefix: sds_identity_cert
                            targetUri: kuma-control-plane:5677
`,
			}),
		)
	})
})
