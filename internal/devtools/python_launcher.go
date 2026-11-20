package devtools

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

func windowsPythonPath(ctx context.Context, path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.Mode().IsRegular() {
		return path, nil
	}

	// App execution aliases are reparse points, so query installed runtimes
	// without using the selected alias to launch Python.
	manager, err := findInPath("pymanager.exe")
	if err != nil || manager == "" {
		return "", err
	}
	output, err := runCommand(ctx, manager, []string{"list", "--one", "--format=json"})
	if err != nil {
		return "", err
	}

	var installed struct {
		Versions []struct {
			Executable string `json:"executable"`
		} `json:"versions"`
	}
	if err := json.Unmarshal(output, &installed); err != nil {
		return "", err
	}
	if len(installed.Versions) == 0 {
		return "", nil
	}
	executable := installed.Versions[0].Executable
	if !filepath.IsAbs(executable) {
		return "", errors.New("安装管理器未返回绝对解释器路径")
	}
	info, err = os.Stat(executable)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", errors.New("安装管理器未返回可直接执行的解释器")
	}
	return executable, nil
}
