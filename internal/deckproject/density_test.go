package deckproject

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestSlideDensityYAMLAndCompilationOverlays(t *testing.T) {
	p := example(t)
	withDensity := strings.Replace(string(p.Raw),
		"  - id: maintain-the-source\n    content_kind: synthetic_example\n",
		"  - id: maintain-the-source\n    content_kind: synthetic_example\n    density: compact\n    header_density: dense\n    auto_density: false\n", 1)
	withDensity = strings.Replace(withDensity,
		"  - id: local-composition\n    content_kind: synthetic_example\n",
		"  - id: local-composition\n    content_kind: synthetic_example\n    density: dense\n    header_density: comfortable\n    auto_density: true\n", 1)
	if withDensity == string(p.Raw) || withDensity == "" {
		t.Fatal("fixture density fields were not inserted")
	}
	if err := os.WriteFile(p.SourcePath, []byte(withDensity), 0644); err != nil {
		t.Fatal(err)
	}
	p, err := Load(p.SourcePath)
	if err != nil {
		t.Fatal(err)
	}
	if p.Document.Slides[0].Density != "compact" || p.Document.Slides[0].HeaderDensity != "dense" || p.Document.Slides[0].AutoDensity == nil || *p.Document.Slides[0].AutoDensity {
		t.Fatalf("shared slide density fields not preserved: %+v", p.Document.Slides[0])
	}
	compiled, err := Compile(p, bundle(t), wmdesign.CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	shared := compiled.Document.Slides[0]
	if shared.Density != "compact" || shared.Frame.HeaderDensity != "dense" || shared.AutoDensity == nil || *shared.AutoDensity {
		t.Fatalf("shared slide density overlay not applied: %+v", shared)
	}
	local := compiled.Document.Slides[1]
	if local.Density != "dense" || local.Frame.HeaderDensity != "comfortable" || local.AutoDensity == nil || !*local.AutoDensity {
		t.Fatalf("local slide density overlay not applied: %+v", local)
	}
}

func TestSlideDensityRejectsUnknownLevels(t *testing.T) {
	p := example(t)
	for _, field := range []string{"density", "header_density"} {
		t.Run(field, func(t *testing.T) {
			path := filepath.Join(p.Root, "bad-density.yaml")
			raw := strings.Replace(string(p.Raw),
				"  - id: maintain-the-source\n    content_kind: synthetic_example\n",
				"  - id: maintain-the-source\n    content_kind: synthetic_example\n    "+field+": tiny\n", 1)
			if err := os.WriteFile(path, []byte(raw), 0644); err != nil {
				t.Fatal(err)
			}
			_, err := Load(path)
			if err == nil || !strings.Contains(err.Error(), field) || !strings.Contains(err.Error(), "comfortable, compact, or dense") {
				t.Fatalf("expected invalid %s error, got %v", field, err)
			}
		})
	}
}
