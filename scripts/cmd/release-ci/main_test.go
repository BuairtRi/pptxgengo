package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractionRejectsTraversalAndSymlinks(t *testing.T) {
	for _, tc := range []struct {
		name, member string
		mode         os.FileMode
	}{{"traversal", "../escaped", 0644}, {"absolute", "/escaped", 0644}, {"symlink", "bin/tool", os.ModeSymlink | 0777}} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			p := filepath.Join(root, "input.zip")
			f, err := os.Create(p)
			if err != nil {
				t.Fatal(err)
			}
			z := zip.NewWriter(f)
			h := &zip.FileHeader{Name: tc.member}
			h.SetMode(tc.mode)
			w, err := z.CreateHeader(h)
			if err != nil {
				t.Fatal(err)
			}
			w.Write([]byte("malicious"))
			z.Close()
			f.Close()
			if err := extract(p, filepath.Join(root, "out")); err == nil {
				t.Fatal("unsafe archive was accepted")
			}
			if _, err := os.Stat(filepath.Join(root, "escaped")); !os.IsNotExist(err) {
				t.Fatal("archive escaped destination")
			}
		})
	}
}
func TestArchivesReproducibleAndPreserveExecutableMode(t *testing.T) {
	root := t.TempDir()
	for _, target := range []string{"linux-amd64", "windows-amd64"} {
		t.Run(target, func(t *testing.T) {
			ext := ".tar.gz"
			if target == "windows-amd64" {
				ext = ".zip"
			}
			a, b := filepath.Join(root, target+"-a"+ext), filepath.Join(root, target+"-b"+ext)
			files := map[string]Input{"bin/tool": {Data: []byte("payload")}, "VERSION": {Data: []byte("v4.1.0-rc.1\n")}}
			if err := writeArchive(files, target, a); err != nil {
				t.Fatal(err)
			}
			if err := writeArchive(files, target, b); err != nil {
				t.Fatal(err)
			}
			da, _ := digest(a)
			db, _ := digest(b)
			if da != db {
				t.Fatal("identical payload produced different archives")
			}
			out := filepath.Join(root, target)
			if err := extract(a, out); err != nil {
				t.Fatal(err)
			}
			info, err := os.Stat(filepath.Join(out, "bin/tool"))
			if err != nil || info.Mode().Perm() != 0755 {
				t.Fatal("executable mode was lost", err)
			}
		})
	}
}
func TestManifestTamperingIsRejected(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "artifact.zip")
	os.WriteFile(path, []byte("original"), 0644)
	hash, _ := digest(path)
	m := Manifest{Schema: "pptxgengo.release-manifest/v1", Project: "riscott/pptxgengo", Version: "v4.1.0-rc.1", Commit: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", PackageKind: "cli-only", Targets: targets, Files: map[string]string{"artifact.zip": hash}}
	if err := writeJSON(filepath.Join(root, "manifest.json"), m); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(path, []byte("changed"), 0644)
	if err := verify(root, m.Version, m.Commit); err == nil {
		t.Fatal("modified artifact was accepted")
	}
}
