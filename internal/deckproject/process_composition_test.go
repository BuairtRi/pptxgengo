package deckproject

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"math"
	"os"
	"strings"
	"testing"
)

func processTestModel() wmdesign.ProcessSpec {
	return wmdesign.ProcessSpec{Columns: 5, LabelWidth: 90, StepHeight: 48, Lanes: []wmdesign.ProcessLane{{Key: "requester", Label: "Requester"}, {Key: "reviewer", Label: "Reviewer"}}, Steps: []wmdesign.ProcessStep{{Key: "begin", Label: "Start", Lane: "requester", Column: 0, Kind: "start"}, {Key: "check", Label: "Review", Lane: "requester", Column: 1, Kind: "decision"}, {Key: "approve", Label: "Approve", Lane: "requester", Column: 2, Kind: "process"}, {Key: "revise", Label: "Revise", Lane: "reviewer", Column: 2, Kind: "process"}, {Key: "join", Label: "Join", Lane: "requester", Column: 3, Kind: "join"}, {Key: "finish", Label: "Finish", Lane: "requester", Column: 4, Kind: "end"}}, Links: []wmdesign.ProcessLink{{Key: "begin-check", From: "begin", To: "check"}, {Key: "yes", From: "check", To: "approve", Outcome: "Yes"}, {Key: "no", From: "check", To: "revise", Outcome: "No"}, {Key: "approve-join", From: "approve", To: "join"}, {Key: "revise-join", From: "revise", To: "join"}, {Key: "join-end", From: "join", To: "finish"}}, Start: "begin", Current: "check", End: []string{"finish"}}
}
func processFixture(t *testing.T) *Project { p := processFixtureUnpinned(t); pin(t, p); return p }
func processFixtureUnpinned(t *testing.T) *Project {
	t.Helper()
	p := example(t)
	m := processTestModel()
	processTestExplicitOutcomeRoutes(&m)
	var args map[string]any
	_ = json.Unmarshal(canonical(m), &args)
	r := wmdesign.Rect{X: 0, Y: 20, W: 846, H: 300}
	options := wmdesign.FrameRequest{Rail: "none", Footer: "compact", TitleLines: 1, Density: "appendix", Surface: "light"}
	p.Document.LocalTemplates = map[string]LocalTemplate{"process": {Name: "Illustrative process", Frame: Reference{Scope: "shared", ID: "wmds/frame/none-compact"}, FrameOptions: &options, Grid: Reference{Scope: "shared", ID: "wmds/grid/12-columns"}, Zones: map[string]Zone{"title": {Role: "slide-title", Required: true, Schema: map[string]any{"type": "string"}}}, Nodes: []Node{{ID: "flow", Kind: "component", Definition: &Reference{Scope: "shared", ID: "wmds/component/process"}, Placement: &Placement{Zone: "body", Rect: &r}, Arguments: args}}}}
	p.Document.Slides = []Slide{{ID: "flow-slide", ContentKind: "synthetic_example", Template: Reference{Scope: "local", ID: "process"}, Values: map[string]any{"title": "Illustrative process joins two review outcomes"}}}
	if e := os.WriteFile(p.SourcePath, canonical(p.Document), 0600); e != nil {
		t.Fatal(e)
	}
	p, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func processTestPatch(p *Project, ops ...ProcessOperation) ProcessPatch {
	return ProcessPatch{Schema: ProcessPatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "Test", Reason: "Illustrative customization", NodeID: "flow", Operations: ops}
}
func TestProcessCompositionPreviewApplyRelationships(t *testing.T) {
	p := processFixture(t)
	b := append([]byte{}, p.Raw...)
	m, e := InspectProcess(p, "flow-slide", "flow", bundle(t), wmdesign.CandidateEngine)
	if e != nil || m.RenderError != "" {
		t.Fatalf("%v %s", e, m.RenderError)
	}
	model := m.Model
	model.Lanes[0], model.Lanes[1] = model.Lanes[1], model.Lanes[0]
	processTestExplicitOutcomeRoutes(&model)
	yes := model.Links[1]
	patch := processTestPatch(p, ProcessOperation{Action: "set", Entity: "current", Key: "revise"}, ProcessOperation{Action: "reorder", Entity: "lane", Order: []string{"reviewer", "requester"}}, ProcessOperation{Action: "set", Entity: "link", Key: "yes", Link: &yes})
	r, e := PatchProcess(p, "flow-slide", patch, bundle(t), wmdesign.CandidateEngine, false)
	if e != nil {
		t.Fatal(e)
	}
	disk, _ := os.ReadFile(p.SourcePath)
	if r.Applied || !bytes.Equal(b, disk) {
		t.Fatal("preview altered source")
	}
	if _, e = PatchProcess(p, "flow-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	m, e = InspectProcess(p, "flow-slide", "flow", bundle(t), wmdesign.CandidateEngine)
	if e != nil || m.Model.Current != "revise" || m.Model.Lanes[0].Key != "reviewer" || m.Model.Links[2].Outcome != "No" {
		t.Fatalf("lost meaning %v %+v", e, m.Model)
	}
	if _, e = PatchProcess(p, "flow-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e == nil {
		t.Fatal("stale accepted")
	}
}
func TestProcessCompositionCascadeAndInvalid(t *testing.T) {
	m := processTestModel()
	if e := processRemoveStep(&m, "check", false); e == nil {
		t.Fatal("incident removal accepted")
	}
	if e := processRemoveStep(&m, "check", true); e != nil {
		t.Fatal(e)
	}
	if m.Current != "" || len(m.Links) != 3 {
		t.Fatal("cascade incomplete")
	}
	m = processTestModel()
	m.Links[1].Outcome = ""
	if e := wmdesign.ValidateProcess(m); e == nil {
		t.Fatal("unnamed decision")
	}
	m = processTestModel()
	m.Steps[2].Column = 1
	if e := wmdesign.ValidateProcess(m); e == nil {
		t.Fatal("overlap accepted")
	}
	p := processFixture(t)
	label := strings.Repeat("long review description ", 30)
	step := processTestModel().Steps[1]
	step.Label = label
	_, e := PatchProcess(p, "flow-slide", processTestPatch(p, ProcessOperation{Action: "set", Entity: "step", Key: "check", Step: &step}), bundle(t), wmdesign.CandidateEngine, false)
	if e == nil {
		t.Fatal("overflow accepted")
	}
}
func TestProcessCompositionStrictOperationPresence(t *testing.T) {
	p := processFixture(t)
	raw := canonical(processTestPatch(p, ProcessOperation{Action: "set", Entity: "current", Key: "check"}))
	raw = bytes.Replace(raw, []byte(`"key":"check"`), []byte(`"key":"check","cascade":false`), 1)
	if _, e := DecodeProcessPatch(raw, "bad"); e == nil {
		t.Fatal("irrelevant false accepted")
	}
}

func TestProcessCompositionVariablePreDecisionAndUnequalBranches(t *testing.T) {
	p := processFixture(t)
	m := processTestModel()
	m.Columns = 10
	m.Steps[1].Column = 5
	m.Steps[1].Label = "Go?"
	m.Steps[2].Column = 6
	m.Steps[2].Label = "Yes"
	m.Steps[3].Column = 6
	m.Steps[3].Label = "Fix"
	m.Steps[4].Column = 8
	m.Steps[4].Label = "Join"
	m.Steps[5].Column = 9
	m.Steps[5].Label = "End"
	m.Links = nil
	labels := []string{"Plan", "Prep", "Scope", "Check"}
	previous := "begin"
	for i, label := range labels {
		k := fmt.Sprintf("pre-%d", i+1)
		m.Steps = append(m.Steps, wmdesign.ProcessStep{Key: k, Label: label, Lane: "requester", Column: i + 1, Kind: "process"})
		m.Links = append(m.Links, wmdesign.ProcessLink{Key: "to-" + k, From: previous, To: k})
		previous = k
	}
	m.Links = append(m.Links, wmdesign.ProcessLink{Key: "to-check", From: previous, To: "check"}, wmdesign.ProcessLink{Key: "yes", From: "check", To: "approve", Outcome: "Yes"}, wmdesign.ProcessLink{Key: "no", From: "check", To: "revise", Outcome: "No"}, wmdesign.ProcessLink{Key: "approved", From: "approve", To: "join"}, wmdesign.ProcessLink{Key: "join-end", From: "join", To: "finish"})
	m.Steps = append(m.Steps, wmdesign.ProcessStep{Key: "retry", Label: "Retry", Lane: "reviewer", Column: 7, Kind: "process"})
	m.Links = append(m.Links, wmdesign.ProcessLink{Key: "fix-retry", From: "revise", To: "retry"}, wmdesign.ProcessLink{Key: "retry-join", From: "retry", To: "join"})
	ops := []ProcessOperation{{Action: "set", Entity: "layout", Columns: &m.Columns}}
	for i := range m.Steps {
		s := m.Steps[i]
		ops = append(ops, ProcessOperation{Action: "set", Entity: "step", Key: s.Key, Step: &s})
	}
	for _, key := range []string{"begin-check", "approve-join", "revise-join"} {
		ops = append(ops, ProcessOperation{Action: "remove", Entity: "link", Key: key})
	}
	processTestExplicitOutcomeRoutes(&m)
	for i := range m.Links {
		l := m.Links[i]
		ops = append(ops, ProcessOperation{Action: "set", Entity: "link", Key: l.Key, Link: &l})
	}
	if _, e := PatchProcess(p, "flow-slide", processTestPatch(p, ops...), bundle(t), wmdesign.CandidateEngine, false); e != nil {
		t.Fatal(e)
	}
}

// Keep same-lane outcomes on an explicit clear upper detour. Coordinates are
// authored in the component allocation; they are not a semantic inference.
func processTestExplicitOutcomeRoutes(m *wmdesign.ProcessSpec) {
	cw := (846 - m.LabelWidth) / float64(m.Columns)
	w := math.Min(cw-18, 150.)
	steps := map[string]wmdesign.ProcessStep{}
	lanes := map[string]int{}
	for i, l := range m.Lanes {
		lanes[l.Key] = i
	}
	for _, s := range m.Steps {
		steps[s.Key] = s
	}
	for i, l := range m.Links {
		a, z := steps[l.From], steps[l.To]
		if l.Outcome == "" || a.Lane != z.Lane {
			continue
		}
		x1 := m.LabelWidth + float64(a.Column)*cw + (cw+w)/2
		x2 := m.LabelWidth + float64(z.Column)*cw + (cw-w)/2
		y := float64(lanes[a.Lane])*300/float64(len(m.Lanes)) + 25
		m.Links[i].Route = [][2]float64{{x1, y}, {x2, y}}
		m.Links[i].LabelPosition = &[2]float64{(x1 + x2) / 2, y}
	}
}

func TestProcessCompositionOutcomeLabelsRequireMeasuredClearance(t *testing.T) {
	p := processFixture(t)
	before := append([]byte{}, p.Raw...)
	base := processTestModel().Links[1]
	for _, scenario := range []struct {
		name string
		link wmdesign.ProcessLink
	}{
		{"narrow-default-gap", base},
		{"long-outcome", func() wmdesign.ProcessLink {
			l := base
			l.Outcome = "Approve after governance and security review"
			return l
		}()},
		{"explicit-on-node", func() wmdesign.ProcessLink { l := base; l.LabelPosition = &[2]float64{240, 75}; return l }()},
		{"explicit-outside-allocation", func() wmdesign.ProcessLink { l := base; l.LabelPosition = &[2]float64{-30, 25}; return l }()},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			_, err := PatchProcess(p, "flow-slide", processTestPatch(p, ProcessOperation{Action: "set", Entity: "link", Key: "yes", Link: &scenario.link}), bundle(t), wmdesign.CandidateEngine, true)
			if err == nil {
				t.Fatal("unreadable outcome accepted")
			}
			disk, e := os.ReadFile(p.SourcePath)
			if e != nil || !bytes.Equal(before, disk) {
				t.Fatal("refusal changed source", e)
			}
		})
	}
	model := processTestModel()
	processTestExplicitOutcomeRoutes(&model)
	link := model.Links[1]
	if _, err := PatchProcess(p, "flow-slide", processTestPatch(p, ProcessOperation{Action: "set", Entity: "link", Key: "yes", Link: &link}), bundle(t), wmdesign.CandidateEngine, false); err != nil {
		t.Fatal("clear explicit detour rejected", err)
	}
	model.Links[1].LabelPosition = &[2]float64{math.Inf(1), 25}
	if wmdesign.ValidateProcess(model) == nil {
		t.Fatal("nonfinite label anchor accepted")
	}
	model.Links[0].LabelPosition = &[2]float64{200, 25}
	model.Links[1].LabelPosition = nil
	if wmdesign.ValidateProcess(model) == nil {
		t.Fatal("unnamed outcome anchor accepted")
	}
}

func TestProcessCompositionOutcomeLabelAnchorStrictCoordinateCount(t *testing.T) {
	p := processFixture(t)
	link := processTestModel().Links[1]
	link.LabelPosition = &[2]float64{1, 2}
	raw := canonical(processTestPatch(p, ProcessOperation{Action: "set", Entity: "link", Key: "yes", Link: &link}))
	for _, bad := range []string{"[1]", "[1,2,3]", "[null,2]", "null", "[\"1\",2]"} {
		changed := bytes.Replace(raw, []byte(`"label_position":[1,2]`), []byte(`"label_position":`+bad), 1)
		if bytes.Equal(raw, changed) {
			t.Fatal("test did not replace coordinates")
		}
		if _, err := DecodeProcessPatch(changed, "invalid-label-position"); err == nil {
			t.Fatal("invalid coordinates accepted", bad)
		}
	}
	changed := bytes.Replace(raw, []byte(`"label_position":[1,2]`), []byte(`"unknown_label_position":[1,2]`), 1)
	if _, err := DecodeProcessPatch(changed, "unknown-label-position"); err == nil {
		t.Fatal("unknown link field accepted")
	}
	link.Route = [][2]float64{{1, 2}}
	raw = canonical(processTestPatch(p, ProcessOperation{Action: "set", Entity: "link", Key: "yes", Link: &link}))
	for _, bad := range []string{"[[1]]", "[[1,2,3]]", "[[null,2]]", "[null]"} {
		changed = bytes.Replace(raw, []byte(`"route":[[1,2]]`), []byte(`"route":`+bad), 1)
		if _, err := DecodeProcessPatch(changed, "invalid-route-coordinate"); err == nil {
			t.Fatal("invalid waypoint accepted", bad)
		}
	}
}

func TestProcessCompositionRefusesOccludedLogicalLinks(t *testing.T) {
	p := processFixture(t)
	before := append([]byte{}, p.Raw...)
	m := processTestModel()
	m.Columns = 6
	m.Steps[2].Column = 3
	m.Steps[4].Column = 4
	m.Steps[5].Column = 5
	processTestExplicitOutcomeRoutes(&m)
	apply := func(model wmdesign.ProcessSpec) error {
		ops := []ProcessOperation{{Action: "set", Entity: "layout", Columns: &model.Columns}}
		for i := range model.Steps {
			step := model.Steps[i]
			ops = append(ops, ProcessOperation{Action: "set", Entity: "step", Key: step.Key, Step: &step})
		}
		for i := range model.Links {
			link := model.Links[i]
			ops = append(ops, ProcessOperation{Action: "set", Entity: "link", Key: link.Key, Link: &link})
		}
		_, err := PatchProcess(p, "flow-slide", processTestPatch(p, ops...), bundle(t), wmdesign.CandidateEngine, true)
		return err
	}
	if err := apply(m); err == nil || !strings.Contains(err.Error(), "revise-join") || !strings.Contains(err.Error(), "approve") {
		t.Fatal("default path through unrelated Approve accepted", err)
	}
	disk, _ := os.ReadFile(p.SourcePath)
	if !bytes.Equal(before, disk) {
		t.Fatal("refusal changed source")
	}
	m.Links[4].Route = [][2]float64{{531, 225}, {531, 75}}
	if err := apply(m); err == nil {
		t.Fatal("explicit occluded route accepted")
	}
	disk, _ = os.ReadFile(p.SourcePath)
	if !bytes.Equal(before, disk) {
		t.Fatal("explicit route refusal changed source")
	}
	m.Links[4].Route = [][2]float64{{594, 225}, {594, 75}}
	if err := apply(m); err != nil {
		t.Fatal("clear explicit detour refused", err)
	}
	// Exact semantic edges attach at the declared boundary. A waypoint inside
	// the source or target body is not a legitimate attached endpoint.
	p, _ = Load(p.SourcePath)
	before = append([]byte{}, p.Raw...)
	m.Links[4].Route = [][2]float64{{400, 225}, {594, 225}, {594, 75}}
	if err := apply(m); err == nil {
		t.Fatal("path through own source body accepted")
	}
	disk, _ = os.ReadFile(p.SourcePath)
	if !bytes.Equal(before, disk) {
		t.Fatal("source-body refusal changed source")
	}
}
func TestProcessCompositionInitializeAndBindings(t *testing.T) {
	p := processFixtureUnpinned(t)
	local := p.Document.LocalTemplates["process"]
	local.Nodes[0].Definition.ID = "wmds/component/block"
	local.Nodes[0].Arguments = map[string]any{"text": "Existing example", "surface": "light"}
	p.Document.LocalTemplates["process"] = local
	if e := os.WriteFile(p.SourcePath, canonical(p.Document), 0600); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	pin(t, p)
	m := processTestModel()
	processTestExplicitOutcomeRoutes(&m)
	patch := processTestPatch(p, ProcessOperation{Action: "initialize", Entity: "source", Cascade: true, Model: &m})
	if _, e := PatchProcess(p, "flow-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	local = p.Document.LocalTemplates["process"]
	local.Nodes[0].Arguments["start"] = map[string]any{"binding": "start"}
	local.Zones["start"] = Zone{Role: "body", Required: true, Schema: map[string]any{"type": "string"}}
	p.Document.LocalTemplates["process"] = local
	p.Document.Slides[0].Values["start"] = "begin"
	if e := os.WriteFile(p.SourcePath, canonical(p.Document), 0600); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	patch = processTestPatch(p, ProcessOperation{Action: "set", Entity: "current", Key: "revise"})
	if _, e := PatchProcess(p, "flow-slide", patch, bundle(t), wmdesign.CandidateEngine, false); e == nil {
		t.Fatal("bound source accepted")
	}
	patch.Operations = append([]ProcessOperation{{Action: "materialize", Entity: "source"}}, patch.Operations...)
	if _, e := PatchProcess(p, "flow-slide", patch, bundle(t), wmdesign.CandidateEngine, false); e != nil {
		t.Fatal(e)
	}
}

func TestProcessCompositionCapturedSourceGeometryRetained(t *testing.T) {
	p := processFixture(t)
	local := p.Document.LocalTemplates["process"]
	geometry := map[string]any{"schema": wmdesign.SceneSourceGeometrySchema, "x_fraction": 0, "y_fraction": 0, "width_fraction": 1, "height_fraction": 1}
	local.Nodes[0].Arguments[wmdesign.SceneSourceGeometryArgument] = geometry
	p.Document.LocalTemplates["process"] = local
	if e := os.WriteFile(p.SourcePath, canonical(p.Document), 0600); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	if _, e := InspectProcess(p, "flow-slide", "flow", bundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	if _, e := PatchProcess(p, "flow-slide", processTestPatch(p, ProcessOperation{Action: "set", Entity: "current", Key: "approve"}), bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	if !bytes.Equal(canonical(geometry), canonical(p.Document.LocalTemplates["process"].Nodes[0].Arguments[wmdesign.SceneSourceGeometryArgument])) {
		t.Fatal("source geometry lost")
	}
}
