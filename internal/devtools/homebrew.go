package devtools

import (
	"context"
	"runtime"
)

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
