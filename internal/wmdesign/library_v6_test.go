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

func v6IntakeBundle() string {
	return filepath.Join("..", "..", "planning", "wm-design-contracts", "v6", "intake-20261004-602-frozen", "bundle")
}

func v6ComparisonSlide(t *testing.T, slide SlideSpec, ignoreBindingKeys bool) any {
	t.Helper()
	raw, err := json.Marshal(slide)
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatal(err)
	}
	if ignoreBindingKeys {
		delete(value, "id")
		delete(value, "template_binding")
		if nodes, ok := value["nodes"].([]any); ok {
			for _, rawNode := range nodes {
				if node, ok := rawNode.(map[string]any); ok {
					if scene, ok := node["scene"].(map[string]any); ok {
						delete(scene, "keys")
					}
				}
			}
		}
	}
	return value
}

func TestLibraryV6PinnedIntakeAndV5Inheritance(t *testing.T) {
	bundle := v6IntakeBundle()
	s, err := Load(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	if s.Revision != LibraryRevisionV6 || s.Commit != "c355c881d543dceccb82dbe12d7129bca9b1fcac" {
		t.Fatal("candidate identity changed")
	}
	for file, want := range map[string]string{
		"bundle.json":    "9b1c303958152e6adddbb84b1fcde1bdb73c1a6f43ac2906ecbbd9b446c28716",
		"inventory.json": "586a742ee2bdfb05ee955af613f4c36e2059ed3504abd69cf652966e9781132e",
	} {
		data, err := os.ReadFile(filepath.Join(bundle, file))
		if err != nil || fmt.Sprintf("%x", sha256.Sum256(data)) != want {
			t.Fatalf("pin changed: %s: %v", file, err)
		}
	}
	current, err := LibraryCatalog(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	previous, err := LibraryCatalog(filepath.Join("..", "..", "library", "wm-design-system", "v5"), "")
	if err != nil {
		t.Fatal(err)
	}
	old := make(map[string]LibraryTemplate)
	for _, def := range previous {
		old[def.Key] = def
	}
	added, revised, retained := 0, 0, 0
	for _, def := range current {
		before, exists := old[def.Key]
		if !exists {
			added++
			if !isV6HeatmapDelta(def.Key) {
				t.Fatalf("addition missing explicit revision map: %s", def.Key)
			}
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
		if !reflect.DeepEqual(a, b) {
			revised++
			if def.Key != "capability-heat/dense" || !isV6HeatmapDelta(def.Key) {
				t.Fatalf("unexpected composition revision: %s", def.Key)
			}
			continue
		}
		retained++
		if isV6HeatmapDelta(def.Key) || def.ContentContract != before.ContentContract || !reflect.DeepEqual(def.Slots, before.Slots) || !reflect.DeepEqual(def.Arrays, before.Arrays) || !reflect.DeepEqual(def.ValueSchema, before.ValueSchema) {
			t.Fatalf("unchanged binding API changed: %s", def.Key)
		}
		afterSlide, err := compileLibrarySlide(def.RawSlide, nil)
		if err != nil {
			t.Fatal(err)
		}
		beforeSlide, err := compileLibrarySlide(before.RawSlide, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := applyLibraryRefinements(def.Key, def.SourceRevision, &afterSlide); err != nil {
			t.Fatal(err)
		}
		if err := applyLibraryRefinements(before.Key, before.SourceRevision, &beforeSlide); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(v6ComparisonSlide(t, beforeSlide, false), v6ComparisonSlide(t, afterSlide, false)) {
			t.Fatalf("accepted composition amendment changed: %s", def.Key)
		}
	}
	if len(current) != 602 || added != 15 || revised != 1 || retained != 586 || len(old) != 0 {
		t.Fatalf("intake drift: total=%d added=%d revised=%d retained=%d removed=%d", len(current), added, revised, retained, len(old))
	}
}

func TestLibraryV6HeatmapsRoundTripAndBuild(t *testing.T) {
	bundle := v6IntakeBundle()
	source, err := LibrarySourceReference(bundle, "", "heatmaps", 2026)
	if err != nil {
		t.Fatal(err)
	}
	input, err := LibraryReference(bundle, "", "heatmaps", 2026)
	if err != nil {
		t.Fatal(err)
	}
	bound, _, err := BindTemplates(bundle, "", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(source.Slides) != 37 || len(bound.Slides) != 37 {
		t.Fatal("heatmap inventory drift")
	}
	for i := range source.Slides {
		a, b := source.Slides[i], bound.Slides[i]
		key := a.TemplateBinding.Template
		if key != b.TemplateBinding.Template {
			t.Fatal("template ordering changed")
		}
		// Binding provenance deliberately differs; authored copy, numeric values,
		// reference identity, priority labels and composition must remain exact.
		a.ID, b.ID = "roundtrip", "roundtrip"
		a.TemplateBinding, b.TemplateBinding = nil, nil
		if !reflect.DeepEqual(v6ComparisonSlide(t, a, true), v6ComparisonSlide(t, b, true)) {
			t.Fatalf("source content changed through binding: %s", key)
		}
	}
	for name, doc := range map[string]Document{"source": source, "bound": bound} {
		for _, slide := range doc.Slides {
			t.Run(name+"/"+slide.TemplateBinding.Template, func(t *testing.T) {
				one := doc
				one.Slides = []SlideSpec{slide}
				deck, report, err := BuildWithEngine(bundle, "", one, CandidateEngine)
				if err != nil {
					t.Fatal(err)
				}
				if len(deck) == 0 || len(report.Slides) != 1 || report.PowerPointVerified || report.VisuallyReviewed {
					t.Fatal("incorrect build or native qualification claim")
				}
			})
		}
	}
}

func TestLibraryV6AllTemplateBuilds(t *testing.T) {
	if testing.Short() {
		t.Skip("602-source/601-bound catalog qualification belongs to integration checks")
	}
	bundle := v6IntakeBundle()
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
	if len(source.Slides) != 602 || len(bound.Slides) != 601 {
		t.Fatalf("qualification inventory changed: %d source, %d active bound", len(source.Slides), len(bound.Slides))
	}
	for name, doc := range map[string]Document{"source": source, "bound": bound} {
		t.Run(name, func(t *testing.T) {
			deck, report, err := BuildWithEngine(bundle, "", doc, CandidateEngine)
			if err != nil {
				t.Fatal(err)
			}
			if len(deck) == 0 || len(report.Slides) != len(doc.Slides) || report.PowerPointVerified || report.VisuallyReviewed {
				t.Fatal("incorrect catalog build or native acceptance claim")
			}
		})
	}
}

func TestLibraryV6AmendmentsAreAtomicIdempotentAndCopyPreserving(t *testing.T) {
	catalog, err := LibraryCatalog(v6IntakeBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	for _, def := range catalog {
		if !isV6HeatmapDelta(def.Key) {
			continue
		}
		t.Run(def.Key, func(t *testing.T) {
			original, err := compileLibrarySlide(def.RawSlide, nil)
			if err != nil {
				t.Fatal(err)
			}
			before := v6ComparisonSlide(t, original, false)
			amended := original
			if err := applyV6HeatmapRefinements(def.Key, &amended); err != nil {
				t.Fatal(err)
			}
			once := v6ComparisonSlide(t, amended, false)
			if err := applyV6HeatmapRefinements(def.Key, &amended); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(once, v6ComparisonSlide(t, amended, false)) {
				t.Fatal("amendment is not idempotent")
			}
			// The source slide's node slice is never mutated while applying a
			// candidate amendment, including on successful publication.
			if !reflect.DeepEqual(before, v6ComparisonSlide(t, original, false)) {
				t.Fatal("source slide mutated")
			}
			for i, node := range original.Nodes {
				a, err := libraryObject(node.Scene.Node)
				if err != nil {
					t.Fatal(err)
				}
				b, err := libraryObject(amended.Nodes[i].Scene.Node)
				if err != nil {
					t.Fatal(err)
				}
				delete(a, "y")
				delete(b, "y")
				if !reflect.DeepEqual(a, b) {
					t.Fatal("amendment changed copy, styles or structural content")
				}
			}
		})
	}
	for _, def := range catalog {
		if def.Key != "capability-heat/grouped-split" {
			continue
		}
		broken, err := compileLibrarySlide(def.RawSlide, nil)
		if err != nil {
			t.Fatal(err)
		}
		obj, err := libraryObject(broken.Nodes[3].Scene.Node)
		if err != nil {
			t.Fatal(err)
		}
		obj["y"] = 999
		broken.Nodes[3].Scene.Node, err = json.Marshal(obj)
		if err != nil {
			t.Fatal(err)
		}
		before := v6ComparisonSlide(t, broken, false)
		if applyV6HeatmapRefinements(def.Key, &broken) == nil {
			t.Fatal("unexpected geometry accepted")
		}
		if !reflect.DeepEqual(before, v6ComparisonSlide(t, broken, false)) {
			t.Fatal("failed second target partially published first")
		}
	}
}
