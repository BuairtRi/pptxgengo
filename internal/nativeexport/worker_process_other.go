//go:build !darwin && !windows

package nativeexport

import "os/exec"

func configureWorkerProcess(cmd *exec.Cmd) {}
