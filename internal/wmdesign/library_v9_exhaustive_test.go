package wmdesign

import "testing"

// Full-catalog generation is deliberately outside the everyday short suite.
// The source includes one deprecated specimen, which active bindings omit.
func TestLibraryV9AllTemplateBuilds(t *testing.T) {
	if testing.Short() {
		t.Skip("exhaustive source and bound catalog generation")
	}
	bundle := v9IntakeBundle()
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
	if len(source.Slides) != 631 || len(bound.Slides) != 630 {
		t.Fatalf("catalog drift: %d source, %d active bound", len(source.Slides), len(bound.Slides))
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
				t.Fatal("incorrect catalog generation or native acceptance claim")
			}
		})
	}
}
