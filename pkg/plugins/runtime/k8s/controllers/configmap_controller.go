package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_record "k8s.io/client-go/tools/record"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_handler "sigs.k8s.io/controller-runtime/pkg/handler"
	kube_reconile "sigs.k8s.io/controller-runtime/pkg/reconcile"
	kube_source "sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/kumahq/kuma/pkg/dns/vips"

	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/dns"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
)

// ConfigMapReconciler reconciles a ConfigMap object
type ConfigMapReconciler struct {
	kube_client.Client
	kube_record.EventRecorder
	Scheme          *kube_runtime.Scheme
	Log             logr.Logger
	ResourceManager manager.ResourceManager
	VIPsAllocator   *dns.VIPsAllocator
	SystemNamespace string
}

func (r *ConfigMapReconciler) Reconcile(req kube_ctrl.Request) (kube_ctrl.Result, error) {
	mesh, ok := vips.MeshFromConfigKey(req.Name)
	if !ok {
		return kube_ctrl.Result{}, nil
	}

	if err := r.VIPsAllocator.CreateOrUpdateVIPConfig(mesh); err != nil {
		return kube_ctrl.Result{}, err
	}

	return kube_ctrl.Result{}, nil
}

func (r *ConfigMapReconciler) SetupWithManager(mgr kube_ctrl.Manager) error {
	for _, addToScheme := range []func(*kube_runtime.Scheme) error{kube_core.AddToScheme, mesh_k8s.AddToScheme} {
		if err := addToScheme(mgr.GetScheme()); err != nil {
			return err
		}
	}
	return kube_ctrl.NewControllerManagedBy(mgr).
		For(&kube_core.ConfigMap{}).
		Watches(&kube_source.Kind{Type: &kube_core.Service{}}, &kube_handler.EnqueueRequestsFromMapFunc{
			ToRequests: &ServiceToConfigMapsMapper{
				Client:          mgr.GetClient(),
				Log:             r.Log.WithName("service-to-configmap-mapper"),
				SystemNamespace: r.SystemNamespace,
			},
		}).
		Watches(&kube_source.Kind{Type: &mesh_k8s.Dataplane{}}, &kube_handler.EnqueueRequestsFromMapFunc{
			ToRequests: &DataplaneToMeshMapper{
				Client:          mgr.GetClient(),
				Log:             r.Log.WithName("dataplane-to-configmap-mapper"),
				SystemNamespace: r.SystemNamespace,
			},
		}).
		Watches(&kube_source.Kind{Type: &mesh_k8s.ExternalService{}}, &kube_handler.EnqueueRequestsFromMapFunc{
			ToRequests: &ExternalServiceToConfigMapsMapper{
				Client:          mgr.GetClient(),
				Log:             r.Log.WithName("external-service-to-configmap-mapperr"),
				SystemNamespace: r.SystemNamespace,
			},
		}).
		Complete(r)
}

type ServiceToConfigMapsMapper struct {
	kube_client.Client
	Log             logr.Logger
	SystemNamespace string
}

func (m *ServiceToConfigMapsMapper) Map(obj kube_handler.MapObject) []kube_reconile.Request {
	cause, ok := obj.Object.(*kube_core.Service)
	if !ok {
		m.Log.WithValues("dataplane", obj.Meta).Error(errors.Errorf("wrong argument type: expected %T, got %T", cause, obj.Object), "wrong argument type")
		return nil
	}

	ctx := context.Background()
	svcName := fmt.Sprintf("%s/%s", cause.Namespace, cause.Name)
	// List Pods in the same namespace
	pods := &kube_core.PodList{}
	if err := m.Client.List(ctx, pods, kube_client.InNamespace(obj.Meta.GetNamespace())); err != nil {
		m.Log.WithValues("service", svcName).Error(err, "failed to fetch Dataplanes in namespace")
		return nil
	}

	meshSet := map[string]bool{}
	for _, pod := range pods.Items {
		if mesh, exist := metadata.Annotations(pod.Annotations).GetString(metadata.KumaMeshAnnotation); exist {
			meshSet[mesh] = true
		}
	}
	var req []kube_reconile.Request
	for mesh := range meshSet {
		req = append(req, kube_reconile.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: m.SystemNamespace, Name: vips.ConfigKey(mesh)},
		})
	}

	return req
}

type DataplaneToMeshMapper struct {
	kube_client.Client
	Log             logr.Logger
	SystemNamespace string
}

func (m *DataplaneToMeshMapper) Map(obj kube_handler.MapObject) []kube_reconile.Request {
	cause, ok := obj.Object.(*mesh_k8s.Dataplane)
	if !ok {
		m.Log.WithValues("dataplane", obj.Meta).Error(errors.Errorf("wrong argument type: expected %T, got %T", cause, obj.Object), "wrong argument type")
		return nil
	}

	return []kube_reconile.Request{{
		NamespacedName: kube_types.NamespacedName{Namespace: m.SystemNamespace, Name: vips.ConfigKey(cause.Mesh)},
	}}
}

type ExternalServiceToConfigMapsMapper struct {
	kube_client.Client
	Log             logr.Logger
	SystemNamespace string
}

func (m *ExternalServiceToConfigMapsMapper) Map(obj kube_handler.MapObject) []kube_reconile.Request {
	cause, ok := obj.Object.(*mesh_k8s.ExternalService)
	if !ok {
		m.Log.WithValues("externalService", obj.Meta).Error(errors.Errorf("wrong argument type: expected %T, got %T", cause, obj.Object), "wrong argument type")
		return nil
	}

	return []kube_reconile.Request{{
		NamespacedName: kube_types.NamespacedName{Namespace: m.SystemNamespace, Name: vips.ConfigKey(cause.Mesh)},
	}}
}
