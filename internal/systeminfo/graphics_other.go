//go:build !darwin && !linux && !windows

package systeminfo

import "context"

func graphicsInfo(context.Context) []GPUInfo {
	return nil
}
