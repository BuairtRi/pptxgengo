package wmdesign

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func v12Bundle() string {
	return filepath.Join("..", "..", "planning", "wm-design-contracts", "v12", "intake-20261008-677-frozen", "bundle")
}

func v12AddedKeys(t *testing.T) []string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "planning", "wm-design-contracts", "v12", "intake-20261008-677-frozen", "new-template-keys.json"))
	if err != nil {
		t.Fatal(err)
	}
	var keys []string
	if err := json.Unmarshal(data, &keys); err != nil {
		t.Fatal(err)
	}
	if len(keys) != 28 {
		t.Fatal("added identity count changed")
	}
	return keys
}

// Verify exact source inheritance as well as executable bindings: family files
// change their hashes when additions are appended, even for retained entries.
func TestLibraryV12PinnedCatalogAndInheritance(t *testing.T) {
	source, err := Load(v12Bundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	if source.Revision != LibraryRevisionV12 || source.Commit != "efec671fe40d2145d14780dc39c3128bc9c65308" {
		t.Fatal("source identity changed")
	}
	current, err := LibraryCatalog(v12Bundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	previous, err := LibraryCatalog(filepath.Join("..", "..", "library", "wm-design-system", "v11"), "")
	if err != nil {
		t.Fatal(err)
	}
	old := map[string]LibraryTemplate{}
	for _, def := range previous {
		old[def.Key] = def
	}
	added := map[string]bool{}
	for _, key := range v12AddedKeys(t) {
		added[key] = true
	}
	counts := map[string]int{}
	retained, active := 0, 0
	for _, def := range current {
		if def.Status != "deprecated" {
			active++
		}
		if before, exists := old[def.Key]; exists {
			retained++
			if !v12JSONEqual(t, def.RawSlide, before.RawSlide) || !reflect.DeepEqual(def.Slots, before.Slots) || !reflect.DeepEqual(def.Arrays, before.Arrays) {
				t.Fatalf("retained composition or binding changed: %s", def.Key)
			}
		} else {
			if !added[def.Key] {
				t.Fatalf("unexpected addition: %s", def.Key)
			}
			counts[def.Family]++
		}
	}
	if len(current) != 677 || active != 676 || retained != 649 || !reflect.DeepEqual(counts, map[string]int{"software": 10, "roadmaps": 9, "team": 9}) {
		t.Fatalf("catalog delta changed: %d/%d/%d, %v", len(current), active, retained, counts)
	}
}

func TestLibraryV12NewBindingsRoundTripAndBuild(t *testing.T) {
	bundle := v12Bundle()
	source, err := LibrarySourceReference(bundle, "", "", 2026)
	if err != nil {
		t.Fatal(err)
	}
	input, err := LibraryReference(bundle, "", "", 2026)
	if err != nil {
		t.Fatal(err)
	}
	keys := map[string]bool{}
	for _, key := range v12AddedKeys(t) {
		keys[key] = true
	}
	var selected []SlideSpec
	for _, slide := range source.Slides {
		if keys[slide.TemplateBinding.Template] {
			selected = append(selected, slide)
		}
	}
	source.Slides = selected
	var values []BoundSlide
	for _, slide := range input.Slides {
		if keys[slide.Template] {
			values = append(values, slide)
		}
	}
	input.Slides = values
	bound, _, err := BindTemplates(bundle, "", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(source.Slides) != 28 || len(bound.Slides) != 28 {
		t.Fatal("new source/bound count changed")
	}
	for i := range source.Slides {
		if !reflect.DeepEqual(v6ComparisonSlide(t, source.Slides[i], true), v6ComparisonSlide(t, bound.Slides[i], true)) {
			t.Fatalf("content roundtrip changed %s", source.Slides[i].TemplateBinding.Template)
		}
	}
	if testing.Short() {
		return
	}
	for _, doc := range []Document{source, bound} {
		for _, slide := range doc.Slides {
			one := doc
			one.Slides = []SlideSpec{slide}
			t.Run(slide.TemplateBinding.Template, func(t *testing.T) {
				deck, report, err := BuildWithEngine(bundle, "", one, CandidateEngine)
				if err != nil {
					t.Fatal(err)
				}
				if len(deck) == 0 || len(report.Slides) != 1 || report.PowerPointVerified || report.VisuallyReviewed {
					t.Fatal("build count or review claim changed")
				}
			})
		}
	}
}

func v12JSONEqual(t *testing.T, a, b json.RawMessage) bool {
	t.Helper()
	var x, y any
	if err := json.Unmarshal(a, &x); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &y); err != nil {
		t.Fatal(err)
	}
	return reflect.DeepEqual(x, y)
}

func TestLibraryV12RetainedCompiledCompositions(t *testing.T) {
	previous, err := LibrarySourceReference(filepath.Join("..", "..", "library", "wm-design-system", "v11"), "", "", 2026)
	if err != nil {
		t.Fatal(err)
	}
	current, err := LibrarySourceReference(v12Bundle(), "", "", 2026)
	if err != nil {
		t.Fatal(err)
	}
	old := map[string]SlideSpec{}
	for _, slide := range previous.Slides {
		old[slide.TemplateBinding.Template] = slide
	}
	retained := 0
	for _, slide := range current.Slides {
		before, exists := old[slide.TemplateBinding.Template]
		if !exists {
			continue
		}
		retained++
		if !reflect.DeepEqual(v6ComparisonSlide(t, before, true), v6ComparisonSlide(t, slide, true)) {
			t.Fatalf("retained compiled composition changed: %s", slide.TemplateBinding.Template)
		}
	}
	if retained != 649 {
		t.Fatal("retained source identity count changed")
	}
}
