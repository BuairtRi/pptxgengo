package wmdesign

import (
	"encoding/json"
	"fmt"
	"math"
)

type Rect struct {
	X float64 `json:"x_pt"`
	Y float64 `json:"y_pt"`
	W float64 `json:"width_pt"`
	H float64 `json:"height_pt"`
}

// Span is one-based and preserves fractional geometry until serialization.
func (g Grid) Span(start, count int) (Rect, error) {
	if start < 1 || count < 1 || start+count-1 > g.Columns {
		return Rect{}, fmt.Errorf("grid.invalid_span")
	}
	return Rect{X: g.MarginX + float64(start-1)*(g.Column+g.Gutter), W: float64(count)*g.Column + float64(count-1)*g.Gutter}, nil
}
func (g Grid) FiveSpan(start, count int, rail string) (Rect, error) {
	if rail != "none" {
		return Rect{}, fmt.Errorf("grid.five_up_requires_no_rail")
	}
	if start < 1 || count < 1 || count > 2 || start+count-1 > g.FiveUp.Columns {
		return Rect{}, fmt.Errorf("grid.invalid_five_span")
	}
	return Rect{X: g.FiveUp.Starts[start-1], W: float64(count)*g.FiveUp.Column + float64(count-1)*g.FiveUp.Gutter}, nil
}
func (g Grid) OuterBox(r Rect) error {
	on := func(v, origin, step float64) bool {
		return math.Abs((v-origin)/step-math.Round((v-origin)/step)) < 1e-8
	}
	if !on(r.X, 3, g.Module) || !on(r.Y, 0, g.Module) || !on(r.W, 0, g.Module) || !on(r.H, 0, g.Module) || r.W <= 0 || r.H <= 0 {
		return fmt.Errorf("grid.off_lattice: %+v", r)
	}
	return nil
}

type NavTab struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}
type FrameRequest struct {
	Rail        string   `json:"rail"`
	Footer      string   `json:"footer"`
	Surface     string   `json:"surface"`
	RailSurface string   `json:"rail_surface"`
	TitleLines  int      `json:"title_lines"`
	Density     string   `json:"density"`
	SourceLines int      `json:"source_lines"`
	NoHeader    bool     `json:"no_header"`
	NoPage      bool     `json:"no_page"`
	Nav         []NavTab `json:"nav,omitempty"`
	Active      string   `json:"active,omitempty"`
}
type ResolvedFrame struct {
	Request    FrameRequest `json:"request"`
	Header     Rect         `json:"header"`
	Body       Rect         `json:"body"`
	Rail       Rect         `json:"rail"`
	Panel      Rect         `json:"panel"`
	Source     Rect         `json:"source"`
	FooterBand Rect         `json:"footer_band"`
	FooterRow  Rect         `json:"footer_row"`
	Whiteboard Rect         `json:"whiteboard"`
	TitleRule  float64      `json:"title_rule_pt"`
	FooterRule float64      `json:"footer_rule_pt"`
	TitleStyle string       `json:"title_style"`
}

func (s *Source) ResolveFrame(q FrameRequest) (ResolvedFrame, error) {
	if q.Rail == "" {
		q.Rail = "none"
	}
	if q.Footer == "" {
		q.Footer = "compact"
	}
	if q.Surface == "" {
		q.Surface = "light"
	}
	if q.RailSurface == "" {
		q.RailSurface = "inverse"
	}
	if q.Density == "" {
		q.Density = "standard"
	}
	if q.TitleLines == 0 {
		q.TitleLines = 1
	}
	if _, e := s.Ink(q.Surface, "bg"); e != nil {
		return ResolvedFrame{}, e
	}
	r, ok := s.Frames.Rails[q.Rail]
	if !ok {
		return ResolvedFrame{}, fmt.Errorf("frame.unknown_rail: %s", q.Rail)
	}
	foot, ok := s.Frames.Footers[q.Footer]
	if !ok {
		return ResolvedFrame{}, fmt.Errorf("frame.unknown_footer: %s", q.Footer)
	}
	if q.TitleLines < 1 || q.TitleLines > 2 || q.SourceLines < 0 || q.SourceLines > 2 {
		return ResolvedFrame{}, fmt.Errorf("frame.invalid_line_allocation")
	}
	if q.Density != "standard" && q.Density != "appendix" {
		return ResolvedFrame{}, fmt.Errorf("frame.unknown_density")
	}
	if q.Density == "appendix" && q.TitleLines != 1 {
		return ResolvedFrame{}, fmt.Errorf("frame.appendix_requires_one_title_line")
	}
	if q.Rail == "left" || q.Rail == "right" {
		valid := false
		for _, v := range r.Surfaces {
			valid = valid || v == q.RailSurface
		}
		if !valid {
			return ResolvedFrame{}, fmt.Errorf("frame.unsupported_rail_surface")
		}
	}
	if q.Rail == "nav" {
		seen := map[string]bool{}
		active := false
		for _, tab := range q.Nav {
			if tab.ID == "" || tab.Label == "" || seen[tab.ID] {
				return ResolvedFrame{}, fmt.Errorf("frame.invalid_nav")
			}
			seen[tab.ID] = true
			active = active || tab.ID == q.Active
		}
		if len(q.Nav) == 0 || !active {
			return ResolvedFrame{}, fmt.Errorf("frame.nav_requires_labels_and_active")
		}
	} else if len(q.Nav) > 0 || q.Active != "" {
		return ResolvedFrame{}, fmt.Errorf("frame.nav_without_nav_rail")
	}
	var title struct {
		One      struct{ Rule, BodyTop float64 } `json:"oneLine"`
		Two      struct{ Rule, BodyTop float64 } `json:"twoLine"`
		Appendix struct {
			Rule, BodyTop float64
			Title         string
		} `json:"appendix"`
	}
	var source struct {
		Compact, Tall struct{ Bottom float64 }
		LineHeight    float64
		BodyBottom    map[string]float64
	}
	for _, f := range s.Frames.Features {
		switch f.ID {
		case "title-zone":
			if e := json.Unmarshal(f.Geometry, &title); e != nil {
				return ResolvedFrame{}, e
			}
		case "zone.source":
			if e := json.Unmarshal(f.Geometry, &source); e != nil {
				return ResolvedFrame{}, e
			}
		}
	}
	top, rule, style := title.One.BodyTop, title.One.Rule, "title"
	if q.TitleLines == 2 {
		top, rule = title.Two.BodyTop, title.Two.Rule
	}
	if q.Density == "appendix" {
		top, rule, style = title.Appendix.BodyTop, title.Appendix.Rule, title.Appendix.Title
	}
	if q.NoHeader {
		top = s.Tokens.Grid.MarginY
		rule = 0
	}
	bottom := foot.Bottom
	if q.SourceLines > 0 {
		bottom = source.BodyBottom[fmt.Sprint(q.SourceLines)]
	}
	f := ResolvedFrame{Request: q, Header: Rect{r.Main[0], 36, r.Main[1] - r.Main[0], rule - 36}, Body: Rect{r.Main[0], top, r.Main[1] - r.Main[0], bottom - top}, TitleRule: rule, TitleStyle: style, FooterRule: foot.Rule}
	if q.NoHeader {
		f.Header = Rect{}
	}
	f.Panel = Rect{r.Panel[0], 0, r.Panel[1] - r.Panel[0], 540}
	f.Rail = Rect{r.Content[0], 36, r.Content[1] - r.Content[0], bottom - 36}
	x, w := 0.0, 960.0
	if q.Rail == "left" {
		x, w = r.Panel[1], 960-r.Panel[1]
	}
	if q.Rail == "right" {
		w = r.Panel[0]
	}
	f.FooterRow = Rect{x, foot.Row[0], w, foot.Row[1] - foot.Row[0]}
	if foot.Band[1] > 0 {
		f.FooterBand = Rect{x, foot.Band[0], w, foot.Band[1] - foot.Band[0]}
	}
	if q.SourceLines > 0 {
		b := source.Compact.Bottom
		if q.Footer == "tall" {
			b = source.Tall.Bottom
		}
		h := float64(q.SourceLines) * source.LineHeight
		f.Source = Rect{r.Main[0], b - h, r.Main[1] - r.Main[0], h}
	}
	wx := 21.0
	if q.Rail == "left" {
		wx = 291
	}
	if q.Rail == "nav" {
		wx = 57
	}
	f.Whiteboard = Rect{wx, 18, 201, 165}
	if f.Body.H <= 0 {
		return ResolvedFrame{}, fmt.Errorf("frame.empty_body")
	}
	return f, nil
}
