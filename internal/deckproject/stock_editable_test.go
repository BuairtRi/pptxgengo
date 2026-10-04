package deckproject

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestStockEditableEntireBindableCatalogRoundTrips(t *testing.T) {
	catalog, err := wmdesign.LibraryCatalog(bundle(t), "")
	if err != nil {
		t.Fatal(err)
	}
	examples, err := wmdesign.LibraryReference(bundle(t), "", "", 2026)
	if err != nil {
		t.Fatal(err)
	}
	defs := map[string]wmdesign.LibraryTemplate{}
	for _, def := range catalog {
		defs[def.Key] = def
	}
	// The 587-page gallery includes one source-only template without a bound
	// example. Exercise every available stock value contract, not a fabricated
	// example for the source-only entry.
	if len(catalog) != 587 || len(examples.Slides) != 586 {
		t.Fatalf("expected 587 catalog entries and 586 bound examples, got %d / %d", len(catalog), len(examples.Slides))
	}
	for _, example := range examples.Slides {
		t.Run(example.Template, func(t *testing.T) {
			var values map[string]any
			if err := json.Unmarshal(example.Values, &values); err != nil {
				t.Fatal(err)
			}
			slide := Slide{ID: example.ID, ContentKind: example.ContentKind, Template: Reference{Scope: "shared", ID: example.Template}, Values: values}
			authored, err := StockEditableSlide(slide, defs[example.Template])
			if err != nil {
				t.Fatal(err)
			}
			p := &Project{SourcePath: "stock-test.yaml", Positions: map[string]Position{}, positionFiles: map[string]string{}}
			if err := p.expandContentAliases(authored, "/slide"); err != nil {
				t.Fatal(err)
			}
			var expanded Slide
			if err := json.Unmarshal(canonical(authored), &expanded); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(canonical(expanded), canonical(slide)) {
				t.Fatal("editable content changed the original stock values")
			}
		})
	}
}

func TestStockEditableSlideKeepsSharedIdentityAndSuppliedCopy(t *testing.T) {
	def := wmdesign.LibraryTemplate{
		TemplateDefinition: wmdesign.TemplateDefinition{Key: "interviews/test"},
		RawSlide:           json.RawMessage(`{"body":[{"type":"table","cols":[{"key":"n","label":"Person"},{"key":"r","label":"Role"}],"rows":[{"n":"Specimen name","r":"Specimen role"}]}]}`),
		Slots: []wmdesign.LibrarySlot{
			{Name: "title", SourcePointer: "/title"},
			{Name: "node01.cols.item01.label", SourcePointer: "/body/0/cols/0/label"},
			{Name: "node01.cols.item02.label", SourcePointer: "/body/0/cols/1/label"},
			{Name: "node01.rows.item01.n", SourcePointer: "/body/0/rows/0/n"},
			{Name: "node01.rows.item01.r", SourcePointer: "/body/0/rows/0/r"},
		},
	}
	slide := Slide{ID: "interviews", Template: Reference{Scope: "shared", ID: def.Key}, ContentKind: "supplied_content", Notes: "Exact original notes", Values: map[string]any{
		"slots": map[string]any{"title": "Who we interviewed", "node01.cols.item01.label": "Name", "node01.cols.item02.label": "Role", "node01.rows.item01.n": "Sam", "node01.rows.item01.r": "Director"},
		"keys":  map[string]any{"node01.rows": []any{"sam"}},
	}}
	before := string(canonical(slide))
	authored, err := StockEditableSlide(slide, def)
	if err != nil {
		t.Fatal(err)
	}
	if before != string(canonical(slide)) {
		t.Fatal("authoring helper mutated the original supplied slide")
	}
	if !reflect.DeepEqual(authored["template"], map[string]any{"scope": "shared", "id": def.Key}) {
		t.Fatal("stock template identity changed")
	}
	content := authored["content"].(map[string]any)
	if content["headline"] != "Who we interviewed" || authored["notes"] != slide.Notes {
		t.Fatal("headline or notes changed")
	}
	table := content["tables"].(map[string]any)["item_01"].(map[string]any)
	row := table["rows"].(map[string]any)["item_01"].(map[string]any)
	if row["person"] != "Sam" || row["role"] != "Director" {
		t.Fatalf("interview fields were not exposed as supplied readable values: %#v", row)
	}
	bindings := authored["bindings"].(map[string]any)
	if bindings["/tables/item_01/rows/item_01/person"] != "/slots/node01.rows.item01.n" || len(bindings) != len(def.Slots) {
		t.Fatal("bindings do not cover exact original stock slots")
	}
	if _, exists := authored["local_templates"]; exists {
		t.Fatal("stock helper introduced a local definition")
	}
	if _, exists := authored["values"].(map[string]any)["slots"]; exists {
		t.Fatal("copy duplicated in technical values")
	}
}

func TestStockEditableSlideTypedArraysRemainReadableAndExact(t *testing.T) {
	def := wmdesign.LibraryTemplate{TemplateDefinition: wmdesign.TemplateDefinition{Key: "cards/3"}}
	slide := Slide{ID: "controls", Template: Reference{Scope: "shared", ID: def.Key}, ContentKind: "supplied_content", Values: map[string]any{
		"title": "Three controls", "cards": []any{map[string]any{"title": "Trace", "body": "Keep evidence."}}, "keys": map[string]any{"cards": []any{"trace"}},
	}}
	authored, err := StockEditableSlide(slide, def)
	if err != nil {
		t.Fatal(err)
	}
	content := authored["content"].(map[string]any)
	if content["headline"] != "Three controls" || !reflect.DeepEqual(content["cards"], slide.Values["cards"]) {
		t.Fatal("typed source values changed")
	}
	if authored["bindings"].(map[string]any)["cards"] != "/cards" {
		t.Fatal("typed array binding missing")
	}
	if len(authored["values"].(map[string]any)) != 1 {
		t.Fatal("technical stable keys not preserved separately")
	}
}

func TestStockEditableSlideRejectsLocalAndUndeclaredCopy(t *testing.T) {
	def := wmdesign.LibraryTemplate{TemplateDefinition: wmdesign.TemplateDefinition{Key: "test"}, Slots: []wmdesign.LibrarySlot{{Name: "title", SourcePointer: "/title"}}}
	local := Slide{ID: "one", Template: Reference{Scope: "local", ID: "test"}, Values: map[string]any{}}
	if _, err := StockEditableSlide(local, def); err == nil {
		t.Fatal("local derivative was represented as stock")
	}
	shared := Slide{ID: "one", Template: Reference{Scope: "shared", ID: "test"}, Values: map[string]any{"slots": map[string]any{"invented": "Copy"}}}
	if _, err := StockEditableSlide(shared, def); err == nil {
		t.Fatal("undeclared stock slot accepted")
	}
}

func TestStockEditableCoverUsesExplicitUniqueSourceStyles(t *testing.T) {
	authored, err := StockScaffoldSlide(bundle(t), "cover/grid", "opening", 2026)
	if err != nil {
		t.Fatal(err)
	}
	copy := authored["content"].(map[string]any)
	if copy["headline"] == nil || copy["section_label"] == nil || copy["subtitle"] == nil {
		t.Fatalf("cover fields must be named: %#v", copy)
	}
}

func TestStockRowAliasesUseAllSuppliedHeadingsForCollisions(t *testing.T) {
	node := map[string]any{"cols": []any{map[string]any{"key": "n", "label": "Person"}, map[string]any{"key": "r", "label": "Role"}}}
	def := wmdesign.LibraryTemplate{Slots: []wmdesign.LibrarySlot{{Name: "first", SourcePointer: "/body/0/cols/0/label"}, {Name: "second", SourcePointer: "/body/0/cols/1/label"}}}
	values := map[string]any{"first": "Owner", "second": "Owner"}
	if stockRowField("n", node, 0, def, values) != "owner_n" || stockRowField("r", node, 0, def, values) != "owner_r" {
		t.Fatal("duplicate supplied headings must retain distinct meaningful field names")
	}
}
