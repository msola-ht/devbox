package devtools

import (
	"context"
	"path/filepath"
	"runtime"
)

// DetectGit returns cached or newly detected Git executables.
func (s *Service) DetectGit(ctx context.Context, refresh bool) ([]ToolDetection, error) {
	return s.git.Read(ctx, refresh, 0, func(ctx context.Context) ([]ToolDetection, error) {
		return detectTools(ctx, "git", "应用的 PATH 中没有找到 Git。", probeGit)
	})
}

func probeGit(ctx context.Context, path string) ToolDetection {
	if runtime.GOOS == "darwin" {
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil {
			return pathFailure(path, err)
		}
		if resolved == "/usr/bin/git" {
			executable, err := developerToolPath(ctx, "git")
			if err != nil {
				return ToolDetection{
					Status:  statusFailed,
					Path:    path,
					Message: "无法从当前 Xcode 或命令行工具目录找到可用的系统 Git。",
				}
			}
			path = executable
		}
	}
	return detectVersion(ctx, path, "git version ")
}
