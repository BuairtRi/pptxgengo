package wmdesign

import (
	"errors"
	"strings"
	"testing"
)

func TestNavigationPreservesPreviousRenderFailure(t *testing.T) {
	want := errors.New("title did not fit")
	r := &renderer{err: want}
	r.nav(ResolvedFrame{Request: FrameRequest{Rail: "nav", Nav: []NavTab{{ID: "one", Label: "One"}}}})
	if r.err != want {
		t.Fatalf("navigation replaced the previous render failure: %v", r.err)
	}
}

func TestNavigationCannotHideTitleOverflowInBuild(t *testing.T) {
	doc, err := LibrarySourceReference(v10IntakeBundle(), "", "", 2026)
	if err != nil {
		t.Fatal(err)
	}
	for _, slide := range doc.Slides {
		if slide.TemplateBinding.Template != "roadmap-narrative/waves-criteria" {
			continue
		}
		slide.Title = strings.Repeat("This title cannot fit in its allotted header. ", 10)
		doc.Slides = []SlideSpec{slide}
		deck, _, err := BuildWithEngine(v10IntakeBundle(), "", doc, CandidateEngine)
		if err == nil || !strings.Contains(err.Error(), "text.line_allocation_exceeded: title") || len(deck) != 0 {
			t.Fatalf("title overflow was hidden: deck=%d, error=%v", len(deck), err)
		}
		return
	}
	t.Fatal("waves-and-criteria fixture is missing")
}
