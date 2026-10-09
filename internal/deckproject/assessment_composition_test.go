package deckproject

import (
	"bytes"
	"encoding/json"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"os"
	"strings"
	"testing"
)

func assessmentScore(v int) *int { return &v }
func assessmentModel() wmdesign.AssessmentSpec {
	return wmdesign.AssessmentSpec{RowLabel: "Capability", LabelWidth: 180, RowHeight: 36, ShowScores: true, Domain: wmdesign.AssessmentDomain{Max: 4, Scale: "seq", Labels: []string{"Absent", "Initial", "Developing", "Established", "Leading"}, MissingLabel: "Not assessed"}, Columns: []wmdesign.AssessmentAxis{{Key: "north", Label: "North"}, {Key: "south", Label: "South"}}, Rows: []wmdesign.AssessmentRow{{Key: "access", Label: "Access", Scores: map[string]*int{"north": assessmentScore(0), "south": nil}}, {Key: "scheduling", Label: "Scheduling", Scores: map[string]*int{"north": assessmentScore(3), "south": assessmentScore(2)}}}}
}
func assessmentFixture(t *testing.T) (*Project, string) {
	t.Helper()
	p := example(t)
	// Real catalog table/legend primitives, selected illustrative rows/sites and
	// a local rectangular allocation. The source catalogue remains immutable.
	catalog, e := wmdesign.LibraryCatalog(bundle(t), "")
	if e != nil {
		t.Fatal(e)
	}
	var raw struct {
		Body []map[string]any `json:"body"`
	}
	for _, def := range catalog {
		if def.Key == "capability-heat/callouts" {
			if e = json.Unmarshal(def.RawSlide, &raw); e != nil {
				t.Fatal(e)
			}
		}
	}
	if len(raw.Body) < 2 {
		t.Fatal("catalog heat specimen missing")
	}
	table := map[string]any{}
	legend := map[string]any{}
	for _, node := range raw.Body {
		if node["type"] == "table" && len(table) == 0 {
			table = node
		}
		if node["type"] == "legend" && len(legend) == 0 {
			legend = node
		}
	}
	if len(table) == 0 || len(legend) == 0 {
		t.Fatal("catalog heat table/legend missing")
	}
	for _, k := range []string{"type", "x", "y", "w", "h"} {
		delete(table, k)
	}
	// Preserve the stock cell/heading language; intentionally reduce the source
	// axes to demonstrate a small assessment, not a fixed stock count.
	table["cols"] = []any{map[string]any{"k": "c0", "label": "Capability", "w": 180}, map[string]any{"k": "c1", "label": "North", "w": 330, "type": "heat", "max": 4, "showValue": true}, map[string]any{"k": "c2", "label": "South", "w": 330, "type": "heat", "max": 4, "showValue": true}}
	table["rows"] = []any{map[string]any{"c0": "Access", "c1": 0, "c2": nil}, map[string]any{"c0": "Scheduling", "c1": 3, "c2": 2}}
	table["dense"] = false
	table["rowH"] = 36

	for _, k := range []string{"type", "x", "y", "w", "h"} {
		delete(legend, k)
	}
	options := wmdesign.FrameRequest{Rail: "none", Footer: "compact", TitleLines: 1, Density: "appendix", Surface: "light"}
	local := LocalTemplate{Name: "Catalog capability assessment derivative", Frame: Reference{Scope: "shared", ID: "wmds/frame/none-compact"}, FrameOptions: &options, Grid: Reference{Scope: "shared", ID: "wmds/grid/12-columns"}, Zones: map[string]Zone{"title": {Role: "slide-title", Required: true, Schema: map[string]any{"type": "string"}}}, Nodes: []Node{{ID: "heat", Kind: "component", Definition: &Reference{Scope: "shared", ID: "wmds/component/table"}, Placement: &Placement{Zone: "body", Rect: &wmdesign.Rect{X: 0, Y: 0, W: 840, H: 260}}, Arguments: table}, {ID: "key", Kind: "component", Definition: &Reference{Scope: "shared", ID: "wmds/component/legend"}, Placement: &Placement{Zone: "body", Rect: &wmdesign.Rect{X: 0, Y: 170, W: 840, H: 60}}, Arguments: legend}}}
	p.Document.LocalTemplates = map[string]LocalTemplate{"assessment": local}
	p.Document.Slides = []Slide{{ID: "assessment-slide", ContentKind: "synthetic_example", Template: Reference{Scope: "local", ID: "assessment"}, Values: map[string]any{"title": "Illustrative capability assessment"}}}
	rawBytes, e := json.Marshal(p.Document)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(p.SourcePath, rawBytes, 0600); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	pin(t, p)
	return p, "heat"
}
func assessmentPatch(p *Project, id string, ops ...AssessmentOperation) AssessmentPatch {
	return AssessmentPatch{Schema: AssessmentPatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "Assessment qualification", Reason: "Customize synthetic capability comparison", NodeID: id, Operations: ops}
}
func TestAssessmentCompositionInitializePreviewApplyReorderRebuild(t *testing.T) {
	p, id := assessmentFixture(t)
	before := append([]byte(nil), p.Raw...)
	m := assessmentModel()
	patch := assessmentPatch(p, id, AssessmentOperation{Action: "initialize", Entity: "source", Model: &m, LegendNode: "key", Cascade: true}, AssessmentOperation{Action: "reorder", Entity: "column", Order: []string{"south", "north"}}, AssessmentOperation{Action: "set", Entity: "row", Key: "data", Label: "Data and analytics"}, AssessmentOperation{Action: "set", Entity: "score", Key: "data", Column: "south", Score: assessmentScore(4)})
	out, e := PatchAssessment(p, "assessment-slide", patch, bundle(t), wmdesign.CandidateEngine, false)
	if e != nil {
		t.Fatal(e)
	}
	disk, e := os.ReadFile(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if out.Applied || !bytes.Equal(disk, before) {
		t.Fatal("preview changed source")
	}
	out, e = PatchAssessment(p, "assessment-slide", patch, bundle(t), wmdesign.CandidateEngine, true)
	if e != nil {
		t.Fatal(e)
	}
	if !out.Applied || out.Decision == "" {
		t.Fatal("guarded decision missing")
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	inspect, e := InspectAssessment(p, "assessment-slide", id, bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if inspect.RenderError != "" || len(inspect.Model.Rows) != 3 || inspect.Model.Columns[0].Key != "south" || inspect.Model.Rows[0].Scores["south"] != nil || *inspect.Model.Rows[0].Scores["north"] != 0 {
		t.Fatalf("identity/missing/zero lost: %+v", inspect)
	}
	if len(p.Document.LocalTemplates["assessment"].Nodes) != 1 || p.Document.Slides[0].Values["title"] != "Illustrative capability assessment" {
		t.Fatal("companion/chrome preservation failed")
	}
	if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	if _, e = PatchAssessment(p, "assessment-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e == nil || !strings.Contains(e.Error(), "hash mismatch") {
		t.Fatalf("stale patch accepted: %v", e)
	}
}
func TestAssessmentCompositionOperationMeaningAndCascade(t *testing.T) {
	s := assessmentModel()
	if e := applyAssessmentOperation(&s, AssessmentOperation{Action: "remove", Entity: "column", Key: "north"}); e == nil {
		t.Fatal("zero scored column deleted without cascade")
	}
	if e := applyAssessmentOperation(&s, AssessmentOperation{Action: "set", Entity: "score", Key: "access", Column: "north", Missing: true}); e != nil {
		t.Fatal(e)
	}
	if s.Rows[0].Scores["north"] != nil {
		t.Fatal("not assessed conversion failed")
	}
	for _, op := range []AssessmentOperation{{Action: "set", Entity: "score", Key: "access", Column: "south"}, {Action: "set", Entity: "score", Key: "access", Column: "south", Score: assessmentScore(5)}, {Action: "set", Entity: "score", Key: "access", Column: "south", Score: assessmentScore(0), Missing: true}, {Action: "reorder", Entity: "row", Order: []string{"access", "access"}}, {Action: "remove", Entity: "row", Key: "scheduling"}} {
		candidate := assessmentModel()
		if e := applyAssessmentOperation(&candidate, op); e == nil {
			t.Fatalf("invalid operation accepted: %+v", op)
		}
	}
	if e := applyAssessmentOperation(&s, AssessmentOperation{Action: "remove", Entity: "column", Key: "south", Cascade: true}); e != nil {
		t.Fatal(e)
	}
	for _, r := range s.Rows {
		if _, ok := r.Scores["south"]; ok {
			t.Fatal("dangling score after removal")
		}
	}
}
func TestAssessmentCompositionStrictDecode(t *testing.T) {
	p, id := assessmentFixture(t)
	good := assessmentPatch(p, id, AssessmentOperation{Action: "set", Entity: "score", Key: "access", Column: "north", Score: assessmentScore(0)})
	if _, e := DecodeAssessmentPatch(canonical(good), "patch"); e != nil {
		t.Fatal(e)
	}
	for _, field := range []string{`,"cascade":false`, `,"label":""`, `,"order":[]`} {
		raw := strings.Replace(string(canonical(good)), `"score":0`, `"score":0`+field, 1)
		if _, e := DecodeAssessmentPatch([]byte(raw), "patch"); e == nil {
			t.Fatal("irrelevant authored field accepted", field)
		}
	}
	raw := strings.Replace(string(canonical(good)), `"score":0`, `"score":0.5`, 1)
	if _, e := DecodeAssessmentPatch([]byte(raw), "patch"); e == nil {
		t.Fatal("fractional ordinal accepted")
	}
}
func TestAssessmentCompositionRefusesOverflowAndBadLegend(t *testing.T) {
	p, id := assessmentFixture(t)
	m := assessmentModel()
	m.Rows[0].Label = strings.Repeat("Long capability label ", 30)
	patch := assessmentPatch(p, id, AssessmentOperation{Action: "initialize", Entity: "source", Model: &m, LegendNode: "key", Cascade: true})
	if _, e := PatchAssessment(p, "assessment-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e == nil {
		t.Fatal("unsuitable long copy applied")
	}
	m = assessmentModel()
	m.Domain.Labels = m.Domain.Labels[:4]
	patch.Operations[0].Model = &m
	if _, e := PatchAssessment(p, "assessment-slide", patch, bundle(t), wmdesign.CandidateEngine, false); e == nil {
		t.Fatal("incomplete score legend accepted")
	}
}

func TestAssessmentCompositionSafetyAndLegendOrder(t *testing.T) {
	for _, name := range []string{"legend-first", "shared", "pinned", "native", "unacknowledged", "palette-drift", "global-min", "global-max", "show-scores-drift", "inline-group", "inline-total", "inline-scale", "inline-ink", "inline-h"} {
		t.Run(name, func(t *testing.T) {
			p, id := assessmentFixture(t)
			m := assessmentModel()
			patch := assessmentPatch(p, id, AssessmentOperation{Action: "initialize", Entity: "source", Model: &m, LegendNode: "key", Cascade: true})
			switch name {
			case "legend-first":
				local := p.Document.LocalTemplates["assessment"]
				local.Nodes[0], local.Nodes[1] = local.Nodes[1], local.Nodes[0]
				p.Document.LocalTemplates["assessment"] = local
			case "shared":
				other := p.Document.Slides[0]
				other.ID = "second"
				p.Document.Slides = append(p.Document.Slides, other)
			case "pinned":
				p.Document.Slides[0].Template.Revision = "sha256:pinned"
			case "native":
				p.Document.Slides[0].NativeOrder = map[string][]string{"root": {"heat"}}
			case "unacknowledged":
				patch.Operations[0].Cascade = false
			case "palette-drift":
				m.Domain.Scale = "risk"
			case "global-min":
				local := p.Document.LocalTemplates["assessment"]
				local.Nodes[0].Arguments["heatMin"] = 1
				p.Document.LocalTemplates["assessment"] = local
			case "global-max":
				local := p.Document.LocalTemplates["assessment"]
				local.Nodes[0].Arguments["heatMax"] = 100
				for _, col := range local.Nodes[0].Arguments["cols"].([]any)[1:] {
					delete(col.(map[string]any), "max")
				}
				p.Document.LocalTemplates["assessment"] = local
			case "show-scores-drift":
				m.ShowScores = false
			case "inline-group", "inline-total", "inline-scale", "inline-ink", "inline-h":
				local := p.Document.LocalTemplates["assessment"]
				local.Nodes[0].Arguments["rows"].([]any)[0].(map[string]any)[strings.TrimPrefix(name, "inline-")] = true
				p.Document.LocalTemplates["assessment"] = local
			}
			if name == "legend-first" {
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
				patch.ExpectedSourceSHA256 = p.SourceHash()
			}
			before := append([]byte(nil), p.Raw...)
			_, e := PatchAssessment(p, "assessment-slide", patch, bundle(t), wmdesign.CandidateEngine, false)
			if name == "legend-first" {
				if e != nil {
					t.Fatal(e)
				}
			} else if e == nil {
				t.Fatal("unsafe change accepted", name)
			}
			disk, readErr := os.ReadFile(p.SourcePath)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if !bytes.Equal(disk, before) {
				t.Fatal("refusal/preview changed source")
			}
		})
	}
}
