// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	v1alpha1 "open-cluster-management.io/cluster-proxy/pkg/apis/proxy/v1alpha1"
)

// ManagedProxyConfigurationLister helps list ManagedProxyConfigurations.
// All objects returned here must be treated as read-only.
type ManagedProxyConfigurationLister interface {
	// List lists all ManagedProxyConfigurations in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.ManagedProxyConfiguration, err error)
	// Get retrieves the ManagedProxyConfiguration from the index for a given name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.ManagedProxyConfiguration, error)
	ManagedProxyConfigurationListerExpansion
}

// managedProxyConfigurationLister implements the ManagedProxyConfigurationLister interface.
type managedProxyConfigurationLister struct {
	indexer cache.Indexer
}

// NewManagedProxyConfigurationLister returns a new ManagedProxyConfigurationLister.
func NewManagedProxyConfigurationLister(indexer cache.Indexer) ManagedProxyConfigurationLister {
	return &managedProxyConfigurationLister{indexer: indexer}
}

// List lists all ManagedProxyConfigurations in the indexer.
func (s *managedProxyConfigurationLister) List(selector labels.Selector) (ret []*v1alpha1.ManagedProxyConfiguration, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.ManagedProxyConfiguration))
	})
	return ret, err
}

// Get retrieves the ManagedProxyConfiguration from the index for a given name.
func (s *managedProxyConfigurationLister) Get(name string) (*v1alpha1.ManagedProxyConfiguration, error) {
	obj, exists, err := s.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("managedproxyconfiguration"), name)
	}
	return obj.(*v1alpha1.ManagedProxyConfiguration), nil
}