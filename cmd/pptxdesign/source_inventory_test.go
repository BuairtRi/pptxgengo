package main

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/buairtri/pptxgengo/internal/deckproject"
	"github.com/buairtri/pptxgengo/internal/nativepkg"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestSourceInventoryWritesJSONAndMarkdownWithoutChangingInput(t *testing.T) {
	input, err := filepath.Abs("../../samples/final/dentalxchange/deck.pptx")
	if err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(input)
	if err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "inventory")
	if err = runSourceInventory([]string{"--in", input, "--out", output}); err != nil {
		t.Fatal(err)
	}
	var report nativepkg.SourceInventory
	data, err := os.ReadFile(filepath.Join(output, "source_inventory.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if report.SlideCount == 0 || report.SourceSHA256 == "" || len(report.Slides) != report.SlideCount {
		t.Fatalf("incomplete source inventory: %+v", report)
	}
	markdown, err := os.ReadFile(filepath.Join(output, "source_inventory.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(markdown) == 0 {
		t.Fatal("Markdown report is empty")
	}
	after, err := os.ReadFile(input)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("source file changed")
	}
	if err = runSourceInventory([]string{"--in", input, "--out", output}); err == nil {
		t.Fatal("accepted an existing output directory")
	}
}

func TestInventoryMappingKeepsSourcePagesExplicitlyUnmapped(t *testing.T) {
	oldReleaseRoot := os.Getenv("PPTXGENGO_RELEASE_ROOT")
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Setenv("PPTXGENGO_RELEASE_ROOT", repoRoot); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Setenv("PPTXGENGO_RELEASE_ROOT", oldReleaseRoot) })
	project := filepath.Join(t.TempDir(), "project")
	copyExampleProject(t, "../../examples/deck-project", project)
	p, err := deckproject.Load(project)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := filepath.Abs("../../library/wm-design-system/v5")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = deckproject.Pin(p, bundle, wmdesign.CandidateEngine); err != nil {
		t.Fatal(err)
	}
	source := nativepkg.SourceInventory{SourceSHA256: "source-hash", Slides: []nativepkg.SourceSlideInventory{{OriginalSlideIndex: 1, Title: "A", Hidden: false}, {OriginalSlideIndex: 2, Title: "B", Hidden: true}}}
	result, err := buildInventoryMapping(project, source)
	if err != nil {
		t.Fatal(err)
	}
	if result.MappingStatus != "source_slides_unmapped" || len(result.SourceSlides) != 2 || len(result.ProjectSlides) != len(p.Document.Slides) {
		t.Fatalf("unexpected mapping report: %+v", result)
	}
	for _, slide := range result.SourceSlides {
		if slide.Status != "unmapped" || slide.ProjectSlideID != "" {
			t.Fatalf("source page was heuristically matched: %+v", slide)
		}
	}
	for _, slide := range result.ProjectSlides {
		if slide.MappingStatus != "unmapped" || slide.TemplateID == "" || slide.TemplateScope == "" {
			t.Fatalf("project source mapping lacks explicit refs or honest status: %+v", slide)
		}
	}
}

func copyExampleProject(t *testing.T, source, target string) {
	t.Helper()
	err := filepath.WalkDir(source, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, current)
		if err != nil {
			return err
		}
		path := filepath.Join(target, relative)
		if entry.IsDir() {
			return os.MkdirAll(path, 0755)
		}
		data, err := os.ReadFile(current)
		if err != nil {
			return err
		}
		return os.WriteFile(path, data, 0644)
	})
	if err != nil {
		t.Fatal(err)
	}
}
