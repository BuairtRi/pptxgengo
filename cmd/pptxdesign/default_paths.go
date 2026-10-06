package main

import (
	"os"
	"path/filepath"
)

const currentDesignBundle = "v11"

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
	for _, revision := range []string{currentDesignBundle, "v10", "v9"} {
		info, err := os.Stat(filepath.Join(root, "library", "wm-design-system", revision, "bundle.json"))
		if err == nil && info.Mode().IsRegular() {
			return root
		}
	}
	return ""
}

// Historic source snapshots remain available in a repository checkout. A
// packaged release ships only its current bundle; projects retain their own
// locked snapshot for continued offline operation.
func designBundlePath(revision string) string {
	root := designReleaseRoot()
	direct := filepath.Join(root, "library", "wm-design-system", revision)
	if _, err := os.Stat(filepath.Join(direct, "bundle.json")); err == nil {
		return direct
	}
	historical := map[string]string{
		"v10": "intake-20261006-649-frozen",
		"v5":  "intake-20261003-587-frozen",
		"v6":  "intake-20261004-602-frozen",
		"v7":  "intake-20261005-616-frozen",
		"v8":  "intake-20261006-623-frozen",
		"v9":  "intake-20261006-631-frozen",
	}
	if intake, ok := historical[revision]; ok {
		path := filepath.Join(root, "planning", "wm-design-contracts", revision, intake, "bundle")
		if _, err := os.Stat(filepath.Join(path, "bundle.json")); err == nil {
			return path
		}
	}
	return direct
}
