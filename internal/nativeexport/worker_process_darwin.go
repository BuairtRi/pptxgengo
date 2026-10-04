package nativeexport

import (
	"os"
	"os/exec"
	"syscall"
)

// Deadline cancellation terminates only this worker and its helper descendants.
// PowerPoint is addressed through Apple events and is never in this process group.
func configureWorkerProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
}
