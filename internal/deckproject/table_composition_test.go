package deckproject

import (
	"bytes"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"os"
	"strings"
	"testing"
)

func tableFixture(t *testing.T) *Project {
	p := processFixtureUnpinned(t)
	local := p.Document.LocalTemplates["process"]
	n := &local.Nodes[0]
	n.Definition.ID = "wmds/component/table"
	n.Arguments = map[string]any{"header": "light", "rowHeader": true, "rowH": 36, "heatMax": 4, "groupW": 24, "cols": []any{map[string]any{"k": "name", "label": "Item", "w": 272}, map[string]any{"k": "status", "label": "Status", "w": 272, "type": "status", "labels": map[string]any{"pass": "Ready", "risk": "At risk", "blocked": "Hold"}}, map[string]any{"k": "score", "label": "Score", "w": 272, "type": "heat", "max": 4, "scale": "risk", "showValue": true}}, "rows": []any{map[string]any{"name": "One", "status": "pass", "score": 0, "h": 42}, map[string]any{"name": "Two", "status": "risk", "score": nil, "ink": "primary"}, map[string]any{"name": "Three", "status": "blocked", "score": 2}}, "rowGroups": []any{map[string]any{"label": "Group A", "from": 0, "to": 1, "fill": "deemph.1"}}, "groups": []any{map[string]any{"label": "Signals", "from": 1, "to": 2}}}
	n.Keys = map[string][]string{"rows": {"one", "two", "three"}, "rowGroups": {"group-a"}, "groups": {"signals"}}
	p.Document.LocalTemplates["process"] = local
	if e := os.WriteFile(p.SourcePath, canonical(p.Document), 0600); e != nil {
		t.Fatal(e)
	}
	p, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if _, e := Pin(p, journeyBundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	return p
}
func tablePatch(p *Project, ops ...TableOperation) TablePatch {
	return TablePatch{Schema: TablePatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "Test", Reason: "Illustrative table customization", SemanticReview: "Scores retain 0..4 risk domain; independent rows and status legend remain consistent", NodeID: "flow", Operations: ops}
}
func TestTableCompositionReorderRetainsCellsMetadataAndGroups(t *testing.T) {
	t.Parallel()
	p := tableFixture(t)
	inspection, e := InspectTable(p, "flow-slide", "flow", journeyBundle(t), wmdesign.CandidateEngine)
	if e != nil || inspection.RenderError != "" {
		t.Fatalf("%v %s", e, inspection.RenderError)
	}
	patch := tablePatch(p, TableOperation{Action: "reorder", Entity: "row", Order: []string{"two", "one", "three"}}, TableOperation{Action: "reorder", Entity: "column", Order: []string{"name", "score", "status"}})
	before := append([]byte{}, p.Raw...)
	if _, e = PatchTable(p, "flow-slide", patch, journeyBundle(t), wmdesign.CandidateEngine, false); e != nil {
		t.Fatal(e)
	}
	disk, _ := os.ReadFile(p.SourcePath)
	if !bytes.Equal(before, disk) {
		t.Fatal("preview wrote source")
	}
	if _, e = PatchTable(p, "flow-slide", patch, journeyBundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	inspection, e = InspectTable(p, "flow-slide", "flow", journeyBundle(t), wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if inspection.Model.Rows[1].Values["score"] != float64(0) || inspection.Model.Rows[0].Values["score"] != nil || inspection.Model.Rows[1].Values["h"] != float64(42) || inspection.Model.Columns[1]["max"] != float64(4) || inspection.Model.RowGroups[0].Members[0] != "two" || inspection.Model.ColumnGroups[0].Members[0] != "score" {
		t.Fatalf("lost semantic data %+v", inspection.Model)
	}
	if _, e = PatchTable(p, "flow-slide", patch, journeyBundle(t), wmdesign.CandidateEngine, true); e == nil {
		t.Fatal("stale accepted")
	}
}
func TestTableCompositionVariableRowsColumnsAndExplicitNull(t *testing.T) {
	t.Parallel()
	p := tableFixture(t)
	row := map[string]any{"name": "Four", "status": "pass", "score": nil}
	patch := tablePatch(p, TableOperation{Action: "set", Entity: "row", Key: "four", Row: row}, TableOperation{Action: "set", Entity: "cell", Key: "one", Column: "score", Value: nil})
	if _, e := PatchTable(p, "flow-slide", patch, journeyBundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	m, e := InspectTable(p, "flow-slide", "flow", journeyBundle(t), wmdesign.CandidateEngine)
	if e != nil || len(m.Model.Rows) != 4 || m.Model.Rows[0].Values["score"] != nil {
		t.Fatalf("row/null %v %+v", e, m.Model)
	}
	col := map[string]any{"k": "owner", "label": "Owner", "w": 272}
	ops := []TableOperation{{Action: "set", Entity: "column", Key: "owner", ColumnValue: col, Cells: map[string]any{"one": "A", "two": "B", "three": nil, "four": "D"}}, {Action: "remove", Entity: "column", Key: "status", Cascade: true}, {Action: "remove", Entity: "row", Key: "two", Cascade: true}}
	if _, e = PatchTable(p, "flow-slide", tablePatch(p, ops...), journeyBundle(t), wmdesign.CandidateEngine, false); e != nil {
		t.Fatal(e)
	}
}
func TestTableCompositionInvalidDependenciesDomainsAndPresence(t *testing.T) {
	t.Parallel()
	p := tableFixture(t)
	cases := []TableOperation{{Action: "reorder", Entity: "row", Order: []string{"one", "three", "two"}}, {Action: "remove", Entity: "column", Key: "score"}, {Action: "set", Entity: "cell", Key: "one", Column: "score", Value: 5}, {Action: "set", Entity: "cell", Key: "one", Column: "missing", Value: 0}, {Action: "set", Entity: "row", Key: "new", Row: map[string]any{"name": "Incomplete"}}, {Action: "set", Entity: "row", Key: "one", Row: map[string]any{"name": strings.Repeat("Too long ", 200)}}}
	for _, op := range cases {
		if _, e := PatchTable(p, "flow-slide", tablePatch(p, op), journeyBundle(t), wmdesign.CandidateEngine, false); e == nil {
			t.Fatal("invalid table accepted", op)
		}
	}
	raw := canonical(tablePatch(p, TableOperation{Action: "set", Entity: "cell", Key: "one", Column: "score", Value: nil}))
	if _, e := DecodeTablePatch(raw, "null"); e != nil {
		t.Fatal(e)
	}
	raw = bytes.Replace(raw, []byte(`"column":"score"`), []byte(`"column":"score","cascade":false`), 1)
	if _, e := DecodeTablePatch(raw, "bad"); e == nil {
		t.Fatal("irrelevant false accepted")
	}
}

func TestTableCompositionHeatDomainsAndHeaderMeaning(t *testing.T) {
	t.Parallel()
	p := tableFixture(t)
	for _, op := range []TableOperation{
		{Action: "reorder", Entity: "column", Order: []string{"status", "name", "score"}},
		{Action: "remove", Entity: "column", Key: "name", Cascade: true},
		{Action: "set", Entity: "layout", Options: map[string]any{"heatMax": 1}},
		{Action: "set", Entity: "column", Key: "score", ColumnValue: map[string]any{"k": "score", "label": "Score", "w": 272, "type": "heat", "max": 1, "scale": "risk"}},
		{Action: "set", Entity: "cell", Key: "one", Column: "score", Value: map[string]any{"value": -1, "text": "Low"}},
	} {
		// Column max overrides global heatMax, so changing just the global setting is
		// valid here. The retained explicit max stays authoritative.
		_, e := PatchTable(p, "flow-slide", tablePatch(p, op), journeyBundle(t), wmdesign.CandidateEngine, false)
		if op.Entity == "layout" {
			if e != nil {
				t.Fatal(e)
			}
		} else if e == nil {
			t.Fatal("meaning change accepted", op)
		}
	}
	if _, e := PatchTable(p, "flow-slide", tablePatch(p, TableOperation{Action: "set", Entity: "cell", Key: "one", Column: "score", Value: map[string]any{"value": 1.5, "text": "Mid"}}), journeyBundle(t), wmdesign.CandidateEngine, false); e != nil {
		t.Fatal(e)
	}
}

func TestTableCompositionNestedCellKeysFollowRows(t *testing.T) {
	t.Parallel()
	p := tableFixture(t)
	local := p.Document.LocalTemplates["process"]
	n := &local.Nodes[0]
	n.Arguments["rowGroups"].([]any)[0].(map[string]any)["label"] = "G"
	cols := n.Arguments["cols"].([]any)
	cols[1].(map[string]any)["type"] = "bullets"
	delete(cols[1].(map[string]any), "labels")
	rows := n.Arguments["rows"].([]any)
	for i, v := range rows {
		v.(map[string]any)["status"] = []any{"A", "B"}
		if n.Keys == nil {
			n.Keys = map[string][]string{}
		}
		n.Keys["rows/"+string(rune('0'+i))+"/status"] = []string{"a", "b"}
	}
	p.Document.LocalTemplates["process"] = local
	if e := os.WriteFile(p.SourcePath, canonical(p.Document), 0600); e != nil {
		t.Fatal(e)
	}
	var e error
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	patch := tablePatch(p, TableOperation{Action: "reorder", Entity: "row", Order: []string{"two", "one", "three"}}, TableOperation{Action: "remove", Entity: "row", Key: "one", Cascade: true})
	if _, e = PatchTable(p, "flow-slide", patch, journeyBundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	n = &p.Document.LocalTemplates["process"].Nodes[0]
	if !bytes.Equal(canonical(n.Keys["rows/0/status"]), canonical([]string{"a", "b"})) {
		t.Fatalf("cell keys lost %+v", n.Keys)
	}
	if _, ok := n.Keys["rows/2/status"]; ok {
		t.Fatal("removed row cell keys retained")
	}
}

func TestTableCompositionSourceGeometryRetained(t *testing.T) {
	t.Parallel()
	p := tableFixture(t)
	local := p.Document.LocalTemplates["process"]
	metadata := map[string]any{"schema": wmdesign.SceneSourceGeometrySchema, "x_fraction": 0, "y_fraction": 0, "width_fraction": 1}
	local.Nodes[0].Arguments[wmdesign.SceneSourceGeometryArgument] = metadata
	p.Document.LocalTemplates["process"] = local
	if e := os.WriteFile(p.SourcePath, canonical(p.Document), 0600); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	if _, e := PatchTable(p, "flow-slide", tablePatch(p, TableOperation{Action: "set", Entity: "cell", Key: "one", Column: "name", Value: "Updated"}), journeyBundle(t), wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.SourcePath)
	if !bytes.Equal(canonical(metadata), canonical(p.Document.LocalTemplates["process"].Nodes[0].Arguments[wmdesign.SceneSourceGeometryArgument])) {
		t.Fatal("table source anchor lost")
	}
}
