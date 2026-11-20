package devtools

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Service exposes local developer tool detection.
type Service struct{}

// HomebrewDetection reports availability and the executable used for detection.
type HomebrewDetection struct {
	Status  string `json:"status"`
	Path    string `json:"path"`
	Version string `json:"version"`
	Message string `json:"message"`
}

// DetectHomebrew looks up Homebrew and reads its version on macOS.
func (Service) DetectHomebrew(ctx context.Context) (HomebrewDetection, error) {
	if err := ctx.Err(); err != nil {
		return HomebrewDetection{}, err
	}
	if runtime.GOOS != "darwin" {
		return HomebrewDetection{Status: "unsupported"}, nil
	}

	path, err := findHomebrew()
	if err != nil {
		message := fmt.Sprintf("无法访问 Homebrew 入口（%v）。", err)
		if errors.Is(err, exec.ErrDot) {
			message = "应用的 PATH 通过相对路径指向 brew，请改用绝对路径。"
		}
		return HomebrewDetection{Status: "failed", Path: path, Message: message}, nil
	}
	if path == "" {
		return HomebrewDetection{Status: "not_found"}, nil
	}

	commandCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(commandCtx, path, "--version")
	cmd.WaitDelay = time.Second
	output, err := cmd.Output()
	if err := ctx.Err(); err != nil {
		return HomebrewDetection{}, err
	}
	if err != nil {
		detail := err.Error()
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if stderr := strings.TrimSpace(string(exitErr.Stderr)); stderr != "" {
				detail = stderr
			}
		}
		message := fmt.Sprintf("无法执行 brew --version（%s）。", detail)
		if errors.Is(commandCtx.Err(), context.DeadlineExceeded) {
			message = "读取 Homebrew 版本超时，请重试。"
		}
		return HomebrewDetection{Status: "failed", Path: path, Message: message}, nil
	}

	firstLine, _, _ := strings.Cut(strings.TrimSpace(string(output)), "\n")
	version, found := strings.CutPrefix(firstLine, "Homebrew ")
	version = strings.TrimSpace(version)
	if !found || version == "" {
		return HomebrewDetection{
			Status:  "failed",
			Path:    path,
			Message: "brew --version 未返回可识别的 Homebrew 版本。",
		}, nil
	}
	return HomebrewDetection{Status: "available", Path: path, Version: version}, nil
}

func findHomebrew() (string, error) {
	path, err := exec.LookPath("brew")
	if err == nil {
		if filepath.IsAbs(path) {
			return path, nil
		}
		err = exec.ErrDot
	}
	if errors.Is(err, exec.ErrNotFound) {
		path, err = "", nil
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
