package systeminfo

import "time"

// GPUInfo identifies a graphics processor by its reported model.
type GPUInfo struct {
	Model string `json:"model"`
}

// SystemInfo describes the machine running the app.
// Numeric measurements are null when unavailable.
type SystemInfo struct {
	OS               string    `json:"os"`
	OSName           string    `json:"osName"`
	OSVersion        string    `json:"osVersion"`
	Arch             string    `json:"arch"`
	CPUModel         string    `json:"cpuModel"`
	GPUs             []GPUInfo `json:"gpus"`
	MemoryTotalBytes *uint64   `json:"memoryTotalBytes"`
	DiskTotalBytes   *uint64   `json:"diskTotalBytes"`
}

// SystemSnapshot contains system information and the time collection completed.
type SystemSnapshot struct {
	Info        SystemInfo `json:"info"`
	CollectedAt time.Time  `json:"collectedAt"`
}
