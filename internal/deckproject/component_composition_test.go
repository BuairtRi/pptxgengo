package deckproject

import (
	"bytes"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"os"
	"strings"
	"testing"
)

func componentFixture(t *testing.T) *Project {
	p := processFixtureUnpinned(t)
	local := p.Document.LocalTemplates["process"]
	n := &local.Nodes[0]
	n.Definition.ID = "wmds/component/card"
	n.Arguments = map[string]any{"surface": "light", "title": "Illustrative card", "body": []any{map[string]any{"bullets": []any{"First observation", "Second observation"}}, map[string]any{"p": "Supporting note"}}}
	n.Keys = map[string][]string{"body": {"observations", "note"}, "body/0/bullets": {"first", "second"}}
	p.Document.LocalTemplates["process"] = local
	if e := os.WriteFile(p.SourcePath, canonical(p.Document), 0600); e != nil {
		t.Fatal(e)
	}
	p, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	pin(t, p)
	return p
}
func componentPatch(p *Project, ops ...ComponentOperation) ComponentPatch {
	return ComponentPatch{Schema: ComponentPatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "Test", Reason: "Illustrative content customization", SemanticReview: "Observations are independently authored; no dependent facts or legends", NodeID: "flow", Operations: ops}
}
func TestComponentCompositionNestedKeysPreviewApply(t *testing.T) {
	p := componentFixture(t)
	before := append([]byte{}, p.Raw...)
	patch := componentPatch(p, ComponentOperation{Action: "set", Entity: "item", Path: "/body/@observations/bullets", Key: "third", Value: "Third observation"}, ComponentOperation{Action: "reorder", Entity: "item", Path: "/body", Order: []string{"note", "observations"}}, ComponentOperation{Action: "set", Entity: "argument", Path: "/body/@note/p", Value: "Updated supporting note"})
	if _, e := PatchComponent(p, "flow-slide", patch, bundle(t), wmdesign.CandidateEngine, false); e != nil {
		t.Fatal(e)
	}
	disk, _ := os.ReadFile(p.SourcePath)
	if !bytes.Equal(before, disk) {
		t.Fatal("preview altered source")
	}
	if _, e := PatchComponent(p, "flow-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	n := p.Document.LocalTemplates["process"].Nodes[0]
	if !bytes.Equal(canonical(n.Keys["body/1/bullets"]), canonical([]string{"first", "second", "third"})) {
		t.Fatalf("nested identities lost %+v", n.Keys)
	}
	if _, exists := n.Keys["body/0/bullets"]; exists {
		t.Fatal("stale child keys retained")
	}
	if _, e := PatchComponent(p, "flow-slide", patch, bundle(t), wmdesign.CandidateEngine, true); e == nil {
		t.Fatal("stale accepted")
	}
	remove := componentPatch(p, ComponentOperation{Action: "remove", Entity: "item", Path: "/body/@observations/bullets", Key: "second", Cascade: true})
	if _, e := PatchComponent(p, "flow-slide", remove, bundle(t), wmdesign.CandidateEngine, false); e != nil {
		t.Fatal(e)
	}
}
func TestComponentCompositionRefusesNumericPathsSpecializedAndOverflow(t *testing.T) {
	p := componentFixture(t)
	for _, op := range []ComponentOperation{{Action: "set", Entity: "argument", Path: "/body/0/bullets/0", Value: "x"}, {Action: "remove", Entity: "item", Path: "/body", Key: "note"}, {Action: "set", Entity: "argument", Path: "/title", Value: strings.Repeat("Very long title ", 100)}} {
		if _, e := PatchComponent(p, "flow-slide", componentPatch(p, op), bundle(t), wmdesign.CandidateEngine, false); e == nil {
			t.Fatal("invalid operation accepted", op)
		}
	}
	q := processFixture(t)
	if _, e := PatchComponent(q, "flow-slide", componentPatch(q, ComponentOperation{Action: "set", Entity: "argument", Path: "/current", Value: "check"}), bundle(t), wmdesign.CandidateEngine, false); e == nil {
		t.Fatal("semantic model bypass accepted")
	}
}
func TestComponentCompositionExplicitNullAndStrictPresence(t *testing.T) {
	p := componentFixture(t)
	raw := canonical(componentPatch(p, ComponentOperation{Action: "set", Entity: "argument", Path: "/title", Value: nil}))
	if !bytes.Contains(raw, []byte(`"value":null`)) {
		t.Fatal("null omitted")
	}
	if _, e := DecodeComponentPatch(raw, "null"); e != nil {
		t.Fatal(e)
	}
	raw = bytes.Replace(raw, []byte(`"path":"/title"`), []byte(`"path":"/title","cascade":false`), 1)
	if _, e := DecodeComponentPatch(raw, "bad"); e == nil {
		t.Fatal("irrelevant false accepted")
	}
}

func TestComponentCompositionInspectionUsesStablePublicPaths(t *testing.T) {
	p := componentFixture(t)
	inspection, e := InspectComponent(p, "flow-slide", "flow", bundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	found := false
	for _, c := range inspection.Collections {
		if c.Path == "/body/@observations/bullets" {
			found = true
		}
	}
	if !found {
		t.Fatalf("nested stable selector absent %+v", inspection.Collections)
	}
}

func TestComponentCompositionSourceGeometryReservedAndRetained(t *testing.T) {
	p := componentFixture(t)
	local := p.Document.LocalTemplates["process"]
	metadata := map[string]any{"schema": wmdesign.SceneSourceGeometrySchema, "x_fraction": 0, "y_fraction": 0, "width_fraction": 1}
	local.Nodes[0].Arguments[wmdesign.SceneSourceGeometryArgument] = metadata
	p.Document.LocalTemplates["process"] = local
	if e := os.WriteFile(p.SourcePath, canonical(p.Document), 0600); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	if _, e := PatchComponent(p, "flow-slide", componentPatch(p, ComponentOperation{Action: "set", Entity: "argument", Path: "/_source_geometry/x_fraction", Value: 0.1}), bundle(t), wmdesign.CandidateEngine, false); e == nil {
		t.Fatal("reserved metadata mutated")
	}
	if _, e := PatchComponent(p, "flow-slide", componentPatch(p, ComponentOperation{Action: "set", Entity: "argument", Path: "/title", Value: "Updated card"}), bundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	if !bytes.Equal(canonical(metadata), canonical(p.Document.LocalTemplates["process"].Nodes[0].Arguments[wmdesign.SceneSourceGeometryArgument])) {
		t.Fatal("source anchor lost")
	}
}
