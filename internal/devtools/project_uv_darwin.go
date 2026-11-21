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

func (s *Service) installProjectUV(ctx context.Context, project string, job *uvInstallJob) (UVInstallStatus, string) {
	if err := s.prepareProjectUV(ctx, project, job); err != nil {
		return uvInstallFailed, fmt.Sprintf("项目 uv 安装未完成（%v）。", err)
	}
	return uvInstalled, "项目 uv 已可用。"
}

func (s *Service) prepareProjectUV(ctx context.Context, project string, job *uvInstallJob) error {
	root, err := projectdir.Open(project)
	if err != nil {
		return err
	}
	defer root.Close()
	if current := readProjectUV(ctx, root); current.Status != statusNotFound {
		if current.Status == statusAvailable {
			return nil
		}
		return errors.New(current.Message)
	}
	tools, err := projectdir.Tools(root, true)
	if err != nil {
		return err
	}
	defer tools.Close()

	install := &s.uvInstall
	env, err := install.environment(ctx)
	if err != nil {
		return err
	}
	temp, err := os.MkdirTemp("", "devbox-uv-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temp)
	var downloadEnv []string
	for _, entry := range env {
		name, _, _ := strings.Cut(entry, "=")
		switch strings.ToLower(name) {
		case "home", "http_proxy", "https_proxy", "all_proxy", "no_proxy", "ssl_cert_file", "ssl_cert_dir", "curl_ca_bundle":
			downloadEnv = append(downloadEnv, entry)
		}
	}
	bin := filepath.Join(temp, "bin")
	downloadEnv = append(downloadEnv, "PATH=/usr/bin:/bin:/usr/sbin:/sbin", "UV_UNMANAGED_INSTALL="+bin)
	script := filepath.Join(temp, "install.sh")
	install.updateMessage(job, "正在下载 uv 安装器…")
	args := []string{"-q", "--fail", "--silent", "--show-error", "--location",
		"--proto", "=https", "--proto-redir", "=https", "--output", script, "https://astral.sh/uv/install.sh"}
	if err := localcmd.RunInDirectory(ctx, temp, "/usr/bin/curl", args, downloadEnv, &job.output, &job.output); err != nil {
		return err
	}
	install.updateMessage(job, "正在安装并校验项目 uv…")
	if err := localcmd.RunInDirectory(ctx, temp, "/bin/sh", []string{script}, downloadEnv, &job.output, &job.output); err != nil {
		return err
	}
	var versionOutput bytes.Buffer
	probeCtx, cancelProbe := context.WithTimeout(ctx, 5*time.Second)
	err = localcmd.RunInDirectory(probeCtx, temp, filepath.Join(bin, "uv"), []string{"--version"}, downloadEnv, &versionOutput, &job.output)
	cancelProbe()
	if err != nil {
		return err
	}
	version, ok := strings.CutPrefix(firstOutputLine(versionOutput.Bytes()), "uv ")
	if !ok || version == "" {
		return errors.New("无法读取下载的 uv 版本")
	}

	stageName := ".uv-" + rand.Text()
	stage, err := projectdir.OpenChild(tools, stageName, true)
	if err != nil {
		return err
	}
	defer tools.RemoveAll(stageName)
	defer stage.Close()
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
	if err := stage.WriteFile(uvRecordName, recordBytes, 0644); err != nil {
		return err
	}
	dir, err := tools.Open(".")
	if err != nil {
		return err
	}
	defer dir.Close()

	install.mu.Lock()
	defer install.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := unix.RenameatxNp(int(dir.Fd()), stageName, int(dir.Fd()), "uv", unix.RENAME_EXCL); err != nil {
		if errors.Is(err, fs.ErrExist) {
			return errors.New("项目 uv 目录已存在，请刷新检测")
		}
		return err
	}
	job.committed = true
	return nil
}
