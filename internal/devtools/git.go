package devtools

import (
	"context"
	"path/filepath"
	"runtime"
)

// DetectGit checks the Git executable found in the app's PATH.
func (Service) DetectGit(ctx context.Context) (ToolDetection, error) {
	if err := ctx.Err(); err != nil {
		return ToolDetection{}, err
	}
	path, err := findInPath("git")
	if err != nil {
		return pathFailure(path, err), nil
	}
	if path == "" {
		return ToolDetection{
			Status:  statusNotFound,
			Message: "应用的 PATH 中没有找到 Git。",
		}, nil
	}

	if runtime.GOOS == "darwin" {
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil {
			return pathFailure(path, err), nil
		}
		if resolved == "/usr/bin/git" {
			executable, err := developerToolPath(ctx, "git")
			if err := ctx.Err(); err != nil {
				return ToolDetection{}, err
			}
			if err != nil {
				return ToolDetection{
					Status:  statusFailed,
					Path:    path,
					Message: "无法从当前 Xcode 或命令行工具目录找到可用的系统 Git。",
				}, nil
			}
			path = executable
		}
	}
	return detectVersion(ctx, path, "git version ")
}
