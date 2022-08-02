package client

import (
	"context"
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"open-cluster-management.io/cluster-proxy/pkg/generated/clientset/versioned"
	"open-cluster-management.io/cluster-proxy/pkg/util"
)

func GetServiceURL(ctx context.Context, kubeconfig *rest.Config, clusterName string, namespace string, serviceName string) (string, error) {
	managedproxyconfigurateionClient := versioned.NewForConfigOrDie(kubeconfig)
	config, err := managedproxyconfigurateionClient.ProxyV1alpha1().ManagedProxyConfigurations().Get(ctx, "cluster-proxy", v1.GetOptions{})
	if err != nil {
		return "", err
	}

	for _, sr := range config.Spec.ServiceResolvers {
		if sr.Namespace != namespace {
			continue
		}
		if sr.ServiceName != serviceName {
			continue
		}
		if sr.ManagedCluster != "" && sr.ManagedCluster != clusterName {
			continue
		}
		// TODO consider how to using lableSeletor to find if a service is exist on target managedCluster
		// 1. get managedclusters that can pass the lableSelector
		// 2. if one of the above step's result matching current
		managdclusterClient, err := versioned.NewForConfigOrDie(kubeconfig)
		return util.GenerateServiceURL(clusterName, namespace, serviceName), nil
	}

	return "", fmt.Errorf("target (in cluster:%s, namespace: %s) service %s, url not found", clusterName, namespace, serviceName)
}
