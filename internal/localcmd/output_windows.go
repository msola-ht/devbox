package localcmd

import (
	"context"
	"errors"
	"os/exec"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func run(ctx context.Context, cmd *exec.Cmd) error {
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(job)
	limits := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	limits.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	if _, err := windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&limits)), uint32(unsafe.Sizeof(limits))); err != nil {
		return err
	}

	// Attach before resuming so the command cannot start an untracked child.
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: windows.CREATE_NO_WINDOW | windows.CREATE_SUSPENDED}
	cmd.Cancel = nil
	if err := cmd.Start(); err != nil {
		return err
	}
	stop := terminateOnCancel(ctx, func() { _ = windows.TerminateJobObject(job, 1) })
	defer stop()
	if err := startJobProcess(ctx, cmd, job); err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return err
	}
	return cmd.Wait()
}

func startJobProcess(ctx context.Context, cmd *exec.Cmd, job windows.Handle) error {
	var attachErr error
	if err := cmd.Process.WithHandle(func(handle uintptr) {
		attachErr = windows.AssignProcessToJobObject(job, windows.Handle(handle))
	}); err != nil {
		return err
	}
	if attachErr != nil {
		return attachErr
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPTHREAD, 0)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(snapshot)
	entry := windows.ThreadEntry32{}
	entry.Size = uint32(unsafe.Sizeof(entry))
	for err = windows.Thread32First(snapshot, &entry); err == nil; err = windows.Thread32Next(snapshot, &entry) {
		if entry.OwnerProcessID != uint32(cmd.Process.Pid) {
			continue
		}
		thread, err := windows.OpenThread(windows.THREAD_SUSPEND_RESUME, false, entry.ThreadID)
		if err != nil {
			return err
		}
		defer windows.CloseHandle(thread)
		_, err = windows.ResumeThread(thread)
		return err
	}
	if !errors.Is(err, windows.ERROR_NO_MORE_FILES) {
		return err
	}
	return errors.New("找不到命令的初始线程")
}
