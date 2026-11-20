package systeminfo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

const snapshotVersion = 2

type storedSnapshot struct {
	Version int `json:"version"`
	SystemSnapshot
}

func snapshotPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "devbox", "system-info.json"), nil
}

func loadSnapshot() (*SystemSnapshot, error) {
	path, err := snapshotPath()
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
	path, err := snapshotPath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".system-info-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	encodeErr := json.NewEncoder(file).Encode(storedSnapshot{Version: snapshotVersion, SystemSnapshot: snapshot})
	closeErr := file.Close()
	if encodeErr != nil {
		return encodeErr
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(file.Name(), path)
}
