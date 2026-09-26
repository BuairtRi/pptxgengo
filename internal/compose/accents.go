package compose

import "fmt"

// AccentSpec attaches artwork to a complete, single-line canvas text block.
// Alpha bounds are normalized source-image coordinates, not the image frame.
// Phrase fragments inside rich text still use pptxanchor's native phrase adapter.
type AccentSpec struct {
	ID             string  `json:"id"`
	Target         string  `json:"target"`
	Mode           string  `json:"mode"`
	AssetPath      string  `json:"asset_path"`
	AssetSHA256    string  `json:"asset_sha256"`
	AlphaBounds    Rect    `json:"alpha_bounds"`
	PaddingXPt     float64 `json:"padding_x_pt"`
	PaddingYPt     float64 `json:"padding_y_pt"`
	OffsetXPt      float64 `json:"offset_x_pt"`
	OffsetYPt      float64 `json:"offset_y_pt"`
	StrokeHeightPt float64 `json:"stroke_height_pt,omitempty"`
}
type PlannedAccent struct {
	AccentSpec
	Bounds        Rect `json:"bounds"`
	VisibleBounds Rect `json:"visible_bounds"`
	TargetBounds  Rect `json:"target_bounds"`
}

func validateAccents(s SlideSpec, ids map[string]bool) error {
	targets := map[string]CanvasSpec{}
	for _, c := range s.Canvas {
		targets[c.ID] = c
	}
	for _, a := range s.Accents {
		if !validID(a.ID) || ids[a.ID] {
			return fmt.Errorf("accent ID duplicated/empty: %s", a.ID)
		}
		ids[a.ID] = true
		t, ok := targets[a.Target]
		if !ok || t.Kind != "text" || t.Valign != "top" || t.Background != "" {
			return fmt.Errorf("accent %s requires a top-aligned canvas text target", a.ID)
		}
		if a.Mode != "underline" && a.Mode != "highlight" {
			return fmt.Errorf("accent %s mode must be underline or highlight", a.ID)
		}
		if a.AssetPath == "" || len(a.AssetSHA256) != 64 || !validRect(a.AlphaBounds) || !inside(a.AlphaBounds, Rect{Width: 1, Height: 1}) {
			return fmt.Errorf("accent %s requires pinned asset and normalized alpha bounds", a.ID)
		}
		if !finite(a.PaddingXPt) || !finite(a.PaddingYPt) || !finite(a.OffsetXPt) || !finite(a.OffsetYPt) || (a.Mode == "underline" && !positive(a.StrokeHeightPt)) {
			return fmt.Errorf("accent %s invalid optical calibration", a.ID)
		}
	}
	return nil
}
func planAccents(s SlideSpec, p *PlannedSlide, m Measurements) error {
	for _, a := range s.Accents {
		var t PlannedCanvas
		for _, c := range p.Canvas {
			if c.ID == a.Target {
				t = c
			}
		}
		z := m.ByRequestID[t.MeasurementID]
		if z.RenderedHeightPt > t.FontSizePt*1.5 {
			return fmt.Errorf("accent %s manual_required: target is multiline; split the intended phrase or use native phrase measurement", a.ID)
		}
		target := Rect{t.Bounds.X + t.InsetX + z.OffsetXPt, t.Bounds.Y + t.InsetY + z.OffsetYPt, z.RenderedWidthPt, z.RenderedHeightPt}
		v := Rect{target.X - a.PaddingXPt + a.OffsetXPt, target.Y - a.PaddingYPt + a.OffsetYPt, target.Width + 2*a.PaddingXPt, target.Height + 2*a.PaddingYPt}
		if a.Mode == "underline" {
			v.Y = target.Y + target.Height + a.OffsetYPt
			v.Height = a.StrokeHeightPt
		}
		if !validRect(v) || !inside(v, Rect{Width: s.WidthPt, Height: s.HeightPt}) {
			return fmt.Errorf("accent %s visible bounds exceed slide or collapse", a.ID)
		}
		for _, c := range p.Canvas {
			if c.ID != a.Target && (c.Kind == "text" || c.Kind == "image") && overlap(v, c.Bounds) {
				return fmt.Errorf("accent %s visible ink collides with %s", a.ID, c.ID)
			}
		}
		type item struct {
			id string
			r  Rect
		}
		var obstacles []item
		if p.TitleMeasurementID != "" {
			obstacles = append(obstacles, item{"slide-title", p.TitleBounds})
		}
		for _, c := range p.Cards {
			obstacles = append(obstacles, item{c.ID, c.Bounds})
		}
		for _, c := range p.Pods {
			obstacles = append(obstacles, item{c.ID, c.Bounds})
		}
		for _, c := range p.Roles {
			obstacles = append(obstacles, item{c.ID, c.Bounds})
		}
		for _, c := range p.Phases {
			obstacles = append(obstacles, item{c.ID, c.Bounds})
		}
		if p.Legend != nil {
			obstacles = append(obstacles, item{"staffing-legend", p.Legend.Bounds})
		}
		for _, o := range obstacles {
			if overlap(v, o.r) {
				return fmt.Errorf("accent %s visible ink collides with %s", a.ID, o.id)
			}
		}
		for _, c := range p.Connections {
			for i := 1; i < len(c.Points); i++ {
				if segmentHits(c.Points[i-1], c.Points[i], expand(v, c.WidthPt/2)) {
					return fmt.Errorf("accent %s collides with reporting line %s", a.ID, c.ID)
				}
			}
		}
		b := Rect{Width: v.Width / a.AlphaBounds.Width, Height: v.Height / a.AlphaBounds.Height}
		b.X = v.X - a.AlphaBounds.X*b.Width
		b.Y = v.Y - a.AlphaBounds.Y*b.Height
		// Transparent padding may extend beyond the slide; visible ink cannot.
		p.Accents = append(p.Accents, PlannedAccent{AccentSpec: a, Bounds: b, VisibleBounds: v, TargetBounds: target})
	}
	return nil
}
