package systeminfo

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
)

func graphicsInfo(ctx context.Context) []GPUInfo {
	output, err := exec.CommandContext(ctx, "/usr/sbin/system_profiler", "SPDisplaysDataType", "-json").Output()
	if err != nil {
		return nil
	}

	var displays struct {
		Devices []struct {
			Model string `json:"sppci_model"`
		} `json:"SPDisplaysDataType"`
	}
	if err := json.Unmarshal(output, &displays); err != nil {
		return nil
	}

	var devices []GPUInfo
	for _, device := range displays.Devices {
		if model := strings.TrimSpace(device.Model); model != "" {
			devices = append(devices, GPUInfo{Model: model})
		}
	}
	return devices
}
