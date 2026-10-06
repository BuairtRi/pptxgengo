package wmdesign

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestLibraryV8GuideAmendmentsPreserveContentAndAreAtomic(t *testing.T) {
	catalog, err := LibraryCatalog(v8IntakeBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	for _, def := range catalog {
		if def.Key != "workshop-guide/question-bank" && def.Key != "workshop-guide/question-ladder" {
			continue
		}
		t.Run(def.Key, func(t *testing.T) {
			original, err := compileLibrarySlide(def.RawSlide, nil)
			if err != nil {
				t.Fatal(err)
			}
			before := v6ComparisonSlide(t, original, false)
			amended := original
			if err := applyV8WorkshopGuideRefinements(def.Key, &amended); err != nil {
				t.Fatal(err)
			}
			once := v6ComparisonSlide(t, amended, false)
			if err := applyV8WorkshopGuideRefinements(def.Key, &amended); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(once, v6ComparisonSlide(t, amended, false)) || !reflect.DeepEqual(before, v6ComparisonSlide(t, original, false)) {
				t.Fatal("amendment is not idempotent or mutated the original slide")
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
					if cols, ok := obj["cols"].([]any); ok {
						for _, col := range cols {
							delete(col.(map[string]any), "w")
						}
					}
				}
				if !reflect.DeepEqual(a, b) {
					t.Fatal("copy, font or semantic value changed")
				}
			}
			if def.Key == "workshop-guide/question-ladder" {
				broken, err := compileLibrarySlide(def.RawSlide, nil)
				if err != nil {
					t.Fatal(err)
				}
				obj, err := libraryObject(broken.Nodes[6].Scene.Node)
				if err != nil {
					t.Fatal(err)
				}
				obj["y"] = 199
				broken.Nodes[6].Scene.Node, err = json.Marshal(obj)
				if err != nil {
					t.Fatal(err)
				}
				beforeBroken := v6ComparisonSlide(t, broken, false)
				if err := applyV8WorkshopGuideRefinements(def.Key, &broken); err == nil {
					t.Fatal("unexpected source geometry accepted")
				}
				if !reflect.DeepEqual(beforeBroken, v6ComparisonSlide(t, broken, false)) {
					t.Fatal("failed amendment partially changed the slide")
				}
			}
		})
	}
}
