package devtools

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"devbox/internal/localcmd"
	"devbox/internal/projectdir"

	"golang.org/x/sys/unix"
)

func (s *Service) manageProjectUV(ctx context.Context, project string, job *uvInstallJob) (UVInstallStatus, string) {
	message, err := s.applyProjectUV(ctx, project, job)
	if err != nil {
		return uvInstallFailed, fmt.Sprintf("项目 uv 操作未完成（%v）。", err)
	}
	return uvInstalled, message
}

func (s *Service) applyProjectUV(ctx context.Context, project string, job *uvInstallJob) (string, error) {
	root, err := projectdir.Open(project)
	if err != nil {
		return "", err
	}
	defer root.Close()
	action := job.task.Action
	if action == uvInstallAction {
		if current := readProjectUV(ctx, root); current.Status != statusNotFound {
			if current.Status == statusAvailable {
				return "项目 uv 已可用。", nil
			}
			return "", errors.New(current.Message)
		}
	}
	tools, err := projectdir.Tools(root, action == uvInstallAction)
	if err != nil {
		return "", err
	}
	defer tools.Close()

	install := &s.uvInstall
	var current *os.Root
	var original projectUVSnapshot
	var currentVersion string
	if action != uvInstallAction {
		current, err = projectdir.OpenChild(tools, "uv", false)
		if errors.Is(err, fs.ErrNotExist) {
			return "", errors.New("当前项目未安装 uv")
		}
		if err != nil {
			return "", err
		}
		defer current.Close()
		original, err = inspectProjectUV(ctx, current)
		if err != nil {
			return "", err
		}
		if original.compatible() {
			currentVersion = original.record.Version
		}
	}

	stageName := ".uv-" + rand.Text()
	if action == uvUninstallAction {
		if err := install.commitProjectUV(ctx, project, tools, current, original, job, stageName); err != nil {
			return "", err
		}
		if err := removePreviousUV(context.WithoutCancel(ctx), tools, current, stageName, original); err != nil {
			return "", fmt.Errorf("uv 已移出安装位置，部分文件保留在 %s：%w", filepath.Join(tools.Name(), stageName), err)
		}
		return "已卸载项目 uv。", nil
	}

	stage, err := install.stageProjectUV(ctx, tools, stageName, currentVersion, job)
	if err != nil {
		return "", err
	}
	if stage == nil {
		return "项目 uv 已是最新版本。", nil
	}
	defer stage.Close()
	if err := install.commitProjectUV(ctx, project, tools, current, original, job, stageName); err != nil {
		_ = removeUVFiles(tools, stage, stageName)
		return "", err
	}
	if current != nil {
		if err := removePreviousUV(context.WithoutCancel(ctx), tools, current, stageName, original); err != nil {
			return "", fmt.Errorf("新版本已可用，旧版本部分文件保留在 %s：%w", filepath.Join(tools.Name(), stageName), err)
		}
		return "项目 uv 已更新。", nil
	}
	return "项目 uv 已可用。", nil
}

func (install *uvInstallation) stageProjectUV(ctx context.Context, tools *os.Root, stageName, currentVersion string, job *uvInstallJob) (*os.Root, error) {
	client, err := install.downloadClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.CloseIdleConnections()
	temp, err := os.MkdirTemp("", "devbox-uv-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(temp)
	archive, err := install.downloadUVRelease(ctx, client, temp, currentVersion, job)
	if err != nil || archive == "" {
		return nil, err
	}
	install.updateMessage(job, "正在解包并检查项目 uv…")
	bin := filepath.Join(temp, "bin")
	if err := extractUVRelease(ctx, archive, bin); err != nil {
		return nil, err
	}
	var versionOutput bytes.Buffer
	probeCtx, cancelProbe := context.WithTimeout(ctx, 5*time.Second)
	err = localcmd.RunInDirectory(probeCtx, temp, filepath.Join(bin, "uv"), []string{"--version"}, nil, &versionOutput, &job.output)
	cancelProbe()
	if err != nil {
		return nil, err
	}
	version, ok := strings.CutPrefix(firstOutputLine(versionOutput.Bytes()), "uv ")
	if !ok || version == "" {
		return nil, errors.New("无法读取下载的 uv 版本")
	}

	if err := tools.Mkdir(stageName, 0755); err != nil {
		return nil, err
	}
	stage, err := projectdir.OpenChild(tools, stageName, false)
	if err != nil {
		return nil, err
	}
	if err := writeUVFiles(ctx, stage, bin, version); err != nil {
		_ = removeUVFiles(tools, stage, stageName)
		_ = stage.Close()
		return nil, err
	}
	return stage, nil
}

func writeUVFiles(ctx context.Context, stage *os.Root, bin, version string) error {
	record := projectUVRecord{Version: version, OS: runtime.GOOS, Arch: runtime.GOARCH}
	for name, digest := range map[string]*string{"uv": &record.UV, "uvx": &record.UVX} {
		data, err := os.ReadFile(filepath.Join(bin, name))
		if err != nil {
			return err
		}
		if err := stage.WriteFile(name, data, 0755); err != nil {
			return err
		}
		*digest, err = executableDigest(ctx, stage, name)
		if err != nil {
			return err
		}
	}
	recordBytes, err := json.Marshal(record)
	if err != nil {
		return err
	}
	return stage.WriteFile(uvRecordName, recordBytes, 0644)
}

func (install *uvInstallation) commitProjectUV(ctx context.Context, project string, tools, current *os.Root, original projectUVSnapshot, job *uvInstallJob, stageName string) error {
	if current != nil {
		install.updateMessage(job, "正在复核项目 uv 文件…")
		if err := verifyUVSnapshot(ctx, current, original); err != nil {
			return err
		}
	}
	dir, err := tools.Open(".")
	if err != nil {
		return err
	}
	defer dir.Close()
	if err := verifyUVLocation(project, dir, current); err != nil {
		return err
	}

	from, to, flags := stageName, "uv", uint32(unix.RENAME_EXCL)
	switch job.task.Action {
	case uvUpdateAction:
		flags = unix.RENAME_SWAP
	case uvUninstallAction:
		from, to = to, from
	}

	install.mu.Lock()
	defer install.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := unix.RenameatxNp(int(dir.Fd()), from, int(dir.Fd()), to, flags); err != nil {
		if errors.Is(err, fs.ErrExist) {
			return errors.New("项目 uv 目录已存在，请刷新检测")
		}
		return err
	}
	job.committed = true
	job.task.Message = "正在完成项目 uv 操作…"
	return nil
}

func verifyUVLocation(project string, dir *os.File, current *os.Root) error {
	root, err := projectdir.Open(project)
	if err != nil {
		return err
	}
	defer root.Close()
	tools, err := projectdir.Tools(root, false)
	if err != nil {
		return err
	}
	defer tools.Close()
	before, err := dir.Stat()
	if err != nil {
		return err
	}
	now, err := tools.Stat(".")
	if err != nil {
		return err
	}
	if !os.SameFile(before, now) {
		return errors.New("工作目录已变化，请刷新检测")
	}
	if current == nil {
		return nil
	}
	return verifyUVDirectory(tools, current, "uv")
}

func verifyUVDirectory(tools, current *os.Root, name string) error {
	before, err := current.Stat(".")
	if err != nil {
		return err
	}
	now, err := tools.Lstat(name)
	if err != nil {
		return err
	}
	if !os.SameFile(before, now) {
		return errors.New("项目 uv 目录已变化，请刷新检测")
	}
	return nil
}

func verifyUVSnapshot(ctx context.Context, current *os.Root, original projectUVSnapshot) error {
	entries, err := fs.ReadDir(current.FS(), ".")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		switch entry.Name() {
		case "uv", "uvx", uvRecordName:
			continue
		default:
			return errors.New("项目 uv 目录中有额外文件，请移走后重试")
		}
	}
	snapshot, err := inspectProjectUV(ctx, current)
	if canceled := ctx.Err(); canceled != nil {
		return canceled
	}
	if err != nil || snapshot != original {
		return errors.New("项目 uv 文件已变化，请刷新检测")
	}
	return nil
}

func removePreviousUV(ctx context.Context, tools, current *os.Root, name string, original projectUVSnapshot) error {
	if err := verifyUVSnapshot(ctx, current, original); err != nil {
		return err
	}
	if err := verifyUVDirectory(tools, current, name); err != nil {
		return err
	}
	return removeUVFiles(tools, current, name)
}

func removeUVFiles(tools, target *os.Root, name string) error {
	for _, file := range []string{"uv", "uvx", uvRecordName} {
		if err := target.Remove(file); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}
	return tools.Remove(name)
}
