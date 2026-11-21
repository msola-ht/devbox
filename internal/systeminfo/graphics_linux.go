package systeminfo

import (
	"context"
	"strings"

	"devbox/internal/localcmd"
)

func graphicsInfo(ctx context.Context) []GPUInfo {
	output, err := localcmd.Output(ctx, "lspci", []string{"-vmm", "-d", "::03xx"}, nil)
	if err != nil {
		return nil
	}

	var devices []GPUInfo
	for _, record := range strings.Split(string(output), "\n\n") {
		var vendor, device string
		for _, line := range strings.Split(record, "\n") {
			tag, value, found := strings.Cut(line, "\t")
			if !found {
				continue
			}
			switch tag {
			case "Vendor:":
				vendor = strings.TrimSpace(value)
			case "Device:":
				device = strings.TrimSpace(value)
			}
		}
		if device != "" {
			devices = append(devices, GPUInfo{Model: strings.TrimSpace(vendor + " " + device)})
		}
	}
	return devices
}
