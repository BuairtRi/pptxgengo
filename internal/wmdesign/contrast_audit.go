package wmdesign

import (
	"fmt"
	"math"
	"strings"
)

// contrastProbe is a private, non-serializing audit path. It preserves fonts,
// role sizes, weights, inks and surfaces while excluding wrapping/capacity from
// contrast assessment. No build constructor enables it.
type contrastProbe struct {
	Checks         []contrastCheck
	SeenText       map[string]bool
	SuppressChecks bool
	CoverageErrors []string
}
type contrastCheck struct {
	ID         string  `json:"id"`
	Role       string  `json:"role"`
	Size       float64 `json:"size_pt"`
	Weight     int     `json:"weight"`
	Ink        string  `json:"ink"`
	Background string  `json:"background"`
	Ratio      float64 `json:"ratio"`
	Minimum    float64 `json:"minimum"`
	Passed     bool    `json:"passed"`
}

// Verify actual emitted text objects, including individual rich runs, rather
// than assuming a successful component planner checked all of its copy.
func (r *renderer) auditPlanContrastCoverage(p *scenePlan) {
	if p == nil || r.contrastProbe == nil {
		return
	}
	check := func(tr TextRecord) {
		covered := func(id string) bool {
			for _, c := range r.contrastProbe.Checks {
				if c.ID == id || strings.HasPrefix(id, c.ID+".part-") {
					return true
				}
			}
			return false
		}
		if !covered(tr.ID) {
			// Some bounded diagram labels are assembled directly rather than
			// through sceneText. Audit their final paint color and font too.
			r.auditLiteralContrast(p, tr.ID, tr.Layout.Style, tr.Rect, tr.Color)
		}
		if tr.Rich != nil {
			for _, paragraph := range tr.Rich.Paragraphs {
				for _, run := range paragraph.Runs {
					if strings.TrimSpace(run.Original) == "" {
						continue
					} // layout spacers
					matched := false
					for _, c := range r.contrastProbe.Checks {
						if strings.HasPrefix(c.ID, tr.ID+"/") && c.Size == run.Style.Size && c.Weight == run.Style.Weight && c.Ink == run.Color {
							matched = true
							break
						}
					}
					if !matched {
						r.auditLiteralContrast(p, tr.ID+"/"+paragraph.Key+"/"+run.Key, run.Style, tr.Rect, run.Color)
					}
				}
			}
		}
		if !covered(tr.ID) {
			r.contrastProbe.CoverageErrors = append(r.contrastProbe.CoverageErrors, tr.ID+": emitted text not contrast-checked")
		}
	}
	for _, it := range p.Items {
		if it.Text != nil {
			check(*it.Text)
		}
		if it.Table != nil {
			for _, tr := range it.Table.Texts {
				check(tr)
			}
		}
		if it.Chart != nil {
			r.auditChartContrast(it.Chart)
		}
	}
}

func (r *renderer) auditChartContrast(chart *sceneChart) {
	if r.contrastProbe == nil {
		return
	}
	o := chart.Options
	bg, err := r.sceneColor(r.sceneContext.Surface, "bg")
	if err != nil {
		r.contrastProbe.CoverageErrors = append(r.contrastProbe.CoverageErrors, chart.ID+": "+err.Error())
		return
	}
	if o.ChartArea != nil && o.ChartArea.Fill != nil && o.ChartArea.Fill.Type != "none" && o.ChartArea.Fill.Color != "" {
		bg = o.ChartArea.Fill.Color
	}
	plotBG := bg
	if o.PlotArea != nil && o.PlotArea.Fill != nil && o.PlotArea.Fill.Type != "none" && o.PlotArea.Fill.Color != "" {
		plotBG = o.PlotArea.Fill.Color
	}
	on := func(p *bool) bool { return p != nil && *p }
	hidden := func(p *bool) bool { return p != nil && *p }
	check := func(role string, size float64, bold *bool, ink, background string) {
		if ink == "" {
			ink = o.Color
		}
		if size <= 0 || ink == "" || background == "" {
			r.contrastProbe.CoverageErrors = append(r.contrastProbe.CoverageErrors, chart.ID+"/"+role+": unspecified native chart typography/color")
			return
		}
		weight := 400
		if on(bold) {
			weight = 700
		}
		st := Style{ID: "native-chart." + role, Size: size, Weight: weight}
		r.contrastAllows(chart.ID+"/"+role, st, ink, background, textContrastMinimum(st))
	}
	if !hidden(o.CatAxisHidden) {
		check("category-axis", o.CatAxisLabelFontSize, o.CatAxisLabelFontBold, o.CatAxisLabelColor, bg)
	}
	if !hidden(o.ValAxisHidden) {
		check("value-axis", o.ValAxisLabelFontSize, o.ValAxisLabelFontBold, o.ValAxisLabelColor, bg)
	}
	if on(o.ShowLegend) {
		check("legend", o.LegendFontSize, nil, o.LegendColor, bg)
	}
	if on(o.ShowTitle) {
		check("title", o.TitleFontSize, o.TitleBold, o.TitleColor, bg)
	}
	if on(o.ShowValue) || on(o.ShowLabel) || on(o.ShowPercent) || on(o.ShowSerName) {
		switch o.DataLabelPosition {
		case "ctr", "inBase", "inEnd":
			if len(o.ChartColors) == 0 {
				r.contrastProbe.CoverageErrors = append(r.contrastProbe.CoverageErrors, chart.ID+": in-bar data label background missing")
			}
			for i, c := range o.ChartColors {
				check(fmt.Sprintf("data-label.series-%03d", i+1), o.DataLabelFontSize, o.DataLabelFontBold, o.DataLabelColor, c)
			}
		default:
			check("data-label", o.DataLabelFontSize, o.DataLabelFontBold, o.DataLabelColor, plotBG)
		}
	}
}

func textContrastMinimum(st Style) float64 {
	if st.Size >= 18 || st.Size >= 14 && st.Weight >= 700 {
		return 3
	}
	return 4.5
}
func (r *renderer) contrastAllows(id string, st Style, ink, bg string, minimum float64) bool {
	ratio := contrast(ink, bg)
	passed := ratio >= minimum
	if r.contrastProbe != nil {
		if r.contrastProbe.SuppressChecks {
			return true
		}
		r.contrastProbe.Checks = append(r.contrastProbe.Checks, contrastCheck{id, st.ID, st.Size, st.Weight, ink, bg, ratio, minimum, passed})
		return true // collect every contrast failure rather than stop at the first
	}
	return passed
}
func (r *renderer) auditTextContrast(id string, st Style, ink, surface string) {
	if r.contrastProbe == nil {
		return
	}
	bg, err := r.sceneColor(surface, "bg")
	if err != nil {
		r.contrastProbe.CoverageErrors = append(r.contrastProbe.CoverageErrors, fmt.Sprintf("%s: %v", id, err))
		return
	}
	r.contrastAllows(id, st, ink, bg, textContrastMinimum(st))
}
func (r *renderer) measureText(text string, st Style, width float64) (TextLayout, error) {
	if r.contrastProbe == nil {
		return r.typeEngine.Measure(text, st, width)
	}
	if r.contrastProbe.SeenText == nil {
		r.contrastProbe.SeenText = map[string]bool{}
	}
	r.contrastProbe.SeenText[text] = true
	if width <= 0 || st.Size <= 0 || st.Leading <= 0 || st.Size > 4096 || st.Leading > 8191 || math.Abs(st.TrackingPt) > 4096 || width > 8191 || !intakeFinite(width, st.Size, st.Leading, st.TrackingPt) {
		return TextLayout{}, fmt.Errorf("contrast.invalid_geometry")
	}
	font, err := r.typeEngine.Resolve(st)
	if err != nil {
		return TextLayout{}, err
	}
	display := strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
	if st.Case == "upper" {
		display = strings.ToUpper(display)
	}
	return TextLayout{Original: text, Displayed: display, Style: st, Font: font, Lines: []TextLine{{Text: display, Advance: math.Min(.001, width/2)}}, AllocationHeight: .001, EstimatedOccupiedHeight: .001, VerticalPolicy: "contrast-only probe; capacity deliberately not assessed"}, nil
}

func (r *renderer) auditLiteralContrast(p *scenePlan, id string, st Style, b Rect, color string) {
	if r.contrastProbe == nil {
		return
	}
	bg, err := r.sceneColor(r.sceneContext.Surface, "bg")
	if err != nil {
		r.contrastProbe.CoverageErrors = append(r.contrastProbe.CoverageErrors, fmt.Sprintf("%s: %v", id, err))
		return
	}
	x, y := b.X+b.W/2, b.Y+b.H/2
	for i := len(p.Items) - 1; i >= 0; i-- {
		shape := p.Items[i].Shape
		if shape == nil || shape.Props.Fill == nil || shape.Props.Fill.Type == "none" || shape.Record.Color == "" {
			continue
		}
		sb := shape.Record.Rect
		if x >= sb.X && x <= sb.X+sb.W && y >= sb.Y && y <= sb.Y+sb.H {
			bg = shape.Record.Color
			break
		}
	}
	r.contrastAllows(id, st, color, bg, textContrastMinimum(st))
}
func (r *renderer) measureRichParagraph(p RichParagraphLayout, width float64) ([]TextLine, error) {
	if r.contrastProbe == nil {
		return r.typeEngine.richParagraphLines(p, width)
	}
	if r.contrastProbe.SeenText == nil {
		r.contrastProbe.SeenText = map[string]bool{}
	}
	r.contrastProbe.SeenText[p.Displayed] = true
	if width <= 0 || !intakeFinite(width) {
		return nil, fmt.Errorf("contrast.invalid_geometry")
	}
	return []TextLine{{Text: p.Displayed, Advance: math.Min(.001, width/2)}}, nil
}
