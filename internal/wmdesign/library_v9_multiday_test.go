package wmdesign

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestLibraryV9MultidayAmendmentsPreserveContentAndAreAtomic(t *testing.T) {
	catalog, err := LibraryCatalog(v9IntakeBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	for _, def := range catalog {
		if def.Key != "workshop-multiday/three-day-grid" && def.Key != "workshop-multiday/two-day-onsite-remote" {
			continue
		}
		t.Run(def.Key, func(t *testing.T) {
			original, err := compileLibrarySlide(def.RawSlide, nil)
			if err != nil {
				t.Fatal(err)
			}
			before := v6ComparisonSlide(t, original, false)
			amended := original
			if err := applyV9MultidayWorkshopRefinements(def.Key, &amended); err != nil {
				t.Fatal(err)
			}
			once := v6ComparisonSlide(t, amended, false)
			if err := applyV9MultidayWorkshopRefinements(def.Key, &amended); err != nil {
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
				var heights []float64
				for _, obj := range []map[string]any{a, b} {
					if cols, ok := obj["cols"].([]any); ok {
						for _, col := range cols {
							delete(col.(map[string]any), "w")
						}
					}
					if rows, ok := obj["rows"].([]any); ok {
						total := 0.0
						for _, row := range rows {
							r := row.(map[string]any)
							total += intakeRepairNumber(r, "h")
							delete(r, "h")
						}
						heights = append(heights, total)
					}
				}
				if !reflect.DeepEqual(a, b) || (len(heights) == 2 && heights[0] != heights[1]) {
					t.Fatal("copy, font, semantic value or total table height changed")
				}
			}
			if def.Key == "workshop-multiday/three-day-grid" {
				broken, err := compileLibrarySlide(def.RawSlide, nil)
				if err != nil {
					t.Fatal(err)
				}
				obj, err := libraryObject(broken.Nodes[3].Scene.Node)
				if err != nil {
					t.Fatal(err)
				}
				obj["rows"].([]any)[4].(map[string]any)["h"] = 55
				broken.Nodes[3].Scene.Node, err = json.Marshal(obj)
				if err != nil {
					t.Fatal(err)
				}
				beforeBroken := v6ComparisonSlide(t, broken, false)
				if err := applyV9MultidayWorkshopRefinements(def.Key, &broken); err == nil {
					t.Fatal("unexpected source geometry accepted")
				}
				if !reflect.DeepEqual(beforeBroken, v6ComparisonSlide(t, broken, false)) {
					t.Fatal("failed amendment partially changed the slide")
				}
			}
		})
	}
}
