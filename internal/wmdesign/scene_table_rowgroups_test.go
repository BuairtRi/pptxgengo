package wmdesign

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestTableV6RowGroupsAndHeights(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source.Revision = LibraryRevisionV6
	p := enhancementPlan(t, r, `{"type":"table","x":57,"y":126,"w":390,"rowH":36,"rowGroups":[{"label":"Care","from":0,"to":1}],"cols":[{"k":"name","label":"Name","w":180},{"k":"score","label":"Score","w":180,"type":"heat","showValue":true}],"rows":[{"name":"First","score":0,"h":48},{"name":"Second","score":4,"h":60}]}`)
	tab := enhancementTable(t, p)
	if tab.Rect != (Rect{87, 126, 360, 144}) || !reflect.DeepEqual(tab.Options.RowH, []float64{.5, 48. / 72, 60. / 72}) {
		t.Fatalf("table geometry %+v rowH%v", tab.Rect, tab.Options.RowH)
	}
	rail := enhancementShape(t, p, ".row-group.source-001.fill")
	if rail.Record.Rect != (Rect{57, 163.5, 24, 105}) {
		t.Fatalf("rail geometry %+v", rail.Record.Rect)
	}
	found := false
	for _, it := range p.Items {
		if it.Text != nil && strings.HasSuffix(it.Text.ID, ".row-group.source-001.label") {
			found = true
			if it.Text.Layout.Displayed != "CARE" || it.Text.Rotation != -90 || it.Text.Color != "FFFFFF" || it.Text.Layout.Style.Size != 8 {
				t.Fatalf("group label %+v", it.Text)
			}
		}
	}
	if !found {
		t.Fatal("editable group label missing")
	}
	if p.Bounds.X > 57 || p.Bounds.X+p.Bounds.W < 447 {
		t.Fatalf("rail outside group bounds %+v", p.Bounds)
	}
	for _, tc := range tab.CellRecords {
		if tc.Row == 2 && tc.Text.Rect.Y != 210 {
			t.Fatalf("variable row starts %+v", tc.Text.Rect)
		}
	}
}

func TestTableV6RowGroupValidation(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source.Revision = LibraryRevisionV6
	base := `{"type":"table","x":57,"y":126,"w":210,"cols":[{"k":"name","label":"Name","w":180}],"rows":[{"name":"First"},{"name":"Second"}],%s}`
	for _, fields := range []string{
		`"rowGroups":[{"label":"A","from":-1,"to":0}]`,
		`"rowGroups":[{"label":"A","from":0,"to":2}]`,
		`"rowGroups":[{"label":"A","from":1,"to":0}]`,
		`"rowGroups":[{"label":"A","from":0.5,"to":1}]`,
		`"rowGroups":[{"label":"A","to":1}]`,
		`"rowGroups":[{"label":"A","from":0,"to":null}]`,
		`"rowGroups":[{"label":"A","from":0,"to":1},{"label":"B","from":1,"to":1}]`,
		`"rowGroups":[{"label":" ","from":0,"to":1}]`,
		`"groupW":-1,"rowGroups":[{"label":"A","from":0,"to":1}]`,
		`"groupW":205,"rowGroups":[{"label":"A","from":0,"to":1}]`,
		`"groupW":24`,
	} {
		raw := strings.Replace(base, "%s", fields, 1)
		if _, err := r.planSceneNode("invalid", json.RawMessage(raw), SceneContext{Surface: "light"}); err == nil {
			t.Errorf("accepted %s", fields)
		}
	}
	for _, h := range []string{"0", "-1", "null", `"48"`} {
		raw := `{"type":"table","x":57,"y":126,"w":180,"cols":[{"k":"name","label":"Name","w":180}],"rows":[{"name":"First","h":` + h + `}]}`
		if _, err := r.planSceneNode("invalid", json.RawMessage(raw), SceneContext{Surface: "light"}); err == nil {
			t.Errorf("accepted row h%s", h)
		}
	}
}

func TestTableV6GroupWidthCoordinatesAndColor(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source.Revision = LibraryRevisionV6
	p := enhancementPlan(t, r, `{"type":"table","x":57,"y":126,"w":216,"groupW":30,"rowGroups":[{"label":"Care","from":0,"to":1,"fill":"#CED7E6"}],"cols":[{"k":"name","label":"Name","w":180}],"rows":[{"name":"First"},{"name":"Second"}]}`)
	if enhancementTable(t, p).Rect.X != 93 || enhancementShape(t, p, ".row-group.source-001.fill").Record.Rect.W != 30 {
		t.Fatal("authored group width not retained")
	}
	for _, it := range p.Items {
		if it.Text != nil && strings.HasSuffix(it.Text.ID, ".row-group.source-001.label") && it.Text.Color != "070154" {
			t.Fatal("pale group contrast")
		}
	}
}

func TestTableV6RowGroupStableIdentity(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source.Revision = LibraryRevisionV6
	raw := json.RawMessage(`{"type":"table","x":57,"y":126,"w":210,"rowGroups":[{"label":"Care","from":0,"to":1}],"cols":[{"k":"name","label":"Name","w":180}],"rows":[{"name":"First"},{"name":"Second"}]}`)
	ctx := SceneContext{Surface: "light", Path: "/body/0", Keys: map[string][]string{"/body/0/rows": {"first", "second"}, "/body/0/rowGroups": {"patient-care"}}}
	p, err := r.planSceneNode("stable", raw, ctx)
	if err != nil {
		t.Fatal(err)
	}
	if enhancementShape(t, p, ".row-group.patient-care.fill").Record.Rect.X != 57 {
		t.Fatal("stable group geometry")
	}
	delete(ctx.Keys, "/body/0/rowGroups")
	if _, err = r.planSceneNode("missing", raw, ctx); err == nil {
		t.Fatal("bound group missing keys accepted")
	}
}

func TestTableV5GeometryUnchangedAndV6OptionsGated(t *testing.T) {
	r := intakeTestRenderer(t)
	raw := `{"type":"table","x":57,"y":126,"w":360,"header":"none","dense":true,"rowH":30,"cols":[{"k":"name","label":"Name","w":180},{"k":"score","label":"Score","w":180,"type":"heat","showValue":true}],"rows":[{"name":"First","score":0},{"name":"Second","score":4}]}`
	r.source.Revision = LibraryRevisionV5
	a := enhancementPlan(t, r, raw)
	r.source.Revision = LibraryRevisionV6
	b := enhancementPlan(t, r, raw)
	if !reflect.DeepEqual(a, b) {
		t.Fatal("new revision changed legacy table plan")
	}
	r.source.Revision = LibraryRevisionV5
	for _, field := range []string{`"heatMin":1,`, `"rowGroups":[],`, `"groupW":24,`, `"heatMin":null,`, `"rowGroups":null,`} {
		if _, err := r.planSceneNode("old", json.RawMessage(strings.Replace(raw, `"type":"table",`, `"type":"table",`+field, 1)), SceneContext{Surface: "light"}); err == nil {
			t.Errorf("old revision accepted %s", field)
		}
	}
	if sceneTableAlign("priority") != "center" {
		t.Fatal("priority header alignment")
	}
	if _, err := sceneTableRowHeights(sceneTableSource{Rows: []map[string]json.RawMessage{{"h": json.RawMessage("48")}}}, 36, LibraryRevisionV5); err == nil {
		t.Fatal("old row height accepted")
	}
	for _, revision := range []string{LibraryRevisionV5, LibraryRevisionV6} {
		r.source.Revision = revision
		legacy := enhancementPlan(t, r, `{"type":"table","x":57,"y":126,"w":180,"cols":[{"k":"h","label":"Height","w":180}],"rows":[{"h":"Tall"}]}`)
		if enhancementTable(t, legacy).Rows[1][0].Text != "Tall" || enhancementTable(t, legacy).Options.RowH[1] != .5 {
			t.Fatal("legacy h column reinterpreted", revision)
		}
	}
	n := sceneTableSource{W: 100, Rows: []map[string]json.RawMessage{{}}, RowGroups: []sceneTableRowGroup{{Label: "A", From: 0, To: 0}}, GroupW: func() *float64 { v := math.Inf(1); return &v }()}
	if _, _, err := sceneTableRowGroupGeometry(n); err == nil {
		t.Fatal("nonfinite rail width")
	}
}
