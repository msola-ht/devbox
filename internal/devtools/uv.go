package devtools

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"
)

// DetectUV returns cached or newly detected uv executables.
func (s *Service) DetectUV(ctx context.Context, refresh bool) ([]ToolDetection, error) {
	var paths []string
	if runtime.GOOS == "linux" {
		if dir := userBinDir(); dir != "" {
			paths = append(paths, filepath.Join(dir, "uv"))
		}
	}
	return s.uv.Read(ctx, refresh, 0, func(ctx context.Context) ([]ToolDetection, error) {
		return s.detectVersionedTool(ctx, toolUV, "uv", "uv ", paths)
	})
}

// StartUVInstall starts a uv installation or returns the active task in this scope.
func (s *Service) StartUVInstall(ctx context.Context, project string) (ToolInstallTask, error) {
	return s.startUVTask(ctx, project, installAction)
}

// StartProjectUVUpdate checks for and installs the latest project uv release.
func (s *Service) StartProjectUVUpdate(ctx context.Context, project string) (ToolInstallTask, error) {
	return s.startUVTask(ctx, project, updateAction)
}

// StartProjectUVUninstall removes the managed uv installation from a project.
func (s *Service) StartProjectUVUninstall(ctx context.Context, project string) (ToolInstallTask, error) {
	return s.startUVTask(ctx, project, uninstallAction)
}

func (s *Service) startUVTask(ctx context.Context, project string, action ToolInstallAction) (ToolInstallTask, error) {
	if runtime.GOOS != "darwin" {
		return ToolInstallTask{}, errors.New("当前仅在 macOS 上支持管理 uv 安装。")
	}
	if project == "" && action == installAction {
		return s.StartToolInstall(ctx, toolUV)
	}
	if !filepath.IsAbs(project) {
		return ToolInstallTask{}, errors.New("请选择项目工作目录。")
	}
	return s.installation.start(ctx, installTarget{tool: toolUV, project: project}, action, func(ctx context.Context, job *toolInstallJob) (ToolInstallStatus, string) {
		return s.manageProjectUV(ctx, project, job)
	})
}

// GetUVInstall returns the most recent task, or null if none has been started.
func (s *Service) GetUVInstall(project string) *ToolInstallTask {
	return s.installation.get(installTarget{tool: toolUV, project: project})
}

// CancelUVInstall requests cancellation and returns the current task snapshot.
func (s *Service) CancelUVInstall(project string, id int) *ToolInstallTask {
	return s.installation.cancel(installTarget{tool: toolUV, project: project}, id)
}
