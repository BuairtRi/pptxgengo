package deckproject

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func curveCompositionBundle(t *testing.T) string {
	t.Helper()
	path := os.Getenv("PPTXGENGO_QUALIFICATION_BUNDLE")
	if path == "" {
		path = "../../library/wm-design-system/v11"
	}
	absolute, e := filepath.Abs(path)
	if e != nil {
		t.Fatal(e)
	}
	return absolute
}
func curveFixture(t *testing.T, catalogKey, kind string) (*Project, string) {
	t.Helper()
	p := example(t)
	catalog, e := wmdesign.LibraryCatalog(curveCompositionBundle(t), "")
	if e != nil {
		t.Fatal(e)
	}
	var args map[string]any
	for _, d := range catalog {
		if d.Key != catalogKey {
			continue
		}
		var slide struct {
			Body []map[string]any `json:"body"`
		}
		if e = json.Unmarshal(d.RawSlide, &slide); e != nil {
			t.Fatal(e)
		}
		for _, n := range slide.Body {
			if n["type"] == kind {
				args = n
				break
			}
		}
	}
	if args == nil {
		t.Fatalf("missing catalog %s/%s", catalogKey, kind)
	}
	for _, k := range []string{"type", "id", "x", "y", "w", "h"} {
		delete(args, k)
	}
	rect := wmdesign.Rect{X: 0, Y: 0, W: 846, H: 330}
	options := wmdesign.FrameRequest{Rail: "none", Footer: "compact", TitleLines: 1, Density: "appendix", Surface: "light"}
	n := Node{ID: "model", Kind: "component", Definition: &Reference{Scope: "shared", ID: "wmds/component/" + kind}, Placement: &Placement{Zone: "body", Rect: &rect}, Arguments: args, Keys: map[string][]string{}}
	for _, path := range []string{"stages", "series", "phases"} {
		if list, ok := args[path].([]any); ok {
			for i := range list {
				n.Keys[path] = append(n.Keys[path], "source-"+strings.Repeat("x", i+1))
			}
		}
	}
	local := LocalTemplate{Name: "Catalog-derived curve fixture", Frame: Reference{Scope: "shared", ID: "wmds/frame/none-compact"}, FrameOptions: &options, Grid: Reference{Scope: "shared", ID: "wmds/grid/12-columns"}, Zones: map[string]Zone{"title": {Role: "slide-title", Required: true, Schema: map[string]any{"type": "string"}}}, Nodes: []Node{n}}
	p.Document.LocalTemplates = map[string]LocalTemplate{"fixture": local}
	p.Document.Slides = []Slide{{ID: "slide", ContentKind: "synthetic_example", Template: Reference{Scope: "local", ID: "fixture"}, Values: map[string]any{"title": "Illustrative model changes with explicit meaning"}}}
	return writeCurveFixture(t, p), "model"
}
func writeCurveFixture(t *testing.T, p *Project) *Project {
	t.Helper()
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
	if _, err := os.Stat(p.Root + "/toolchain.lock.json"); os.IsNotExist(err) {
		if _, e := Pin(p, curveCompositionBundle(t), wmdesign.CandidateEngine); e != nil {
			t.Fatal(e)
		}
	}
	return p
}
func maturityPatch(p *Project, ops ...MaturityOperation) MaturityPatch {
	return MaturityPatch{Schema: MaturityPatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "qualification", Reason: "Adapt synthetic catalog maturity", NodeID: "model", Operations: ops}
}
func staffingPatch(p *Project, ops ...StaffingOperation) StaffingPatch {
	return StaffingPatch{Schema: StaffingPatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "qualification", Reason: "Adapt synthetic catalog staffing", NodeID: "model", Operations: ops}
}
func TestMaturityCompositionCatalogCountReferencesAndAtomicApply(t *testing.T) {
	t.Parallel()
	p, id := curveFixture(t, "maturity/four-stage", "maturity")
	before := append([]byte(nil), p.Raw...)
	ins, e := InspectMaturity(p, "slide", id, curveCompositionBundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if len(ins.Model.Stages) != 4 {
		t.Fatal(ins.Model)
	}
	keys := []string{}
	for _, s := range ins.Model.Stages {
		keys = append(keys, s.Key)
	}
	sh := 2.0
	lw := 110.
	r := wmdesign.Rect{X: 0, Y: 0, W: 846, H: 330}
	patch := maturityPatch(p, MaturityOperation{Action: "set", Entity: "stage", Key: "future", Stage: &MaturityStage{Key: "future", Label: "Future", At: .98}}, MaturityOperation{Action: "set", Entity: "stage", Key: "scale", Stage: &MaturityStage{Key: "scale", Label: "Scale", At: .99}}, MaturityOperation{Action: "reorder", Entity: "stage", Order: append([]string{keys[0], "future", keys[1], keys[2], "scale"}, keys[3]), Spacing: "even"}, MaturityOperation{Action: "set", Entity: "current", Key: keys[1]}, MaturityOperation{Action: "set", Entity: "here", Marker: &MaturityMarker{keys[1], "Current"}}, MaturityOperation{Action: "set", Entity: "target", Marker: &MaturityMarker{"scale", "Target"}}, MaturityOperation{Action: "set", Entity: "branch", Branch: &MaturityBranch{From: keys[2], Label: "Alternate", Text: ""}}, MaturityOperation{Action: "set", Entity: "layout", Layout: &MaturityLayout{Rect: &r, Shape: &sh, LabelWidth: &lw}})
	preview, e := PatchMaturity(p, "slide", patch, curveCompositionBundle(t), wmdesign.CandidateEngine, false)
	if e != nil {
		t.Fatal(e)
	}
	if preview.Applied {
		t.Fatal("preview applied")
	}
	disk, _ := os.ReadFile(p.SourcePath)
	if !bytes.Equal(disk, before) {
		t.Fatal("preview changed source")
	}
	ap, e := PatchMaturity(p, "slide", patch, curveCompositionBundle(t), wmdesign.CandidateEngine, true)
	if e != nil {
		t.Fatal(e)
	}
	if !ap.Applied || ap.Decision == "" {
		t.Fatal(ap)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	ins, e = InspectMaturity(p, "slide", id, curveCompositionBundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if len(ins.Model.Stages) != 6 || ins.Model.Current != keys[1] || ins.Model.Target.Stage != "scale" || ins.Model.Branch.From != keys[2] {
		t.Fatal(ins.Model)
	}
	if _, e = PatchMaturity(p, "slide", patch, curveCompositionBundle(t), wmdesign.CandidateEngine, false); e == nil {
		t.Fatal("stale accepted")
	}
	ops := []MaturityOperation{{Action: "remove", Entity: "stage", Key: keys[1]}}
	if _, e = PatchMaturity(p, "slide", maturityPatch(p, ops...), curveCompositionBundle(t), wmdesign.CandidateEngine, false); e == nil {
		t.Fatal("incident deletion accepted")
	}
	ops[0].Cascade = true
	ops = append(ops, MaturityOperation{Action: "remove", Entity: "stage", Key: "future"}, MaturityOperation{Action: "remove", Entity: "stage", Key: "scale", Cascade: true}, MaturityOperation{Action: "set", Entity: "spacing", Spacing: "curve"})
	if _, e = PatchMaturity(p, "slide", maturityPatch(p, ops...), curveCompositionBundle(t), wmdesign.CandidateEngine, false); e != nil {
		t.Fatal(e)
	}
}
func TestMaturityCompositionStrictInvalidAndBinding(t *testing.T) {
	t.Parallel()
	p, _ := curveFixture(t, "maturity/four-stage", "maturity")
	raw := canonical(maturityPatch(p, MaturityOperation{Action: "remove", Entity: "current"}))
	raw = bytes.Replace(raw, []byte(`"entity":"current"`), []byte(`"entity":"current","cascade":false`), 1)
	if _, e := DecodeMaturityPatch(raw, "patch"); e == nil {
		t.Fatal("authored false accepted")
	}
	local := p.Document.LocalTemplates["fixture"]
	local.Nodes[0].Arguments["axisLabel"] = map[string]any{"binding": "axis"}
	p.Document.LocalTemplates["fixture"] = local
	local.Zones["axis"] = Zone{Role: "body", Schema: map[string]any{"type": "string"}}
	p.Document.LocalTemplates["fixture"] = local
	p.Document.Slides[0].Values["axis"] = "Stage"
	p = writeCurveFixture(t, p)
	v := 1.
	patch := maturityPatch(p, MaturityOperation{Action: "set", Entity: "layout", Layout: &MaturityLayout{Shape: &v}})
	if _, e := PatchMaturity(p, "slide", patch, curveCompositionBundle(t), wmdesign.CandidateEngine, false); e == nil {
		t.Fatal("binding implicitly adopted")
	}
	patch.Operations = append([]MaturityOperation{{Action: "materialize", Entity: "source"}}, patch.Operations...)
	if _, e := PatchMaturity(p, "slide", patch, curveCompositionBundle(t), wmdesign.CandidateEngine, false); e != nil {
		t.Fatal(e)
	}
	v = -1
	if _, e := PatchMaturity(p, "slide", patch, curveCompositionBundle(t), wmdesign.CandidateEngine, false); e == nil {
		t.Fatal("negative curve accepted")
	}
}
func TestStaffingCompositionCatalogPointsSeriesPhasesScale(t *testing.T) {
	t.Parallel()
	p, id := curveFixture(t, "team-curve/build-together", "teamcurve")
	ins, e := InspectStaffing(p, "slide", id, curveCompositionBundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if ins.MutationBlocked == "" {
		t.Fatal("legacy units guessed")
	}
	pk, sk := staffingPointKeys(ins.Model), staffingSeriesKeys(ins.Model)
	unit := StaffingScale{Unit: "FTE", TimeUnit: "delivery fraction", Source: "Illustrative staffing assumption"}
	last := ins.Model.Points[len(ins.Model.Points)-1]
	point := StaffingPoint{Key: "review", At: .95, Label: "Review", Values: map[string]float64{}}
	for _, k := range sk {
		point.Values[k] = last.Values[k]
	}
	order := append(append([]string{}, pk[:len(pk)-1]...), "review", pk[len(pk)-1])
	phase := StaffingPhase{Key: "transition", Label: "Transition", Point: "review"}
	patch := staffingPatch(p, StaffingOperation{Action: "set", Entity: "scale", Scale: &unit}, StaffingOperation{Action: "set", Entity: "point", Key: "review", Point: &point}, StaffingOperation{Action: "reorder", Entity: "point", Order: order, Spacing: "even"}, StaffingOperation{Action: "set", Entity: "phase", Key: "transition", Phase: &phase})
	before := append([]byte(nil), p.Raw...)
	if _, e = PatchStaffing(p, "slide", patch, curveCompositionBundle(t), wmdesign.CandidateEngine, false); e != nil {
		t.Fatal(e)
	}
	disk, _ := os.ReadFile(p.SourcePath)
	if !bytes.Equal(before, disk) {
		t.Fatal("staffing preview changed bytes")
	}
	if _, e = PatchStaffing(p, "slide", patch, curveCompositionBundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	ins, e = InspectStaffing(p, "slide", id, curveCompositionBundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if len(ins.Model.Points) != len(pk)+1 || ins.Model.Scale.Unit != "FTE" {
		t.Fatal(ins.Model)
	}
	if _, e = PatchStaffing(p, "slide", staffingPatch(p, StaffingOperation{Action: "remove", Entity: "point", Key: "review"}), curveCompositionBundle(t), wmdesign.CandidateEngine, false); e == nil {
		t.Fatal("referenced point deletion accepted")
	}
	patch = staffingPatch(p, StaffingOperation{Action: "remove", Entity: "point", Key: "review", Cascade: true}, StaffingOperation{Action: "remove", Entity: "series", Key: sk[0]})
	if _, e = PatchStaffing(p, "slide", patch, curveCompositionBundle(t), wmdesign.CandidateEngine, false); e != nil {
		t.Fatal(e)
	}
}
func TestStaffingCompositionStrictMissingScaleAndInvalidValues(t *testing.T) {
	t.Parallel()
	p, _ := curveFixture(t, "team-curve/build-together", "teamcurve")
	hide := true
	if _, e := PatchStaffing(p, "slide", staffingPatch(p, StaffingOperation{Action: "set", Entity: "layout", Layout: &StaffingLayout{PhaseLabels: &hide}}), curveCompositionBundle(t), wmdesign.CandidateEngine, false); e == nil {
		t.Fatal("legacy unspecified units allowed")
	}
	raw := canonical(staffingPatch(p, StaffingOperation{Action: "remove", Entity: "series", Key: "source-x"}))
	raw = bytes.Replace(raw, []byte(`"entity":"series"`), []byte(`"entity":"series","cascade":false`), 1)
	if _, e := DecodeStaffingPatch(raw, "patch"); e == nil {
		t.Fatal("authored inapplicable false accepted")
	}
	ins, e := InspectStaffing(p, "slide", "model", curveCompositionBundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	m := ins.Model
	m.Scale = StaffingScale{Unit: "FTE", TimeUnit: "week", Source: "Assumption"}
	m.Points[0].Values[m.Series[0].Key] = -1
	if e = validateStaffing(m, true); e == nil {
		t.Fatal("negative accepted")
	}
	m.Points[0].Values[m.Series[0].Key] = 0
	m.Points[1].At = m.Points[0].At
	if e = validateStaffing(m, true); e == nil {
		t.Fatal("duplicate positions accepted")
	}
}
func layerFixture(t *testing.T) (*Project, []LayerSelection) {
	p := example(t)
	options := wmdesign.FrameRequest{Rail: "none", Footer: "compact", TitleLines: 1, Density: "appendix", Surface: "light"}
	nodes := []Node{}
	selections := []LayerSelection{}
	for i, label := range []string{"Experience", "Services", "Data"} {
		id := []string{"experience", "services", "data"}[i]
		r := wmdesign.Rect{X: 0, Y: 0 + float64(i)*66, W: 600, H: 66}
		nodes = append(nodes, Node{ID: id, Kind: "component", Definition: &Reference{Scope: "shared", ID: "wmds/component/layerrow"}, Placement: &Placement{Zone: "body", Rect: &r}, Arguments: map[string]any{"label": label, "text": "Illustrative layer", "n": string(rune('1' + i)), "surface": "subtle"}})
		selections = append(selections, LayerSelection{id, []string{id}, id, id, id})
	}
	foundation := wmdesign.Rect{X: 0, Y: 220, W: 600, H: 36}
	controls := wmdesign.Rect{X: 620, Y: 0, W: 226, H: 198}
	nodes = append(nodes, Node{ID: "foundation", Kind: "component", Definition: &Reference{Scope: "shared", ID: "wmds/component/layerrow"}, Placement: &Placement{Zone: "body", Rect: &foundation}, Arguments: map[string]any{"label": "Foundation", "text": "Identity and network", "surface": "deep"}}, Node{ID: "controls", Kind: "component", Keys: map[string][]string{"body": {"copy"}}, Definition: &Reference{Scope: "shared", ID: "wmds/component/card"}, Placement: &Placement{Zone: "body", Rect: &controls}, Arguments: map[string]any{"label": "Across layers 1–3", "title": "Controls", "titleStyle": "subhead", "surface": "subtle", "pad": 12, "body": []any{map[string]any{"p": "Security and audit"}}}})
	local := LocalTemplate{Name: "Layer qualification", Frame: Reference{Scope: "shared", ID: "wmds/frame/none-compact"}, FrameOptions: &options, Grid: Reference{Scope: "shared", ID: "wmds/grid/12-columns"}, Zones: map[string]Zone{"title": {Role: "slide-title", Required: true, Schema: map[string]any{"type": "string"}}}, Nodes: nodes}
	p.Document.LocalTemplates = map[string]LocalTemplate{"fixture": local}
	p.Document.Slides = []Slide{{ID: "slide", ContentKind: "synthetic_example", Template: Reference{Scope: "local", ID: "fixture"}, Values: map[string]any{"title": "Illustrative architecture changes layers"}}}
	return writeCurveFixture(t, p), selections
}
func TestLayerCompositionVariableCountsControlsAndFoundation(t *testing.T) {
	t.Parallel()
	p, sel := layerFixture(t)
	label := "Integration"
	text := "Shared APIs"
	layout := LayerLayout{Rect: wmdesign.Rect{X: 0, Y: 0, W: 600, H: 270}, Palette: "sequence", FoundationNode: "foundation", FoundationGap: 18, ControlsNode: "controls", ControlsMode: "span"}
	ops := []LayerOperation{}
	for _, k := range []string{"integration", "analytics", "workflow"} {
		v := label
		if k != "integration" {
			v = k
		}
		ops = append(ops, LayerOperation{Action: "add", Entity: "layer", Key: k, Prototype: "services", Label: &v, Text: &text})
	}
	ops = append(ops, LayerOperation{Action: "reorder", Entity: "layer", Order: []string{"experience", "workflow", "services", "integration", "analytics", "data"}}, LayerOperation{Action: "set", Entity: "layout", Layout: &layout})
	patch := LayerPatch{Schema: LayerPatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "qualification", Reason: "Customize layer counts", Layers: sel, Operations: ops}
	before := append([]byte(nil), p.Raw...)
	preview, e := PatchLayers(p, "slide", patch, curveCompositionBundle(t), wmdesign.CandidateEngine, false)
	if e != nil {
		t.Fatal(e)
	}
	if preview.Applied {
		t.Fatal("preview applied")
	}
	disk, _ := os.ReadFile(p.SourcePath)
	if !bytes.Equal(before, disk) {
		t.Fatal("preview writes")
	}
	if _, e = PatchLayers(p, "slide", patch, curveCompositionBundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	local := p.Document.LocalTemplates["fixture"]
	f, _ := layerLeaf(&local, "foundation")
	c, _ := layerLeaf(&local, "controls")
	if f.Placement.Rect.Y != 288 || c.Placement.Rect.H != 270 || c.Arguments["label"] != "Across layers 1–6" {
		t.Fatal(f, c)
	}
	if len(DiscoverLayerSelections(local)) != 6 {
		t.Fatal(local.Nodes)
	}
	if _, e = PatchLayers(p, "slide", patch, curveCompositionBundle(t), wmdesign.CandidateEngine, false); e == nil {
		t.Fatal("stale accepted")
	}
}
func TestLayerCompositionExplicitSelectionsInvalidAndOverflowAtomic(t *testing.T) {
	t.Parallel()
	p, sel := layerFixture(t)
	layout := LayerLayout{Rect: wmdesign.Rect{X: 0, Y: 0, W: 600, H: 198}, Palette: "preserve"}
	label := strings.Repeat("long label ", 1000)
	patch := LayerPatch{Schema: LayerPatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "qualification", Reason: "Validate fit", Layers: sel, Operations: []LayerOperation{{Action: "set", Entity: "layer", Key: sel[0].Key, Label: &label}, {Action: "set", Entity: "layout", Layout: &layout}}}
	before := append([]byte(nil), p.Raw...)
	if _, e := PatchLayers(p, "slide", patch, curveCompositionBundle(t), wmdesign.CandidateEngine, true); e == nil {
		t.Fatal("overlong label accepted")
	}
	disk, _ := os.ReadFile(p.SourcePath)
	if !bytes.Equal(before, disk) {
		t.Fatal("failed fit changed bytes")
	}
	patch.Layers[1].Members = []string{sel[0].Members[0]}
	if _, e := PatchLayers(p, "slide", patch, curveCompositionBundle(t), wmdesign.CandidateEngine, false); e == nil {
		t.Fatal("ambiguous membership accepted")
	}
}

func TestMaturityCompositionAllCurveProfileSourceContracts(t *testing.T) {
	t.Parallel()
	// All fifteen actual maturity curves are exercised. Four AI companion-only
	// variants and readiness/adoption-curve need their actual assessment/chart models.
	for _, key := range []string{"maturity/four-stage", "maturity/five-active", "maturity/split", "maturity/nav", "maturity/table", "maturity/ai-simple", "maturity/ai-beyond", "maturity/ai-capabilities", "maturity/three-stage", "maturity/six-stage", "maturity/table-left", "maturity/table-tall", "maturity/insights", "maturity/ai-insights", "maturity/ai-here-insights"} {
		t.Run(key, func(t *testing.T) {
			p, id := curveFixture(t, key, "maturity")
			ins, e := InspectMaturity(p, "slide", id, curveCompositionBundle(t), wmdesign.CandidateEngine)
			if e != nil {
				t.Fatal(e)
			}
			if ins.RenderError != "" {
				t.Fatal(ins.RenderError)
			}
			shape := 2.0
			ops := []MaturityOperation{{Action: "set", Entity: "layout", Layout: &MaturityLayout{Shape: &shape}}}
			if _, e = PatchMaturity(p, "slide", maturityPatch(p, ops...), curveCompositionBundle(t), wmdesign.CandidateEngine, false); e != nil {
				t.Fatal(e)
			}
		})
	}
}
func TestStaffingCompositionAllNineCatalogSourceContracts(t *testing.T) {
	t.Parallel()
	for _, key := range []string{"team-curve/build-together", "team-curve/build-together-bands", "team-curve/build-together-roles", "team-curve/agents", "team-curve/agents-bands", "team-curve/agents-roles", "team-curve/agents-nav", "team-curve/agents-split", "team-curve/before-after"} {
		t.Run(key, func(t *testing.T) {
			p, id := curveFixture(t, key, "teamcurve")
			ins, e := InspectStaffing(p, "slide", id, curveCompositionBundle(t), wmdesign.CandidateEngine)
			if e != nil {
				t.Fatal(e)
			}
			if ins.RenderError != "" {
				t.Fatal(ins.RenderError)
			}
			scale := StaffingScale{Unit: "relative capacity", TimeUnit: "normalized delivery fraction", Source: "Illustrative catalog derivative", Maximum: ins.Model.Scale.Maximum}
			smooth := 0.
			curve := "monotone"
			patch := staffingPatch(p, StaffingOperation{Action: "set", Entity: "scale", Scale: &scale}, StaffingOperation{Action: "set", Entity: "layout", Layout: &StaffingLayout{Smooth: &smooth, Curve: &curve}})
			if _, e = PatchStaffing(p, "slide", patch, curveCompositionBundle(t), wmdesign.CandidateEngine, false); e != nil {
				t.Fatal(e)
			}
		})
	}
}
func TestLayerCompositionPlanePainterOrderAndCompoundRows(t *testing.T) {
	t.Parallel()
	p, sel := layerFixture(t)
	local := p.Document.LocalTemplates["fixture"]
	for i := 0; i < 3; i++ {
		local.Nodes[i].Definition.ID = "wmds/component/plane"
		delete(local.Nodes[i].Arguments, "n")
		sel[i].NumberNode = ""
		local.Nodes[i].Placement.Rect.H = 80
		local.Nodes[i].Placement.Rect.Y = float64(i) * 62
	}
	p.Document.LocalTemplates["fixture"] = local
	p = writeCurveFixture(t, p)
	label := "Integration"
	layout := LayerLayout{Rect: wmdesign.Rect{X: 0, Y: 0, W: 600, H: 270}, Overlap: 18, Palette: "sequence"}
	patch := LayerPatch{Schema: LayerPatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "qualification", Reason: "Adapt native 3-D planes", Layers: sel, Operations: []LayerOperation{{Action: "add", Entity: "layer", Key: "integration", Prototype: "services", Label: &label}, {Action: "set", Entity: "layout", Layout: &layout}}}
	if _, e := PatchLayers(p, "slide", patch, curveCompositionBundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	local = p.Document.LocalTemplates["fixture"]
	order := []string{}
	for _, n := range local.Nodes {
		if layerKind(&n) == "plane" {
			order = append(order, n.ID)
		}
	}
	if strings.Join(order, ",") != "integration,data,services,experience" {
		t.Fatal(order)
	}
	// Compound legacy layer-map rows keep their separate native label/description
	// objects and optional numbering with explicit semantic membership.
	p, _ = layerFixture(t)
	local = p.Document.LocalTemplates["fixture"]
	local.Nodes = local.Nodes[3:]
	sel = nil
	for i, id := range []string{"experience", "services", "data"} {
		y := float64(i) * 66
		labelRect := wmdesign.Rect{X: 45, Y: y, W: 180, H: 26}
		textRect := wmdesign.Rect{X: 240, Y: y, W: 300, H: 26}
		numberRect := wmdesign.Rect{X: 0, Y: y, W: 36, H: 36}
		local.Nodes = append(local.Nodes, Node{ID: id + "-label", Kind: "text", Style: "label", Text: id, Placement: &Placement{Zone: "body", Rect: &labelRect}}, Node{ID: id + "-text", Kind: "text", Style: "small", Text: "Illustrative mapping", Placement: &Placement{Zone: "body", Rect: &textRect}}, Node{ID: id + "-number", Kind: "component", Definition: &Reference{Scope: "shared", ID: "wmds/component/block"}, Placement: &Placement{Zone: "body", Rect: &numberRect}, Arguments: map[string]any{"text": string(rune('1' + i)), "surface": "inverse", "style": "label"}})
		sel = append(sel, LayerSelection{Key: id, Members: []string{id + "-number", id + "-label", id + "-text"}, LabelNode: id + "-label", TextNode: id + "-text", NumberNode: id + "-number"})
	}
	p.Document.LocalTemplates["fixture"] = local
	p = writeCurveFixture(t, p)
	layout = LayerLayout{Rect: wmdesign.Rect{X: 0, Y: 0, W: 540, H: 198}, Palette: "preserve"}
	patch = LayerPatch{Schema: LayerPatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "qualification", Reason: "Adapt compound legacy layer list", Layers: sel, Operations: []LayerOperation{{Action: "add", Entity: "layer", Key: "integration", Prototype: "services", Label: &label}, {Action: "set", Entity: "layout", Layout: &layout}}}
	if _, e := PatchLayers(p, "slide", patch, curveCompositionBundle(t), wmdesign.CandidateEngine, false); e != nil {
		t.Fatal(e)
	}
}
func TestStaffingCompositionRemoveAllPhasesAndNonuniformSpacing(t *testing.T) {
	t.Parallel()
	p, id := curveFixture(t, "team-curve/agents", "teamcurve")
	ins, e := InspectStaffing(p, "slide", id, curveCompositionBundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	ops := []StaffingOperation{{Action: "set", Entity: "scale", Scale: &StaffingScale{Unit: "relative capacity", TimeUnit: "normalized duration", Source: "Illustrative assumption"}}}
	for _, ph := range ins.Model.Phases {
		ops = append(ops, StaffingOperation{Action: "remove", Entity: "phase", Key: ph.Key})
	}
	order := staffingPointKeys(ins.Model)
	at := []float64{}
	for i := range order {
		u := float64(i) / float64(len(order)-1)
		at = append(at, u*u)
	}
	ops = append(ops, StaffingOperation{Action: "reorder", Entity: "point", Order: order, Spacing: "explicit", At: at})
	if _, e = PatchStaffing(p, "slide", staffingPatch(p, ops...), curveCompositionBundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	ins, e = InspectStaffing(p, "slide", id, curveCompositionBundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if len(ins.Model.Phases) != 0 || ins.Model.Points[1].At != at[1] {
		t.Fatal(ins.Model)
	}
}
func layerCatalogFixture(t *testing.T, key string) (*Project, []LayerSelection) {
	t.Helper()
	p, _ := layerFixture(t)
	catalog, e := wmdesign.LibraryCatalog(curveCompositionBundle(t), "")
	if e != nil {
		t.Fatal(e)
	}
	var body []map[string]any
	for _, d := range catalog {
		if d.Key != key {
			continue
		}
		var s struct {
			Body []map[string]any `json:"body"`
		}
		if e = json.Unmarshal(d.RawSlide, &s); e != nil {
			t.Fatal(e)
		}
		body = s.Body
	}
	if len(body) == 0 {
		t.Fatal("missing catalog layer profile", key)
	}
	selected := []int{}
	compound := key == "architecture/layer-map"
	for i, n := range body {
		typ, _ := n["type"].(string)
		if compound {
			if i < 15 {
				selected = append(selected, i)
			}
			continue
		}
		if typ == "layerrow" && n["n"] != nil || key == "architecture/layers-icons" && typ == "node" && n["y"].(float64) < 390 || strings.Contains(key, "layers-3d") && typ == "plane" && n["x"].(float64) < 600 {
			selected = append(selected, i)
		}
	}
	if len(selected) == 0 {
		t.Fatal("missing selected source layers", key)
	}
	minX, minY := 1e9, 1e9
	for _, i := range selected {
		n := body[i]
		if x := n["x"].(float64); x < minX {
			minX = x
		}
		if y := n["y"].(float64); y < minY {
			minY = y
		}
	}
	nodes := []Node{}
	sel := []LayerSelection{}
	for j, i := range selected {
		n := body[i]
		typ := n["type"].(string)
		id := fmt.Sprintf("source-%03d", j+1)
		h, ok := n["h"].(float64)
		if !ok {
			h = 24
		}
		r := wmdesign.Rect{X: n["x"].(float64) - minX, Y: n["y"].(float64) - minY, W: n["w"].(float64), H: h}
		for _, k := range []string{"type", "id", "x", "y", "w", "h"} {
			delete(n, k)
		}
		nodes = append(nodes, Node{ID: id, Kind: "component", Definition: &Reference{Scope: "shared", ID: "wmds/component/" + typ}, Placement: &Placement{Zone: "body", Rect: &r}, Arguments: n})
		if !compound {
			num := ""
			if typ == "layerrow" {
				num = id
			}
			sel = append(sel, LayerSelection{Key: id, Members: []string{id}, LabelNode: id, TextNode: id, NumberNode: num})
		}
	}
	if compound {
		for j := 0; j < len(nodes); j += 3 {
			sel = append(sel, LayerSelection{Key: fmt.Sprintf("layer-%d", j/3+1), Members: []string{nodes[j].ID, nodes[j+1].ID, nodes[j+2].ID}, LabelNode: nodes[j+1].ID, TextNode: nodes[j+2].ID, NumberNode: nodes[j].ID})
		}
	}
	// Explicitly define the visual order of original 3-D source planes. Its body
	// paints bottom-up; semantic order is top-down.
	if strings.Contains(key, "layers-3d") {
		for i, j := 0, len(sel)-1; i < j; i, j = i+1, j-1 {
			sel[i], sel[j] = sel[j], sel[i]
		}
	}
	local := p.Document.LocalTemplates["fixture"]
	local.Nodes = nodes
	p.Document.LocalTemplates["fixture"] = local
	return writeCurveFixture(t, p), sel
}
func TestLayerCompositionAllEightCatalogProfiles(t *testing.T) {
	t.Parallel()
	for _, key := range []string{"architecture/layers", "architecture/layers-left", "architecture/layers-split", "architecture/layers-nav", "architecture/layers-icons", "architecture/layers-3d", "architecture/layers-3d-systems", "architecture/layer-map"} {
		t.Run(key, func(t *testing.T) {
			p, sel := layerCatalogFixture(t, key)
			ins, e := InspectLayers(p, "slide", sel, curveCompositionBundle(t), wmdesign.CandidateEngine)
			if e != nil {
				t.Fatal(e)
			}
			if ins.RenderError != "" {
				t.Fatal(ins.RenderError)
			}
			layout := ins.Layout
			layout.Palette = "preserve"
			if strings.Contains(key, "layers-3d") {
				layout.Overlap = 18
			}
			patch := LayerPatch{Schema: LayerPatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "qualification", Reason: "Verify actual catalog native layer source", Layers: sel, Operations: []LayerOperation{{Action: "set", Entity: "layout", Layout: &layout}}}
			if _, e = PatchLayers(p, "slide", patch, curveCompositionBundle(t), wmdesign.CandidateEngine, false); e != nil {
				t.Fatal(e)
			}
		})
	}
}

func TestStaffingCompositionCommentsFollowStableSeriesAfterPointEdit(t *testing.T) {
	t.Parallel()
	p, _ := curveFixture(t, "team-curve/build-together", "teamcurve")
	doc, e := sourceYAML(p.Raw)
	if e != nil {
		t.Fatal(e)
	}
	source := mappingNode(mappingNode(mappingNode(doc.Content[0], "local_templates"), "fixture"), "nodes").Content[0]
	series := mappingNode(mappingNode(source, "arguments"), "series")
	series.Content[0].HeadComment = "Retain this series ownership rationale"
	mappingNode(series.Content[0], "name").LineComment = "Keep authored capacity meaning"
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
	in, e := InspectStaffing(p, "slide", "model", curveCompositionBundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	point := in.Model.Points[1]
	for key := range point.Values {
		point.Values[key] += .1
	}
	patch := staffingPatch(p, StaffingOperation{Action: "set", Entity: "point", Key: point.Key, Point: &point}, StaffingOperation{Action: "set", Entity: "scale", Scale: &StaffingScale{Unit: "relative capacity", TimeUnit: "illustrative phase", Source: "Synthetic catalog example"}})
	if _, e = PatchStaffing(p, "slide", patch, curveCompositionBundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	doc, e = sourceYAML(p.Raw)
	if e != nil {
		t.Fatal(e)
	}
	source = mappingNode(mappingNode(mappingNode(doc.Content[0], "local_templates"), "fixture"), "nodes").Content[0]
	first := mappingNode(mappingNode(source, "arguments"), "series").Content[0]
	if !strings.Contains(first.HeadComment, "Retain this series ownership rationale") || !strings.Contains(mappingNode(first, "name").LineComment, "Keep authored capacity meaning") {
		t.Fatal("stable series comments lost after point edit", first.HeadComment)
	}
}

func TestMaturityCompositionNativeAndPinGuards(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"maturity", "staffing", "layers"} {
		for _, refusal := range []string{"native-order", "native-geometry", "pinned-template", "shared-template"} {
			t.Run(kind+"/"+refusal, func(t *testing.T) {
				var p *Project
				var selections []LayerSelection
				switch kind {
				case "maturity":
					p, _ = curveFixture(t, "maturity/four-stage", "maturity")
				case "staffing":
					p, _ = curveFixture(t, "team-curve/build-together", "teamcurve")
				case "layers":
					p, selections = layerFixture(t)
				}
				switch refusal {
				case "native-order":
					p.Document.Slides[0].NativeOrder = map[string][]string{"root": {"model"}}
				case "native-geometry":
					p.Document.Slides[0].NativeGeometry = map[string]NativeGeometry{"model": {Kind: "sp"}}
				case "pinned-template":
					p.Document.Slides[0].Template.Revision = "r1"
				case "shared-template":
					second := p.Document.Slides[0]
					second.ID = "other"
					p.Document.Slides = append(p.Document.Slides, second)
				}
				before, _ := os.ReadFile(p.SourcePath)
				var e error
				switch kind {
				case "maturity":
					shape := 1.
					_, e = PatchMaturity(p, "slide", maturityPatch(p, MaturityOperation{Action: "set", Entity: "layout", Layout: &MaturityLayout{Shape: &shape}}), curveCompositionBundle(t), wmdesign.CandidateEngine, true)
				case "staffing":
					_, e = PatchStaffing(p, "slide", staffingPatch(p, StaffingOperation{Action: "set", Entity: "scale", Scale: &StaffingScale{Unit: "relative capacity", TimeUnit: "illustrative phase", Source: "Synthetic catalog example"}}), curveCompositionBundle(t), wmdesign.CandidateEngine, true)
				case "layers":
					_, e = PatchLayers(p, "slide", LayerPatch{Schema: LayerPatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "qualification", Reason: "Respect native overrides and template pins", Layers: selections, Operations: []LayerOperation{{Action: "set", Entity: "layout", Layout: &LayerLayout{Rect: wmdesign.Rect{X: 0, Y: 0, W: 600, H: 198}, Palette: "preserve"}}}}, curveCompositionBundle(t), wmdesign.CandidateEngine, true)
				}
				if e == nil {
					t.Fatal("source safety refusal bypassed", refusal)
				}
				after, _ := os.ReadFile(p.SourcePath)
				if !bytes.Equal(before, after) {
					t.Fatal("refusal changed source")
				}
			})
		}
	}
}
