package wmdesign

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLibraryV11ChangedTemplatesRoundTripAndBuild(t *testing.T) {
	bundle := v11IntakeBundle()
	keys := v11TemplateKeys(t, "changed", 231)
	source, err := LibrarySourceReference(bundle, "", "", 2026)
	if err != nil {
		t.Fatal(err)
	}
	input, err := LibraryReference(bundle, "", "", 2026)
	if err != nil {
		t.Fatal(err)
	}
	bySource := map[string]SlideSpec{}
	for _, slide := range source.Slides {
		bySource[slide.TemplateBinding.Template] = slide
	}
	byBound := map[string]BoundSlide{}
	for _, slide := range input.Slides {
		byBound[slide.Template] = slide
	}
	source.Slides, input.Slides = nil, nil
	for _, key := range keys {
		left, ok1 := bySource[key]
		right, ok2 := byBound[key]
		if !ok1 || !ok2 {
			t.Fatalf("changed key absent from source/bound references: %s", key)
		}
		source.Slides = append(source.Slides, left)
		input.Slides = append(input.Slides, right)
	}
	bound, _, err := BindTemplates(bundle, "", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(bound.Slides) != len(keys) {
		t.Fatal("changed binding cardinality drift")
	}
	for i := range source.Slides {
		if !reflect.DeepEqual(v6ComparisonSlide(t, source.Slides[i], true), v6ComparisonSlide(t, bound.Slides[i], true)) {
			t.Fatalf("source/bound composition differs: %s", keys[i])
		}
	}
	// The source/binding assertions above remain hermetic. Rendering the
	// source specimens needs original private artwork in the integration lane.
	if testing.Short() {
		return
	}
	for _, entry := range []struct {
		name string
		doc  Document
	}{{"source", source}, {"bound", bound}} {
		t.Run(entry.name, func(t *testing.T) {
			deck, report, err := BuildWithEngine(bundle, "", entry.doc, CandidateEngine)
			if err != nil {
				t.Fatal(err)
			}
			if len(deck) == 0 || len(report.Slides) != len(keys) || report.PowerPointVerified || report.VisuallyReviewed {
				t.Fatal("invalid generation count or native acceptance claim")
			}
		})
	}
}

func TestLibraryV11AllTemplateBuilds(t *testing.T) {
	if testing.Short() {
		t.Skip("exhaustive 649 source and 648 active bound generation")
	}
	bundle := v11IntakeBundle()
	source, err := LibrarySourceReference(bundle, "", "", 2026)
	if err != nil {
		t.Fatal(err)
	}
	input, err := LibraryReference(bundle, "", "", 2026)
	if err != nil {
		t.Fatal(err)
	}
	bound, _, err := BindTemplates(bundle, "", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(source.Slides) != 649 || len(bound.Slides) != 648 {
		t.Fatal("full catalog cardinality drift")
	}
	for _, entry := range []struct {
		name string
		doc  Document
	}{{"source", source}, {"bound", bound}} {
		t.Run(entry.name, func(t *testing.T) {
			deck, report, err := BuildWithEngine(bundle, "", entry.doc, CandidateEngine)
			if err != nil {
				t.Fatal(err)
			}
			if len(deck) == 0 || len(report.Slides) != len(entry.doc.Slides) || report.PowerPointVerified || report.VisuallyReviewed {
				t.Fatal("invalid generation count or native acceptance claim")
			}
		})
	}
}

func TestLibraryV11UnchangedPublicationRenderInheritance(t *testing.T) {
	if testing.Short() {
		t.Skip("paired visible dependency comparison of 297 render-inheritance specimens")
	}
	keys := v11TemplateKeys(t, "inherit-eligible", 297)
	count, err := VerifyLibraryPublicationRenderInheritance(LibraryPublicationOptions{
		Bundle: v11IntakeBundle(), PreviousBundle: v10IntakeBundle(), Year: 2026,
	}, keys)
	if err != nil {
		t.Fatal(err)
	}
	if count != 297 {
		t.Fatalf("inheritance cardinality: %d", count)
	}
}

func TestLibraryV11ChangedIndividualSourceBuilds(t *testing.T) {
	if testing.Short() {
		t.Skip("individual generation of all 231 changed specimens")
	}
	source, err := LibrarySourceReference(v11IntakeBundle(), "", "", 2026)
	if err != nil {
		t.Fatal(err)
	}
	byKey := map[string]SlideSpec{}
	for _, slide := range source.Slides {
		byKey[slide.TemplateBinding.Template] = slide
	}
	type result struct {
		Key         string                   `json:"key"`
		Error       string                   `json:"error,omitempty"`
		Density     *SlideDensityRecord      `json:"density,omitempty"`
		Adjustments []SlideDensityAdjustment `json:"adjustments,omitempty"`
	}
	var results []result
	for _, key := range v11TemplateKeys(t, "changed", 231) {
		t.Run(key, func(t *testing.T) {
			slide, ok := byKey[key]
			if !ok {
				t.Fatal("audited key absent")
			}
			doc := source
			doc.Slides = []SlideSpec{slide}
			_, report, err := BuildWithEngine(v11IntakeBundle(), "", doc, CandidateEngine)
			entry := result{Key: key, Adjustments: report.DensityAdjustments}
			if len(report.Slides) == 1 {
				entry.Density = report.Slides[0].Density
			}
			if err != nil {
				entry.Error = err.Error()
			}
			results = append(results, entry)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
	dir, err := os.MkdirTemp("/private/tmp", "wmds-v11-changed-individual-build-")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "audit.json"), append(raw, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	t.Logf("complete individual source fit audit: %s", dir)
}
