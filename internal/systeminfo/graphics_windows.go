package systeminfo

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"syscall"
)

func graphicsInfo(ctx context.Context) []GPUInfo {
	const query = `$ErrorActionPreference = 'Stop'; [Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false); ConvertTo-Json -Compress -InputObject @(Get-CimInstance -ClassName Win32_VideoController | Select-Object -ExpandProperty Name)`
	command := exec.CommandContext(ctx, "powershell.exe", "-NoLogo", "-NoProfile", "-NonInteractive", "-Command", query)
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := command.Output()
	if err != nil {
		return nil
	}

	var models []string
	if err := json.Unmarshal(output, &models); err != nil {
		return nil
	}
	var devices []GPUInfo
	for _, name := range models {
		if model := strings.TrimSpace(name); model != "" {
			devices = append(devices, GPUInfo{Model: model})
		}
	}
	return devices
}
