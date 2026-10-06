package wmdesign

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func v10IntakeBundle() string {
	return filepath.Join("..", "..", "planning", "wm-design-contracts", "v10", "intake-20261006-649-frozen", "bundle")
}

func TestLibraryV10FrozenSnapshotClosure(t *testing.T) {
	bundle := v10IntakeBundle()
	for file, want := range map[string]string{
		"bundle.json":    "45d25e4d920165425a661aa8ceea36a24b979070b673f547f2294d0a65d109c0",
		"inventory.json": "aea092e8e5e1aca02900ab87b90b294d014ac2c19937049f735e1ad30a7c9124",
	} {
		data, err := os.ReadFile(filepath.Join(bundle, file))
		if err != nil || fmt.Sprintf("%x", sha256.Sum256(data)) != want {
			t.Fatalf("pin changed: %s: %v", file, err)
		}
	}
	source, err := Load(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(source.Files) != 40 {
		t.Fatalf("inventory drift: %d sources", len(source.Files))
	}
	var inventory struct {
		Sources []struct {
			Path  string `json:"path"`
			Bytes int    `json:"bytes"`
		} `json:"sources"`
	}
	raw, err := os.ReadFile(filepath.Join(bundle, "inventory.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &inventory); err != nil {
		t.Fatal(err)
	}
	for _, file := range inventory.Sources {
		data, err := os.ReadFile(filepath.Join(source.Root, file.Path))
		if err != nil || len(data) != file.Bytes {
			t.Fatalf("source size drift: %s: %v", file.Path, err)
		}
	}
	previous, err := Load(v9IntakeBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(source.Tokens, previous.Tokens) {
		t.Fatal("intake changed typography tokens")
	}
	var manifests [2]struct {
		Files []SourceFile `json:"files"`
	}
	for i, root := range []string{v9IntakeBundle(), bundle} {
		data, err := os.ReadFile(filepath.Join(root, "bundle.json"))
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(data, &manifests[i]); err != nil {
			t.Fatal(err)
		}
	}
	if len(manifests[1].Files) != 15 || !reflect.DeepEqual(manifests[0].Files, manifests[1].Files) {
		t.Fatal("intake changed font, license or asset identities")
	}
}

func TestLibraryV10PinnedIntakeAndV9Inheritance(t *testing.T) {
	source, err := Load(v10IntakeBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	if source.Revision != LibraryRevisionV10 || source.Commit != "c14fb286fb38e15800a6fd476a1ed67956f1165f" {
		t.Fatal("candidate identity changed")
	}
	current, err := LibraryCatalog(v10IntakeBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	previous, err := LibraryCatalog(v9IntakeBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	old := map[string]LibraryTemplate{}
	for _, def := range previous {
		old[def.Key] = def
	}
	additions := 0
	expectedAdditions := map[string]bool{}
	for _, key := range v10AddedTemplateKeys {
		expectedAdditions[key] = true
	}
	for _, def := range current {
		before, exists := old[def.Key]
		if !exists {
			additions++
			if (def.Family != "argument" && def.Family != "roadmaps") || def.SourceRevision != LibraryRevisionV10 {
				t.Fatalf("unexpected addition %s", def.Key)
			}
			if !expectedAdditions[def.Key] {
				t.Fatalf("unexpected added key %s", def.Key)
			}
			delete(expectedAdditions, def.Key)
			continue
		}
		delete(old, def.Key)
		var a, b any
		if err := json.Unmarshal(before.RawSlide, &a); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(def.RawSlide, &b); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(a, b) || def.ContentContract != before.ContentContract || !reflect.DeepEqual(def.Slots, before.Slots) || !reflect.DeepEqual(def.Arrays, before.Arrays) || !reflect.DeepEqual(def.ValueSchema, before.ValueSchema) {
			t.Fatalf("retained source or binding changed %s", def.Key)
		}
		left, err := compileLibrarySlide(before.RawSlide, nil)
		if err != nil {
			t.Fatal(err)
		}
		right, err := compileLibrarySlide(def.RawSlide, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := applyLibraryRefinements(before.Key, before.SourceRevision, &left); err != nil {
			t.Fatal(err)
		}
		if err := applyLibraryRefinements(def.Key, def.SourceRevision, &right); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(v6ComparisonSlide(t, left, false), v6ComparisonSlide(t, right, false)) {
			t.Fatalf("retained amendments changed %s", def.Key)
		}
	}
	if len(current) != 649 || additions != 18 || len(old) != 0 || len(expectedAdditions) != 0 {
		t.Fatalf("intake drift: total %d, added %d, removed %d", len(current), additions, len(old))
	}
}

var v10AddedTemplateKeys = []string{
	"pillars/three-why-overlap",
	"pillars/four-why-overlap",
	"pillars/three-why-matters",
	"pillars/three-why-overlap-dense",
	"pillars/three-why-overlap-nav",
	"road-fork/parallel",
	"road-fork/decision",
	"road-fork/decision-chosen",
	"road-fork/parallel-split",
	"road-fork/decision-criteria",
	"road-fork/parallel-detail",
	"road-fork/foundation-then-waves",
	"roadmap-narrative/horizon-table",
	"roadmap-narrative/memo",
	"roadmap-narrative/waves-criteria",
	"roadmap-narrative/objectives-tree",
	"roadmap-narrative/decisions-roadmap",
	"roadmap-narrative/kpi-roadmap",
}

func TestLibraryV10NewTemplatesRoundTripAndBuild(t *testing.T) {
	bundle := v10IntakeBundle()
	source, err := LibrarySourceReference(bundle, "", "", 2026)
	if err != nil {
		t.Fatal(err)
	}
	input, err := LibraryReference(bundle, "", "", 2026)
	if err != nil {
		t.Fatal(err)
	}
	selected := map[string]bool{}
	for _, key := range v10AddedTemplateKeys {
		selected[key] = true
	}
	if len(selected) != 18 {
		t.Fatal("new-key inventory drift")
	}
	var sourceSlides []SlideSpec
	for _, slide := range source.Slides {
		if selected[slide.TemplateBinding.Template] {
			sourceSlides = append(sourceSlides, slide)
		}
	}
	source.Slides = sourceSlides
	var boundSlides []BoundSlide
	for _, slide := range input.Slides {
		if selected[slide.Template] {
			boundSlides = append(boundSlides, slide)
		}
	}
	input.Slides = boundSlides
	bound, _, err := BindTemplates(bundle, "", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(source.Slides) != 18 || len(bound.Slides) != 18 {
		t.Fatal("new-template inventory drift")
	}
	for i := range source.Slides {
		if !reflect.DeepEqual(v6ComparisonSlide(t, source.Slides[i], true), v6ComparisonSlide(t, bound.Slides[i], true)) {
			t.Fatalf("binding changed authored template %s", source.Slides[i].TemplateBinding.Template)
		}
	}
	for name, doc := range map[string]Document{"source": source, "bound": bound} {
		t.Run(name, func(t *testing.T) {
			deck, report, err := BuildWithEngine(bundle, "", doc, CandidateEngine)
			if err != nil {
				t.Fatal(err)
			}
			if len(deck) == 0 || len(report.Slides) != 18 || report.PowerPointVerified || report.VisuallyReviewed {
				t.Fatal("incorrect build or native acceptance claim")
			}
		})
	}
}
