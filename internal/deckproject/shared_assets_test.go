package deckproject

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestSharedAssetValuesOnlyResolvesDeclaredMediaFields(t *testing.T) {
	def := wmdesign.LibraryTemplate{RawSlide: json.RawMessage(`{"body":[{"type":"imageframe","photo":"photo-team"},{"type":"text","text":"client-logo"},{"type":"table","rows":[{"photo":"client-logo"}]},{"type":"card","media":{"src":"photo-team"}}]}`), Slots: []wmdesign.LibrarySlot{
		{Name: "image", Kind: "string", SourcePointer: "/body/0/photo"},
		{Name: "text", Kind: "string", SourcePointer: "/body/1/text"},
		{Name: "cell", Kind: "string", SourcePointer: "/body/2/rows/0/photo"},
		{Name: "card", Kind: "string", SourcePointer: "/body/3/media/src"},
	}}
	values := map[string]any{"slots": map[string]any{"image": "client-logo", "text": "client-logo", "cell": "client-logo", "card": "project:alias"}}
	before := canonical(values)
	keys := map[string]string{"client-logo": "project:client-logo", "alias": "photo-team"}
	registry := map[string]wmdesign.PrimitiveAssetReference{"photo-team": {Key: "photo-team"}}
	resolved, err := sharedAssetValues(values, def, keys, registry)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, canonical(values)) {
		t.Fatal("authored values mutated")
	}
	slots := resolved["slots"].(map[string]any)
	if slots["image"] != "project:client-logo" || slots["card"] != "photo-team" || slots["text"] != "client-logo" || slots["cell"] != "client-logo" {
		t.Fatalf("unbounded asset resolution: %+v", slots)
	}
	values["slots"].(map[string]any)["image"] = "photo-team"
	keys["photo-team"] = "project:photo-team"
	resolved, err = sharedAssetValues(values, def, keys, registry)
	if err != nil || resolved["slots"].(map[string]any)["image"] != "photo-team" {
		t.Fatal("existing registry key was shadowed")
	}
	values["slots"].(map[string]any)["image"] = "project:photo-team"
	resolved, err = sharedAssetValues(values, def, keys, registry)
	if err != nil || resolved["slots"].(map[string]any)["image"] != "project:photo-team" {
		t.Fatal("explicit project asset was not retained")
	}
	values["slots"].(map[string]any)["image"] = "undeclared-image"
	resolved, err = sharedAssetValues(values, def, keys, registry)
	if err != nil || resolved["slots"].(map[string]any)["image"] != "undeclared-image" {
		t.Fatal("unknown image should remain subject to renderer registry validation")
	}
}

func TestSharedSlideCompileResolvesProjectImageWithoutChangingSource(t *testing.T) {
	p := example(t)
	catalog, err := wmdesign.LibraryCatalog(bundle(t), "")
	if err != nil {
		t.Fatal(err)
	}
	examples, err := wmdesign.LibraryReference(bundle(t), "", "covers", 2026)
	if err != nil {
		t.Fatal(err)
	}
	defs := map[string]wmdesign.LibraryTemplate{}
	for _, def := range catalog {
		defs[def.Key] = def
	}
	for _, example := range examples.Slides {
		def := defs[example.Template]
		var source map[string]any
		if err := json.Unmarshal(def.RawSlide, &source); err != nil {
			t.Fatal(err)
		}
		for _, slot := range def.Slots {
			parts, err := contentPointerParts(slot.SourcePointer)
			if err != nil || !sharedImagePointer(source, parts) {
				continue
			}
			var values map[string]any
			if err := json.Unmarshal(example.Values, &values); err != nil {
				t.Fatal(err)
			}
			assetID := "client-image"
			p.Document.Assets = map[string]Asset{assetID: {Path: "assets/originals/sample.png"}}
			values["slots"].(map[string]any)[slot.Name] = assetID
			p.Document.Slides = []Slide{{ID: "cover", Template: Reference{Scope: "shared", ID: example.Template}, ContentKind: "supplied_content", Values: values}}
			before := canonical(p.Document)
			compiled, err := Compile(p, bundle(t), wmdesign.CandidateEngine)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(before, canonical(p.Document)) || !bytes.Contains(canonical(compiled.Document), []byte("project:client-image")) || len(compiled.Assets["project:client-image"].Data) == 0 {
				t.Fatal("shared image registry binding/source preservation failed")
			}
			pptx, _, err := wmdesign.BuildWithEngineAndAssets(bundle(t), "", compiled.Document, wmdesign.CandidateEngine, compiled.Assets)
			if err != nil || len(pptx) == 0 {
				t.Fatalf("resolved shared image cannot build: %v", err)
			}
			return
		}
	}
	t.Fatal("expected a genuine cover image slot")
}
