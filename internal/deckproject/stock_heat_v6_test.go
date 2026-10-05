package deckproject

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestStockEditableV6HeatmapsPreserveSuppliedCopyAndTopology(t *testing.T) {
	bundle := filepath.Join("..", "..", "planning", "wm-design-contracts", "v6", "intake-20261004-602-frozen", "bundle")
	catalog, err := wmdesign.LibraryCatalog(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	examples, err := wmdesign.LibraryReference(bundle, "", "heatmaps", 2026)
	if err != nil {
		t.Fatal(err)
	}
	if len(examples.Slides) != 37 {
		t.Fatal("heatmap inventory changed")
	}
	defs := map[string]wmdesign.LibraryTemplate{}
	for _, def := range catalog {
		defs[def.Key] = def
	}
	roles := map[string]int{}
	for _, example := range examples.Slides {
		t.Run(example.Template, func(t *testing.T) {
			def := defs[example.Template]
			metadata, err := wmdesign.LibraryAuthoringMetadata(def)
			if err != nil {
				t.Fatal(err)
			}
			if metadata.ReviewStatus != "inferred_from_pinned_source" {
				t.Fatal("semantic review invented")
			}
			var values map[string]any
			if err := json.Unmarshal(example.Values, &values); err != nil {
				t.Fatal(err)
			}
			slots := values["slots"].(map[string]any)
			for name, value := range slots {
				switch value := value.(type) {
				case string:
					slots[name] = "Authored: " + value
				case bool:
					slots[name] = !value
				}
			}
			slide := Slide{ID: example.ID, ContentKind: "supplied_content", Template: Reference{Scope: "shared", ID: example.Template}, Values: values}
			before := canonical(slide)
			authored, err := StockEditableSlide(slide, def)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(before, canonical(slide)) {
				t.Fatal("caller values mutated")
			}
			bindings := authored["bindings"].(map[string]any)
			for _, slot := range metadata.Slots {
				roles[slot.Role]++
				if bindings[slot.Alias] != "/slots/"+escape(slot.Name) {
					t.Fatalf("source copy lacks its semantic binding: %s %s", slot.SourcePointer, slot.Alias)
				}
				if strings.Contains(slot.SourcePointer, "/rowGroups/") && !strings.HasSuffix(slot.SourcePointer, "/label") {
					t.Fatal("structural row range exposed as copy")
				}
			}
			p := &Project{SourcePath: "stock-heat.yaml", Positions: map[string]Position{}, positionFiles: map[string]string{}}
			if err := p.expandContentAliases(authored, "/slide"); err != nil {
				t.Fatal(err)
			}
			var expanded Slide
			if err := json.Unmarshal(canonical(authored), &expanded); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(before, canonical(expanded)) {
				t.Fatal("friendly projection changed supplied values or stable array identities")
			}
		})
	}
	for _, role := range []string{"priority_value", "reference_id", "reference_active", "group_label", "evidence_body"} {
		if roles[role] == 0 {
			t.Fatal("new source meaning missing", role)
		}
	}
}
