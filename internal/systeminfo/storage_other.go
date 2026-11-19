//go:build !darwin && !linux && !windows

package systeminfo

import "context"

func physicalDiskCapacity(context.Context) *uint64 {
	return nil
}
