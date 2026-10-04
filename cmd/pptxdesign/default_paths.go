package main

import (
	"os"
	"path/filepath"
)

// Installed tools can be invoked directly as well as through the wrapper.
// Repository development keeps relative paths when no release tree is present.
func designReleaseRoot() string {
	if root := os.Getenv("PPTXGENGO_RELEASE_ROOT"); root != "" {
		return root
	}
	executable, err := os.Executable()
	if err != nil {
		return ""
	}
	return releaseRootForDesignExecutable(executable)
}

func releaseRootForDesignExecutable(executable string) string {
	resolved, err := filepath.EvalSymlinks(executable)
	if err != nil {
		return ""
	}
	root := filepath.Dir(filepath.Dir(resolved))
	info, err := os.Stat(filepath.Join(root, "library", "wm-design-system", "v5", "bundle.json"))
	if err == nil && info.Mode().IsRegular() {
		return root
	}
	return ""
}
