package wmdesign

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalTeamCurveOwnsCompleteStrokeEnvelope(t *testing.T) {
	allocation := Rect{100, 180, 700, 200}
	args := map[string]any{"max": 1, "curve": "monotone", "phaseLabels": false, "phaseH": 0, "series": []any{
		map[string]any{"name": "Original qualitative profile", "values": []float64{0, 1, 1, 0}, "style": "line", "dashed": true, "labelAt": false},
	}}
	raw, err := ComposeSceneNode("teamcurve", args, allocation)
	if err != nil {
		t.Fatal(err)
	}
	path := "/local_templates/original-profile/nodes/curve"
	doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{{ID: "local-curve", Frame: FrameRequest{NoHeader: true}, Nodes: []Node{{ID: "profile", Kind: "scene", Scene: &SceneSpec{Node: raw, Path: path, Allocation: &allocation, Keys: map[string][]string{path + "/series": {"original"}}}}}}}}
	bundle := filepath.Join("..", "..", "library", "wm-design-system", "v5")
	_, report, err := BuildWithEngine(bundle, "", doc, CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, shape := range report.Slides[0].Shapes {
		if strings.HasPrefix(shape.ID, "profile.") {
			count++
			if !inside(shape.Rect, allocation) {
				t.Fatalf("native ink exceeds local allocation: %+v", shape.Rect)
			}
		}
	}
	if count == 0 {
		t.Fatal("no editable native curve")
	}
	args["type"], args["x"], args["y"], args["w"], args["h"] = "teamcurve", allocation.X, allocation.Y, allocation.W, allocation.H
	raw, err = json.Marshal(args)
	if err != nil {
		t.Fatal(err)
	}
	doc.Slides[0].Nodes[0].Scene.Node = raw
	if _, _, err := BuildWithEngine(bundle, "", doc, CandidateEngine); err == nil || !strings.Contains(err.Error(), "component_exceeds_allocation") {
		t.Fatalf("unreserved stroke must remain rejected: %v", err)
	}
	if _, err := ComposeSceneNode("teamcurve", nil, Rect{0, 0, 2.5, 200}); err == nil {
		t.Fatal("stroke-only allocation accepted")
	}
}
