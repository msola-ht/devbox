package devtools

import (
	"context"
	"runtime"
)

// DetectZsh returns cached or newly detected Zsh executables on macOS and Linux.
func (s *Service) DetectZsh(ctx context.Context, refresh bool) ([]ToolDetection, error) {
	return s.zsh.Read(ctx, refresh, 0, detectZsh)
}

func detectZsh(ctx context.Context) ([]ToolDetection, error) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		return []ToolDetection{{
			Status: statusUnsupported, Message: "此项检测适用于 macOS 和 Linux。",
		}}, nil
	}
	paths := append(macOSBinPaths("zsh"), "/bin/zsh", "/usr/bin/zsh")
	return detectTools(ctx, "zsh", "在已检查的路径中没有找到 Zsh。",
		func(ctx context.Context, path string) ToolDetection {
			return detectVersion(ctx, path, "zsh ")
		}, paths...)
}
