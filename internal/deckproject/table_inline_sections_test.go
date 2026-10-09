package deckproject

import (
	"bytes"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"os"
	"strings"
	"testing"
)

func tableInlineFixture(t *testing.T) *Project {
	t.Helper()
	p := tableFixture(t)
	local := p.Document.LocalTemplates["process"]
	n := &local.Nodes[0]
	delete(n.Arguments, "rowGroups")
	delete(n.Arguments, "groupW")
	for _, column := range n.Arguments["cols"].([]any) {
		column.(map[string]any)["w"] = 282
	}
	n.Arguments["rows"] = []any{
		map[string]any{"group": "Clinical"},
		map[string]any{"name": "One", "status": "pass", "score": 0},
		map[string]any{"name": "Two", "status": "risk", "score": nil},
		map[string]any{"group": "Corporate"},
		map[string]any{"name": "Three", "status": "blocked", "score": 2},
	}
	n.Keys["rows"] = []string{"clinical", "one", "two", "corporate", "three"}
	delete(n.Keys, "rowGroups")
	p.Document.LocalTemplates["process"] = local
	if e := os.WriteFile(p.SourcePath, canonical(p.Document), 0600); e != nil {
		t.Fatal(e)
	}
	p, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func TestTableInlineSectionsStableMembershipAndCountChanges(t *testing.T) {
	p := tableInlineFixture(t)
	b := journeyBundle(t)
	before, e := InspectTable(p, "flow-slide", "flow", b, wmdesign.CandidateEngine)
	if e != nil || len(before.Model.InlineSections) != 2 {
		t.Fatal(e, before)
	}
	ops := []TableOperation{
		{Action: "set", Entity: "row", Key: "four", Section: "clinical", Row: map[string]any{"name": "Four", "status": "pass", "score": 0}},
		{Action: "reorder", Entity: "row", Order: []string{"corporate", "three", "clinical", "four", "two", "one"}},
		{Action: "remove", Entity: "row", Key: "two", Cascade: true},
		{Action: "move", Entity: "row", Key: "one", Section: "corporate"},
	}
	if _, e = PatchTable(p, "flow-slide", tablePatch(p, ops...), b, wmdesign.CandidateEngine, true); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	after, e := InspectTable(p, "flow-slide", "flow", b, wmdesign.CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	if len(after.Model.Rows) != 5 || after.Model.InlineSections[0].Header != "corporate" || strings.Join(after.Model.InlineSections[0].Members, ",") != "three,one" || strings.Join(after.Model.InlineSections[1].Members, ",") != "four" || after.Model.Rows[2].Values["score"] != float64(0) {
		t.Fatalf("lost explicit group meaning %+v", after.Model)
	}
}
func TestTableInlineSectionsRefuseImplicitMembershipAndKeepSource(t *testing.T) {
	p := tableInlineFixture(t)
	b := journeyBundle(t)
	original := append([]byte{}, p.Raw...)
	cases := []TableOperation{
		{Action: "reorder", Entity: "row", Order: []string{"three", "corporate", "two", "one", "clinical"}},
		{Action: "reorder", Entity: "row", Order: []string{"clinical", "one", "corporate", "two", "three"}},
		{Action: "remove", Entity: "row", Key: "clinical", Cascade: true},
		{Action: "move", Entity: "row", Key: "one", Section: "missing"},
		{Action: "move", Entity: "row", Key: "clinical", Section: "corporate"},
		{Action: "set", Entity: "row", Key: "new", Row: map[string]any{"name": "New", "status": "pass", "score": nil}},
		{Action: "set", Entity: "row", Key: "one", Section: "corporate", Row: map[string]any{"name": "One"}},
		{Action: "set", Entity: "cell", Key: "one", Column: "name", Value: strings.Repeat("Long label ", 200)},
	}
	for _, op := range cases {
		if _, e := PatchTable(p, "flow-slide", tablePatch(p, op), b, wmdesign.CandidateEngine, true); e == nil {
			t.Fatal("invalid meaning/fit accepted", op)
		}
		disk, _ := os.ReadFile(p.SourcePath)
		if !bytes.Equal(original, disk) {
			t.Fatal("refusal changed source")
		}
	}
	if _, e := PatchTable(p, "flow-slide", tablePatch(p, TableOperation{Action: "remove", Entity: "row", Key: "one", Cascade: true}, TableOperation{Action: "remove", Entity: "row", Key: "two", Cascade: true}, TableOperation{Action: "remove", Entity: "row", Key: "clinical", Cascade: true}), b, wmdesign.CandidateEngine, false); e != nil {
		t.Fatal("empty reviewed section removal refused", e)
	}
}
func TestTableSparseLeadingLabelsRequireExplicitSourceMapping(t *testing.T) {
	p := tableFixture(t)
	local := p.Document.LocalTemplates["process"]
	rows := local.Nodes[0].Arguments["rows"].([]any)
	rows[1].(map[string]any)["name"] = ""
	p.Document.LocalTemplates["process"] = local
	if e := os.WriteFile(p.SourcePath, canonical(p.Document), 0600); e != nil {
		t.Fatal(e)
	}
	p, e := Load(p.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	b := journeyBundle(t)
	inspection, e := InspectTable(p, "flow-slide", "flow", b, wmdesign.CandidateEngine)
	if e != nil || strings.Join(inspection.Model.SparseLeadingLabels, ",") != "two" {
		t.Fatal(e, inspection.Model)
	}
	order := TableOperation{Action: "reorder", Entity: "row", Order: []string{"two", "one", "three"}}
	original := append([]byte{}, p.Raw...)
	if _, e = PatchTable(p, "flow-slide", tablePatch(p, order), b, wmdesign.CandidateEngine, true); e == nil || !strings.Contains(e.Error(), "sparse leading") {
		t.Fatal("blank label meaning was inferred", e)
	}
	disk, _ := os.ReadFile(p.SourcePath)
	if !bytes.Equal(original, disk) {
		t.Fatal("refusal changed source")
	}
	if _, e = PatchTable(p, "flow-slide", tablePatch(p, TableOperation{Action: "set", Entity: "cell", Key: "two", Column: "name", Value: "Two"}, order), b, wmdesign.CandidateEngine, false); e != nil {
		t.Fatal("explicit per-row source mapping refused", e)
	}
}

func TestTableActualCatalogSectionAndPhaseSemantics(t *testing.T) {
	for _, key := range []string{"app-portfolio-heat/by-domain", "activities/by-phase-table"} {
		t.Run(key, func(t *testing.T) {
			b := journeyBundle(t)
			scaffold, e := ScaffoldTemplateWithOptions(b, key, wmdesign.CandidateEngine, "Review actual source declared groups and sparse phase context", 2026, ScaffoldOptions{PlaceholderMedia: true})
			if e != nil {
				t.Fatal(e)
			}
			p := example(t)
			p.Document.Assets = scaffold.Assets
			p.Document.Context = nil
			p.Document.LocalTemplates = map[string]LocalTemplate{"catalog": scaffold.Template}
			p.Document.Slides = []Slide{{ID: "catalog-slide", ContentKind: "synthetic_example", Template: Reference{Scope: "local", ID: "catalog"}, Values: scaffold.SyntheticSourceValues}}
			if e = os.WriteFile(p.SourcePath, canonical(p.Document), 0600); e != nil {
				t.Fatal(e)
			}
			p, e = Load(p.SourcePath)
			if e != nil {
				t.Fatal(e)
			}
			if _, e = Pin(p, b, wmdesign.CandidateEngine); e != nil {
				t.Fatal(e)
			}
			inspection, e := InspectTable(p, "catalog-slide", "node01", b, wmdesign.CandidateEngine)
			if e != nil || inspection.RenderError != "" {
				t.Fatal(e, inspection.RenderError)
			}
			order := []string{}
			for i := len(inspection.Model.Rows) - 1; i >= 0; i-- {
				order = append(order, inspection.Model.Rows[i].Key)
			}
			patch := TablePatch{Schema: TablePatchSchema, ExpectedSourceSHA256: p.SourceHash(), Actor: "Qualification operator", Reason: "Reject flat source reversal that loses group or phase context", SemanticReview: "Source scores, phase associations and source facts remain explicit", NodeID: "node01", Operations: []TableOperation{{Action: "materialize", Entity: "source"}, {Action: "reorder", Entity: "row", Order: order}}}
			before := append([]byte{}, p.Raw...)
			if _, e = PatchTable(p, "catalog-slide", patch, b, wmdesign.CandidateEngine, true); e == nil {
				t.Fatal("actual source flat reversal accepted")
			}
			disk, _ := os.ReadFile(p.SourcePath)
			if !bytes.Equal(before, disk) {
				t.Fatal("refusal changed actual source")
			}
			if key == "app-portfolio-heat/by-domain" {
				if len(inspection.Model.InlineSections) != 3 {
					t.Fatal("missing source group declarations", inspection.Model)
				}
				order = nil
				for _, section := range inspection.Model.InlineSections {
					order = append(order, section.Header)
					for i := len(section.Members) - 1; i >= 0; i-- {
						order = append(order, section.Members[i])
					}
				}
				patch.Operations = []TableOperation{{Action: "materialize", Entity: "source"}, {Action: "reorder", Entity: "row", Order: order}}
			} else {
				if len(inspection.Model.SparseLeadingLabels) != 9 {
					t.Fatal("actual sparse phase labels not exposed", inspection.Model)
				}
				// These exact source example keys have an explicitly reviewed phase
				// map. This fixture authors each membership; production does not
				// derive phase membership from blank cells or fill colors.
				phases := []string{"Rationalize", "Rationalize", "Rationalize", "Rationalize", "Build", "Build", "Build", "Build", "Sustain", "Sustain", "Sustain", "Sustain"}
				patch.Operations = []TableOperation{{Action: "materialize", Entity: "source"}}
				for i, row := range inspection.Model.Rows {
					patch.Operations = append(patch.Operations, TableOperation{Action: "set", Entity: "cell", Key: row.Key, Column: "ph", Value: phases[i]})
				}
				patch.Operations = append(patch.Operations, TableOperation{Action: "reorder", Entity: "row", Order: order})
			}
			if _, e = PatchTable(p, "catalog-slide", patch, b, wmdesign.CandidateEngine, false); e != nil {
				t.Fatal("explicit context-preserving source preview refused", e)
			}
		})
	}
}
