package systeminfo

import (
	"context"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

// Service provides information about the machine running the app.
type Service struct{}

// GPUInfo describes a graphics processor. CoreCount is null when unavailable.
type GPUInfo struct {
	Model     string `json:"model"`
	CoreCount *int   `json:"coreCount"`
}

// SystemInfo describes the machine running the app.
type SystemInfo struct {
	OS                   string    `json:"os"`
	OSName               string    `json:"osName"`
	OSVersion            string    `json:"osVersion"`
	Hostname             string    `json:"hostname"`
	Arch                 string    `json:"arch"`
	GoVersion            string    `json:"goVersion"`
	CPUModel             string    `json:"cpuModel"`
	GPUs                 []GPUInfo `json:"gpus"`
	PhysicalCores        *int      `json:"physicalCores"`
	LogicalCores         *int      `json:"logicalCores"`
	MemoryTotalBytes     *uint64   `json:"memoryTotalBytes"`
	MemoryAvailableBytes *uint64   `json:"memoryAvailableBytes"`
	DiskTotalBytes       *uint64   `json:"diskTotalBytes"`
	UptimeSeconds        *uint64   `json:"uptimeSeconds"`
}

// Info returns a snapshot of the host's system, hardware and runtime information.
// Numeric measurements are null when unavailable. Disk capacity is the combined
// capacity of physical disks, counting each whole disk once.
func (Service) Info(ctx context.Context) SystemInfo {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	info := SystemInfo{
		OS:        runtime.GOOS,
		OSName:    platformName(runtime.GOOS),
		Arch:      runtime.GOARCH,
		GoVersion: runtime.Version(),
	}
	if ctx.Err() != nil {
		return info
	}

	// Slow platform commands share the deadline but run independently, so one
	// unavailable device category does not consume the other's time budget.
	var collectors sync.WaitGroup
	collectors.Go(func() { info.DiskTotalBytes = physicalDiskCapacity(ctx) })
	collectors.Go(func() { info.GPUs = graphicsInfo(ctx) })

	if hostname, err := os.Hostname(); err == nil {
		info.Hostname = hostname
	}
	if _, _, version, err := host.PlatformInformationWithContext(ctx); err == nil {
		info.OSVersion = version
	}
	if processors, err := cpu.InfoWithContext(ctx); err == nil && len(processors) > 0 {
		info.CPUModel = processors[0].ModelName
	}
	if count, err := cpu.CountsWithContext(ctx, false); err == nil && count > 0 {
		info.PhysicalCores = &count
	}
	if count, err := cpu.CountsWithContext(ctx, true); err == nil && count > 0 {
		info.LogicalCores = &count
	}
	if memory, err := mem.VirtualMemoryWithContext(ctx); err == nil {
		info.MemoryTotalBytes = &memory.Total
		info.MemoryAvailableBytes = &memory.Available
	}
	if uptime, err := host.UptimeWithContext(ctx); err == nil {
		info.UptimeSeconds = &uptime
	}
	collectors.Wait()

	return info
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
