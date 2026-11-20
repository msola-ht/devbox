package devtools

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"devbox/internal/localcmd"
)

// Service exposes local developer tool detection.
type Service struct{}

// ToolStatus describes the outcome of detecting a tool.
type ToolStatus string

const (
	statusAvailable   ToolStatus = "available"
	statusNotFound    ToolStatus = "not_found"
	statusFailed      ToolStatus = "failed"
	statusUnsupported ToolStatus = "unsupported"
)

// ToolDetection reports availability, executable path, version, and diagnostics.
type ToolDetection struct {
	Status  ToolStatus `json:"status"`
	Path    string     `json:"path"`
	Version string     `json:"version"`
	Message string     `json:"message"`
}

func findInPath(name string) (string, error) {
	path, err := exec.LookPath(name)
	if errors.Is(err, exec.ErrNotFound) {
		return "", nil
	}
	if err == nil && !filepath.IsAbs(path) {
		return path, exec.ErrDot
	}
	return path, err
}

func pathFailure(path string, err error) ToolDetection {
	message := fmt.Sprintf("无法访问命令入口（%v）。", err)
	if errors.Is(err, exec.ErrDot) {
		message = "应用的 PATH 通过相对路径指向此命令，请改用绝对路径。"
	}
	return ToolDetection{Status: statusFailed, Path: path, Message: message}
}

func detectVersion(ctx context.Context, path, prefix string) (ToolDetection, error) {
	output, err := runCommand(ctx, path, []string{"--version"})
	if err := ctx.Err(); err != nil {
		return ToolDetection{}, err
	}
	if err != nil {
		return versionFailure(path, err), nil
	}

	firstLine, _, _ := strings.Cut(strings.TrimSpace(string(output)), "\n")
	version, found := strings.CutPrefix(firstLine, prefix)
	version = strings.TrimSpace(version)
	if !found || version == "" {
		return ToolDetection{
			Status:  statusFailed,
			Path:    path,
			Message: "命令未返回可识别的版本。",
		}, nil
	}
	return ToolDetection{Status: statusAvailable, Path: path, Version: version}, nil
}

func versionFailure(path string, err error) ToolDetection {
	message := "读取版本超时，请重试。"
	if !errors.Is(err, context.DeadlineExceeded) {
		detail := err.Error()
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if stderr := strings.TrimSpace(string(exitErr.Stderr)); stderr != "" {
				detail = stderr
			}
		}
		message = fmt.Sprintf("无法读取版本（%s）。", detail)
	}
	return ToolDetection{Status: statusFailed, Path: path, Message: message}
}

func runCommand(ctx context.Context, path string, args []string, environment ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var env []string
	if len(environment) > 0 {
		env = append(os.Environ(), environment...)
	}
	return localcmd.Output(ctx, path, args, env)
}
