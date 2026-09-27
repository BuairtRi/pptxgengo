package library

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFingerprintSemanticVariantsAndDifferences(t *testing.T) {
	root := t.TempDir()
	s, err := NewStore(root)
	if err != nil {
		t.Fatal(err)
	}
	asset := []byte("identical artwork")
	put(t, filepath.Join(root, "a.png"), asset)
	put(t, filepath.Join(root, "b.png"), asset)
	base := map[string]any{
		"id": "slide-a", "notes": "source A", "width_pt": 960, "height_pt": 540,
		"canvas": []any{
			map[string]any{"id": "node-a", "kind": "text", "bounds": map[string]any{"x": 10, "y": 20, "width": 200, "height": 50}, "text": "Source copy A", "font_face": "Arial", "font_size_pt": 20, "foreground": "#111111", "allow_overlap": []any{"image-a"}},
			map[string]any{"id": "image-a", "kind": "image", "bounds": map[string]any{"x": 300, "y": 20, "width": 50, "height": 50}, "asset_path": "a.png", "asset_sha256": hashBytes(asset), "fallback_asset_path": "a.png", "fallback_asset_sha256": hashBytes(asset)},
		},
		"connections": []any{map[string]any{"id": "edge-a", "from": "node-a", "to": "image-a", "relationship": "flow", "preferred_direction": "right"}},
		"accents":     []any{map[string]any{"id": "accent-a", "target": "node-a", "mode": "underline"}},
	}
	makeContract := func(name string, slide map[string]any, token string) Contract {
		t.Helper()
		spec, _ := json.Marshal(map[string]any{"slides": []any{slide}})
		path := name + ".json"
		put(t, filepath.Join(root, path), spec)
		return Contract{ID: name, Composition: Composition{SpecPath: path, SlideID: slide["id"].(string), Slots: []Slot{{Name: name + "/copy", Role: "claim", Pointer: "/canvas/0/text", ValueType: "string", Required: true, MaxChars: 80}}}, StyleVariants: []StyleVariant{{ID: "source", SemanticProfile: "brand", Tokens: map[string]string{"brand.ink": token}}}, Transforms: Transforms{Translation: "tested", Resize: "unsupported", Rotation: "unsupported"}}
	}
	clone := func(v map[string]any) map[string]any {
		b, _ := json.Marshal(v)
		var out map[string]any
		json.Unmarshal(b, &out)
		return out
	}
	c1 := makeContract("first", clone(base), "#111111")
	other := clone(base)
	other["id"] = "slide-b"
	other["notes"] = "source B"
	canvas := other["canvas"].([]any)
	canvas[0].(map[string]any)["id"] = "node-b"
	canvas[0].(map[string]any)["text"] = "Different slot copy"
	canvas[0].(map[string]any)["foreground"] = "#222222"
	canvas[0].(map[string]any)["allow_overlap"] = []any{"image-b"}
	canvas[1].(map[string]any)["id"] = "image-b"
	canvas[1].(map[string]any)["asset_path"] = "b.png"
	canvas[1].(map[string]any)["fallback_asset_path"] = "b.png"
	connection := other["connections"].([]any)[0].(map[string]any)
	connection["id"] = "edge-b"
	connection["from"] = "node-b"
	connection["to"] = "image-b"
	other["accents"].([]any)[0].(map[string]any)["id"] = "accent-b"
	other["accents"].([]any)[0].(map[string]any)["target"] = "node-b"
	c2 := makeContract("second", other, "#222222")
	fp1, err := s.Fingerprint(c1)
	if err != nil {
		t.Fatal(err)
	}
	fp2, err := s.Fingerprint(c2)
	if err != nil {
		t.Fatal(err)
	}
	if fp1 != fp2 {
		t.Fatal("renamed IDs, source notes, slot copy, verified asset path and declared semantic color should deduplicate")
	}
	for name, mutate := range map[string]func(map[string]any){
		"geometry": func(v map[string]any) {
			v["canvas"].([]any)[0].(map[string]any)["bounds"].(map[string]any)["width"] = float64(201)
		},
		"shape count": func(v map[string]any) { v["canvas"] = v["canvas"].([]any)[:1] },
		"routing":     func(v map[string]any) { v["connections"].([]any)[0].(map[string]any)["preferred_direction"] = "left" },
		"target":      func(v map[string]any) { v["accents"].([]any)[0].(map[string]any)["target"] = "image-a" },
		"font":        func(v map[string]any) { v["canvas"].([]any)[0].(map[string]any)["font_face"] = "Aptos" },
		"asset": func(v map[string]any) {
			v["canvas"].([]any)[1].(map[string]any)["asset_sha256"] = strings.Repeat("a", 64)
		},
	} {
		t.Run(name, func(t *testing.T) {
			changed := clone(base)
			mutate(changed)
			c := makeContract(strings.ReplaceAll(name, " ", "-"), changed, "#111111")
			fp, err := s.Fingerprint(c)
			if name == "asset" {
				if err == nil {
					t.Fatal("unverified asset hash accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if fp == fp1 {
				t.Fatal("distinct composition deduplicated")
			}
		})
	}
	slotDifference := c1
	slotDifference.Composition.Slots = append([]Slot(nil), c1.Composition.Slots...)
	slotDifference.Composition.Slots[0].MaxChars = 40
	fp, err := s.Fingerprint(slotDifference)
	if err != nil {
		t.Fatal(err)
	}
	if fp == fp1 {
		t.Fatal("different slot capacity deduplicated")
	}
	arrayA := c1
	arrayA.Composition.Slots = []Slot{{Name: "repeated a", Role: "claims", Pointer: "/canvas/0/text", ValueType: "string_array", Required: true, MaxChars: 80, MaxItems: 3, ItemTemplate: json.RawMessage(`{"id":"source-a","label":"Original A","bounds":{"x":1,"y":2,"width":3,"height":4}}`), ItemValuePointer: "/label", ItemIDPointer: "/id", ItemIDPrefix: "source-a-"}}
	arrayB := arrayA
	arrayB.Composition.Slots = append([]Slot(nil), arrayA.Composition.Slots...)
	arrayB.Composition.Slots[0].Name = "repeated b"
	arrayB.Composition.Slots[0].ItemTemplate = json.RawMessage(`{"id":"source-b","label":"Original B","bounds":{"x":1,"y":2,"width":3,"height":4}}`)
	arrayB.Composition.Slots[0].ItemIDPrefix = "source-b-"
	arrayAFP, err := s.Fingerprint(arrayA)
	if err != nil {
		t.Fatal(err)
	}
	arrayBFP, err := s.Fingerprint(arrayB)
	if err != nil {
		t.Fatal(err)
	}
	if arrayAFP != arrayBFP {
		t.Fatal("repeat item source copy and generated IDs did not deduplicate")
	}
	arrayB.Composition.Slots[0].ItemTemplate = json.RawMessage(`{"id":"source-b","label":"Original B","bounds":{"x":1,"y":2,"width":4,"height":4}}`)
	arrayBFP, err = s.Fingerprint(arrayB)
	if err != nil {
		t.Fatal(err)
	}
	if arrayAFP == arrayBFP {
		t.Fatal("repeat item geometry deduplicated")
	}
	undeclared := c2
	undeclared.StyleVariants = nil
	fp, err = s.Fingerprint(undeclared)
	if err != nil {
		t.Fatal(err)
	}
	if fp == fp1 {
		t.Fatal("undeclared color change deduplicated")
	}
}

func TestFingerprintKeepsUnverifiedAssetPath(t *testing.T) {
	root := t.TempDir()
	s, _ := NewStore(root)
	makeOne := func(name, path string) Contract {
		b, _ := json.Marshal(map[string]any{"slides": []any{map[string]any{"id": "s", "canvas": []any{map[string]any{"id": "i", "kind": "image", "asset_path": path}}}}})
		if err := os.WriteFile(filepath.Join(root, name+".json"), b, 0644); err != nil {
			t.Fatal(err)
		}
		return Contract{Composition: Composition{SpecPath: name + ".json", SlideID: "s"}}
	}
	a, err := s.Fingerprint(makeOne("a", "a.png"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.Fingerprint(makeOne("b", "b.png"))
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatal("unverified asset paths deduplicated")
	}
}

func TestFingerprintScopedContainerReferences(t *testing.T) {
	root := t.TempDir()
	s, _ := NewStore(root)
	makeOne := func(name, container, cell, block string) Contract {
		slide := map[string]any{
			"id":          name,
			"layouts":     []any{map[string]any{"id": container, "cells": []any{map[string]any{"id": cell, "blocks": []any{map[string]any{"id": block, "text": "fixed"}}}}}},
			"connections": []any{map[string]any{"id": "edge-" + name, "from": container + "/" + cell, "to": container + "/" + cell + "/" + block}},
		}
		b, _ := json.Marshal(map[string]any{"slides": []any{slide}})
		put(t, filepath.Join(root, name+".json"), b)
		return Contract{Composition: Composition{SpecPath: name + ".json", SlideID: name}}
	}
	a, err := s.Fingerprint(makeOne("a", "container-a", "cell-a", "block-a"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.Fingerprint(makeOne("b", "container-b", "cell-b", "block-b"))
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatal("renamed scoped container/cell/block references did not deduplicate")
	}
}

func TestIndexAliasRetainsIndependentSourceAndReview(t *testing.T) {
	root := t.TempDir()
	s, _ := NewStore(root)
	catalog := filepath.Join(root, "catalog.sqlite")
	if _, err := sqlite(catalog, "CREATE TABLE items(id TEXT PRIMARY KEY,kind TEXT,source_id TEXT,category TEXT,title TEXT,body TEXT,json TEXT);", false); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(catalog)
	manifest, _ := json.Marshal(map[string]any{"artifacts": map[string]Artifact{"sqlite": {Path: "catalog.sqlite", SHA256: hashBytes(b)}}})
	put(t, filepath.Join(root, "library/catalog-manifest.json"), manifest)
	short := []byte(`{"references":[]}`)
	put(t, filepath.Join(root, "library/reference-shortlist.json"), short)
	prefs, _ := json.Marshal(map[string]any{"shortlist_sha256": hashBytes(short), "choices": []any{}})
	put(t, filepath.Join(root, "library/reference-preferences.json"), prefs)
	for i, label := range []string{"a", "b"} {
		slide := map[string]any{"id": "slide-" + label, "canvas": []any{map[string]any{"id": "box-" + label, "kind": "text", "bounds": map[string]any{"x": 10, "y": 10, "width": 50, "height": 20}, "text": "copy " + label}}}
		spec, _ := json.Marshal(map[string]any{"slides": []any{slide}})
		put(t, filepath.Join(root, "spec-"+label+".json"), spec)
		source := []byte("source " + label)
		put(t, filepath.Join(root, "source-"+label+".txt"), source)
		state, pref := "measured_fixture", "preferred"
		if i == 1 {
			state, pref = "inventory", "avoid"
		}
		c := Contract{Schema: ContractSchema, ID: label, Version: "1", Kind: "layout", Name: label, Purpose: "alias fixture", Source: Source{SourceID: "source-" + label, Path: "source-" + label + ".txt", SourceSHA256: hashBytes(source), Slide: 1}, Composition: Composition{SpecPath: "spec-" + label + ".json", SpecSHA256: hashBytes(spec), SlideID: "slide-" + label, Slots: []Slot{{Name: "copy-" + label, Role: "claim", Pointer: "/canvas/0/text", ValueType: "string", Required: true, MaxChars: 20}}}, FitEnvelope: FitEnvelope{FontPolicy: "no_silent_shrink"}, Transforms: Transforms{Translation: "tested", Resize: "unsupported", Rotation: "unsupported"}, Preference: Preference{Value: pref}, Qualification: Qualification{State: state}}
		contract, _ := json.Marshal(c)
		put(t, filepath.Join(root, "library/contracts/"+label+".json"), contract)
	}
	s.CatalogPath = catalog
	index := filepath.Join(root, "index.sqlite")
	report, err := s.BuildIndex(index)
	if err != nil {
		t.Fatal(err)
	}
	if report.AliasCount != 1 {
		t.Fatalf("expected one alias: %+v", report)
	}
	a, err := s.Inspect(index, "a")
	if err != nil {
		t.Fatal(err)
	}
	bins, err := s.Inspect(index, "b")
	if err != nil {
		t.Fatal(err)
	}
	if a.CanonicalID != "a" || bins.CanonicalID != "a" || bins.Contract.Source.SourceID != "source-b" || bins.Contract.Qualification.State != "inventory" || bins.Contract.Preference.Value != "avoid" {
		t.Fatalf("alias inherited review or lost provenance: %+v", bins)
	}
	rows, err := sqlite(index, "SELECT id,source_id,state,preference,canonical_id FROM contracts ORDER BY id;", true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(rows), `"source-b"`) || !strings.Contains(string(rows), `"avoid"`) {
		t.Fatalf("index lost alias metadata: %s", rows)
	}
}
