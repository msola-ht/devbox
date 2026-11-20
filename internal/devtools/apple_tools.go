package devtools

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
)

func developerToolPath(ctx context.Context, name string) (string, error) {
	// Locate the installed tool without running Apple's installation launcher.
	output, err := runCommand(ctx, "/usr/bin/xcode-select", []string{"--print-path"})
	if err != nil {
		return "", err
	}
	dir := strings.TrimSpace(string(output))
	if !filepath.IsAbs(dir) {
		return "", errors.New("invalid developer directory")
	}
	return exec.LookPath(filepath.Join(dir, "usr", "bin", name))
}
