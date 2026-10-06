//go:build !windows

package nativeexport

import (
	"fmt"
	"os"
)

func protectTrustPermissions(path string) error { return nil }

func validateTrustPermissions(path string, info os.FileInfo) error {
	if info.Mode().Perm()&0077 != 0 {
		return fmt.Errorf("native trust file/directory must be private (0600/0700): %s", path)
	}
	return nil
}
