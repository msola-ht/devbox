package systeminfo

import (
	"encoding/json"
	"os"
	"runtime"

	"devbox/internal/configfile"
)

const snapshotVersion = 2

type storedSnapshot struct {
	Version int `json:"version"`
	SystemSnapshot
}

func loadSnapshot() (*SystemSnapshot, error) {
	path, err := configfile.Path("system-info.json")
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var saved storedSnapshot
	if err := json.Unmarshal(data, &saved); err != nil {
		return nil, err
	}
	if saved.Version != snapshotVersion || saved.CollectedAt.IsZero() || saved.Info.OS != runtime.GOOS || saved.Info.Arch != runtime.GOARCH {
		return nil, nil
	}
	if saved.Info.GPUs == nil {
		saved.Info.GPUs = []GPUInfo{}
	}
	return &saved.SystemSnapshot, nil
}

func saveSnapshot(snapshot SystemSnapshot) error {
	path, err := configfile.Path("system-info.json")
	if err != nil {
		return err
	}
	return configfile.WriteJSON(path, storedSnapshot{Version: snapshotVersion, SystemSnapshot: snapshot})
}
