package systeminfo

import (
	"context"
	"runtime"
	"sync"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

// collect reads the host without accessing the snapshot store. Disk capacity is
// the combined capacity of physical disks, counting each whole disk once.
func collect(ctx context.Context) (SystemInfo, error) {
	if err := ctx.Err(); err != nil {
		return SystemInfo{}, err
	}
	info := SystemInfo{
		OS:     runtime.GOOS,
		OSName: platformName(runtime.GOOS),
		Arch:   runtime.GOARCH,
	}

	// Run slow platform commands independently within the caller's deadline.
	var collectors sync.WaitGroup
	collectors.Go(func() { info.DiskTotalBytes = physicalDiskCapacity(ctx) })
	collectors.Go(func() { info.GPUs = graphicsInfo(ctx) })

	if _, _, version, err := host.PlatformInformationWithContext(ctx); err == nil {
		info.OSVersion = version
	}
	if processors, err := cpu.InfoWithContext(ctx); err == nil && len(processors) > 0 {
		info.CPUModel = processors[0].ModelName
	}
	if memory, err := mem.VirtualMemoryWithContext(ctx); err == nil {
		info.MemoryTotalBytes = &memory.Total
	}
	collectors.Wait()

	if err := ctx.Err(); err != nil {
		return SystemInfo{}, err
	}
	if info.GPUs == nil {
		info.GPUs = []GPUInfo{}
	}
	return info, nil
}

func platformName(goos string) string {
	switch goos {
	case "darwin":
		return "macOS"
	case "linux":
		return "Linux"
	case "windows":
		return "Windows"
	default:
		return goos
	}
}
