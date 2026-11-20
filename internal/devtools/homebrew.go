package devtools

import (
	"context"
	"errors"
	"io/fs"
	"os/exec"
	"runtime"
)

// DetectHomebrew looks up Homebrew and reads its version on macOS.
func (Service) DetectHomebrew(ctx context.Context) (ToolDetection, error) {
	if err := ctx.Err(); err != nil {
		return ToolDetection{}, err
	}
	if runtime.GOOS != "darwin" {
		return ToolDetection{
			Status:  statusUnsupported,
			Message: "此项检测适用于 macOS。",
		}, nil
	}

	path, err := findHomebrew()
	if err != nil {
		return pathFailure(path, err), nil
	}
	if path == "" {
		return ToolDetection{
			Status:  statusNotFound,
			Message: "已检查应用的 PATH 和 macOS 默认安装路径。",
		}, nil
	}
	return detectVersion(ctx, path, "Homebrew ")
}

func findHomebrew() (string, error) {
	path, err := findInPath("brew")
	if err == nil && path != "" {
		return path, nil
	}

	// Finder launches may use a PATH that omits both Homebrew prefixes.
	paths := [2]string{"/opt/homebrew/bin/brew", "/usr/local/bin/brew"}
	if runtime.GOARCH == "amd64" {
		paths[0], paths[1] = paths[1], paths[0]
	}

	for _, candidate := range paths {
		_, candidateErr := exec.LookPath(candidate)
		if candidateErr == nil {
			return candidate, nil
		}
		if err == nil && !errors.Is(candidateErr, fs.ErrNotExist) {
			path, err = candidate, candidateErr
		}
	}
	return path, err
}
