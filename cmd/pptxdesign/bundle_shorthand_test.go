package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBundleShorthandEveryCommandRoute(t *testing.T) {
	root := t.TempDir()
	t.Setenv("PPTXGENGO_RELEASE_ROOT", root)
	for _, revision := range []string{"v1", "v2", "v3", "v4", "v5", "v6", "v7"} {
		for _, route := range []string{"catalog", "search", "index"} {
			t.Run(revision+"/"+route, func(t *testing.T) {
				var err error
				switch route {
				case "catalog":
					previous := os.Args
					os.Args = []string{"pptxdesign", "library-catalog", "--bundle", revision, "--engine", "wmds-go-foundation.v2"}
					defer func() { os.Args = previous }()
					err = run()
				case "search":
					err = runLibrarySearch([]string{"--bundle", revision, "--query", "maturity"})
				case "index":
					err = runLibraryIndex("library-index", []string{"--bundle", revision, "--out", filepath.Join(t.TempDir(), "new.sqlite")})
				}
				want := filepath.Join(root, "library", "wm-design-system", revision)
				if err == nil || !strings.Contains(err.Error(), want) {
					t.Fatalf("route %s did not resolve %s to %s: %v", route, revision, want, err)
				}
			})
		}
	}
}
