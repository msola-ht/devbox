package systeminfo

import (
	"context"
	"runtime"
	"sync"
	"time"

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

	// Leave time for the core fields if either platform command times out.
	optionalCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	var collectors sync.WaitGroup
	collectors.Go(func() { info.DiskTotalBytes = physicalDiskCapacity(optionalCtx) })
	collectors.Go(func() { info.GPUs = graphicsInfo(optionalCtx) })

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
