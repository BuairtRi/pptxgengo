package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/compose"
	"github.com/buairtri/pptxgengo/internal/textlayout"
	"golang.org/x/image/font/gofont/goregular"
)

func TestGoLayoutCLIWithoutNativeTools(t *testing.T) {
	root := t.TempDir()
	fonts := filepath.Join(root, "fonts")
	if err := os.Mkdir(fonts, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fonts, "Go.ttf"), goregular.TTF, 0600); err != nil {
		t.Fatal(err)
	}
	// No executables can be found. The complete Go build path must still work.
	t.Setenv("PATH", root)
	spec := compose.Spec{Schema: "pptxgengo.compose-spec.v1", Slides: []compose.SlideSpec{{ID: "slide", WidthPt: 960, HeightPt: 540, TitleFontFace: "Go", TitleFontSizePt: 24, Pods: []compose.PodSpec{}, Canvas: []compose.CanvasSpec{{ID: "text", Kind: "text", Bounds: compose.Rect{X: 40, Y: 40, Width: 400, Height: 120}, Text: "Modern typography measured in Go. Editable text in the generated slide.", FontFace: "Go", FontSizePt: 16, Foreground: "#070154", Align: "left", Valign: "top"}}}}}
	path := filepath.Join(root, "spec.json")
	if err := writeNew(path, jsonBytes(spec)); err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{"measure", "fit-report", "build"} {
		out := filepath.Join(root, command)
		if err := run([]string{command, "--engine", "go", "--spec", path, "--font-dir", fonts, "--out", out}); err != nil {
			t.Fatalf("%s: %v", command, err)
		}
		if command == "measure" {
			var report textlayout.Report
			if _, err := readJSON(out, &report); err != nil {
				t.Fatal(err)
			}
			if report.PowerPointVerified || report.SpecSHA256 == "" {
				t.Fatal("incorrect provenance")
			}
		}
		if command == "build" {
			var m manifest
			if _, err := readJSON(filepath.Join(out, "manifest.json"), &m); err != nil {
				t.Fatal(err)
			}
			if m.GoLayout == nil || m.Environment != nil || m.Plan == nil || m.GoLayout.PowerPointVerified {
				t.Fatal("Go result mislabeled as native")
			}
			if _, err := os.Stat(filepath.Join(out, "go-layout.json")); err != nil {
				t.Fatal(err)
			}
			deck, err := os.ReadFile(filepath.Join(out, m.DeckFile))
			if err != nil {
				t.Fatal(err)
			}
			if hash(deck) != m.DeckSHA {
				t.Fatal("deck hash mismatch")
			}
			if err := validateShapeStructure(deck, m.Slides); err != nil {
				t.Fatal(err)
			}
		}
		if err := run([]string{command, "--engine", "go", "--spec", path, "--font-dir", fonts, "--out", out}); err == nil {
			t.Fatal("output overwritten")
		}
	}
	for _, flags := range [][]string{{"--cache", root}, {"--adapter", "never.applescript"}, {"--bundle", root}, {"--evidence", path}} {
		args := append([]string{"build", "--engine", "go", "--spec", path, "--out", filepath.Join(root, "rejected")}, flags...)
		if err := run(args); err == nil {
			t.Fatalf("native-only flags accepted: %v", flags)
		}
	}
	if err := run([]string{"verify", "--engine", "go", "--out", filepath.Join(root, "verify")}); err == nil || !strings.Contains(err.Error(), "supports") {
		t.Fatal("Go verification should be rejected")
	}
	// Report an overflowing fixed zone and refuse to generate that deck.
	spec.Slides[0].Canvas[0].Bounds.Height = 5
	overflowSpec := filepath.Join(root, "overflow-spec.json")
	if err := writeNew(overflowSpec, jsonBytes(spec)); err != nil {
		t.Fatal(err)
	}
	overflowReport := filepath.Join(root, "overflow-report.json")
	if err := run([]string{"fit-report", "--engine", "go", "--spec", overflowSpec, "--font-dir", fonts, "--out", overflowReport}); err != nil {
		t.Fatal(err)
	}
	var fit struct {
		PlannerPassed bool `json:"planner_passed"`
		OverflowCount int  `json:"overflow_count"`
	}
	b, err := os.ReadFile(overflowReport)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &fit); err != nil {
		t.Fatal(err)
	}
	if fit.PlannerPassed || fit.OverflowCount != 1 {
		t.Fatal("overflow not reported")
	}
	failedOut := filepath.Join(root, "overflow-build")
	if err := run([]string{"build", "--engine", "go", "--spec", overflowSpec, "--font-dir", fonts, "--out", failedOut}); err == nil {
		t.Fatal("overflowing deck built")
	}
	if _, err := os.Stat(failedOut); !os.IsNotExist(err) {
		t.Fatal("failed build left an output bundle")
	}
}
