package deckproject

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func quantitativeFixture(t *testing.T, kind string) (*Project, string) {
	t.Helper()
	p := example(t)
	options := wmdesign.FrameRequest{Rail: "none", Footer: "compact", TitleLines: 1, Density: "appendix", Surface: "light"}
	source := map[string]any{"kind": kind, "categories": []any{"Current", "Future"}, "series": []any{map[string]any{"name": "Capacity", "values": []any{0.0, 3.0}}}, "units": "FTE", "source": "Illustrative capacity assumptions"}
	keys := map[string][]string{"categories": {"current", "future"}, "series": {"capacity"}}
	if kind == "scatter" {
		source = map[string]any{"kind": "scatter", "points": []any{[]any{1.0, 2.0, "One"}, []any{3.0, 4.0, "Two"}}, "units": "Index", "source": "Illustrative observations"}
		keys = map[string][]string{"points": {"one", "two"}}
	}
	if kind == "quadrant" {
		source = map[string]any{"kind": "quadrant", "items": []any{map[string]any{"x": .2, "y": .3, "label": "One"}}, "quadrants": map[string]any{"tl": map[string]any{"name": "TL"}, "tr": map[string]any{"name": "TR"}, "bl": map[string]any{"name": "BL"}, "br": map[string]any{"name": "BR"}}, "positionMode": "qualitative"}
		keys = map[string][]string{"items": {"one"}}
	}
	local := LocalTemplate{Name: "Illustrative chart", Frame: Reference{Scope: "shared", ID: "wmds/frame/none-compact"}, FrameOptions: &options, Grid: Reference{Scope: "shared", ID: "wmds/grid/12-columns"}, Zones: map[string]Zone{"title": {Role: "slide-title", Required: true, Schema: map[string]any{"type": "string"}}}, Nodes: []Node{{ID: "chart", Kind: "component", Definition: &Reference{Scope: "shared", ID: "wmds/component/chart"}, Placement: &Placement{Zone: "body", Rect: &wmdesign.Rect{X: 0, Y: 0, W: 840, H: 270}}, Arguments: source, Keys: keys}}}
	p.Document.LocalTemplates = map[string]LocalTemplate{"quantitative": local}
	p.Document.Slides = []Slide{{ID: "quantitative-slide", ContentKind: "synthetic_example", Template: Reference{Scope: "local", ID: "quantitative"}, Values: map[string]any{"title": "Illustrative chart changes retain keyed facts"}}}
	raw, _ := json.Marshal(p.Document)
	if e := os.WriteFile(p.SourcePath, raw, 0600); e != nil {
		t.Fatal(e)
	}
	p, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	pin(t, p)
	return p, "chart"
}
func quantitativePatch(p *Project, ops ...QuantitativeOperation) QuantitativePatch {
	return QuantitativePatch{Schema: QuantitativePatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "Qualification", Reason: "Customize explicit illustrative values", NodeID: "chart", Operations: ops}
}
func qnumber(v float64) *float64 { return &v }
func TestQuantitativeCompositionKeyedPreviewApplyAndRebuild(t *testing.T) {
	t.Parallel()
	p, _ := quantitativeFixture(t, "column")
	before := append([]byte(nil), p.Raw...)
	patch := quantitativePatch(p, QuantitativeOperation{Action: "set", Entity: "category", Key: "new", Label: "New"}, QuantitativeOperation{Action: "set", Entity: "value", Key: "new", Series: "capacity", Value: qnumber(5)}, QuantitativeOperation{Action: "reorder", Entity: "category", Order: []string{"new", "future", "current"}})
	out, e := PatchQuantitative(p, "quantitative-slide", patch, bundle(t), wmdesign.CandidateEngine, false)
	if e != nil {
		t.Fatal(e)
	}
	disk, _ := os.ReadFile(p.SourcePath)
	if out.Applied || !bytes.Equal(disk, before) {
		t.Fatal("preview changed source")
	}
	out, e = PatchQuantitative(p, "quantitative-slide", patch, bundle(t), wmdesign.CandidateEngine, true)
	if e != nil || !out.Applied {
		t.Fatalf("apply %v", e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	inspect, e := InspectQuantitative(p, "quantitative-slide", "chart", bundle(t), wmdesign.CandidateEngine)
	if e != nil || inspect.RenderError != "" {
		t.Fatalf("inspect %+v %v", inspect, e)
	}
	values := inspect.Source["series"].([]any)[0].(map[string]any)["values"].([]any)
	if values[0] != 5.0 || values[2] != 0.0 || inspect.Keys["categories"][2] != "current" {
		t.Fatal("reorder detached values")
	}
	if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	if _, e = PatchQuantitative(p, "quantitative-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e == nil {
		t.Fatal("stale accepted")
	}
}
func TestQuantitativeCompositionNullZeroKindsAndCollectionCounts(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"line", "bar", "pie", "doughnut", "scatter", "quadrant"} {
		t.Run(kind, func(t *testing.T) {
			p, _ := quantitativeFixture(t, kind)
			var ops []QuantitativeOperation
			switch kind {
			case "scatter":
				ops = []QuantitativeOperation{{Action: "set", Entity: "point", Key: "three", Point: []any{5.0, 6.0, "Three"}}, {Action: "reorder", Entity: "point", Order: []string{"three", "two", "one"}}}
			case "quadrant":
				ops = []QuantitativeOperation{{Action: "set", Entity: "item", Key: "two", Data: map[string]any{"x": .8, "y": .7, "label": "Two"}}, {Action: "reorder", Entity: "item", Order: []string{"two", "one"}}}
			case "line":
				ops = []QuantitativeOperation{{Action: "set", Entity: "value", Key: "future", Series: "capacity", Missing: true}}
			default:
				ops = []QuantitativeOperation{{Action: "set", Entity: "value", Key: "current", Series: "capacity", Value: qnumber(0)}}
			}
			if _, e := PatchQuantitative(p, "quantitative-slide", quantitativePatch(p, ops...), bundle(t), wmdesign.CandidateEngine, true); e != nil {
				t.Fatal(e)
			}
		})
	}
}
func TestQuantitativeCompositionInvalidFieldsBindingsAndDeletion(t *testing.T) {
	t.Parallel()
	p, _ := quantitativeFixture(t, "column")
	for _, op := range []QuantitativeOperation{{Action: "remove", Entity: "category", Key: "current"}, {Action: "set", Entity: "value", Key: "current", Series: "capacity"}, {Action: "set", Entity: "value", Key: "current", Series: "capacity", Value: qnumber(0), Missing: true}, {Action: "set", Entity: "value", Key: "current", Series: "capacity", Value: qnumber(1e10)}, {Action: "reorder", Entity: "category", Order: []string{"future", "future"}}, {Action: "set", Entity: "setting", Field: "x", Setting: 0}, {Action: "set", Entity: "value", Key: "future", Series: "capacity", Missing: true}} {
		before, _ := os.ReadFile(p.SourcePath)
		if _, e := PatchQuantitative(p, "quantitative-slide", quantitativePatch(p, op), bundle(t), wmdesign.CandidateEngine, true); e == nil {
			t.Fatalf("invalid accepted %+v", op)
		}
		after, _ := os.ReadFile(p.SourcePath)
		if !bytes.Equal(before, after) {
			t.Fatal("failed candidate changed source")
		}
	}
	good := quantitativePatch(p, QuantitativeOperation{Action: "set", Entity: "category", Key: "current", Label: "Current"})
	raw := strings.Replace(string(canonical(good)), `"label":"Current"`, `"label":"Current","missing":false`, 1)
	if _, e := DecodeQuantitativePatch([]byte(raw), "patch"); e == nil {
		t.Fatal("irrelevant authored false accepted")
	}
	if _, e := PatchQuantitative(p, "quantitative-slide", quantitativePatch(p, QuantitativeOperation{Action: "remove", Entity: "category", Key: "current", Cascade: true}), bundle(t), wmdesign.CandidateEngine, false); e != nil {
		t.Fatal(e)
	}
}

func TestQuantitativeCompositionExplicitSourceUnionMigration(t *testing.T) {
	t.Parallel()
	p, _ := quantitativeFixture(t, "column")
	source := map[string]any{"kind": "scatter", "points": []any{[]any{-2.0, 3.0, "Lower"}, []any{4.0, 5.0, "Upper"}}, "units": "Index", "source": "Explicit illustrative migration", "xMin": -3.0, "xMax": 6.0, "yMin": 0.0, "yMax": 7.0}
	op := QuantitativeOperation{Action: "replace", Entity: "source", Cascade: true, Data: source, Keys: map[string][]string{"points": {"lower", "upper"}}}
	if _, e := PatchQuantitative(p, "quantitative-slide", quantitativePatch(p, op), bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	inspect, e := InspectQuantitative(p, "quantitative-slide", "chart", bundle(t), wmdesign.CandidateEngine)
	if e != nil || inspect.RenderError != "" {
		t.Fatalf("migration %v %s", e, inspect.RenderError)
	}
	if inspect.Source["source"] != source["source"] || inspect.Keys["points"][1] != "upper" {
		t.Fatal("source/key lost")
	}
	bad := op
	bad.Cascade = false
	if _, e = PatchQuantitative(p, "quantitative-slide", quantitativePatch(p, bad), bundle(t), wmdesign.CandidateEngine, true); e == nil {
		t.Fatal("unacknowledged union replacement accepted")
	}
}
func TestQuantitativeCompositionRejectsMalformedIntermediateSeriesWithoutPanic(t *testing.T) {
	t.Parallel()
	p, _ := quantitativeFixture(t, "column")
	ops := []QuantitativeOperation{{Action: "set", Entity: "series", Key: "capacity", Data: map[string]any{"name": "Capacity"}}, {Action: "reorder", Entity: "category", Order: []string{"future", "current"}}}
	if _, e := PatchQuantitative(p, "quantitative-slide", quantitativePatch(p, ops...), bundle(t), wmdesign.CandidateEngine, true); e == nil {
		t.Fatal("malformed series accepted")
	}
}

func TestQuantitativeCompositionCountChangesRetainColorHighlightAndNestedKeys(t *testing.T) {
	t.Parallel()
	p, _ := quantitativeFixture(t, "pie")
	ops := []QuantitativeOperation{{Action: "set", Entity: "setting", Field: "colors", Setting: []any{"series.1", "series.2"}}, {Action: "set", Entity: "setting", Field: "highlight", Setting: []any{1.0}}}
	if _, e := PatchQuantitative(p, "quantitative-slide", quantitativePatch(p, ops...), bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	n := p.Document.LocalTemplates["quantitative"].Nodes[0]
	colorKey := n.Keys["colors"][0]
	ops = []QuantitativeOperation{{Action: "set", Entity: "category", Key: "new", Label: "New"}, {Action: "set", Entity: "value", Key: "new", Series: "capacity", Value: qnumber(2)}, {Action: "remove", Entity: "category", Key: "future", Cascade: true}, {Action: "reorder", Entity: "category", Order: []string{"new", "current"}}}
	if _, e = PatchQuantitative(p, "quantitative-slide", quantitativePatch(p, ops...), bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	n = p.Document.LocalTemplates["quantitative"].Nodes[0]
	if len(n.Keys["colors"]) != 2 || n.Keys["colors"][1] != colorKey || len(n.Keys["highlight"]) != 0 || len(n.Keys["series/0/values"]) != 2 {
		t.Fatalf("color/highlight/nested keys detached %+v args=%+v", n.Keys, n.Arguments)
	}
	colors := n.Arguments["colors"].([]any)
	if colors[1] != "series.1" {
		t.Fatal("original category color moved")
	}
}
