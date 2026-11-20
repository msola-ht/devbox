package proxy

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"
)

func readSystemSettings(ctx context.Context) ([]SystemEntry, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, "/usr/sbin/scutil", "--proxy").Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, true, fmt.Errorf("读取系统代理超时。")
		}
		if ctx.Err() != nil {
			return nil, true, ctx.Err()
		}
		return nil, true, fmt.Errorf("无法读取系统代理。")
	}
	if !strings.HasPrefix(strings.TrimSpace(string(output)), "<dictionary> {") {
		return nil, true, fmt.Errorf("无法解析系统代理。")
	}
	values := make(map[string]string)
	depth := 0
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasSuffix(line, "{") {
			depth++
			continue
		}
		if line == "}" {
			depth--
			continue
		}
		if depth != 1 {
			continue
		}
		key, value, ok := strings.Cut(line, " : ")
		if ok {
			values[key] = value
		}
	}
	if depth != 0 {
		return nil, true, fmt.Errorf("无法解析系统代理。")
	}
	entries := make([]SystemEntry, 0, 5)
	for _, item := range []struct{ name, prefix string }{
		{"HTTP", "HTTP"},
		{"HTTPS", "HTTPS"},
		{"SOCKS", "SOCKS"},
	} {
		if values[item.prefix+"Enable"] != "1" {
			continue
		}
		host, port := values[item.prefix+"Proxy"], values[item.prefix+"Port"]
		address := "地址不可用"
		if host != "" && port != "" {
			address = net.JoinHostPort(strings.Trim(host, "[]"), port)
		}
		entries = append(entries, SystemEntry{Name: item.name, Address: address})
	}
	if values["ProxyAutoConfigEnable"] == "1" {
		entries = append(entries, SystemEntry{Name: "自动配置", Address: "已开启"})
	}
	if values["ProxyAutoDiscoveryEnable"] == "1" {
		entries = append(entries, SystemEntry{Name: "自动发现", Address: "已开启"})
	}
	return entries, true, nil
}
