package wmdesign

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestV5DeltaSourceQualification(t *testing.T) {
	bundle := filepath.Join("..", "..", "library", "wm-design-system", "v5")
	source, err := Load(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	entries := intakeRepairEntries(t, filepath.Join(bundle, "source", "templates", "library"))
	keys := make([]string, 0, len(v5DeltaNodeCounts))
	for key := range v5DeltaNodeCounts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	type result struct {
		Key   string `json:"key"`
		Error string `json:"error,omitempty"`
	}
	var results []result
	for _, key := range keys {
		t.Run(key, func(t *testing.T) {
			slide := intakeRepairSlide(t, entries[key])
			err := applyLibraryRefinements(key, LibraryRevisionV5, &slide)
			if err == nil {
				_, _, err = buildWithLoadedSource(bundle, source, Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{slide}}, CandidateEngine, nil)
			}
			row := result{Key: key}
			if err != nil {
				row.Error = err.Error()
				t.Error(err)
			}
			results = append(results, row)
		})
	}
	if path := os.Getenv("WMDS_V5_DELTA_RESULTS"); path != "" {
		b, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(path, append(b, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if len(keys) != 68 {
		t.Fatalf("delta templates %d", len(keys))
	}
}

func TestV5DeltaRepairsAtomicIdempotentAndCopyPreserving(t *testing.T) {
	entries := intakeRepairEntries(t, filepath.Join("..", "..", "library", "wm-design-system", "v5", "source", "templates", "library"))
	for key := range v5DeltaNodeCounts {
		t.Run(key, func(t *testing.T) {
			original := intakeRepairSlide(t, entries[key])
			slide := intakeRepairSlide(t, entries[key])
			if err := ApplyIncomingV5Repairs(key, LibraryRevisionV5, &slide); err != nil {
				t.Fatal(err)
			}
			first, _ := json.Marshal(slide)
			if err := ApplyIncomingV5Repairs(key, LibraryRevisionV5, &slide); err != nil {
				t.Fatal(err)
			}
			second, _ := json.Marshal(slide)
			if !bytes.Equal(first, second) {
				t.Fatal("repair is cumulative")
			}
			if slide.Title != original.Title || slide.Source != original.Source || slide.Eyebrow != original.Eyebrow {
				t.Fatal("chrome copy changed")
			}
			if key == "value-curve/break-even-split" {
				obj, err := libraryObject(slide.Nodes[1].Scene.Node)
				if err != nil {
					t.Fatal(err)
				}
				point := obj["points"].([]any)[0].([]any)
				y := intakeRepairNumber(map[string]any{"y": point[1]}, "y")
				if math.Abs((y-84)/(414-84)-(315.7-84)/(432-84)) > 1e-12 {
					t.Fatal("break-even dot lost its original chart ordinate")
				}
			}
			for i, n := range original.Nodes {
				before := v5RepairSemanticNode(t, key, n.Scene.Node)
				after := v5RepairSemanticNode(t, key, slide.Nodes[i].Scene.Node)
				if !bytes.Equal(before, after) {
					t.Fatalf("copy, data or typography changed in%s", n.ID)
				}
			}
			bad := intakeRepairSlide(t, entries[key])
			bad.Nodes[0].Scene = nil
			v5RepairRequireAtomicReject(t, key, &bad)
			bad = intakeRepairSlide(t, entries[key])
			bad.Nodes[0].ID = "wrong"
			v5RepairRequireAtomicReject(t, key, &bad)
			bad = intakeRepairSlide(t, entries[key])
			bad.Nodes = bad.Nodes[1:]
			v5RepairRequireAtomicReject(t, key, &bad)
		})
	}
	for _, key := range []string{"interviews-planned/waves", "interviews-access/escalation-split"} {
		slide := intakeRepairSlide(t, entries[key])
		slide.Frame.TitleLines = 3
		v5RepairRequireAtomicReject(t, key, &slide)
	}
	slide := intakeRepairSlide(t, entries["offers-sku/catalog-split"])
	obj, err := libraryObject(slide.Nodes[0].Scene.Node)
	if err != nil {
		t.Fatal(err)
	}
	obj["rowH"] = 17
	slide.Nodes[0].Scene.Node, _ = json.Marshal(obj)
	v5RepairRequireAtomicReject(t, "offers-sku/catalog-split", &slide)
	for _, revision := range []string{LibraryRevisionV3, LibraryRevisionV4} {
		slide := intakeRepairSlide(t, entries["understanding-statement/inverse"])
		before, _ := json.Marshal(slide)
		if err := ApplyIncomingV5Repairs("understanding-statement/inverse", revision, &slide); err != nil {
			t.Fatal(err)
		}
		after, _ := json.Marshal(slide)
		if !bytes.Equal(before, after) {
			t.Fatal("historical revision changed")
		}
	}
}

func v5RepairRequireAtomicReject(t *testing.T, key string, slide *SlideSpec) {
	t.Helper()
	before, _ := json.Marshal(slide)
	if err := ApplyIncomingV5Repairs(key, LibraryRevisionV5, slide); err == nil {
		t.Fatal("unexpected source topology accepted")
	}
	after, _ := json.Marshal(slide)
	if !bytes.Equal(before, after) {
		t.Fatal("rejected repair partially mutated input")
	}
}

// Exclude only the named geometry fields and the two explicitly documented
// invalid/unused source fields. Every copy leaf, style token and chart/table
// datum must remain byte-equivalent after canonical JSON serialization.
func v5RepairSemanticNode(t *testing.T, key string, raw json.RawMessage) []byte {
	t.Helper()
	obj, err := libraryObject(raw)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"x", "y", "w", "h", "rowH"} {
		delete(obj, field)
	}
	if obj["type"] == "connector" {
		delete(obj, "points")
	}
	if key == "interviews-access/escalation-split" {
		delete(obj, "edge")
	}
	if key == "interviews-readout/full" && obj["type"] == "table" {
		for _, raw := range obj["rows"].([]any) {
			delete(raw.(map[string]any), "p")
		}
	}
	b, err := json.Marshal(obj)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
