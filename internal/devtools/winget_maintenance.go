package devtools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
)

const wingetModuleSetup = `
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'
$module = Get-Module -ListAvailable -Name Microsoft.WinGet.Client |
  Where-Object { $_.ExportedCommands.ContainsKey('Repair-WinGetPackageManager') } |
  Select-Object -First 1
if (-not $module) {
  [Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12
  $gallery = Get-PSRepository -Name PSGallery
  if ($gallery.SourceLocation.TrimEnd('/') -ne 'https://www.powershellgallery.com/api/v2') {
    throw 'PSGallery does not point to the official PowerShell Gallery.'
  }
  if (-not (Get-PackageProvider -Name NuGet -ListAvailable -ErrorAction SilentlyContinue)) {
    Install-PackageProvider -Name NuGet -Scope CurrentUser -Force | Out-Null
  }
  Install-Module -Name Microsoft.WinGet.Client -Repository PSGallery -Scope CurrentUser -Force | Out-Null
}
Import-Module Microsoft.WinGet.Client
Get-Command 'Microsoft.WinGet.Client\Repair-WinGetPackageManager' | Out-Null
`

func (s *Service) maintainWinGet(ctx context.Context, job *toolInstallJob) (ToolInstallStatus, string) {
	i := &s.installation
	if !i.claimInstaller(string(channelWinGet)) {
		return installFailed, "已有 Windows 安装任务进行中，请稍后重试。"
	}
	defer i.releaseInstaller(string(channelWinGet))

	results, err := s.DetectWinGet(ctx, true)
	if err != nil {
		return installFailed, fmt.Sprintf("无法检查当前 WinGet（%v）。", err)
	}
	index := slices.IndexFunc(results, availableTool)
	expectedVersion := ""
	if job.task.Action == updateAction {
		if index < 0 {
			return installFailed, "当前 WinGet 不可用，请先安装或修复。"
		}
		i.updateMessage(job, "正在检查 WinGet 稳定版…")
		latest, err := s.latestWinGetVersion(ctx)
		if err != nil {
			return installFailed, fmt.Sprintf("无法查询 WinGet 稳定版（%v）。", err)
		}
		current := results[index].Version
		currentNumber, valid := winGetVersionNumber(current)
		latestNumber, _ := winGetVersionNumber(latest)
		if !valid {
			return installFailed, "无法比较当前 WinGet 版本。"
		}
		if slices.Compare(currentNumber[:], latestNumber[:]) >= 0 {
			return installSucceeded, fmt.Sprintf("WinGet 当前为 %s，已达到稳定版 %s。", current, latest)
		}
		// The release tag is checked before being included in the PowerShell command.
		expectedVersion = latest
	}
	path, err := installerPowerShell()
	if err != nil {
		return installMissingInstaller, "未找到系统自带 Windows PowerShell，无法运行 App Installer 维护程序。"
	}

	i.updateMessage(job, "正在准备 Microsoft WinGet PowerShell 模块…")
	setupCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	err = installerPowerShellCommand(path, wingetModuleSetup).run(setupCtx, job)
	cancel()
	if err != nil {
		return installFailed, fmt.Sprintf("无法准备 WinGet 维护程序（%v）。请查看任务输出。", err)
	}

	versionArgument := ""
	if expectedVersion != "" {
		// The Store may have refreshed App Installer while the module was being prepared.
		current, err := s.DetectWinGet(ctx, true)
		if err != nil {
			return installFailed, fmt.Sprintf("无法复核当前 WinGet（%v）。", err)
		}
		index := slices.IndexFunc(current, availableTool)
		if index < 0 {
			return installFailed, "当前 WinGet 不可用，请先安装或修复。"
		}
		currentNumber, valid := winGetVersionNumber(current[index].Version)
		expectedNumber, _ := winGetVersionNumber(expectedVersion)
		if !valid {
			return installFailed, "无法比较当前 WinGet 版本。"
		}
		if slices.Compare(currentNumber[:], expectedNumber[:]) >= 0 {
			return installSucceeded, fmt.Sprintf("WinGet 当前为 %s，已达到稳定版 %s。", current[index].Version, expectedVersion)
		}
		versionArgument = " -Version '" + expectedVersion + "'"
	}
	script := `
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'
$env:GH_TOKEN = $null
$env:GITHUB_TOKEN = $null
[Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12
Import-Module Microsoft.WinGet.Client
Microsoft.WinGet.Client\Repair-WinGetPackageManager` + versionArgument + " -Verbose"
	err = i.runNativePackageCommand(ctx, job, installerPowerShellCommand(path, script), "正在维护 App Installer…")
	if errors.Is(err, context.Canceled) {
		return installCanceled, "已取消 WinGet 维护。"
	}
	i.updateMessage(job, "正在重新检测 WinGet…")
	results, detectErr := s.DetectWinGet(i.ctx, true)
	if err != nil {
		return installFailed, fmt.Sprintf("App Installer 维护失败（%v）。请查看任务输出。", err)
	}
	index = slices.IndexFunc(results, availableTool)
	if detectErr != nil || index < 0 {
		return installNotDetected, "App Installer 维护命令已完成，暂未检出可用的 WinGet。"
	}
	version := results[index].Version
	if expectedVersion != "" {
		currentNumber, _ := winGetVersionNumber(version)
		expectedNumber, _ := winGetVersionNumber(expectedVersion)
		if slices.Compare(currentNumber[:], expectedNumber[:]) < 0 {
			return installNotDetected, fmt.Sprintf("更新命令已完成，目前仍检出 WinGet %s。", version)
		}
		return installSucceeded, fmt.Sprintf("WinGet 已更新到 %s。", version)
	}
	return installSucceeded, fmt.Sprintf("WinGet %s 已可用。", version)
}

func (s *Service) latestWinGetVersion(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	client, err := s.installation.downloadClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.CloseIdleConnections()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/microsoft/winget-cli/releases/latest", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub 返回 HTTP %d", response.StatusCode)
	}
	var release struct {
		Tag string `json:"tag_name"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 2<<20)).Decode(&release); err != nil {
		return "", err
	}
	match := windowsPackageVersion.FindStringSubmatch(release.Tag)
	if len(match) == 0 || strings.TrimPrefix(release.Tag, "v") != match[1] {
		return "", errors.New("发布版本格式无效")
	}
	if _, valid := winGetVersionNumber(match[1]); !valid {
		return "", errors.New("发布版本格式无效")
	}
	return match[1], nil
}

func winGetVersionNumber(version string) ([3]int, bool) {
	var numbers [3]int
	match := windowsPackageVersion.FindStringSubmatch(version)
	if len(match) == 0 {
		return numbers, false
	}
	for index, field := range strings.Split(match[1], ".") {
		number, err := strconv.Atoi(field)
		if err != nil {
			return numbers, false
		}
		numbers[index] = number
	}
	return numbers, true
}
