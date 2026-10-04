package wmdesign

import (
	"encoding/json"
	"testing"
)

func TestCompactArchitectureLabels(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source.Revision = IntakeRepairRevision
	cases := []struct {
		node     map[string]any
		text     string
		maxLines int
	}{
		{map[string]any{"type": "block", "x": 57, "y": 162, "w": 36, "h": 36, "surface": "strong", "text": "W1", "style": "small"}, "W1", 1},
		{map[string]any{"type": "node", "x": 255, "y": 162, "w": 72, "h": 198, "surface": "inverse", "icon": "connection-link", "text": "API gateway", "layout": "top", "style": "small"}, "API gateway", 2},
		{map[string]any{"type": "chevron", "x": 165, "y": 162, "w": 102, "h": 54, "surface": "strong", "text": "Approve", "style": "small"}, "Approve", 1},
	}
	for _, c := range cases {
		raw, err := json.Marshal(c.node)
		if err != nil {
			t.Fatal(err)
		}
		p, ok, err := r.planDiagramScene("sample", raw, SceneContext{Surface: "light"})
		if err != nil || !ok {
			t.Fatalf("%s: recognized=%v err=%v", c.text, ok, err)
		}
		found := false
		for _, item := range p.Items {
			if item.Text != nil && item.Text.Layout.Original == c.text {
				found = true
				if len(item.Text.Layout.Lines) > c.maxLines {
					t.Fatalf("%s: got %d lines", c.text, len(item.Text.Layout.Lines))
				}
			}
		}
		if !found {
			t.Fatalf("missing %s", c.text)
		}
	}
}
