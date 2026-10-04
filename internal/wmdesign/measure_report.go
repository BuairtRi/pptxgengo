package wmdesign

import (
	"math"
	"strings"
)

type TextCollision struct {
	First   string `json:"first"`
	Second  string `json:"second"`
	Overlap Rect   `json:"overlap"`
}
type SlideMeasurements struct {
	ID          string           `json:"id"`
	Page        int              `json:"page"`
	TextCount   int              `json:"text_count"`
	Collisions  []TextCollision  `json:"estimated_text_collisions"`
	PanelSpills []TextPanelSpill `json:"estimated_panel_spills"`
}
type TextPanelSpill struct {
	TextID        string  `json:"text_id"`
	PanelID       string  `json:"panel_id"`
	BottomOverrun float64 `json:"bottom_overrun_pt"`
}
type MeasurementReport struct {
	Schema          string              `json:"schema"`
	Engine          string              `json:"engine"`
	SourceRevision  string              `json:"source_revision"`
	Slides          []SlideMeasurements `json:"slides"`
	CollisionCount  int                 `json:"collision_count"`
	PanelSpillCount int                 `json:"panel_spill_count"`
	Policy          []string            `json:"policy"`
}

// MeasureReport is an advisory analysis of the exact renderer's shaped line
// advances and occupied heights. It makes no native or semantic acceptance claim.
func MeasureReport(report Report) MeasurementReport {
	out := MeasurementReport{Schema: "pptxgengo.text-measurements.v1", Engine: report.Engine, SourceRevision: report.SourceRevision, Slides: []SlideMeasurements{}, Policy: []string{"Estimated occupied text intersections are review findings, not automatic layout failures.", "Rotated text is excluded; chart and table internal layouts may require separate inspection.", "Native PowerPoint and semantic design review remain required."}}
	for _, slide := range report.Slides {
		measurement := SlideMeasurements{ID: slide.ID, Page: slide.Page, TextCount: len(slide.Texts), Collisions: []TextCollision{}, PanelSpills: []TextPanelSpill{}}
		lines := make([][]Rect, len(slide.Texts))
		for i, text := range slide.Texts {
			lines[i] = estimatedTextLines(text)
			if len(lines[i]) == 0 {
				continue
			}
			anchor := lines[i][0]
			for _, shape := range slide.Shapes {
				if !strings.HasSuffix(shape.ID, ".container") {
					continue
				}
				panel := shape.Rect
				if anchor.X < panel.X || anchor.X+anchor.W > panel.X+panel.W || anchor.Y < panel.Y || anchor.Y >= panel.Y+panel.H {
					continue
				}
				overrun := 0.0
				for _, line := range lines[i] {
					overrun = math.Max(overrun, line.Y+line.H-panel.Y-panel.H)
				}
				if overrun > .5 {
					measurement.PanelSpills = append(measurement.PanelSpills, TextPanelSpill{text.ID, shape.ID, overrun})
				}
			}
		}
		for i := range slide.Texts {
			for j := i + 1; j < len(slide.Texts); j++ {
				found := false
				for _, a := range lines[i] {
					for _, b := range lines[j] {
						left, top := math.Max(a.X, b.X), math.Max(a.Y, b.Y)
						w, h := math.Min(a.X+a.W, b.X+b.W)-left, math.Min(a.Y+a.H, b.Y+b.H)-top
						if w > .5 && h > .5 {
							measurement.Collisions = append(measurement.Collisions, TextCollision{slide.Texts[i].ID, slide.Texts[j].ID, Rect{left, top, w, h}})
							found = true
							break
						}
					}
					if found {
						break
					}
				}
			}
		}
		out.CollisionCount += len(measurement.Collisions)
		out.PanelSpillCount += len(measurement.PanelSpills)
		out.Slides = append(out.Slides, measurement)
	}
	return out
}

func estimatedTextLines(text TextRecord) []Rect {
	if text.Rotation != 0 || len(text.Layout.Lines) == 0 || text.Layout.EstimatedOccupiedHeight <= 0 {
		return nil
	}
	leading := text.Layout.Style.Leading
	if leading <= 0 {
		return nil
	}
	first, last := -1, -1
	for i, line := range text.Layout.Lines {
		if strings.TrimSpace(line.Text) != "" {
			if first < 0 {
				first = i
			}
			last = i
		}
	}
	if first < 0 {
		return nil
	}
	terminal := text.Layout.EstimatedOccupiedHeight - float64(last-first)*leading
	if terminal <= 0 {
		return nil
	}
	top := text.Rect.Y
	switch text.VerticalAlign {
	case "middle", "center", "ctr":
		top += math.Max(0, text.Rect.H-text.Layout.AllocationHeight) / 2
	case "bottom", "b":
		top += math.Max(0, text.Rect.H-text.Layout.AllocationHeight)
	}
	result := []Rect{}
	for i, line := range text.Layout.Lines {
		if strings.TrimSpace(line.Text) == "" {
			continue
		}
		x := text.Rect.X
		switch text.Align {
		case "center":
			x += (text.Rect.W - line.Advance) / 2
		case "right":
			x += text.Rect.W - line.Advance
		}
		result = append(result, Rect{x, top + float64(i)*leading, line.Advance, terminal})
	}
	return result
}
