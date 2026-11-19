package systeminfo

import (
	"context"
	"encoding/json"
	"os/exec"
	"strconv"
)

func physicalDiskCapacity(ctx context.Context) *uint64 {
	output, err := exec.CommandContext(ctx, "lsblk", "--bytes", "--json", "--nodeps", "--output", "TYPE,SIZE").Output()
	if err != nil {
		return nil
	}

	var devices struct {
		Disks []struct {
			Type string      `json:"type"`
			Size json.Number `json:"size"`
		} `json:"blockdevices"`
	}
	if err := json.Unmarshal(output, &devices); err != nil {
		return nil
	}
	var total uint64
	for _, disk := range devices.Disks {
		if disk.Type != "disk" {
			continue
		}
		capacity, err := strconv.ParseUint(string(disk.Size), 10, 64)
		if err != nil {
			return nil
		}
		total += capacity
	}
	if total == 0 {
		return nil
	}
	return &total
}
