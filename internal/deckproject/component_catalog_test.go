package deckproject

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// Whole-frame sources retain companion nodes. This checks one actual source per
// family; it is not a claim that every catalog variant has passed desktop QA.
func TestComponentTableCatalogEveryFamilyRepresentative(t *testing.T) {
	if testing.Short() {
		t.Skip("29-family composition catalog exercise belongs to nightly/on-demand")
	}
	b := journeyBundle(t)
	catalog, e := wmdesign.LibraryCatalog(b, "")
	if e != nil {
		t.Fatal(e)
	}
	byFamily := map[string][]wmdesign.LibraryTemplate{}
	for _, def := range catalog {
		byFamily[def.Family] = append(byFamily[def.Family], def)
	}
	families := []string{}
	for family := range byFamily {
		families = append(families, family)
	}
	sort.Strings(families)
	if len(families) != 29 {
		t.Fatal("catalog family count changed", len(families))
	}
	qualified := 0
	combined := example(t)
	combined.Document.Assets = map[string]Asset{}
	combined.Document.Context = nil
	combined.Document.LocalTemplates = map[string]LocalTemplate{}
	combined.Document.Slides = nil
	combinedReceipts := []map[string]any{}
	for _, family := range families {
		t.Run(family, func(t *testing.T) {
			var selected *Project
			nodeID, kind, key := "", "", ""
			var placeholderReceipts []ScaffoldPlaceholder
			failures := []string{}
			for _, def := range byFamily[family] {
				scaffold, e := ScaffoldTemplateWithOptions(b, def.Key, wmdesign.CandidateEngine, "Explicit illustrative family customization qualification; declared distribution-excluded media uses labeled placeholders", 2026, ScaffoldOptions{PlaceholderMedia: true})
				if e != nil {
					failures = append(failures, def.Key+": "+e.Error())
					continue
				}
				for _, node := range scaffold.Template.Nodes {
					if node.Definition == nil {
						continue
					}
					candidateKind := strings.TrimPrefix(node.Definition.ID, "wmds/component/")
					if !catalogQualificationLeafKind(candidateKind) {
						continue
					}
					if command := componentSpecialized(candidateKind); command != "" && command != "table" {
						continue
					}
					args, bound, e := componentArgs(&node, scaffold.SyntheticSourceValues)
					_ = bound
					if e != nil {
						continue
					}
					if candidateKind == "table" || candidateKind == "editable-table" {
						rows, ok := args["rows"].([]any)
						if !ok || len(rows) < 1 {
							continue
						}
					} else {
						// Empty copy can be an intentional decorative backing for a
						// separate companion heading. Customize authored visible copy,
						// never populate a blank backing and collide with that heading.
						scalar := false
						for _, field := range []string{"text", "title", "label", "name", "meta", "role", "value"} {
							if value, ok := args[field].(string); ok && strings.TrimSpace(value) != "" {
								scalar = true
							}
						}
						collections, e := componentCollections(&node, args)
						if e != nil || (!scalar && len(collections) == 0) {
							continue
						}
					}
					p := example(t)
					p.Document.Assets = scaffold.Assets
					for path, payload := range scaffold.AssetPayloads {
						target := filepath.Join(p.Root, path)
						if e = os.MkdirAll(filepath.Dir(target), 0700); e != nil {
							t.Fatal(e)
						}
						if e = os.WriteFile(target, payload, 0600); e != nil {
							t.Fatal(e)
						}
					}
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
					selected = p
					nodeID = node.ID
					kind = candidateKind
					key = def.Key
					placeholderReceipts = scaffold.PlaceholderMedia
					break
				}
				if selected != nil {
					break
				}
			}
			if selected == nil {
				t.Fatalf("no usable actual source for family %s: %v", family, failures)
			}
			qualified++
			t.Logf("selected %s node %s (%s), full source retained", key, nodeID, kind)
			if kind == "table" || kind == "editable-table" {
				inspection, e := InspectTable(selected, "catalog-slide", nodeID, b, wmdesign.CandidateEngine)
				if e != nil || inspection.RenderError != "" {
					t.Fatalf("table %v %s", e, inspection.RenderError)
				}
				// Source group headers and sparse phase labels carry context.
				// Customize one ordinary visible cell, retaining every row order,
				// group membership, phase label, score and unit. Dedicated table
				// tests exercise semantic reordering rather than a flat reversal.
				ops := []TableOperation{{Action: "materialize", Entity: "source"}}
				foundCell := false
				for _, row := range inspection.Model.Rows {
					if _, header := row.Values["group"]; header {
						continue
					}
					for _, column := range inspection.Model.Columns {
						k, _ := column["k"].(string)
						typ, _ := column["type"].(string)
						if typ == "" || typ == "text" {
							if obj, ok := row.Values[k].(map[string]any); ok {
								if value, ok := obj["text"].(string); ok && strings.TrimSpace(value) != "" && len(value) <= 64 {
									copy := map[string]any{}
									for key, value := range obj {
										copy[key] = value
									}
									copy["text"] = "Demo " + value
									ops = append(ops, TableOperation{Action: "set", Entity: "cell", Key: row.Key, Column: k, Value: copy})
									foundCell = true
									break
								}
							}
						}
						value, ok := row.Values[k].(string)
						if !ok || strings.TrimSpace(value) == "" || len(value) > 64 || typ != "" && typ != "text" || k == "ph" || k == "phase" {
							continue
						}
						ops = append(ops, TableOperation{Action: "set", Entity: "cell", Key: row.Key, Column: k, Value: "Demo " + value})
						foundCell = true
						break
					}
					if foundCell {
						break
					}
				}
				if !foundCell {
					for _, row := range inspection.Model.Rows {
						for _, column := range inspection.Model.Columns {
							if column["type"] != "status" {
								continue
							}
							k, _ := column["k"].(string)
							status, ok := row.Values[k].(string)
							if !ok || status == "" {
								continue
							}
							label := map[string]string{"risk": "At risk", "on": "On track", "off": "Off track", "done": "Done", "progress": "In progress", "notstarted": "Not started", "blocked": "Blocked", "pass": "Pass", "fail": "Fail"}[status]
							if labels, ok := column["labels"].(map[string]any); ok {
								if v, ok := labels[status].(string); ok {
									label = v
								}
							}
							if label == "" {
								continue
							}
							ops = append(ops, TableOperation{Action: "set", Entity: "cell", Key: row.Key, Column: k, Value: map[string]any{"status": status, "label": "Demo · " + label}})
							foundCell = true
							break
						}
						if foundCell {
							break
						}
					}
				}
				if !foundCell {
					t.Fatal("table has no ordinary visible text cell for meaningful source customization")
				}

				patch := TablePatch{Schema: TablePatchSchema, ExpectedSourceSHA256: selected.SourceHash(), Actor: "Catalog qualification", Reason: "Customize one illustrative caption with stable keyed cell ownership while preserving source table context", SemanticReview: "Illustrative rows retain source order, phase and group associations, scores, units and metadata; source facts are not recalculated", NodeID: nodeID, Operations: ops}
				if _, e = PatchTable(selected, "catalog-slide", patch, b, wmdesign.CandidateEngine, false); e != nil {
					t.Fatal(e)
				}
				bad := patch
				bad.ExpectedSourceSHA256 = strings.Repeat("0", 64)
				if _, e = PatchTable(selected, "catalog-slide", bad, b, wmdesign.CandidateEngine, false); e == nil {
					t.Fatal("stale source accepted")
				}
				{
					if _, e = PatchTable(selected, "catalog-slide", patch, b, wmdesign.CandidateEngine, true); e != nil {
						t.Fatal(e)
					}
					selected, e = Load(selected.SourcePath)
					if e != nil {
						t.Fatal(e)
					}
				}
			} else {
				inspection, e := InspectComponent(selected, "catalog-slide", nodeID, b, wmdesign.CandidateEngine)
				if e != nil || inspection.RenderError != "" {
					t.Fatalf("component %v %s", e, inspection.RenderError)
				}
				ops := []ComponentOperation{{Action: "materialize", Entity: "source"}}
				chosen := ""
				for _, field := range []string{"text", "title", "label", "name", "meta", "role", "value"} {
					if value, ok := inspection.Arguments[field].(string); ok && strings.TrimSpace(value) != "" {
						chosen = field
						ops = append(ops, ComponentOperation{Action: "set", Entity: "argument", Path: "/" + field, Value: func() any {
							if field == "value" {
								return value
							}
							return catalogQualificationScalar(key, nodeID, field, value)
						}()})
						break
					}
				}
				if chosen == "" {
					for _, c := range inspection.Collections {
						items, err := lookupPointer(inspection.Arguments, c.sourcePath)
						if err != nil {
							t.Fatal(err)
						}
						for i, raw := range items.([]any) {
							item, ok := raw.(map[string]any)
							if !ok {
								continue
							}
							for _, field := range []string{"title", "label", "name", "text", "p", "meta", "role"} {
								value, ok := item[field].(string)
								if !ok || strings.TrimSpace(value) == "" || len(value) > 64 {
									continue
								}
								chosen = c.Path + "/@" + c.Keys[i] + "/" + field
								ops = append(ops, ComponentOperation{Action: "set", Entity: "argument", Path: chosen, Value: "Demo " + value})
								break
							}
							if chosen != "" {
								break
							}
						}
						if chosen != "" {
							break
						}
					}
				}
				if chosen == "" {
					t.Fatal("selected component has no editable source field")
				}
				patch := ComponentPatch{Schema: ComponentPatchSchema, ExpectedSourceSHA256: selected.SourceHash(), Actor: "Catalog qualification", Reason: "Exercise typed source adapter and immutable measured preview", SemanticReview: "Illustrative source copy and interdependent facts retain existing meaning; no inferred dates/owners/totals", NodeID: nodeID, Operations: ops}
				if _, e = PatchComponent(selected, "catalog-slide", patch, b, wmdesign.CandidateEngine, false); e != nil {
					t.Fatal(e)
				}
				{
					long := patch
					long.Operations = append([]ComponentOperation{}, patch.Operations...)
					long.Operations[len(long.Operations)-1].Value = strings.Repeat("A deliberately overlong qualification label ", 500)
					if _, e = PatchComponent(selected, "catalog-slide", long, b, wmdesign.CandidateEngine, false); e == nil {
						t.Fatal("overlong source unexpectedly fits")
					}
				}
				{
					if _, e = PatchComponent(selected, "catalog-slide", patch, b, wmdesign.CandidateEngine, true); e != nil {
						t.Fatal(e)
					}
					selected, e = Load(selected.SourcePath)
					if e != nil {
						t.Fatal(e)
					}
				}
			}
			// The pinned stock maturity specimen's first gap is one stage,
			// unlike its headline. Correct this local narrative explicitly,
			// retaining all original scores, targets and parent provenance.
			if key == "maturity/ai-assessment" {
				values := map[string]any{}
				for field, value := range selected.Document.Slides[0].Values {
					values[field] = value
				}
				values["title"] = "Six capabilities need two stages; adoption needs one"
				if _, e = EditSlidesWithOptions(selected, map[string]SlideEdit{"catalog-slide": {Values: values}}, b, wmdesign.CandidateEngine, EditOptions{CheckFit: true}); e != nil {
					t.Fatal(e)
				}
				selected, e = Load(selected.SourcePath)
				if e != nil {
					t.Fatal(e)
				}
			}
			if key == "assumptions/change-control" {
				inspection, e := InspectComponent(selected, "catalog-slide", nodeID, b, wmdesign.CandidateEngine)
				if e != nil {
					t.Fatal(e)
				}
				items := inspection.Arguments["items"].([]any)
				for i, want := range []string{"Request", "Assess", "Approve", "Update"} {
					got := items[i].(map[string]any)["title"].(string)
					if strings.TrimPrefix(got, "Demo ") != want {
						t.Fatalf("workflow step%d lost chronology: %q, want %q", i, got, want)
					}
				}
			}
			if key == "maturity/ai-assessment" {
				inspection, e := InspectTable(selected, "catalog-slide", nodeID, b, wmdesign.CandidateEngine)
				if e != nil {
					t.Fatal(e)
				}
				if len(inspection.Model.Rows) != 7 || !bytes.Equal(canonical(inspection.Model.Rows[0].Values["m"]), canonical([]any{float64(2), float64(3)})) {
					t.Fatal("maturity scenario scores/targets changed")
				}
				if selected.Document.Slides[0].Values["title"] != "Six capabilities need two stages; adoption needs one" {
					t.Fatal("maturity narrative contradicts retained scenario")
				}
			}
			if key == "architecture/app-ecosystem" {
				if nodeID != "node02" || kind != "device" || len(selected.Document.LocalTemplates["catalog"].Nodes) != 30 {
					t.Fatal("architecture leaf fixture lost original full source")
				}
				for id, want := range map[string]string{"node01": "Channels", "node02": "Portal", "node22": "External systems"} {
					got, err := InspectComponent(selected, "catalog-slide", id, b, wmdesign.CandidateEngine)
					if err != nil || got.Arguments["label"] != want {
						t.Fatalf("architecture %s label %v, want %q: %v", id, got.Arguments["label"], want, err)
					}
				}
			}
			writeProcessJourneyQualification(t, selected, "catalog-"+family)
			receipt := map[string]any{"family": family, "template": key, "node": nodeID, "kind": kind, "placeholder_media": placeholderReceipts, "scope": "one full-source representative with guarded source customization; desktop QA separate"}
			combinedReceipts = append(combinedReceipts, receipt)
			id := "catalog-" + family
			combined.Document.LocalTemplates[id] = selected.Document.LocalTemplates["catalog"]
			slide := selected.Document.Slides[0]
			slide.ID = id
			slide.Template.ID = id
			combined.Document.Slides = append(combined.Document.Slides, slide)
			for assetID, asset := range selected.Document.Assets {
				if old, exists := combined.Document.Assets[assetID]; exists && !bytes.Equal(canonical(old), canonical(asset)) {
					t.Fatalf("conflicting portable asset %s", assetID)
				}
				combined.Document.Assets[assetID] = asset
				raw, e := os.ReadFile(filepath.Join(selected.Root, asset.Path))
				if e != nil {
					t.Fatal(e)
				}
				target := filepath.Join(combined.Root, asset.Path)
				if e = os.MkdirAll(filepath.Dir(target), 0700); e != nil {
					t.Fatal(e)
				}
				if e = os.WriteFile(target, raw, 0600); e != nil {
					t.Fatal(e)
				}
			}
			if root := os.Getenv("PPTXGENGO_FAMILY_CATALOG_EVIDENCE"); root != "" {
				if e = os.MkdirAll(root, 0700); e != nil {
					t.Fatal(e)
				}
				raw, _ := json.MarshalIndent(receipt, "", "  ")
				if e = os.WriteFile(root+"/"+family+".json", raw, 0600); e != nil {
					t.Fatal(e)
				}
			}
		})
	}
	if qualified != 29 || t.Failed() {
		t.Fatal(fmt.Sprintf("qualified %d families, wanted29", qualified))
	}
	if e = os.WriteFile(combined.SourcePath, canonical(combined.Document), 0600); e != nil {
		t.Fatal(e)
	}
	combined, e = Load(combined.SourcePath)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = Pin(combined, b, wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	if _, e = Check(combined, b, wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	if os.Getenv("PPTXGENGO_FAMILY_QUALIFICATION_DIR") != "" {
		receipt, e := Build(combined, BuildOptions{Bundle: b, Engine: wmdesign.CandidateEngine})
		if e != nil {
			t.Fatal(e)
		}
		t.Logf("combined full-source deck: %s/builds/%s/deck.pptx", combined.Root, receipt.BuildID)
	}
	if e = os.WriteFile(filepath.Join(combined.Root, "qualification-inputs.json"), canonical(combinedReceipts), 0600); e != nil {
		t.Fatal(e)
	}
	writeProcessJourneyQualification(t, combined, "qualification-families-29")
}

func TestComponentCatalogWorkshopPreservesDecorativeEmptyBacking(t *testing.T) {
	b := journeyBundle(t)
	source, err := ScaffoldTemplateWithOptions(b, "workshop-guide/breakout-tables", wmdesign.CandidateEngine, "Inspect actual authored heading and decorative background", 2026, ScaffoldOptions{PlaceholderMedia: true})
	if err != nil {
		t.Fatal(err)
	}
	blank, heading := false, false
	for _, node := range source.Template.Nodes {
		args, _, e := componentArgs(&node, source.SyntheticSourceValues)
		if e != nil {
			t.Fatal(e)
		}
		if node.ID == "node01" {
			blank = args["text"] == "" && node.Definition.ID == "wmds/component/block"
		}
		if node.ID == "node02" {
			heading = strings.Contains(fmt.Sprint(args["text"]), "SCHEDULING")
		}
	}
	if !blank || !heading {
		t.Fatal("actual source decorative backing/companion heading contract changed", blank, heading)
	}
}

// Fixed grouping captions share space with their companion contents. The actual
// Office v3 fixture showed that a locally fitting "Demo Channels" caption could
// wrap into its first device icon. Exercise a leaf instead of blindly lengthening
// every fixed container/frame header; this is fixture selection, not inference.
func catalogQualificationLeafKind(kind string) bool {
	return kind != "container" && kind != "frame"
}

func catalogQualificationScalar(key, nodeID, field, value string) string {
	if key == "architecture/app-ecosystem" && nodeID == "node02" && field == "label" {
		return "Portal" // Explicit illustrative customization; grouping stays Channels.
	}
	return "Demo " + value
}

func TestComponentCatalogArchitectureCustomizesLeafNotContainerHeader(t *testing.T) {
	b := journeyBundle(t)
	source, err := ScaffoldTemplateWithOptions(b, "architecture/app-ecosystem", wmdesign.CandidateEngine, "Reviewed leaf customization retains grouping and source geometry", 2026, ScaffoldOptions{PlaceholderMedia: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(source.Template.Nodes) != 30 {
		t.Fatal("original full architecture source changed")
	}
	selected := ""
	for _, node := range source.Template.Nodes {
		if node.Definition == nil {
			continue
		}
		kind := strings.TrimPrefix(node.Definition.ID, "wmds/component/")
		args, _, e := componentArgs(&node, source.SyntheticSourceValues)
		if e != nil {
			t.Fatal(e)
		}
		if node.ID == "node01" && (kind != "container" || args["label"] != "Channels" || catalogQualificationLeafKind(kind)) {
			t.Fatal("fixed grouping caption must remain original Channels", args)
		}
		if node.ID == "node22" && (kind != "frame" || args["label"] != "External systems" || catalogQualificationLeafKind(kind)) {
			t.Fatal("fixed external grouping frame must remain original", args)
		}
		if selected == "" && catalogQualificationLeafKind(kind) {
			if value, ok := args["label"].(string); ok && strings.TrimSpace(value) != "" {
				selected = node.ID
				if kind != "device" || value != "Web app" || catalogQualificationScalar("architecture/app-ecosystem", node.ID, "label", value) != "Portal" {
					t.Fatal("architecture representative must customize the first device leaf", node.ID, kind, args)
				}
			}
		}
	}
	if selected != "node02" {
		t.Fatal("expected node02 leaf", selected)
	}
	if catalogQualificationScalar("architecture/app-ecosystem", "node03", "label", "Mobile app") != "Demo Mobile app" {
		t.Fatal("explicit fixture mapping escaped its reviewed leaf")
	}
}
