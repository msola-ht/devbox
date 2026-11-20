// Package localcmd runs short local commands with scoped cancellation.
package localcmd

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"time"
)

// Output runs a command and captures stdout. A nil env inherits the current
// environment. Cancellation terminates the command and its subprocesses.
func Output(ctx context.Context, path string, args, env []string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Env = env
	cmd.WaitDelay = time.Second
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := run(ctx, cmd)
	if canceled := ctx.Err(); canceled != nil {
		return nil, canceled
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		exitErr.Stderr = stderr.Bytes()
	}
	return stdout.Bytes(), err
}

// terminateOnCancel keeps the command group cancellable until all output is
// collected. The stop function waits for any cleanup already in progress.
func terminateOnCancel(ctx context.Context, terminate func()) func() bool {
	done := make(chan struct{})
	stop := context.AfterFunc(ctx, func() {
		terminate()
		close(done)
	})
	return func() bool {
		if stop() {
			return true
		}
		<-done
		return false
	}
}
