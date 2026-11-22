package devtools

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"devbox/internal/localcmd"
)

type packageCommand struct {
	path string
	args []string
	env  []string
	dir  string
}

func (s *Service) installLocalTool(ctx context.Context, tool InstallTool, recipe installRecipe, channel InstallChannel, job *toolInstallJob) (ToolInstallStatus, string) {
	installation := &s.installation
	if channel != channelHomebrew && channel != channelAstral {
		key := string(channel)
		if !installation.claimInstaller(key) {
			return installFailed, "同一安装渠道已有任务进行中，请稍后重试。"
		}
		defer installation.releaseInstaller(key)
	}
	results, err := recipe.detect(s, ctx, true)
	if err != nil {
		return installFailed, fmt.Sprintf("安装前无法检测 %s（%v）。", recipe.name, err)
	}
	if slices.ContainsFunc(results, func(result ToolDetection) bool { return availableForInstall(tool, result) }) {
		s.refreshCompanionTool(ctx, tool)
		return installSucceeded, fmt.Sprintf("已检测到可用的 %s。", recipe.name)
	}

	switch channel {
	case channelHomebrew:
		return s.installHomebrewTool(ctx, tool, recipe, job)
	case channelAstral:
		return s.installUserUV(ctx, job)
	}

	var cmd packageCommand
	switch channel {
	case channelWinGet:
		cmd, err = s.winGetCommand(ctx, recipe.winget, tool == toolNode || tool == toolNPM)
	case channelAPT:
		cmd, err = aptCommand(recipe.apt)
	case channelNPM:
		cmd, err = s.npmCommand(ctx)
	default:
		return installFailed, "当前平台没有适用的安装渠道。"
	}
	if errors.Is(err, errMissingInstaller) {
		return installMissingInstaller, err.Error()
	}
	if err != nil {
		return installFailed, err.Error()
	}

	message := fmt.Sprintf("正在安装 %s…", recipe.name)
	var runErr error
	if channel == channelAPT || channel == channelWinGet {
		runErr = installation.runNativePackageCommand(ctx, job, cmd, message+" 如系统要求，请完成授权。")
	} else {
		installation.updateMessage(job, message)
		runErr = cmd.run(ctx, job)
	}
	if errors.Is(runErr, context.Canceled) && ctx.Err() == nil {
		return installCanceled, "安装已取消。"
	}
	installation.closeCancellation(ctx, job, fmt.Sprintf("正在重新检测 %s…", recipe.name))
	results, companion, err := s.refreshInstalledTool(tool, recipe)
	var exitErr *exec.ExitError
	if errors.As(runErr, &exitErr) {
		switch {
		case channel == channelAPT && cmd.path == "/usr/bin/pkexec" && exitErr.ExitCode() == 126:
			return installCanceled, "已取消系统授权。"
		case channel == channelAPT && cmd.path == "/usr/bin/pkexec" && exitErr.ExitCode() == 127:
			return installFailed, "无法完成系统授权。请查看任务输出，或在终端使用 APT 安装。"
		case channel == channelWinGet && uint32(exitErr.ExitCode()) == 0x8a150046:
			return installFailed, "WinGet 源条款尚未接受。请在终端运行 winget search --source winget --id " + recipe.winget + "，阅读提示后重试。"
		case channel == channelWinGet && uint32(exitErr.ExitCode()) == 0x8a150041:
			return installFailed, "此软件包需要接受许可条款。请在终端运行 winget install --exact --source winget --id " + recipe.winget + "，查看并处理提示。"
		}
	}
	if runErr != nil {
		return installFailed, fmt.Sprintf("安装 %s 失败（%v）。请查看任务输出。", recipe.name, runErr)
	}
	if err == nil && slices.ContainsFunc(results, func(result ToolDetection) bool { return availableForInstall(tool, result) }) {
		if channel == channelNPM {
			bin := userBinDir()
			if availableAt(results, filepath.Join(bin, "pnpm")) {
				return installSucceeded, fmt.Sprintf("已安装 pnpm 到 %s。", bin)
			}
			return installNotDetected, fmt.Sprintf("npm 安装已完成，暂未在 %s 检出 pnpm。", bin)
		}
		if (tool == toolNode || tool == toolNPM && channel == channelWinGet) && !slices.ContainsFunc(companion, availableTool) {
			return installNotDetected, fmt.Sprintf("已检测到 %s，暂未检出同一安装包提供的 %s。", recipe.name, installRecipes[companionTool(tool)].name)
		}
		return installSucceeded, fmt.Sprintf("已安装 %s。", recipe.name)
	}
	return installNotDetected, fmt.Sprintf("安装命令已完成，暂未检出可用的 %s。", recipe.name)
}

func (s *Service) refreshInstalledTool(tool InstallTool, recipe installRecipe) ([]ToolDetection, []ToolDetection, error) {
	ctx := s.installation.ctx
	var companion []ToolDetection
	var checks sync.WaitGroup
	if companionTool(tool) != "" {
		checks.Go(func() { companion = s.refreshCompanionTool(ctx, tool) })
	}
	results, err := recipe.detect(s, ctx, true)
	checks.Wait()
	return results, companion, err
}

func (s *Service) refreshCompanionTool(ctx context.Context, tool InstallTool) []ToolDetection {
	if other := companionTool(tool); other != "" {
		results, _ := installRecipes[other].detect(s, ctx, true)
		return results
	}
	return nil
}

func availableTool(result ToolDetection) bool { return result.Status == statusAvailable }

func availableAt(results []ToolDetection, expected string) bool {
	resolved, _ := filepath.EvalSymlinks(expected)
	return slices.ContainsFunc(results, func(result ToolDetection) bool {
		if !availableTool(result) {
			return false
		}
		if result.Path == expected {
			return true
		}
		actual := result.resolvedPath
		if actual == "" {
			actual, _ = filepath.EvalSymlinks(result.Path)
		}
		return resolved != "" && actual == resolved
	})
}

var errMissingInstaller = errors.New("缺少安装程序")

func (cmd packageCommand) run(ctx context.Context, job *toolInstallJob) error {
	return localcmd.RunInDirectory(ctx, cmd.dir, cmd.path, cmd.args, cmd.env, &job.output, &job.output)
}

func (i *toolInstallation) runNativePackageCommand(ctx context.Context, job *toolInstallJob, cmd packageCommand, message string) error {
	if !i.beginNativeInstall(ctx, job, message) {
		return context.Canceled
	}
	defer i.endNativeInstall()
	return cmd.run(context.WithoutCancel(ctx), job)
}

func (s *Service) winGetCommand(ctx context.Context, packageID string, nodeBundle bool) (packageCommand, error) {
	results, err := s.DetectWinGet(ctx, true)
	if err != nil {
		return packageCommand{}, fmt.Errorf("无法检测 WinGet：%w", err)
	}
	index := slices.IndexFunc(results, availableTool)
	if index < 0 {
		return packageCommand{}, fmt.Errorf("WinGet 当前不可用，请在「系统包管理器」中安装或修复：%w", errMissingInstaller)
	}
	args := []string{
		"install", "--id", packageID, "--exact", "--source", "winget", "--no-upgrade", "--disable-interactivity",
	}
	if nodeBundle {
		args = append(args, "--installer-type", "wix")
	}
	return packageCommand{path: results[index].Path, args: args}, nil
}

func aptCommand(packages []string) (packageCommand, error) {
	const apt = "/usr/bin/apt-get"
	if _, err := exec.LookPath(apt); err != nil {
		return packageCommand{}, fmt.Errorf("未找到 APT：%w", errMissingInstaller)
	}
	args := append([]string{"--assume-yes", "--no-upgrade", "--no-remove", "install"}, packages...)
	if os.Geteuid() == 0 {
		return packageCommand{path: apt, args: args}, nil
	}
	const pkexec = "/usr/bin/pkexec"
	if _, err := exec.LookPath(pkexec); err != nil {
		return packageCommand{}, fmt.Errorf("未找到系统授权程序 pkexec，请通过终端使用 APT：%w", errMissingInstaller)
	}
	return packageCommand{path: pkexec, args: append([]string{"--disable-internal-agent", apt}, args...)}, nil
}

func (s *Service) npmCommand(ctx context.Context) (packageCommand, error) {
	results, err := s.DetectNPM(ctx, true)
	if err != nil {
		return packageCommand{}, err
	}
	if !slices.ContainsFunc(results, availableTool) {
		return packageCommand{}, fmt.Errorf("请先在开发工具页安装 Node.js 和 npm：%w", errMissingInstaller)
	}
	probeCtx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()
	var path string
	for _, result := range results {
		if !availableTool(result) || probeCtx.Err() != nil {
			continue
		}
		node, err := exec.LookPath(filepath.Join(filepath.Dir(result.Path), "node"))
		if err != nil {
			node, err = exec.LookPath("node")
		}
		if err != nil || !filepath.IsAbs(node) {
			continue
		}
		output, err := runJavaScriptVersion(probeCtx, node, nil)
		major, _, _ := strings.Cut(strings.TrimPrefix(firstOutputLine(output), "v"), ".")
		version, _ := strconv.Atoi(major)
		if err == nil && version >= 18 {
			path = result.Path
			break
		}
	}
	if err := probeCtx.Err(); err != nil {
		return packageCommand{}, fmt.Errorf("无法检查 npm 使用的 Node.js：%w", err)
	}
	if path == "" {
		return packageCommand{}, fmt.Errorf("未找到使用 Node.js 18 或更新版本的 npm，请先安装兼容的 Node.js 和 npm：%w", errMissingInstaller)
	}
	bin := userBinDir()
	if bin == "" {
		return packageCommand{}, errors.New("无法确定用户工具目录。")
	}
	prefix := filepath.Dir(bin)
	env, err := s.installation.environment(ctx)
	if err != nil {
		return packageCommand{}, err
	}
	return packageCommand{
		path: path,
		args: []string{"install", "--global", "--prefix", prefix, "pnpm@12"},
		dir:  filepath.Dir(prefix),
		env:  append(env, "PATH="+filepath.Dir(path)+string(os.PathListSeparator)+os.Getenv("PATH")),
	}, nil
}
