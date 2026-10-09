package deckproject

import (
	"bytes"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"os"
	"testing"
)

func portfolioTestModel() wmdesign.PortfolioSpec {
	return wmdesign.PortfolioSpec{CardHeight: 72, Horizons: []wmdesign.PortfolioHorizon{{Key: "now", Label: "Now", Meaning: "Committed focus"}, {Key: "next", Label: "Next", Meaning: "Planned focus"}, {Key: "later", Label: "Later", Meaning: "Options, not dates"}}, Statuses: []wmdesign.PortfolioStatus{{Key: "active", Label: "Active", Surface: "light"}, {Key: "planned", Label: "Planned", Surface: "subtle"}}, Initiatives: []wmdesign.PortfolioInitiative{{Key: "access", Label: "Access", Horizon: "now", Owner: "Operations", Status: "active", Confidence: "High confidence", Slot: 0}, {Key: "data", Label: "Data", Horizon: "next", Owner: "Technology", Status: "planned", Confidence: "Medium confidence", Slot: 0}, {Key: "enable", Label: "Enablement", Horizon: "later", Owner: "People", Status: "planned", Confidence: "Low confidence", Slot: 1}}, Dependencies: []wmdesign.PortfolioDependency{{Key: "access-data", From: "access", To: "data", Label: "Informs"}}}
}
func portfolioFixture(t *testing.T) *Project { p := portfolioFixtureUnpinned(t); pin(t, p); return p }
func portfolioFixtureUnpinned(t *testing.T) *Project {
	t.Helper()
	p := processFixtureUnpinned(t)
	m := portfolioTestModel()
	n := p.Document.LocalTemplates["process"].Nodes[0]
	n.Definition.ID = "wmds/component/portfolio"
	n.Arguments = map[string]any{}
	if e := strictInto(m, &n.Arguments); e != nil {
		t.Fatal(e)
	}
	local := p.Document.LocalTemplates["process"]
	local.Nodes[0] = n
	p.Document.LocalTemplates["process"] = local
	p.Document.Slides[0].Values["title"] = "Illustrative horizons keep commitments separate from options"
	if e := os.WriteFile(p.SourcePath, canonical(p.Document), 0600); e != nil {
		t.Fatal(e)
	}
	p, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func portfolioPatch(p *Project, ops ...PortfolioOperation) PortfolioPatch {
	return PortfolioPatch{Schema: PortfolioPatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "Test", Reason: "Illustrative portfolio customization", NodeID: "flow", Operations: ops}
}
func TestPortfolioCompositionTransferRetainsDependencies(t *testing.T) {
	p := portfolioFixture(t)
	m, e := InspectPortfolio(p, "flow-slide", "flow", bundle(t), wmdesign.CandidateEngine)
	if e != nil || m.RenderError != "" {
		t.Fatalf("%v %s", e, m.RenderError)
	}
	move := m.Model.Initiatives[1]
	move.Horizon = "later"
	move.Slot = 0
	patch := portfolioPatch(p, PortfolioOperation{Action: "set", Entity: "initiative", Key: move.Key, Initiative: &move}, PortfolioOperation{Action: "reorder", Entity: "horizon", Order: []string{"later", "next", "now"}})
	before := append([]byte{}, p.Raw...)
	if _, e = PatchPortfolio(p, "flow-slide", patch, bundle(t), wmdesign.CandidateEngine, false); e != nil {
		t.Fatal(e)
	}
	disk, _ := os.ReadFile(p.SourcePath)
	if !bytes.Equal(disk, before) {
		t.Fatal("preview altered source")
	}
	if _, e = PatchPortfolio(p, "flow-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	m, e = InspectPortfolio(p, "flow-slide", "flow", bundle(t), wmdesign.CandidateEngine)
	if e != nil || m.Model.Initiatives[1].Horizon != "later" || m.Model.Dependencies[0].To != "data" || m.Model.Horizons[0].Meaning != "Options, not dates" {
		t.Fatalf("lost meaning %v %+v", e, m.Model)
	}
	if _, e = PatchPortfolio(p, "flow-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e == nil {
		t.Fatal("stale accepted")
	}
}
func TestPortfolioCompositionFiveHorizonsAndCascade(t *testing.T) {
	p := portfolioFixture(t)
	op1 := PortfolioOperation{Action: "set", Entity: "horizon", Key: "discover", Horizon: &wmdesign.PortfolioHorizon{Key: "discover", Label: "Discover", Meaning: "Ideas under review"}}
	op2 := PortfolioOperation{Action: "set", Entity: "horizon", Key: "explore", Horizon: &wmdesign.PortfolioHorizon{Key: "explore", Label: "Explore", Meaning: "Evidence needed"}}
	if _, e := PatchPortfolio(p, "flow-slide", portfolioPatch(p, op1, op2), bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	m, e := InspectPortfolio(p, "flow-slide", "flow", bundle(t), wmdesign.CandidateEngine)
	if e != nil || len(m.Model.Horizons) != 5 {
		t.Fatalf("five horizons %v", e)
	}
	if e = removePortfolioInitiative(&m.Model, "access", false); e == nil {
		t.Fatal("incident deletion accepted")
	}
	if e = removePortfolioInitiative(&m.Model, "access", true); e != nil || len(m.Model.Dependencies) != 0 {
		t.Fatal("cascade incomplete", e)
	}
	if e = applyPortfolioOperation(&m.Model, &Node{}, PortfolioOperation{Action: "remove", Entity: "horizon", Key: "later"}); e == nil {
		t.Fatal("populated horizon accepted")
	}
}
func TestPortfolioCompositionInvalidAndStrictPresence(t *testing.T) {
	m := portfolioTestModel()
	m.Horizons[0].Meaning = ""
	if e := wmdesign.ValidatePortfolio(m); e == nil {
		t.Fatal("undefined horizon meaning")
	}
	m = portfolioTestModel()
	m.Initiatives[1].Horizon = "missing"
	if e := wmdesign.ValidatePortfolio(m); e == nil {
		t.Fatal("invalid horizon")
	}
	p := portfolioFixture(t)
	raw := canonical(portfolioPatch(p, PortfolioOperation{Action: "remove", Entity: "dependency", Key: "access-data"}))
	raw = bytes.Replace(raw, []byte(`"key":"access-data"`), []byte(`"key":"access-data","cascade":false`), 1)
	if _, e := DecodePortfolioPatch(raw, "bad"); e == nil {
		t.Fatal("irrelevant false accepted")
	}
}

func TestPortfolioCompositionCapturedSourceGeometryRetained(t *testing.T) {
	p := portfolioFixture(t)
	local := p.Document.LocalTemplates["process"]
	geometry := map[string]any{"schema": wmdesign.SceneSourceGeometrySchema, "x_fraction": 0, "y_fraction": 0, "width_fraction": 1, "height_fraction": 1}
	local.Nodes[0].Arguments[wmdesign.SceneSourceGeometryArgument] = geometry
	p.Document.LocalTemplates["process"] = local
	if e := os.WriteFile(p.SourcePath, canonical(p.Document), 0600); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	if _, e := InspectPortfolio(p, "flow-slide", "flow", bundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	if _, e := PatchPortfolio(p, "flow-slide", portfolioPatch(p, PortfolioOperation{Action: "reorder", Entity: "horizon", Order: []string{"later", "next", "now"}}), bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	if !bytes.Equal(canonical(geometry), canonical(p.Document.LocalTemplates["process"].Nodes[0].Arguments[wmdesign.SceneSourceGeometryArgument])) {
		t.Fatal("source geometry lost")
	}
}
