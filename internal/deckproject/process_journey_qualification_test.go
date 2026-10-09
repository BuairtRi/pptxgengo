package deckproject

import (
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// Explicit opt-in exports private synthetic qualification projects. No desktop
// result is inferred from this setup: the caller must build/open/edit/save them.
func writeProcessJourneyQualification(t *testing.T, p *Project, name string) {
	t.Helper()
	root := os.Getenv("PPTXGENGO_FAMILY_QUALIFICATION_DIR")
	if root == "" {
		return
	}
	if !filepath.IsAbs(root) {
		t.Fatal("qualification output must be an absolute private directory")
	}
	dst := filepath.Join(root, name)
	if _, e := os.Stat(dst); e == nil {
		t.Fatal("qualification output already exists", dst)
	}
	if e := filepath.WalkDir(p.Root, func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		rel, e := filepath.Rel(p.Root, path)
		if e != nil {
			return e
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0700)
		}
		b, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		return os.WriteFile(target, b, 0600)
	}); e != nil {
		t.Fatal(e)
	}
}
func TestProcessJourneyQualificationExamples(t *testing.T) {
	if os.Getenv("PPTXGENGO_FAMILY_QUALIFICATION_DIR") == "" {
		t.Skip("set explicit private qualification output to export examples")
	}
	p := processFixtureUnpinned(t)
	if _, e := Pin(p, journeyBundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	writeProcessJourneyQualification(t, p, "process-branching")
	p = portfolioFixtureUnpinned(t)
	if _, e := Pin(p, journeyBundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	writeProcessJourneyQualification(t, p, "portfolio-horizons")
	p, _ = journeyCatalogFixture(t, "road-fork/decision")
	m, e := InspectJourney(p, "journey-slide", "journey", journeyBundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	chosen := m.Model.Branches[1].Key
	current := m.Model.Branches[1].Milestones[0].Key
	if _, e = PatchJourney(p, "journey-slide", journeyTestPatch(p, JourneyOperation{Action: "set", Entity: "chosen", Key: chosen}, JourneyOperation{Action: "set", Entity: "current", Key: current}), journeyBundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	writeProcessJourneyQualification(t, p, "journey-fork")
	p, _ = journeyCatalogFixture(t, "road/left")
	writeProcessJourneyQualification(t, p, "journey-road")
}

func TestComponentTableQualificationExamples(t *testing.T) {
	if os.Getenv("PPTXGENGO_FAMILY_QUALIFICATION_DIR") == "" {
		t.Skip("set explicit private qualification output to export examples")
	}
	p := semanticComponentFixture(t, "card", map[string]any{"surface": "light", "title": "Illustrative card", "body": []any{map[string]any{"bullets": []any{"First observation", "Second observation"}}, map[string]any{"p": "Supporting note"}}}, map[string][]string{"body": {"observations", "note"}, "body/0/bullets": {"first", "second"}})
	if _, e := PatchComponent(p, "flow-slide", componentPatch(p, ComponentOperation{Action: "set", Entity: "item", Path: "/body/@observations/bullets", Key: "third", Value: "Third observation"}, ComponentOperation{Action: "reorder", Entity: "item", Path: "/body", Order: []string{"note", "observations"}}), journeyBundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	writeProcessJourneyQualification(t, p, "component-native-card")
	p = tableFixture(t)
	if _, e := PatchTable(p, "flow-slide", tablePatch(p, TableOperation{Action: "set", Entity: "row", Key: "four", Row: map[string]any{"name": "Four", "status": "pass", "score": nil}}, TableOperation{Action: "reorder", Entity: "column", Order: []string{"name", "score", "status"}}), journeyBundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	writeProcessJourneyQualification(t, p, "table-native-groups")
}
