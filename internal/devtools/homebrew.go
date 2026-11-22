package devtools

import (
	"context"
	"path/filepath"
	"runtime"
	"sync"
)

type homebrewLocation struct {
	bin    string
	cellar string
}

type homebrewEntries struct {
	mu    sync.Mutex
	paths map[InstallTool]homebrewLocation
}

func (entries *homebrewEntries) location(tool InstallTool) homebrewLocation {
	entries.mu.Lock()
	defer entries.mu.Unlock()
	return entries.paths[tool]
}

func (entries *homebrewEntries) remember(tool InstallTool, location homebrewLocation) {
	entries.mu.Lock()
	defer entries.mu.Unlock()
	if entries.paths == nil {
		entries.paths = make(map[InstallTool]homebrewLocation)
	}
	entries.paths[tool] = location
	if other := companionTool(tool); other != "" {
		entries.paths[other] = homebrewLocation{
			bin:    filepath.Join(filepath.Dir(location.bin), installRecipes[other].bin),
			cellar: location.cellar,
		}
	}
}

func (location homebrewLocation) paths(paths []string) []string {
	if location.bin != "" {
		return append(paths, location.bin)
	}
	return paths
}

func (location homebrewLocation) probe(probe func(context.Context, string) ToolDetection) func(context.Context, string) ToolDetection {
	if location.cellar == "" {
		return probe
	}
	cellar, err := filepath.EvalSymlinks(location.cellar)
	if err != nil {
		return probe
	}
	return func(ctx context.Context, path string) ToolDetection {
		result := probe(ctx, path)
		if result.Status == statusAvailable && result.Source == sourceUnchecked {
			resolved := result.resolvedPath
			if resolved == "" {
				resolved, _ = filepath.EvalSymlinks(result.Path)
			}
			if withinDirectory(resolved, cellar) {
				result.Source = sourceHomebrew
			}
		}
		return result
	}
}

// DetectHomebrew returns cached or newly detected Homebrew installations.
func (s *Service) DetectHomebrew(ctx context.Context, refresh bool) ([]ToolDetection, error) {
	return s.homebrew.Read(ctx, refresh, 0, detectHomebrew)
}

func detectHomebrew(ctx context.Context) ([]ToolDetection, error) {
	if runtime.GOOS != "darwin" {
		return []ToolDetection{{
			Status:  statusUnsupported,
			Message: "此项检测适用于 macOS。",
		}}, nil
	}
	return detectTools(ctx, "brew", "已检查应用的 PATH 和 macOS 默认安装路径。",
		func(ctx context.Context, path string) ToolDetection {
			return detectVersion(ctx, path, "Homebrew ")
		}, macOSBinPaths("brew")...)
}
