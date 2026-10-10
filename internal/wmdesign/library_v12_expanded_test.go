package wmdesign

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"
)

func expandedV12Bundle() string {
	return filepath.Join("..", "..", "library", "wm-design-system", "v12")
}

func TestLibraryV12ExpandedCatalogAndRetainedCompositions(t *testing.T) {
	bundle := expandedV12Bundle()
	s, err := Load(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	if s.Commit != "36132d5637abdbdeb70945ad650b795cabea05ef" || s.Revision != LibraryRevisionV12 {
		t.Fatal("expanded source identity changed")
	}
	current, err := LibraryCatalog(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	previous, err := LibraryCatalog(v12Bundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	old := map[string]LibraryTemplate{}
	for _, def := range previous {
		old[def.Key] = def
	}
	added := map[string]int{}
	retained, revised, active := 0, 0, 0
	for _, def := range current {
		if def.Status != "deprecated" {
			active++
		}
		before, ok := old[def.Key]
		if !ok {
			added[def.Family]++
			continue
		}
		if def.Key == "guide/section-guide" || def.Key == "plan/gantt" {
			if v12JSONEqual(t, def.RawSlide, before.RawSlide) {
				t.Fatalf("revision missing: %s", def.Key)
			}
			revised++
			continue
		}
		if !v12JSONEqual(t, def.RawSlide, before.RawSlide) || !reflect.DeepEqual(def.Slots, before.Slots) || !reflect.DeepEqual(def.Arrays, before.Arrays) {
			t.Fatalf("retained source/bindings changed: %s", def.Key)
		}
		retained++
	}
	if len(current) != 735 || active != 734 || retained != 675 || revised != 2 || !reflect.DeepEqual(added, map[string]int{"approach": 11, "covers": 24, "indexes": 11, "roadmaps": 12}) {
		t.Fatalf("catalog delta changed: %d/%d retained=%d revised=%d added=%v", len(current), active, retained, revised, added)
	}
	prior, err := LibrarySourceReference(v12Bundle(), "", "", 2026)
	if err != nil {
		t.Fatal(err)
	}
	now, err := LibrarySourceReference(bundle, "", "", 2026)
	if err != nil {
		t.Fatal(err)
	}
	compiled := map[string]SlideSpec{}
	for _, slide := range prior.Slides {
		compiled[slide.TemplateBinding.Template] = slide
	}
	for _, slide := range now.Slides {
		key := slide.TemplateBinding.Template
		before, ok := compiled[key]
		if !ok || key == "guide/section-guide" || key == "plan/gantt" {
			continue
		}
		if !samePublicationComposition(before, slide) {
			t.Fatalf("retained compiled composition changed: %s", key)
		}
	}
}

func TestLibraryV12ExpandedAllContentRoundTrips(t *testing.T) {
	bundle := expandedV12Bundle()
	source, err := LibrarySourceReference(bundle, "", "", 2026)
	if err != nil {
		t.Fatal(err)
	}
	input, err := LibraryReference(bundle, "", "", 2026)
	if err != nil {
		t.Fatal(err)
	}
	bound, _, err := BindTemplates(bundle, "", input)
	if err != nil {
		t.Fatal(err)
	}
	stock := map[string]SlideSpec{}
	for _, slide := range source.Slides {
		stock[slide.TemplateBinding.Template] = slide
	}
	if len(source.Slides) != 735 || len(bound.Slides) != 734 {
		t.Fatal("source/bound count changed")
	}
	for _, slide := range bound.Slides {
		key := slide.TemplateBinding.Template
		if usesLegacyTemplate(LibraryRevisionV12, key) {
			continue
		}
		if !reflect.DeepEqual(v6ComparisonSlide(t, stock[key], true), v6ComparisonSlide(t, slide, true)) {
			t.Fatalf("content roundtrip changed: %s", key)
		}
	}
}

func TestLibraryV12CatalogGeometryGuards(t *testing.T) {
	bundle := expandedV12Bundle()
	s, err := Load(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	typeEngine, err := NewTypographyEngine(filepath.Join(bundle, "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	r := renderer{source: s, typeEngine: typeEngine, bundle: bundle}
	ctx := SceneContext{Surface: "light"}
	for _, raw := range []string{
		`{"type":"preset","shape":"diamond","x":10,"y":20,"w":0,"h":7}`,
		`{"type":"preset","shape":"ellipse","x":10,"y":20,"w":7,"h":7}`,
		`{"type":"toclist","x":57,"y":126,"w":400,"items":[{"n":"1","text":"One","page":"3","level":3}]}`,
		`{"type":"toclist","x":57,"y":126,"w":90,"items":[{"n":"1","text":"A heading that cannot fit this allocation","page":"3"}]}`,
		`{"type":"toclist","x":57,"y":126,"w":400,"unknown":true,"items":[]}`,
	} {
		if _, handled, err := r.planV12CatalogScene("guard", json.RawMessage(raw), ctx); !handled || err == nil {
			t.Fatalf("invalid source geometry accepted: %s, %v", raw, err)
		}
	}
	for _, raw := range []string{
		`{"type":"preset","shape":"diamond","surface":"inverse","x":10,"y":20,"w":7,"h":7}`,
		`{"type":"toclist","x":57,"y":126,"w":400,"rules":true,"items":[{"n":"01","text":"One","page":"3"},{"n":"","text":"Detail","page":"4","level":2}]}`,
	} {
		if plan, handled, err := r.planV12CatalogScene("valid", json.RawMessage(raw), ctx); !handled || err != nil || len(plan.Items) == 0 {
			t.Fatalf("valid source geometry rejected: %s, %v", raw, err)
		}
	}
	r.source.Revision = LibraryRevisionV11
	if _, handled, err := r.planV12CatalogScene("historical", json.RawMessage(`{"type":"toclist"}`), ctx); handled || err != nil {
		t.Fatal("V12-only primitive changed historical revision")
	}
}
