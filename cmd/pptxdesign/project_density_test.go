package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/deckproject"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestProjectDensityWarningsUseTypedAdjustments(t *testing.T) {
	report := wmdesign.Report{
		Slides:             []wmdesign.SlideReport{{ID: "body-fit", Density: &wmdesign.SlideDensityRecord{Header: "comfortable"}}},
		DensityAdjustments: []wmdesign.SlideDensityAdjustment{{SlideID: "body-fit", Requested: "comfortable", Resolved: "compact", Reason: "body overflow"}},
	}
	warnings := projectDensityWarningMessages(report)
	if len(warnings) != 1 {
		t.Fatalf("wanted one density warning, got %#v", warnings)
	}
	want := "warning: Slide body-fit: body density changed comfortable → compact to fit supplied content (header comfortable unchanged)."
	if warnings[0] != want {
		t.Fatalf("warning = %q, want %q", warnings[0], want)
	}
	if got := projectDensityWarningMessages(wmdesign.Report{}); len(got) != 0 {
		t.Fatalf("unexpected warning without adjustments: %#v", got)
	}
}

func TestLibraryMatchDensityWarningsReadCandidateLayout(t *testing.T) {
	path := filepath.Join(t.TempDir(), "candidate-layout.json")
	report := wmdesign.Report{
		Slides: []wmdesign.SlideReport{{ID: "candidate-001", Density: &wmdesign.SlideDensityRecord{Header: "comfortable"}}},
		DensityAdjustments: []wmdesign.SlideDensityAdjustment{{
			SlideID: "candidate-001", Requested: "comfortable", Resolved: "dense",
			Steps: []wmdesign.DensityStep{{From: "comfortable", To: "compact"}, {From: "compact", To: "dense"}},
		}},
	}
	if err := wmdesign.WriteJSON(path, report); err != nil {
		t.Fatal(err)
	}
	warnings, err := libraryMatchDensityWarningMessages(deckproject.LibraryMatchReport{
		Candidates: []deckproject.MatchCandidate{{LayoutReport: path}},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := "warning: Slide candidate-001: body density changed comfortable → compact → dense to fit supplied content (header comfortable unchanged)."
	if len(warnings) != 1 || warnings[0] != want {
		t.Fatalf("warnings = %#v, want one warning %q", warnings, want)
	}
	stderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	err = emitLibraryMatchDensityWarnings(deckproject.LibraryMatchReport{
		Candidates: []deckproject.MatchCandidate{{LayoutReport: path}},
	})
	_ = w.Close()
	os.Stderr = stderr
	if err != nil {
		t.Fatal(err)
	}
	stderrOutput, err := io.ReadAll(r)
	_ = r.Close()
	if err != nil || strings.TrimSpace(string(stderrOutput)) != want {
		t.Fatalf("stderr warning = %q, err=%v; want %q", stderrOutput, err, want)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}
