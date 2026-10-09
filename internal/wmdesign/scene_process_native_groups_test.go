package wmdesign

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestProcessNativeStepGroupsOwnLabelsMarkersAndAttachedEndpoints(t *testing.T) {
	r := intakeTestRenderer(t)
	s := ProcessSpec{Type: "process", X: 57, Y: 130, W: 846, H: 280, Columns: 3, LabelWidth: 80, StepHeight: 50,
		Lanes: []ProcessLane{{Key: "delivery", Label: "Delivery"}, {Key: "review", Label: "Review"}},
		Steps: []ProcessStep{{Key: "start", Label: "Start", Lane: "delivery", Kind: "start", Column: 0}, {Key: "decide", Label: "Approve?", Lane: "delivery", Kind: "decision", Column: 1}, {Key: "deliver", Label: "Deliver", Lane: "delivery", Kind: "end", Column: 2}, {Key: "revise", Label: "Revise", Lane: "review", Kind: "process", Column: 2}},
		Links: []ProcessLink{{Key: "start-decision", From: "start", To: "decide"}, {Key: "approved", From: "decide", To: "deliver", Outcome: "Yes"}, {Key: "revise", From: "decide", To: "revise", Outcome: "No"}},
		Start: "start", Current: "decide", End: []string{"deliver"}}
	raw, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	p, handled, err := r.planProcessScene("workflow", raw, SceneContext{Surface: "light"})
	if err != nil || !handled {
		t.Fatalf("handled=%t err=%v", handled, err)
	}
	groups := map[string]ComponentRecord{}
	for _, g := range p.Groups {
		groups[g.ID] = g
	}
	for _, step := range s.Steps {
		id := "workflow.steps." + step.Key
		g, ok := groups[id]
		if !ok || g.Definition != "process.step" {
			t.Fatalf("missing owned step group %s", id)
		}
		parts := map[string]bool{}
		for _, part := range g.Parts {
			parts[part] = true
		}
		if !parts[id+".surface"] || !parts[id+".label"] {
			t.Fatalf("step surface/label split: %+v", g)
		}
		marked := step.Key != "revise"
		if parts[id+".marker"] != marked {
			t.Fatalf("step marker ownership %s", id)
		}
		wantHeight := s.StepHeight
		if marked {
			wantHeight += 17
		}
		if g.Rect.H != wantHeight {
			t.Fatalf("group does not enclose marker: %+v", g.Rect)
		}
	}
	attached := 0
	for _, item := range p.Items {
		if item.Shape == nil || item.Shape.Connection == nil {
			continue
		}
		c := item.Shape.Connection
		if !strings.HasPrefix(c.Begin.ObjectName, "workflow.steps.") || !strings.HasSuffix(c.Begin.ObjectName, ".surface") || !strings.HasSuffix(c.End.ObjectName, ".surface") {
			t.Fatalf("relationship detached from stable step surface: %+v", c)
		}
		attached++
	}
	if attached != len(s.Links) {
		t.Fatalf("attached links %d, want %d", attached, len(s.Links))
	}
}
