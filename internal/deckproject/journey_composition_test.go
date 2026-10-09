package deckproject

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"os"
	"path/filepath"
	"testing"
)

func journeyBundle(t *testing.T) string {
	t.Helper()
	p, e := filepath.Abs("../../library/wm-design-system/v11")
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func journeyCatalogFixture(t *testing.T, key string) (*Project, string) {
	t.Helper()
	p := example(t)
	catalog, e := wmdesign.LibraryCatalog(journeyBundle(t), "")
	if e != nil {
		t.Fatal(e)
	}
	var node map[string]any
	split := ""
	for _, def := range catalog {
		if def.Key == key {
			var slide struct {
				Body  []map[string]any `json:"body"`
				Split string           `json:"split"`
			}
			if e = json.Unmarshal(def.RawSlide, &slide); e != nil {
				t.Fatal(e)
			}
			split = slide.Split
			for _, n := range slide.Body {
				if n["type"] == "road" || n["type"] == "roadfork" {
					node = n
					break
				}
			}
		}
	}
	if node == nil {
		t.Fatal("no catalog journey", key)
	}
	kind := node["type"].(string)
	r := wmdesign.Rect{X: node["x"].(float64) - 57, Y: node["y"].(float64) - 126, W: node["w"].(float64), H: node["h"].(float64)}
	keys := map[string][]string{}
	if kind == "road" {
		for i := range node["milestones"].([]any) {
			keys["milestones"] = append(keys["milestones"], fmt.Sprintf("step-%03d", i+1))
		}
	} else {
		for i := range node["trunk"].([]any) {
			keys["trunk"] = append(keys["trunk"], fmt.Sprintf("trunk-%03d", i+1))
		}
		for i, b := range node["branches"].([]any) {
			keys["branches"] = append(keys["branches"], fmt.Sprintf("option-%03d", i+1))
			path := fmt.Sprintf("branches/%d/milestones", i)
			for j := range b.(map[string]any)["milestones"].([]any) {
				keys[path] = append(keys[path], fmt.Sprintf("option-%03d-step-%03d", i+1, j+1))
			}
		}
	}
	if split != "" {
		r.Y = node["y"].(float64) - 36
	}
	for _, k := range []string{"type", "id", "x", "y", "w", "h"} {
		delete(node, k)
	}
	options := wmdesign.FrameRequest{Rail: "none", Footer: "compact", TitleLines: 1, Density: "appendix", Surface: "light", Split: split}
	if split != "" {
		options.Density = "standard"
	}
	p.Document.LocalTemplates = map[string]LocalTemplate{"journey": {Name: "Catalog-derived illustrative journey", Frame: Reference{Scope: "shared", ID: "wmds/frame/none-compact"}, FrameOptions: &options, Grid: Reference{Scope: "shared", ID: "wmds/grid/12-columns"}, Zones: map[string]Zone{"title": {Role: "slide-title", Required: true, Schema: map[string]any{"type": "string"}}}, Nodes: []Node{{ID: "journey", Kind: "component", Definition: &Reference{Scope: "shared", ID: "wmds/component/" + kind}, Placement: &Placement{Zone: "body", Rect: &r}, Arguments: node, Keys: keys}}}}
	p.Document.Slides = []Slide{{ID: "journey-slide", ContentKind: "synthetic_example", Template: Reference{Scope: "local", ID: "journey"}, Values: map[string]any{"title": "Illustrative journey"}}}
	if e = os.WriteFile(p.SourcePath, canonical(p.Document), 0600); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = Pin(p, journeyBundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	return p, "journey"
}
func journeyTestPatch(p *Project, ops ...JourneyOperation) JourneyPatch {
	return JourneyPatch{Schema: JourneyPatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "Test", Reason: "Illustrative journey customization", NodeID: "journey", Operations: ops}
}
func TestJourneyCompositionRoadCatalogRoundTrip(t *testing.T) {
	t.Parallel()
	p, id := journeyCatalogFixture(t, "road/right")
	m, e := InspectJourney(p, "journey-slide", id, journeyBundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if m.RenderError != "" {
		t.Fatal(m.RenderError)
	}
	before := append([]byte{}, p.Raw...)
	k := m.Model.Trunk[1].Key
	patch := journeyTestPatch(p, JourneyOperation{Action: "set", Entity: "current", Key: k})
	if _, e = PatchJourney(p, "journey-slide", patch, journeyBundle(t), wmdesign.CandidateEngine, false); e != nil {
		t.Fatal(e)
	}
	disk, _ := os.ReadFile(p.SourcePath)
	if !bytes.Equal(before, disk) {
		t.Fatal("preview changed source")
	}
	if _, e = PatchJourney(p, "journey-slide", patch, journeyBundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	m, e = InspectJourney(p, "journey-slide", id, journeyBundle(t), wmdesign.CandidateEngine)
	if e != nil || m.Model.Current != k {
		t.Fatalf("lost current %v %+v", e, m.Model)
	}
	if _, e = PatchJourney(p, "journey-slide", patch, journeyBundle(t), wmdesign.CandidateEngine, true); e == nil {
		t.Fatal("stale accepted")
	}
}
func TestJourneyCompositionForkChosenFollowsReorder(t *testing.T) {
	t.Parallel()
	p, id := journeyCatalogFixture(t, "road-fork/decision")
	m, e := InspectJourney(p, "journey-slide", id, journeyBundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	a, b := m.Model.Branches[0].Key, m.Model.Branches[1].Key
	current := m.Model.Branches[1].Milestones[0].Key
	patch := journeyTestPatch(p, JourneyOperation{Action: "set", Entity: "chosen", Key: b}, JourneyOperation{Action: "set", Entity: "current", Key: current}, JourneyOperation{Action: "reorder", Entity: "branch", Order: append([]string{b, a}, func() []string {
		var out []string
		for _, br := range m.Model.Branches[2:] {
			out = append(out, br.Key)
		}
		return out
	}()...)})
	if _, e = PatchJourney(p, "journey-slide", patch, journeyBundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	m, e = InspectJourney(p, "journey-slide", id, journeyBundle(t), wmdesign.CandidateEngine)
	if e != nil || m.Model.Current != current || m.Model.Chosen != b || m.Model.Branches[0].Key != b {
		t.Fatalf("lost keyed refs %v %+v", e, m.Model)
	}
	if e = applyJourneyOperation(&m.Model, JourneyOperation{Action: "remove", Entity: "milestone", Key: current}); e == nil {
		t.Fatal("current removal accepted")
	}
	if e = applyJourneyOperation(&m.Model, JourneyOperation{Action: "remove", Entity: "milestone", Key: current, Cascade: true}); e != nil {
		t.Fatal(e)
	}
}
func TestJourneyCompositionAllCatalogTypedModels(t *testing.T) {
	t.Parallel()
	keys := []string{"road/right", "road/left", "road/phases", "road/cards", "road/narrative-split", "road/here", "road-fork/parallel", "road-fork/decision", "road-fork/decision-chosen", "road-fork/parallel-split", "road-fork/decision-criteria", "road-fork/parallel-detail", "road-fork/foundation-then-waves"}
	for _, key := range keys {
		t.Run(key, func(t *testing.T) {
			p, id := journeyCatalogFixture(t, key)
			m, e := InspectJourney(p, "journey-slide", id, journeyBundle(t), wmdesign.CandidateEngine)
			if e != nil {
				t.Fatal(e)
			}
			if m.RenderError != "" {
				t.Fatal(m.RenderError)
			}
			current := ""
			if len(m.Model.Trunk) > 0 {
				current = m.Model.Trunk[0].Key
			} else {
				current = m.Model.Branches[0].Milestones[0].Key
			}
			patch := journeyTestPatch(p, JourneyOperation{Action: "set", Entity: "current", Key: current})
			if _, e = PatchJourney(p, "journey-slide", patch, journeyBundle(t), wmdesign.CandidateEngine, false); e != nil {
				t.Fatal(e)
			}
		})
	}
}

func TestJourneyCompositionFourTrunkAndUnequalOptions(t *testing.T) {
	t.Parallel()
	p, id := journeyCatalogFixture(t, "road-fork/decision-chosen")
	m, e := InspectJourney(p, "journey-slide", id, journeyBundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	ops := []JourneyOperation{}
	for _, v := range m.Model.Trunk {
		ops = append(ops, JourneyOperation{Action: "remove", Entity: "milestone", Key: v.Key, Cascade: true})
	}
	labels := []string{"Scope", "Plan", "Prepare", "Start"}
	for i, label := range labels {
		k := fmt.Sprintf("prep-%d", i+1)
		ops = append(ops, JourneyOperation{Action: "set", Entity: "milestone", Key: k, Milestone: &JourneyMilestone{Key: k, Label: label}})
	}
	at := .48
	ops = append(ops, JourneyOperation{Action: "set", Entity: "fork", Fork: &JourneyFork{Label: "Choose", At: &at}})
	for i, b := range m.Model.Branches {
		b.Title = fmt.Sprintf("Option %d", i+1)
		b.Text = ""
		b.Milestones = []JourneyMilestone{}
		for j := 0; j < i+1; j++ {
			b.Milestones = append(b.Milestones, JourneyMilestone{Key: fmt.Sprintf("option-%d-%d", i, j), Label: fmt.Sprintf("Step %d", j+1)})
		}
		ops = append(ops, JourneyOperation{Action: "set", Entity: "branch", Key: b.Key, BranchValue: &b})
	}
	ops = append(ops, JourneyOperation{Action: "remove", Entity: "current"})
	if _, e = PatchJourney(p, "journey-slide", journeyTestPatch(p, ops...), journeyBundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	m, e = InspectJourney(p, "journey-slide", id, journeyBundle(t), wmdesign.CandidateEngine)
	if e != nil || len(m.Model.Trunk) != 4 || len(m.Model.Branches[0].Milestones) != 1 || len(m.Model.Branches[1].Milestones) != 2 {
		t.Fatalf("variable fork %v %+v", e, m.Model)
	}
}
