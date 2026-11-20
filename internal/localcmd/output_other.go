//go:build !unix && !windows

package localcmd

import (
	"context"
	"os/exec"
)

func run(_ context.Context, cmd *exec.Cmd) error {
	return cmd.Run()
}
