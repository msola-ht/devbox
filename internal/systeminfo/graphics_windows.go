package systeminfo

import (
	"context"
	"encoding/json"
	"strings"
)

func graphicsInfo(ctx context.Context) []GPUInfo {
	const query = `Get-CimInstance -ClassName Win32_VideoController | Select-Object -ExpandProperty Name`
	output, err := powershellJSON(ctx, query)
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
