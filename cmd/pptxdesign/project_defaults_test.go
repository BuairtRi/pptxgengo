package main

import (
	"archive/zip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/buairtri/pptxgengo/internal/deckproject"
)

const defaultPinFixture = `schema: pptxgengo.deck-document.v1
id: pin-resumption-fixture
title: Synthetic pin resumption fixture
year: 2026
toolchain: {lockfile: toolchain.lock.json}
local_templates:
  plain:
    name: Synthetic editable text
    frame: {scope: shared, id: wmds/frame/none-compact}
    grid: {scope: shared, id: wmds/grid/12-columns}
    zones:
      title: {role: slide-title, required: true, schema: {type: string, maxLength: 70}}
      copy: {role: body-copy, required: true, schema: {type: string, maxLength: 100}}
    nodes:
      - id: body
        kind: text
        placement: {zone: body, span: {start: 1, count: 8, y_pt: 24, height_pt: 96}}
        style: body
        ink: primary
        text: {binding: copy}
slides:
  - id: stable-page
    content_kind: synthetic_example
    template: {scope: local, id: plain}
    values: {title: Synthetic maintained source, copy: Exact project pins survive published defaults.}
`

func projectDefaultFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "deck.yaml"), []byte(defaultPinFixture), 0600); err != nil {
		t.Fatal(err)
	}
	return root
}

func copyBundleFixture(t *testing.T, source, destination string) {
	t.Helper()
	if err := os.MkdirAll(destination, 0700); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(source)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(source, entry.Name())
		destinationPath := filepath.Join(destination, entry.Name())
		info, err := os.Stat(sourcePath) // Deliberately materialize historical asset symlinks.
		if err != nil {
			t.Fatal(err)
		}
		if info.IsDir() {
			copyBundleFixture(t, sourcePath, destinationPath)
			continue
		}
		data, err := os.ReadFile(sourcePath)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(destinationPath, data, info.Mode().Perm()); err != nil {
			t.Fatal(err)
		}
	}
}

func projectCommandJSON(t *testing.T, args ...string) json.RawMessage {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	previous := os.Stdout
	os.Stdout = f
	err = runProject(args)
	os.Stdout = previous
	if err != nil {
		t.Fatalf("%v: %v", args, err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	b, err := io.ReadAll(f)
	if err != nil || !json.Valid(b) {
		t.Fatalf("invalid command JSON %q: %v", b, err)
	}
	return b
}

func TestPublishedProjectBundleMetadata(t *testing.T) {
	root := t.TempDir()
	if got, err := publishedProjectBundle(root); err != nil || got != "v9" {
		t.Fatalf("latest default: %q %v", got, err)
	}
	if err := os.Mkdir(filepath.Join(root, "release"), 0700); err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"v1", "v2", "v3", "v4", "v5", "v6", "v7", "v8", "v9", "../v5", "v5 v3", ""} {
		if err := os.WriteFile(filepath.Join(root, "release/default-bundle.txt"), []byte(value+"\n"), 0600); err != nil {
			t.Fatal(err)
		}
		got, err := publishedProjectBundle(root)
		if validPublishedBundle(value) {
			if err != nil || got != value {
				t.Errorf("metadata %q: %q %v", value, got, err)
			}
		} else if err == nil {
			t.Errorf("invalid metadata %q accepted", value)
		}
	}
}

func TestProjectPublishedV9PreservesLockedRevisions(t *testing.T) {
	stage := t.TempDir()
	repoLibrary, err := filepath.Abs("../../library")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(repoLibrary, filepath.Join(stage, "library")); err != nil {
		t.Fatal(err)
	}
	repoPlanning, err := filepath.Abs("../../planning/wm-design-contracts")
	if err != nil {
		t.Fatal(err)
	}
	for _, historical := range []struct{ revision, intake string }{
		{"v5", "intake-20261003-587-frozen"},
		{"v7", "intake-20261005-616-frozen"},
		{"v8", "intake-20261006-623-frozen"},
	} {
		from := filepath.Join(repoPlanning, historical.revision, historical.intake, "bundle")
		to := filepath.Join(stage, "planning/wm-design-contracts", historical.revision, historical.intake, "bundle")
		copyBundleFixture(t, from, to)
	}
	if err := os.Mkdir(filepath.Join(stage, "release"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stage, "release/default-bundle.txt"), []byte("v9\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PPTXGENGO_RELEASE_ROOT", stage)
	for _, revision := range []string{"v5", "v7", "v8", "v9"} {
		t.Run(revision, func(t *testing.T) {
			root := projectDefaultFixture(t)
			projectCommandJSON(t, "init", "--project", root, "--bundle", revision)
			lockBefore, err := os.ReadFile(filepath.Join(root, "toolchain.lock.json"))
			if err != nil {
				t.Fatal(err)
			}
			projectCommandJSON(t, "check", "--project", root)
			projectCommandJSON(t, "build", "--project", root)
			projectCommandJSON(t, "export", "--project", root, "--mode", "client", "--out", filepath.Join(t.TempDir(), "client.zip"))
			lockAfter, err := os.ReadFile(filepath.Join(root, "toolchain.lock.json"))
			if err != nil || string(lockBefore) != string(lockAfter) {
				t.Fatal("omitted bundle changed existing lock")
			}
			if revision != "v5" && runProject([]string{"check", "--project", root, "--bundle", "v5"}) == nil {
				t.Fatal("explicit mismatched bundle accepted")
			}
			if revision == "v5" {
				archive := filepath.Join(t.TempDir(), "offline.zip")
				projectCommandJSON(t, "export", "--project", root, "--mode", "offline", "--out", archive)
				z, err := zip.OpenReader(archive)
				if err != nil {
					t.Fatal(err)
				}
				defer z.Close()
				relocated := t.TempDir()
				for _, entry := range z.File {
					path := filepath.Join(relocated, filepath.FromSlash(entry.Name))
					if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
						t.Fatal(err)
					}
					in, err := entry.Open()
					if err != nil {
						t.Fatal(err)
					}
					b, err := io.ReadAll(in)
					in.Close()
					if err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(path, b, 0600); err != nil {
						t.Fatal(err)
					}
				}
				t.Setenv("PPTXGENGO_RELEASE_ROOT", t.TempDir())
				projectCommandJSON(t, "check", "--project", relocated)
				projectCommandJSON(t, "build", "--project", relocated)
			}
		})
	}
	root := projectDefaultFixture(t)
	projectCommandJSON(t, "init", "--project", root)
	p, err := deckproject.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	lock, _, err := deckproject.ReadLock(p)
	if err != nil || lock.BundleRevision != "wmds-library.v9" {
		t.Fatalf("new staged project pin: %+v %v", lock, err)
	}
}
