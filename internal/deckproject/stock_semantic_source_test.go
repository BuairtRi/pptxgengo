package deckproject

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestStockAliasesIndependentOfSuppliedHeadingCopy(t *testing.T) {
	def := wmdesign.LibraryTemplate{TemplateDefinition: wmdesign.TemplateDefinition{Key: "table/test"}, RawSlide: json.RawMessage(`{"body":[{"type":"table","cols":[{"key":"a","label":"Owner"}],"rows":[{"a":"Name"}]}]}`), Slots: []wmdesign.LibrarySlot{{Name: "heading", SourcePointer: "/body/0/cols/0/label", Kind: "string"}, {Name: "owner", SourcePointer: "/body/0/rows/0/a", Kind: "string"}}}
	makeSlide := func(heading string) Slide {
		return Slide{ID: "one", Template: Reference{Scope: "shared", ID: def.Key}, Values: map[string]any{"slots": map[string]any{"heading": heading, "owner": "Pat"}}}
	}
	a, e := StockEditableSlide(makeSlide("Accountable person"), def)
	if e != nil {
		t.Fatal(e)
	}
	b, e := StockEditableSlide(makeSlide("Decision lead"), def)
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(a["bindings"], b["bindings"]) {
		t.Fatal("supplied heading changed stable alias")
	}
}
func TestStockScaffoldCommentsAndDecorativeContract(t *testing.T) {
	data, e := StockScaffoldSlideSource(bundle(t), "lifecycle/three-phases", "lifecycle", 2026)
	if e != nil {
		t.Fatal(e)
	}
	text := string(data)
	if !strings.Contains(text, "approximate capacity:") || !strings.Contains(text, "Phase 1") || !strings.Contains(text, "native fit not evaluated") || strings.Contains(text, "block_") {
		t.Fatalf("missing/opaque comments: %s", text)
	}
	authored, e := StockScaffoldSlide(bundle(t), "lifecycle/three-phases", "lifecycle", 2026)
	if e != nil {
		t.Fatal(e)
	}
	content := authored["content"].(map[string]any)
	if content["phases"] == nil || content["panels"] != nil {
		t.Fatal("empty panels exposed as editable copy")
	}
	slots := authored["values"].(map[string]any)["slots"].(map[string]any)
	if len(slots) != 15 {
		t.Fatalf("decorative original slots lost: %d", len(slots))
	}
	for _, v := range slots {
		if v != "" {
			t.Fatal("nonempty copy hidden as decoration")
		}
	}
}
