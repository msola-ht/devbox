// Package localcmd runs local commands with scoped cancellation.
package localcmd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
	"time"
)

// Output runs a command and captures stdout. A nil env inherits the current
// environment. Cancellation terminates the command and its subprocesses.
func Output(ctx context.Context, path string, args, env []string) ([]byte, error) {
	var stdout, stderr bytes.Buffer
	err := Run(ctx, path, args, env, &stdout, &stderr)
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		exitErr.Stderr = stderr.Bytes()
	}
	if canceled := ctx.Err(); canceled != nil {
		return nil, canceled
	}
	return stdout.Bytes(), err
}

// Run forwards command output to the given writers and cancels the process group.
func Run(ctx context.Context, path string, args, env []string, stdout, stderr io.Writer) error {
	return RunInDirectory(ctx, "", path, args, env, stdout, stderr)
}

// RunInDirectory runs a command with an explicit working directory.
func RunInDirectory(ctx context.Context, dir, path string, args, env []string, stdout, stderr io.Writer) error {
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Dir = dir
	cmd.Env = env
	cmd.WaitDelay = time.Second
	cmd.Stdout, cmd.Stderr = stdout, stderr
	err := run(ctx, cmd)
	if canceled := ctx.Err(); canceled != nil {
		return canceled
	}
	return err
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
