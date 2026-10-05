package wmdesign

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"
)

func v7IntakeBundle() string {
	return filepath.Join("..", "..", "planning", "wm-design-contracts", "v7", "intake-20261005-616-frozen", "bundle")
}

func TestLibraryV7PinnedIntakeAndV6Inheritance(t *testing.T) {
	source, err := Load(v7IntakeBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	if source.Revision != LibraryRevisionV7 || source.Commit != "c788cefeb5bb409118ac217adb53216d8156eec3" {
		t.Fatal("candidate identity changed")
	}
	current, err := LibraryCatalog(v7IntakeBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	previous, err := LibraryCatalog(v6IntakeBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	old := map[string]LibraryTemplate{}
	for _, def := range previous {
		old[def.Key] = def
	}
	additions := 0
	for _, def := range current {
		before, exists := old[def.Key]
		if !exists {
			additions++
			if def.Family != "workshops" || def.SourceRevision != LibraryRevisionV7 {
				t.Fatalf("unexpected addition %s", def.Key)
			}
			continue
		}
		delete(old, def.Key)
		var a, b any
		if err := json.Unmarshal(before.RawSlide, &a); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(def.RawSlide, &b); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(a, b) || def.ContentContract != before.ContentContract || !reflect.DeepEqual(def.Slots, before.Slots) || !reflect.DeepEqual(def.Arrays, before.Arrays) || !reflect.DeepEqual(def.ValueSchema, before.ValueSchema) {
			t.Fatalf("retained source or binding changed %s", def.Key)
		}
		left, err := compileLibrarySlide(before.RawSlide, nil)
		if err != nil {
			t.Fatal(err)
		}
		right, err := compileLibrarySlide(def.RawSlide, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := applyLibraryRefinements(before.Key, before.SourceRevision, &left); err != nil {
			t.Fatal(err)
		}
		if err := applyLibraryRefinements(def.Key, def.SourceRevision, &right); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(v6ComparisonSlide(t, left, false), v6ComparisonSlide(t, right, false)) {
			t.Fatalf("retained amendments changed %s", def.Key)
		}
	}
	if len(current) != 616 || additions != 14 || len(old) != 0 {
		t.Fatalf("intake drift: total %d, added %d, removed %d", len(current), additions, len(old))
	}
}

func TestLibraryV7WorkshopsRoundTripAndBuild(t *testing.T) {
	bundle := v7IntakeBundle()
	source, err := LibrarySourceReference(bundle, "", "workshops", 2026)
	if err != nil {
		t.Fatal(err)
	}
	input, err := LibraryReference(bundle, "", "workshops", 2026)
	if err != nil {
		t.Fatal(err)
	}
	bound, _, err := BindTemplates(bundle, "", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(source.Slides) != 14 || len(bound.Slides) != 14 {
		t.Fatal("workshop inventory drift")
	}
	for i := range source.Slides {
		if !reflect.DeepEqual(v6ComparisonSlide(t, source.Slides[i], true), v6ComparisonSlide(t, bound.Slides[i], true)) {
			t.Fatalf("binding changed authored workshop %s", source.Slides[i].TemplateBinding.Template)
		}
	}
	for name, doc := range map[string]Document{"source": source, "bound": bound} {
		t.Run(name, func(t *testing.T) {
			deck, report, err := BuildWithEngine(bundle, "", doc, CandidateEngine)
			if err != nil {
				t.Fatal(err)
			}
			if len(deck) == 0 || len(report.Slides) != 14 || report.PowerPointVerified || report.VisuallyReviewed {
				t.Fatal("incorrect build or native acceptance claim")
			}
		})
	}
}

func TestLibraryV7WorkshopAmendmentsAreAtomicAndIdempotent(t *testing.T) {
	catalog, err := LibraryCatalog(v7IntakeBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	for _, def := range catalog {
		if def.Family != "workshops" {
			continue
		}
		t.Run(def.Key, func(t *testing.T) {
			original, err := compileLibrarySlide(def.RawSlide, nil)
			if err != nil {
				t.Fatal(err)
			}
			before := v6ComparisonSlide(t, original, false)
			amended := original
			if err := applyV7WorkshopRefinements(def.Key, &amended); err != nil {
				t.Fatal(err)
			}
			once := v6ComparisonSlide(t, amended, false)
			if err := applyV7WorkshopRefinements(def.Key, &amended); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(once, v6ComparisonSlide(t, amended, false)) {
				t.Fatal("non-idempotent amendment")
			}
			if !reflect.DeepEqual(before, v6ComparisonSlide(t, original, false)) {
				t.Fatal("source mutation")
			}
			for i, node := range original.Nodes {
				a, err := libraryObject(node.Scene.Node)
				if err != nil {
					t.Fatal(err)
				}
				b, err := libraryObject(amended.Nodes[i].Scene.Node)
				if err != nil {
					t.Fatal(err)
				}
				for _, obj := range []map[string]any{a, b} {
					delete(obj, "y")
					delete(obj, "rowHeight")
					if cols, ok := obj["cols"].([]any); ok {
						for _, col := range cols {
							delete(col.(map[string]any), "w")
						}
					}
				}
				if !reflect.DeepEqual(a, b) {
					t.Fatal("copy, typography, or semantic values changed")
				}
			}
		})
	}
	for _, def := range catalog {
		if def.Key != "workshop-overview/dense-leave-with" {
			continue
		}
		broken, err := compileLibrarySlide(def.RawSlide, nil)
		if err != nil {
			t.Fatal(err)
		}
		obj, err := libraryObject(broken.Nodes[5].Scene.Node)
		if err != nil {
			t.Fatal(err)
		}
		obj["rowHeight"] = 999
		broken.Nodes[5].Scene.Node, err = json.Marshal(obj)
		if err != nil {
			t.Fatal(err)
		}
		before := v6ComparisonSlide(t, broken, false)
		if applyV7WorkshopRefinements(def.Key, &broken) == nil {
			t.Fatal("unexpected geometry accepted")
		}
		if !reflect.DeepEqual(before, v6ComparisonSlide(t, broken, false)) {
			t.Fatal("failed amendment partially published")
		}
	}
}
