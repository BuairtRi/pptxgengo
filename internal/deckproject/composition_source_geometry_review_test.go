package deckproject

import (
	"bytes"
	"os"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func reviewAddGeometry(t *testing.T, p *Project, template, node string) (*Project, map[string]any) {
	t.Helper()
	geometry := map[string]any{"schema": wmdesign.SceneSourceGeometrySchema, "x_fraction": 0, "y_fraction": 0, "width_fraction": 1}
	local := p.Document.LocalTemplates[template]
	for i := range local.Nodes {
		if local.Nodes[i].ID == node {
			local.Nodes[i].Arguments[wmdesign.SceneSourceGeometryArgument] = geometry
		}
	}
	p.Document.LocalTemplates[template] = local
	if e := os.WriteFile(p.SourcePath, canonical(p.Document), 0600); e != nil {
		t.Fatal(e)
	}
	p, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	return p, geometry
}
func reviewAssertGeometry(t *testing.T, p *Project, template, node string, geometry map[string]any) {
	t.Helper()
	p, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	for _, n := range p.Document.LocalTemplates[template].Nodes {
		if n.ID == node {
			if !bytes.Equal(canonical(geometry), canonical(n.Arguments[wmdesign.SceneSourceGeometryArgument])) {
				t.Fatal("captured source geometry lost")
			}
			return
		}
	}
	t.Fatal("node lost")
}
func TestCompositionSourceGeometryReviewCycle(t *testing.T) {
	p, id := cycleCompositionFixture(t)
	p, geometry := reviewAddGeometry(t, p, "cycle-plan", id)
	m, e := InspectCycle(p, "cycle-slide", id, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	patch := CyclePatch{Schema: CyclePatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "Review", Reason: "Retain captured geometry when semantic current changes", NodeID: id, Operations: []CycleOperation{{Action: "materialize", Entity: "source"}, {Action: "set", Entity: "active", Key: m.Model.Steps[1].Key}}}
	if _, e = PatchCycle(p, "cycle-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	reviewAssertGeometry(t, p, "cycle-plan", id, geometry)
}
func TestCompositionSourceGeometryReviewGantt(t *testing.T) {
	p, id := ganttCompositionFixture(t)
	p, geometry := reviewAddGeometry(t, p, "plan", id)
	if _, e := InspectGantt(p, "plan-slide", id, bundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	at := 2.25
	if _, e := PatchGantt(p, "plan-slide", ganttPatch(p, id, GanttOperation{Action: "set", Entity: "today", At: &at}), bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	reviewAssertGeometry(t, p, "plan", id, geometry)
}
func TestCompositionSourceGeometryReviewAssessment(t *testing.T) {
	p, id := assessmentFixture(t)
	p, geometry := reviewAddGeometry(t, p, "assessment", id)
	m := assessmentModel()
	if _, e := PatchAssessment(p, "assessment-slide", assessmentPatch(p, id, AssessmentOperation{Action: "initialize", Entity: "source", Model: &m, LegendNode: "key", Cascade: true}), bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	if _, e := InspectAssessment(p, "assessment-slide", id, bundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	if _, e := PatchAssessment(p, "assessment-slide", assessmentPatch(p, id, AssessmentOperation{Action: "set", Entity: "score", Key: "access", Column: "north", Score: assessmentScore(1)}), bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	reviewAssertGeometry(t, p, "assessment", id, geometry)
}
