package wmdesign

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestIntakeGaugeClosedBindingPreservesContentAndPalette(t *testing.T) {
	raw := json.RawMessage(`{"type":"slide","rail":"none","footer":"compact","body":[` + frozenGaugeNode + `]}`)
	obj, err := libraryObject(raw)
	if err != nil {
		t.Fatal(err)
	}
	def := LibraryTemplate{TemplateDefinition: TemplateDefinition{Key: "test/gauge", SourceRevision: IntakeRepairRevision}, RawSlide: raw}
	libraryContentWalk(&def, obj["body"].([]any)[0], "/body/0", "dial", "", libraryProjectionContext{})
	wanted := map[string]string{"/body/0/value": "0.75", "/body/0/valueText": `"3.0"`, "/body/0/caption": `"Readiness[^1]"`, "/body/0/ends/0": `"Lower"`, "/body/0/ends/1": `"Higher"`}
	values := LibraryValues{Slots: map[string]json.RawMessage{}, Keys: map[string][]string{}}
	for _, slot := range def.Slots {
		value, ok := wanted[slot.SourcePointer]
		if !ok {
			t.Fatalf("unexpected gauge editable slot %s", slot.SourcePointer)
		}
		values.Slots[slot.Name] = json.RawMessage(value)
		delete(wanted, slot.SourcePointer)
	}
	if len(wanted) != 0 {
		t.Fatalf("missing gauge slots: %+v", wanted)
	}
	for _, array := range def.Arrays {
		if array.SourcePointer != "/body/0/ends" || array.Count != 2 {
			t.Fatalf("palette or geometry became keyed content: %+v", array)
		}
		values.Keys[array.Name] = []string{"low", "high"}
	}
	if len(def.Arrays) != 1 {
		t.Fatalf("endpoint identity missing: %+v", def.Arrays)
	}
	body, _ := json.Marshal(values)
	slide, record, err := bindLibraryTemplate(def, BoundSlide{ID: "supplied-dial", Template: def.Key, ContentKind: "supplied_content", Values: body})
	if err != nil {
		t.Fatal(err)
	}
	if len(record.Assignments) != 5 || len(slide.Nodes) != 1 || slide.Nodes[0].Scene == nil {
		t.Fatalf("missing authored content mappings: %+v", record)
	}
	var gauge intakeGaugeSource
	if err := json.Unmarshal(slide.Nodes[0].Scene.Node, &gauge); err != nil {
		t.Fatal(err)
	}
	if gauge.Value == nil || *gauge.Value != .75 || gauge.ValueText != "3.0" || gauge.Caption != "Readiness[^1]" || strings.Join(gauge.Ends, "/") != "Lower/Higher" || strings.Join(gauge.Segments, "/") != "kpi.off/kpi.off/kpi.risk/kpi.risk/kpi.on/kpi.on" {
		t.Fatalf("binding altered content/palette: %+v", gauge)
	}
	r := intakeTestRenderer(t)
	p := gaugePlan(t, r, string(slide.Nodes[0].Scene.Node), SceneContext{Surface: "light", Notes: []string{"Survey evidence"}})
	seenFootnote := false
	for _, item := range p.Items {
		if item.Text != nil && strings.HasSuffix(item.Text.ID, ".value.part-2") {
			seenFootnote = item.Text.Rich != nil && strings.Contains(item.Text.Layout.Displayed, "Readiness")
		}
	}
	if !seenFootnote {
		t.Fatal("bound caption footnote lost")
	}
}

func TestIntakeGaugeLocalCompositionHasDerivedHeight(t *testing.T) {
	var args map[string]any
	if err := json.Unmarshal([]byte(frozenGaugeNode), &args); err != nil {
		t.Fatal(err)
	}
	delete(args, "type")
	delete(args, "x")
	delete(args, "y")
	delete(args, "w")
	raw, err := ComposeSceneNode("gauge", args, Rect{100, 120, 216, 216})
	if err != nil {
		t.Fatal(err)
	}
	var node map[string]json.RawMessage
	if err := json.Unmarshal(raw, &node); err != nil {
		t.Fatal(err)
	}
	if _, exists := node["h"]; exists {
		t.Fatal("composition injected unsupported h into derived-height gauge")
	}
	r := intakeTestRenderer(t)
	p := gaugePlan(t, r, string(raw), SceneContext{Surface: "light"})
	if p.Bounds.Y+p.Bounds.H > 336 {
		t.Fatalf("gauge escaped sufficient component allocation: %+v", p.Bounds)
	}
	args["h"] = 216
	raw, err = ComposeSceneNode("gauge", args, Rect{100, 120, 216, 216})
	if err != nil {
		t.Fatal(err)
	}
	if _, handled, err := r.planIntakeGaugeScene("gauge", raw, SceneContext{Surface: "light"}); !handled || err == nil {
		t.Fatal("unknown caller field bypassed strict gauge planner")
	}
}
