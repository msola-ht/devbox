package devtools

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
)

var javascriptVersion = regexp.MustCompile(`^v?[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$`)

// DetectNPM returns cached or newly detected npm installations.
func (s *Service) DetectNPM(ctx context.Context, refresh bool) ([]ToolDetection, error) {
	return s.npm.Read(ctx, refresh, 0, func(ctx context.Context) ([]ToolDetection, error) {
		return detectNodeTool(ctx, "npm", "在已检查的路径中没有找到 npm。",
			func(ctx context.Context, path string) ToolDetection {
				return probePackageManagerVersion(ctx, "npm", path)
			})
	})
}

// DetectPNPM returns cached or newly detected pnpm installations.
func (s *Service) DetectPNPM(ctx context.Context, refresh bool) ([]ToolDetection, error) {
	return s.pnpm.Read(ctx, refresh, 0, func(ctx context.Context) ([]ToolDetection, error) {
		paths := macOSBinPaths("pnpm")
		if dir := os.Getenv("PNPM_HOME"); filepath.IsAbs(dir) {
			paths = append(paths, filepath.Join(dir, "pnpm"))
		}
		return detectTools(ctx, "pnpm", "在已检查的路径中没有找到 pnpm。",
			func(ctx context.Context, path string) ToolDetection {
				return probePackageManagerVersion(ctx, "pnpm", path,
					"npm_config_manage_package_manager_versions=false", "pnpm_config_pm_on_fail=ignore")
			}, paths...)
	})
}

func probePackageManagerVersion(ctx context.Context, name, path string, settings ...string) ToolDetection {
	// Prefer the Node.js beside the package manager and keep Corepack offline.
	commandPath := filepath.Dir(path)
	if inherited := os.Getenv("PATH"); inherited != "" {
		commandPath += string(os.PathListSeparator) + inherited
	}
	environment := []string{"PATH=" + commandPath, "COREPACK_ENABLE_NETWORK=0", "COREPACK_ENABLE_PROJECT_SPEC=0"}
	result := probeJavaScriptVersion(ctx, path, append(environment, settings...)...)
	if result.Status == statusAvailable {
		dir := filepath.Dir(result.resolvedPath)
		if filepath.Base(result.resolvedPath) == name+".js" &&
			filepath.Base(dir) == "dist" && filepath.Base(filepath.Dir(dir)) == "corepack" {
			result.Source = sourceCorepack
		}
	}
	return result
}

func probeJavaScriptVersion(ctx context.Context, path string, environment ...string) ToolDetection {
	output, err := runJavaScriptVersion(ctx, path, environment)
	version := firstOutputLine(output)
	if !javascriptVersion.MatchString(version) {
		version = ""
	}
	return versionDetection(path, version, err)
}
