package wmdesign

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestAuthoringMetadataAll587PinnedTemplates(t *testing.T) {
	if testing.Short() {
		t.Skip("exhaustive metadata and alias sweep for all 587 templates; run make test-integration")
	}
	bundle := filepath.Join("..", "..", "planning", "wm-design-contracts", "v5", "intake-20261003-587-frozen", "bundle")
	catalog, e := LibraryCatalog(bundle, "")
	if e != nil {
		t.Fatal(e)
	}
	coverage, e := AuthoringCoverage(catalog)
	if e != nil {
		t.Fatal(e)
	}
	if coverage.Templates != 587 || len(coverage.GenericTemplates) != 0 || coverage.ReviewedTemplates != 0 || coverage.InferredTemplates != 587 || coverage.DecorativeSlots == 0 {
		t.Fatalf("dishonest or incomplete coverage: %+v", coverage)
	}
	for _, def := range catalog {
		a, e := LibraryAuthoringMetadata(def)
		if e != nil {
			t.Fatal(e)
		}
		if a.SourceSHA256 != def.SourceSHA256 || a.SourceRevision != def.SourceRevision {
			t.Fatalf("unpinned metadata %s", def.Key)
		}
		seen := map[string]bool{}
		for _, s := range a.Slots {
			if s.Alias == "" || s.Description == "" || s.Role == "" || s.ClassificationBasis == "" || seen[s.Alias] {
				t.Fatalf("incomplete/duplicate semantic slot: %s %+v", def.Key, s)
			}
			seen[s.Alias] = true
			if s.Capacity.NativeFit != "not_evaluated" {
				t.Fatal("native fit invented")
			}
		}
	}
}
func TestLifecycleAuthoringRecipe(t *testing.T) {
	catalog, e := LibraryCatalog(filepath.Join("..", "..", "planning", "wm-design-contracts", "v5", "intake-20261003-587-frozen", "bundle"), "")
	if e != nil {
		t.Fatal(e)
	}
	for _, def := range catalog {
		if def.Key != "lifecycle/three-phases" {
			continue
		}
		a, e := LibraryAuthoringMetadata(def)
		if e != nil {
			t.Fatal(e)
		}
		titles, bodies, decorative := 0, 0, 0
		for _, s := range a.Slots {
			if s.Classification == "decorative" {
				decorative++
				continue
			}
			if strings.HasPrefix(s.Alias, "/phases/") {
				if strings.HasSuffix(s.Alias, "/title") {
					titles++
				}
				if s.Role == "body" {
					bodies++
				}
				if s.GroupIndex < 0 || s.GroupIndex > 2 || s.Cardinality != 3 || s.ReviewStatus != "explicit_source_recipe_unreviewed" {
					t.Fatalf("bad lifecycle recipe: %+v", s)
				}
			}
		}
		if a.Relationship != "sequence" || titles != 15 || bodies != 15 || decorative != 15 {
			t.Fatalf("unexpected recipe %d titles %d bodies %d decorative", titles, bodies, decorative)
		}
		return
	}
	t.Fatal("missing lifecycle source")
}
func TestSlotCapacityShapedEstimateAndOverrun(t *testing.T) {
	bundle := filepath.Join("..", "..", "planning", "wm-design-contracts", "v5", "intake-20261003-587-frozen", "bundle")
	source, e := Load(bundle, "")
	if e != nil {
		t.Fatal(e)
	}
	typography, e := NewTypography(filepath.Join(bundle, "fonts"))
	if e != nil {
		t.Fatal(e)
	}
	style, e := source.Style("body")
	if e != nil {
		t.Fatal(e)
	}
	slot := LibraryAuthoringSlot{Alias: "/body", Capacity: LibrarySlotCapacity{Status: "estimated_geometry", Style: &style, WidthPt: 270, HeightPt: 54, LineBudget: 2, NativeFit: "not_evaluated", Assumptions: []string{"Fixed point size."}}}
	capacity, e := EstimateLibrarySlotCapacity(typography, slot.Capacity)
	if e != nil {
		t.Fatal(e)
	}
	if capacity.ApproxCharacters < 10 || capacity.Status != "estimated_go_shaping" {
		t.Fatalf("bad probe: %+v", capacity)
	}
	short, e := MeasureLibrarySlot(typography, slot, "A clear point.")
	if e != nil {
		t.Fatal(e)
	}
	long, e := MeasureLibrarySlot(typography, slot, strings.Repeat("An observation requires careful review. ", 15))
	if e != nil {
		t.Fatal(e)
	}
	if short.Status != "fits_estimate" || long.Status != "overflow_estimate" || long.LineOverrun <= 0 || long.HeightOverrunPt <= 0 || long.NativeFit != "not_evaluated" {
		t.Fatalf("incorrect fits: %+v %+v", short, long)
	}
	unknown, e := MeasureLibrarySlot(typography, LibraryAuthoringSlot{Capacity: unknownSlotCapacity("not_modeled")}, "Copy")
	if e != nil || unknown.Status != "unsupported" {
		t.Fatalf("unsupported capacity claimed fit: %+v %v", unknown, e)
	}
	rich, e := MeasureLibrarySlot(typography, slot, "[[Important]] claim")
	if e != nil || rich.Status != "unsupported" {
		t.Fatal("markup incorrectly measured as ordinary text")
	}
}

func TestSelectionHydratesPreAuthoringProjection(t *testing.T) {
	bundle := filepath.Join("..", "..", "planning", "wm-design-contracts", "v5", "intake-20261003-587-frozen", "bundle")
	path, report := indexFixture(t)
	// Model an older projection explicitly; a maintained production index
	// already contains authoring metadata and would bypass this fallback.
	fixture, e := indexDB(path, false)
	if e != nil {
		t.Fatal(e)
	}
	defer fixture.Close()
	if _, e = fixture.Exec("UPDATE entities SET json=json_remove(json, '$.template.authoring') WHERE key=?", "lifecycle/three-phases"); e != nil {
		t.Fatal(e)
	}
	entities, e := (&LibraryIndex{db: fixture}).entities("", nil)
	if e != nil {
		t.Fatal(e)
	}
	report.ProjectionSHA256 = projectionHash(entities)
	// A pre-authoring index predates the versioned retrieval-text projection.
	// Keep the legacy fixture honest instead of retaining new alias/FTS pins.
	report.RetrievalText = ""
	report.RetrievalSHA256 = ""
	if _, e = fixture.Exec("UPDATE meta SET value=? WHERE key='report'", string(indexJSON(report))); e != nil {
		t.Fatal(e)
	}
	index, e := OpenLibraryIndex(path, LibraryIndexOptions{Bundle: bundle})
	if e != nil {
		t.Fatal(e)
	}
	defer index.Close()
	entity, e := index.Inspect("lifecycle/three-phases")
	if e != nil || entity.Template == nil || entity.Template.Authoring != nil {
		t.Fatalf("pre-authoring fixture not established: %v", e)
	}
	card, e := index.SelectionCard("lifecycle/three-phases")
	if e != nil {
		t.Fatal(e)
	}
	if card.Authoring == nil || len(card.Authoring.Slots) == 0 || card.Authoring.SourceSHA256 == "" {
		t.Fatal("old projection not hydrated")
	}
	for _, s := range card.Authoring.Slots {
		if s.Role == "headline" {
			if s.Capacity.ApproxCharacters <= 0 || s.Capacity.Style == nil {
				t.Fatalf("missing pinned capacity: %+v", s)
			}
			return
		}
	}
	t.Fatal("missing headline")
}

func TestPrimitiveSemanticRolesPrecedeGenericTextRole(t *testing.T) {
	for style, want := range map[string]string{"subhead": "lead", "lead": "lead", "eyebrow": "label", "label": "label", "body": "body"} {
		if got := authoringRole("text", map[string]any{"type": "text", "style": style}); got != want {
			t.Fatalf("%s role %s, want %s", style, got, want)
		}
	}
}
func TestNamedHighUseRecipesRetainReviewGaps(t *testing.T) {
	catalog, e := LibraryCatalog(filepath.Join("..", "..", "planning", "wm-design-contracts", "v5", "intake-20261003-587-frozen", "bundle"), "")
	if e != nil {
		t.Fatal(e)
	}
	coverage, e := AuthoringCoverage(catalog)
	if e != nil {
		t.Fatal(e)
	}
	if coverage.RecipeTemplates == 0 || len(coverage.AmbiguousTemplates) == 0 || coverage.ReviewedTemplates != 0 {
		t.Fatalf("semantic uncertainty concealed: %+v", coverage)
	}
	t.Logf("templates=%d slots=%d recipes=%d ambiguous=%d reviewed=%d generic_names=%d", coverage.Templates, coverage.Slots, coverage.RecipeTemplates, len(coverage.AmbiguousTemplates), coverage.ReviewedTemplates, len(coverage.GenericTemplates))
	for _, def := range catalog {
		if def.Key != "interviews-readout/full" {
			continue
		}
		a, e := LibraryAuthoringMetadata(def)
		if e != nil {
			t.Fatal(e)
		}
		for _, s := range a.Slots {
			if s.SourcePointer == "/body/0/rows/0/n" {
				if s.Alias != "/interviewees/item_01/person" || s.ReviewStatus != "explicit_source_recipe_unreviewed" {
					t.Fatalf("opaque/unreviewed interview alias: %+v", s)
				}
				return
			}
		}
		t.Fatal("missing drawn person field")
	}
	t.Fatal("missing source template")
}

func TestHighUseStatementStatusAndInterviewTopology(t *testing.T) {
	catalog, err := LibraryCatalog(filepath.Join("..", "..", "planning", "wm-design-contracts", "v5", "intake-20261003-587-frozen", "bundle"), "")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]map[string]string{
		"key-message/statement":   {"/body/1/text": "/headline", "/body/0/text": "/eyebrow", "/body/2/text": "/supporting_statement"},
		"status/classic":          {"/body/6/rows/0/own": "/status/risks/item_01/owner", "/body/6/rows/0/li": "/status/risks/item_01/likelihood", "/body/6/rows/0/mit": "/status/risks/item_01/mitigation", "/body/6/rows/0/st": "/status/risks/item_01/state"},
		"interviews-detail/dense": {"/body/1/text": "/interview/stakeholder_name", "/body/5/text": "/interview/date_and_duration", "/body/11/title": "/topics/item_01/title", "/body/20/items/0": "/pain_and_opportunities/items/item_01"},
	}
	for _, def := range catalog {
		fields, ok := want[def.Key]
		if !ok {
			continue
		}
		metadata, err := LibraryAuthoringMetadata(def)
		if err != nil {
			t.Fatal(err)
		}
		if metadata.SemanticStatus != "engineering_recipe_requires_review" {
			t.Fatal("recipe claimed review", def.Key)
		}
		for _, slot := range metadata.Slots {
			if alias, exists := fields[slot.SourcePointer]; exists {
				if slot.Alias != alias || slot.ReviewStatus != "explicit_source_recipe_unreviewed" {
					t.Fatalf("%s %s: %+v", def.Key, alias, slot)
				}
				delete(fields, slot.SourcePointer)
			}
		}
		if len(fields) != 0 {
			t.Fatal("expected slots missing", def.Key, fields)
		}
		delete(want, def.Key)
	}
	if len(want) != 0 {
		t.Fatal("expected templates missing")
	}
}
