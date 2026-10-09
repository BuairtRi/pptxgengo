package deckproject

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func ganttReviewSource() wmdesign.GanttSpec {
	return wmdesign.GanttSpec{Groups: []wmdesign.GanttGroup{{Key: "delivery", Label: "Delivery", Fill: "series.1", Lanes: []wmdesign.GanttLane{{Key: "build", Title: "Build", Icon: "none", Items: []wmdesign.GanttItem{{Key: "api", Kind: "build", Label: "API", From: 1, To: 2}}}}}}}
}
func TestGanttReviewDestructiveReplacementRequiresExplicitCascade(t *testing.T) {
	for _, tc := range []struct {
		name string
		op   GanttOperation
	}{
		{"group", GanttOperation{Action: "set", Entity: "group", Key: "delivery", GroupValue: &wmdesign.GanttGroup{Key: "delivery", Label: "Delivery", Fill: "series.1"}}},
		{"lane", GanttOperation{Action: "set", Entity: "lane", Group: "delivery", Key: "build", LaneValue: &wmdesign.GanttLane{Key: "build", Title: "Build", Icon: "none"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := ganttReviewSource()
			before := canonical(s)
			if e := applyGanttOperation(&s, tc.op); e == nil {
				t.Fatal("set erased children without explicit cascade")
			}
			if !bytes.Equal(before, canonical(s)) {
				t.Fatal("rejected destructive set changed source model")
			}
			tc.op.Cascade = true
			if e := applyGanttOperation(&s, tc.op); e != nil {
				t.Fatal("explicit cascade refused", e)
			}
		})
	}
}
func TestGanttReviewLayoutPitchContract(t *testing.T) {
	for _, pitch := range []float64{0, 23.99, 60.01, -1} {
		s := ganttReviewSource()
		if e := applyGanttOperation(&s, GanttOperation{Action: "set", Entity: "layout", TrackPitch: &pitch}); e == nil {
			t.Fatalf("accepted out-of-contract pitch %v", pitch)
		}
	}
	for _, pitch := range []float64{24, 30, 60} {
		s := ganttReviewSource()
		if e := applyGanttOperation(&s, GanttOperation{Action: "set", Entity: "layout", TrackPitch: &pitch}); e != nil {
			t.Fatal(e)
		}
	}
}
func TestGanttReviewMalformedOperationsAndIntervalsAreAtomic(t *testing.T) {
	p, id := ganttCompositionFixture(t)
	inspection, e := InspectGantt(p, "plan-slide", id, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	invalid := []GanttOperation{
		{Action: "set", Entity: "gate", Key: "out-of-range", Gate: &GanttGateValue{Key: "out-of-range", At: float64(len(inspection.Schedule.Periods.Labels)) + 0.1, Label: "Approve"}},
		{Action: "set", Entity: "today", At: func() *float64 { x := -0.1; return &x }()},
		{Action: "materialize", Entity: "source", Key: "ignored"},
		{Action: "reorder", Entity: "group", Order: []string{"missing"}},
	}
	before := append([]byte(nil), p.Raw...)
	for _, op := range invalid {
		patch := ganttPatch(p, id, op)
		if _, e := PatchGantt(p, "plan-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e == nil {
			t.Fatalf("accepted invalid operation %+v", op)
		}
		raw, e := os.ReadFile(p.SourcePath)
		if e != nil {
			t.Fatal(e)
		}
		if !bytes.Equal(before, raw) {
			t.Fatal("rejected operation changed source")
		}
	}
}
func TestGanttReviewInspectRetainsSourceWithoutFalseMeasurement(t *testing.T) {
	p, id := ganttCompositionFixture(t)
	template := p.Document.LocalTemplates["plan"]
	node, e := ganttNode(&template, id)
	if e != nil {
		t.Fatal(e)
	}
	node.Arguments["trackPitch"] = 100.0
	p.Document.LocalTemplates["plan"] = template
	if e = os.WriteFile(p.SourcePath, canonical(p.Document), 0600); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	inspection, e := InspectGantt(p, "plan-slide", id, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if inspection.SourceSHA256 != p.SourceHash() || inspection.Schedule.TrackPitch != 100 || len(inspection.Schedule.Groups) == 0 {
		t.Fatal("source inspection was lost on renderer failure")
	}
	if inspection.RenderError == "" || !strings.Contains(inspection.Geometry.Validation, "failed_no_measurement") {
		t.Fatal("renderer failure advertised measurement success")
	}
	if len(inspection.Geometry.FinalNative) > 0 || len(inspection.Geometry.MeasuredText) > 0 {
		t.Fatal("invented measured output")
	}
}

func TestGanttReviewTodayEndBoundaryIsMeasured(t *testing.T) {
	p, id := ganttCompositionFixture(t)
	inspection, e := InspectGantt(p, "plan-slide", id, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	at := float64(len(inspection.Schedule.Periods.Labels))
	patch := ganttPatch(p, id, GanttOperation{Action: "set", Entity: "today", At: &at})
	if _, e = PatchGantt(p, "plan-slide", patch, bundle(t), wmdesign.CandidateEngine, false); e != nil {
		t.Fatalf("valid last period boundary needs a bounded Today label: %v", e)
	}
}

func TestGanttReviewContradictoryTaskMeaningRefused(t *testing.T) {
	cases := []wmdesign.GanttItem{
		{Key: "api", Kind: "build", Label: "API", From: 1, To: 2, Milestone: true},
		{Key: "api", Kind: "build", Label: "API", From: 1, To: 2, Event: "star"},
		{Key: "api", Kind: "build", Label: "API", From: 1, To: 2, At: 3},
		{Key: "api", Kind: "build", Label: "API", From: 1, To: 2, LabelSide: "left"},
		{Key: "api", Milestone: true, At: 2, From: 1, To: 3},
		{Key: "api", Milestone: true, At: 2, Progress: func() *float64 { x := 0.5; return &x }()},
		{Key: "api", Milestone: true, Event: "star", At: 2},
		{Key: "api", Tag: true, Milestone: true, At: 2},
	}
	for _, task := range cases {
		s := ganttReviewSource()
		before := canonical(s)
		op := GanttOperation{Action: "set", Entity: "task", Group: "delivery", Lane: "build", Key: "api", Task: &task}
		if e := applyGanttOperation(&s, op); e == nil {
			t.Fatalf("contradictory task was accepted: %+v", task)
		}
		if !bytes.Equal(before, canonical(s)) {
			t.Fatal("rejected task mutated model")
		}
	}
	// Parent record replacement must enforce the same semantic task invariant.
	task := cases[0]
	for _, entity := range []string{"group", "lane"} {
		s := ganttReviewSource()
		op := GanttOperation{Action: "set", Entity: entity, Key: "delivery", GroupValue: &wmdesign.GanttGroup{Key: "delivery", Lanes: []wmdesign.GanttLane{{Key: "build", Items: []wmdesign.GanttItem{task}}}}}
		if entity == "lane" {
			op = GanttOperation{Action: "set", Entity: "lane", Group: "delivery", Key: "build", LaneValue: &wmdesign.GanttLane{Key: "build", Items: []wmdesign.GanttItem{task}}}
		}
		if e := applyGanttOperation(&s, op); e == nil {
			t.Fatal("parent set bypassed task semantic validation")
		}
	}
}
