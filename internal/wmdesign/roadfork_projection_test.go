package wmdesign

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRoadForkProjectionKeepsSemanticContentAndStableArrays(t *testing.T) {
	obj, err := libraryObject(json.RawMessage(`{"type":"roadfork","x":57,"y":126,"w":846,"h":264,"mode":"decision","chosen":0,"small":true,"roadW":18,"forkW":162,"trunk":[{"label":"Foundation","date":"Now","text":"One base","n":7,"here":true,"hereLabel":"Current"}],"fork":{"label":"Choose","text":"One option","at":0.3,"tag":"Approved"},"branches":[{"title":"First","text":"Build","milestones":[{"label":"Start","date":"Next","text":"Deliver","n":"A1"}]},{"title":"Second","milestones":[{"label":"Buy"}]}]}`))
	if err != nil {
		t.Fatal(err)
	}
	def := LibraryTemplate{TemplateDefinition: TemplateDefinition{SourceRevision: LibraryRevisionV10}}
	if !libraryContentWalk(&def, obj, "/body/0", "node01", "", libraryProjectionContext{}) {
		t.Fatal("no roadfork content")
	}
	slots := map[string]string{}
	for _, slot := range def.Slots {
		slots[slot.SourcePointer] = slot.Kind
	}
	for pointer, kind := range map[string]string{
		"/body/0/chosen": "number", "/body/0/trunk/0/n": "number", "/body/0/trunk/0/here": "boolean", "/body/0/trunk/0/hereLabel": "string",
		"/body/0/trunk/0/label": "string", "/body/0/trunk/0/date": "string", "/body/0/trunk/0/text": "string",
		"/body/0/fork/label": "string", "/body/0/fork/text": "string", "/body/0/fork/tag": "string",
		"/body/0/branches/0/title": "string", "/body/0/branches/0/text": "string", "/body/0/branches/0/milestones/0/n": "string",
	} {
		if slots[pointer] != kind {
			t.Errorf("slot %s: got %s want %s", pointer, slots[pointer], kind)
		}
	}
	for _, field := range []string{"at", "roadW", "forkW", "small", "mode"} {
		for path := range slots {
			if strings.HasSuffix(path, "/"+field) {
				t.Errorf("geometry/style projected: %s", path)
			}
		}
	}
	arrays := map[string]int{}
	for _, array := range def.Arrays {
		arrays[array.SourcePointer] = array.Count
	}
	for path, count := range map[string]int{"/body/0/trunk": 1, "/body/0/branches": 2, "/body/0/branches/0/milestones": 1, "/body/0/branches/1/milestones": 1} {
		if arrays[path] != count {
			t.Errorf("array %s: %d want %d", path, arrays[path], count)
		}
	}
}

func TestRoadForkStrictInputsAndRevisionGate(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source.Revision = LibraryRevisionV10
	base := `{"type":"roadfork","x":57,"y":126,"w":846,"h":264,"mode":"decision","trunk":[{"label":"Base"}],"branches":[{"title":"First","milestones":[{"label":"Start"}]},{"title":"Second","milestones":[{"label":"Buy"}]}]}`
	for _, raw := range []string{
		strings.Replace(base, `"x":57`, `"x":null`, 1), strings.Replace(base, `"x":57`, `"x":1e100`, 1), strings.Replace(base, `"y":126`, `"y":1e100`, 1),
		strings.Replace(base, `"h":264`, `"h":264,"_h":2161`, 1), strings.Replace(base, `"mode":"decision"`, `"mode":"decision","chosen":2`, 1),
		strings.Replace(base, `"mode":"decision"`, `"mode":"decision","chosen":0.5`, 1), strings.Replace(base, `"mode":"decision"`, `"mode":"parallel","chosen":0`, 1),
		strings.Replace(base, `"mode":"decision"`, `"mode":"decision","roadW":0`, 1), strings.Replace(base, `"label":"Base"`, `"label":"`+strings.Repeat("X", 2049)+`"`, 1),
		strings.Replace(base, `"label":"Base"`, `"label":"Base","here":true,"hereLabel":"`+strings.Repeat("M", 80)+`"`, 1),
		strings.Replace(base, `"type":"roadfork"`, `"type":"roadfork","unknown":true`, 1),
	} {
		if _, handled, err := r.planRoadForkScene("bad", json.RawMessage(raw), SceneContext{Surface: "light"}); !handled || err == nil {
			t.Errorf("accepted invalid input %s", raw[:min(len(raw), 100)])
		}
	}
	r.source.Revision = LibraryRevisionV9
	if _, handled, err := r.planRoadForkScene("old", json.RawMessage(base), SceneContext{Surface: "light"}); !handled || err == nil {
		t.Fatal("old revision accepted new node")
	}
}
