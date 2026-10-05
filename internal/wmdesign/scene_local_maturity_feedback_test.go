package wmdesign

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalMaturityCompleteAllocationAndLabelHeadroom(t *testing.T) {
	bundle := filepath.Join("..", "..", "planning", "wm-design-contracts", "v5", "intake-20261003-587-frozen", "bundle")
	allocation := Rect{100, 190, 700, 240}
	args := map[string]any{"stages": []any{
		map[string]any{"label": "Manual", "text": "People perform each step."},
		map[string]any{"label": "Assisted", "text": "Tools support bounded tasks."},
		map[string]any{"label": "Connected", "text": "Teams coordinate handoffs."},
		map[string]any{"label": "Measured", "text": "Owners review delivery outcomes."},
	}}
	raw, err := ComposeSceneNode("maturity", args, allocation)
	if err != nil {
		t.Fatal(err)
	}
	path := "/local_templates/progression/nodes/curve"
	doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{{ID: "local-maturity", Frame: FrameRequest{NoHeader: true}, Nodes: []Node{{ID: "progression", Kind: "scene", Scene: &SceneSpec{Node: raw, Path: path, Allocation: &allocation, Keys: map[string][]string{path + "/stages": {"manual", "assisted", "connected", "measured"}}}}}}}}
	_, report, err := BuildWithEngine(bundle, "", doc, CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, text := range report.Slides[0].Texts {
		if !strings.HasPrefix(text.ID, "progression.") {
			continue
		}
		count++
		if !inside(text.Rect, allocation) {
			t.Fatalf("local maturity text outside allocation: %s %+v", text.ID, text.Rect)
		}
	}
	for _, shape := range report.Slides[0].Shapes {
		if strings.HasPrefix(shape.ID, "progression.") && !inside(shape.Rect, allocation) {
			t.Fatalf("local maturity ink outside allocation: %s %+v", shape.ID, shape.Rect)
		}
	}
	if count < 8 {
		t.Fatalf("only%d local labels checked", count)
	}
	// A direct source node without local stroke padding must still fail the
	// allocation gate. Never hide its ink overflow by changing reported bounds.
	args["type"], args["x"], args["y"], args["w"], args["h"] = "maturity", allocation.X, allocation.Y, allocation.W, allocation.H
	raw, err = json.Marshal(args)
	if err != nil {
		t.Fatal(err)
	}
	doc.Slides[0].Nodes[0].Scene.Node = raw
	if _, _, err = BuildWithEngine(bundle, "", doc, CandidateEngine); err == nil || !strings.Contains(err.Error(), "component_exceeds_allocation") {
		t.Fatalf("unreserved source stroke should exceed allocation: %v", err)
	}
}
