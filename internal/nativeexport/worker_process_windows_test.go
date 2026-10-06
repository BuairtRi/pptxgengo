package nativeexport

import (
	"context"
	"errors"
	"golang.org/x/sys/windows"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestWindowsWorkerCancellationClosesOnlyOwnedHelper(t *testing.T) {
	sibling := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", "Start-Sleep -Seconds 60")
	if err := sibling.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { sibling.Process.Kill(); sibling.Wait() }()
	t.Setenv("PPTXGENGO_TEST_BLOCK_WORKER", "1")
	ready := filepath.Join(t.TempDir(), "ready.txt")
	t.Setenv("PPTXGENGO_TEST_WORKER_READY", ready)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := workerProcess(ctx, workerRequest{Doctor: &DoctorOptions{}}); done <- err }()
	var pid int
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(ready)
		if err == nil {
			pid, err = strconv.Atoi(strings.TrimSpace(string(data)))
			if err != nil {
				t.Fatal(err)
			}
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if pid == 0 {
		t.Fatal("Windows helper did not become ready")
	}
	cancel()
	if err := <-done; err == nil || !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	active := func(pid uint32) bool {
		handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
		if err != nil {
			return false
		}
		defer windows.CloseHandle(handle)
		var exit uint32
		if err := windows.GetExitCodeProcess(handle, &exit); err != nil {
			t.Fatal(err)
		}
		return exit == 259
	}
	if active(uint32(pid)) {
		t.Fatal("owned PowerShell helper survived cancellation")
	}
	if !active(uint32(sibling.Process.Pid)) {
		t.Fatal("unrelated PowerShell process was terminated")
	}
}
