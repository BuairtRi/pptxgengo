package deckproject

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/finishedslide"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func finishedDerivedProject(t *testing.T) *Project {
	t.Helper()
	p := reuseMediaProject(t)
	ids := []string{"sample-image", "z-middle", "a-final"}
	for i, id := range ids {
		n := 8 >> i
		img := image.NewRGBA(image.Rect(0, 0, n, n))
		for y := 0; y < n; y++ {
			for x := 0; x < n; x++ {
				img.Set(x, y, color.RGBA{R: 80, G: 120, B: 200, A: 255})
			}
		}
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			t.Fatal(err)
		}
		path := "assets/" + id + ".png"
		if err := os.WriteFile(filepath.Join(p.Root, path), buf.Bytes(), 0644); err != nil {
			t.Fatal(err)
		}
		a := Asset{Path: path, SHA256: digest(buf.Bytes()), Description: "Synthetic derived image fixture", Focus: &AssetFocus{X: 0.5, Y: 0.5}}
		if i > 0 {
			a.DerivedFrom = ids[i-1]
			a.DerivationReceipt = "assets/" + id + "-receipt.json"
			r := DerivationReceipt{Schema: "pptxgengo.asset-derivation.v1", SourceAsset: ids[i-1], SourceSHA256: p.Document.Assets[ids[i-1]].SHA256, ResultSHA256: a.SHA256, Operation: "crop", Parameters: map[string]any{"x": 0, "y": 0, "width": n, "height": n}}
			if err := os.WriteFile(filepath.Join(p.Root, a.DerivationReceipt), append([]byte(" \n"), canonical(r)...), 0644); err != nil {
				t.Fatal(err)
			}
		}
		p.Document.Assets[id] = a
	}
	def, err := finishedTemplate(bundle(t), p.Document.Slides[0].Template.ID)
	if err != nil {
		t.Fatal(err)
	}
	media, err := finishedMediaSlots(p.Document.Slides[0], def)
	if err != nil {
		t.Fatal(err)
	}
	for slot := range media {
		p.Document.Slides[0].Values["slots"].(map[string]any)[slot] = "project:a-final"
	}
	if err := os.WriteFile(p.SourcePath, canonical(p.Document), 0644); err != nil {
		t.Fatal(err)
	}
	p, err = Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestFinishedSlideDerivedAssetsIndependentInsertion(t *testing.T) {
	source := finishedDerivedProject(t)
	sourceRaw := mustRead(t, source.SourcePath)
	library, m := publishReuse(t, source, 1)
	_, deps, payload, err := readFinishedSource(library, m)
	if err != nil {
		t.Fatal(err)
	}
	if len(deps.Assets) != 3 {
		t.Fatalf("closure: %+v", deps.Assets)
	}
	if !bytes.Equal(sourceRaw, mustRead(t, source.SourcePath)) {
		t.Fatal("publication changed source")
	}
	for _, id := range []string{"z-middle", "a-final"} {
		if !bytes.Equal(payload[deps.Assets[id].DerivationReceipt], mustRead(t, filepath.Join(source.Root, source.Document.Assets[id].DerivationReceipt))) {
			t.Fatal("source receipt rewritten")
		}
	}
	var previous map[string]string
	var firstPath string
	for i := 0; i < 2; i++ {
		p := reuseProject(t)
		r, err := InsertFinishedSlide(p, FinishedSlideInsertOptions{Package: library, ID: "derived-copy", Rationale: "Independent synthetic image reuse", AllowDraft: true, Bundle: bundle(t), Engine: wmdesign.CandidateEngine})
		if err != nil {
			t.Fatal(err)
		}
		if len(r.Library.AssetRemaps) != 3 || len(r.Library.Derivations) != 2 {
			t.Fatal("incomplete lineage", r.Library)
		}
		p, err = Load(p.Root)
		if err != nil {
			t.Fatal(err)
		}
		for id, old := range deps.Assets {
			newID := r.Library.AssetRemaps[id]
			a := p.Document.Assets[newID]
			if previous != nil && previous[id] == newID {
				t.Fatal("shared identity")
			}
			if !bytes.Equal(mustRead(t, filepath.Join(p.Root, a.Path)), payload[old.Path]) || !reflect.DeepEqual(a.Focus, old.Focus) {
				t.Fatal("image or focus adapted")
			}
			if old.DerivedFrom == "" {
				continue
			}
			if a.DerivedFrom != r.Library.AssetRemaps[old.DerivedFrom] {
				t.Fatal("ancestry not remapped")
			}
			line := r.Library.Derivations[id]
			original := mustRead(t, filepath.Join(p.Root, line.OriginalReceiptPath))
			if !bytes.Equal(original, payload[old.DerivationReceipt]) || digest(original) != line.SourceReceiptSHA256 {
				t.Fatal("original receipt lost")
			}
			raw := mustRead(t, filepath.Join(p.Root, a.DerivationReceipt))
			if digest(raw) != line.RemappedReceiptSHA256 || a.DerivationReceipt != line.RemappedReceiptPath {
				t.Fatal("remapped receipt pin wrong")
			}
			var before, after DerivationReceipt
			if err := json.Unmarshal(original, &before); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(raw, &after); err != nil {
				t.Fatal(err)
			}
			before.SourceAsset = after.SourceAsset
			if !reflect.DeepEqual(before, after) {
				t.Fatal("receipt operation or parameters adapted")
			}
		}
		if _, err := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); err != nil {
			t.Fatal(err)
		}
		exported, err := Export(p, ExportOptions{Mode: "maintainer", Out: filepath.Join(t.TempDir(), "portable.zip")})
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range r.Library.Derivations {
			if exported.Files[d.OriginalReceiptPath] != d.SourceReceiptSHA256 || exported.Files[d.RemappedReceiptPath] != d.RemappedReceiptSHA256 {
				t.Fatal("portable export lost derivation", exported)
			}
		}
		current := filepath.Join(p.Root, p.Document.Assets[r.Library.AssetRemaps["a-final"]].Path)
		if i == 0 {
			firstPath = current
		} else {
			if err := os.WriteFile(firstPath, []byte("user edit in first copy"), 0644); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(mustRead(t, current), payload[deps.Assets["a-final"].Path]) {
				t.Fatal("copy mutation leaked")
			}
		}
		previous = r.Library.AssetRemaps
	}
	if !bytes.Equal(mustRead(t, filepath.Join(library, deps.Assets["a-final"].Path)), payload[deps.Assets["a-final"].Path]) {
		t.Fatal("library changed")
	}
}

func TestFinishedSlideDerivedAssetGraphRefusals(t *testing.T) {
	base := map[string]Asset{"root": {Path: "root.png", SHA256: strings.Repeat("a", 64)}, "leaf": {Path: "leaf.png", SHA256: strings.Repeat("b", 64), DerivedFrom: "root", DerivationReceipt: "receipt.json"}}
	for _, tc := range []struct {
		name   string
		change func(map[string]Asset)
	}{
		{"cycle", func(a map[string]Asset) {
			v := a["root"]
			v.DerivedFrom = "leaf"
			v.DerivationReceipt = "r.json"
			a["root"] = v
		}},
		{"missing-parent", func(a map[string]Asset) { delete(a, "root") }},
		{"unpaired", func(a map[string]Asset) { v := a["leaf"]; v.DerivationReceipt = ""; a["leaf"] = v }},
		{"ambiguous-location", func(a map[string]Asset) { v := a["leaf"]; v.RegistryID = "also-registry"; a["leaf"] = v }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := map[string]Asset{}
			for k, v := range base {
				a[k] = v
			}
			tc.change(a)
			if _, err := finishedAssetOrder(map[string]string{"image": "project:leaf"}, a); err == nil {
				t.Fatal("invalid graph accepted")
			}
		})
	}
	r := DerivationReceipt{Schema: "pptxgengo.asset-derivation.v1", SourceAsset: "root", SourceSHA256: base["root"].SHA256, ResultSHA256: base["leaf"].SHA256, Operation: "crop"}
	for _, tc := range []struct {
		name string
		raw  []byte
	}{
		{"wrong-source", bytes.Replace(canonical(r), []byte("root"), []byte("other"), 1)},
		{"wrong-result", bytes.Replace(canonical(r), []byte(base["leaf"].SHA256), []byte(strings.Repeat("c", 64)), 1)},
		{"empty", nil}, {"oversize", bytes.Repeat([]byte(" "), 1<<20+1)},
		{"unknown-field", append(canonical(r)[:len(canonical(r))-1], []byte(",\"unexpected\":true}")...)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := finishedDerivation(tc.raw, "root", base["root"].SHA256, base["leaf"].SHA256); err == nil {
				t.Fatal("invalid receipt accepted")
			}
		})
	}
	// Closed packages cannot smuggle an unrelated asset into the destination.
	deps := map[string]Asset{"unreferenced": base["root"]}
	if err := insertFinishedAssets(nil, deps, nil, nil, map[string]Asset{}, map[string][]byte{}, &LibraryLineage{}); err == nil {
		t.Fatal("unreferenced asset accepted")
	}
}

func TestFinishedSlideDerivedAssetsResealedRefusals(t *testing.T) {
	source := finishedDerivedProject(t)
	root, manifest := publishReuse(t, source, 1)
	_, deps, payload, err := readFinishedSource(root, manifest)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"wrong-parent-hash", "wrong-parent-id", "cycle", "missing-parent", "unpaired", "wrong-role", "unrelated"} {
		t.Run(name, func(t *testing.T) {
			files := map[string][]byte{}
			for k, v := range payload {
				files[k] = append([]byte(nil), v...)
			}
			m := manifest
			m.Files = append([]finishedslide.File(nil), manifest.Files...)
			altered := deps
			altered.Assets = map[string]Asset{}
			for k, v := range deps.Assets {
				altered.Assets[k] = v
			}
			switch name {
			case "wrong-parent-hash", "wrong-parent-id":
				path := deps.Assets["a-final"].DerivationReceipt
				var r DerivationReceipt
				if err := json.Unmarshal(files[path], &r); err != nil {
					t.Fatal(err)
				}
				if name == "wrong-parent-hash" {
					r.SourceSHA256 = strings.Repeat("c", 64)
				} else {
					r.SourceAsset = "sample-image"
				}
				files[path] = canonical(r)
			case "cycle":
				a := altered.Assets["sample-image"]
				a.DerivedFrom = "a-final"
				a.DerivationReceipt = altered.Assets["a-final"].DerivationReceipt
				altered.Assets["sample-image"] = a
			case "missing-parent":
				delete(altered.Assets, "z-middle")
			case "unpaired":
				a := altered.Assets["a-final"]
				a.DerivationReceipt = ""
				altered.Assets["a-final"] = a
			case "wrong-role":
				for i := range m.Files {
					if m.Files[i].Path == deps.Assets["a-final"].DerivationReceipt {
						m.Files[i].Role = "evidence"
					}
				}
			case "unrelated":
				altered.Assets["unrelated"] = deps.Assets["sample-image"]
			}
			files["dependencies.json"] = canonical(altered)
			tampered := filepath.Join(t.TempDir(), "resealed")
			if _, err := finishedslide.Create(tampered, m, files); err != nil {
				t.Fatal(err)
			}
			p := reuseProject(t)
			before := mustRead(t, p.SourcePath)
			beforeLog := mustRead(t, filepath.Join(p.Root, "composition-log.yaml"))
			if _, err := InsertFinishedSlide(p, FinishedSlideInsertOptions{Package: tampered, ID: "rejected-copy", Rationale: "Reject false asset ancestry", AllowDraft: true, Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); err == nil {
				t.Fatal("false graph accepted")
			}
			if !bytes.Equal(before, mustRead(t, p.SourcePath)) || !bytes.Equal(beforeLog, mustRead(t, filepath.Join(p.Root, "composition-log.yaml"))) {
				t.Fatal("refusal mutated source")
			}
		})
	}
}
