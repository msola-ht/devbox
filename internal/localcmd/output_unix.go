//go:build unix

package localcmd

import (
	"context"
	"errors"
	"os/exec"
	"syscall"
)

func run(ctx context.Context, cmd *exec.Cmd) error {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = nil
	if err := cmd.Start(); err != nil {
		return err
	}
	terminate := func() { _ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	stop := terminateOnCancel(ctx, terminate)
	err := cmd.Wait()
	if stop() && errors.Is(err, exec.ErrWaitDelay) {
		terminate()
	}
	return err
}
