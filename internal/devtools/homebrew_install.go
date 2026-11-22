package devtools

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"

	"devbox/internal/localcmd"
)

func (s *Service) installHomebrewTool(ctx context.Context, tool InstallTool, recipe installRecipe, job *toolInstallJob) (ToolInstallStatus, string) {
	brews, err := s.DetectHomebrew(ctx, true)
	if err != nil {
		return installFailed, fmt.Sprintf("无法检测 Homebrew（%v）。", err)
	}
	index := slices.IndexFunc(brews, availableTool)
	if index < 0 {
		return installMissingHomebrew, "未找到可用的 Homebrew。请先安装 Homebrew，再重试。"
	}
	brew := brews[index].Path
	output, err := runCommand(ctx, brew, []string{"--prefix"})
	prefix := firstOutputLine(output)
	if err != nil || !filepath.IsAbs(prefix) {
		return installFailed, "无法读取 Homebrew 安装位置。请检查 Homebrew 后重试。"
	}
	prefix, err = filepath.EvalSymlinks(prefix)
	if err != nil {
		return installFailed, fmt.Sprintf("无法访问 Homebrew 安装位置（%v）。", err)
	}
	install := &s.installation
	if !install.claimInstaller(prefix) {
		return installFailed, "该 Homebrew 正在安装另一个工具。请在当前任务完成后重试。"
	}
	defer install.releaseInstaller(prefix)

	bindir := filepath.Join(prefix, "bin")
	if recipe.brew == "node@24" {
		bindir = filepath.Join(prefix, "opt", recipe.brew, "bin")
	}
	location := homebrewLocation{bin: filepath.Join(bindir, recipe.bin)}
	if output, err := runCommand(ctx, brew, []string{"--cellar"}); err == nil {
		if cellar := firstOutputLine(output); filepath.IsAbs(cellar) {
			location.cellar = cellar
		}
	}
	s.homebrewPaths.remember(tool, location)

	if results, err := recipe.detect(s, ctx, true); err == nil && slices.ContainsFunc(results, availableTool) {
		s.refreshCompanionTool(ctx, tool)
		return installSucceeded, fmt.Sprintf("已检测到可用的 %s。", recipe.name)
	}
	env, err := install.environment(ctx)
	if err != nil {
		return installFailed, fmt.Sprintf("无法准备 Homebrew 安装环境（%v）。", err)
	}
	env = append(env,
		"HOMEBREW_NO_ASK=1",
		"HOMEBREW_NO_AUTO_UPDATE=1",
		"HOMEBREW_NO_INSTALL_CLEANUP=1",
		"HOMEBREW_NO_INSTALL_UPGRADE=1",
		"HOMEBREW_NO_INSTALLED_DEPENDENTS_CHECK=1",
		"NONINTERACTIVE=1",
	)
	install.updateMessage(job, fmt.Sprintf("正在通过 Homebrew 安装 %s…", recipe.name))
	runErr := localcmd.Run(ctx, brew, []string{"install", "--formula", recipe.brew}, env, &job.output, &job.output)

	// The command has finished changing the system. Finish detection even if the user canceled it.
	install.closeCancellation(ctx, job, fmt.Sprintf("正在重新检测 %s…", recipe.name))
	results, companion, err := s.refreshInstalledTool(tool, recipe)
	if runErr != nil {
		return installFailed, fmt.Sprintf("Homebrew 安装 %s 失败（%v）。", recipe.name, runErr)
	}
	if err != nil {
		return installNotDetected, fmt.Sprintf("Homebrew 安装命令已完成，重新检测 %s 失败（%v）。", recipe.name, err)
	}
	if availableAt(results, location.bin) {
		if tool == toolNode || tool == toolNPM {
			other := installRecipes[companionTool(tool)]
			if !availableAt(companion, filepath.Join(bindir, other.bin)) {
				return installNotDetected, fmt.Sprintf("已检测到 %s，暂未检出同一 formula 提供的 %s。", recipe.name, other.name)
			}
			return installSucceeded, fmt.Sprintf("已通过 Homebrew 安装 %s。终端若未检出，可将 %s 加入 PATH。", recipe.name, bindir)
		}
		return installSucceeded, fmt.Sprintf("已通过 Homebrew 安装 %s。", recipe.name)
	}
	return installNotDetected, fmt.Sprintf("Homebrew 安装命令已完成，未能读取该安装的 %s 版本。", recipe.name)
}
