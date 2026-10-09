package deckproject

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func ganttCompositionFixture(t *testing.T) (*Project, string) {
	t.Helper()
	p := example(t)
	catalog, e := wmdesign.LibraryCatalog(bundle(t), "")
	if e != nil {
		t.Fatal(e)
	}
	var source struct {
		Body []map[string]any `json:"body"`
	}
	for _, def := range catalog {
		if def.Key == "plan/gantt" {
			if e = json.Unmarshal(def.RawSlide, &source); e != nil {
				t.Fatal(e)
			}
		}
	}
	if len(source.Body) != 1 {
		t.Fatal("catalog Gantt specimen missing")
	}
	var schedule wmdesign.GanttSpec
	if e = json.Unmarshal(canonical(source.Body[0]), &schedule); e != nil {
		t.Fatal(e)
	}
	// The source specimen's optional icon graphics are omitted so this test is
	// portable without the private graphical asset package; schedule/style remain.
	for gi := range schedule.Groups {
		// This composition selects one illustrative lane per workstream. The
		// full five-lane stock example uses a header-adjacent legend exception;
		// that is not a valid allocation for a local body component.
		if len(schedule.Groups[gi].Lanes) > 1 {
			schedule.Groups[gi].Lanes = schedule.Groups[gi].Lanes[:1]
		}
		for li := range schedule.Groups[gi].Lanes {
			schedule.Groups[gi].Lanes[li].Icon = "none"
			if len(schedule.Groups[gi].Lanes[li].Items) > 1 {
				schedule.Groups[gi].Lanes[li].Items = schedule.Groups[gi].Lanes[li].Items[:1]
			}
		}
	}
	keys, e := normalizeGanttKeys(&schedule, nil)
	if e != nil {
		t.Fatal(e)
	}
	var args map[string]any
	if e = json.Unmarshal(canonical(schedule), &args); e != nil {
		t.Fatal(e)
	}
	for _, k := range []string{"type", "x", "y", "w"} {
		delete(args, k)
	}
	args["sidebarLabel"] = map[string]any{"binding": "sidebar"}
	options := wmdesign.FrameRequest{Rail: "none", Footer: "compact", TitleLines: 1, Density: "appendix", Surface: "light"}
	foundation, e := wmdesign.Load(bundle(t), "")
	if e != nil {
		t.Fatal(e)
	}
	frame, e := foundation.ResolveFrame(options)
	if e != nil {
		t.Fatal(e)
	}
	rect := wmdesign.Rect{X: schedule.X - frame.Body.X, Y: 0, W: schedule.W, H: frame.Body.H}
	args["trackPitch"] = 48
	args["legendFullWidth"] = true
	local := LocalTemplate{Name: "Catalog work plan fixture (optional icons omitted)", Frame: Reference{Scope: "shared", ID: "wmds/frame/none-compact"}, FrameOptions: &options, Grid: Reference{Scope: "shared", ID: "wmds/grid/12-columns"}, Zones: map[string]Zone{"title": {Role: "slide-title", Required: true, Schema: map[string]any{"type": "string"}}, "sidebar": {Role: "body", Required: true, Schema: map[string]any{"type": "string"}}}, Nodes: []Node{{ID: "schedule", Kind: "component", Definition: &Reference{Scope: "shared", ID: "wmds/component/gantt"}, Placement: &Placement{Zone: "body", Rect: &rect}, Arguments: args, Keys: keys}}}
	p.Document.LocalTemplates = map[string]LocalTemplate{"plan": local}
	p.Document.Slides = []Slide{{ID: "plan-slide", ContentKind: "synthetic_example", Template: Reference{Scope: "local", ID: "plan"}, Values: map[string]any{"title": "Eight weeks from kickoff to a signed-off roadmap", "sidebar": "Workstreams"}}}
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
	return p, "schedule"
}

func ganttPatch(p *Project, id string, ops ...GanttOperation) GanttPatch {
	return GanttPatch{Schema: GanttPatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "Gantt qualification", Reason: "Customize synthetic work plan", NodeID: id, Timebase: "periods", Operations: append([]GanttOperation{{Action: "materialize", Entity: "source"}}, ops...)}
}

func TestGanttCompositionCatalogPreviewApplyStableReorder(t *testing.T) {
	p, id := ganttCompositionFixture(t)
	before := p.SourceHash()
	raw := append([]byte(nil), p.Raw...)
	inspect, e := InspectGantt(p, "plan-slide", id, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if len(inspect.Schedule.Groups) < 2 {
		t.Fatal("catalog groups missing")
	}
	order := []string{}
	for i := len(inspect.Schedule.Groups) - 1; i >= 0; i-- {
		order = append(order, inspect.Schedule.Groups[i].Key)
	}
	gate := GanttGateValue{Key: "approval", At: 4.5, Label: "Approve", Callout: true}
	patch := ganttPatch(p, id, GanttOperation{Action: "remove", Entity: "gate", Key: ganttKey(inspect.Schedule.Gates[1].Key)}, GanttOperation{Action: "reorder", Entity: "group", Order: order}, GanttOperation{Action: "set", Entity: "gate", Key: "approval", Gate: &gate})
	preview, e := PatchGantt(p, "plan-slide", patch, bundle(t), wmdesign.CandidateEngine, false)
	if e != nil {
		t.Fatal(e)
	}
	disk, e := os.ReadFile(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(disk, raw) || p.SourceHash() != before || preview.Applied || preview.AfterSHA256 == before {
		t.Fatal("preview changed disk or omitted changes")
	}
	result, e := PatchGantt(p, "plan-slide", patch, bundle(t), wmdesign.CandidateEngine, true)
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
	after, e := InspectGantt(p, "plan-slide", id, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if after.Schedule.Groups[0].Key != order[0] || after.MutationBlocked != "" {
		t.Fatal("stable group keys/bindings not preserved")
	}
	if !after.Schedule.Gates[len(after.Schedule.Gates)-1].Callout {
		t.Fatal("callout emphasis lost")
	}
	if _, e = PatchGantt(p, "plan-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e == nil || !strings.Contains(e.Error(), "hash mismatch") {
		t.Fatalf("stale apply accepted: %v", e)
	}
	if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
}

func TestGanttCompositionOperations(t *testing.T) {
	s := wmdesign.GanttSpec{Groups: []wmdesign.GanttGroup{{Key: "delivery", Lanes: []wmdesign.GanttLane{{Key: "build", Items: []wmdesign.GanttItem{{Key: "api", Kind: "build", Label: "API", From: 1, To: 2}}}, {Key: "test"}}}}}
	op := GanttOperation{Action: "move", Entity: "task", Key: "api", Group: "delivery", Lane: "build", ToGroup: "delivery", ToLane: "test"}
	if e := applyGanttOperation(&s, op); e != nil {
		t.Fatal(e)
	}
	if len(s.Groups[0].Lanes[0].Items) != 0 || s.Groups[0].Lanes[1].Items[0].Key != "api" {
		t.Fatal("reassignment lost identity")
	}
	task := s.Groups[0].Lanes[1].Items[0]
	task.From, task.To = 2.25, 4.75
	if e := applyGanttOperation(&s, GanttOperation{Action: "set", Entity: "task", Group: "delivery", Lane: "test", Key: "api", Task: &task}); e != nil {
		t.Fatal(e)
	}
	if s.Groups[0].Lanes[1].Items[0].To != 4.75 {
		t.Fatal("retiming failed")
	}
	if e := applyGanttOperation(&s, GanttOperation{Action: "remove", Entity: "lane", Group: "delivery", Key: "test"}); e == nil {
		t.Fatal("implicit child deletion accepted")
	}
	if e := applyGanttOperation(&s, GanttOperation{Action: "remove", Entity: "lane", Group: "delivery", Key: "test", Cascade: true}); e != nil {
		t.Fatal(e)
	}
	if e := applyGanttOperation(&s, GanttOperation{Action: "reorder", Entity: "lane", Group: "delivery", Order: []string{"build", "build"}}); e == nil {
		t.Fatal("duplicate reorder accepted")
	}
	if e := applyGanttOperation(&s, GanttOperation{Action: "remove", Entity: "group", Key: "delivery", At: new(float64)}); e == nil {
		t.Fatal("irrelevant operation field accepted")
	}
}

func TestGanttCompositionRejectsAuthoredZeroFields(t *testing.T) {
	base := GanttPatch{Schema: GanttPatchSchema, ExpectedSourceSHA256: strings.Repeat("a", 64), Actor: "Operator", Reason: "Remove marker", NodeID: "schedule", Timebase: "periods", Operations: []GanttOperation{{Action: "remove", Entity: "today"}}}
	for field, value := range map[string]any{"at": 0, "cascade": false, "key": "", "order": []any{}} {
		var obj map[string]any
		if e := json.Unmarshal(canonical(base), &obj); e != nil {
			t.Fatal(e)
		}
		obj["operations"].([]any)[0].(map[string]any)[field] = value
		if _, e := DecodeGanttPatch(canonical(obj), "patch"); e == nil {
			t.Fatalf("irrelevant authored zero field %s accepted", field)
		}
	}
}

func TestGanttCompositionEightLaneCapacityFailureRetainsSource(t *testing.T) {
	p, id := ganttCompositionFixture(t)
	before := append([]byte(nil), p.Raw...)
	inspected, e := InspectGantt(p, "plan-slide", id, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	group := inspected.Schedule.Groups[0]
	original := group.Lanes[0]
	for i := 1; i < 8; i++ {
		lane := original
		lane.Key = fmt.Sprintf("lane-%d", i+1)
		group.Lanes = append(group.Lanes, lane)
	}
	patch := ganttPatch(p, id, GanttOperation{Action: "set", Entity: "group", Key: group.Key, GroupValue: &group})
	if _, e = PatchGantt(p, "plan-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e == nil || !strings.Contains(e.Error(), "overflow") {
		t.Fatalf("eight-lane allocation was not rejected explicitly: %v", e)
	}
	raw, e := os.ReadFile(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(before, raw) {
		t.Fatal("oversized schedule changed source")
	}
}

func TestGanttCompositionValidationAndAtomicOverflow(t *testing.T) {
	p, id := ganttCompositionFixture(t)
	before, e := os.ReadFile(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	inspect, e := InspectGantt(p, "plan-slide", id, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if inspect.MutationBlocked != "" {
		patch := ganttPatch(p, id, GanttOperation{Action: "remove", Entity: "gate", Key: ganttKey(inspect.Schedule.Gates[0].Key)})
		patch.Operations = patch.Operations[1:]
		if _, e = PatchGantt(p, "plan-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e == nil {
			t.Fatal("implicit binding materialization accepted")
		}
	}
	pitch := 100.0
	patch := ganttPatch(p, id, GanttOperation{Action: "set", Entity: "layout", TrackPitch: &pitch})
	if _, e = PatchGantt(p, "plan-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e == nil {
		t.Fatal("invalid geometry committed")
	}
	after, e := os.ReadFile(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("failed fit changed source")
	}
	for _, raw := range []string{`{"schema":"pptxgengo.gantt-patch.v1","unknown":true}`, `schema: pptxgengo.gantt-patch.v1\n---\nschema: x`, strings.Replace(string(canonical(patch)), `"periods"`, `"dates"`, 1)} {
		if _, e = DecodeGanttPatch([]byte(raw), "patch"); e == nil {
			t.Fatal("invalid/unknown/timebase patch accepted")
		}
	}
}

// Opt-in retained catalog-derived packet for actual CLI/native qualification.
// It omits only unavailable optional icons, never substitutes a raster timeline.
func TestGanttCompositionDemo(t *testing.T) {
	out := os.Getenv("PPTXGENGO_GANTT_DEMO_OUT")
	if out == "" {
		t.Skip("set a fresh private Documents output directory")
	}
	if e := os.Mkdir(out, 0700); e != nil {
		t.Fatal(e)
	}
	p, id := ganttCompositionFixture(t)
	first, e := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine})
	if e != nil {
		t.Fatal(e)
	}
	before, e := InspectGantt(p, "plan-slide", id, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	g := before.Schedule.Groups[0]
	lane := g.Lanes[0]
	task := lane.Items[0]
	task.Label = "Mobilize"
	task.From, task.To = 0.25, 1.25
	stockGate := before.Schedule.Gates[1]
	gate := GanttGateValue{Key: ganttKey(stockGate.Key), At: stockGate.At, Label: stockGate.Label, Callout: stockGate.Callout}
	gate.At = 4.5
	gate.Label = "Approve"
	gate.Callout = true
	today := 2.25
	patch := ganttPatch(p, id, GanttOperation{Action: "set", Entity: "task", Group: g.Key, Lane: lane.Key, Key: task.Key, Task: &task}, GanttOperation{Action: "set", Entity: "gate", Key: gate.Key, Gate: &gate}, GanttOperation{Action: "set", Entity: "today", At: &today})
	preview, e := PatchGantt(p, "plan-slide", patch, bundle(t), wmdesign.CandidateEngine, false)
	if e != nil {
		t.Fatal(e)
	}
	applied, e := PatchGantt(p, "plan-slide", patch, bundle(t), wmdesign.CandidateEngine, true)
	if e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.Root)
	if e != nil {
		t.Fatal(e)
	}
	last, e := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = SaveVersion(p, "Gantt qualification", "Catalog-derived source demo; desktop review pending"); e != nil {
		t.Fatal(e)
	}
	zip := filepath.Join(out, "project.zip")
	if _, e = ShareProject(p, zip); e != nil {
		t.Fatal(e)
	}
	if _, e = ExtractShare(zip, filepath.Join(out, "project")); e != nil {
		t.Fatal(e)
	}
	for name, v := range map[string]any{"inspection.json": before, "patch.json": patch, "preview.json": preview, "applied.json": applied, "builds.json": map[string]any{"before": first, "after": last, "limits": "Catalog-derived illustrative plan customized to three lanes, one task per lane, explicit 48pt track pitch and full-width legend; optional private icon graphics omitted. Source and Go renderer evidence only; native desktop qualification pending."}} {
		if e = os.WriteFile(filepath.Join(out, name), canonical(v), 0600); e != nil {
			t.Fatal(e)
		}
	}
	t.Log("Retained Gantt qualification packet:", out)
}

func TestGanttCompositionRejectsCollidingGateLabels(t *testing.T) {
	p, id := ganttCompositionFixture(t)
	inspect, e := InspectGantt(p, "plan-slide", id, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	original := inspect.Schedule.Gates[1]
	for _, offset := range []float64{0, 0.1} {
		gate := GanttGateValue{Key: "collision", At: original.At + offset, Label: "Approve"}
		_, e = PatchGantt(p, "plan-slide", ganttPatch(p, id, GanttOperation{Action: "set", Entity: "gate", Key: gate.Key, Gate: &gate}), bundle(t), wmdesign.CandidateEngine, false)
		if e == nil || !strings.Contains(e.Error(), "gantt_gate_labels_overlap") {
			t.Fatal("unreadable measured gate preview accepted", offset, e)
		}
	}
}
