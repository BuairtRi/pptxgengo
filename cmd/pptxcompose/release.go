package main

import (
	"os"
	"path/filepath"
	"strings"
)

func releaseScript(name string) string {
	if root := os.Getenv("PPTXGENGO_RELEASE_ROOT"); root != "" {
		return filepath.Join(root, "scripts", name)
	}
	return filepath.Join("scripts", name)
}

// A packaged recipe can keep its source-relative asset name while caller
// paths continue to resolve against the caller's working directory first.
func readAsset(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err == nil || !os.IsNotExist(err) {
		return data, err
	}
	root := os.Getenv("PPTXGENGO_RELEASE_ROOT")
	if root == "" || !filepath.IsLocal(path) ||
		!(strings.HasPrefix(filepath.ToSlash(path), "samples/") || strings.HasPrefix(filepath.ToSlash(path), "library/")) {
		return nil, err
	}
	return os.ReadFile(filepath.Join(root, path))
}
