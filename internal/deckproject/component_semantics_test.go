package deckproject

import (
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"os"
	"testing"
)

func semanticComponentFixture(t *testing.T, kind string, args map[string]any, keys map[string][]string) *Project {
	p := processFixtureUnpinned(t)
	local := p.Document.LocalTemplates["process"]
	n := &local.Nodes[0]
	n.Definition.ID = "wmds/component/" + kind
	n.Arguments = args
	n.Keys = keys
	if kind == "phases" {
		n.Placement.Rect.H = 340
	}
	p.Document.LocalTemplates["process"] = local
	if e := os.WriteFile(p.SourcePath, canonical(p.Document), 0600); e != nil {
		t.Fatal(e)
	}
	p, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = Pin(p, journeyBundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	return p
}
func TestComponentSemanticVennMemberKeysSurviveReorder(t *testing.T) {
	p := semanticComponentFixture(t, "venn", map[string]any{"sets": []any{map[string]any{"label": "Access"}, map[string]any{"label": "Quality"}, map[string]any{"label": "Cost"}}, "regions": []any{map[string]any{"in": []any{0, 1}, "label": "Timely"}}}, map[string][]string{"sets": {"access", "quality", "cost"}, "regions": {"timely"}})
	patch := componentPatch(p, ComponentOperation{Action: "reorder", Entity: "item", Path: "/sets", Order: []string{"cost", "quality", "access"}})
	if _, e := PatchComponent(p, "flow-slide", patch, journeyBundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	m, e := InspectComponent(p, "flow-slide", "flow", journeyBundle(t), wmdesign.CandidateEngine)
	if e != nil || len(m.Relationships["timely"]) != 2 || m.Relationships["timely"][0] != "access" {
		t.Fatalf("wrong membership %v %+v", e, m.Relationships)
	}
	region := map[string]any{"members": []any{"quality", "cost"}, "label": "Both"}
	if _, e = PatchComponent(p, "flow-slide", componentPatch(p, ComponentOperation{Action: "set", Entity: "item", Path: "/regions", Key: "balanced", Value: region}), journeyBundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	if _, e = PatchComponent(p, "flow-slide", componentPatch(p, ComponentOperation{Action: "remove", Entity: "item", Path: "/sets", Key: "access", Cascade: true}), journeyBundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	m, e = InspectComponent(p, "flow-slide", "flow", journeyBundle(t), wmdesign.CandidateEngine)
	if e != nil || len(m.Relationships) != 1 || len(m.Relationships["balanced"]) != 2 {
		t.Fatalf("cascade wrong %v %+v", e, m.Relationships)
	}
}
func TestComponentSemanticPhaseCurrentKeyReorder(t *testing.T) {
	p := semanticComponentFixture(t, "phases", map[string]any{"phases": []any{map[string]any{"name": "Discover", "n": "1"}, map[string]any{"name": "Design", "n": "2"}, map[string]any{"name": "Deliver", "n": "3"}}, "current": 1}, map[string][]string{"phases": {"discover", "design", "deliver"}})
	patch := componentPatch(p, ComponentOperation{Action: "reorder", Entity: "item", Path: "/phases", Order: []string{"deliver", "discover", "design"}})
	if _, e := PatchComponent(p, "flow-slide", patch, journeyBundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	m, e := InspectComponent(p, "flow-slide", "flow", journeyBundle(t), wmdesign.CandidateEngine)
	if e != nil || m.Markers["current"] != "design" {
		t.Fatalf("current lost %v %+v", e, m.Markers)
	}
	if _, e = PatchComponent(p, "flow-slide", componentPatch(p, ComponentOperation{Action: "set", Entity: "marker", Path: "/current", Key: "discover"}), journeyBundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
}
func TestComponentSemanticMatrixColumnsKeepCells(t *testing.T) {
	p := semanticComponentFixture(t, "matrix", map[string]any{"labels": "left", "labelW": 100, "cols": []any{"Now", "Next"}, "cellH": 30, "rows": []any{map[string]any{"key": "delivery", "label": "Delivery", "cells": []any{"A", "B"}}, map[string]any{"key": "operations", "label": "Operations", "cells": []any{"C", "D"}}}}, map[string][]string{"cols": {"now", "next"}, "rows": {"delivery", "operations"}, "rows/0/cells": {"a", "b"}, "rows/1/cells": {"c", "d"}})
	patch := componentPatch(p, ComponentOperation{Action: "reorder", Entity: "item", Path: "/cols", Order: []string{"next", "now"}}, ComponentOperation{Action: "set", Entity: "item", Path: "/cols", Key: "later", Value: map[string]any{"label": "Later", "cells": map[string]any{"delivery": "E", "operations": "F"}}})
	if _, e := PatchComponent(p, "flow-slide", patch, journeyBundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	m, e := InspectComponent(p, "flow-slide", "flow", journeyBundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	rows := m.Arguments["rows"].([]any)
	cells := rows[0].(map[string]any)["cells"].([]any)
	if len(cells) != 3 || cells[0] != "B" || cells[1] != "A" || cells[2] != "E" {
		t.Fatalf("columns misassigned %v", cells)
	}
	n := p.Document.LocalTemplates["process"].Nodes[0]
	if n.Keys["rows/0/cells"][0] != "b" {
		t.Fatal("native cell identity lost")
	}
}

func TestComponentSemanticRefusesPositionalAndIncompleteMappings(t *testing.T) {
	p := semanticComponentFixture(t, "venn", map[string]any{"sets": []any{map[string]any{"label": "A"}, map[string]any{"label": "B"}, map[string]any{"label": "C"}}, "regions": []any{map[string]any{"in": []any{0, 1}, "label": "Both"}}}, map[string][]string{"sets": {"a", "b", "c"}, "regions": {"both"}})
	for _, op := range []ComponentOperation{
		{Action: "set", Entity: "item", Path: "/regions", Key: "new", Value: map[string]any{"in": []any{0, 1}, "label": "Two"}},
		{Action: "set", Entity: "argument", Path: "/regions/@both/in", Value: []any{1, 2}},
		{Action: "set", Entity: "item", Path: "/regions", Key: "new", Value: map[string]any{"members": []any{"a", "missing"}, "label": "Two"}},
		{Action: "remove", Entity: "item", Path: "/sets", Key: "a"},
	} {
		if _, e := PatchComponent(p, "flow-slide", componentPatch(p, op), journeyBundle(t), wmdesign.CandidateEngine, false); e == nil {
			t.Fatal("invalid semantic mapping accepted", op)
		}
	}
	p = semanticComponentFixture(t, "matrix", map[string]any{"labels": "left", "labelW": 100, "cols": []any{"Now"}, "cellH": 30, "rows": []any{map[string]any{"key": "delivery", "label": "Delivery", "cells": []any{"A"}}}}, map[string][]string{"cols": {"now"}, "rows": {"delivery"}, "rows/0/cells": {"a"}})
	if _, e := PatchComponent(p, "flow-slide", componentPatch(p, ComponentOperation{Action: "set", Entity: "item", Path: "/cols", Key: "next", Value: map[string]any{"label": "Next", "cells": map[string]any{}}}), journeyBundle(t), wmdesign.CandidateEngine, false); e == nil {
		t.Fatal("matrix omitted cell value accepted")
	}
}

func TestComponentSemanticActualVennSourceVariableCount(t *testing.T) {
	p := sourceProfileProject(t, journeyBundle(t), "venn/three-text")
	node := ""
	for _, n := range p.Document.LocalTemplates["catalog"].Nodes {
		if n.Definition != nil && n.Definition.ID == "wmds/component/venn" {
			node = n.ID
			break
		}
	}
	if node == "" {
		t.Fatal("actual source Venn missing")
	}
	inspect, e := InspectComponent(p, "catalog-slide", node, journeyBundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	keys := []string{}
	for _, c := range inspect.Collections {
		if c.Path == "/sets" {
			keys = c.Keys
		}
	}
	if len(keys) != 3 {
		t.Fatal("actual source did not have three sets")
	}
	patch := ComponentPatch{Schema: ComponentPatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "Qualification", Reason: "Explicit illustrative3-to-2 concept change with incident region removal", SemanticReview: "Remaining sets and regions retain membership; source copy/provenance remains authored; source metrics are not inferred", NodeID: node, Operations: []ComponentOperation{{Action: "materialize", Entity: "source"}, {Action: "remove", Entity: "item", Path: "/sets", Key: keys[2], Cascade: true}}}
	if _, e = PatchComponent(p, "catalog-slide", patch, journeyBundle(t), wmdesign.CandidateEngine, false); e != nil {
		t.Fatal(e)
	}
}
