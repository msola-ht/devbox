package proxy

import (
	"encoding/json"
	"errors"
	"os"

	"devbox/internal/configfile"
)

const settingsVersion = 1
const settingsFile = "proxy.json"

var errInvalidSettings = errors.New("代理设置文件无效或版本不兼容。")

type storedSettings struct {
	Version int `json:"version"`
	Settings
}

func loadSettings() (Settings, error) {
	path, err := configfile.Path(settingsFile)
	if err != nil {
		return Settings{}, errors.New("无法确定代理设置位置。")
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return defaultSettings(), nil
	}
	if err != nil {
		return Settings{}, errors.New("无法读取代理设置。")
	}
	cfg := storedSettings{Settings: Settings{Manual: defaultSettings().Manual}}
	if err := json.Unmarshal(data, &cfg); err != nil || cfg.Version != settingsVersion || !validMode(cfg.Mode) {
		return Settings{}, errInvalidSettings
	}
	manual, err := normalizeManual(cfg.Manual)
	if err != nil {
		return Settings{}, errInvalidSettings
	}
	cfg.Manual = manual
	return cfg.Settings, nil
}

func saveSettings(settings Settings) error {
	path, err := configfile.Path(settingsFile)
	if err != nil {
		return errors.New("无法确定代理设置位置。")
	}
	if err := configfile.WriteJSON(path, storedSettings{Version: settingsVersion, Settings: settings}); err != nil {
		return errors.New("无法保存代理设置。")
	}
	return nil
}
