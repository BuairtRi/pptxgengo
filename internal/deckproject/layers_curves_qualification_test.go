package deckproject

import (
	"encoding/json"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func writeLayerCurveQualification(t *testing.T, p *Project, name string) {
	t.Helper()
	root := os.Getenv("PPTXGENGO_FAMILY_QUALIFICATION_DIR")
	if !filepath.IsAbs(root) {
		t.Fatal("qualification output must be explicit absolute private directory")
	}
	dst := filepath.Join(root, name)
	if _, e := os.Stat(dst); e == nil {
		t.Fatal("qualification output exists", dst)
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
func TestLayersCurvesQualificationExamples(t *testing.T) {
	if os.Getenv("PPTXGENGO_FAMILY_QUALIFICATION_DIR") == "" {
		t.Skip("explicit private output required")
	}
	b := curveCompositionBundle(t)
	for _, count := range []int{3, 5, 6} {
		p, id := curveFixture(t, "maturity/six-stage", "maturity")
		ins, e := InspectMaturity(p, "slide", id, b, wmdesign.CandidateEngine)
		if e != nil {
			t.Fatal(e)
		}
		ops := []MaturityOperation{}
		for _, s := range ins.Model.Stages[count:] {
			ops = append(ops, MaturityOperation{Action: "remove", Entity: "stage", Key: s.Key, Cascade: true})
		}
		shape := float64(count) / 2
		ops = append(ops, MaturityOperation{Action: "set", Entity: "spacing", Spacing: "even"}, MaturityOperation{Action: "set", Entity: "current", Key: ins.Model.Stages[0].Key}, MaturityOperation{Action: "set", Entity: "here", Marker: &MaturityMarker{ins.Model.Stages[0].Key, "Current"}}, MaturityOperation{Action: "set", Entity: "target", Marker: &MaturityMarker{ins.Model.Stages[count-1].Key, "Target"}}, MaturityOperation{Action: "set", Entity: "layout", Layout: &MaturityLayout{Shape: &shape}})
		if _, e = PatchMaturity(p, "slide", maturityPatch(p, ops...), b, wmdesign.CandidateEngine, true); e != nil {
			t.Fatal(e)
		}
		p, e = Load(p.SourcePath)
		if e != nil {
			t.Fatal(e)
		}
		writeLayerCurveQualification(t, p, fmt.Sprintf("maturity-%d-stages", count))
	}
	p, id := curveFixture(t, "team-curve/agents", "teamcurve")
	ins, e := InspectStaffing(p, "slide", id, b, wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	pk := staffingPointKeys(ins.Model)
	at := []float64{}
	for i := range pk {
		u := float64(i) / float64(len(pk)-1)
		at = append(at, u*u)
	}
	patch := staffingPatch(p, StaffingOperation{Action: "set", Entity: "scale", Scale: &StaffingScale{Unit: "relative capacity", TimeUnit: "normalized duration", Source: "Illustrative assumption"}}, StaffingOperation{Action: "reorder", Entity: "point", Order: pk, Spacing: "explicit", At: at})
	if _, e = PatchStaffing(p, "slide", patch, b, wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	writeLayerCurveQualification(t, p, "staffing-unequal-spacing")
	// Both actual before/after sources are retained as separate keyed nodes in a
	// comparable full-width fixture; semantics are authored for each independently.
	p, id = curveFixture(t, "team-curve/before-after", "teamcurve")
	catalog, e := wmdesign.LibraryCatalog(b, "")
	if e != nil {
		t.Fatal(e)
	}
	var args map[string]any
	for _, d := range catalog {
		if d.Key != "team-curve/before-after" {
			continue
		}
		var slide struct {
			Body []map[string]any `json:"body"`
		}
		if e = json.Unmarshal(d.RawSlide, &slide); e != nil {
			t.Fatal(e)
		}
		seen := 0
		for _, n := range slide.Body {
			if n["type"] != "teamcurve" {
				continue
			}
			seen++
			if seen == 2 {
				args = n
				break
			}
		}
	}
	if args == nil {
		t.Fatal("second actual before/after curve missing")
	}
	for _, k := range []string{"type", "id", "x", "y", "w", "h"} {
		delete(args, k)
	}
	local := p.Document.LocalTemplates["fixture"]
	local.Nodes[0].Placement.Rect = &wmdesign.Rect{X: 0, Y: 0, W: 405, H: 330}
	keys := map[string][]string{}
	for _, path := range []string{"series", "phases"} {
		for i := range args[path].([]any) {
			keys[path] = append(keys[path], fmt.Sprintf("source-%03d", i+1))
		}
	}
	local.Nodes = append(local.Nodes, Node{ID: "after", Kind: "component", Definition: &Reference{Scope: "shared", ID: "wmds/component/teamcurve"}, Placement: &Placement{Zone: "body", Rect: &wmdesign.Rect{X: 435, Y: 0, W: 411, H: 330}}, Arguments: args, Keys: keys})
	p.Document.LocalTemplates["fixture"] = local
	p = writeCurveFixture(t, p)
	for _, nodeID := range []string{id, "after"} {
		patch := staffingPatch(p, StaffingOperation{Action: "set", Entity: "scale", Scale: &StaffingScale{Unit: "relative capacity", TimeUnit: "delivery fraction", Source: "Illustrative comparison"}})
		patch.NodeID = nodeID
		if _, e = PatchStaffing(p, "slide", patch, b, wmdesign.CandidateEngine, true); e != nil {
			t.Fatal(e)
		}
		p, e = Load(p.SourcePath)
		if e != nil {
			t.Fatal(e)
		}
	}
	writeLayerCurveQualification(t, p, "staffing-before-after")
	for _, count := range []int{3, 5, 6} {
		p, sel := layerFixture(t)
		ops := []LayerOperation{}
		for i := 3; i < count; i++ {
			label := fmt.Sprintf("Layer %d", i+1)
			text := "Explicit responsibility"
			ops = append(ops, LayerOperation{Action: "add", Entity: "layer", Key: fmt.Sprintf("layer-%d", i+1), Prototype: "services", Label: &label, Text: &text})
		}
		layout := LayerLayout{Rect: wmdesign.Rect{X: 0, Y: 0, W: 600, H: 270}, Palette: "sequence", FoundationNode: "foundation", FoundationGap: 18, ControlsNode: "controls", ControlsMode: "span"}
		ops = append(ops, LayerOperation{Action: "set", Entity: "layout", Layout: &layout})
		patch := LayerPatch{Schema: LayerPatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "qualification", Reason: "Illustrative variable architecture", Layers: sel, Operations: ops}
		if _, e = PatchLayers(p, "slide", patch, b, wmdesign.CandidateEngine, true); e != nil {
			t.Fatal(e)
		}
		p, e = Load(p.SourcePath)
		if e != nil {
			t.Fatal(e)
		}
		writeLayerCurveQualification(t, p, fmt.Sprintf("layers-%d", count))
	}
	for _, key := range []string{"architecture/layers-3d", "architecture/layer-map", "architecture/layers-icons"} {
		p, sel := layerCatalogFixture(t, key)
		ins, e := InspectLayers(p, "slide", sel, b, wmdesign.CandidateEngine)
		if e != nil {
			t.Fatal(e)
		}
		layout := ins.Layout
		layout.Palette = "preserve"
		if key == "architecture/layers-3d" {
			layout.Overlap = 18
		}
		patch := LayerPatch{Schema: LayerPatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "qualification", Reason: "Preserve actual catalog visual types", Layers: sel, Operations: []LayerOperation{{Action: "set", Entity: "layout", Layout: &layout}}}
		if _, e = PatchLayers(p, "slide", patch, b, wmdesign.CandidateEngine, true); e != nil {
			t.Fatal(e)
		}
		p, e = Load(p.SourcePath)
		if e != nil {
			t.Fatal(e)
		}
		writeLayerCurveQualification(t, p, filepath.Base(key))
	}
}
