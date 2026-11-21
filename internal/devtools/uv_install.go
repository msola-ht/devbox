package devtools

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"

	"devbox/internal/localcmd"
)

// UVInstallStatus is the current phase or outcome of a uv installation.
type UVInstallStatus string

// UVInstallAction identifies an operation on uv.
type UVInstallAction string

const (
	uvInstallAction   UVInstallAction = "install"
	uvUpdateAction    UVInstallAction = "update"
	uvUninstallAction UVInstallAction = "uninstall"
)

const (
	uvInstalling      UVInstallStatus = "running"
	uvCanceling       UVInstallStatus = "canceling"
	uvInstalled       UVInstallStatus = "succeeded"
	uvInstallFailed   UVInstallStatus = "failed"
	uvInstallCanceled UVInstallStatus = "canceled"
	uvMissingHomebrew UVInstallStatus = "missing_homebrew"
	uvNotDetected     UVInstallStatus = "not_detected"
)

// UVInstallTask reports the latest app-owned installation and its recent output.
type UVInstallTask struct {
	ID              int             `json:"id"`
	Action          UVInstallAction `json:"action"`
	Status          UVInstallStatus `json:"status"`
	Message         string          `json:"message"`
	Output          string          `json:"output"`
	DownloadPercent *int            `json:"downloadPercent"`
	Cancelable      bool            `json:"cancelable"`
}

type uvInstallation struct {
	mu             sync.Mutex
	ctx            context.Context
	environment    func(context.Context) ([]string, error)
	downloadClient func(context.Context) (*http.Client, error)
	latest         map[string]*uvInstallJob
	uvPath         string
}

type uvInstallJob struct {
	task      UVInstallTask
	cancel    context.CancelFunc
	done      chan struct{}
	output    installOutput
	committed bool
}

// NewService creates the tool service and returns its synchronous shutdown hook.
func NewService(environment func(context.Context) ([]string, error), downloadClient func(context.Context) (*http.Client, error)) (*Service, func()) {
	ctx, cancel := context.WithCancel(context.Background())
	service := &Service{uvInstall: uvInstallation{
		ctx: ctx, environment: environment, downloadClient: downloadClient,
		latest: make(map[string]*uvInstallJob),
	}}
	return service, func() {
		cancel()
		service.uvInstall.mu.Lock()
		jobs := slices.Collect(maps.Values(service.uvInstall.latest))
		service.uvInstall.mu.Unlock()
		for _, job := range jobs {
			<-job.done
		}
	}
}

// StartUVInstall starts a uv installation or returns the active task in this scope.
func (s *Service) StartUVInstall(ctx context.Context, project string) (UVInstallTask, error) {
	return s.startUVTask(ctx, project, uvInstallAction)
}

// StartProjectUVUpdate checks for and installs the latest project uv release.
func (s *Service) StartProjectUVUpdate(ctx context.Context, project string) (UVInstallTask, error) {
	return s.startUVTask(ctx, project, uvUpdateAction)
}

// StartProjectUVUninstall removes the managed uv installation from a project.
func (s *Service) StartProjectUVUninstall(ctx context.Context, project string) (UVInstallTask, error) {
	return s.startUVTask(ctx, project, uvUninstallAction)
}

func (s *Service) startUVTask(ctx context.Context, project string, action UVInstallAction) (UVInstallTask, error) {
	if runtime.GOOS != "darwin" {
		return UVInstallTask{}, errors.New("当前仅在 macOS 上支持管理 uv 安装。")
	}
	if action != uvInstallAction && !filepath.IsAbs(project) {
		return UVInstallTask{}, errors.New("请选择项目工作目录。")
	}
	install := &s.uvInstall
	install.mu.Lock()
	defer install.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return UVInstallTask{}, err
	}
	if err := install.ctx.Err(); err != nil {
		return UVInstallTask{}, err
	}

	id := 1
	if latest := install.latest[project]; latest != nil {
		switch latest.task.Status {
		case uvInstalling, uvCanceling:
			return latest.snapshot(), nil
		}
		id = latest.task.ID + 1
	}
	taskCtx, cancel := context.WithCancel(install.ctx)
	task := UVInstallTask{ID: id, Action: action, Status: uvInstalling, Message: "正在检查 uv 安装位置…"}
	job := &uvInstallJob{task: task, cancel: cancel, done: make(chan struct{})}
	install.latest[project] = job
	go s.runUVInstall(taskCtx, project, job)
	return job.snapshot(), nil
}

// GetUVInstall returns the most recent task, or null if none has been started.
func (s *Service) GetUVInstall(project string) *UVInstallTask {
	install := &s.uvInstall
	install.mu.Lock()
	defer install.mu.Unlock()
	job := install.latest[project]
	if job == nil {
		return nil
	}
	snapshot := job.snapshot()
	return &snapshot
}

// CancelUVInstall requests cancellation and returns the current task snapshot.
func (s *Service) CancelUVInstall(project string, id int) *UVInstallTask {
	install := &s.uvInstall
	install.mu.Lock()
	defer install.mu.Unlock()
	job := install.latest[project]
	if job == nil {
		return nil
	}
	if job.task.ID == id && job.cancelable() {
		job.task.Status = uvCanceling
		job.task.Message = "正在取消操作…"
		job.task.DownloadPercent = nil
		job.cancel()
	}
	snapshot := job.snapshot()
	return &snapshot
}

func (i *uvInstallation) binPath() string {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.uvPath
}

func (j *uvInstallJob) snapshot() UVInstallTask {
	task := j.task
	task.Output = j.output.text()
	task.Cancelable = j.cancelable()
	return task
}

func (j *uvInstallJob) cancelable() bool {
	return j.task.Status == uvInstalling && !j.committed
}

func (s *Service) runUVInstall(ctx context.Context, project string, job *uvInstallJob) {
	defer close(job.done)
	defer job.cancel()
	var status UVInstallStatus
	var message string
	if project == "" {
		status, message = s.installUV(ctx, job)
	} else {
		status, message = s.manageProjectUV(ctx, project, job)
	}
	s.uvInstall.mu.Lock()
	defer s.uvInstall.mu.Unlock()
	if ctx.Err() != nil && !job.committed {
		status, message = uvInstallCanceled, "已取消项目 uv 操作。"
		if project == "" {
			message = "已取消 uv 安装。安装期间的本机变更可能已生效。"
		}
	}
	job.task.Status = status
	job.task.Message = message
	if status != uvInstalled {
		job.task.DownloadPercent = nil
	}
}

func (s *Service) installUV(ctx context.Context, job *uvInstallJob) (UVInstallStatus, string) {
	results, err := s.DetectUV(ctx, true)
	if err != nil {
		return uvInstallFailed, fmt.Sprintf("安装前无法检测 uv（%v）。", err)
	}
	if slices.ContainsFunc(results, func(result ToolDetection) bool { return result.Status == statusAvailable }) {
		return uvInstalled, "已检测到可用的 uv。"
	}

	brews, err := detectHomebrew(ctx)
	if err != nil {
		return uvInstallFailed, fmt.Sprintf("无法检测 Homebrew（%v）。", err)
	}
	index := slices.IndexFunc(brews, func(result ToolDetection) bool { return result.Status == statusAvailable })
	if index < 0 {
		return uvMissingHomebrew, "未找到可用的 Homebrew。请先安装 Homebrew，再重试。"
	}
	brew := brews[index].Path
	output, err := runCommand(ctx, brew, []string{"--prefix"})
	prefix := firstOutputLine(output)
	if err != nil || !filepath.IsAbs(prefix) {
		return uvInstallFailed, "无法读取 Homebrew 安装位置。请检查 Homebrew 后重试。"
	}
	uvPath := filepath.Join(prefix, "bin", "uv")
	install := &s.uvInstall
	install.mu.Lock()
	install.uvPath = uvPath
	install.mu.Unlock()

	if result := detectVersion(ctx, uvPath, "uv "); result.Status == statusAvailable {
		_, _ = s.DetectUV(ctx, true)
		return uvInstalled, "已检测到可用的 uv。"
	}
	env, err := install.environment(ctx)
	if err != nil {
		return uvInstallFailed, fmt.Sprintf("无法准备 Homebrew 安装环境（%v）。", err)
	}
	env = append(env,
		"HOMEBREW_NO_ASK=1",
		"HOMEBREW_NO_AUTO_UPDATE=1",
		"HOMEBREW_NO_INSTALL_CLEANUP=1",
		"HOMEBREW_NO_INSTALLED_DEPENDENTS_CHECK=1",
		"NONINTERACTIVE=1",
	)
	install.updateMessage(job, "正在通过 Homebrew 安装 uv…")
	runErr := localcmd.Run(ctx, brew, []string{"install", "--formula", "uv"}, env, &job.output, &job.output)

	install.updateMessage(job, "正在重新检测 uv…")
	results, err = s.DetectUV(install.ctx, true)
	if runErr != nil {
		return uvInstallFailed, fmt.Sprintf("Homebrew 安装 uv 失败（%v）。", runErr)
	}
	if err != nil {
		return uvNotDetected, fmt.Sprintf("Homebrew 安装命令已完成，重新检测 uv 失败（%v）。", err)
	}
	resolvedPath, _ := filepath.EvalSymlinks(uvPath)
	for _, result := range results {
		if result.Status == statusAvailable &&
			(result.Path == uvPath || resolvedPath != "" && result.resolvedPath == resolvedPath) {
			return uvInstalled, "已通过 Homebrew 安装 uv。"
		}
	}
	return uvNotDetected, "Homebrew 安装命令已完成，未能读取该安装的 uv 版本。"
}

func (i *uvInstallation) updateMessage(job *uvInstallJob, message string) {
	i.updateProgress(job, message, nil)
}

func (i *uvInstallation) updateProgress(job *uvInstallJob, message string, percent *int) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if job.task.Status != uvInstalling {
		return
	}
	if percent != nil {
		if previous := job.task.DownloadPercent; previous != nil && *percent < *previous {
			return
		}
		job.task.DownloadPercent = percent
	}
	job.task.Message = message
}

// Keep the most recent output available while the command is running.
type installOutput struct {
	mu   sync.Mutex
	tail []byte
}

func (o *installOutput) Write(p []byte) (int, error) {
	const limit = 16 * 1024
	o.mu.Lock()
	defer o.mu.Unlock()
	if len(p) >= limit {
		o.tail = append(o.tail[:0], p[len(p)-limit:]...)
	} else {
		if excess := len(o.tail) + len(p) - limit; excess > 0 {
			o.tail = o.tail[:copy(o.tail, o.tail[excess:])]
		}
		o.tail = append(o.tail, p...)
	}
	return len(p), nil
}

func (o *installOutput) text() string {
	o.mu.Lock()
	defer o.mu.Unlock()
	return strings.TrimSpace(strings.ToValidUTF8(string(o.tail), "�"))
}
