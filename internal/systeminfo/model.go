package systeminfo

import "time"

// GPUInfo describes a graphics processor. CoreCount is null when unavailable.
type GPUInfo struct {
	Model     string `json:"model"`
	CoreCount *int   `json:"coreCount"`
}

// SystemInfo describes the machine running the app.
// Numeric measurements are null when unavailable.
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

// SystemSnapshot contains system information and the time collection completed.
type SystemSnapshot struct {
	Info        SystemInfo `json:"info"`
	CollectedAt time.Time  `json:"collectedAt"`
}
