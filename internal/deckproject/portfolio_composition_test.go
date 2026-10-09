package deckproject

import (
	"bytes"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"math"
	"os"
	"strings"
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
	before := append([]byte{}, p.Raw...)
	if _, err := PatchPortfolio(p, "flow-slide", portfolioPatch(p, op1, op2), bundle(t), wmdesign.CandidateEngine, true); err == nil || !strings.Contains(err.Error(), "measured outcome label") {
		t.Fatal("unreadable adjacent dependency label accepted", err)
	}
	disk, _ := os.ReadFile(p.SourcePath)
	if !bytes.Equal(before, disk) {
		t.Fatal("refusal mutated source")
	}
	// Keep five horizons in their authored order and all memberships unchanged.
	// Route the named dependency through open space above the adjacent cards.
	dependency := portfolioTestModel().Dependencies[0]
	dependency.Route = [][2]float64{{159.6, 9}, {169.2, 9}, {169.2, 64.5}}
	dependency.LabelPosition = &[2]float64{169.2, 9}
	route := PortfolioOperation{Action: "set", Entity: "dependency", Key: dependency.Key, Dependency: &dependency}
	if _, e := PatchPortfolio(p, "flow-slide", portfolioPatch(p, op1, op2, route), bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	m, e := InspectPortfolio(p, "flow-slide", "flow", bundle(t), wmdesign.CandidateEngine)
	if e != nil || len(m.Model.Horizons) != 5 {
		t.Fatalf("five horizons %v", e)
	}
	if m.Model.Dependencies[0].From != "access" || m.Model.Dependencies[0].To != "data" || m.Model.Dependencies[0].Label != "Informs" || m.Model.Initiatives[0].Horizon != "now" || m.Model.Initiatives[1].Horizon != "next" {
		t.Fatal("spacing changed portfolio facts")
	}
	for i, key := range []string{"now", "next", "later", "discover", "explore"} {
		if m.Model.Horizons[i].Key != key {
			t.Fatal("routing changed horizon order")
		}
	}
	if !bytes.Equal(canonical(m.Model.Initiatives), canonical(portfolioTestModel().Initiatives)) || m.Model.Dependencies[0].LabelPosition == nil || *m.Model.Dependencies[0].LabelPosition != [2]float64{169.2, 9} {
		t.Fatal("routing changed initiative facts or lost explicit label anchor")
	}
	if _, err := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); err != nil {
		t.Fatal("five-horizon applied source did not build", err)
	}
	writeProcessJourneyQualification(t, p, "portfolio-five-horizons-explicit-dependency")
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

func TestPortfolioDependencyLabelPositionStrictCoordinatesAndGuards(t *testing.T) {
	p := portfolioFixture(t)
	dependency := portfolioTestModel().Dependencies[0]
	dependency.LabelPosition = &[2]float64{1, 2}
	raw := canonical(portfolioPatch(p, PortfolioOperation{Action: "set", Entity: "dependency", Key: dependency.Key, Dependency: &dependency}))
	for _, bad := range []string{"[1]", "[1,2,3]", "null", "[null,2]", "[\"1\",2]"} {
		changed := bytes.Replace(raw, []byte(`"label_position":[1,2]`), []byte(`"label_position":`+bad), 1)
		if _, err := DecodePortfolioPatch(changed, "invalid-label-anchor"); err == nil {
			t.Fatal("invalid dependency anchor accepted", bad)
		}
	}
	dependency.Route = [][2]float64{{1, 2}}
	raw = canonical(portfolioPatch(p, PortfolioOperation{Action: "set", Entity: "dependency", Key: dependency.Key, Dependency: &dependency}))
	for _, bad := range []string{"[[1]]", "[[1,2,3]]", "[[null,2]]", "[null]"} {
		changed := bytes.Replace(raw, []byte(`"route":[[1,2]]`), []byte(`"route":`+bad), 1)
		if _, err := DecodePortfolioPatch(changed, "invalid-dependency-route"); err == nil {
			t.Fatal("invalid dependency waypoint accepted", bad)
		}
	}
	before := append([]byte{}, p.Raw...)
	for _, anchor := range [][2]float64{{141, 64.5}, {-20, 9}} {
		dependency.Route = nil
		dependency.LabelPosition = &anchor
		if _, err := PatchPortfolio(p, "flow-slide", portfolioPatch(p, PortfolioOperation{Action: "set", Entity: "dependency", Key: dependency.Key, Dependency: &dependency}), bundle(t), wmdesign.CandidateEngine, true); err == nil {
			t.Fatal("unsafe dependency label accepted", anchor)
		}
		disk, _ := os.ReadFile(p.SourcePath)
		if !bytes.Equal(before, disk) {
			t.Fatal("refusal altered source")
		}
	}
	m := portfolioTestModel()
	m.Dependencies[0].LabelPosition = &[2]float64{math.Inf(1), 9}
	if wmdesign.ValidatePortfolio(m) == nil {
		t.Fatal("nonfinite anchor accepted")
	}
	m.Dependencies[0].LabelPosition = &[2]float64{169, 9}
	m.Dependencies[0].Label = ""
	if wmdesign.ValidatePortfolio(m) == nil {
		t.Fatal("unnamed dependency anchor accepted")
	}
}

func TestPortfolioDependencyNonadjacentRoutePreservesMembership(t *testing.T) {
	p := portfolioFixture(t)
	before := append([]byte{}, p.Raw...)
	dependency := wmdesign.PortfolioDependency{Key: "access-enable", From: "access", To: "enable", Label: "Supports"}
	patch := func(d wmdesign.PortfolioDependency) PortfolioPatch {
		return portfolioPatch(p, PortfolioOperation{Action: "set", Entity: "dependency", Key: d.Key, Dependency: &d})
	}
	if _, err := PatchPortfolio(p, "flow-slide", patch(dependency), bundle(t), wmdesign.CandidateEngine, true); err == nil || !strings.Contains(err.Error(), "crosses step data") {
		t.Fatal("default route hid behind intermediate card", err)
	}
	dependency.Route = [][2]float64{{427, 64.5}, {427, 193.5}}
	if _, err := PatchPortfolio(p, "flow-slide", patch(dependency), bundle(t), wmdesign.CandidateEngine, true); err == nil {
		t.Fatal("explicit route through intermediate card accepted")
	}
	disk, _ := os.ReadFile(p.SourcePath)
	if !bytes.Equal(before, disk) {
		t.Fatal("crossing refusal changed source")
	}
	dependency.Route = [][2]float64{{225, 64.5}, {225, 9}, {621, 9}, {621, 193.5}}
	dependency.LabelPosition = &[2]float64{423, 9}
	accepted := patch(dependency)
	if _, err := PatchPortfolio(p, "flow-slide", accepted, bundle(t), wmdesign.CandidateEngine, true); err != nil {
		t.Fatal("clear nonadjacent route refused", err)
	}
	p, _ = Load(p.SourcePath)
	inspection, err := InspectPortfolio(p, "flow-slide", "flow", bundle(t), wmdesign.CandidateEngine)
	if err != nil || inspection.RenderError != "" {
		t.Fatal(err, inspection.RenderError)
	}
	if !bytes.Equal(canonical(inspection.Model.Horizons), canonical(portfolioTestModel().Horizons)) || !bytes.Equal(canonical(inspection.Model.Initiatives), canonical(portfolioTestModel().Initiatives)) {
		t.Fatal("routing reclassified initiatives or horizons")
	}
	if len(inspection.Model.Dependencies) != 2 || !bytes.Equal(canonical(inspection.Model.Dependencies[1]), canonical(dependency)) {
		t.Fatal("clear route/label was not persistent")
	}
	after := append([]byte{}, p.Raw...)
	if _, err = PatchPortfolio(p, "flow-slide", accepted, bundle(t), wmdesign.CandidateEngine, true); err == nil {
		t.Fatal("stale route accepted")
	}
	disk, _ = os.ReadFile(p.SourcePath)
	if !bytes.Equal(after, disk) {
		t.Fatal("stale refusal changed source")
	}
}
