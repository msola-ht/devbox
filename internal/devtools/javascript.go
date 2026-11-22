package devtools

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

var javascriptVersion = regexp.MustCompile(`^v?[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$`)

// DetectNPM returns cached or newly detected npm installations.
func (s *Service) DetectNPM(ctx context.Context, refresh bool) ([]ToolDetection, error) {
	return s.npm.Read(ctx, refresh, 0, func(ctx context.Context) ([]ToolDetection, error) {
		location := s.homebrewPaths.location(toolNPM)
		return detectNodeTool(ctx, "npm", "在已检查的路径中没有找到 npm。",
			location.probe(func(ctx context.Context, path string) ToolDetection {
				return probePackageManagerVersion(ctx, "npm", path)
			}), location.paths(nil)...)
	})
}

// DetectPNPM returns cached or newly detected pnpm installations.
func (s *Service) DetectPNPM(ctx context.Context, refresh bool) ([]ToolDetection, error) {
	return s.pnpm.Read(ctx, refresh, 0, func(ctx context.Context) ([]ToolDetection, error) {
		location := s.homebrewPaths.location(toolPNPM)
		paths := location.paths(macOSBinPaths("pnpm"))
		if runtime.GOOS == "linux" {
			if dir := userBinDir(); dir != "" {
				paths = append(paths, filepath.Join(dir, "pnpm"))
			}
		}
		if dir := os.Getenv("PNPM_HOME"); filepath.IsAbs(dir) {
			paths = append(paths, filepath.Join(dir, "pnpm"))
		}
		return detectTools(ctx, "pnpm", "在已检查的路径中没有找到 pnpm。",
			location.probe(func(ctx context.Context, path string) ToolDetection {
				return probePackageManagerVersion(ctx, "pnpm", path,
					"npm_config_manage_package_manager_versions=false", "pnpm_config_pm_on_fail=ignore")
			}), paths...)
	})
}

func probePackageManagerVersion(ctx context.Context, name, path string, settings ...string) ToolDetection {
	// Prefer the Node.js beside the package manager and keep Corepack offline.
	commandPath := filepath.Dir(path)
	if inherited := os.Getenv("PATH"); inherited != "" {
		commandPath += string(os.PathListSeparator) + inherited
	}
	environment := []string{
		"PATH=" + commandPath,
		"COREPACK_ENABLE_NETWORK=0",
		"COREPACK_ENABLE_PROJECT_SPEC=0",
		"COREPACK_DEFAULT_TO_LATEST=0",
		"COREPACK_ENABLE_DOWNLOAD_PROMPT=0",
	}
	output, err := runJavaScriptVersion(ctx, path, append(environment, settings...))
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && strings.Contains(string(exitErr.Stderr), "Network access disabled by the environment; can't reach ") {
		return ToolDetection{
			Status:  statusFailed,
			Path:    path,
			Source:  sourceCorepack,
			Message: fmt.Sprintf("Corepack 的本地缓存中缺少 %s 的默认版本。请先用 Corepack 安装所需版本，再刷新检测。", name),
		}
	}
	result := javascriptVersionDetection(path, output, err)
	if result.Status == statusAvailable {
		dir := filepath.Dir(result.resolvedPath)
		if filepath.Base(result.resolvedPath) == name+".js" &&
			filepath.Base(dir) == "dist" && filepath.Base(filepath.Dir(dir)) == "corepack" {
			result.Source = sourceCorepack
		}
	}
	return result
}

func javascriptVersionDetection(path string, output []byte, err error) ToolDetection {
	version := firstOutputLine(output)
	if !javascriptVersion.MatchString(version) {
		version = ""
	}
	return versionDetection(path, version, err)
}
