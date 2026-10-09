package wmdesign

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestGanttNativeItemsKeepBarsMarkersAndCaptionsTogether(t *testing.T) {
	r := intakeTestRenderer(t)
	raw := json.RawMessage(`{"type":"gantt","x":57,"y":130,"w":846,"cols":{"group":24,"lane":180},"periods":{"labels":["1","2","3","4","5","6","7","8","9","10","11","12"]},"kinds":{"task":{"label":"Task","fill":"ramp.100","text":"strong"}},"groups":[{"key":"delivery","label":"IT","fill":"strong","lanes":[{"key":"build","icon":"none","title":"Build","items":[{"key":"api","kind":"task","label":"API","from":1,"to":3},{"key":"data","kind":"task","label":"Data","from":3,"to":5,"progress":0.5},{"key":"review","tag":true,"label":"Review","at":6},{"key":"signoff","milestone":true,"label":"Signoff","at":10,"labelSide":"left"}]}]}]}`)
	p, err := r.planGanttScene("schedule", raw, SceneContext{Surface: "light"})
	if err != nil {
		t.Fatal(err)
	}
	groups := map[string]ComponentRecord{}
	for _, g := range p.Groups {
		groups[g.ID] = g
	}
	for _, key := range []string{"api", "data", "review", "signoff"} {
		id := "schedule.groups.delivery.lanes.build.items." + key
		g, ok := groups[id]
		if !ok || g.Definition != "gantt.item" {
			t.Fatalf("missing task group %s", id)
		}
		label, graphic := false, false
		for _, part := range g.Parts {
			if part == id+".label" {
				label = true
			}
			if strings.HasPrefix(part, id+".") && part != id+".label" {
				graphic = true
			}
		}
		if !label || !graphic || g.Rect.W <= 0 || g.Rect.H < 18 {
			t.Fatalf("task content split or invalid bounds: %+v", g)
		}
	}
	root := groups["schedule"]
	for _, part := range root.Parts {
		if strings.Contains(part, ".items.") && strings.HasSuffix(part, ".label") {
			t.Fatalf("caption escaped its task group: %s", part)
		}
	}
}
