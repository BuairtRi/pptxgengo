package wmdesign

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestLibraryV10AllocationsPreserveContentAndAreAtomic(t *testing.T) {
	catalog, err := LibraryCatalog(v10IntakeBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	for _, def := range catalog {
		if def.Key != "road-fork/parallel" && def.Key != "road-fork/decision" && def.Key != "road-fork/foundation-then-waves" && def.Key != "roadmap-narrative/objectives-tree" && def.Key != "roadmap-narrative/waves-criteria" && def.Key != "roadmap-narrative/horizon-table" {
			continue
		}
		t.Run(def.Key, func(t *testing.T) {
			original, err := compileLibrarySlide(def.RawSlide, nil)
			if err != nil {
				t.Fatal(err)
			}
			before := v6ComparisonSlide(t, original, false)
			amended := original
			if err := applyV10AllocationRefinements(def.Key, &amended); err != nil {
				t.Fatal(err)
			}
			once := v6ComparisonSlide(t, amended, false)
			if err := applyV10AllocationRefinements(def.Key, &amended); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(once, v6ComparisonSlide(t, amended, false)) || !reflect.DeepEqual(before, v6ComparisonSlide(t, original, false)) {
				t.Fatal("amendment is not idempotent or mutated the original slide")
			}
			beforeChrome, afterChrome := original, amended
			beforeChrome.Nodes, afterChrome.Nodes = nil, nil
			if def.Key == "roadmap-narrative/waves-criteria" {
				if amended.Frame.TitleLines != 2 {
					t.Fatal("missing authored two-line title allocation")
				}
				afterChrome.Frame.TitleLines = original.Frame.TitleLines
			}
			if !reflect.DeepEqual(beforeChrome, afterChrome) {
				t.Fatal("slide copy, fonts, chrome or unrelated frame allocation changed")
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
				if i == 0 {
					for _, obj := range []map[string]any{a, b} {
						switch def.Key {
						case "roadmap-narrative/objectives-tree":
							delete(obj["cols"].([]any)[1].(map[string]any), "w")
						case "roadmap-narrative/horizon-table":
							cols := obj["cols"].([]any)
							delete(cols[0].(map[string]any), "w")
							delete(cols[5].(map[string]any), "w")
						case "roadmap-narrative/waves-criteria":
							delete(obj, "y")
							for _, row := range obj["rows"].([]any) {
								delete(row.(map[string]any), "h")
							}
						default:
							delete(obj, "h")
						}
					}
				}
				if !reflect.DeepEqual(a, b) {
					t.Fatal("copy, font, semantic value, topology or unrelated geometry changed")
				}
			}
			broken, err := compileLibrarySlide(def.RawSlide, nil)
			if err != nil {
				t.Fatal(err)
			}
			obj, err := libraryObject(broken.Nodes[0].Scene.Node)
			if err != nil {
				t.Fatal(err)
			}
			switch def.Key {
			case "roadmap-narrative/objectives-tree":
				obj["cols"].([]any)[1].(map[string]any)["w"] = 249
			case "roadmap-narrative/horizon-table":
				obj["cols"].([]any)[5].(map[string]any)["w"] = 91
			case "roadmap-narrative/waves-criteria":
				obj["rows"].([]any)[3].(map[string]any)["h"] = 66
			default:
				obj["h"] = 325
			}
			broken.Nodes[0].Scene.Node, err = json.Marshal(obj)
			if err != nil {
				t.Fatal(err)
			}
			beforeBroken := v6ComparisonSlide(t, broken, false)
			if err := applyV10AllocationRefinements(def.Key, &broken); err == nil {
				t.Fatal("unexpected source geometry accepted")
			}
			if !reflect.DeepEqual(beforeBroken, v6ComparisonSlide(t, broken, false)) {
				t.Fatal("failed amendment partially changed the slide")
			}
		})
	}
}
