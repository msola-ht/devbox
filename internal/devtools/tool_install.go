package devtools

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strings"
	"sync"
)

// ToolInstallStatus is the current phase or outcome of a tool installation.
type ToolInstallStatus string

// ToolInstallAction identifies the operation performed by an installation task.
type ToolInstallAction string

const (
	installAction   ToolInstallAction = "install"
	updateAction    ToolInstallAction = "update"
	uninstallAction ToolInstallAction = "uninstall"
)

const (
	installRunning          ToolInstallStatus = "running"
	installCanceling        ToolInstallStatus = "canceling"
	installSucceeded        ToolInstallStatus = "succeeded"
	installFailed           ToolInstallStatus = "failed"
	installCanceled         ToolInstallStatus = "canceled"
	installMissingHomebrew  ToolInstallStatus = "missing_homebrew"
	installMissingInstaller ToolInstallStatus = "missing_installer"
	installNotDetected      ToolInstallStatus = "not_detected"
)

// ToolInstallTask reports the latest app-owned installation and its recent output.
type ToolInstallTask struct {
	ID              int               `json:"id"`
	Action          ToolInstallAction `json:"action"`
	Status          ToolInstallStatus `json:"status"`
	Message         string            `json:"message"`
	Output          string            `json:"output"`
	DownloadPercent *int              `json:"downloadPercent"`
	Cancelable      bool              `json:"cancelable"`
}

type toolInstallation struct {
	mu               sync.Mutex
	ctx              context.Context
	environment      func(context.Context) ([]string, error)
	downloadClient   func(context.Context) (*http.Client, error)
	latest           map[installTarget]*toolInstallJob
	activeInstallers map[string]bool
	nativeJob        *toolInstallJob
	quitting         bool
}

type installTarget struct {
	tool    InstallTool
	project string
}

type toolInstallJob struct {
	task         ToolInstallTask
	cancel       context.CancelFunc
	done         chan struct{}
	output       installOutput
	cancelClosed bool
}

type installWork func(context.Context, *toolInstallJob) (ToolInstallStatus, string)

// NewService creates the tool service and returns its synchronous shutdown hook.
func NewService(environment func(context.Context) ([]string, error), downloadClient func(context.Context) (*http.Client, error)) (*Service, func()) {
	ctx, cancel := context.WithCancel(context.Background())
	service := &Service{installation: toolInstallation{
		ctx: ctx, environment: environment, downloadClient: downloadClient,
		latest:           make(map[installTarget]*toolInstallJob),
		activeInstallers: make(map[string]bool),
	}}
	return service, func() {
		cancel()
		service.installation.mu.Lock()
		jobs := slices.Collect(maps.Values(service.installation.latest))
		native := service.installation.nativeJob
		service.installation.mu.Unlock()
		for _, job := range jobs {
			// A forced end-session can bypass the normal window close guard.
			if job != native {
				<-job.done
			}
		}
	}
}

// PrepareQuit permits closing the app once native installers have exited.
func PrepareQuit(s *Service) bool {
	i := &s.installation
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.nativeJob != nil {
		return false
	}
	i.quitting = true
	return true
}

// ResumeAfterCanceledQuit reopens installation submission if the app stays open.
func ResumeAfterCanceledQuit(s *Service) {
	i := &s.installation
	i.mu.Lock()
	defer i.mu.Unlock()
	i.quitting = false
}

func (install *toolInstallation) start(ctx context.Context, target installTarget, action ToolInstallAction, work installWork) (ToolInstallTask, error) {
	install.mu.Lock()
	defer install.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return ToolInstallTask{}, err
	}
	if err := install.ctx.Err(); err != nil {
		return ToolInstallTask{}, err
	}
	if install.quitting {
		return ToolInstallTask{}, context.Canceled
	}

	id := 1
	if latest := install.latest[target]; latest != nil {
		switch latest.task.Status {
		case installRunning, installCanceling:
			return latest.snapshot(), nil
		}
		id = latest.task.ID + 1
	}
	taskCtx, cancel := context.WithCancel(install.ctx)
	task := ToolInstallTask{ID: id, Action: action, Status: installRunning, Message: fmt.Sprintf("正在检查 %s 安装位置…", target.tool)}
	job := &toolInstallJob{task: task, cancel: cancel, done: make(chan struct{})}
	install.latest[target] = job
	go install.run(taskCtx, target, job, work)
	return job.snapshot(), nil
}

func (install *toolInstallation) get(target installTarget) *ToolInstallTask {
	install.mu.Lock()
	defer install.mu.Unlock()
	job := install.latest[target]
	if job == nil {
		return nil
	}
	snapshot := job.snapshot()
	return &snapshot
}

func (install *toolInstallation) cancel(target installTarget, id int) *ToolInstallTask {
	install.mu.Lock()
	defer install.mu.Unlock()
	job := install.latest[target]
	if job == nil {
		return nil
	}
	if job.task.ID == id && job.cancelable() {
		job.task.Status = installCanceling
		job.task.Message = "正在取消操作…"
		job.task.DownloadPercent = nil
		job.cancel()
	}
	snapshot := job.snapshot()
	return &snapshot
}

func (j *toolInstallJob) snapshot() ToolInstallTask {
	task := j.task
	task.Output = j.output.text()
	task.Cancelable = j.cancelable()
	return task
}

func (j *toolInstallJob) cancelable() bool {
	return j.task.Status == installRunning && !j.cancelClosed
}

func (i *toolInstallation) run(ctx context.Context, target installTarget, job *toolInstallJob, work installWork) {
	defer close(job.done)
	defer job.cancel()
	status, message := work(ctx, job)
	i.mu.Lock()
	defer i.mu.Unlock()
	if ctx.Err() != nil && !job.cancelClosed {
		status, message = installCanceled, "已取消项目 uv 操作。"
		if target.project == "" {
			message = fmt.Sprintf("已取消 %s 安装。安装期间的本机变更可能已生效。", target.tool)
			if target.tool == toolWinGet {
				message = "已取消 WinGet 操作。"
			}
		}
	}
	job.task.Status = status
	job.task.Message = message
	if status != installSucceeded {
		job.task.DownloadPercent = nil
	}
}

func (i *toolInstallation) claimInstaller(key string) bool {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.activeInstallers[key] {
		return false
	}
	i.activeInstallers[key] = true
	return true
}

func (i *toolInstallation) releaseInstaller(key string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	delete(i.activeInstallers, key)
}

func (i *toolInstallation) closeCancellation(ctx context.Context, job *toolInstallJob, message string) bool {
	i.mu.Lock()
	defer i.mu.Unlock()
	if ctx.Err() != nil {
		return false
	}
	job.cancelClosed = true
	job.task.Message = message
	return true
}

func (i *toolInstallation) beginNativeInstall(ctx context.Context, job *toolInstallJob, message string) bool {
	i.mu.Lock()
	defer i.mu.Unlock()
	if ctx.Err() != nil || i.quitting {
		return false
	}
	job.cancelClosed = true
	job.task.Message = message
	i.nativeJob = job
	return true
}

func (i *toolInstallation) endNativeInstall() {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.nativeJob = nil
}

func (i *toolInstallation) updateMessage(job *toolInstallJob, message string) {
	i.updateProgress(job, message, nil)
}

func (i *toolInstallation) updateProgress(job *toolInstallJob, message string, percent *int) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if job.task.Status != installRunning {
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
