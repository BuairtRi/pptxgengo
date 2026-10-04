package deckproject

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestCalibrationLocationsPreserveLogicalLockKey(t *testing.T) {
	root := t.TempDir()
	fontRoot := filepath.Join(root, "pinned", "fonts")
	legacy := filepath.Join(root, filepath.FromSlash(calibrationLockKey))
	current := filepath.Join(root, "pinned", "typography", "calibration.json")
	for _, path := range []string{legacy, current} {
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(legacy, []byte("old export"), 0600); err != nil {
		t.Fatal(err)
	}
	legacy, err := filepath.EvalSymlinks(legacy)
	if err != nil {
		t.Fatal(err)
	}
	got, err := runtimeFilePath(filepath.Dir(fontRoot), calibrationLockKey)
	if err != nil || got != legacy {
		t.Fatalf("legacy offline path: %q %v", got, err)
	}
	if err := os.WriteFile(current, []byte("current bundled calibration"), 0600); err != nil {
		t.Fatal(err)
	}
	current, err = filepath.EvalSymlinks(current)
	if err != nil {
		t.Fatal(err)
	}
	got, err = runtimeFilePath(filepath.Dir(fontRoot), calibrationLockKey)
	if err != nil || got != current {
		t.Fatalf("current physical path: %q %v", got, err)
	}
	if _, err := runtimeFilePath(filepath.Dir(fontRoot), "../outside"); err == nil {
		t.Fatal("unsafe noncalibration runtime path accepted")
	}
}

func TestPinnedBundleExcludesOnlyAuxiliaryResources(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"bundle.json": "source", "source/templates/library/core.json": "contract",
		"fonts/font.ttf": "font", "catalog/preview.png": "preview",
		"library.sqlite": "discovery", "typography/calibration.json": "separate runtime pin",
	}
	for rel, content := range files {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	pinned, err := filteredTreeHashes(root, bundleAuxiliaryPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(pinned) != 3 {
		t.Fatalf("unexpected source closure: %v", pinned)
	}
	all, err := treeHashes(root)
	if err != nil || len(all) != len(files) {
		t.Fatalf("generic tree hashing changed: %v %v", all, err)
	}
	if err := os.WriteFile(filepath.Join(root, "unexpected.json"), []byte("drift"), 0600); err != nil {
		t.Fatal(err)
	}
	next, err := filteredTreeHashes(root, bundleAuxiliaryPath)
	if err != nil || reflect.DeepEqual(pinned, next) {
		t.Fatal("unrecognized added source file escaped the pin")
	}
}

// The qualified delivery locks keep their original compiler pin. This verifies
// source relocation and all independent resource pins without rewriting locks,
// invoking a different compiler as if qualified, or touching delivery outputs.
func TestFinalRelocatedProjectsCompileAndPreserveResourcePins(t *testing.T) {
	if testing.Short() {
		t.Skip("three final relocated project builds; run make test-integration")
	}
	for name, count := range map[string]int{"software-modernization": 83, "patterson-eaglesoft": 39, "dentalxchange": 75} {
		t.Run(name, func(t *testing.T) {
			src := filepath.Join("..", "..", "samples", "final", name)
			dst := t.TempDir()
			err := filepath.WalkDir(src, func(path string, entry os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				rel, err := filepath.Rel(src, path)
				if err != nil {
					return err
				}
				target := filepath.Join(dst, rel)
				if entry.IsDir() {
					return os.MkdirAll(target, 0700)
				}
				if ext := strings.ToLower(filepath.Ext(path)); ext == ".zip" || ext == ".pptx" || ext == ".pdf" {
					return nil
				}
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				return os.WriteFile(target, data, 0600)
			})
			if err != nil {
				t.Fatal(err)
			}
			p, err := Load(dst)
			if err != nil {
				t.Fatal(err)
			}
			lock, before, err := ReadLock(p)
			if err != nil {
				t.Fatal(err)
			}
			actual, err := makeLock(bundle(t), lock.Engine)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(lock.BundleFiles, actual.BundleFiles) || lock.BundleSHA256 != actual.BundleSHA256 || !reflect.DeepEqual(lock.RuntimeFiles, actual.RuntimeFiles) {
				t.Fatal("qualified source or runtime resource pins changed")
			}
			compiled, err := Compile(p, bundle(t), lock.Engine)
			if err != nil || len(compiled.Document.Slides) != count {
				t.Fatalf("relocated compilation: %d slides, %v", len(compiled.Document.Slides), err)
			}
			_, after, err := ReadLock(p)
			if err != nil || string(before) != string(after) {
				t.Fatal("qualified compiler lock was changed")
			}
			if lock.RuntimeFiles[calibrationLockKey] != wmdesign.CandidateCalibrationSHA {
				t.Fatal("qualified calibration logical key/hash changed")
			}
		})
	}
}
