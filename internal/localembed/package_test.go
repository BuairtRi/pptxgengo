package localembed

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPackageOptionsAndNonReplacingPublication(t *testing.T) {
	for _, v := range []struct {
		from     string
		download bool
	}{{"", false}, {"source", true}} {
		if _, err := PreparePackage(context.Background(), filepath.Join(t.TempDir(), "model"), v.from, v.download); err == nil {
			t.Fatal("ambiguous inputs accepted")
		}
	}
	target := filepath.Join(t.TempDir(), "snapshot.json")
	if err := WriteSnapshot(target, []byte("original")); err != nil {
		t.Fatal(err)
	}
	if err := WriteSnapshot(target, []byte("replacement")); err == nil {
		t.Fatal("existing file overwritten")
	}
	data, err := os.ReadFile(target)
	if err != nil || string(data) != "original" {
		t.Fatal(string(data), err)
	}
	if err := publishDirectory(t.TempDir(), t.TempDir()); err == nil {
		t.Fatal("existing directory replaced")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	out := filepath.Join(t.TempDir(), "cancelled")
	if _, err := PreparePackage(ctx, out, "source", false); err != context.Canceled {
		t.Fatal(err)
	}
	if _, err := os.Lstat(out); !os.IsNotExist(err) {
		t.Fatal("failed operation published output")
	}
}

func TestPinnedPackageOfflineCopy(t *testing.T) {
	from := os.Getenv("PPTXGENGO_EMBED_MODEL_DIR")
	if from == "" {
		t.Skip("optional pinned weights")
	}
	out := filepath.Join(t.TempDir(), "offline model")
	r, err := PreparePackage(context.Background(), out, from, false)
	if err != nil {
		t.Fatal(err)
	}
	if r.Identity.Revision != ModelRevision {
		t.Fatal(r)
	}
	for _, name := range []string{"manifest.json", "LICENSE", "README.txt"} {
		data, err := os.ReadFile(filepath.Join(out, name))
		if err != nil || len(data) == 0 {
			t.Fatal(name, err)
		}
	}
	license, _ := os.ReadFile(filepath.Join(out, "LICENSE"))
	if !strings.Contains(string(license), "Apache License") {
		t.Fatal("license missing")
	}
	if _, err = PreparePackage(context.Background(), out, from, false); err == nil {
		t.Fatal("package overwritten")
	}
	m, err := Load(out)
	if err != nil {
		t.Fatal(err)
	}
	m.Close()
}
