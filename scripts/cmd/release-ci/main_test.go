package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"path/filepath"
	"runtime"
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
			// Check serialized Unix mode on every OS. Windows filesystems do
			// not expose Unix executable bits on extracted regular files.
			if target == "windows-amd64" {
				z, e := zip.OpenReader(a)
				if e != nil {
					t.Fatal(e)
				}
				found := false
				for _, f := range z.File {
					if f.Name == "bin/tool" {
						found = true
						if f.Mode().Perm() != 0755 {
							t.Fatal("ZIP executable mode lost")
						}
					}
				}
				z.Close()
				if !found {
					t.Fatal("ZIP executable missing")
				}
			} else {
				f, e := os.Open(a)
				if e != nil {
					t.Fatal(e)
				}
				gz, e := gzip.NewReader(f)
				if e != nil {
					t.Fatal(e)
				}
				tr, found := tar.NewReader(gz), false
				for {
					h, e := tr.Next()
					if e != nil {
						break
					}
					if h.Name == "bin/tool" {
						found = true
						if os.FileMode(h.Mode).Perm() != 0755 {
							t.Fatal("TAR executable mode lost")
						}
					}
				}
				gz.Close()
				f.Close()
				if !found {
					t.Fatal("TAR executable missing")
				}
			}
			out := filepath.Join(root, target)
			if err := extract(a, out); err != nil {
				t.Fatal(err)
			}
			info, err := os.Stat(filepath.Join(out, "bin/tool"))
			if err != nil || !info.Mode().IsRegular() || (runtime.GOOS != "windows" && info.Mode().Perm() != 0755) {
				t.Fatal("executable mode was lost", err)
			}
			data, e := os.ReadFile(filepath.Join(out, "bin/tool"))
			if e != nil || string(data) != "payload" {
				t.Fatal("extracted executable payload changed", e)
			}
		})
	}
}
func manifestFixture(t *testing.T) (string, Manifest) {
	t.Helper()
	root := t.TempDir()
	m := Manifest{Schema: "pptxgengo.release-manifest/v1", Project: "riscott/pptxgengo", Version: "v4.1.0-rc.1", Commit: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", PackageKind: "cli-only", Targets: targets, Files: map[string]string{}}
	for _, target := range targets {
		for _, name := range []string{archiveName(target, m.Version), archiveName(target, m.Version) + ".sbom.json", archiveName(target, m.Version) + ".vulnerabilities.json", target + ".build-evidence.json", "security-policy.json"} {
			path := filepath.Join(root, name)
			os.WriteFile(path, []byte("original"), 0644)
			hash, _ := digest(path)
			m.Files[name] = hash
		}
		if target[:6] == "darwin" {
			name := target + ".notarization.json"
			path := filepath.Join(root, name)
			os.WriteFile(path, []byte("original"), 0644)
			hash, _ := digest(path)
			m.Files[name] = hash
		}
	}
	if err := writeJSON(filepath.Join(root, "manifest.json"), m); err != nil {
		t.Fatal(err)
	}
	return root, m
}
func TestManifestTamperingIsRejected(t *testing.T) {
	root, m := manifestFixture(t)
	if err := verify(root, m.Version, m.Commit); err != nil {
		t.Fatal("valid manifest rejected", err)
	}
	os.WriteFile(filepath.Join(root, archiveName(targets[0], m.Version)), []byte("changed"), 0644)
	if err := verify(root, m.Version, m.Commit); err == nil {
		t.Fatal("modified artifact was accepted")
	}
}

func TestManifestRefusesUnattestedFinalArtifacts(t *testing.T) {
	root, m := manifestFixture(t)
	if e := os.WriteFile(filepath.Join(root, "unlisted-download.txt"), []byte("not covered by the manifest"), 0600); e != nil {
		t.Fatal(e)
	}
	if e := verify(root, m.Version, m.Commit); e == nil {
		t.Fatal("unattested file could be published")
	}
}
