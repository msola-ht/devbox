package proxy

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	"devbox/internal/localcmd"

	"github.com/egoist/mygo/netproxy"
)

const systemRouteCommand = "--devbox-system-proxy-route"

// systemRoute uses a child environment because netproxy.Proxy skips the OS
// resolver if any proxy environment variable is set.
func systemRoute(req *http.Request) (*url.URL, error) {
	switch runtime.GOOS {
	case "darwin", "linux", "windows":
	default:
		return nil, nil // MyGo has no system resolver on these platforms.
	}

	executable, err := os.Executable()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(req.Context(), 6*time.Second)
	defer cancel()
	target := url.URL{Scheme: req.URL.Scheme, Host: req.URL.Host, Path: "/"}
	environ := os.Environ()
	env := make([]string, 0, len(environ))
	for _, entry := range environ {
		name, _, _ := strings.Cut(entry, "=")
		switch strings.ToUpper(name) {
		case "HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "NO_PROXY":
			continue
		}
		env = append(env, entry)
	}
	output, err := localcmd.Output(ctx, executable, []string{systemRouteCommand, target.String()}, env)
	if err != nil || len(output) == 0 {
		return nil, err
	}
	return url.Parse(string(output))
}

// HandleSystemRouteCommand handles the child entry before MyGo starts.
func HandleSystemRouteCommand(args []string) (bool, error) {
	if len(args) < 2 || args[1] != systemRouteCommand {
		return false, nil
	}
	if len(args) != 3 {
		return true, errors.New("system proxy target is required")
	}
	req, err := http.NewRequest(http.MethodGet, args[2], nil)
	if err != nil {
		return true, err
	}
	if req.URL.Host == "" || (req.URL.Scheme != "http" && req.URL.Scheme != "https") {
		return true, errors.New("invalid system proxy target")
	}
	address, err := netproxy.Proxy(req)
	if err != nil || address == nil {
		return true, err
	}
	_, err = fmt.Fprint(os.Stdout, address.String())
	return true, err
}
