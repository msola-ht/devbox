package runtimeenv

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Service exposes read-only information about the app's runtime environment.
type Service struct{}

// EnvironmentInfo describes the user's home directory and account shell.
type EnvironmentInfo struct {
	LoginShell string `json:"loginShell"`
	HomeDir    string `json:"homeDir"`
}

// Get reads the environment and the system account's login shell.
func (Service) Get(ctx context.Context) (EnvironmentInfo, error) {
	homeDir, _ := os.UserHomeDir()
	loginShell := readLoginShell(ctx)
	if err := ctx.Err(); err != nil {
		return EnvironmentInfo{}, err
	}
	return EnvironmentInfo{
		LoginShell: loginShell,
		HomeDir:    homeDir,
	}, nil
}

func readLoginShell(ctx context.Context) string {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	uid := strconv.Itoa(os.Getuid())
	switch runtime.GOOS {
	case "darwin":
		output, err := exec.CommandContext(ctx, "/usr/bin/dscacheutil", "-q", "user", "-a", "uid", uid).Output()
		if err != nil {
			return ""
		}
		for line := range strings.SplitSeq(string(output), "\n") {
			if shell, ok := strings.CutPrefix(line, "shell:"); ok {
				return strings.TrimSpace(shell)
			}
		}
	case "linux":
		output, err := exec.CommandContext(ctx, "/usr/bin/getent", "passwd", uid).Output()
		if err != nil {
			return ""
		}
		fields := strings.Split(strings.TrimRight(string(output), "\r\n"), ":")
		if len(fields) == 7 {
			return fields[6]
		}
	}
	return ""
}
