package client

import (
	"context"
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/rest"
	clusterv1client "open-cluster-management.io/api/client/cluster/clientset/versioned"
	"open-cluster-management.io/cluster-proxy/pkg/generated/clientset/versioned"
	"open-cluster-management.io/cluster-proxy/pkg/util"
	clustersdkv1beta2 "open-cluster-management.io/sdk-go/pkg/apis/cluster/v1beta2"
)

func GetProxyHost(ctx context.Context, kubeconfig *rest.Config, clusterName string, namespace string, serviceName string) (string, error) {
	client := versioned.NewForConfigOrDie(kubeconfig)
	mpsrList, err := client.ProxyV1alpha1().ManagedProxyServiceResolvers().List(ctx, v1.ListOptions{})
	if err != nil {
		return "", err
	}

	// Get labels of the managedCluster
	clusterClient, err := clusterv1client.NewForConfig(kubeconfig)
	if err != nil {
		return "", err
	}
	managedCluster, err := clusterClient.ClusterV1().ManagedClusters().Get(ctx, clusterName, v1.GetOptions{})
	if err != nil {
		return "", err
	}

	fmt.Printf("DEBUG: Found %d ManagedProxyServiceResolvers\n", len(mpsrList.Items))
	for _, mpsr := range mpsrList.Items {
		fmt.Printf("DEBUG: MPSR %s: serviceRef=%s/%s, clusterSet=%s\n",
			mpsr.Name,
			mpsr.Spec.ServiceSelector.ServiceRef.Namespace,
			mpsr.Spec.ServiceSelector.ServiceRef.Name,
			mpsr.Spec.ManagedClusterSelector.ManagedClusterSet.Name)
	}
	fmt.Printf("DEBUG: ManagedCluster %s labels: %v\n", clusterName, managedCluster.Labels)

	// Return when namespace, serviceName and labels of the managedCluster are all matched
	for _, sr := range mpsrList.Items {
		if !util.IsServiceResolverLegal(&sr) {
			fmt.Printf("DEBUG: MPSR %s is not legal\n", sr.Name)
			continue
		}

		set, err := clusterClient.ClusterV1beta2().ManagedClusterSets().Get(ctx, sr.Spec.ManagedClusterSelector.ManagedClusterSet.Name, v1.GetOptions{})
		if err != nil {
			fmt.Printf("DEBUG: Failed to get ManagedClusterSet %s: %v\n", sr.Spec.ManagedClusterSelector.ManagedClusterSet.Name, err)
			return "", err
		}
		selector, err := clustersdkv1beta2.BuildClusterSelector(set)
		if err != nil {
			fmt.Printf("DEBUG: Failed to build ClusterSelector for ManagedClusterSet %s: %v\n", sr.Spec.ManagedClusterSelector.ManagedClusterSet.Name, err)
			return "", err
		}
		if !selector.Matches(labels.Set(managedCluster.Labels)) {
			fmt.Printf("DEBUG: ManagedCluster %s does not match ManagedClusterSet %s\n", clusterName, sr.Spec.ManagedClusterSelector.ManagedClusterSet.Name)
			continue
		}

		if sr.Spec.ServiceSelector.ServiceRef.Namespace != namespace || sr.Spec.ServiceSelector.ServiceRef.Name != serviceName {
			fmt.Printf("DEBUG: Service %s/%s does not match MPSR %s\n", namespace, serviceName, sr.Name)
			continue
		}

		return util.GenerateServiceURL(clusterName, namespace, serviceName), nil
	}

	return "", fmt.Errorf("Not found any suitable ManagedProxyServiceResolver for (cluster:%s, namespace: %s, service: %s)", clusterName, namespace, serviceName)
}
