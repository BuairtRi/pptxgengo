package wmdesign

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestTableStatusSourceLabelPrecedence(t *testing.T) {
	labels := map[string]string{"on": "Received", "risk": "Pending", "off": "Overdue"}
	for _, example := range []struct{ raw, status, label string }{
		{`"on"`, "on", "Received"},
		{` {"status":"on","label":"Confirmed"} `, "on", "Confirmed"},
		{`{"status":"risk","label":""}`, "risk", "Pending"},
		{`{"status":"done"}`, "done", "Done"},
		{`"blocked"`, "blocked", "Blocked"},
	} {
		status, label, err := sceneTableStatusValue(json.RawMessage(example.raw), labels)
		if err != nil || status != example.status || label != example.label {
			t.Errorf("%s: status=%q label=%q error=%v", example.raw, status, label, err)
		}
	}
	status, label, err := sceneTableStatusValue(json.RawMessage(`"on"`), map[string]string{"on": ""})
	if err != nil || status != "on" || label != "On track" {
		t.Fatal("empty map label must preserve source fallback", status, label, err)
	}
}

func TestTableStatusSourceLabelsNativeTextAndMarks(t *testing.T) {
	r := intakeTestRenderer(t)
	p := enhancementPlan(t, r, `{"type":"table","x":117,"y":126,"w":480,"dense":true,"cols":[{"k":"a","label":"Request","w":240,"type":"status","labels":{"on":"Received"}},{"k":"b","label":"Owner","w":240,"type":"status","labels":{"on":"Received"}}],"rows":[{"a":"on","b":{"status":"on","label":"Confirmed"}}]}`)
	tab := enhancementTable(t, p)
	if tab.Rows[1][0].Text != "Received" || tab.Rows[1][1].Text != "Confirmed" {
		t.Fatal("visible labels not retained in native cells", tab.Rows[1])
	}
	a := enhancementShape(t, p, ".a.status-mark")
	b := enhancementShape(t, p, ".b.status-mark")
	if a.Record.Color != b.Record.Color || a.Record.Geometry != b.Record.Geometry {
		t.Fatal("custom label changed canonical status mark")
	}
}

func TestTableStatusOptionsStrictlyRejected(t *testing.T) {
	for _, raw := range []string{`42`, `true`, `[]`, `null`, `{"status":"other"}`, `{"label":"Ready"}`, `{"status":"on","label":42}`, `{"status":"on","Label":"Ready"}`, `{"status":"on","unknown":true}`, `{"status":"on","status":"off"}`} {
		if _, _, err := sceneTableStatusValue(json.RawMessage(raw), nil); err == nil {
			t.Fatalf("invalid status union accepted: %s", raw)
		}
	}
	r := intakeTestRenderer(t)
	for _, column := range []string{
		`{"k":"a","label":"A","w":240,"labels":{"on":"Ready"}}`,
		`{"k":"a","label":"A","w":240,"type":"status","labels":{"unknown":"Ready"}}`,
		`{"k":"a","label":"A","w":240,"type":"status","labels":{"on":42}}`,
		`{"k":"a","label":"A","w":240,"type":"status","Labels":{"on":"Ready"}}`,
		`{"k":"a","label":"A","w":240,"type":"status","labels":{"on":"Ready","on":"Done"}}`,
	} {
		raw := `{"type":"table","x":117,"y":126,"w":240,"cols":[` + column + `],"rows":[{"a":"on"}]}`
		if _, err := r.planSceneNode("sample", json.RawMessage(raw), SceneContext{Surface: "light"}); err == nil {
			t.Fatalf("invalid column labels accepted: %s", column)
		}
	}
}

func TestTableStatusExistingRevisionPlansUnchanged(t *testing.T) {
	r := intakeTestRenderer(t)
	base := `{"type":"table","x":117,"y":126,"w":240,"cols":[{"k":"a","label":"Request","w":240,"type":"status"}],"rows":[{"a":"on"},{"a":"risk"}]}`
	for _, revision := range []string{LibraryRevisionV1, LibraryRevisionV2, LibraryRevisionV3, LibraryRevisionV4, LibraryRevisionV5} {
		r.source.Revision = revision
		original := enhancementPlan(t, r, base)
		explicitEmpty := enhancementPlan(t, r, strings.Replace(base, `"type":"status"`, `"type":"status","labels":{}`, 1))
		if !reflect.DeepEqual(original, explicitEmpty) {
			t.Fatalf("empty label override changed the %s existing scalar plan", revision)
		}
	}
}

func TestTableV5NoneHeaderMatchesFrozenLightHeader(t *testing.T) {
	r := intakeTestRenderer(t)
	raw := `{"type":"table","x":117,"y":126,"w":240,"header":"none","dense":true,"cols":[{"k":"a","label":"Follow-up","w":240}],"rows":[{"a":"Denial extract"}]}`
	for _, revision := range []string{LibraryRevisionV1, LibraryRevisionV2, LibraryRevisionV3, LibraryRevisionV4} {
		r.source.Revision = revision
		if _, err := r.planSceneNode("sample", json.RawMessage(raw), SceneContext{Surface: "light"}); err == nil {
			t.Fatalf("none header expanded prior %s contract", revision)
		}
	}
	r.source.Revision = LibraryRevisionV5
	none := enhancementPlan(t, r, raw)
	light := enhancementPlan(t, r, strings.Replace(raw, `"header":"none"`, `"header":"light"`, 1))
	if !reflect.DeepEqual(none, light) {
		t.Fatal("v5 none header differs from the frozen source renderer's light header")
	}
	if tab := enhancementTable(t, none); len(tab.Rows) != 2 || tab.Rows[0][0].Text != "FOLLOW-UP" {
		t.Fatal("source header row was lost", tab.Rows)
	}
}
