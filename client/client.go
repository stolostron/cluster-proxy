package client

import (
	"context"
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"open-cluster-management.io/cluster-proxy/pkg/generated/clientset/versioned"
)

func GetServiceURL(ctx context.Context, kubeconfig *rest.Config, clusterName string, namespace string, serviceName string) (string, error) {
	managedproxyconfigurateionClient := versioned.NewForConfigOrDie(kubeconfig)
	config, err := managedproxyconfigurateionClient.ProxyV1alpha1().ManagedProxyConfigurations().Get(ctx, "cluster-proxy", v1.GetOptions{})
	if err != nil {
		return "", err
	}

	for _, su := range config.Status.ServiceURLs {
		if su.ManagedCluster == clusterName && su.Namespace == namespace && su.ServiceName == serviceName {
			return su.URL, nil
		}
	}

	return "", fmt.Errorf("target (in cluster:%s, namespace: %s) service %s, url not found", clusterName, namespace, serviceName)
}
