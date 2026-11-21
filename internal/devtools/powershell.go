package devtools

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

var powerShellVersion = regexp.MustCompile(`^[0-9]+\.[0-9]+(\.[0-9]+){0,2}(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$`)

// DetectPowerShell returns cached or newly detected PowerShell installations on Windows.
func (s *Service) DetectPowerShell(ctx context.Context, refresh bool) ([]ToolDetection, error) {
	return s.pwsh.Read(ctx, refresh, 0, detectPowerShell)
}

func detectPowerShell(ctx context.Context) ([]ToolDetection, error) {
	if runtime.GOOS != "windows" {
		return []ToolDetection{{
			Status:  statusUnsupported,
			Message: "此项检测适用于 Windows。",
		}}, nil
	}
	var paths []string
	if dir := os.Getenv("ProgramFiles"); filepath.IsAbs(dir) {
		paths = append(paths, filepath.Join(dir, "PowerShell", "7", "pwsh.exe"))
	}
	if path := systemPowerShellPath(); path != "" {
		paths = append(paths, path)
	}
	return detectTools(ctx, "pwsh", "在已检查的路径中没有找到 PowerShell。", probePowerShell, paths...)
}

func systemPowerShellPath() string {
	root := os.Getenv("SystemRoot")
	if !filepath.IsAbs(root) {
		return ""
	}
	return filepath.Join(root, "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
}

func probePowerShell(ctx context.Context, path string) ToolDetection {
	args := []string{"-NoLogo", "-NoProfile", "-NonInteractive", "-Command", "$PSVersionTable.PSVersion.ToString()"}
	output, err := runCommand(ctx, path, args, "POWERSHELL_TELEMETRY_OPTOUT=1", "POWERSHELL_UPDATECHECK=Off")
	version := firstOutputLine(output)
	if !powerShellVersion.MatchString(version) {
		version = ""
	}
	result := versionDetection(path, version, err)
	if result.Status == statusAvailable && strings.EqualFold(path, systemPowerShellPath()) {
		result.Source = sourceSystem
	}
	return result
}
