package devtools

import (
	"context"
	"runtime"
)

// DetectHomebrew checks Homebrew installations on macOS.
func (Service) DetectHomebrew(ctx context.Context) ([]ToolDetection, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if runtime.GOOS != "darwin" {
		return []ToolDetection{{
			Status:  statusUnsupported,
			Message: "此项检测适用于 macOS。",
		}}, nil
	}
	// Finder launches may use a PATH that omits both Homebrew prefixes.
	paths := [2]string{"/opt/homebrew/bin/brew", "/usr/local/bin/brew"}
	if runtime.GOARCH == "amd64" {
		paths[0], paths[1] = paths[1], paths[0]
	}

	return detectTools(ctx, "brew", "已检查应用的 PATH 和 macOS 默认安装路径。",
		func(ctx context.Context, path string) ToolDetection {
			return detectVersion(ctx, path, "Homebrew ")
		}, paths[:]...)
}
