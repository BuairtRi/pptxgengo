package wmdesign

import (
	"path/filepath"
	"strings"
	"testing"
)

func densityBundle(t *testing.T) string {
	t.Helper()
	return filepath.Join("..", "..", "planning", "wm-design-contracts", "v11", "intake-20261006-649-frozen", "bundle")
}

func TestAuthoringCapacityUsesTemplatePreferredDensity(t *testing.T) {
	source, err := Load(densityBundle(t), "")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := LibraryCatalogFromSource(source)
	if err != nil {
		t.Fatal(err)
	}
	var definition LibraryTemplate
	for _, candidate := range catalog {
		if candidate.Key == "offers-onepager/classic" {
			definition = candidate
			break
		}
	}
	if definition.Key == "" {
		t.Fatal("compact offers one-pager not found")
	}
	if definition.Authoring == nil {
		t.Fatal("template authoring metadata missing")
	}
	var body LibrarySlotCapacity
	for _, slot := range definition.Authoring.Slots {
		if slot.Capacity.Style != nil && slot.Capacity.Style.ID == "body" && slot.SourcePointer != "/title" && slot.SourcePointer != "/eyebrow" {
			body = slot.Capacity
			break
		}
	}
	if body.Style == nil || body.Density != "compact" || body.Style.Size != 12 || body.Style.Leading != 17 {
		preferred := preferredTemplateDensity(definition, source)
		t.Fatalf("compact template capacity retained stale style: %s preferred=%+v capacity=%+v", definition.Key, preferred, body)
	}
	joined := strings.Join(body.Assumptions, " ")
	if !strings.Contains(joined, "preferred authored body density") || strings.Contains(joined, "automatic fit adjustment is included") {
		t.Fatalf("capacity must describe static preferred-tier estimate: %v", body.Assumptions)
	}
}

func TestRestyleCapacityUsesRequestedCellTierAndLineBudget(t *testing.T) {
	source, err := Load(densityBundle(t), "")
	if err != nil {
		t.Fatal(err)
	}
	typography, err := NewSourceTypographyEngine(source, filepath.Join(densityBundle(t), "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	base, err := source.Style("small")
	if err != nil {
		t.Fatal(err)
	}
	capacity := LibrarySlotCapacity{Status: "estimated_geometry", Basis: "pinned_plain_native_table_cell_plan", WidthPt: 198, HeightPt: 56, LineBudget: 2, Style: &base, NativeFit: "not_evaluated"}
	updated, err := RestyleLibrarySlotCapacityAtDensity(source, typography, capacity, "dense", "cell", 0)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Style == nil || updated.Style.Size != 8 || updated.Style.Leading != 11 || updated.Density != "dense" || updated.LineBudget <= capacity.LineBudget {
		t.Fatalf("cell capacity did not use dense cell-small role/line budget: before=%+v after=%+v", capacity, updated)
	}
}

func TestHistoricalCapacityDoesNotExposeDensityMetadata(t *testing.T) {
	bundle := filepath.Join("..", "..", "planning", "wm-design-contracts", "v10", "intake-20261006-649-frozen", "bundle")
	source, err := Load(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := LibraryCatalogFromSource(source)
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range catalog {
		if candidate.Key != "offers-onepager/classic" {
			continue
		}
		for _, slot := range candidate.Authoring.Slots {
			if slot.Capacity.Density != "" {
				t.Fatalf("historical V10 capacity unexpectedly records density: %+v", slot.Capacity)
			}
		}
		return
	}
	t.Fatal("V10 offers one-pager not found")
}
