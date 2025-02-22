package context

import (
	"io/ioutil"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

type Context struct {
	ControlPlane   *ControlPlaneContext
	Mesh           MeshContext
	ConnectionInfo ConnectionInfo
}

type ConnectionInfo struct {
	// Authority defines the URL that was used by the data plane to connect to the control plane
	Authority string
}

type ControlPlaneContext struct {
	SdsTlsCert []byte
}

func (c Context) SDSLocation() string {
	// SDS lives on the same server as XDS so we can use the URL that Dataplane used to connect to XDS
	return c.ConnectionInfo.Authority
}

type MeshContext struct {
	Resource   *mesh_core.MeshResource
	Dataplanes *mesh_core.DataplaneResourceList
}

func BuildControlPlaneContext(config kuma_cp.Config) (*ControlPlaneContext, error) {
	var cert []byte
	if config.DpServer.TlsCertFile != "" {
		c, err := ioutil.ReadFile(config.DpServer.TlsCertFile)
		if err != nil {
			return nil, err
		}
		cert = c
	}

	return &ControlPlaneContext{
		SdsTlsCert: cert,
	}, nil
}
