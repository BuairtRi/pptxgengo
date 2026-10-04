package wmdesign

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// TestIntakeManualSourceSmoke is an opt-in diagnostic, not a claim that every
// observed template fits. A successful run writes each rejected source and its
// exact failure; only compiled slides enter the unreleased prototype deck.
func TestIntakeManualSourceSmoke(t *testing.T) {
	out := os.Getenv("WMDS_INTAKE_SMOKE_OUT")
	if out == "" {
		t.Skip("manual diagnostic output only")
	}
	if err := os.Mkdir(out, 0755); err != nil {
		t.Fatal(err)
	}
	snapshot := filepath.Join("..", "..", "planning", "wm-design-contracts", "v4", "intake-20261003", "source", "templates", "library")
	if selected := os.Getenv("WMDS_INTAKE_SNAPSHOT"); selected != "" {
		snapshot = filepath.Join(selected, "source", "templates", "library")
	}
	bundle := filepath.Join("..", "..", "library", "wm-design-system", "v5")
	source, err := Load(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	round12 := os.Getenv("WMDS_INTAKE_ROUND12") == "1"
	retained := 0
	baseline := map[string]libraryEntry{}
	if os.Getenv("WMDS_INTAKE_RETAIN_V3") == "1" {
		// Preserve the historical diagnostic flag using qualified lineage
		// evidence, rather than treating every current v5 source as v3 input.
		entries := intakeRepairEntries(t, filepath.Join("..", "..", "planning", "wm-design-contracts", "v4", "intake-20261003-frozen", "source", "templates", "library"))
		for key, entry := range entries {
			if v4RetainedV3Compositions[key] {
				baseline[key] = entry
			}
		}
	}
	if round12 {
		source.Revision = IntakeTeamCurveMonotoneRevision
		observationRoot := filepath.Clean(filepath.Join(snapshot, "..", "..", ".."))
		data, err := os.ReadFile(filepath.Join(observationRoot, "observation.json"))
		if err != nil {
			t.Fatal(err)
		}
		var observation struct {
			Head  string       `json:"head"`
			Files []SourceFile `json:"files"`
		}
		if err = json.Unmarshal(data, &observation); err != nil {
			t.Fatal(err)
		}
		if len(observation.Files) == 0 || len(observation.Head) != 40 {
			t.Fatal("invalid intake observation")
		}
		for _, file := range observation.Files {
			if filepath.IsAbs(file.Path) || strings.Contains(file.Path, "..") {
				t.Fatal("unsafe observation path")
			}
			b, err := os.ReadFile(filepath.Join(observationRoot, "source", file.Path))
			if err != nil {
				t.Fatal(err)
			}
			if fmt.Sprintf("%x", sha256.Sum256(b)) != file.SHA256 {
				t.Fatalf("observation drift: %s", file.Path)
			}
		}
		source.Commit = observation.Head
		data, err = os.ReadFile(filepath.Join(snapshot, "..", "..", "frames", "v0", "frames.json"))
		if err != nil {
			t.Fatal(err)
		}
		if err = json.Unmarshal(data, &source.Frames); err != nil {
			t.Fatal(err)
		}
	}
	type result struct {
		Key    string `json:"key"`
		Status string `json:"status"`
		Error  string `json:"error,omitempty"`
		Slide  int    `json:"slide,omitempty"`
	}
	var results []result
	doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Title: "WMDS incoming capability prototypes", BuildIdentity: &BuildIdentity{Timestamp: "2000-01-01T00:00:00Z", Seed: "intake-20261003-capabilities"}}
	families := []string{"venn", "maturity", "heatmaps", "argument", "approach", "team-curves"}
	all := os.Getenv("WMDS_INTAKE_ALL") == "1"
	selected := map[string]bool{}
	if list := os.Getenv("WMDS_INTAKE_KEYS"); list != "" {
		for _, key := range strings.Split(list, ",") {
			if key == "" || selected[key] {
				t.Fatal("empty or duplicate selected intake key")
			}
			selected[key] = true
		}
	}
	seen := map[string]bool{}
	if all {
		families = nil
		paths, err := filepath.Glob(filepath.Join(snapshot, "*.json"))
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range paths {
			families = append(families, strings.TrimSuffix(filepath.Base(path), ".json"))
		}
	}
	for _, family := range families {
		data, err := os.ReadFile(filepath.Join(snapshot, family+".json"))
		if err != nil {
			t.Fatal(err)
		}
		var f struct {
			Templates []libraryEntry `json:"templates"`
		}
		if err = json.Unmarshal(data, &f); err != nil {
			t.Fatal(err)
		}
		for _, entry := range f.Templates {
			key := entry.ID + "/" + entry.Variant
			if len(selected) > 0 && !selected[key] {
				continue
			}
			seen[key] = true
			if !all && (family == "argument" && entry.ID != "vendors" || family == "approach" && (entry.ID != "roles" || !strings.HasPrefix(entry.Variant, "dot-scores"))) {
				continue
			}
			obj, e := libraryObject(entry.Slide)
			if e != nil {
				t.Fatal(e)
			}
			for _, n := range obj["body"].([]any) {
				if node, ok := n.(map[string]any); !round12 && ok && node["type"] == "teamcurve" {
					node["curve"] = "monotone"
				}
			}
			entry.Slide, _ = json.Marshal(obj)
			def := LibraryTemplate{SourceRevision: source.Revision, RawSlide: entry.Slide}
			for i, n := range obj["body"].([]any) {
				libraryContentWalk(&def, n, fmt.Sprintf("/body/%d", i), fmt.Sprintf("node%02d", i+1), "", libraryProjectionContext{})
			}
			keys := map[string][]string{}
			for _, a := range def.Arrays {
				for i := 0; i < a.Count; i++ {
					keys[a.SourcePointer] = append(keys[a.SourcePointer], fmt.Sprintf("item-%03d", i+1))
				}
			}
			slide, e := compileLibrarySlide(entry.Slide, keys)
			if e == nil {
				slide.ID = "intake-" + strings.ReplaceAll(key, "/", "-")
				if old, ok := baseline[key]; ok {
					original, err := libraryObject(old.Slide)
					if err != nil {
						t.Fatal(err)
					}
					if reflect.DeepEqual(original, obj) {
						e = applyLibraryRefinements(key, LibraryRevisionV3, &slide)
						retained++
					}
				}
				if os.Getenv("WMDS_INTAKE_REPAIRS") == "1" {
					e = ApplyIncomingIntakeRepairs(key, IntakeRepairRevision, &slide)
				}
				if e == nil {
					_, _, e = buildWithLoadedSource(bundle, source, Document{Schema: doc.Schema, Year: doc.Year, Slides: []SlideSpec{slide}}, CandidateEngine, nil)
				}
			}
			if e != nil {
				results = append(results, result{Key: key, Status: "rejected", Error: e.Error()})
				t.Logf("REJECT %s: %v", key, e)
				continue
			}
			doc.Slides = append(doc.Slides, slide)
			results = append(results, result{Key: key, Status: "compiled", Slide: len(doc.Slides)})
		}
	}
	for key := range selected {
		if !seen[key] {
			t.Fatalf("selected intake key not found: %s", key)
		}
	}
	write := func(name string, v any) {
		b, e := json.MarshalIndent(v, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(out, name), append(b, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	}
	write("smoke-results.json", results)
	write("foundation.json", doc)
	if len(doc.Slides) == 0 {
		t.Fatal("no capability sources compiled")
	}
	data, report, err := buildWithLoadedSource(bundle, source, doc, CandidateEngine, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(out, "capability-reference.pptx"), data, 0644); err != nil {
		t.Fatal(err)
	}
	write("layout-report.json", report)
	write("receipt.json", map[string]any{"attempted": len(results), "compiled": len(doc.Slides), "rejected": len(results) - len(doc.Slides), "pptx_sha256": fmt.Sprintf("%x", sha256.Sum256(data)), "bytes": len(data), "source_snapshot": snapshot, "native_review": "pending", "status": "unreleased_capability_prototype", "round12_renderer": round12, "all_templates": all && len(selected) == 0, "selected_key_count": len(selected), "incoming_repairs": os.Getenv("WMDS_INTAKE_REPAIRS") == "1", "unchanged_v3_specimens_retaining_qualified_amendments": retained, "assets_and_fonts": "pinned v5; historical source overlays and optional Round12 frame/renderer semantics are an internal diagnostic only"})
	t.Logf("Attempted%d compiled%d rejected%d, output%s", len(results), len(doc.Slides), len(results)-len(doc.Slides), out)
}
