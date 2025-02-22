package cla

import (
	"context"
	"fmt"
	"time"

	"github.com/kumahq/kuma/pkg/core/datasource"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/metrics"

	envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/patrickmn/go-cache"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/xds/cache/once"
	"github.com/kumahq/kuma/pkg/xds/envoy/endpoints"
	"github.com/kumahq/kuma/pkg/xds/topology"
)

var (
	claCacheLog = core.Log.WithName("cla-cache")
)

// Cache is needed to share and cache ClusterLoadAssignments among goroutines
// which reconcile Dataplane's state. In scope of one mesh ClusterLoadAssignment
// will be the same for each service so no need to reconcile for each dataplane.
type Cache struct {
	cache   *cache.Cache
	rm      manager.ReadOnlyResourceManager
	dsl     datasource.Loader
	ipFunc  lookup.LookupIPFunc
	zone    string
	onceMap *once.Map
	metrics *prometheus.GaugeVec
}

func NewCache(
	rm manager.ReadOnlyResourceManager,
	dsl datasource.Loader,
	zone string, expirationTime time.Duration,
	ipFunc lookup.LookupIPFunc,
	metrics metrics.Metrics,
) (*Cache, error) {
	metric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cla_cache",
		Help: "Summary of CLA Cache",
	}, []string{"operation", "result"})
	if err := metrics.Register(metric); err != nil {
		return nil, err
	}
	return &Cache{
		cache:   cache.New(expirationTime, time.Duration(int64(float64(expirationTime)*0.9))),
		rm:      rm,
		dsl:     dsl,
		zone:    zone,
		ipFunc:  ipFunc,
		onceMap: once.NewMap(),
		metrics: metric,
	}, nil
}

func (c *Cache) GetCLA(ctx context.Context, meshName, service string) (*envoy_api_v2.ClusterLoadAssignment, error) {
	key := fmt.Sprintf("%s:%s", meshName, service)
	value, found := c.cache.Get(key)
	if found {
		c.metrics.WithLabelValues("get", "hit").Inc()
		return value.(*envoy_api_v2.ClusterLoadAssignment), nil
	}
	o := c.onceMap.Get(key)
	c.metrics.WithLabelValues("get", "hit-wait").Inc()
	o.Do(func() (interface{}, error) {
		c.metrics.WithLabelValues("get", "hit-wait").Dec()
		c.metrics.WithLabelValues("get", "miss").Inc()
		dataplanes, err := topology.GetDataplanes(claCacheLog, ctx, c.rm, c.ipFunc, meshName)
		if err != nil {
			return nil, err
		}
		mesh := &core_mesh.MeshResource{}
		if err := c.rm.Get(ctx, mesh, core_store.GetByKey(meshName, model.NoMesh)); err != nil {
			return nil, err
		}
		externalServices := &core_mesh.ExternalServiceResourceList{}
		if err := c.rm.List(ctx, externalServices, core_store.ListByMesh(meshName)); err != nil {
			return nil, err
		}
		endpointMap := topology.BuildEndpointMap(mesh, c.zone, dataplanes.Items, externalServices.Items, c.dsl)
		cla := endpoints.CreateClusterLoadAssignment(service, endpointMap[service])
		c.cache.SetDefault(key, cla)
		c.onceMap.Delete(key)
		return cla, nil
	})
	if o.Err != nil {
		return nil, o.Err
	}
	return o.Value.(*envoy_api_v2.ClusterLoadAssignment), nil
}
