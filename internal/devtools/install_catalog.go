package devtools

import (
	"bytes"
	"context"
	"errors"
	"os"
	"runtime"
	"strconv"
	"strings"
)

// InstallTool is the fixed identifier of a tool Devbox can install.
type InstallTool string

const (
	toolHomebrew   InstallTool = "homebrew"
	toolWinGet     InstallTool = "winget"
	toolChocolatey InstallTool = "chocolatey"
	toolUV         InstallTool = "uv"
	toolRG         InstallTool = "ripgrep"
	toolJQ         InstallTool = "jq"
	toolGit        InstallTool = "git"
	toolGH         InstallTool = "gh"
	toolPython     InstallTool = "python"
	toolNode       InstallTool = "node"
	toolNPM        InstallTool = "npm"
	toolPNPM       InstallTool = "pnpm"
	toolZsh        InstallTool = "zsh"
	toolPowerShell InstallTool = "powershell"
)

// InstallChannel identifies the platform installer used by a tool card.
type InstallChannel string

const (
	channelHomebrew      InstallChannel = "homebrew"
	channelHomebrewSetup InstallChannel = "homebrew_setup"
	channelWinGet        InstallChannel = "winget"
	channelAppInstaller  InstallChannel = "app_installer"
	channelChocolatey    InstallChannel = "chocolatey"
	channelAPT           InstallChannel = "apt"
	channelAstral        InstallChannel = "astral"
	channelNPM           InstallChannel = "npm"
)

// ToolInstallOption is the installation option exposed on this machine.
type ToolInstallOption struct {
	Tool    InstallTool    `json:"tool"`
	Channel InstallChannel `json:"channel"`
}

type installRecipe struct {
	name   string
	bin    string
	detect func(*Service, context.Context, bool) ([]ToolDetection, error)
	brew   string
	winget string
	apt    []string
	user   InstallChannel
}

var installRecipes = map[InstallTool]installRecipe{
	toolUV:         {"uv", "uv", (*Service).DetectUV, "uv", "astral-sh.uv", nil, channelAstral},
	toolRG:         {"ripgrep", "rg", (*Service).DetectRG, "ripgrep", "BurntSushi.ripgrep.MSVC", []string{"ripgrep"}, ""},
	toolJQ:         {"jq", "jq", (*Service).DetectJQ, "jq", "jqlang.jq", []string{"jq"}, ""},
	toolGit:        {"Git", "git", (*Service).DetectGit, "git", "Git.Git", []string{"git"}, ""},
	toolGH:         {"GitHub CLI", "gh", (*Service).DetectGH, "gh", "GitHub.cli", []string{"gh"}, ""},
	toolPython:     {"Python", "python3", (*Service).DetectPython, "python", "Python.Python.3.14", []string{"python3"}, ""},
	toolNode:       {"Node.js", "node", (*Service).DetectNode, "node@24", "OpenJS.NodeJS.LTS", []string{"nodejs", "npm"}, ""},
	toolNPM:        {"npm", "npm", (*Service).DetectNPM, "node@24", "OpenJS.NodeJS.LTS", []string{"npm"}, ""},
	toolPNPM:       {"pnpm", "pnpm", (*Service).DetectPNPM, "pnpm", "pnpm.pnpm", nil, channelNPM},
	toolZsh:        {"Zsh", "zsh", (*Service).DetectZsh, "", "", []string{"zsh"}, ""},
	toolPowerShell: {"PowerShell 7", "pwsh", (*Service).DetectPowerShell, "", "Microsoft.PowerShell", nil, ""},
}

func (recipe installRecipe) channel(ubuntu bool) InstallChannel {
	switch runtime.GOOS {
	case "darwin":
		if recipe.brew != "" {
			return channelHomebrew
		}
	case "windows":
		if recipe.winget != "" {
			return channelWinGet
		}
	case "linux":
		if ubuntu {
			if len(recipe.apt) > 0 {
				return channelAPT
			}
			if recipe.user != channelAstral || runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64" {
				return recipe.user
			}
		}
	}
	return ""
}

func isUbuntu() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return false
	}
	for line := range bytes.SplitSeq(data, []byte("\n")) {
		if id, ok := strings.CutPrefix(string(line), "ID="); ok {
			return strings.Trim(id, "\"'") == "ubuntu"
		}
	}
	return false
}

// GetToolInstallOptions lists the installation channels supported on this machine.
func (s *Service) GetToolInstallOptions() []ToolInstallOption {
	options := make([]ToolInstallOption, 0, len(installRecipes)+1)
	ubuntu := isUbuntu()
	for tool, recipe := range installRecipes {
		if channel := recipe.channel(ubuntu); channel != "" {
			options = append(options, ToolInstallOption{Tool: tool, Channel: channel})
		}
	}
	if runtime.GOOS == "darwin" {
		options = append(options, ToolInstallOption{Tool: toolHomebrew, Channel: channelHomebrewSetup})
	}
	if runtime.GOOS == "windows" {
		options = append(options,
			ToolInstallOption{Tool: toolWinGet, Channel: channelAppInstaller},
			ToolInstallOption{Tool: toolChocolatey, Channel: channelChocolatey},
		)
	}
	return options
}

// StartToolInstall starts the supported installation for a tool.
func (s *Service) StartToolInstall(ctx context.Context, tool InstallTool) (ToolInstallTask, error) {
	target := installTarget{tool: tool}
	if runtime.GOOS == "windows" {
		switch tool {
		case toolWinGet:
			return s.installation.start(ctx, target, installAction, s.maintainWinGet)
		case toolChocolatey:
			return s.installation.start(ctx, target, installAction, s.installChocolatey)
		}
	}
	recipe := installRecipes[tool]
	channel := recipe.channel(isUbuntu())
	if channel == "" {
		return ToolInstallTask{}, errors.New("当前平台没有适用的安装渠道。")
	}
	return s.installation.start(ctx, target, installAction, func(ctx context.Context, job *toolInstallJob) (ToolInstallStatus, string) {
		return s.installLocalTool(ctx, tool, recipe, channel, job)
	})
}

// StartWinGetUpdate updates the current user's App Installer to a newer stable release.
func (s *Service) StartWinGetUpdate(ctx context.Context) (ToolInstallTask, error) {
	if runtime.GOOS != "windows" {
		return ToolInstallTask{}, errors.New("WinGet 更新仅适用于 Windows。")
	}
	return s.installation.start(ctx, installTarget{tool: toolWinGet}, updateAction, s.maintainWinGet)
}

// GetToolInstall reads the most recent local installation for a tool.
func (s *Service) GetToolInstall(tool InstallTool) *ToolInstallTask {
	return s.installation.get(installTarget{tool: tool})
}

// CancelToolInstall requests cancellation of the matching local installation.
func (s *Service) CancelToolInstall(tool InstallTool, id int) *ToolInstallTask {
	return s.installation.cancel(installTarget{tool: tool}, id)
}

func availableForInstall(tool InstallTool, result ToolDetection) bool {
	if result.Status != statusAvailable {
		return false
	}
	if tool != toolPowerShell {
		return true
	}
	major, _, _ := strings.Cut(result.Version, ".")
	version, _ := strconv.Atoi(major)
	return result.Source != sourceSystem && version >= 7
}

func companionTool(tool InstallTool) InstallTool {
	switch tool {
	case toolNode:
		return toolNPM
	case toolNPM:
		return toolNode
	default:
		return ""
	}
}
