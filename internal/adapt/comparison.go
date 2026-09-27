package adapt

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/buairtri/pptxgengo/internal/compose"
)

type comparisonInput struct {
	GaugeMode string                `json:"gauge_mode,omitempty"`
	Options   []comparisonOption    `json:"options"`
	Criteria  []comparisonCriterion `json:"criteria"`
}
type comparisonOption struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}
type comparisonCriterion struct {
	ID       string              `json:"id"`
	Label    string              `json:"label"`
	Gauge    comparisonGauge     `json:"gauge"`
	Findings []comparisonFinding `json:"findings"`
}
type comparisonGauge struct {
	Value     *float64 `json:"value"`
	ScaleMin  *float64 `json:"scale_min"`
	ScaleMax  *float64 `json:"scale_max"`
	TargetMin *float64 `json:"target_min"`
	TargetMax *float64 `json:"target_max"`
	Unit      string   `json:"unit,omitempty"`
}
type comparisonFinding struct {
	OptionID string `json:"option_id"`
	Text     string `json:"text"`
	Status   string `json:"status"`
}

var comparisonStatuses = map[string]bool{"strong": true, "mixed": true, "limited": true, "unknown": true}

func comparisonNumber(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }
func comparisonFinite(v float64) bool   { return !math.IsNaN(v) && !math.IsInf(v, 0) }

func comparisonScaleLabel(g comparisonGauge) string {
	return "Scale " + comparisonNumber(*g.ScaleMin) + "–" + comparisonNumber(*g.ScaleMax) + g.Unit
}

func comparisonTargetLabel(g comparisonGauge) string {
	return "Target " + comparisonNumber(*g.TargetMin) + "–" + comparisonNumber(*g.TargetMax) + g.Unit
}

func comparisonNote(id, value string, bounds compose.Rect, ink, surface string) compose.CanvasSpec {
	return compose.CanvasSpec{ID: id, Kind: "text", Bounds: bounds, Text: value, FontFace: "Arial", FontSizePt: 9, Foreground: ink, Background: surface, Align: "left", Valign: "middle"}
}

func comparisonFindingColors(style Style, status string) (string, string) {
	switch status {
	case "strong":
		return roadmapInk(style, style.Active), style.Active
	case "mixed":
		return roadmapInk(style, style.Surface), style.Surface
	case "limited":
		return roadmapInk(style, style.MutedSurface), style.MutedSurface
	default:
		return roadmapInk(style, style.Inactive), style.Inactive
	}
}

func buildComparison(c Context, raw json.RawMessage) (Content, error) {
	var in comparisonInput
	if err := Decode(raw, &in); err != nil {
		return Content{}, fmt.Errorf("comparison content: %w", err)
	}
	if len(in.Options) < 2 || len(in.Options) > 4 {
		return Content{}, fmt.Errorf("comparison options must contain 2..4 entries")
	}
	if len(in.Criteria) < 1 || len(in.Criteria) > 7 {
		return Content{}, fmt.Errorf("comparison criteria must contain 1..7 entries")
	}
	if in.GaugeMode != "" && in.GaugeMode != "linear" && in.GaugeMode != "dial" {
		return Content{}, fmt.Errorf("comparison gauge_mode must be linear or dial")
	}
	optionIndex := map[string]int{}
	for i, o := range in.Options {
		if !roadmapID.MatchString(o.ID) || strings.TrimSpace(o.Label) == "" {
			return Content{}, fmt.Errorf("comparison option %d requires a safe id and label", i+1)
		}
		if _, found := optionIndex[o.ID]; found {
			return Content{}, fmt.Errorf("duplicate comparison option id %q", o.ID)
		}
		optionIndex[o.ID] = i
	}
	criteriaSeen := map[string]bool{}
	for i, r := range in.Criteria {
		if !roadmapID.MatchString(r.ID) || criteriaSeen[r.ID] || strings.TrimSpace(r.Label) == "" {
			return Content{}, fmt.Errorf("comparison criterion %d requires a unique safe id and label", i+1)
		}
		criteriaSeen[r.ID] = true
		g := r.Gauge
		if g.Value == nil || g.ScaleMin == nil || g.ScaleMax == nil || g.TargetMin == nil || g.TargetMax == nil {
			return Content{}, fmt.Errorf("comparison criterion %s gauge requires value, scale_min/max and target_min/max", r.ID)
		}
		if !comparisonFinite(*g.Value) || !comparisonFinite(*g.ScaleMin) || !comparisonFinite(*g.ScaleMax) || !comparisonFinite(*g.TargetMin) || !comparisonFinite(*g.TargetMax) || *g.ScaleMin >= *g.ScaleMax || *g.Value < *g.ScaleMin || *g.Value > *g.ScaleMax || *g.TargetMin < *g.ScaleMin || *g.TargetMax > *g.ScaleMax || *g.TargetMin > *g.TargetMax {
			return Content{}, fmt.Errorf("comparison criterion %s has invalid gauge values or target range", r.ID)
		}
		if strings.TrimSpace(g.Unit) != g.Unit {
			return Content{}, fmt.Errorf("comparison criterion %s gauge unit has surrounding whitespace", r.ID)
		}
		if len(r.Findings) != len(in.Options) {
			return Content{}, fmt.Errorf("comparison criterion %s needs one finding per option", r.ID)
		}
		seen := map[string]bool{}
		for j, f := range r.Findings {
			if _, ok := optionIndex[f.OptionID]; !ok {
				return Content{}, fmt.Errorf("comparison criterion %s finding %d references unknown option %q", r.ID, j+1, f.OptionID)
			}
			if seen[f.OptionID] {
				return Content{}, fmt.Errorf("comparison criterion %s repeats option %q", r.ID, f.OptionID)
			}
			seen[f.OptionID] = true
			if strings.TrimSpace(f.Text) == "" || !comparisonStatuses[f.Status] {
				return Content{}, fmt.Errorf("comparison criterion %s option %s needs text and status strong|mixed|limited|unknown", r.ID, f.OptionID)
			}
		}
	}
	b := c.Bounds
	if !roadmapFinite(b.X) || !roadmapFinite(b.Y) || !roadmapFinite(b.Width) || !roadmapFinite(b.Height) || b.Width < 500 || b.Height < 230 {
		return Content{}, fmt.Errorf("comparison bounds require at least 500 × 230 points")
	}
	labelW := math.Max(175, math.Min(255, b.Width*.28))
	optionW := (b.Width - labelW) / float64(len(in.Options))
	if optionW < 115 {
		return Content{}, fmt.Errorf("comparison option cells are narrower than 115 points; enlarge bounds or reduce options")
	}
	headerH := 35.0
	rowH := math.Min(95, (b.Height-headerH)/float64(len(in.Criteria)))
	if rowH < 45 {
		return Content{}, fmt.Errorf("comparison rows are shorter than 45 points needed for gauge scale and target labels; enlarge bounds or reduce criteria")
	}
	if in.GaugeMode == "dial" && rowH < 58 {
		return Content{}, fmt.Errorf("comparison dial rows need at least 58 points; enlarge bounds or reduce criteria")
	}
	mode := in.GaugeMode
	if mode == "" {
		mode = "linear"
	}
	content := Content{Controls: map[string]any{"criterion_count": len(in.Criteria), "option_count": len(in.Options), "gauge_mode": mode, "gauge_policy": "Each needle and any nonzero target band are positioned from caller supplied numeric value and scale.", "source_vocabulary": "T045 five-row comparison, pale cells and left-side gauge"}, Limitations: []string{"Gauge geometry represents the supplied values; this adapter does not invent or score evidence.", "The 9-point gauge labels and qualitative finding text require native measurement and visual review."}}
	add := func(v compose.CanvasSpec) { content.Canvas = append(content.Canvas, v) }
	add(roadmapText("comparison-header-criterion", "CRITERION / ACTUAL VALUE", compose.Rect{X: b.X, Y: b.Y, Width: labelW, Height: headerH}, c.Style.LabelFontPt, true, c.Style.Ink(c.Style.Navy), c.Style.Navy))
	for i, o := range in.Options {
		x := b.X + labelW + float64(i)*optionW
		add(roadmapText(fmt.Sprintf("comparison-header-option-%02d", i), o.Label, compose.Rect{X: x, Y: b.Y, Width: optionW, Height: headerH}, c.Style.LabelFontPt, true, c.Style.Ink(c.Style.Navy), c.Style.Navy))
	}
	for i, r := range in.Criteria {
		y := b.Y + headerH + float64(i)*rowH
		rowBG := c.Style.MutedSurface
		if i%2 == 1 {
			rowBG = c.Style.Surface
		}
		g := r.Gauge
		valueLabel := comparisonNumber(*g.Value) + g.Unit
		label := r.Label + "  ·  " + valueLabel
		labelH := math.Min(25, rowH*.52)
		add(roadmapText(fmt.Sprintf("comparison-criterion-%02d", i), label, compose.Rect{X: b.X, Y: y, Width: labelW, Height: labelH}, c.Style.BodyFontPt, true, roadmapInk(c.Style, rowBG), rowBG))
		if mode == "dial" {
			paneID := fmt.Sprintf("comparison-dial-pane-%02d", i)
			add(compose.CanvasSpec{ID: paneID, Kind: "surface", Bounds: compose.Rect{X: b.X, Y: y + labelH, Width: labelW, Height: rowH - labelH}, Background: rowBG})
			dialH := rowH - labelH - 4
			dialW := math.Min(labelW*.42, dialH*2)
			dialX := b.X + 8
			dialY := y + labelH + 2
			arcBounds := compose.Rect{X: dialX, Y: dialY, Width: dialW, Height: dialH}
			baseID := fmt.Sprintf("comparison-dial-base-%02d", i)
			add(compose.CanvasSpec{ID: baseID, Kind: "shape", Preset: "blockArc", Bounds: arcBounds, Background: c.Style.Inactive, ArcStartDeg: 180, ArcEndDeg: 360, ArcThicknessRatio: .38, AllowOverlap: []string{paneID}})
			scale := *g.ScaleMax - *g.ScaleMin
			targetStart := 180 + ((*g.TargetMin-*g.ScaleMin)/scale)*180
			targetEnd := 180 + ((*g.TargetMax-*g.ScaleMin)/scale)*180
			targetID := fmt.Sprintf("comparison-dial-target-%02d", i)
			if *g.TargetMin < *g.TargetMax {
				add(compose.CanvasSpec{ID: targetID, Kind: "shape", Preset: "blockArc", Bounds: arcBounds, Background: c.Style.Active, ArcStartDeg: targetStart, ArcEndDeg: targetEnd, ArcThicknessRatio: .38, AllowOverlap: []string{paneID, baseID}})
			}
			angle := math.Round((180+((*g.Value-*g.ScaleMin)/scale)*180)*1000) / 1000
			rad := angle * math.Pi / 180
			pivotX, pivotY := dialX+dialW/2, dialY+dialH/2
			needleH := math.Min(15, math.Max(10, dialH*.4))
			needleX := pivotX + math.Cos(rad)*dialW*.275
			needleY := pivotY + math.Sin(rad)*dialH*.275
			needleOverlap := []string{paneID, baseID}
			if *g.TargetMin < *g.TargetMax {
				needleOverlap = append(needleOverlap, targetID)
			}
			add(compose.CanvasSpec{ID: fmt.Sprintf("comparison-dial-actual-%02d", i), Kind: "shape", Preset: "triangle", Bounds: compose.Rect{X: needleX - 4, Y: needleY - needleH/2, Width: 8, Height: needleH}, Background: c.Style.Ink(rowBG), RotationDeg: angle - 270, AllowOverlap: needleOverlap})
			noteX := dialX + dialW + 12
			noteW := b.X + labelW - noteX - 5
			scaleNote := comparisonNote(fmt.Sprintf("comparison-scale-%02d", i), comparisonScaleLabel(g), compose.Rect{X: noteX, Y: y + labelH + 5, Width: noteW, Height: 12}, c.Style.Ink(rowBG), rowBG)
			scaleNote.AllowOverlap = []string{paneID}
			add(scaleNote)
			targetNote := comparisonNote(fmt.Sprintf("comparison-target-%02d", i), comparisonTargetLabel(g), compose.Rect{X: noteX, Y: y + labelH + 19, Width: noteW, Height: 12}, c.Style.Ink(rowBG), rowBG)
			targetNote.AllowOverlap = []string{paneID}
			add(targetNote)
		} else {
			paneID := fmt.Sprintf("comparison-linear-pane-%02d", i)
			add(compose.CanvasSpec{ID: paneID, Kind: "surface", Bounds: compose.Rect{X: b.X, Y: y + labelH, Width: labelW, Height: rowH - labelH}, Background: rowBG})
			trackX := b.X + 8
			trackW := labelW - 16
			trackY := y + labelH + 2
			trackH := math.Min(11, rowH-labelH-16)
			if trackH < 5 {
				return Content{}, fmt.Errorf("comparison criterion %s has no gauge room", r.ID)
			}
			baseID := fmt.Sprintf("comparison-gauge-base-%02d", i)
			add(compose.CanvasSpec{ID: baseID, Kind: "surface", Bounds: compose.Rect{X: trackX, Y: trackY, Width: trackW, Height: trackH}, Background: c.Style.Inactive, AllowOverlap: []string{paneID}})
			scale := *g.ScaleMax - *g.ScaleMin
			targetX := trackX + ((*g.TargetMin-*g.ScaleMin)/scale)*trackW
			targetW := ((*g.TargetMax - *g.TargetMin) / scale) * trackW
			if targetW > 0 {
				add(compose.CanvasSpec{ID: fmt.Sprintf("comparison-gauge-target-%02d", i), Kind: "surface", Bounds: compose.Rect{X: targetX, Y: trackY, Width: targetW, Height: trackH}, Background: c.Style.Active, AllowOverlap: []string{paneID, baseID}})
			}
			needleX := trackX + ((*g.Value-*g.ScaleMin)/scale)*trackW
			if needleX > trackX+trackW-2 {
				needleX = trackX + trackW - 2
			}
			needleGround := c.Style.Inactive
			if *g.Value >= *g.TargetMin && *g.Value <= *g.TargetMax {
				needleGround = c.Style.Active
			}
			// One-point end caps keep the actual marker distinct from the track
			// without entering the criterion or numeric-label text lanes.
			needle := compose.CanvasSpec{ID: fmt.Sprintf("comparison-gauge-actual-%02d", i), Kind: "shape", Preset: "rect", Bounds: compose.Rect{X: needleX, Y: trackY - 1, Width: 2, Height: trackH + 2}, Background: c.Style.Ink(needleGround), AllowOverlap: []string{paneID, baseID}}
			if targetW > 0 {
				needle.AllowOverlap = append(needle.AllowOverlap, fmt.Sprintf("comparison-gauge-target-%02d", i))
			}
			add(needle)
			scaleNote := comparisonNote(fmt.Sprintf("comparison-scale-target-%02d", i), comparisonScaleLabel(g)+"  ·  "+comparisonTargetLabel(g), compose.Rect{X: trackX, Y: trackY + trackH + 2, Width: trackW, Height: 12}, c.Style.Ink(rowBG), rowBG)
			scaleNote.AllowOverlap = []string{paneID}
			add(scaleNote)
		}
		findings := make(map[string]comparisonFinding, len(r.Findings))
		for _, f := range r.Findings {
			findings[f.OptionID] = f
		}
		for j, o := range in.Options {
			f := findings[o.ID]
			fg, bg := comparisonFindingColors(c.Style, f.Status)
			text := f.Text
			add(roadmapText(fmt.Sprintf("comparison-finding-%02d-%02d", i, j), text, compose.Rect{X: b.X + labelW + float64(j)*optionW, Y: y, Width: optionW, Height: rowH}, c.Style.BodyFontPt, false, fg, bg))
		}
	}
	legendY := b.Y + headerH + float64(len(in.Criteria))*rowH + 2
	legendX := b.X + labelW
	legendW := (b.Width - labelW) / 4
	for j, status := range []string{"strong", "mixed", "limited", "unknown"} {
		_, swatch := comparisonFindingColors(c.Style, status)
		x := legendX + float64(j)*legendW
		borderID := fmt.Sprintf("comparison-legend-border-%02d", j)
		add(compose.CanvasSpec{ID: borderID, Kind: "surface", Bounds: compose.Rect{X: x, Y: legendY + 1, Width: 10, Height: 10}, Background: c.Style.Navy})
		add(compose.CanvasSpec{ID: fmt.Sprintf("comparison-legend-swatch-%02d", j), Kind: "surface", Bounds: compose.Rect{X: x + 1, Y: legendY + 2, Width: 8, Height: 8}, Background: swatch, AllowOverlap: []string{borderID}})
		add(comparisonNote(fmt.Sprintf("comparison-legend-label-%02d", j), strings.ToUpper(status[:1])+status[1:], compose.Rect{X: x + 14, Y: legendY, Width: legendW - 14, Height: 12}, c.Style.Ink(c.Style.White), c.Style.White))
	}
	return content, nil
}
