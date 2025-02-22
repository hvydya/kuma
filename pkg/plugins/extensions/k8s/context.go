package k8s

import (
	"context"

	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"

	k8s_common "github.com/kumahq/kuma/pkg/plugins/common/k8s"
)

type managerKey struct{}

func NewManagerContext(ctx context.Context, manager kube_ctrl.Manager) context.Context {
	return context.WithValue(ctx, managerKey{}, manager)
}

func FromManagerContext(ctx context.Context) (manager kube_ctrl.Manager, ok bool) {
	manager, ok = ctx.Value(managerKey{}).(kube_ctrl.Manager)
	return
}

// One instance of Converter needs to be shared across resource plugin and runtime
// plugin if CachedConverter is used, only one instance is created, otherwise we would
// have all cached resources in the memory twice.

type converterKey struct{}

func NewResourceConverterContext(ctx context.Context, converter k8s_common.Converter) context.Context {
	return context.WithValue(ctx, converterKey{}, converter)
}

func FromResourceConverterContext(ctx context.Context) (converter k8s_common.Converter, ok bool) {
	converter, ok = ctx.Value(converterKey{}).(k8s_common.Converter)
	return
}

type nonCachedClient struct{}

func NewNonCachedClientContext(ctx context.Context, client kube_client.Client) context.Context {
	return context.WithValue(ctx, nonCachedClient{}, client)
}

func FromNonCachedClientContext(ctx context.Context) (client kube_client.Client, ok bool) {
	client, ok = ctx.Value(nonCachedClient{}).(kube_client.Client)
	return
}
