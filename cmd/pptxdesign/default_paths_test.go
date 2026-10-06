package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirectInstalledExecutableResolvesLatestBundle(t *testing.T) {
	root := t.TempDir()
	canonical, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	executable := filepath.Join(root, "bin", "pptxdesign")
	bundle := filepath.Join(root, "library", "wm-design-system", "v10", "bundle.json")
	for _, path := range []string{executable, bundle} {
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("fixture"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if got := releaseRootForDesignExecutable(executable); got != canonical {
		t.Fatalf("direct release root: %q; want %q", got, canonical)
	}
	link := filepath.Join(t.TempDir(), "pptxdesign")
	if err := os.Symlink(executable, link); err != nil {
		t.Fatal(err)
	}
	if got := releaseRootForDesignExecutable(link); got != canonical {
		t.Fatalf("linked release root: %q; want %q", got, canonical)
	}
	if err := os.Remove(bundle); err != nil {
		t.Fatal(err)
	}
	if got := releaseRootForDesignExecutable(executable); got != "" {
		t.Fatalf("invented release tree: %q", got)
	}
	t.Setenv("PPTXGENGO_RELEASE_ROOT", "explicit-release-root")
	if got := designReleaseRoot(); got != "explicit-release-root" {
		t.Fatalf("explicit root lost: %q", got)
	}
}
