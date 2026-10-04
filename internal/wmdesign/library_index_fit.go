package wmdesign

import (
	"fmt"
	"os"
	"path/filepath"
)

type LibraryFitCandidate struct {
	SlideID      string   `json:"slide_id"`
	Template     string   `json:"template"`
	Status       string   `json:"status"`
	Error        string   `json:"error,omitempty"`
	Deck         string   `json:"deck,omitempty"`
	LayoutReport string   `json:"layout_report,omitempty"`
	Warnings     []string `json:"warnings,omitempty"`
}

type LibraryFitReport struct {
	Schema         string                `json:"schema"`
	InputSHA256    string                `json:"input_sha256"`
	Engine         string                `json:"engine"`
	SourceRevision string                `json:"source_revision"`
	Candidates     []LibraryFitCandidate `json:"candidates"`
	Passed         int                   `json:"passed"`
	Failed         int                   `json:"failed"`
	Policy         []string              `json:"policy"`
}

// FitLibraryCandidates checks independently supplied candidate bindings. It
// never remaps meanings from a different template or borrows specimen copy.
func FitLibraryCandidates(bundle, source, engine, out string, input BoundDocument) (LibraryFitReport, error) {
	report := LibraryFitReport{Schema: "pptxgengo.library-content-fit.v1", InputSHA256: indexDigest(indexJSON(input)), Engine: engine, Candidates: []LibraryFitCandidate{}, Policy: []string{"Every candidate must carry explicit supplied-content values for its own closed contract.", "Successful Go binding/layout is not native visual review or general template qualification.", "Compare rendered alternatives using the actual content; retain material caveats and evidence."}}
	if out == "" || len(input.Slides) == 0 {
		return report, fmt.Errorf("library.fit_requires_candidates_and_new_output")
	}
	if _, e := os.Stat(out); !os.IsNotExist(e) {
		return report, fmt.Errorf("library.fit_new_output_required")
	}
	s, e := Load(bundle, source)
	if e != nil {
		return report, e
	}
	report.SourceRevision = s.Revision
	seen := map[string]bool{}
	for _, candidate := range input.Slides {
		if candidate.ContentKind != "supplied_content" || candidate.ID == "" || seen[candidate.ID] {
			return report, fmt.Errorf("library.fit_requires_unique_supplied_content_ids")
		}
		seen[candidate.ID] = true
	}
	if e = os.MkdirAll(filepath.Dir(out), 0755); e != nil {
		return report, e
	}
	if e = os.Mkdir(out, 0755); e != nil {
		return report, e
	}
	if e = WriteJSON(filepath.Join(out, "candidate-content.json"), input); e != nil {
		return report, e
	}
	for i, candidate := range input.Slides {
		result := LibraryFitCandidate{SlideID: candidate.ID, Template: candidate.Template, Status: "binding_or_layout_failed"}
		one := input
		one.Slides = []BoundSlide{candidate}
		doc, _, err := BindTemplates(bundle, source, one)
		if err == nil {
			var deck []byte
			var layout Report
			deck, layout, err = BuildWithEngine(bundle, source, doc, engine)
			if err == nil {
				prefix := fmt.Sprintf("candidate-%03d", i+1)
				deckPath, layoutPath := filepath.Join(out, prefix+".pptx"), filepath.Join(out, prefix+".layout.json")
				if e = os.WriteFile(deckPath, deck, 0644); e != nil {
					return report, e
				}
				if e = WriteJSON(layoutPath, layout); e != nil {
					return report, e
				}
				if e = WriteJSON(filepath.Join(out, prefix+".foundation.json"), doc); e != nil {
					return report, e
				}
				result.Status, result.Deck, result.LayoutReport, result.Warnings = "go_layout_succeeded_native_review_pending", deckPath, layoutPath, layout.Warnings
			}
		}
		if err != nil {
			result.Error = err.Error()
			report.Failed++
		} else {
			report.Passed++
		}
		report.Candidates = append(report.Candidates, result)
	}
	if e = WriteJSON(filepath.Join(out, "fit-report.json"), report); e != nil {
		return report, e
	}
	return report, nil
}
