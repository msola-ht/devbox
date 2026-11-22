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
	"devbox/internal/readcache"
)

// Service exposes local developer tool detection and installation.
type Service struct {
	chocolatey    readcache.Cache[[]ToolDetection]
	gh            readcache.Cache[[]ToolDetection]
	git           readcache.Cache[[]ToolDetection]
	homebrew      readcache.Cache[[]ToolDetection]
	homebrewPaths homebrewEntries
	jq            readcache.Cache[[]ToolDetection]
	node          readcache.Cache[[]ToolDetection]
	npm           readcache.Cache[[]ToolDetection]
	pnpm          readcache.Cache[[]ToolDetection]
	pwsh          readcache.Cache[[]ToolDetection]
	python        readcache.Cache[[]ToolDetection]
	rg            readcache.Cache[[]ToolDetection]
	uv            readcache.Cache[[]ToolDetection]
	winget        readcache.Cache[[]ToolDetection]
	installation  toolInstallation
	zsh           readcache.Cache[[]ToolDetection]
}

func (s *Service) detectVersionedTool(ctx context.Context, tool InstallTool, bin, prefix string, paths []string, environment ...string) ([]ToolDetection, error) {
	location := s.homebrewPaths.location(tool)
	paths = location.paths(append(macOSBinPaths(bin), paths...))
	return detectTools(ctx, bin, fmt.Sprintf("在已检查的路径中没有找到 %s。", bin),
		location.probe(func(ctx context.Context, path string) ToolDetection {
			return detectVersion(ctx, path, prefix, environment...)
		}), paths...)
}

// ToolStatus describes the outcome of detecting a tool.
type ToolStatus string

const (
	statusAvailable          ToolStatus = "available"
	statusNotFound           ToolStatus = "not_found"
	statusFailed             ToolStatus = "failed"
	statusUnsupported        ToolStatus = "unsupported"
	statusIncompatible       ToolStatus = "incompatible"
	statusProjectUnavailable ToolStatus = "project_unavailable"
)

// ToolDetection reports availability and details of a local tool installation.
type ToolDetection struct {
	Status  ToolStatus `json:"status"`
	Path    string     `json:"path"`
	Version string     `json:"version"`
	Source  ToolSource `json:"source"`
	Message string     `json:"message"`

	resolvedPath string
}

func pathFailure(path string, err error) ToolDetection {
	message := fmt.Sprintf("无法访问命令入口（%v）。", err)
	if errors.Is(err, exec.ErrDot) {
		message = "应用的 PATH 通过相对路径指向此命令，请改用绝对路径。"
	}
	return ToolDetection{Status: statusFailed, Path: path, Message: message}
}

func detectVersion(ctx context.Context, path, prefix string, environment ...string) ToolDetection {
	output, err := runCommand(ctx, path, []string{"--version"}, environment...)
	version, found := strings.CutPrefix(firstOutputLine(output), prefix)
	if !found {
		version = ""
	}
	return versionDetection(path, strings.TrimSpace(version), err)
}

func firstOutputLine(output []byte) string {
	line, _, _ := strings.Cut(strings.TrimSpace(string(output)), "\n")
	return strings.TrimSpace(line)
}

func versionDetection(path, version string, err error) ToolDetection {
	if err != nil {
		return versionFailure(path, err)
	}
	if version == "" {
		return ToolDetection{
			Status:  statusFailed,
			Path:    path,
			Message: "命令未返回可识别的版本。",
		}
	}
	resolvedPath, _ := filepath.EvalSymlinks(path)
	return ToolDetection{
		Status: statusAvailable, Path: path, Version: version,
		Source: sourceForExecutable(resolvedPath), resolvedPath: resolvedPath,
	}
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
