package deckproject

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func placeholderScaffoldProject(t *testing.T, key string, options ...ScaffoldOptions) (*Project, TemplateScaffold) {
	t.Helper()
	b := curveCompositionBundle(t)
	selected := ScaffoldOptions{PlaceholderMedia: true}
	if len(options) > 0 {
		selected = options[0]
	}
	scaffold, e := ScaffoldTemplateWithOptions(b, key, wmdesign.CandidateEngine, "Explicit schematic private-media example; no original artwork claim", 2026, selected)
	if e != nil {
		t.Fatal(e)
	}
	p := example(t)
	p.Document.Assets = scaffold.Assets
	p.Document.Context = nil
	p.Document.LocalTemplates = map[string]LocalTemplate{"portable": scaffold.Template}
	p.Document.Slides = []Slide{{ID: "portable-slide", ContentKind: "synthetic_example", Template: Reference{Scope: "local", ID: "portable"}, Values: scaffold.SyntheticSourceValues}}
	for relative, payload := range scaffold.AssetPayloads {
		path, e := SafePath(p.Root, relative)
		if e != nil {
			t.Fatal(e)
		}
		if e = os.MkdirAll(filepath.Dir(path), 0700); e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(path, payload, 0600); e != nil {
			t.Fatal(e)
		}
	}
	if e = os.WriteFile(p.SourcePath, canonical(p.Document), 0600); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = Pin(p, b, wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	return p, scaffold
}
func TestScaffoldPlaceholderMediaExplicitPortableAndGenuineIcons(t *testing.T) {
	t.Setenv("WMDS_BRANDING_ROOT", t.TempDir())
	for _, key := range []string{"team-curve/build-together", "runbook/cutover-timeline", "architecture/layers-icons"} {
		t.Run(key, func(t *testing.T) {
			p, scaffold := placeholderScaffoldProject(t, key)
			if key != "architecture/layers-icons" && len(scaffold.PlaceholderMedia) == 0 {
				t.Fatal("rendered placeholder receipt missing")
			}
			for _, receipt := range scaffold.PlaceholderMedia {
				if strings.HasPrefix(receipt.RegistryID, "icon/") || strings.HasPrefix(receipt.RegistryID, "arrow-") || scaffold.Assets[receipt.AssetID].RegistryID != "" || !strings.Contains(receipt.Description, "not original artwork") {
					t.Fatal("original identity claimed or genuine glyph replaced", receipt)
				}
				if digest(scaffold.AssetPayloads[receipt.Path]) != receipt.SHA256 {
					t.Fatal("payload identity mismatch")
				}
			}
			compile, e := Compile(p, curveCompositionBundle(t), wmdesign.CandidateEngine)
			if e != nil {
				t.Fatal(e)
			}
			if _, _, e = wmdesign.BuildWithEngineAndAssets(curveCompositionBundle(t), "", compile.Document, wmdesign.CandidateEngine, compile.Assets); e != nil {
				t.Fatal("portable build", e)
			}
			// Copy only authored project sources/payloads to another owned root. The
			// rendering policy must survive without any ambient placeholder override.
			portable := t.TempDir()
			if e = copyDemoProject(p.Root, portable); e != nil {
				t.Fatal(e)
			}
			relocated, e := Load(filepath.Join(portable, filepath.Base(p.SourcePath)))
			if e != nil {
				t.Fatal(e)
			}
			compile, e = Compile(relocated, curveCompositionBundle(t), wmdesign.CandidateEngine)
			if e != nil {
				t.Fatal(e)
			}
			if _, _, e = wmdesign.BuildWithEngineAndAssets(curveCompositionBundle(t), "", compile.Document, wmdesign.CandidateEngine, compile.Assets); e != nil {
				t.Fatal("relocated build", e)
			}
		})
	}
	if _, e := ScaffoldTemplate(curveCompositionBundle(t), "team-curve/build-together", wmdesign.CandidateEngine, "Normal strict source", 2026); e == nil {
		t.Fatal("normal scaffold silently replaced private media")
	}
}
func TestScaffoldPlaceholderAliasRefusalsAndTampering(t *testing.T) {
	t.Setenv("WMDS_BRANDING_ROOT", t.TempDir())
	p, scaffold := placeholderScaffoldProject(t, "team-curve/build-together")
	if len(scaffold.Assets) == 0 {
		t.Fatal("missing test asset")
	}
	id := ""
	for key := range scaffold.Assets {
		id = key
		break
	}
	original := scaffold.Assets[id]
	for _, scenario := range []string{"unknown", "icon", "arrow", "original-hash", "missing-hash", "description", "registry-claim", "unsafe-path", "duplicate", "competing-original", "symlink"} {
		t.Run(scenario, func(t *testing.T) {
			document := p.Document
			document.Assets = map[string]Asset{}
			for key, a := range scaffold.Assets {
				document.Assets[key] = a
			}
			a := original
			switch scenario {
			case "unknown":
				a.PlaceholderFor = "not-registered"
			case "icon":
				a.PlaceholderFor = "icon/browser-gear/white"
			case "arrow":
				a.PlaceholderFor = "arrow-straight"
			case "original-hash":
				for _, ref := range wmdesign.PrimitiveAssetCatalog() {
					if ref.Key == a.PlaceholderFor {
						a.SHA256 = ref.SHA256
						a.Path = "assets/objects/sha256/" + a.SHA256
					}
				}
			case "missing-hash":
				a.SHA256 = ""
			case "description":
				a.Description = "Original approved artwork"
			case "registry-claim":
				a.RegistryID = a.PlaceholderFor
			case "unsafe-path":
				a.Path = "../outside.png"
			case "duplicate":
				document.Assets["another-placeholder"] = a
			case "competing-original":
				document.Assets["declared-original"] = Asset{RegistryID: a.PlaceholderFor}
			case "symlink":
				path := filepath.Join(p.Root, a.Path)
				old, e := os.ReadFile(path)
				if e != nil {
					t.Fatal(e)
				}
				target := filepath.Join(t.TempDir(), "outside")
				os.WriteFile(target, old, 0600)
				if e = os.Remove(path); e != nil {
					t.Fatal(e)
				}
				if e = os.Symlink(target, path); e != nil {
					t.Fatal(e)
				}
				t.Cleanup(func() { os.Remove(path); os.WriteFile(path, old, 0600) })
			}
			document.Assets[id] = a
			source := canonical(document)
			if e := os.WriteFile(p.SourcePath, source, 0600); e != nil {
				t.Fatal(e)
			}
			if _, e := Load(p.SourcePath); e == nil {
				t.Fatal("unsafe placeholder alias accepted", scenario)
			}
		})
	}
	if e := os.WriteFile(p.SourcePath, canonical(p.Document), 0600); e != nil {
		t.Fatal(e)
	}
	path := filepath.Join(p.Root, original.Path)
	payload, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(path, append(bytes.Clone(payload), []byte("tamper")...), 0600); e != nil {
		t.Fatal(e)
	}
	loaded, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = Compile(loaded, curveCompositionBundle(t), wmdesign.CandidateEngine); e == nil || !strings.Contains(e.Error(), "hash mismatch") {
		t.Fatal("tampered placeholder payload accepted", e)
	}
}

func copyDemoProject(source, destination string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		relative, e := filepath.Rel(source, path)
		if e != nil {
			return e
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0700)
		}
		raw, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		return os.WriteFile(target, raw, 0600)
	})
}
