package deckproject

import (
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestStaffingScaleContextUsesActualInkAndRetainsIntentionalOverlays(t *testing.T) {
	text := func(id string, x, y, w, h, advance float64) wmdesign.TextRecord {
		return wmdesign.TextRecord{ID: id, Rect: wmdesign.Rect{X: x, Y: y, W: w, H: h}, Layout: wmdesign.TextLayout{Lines: []wmdesign.TextLine{{Advance: advance}}, EstimatedOccupiedHeight: h}}
	}
	in := DiagramInspection{MeasuredText: []wmdesign.TextRecord{text("slide.body.model.scale", 57, 450, 846, 10, 350), text("slide.body.context.text", 60, 452, 100, 10, 100)}}
	if e := staffingScaleContext(in, "model"); e == nil || !strings.Contains(e.Error(), "context.text") {
		t.Fatalf("context intersection ignored: %v", e)
	}
	in.MeasuredText[1].Rect.Y = 430
	if e := staffingScaleContext(in, "model"); e != nil {
		t.Fatal(e)
	}
	// An unrelated object beyond the actual caption ink is clear, despite
	// overlapping the much wider caption allocation box.
	in.MeasuredText[1] = text("slide.body.context.text", 600, 452, 100, 10, 100)
	if e := staffingScaleContext(in, "model"); e != nil {
		t.Fatal(e)
	}
	in.MeasuredText[1] = text("slide.body.model.phases.0.label", 60, 452, 100, 10, 100)
	if e := staffingScaleContext(in, "model"); e != nil {
		t.Fatal("intentional own annotations", e)
	}
}

func TestStaffingScaleContextRefusalLeavesSourceUnchanged(t *testing.T) {
	p, id := curveFixture(t, "team-curve/agents", "teamcurve")
	local := p.Document.LocalTemplates["fixture"]
	rect := wmdesign.Rect{X: 0, Y: 314, W: 300, H: 16}
	local.Nodes = append(local.Nodes, Node{ID: "context", Kind: "component", Definition: &Reference{Scope: "shared", ID: "wmds/component/text"}, Placement: &Placement{Zone: "body", Rect: &rect}, Arguments: map[string]any{"text": "Retained contextual source caption", "style": "small", "ink": "primary"}})
	p.Document.LocalTemplates["fixture"] = local
	p = writeCurveFixture(t, p)
	before := p.SourceHash()
	patch := staffingPatch(p, StaffingOperation{Action: "materialize", Entity: "source"}, StaffingOperation{Action: "set", Entity: "scale", Scale: &StaffingScale{Unit: "capacity", TimeUnit: "delivery", Source: "Illustrative"}})
	patch.NodeID = id
	for _, apply := range []bool{false, true} {
		if _, e := PatchStaffing(p, "slide", patch, curveCompositionBundle(t), wmdesign.CandidateEngine, apply); e == nil || !strings.Contains(e.Error(), "intersects contextual text") {
			t.Fatalf("apply=%t context not refused %v", apply, e)
		}
		after, e := Load(p.SourcePath)
		if e != nil || after.SourceHash() != before {
			t.Fatalf("context refusal changed source %v", e)
		}
	}
}
