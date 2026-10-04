//go:build !darwin

package nativeexport

import "os/exec"

func configureWorkerProcess(cmd *exec.Cmd) {}
