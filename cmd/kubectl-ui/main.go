package main

import (
	"context"
	"flag"
	"os"

	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	"open-cluster-management.io/cluster-proxy/pkg/kubectlui"
	ctrl "sigs.k8s.io/controller-runtime"
)

var proxyServerHost string
var proxyServerPort int
var proxyCACertPath string
var proxyCertPath string
var proxyKeyPath string
var serverCert string
var serverKey string
var serverPort int

func main() {
	var err error
	logger := klogr.New()
	klog.SetOutput(os.Stdout)
	klog.InitFlags(flag.CommandLine)
	flag.StringVar(&proxyServerHost, "host", "", "proxy server host")
	flag.IntVar(&proxyServerPort, "port", 8090, "proxy server port")
	flag.StringVar(&proxyCACertPath, "proxy-ca-cert", "", "the path to ca cert")
	flag.StringVar(&proxyCertPath, "proxy-cert", "", "the path to tls cert")
	flag.StringVar(&proxyKeyPath, "proxy-key", "", "the path to tls key")
	flag.StringVar(&serverCert, "server-cert", "", "the cert for server")
	flag.StringVar(&serverKey, "server-key", "", "the key for server")
	flag.IntVar(&serverPort, "server-port", 9090, "the port for server")
	flag.Parse()

	// pipe controller-runtime logs to klog
	ctrl.SetLogger(logger)

	kui, err := kubectlui.NewKubectlUI(proxyServerHost, proxyServerPort,
		proxyCACertPath, proxyCertPath, proxyKeyPath,
		serverCert, serverKey, serverPort)
	if err != nil {
		klog.Fatal(err)
	}
	err = kui.Start(context.Background())
	if err != nil {
		klog.Fatal(err)
	}
}
