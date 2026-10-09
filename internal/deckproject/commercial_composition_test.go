package deckproject

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func commercialValue(v string) *string { return &v }
func commercialModel() wmdesign.CommercialModel {
	return wmdesign.CommercialModel{Precision: 2, Rounding: "half_even", Assumptions: []string{"Illustrative labor estimate, excludes taxes."}, Rows: []wmdesign.CommercialRow{
		{Key: "hours", Label: "Hours", Unit: "hours", Value: commercialValue("10.25"), Assumption: "Estimated hours"},
		{Key: "rate", Label: "Rate", Unit: "USD/hour", Value: commercialValue("100.10"), Source: "Illustrative rate card"},
		{Key: "fee", Label: "Fee", Unit: "USD", Formula: &wmdesign.CommercialFormula{Operation: "convert", Inputs: []string{"hours", "rate"}}, Assumption: "hours multiplied by USD/hour produces USD"},
	}, Targets: []wmdesign.CommercialTarget{{Row: "fee", Path: "/value", Prefix: "$"}}}
}
func commercialFixture(t *testing.T) (*Project, string) {
	t.Helper()
	p := example(t)
	options := wmdesign.FrameRequest{Rail: "none", Footer: "compact", TitleLines: 1, Density: "appendix", Surface: "light"}
	local := LocalTemplate{Name: "Illustrative commercial metric", Frame: Reference{Scope: "shared", ID: "wmds/frame/none-compact"}, FrameOptions: &options, Grid: Reference{Scope: "shared", ID: "wmds/grid/12-columns"}, Zones: map[string]Zone{"title": {Role: "slide-title", Required: true, Schema: map[string]any{"type": "string"}}}, Nodes: []Node{{ID: "fee", Kind: "component", Definition: &Reference{Scope: "shared", ID: "wmds/component/metric"}, Placement: &Placement{Zone: "body", Rect: &wmdesign.Rect{X: 0, Y: 0, W: 300, H: 120}}, Arguments: map[string]any{"value": "$1,026.03", "label": "Illustrative fee"}}, {ID: "assumption", Kind: "text", Style: "body", Text: "Illustrative only; excludes taxes", Placement: &Placement{Zone: "body", Rect: &wmdesign.Rect{X: 320, Y: 0, W: 500, H: 100}}}}}
	p.Document.LocalTemplates = map[string]LocalTemplate{"commercial": local}
	p.Document.Slides = []Slide{{ID: "commercial-slide", ContentKind: "synthetic_example", Template: Reference{Scope: "local", ID: "commercial"}, Values: map[string]any{"title": "Illustrative fee retains the decimal source model"}}}
	raw, _ := json.Marshal(p.Document)
	if e := os.WriteFile(p.SourcePath, raw, 0600); e != nil {
		t.Fatal(e)
	}
	p, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	pin(t, p)
	return p, "fee"
}
func commercialPatch(p *Project, ops ...CommercialOperation) CommercialPatch {
	return CommercialPatch{Schema: CommercialPatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "Finance qualification", Reason: "Customize illustrative decimal source facts", NodeID: "fee", Operations: ops}
}
func TestCommercialCompositionInitializePreviewApplyRebuild(t *testing.T) {
	t.Parallel()
	p, _ := commercialFixture(t)
	before := append([]byte(nil), p.Raw...)
	m := commercialModel()
	patch := commercialPatch(p, CommercialOperation{Action: "initialize", Entity: "source", Model: &m, Cascade: true}, CommercialOperation{Action: "set", Entity: "row", Key: "hours", Row: &wmdesign.CommercialRow{Key: "hours", Label: "Hours", Unit: "hours", Value: commercialValue("12.5"), Assumption: "Revised illustrative estimate"}}, CommercialOperation{Action: "reorder", Entity: "row", Order: []string{"fee", "rate", "hours"}})
	out, e := PatchCommercial(p, "commercial-slide", patch, bundle(t), wmdesign.CandidateEngine, false)
	if e != nil {
		t.Fatal(e)
	}
	disk, _ := os.ReadFile(p.SourcePath)
	if out.Applied || !bytes.Equal(before, disk) {
		t.Fatal("preview changed source")
	}
	out, e = PatchCommercial(p, "commercial-slide", patch, bundle(t), wmdesign.CandidateEngine, true)
	if e != nil || !out.Applied {
		t.Fatalf("apply %v", e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	inspect, e := InspectCommercial(p, "commercial-slide", "fee", bundle(t), wmdesign.CandidateEngine)
	if e != nil || inspect.RenderError != "" {
		t.Fatalf("inspect %+v %v", inspect, e)
	}
	if inspect.Results[0].Rounded != "1251.25" || inspect.Presentation["value"] != "$1251.25" || inspect.Presentation["label"] != "Illustrative fee" {
		t.Fatalf("bad mapped decimal %+v", inspect)
	}
	if p.Document.LocalTemplates["commercial"].Nodes[1].Text != "Illustrative only; excludes taxes" {
		t.Fatal("other copy changed")
	}
	if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	if _, e = PatchCommercial(p, "commercial-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e == nil {
		t.Fatal("stale accepted")
	}
}
func TestCommercialCompositionCollectionTargetsFollowKeys(t *testing.T) {
	t.Parallel()
	m := commercialModel()
	m.Targets = []wmdesign.CommercialTarget{{Row: "fee", Path: "/rows/0/amount", Prefix: "$"}}
	s := wmdesign.CommercialSpec{Model: m, Presentation: map[string]any{"type": "table", "rows": []any{map[string]any{"label": "Build", "amount": "old"}, map[string]any{"label": "Run", "amount": "unchanged"}}}}
	keys := map[string][]string{"presentation/rows": {"build", "run"}}
	if e := applyCommercial(&s, keys, CommercialOperation{Action: "reorder", Entity: "collection", Path: "/rows", Order: []string{"run", "build"}}); e != nil {
		t.Fatal(e)
	}
	p, _, e := wmdesign.MaterializeCommercialPresentation(s)
	if e != nil {
		t.Fatal(e)
	}
	rows := p["rows"].([]any)
	if s.Model.Targets[0].Path != "/rows/1/amount" || rows[1].(map[string]any)["amount"] != "$1026.02" || rows[0].(map[string]any)["amount"] != "unchanged" {
		t.Fatalf("mapping moved identity %+v %+v", s.Model.Targets, rows)
	}
	if e := applyCommercial(&s, keys, CommercialOperation{Action: "remove", Entity: "collection", Path: "/rows", Key: "build", Cascade: true}); e == nil {
		t.Fatal("mapped item removed")
	}
}
func TestCommercialCompositionInvalidUnitsFieldsMissingAndCascade(t *testing.T) {
	t.Parallel()
	p, _ := commercialFixture(t)
	for _, mutate := range []func(*wmdesign.CommercialModel){func(m *wmdesign.CommercialModel) { m.Rows[0].Value = nil }, func(m *wmdesign.CommercialModel) { m.Rows[2].Assumption = "" }, func(m *wmdesign.CommercialModel) { m.Targets[0].Path = "/x" }, func(m *wmdesign.CommercialModel) { m.Rounding = "truncate" }} {
		m := commercialModel()
		mutate(&m)
		if _, e := PatchCommercial(p, "commercial-slide", commercialPatch(p, CommercialOperation{Action: "initialize", Entity: "source", Model: &m, Cascade: true}), bundle(t), wmdesign.CandidateEngine, true); e == nil {
			t.Fatal("invalid model applied")
		}
	}
	m := commercialModel()
	s := wmdesign.CommercialSpec{Model: m}
	if e := applyCommercial(&s, map[string][]string{}, CommercialOperation{Action: "remove", Entity: "row", Key: "hours", Cascade: true}); e == nil {
		t.Fatal("dependent input removed")
	}
	good := commercialPatch(p, CommercialOperation{Action: "initialize", Entity: "source", Model: &m, Cascade: true})
	raw := strings.Replace(string(canonical(good)), `"entity":"source"`, `"entity":"source","precision":0`, 1)
	if _, e := DecodeCommercialPatch([]byte(raw), "patch"); e == nil {
		t.Fatal("irrelevant explicit zero accepted")
	}
}
func TestCommercialCompositionMigratesFieldComments(t *testing.T) {
	t.Parallel()
	p, _ := commercialFixture(t)
	doc, e := sourceYAML(p.Raw)
	if e != nil {
		t.Fatal(e)
	}
	args := mappingNode(mappingNode(mappingNode(doc.Content[0], "local_templates"), "commercial"), "nodes").Content[0]
	value := mappingNode(mappingNode(args, "arguments"), "value")
	value.HeadComment = "Review the exact decimal basis"
	raw, e := encodeSourceYAML(doc)
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
	m := commercialModel()
	if _, e = PatchCommercial(p, "commercial-slide", commercialPatch(p, CommercialOperation{Action: "initialize", Entity: "source", Model: &m, Cascade: true}), bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	raw, _ = os.ReadFile(p.SourcePath)
	if !bytes.Contains(raw, []byte("Review the exact decimal basis")) {
		t.Fatal("migrated presentation field comment lost")
	}
}

func TestCommercialCompositionGroupUpdatesAllMappedDisplaysAtomically(t *testing.T) {
	t.Parallel()
	p, _ := commercialFixture(t)
	m := commercialModel()
	m.Targets[0].NodeID = "fee"
	m.Targets = append(m.Targets, wmdesign.CommercialTarget{NodeID: "assumption", Row: "fee", Path: "/text", Prefix: "Fee is $", Suffix: "; illustrative, excludes taxes"})
	patch := commercialPatch(p, CommercialOperation{Action: "initialize", Entity: "source", Nodes: []string{"assumption"}, Model: &m, Cascade: true})
	if _, e := PatchCommercial(p, "commercial-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	hours := m.Rows[0]
	hours.Value = commercialValue("20")
	patch = commercialPatch(p, CommercialOperation{Action: "set", Entity: "row", Key: "hours", Row: &hours})
	if _, e = PatchCommercial(p, "commercial-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	a, e := InspectCommercial(p, "commercial-slide", "fee", bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	b, e := InspectCommercial(p, "commercial-slide", "assumption", bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if a.Presentation["value"] != "$2002.00" || b.Presentation["text"] != "Fee is $2002.00; illustrative, excludes taxes" || !bytes.Equal(canonical(a.Model), canonical(b.Model)) {
		t.Fatalf("group drift %+v %+v", a, b)
	}
	clone := p.Document.LocalTemplates["commercial"]
	clone.Nodes[1].Arguments["model"].(map[string]any)["precision"] = 0
	p.Document.LocalTemplates["commercial"] = clone
	raw, _ := json.Marshal(p.Document)
	if e = os.WriteFile(p.SourcePath, raw, 0600); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = PatchCommercial(p, "commercial-slide", commercialPatch(p, CommercialOperation{Action: "set", Entity: "row", Key: "hours", Row: &hours}), bundle(t), wmdesign.CandidateEngine, true); e == nil {
		t.Fatal("inconsistent mirrored model accepted")
	}
}

func TestCommercialCompositionRepeatedRowAndTargetCountsKeepKeys(t *testing.T) {
	t.Parallel()
	p, _ := commercialFixture(t)
	m := commercialModel()
	if _, e := PatchCommercial(p, "commercial-slide", commercialPatch(p, CommercialOperation{Action: "initialize", Entity: "source", Model: &m, Cascade: true}), bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	row := wmdesign.CommercialRow{Key: "reserve", Label: "Reserve", Unit: "USD", Value: commercialValue("0"), Assumption: "Explicit zero reserve"}
	ops := []CommercialOperation{{Action: "set", Entity: "row", Key: "reserve", Row: &row}, {Action: "reorder", Entity: "row", Order: []string{"reserve", "fee", "rate", "hours"}}, {Action: "set", Entity: "targets", Targets: []wmdesign.CommercialTarget{{Row: "reserve", Path: "/value", Prefix: "$"}, {Row: "fee", Path: "/label", Prefix: "Fee $"}}}}
	if _, e = PatchCommercial(p, "commercial-slide", commercialPatch(p, ops...), bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	n := p.Document.LocalTemplates["commercial"].Nodes[0]
	if n.Keys["model/rows"][0] != "reserve" || len(n.Keys["model/targets"]) != 2 {
		t.Fatalf("keyed counts %+v", n.Keys)
	}
	ops = []CommercialOperation{{Action: "set", Entity: "targets", Targets: []wmdesign.CommercialTarget{{Row: "fee", Path: "/value", Prefix: "$"}}}, {Action: "remove", Entity: "row", Key: "reserve", Cascade: true}}
	if _, e = PatchCommercial(p, "commercial-slide", commercialPatch(p, ops...), bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	n = p.Document.LocalTemplates["commercial"].Nodes[0]
	if len(n.Keys["model/rows"]) != 3 || n.Keys["model/rows"][0] != "fee" || len(n.Keys["model/targets"]) != 1 {
		t.Fatalf("retained keys %+v", n.Keys)
	}
}
