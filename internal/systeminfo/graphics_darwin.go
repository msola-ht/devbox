package systeminfo

import (
	"context"
	"encoding/json"
	"os/exec"
	"strconv"
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
			Cores string `json:"sppci_cores"`
		} `json:"SPDisplaysDataType"`
	}
	if err := json.Unmarshal(output, &displays); err != nil {
		return nil
	}

	var devices []GPUInfo
	for _, device := range displays.Devices {
		if model := strings.TrimSpace(device.Model); model != "" {
			gpu := GPUInfo{Model: model}
			if cores, err := strconv.Atoi(strings.TrimSpace(device.Cores)); err == nil && cores > 0 {
				gpu.CoreCount = &cores
			}
			devices = append(devices, gpu)
		}
	}
	return devices
}
