package devtools

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"unicode/utf16"
)

const chocolateyBootstrap = `
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'
[Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12
$env:chocolateyDownloadUrl = $null
$env:chocolateyVersion = $null
$env:ChocolateyInstall = Join-Path ([Environment]::GetFolderPath('CommonApplicationData')) 'chocolatey'
$bootstrap = Invoke-WebRequest -UseBasicParsing -TimeoutSec 30 -Uri 'https://community.chocolatey.org/install.ps1'
& ([ScriptBlock]::Create($bootstrap.Content))
`

func (s *Service) installChocolatey(ctx context.Context, job *toolInstallJob) (ToolInstallStatus, string) {
	i := &s.installation
	if !i.claimInstaller(string(channelWinGet)) {
		return installFailed, "已有 Windows 安装任务进行中，请稍后重试。"
	}
	defer i.releaseInstaller(string(channelWinGet))

	results, err := s.DetectChocolatey(ctx, true)
	if err != nil {
		return installFailed, fmt.Sprintf("无法检测 Chocolatey（%v）。", err)
	}
	if index := slices.IndexFunc(results, availableTool); index >= 0 {
		return installSucceeded, fmt.Sprintf("Chocolatey %s 已可用。", results[index].Version)
	}
	path, err := installerPowerShell()
	if err != nil {
		return installMissingInstaller, "未找到系统自带 Windows PowerShell，无法运行 Chocolatey 安装程序。"
	}

	cmd := installerPowerShellCommand(path, chocolateyElevatedScript())
	err = i.runNativePackageCommand(ctx, job, cmd, "正在安装 Chocolatey，请完成系统授权。安装窗口会显示输出。")
	i.updateMessage(job, "正在重新检测 Chocolatey…")
	results, detectErr := s.DetectChocolatey(i.ctx, true)
	var exitErr *exec.ExitError
	if errors.Is(err, context.Canceled) ||
		errors.As(err, &exitErr) && exitErr.ExitCode() == 126 {
		return installCanceled, "已取消 Chocolatey 安装。"
	}
	if err != nil {
		return installFailed, fmt.Sprintf("Chocolatey 安装失败（%v）。可查看官方安装说明。", err)
	}
	index := slices.IndexFunc(results, availableTool)
	if detectErr != nil || index < 0 {
		return installNotDetected, "安装程序已退出，暂未检出可用的 Chocolatey。"
	}
	return installSucceeded, fmt.Sprintf("Chocolatey %s 已可用。", results[index].Version)
}

func chocolateyElevatedScript() string {
	var payload []byte
	for _, unit := range utf16.Encode([]rune(chocolateyBootstrap)) {
		payload = binary.LittleEndian.AppendUint16(payload, unit)
	}
	return fmt.Sprintf(`
$ErrorActionPreference = 'Stop'
try {
  $process = Start-Process -FilePath (Join-Path $PSHOME 'powershell.exe') -WorkingDirectory $PSHOME -Verb RunAs -Wait -PassThru -ArgumentList @('-NoLogo','-NoProfile','-NonInteractive','-ExecutionPolicy','Bypass','-EncodedCommand','%s')
} catch {
  for ($cause = $_.Exception; $cause; $cause = $cause.InnerException) {
    if ($cause -is [ComponentModel.Win32Exception] -and $cause.NativeErrorCode -eq 1223) { exit 126 }
  }
  throw
}
exit $process.ExitCode
`, base64.StdEncoding.EncodeToString(payload))
}
