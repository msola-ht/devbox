package devtools

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

// DetectPython returns cached or newly detected Python 3 interpreters.
func (s *Service) DetectPython(ctx context.Context, refresh bool) ([]ToolDetection, error) {
	return s.python.Read(ctx, refresh, 0, detectPython)
}

func detectPython(ctx context.Context) ([]ToolDetection, error) {
	name := "python3"
	if runtime.GOOS == "windows" {
		name = "python.exe"
	}
	return detectTools(ctx, name, fmt.Sprintf("应用的 PATH 中没有找到 %s。", name), probePythonEntry)
}

func probePythonEntry(ctx context.Context, path string) ToolDetection {
	executable, err := resolvePython(ctx, path)
	if err != nil {
		return ToolDetection{
			Status:  statusFailed,
			Path:    path,
			Message: fmt.Sprintf("无法定位 Python 解释器（%v）。", err),
		}
	}
	if executable == "" {
		return ToolDetection{
			Status:  statusNotFound,
			Path:    path,
			Message: "找到了 Windows Python 启动别名，未查到可用的已安装解释器。",
		}
	}
	return probePython(ctx, executable)
}

func resolvePython(ctx context.Context, path string) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil {
			return "", err
		}
		if resolved == "/usr/bin/python3" {
			return developerToolPath(ctx, "python3")
		}
	case "windows":
		return windowsPythonPath(ctx, path)
	}
	return path, nil
}

func probePython(ctx context.Context, path string) ToolDetection {
	const script = `import json, sys; print(json.dumps({"version": sys.version.split()[0], "path": sys.executable}))`
	args := []string{"-I", "-B", "-S", "-c", script}
	var environment []string
	if runtime.GOOS == "windows" {
		environment = []string{"PYTHON_MANAGER_AUTOMATIC_INSTALL=false"}
	}
	output, err := runCommand(ctx, path, args, environment...)
	if err != nil {
		return versionFailure(path, err)
	}

	var interpreter struct {
		Version string `json:"version"`
		Path    string `json:"path"`
	}
	if err := json.Unmarshal(output, &interpreter); err != nil ||
		!strings.HasPrefix(interpreter.Version, "3.") || !filepath.IsAbs(interpreter.Path) {
		return ToolDetection{
			Status:  statusFailed,
			Path:    path,
			Message: "命令未返回可识别的 Python 3 版本和解释器路径。",
		}
	}
	return ToolDetection{
		Status:  statusAvailable,
		Path:    interpreter.Path,
		Version: interpreter.Version,
	}
}
