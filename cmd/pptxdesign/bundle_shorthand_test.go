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
	for _, revision := range []string{"v1", "v2", "v3", "v4", "v5", "v6", "v7", "v8", "v9"} {
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

func TestV7ShorthandResolvesFrozenIntake(t *testing.T) {
	testHistoricalShorthandResolvesFrozenIntake(t, "v7", "intake-20261005-616-frozen")
}

func TestV8ShorthandResolvesFrozenIntake(t *testing.T) {
	testHistoricalShorthandResolvesFrozenIntake(t, "v8", "intake-20261006-623-frozen")
}

func testHistoricalShorthandResolvesFrozenIntake(t *testing.T, revision, intake string) {
	t.Helper()
	root := t.TempDir()
	t.Setenv("PPTXGENGO_RELEASE_ROOT", root)
	path := filepath.Join(root, "planning", "wm-design-contracts", revision, intake, "bundle")
	if err := os.MkdirAll(path, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "bundle.json"), []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	if got := designBundlePath(revision); got != path {
		t.Fatalf("historical %s path = %q; want %q", revision, got, path)
	}
}
