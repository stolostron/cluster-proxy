package eventhandler

import (
	"context"

	"k8s.io/client-go/util/workqueue"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ handler.EventHandler = &ManagedProxyServiceResolverHandler{}

type ManagedProxyServiceResolverHandler struct {
	client.Client
}

func (m ManagedProxyServiceResolverHandler) Create(event event.CreateEvent, limitingInterface workqueue.RateLimitingInterface) {
	m.findClusterProxyAddon(limitingInterface)
}

func (m ManagedProxyServiceResolverHandler) Update(event event.UpdateEvent, limitingInterface workqueue.RateLimitingInterface) {
	m.findClusterProxyAddon(limitingInterface)
}

func (m ManagedProxyServiceResolverHandler) Delete(event event.DeleteEvent, limitingInterface workqueue.RateLimitingInterface) {
	m.findClusterProxyAddon(limitingInterface)
}

func (m ManagedProxyServiceResolverHandler) Generic(event event.GenericEvent, limitingInterface workqueue.RateLimitingInterface) {
	m.findClusterProxyAddon(limitingInterface)
}

func (m *ManagedProxyServiceResolverHandler) findClusterProxyAddon(limitingInterface workqueue.RateLimitingInterface) {
	list := &addonv1alpha1.ClusterManagementAddOnList{}
	err := m.Client.List(context.TODO(), list)
	if err != nil {
		return
	}
	for _, addon := range list.Items {
		if addon.Spec.AddOnConfiguration.CRDName == crdName {
			req := reconcile.Request{}
			req.Namespace = addon.Namespace
			req.Name = addon.Name
			limitingInterface.Add(req)
		}
	}
}
