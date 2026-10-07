// Package noreplacedir publishes a staged sibling directory without replacing
// an existing destination. Callers own staging, validation and cleanup.
package noreplacedir

import "golang.org/x/sys/unix"

func Publish(from, to string) error {
	return unix.RenameatxNp(unix.AT_FDCWD, from, unix.AT_FDCWD, to, unix.RENAME_EXCL)
}
