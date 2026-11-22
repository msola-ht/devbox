package devtools

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func userBinDir() string {
	home, err := os.UserHomeDir()
	if err != nil || !filepath.IsAbs(home) {
		return ""
	}
	return filepath.Join(home, ".local", "bin")
}

func (s *Service) installUserUV(ctx context.Context, job *toolInstallJob) (ToolInstallStatus, string) {
	install := &s.installation
	bin := userBinDir()
	if bin == "" {
		return installFailed, "无法确定用户工具目录。"
	}
	if err := os.MkdirAll(bin, 0755); err != nil {
		return installFailed, fmt.Sprintf("无法创建用户工具目录（%v）。", err)
	}
	for _, name := range []string{"uv", "uvx"} {
		path := filepath.Join(bin, name)
		if _, err := os.Lstat(path); err == nil {
			return installFailed, fmt.Sprintf("安装位置已有 %s。请检查并移走该入口后重试。", path)
		} else if !errors.Is(err, fs.ErrNotExist) {
			return installFailed, fmt.Sprintf("无法检查安装位置 %s（%v）。", path, err)
		}
	}
	temp, err := os.MkdirTemp(bin, ".devbox-uv-")
	if err != nil {
		return installFailed, fmt.Sprintf("无法准备 uv 安装目录（%v）。", err)
	}
	defer os.RemoveAll(temp)

	client, err := install.downloadClient(ctx)
	if err != nil {
		return installFailed, fmt.Sprintf("无法准备 uv 下载（%v）。", err)
	}
	defer client.CloseIdleConnections()
	archive, err := install.downloadUVRelease(ctx, client, temp, "", job)
	if err != nil {
		return installFailed, err.Error()
	}
	stage := filepath.Join(temp, "bin")
	install.updateMessage(job, "正在解包并检查 uv…")
	if err := extractUVRelease(ctx, archive, stage); err != nil {
		return installFailed, err.Error()
	}
	if result := detectVersion(ctx, filepath.Join(stage, "uv"), "uv "); result.Status != statusAvailable {
		return installFailed, result.Message
	}

	if !install.closeCancellation(ctx, job, "正在写入并复检 uv…") {
		return installCanceled, "安装已取消。"
	}
	// Link on the same filesystem so an existing entry is never overwritten.
	stagedUVX, uvx := filepath.Join(stage, "uvx"), filepath.Join(bin, "uvx")
	if err := os.Link(stagedUVX, uvx); err != nil {
		return installFailed, fmt.Sprintf("无法安装 uvx（%v）。", err)
	}
	if err := os.Link(filepath.Join(stage, "uv"), filepath.Join(bin, "uv")); err != nil {
		if original, originalErr := os.Stat(stagedUVX); originalErr == nil {
			if current, currentErr := os.Lstat(uvx); currentErr == nil && os.SameFile(original, current) {
				if removeErr := os.Remove(uvx); removeErr != nil {
					return installFailed, fmt.Sprintf("无法安装 uv（%v）。已写入的 uvx 保留在 %s（%v）。", err, uvx, removeErr)
				}
			}
		}
		_, _ = s.DetectUV(install.ctx, true)
		return installFailed, fmt.Sprintf("无法安装 uv（%v）。", err)
	}
	results, err := s.DetectUV(install.ctx, true)
	if err == nil && availableAt(results, filepath.Join(bin, "uv")) {
		return installSucceeded, fmt.Sprintf("已安装 uv 到 %s。", bin)
	}
	return installNotDetected, "uv 已写入用户工具目录，暂未检出可用版本。"
}
