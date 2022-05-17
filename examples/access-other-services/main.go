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

	konnectivity "sigs.k8s.io/apiserver-network-proxy/konnectivity-client/pkg/client"
	"sigs.k8s.io/apiserver-network-proxy/pkg/util"
)

var targetServiceURl string
var targetServicePort string

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

//go run examples/access-other-services/main.go --service-url=local-cluster-default-busybox-p2845 --host=localhost --port=8090 --ca-cert=./temp/ca.crt  --cert=./temp/tls.crt --key=./temp/tls.key

func main() {
	flag.StringVar(&targetServiceURl, "service-url", "", "the url of the target service")
	flag.StringVar(&proxyServerHost, "host", "", "proxy server host")
	flag.StringVar(&proxyServerPort, "port", "", "proxy server port")
	flag.StringVar(&proxyCACertPath, "ca-cert", "", "the path to ca cert")
	flag.StringVar(&proxyCertPath, "cert", "", "the path to tls cert")
	flag.StringVar(&proxyKeyPath, "key", "", "the path to tls key")
	flag.Parse()

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
