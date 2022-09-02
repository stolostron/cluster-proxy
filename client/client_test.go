package client

import (
	"context"
	"fmt"
	"testing"

	"k8s.io/client-go/tools/clientcmd"
)

func TestGetServiceURL(t *testing.T) {
	kubeconfigFilePath := "/Users/xuezhao/configs/kubeconfigs/cluster-proxy.kubeconfig"
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigFilePath)
	if err != nil {
		t.Fatal(err)
	}
	url, err := GetProxyHost(context.Background(), cfg, "local-cluster", "default", "busybox")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(url)
}
