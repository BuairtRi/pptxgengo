package wmdesign

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"
)

func TestV5NativeDonutRefinementIndependentPreservation(t *testing.T) {
	entries := intakeRepairEntries(t, filepath.Join("..", "..", "planning", "wm-design-contracts", "v5", "intake-20261003-587-frozen", "bundle", "source", "templates", "library"))
	const key = "value-types/hard-soft-right"
	original := intakeRepairSlide(t, entries[key])
	doc := intakeRepairSlide(t, entries[key])
	if err := applyLibraryRefinements(key, LibraryRevisionV5, &doc); err != nil {
		t.Fatal(err)
	}
	first, _ := json.Marshal(doc)
	if err := applyLibraryRefinements(key, LibraryRevisionV5, &doc); err != nil {
		t.Fatal(err)
	}
	second, _ := json.Marshal(doc)
	if !bytes.Equal(first, second) {
		t.Fatal("native refinement is cumulative")
	}
	originalNode, _ := libraryObject(original.Nodes[0].Scene.Node)
	refinedNode, _ := libraryObject(doc.Nodes[0].Scene.Node)
	if fmt.Sprint(refinedNode["holeSize"]) != "50" {
		t.Fatalf("hole size %v", refinedNode["holeSize"])
	}
	delete(refinedNode, "holeSize")
	if !reflect.DeepEqual(originalNode, refinedNode) {
		t.Fatal("refinement changed source chart fields beyond hole size")
	}
	doc.Nodes[0].Scene.Node = append(json.RawMessage(nil), original.Nodes[0].Scene.Node...)
	if !reflect.DeepEqual(original, doc) {
		t.Fatal("refinement changed other slide nodes, copy or geometry")
	}
	for _, revision := range []string{LibraryRevisionV1, LibraryRevisionV2, LibraryRevisionV3, LibraryRevisionV4} {
		d := intakeRepairSlide(t, entries[key])
		if err := applyLibraryRefinements(key, revision, &d); err != nil {
			t.Fatal(err)
		}
		obj, _ := libraryObject(d.Nodes[0].Scene.Node)
		if _, present := obj["holeSize"]; present {
			t.Fatalf("hole refinement leaked into %s", revision)
		}
	}
	neighbor := intakeRepairSlide(t, entries["value-types/hard-soft"])
	before, _ := json.Marshal(neighbor)
	if err := applyV5NativeChartRefinement("value-types/hard-soft", &neighbor); err != nil {
		t.Fatal(err)
	}
	after, _ := json.Marshal(neighbor)
	if !bytes.Equal(before, after) {
		t.Fatal("native refinement changed neighboring template")
	}
	for _, raw := range []json.RawMessage{json.RawMessage(`{"type":"chart","kind":"pie"}`), json.RawMessage(`{"type":`)} {
		d := intakeRepairSlide(t, entries[key])
		d.Nodes[0].Scene.Node = raw
		originalRaw := append(json.RawMessage(nil), raw...)
		before, _ := json.Marshal(d)
		if err := applyV5NativeChartRefinement(key, &d); err == nil {
			t.Fatal("invalid target accepted")
		}
		after, _ := json.Marshal(d)
		if !bytes.Equal(before, after) || !bytes.Equal(originalRaw, d.Nodes[0].Scene.Node) {
			t.Fatal("rejected target partially mutated")
		}
	}
}

func TestV5NativeDonutHoleRuntimeIndependentScope(t *testing.T) {
	r := intakeTestRenderer(t)
	base := map[string]any{"type": "chart", "kind": "doughnut", "x": 57, "y": 126, "w": 270, "h": 306, "categories": []string{"Hard", "Capacity", "Risk"}, "series": []any{map[string]any{"name": "Value", "values": []int{1500000, 1000000, 100000}}}, "format": map[string]any{"kind": "currency"}}
	r.source.Revision = LibraryRevisionV5
	beforeRaw, _ := json.Marshal(base)
	before, _, err := r.planChartScene("chart", beforeRaw, SceneContext{Surface: "light"})
	if err != nil {
		t.Fatal(err)
	}
	base["holeSize"] = 50
	afterRaw, _ := json.Marshal(base)
	after, _, err := r.planChartScene("chart", afterRaw, SceneContext{Surface: "light"})
	if err != nil {
		t.Fatal(err)
	}
	if len(before.Items) != len(after.Items) {
		t.Fatal("chart item count changed")
	}
	found := false
	for i := range before.Items {
		b, a := before.Items[i].Chart, after.Items[i].Chart
		if b == nil || a == nil {
			continue
		}
		found = true
		if *b.Options.HoleSize != 60 || *a.Options.HoleSize != 50 {
			t.Fatal("unexpected native hole sizes")
		}
		b.Options.HoleSize, a.Options.HoleSize = nil, nil
		if !reflect.DeepEqual(b, a) {
			t.Fatal("hole option changed native data, fonts, position or other chart options")
		}
	}
	if !found {
		t.Fatal("native chart missing")
	}
	for _, tc := range []struct {
		revision, kind string
		size           float64
		valid          bool
	}{
		{LibraryRevisionV5, "doughnut", 30, true}, {LibraryRevisionV5, "doughnut", 80, true},
		{LibraryRevisionV5, "doughnut", 29.99, false}, {LibraryRevisionV5, "doughnut", 80.01, false},
		{LibraryRevisionV4, "doughnut", 50, false}, {LibraryRevisionV3, "doughnut", 50, false},
		{LibraryRevisionV2, "doughnut", 50, false}, {LibraryRevisionV1, "doughnut", 50, false},
		{LibraryRevisionV5, "pie", 50, false}, {LibraryRevisionV5, "column", 50, false},
	} {
		r.source.Revision = tc.revision
		base["kind"], base["holeSize"] = tc.kind, tc.size
		raw, _ := json.Marshal(base)
		_, _, err := r.planChartScene("chart", raw, SceneContext{Surface: "light"})
		if (err == nil) != tc.valid {
			t.Fatalf("%s %s %v: %v", tc.revision, tc.kind, tc.size, err)
		}
	}
}
