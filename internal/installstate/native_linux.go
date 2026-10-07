package installstate

import (
	"fmt"
	"golang.org/x/sys/unix"
)

func nativeArchitecture() (string, error) {
	var u unix.Utsname
	if e := unix.Uname(&u); e != nil {
		return "", e
	}
	b := []byte{}
	for _, c := range u.Machine {
		if c == 0 {
			break
		}
		b = append(b, byte(c))
	}
	switch string(b) {
	case "x86_64":
		return "amd64", nil
	case "aarch64", "arm64":
		return "arm64", nil
	}
	return "", fmt.Errorf("unsupported native architecture %q", string(b))
}

func publishDirectory(from, to string) error {
	return unix.Renameat2(unix.AT_FDCWD, from, unix.AT_FDCWD, to, unix.RENAME_NOREPLACE)
}
