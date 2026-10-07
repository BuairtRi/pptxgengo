package installstate

import (
	"errors"
	"golang.org/x/sys/unix"
)

func nativeArchitecture() (string, error) {
	arm, e := unix.SysctlUint32("hw.optional.arm64")
	if e == nil && arm == 1 {
		return "arm64", nil
	}
	if e != nil && !errors.Is(e, unix.ENOENT) {
		return "", e
	}
	return "amd64", nil
}

func publishDirectory(from, to string) error {
	return unix.RenameatxNp(unix.AT_FDCWD, from, unix.AT_FDCWD, to, unix.RENAME_EXCL)
}
