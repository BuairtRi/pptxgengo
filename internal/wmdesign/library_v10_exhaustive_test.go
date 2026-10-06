package wmdesign

import "testing"

// Full-catalog generation is deliberately outside the everyday short suite.
// The source includes one deprecated specimen, which active bindings omit.
func TestLibraryV10AllTemplateBuilds(t *testing.T) {
	if testing.Short() {
		t.Skip("exhaustive source and bound catalog generation")
	}
	bundle := v10IntakeBundle()
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
	if len(source.Slides) != 649 || len(bound.Slides) != 648 {
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

// Source-file and family metadata moves must not change the rendered dependency
// closure used to inherit the accepted 631 specimens from v9.
func TestLibraryV10RetainedPublicationRenderInheritance(t *testing.T) {
	if testing.Short() {
		t.Skip("paired 631-specimen visible package comparison")
	}
	previous, err := LibraryCatalog(v9IntakeBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	keys := make([]string, 0, len(previous))
	for _, def := range previous {
		keys = append(keys, def.Key)
	}
	if len(keys) != 631 {
		t.Fatal("previous catalog drift")
	}
	count, err := VerifyLibraryPublicationRenderInheritance(LibraryPublicationOptions{
		Bundle: v10IntakeBundle(), PreviousBundle: v9IntakeBundle(), Year: 2026,
	}, keys)
	if err != nil {
		t.Fatal(err)
	}
	if count != 631 {
		t.Fatalf("inheritance drift: %d", count)
	}
}

// Check specimens individually so a chrome error does not conceal later
// failures in a full-catalog deck.
func TestLibraryV10RetainedIndividualSourceBuilds(t *testing.T) {
	if testing.Short() {
		t.Skip("individual generation of all retained 631 specimens")
	}
	source, err := LibrarySourceReference(v10IntakeBundle(), "", "", 2026)
	if err != nil {
		t.Fatal(err)
	}
	added := map[string]bool{}
	for _, key := range v10AddedTemplateKeys {
		added[key] = true
	}
	count := 0
	for _, slide := range source.Slides {
		if added[slide.TemplateBinding.Template] {
			continue
		}
		count++
		t.Run(slide.TemplateBinding.Template, func(t *testing.T) {
			doc := source
			doc.Slides = []SlideSpec{slide}
			if _, _, err := BuildWithEngine(v10IntakeBundle(), "", doc, CandidateEngine); err != nil {
				t.Fatal(err)
			}
		})
	}
	if count != 631 {
		t.Fatalf("retained catalog drift: %d", count)
	}
}
