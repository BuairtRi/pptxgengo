package wmdesign

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestQuantitativeAxesNoClipMissingAndLongEndLabelBounds(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source.Revision = LibraryRevisionV5
	raw := json.RawMessage(`{"type":"chart","kind":"column","x":57,"y":126,"w":558,"h":252,"categories":["A","B","C"],"series":[{"name":"Capacity","values":[0,null,3]}],"allowMissing":true,"preserveWorkbookZeros":true,"yMin":0,"yMax":4}`)
	p, _, e := r.planChartScene("chart", raw, SceneContext{Surface: "light"})
	if e != nil {
		t.Fatal(e)
	}
	found := false
	for _, it := range p.Items {
		if it.Chart != nil {
			found = true
			if *it.Chart.Options.ValAxisMaxVal != 4 || !it.Chart.Data[0].MissingValues[1] || it.Chart.Data[0].Values[0] != 0 || it.Chart.Options.DisplayBlanksAs != "gap" {
				t.Fatal("axes/missing semantics lost")
			}
		}
	}
	if !found {
		t.Fatal("native chart missing")
	}
	for _, value := range []string{"2", "-1"} {
		changed := strings.Replace(string(raw), `"yMax":4`, `"yMax":`+value, 1)
		if _, _, e = r.planChartScene("chart", json.RawMessage(changed), SceneContext{Surface: "light"}); e == nil {
			t.Fatal("clipped data accepted")
		}
	}
	line := json.RawMessage(`{"type":"chart","kind":"line","x":57,"y":126,"w":558,"h":252,"categories":["A","B"],"series":[{"name":"Very long series one","values":[3,0]},{"name":"Another long series two","values":[4,0]}],"yMin":0,"yMax":4}`)
	p, _, e = r.planChartScene("chart", line, SceneContext{Surface: "light"})
	if e != nil {
		t.Fatal(e)
	}
	for _, it := range p.Items {
		if it.Text != nil && strings.Contains(it.Text.ID, ".direct-") {
			if it.Text.Rect.Y < 126 || it.Text.Rect.Y+it.Text.Rect.H > 378.02 {
				t.Fatalf("direct label outside chart %+v", it.Text.Rect)
			}
		}
	}
}

func TestQuantitativeAutoUpdateWorkbookTypedOption(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source.Revision = LibraryRevisionV5
	base := `{"type":"chart","kind":"column","x":57,"y":126,"w":558,"h":252,"categories":["A","B"],"series":[{"name":"Capacity","values":[0,3]}],"autoUpdateWorkbook":VALUE}`
	for _, value := range []string{"true", "false", "null", "1", `"true"`} {
		p, _, e := r.planChartScene("chart", json.RawMessage(strings.Replace(base, "VALUE", value, 1)), SceneContext{Surface: "light"})
		if value != "true" && value != "false" {
			if e == nil {
				t.Fatalf("nonboolean %s accepted", value)
			}
			continue
		}
		if e != nil {
			t.Fatal(e)
		}
		found := false
		for _, it := range p.Items {
			if it.Chart != nil {
				found = true
				if it.Chart.Options.AutoUpdateWorkbook != (value == "true") {
					t.Fatal("autoUpdate flag lost")
				}
			}
		}
		if !found {
			t.Fatal("native chart missing")
		}
	}
}
