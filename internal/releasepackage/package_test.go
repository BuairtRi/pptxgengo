package releasepackage

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func packageFixture(t *testing.T) Options {
	t.Helper()
	root := t.TempDir()
	release := filepath.Join(root, "release")
	bins := filepath.Join(root, "binaries")
	skill := filepath.Join(root, "skill")
	hashes := map[string]string{}
	for rel, value := range map[string]string{"VERSION": "old\n", "library/wm-design-system/v11/library.sqlite": "database fixture", "branding/West Monroe Photos/People/test.jpg": "original photo", "branding/logos/test.svg": "logo", "skills/west-monroe-presentations/SKILL.md": "old skill", "bin/pptxgengo": "old binary"} {
		path := filepath.Join(release, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(value), 0600); err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256([]byte(value))
		hashes[rel] = hex.EncodeToString(sum[:])
	}
	raw, _ := json.Marshal(map[string]any{"schema": "pptxgengo.local-release-manifest.v1", "files_sha256": hashes, "selected_bundle": "v11"})
	if err := os.WriteFile(filepath.Join(release, "release-manifest.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(bins, 0700); err != nil {
		t.Fatal(err)
	}
	for _, tool := range []string{"pptxgengo", "pptxdesign", "wmdsdocs"} {
		writeTestPE(t, filepath.Join(bins, tool+".exe"), 0x8664)
	}
	if err := os.Mkdir(skill, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skill, "SKILL.md"), []byte("latest skill"), 0600); err != nil {
		t.Fatal(err)
	}
	return Options{Release: release, Binaries: bins, Skill: skill, Out: filepath.Join(root, "tester.zip"), Architecture: "amd64", Version: "0.1.0-local.21-windows-preview.1"}
}
func writeTestPE(t *testing.T, path string, machine uint16) {
	t.Helper()
	data := make([]byte, 152)
	copy(data, []byte("MZ"))
	binary.LittleEndian.PutUint32(data[60:], 128)
	copy(data[128:], []byte{'P', 'E', 0, 0})
	binary.LittleEndian.PutUint16(data[132:], machine)
	binary.LittleEndian.PutUint16(data[150:], 2)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}
func TestPortablePackageReplacesSkillAndExecutablesWithMatchingManifest(t *testing.T) {
	for _, light := range []bool{false, true} {
		t.Run(map[bool]string{false: "full", true: "without-photos"}[light], func(t *testing.T) {
			o := packageFixture(t)
			o.WithoutPhotos = light
			report, err := Create(o)
			if err != nil {
				t.Fatal(err)
			}
			archive, err := zip.OpenReader(o.Out)
			if err != nil {
				t.Fatal(err)
			}
			defer archive.Close()
			files := map[string][]byte{}
			for _, file := range archive.File {
				r, err := file.Open()
				if err != nil {
					t.Fatal(err)
				}
				data, err := io.ReadAll(r)
				r.Close()
				if err != nil {
					t.Fatal(err)
				}
				files[strings.TrimPrefix(file.Name, "pptxgengo-windows-amd64/")] = data
			}
			if string(files["skills/west-monroe-presentations/SKILL.md"]) != "latest skill" || files["bin/pptxgengo"] != nil || files["bin/pptxgengo.exe"] == nil {
				t.Fatal("wrong replacement files")
			}
			if _, present := files["branding/West Monroe Photos/People/test.jpg"]; present == light {
				t.Fatal("photo omission mismatch")
			}
			var manifest struct {
				Files     map[string]string `json:"files_sha256"`
				Count     int               `json:"file_count"`
				Qualified bool              `json:"windows_runtime_qualified"`
			}
			if err := json.Unmarshal(files["release-manifest.json"], &manifest); err != nil {
				t.Fatal(err)
			}
			if report.Files != manifest.Count || report.WindowsRuntimeQualified || manifest.Qualified {
				t.Fatal(report, manifest)
			}
			for name, expected := range manifest.Files {
				sum := sha256.Sum256(files[name])
				if hex.EncodeToString(sum[:]) != expected {
					t.Fatal("hash mismatch", name)
				}
			}
			before, _ := os.ReadFile(filepath.Join(o.Release, "skills/west-monroe-presentations/SKILL.md"))
			if string(before) != "old skill" {
				t.Fatal("installed release mutated")
			}
			if _, err := Create(o); err == nil {
				t.Fatal("existing output overwritten")
			}
		})
	}
}
func TestPortablePackageRejectsDriftAndWrongArchitecture(t *testing.T) {
	for _, scenario := range []string{"drift", "architecture"} {
		t.Run(scenario, func(t *testing.T) {
			o := packageFixture(t)
			if scenario == "drift" {
				if err := os.WriteFile(filepath.Join(o.Release, "library/wm-design-system/v11/library.sqlite"), []byte("changed"), 0600); err != nil {
					t.Fatal(err)
				}
			} else {
				writeTestPE(t, filepath.Join(o.Binaries, "pptxdesign.exe"), 0xaa64)
			}
			if _, err := Create(o); err == nil {
				t.Fatal("invalid package accepted")
			}
			if _, err := os.Stat(o.Out); !os.IsNotExist(err) {
				t.Fatal("failed package was retained")
			}
		})
	}
}
func TestWindowsPackagePaths(t *testing.T) {
	for _, path := range []string{"../x", "/x", `C:\foo`, `a\b`, "NUL.png", "a/COM1.txt", "a/x.", "a/x ", "a//b", "a:x", "a?b"} {
		if err := validateArchivePath(path); err == nil {
			t.Fatal("unsafe path accepted", path)
		}
	}
	if err := validateArchivePath("branding/Photos/Unicode-é & spaces.png"); err != nil {
		t.Fatal(err)
	}
}
