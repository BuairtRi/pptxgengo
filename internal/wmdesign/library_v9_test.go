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

func v9IntakeBundle() string {
	return filepath.Join("..", "..", "planning", "wm-design-contracts", "v9", "intake-20261006-631-frozen", "bundle")
}

func TestLibraryV9FrozenSnapshotClosure(t *testing.T) {
	bundle := v9IntakeBundle()
	for file, want := range map[string]string{
		"bundle.json":    "0ad33b662d7e3a9b47a47237f0037c09cba959ee2693d39b4cc36ed230ab75c1",
		"inventory.json": "9ae0d5af692d4dfe3c9807d232202c2461f55168fe926689ee833f82f8001fb3",
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
	if len(source.Files) != 39 {
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
	previous, err := Load(v8IntakeBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(source.Tokens, previous.Tokens) {
		t.Fatal("intake changed typography tokens")
	}
}

func TestLibraryV9PinnedIntakeAndV8Inheritance(t *testing.T) {
	source, err := Load(v9IntakeBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	if source.Revision != LibraryRevisionV9 || source.Commit != "03c25c39643eb0edd75ece71663ddfd0c4cb5358" {
		t.Fatal("candidate identity changed")
	}
	current, err := LibraryCatalog(v9IntakeBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	previous, err := LibraryCatalog(v8IntakeBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	old := map[string]LibraryTemplate{}
	for _, def := range previous {
		old[def.Key] = def
	}
	additions := 0
	for _, def := range current {
		before, exists := old[def.Key]
		if !exists {
			additions++
			if def.Family != "workshops" || def.SourceRevision != LibraryRevisionV9 {
				t.Fatalf("unexpected addition %s", def.Key)
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
	if len(current) != 631 || additions != 8 || len(old) != 0 {
		t.Fatalf("intake drift: total %d, added %d, removed %d", len(current), additions, len(old))
	}
}

func TestLibraryV9WorkshopsRoundTripAndBuild(t *testing.T) {
	bundle := v9IntakeBundle()
	source, err := LibrarySourceReference(bundle, "", "workshops", 2026)
	if err != nil {
		t.Fatal(err)
	}
	input, err := LibraryReference(bundle, "", "workshops", 2026)
	if err != nil {
		t.Fatal(err)
	}
	bound, _, err := BindTemplates(bundle, "", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(source.Slides) != 29 || len(bound.Slides) != 29 {
		t.Fatal("workshop inventory drift")
	}
	for i := range source.Slides {
		if !reflect.DeepEqual(v6ComparisonSlide(t, source.Slides[i], true), v6ComparisonSlide(t, bound.Slides[i], true)) {
			t.Fatalf("binding changed authored workshop %s", source.Slides[i].TemplateBinding.Template)
		}
	}
	for name, doc := range map[string]Document{"source": source, "bound": bound} {
		t.Run(name, func(t *testing.T) {
			deck, report, err := BuildWithEngine(bundle, "", doc, CandidateEngine)
			if err != nil {
				t.Fatal(err)
			}
			if len(deck) == 0 || len(report.Slides) != 29 || report.PowerPointVerified || report.VisuallyReviewed {
				t.Fatal("incorrect build or native acceptance claim")
			}
		})
	}
}
