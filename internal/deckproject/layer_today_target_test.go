package deckproject

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

// Keep the complete genuine catalog source, rather than reconstructing only its
// four convenient target rows. Today links belong to the retained Today model;
// target row authoring explicitly preserves them, without inferring transition.
func todayTargetLayerProject(t *testing.T) (*Project, []LayerSelection) {
	t.Helper()
	b := curveCompositionBundle(t)
	s, err := ScaffoldTemplate(b, "architecture/today-vs-target", wmdesign.CandidateEngine, "Explicit target-layer composition; retain current-state topology and transition context", 2026)
	if err != nil {
		t.Fatal(err)
	}
	p := example(t)
	p.Document.Context = nil
	p.Document.EditingProfile = wmdesign.NativeEditingProfile
	p.Document.LocalTemplates = map[string]LocalTemplate{"catalog": s.Template}
	p.Document.Slides = []Slide{{ID: "slide", ContentKind: "synthetic_example", Template: Reference{Scope: "local", ID: "catalog"}, Values: s.SyntheticSourceValues}}
	p = writeCurveFixture(t, p)
	selections := []LayerSelection{}
	for i := 18; i <= 21; i++ {
		id := fmt.Sprintf("node%02d", i)
		selections = append(selections, LayerSelection{Key: id, Members: []string{id}, LabelNode: id, TextNode: id, NumberNode: id})
	}
	if len(p.Document.LocalTemplates["catalog"].Nodes) != 24 {
		t.Fatal("incomplete actual today/target source")
	}
	return p, selections
}

func TestLayerCompositionActualTodayTargetPreservesCurrentState(t *testing.T) {
	p, selections := todayTargetLayerProject(t)
	b := curveCompositionBundle(t)
	baseline := p.Document.LocalTemplates["catalog"]
	// Explicit materialization prunes selected-row binding values; all unrelated
	// narrative/current-state values must stay exact, not merely still render.
	contextValues := func(values map[string]any) []byte {
		out := map[string]any{}
		for key, value := range values {
			selected := false
			for i := 18; i <= 21; i++ {
				selected = selected || strings.HasPrefix(key, fmt.Sprintf("node%02d.", i))
			}
			if !selected {
				out[key] = value
			}
		}
		raw, _ := json.Marshal(out)
		return raw
	}
	baselineValues := contextValues(p.Document.Slides[0].Values)
	retained := map[string]Node{}
	for _, n := range baseline.Nodes {
		if n.ID < "node18" || n.ID > "node21" {
			retained[n.ID] = n
		}
	}
	verifyRetained := func(p *Project, count int) {
		t.Helper()
		local := p.Document.LocalTemplates["catalog"]
		if !reflect.DeepEqual(local.Provenance, baseline.Provenance) || !reflect.DeepEqual(local.FrameOptions, baseline.FrameOptions) || !reflect.DeepEqual(local.Frame, baseline.Frame) || !reflect.DeepEqual(local.Grid, baseline.Grid) {
			t.Fatal("parent pins/frame/grid changed")
		}
		values := contextValues(p.Document.Slides[0].Values)
		if !bytes.Equal(values, baselineValues) {
			t.Fatal("source narrative facts changed")
		}
		seen := 0
		for _, n := range local.Nodes {
			if original, ok := retained[n.ID]; ok {
				seen++
				if !reflect.DeepEqual(n, original) {
					t.Fatalf("Today/Target frame, current object, raw connector, transition or companion %s changed", n.ID)
				}
			}
		}
		if seen != len(retained) || len(local.Nodes) != len(retained)+count {
			t.Fatalf("unrelated objects dropped or target count wrong: retained=%d nodes=%d", seen, len(local.Nodes))
		}
	}
	inspect, err := InspectLayers(p, "slide", selections, b, wmdesign.CandidateEngine)
	if err != nil || inspect.RenderError != "" || len(inspect.Layers) != 4 {
		t.Fatalf("actual target inspection: %+v %v", inspect, err)
	}
	layout := inspect.Layout
	layout.Rect.H = 216 // Six 36pt target rows remain inside the unchanged Target frame.
	layout.Gap = 0
	layout.Palette = "sequence"
	labelA, labelB := "Governance", "Observability"
	description := "Explicit target responsibility"
	order := []string{"node18", "target-governance", "node19", "node20", "target-observability", "node21"}
	patch := LayerPatch{Schema: LayerPatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "qualification", Reason: "Compose six explicitly selected Target layers; preserve all Today raw links, both frames and transition semantics", Layers: selections, Operations: []LayerOperation{
		{Action: "materialize", Entity: "source"},
		{Action: "add", Entity: "layer", Key: "target-governance", Prototype: "node19", Label: &labelA, Text: &description},
		{Action: "add", Entity: "layer", Key: "target-observability", Prototype: "node19", Label: &labelB, Text: &description},
		{Action: "reorder", Entity: "layer", Order: order},
		{Action: "set", Entity: "layout", Layout: &layout},
	}}
	before, _ := os.ReadFile(p.SourcePath)
	if preview, err := PatchLayers(p, "slide", patch, b, wmdesign.CandidateEngine, false); err != nil || preview.Applied {
		t.Fatalf("six-layer measured preview: %+v %v", preview, err)
	}
	if after, _ := os.ReadFile(p.SourcePath); !bytes.Equal(before, after) {
		t.Fatal("preview changed source")
	}
	if _, err = PatchLayers(p, "slide", patch, b, wmdesign.CandidateEngine, true); err != nil {
		t.Fatal(err)
	}
	p, err = Load(p.SourcePath)
	if err != nil {
		t.Fatal(err)
	}
	verifyRetained(p, 6)
	selections = DiscoverLayerSelections(p.Document.LocalTemplates["catalog"])
	if len(selections) != 6 {
		t.Fatal("six stable target identities absent", selections)
	}
	order = []string{"node21", "target-observability", "node20", "node19", "node18"}
	patch.ExpectedSourceSHA256 = p.SourceHash()
	patch.Layers = selections
	patch.Operations = []LayerOperation{{Action: "remove", Entity: "layer", Key: "target-governance"}, {Action: "reorder", Entity: "layer", Order: order}, {Action: "set", Entity: "layout", Layout: &layout}}
	if _, err = PatchLayers(p, "slide", patch, b, wmdesign.CandidateEngine, true); err != nil {
		t.Fatal("five-layer remove/reorder", err)
	}
	p, err = Load(p.SourcePath)
	if err != nil {
		t.Fatal(err)
	}
	verifyRetained(p, 5)
	inspect, err = InspectLayers(p, "slide", nil, b, wmdesign.CandidateEngine)
	if err != nil || inspect.RenderError != "" || len(inspect.Layers) != 5 {
		t.Fatalf("final source render failed: %+v %v", inspect, err)
	}
	for i, entry := range inspect.Layers {
		if entry.Key != order[i] {
			t.Fatalf("auto-inspection forgot authored reorder: got %s at %d, want %s", entry.Key, i, order[i])
		}
	}
	// A future layout-only edit must retain numbered semantic order even though
	// the unrelated paint array and all Today objects were preserved verbatim.
	patch.ExpectedSourceSHA256 = p.SourceHash()
	patch.Layers = DiscoverLayerSelections(p.Document.LocalTemplates["catalog"])
	patch.Operations = []LayerOperation{{Action: "set", Entity: "layout", Layout: &layout}}
	if _, err = PatchLayers(p, "slide", patch, b, wmdesign.CandidateEngine, true); err != nil {
		t.Fatal("future layout with retained authored order", err)
	}
	p, err = Load(p.SourcePath)
	if err != nil {
		t.Fatal(err)
	}
	verifyRetained(p, 5)
	inspect, err = InspectLayers(p, "slide", nil, b, wmdesign.CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	for i, entry := range inspect.Layers {
		if entry.Key != order[i] || math.Abs(entry.Rect.Y-(layout.Rect.Y+float64(i)*layout.Rect.H/5)) > 1e-8 {
			t.Fatal("future layout reset authored row order", entry)
		}
	}
}

func TestLayerCompositionAutoNumberingAmbiguityAtomic(t *testing.T) {
	for _, kind := range []string{"duplicate", "zero", "negative", "label", "fraction"} {
		t.Run(kind, func(t *testing.T) {
			p, _ := todayTargetLayerProject(t)
			local := p.Document.LocalTemplates["catalog"]
			for i := range local.Nodes {
				if local.Nodes[i].ID != "node19" {
					continue
				}
				value := any("1")
				switch kind {
				case "zero":
					value = "0"
				case "negative":
					value = "-1"
				case "label":
					value = "Integration"
				case "fraction":
					value = "2.5"
				}
				// Retain the genuine scaffold's n binding and its declared zone;
				// author the invalid observation rather than orphaning a zone.
				binding := local.Nodes[i].Arguments["n"].(map[string]any)["binding"].(string)
				p.Document.Slides[0].Values[binding] = value
			}
			p.Document.LocalTemplates["catalog"] = local
			p = writeCurveFixture(t, p)
			before, _ := os.ReadFile(p.SourcePath)
			if _, err := InspectLayers(p, "slide", nil, curveCompositionBundle(t), wmdesign.CandidateEngine); err == nil {
				t.Fatal("ambiguous automatic numbering accepted")
			}
			if len(DiscoverLayerSelections(local, p.Document.Slides[0].Values)) != 0 {
				t.Fatal("ambiguous automatic numbering silently used declaration order")
			}
			if after, _ := os.ReadFile(p.SourcePath); !bytes.Equal(before, after) {
				t.Fatal("inspection changed source")
			}
		})
	}
}

func TestLayerCompositionAuthoredNumberBindingsOrder(t *testing.T) {
	p, _ := todayTargetLayerProject(t)
	local := p.Document.LocalTemplates["catalog"]
	values := map[string]any{}
	for i := range local.Nodes {
		n := &local.Nodes[i]
		if n.ID < "node18" || n.ID > "node21" {
			continue
		}
		key := n.ID + ".authored-number"
		n.Arguments["n"] = map[string]any{"binding": key}
		values[key] = map[string]string{"node18": "4", "node19": "1", "node20": "3", "node21": "2"}[n.ID]
	}
	got, err := discoverNumberedLayerSelections(local, values)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"node19", "node21", "node20", "node18"}
	for i, item := range got {
		if item.Key != want[i] {
			t.Fatal("explicit authored binding order lost", got)
		}
	}
	if len(got) != 4 || !reflect.DeepEqual(got, DiscoverLayerSelections(local, values)) {
		t.Fatal("resolved ordinal discovery differs", got)
	}
	if _, err = discoverNumberedLayerSelections(local, nil); err == nil {
		t.Fatal("unresolved numbering fell back to incidental declaration order")
	}
}

func TestLayerCompositionActualTodayTargetInvalidAtomic(t *testing.T) {
	for _, kind := range []string{"duplicate-member", "duplicate-order", "oversize", "stale"} {
		t.Run(kind, func(t *testing.T) {
			p, selections := todayTargetLayerProject(t)
			b := curveCompositionBundle(t)
			inspect, err := InspectLayers(p, "slide", selections, b, wmdesign.CandidateEngine)
			if err != nil {
				t.Fatal(err)
			}
			layout := inspect.Layout
			patch := LayerPatch{Schema: LayerPatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "qualification", Reason: "Refuse unsafe actual target-layer customization atomically", Layers: selections, Operations: []LayerOperation{{Action: "materialize", Entity: "source"}, {Action: "set", Entity: "layout", Layout: &layout}}}
			switch kind {
			case "duplicate-member":
				patch.Layers[1].Members = []string{"node18"}
			case "duplicate-order":
				patch.Operations = append(patch.Operations, LayerOperation{Action: "reorder", Entity: "layer", Order: []string{"node18", "node18", "node20", "node21"}})
			case "oversize":
				layout.Rect.H = 600
			case "stale":
				patch.ExpectedSourceSHA256 = "stale"
			}
			before, _ := os.ReadFile(p.SourcePath)
			if _, err = PatchLayers(p, "slide", patch, b, wmdesign.CandidateEngine, true); err == nil {
				t.Fatal("unsafe patch accepted")
			}
			if after, _ := os.ReadFile(p.SourcePath); !bytes.Equal(before, after) {
				t.Fatal("refused patch changed source")
			}
		})
	}
}
