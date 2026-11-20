package proxy

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// ManualConfig identifies the server shared by HTTP and SOCKS5 proxies.
type ManualConfig struct {
	Address string `json:"address"`
}

// Settings contains the effective mode and the last manual proxy.
type Settings struct {
	Mode   string       `json:"mode"`
	Manual ManualConfig `json:"manual"`
}

// SystemEntry is one enabled system proxy setting.
type SystemEntry struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

// SystemProxyInfo describes system settings and the routes used for example.com.
type SystemProxyInfo struct {
	SettingsAvailable bool          `json:"settingsAvailable"`
	Settings          []SystemEntry `json:"settings"`
	HTTP              string        `json:"http"`
	HTTPS             string        `json:"https"`
}

var errInvalidMode = errors.New("请选择有效的代理模式。")

func defaultSettings() Settings {
	return Settings{
		Mode:   "system",
		Manual: ManualConfig{Address: "127.0.0.1:7890"},
	}
}

func validMode(mode string) bool {
	return mode == "system" || mode == "manual" || mode == "direct"
}

func normalizeManual(manual ManualConfig) (ManualConfig, error) {
	address := strings.ReplaceAll(strings.TrimSpace(manual.Address), "：", ":")
	if !strings.Contains(address, "://") {
		address = "http://" + address
	}
	parsed, err := url.Parse(address)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "socks5") || parsed.Hostname() == "" ||
		parsed.User != nil || (parsed.Path != "" && parsed.Path != "/") ||
		parsed.RawQuery != "" || parsed.Fragment != "" ||
		strings.ContainsAny(parsed.Hostname(), " \t\r\n") {
		return ManualConfig{}, fmt.Errorf("请输入有效的代理地址，例如 127.0.0.1:7890。")
	}
	port, err := strconv.Atoi(parsed.Port())
	if err != nil || port < 1 || port > 65535 {
		return ManualConfig{}, fmt.Errorf("代理地址需要包含 1 到 65535 之间的端口。")
	}
	manual.Address = net.JoinHostPort(parsed.Hostname(), strconv.Itoa(port))
	return manual, nil
}
