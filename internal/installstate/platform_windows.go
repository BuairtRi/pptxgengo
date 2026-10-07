//go:build windows

package installstate

import (
	"fmt"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"os"
	"path/filepath"
	"unsafe"
)

func lockRoot(root string) (func(), error) {
	f, e := os.OpenFile(filepath.Join(root, "installation.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if e != nil {
		return nil, e
	}
	overlapped := new(windows.Overlapped)
	if e = windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, overlapped); e != nil {
		f.Close()
		return nil, e
	}
	return func() { windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, overlapped); f.Close() }, nil
}
func replaceFile(from, to string) error {
	a, e := windows.UTF16PtrFromString(from)
	if e != nil {
		return e
	}
	b, e := windows.UTF16PtrFromString(to)
	if e != nil {
		return e
	}
	return windows.MoveFileEx(a, b, windows.MOVEFILE_REPLACE_EXISTING|windows.MOVEFILE_WRITE_THROUGH)
}
func userPath() (pathSnapshot, error) {
	k, e := registry.OpenKey(registry.CURRENT_USER, "Environment", registry.QUERY_VALUE)
	if e == registry.ErrNotExist {
		return pathSnapshot{}, nil
	}
	if e != nil {
		return pathSnapshot{}, e
	}
	defer k.Close()
	s, kind, e := k.GetStringValue("Path")
	if e == registry.ErrNotExist {
		return pathSnapshot{}, nil
	}
	if e != nil {
		return pathSnapshot{}, e
	}
	if kind != registry.SZ && kind != registry.EXPAND_SZ {
		return pathSnapshot{}, fmt.Errorf("unsupported user PATH registry type")
	}
	return pathSnapshot{Value: s, Kind: kind, Exists: true}, nil
}
func setUserPath(value pathSnapshot) error {
	k, _, e := registry.CreateKey(registry.CURRENT_USER, "Environment", registry.SET_VALUE)
	if e != nil {
		return e
	}
	defer k.Close()
	if !value.Exists {
		e = k.DeleteValue("Path")
		if e == registry.ErrNotExist {
			e = nil
		}
	} else if value.Kind == registry.SZ {
		e = k.SetStringValue("Path", value.Value)
	} else if value.Kind == registry.EXPAND_SZ {
		e = k.SetExpandStringValue("Path", value.Value)
	} else {
		return fmt.Errorf("unsupported user PATH registry type")
	}
	if e != nil {
		return e
	}
	environment, e := windows.UTF16PtrFromString("Environment")
	if e != nil {
		return e
	}
	var result uintptr
	// Notify the desktop so newly opened terminals inherit the changed user PATH.
	windows.NewLazySystemDLL("user32.dll").NewProc("SendMessageTimeoutW").Call(0xffff, 0x001a, 0, uintptr(unsafe.Pointer(environment)), 2, 5000, uintptr(unsafe.Pointer(&result)))
	return nil
}

func nativeArchitecture() (string, error) {
	var process, native uint16
	if e := windows.IsWow64Process2(windows.CurrentProcess(), &process, &native); e != nil {
		return "", e
	}
	switch native {
	case 0x8664:
		return "amd64", nil
	case 0xaa64:
		return "arm64", nil
	}
	return "", fmt.Errorf("unsupported native Windows machine 0x%x", native)
}

func publishDirectory(from, to string) error {
	a, e := windows.UTF16PtrFromString(from)
	if e != nil {
		return e
	}
	b, e := windows.UTF16PtrFromString(to)
	if e != nil {
		return e
	}
	return windows.MoveFileEx(a, b, windows.MOVEFILE_WRITE_THROUGH)
}
