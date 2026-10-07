//go:build darwin || linux

package installstate

import (
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
)

func lockRoot(root string) (func(), error) {
	f, e := os.OpenFile(filepath.Join(root, "installation.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if e != nil {
		return nil, e
	}
	if e = unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB); e != nil {
		f.Close()
		return nil, e
	}
	return func() { unix.Flock(int(f.Fd()), unix.LOCK_UN); f.Close() }, nil
}
func replaceFile(from, to string) error {
	if e := os.Rename(from, to); e != nil {
		return e
	}
	d, e := os.Open(filepath.Dir(to))
	if e != nil {
		return e
	}
	defer d.Close()
	return d.Sync()
}
func userPath() (pathSnapshot, error)      { return pathSnapshot{}, nil }
func setUserPath(value pathSnapshot) error { return nil }
