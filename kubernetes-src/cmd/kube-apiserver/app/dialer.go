package app

import (
	"context"
	"crypto/tls"
	"errors"
	mathrand "math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"k8s.io/klog"

	utilnet "k8s.io/apimachinery/pkg/util/net"
	kubeoptions "k8s.io/kubernetes/pkg/kubeapiserver/options"
)

func CreateOutboundDialer(s completedServerRunOptions) (*http.Transport, error) {
	proxyDialerFn := createWhitelistDialer(s.ProxyCIDRWhitelist)

	proxyTLSClientConfig := &tls.Config{InsecureSkipVerify: true}

	proxyTransport := utilnet.SetTransportDefaults(&http.Transport{
		DialContext:     proxyDialerFn,
		TLSClientConfig: proxyTLSClientConfig,
	})
	return proxyTransport, nil
}

func createWhitelistDialer(whitelist kubeoptions.IPNetSlice) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		start := time.Now()
		id := mathrand.Int63() // So you can match begins/ends in the log.
		klog.Infof("[%x: %v] Dialing...", id, addr)
		defer func() {
			klog.Infof("[%x: %v] Dialed in %v.", id, addr, time.Since(start))
		}()

		if !whitelist.Contains(strings.Split(addr, ":")[0]) {
			return nil, errors.New("Address is not allowed")
		}
		dialer := &net.Dialer{}
		return dialer.DialContext(ctx, network, addr)
	}
}
