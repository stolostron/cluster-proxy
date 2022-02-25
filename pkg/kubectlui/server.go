package kubectlui

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	grpccredentials "google.golang.org/grpc/credentials"
	"k8s.io/klog/v2"
	konnectivity "sigs.k8s.io/apiserver-network-proxy/konnectivity-client/pkg/client"
	"sigs.k8s.io/apiserver-network-proxy/pkg/util"
)

type KubectlUI struct {
	getTunnel func() (konnectivity.Tunnel, error)

	serverCert string
	serverKey  string
	serverPort int
}

func NewKubectlUI(proxyServerHost string, proxyServerPort int,
	proxyCAPath, proxyCertPath, proxyKeyPath,
	serverCert, serverKey string, serverPort int) (*KubectlUI, error) {
	proxyTLSCfg, err := util.GetClientTLSConfig(proxyCAPath, proxyCertPath, proxyKeyPath, proxyServerHost, nil)
	if err != nil {
		return nil, err
	}

	return &KubectlUI{
		getTunnel: func() (konnectivity.Tunnel, error) {
			// instantiate a gprc proxy dialer
			tunnel, err := konnectivity.CreateSingleUseGrpcTunnel(
				context.TODO(),
				net.JoinHostPort(proxyServerHost, strconv.Itoa(proxyServerPort)),
				grpc.WithTransportCredentials(grpccredentials.NewTLS(proxyTLSCfg)),
			)
			if err != nil {
				return nil, err
			}
			return tunnel, nil
		},
		serverCert: serverCert,
		serverKey:  serverKey,
		serverPort: serverPort,
	}, nil
}

func (k *KubectlUI) handler(wr http.ResponseWriter, req *http.Request) {
	if klog.V(4).Enabled() {
		dump, err := httputil.DumpRequest(req, true)
		if err != nil {
			http.Error(wr, err.Error(), http.StatusBadRequest)
			return
		}
		klog.V(4).Infof("request:\n%s", string(dump))
	}

	// parse clusterID from current requestURL
	clusterID, kubeAPIPath, err := parseRequestURL(req.RequestURI)
	if err != nil {
		http.Error(wr, err.Error(), http.StatusBadRequest)
		return
	}

	target := fmt.Sprintf("https://%s", clusterID)
	apiserverURL, err := url.Parse(target)
	if err != nil {
		http.Error(wr, err.Error(), http.StatusBadRequest)
		return
	}

	tunnel, err := k.getTunnel()
	if err != nil {
		http.Error(wr, err.Error(), http.StatusBadRequest)
		return
	}

	var proxyConn net.Conn
	defer func() {
		if proxyConn != nil {
			err = proxyConn.Close()
			if err != nil {
				klog.Errorf("connection closed: %v", err)
			}
		}
	}()

	proxy := httputil.NewSingleHostReverseProxy(apiserverURL)
	proxy.Transport = &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Skip server-auth for kube-apiserver
		},
		// golang http pkg automaticly upgrade http connection to http2 connection, but http2 can not upgrade to SPDY which used in "kubectl exec".
		// set ForceAttemptHTTP2 = false to prevent auto http2 upgration
		ForceAttemptHTTP2: false,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			proxyConn, err = tunnel.DialContext(ctx, network, addr)
			return proxyConn, err
		},
	}

	proxy.ErrorHandler = func(rw http.ResponseWriter, r *http.Request, e error) {
		rw.Write([]byte(fmt.Sprintf("proxy to anp-proxy-server failed because %v", e)))
		klog.Errorf("proxy to anp-proxy-server failed because %v", e)
	}

	// update request URL path
	req.URL.Path = kubeAPIPath

	klog.V(4).Infof("request scheme:%s; rawQuery:%s; path:%s", req.URL.Scheme, req.URL.RawQuery, req.URL.Path)

	proxy.ServeHTTP(wr, req)
}

func parseRequestURL(requestURL string) (clusterID string, kubeAPIPath string, err error) {
	paths := strings.Split(requestURL, "/")
	if len(paths) <= 2 {
		err = fmt.Errorf("requestURL format not correct, path more than 2: %s", requestURL)
		return
	}
	clusterID = paths[1]                             // <clusterID>
	kubeAPIPath = strings.Join(paths[2:], "/")       // api/pods?timeout=32s
	kubeAPIPath = strings.Split(kubeAPIPath, "?")[0] // api/pods note: we only need path here, the proxy pkg would add params back
	return
}

func (k *KubectlUI) Start(ctx context.Context) error {
	var err error

	klog.Infof("start https server on %d", k.serverPort)
	http.HandleFunc("/", k.handler)

	err = http.ListenAndServeTLS(fmt.Sprintf(":%d", k.serverPort), k.serverCert, k.serverKey, nil)
	if err != nil {
		klog.Fatalf("failed to start user proxy server: %v", err)
	}

	return nil
}
