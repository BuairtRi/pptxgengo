package deckproject

import (
	"encoding/json"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"os"
	"testing"
)

func TestSourceContainerOriginalAndFixedClearance(t *testing.T) {
	if _, err := ScaffoldTemplateWithOptions(curveCompositionBundle(t), "architecture/layer-map", wmdesign.CandidateEngine, "Strict original source", 2026, ScaffoldOptions{PlaceholderMedia: true}); err == nil {
		t.Fatal("original three-point caption clearance silently altered")
	}
	p, scaffold := placeholderScaffoldProject(t, "architecture/layer-map", ScaffoldOptions{PlaceholderMedia: true, SourceContainerClearanceFit: true})
	if len(scaffold.GeometryAdjustments) != 1 || scaffold.GeometryAdjustments[0].Original.H != 294 || scaffold.GeometryAdjustments[0].Authored.H != 291 {
		t.Fatal(scaffold.GeometryAdjustments)
	}
	index := -1
	for i, n := range scaffold.Template.Nodes {
		if n.Placement.Zone == "source_container" {
			index = i
		}
	}
	if index < 0 {
		t.Fatal("actual source container missing")
	}
	original := scaffold.Template.Nodes[index]
	t.Logf("actual original container %s geometry=%s placement=%+v", original.ID, canonical(original.Arguments[wmdesign.SceneSourceGeometryArgument]), original.Placement.Rect)
	if _, err := Compile(p, curveCompositionBundle(t), wmdesign.CandidateEngine); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"extent", "identity", "kind", "ancestor", "no-parent", "no-metadata", "new-node", "snapshot"} {
		t.Run(mode, func(t *testing.T) {
			var doc Document
			if err := json.Unmarshal(canonical(p.Document), &doc); err != nil {
				t.Fatal(err)
			}
			local := doc.LocalTemplates["portable"]
			n := &local.Nodes[index]
			switch mode {
			case "extent":
				n.Placement.Rect.H += .1
			case "identity":
				n.ID = "invented"
			case "kind":
				n.Definition.ID = "wmds/component/block"
			case "ancestor":
				local.Provenance.SourceFileSHA256 = "0000000000000000000000000000000000000000000000000000000000000000"
			case "no-parent":
				local.Provenance = nil
			case "no-metadata":
				delete(n.Arguments, wmdesign.SceneSourceGeometryArgument)
			case "new-node":
				n.ID = "new-local"
			case "snapshot":
				raw := []byte(`{"fake":"ancestor"}`)
				local.Provenance.DefinitionSnapshot = "fake-ancestor.json"
				local.Provenance.DefinitionSHA256 = digest(raw)
				if err := os.WriteFile(p.Root+"/fake-ancestor.json", raw, 0600); err != nil {
					t.Fatal(err)
				}
			}
			doc.LocalTemplates["portable"] = local
			candidate := *p
			candidate.Document = doc
			if _, err := Compile(&candidate, curveCompositionBundle(t), wmdesign.CandidateEngine); err == nil {
				t.Fatal("forged/overflowing source container accepted")
			}
		})
	}
	local := p.Document.LocalTemplates["portable"]
	n := local.Nodes[index]
	n.Arguments[wmdesign.SceneSourceGeometryArgument] = map[string]any{"schema": wmdesign.SceneSourceGeometrySchema, "x_fraction": 0., "y_fraction": 0., "width_fraction": 1., "height_fraction": 2.}
	if err := validateSourceContainerOriginal(local, n, scaffold); err == nil {
		t.Fatal("forged source extent accepted")
	}
	f := wmdesign.ResolvedFrame{Body: wmdesign.Rect{X: 75, Y: 128, W: 783, H: 322}, Source: wmdesign.Rect{X: 75, Y: 462, W: 783, H: 12}}
	if got := sourceContainerZone(f).H; got != 456 {
		t.Fatal(got)
	}
	f.Source.Y = 457
	if got := sourceContainerZone(f).H; got != 451 {
		t.Fatal("caption clearance ignored", got)
	}
	f.Source = wmdesign.Rect{}
	f.FooterRule = 458
	if got := sourceContainerZone(f).H; got != 452 {
		t.Fatal("footer clearance ignored", got)
	}
	if _, _, err := chooseSourceSceneZone("block", map[string]any{}, wmdesign.Rect{X: 75, Y: 128, W: 100, H: 328}, f); err == nil {
		t.Fatal("generic component gained frame exception")
	}
}
