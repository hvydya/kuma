package auth

import (
	"context"
	"sync"

	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_server "github.com/envoyproxy/go-control-plane/pkg/server/v2"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

const authorization = "authorization"

func NewCallbacks(resManager core_manager.ResourceManager, authenticator Authenticator) envoy_server.Callbacks {
	return &authCallbacks{
		resManager:    resManager,
		authenticator: authenticator,
		contexts:      map[core_xds.StreamID]context.Context{},
		authenticated: map[core_xds.StreamID]string{},
	}
}

// authCallback checks if the DiscoveryRequest is authorized, ie. if it has a valid Dataplane Token/Service Account Token.
type authCallbacks struct {
	resManager    core_manager.ResourceManager
	authenticator Authenticator

	sync.RWMutex // protects contexts and authenticated
	// contexts stores context for every stream, since Context from which we can extract auth data is only available in OnStreamOpen
	contexts map[core_xds.StreamID]context.Context
	// authenticated stores authenticated ProxyID for stream. We don't want to authenticate every because since on K8S we execute ReviewToken which is expensive
	// as long as client won't change ProxyID it's safe to authenticate only once.
	authenticated map[core_xds.StreamID]string
}

var _ envoy_server.Callbacks = &authCallbacks{}

func (a *authCallbacks) OnStreamOpen(ctx context.Context, streamID core_xds.StreamID, _ string) error {
	a.Lock()
	defer a.Unlock()

	a.contexts[streamID] = ctx
	return nil
}

func (a *authCallbacks) OnStreamClosed(streamID core_xds.StreamID) {
	a.Lock()
	delete(a.contexts, streamID)
	delete(a.authenticated, streamID)
	a.Unlock()
}

func (a *authCallbacks) OnStreamRequest(streamID core_xds.StreamID, req *envoy_api.DiscoveryRequest) error {
	if id, alreadyAuthenticated := a.authNodeId(streamID); alreadyAuthenticated {
		if req.Node != nil && req.Node.Id != id {
			return errors.Errorf("stream was authenticated for ID %s. Received request is for node with ID %s. Node ID cannot be changed after stream is initialized", id, req.Node.Id)
		}
		return nil
	}

	credential, err := a.credential(streamID)
	if err != nil {
		return err
	}
	err = a.authenticate(credential, req)
	if err != nil {
		return err
	}
	a.Lock()
	a.authenticated[streamID] = req.Node.Id
	a.Unlock()
	return nil
}

func (a *authCallbacks) authNodeId(streamID core_xds.StreamID) (string, bool) {
	a.RLock()
	defer a.RUnlock()
	id, ok := a.authenticated[streamID]
	return id, ok
}

func (a *authCallbacks) credential(streamID core_xds.StreamID) (Credential, error) {
	a.RLock()
	defer a.RUnlock()

	ctx, exists := a.contexts[streamID]
	if !exists {
		return "", errors.Errorf("there is no context for stream ID %d", streamID)
	}
	credential, err := extractCredential(ctx)
	if err != nil {
		return "", errors.Wrap(err, "could not extract credential from DiscoveryRequest")
	}
	return credential, err
}

func (a *authCallbacks) authenticate(credential Credential, req *envoy_api.DiscoveryRequest) error {
	dataplane := &core_mesh.DataplaneResource{}
	md := core_xds.DataplaneMetadataFromNode(req.Node)
	if md.DataplaneResource != nil {
		dataplane = md.DataplaneResource
	} else {
		proxyId, err := core_xds.ParseProxyId(req.Node)
		if err != nil {
			return errors.Wrap(err, "SDS request must have a valid Proxy Id")
		}
		err = a.resManager.Get(context.Background(), dataplane, core_store.GetByKey(proxyId.Name, proxyId.Mesh))
		if err != nil {
			if core_store.IsResourceNotFound(err) {
				return errors.New("dataplane not found. Create Dataplane in Kuma CP first or pass it as an argument to kuma-dp")
			}
			return err
		}
	}

	if err := a.authenticator.Authenticate(context.Background(), dataplane, credential); err != nil {
		return errors.Wrap(err, "authentication failed")
	}
	return nil
}

func (a *authCallbacks) OnStreamResponse(core_xds.StreamID, *envoy_api.DiscoveryRequest, *envoy_api.DiscoveryResponse) {
}

func (a *authCallbacks) OnFetchRequest(ctx context.Context, request *envoy_api.DiscoveryRequest) error {
	credential, err := extractCredential(ctx)
	if err != nil {
		return errors.Wrap(err, "could not extract credential from DiscoveryRequest")
	}
	return a.authenticate(credential, request)
}

func (a *authCallbacks) OnFetchResponse(*envoy_api.DiscoveryRequest, *envoy_api.DiscoveryResponse) {
}

var _ envoy_server.Callbacks = &authCallbacks{}

func extractCredential(ctx context.Context) (Credential, error) {
	metadata, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.Errorf("request has no metadata")
	}
	if values, ok := metadata[authorization]; ok {
		if len(values) != 1 {
			return "", errors.Errorf("request must have exactly 1 %q header, got %d", authorization, len(values))
		}
		return values[0], nil
	}
	return "", nil
}
