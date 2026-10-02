package adapt

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/buairtri/pptxgengo/internal/compose"
)

// Roadmap periods are ordered labels. Intervals name inclusive first and last
// period IDs; milestones pin to one period, rather than to a guessed date.
type roadmapInput struct {
	Periods     []roadmapPeriod     `json:"periods"`
	Workstreams []roadmapWorkstream `json:"workstreams"`
}
type roadmapPeriod struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}
type roadmapWorkstream struct {
	ID         string             `json:"id"`
	Label      string             `json:"label"`
	Group      string             `json:"group,omitempty"`
	Status     string             `json:"status"`
	Intervals  []roadmapInterval  `json:"intervals"`
	Milestones []roadmapMilestone `json:"milestones,omitempty"`
}
type roadmapInterval struct {
	StartPeriod string `json:"start_period"`
	EndPeriod   string `json:"end_period"`
	Status      string `json:"status"`
}
type roadmapMilestone struct {
	Period         string `json:"period"`
	Label          string `json:"label"`
	CaptionPeriods int    `json:"caption_periods,omitempty"`
}

var roadmapID = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*$`)
var roadmapStatuses = map[string]bool{"planned": true, "active": true, "complete": true, "at_risk": true, "blocked": true}

func roadmapStatusColor(style Style, status string) string {
	switch status {
	case "active":
		return style.Active
	case "complete":
		return style.Navy
	case "at_risk":
		return style.Accent
	case "blocked":
		return style.Text
	default:
		return style.Inactive
	}
}

func roadmapInk(style Style, background string) string {
	return style.Ink(background)
}

func roadmapIntervalFill(style Style, intervals []roadmapInterval, index map[string]int, period int) (string, bool) {
	for _, interval := range intervals {
		if index[interval.StartPeriod] <= period && period <= index[interval.EndPeriod] {
			return roadmapStatusColor(style, interval.Status), true
		}
	}
	return "", false
}

func roadmapText(fontFace, id, text string, r compose.Rect, fontSize float64, bold bool, fg, bg string) compose.CanvasSpec {
	return compose.CanvasSpec{ID: id, Kind: "text", Bounds: r, Text: text, FontFace: fontFace, FontSizePt: fontSize, Bold: bold, Foreground: fg, Background: bg, InsetX: 3, InsetY: 2, Align: "left", Valign: "middle"}
}

func buildRoadmap(c Context, raw json.RawMessage) (Content, error) {
	var in roadmapInput
	if err := Decode(raw, &in); err != nil {
		return Content{}, fmt.Errorf("roadmap content: %w", err)
	}
	if len(in.Periods) < 3 || len(in.Periods) > 12 {
		return Content{}, fmt.Errorf("roadmap periods must contain 3..12 entries")
	}
	if len(in.Workstreams) < 1 || len(in.Workstreams) > 8 {
		return Content{}, fmt.Errorf("roadmap workstreams must contain 1..8 entries")
	}
	periodIndex := map[string]int{}
	for i, p := range in.Periods {
		if !roadmapID.MatchString(p.ID) || strings.TrimSpace(p.Label) == "" {
			return Content{}, fmt.Errorf("roadmap period %d requires a safe id and label", i+1)
		}
		if _, found := periodIndex[p.ID]; found {
			return Content{}, fmt.Errorf("duplicate roadmap period id %q", p.ID)
		}
		periodIndex[p.ID] = i
	}
	seen := map[string]bool{}
	groups := 0
	previousGroup := ""
	for i, w := range in.Workstreams {
		if !roadmapID.MatchString(w.ID) || seen[w.ID] || strings.TrimSpace(w.Label) == "" {
			return Content{}, fmt.Errorf("roadmap workstream %d requires a unique safe id and label", i+1)
		}
		seen[w.ID] = true
		if !roadmapStatuses[w.Status] {
			return Content{}, fmt.Errorf("roadmap workstream %s has invalid status %q", w.ID, w.Status)
		}
		if len(w.Intervals) == 0 {
			return Content{}, fmt.Errorf("roadmap workstream %s requires an interval", w.ID)
		}
		if w.Group != previousGroup && w.Group != "" {
			groups++
		}
		previousGroup = w.Group
		occupied := make([]bool, len(in.Periods))
		for j, v := range w.Intervals {
			start, ok := periodIndex[v.StartPeriod]
			if !ok {
				return Content{}, fmt.Errorf("roadmap %s interval %d unknown start period %q", w.ID, j+1, v.StartPeriod)
			}
			end, ok := periodIndex[v.EndPeriod]
			if !ok {
				return Content{}, fmt.Errorf("roadmap %s interval %d unknown end period %q", w.ID, j+1, v.EndPeriod)
			}
			if end < start {
				return Content{}, fmt.Errorf("roadmap %s interval %d ends before it starts", w.ID, j+1)
			}
			if !roadmapStatuses[v.Status] {
				return Content{}, fmt.Errorf("roadmap %s interval %d has invalid status %q", w.ID, j+1, v.Status)
			}
			for n := start; n <= end; n++ {
				if occupied[n] {
					return Content{}, fmt.Errorf("roadmap %s intervals overlap at %s", w.ID, in.Periods[n].ID)
				}
				occupied[n] = true
			}
		}
		milestoneSeen := map[int]bool{}
		for j, m := range w.Milestones {
			at, ok := periodIndex[m.Period]
			if !ok {
				return Content{}, fmt.Errorf("roadmap %s milestone %d unknown period %q", w.ID, j+1, m.Period)
			}
			if strings.TrimSpace(m.Label) == "" || milestoneSeen[at] {
				return Content{}, fmt.Errorf("roadmap %s milestone %d requires a label and unique period", w.ID, j+1)
			}
			if m.CaptionPeriods < 0 || m.CaptionPeriods > 3 {
				return Content{}, fmt.Errorf("roadmap %s milestone %d caption_periods must be 1..3 when set", w.ID, j+1)
			}
			milestoneSeen[at] = true
		}
	}
	b := c.Bounds
	if !roadmapFinite(b.X) || !roadmapFinite(b.Y) || !roadmapFinite(b.Width) || !roadmapFinite(b.Height) || b.Width < 480 || b.Height < 230 {
		return Content{}, fmt.Errorf("roadmap bounds require at least 480 × 230 points")
	}
	labelW := math.Max(165, math.Min(255, b.Width*.29))
	periodW := (b.Width - labelW) / float64(len(in.Periods))
	if periodW < 40 {
		return Content{}, fmt.Errorf("roadmap period cells are narrower than 40 points; enlarge bounds or reduce periods")
	}
	headerH := 35.0
	groupH := 21.0
	rowH := math.Min(70, (b.Height-headerH-float64(groups)*groupH)/float64(len(in.Workstreams)))
	if rowH < 35 {
		return Content{}, fmt.Errorf("roadmap rows are shorter than 35 points; enlarge bounds or reduce workstreams")
	}
	content := Content{Controls: map[string]any{"period_count": len(in.Periods), "workstream_count": len(in.Workstreams), "group_count": groups, "source_vocabulary": []string{"T015 period grid and phase bands", "T039 swimlane milestones"}}, Limitations: []string{"Time labels and workstream copy require native text measurement and visual review.", "Milestone positions are tied to period IDs; this adapter does not infer dates or dependencies."}}
	add := func(v compose.CanvasSpec) { content.Canvas = append(content.Canvas, v) }
	add(roadmapText(c.Style.FontFace, "roadmap-header-label", "WORKSTREAM / PHASE", compose.Rect{X: b.X, Y: b.Y, Width: labelW, Height: headerH}, c.Style.LabelFontPt, true, c.Style.Ink(c.Style.Navy), c.Style.Navy))
	for i, p := range in.Periods {
		x := b.X + labelW + float64(i)*periodW
		bg := c.Style.Navy
		if i%2 == 1 {
			bg = c.Style.Active
		}
		add(roadmapText(c.Style.FontFace, fmt.Sprintf("roadmap-period-%02d", i), p.Label, compose.Rect{X: x, Y: b.Y, Width: periodW, Height: headerH}, c.Style.LabelFontPt, true, c.Style.Ink(bg), bg))
	}
	y := b.Y + headerH
	previousGroup = ""
	for i, w := range in.Workstreams {
		if w.Group != "" && w.Group != previousGroup {
			add(roadmapText(c.Style.FontFace, fmt.Sprintf("roadmap-group-%02d", i), w.Group, compose.Rect{X: b.X, Y: y, Width: b.Width, Height: groupH}, c.Style.LabelFontPt, true, c.Style.Ink(c.Style.Secondary), c.Style.Secondary))
			y += groupH
		}
		previousGroup = w.Group
		rowBG := c.Style.MutedSurface
		if i%2 == 1 {
			rowBG = c.Style.Surface
		}
		// Reserve the same activity lane on every row. Its 31-point ceiling
		// accommodates a measured two-line, 11-point workstream label with
		// four points of vertical inset; dense rows retain a 16-point caption
		// lane. The label, interval, and milestone marker share its centerline.
		activityH := math.Min(31, rowH-17)
		barH := math.Min(18, activityH-5)
		barTop := (activityH - barH) / 2
		labelBaseID := fmt.Sprintf("roadmap-row-%02d-base", i)
		add(compose.CanvasSpec{ID: labelBaseID, Kind: "surface", Bounds: compose.Rect{X: b.X, Y: y, Width: labelW, Height: rowH}, Background: rowBG})
		label := w.Label + "  ·  " + strings.ReplaceAll(w.Status, "_", " ")
		labelSpec := roadmapText(c.Style.FontFace, fmt.Sprintf("roadmap-row-%02d-label", i), label, compose.Rect{X: b.X, Y: y, Width: labelW, Height: activityH}, c.Style.BodyFontPt, true, roadmapInk(c.Style, rowBG), rowBG)
		labelSpec.AllowOverlap = []string{labelBaseID}
		add(labelSpec)
		barIDs := make(map[int]string, len(in.Periods))
		outlineIDs := make(map[int]string, len(in.Periods))
		for p := range in.Periods {
			x := b.X + labelW + float64(p)*periodW
			cellID := fmt.Sprintf("roadmap-cell-%02d-%02d", i, p)
			add(compose.CanvasSpec{ID: cellID, Kind: "surface", Bounds: compose.Rect{X: x, Y: y, Width: periodW, Height: rowH}, Background: rowBG})
		}
		for j, interval := range w.Intervals {
			fill := roadmapStatusColor(c.Style, interval.Status)
			if contrast(fill, rowBG) >= 3 {
				continue
			}
			start := periodIndex[interval.StartPeriod]
			end := periodIndex[interval.EndPeriod]
			outlineID := fmt.Sprintf("roadmap-bar-outline-%02d-%02d", i, j)
			outline := compose.CanvasSpec{ID: outlineID, Kind: "surface", Bounds: compose.Rect{X: b.X + labelW + float64(start)*periodW, Y: y + barTop - 1, Width: float64(end-start+1) * periodW, Height: barH + 2}, Background: c.Style.Ink(rowBG)}
			for p := start; p <= end; p++ {
				outlineIDs[p] = outlineID
				outline.AllowOverlap = append(outline.AllowOverlap, fmt.Sprintf("roadmap-cell-%02d-%02d", i, p))
			}
			add(outline)
		}
		for p := range in.Periods {
			x := b.X + labelW + float64(p)*periodW
			cellID := fmt.Sprintf("roadmap-cell-%02d-%02d", i, p)
			if fill, on := roadmapIntervalFill(c.Style, w.Intervals, periodIndex, p); on {
				barID := fmt.Sprintf("roadmap-bar-%02d-%02d", i, p)
				barIDs[p] = barID
				overlaps := []string{cellID}
				if outlineID := outlineIDs[p]; outlineID != "" {
					overlaps = append(overlaps, outlineID)
				}
				add(compose.CanvasSpec{ID: barID, Kind: "surface", Bounds: compose.Rect{X: x, Y: y + barTop, Width: periodW, Height: barH}, Background: fill, AllowOverlap: overlaps})
			}
		}
		for j, m := range w.Milestones {
			p := periodIndex[m.Period]
			x := b.X + labelW + float64(p)*periodW
			cellID := fmt.Sprintf("roadmap-cell-%02d-%02d", i, p)
			starID := fmt.Sprintf("roadmap-milestone-%02d-%02d", i, j)
			markerFill := rowBG
			starOverlap := []string{cellID}
			if fill, on := roadmapIntervalFill(c.Style, w.Intervals, periodIndex, p); on {
				markerFill = fill
				starOverlap = append(starOverlap, barIDs[p])
				if outlineID := outlineIDs[p]; outlineID != "" {
					starOverlap = append(starOverlap, outlineID)
				}
			}
			add(compose.CanvasSpec{ID: starID, Kind: "shape", Preset: "star5", Bounds: compose.Rect{X: x + periodW/2 - 6, Y: y + barTop + (barH-12)/2, Width: 12, Height: 12}, Background: c.Style.Ink(markerFill), AllowOverlap: starOverlap})
			// Captions use one period by default, directly under their marker.
			// The caller may reserve up to three periods for measured long copy.
			span := m.CaptionPeriods
			if span == 0 {
				span = 1
			}
			laneLeft := b.X + labelW
			laneRight := b.X + b.Width
			for k, other := range w.Milestones {
				if k == j {
					continue
				}
				otherP := periodIndex[other.Period]
				mid := b.X + labelW + (float64(p+otherP+1)/2)*periodW
				if otherP < p && mid > laneLeft {
					laneLeft = mid
				}
				if otherP > p && mid < laneRight {
					laneRight = mid
				}
			}
			captionW := math.Min(400, math.Min(float64(span)*periodW-4, laneRight-laneLeft-4))
			idealX := x + periodW/2 - captionW/2
			minX := laneLeft + 2
			maxX := laneRight - captionW - 2
			captionX := math.Max(minX, math.Min(idealX, maxX))
			captionAlign := "center"
			if span > 1 && idealX < minX-0.01 {
				captionAlign = "left"
			} else if span > 1 && idealX > maxX+0.01 {
				captionAlign = "right"
			}
			labelH := math.Min(42, rowH-activityH-1)
			labelID := fmt.Sprintf("roadmap-milestone-label-%02d-%02d", i, j)
			caption := roadmapText(c.Style.FontFace, labelID, m.Label, compose.Rect{X: captionX, Y: y + rowH - labelH - 1, Width: captionW, Height: labelH}, math.Min(c.Style.LabelFontPt, 10), false, roadmapInk(c.Style, rowBG), rowBG)
			caption.Align = captionAlign
			for period := range in.Periods {
				cellX := b.X + labelW + float64(period)*periodW
				if captionX < cellX+periodW && captionX+captionW > cellX {
					caption.AllowOverlap = append(caption.AllowOverlap, fmt.Sprintf("roadmap-cell-%02d-%02d", i, period))
				}
			}
			add(caption)
		}
		y += rowH
	}
	return content, nil
}

func roadmapFinite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
