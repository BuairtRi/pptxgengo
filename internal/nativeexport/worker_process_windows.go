package nativeexport

import (
	"errors"
	"golang.org/x/sys/windows"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"
)

func configureWorkerProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x00000200} // CREATE_NEW_PROCESS_GROUP
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		// Terminate only direct PowerShell helpers, then this worker. A general
		// tree kill could terminate a COM-launched PowerPoint, so it is forbidden.
		helperErr := terminateWindowsHelpers(uint32(cmd.Process.Pid))
		return errors.Join(helperErr, cmd.Process.Kill())
	}
}

func terminateWindowsHelpers(workerPID uint32) error {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(snapshot)
	entry := windows.ProcessEntry32{Size: uint32(unsafe.Sizeof(windows.ProcessEntry32{}))}
	for err = windows.Process32First(snapshot, &entry); err == nil; err = windows.Process32Next(snapshot, &entry) {
		if entry.ParentProcessID != workerPID || !strings.EqualFold(windows.UTF16ToString(entry.ExeFile[:]), "powershell.exe") {
			continue
		}
		process, openErr := windows.OpenProcess(windows.PROCESS_TERMINATE, false, entry.ProcessID)
		if openErr != nil {
			if openErr == windows.ERROR_INVALID_PARAMETER {
				continue
			}
			return openErr
		}
		killErr := windows.TerminateProcess(process, 1)
		windows.CloseHandle(process)
		if killErr != nil {
			return killErr
		}
	}
	if err == windows.ERROR_NO_MORE_FILES {
		return nil
	}
	return err
}
