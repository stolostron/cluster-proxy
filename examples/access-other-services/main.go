package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"google.golang.org/grpc"
	grpccredentials "google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"k8s.io/client-go/tools/clientcmd"
	"open-cluster-management.io/cluster-proxy/client"

	konnectivity "sigs.k8s.io/apiserver-network-proxy/konnectivity-client/pkg/client"
	"sigs.k8s.io/apiserver-network-proxy/pkg/util"
)

var kubeconfig string
var managedcluster string
var namespace string
var serviceName string

var proxyServerHost string
var proxyServerPort string

// Assumes that the cluster-proxy is installed in the open-cluster-management-addon namespace.
// `proxyCACert` could be found in Secret `proxy-server-ca` in the `open-cluster-management-addon`` namespace.
var proxyCACertPath string

// Assumes that the cluster-proxy is installed in the open-cluster-management-addon namespace.
// `proxyCert` and `proxyKey` could be found in Secret `proxy-client` in the `open-cluster-management-addon`` namespace.
var proxyCertPath string
var proxyKeyPath string

var DefaultDialer = &net.Dialer{Timeout: 2 * time.Second, KeepAlive: 2 * time.Second}


func main() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.StringVar(&managedcluster, "managed-cluster", "", "the name of the managed cluster")
	flag.StringVar(&namespace, "namespace", "", "the namespace of the target service")
	flag.StringVar(&serviceName, "service-name", "", "the name of the target service")

	flag.StringVar(&proxyServerHost, "host", "", "proxy server host")
	flag.StringVar(&proxyServerPort, "port", "", "proxy server port")
	flag.StringVar(&proxyCACertPath, "ca-cert", "", "the path to ca cert")
	flag.StringVar(&proxyCertPath, "cert", "", "the path to tls cert")
	flag.StringVar(&proxyKeyPath, "key", "", "the path to tls key")
	flag.Parse()

	// Build the tunnel to ANP proxy-server
	tlsCfg, err := util.GetClientTLSConfig(proxyCACertPath, proxyCertPath, proxyKeyPath, proxyServerHost, nil)
	if err != nil {
		panic(err)
	}
	dialerTunnel, err := konnectivity.CreateSingleUseGrpcTunnel(
		context.TODO(),
		net.JoinHostPort(proxyServerHost, proxyServerPort),
		grpc.WithTransportCredentials(grpccredentials.NewTLS(tlsCfg)),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time: time.Second * 5,
		}),
	)
	if err != nil {
		panic(err)
	}

	// Get the target URL
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}
	targetServiceURl, err := client.GetServiceURL(context.Background(), cfg, managedcluster, namespace, serviceName)
	if err != nil {
		panic(err)
	}

	// Using the tunnel to access the target service
	tr := &http.Transport{
		DialContext:         dialerTunnel.DialContext,
		TLSHandshakeTimeout: 2 * time.Second,
	}
	client := http.Client{Transport: tr}

	resp, err := client.Get("http://" + targetServiceURl + ":8000")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Print("response: ", string(content))
}
