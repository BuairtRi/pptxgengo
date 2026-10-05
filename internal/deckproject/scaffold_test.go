package deckproject

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestScaffoldLoadsBuildsAndPinsActualParent(t *testing.T) {
	bundlePath := filepath.Join("..", "..", "planning", "wm-design-contracts", "v5", "intake-20261003-587-frozen", "bundle")
	for _, key := range []string{"lifecycle/three-phases", "plan/gantt", "intellio/swimlane"} {
		t.Run(key, func(t *testing.T) {
			if testing.Short() && key != "intellio/swimlane" {
				t.Skip("scaffold build uses registered private artwork; run make test-integration")
			}
			draft, err := ScaffoldTemplate(bundlePath, key, wmdesign.CandidateEngine, "Preserve original topology instead of inventing an extra phase", 2026)
			if err != nil {
				t.Fatal(err)
			}
			p := example(t)
			p.Document.LocalTemplates = map[string]LocalTemplate{"adapted": draft.Template}
			p.Document.Slides = []Slide{{ID: "one", ContentKind: "synthetic_example", Template: Reference{Scope: "local", ID: "adapted"}, Values: draft.SyntheticSourceValues}}
			data, err := json.Marshal(p.Document)
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(p.SourcePath, data, 0644); err != nil {
				t.Fatal(err)
			}
			candidate, err := Load(p.SourcePath)
			if err != nil {
				t.Fatal(err)
			}
			compiled, err := Compile(candidate, bundlePath, wmdesign.CandidateEngine)
			if err != nil {
				t.Fatal(err)
			}
			if _, _, err = wmdesign.BuildWithEngineAndAssets(bundlePath, "", compiled.Document, wmdesign.CandidateEngine, compiled.Assets); err != nil {
				t.Fatal(err)
			}
			// No snapshot or decorative parent label can bypass the actual
			// source/definition hash check.
			candidate.Document.LocalTemplates["adapted"].Provenance.DefinitionSHA256 = strings.Repeat("0", 64)
			if _, err = Compile(candidate, bundlePath, wmdesign.CandidateEngine); err == nil {
				t.Fatal("false shared ancestry accepted")
			}
		})
	}
}

func TestScaffoldRequiresReasonAndLeavesTypedRecipesExplicit(t *testing.T) {
	bundlePath := filepath.Join("..", "..", "planning", "wm-design-contracts", "v5", "intake-20261003-587-frozen", "bundle")
	if _, err := ScaffoldTemplate(bundlePath, "lifecycle/three-phases", wmdesign.CandidateEngine, "", 2026); err == nil {
		t.Fatal("missing reason accepted")
	}
	if _, err := ScaffoldTemplate(bundlePath, "cards/3", wmdesign.CandidateEngine, "Adapt three items", 2026); err == nil {
		t.Fatal("typed recipe silently retained specimen content")
	}
	if _, err := ScaffoldTemplate(bundlePath, "lifecycle/three-phases", wmdesign.CandidateEngine, "Omit an unused source relation", 2026, "unknown"); err == nil {
		t.Fatal("unknown omission accepted")
	}
	if _, err := ScaffoldTemplate(bundlePath, "lifecycle/three-phases", wmdesign.CandidateEngine, "Omit an unused source node", 2026, "node01", "node01"); err == nil {
		t.Fatal("duplicate omission accepted")
	}
	if testing.Short() {
		t.Skip("valid lifecycle scaffold build uses registered private artwork; run make test-integration")
	}
	draft, err := ScaffoldTemplate(bundlePath, "lifecycle/three-phases", wmdesign.CandidateEngine, "Omit an unused source node", 2026, "node01")
	if err != nil || len(draft.OmittedNodes) != 1 {
		t.Fatal("explicit omission lost", err)
	}
	for _, node := range draft.Template.Nodes {
		if node.ID == "node01" {
			t.Fatal("omitted node survived")
		}
	}
}
