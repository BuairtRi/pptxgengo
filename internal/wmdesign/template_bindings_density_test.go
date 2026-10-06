package wmdesign

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBoundSlideDensityFieldsValidate(t *testing.T) {
	auto := false
	base := BoundDocument{Schema: BoundDocumentSchema, Year: 2026, Slides: []BoundSlide{{ID: "density-slide", Template: "cards/3", ContentKind: "supplied_content", Density: "compact", HeaderDensity: "dense", AutoDensity: &auto, Values: json.RawMessage(`{}`)}}}
	if err := validateBoundDocument(base); err != nil {
		t.Fatalf("valid density fields rejected: %v", err)
	}
	for _, tc := range []struct {
		field string
		set   func(*BoundSlide)
		want  string
	}{
		{"density", func(s *BoundSlide) { s.Density = "tiny" }, "binding.invalid_density"},
		{"header_density", func(s *BoundSlide) { s.HeaderDensity = "tiny" }, "binding.invalid_header_density"},
	} {
		t.Run(tc.field, func(t *testing.T) {
			invalid := base
			invalid.Slides = append([]BoundSlide(nil), base.Slides...)
			tc.set(&invalid.Slides[0])
			if err := validateBoundDocument(invalid); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("wanted %s, got %v", tc.want, err)
			}
		})
	}
}
