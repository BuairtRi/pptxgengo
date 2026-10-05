package wmdesign

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildHeaderErrorIdentifiesStableSlide(t *testing.T) {
	doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{
		{ID: "first-valid", Frame: FrameRequest{Rail: "none", Footer: "compact", TitleLines: 1}, Title: "A clear headline"},
		{ID: "decision-source-022", Frame: FrameRequest{Rail: "none", Footer: "compact", TitleLines: 1}, Title: strings.Repeat("Modernization strategy and outcomes ", 12)},
	}}
	_, _, err := BuildWithEngine(filepath.Join("..", "..", "planning", "wm-design-contracts", "v5", "intake-20261003-587-frozen", "bundle"), "", doc, CandidateEngine)
	if err == nil || !strings.Contains(err.Error(), "slide decision-source-022:") || !strings.Contains(err.Error(), "text.line_allocation_exceeded") {
		t.Fatalf("missing stable slide context: %v", err)
	}
}
