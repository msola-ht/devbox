package systeminfo

import (
	"context"
	"os/exec"
	"strings"
)

func graphicsInfo(ctx context.Context) []GPUInfo {
	output, err := exec.CommandContext(ctx, "lspci", "-vmm", "-d", "::03xx").Output()
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
