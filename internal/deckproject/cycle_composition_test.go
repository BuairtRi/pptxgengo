package deckproject

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func cycleCompositionFixture(t *testing.T) (*Project, string) {
	t.Helper()
	p := example(t)
	catalog, e := wmdesign.LibraryCatalog(bundle(t), "")
	if e != nil {
		t.Fatal(e)
	}
	var args map[string]any
	for _, def := range catalog {
		if def.Key != "phase/define" {
			continue
		}
		var source struct {
			Body []map[string]any `json:"body"`
		}
		if e = json.Unmarshal(def.RawSlide, &source); e != nil {
			t.Fatal(e)
		}
		for _, n := range source.Body {
			if n["type"] == "cycle" {
				args = n
				break
			}
		}
	}
	if args == nil {
		t.Fatal("catalog phase/define cycle missing")
	}
	for _, k := range []string{"type", "id", "x", "y", "w", "h"} {
		delete(args, k)
	}
	args["center"] = map[string]any{"label": map[string]any{"binding": "center"}, "w": 90}
	options := wmdesign.FrameRequest{Rail: "none", Footer: "compact", TitleLines: 1, Density: "appendix", Surface: "light"}
	rect := wmdesign.Rect{X: 120, Y: 30, W: 600, H: 300}
	local := LocalTemplate{Name: "Catalog cycle operator fixture", Frame: Reference{Scope: "shared", ID: "wmds/frame/none-compact"}, FrameOptions: &options, Grid: Reference{Scope: "shared", ID: "wmds/grid/12-columns"}, Zones: map[string]Zone{"title": {Role: "slide-title", Required: true, Schema: map[string]any{"type": "string"}}, "center": {Role: "body", Required: true, Schema: map[string]any{"type": "string"}}}, Nodes: []Node{{ID: "cycle", Kind: "component", Definition: &Reference{Scope: "shared", ID: "wmds/component/cycle"}, Placement: &Placement{Zone: "body", Rect: &rect}, Arguments: args, Keys: map[string][]string{"items": {"source-001", "source-002", "source-003", "source-004"}}}}}
	p.Document.LocalTemplates = map[string]LocalTemplate{"cycle-plan": local}
	p.Document.Slides = []Slide{{ID: "cycle-slide", ContentKind: "synthetic_example", Template: Reference{Scope: "local", ID: "cycle-plan"}, Values: map[string]any{"title": "Illustrative lifecycle has four stages", "center": "PDLC"}}}
	raw, e := json.Marshal(p.Document)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(p.SourcePath, raw, 0600); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	pin(t, p)
	return p, "cycle"
}
func cyclePatch(p *Project, id string, ops ...CycleOperation) CyclePatch {
	return CyclePatch{Schema: CyclePatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "Cycle qualification", Reason: "Customize synthetic cycle", NodeID: id, Operations: append([]CycleOperation{{Action: "materialize", Entity: "source"}}, ops...)}
}
func TestCycleCompositionCatalogPreviewApplyCountAndRelationships(t *testing.T) {
	p, id := cycleCompositionFixture(t)
	before := append([]byte(nil), p.Raw...)
	inspect, e := InspectCycle(p, "cycle-slide", id, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if inspect.RenderError != "" {
		t.Fatal(inspect.RenderError)
	}
	if len(inspect.Model.Steps) != 4 || inspect.Model.Active != "source-002" || inspect.MutationBlocked == "" {
		t.Fatalf("bad catalog inspection %+v", inspect.Model)
	}
	patch := cyclePatch(p, id, CycleOperation{Action: "set", Entity: "step", Key: "monitor", Step: &CycleStep{Key: "monitor", Label: "Monitor"}}, CycleOperation{Action: "set", Entity: "loop", Key: "feedback", Loop: &CycleLoop{Key: "feedback", From: "monitor", To: "source-002"}}, CycleOperation{Action: "reorder", Entity: "step", Order: []string{"source-001", "monitor", "source-002", "source-003", "source-004"}})
	preview, e := PatchCycle(p, "cycle-slide", patch, bundle(t), wmdesign.CandidateEngine, false)
	if e != nil {
		t.Fatal(e)
	}
	disk, e := os.ReadFile(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(before, disk) || preview.Applied || preview.AfterSHA256 == p.SourceHash() {
		t.Fatal("preview mutates source or misses proposal")
	}
	result, e := PatchCycle(p, "cycle-slide", patch, bundle(t), wmdesign.CandidateEngine, true)
	if e != nil {
		t.Fatal(e)
	}
	if !result.Applied || result.Decision == "" {
		t.Fatal("missing guarded decision")
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	after, e := InspectCycle(p, "cycle-slide", id, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if after.RenderError != "" {
		t.Fatal(after.RenderError)
	}
	if len(after.Model.Steps) != 5 || after.Model.Active != "source-002" || after.Model.Loops[0].From != "monitor" || after.Model.Loops[0].To != "source-002" || after.MutationBlocked != "" {
		t.Fatalf("relationships changed %+v", after.Model)
	}
	n := p.Document.LocalTemplates["cycle-plan"].Nodes[0]
	if n.Arguments["active"] != float64(2) {
		t.Fatal("active marker was not lowered after reorder")
	}
	if _, e = PatchCycle(p, "cycle-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e == nil || !strings.Contains(e.Error(), "hash mismatch") {
		t.Fatalf("stale source accepted %v", e)
	}
	remove := cyclePatch(p, id, CycleOperation{Action: "remove", Entity: "step", Key: "monitor", Cascade: true}, CycleOperation{Action: "remove", Entity: "step", Key: "source-004"})
	if _, e = PatchCycle(p, "cycle-slide", remove, bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	after, e = InspectCycle(p, "cycle-slide", id, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if len(after.Model.Steps) != 3 || len(after.Model.Loops) != 0 || after.Model.Active != "source-002" {
		t.Fatalf("3-step composition failed %+v", after.Model)
	}
	if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
}
func TestCycleCompositionReferenceReorderCascade(t *testing.T) {
	m := CycleModel{Steps: []CycleStep{{Key: "start", Label: "Start"}, {Key: "build", Label: "Build"}, {Key: "review", Label: "Review"}}, Active: "build", Loops: []CycleLoop{{Key: "feedback", From: "review", To: "build"}}}
	if e := applyCycleOperation(&m, CycleOperation{Action: "reorder", Entity: "step", Order: []string{"review", "build", "start"}}); e != nil {
		t.Fatal(e)
	}
	n := Node{Placement: &Placement{Zone: "body", Rect: &wmdesign.Rect{W: 600, H: 300}}}
	if e := lowerCycle(&n, m, cycleSourceSpec{}); e != nil {
		t.Fatal(e)
	}
	m2, _, _, e := cycleSource(&n, nil)
	if e != nil {
		t.Fatal(e)
	}
	if m2.Active != "build" || m2.Loops[0].From != "review" || m2.Loops[0].To != "build" {
		t.Fatal("reorder retargeted relationship")
	}
	if e = applyCycleOperation(&m, CycleOperation{Action: "remove", Entity: "step", Key: "build"}); e == nil {
		t.Fatal("dependent removal accepted")
	}
	if e = applyCycleOperation(&m, CycleOperation{Action: "remove", Entity: "step", Key: "build", Cascade: true}); e != nil {
		t.Fatal(e)
	}
	if m.Active != "" || len(m.Loops) != 0 {
		t.Fatal("cascade did not clear explicit dependents")
	}
	if e = validateCycle(m); e != nil {
		t.Fatal(e)
	}
	for _, model := range []CycleModel{{Steps: m.Steps, Loops: []CycleLoop{{Key: "bad", From: "start", To: "missing"}}}, {Steps: m.Steps, Loops: []CycleLoop{{Key: "bad", From: "start", To: "start"}}}, {Steps: m.Steps, Active: "missing"}, {Steps: []CycleStep{{Key: "one", Label: "One"}}}} {
		if e = validateCycle(model); e == nil {
			t.Fatal("invalid cycle accepted")
		}
	}
	if e = applyCycleOperation(&m, CycleOperation{Action: "reorder", Entity: "step", Order: []string{"start", "start"}}); e == nil {
		t.Fatal("duplicate reorder accepted")
	}
}
func TestCycleCompositionStrictAuthoredFields(t *testing.T) {
	base := CyclePatch{Schema: CyclePatchSchema, ExpectedSourceSHA256: strings.Repeat("a", 64), Actor: "Agent", Reason: "Remove marker", NodeID: "cycle", Operations: []CycleOperation{{Action: "remove", Entity: "active"}}}
	for field, value := range map[string]any{"key": "", "cascade": false, "order": []any{}, "layout": nil, "step": map[string]any{}} {
		var object map[string]any
		if e := json.Unmarshal(canonical(base), &object); e != nil {
			t.Fatal(e)
		}
		object["operations"].([]any)[0].(map[string]any)[field] = value
		if _, e := DecodeCyclePatch(canonical(object), "test"); e == nil {
			t.Fatalf("irrelevant authored %s accepted", field)
		}
	}
	var object map[string]any
	_ = json.Unmarshal(canonical(base), &object)
	object["operations"] = []any{map[string]any{"action": "set", "entity": "step", "key": "new", "step": map[string]any{"key": "new", "label": "New", "unknown": 0}}}
	if _, e := DecodeCyclePatch(canonical(object), "test"); e == nil {
		t.Fatal("unknown nested field accepted")
	}
	if _, e := DecodeCyclePatch(append(canonical(base), []byte("\n---\n{}")...), "test"); e == nil {
		t.Fatal("multiple documents accepted")
	}
}
func TestCycleCompositionRejectsBindingsCapacityAndNativeLayout(t *testing.T) {
	p, id := cycleCompositionFixture(t)
	patch := cyclePatch(p, id, CycleOperation{Action: "set", Entity: "active", Key: "source-003"})
	patch.Operations = patch.Operations[1:]
	if _, e := PatchCycle(p, "cycle-slide", patch, bundle(t), wmdesign.CandidateEngine, false); e == nil || !strings.Contains(e.Error(), "bound arguments") {
		t.Fatalf("implicit materialization accepted %v", e)
	}
	oversized := cyclePatch(p, id, CycleOperation{Action: "set", Entity: "step", Key: "source-001", Step: &CycleStep{Key: "source-001", Label: strings.Repeat("Long label ", 100)}})
	if _, e := PatchCycle(p, "cycle-slide", oversized, bundle(t), wmdesign.CandidateEngine, true); e == nil {
		t.Fatal("overflow silently accepted")
	}
	disk, _ := os.ReadFile(p.SourcePath)
	if !bytes.Equal(disk, p.Raw) {
		t.Fatal("failed measure changed disk")
	}
	bad := cyclePatch(p, id, CycleOperation{Action: "set", Entity: "loop", Key: "bad", Loop: &CycleLoop{Key: "bad", From: "missing", To: "source-001"}})
	if _, e := PatchCycle(p, "cycle-slide", bad, bundle(t), wmdesign.CandidateEngine, true); e == nil {
		t.Fatal("unknown endpoints accepted")
	}
	p.Document.Slides[0].NativeOrder = map[string][]string{"cycle": {"source-001"}}
	if _, e := PatchCycle(p, "cycle-slide", cyclePatch(p, id, CycleOperation{Action: "set", Entity: "active", Key: "source-003"}), bundle(t), wmdesign.CandidateEngine, false); e == nil || !strings.Contains(e.Error(), "reset_native_layout") {
		t.Fatalf("native override overwritten %v", e)
	}
}

func TestCycleCompositionKeysAndCardinality(t *testing.T) {
	n := Node{Arguments: map[string]any{"items": []any{map[string]any{"label": "One"}, map[string]any{"label": "Two"}}}, Keys: map[string][]string{"/items": {"one", "two"}}}
	m, _, _, e := cycleSource(&n, nil)
	if e != nil {
		t.Fatal(e)
	}
	if m.Steps[0].Key != "one" {
		t.Fatal("slash-prefixed authored keys lost")
	}
	n.Keys["items"] = []string{"other", "two"}
	if _, _, _, e = cycleSource(&n, nil); e == nil {
		t.Fatal("conflicting key paths accepted")
	}
	delete(n.Keys, "items")
	n.Keys["loops"] = []string{"old"}
	if _, _, _, e = cycleSource(&n, nil); e == nil {
		t.Fatal("orphan loop keys accepted")
	}
	for _, keys := range [][]string{{"one"}, {"one", "one"}, {"wrong key", "two"}} {
		if _, e = cycleKeys(map[string][]string{"items": keys}, "items", 2); e == nil {
			t.Fatal("invalid key overlay accepted")
		}
	}
	m = CycleModel{}
	for i := 0; i < 13; i++ {
		m.Steps = append(m.Steps, CycleStep{Key: strings.Repeat("x", i+1), Label: "Stage"})
	}
	if e = validateCycle(m); e == nil {
		t.Fatal("13-stage source accepted")
	}
	m.Steps = m.Steps[:2]
	for i := 0; i < 33; i++ {
		m.Loops = append(m.Loops, CycleLoop{Key: strings.Repeat("y", i+1), From: m.Steps[0].Key, To: m.Steps[1].Key})
	}
	if e = validateCycle(m); e == nil {
		t.Fatal("33 loops accepted")
	}
}

func TestCycleCompositionUniqueUnkeyedCommentsFollowValues(t *testing.T) {
	for _, entity := range []string{"step", "loop"} {
		t.Run(entity, func(t *testing.T) {
			p, id := cycleCompositionFixture(t)
			setup := cyclePatch(p, id, CycleOperation{Action: "set", Entity: "loop", Key: "delivery-feedback", Loop: &CycleLoop{Key: "delivery-feedback", From: "source-003", To: "source-002", Label: "Delivery feedback"}}, CycleOperation{Action: "set", Entity: "loop", Key: "optimization-feedback", Loop: &CycleLoop{Key: "optimization-feedback", From: "source-004", To: "source-002", Label: "Optimization feedback"}})
			if _, e := PatchCycle(p, "cycle-slide", setup, bundle(t), wmdesign.CandidateEngine, true); e != nil {
				t.Fatal(e)
			}
			var e error
			p, e = Load(p.SourcePath)
			if e != nil {
				t.Fatal(e)
			}
			doc, e := sourceYAML(p.Raw)
			if e != nil {
				t.Fatal(e)
			}
			arguments := mappingNode(mappingNode(mappingNode(doc.Content[0], "local_templates"), "cycle-plan"), "nodes").Content[0]
			arguments = mappingNode(arguments, "arguments")
			path := "items"
			label := "Identify"
			comment := "Retain this stage context with Identify"
			if entity == "loop" {
				path = "loops"
				label = "Delivery feedback"
				comment = "Retain this feedback rationale with Delivery feedback"
			}
			records := mappingNode(arguments, path)
			records.Content[0].HeadComment = comment
			mappingNode(records.Content[0], "label").LineComment = "Keep nested semantic note"
			raw, e := encodeSourceYAML(doc)
			if e != nil {
				t.Fatal(e)
			}
			if e = os.WriteFile(p.SourcePath, raw, 0600); e != nil {
				t.Fatal(e)
			}
			p, e = Load(p.SourcePath)
			if e != nil {
				t.Fatal(e)
			}
			order := []string{"source-004", "source-003", "source-002", "source-001"}
			if entity == "loop" {
				order = []string{"optimization-feedback", "delivery-feedback"}
			}
			patch := cyclePatch(p, id, CycleOperation{Action: "reorder", Entity: entity, Order: order})
			if _, e = PatchCycle(p, "cycle-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e != nil {
				t.Fatal(e)
			}
			p, e = Load(p.SourcePath)
			if e != nil {
				t.Fatal(e)
			}
			doc, e = sourceYAML(p.Raw)
			if e != nil {
				t.Fatal(e)
			}
			arguments = mappingNode(mappingNode(mappingNode(doc.Content[0], "local_templates"), "cycle-plan"), "nodes").Content[0]
			arguments = mappingNode(arguments, "arguments")
			found := false
			for _, record := range mappingNode(arguments, path).Content {
				if mappingNode(record, "label").Value != label {
					continue
				}
				found = true
				if !strings.Contains(record.HeadComment, comment) || !strings.Contains(mappingNode(record, "label").LineComment, "Keep nested semantic note") {
					t.Fatalf("unique unkeyed %s comments were discarded on reorder: head=%q line=%q", entity, record.HeadComment, mappingNode(record, "label").LineComment)
				}
			}
			if !found {
				t.Fatal("commented semantic record disappeared")
			}
		})
	}
}

func TestCycleCompositionAmbiguousUnkeyedCommentsAreNotReattached(t *testing.T) {
	for _, counts := range [][2]int{{1, 2}, {2, 1}, {2, 2}} {
		oldValues := []any{}
		currentValues := []any{}
		for i := 0; i < counts[0]; i++ {
			oldValues = append(oldValues, map[string]any{"label": "Repeated stage"})
		}
		for i := 0; i < counts[1]; i++ {
			currentValues = append(currentValues, map[string]any{"label": "Repeated stage"})
		}
		old, e := editYAMLNode(oldValues)
		if e != nil {
			t.Fatal(e)
		}
		current, e := editYAMLNode(currentValues)
		if e != nil {
			t.Fatal(e)
		}
		old.Content[0].HeadComment = "This comment belongs to one authored instance"
		mappingNode(old.Content[0], "label").LineComment = "This instance is the selected one"
		preserveDiagramComments(old, current)
		for _, record := range current.Content {
			if record.HeadComment != "" || mappingNode(record, "label").LineComment != "" {
				t.Fatalf("ambiguous comment reattached with old/current counts %v", counts)
			}
		}
	}
}
