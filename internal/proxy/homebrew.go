package proxy

import (
	"context"
	"net/url"
	"os"
	"strings"
)

// HomebrewEnvironment applies the saved proxy mode to a Homebrew subprocess.
func HomebrewEnvironment(ctx context.Context, service *Service) ([]string, error) {
	settings, err := service.Get(ctx)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return nil, errInvalidSettings
	}
	env := os.Environ()
	if settings.Mode == "system" {
		return env, nil
	}
	filtered := env[:0]
	for _, entry := range env {
		name, _, _ := strings.Cut(entry, "=")
		switch strings.ToLower(name) {
		case "http_proxy", "https_proxy", "all_proxy", "ftp_proxy", "no_proxy":
			continue
		}
		filtered = append(filtered, entry)
	}
	if settings.Mode == "direct" {
		return filtered, nil
	}
	scheme, err := proxyProtocol(ctx, settings.Manual.Address)
	if err != nil {
		return nil, err
	}
	address := (&url.URL{Scheme: scheme, Host: settings.Manual.Address}).String()
	return append(filtered, "http_proxy="+address, "https_proxy="+address, "all_proxy="+address,
		"no_proxy=localhost,.localhost,127.0.0.0/8,::1"), nil
}
