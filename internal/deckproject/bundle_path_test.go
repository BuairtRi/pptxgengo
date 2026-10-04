package deckproject

import (
	"path/filepath"
	"testing"
)

func TestBundlePathAllPinnedRevisions(t *testing.T) {
	for _, root := range []string{"", t.TempDir()} {
		t.Setenv("PPTXGENGO_RELEASE_ROOT", root)
		for _, revision := range []string{"v1", "v2", "v3", "v4", "v5"} {
			want := filepath.Join(root, "library", "wm-design-system", revision)
			if got := BundlePath(revision); got != want {
				t.Errorf("BundlePath(%q) = %q; want %q", revision, got, want)
			}
		}
		for _, explicit := range []string{"v6", "./custom-bundle", filepath.Join(t.TempDir(), "bundle")} {
			if got := BundlePath(explicit); got != explicit {
				t.Errorf("explicit bundle %q changed to %q", explicit, got)
			}
		}
	}
}
