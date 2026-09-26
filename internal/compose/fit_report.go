package compose

import (
	"math"
	"strings"
)

// FixedTextFitReport reports every fixed text zone before the planner's first
// failure. Dynamic pod/legend geometry, collisions and routes still require Plan.
// Measurements must be validated against the full probe contract by the caller.
type TextZoneFit struct {
	SlideID           string  `json:"slide_id"`
	ElementID         string  `json:"element_id"`
	RequestID         string  `json:"request_id"`
	AvailableWidthPt  float64 `json:"available_width_pt"`
	AvailableHeightPt float64 `json:"available_height_pt"`
	MeasuredWidthPt   float64 `json:"measured_width_pt"`
	MeasuredHeightPt  float64 `json:"measured_height_pt"`
	OverflowWidthPt   float64 `json:"overflow_width_pt"`
	OverflowHeightPt  float64 `json:"overflow_height_pt"`
	Fits              bool    `json:"fits"`
	TolerancePt       float64 `json:"tolerance_pt"`
}

func FixedTextFitReport(spec Spec, measured Measurements) []TextZoneFit {
	var rows []TextZoneFit
	add := func(slide, id, request string, width, height, tolerance float64) {
		m, ok := measured.ByRequestID[request]
		dw, dh := math.Max(0, m.RenderedWidthPt-width), math.Max(0, m.RenderedHeightPt-height)
		rows = append(rows, TextZoneFit{slide, id, request, width, height, m.RenderedWidthPt, m.RenderedHeightPt, dw, dh, ok && dw <= tolerance && dh <= tolerance, tolerance})
	}
	for _, s := range spec.Slides {
		if strings.TrimSpace(s.Title) != "" {
			add(s.ID, "slide-title", requestID(s.ID, "slide_title"), s.TitleBounds.Width, s.TitleBounds.Height, 1e-6)
		}
		for _, c := range s.Canvas {
			if c.Kind == "text" {
				add(s.ID, c.ID, requestID(s.ID, "canvas", c.ID), c.Bounds.Width-2*c.InsetX, c.Bounds.Height-2*c.InsetY, .01)
			}
		}
		for _, r := range s.Roles {
			add(s.ID, r.ID, requestID(s.ID, "standalone_role", r.ID), r.Bounds.Width-2*r.HorizontalInsetPt, r.Bounds.Height-2*r.VerticalInsetPt, 1e-6)
		}
		for _, c := range s.Cards {
			for _, field := range cardFields(c) {
				r := cardBlockBounds(c, field.role)
				_, _, _, ix, iy := cardStyle(field.role)
				add(s.ID, c.ID+"/"+field.role, requestID(s.ID, "card", c.ID, field.role), r.Width-2*ix, r.Height-2*iy, 1e-6)
			}
		}
	}
	return rows
}
