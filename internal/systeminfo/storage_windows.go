package systeminfo

import (
	"context"
	"encoding/json"
	"os/exec"
	"syscall"
)

func physicalDiskCapacity(ctx context.Context) *uint64 {
	const query = `$ErrorActionPreference = 'Stop'; [Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false); ConvertTo-Json -Compress -InputObject @(Get-CimInstance -ClassName Win32_DiskDrive | Where-Object { $_.Size -gt 0 } | Select-Object -ExpandProperty Size)`
	command := exec.CommandContext(ctx, "powershell.exe", "-NoLogo", "-NoProfile", "-NonInteractive", "-Command", query)
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := command.Output()
	if err != nil {
		return nil
	}

	var capacities []uint64
	if err := json.Unmarshal(output, &capacities); err != nil {
		return nil
	}
	var total uint64
	for _, capacity := range capacities {
		total += capacity
	}
	if total == 0 {
		return nil
	}
	return &total
}
