package systeminfo

import (
	"context"
	"encoding/json"
)

func physicalDiskCapacity(ctx context.Context) *uint64 {
	const query = `Get-CimInstance -ClassName Win32_DiskDrive | Where-Object { $_.Size -gt 0 } | Select-Object -ExpandProperty Size`
	output, err := powershellJSON(ctx, query)
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
