package devtools

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
)

var windowsPackageVersion = regexp.MustCompile(`^v?([0-9]+\.[0-9]+\.[0-9]+)(?:[-+][0-9A-Za-z.+-]+)?$`)

// DetectWinGet reads the version of the Windows App Installer command.
func (s *Service) DetectWinGet(ctx context.Context, refresh bool) ([]ToolDetection, error) {
	return s.winget.Read(ctx, refresh, 0, func(ctx context.Context) ([]ToolDetection, error) {
		if runtime.GOOS != "windows" {
			return []ToolDetection{{Status: statusUnsupported, Message: "WinGet 适用于 Windows。"}}, nil
		}
		return detectTools(ctx, "winget.exe", "未找到 WinGet。", probeWindowsPackageManager)
	})
}

// DetectChocolatey reads the version of the Chocolatey command.
func (s *Service) DetectChocolatey(ctx context.Context, refresh bool) ([]ToolDetection, error) {
	return s.chocolatey.Read(ctx, refresh, 0, func(ctx context.Context) ([]ToolDetection, error) {
		if runtime.GOOS != "windows" {
			return []ToolDetection{{Status: statusUnsupported, Message: "Chocolatey 适用于 Windows。"}}, nil
		}
		var paths []string
		if dir := os.Getenv("ChocolateyInstall"); filepath.IsAbs(dir) {
			paths = append(paths, filepath.Join(dir, "bin", "choco.exe"))
		}
		if dir := os.Getenv("ProgramData"); filepath.IsAbs(dir) {
			paths = append(paths, filepath.Join(dir, "chocolatey", "bin", "choco.exe"))
		}
		return detectTools(ctx, "choco.exe", "未找到 Chocolatey。", probeWindowsPackageManager, paths...)
	})
}

func probeWindowsPackageManager(ctx context.Context, path string) ToolDetection {
	output, err := runCommand(ctx, path, []string{"--version"})
	version := firstOutputLine(output)
	if !windowsPackageVersion.MatchString(version) {
		version = ""
	}
	return versionDetection(path, version, err)
}

func installerPowerShell() (string, error) {
	path := systemPowerShellPath()
	if path == "" {
		return "", errMissingInstaller
	}
	return exec.LookPath(path)
}

func installerPowerShellCommand(path, script string) packageCommand {
	return packageCommand{path: path, args: []string{
		"-NoLogo", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "RemoteSigned", "-Command", script,
	}}
}
