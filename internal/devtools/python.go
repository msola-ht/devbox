package devtools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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
	const script = `import json, sys; print(json.dumps({"version": sys.version.split()[0], "path": sys.executable, "basePath": getattr(sys, "_base_executable", None)}))`
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
		Version  string `json:"version"`
		Path     string `json:"path"`
		BasePath string `json:"basePath"`
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
		Source:  pythonSource(interpreter.Path, interpreter.BasePath),
	}
}

func pythonSource(path, basePath string) ToolSource {
	// Attribute the base runtime while keeping virtual environments distinct.
	if filepath.IsAbs(basePath) {
		path = basePath
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return sourceUnchecked
	}
	if source := sourceForExecutable(resolved); source != sourceUnchecked {
		return source
	}

	paths := []string{os.Getenv("UV_PYTHON_INSTALL_DIR")}
	if runtime.GOOS == "windows" {
		paths = append(paths, filepath.Join(os.Getenv("APPDATA"), "uv", "data", "python"))
	} else {
		home, _ := os.UserHomeDir()
		paths = append(paths, filepath.Join(os.Getenv("XDG_DATA_HOME"), "uv", "python"),
			filepath.Join(home, ".local", "share", "uv", "python"))
	}
	for _, dir := range paths {
		if filepath.IsAbs(dir) {
			root, err := filepath.EvalSymlinks(dir)
			if err == nil && withinDirectory(resolved, root) {
				return sourceUV
			}
		}
	}
	return sourceUnchecked
}
