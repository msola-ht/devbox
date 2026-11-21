package systeminfo

import (
	"context"
	"encoding/json"
	"strings"

	"devbox/internal/localcmd"
)

func graphicsInfo(ctx context.Context) []GPUInfo {
	output, err := localcmd.Output(ctx, "/usr/sbin/system_profiler", []string{"SPDisplaysDataType", "-json"}, nil)
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
