package runtimeenv

import (
	"context"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"devbox/internal/localcmd"
	"devbox/internal/readcache"
)

// Service exposes read-only information about the app's runtime environment.
type Service struct {
	environment readcache.Cache[EnvironmentInfo]
}

// EnvironmentInfo describes the user's home directory and account shell.
type EnvironmentInfo struct {
	LoginShell string `json:"loginShell"`
	HomeDir    string `json:"homeDir"`
}

// Get returns the last environment reading or reads it from the system.
func (s *Service) Get(ctx context.Context, refresh bool) (EnvironmentInfo, error) {
	return s.environment.Read(ctx, refresh, 0, readEnvironment)
}

func readEnvironment(ctx context.Context) (EnvironmentInfo, error) {
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
		output, err := localcmd.Output(ctx, "/usr/bin/dscacheutil", []string{"-q", "user", "-a", "uid", uid}, nil)
		if err != nil {
			return ""
		}
		for line := range strings.SplitSeq(string(output), "\n") {
			if shell, ok := strings.CutPrefix(line, "shell:"); ok {
				return strings.TrimSpace(shell)
			}
		}
	case "linux":
		output, err := localcmd.Output(ctx, "/usr/bin/getent", []string{"passwd", uid}, nil)
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
