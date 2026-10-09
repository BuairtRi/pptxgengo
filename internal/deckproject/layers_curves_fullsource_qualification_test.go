package deckproject

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

// Opt-in desktop input exporter: full actual source parents and companions are
// retained. Its test-process pin is subsequently migrated by the genuine CLI.
func TestLayersCurvesFullSourceCombinedGUIExamples(t *testing.T) {
	if os.Getenv("PPTXGENGO_FAMILY_QUALIFICATION_DIR") == "" {
		t.Skip("explicit private output required")
	}
	b := curveCompositionBundle(t)
	combined := example(t)
	combined.Document.ID = "v430-layers-curves-full-source-11"
	combined.Document.Title = "Illustrative layer, maturity and staffing customization"
	combined.Document.Context = nil
	combined.Document.Assets = map[string]Asset{}
	combined.Document.LocalTemplates = map[string]LocalTemplate{}
	combined.Document.Slides = nil
	combined.Document.EditingProfile = wmdesign.NativeEditingProfile
	receipts := []map[string]any{}
	fresh := func(key string) *Project {
		scaffold, e := ScaffoldTemplateWithOptions(b, key, wmdesign.CandidateEngine, "Illustrative full-source desktop customization; preserve actual parent, companions and stable native ownership", 2026, ScaffoldOptions{PlaceholderMedia: true, SourceContainerClearanceFit: key == "architecture/layer-map"})
		if e != nil {
			t.Fatal(key, e)
		}
		p := example(t)
		p.Document.Context = nil
		p.Document.Assets = scaffold.Assets
		p.Document.EditingProfile = wmdesign.NativeEditingProfile
		if key == "architecture/layer-map" {
			if len(scaffold.GeometryAdjustments) != 1 || scaffold.GeometryAdjustments[0].Original.H != 294 || scaffold.GeometryAdjustments[0].Authored.H != 291 {
				t.Fatal("explicit original294→291 source frame adjustment missing", scaffold.GeometryAdjustments)
			}
			receipts = append(receipts, map[string]any{"template": key, "explicit_source_container_clearance_fit": scaffold.GeometryAdjustments})
		}
		for path, payload := range scaffold.AssetPayloads {
			target := filepath.Join(p.Root, path)
			if e = os.MkdirAll(filepath.Dir(target), 0700); e != nil {
				t.Fatal(e)
			}
			if e = os.WriteFile(target, payload, 0600); e != nil {
				t.Fatal(e)
			}
		}
		p.Document.LocalTemplates = map[string]LocalTemplate{"catalog": scaffold.Template}
		p.Document.Slides = []Slide{{ID: "catalog-slide", ContentKind: "synthetic_example", Template: Reference{Scope: "local", ID: "catalog"}, Values: scaffold.SyntheticSourceValues}}
		return writeCurveFixture(t, p)
	}
	add := func(p *Project, name, key string) {
		p, e := Load(p.SourcePath)
		if e != nil {
			t.Fatal(e)
		}
		local := p.Document.LocalTemplates["catalog"]
		if local.Provenance == nil || local.Provenance.Parent.ID != key || local.Provenance.SourceFileSHA256 == "" {
			t.Fatal("actual source ancestry missing", name)
		}
		slide := p.Document.Slides[0]
		slide.ID = name
		slide.Template.ID = name
		// Identification copy is explicitly illustrative; all source companions stay.
		for zone, contract := range local.Zones {
			if contract.Role == "slide-title" {
				slide.Values[zone] = fmt.Sprintf("Illustrative customization: %s", name)
			}
		}
		combined.Document.LocalTemplates[name] = local
		combined.Document.Slides = append(combined.Document.Slides, slide)
		for id, asset := range p.Document.Assets {
			if prior, ok := combined.Document.Assets[id]; ok && prior != asset {
				t.Fatal("asset alias collision", id)
			}
			combined.Document.Assets[id] = asset
			if asset.Path != "" {
				payload, e := os.ReadFile(filepath.Join(p.Root, asset.Path))
				if e != nil {
					t.Fatal(e)
				}
				target := filepath.Join(combined.Root, asset.Path)
				if e = os.MkdirAll(filepath.Dir(target), 0700); e != nil {
					t.Fatal(e)
				}
				if e = os.WriteFile(target, payload, 0600); e != nil {
					t.Fatal(e)
				}
			}
		}
		receipts = append(receipts, map[string]any{"slide": name, "template": key, "parent": local.Provenance, "retained_nodes": len(local.Nodes), "source_sha256": p.SourceHash(), "native": "pending actual desktop edit/Save As/adoption and visual review"})
	}
	find := func(p *Project, kind string) []string {
		ids := []string{}
		for _, n := range p.Document.LocalTemplates["catalog"].Nodes {
			if n.Definition != nil && n.Definition.ID == "wmds/component/"+kind {
				ids = append(ids, n.ID)
			}
		}
		if len(ids) == 0 {
			t.Fatal("missing actual source component", kind)
		}
		return ids
	}
	for _, count := range []int{3, 5, 6} {
		p := fresh("maturity/six-stage")
		id := find(p, "maturity")[0]
		inspect, e := InspectMaturity(p, "catalog-slide", id, b, wmdesign.CandidateEngine)
		if e != nil {
			t.Fatal(e)
		}
		ops := []MaturityOperation{{Action: "materialize", Entity: "source"}}
		for _, stage := range inspect.Model.Stages[count:] {
			ops = append(ops, MaturityOperation{Action: "remove", Entity: "stage", Key: stage.Key, Cascade: true})
		}
		shape := float64(count) / 2
		ops = append(ops, MaturityOperation{Action: "set", Entity: "spacing", Spacing: "even"}, MaturityOperation{Action: "set", Entity: "current", Key: inspect.Model.Stages[0].Key}, MaturityOperation{Action: "set", Entity: "target", Marker: &MaturityMarker{inspect.Model.Stages[count-1].Key, "Illustrative target"}}, MaturityOperation{Action: "set", Entity: "layout", Layout: &MaturityLayout{Shape: &shape}})
		patch := maturityPatch(p, ops...)
		patch.NodeID = id
		if _, e = PatchMaturity(p, "catalog-slide", patch, b, wmdesign.CandidateEngine, true); e != nil {
			t.Fatal("maturity", count, e)
		}
		add(p, fmt.Sprintf("maturity-%d-stages", count), "maturity/six-stage")
	}
	for _, count := range []int{3, 5, 6} {
		p := fresh("architecture/layers")
		inspect, e := InspectLayers(p, "catalog-slide", nil, b, wmdesign.CandidateEngine)
		if e != nil {
			t.Fatal(e)
		}
		selections := DiscoverLayerSelections(p.Document.LocalTemplates["catalog"], p.Document.Slides[0].Values)
		ops := []LayerOperation{{Action: "materialize", Entity: "source"}}
		for _, layer := range inspect.Layers[countMin(count, len(inspect.Layers)):] {
			ops = append(ops, LayerOperation{Action: "remove", Entity: "layer", Key: layer.Key, Cascade: true})
		}
		for i := len(inspect.Layers); i < count; i++ {
			label := fmt.Sprintf("Illustrative layer %d", i+1)
			text := "Explicit responsibility"
			ops = append(ops, LayerOperation{Action: "add", Entity: "layer", Key: fmt.Sprintf("added-layer-%d", i+1), Prototype: inspect.Layers[1].Key, Label: &label, Text: &text})
		}
		layout := inspect.Layout
		layout.Palette = "sequence"
		for _, n := range p.Document.LocalTemplates["catalog"].Nodes {
			if n.Definition == nil {
				continue
			}
			if n.Definition.ID == "wmds/component/layerrow" && n.Arguments["n"] == nil {
				layout.FoundationNode = n.ID
				layout.FoundationGap = 18
			}
			if n.Definition.ID == "wmds/component/card" {
				layout.ControlsNode = n.ID
				layout.ControlsMode = "span"
			}
		}
		ops = append(ops, LayerOperation{Action: "set", Entity: "layout", Layout: &layout})
		patch := LayerPatch{Schema: LayerPatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "Desktop qualification", Reason: "Explicit illustrative count and ordinal palette; retain actual source framing, controls and foundation", Layers: selections, Operations: ops}
		if _, e = PatchLayers(p, "catalog-slide", patch, b, wmdesign.CandidateEngine, true); e != nil {
			t.Fatal("layers", count, e)
		}
		add(p, fmt.Sprintf("layers-%d", count), "architecture/layers")
	}
	for _, key := range []string{"team-curve/agents", "team-curve/before-after"} {
		p := fresh(key)
		for _, id := range find(p, "teamcurve") {
			inspect, e := InspectStaffing(p, "catalog-slide", id, b, wmdesign.CandidateEngine)
			if e != nil {
				t.Fatal(e)
			}
			keys := staffingPointKeys(inspect.Model)
			at := []float64{}
			for i := range keys {
				u := float64(i) / float64(len(keys)-1)
				at = append(at, u*u)
			}
			ops := []StaffingOperation{{Action: "materialize", Entity: "source"}, {Action: "set", Entity: "scale", Scale: &StaffingScale{Unit: "relative capacity", TimeUnit: "normalized delivery", Source: "Illustrative assumption"}}, {Action: "reorder", Entity: "point", Order: keys, Spacing: "explicit", At: at}}
			patch := staffingPatch(p, ops...)
			patch.NodeID = id
			if _, e = PatchStaffing(p, "catalog-slide", patch, b, wmdesign.CandidateEngine, true); e != nil {
				t.Fatal("staffing", key, id, e)
			}
			p, e = Load(p.SourcePath)
			if e != nil {
				t.Fatal(e)
			}
		}
		if key == "team-curve/agents" {
			// The new nonlinear staffing samples change the colored band at
			// the retained fixed annotations. Explicitly move those source
			// companions into its dark area; do not delete or infer their copy.
			local := p.Document.LocalTemplates["catalog"]
			ops := []DiagramOperation{}
			for _, node := range local.Nodes {
				if node.ID != "node02" && node.ID != "node05" && node.ID != "node06" {
					continue
				}
				r := *node.Placement.Rect
				if node.ID == "node02" {
					r.W = 252 // Retain first-phase scope; never cross the phase separator.
				} else if node.ID == "node05" {
					r.Y = 181
				} else {
					r.Y = 78
				}
				ops = append(ops, DiagramOperation{Action: "move", ID: node.ID, Rect: &r})
			}
			if len(ops) != 3 {
				t.Fatal("expected retained narrative, AI heading and bullet source companions")
			}
			result, e := PatchDiagram(p, "catalog-slide", DiagramPatch{Schema: DiagramPatchSchema, Actor: "Desktop qualification", Reason: "Explicitly retain narrative within first-phase scope and relocate retained AI heading and bullet companions into the colored staffing band after unequal spacing; preserve their original copy and identities", Operations: ops}, b, wmdesign.CandidateEngine, true)
			if e != nil {
				t.Fatal("staffing contextual source relocation", e)
			}
			receipts = append(receipts, map[string]any{"template": key, "explicit_retained_companion_relocation": result})
			p, e = Load(p.SourcePath)
			if e != nil {
				t.Fatal(e)
			}
			caption, e := PatchComponent(p, "catalog-slide", ComponentPatch{Schema: ComponentPatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "Desktop qualification", Reason: "Use existing small note typography for retained first-phase context; preserve every original word, source identity and phase scope", SemanticReview: "Typography only; original Northfield/West Monroe diagnosis, benchmark and future-agent facts unchanged; no divider, date or phase relationship inferred", NodeID: "node02", Operations: []ComponentOperation{{Action: "materialize", Entity: "source"}, {Action: "set", Entity: "argument", Path: "/style", Value: "small"}}}, b, wmdesign.CandidateEngine, true)
			if e != nil {
				t.Fatal("staffing contextual note typography", e)
			}
			receipts = append(receipts, map[string]any{"template": key, "explicit_retained_narrative_typography": caption})
			p, e = Load(p.SourcePath)
			if e != nil {
				t.Fatal(e)
			}
		}
		name := "staffing-unequal-spacing"
		if key == "team-curve/before-after" {
			name = "staffing-before-after"
		}
		add(p, name, key)
	}
	for _, key := range []string{"architecture/layers-3d", "architecture/layer-map", "architecture/layers-icons"} {
		p := fresh(key)
		local := p.Document.LocalTemplates["catalog"]
		selections := []LayerSelection{}
		if key == "architecture/layer-map" {
			for i := 0; i < 15; i += 3 {
				selections = append(selections, LayerSelection{Key: fmt.Sprintf("layer-%d", i/3+1), Members: []string{local.Nodes[i].ID, local.Nodes[i+1].ID, local.Nodes[i+2].ID}, LabelNode: local.Nodes[i+1].ID, TextNode: local.Nodes[i+2].ID})
			}
		} else {
			for _, n := range local.Nodes {
				if n.Definition == nil {
					continue
				}
				kind := layerKind(&n)
				if kind == "plane" && n.Placement.Rect.X < 500 || kind == "node" && n.Placement.Rect.Y < 240 {
					selections = append(selections, LayerSelection{Key: n.ID, Members: []string{n.ID}, LabelNode: n.ID, TextNode: n.ID})
				}
			}
			sort.Slice(selections, func(i, j int) bool {
				a, _ := layerLeaf(&local, selections[i].LabelNode)
				b, _ := layerLeaf(&local, selections[j].LabelNode)
				return a.Placement.Rect.Y < b.Placement.Rect.Y
			})
		}
		inspect, e := InspectLayers(p, "catalog-slide", selections, b, wmdesign.CandidateEngine)
		if e != nil {
			t.Fatal(e)
		}
		layout := inspect.Layout
		layout.Palette = "preserve"
		if key == "architecture/layers-3d" {
			layout.Overlap = 18
		}
		label := inspect.Layers[0].Label + " (illustrative)"
		if key == "architecture/layers-3d" {
			label = "Experience layer"
		}
		patch := LayerPatch{Schema: LayerPatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "Desktop qualification", Reason: "Explicit full-source layer selection, preserved surfaces and companions; labels identify illustrative examples", Layers: selections, Operations: []LayerOperation{{Action: "materialize", Entity: "source"}, {Action: "set", Entity: "layer", Key: selections[0].Key, Label: &label}, {Action: "set", Entity: "layout", Layout: &layout}}}
		if _, e = PatchLayers(p, "catalog-slide", patch, b, wmdesign.CandidateEngine, true); e != nil {
			t.Fatal(key, e)
		}
		add(p, filepath.Base(key), key)
	}
	if len(combined.Document.Slides) != 11 {
		t.Fatal("expected eleven full source cases")
	}
	combined = writeCurveFixture(t, combined)
	if _, e := Build(combined, BuildOptions{Bundle: b, Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal("combined source build", e)
	}
	if e := os.WriteFile(filepath.Join(combined.Root, "qualification-inputs.json"), canonical(receipts), 0600); e != nil {
		t.Fatal(e)
	}
	writeLayerCurveQualification(t, combined, "combined-layers-curves-11")
}

func countMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}
