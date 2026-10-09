package deckproject

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func quantitativeV11(t *testing.T) string {
	t.Helper()
	b, e := filepath.Abs("../../library/wm-design-system/v11")
	if e != nil {
		t.Fatal(e)
	}
	return b
}

type quantitativeProfileEntry struct {
	Key, Family string
	Profiles    []string       `json:"composition_profiles"`
	SourceTypes map[string]int `json:"source_type_counts"`
}

func profileTemplates(t *testing.T) []quantitativeProfileEntry {
	t.Helper()
	raw, e := os.ReadFile("../../docs/template-customization-inventory.json")
	if e != nil {
		t.Fatal(e)
	}
	var inv struct {
		Templates []quantitativeProfileEntry `json:"templates"`
	}
	if e = json.Unmarshal(raw, &inv); e != nil {
		t.Fatal(e)
	}
	return inv.Templates
}
func sourceProfileProject(t *testing.T, b, key string) *Project {
	t.Helper()
	scaffold, e := ScaffoldTemplate(b, key, wmdesign.CandidateEngine, "Qualify explicit illustrative source customization", 2026)
	if e != nil {
		t.Fatal(e)
	}
	p := example(t)
	p.Document.Assets = nil
	p.Document.Context = nil
	p.Document.LocalTemplates = map[string]LocalTemplate{"catalog": scaffold.Template}
	p.Document.Slides = []Slide{{ID: "catalog-slide", ContentKind: "synthetic_example", Template: Reference{Scope: "local", ID: "catalog"}, Values: scaffold.SyntheticSourceValues}}
	raw, _ := json.Marshal(p.Document)
	if e = os.WriteFile(p.SourcePath, raw, 0600); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = Pin(p, b, wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	return p
}
func exportCompositionFixture(t *testing.T, p *Project, name string, patch any) {
	t.Helper()
	root := os.Getenv("PPTXGENGO_COMPOSITION_FIXTURES")
	if root == "" {
		return
	}
	target := filepath.Join(root, name)
	if e := os.MkdirAll(target, 0700); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(filepath.Join(target, "deck.yaml"), p.Raw, 0600); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(filepath.Join(target, "patch.json"), canonical(patch), 0600); e != nil {
		t.Fatal(e)
	}
}
func TestQuantitativeCatalogV11EveryChartSource(t *testing.T) {
	if testing.Short() {
		t.Skip("exhaustive source catalog audit runs nightly/on demand; representative composition regressions remain in short suite")
	}
	b := quantitativeV11(t)
	variants, charts, noncharts := 0, 0, 0
	for _, entry := range profileTemplates(t) {
		if teamFind(entry.Profiles, "quantitative-chart") < 0 {
			continue
		}
		variants++
		t.Run(strings.ReplaceAll(entry.Key, "/", "-"), func(t *testing.T) {
			if entry.SourceTypes["chart"] == 0 {
				noncharts++
				t.Log("No chart source: numeric headline/content/media/table/assessment semantic route")
				return
			}
			p := sourceProfileProject(t, b, entry.Key)
			found := 0
			for _, n := range p.Document.LocalTemplates["catalog"].Nodes {
				if n.Definition == nil || n.Definition.ID != "wmds/component/chart" {
					continue
				}
				found++
				charts++
				inspect, e := InspectQuantitative(p, "catalog-slide", n.ID, b, wmdesign.CandidateEngine)
				if e != nil {
					t.Fatal(e)
				}
				ops := []QuantitativeOperation{{Action: "materialize", Entity: "source"}, {Action: "set", Entity: "setting", Field: "preserveWorkbookZeros", Setting: true}}
				kind, _ := inspect.Source["kind"].(string)
				if kind == "quadrant" {
					ids := inspect.Keys["items"]
					order := append([]string(nil), ids...)
					for i, j := 0, len(order)-1; i < j; i, j = i+1, j-1 {
						order[i], order[j] = order[j], order[i]
					}
					ops = append(ops, QuantitativeOperation{Action: "reorder", Entity: "item", Order: order})
				} else {
					ids := inspect.Keys["categories"]
					order := append([]string(nil), ids...)
					for i, j := 0, len(order)-1; i < j; i, j = i+1, j-1 {
						order[i], order[j] = order[j], order[i]
					}
					ops = append(ops, QuantitativeOperation{Action: "reorder", Entity: "category", Order: order})
				}
				patch := QuantitativePatch{Schema: QuantitativePatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "Catalog qualification", Reason: "Reverse illustrative keyed order while preserving value/category relationships and series styling", NodeID: n.ID, Operations: ops}
				if entry.Key == "chart/column-rail" || entry.Key == "value-curve/scenarios" {
					exportCompositionFixture(t, p, strings.ReplaceAll(entry.Key, "/", "-"), patch)
				}
				if _, e = PatchQuantitative(p, "catalog-slide", patch, b, wmdesign.CandidateEngine, false); e != nil {
					t.Fatal(e)
				}
			}
			if found == 0 {
				noncharts++
				t.Log("No chart source: numeric headline/content/media/table/assessment profile mapping, not a chart data contract")
			}
		})
	}
	if variants != 36 {
		t.Fatalf("inventory changed: %d variants", variants)
	}
	t.Logf("V11 quantitative profile: %d variants, %d chart nodes exercised, %d variants routed to non-chart semantics", variants, charts, noncharts)
}
func firstCommercialLiteral(v any, path string) (string, bool, bool) {
	switch x := v.(type) {
	case map[string]any:
		// Monetary display data may use real authored table column IDs (m/a/c),
		// not only invented c0/c1 fixture fields. Never map source geometry/style.
		names := []string{}
		for key := range x {
			switch key {
			case "type", "id", "k", "x", "y", "w", "h", "size", "_source_geometry", "color", "fill", "style":
				continue
			}
			names = append(names, key)
		}
		sort.Strings(names)
		for _, key := range names {
			if p, n, ok := firstCommercialLiteral(x[key], path+"/"+escape(key)); ok {
				return p, n, true
			}
		}
	case []any:
		for i, item := range x {
			if p, n, ok := firstCommercialLiteral(item, fmt.Sprintf("%s/%d", path, i)); ok {
				return p, n, true
			}
		}
	case string:
		if strings.Contains(x, "$") {
			return path, false, true
		}
	}
	return "", false, false
}
func TestCommercialCatalogV11EveryCommercialAndValueVariant(t *testing.T) {
	if testing.Short() {
		t.Skip("exhaustive source catalog audit runs nightly/on demand; representative composition regressions remain in short suite")
	}
	b := quantitativeV11(t)
	variants, initialized, assumptions := 0, 0, 0
	for _, entry := range profileTemplates(t) {
		if entry.Family != "commercials" && teamFind(entry.Profiles, "value-model") < 0 {
			continue
		}
		variants++
		t.Run(strings.ReplaceAll(entry.Key, "/", "-"), func(t *testing.T) {
			p := sourceProfileProject(t, b, entry.Key)
			var selected *Node
			path := ""
			numeric := false
			for _, node := range p.Document.LocalTemplates["catalog"].Nodes {
				if node.Definition == nil {
					continue
				}
				source, e := resolveArguments(node.Arguments, p.Document.Slides[0].Values)
				if e != nil {
					t.Fatal(e)
				}
				target, num, ok := firstCommercialLiteral(source, "")
				if ok {
					n := node
					selected = &n
					path, numeric = target, num
					break
				}
			}
			if selected == nil {
				assumptions++
				qualified := false
				for _, n := range p.Document.LocalTemplates["catalog"].Nodes {
					if n.Definition == nil {
						continue
					}
					route := componentSpecialized(strings.TrimPrefix(n.Definition.ID, "wmds/component/"))
					if route == "table" {
						inspect, e := InspectTable(p, "catalog-slide", n.ID, b, wmdesign.CandidateEngine)
						if e != nil {
							t.Fatal(e)
						}
						order := []string{}
						for i := len(inspect.Model.Rows) - 1; i >= 0; i-- {
							order = append(order, inspect.Model.Rows[i].Key)
						}
						ops := []TableOperation{{Action: "materialize", Entity: "source"}}
						if len(order) > 1 {
							ops = append(ops, TableOperation{Action: "reorder", Entity: "row", Order: order})
						}
						patch := TablePatch{Schema: TablePatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "Catalog qualification", Reason: "Reorder illustrative scope rows preserving source cells", SemanticReview: "No inferred financial calculation; source row order reviewed", NodeID: n.ID, Operations: ops}
						if _, e = PatchTable(p, "catalog-slide", patch, b, wmdesign.CandidateEngine, false); e != nil {
							t.Fatal(e)
						}
						qualified = true
						break
					}
					if route != "" {
						continue
					}
					inspect, e := InspectComponent(p, "catalog-slide", n.ID, b, wmdesign.CandidateEngine)
					if e != nil {
						t.Fatal(e)
					}
					ops := []ComponentOperation{{Action: "materialize", Entity: "source"}}
					for _, collection := range inspect.Collections {
						if collection.Count < 2 || strings.Count(collection.Path, "/") != 1 {
							continue
						}
						order := append([]string(nil), collection.Keys...)
						for i, j := 0, len(order)-1; i < j; i, j = i+1, j-1 {
							order[i], order[j] = order[j], order[i]
						}
						ops = append(ops, ComponentOperation{Action: "reorder", Entity: "item", Path: collection.Path, Order: order})
						break
					}
					patch := ComponentPatch{Schema: ComponentPatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "Catalog qualification", Reason: "Customize source scope assumptions preserving all supplied facts", SemanticReview: "Illustrative scope order reviewed; no inferred financial calculation", NodeID: n.ID, Operations: ops}
					if _, e = PatchComponent(p, "catalog-slide", patch, b, wmdesign.CandidateEngine, false); e != nil {
						t.Fatal(e)
					}
					qualified = true
					break
				}
				if !qualified {
					t.Fatal("scope-only variant has no qualified source component route")
				}
				t.Log("Scope/assumptions-only composition: no authored monetary field; retains source text/table editing and explicit optional new modeled metric")
				return
			}
			initialized++
			model := wmdesign.CommercialModel{Precision: 0, Rounding: "half_even", Assumptions: []string{"Illustrative source replacement for composition qualification only"}, Rows: []wmdesign.CommercialRow{{Key: "modeled", Label: "Illustrative modeled value", Unit: "USD", Value: commercialValue("123"), Assumption: "Qualification sample, not a business estimate"}}, Targets: []wmdesign.CommercialTarget{{Row: "modeled", Path: path, Numeric: numeric}}}
			if !numeric {
				model.Targets[0].Prefix = "$"
			}
			patch := CommercialPatch{Schema: CommercialPatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "Catalog qualification", Reason: "Explicitly map one illustrative numeric display, preserving all other source components", NodeID: selected.ID, Operations: []CommercialOperation{{Action: "initialize", Entity: "source", Model: &model, Cascade: true}}}
			if entry.Key == "pricing/fixed-fee" || entry.Key == "pricing/options" || entry.Key == "value-summary/big-number" {
				exportCompositionFixture(t, p, strings.ReplaceAll(entry.Key, "/", "-"), patch)
			}
			if _, e := PatchCommercial(p, "catalog-slide", patch, b, wmdesign.CandidateEngine, false); e != nil {
				t.Fatal(e)
			}
		})
	}
	if variants != 55 {
		t.Fatalf("inventory changed: %d variants", variants)
	}
	t.Logf("V11 commercial/value: %d variants, %d explicit display migrations exercised, %d scope/assumption-only variants", variants, initialized, assumptions)
}

func TestQuantitativeCatalogV11CountGrowthSourceKeysAndUnits(t *testing.T) {
	b := quantitativeV11(t)
	p := sourceProfileProject(t, b, "chart/column-rail")
	n := p.Document.LocalTemplates["catalog"].Nodes[0]
	inspect, e := InspectQuantitative(p, "catalog-slide", n.ID, b, wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	ops := []QuantitativeOperation{{Action: "materialize", Entity: "source"}, {Action: "set", Entity: "setting", Field: "preserveWorkbookZeros", Setting: true}, {Action: "set", Entity: "category", Key: "wave-extra", Label: "Extra"}}
	for _, series := range inspect.Keys["series"] {
		ops = append(ops, QuantitativeOperation{Action: "set", Entity: "value", Key: "wave-extra", Series: series, Value: qnumber(4)})
	}
	patch := QuantitativePatch{Schema: QuantitativePatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "Catalog qualification", Reason: "Add one explicitly supplied illustrative wave preserving days units", NodeID: n.ID, Operations: ops}
	exportCompositionFixture(t, p, "chart-column-rail-growth", patch)
	if _, e = PatchQuantitative(p, "catalog-slide", patch, b, wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	inspect, e = InspectQuantitative(p, "catalog-slide", n.ID, b, wmdesign.CandidateEngine)
	if e != nil || inspect.RenderError != "" {
		t.Fatalf("inspect %v %s", e, inspect.RenderError)
	}
	if len(inspect.Keys["categories"]) != 4 || inspect.Source["units"] != "Days to close" {
		t.Fatal("count/unit meaning lost")
	}
	if _, e = Build(p, BuildOptions{Bundle: b, Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
}

func TestQuantitativeCatalogV11QuadrantScaffoldGeometryPreservesPublishedAllocation(t *testing.T) {
	b := quantitativeV11(t)
	for _, key := range []string{"quadrant/subtle", "quadrant/subtle-left", "quadrant/numbered-legend", "quadrant/strong"} {
		t.Run(strings.ReplaceAll(key, "/", "-"), func(t *testing.T) {
			refs, e := wmdesign.LibraryReference(b, "", "evidence", 2026)
			if e != nil {
				t.Fatal(e)
			}
			for _, slide := range refs.Slides {
				if slide.Template == key {
					refs.Slides = []wmdesign.BoundSlide{slide}
					break
				}
			}
			original, _, e := wmdesign.BindTemplates(b, "", refs)
			if e != nil {
				t.Fatal(e)
			}
			p := sourceProfileProject(t, b, key)
			compiled, e := Compile(p, b, wmdesign.CandidateEngine)
			if e != nil {
				t.Fatal(e)
			}
			chartGeometry := func(doc wmdesign.Document) map[string]any {
				for _, slide := range doc.Slides {
					for _, node := range slide.Nodes {
						if node.Scene == nil {
							continue
						}
						var src map[string]any
						if e := json.Unmarshal(node.Scene.Node, &src); e != nil {
							t.Fatal(e)
						}
						if src["type"] == "chart" && src["kind"] == "quadrant" {
							return src
						}
					}
				}
				t.Fatal("quadrant chart missing")
				return nil
			}
			a, c := chartGeometry(original), chartGeometry(compiled.Document)
			for _, field := range []string{"x", "y", "w", "h", "kind", "style", "fill", "xTitle", "yTitle"} {
				if string(canonical(a[field])) != string(canonical(c[field])) {
					t.Fatalf("published %s changed: %v -> %v", field, a[field], c[field])
				}
			}
			if _, _, e = wmdesign.BuildWithEngine(b, "", compiled.Document, wmdesign.CandidateEngine); e != nil {
				t.Fatal(e)
			}
		})
	}
}

func TestCommercialCatalogV11FixedFeeCoherentGroupFixture(t *testing.T) {
	b := quantitativeV11(t)
	p := sourceProfileProject(t, b, "pricing/fixed-fee")
	table, summary := "", ""
	for _, n := range p.Document.LocalTemplates["catalog"].Nodes {
		if n.Definition == nil {
			continue
		}
		switch n.Definition.ID {
		case "wmds/component/table":
			table = n.ID
		case "wmds/component/feesummary":
			summary = n.ID
		}
	}
	if table == "" || summary == "" {
		t.Fatal("actual fixed-fee table/summary missing")
	}
	model := wmdesign.CommercialModel{Precision: 0, Rounding: "half_even", Assumptions: []string{"Illustrative fixed-fee plan; percent allocations unchanged; excludes expenses and taxes"}}
	for i, value := range []string{"100000", "300000", "400000", "200000"} {
		key := fmt.Sprintf("milestone-%d", i+1)
		model.Rows = append(model.Rows, wmdesign.CommercialRow{Key: key, Label: fmt.Sprintf("Milestone %d", i+1), Unit: "USD", Value: commercialValue(value), Assumption: "Explicit illustrative milestone fee"})
		inputs := []string{}
		for j := 0; j <= i; j++ {
			inputs = append(inputs, fmt.Sprintf("milestone-%d", j+1))
		}
		cumulative := fmt.Sprintf("cumulative-%d", i+1)
		model.Rows = append(model.Rows, wmdesign.CommercialRow{Key: cumulative, Label: "Cumulative fee", Unit: "USD", Formula: &wmdesign.CommercialFormula{Operation: "sum", Inputs: inputs}})
		model.Targets = append(model.Targets, wmdesign.CommercialTarget{NodeID: table, Row: key, Path: fmt.Sprintf("/rows/%d/a", i), Prefix: "$", Suffix: "K", Scale: "1000"}, wmdesign.CommercialTarget{NodeID: table, Row: cumulative, Path: fmt.Sprintf("/rows/%d/c", i), Prefix: "$", Suffix: "K", Scale: "1000"})
	}
	model.Rows = append(model.Rows, wmdesign.CommercialRow{Key: "total", Label: "Total fixed fee", Unit: "USD", Formula: &wmdesign.CommercialFormula{Operation: "sum", Inputs: []string{"milestone-1", "milestone-2", "milestone-3", "milestone-4"}}}, wmdesign.CommercialRow{Key: "phase-1", Label: "Diagnose fee", Unit: "USD", Formula: &wmdesign.CommercialFormula{Operation: "sum", Inputs: []string{"milestone-1", "milestone-2"}}}, wmdesign.CommercialRow{Key: "phase-2", Label: "Design fee", Unit: "USD", Formula: &wmdesign.CommercialFormula{Operation: "sum", Inputs: []string{"milestone-3", "milestone-4"}}})
	for _, target := range []wmdesign.CommercialTarget{{NodeID: table, Row: "total", Path: "/rows/4/a"}, {NodeID: summary, Row: "total", Path: "/total"}, {NodeID: summary, Row: "phase-1", Path: "/lines/0/1"}, {NodeID: summary, Row: "phase-2", Path: "/lines/1/1"}} {
		target.Prefix = "$"
		target.Suffix = "K"
		target.Scale = "1000"
		model.Targets = append(model.Targets, target)
	}
	patch := CommercialPatch{Schema: CommercialPatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "Catalog qualification", Reason: "Explicitly price the illustrative fixed-fee plan with synchronized milestone, cumulative and phase totals", NodeID: table, Operations: []CommercialOperation{{Action: "initialize", Entity: "source", Model: &model, Nodes: []string{summary}, Cascade: true}}}
	exportCompositionFixture(t, p, "pricing-fixed-fee", patch)
	metadata := map[string][]byte{}
	for _, n := range p.Document.LocalTemplates["catalog"].Nodes {
		metadata[n.ID] = canonical(n.Arguments["_source_geometry"])
	}
	if _, e := PatchCommercial(p, "catalog-slide", patch, b, wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	for _, id := range []string{table, summary} {
		inspect, e := InspectCommercial(p, "catalog-slide", id, b, wmdesign.CandidateEngine)
		if e != nil || inspect.RenderError != "" {
			t.Fatalf("inspect %v %s", e, inspect.RenderError)
		}
		if string(canonical(inspect.Presentation["_source_geometry"])) != string(metadata[id]) {
			t.Fatal("source geometry metadata changed")
		}
		if id == summary && inspect.Presentation["total"] != "$1000K" {
			t.Fatal("total not synchronized")
		}
	}
	if _, e = Build(p, BuildOptions{Bundle: b, Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
}
