package proxy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/egoist/mygo/netproxy"
	"golang.org/x/net/http/httpproxy"
)

const checkHost = "example.com"
const checkURL = "https://" + checkHost + "/"

var errProxyAuth = errors.New("代理服务器要求认证，当前设置不支持代理认证。")

var environmentProxy = sync.OnceValue(func() *httpproxy.Config {
	config := httpproxy.FromEnvironment()
	allProxy := os.Getenv("ALL_PROXY")
	if allProxy == "" {
		allProxy = os.Getenv("all_proxy")
	}
	if config.HTTPProxy == "" {
		config.HTTPProxy = allProxy
	}
	if config.HTTPSProxy == "" {
		config.HTTPSProxy = allProxy
	}
	return config
})

func systemProxy(req *http.Request) (*url.URL, error) {
	if isLoopback(req.URL.Hostname()) {
		return nil, nil
	}
	config := environmentProxy()
	if (req.URL.Scheme == "http" && config.HTTPProxy != "") ||
		(req.URL.Scheme == "https" && config.HTTPSProxy != "") {
		return config.ProxyFunc()(req.URL)
	}

	if config.HTTPProxy == "" && config.HTTPSProxy == "" && config.NoProxy == "" {
		return netproxy.Proxy(req)
	}

	if config.NoProxy != "" {
		// ProxyFunc exposes NO_PROXY matching through a configured proxy.
		// This address is only used to check whether the target is bypassed.
		bypass := httpproxy.Config{
			HTTPProxy:  "http://proxy.invalid",
			HTTPSProxy: "http://proxy.invalid",
			NoProxy:    config.NoProxy,
		}
		if route, _ := bypass.ProxyFunc()(req.URL); route == nil {
			return nil, nil
		}
	}
	return systemRoute(req)
}

func systemInfo(ctx context.Context) (SystemProxyInfo, error) {
	entries, available, err := readSystemSettings(ctx)
	if err != nil {
		return SystemProxyInfo{}, err
	}
	info := SystemProxyInfo{SettingsAvailable: available, Settings: entries}
	for _, target := range []struct {
		scheme string
		result *SystemRoute
	}{
		{"http", &info.HTTP},
		{"https", &info.HTTPS},
	} {
		if err := ctx.Err(); err != nil {
			return SystemProxyInfo{}, err
		}
		req := &http.Request{URL: &url.URL{Scheme: target.scheme, Host: checkHost, Path: "/"}}
		address, err := systemProxy(req.WithContext(ctx))
		if canceled := ctx.Err(); canceled != nil {
			return SystemProxyInfo{}, canceled
		}
		switch {
		case err != nil:
			target.result.Status = routeError
		case address == nil:
			target.result.Status = routeDirect
		default:
			*target.result = SystemRoute{
				Status:   routeProxy,
				Protocol: address.Scheme,
				Address:  address.Host,
			}
		}
	}
	return info, nil
}

type manualResolver struct {
	address  string
	mu       sync.Mutex
	protocol string
}

func isLoopback(host string) bool {
	host = strings.ToLower(host)
	ip := net.ParseIP(host)
	return host == "localhost" || strings.HasSuffix(host, ".localhost") || ip.IsLoopback()
}

func (r *manualResolver) proxy(req *http.Request) (*url.URL, error) {
	if isLoopback(req.URL.Hostname()) {
		return nil, nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.protocol == "" {
		protocol, err := proxyProtocol(req.Context(), r.address)
		if err != nil {
			return nil, err
		}
		r.protocol = protocol
	}
	return &url.URL{Scheme: r.protocol, Host: r.address}, nil
}

// proxyProtocol checks whether the address accepts an unauthenticated SOCKS5
// greeting. Other responses are treated as an HTTP proxy.
func proxyProtocol(ctx context.Context, address string) (string, error) {
	probeCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	conn, err := (&net.Dialer{}).DialContext(probeCtx, "tcp", address)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	stop := context.AfterFunc(probeCtx, func() { _ = conn.Close() })
	defer stop()
	_, err = conn.Write([]byte{5, 1, 0})
	if err == nil {
		var response [2]byte
		_, err = io.ReadFull(conn, response[:])
		if err == nil && response == [2]byte{5, 0} {
			return "socks5", nil
		}
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return "http", nil
}

func transportFor(settings Settings) (*http.Transport, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	switch settings.Mode {
	case "direct":
		transport.Proxy = nil
	case "system":
		transport.Proxy = systemProxy
	case "manual":
		manual, err := normalizeManual(settings.Manual)
		if err != nil {
			return nil, err
		}
		transport.Proxy = (&manualResolver{address: manual.Address}).proxy
	default:
		return nil, errInvalidMode
	}
	return transport, nil
}

// DownloadClient creates an HTTP client with the saved proxy settings.
func DownloadClient(ctx context.Context, service *Service) (*http.Client, error) {
	settings, err := service.Get(ctx)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return nil, errInvalidSettings
	}
	transport, err := transportFor(*settings)
	if err != nil {
		return nil, err
	}
	return &http.Client{Transport: transport}, nil
}

func checkConnection(ctx context.Context, settings Settings) error {
	transport, err := transportFor(settings)
	if err != nil {
		return err
	}
	transport.OnProxyConnectResponse = func(_ context.Context, _ *url.URL, _ *http.Request, response *http.Response) error {
		if response.StatusCode == http.StatusProxyAuthRequired {
			return errProxyAuth
		}
		return nil
	}
	client := &http.Client{Transport: transport}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	defer client.CloseIdleConnections()
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, checkURL, nil)
	if err != nil {
		return err
	}
	response, err := client.Do(req)
	if err != nil {
		switch ctx.Err() {
		case context.DeadlineExceeded:
			return errors.New("访问 example.com 超时，请检查当前网络和代理。")
		case context.Canceled:
			return ctx.Err()
		}
		if errors.Is(err, errProxyAuth) {
			return errProxyAuth
		}
		return errors.New("无法访问 example.com，请检查当前网络和代理。")
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusProxyAuthRequired {
		return errProxyAuth
	}
	if response.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("example.com 返回 HTTP %d，连接检查未通过。", response.StatusCode)
	}
	return nil
}
