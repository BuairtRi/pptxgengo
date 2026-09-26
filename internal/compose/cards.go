package compose

import (
	"fmt"
	"math"
	"strings"
)

// CardSpec describes one bounded, editable card. Kind is "numbered" for a
// numbered explanation row or "metric" for a value-over-label KPI card.
// Bounds are caller-owned and must be large enough for the measured copy.
type CardSpec struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	Profile    string `json:"profile,omitempty"` // light or dark
	Bounds     Rect   `json:"bounds"`
	Number     string `json:"number,omitempty"`
	Title      string `json:"title,omitempty"`
	Body       string `json:"body,omitempty"`
	Value      string `json:"value,omitempty"`
	Label      string `json:"label,omitempty"`
	Surface    string `json:"surface"`
	Accent     string `json:"accent,omitempty"`
	Foreground string `json:"foreground,omitempty"` // auto or explicit #RRGGBB
}

// PlannedCardBlock is an independently measured editable text shape.
type PlannedCardBlock struct {
	Role              string  `json:"role"`
	Text              string  `json:"text"`
	Bounds            Rect    `json:"bounds"`
	FontFace          string  `json:"font_face"`
	FontSizePt        float64 `json:"font_size_pt"`
	Bold              bool    `json:"bold"`
	Foreground        string  `json:"foreground"`
	Align             string  `json:"align"`
	HorizontalInsetPt float64 `json:"horizontal_inset_pt"`
	VerticalInsetPt   float64 `json:"vertical_inset_pt"`
	MeasurementID     string  `json:"measurement_id"`
}

// PlannedCard keeps the card surface, accent rule and text blocks distinct so
// renderers can emit native PowerPoint shapes for every part.
type PlannedCard struct {
	ID           string             `json:"id"`
	Kind         string             `json:"kind"`
	Bounds       Rect               `json:"bounds"`
	Surface      string             `json:"surface"`
	Accent       string             `json:"accent,omitempty"`
	Foreground   string             `json:"foreground"`
	AccentBounds Rect               `json:"accent_bounds,omitempty"`
	Blocks       []PlannedCardBlock `json:"blocks"`
}

const (
	cardNavy = "#070154"
	cardBlue = "#0047FF"
	cardPale = "#E8EEF8"
)

func cardBlockBounds(c CardSpec, role string) Rect {
	r := c.Bounds
	switch role {
	case "number":
		return Rect{X: r.X + 20, Y: r.Y + 14, Width: 48, Height: r.Height - 28}
	case "title":
		return Rect{X: r.X + 80, Y: r.Y + 14, Width: r.Width - 98, Height: 28}
	case "body":
		return Rect{X: r.X + 80, Y: r.Y + 48, Width: r.Width - 98, Height: r.Height - 62}
	case "value":
		return Rect{X: r.X + 12, Y: r.Y + 20, Width: r.Width - 24, Height: math.Min(48, r.Height*.42)}
	case "label":
		y := r.Y + 20 + math.Min(48, r.Height*.42) + 8
		return Rect{X: r.X + 18, Y: y, Width: r.Width - 36, Height: r.Y + r.Height - 16 - y}
	default:
		return Rect{}
	}
}

func cardFields(c CardSpec) []struct{ role, text string } {
	if c.Kind == "numbered" {
		return []struct{ role, text string }{{"number", c.Number}, {"title", c.Title}, {"body", c.Body}}
	}
	return []struct{ role, text string }{{"value", c.Value}, {"label", c.Label}}
}

func effectiveCard(c CardSpec) (CardSpec, error) {
	if c.Profile == "" {
		c.Profile = "light"
	}
	switch c.Profile {
	case "light":
		if c.Surface == "" {
			c.Surface = "surface.light"
		}
	case "dark":
		if c.Surface == "" {
			c.Surface = cardNavy
		}
	default:
		return c, fmt.Errorf("profile must be light or dark")
	}
	if c.Foreground == "" {
		c.Foreground = "auto"
	}
	if c.Accent == "" && c.Kind == "numbered" {
		c.Accent = cardBlue
	}
	return c, nil
}

func cardStyle(role string) (size float64, bold bool, align string, insetX, insetY float64) {
	switch role {
	case "number":
		return 22, true, "center", 1, 1
	case "title":
		return 16, true, "left", 0, 1
	case "body":
		return 12, false, "left", 0, 1
	case "value":
		return 28, true, "center", 2, 1
	default:
		return 12, false, "center", 4, 1
	}
}

func validateCards(s SlideSpec, ids map[string]bool) error {
	for _, c := range s.Cards {
		if !validID(c.ID) || ids[c.ID] {
			return fmt.Errorf("slide %s card IDs must be nonempty and unique: %q", s.ID, c.ID)
		}
		ids[c.ID] = true
		if c.Kind != "numbered" && c.Kind != "metric" {
			return fmt.Errorf("slide %s card %s kind must be numbered or metric", s.ID, c.ID)
		}
		if !validRect(c.Bounds) || !inside(c.Bounds, Rect{Width: s.WidthPt, Height: s.HeightPt}) || c.Bounds.Width < 180 || c.Bounds.Height < 88 {
			return fmt.Errorf("slide %s card %s bounds are invalid, too small, or outside slide", s.ID, c.ID)
		}
		c, err := effectiveCard(c)
		if err != nil {
			return fmt.Errorf("slide %s card %s: %w", s.ID, c.ID, err)
		}
		bg, err := resolveColor(c.Surface)
		if err != nil {
			return fmt.Errorf("slide %s card %s surface: %w", s.ID, c.ID, err)
		}
		if _, err := foreground(c.Foreground, bg); err != nil {
			return fmt.Errorf("slide %s card %s foreground: %w", s.ID, c.ID, err)
		}
		if c.Kind == "numbered" {
			if !validID(c.Number) || !validID(c.Title) || !validID(c.Body) || c.Value != "" || c.Label != "" {
				return fmt.Errorf("slide %s numbered card %s requires number, title and body", s.ID, c.ID)
			}
		} else {
			if !validID(c.Value) || !validID(c.Label) || c.Number != "" || c.Title != "" || c.Body != "" || c.Accent != "" {
				return fmt.Errorf("slide %s metric card %s requires value and label", s.ID, c.ID)
			}
		}
		if c.Accent != "" {
			if _, err := resolveColor(c.Accent); err != nil {
				return fmt.Errorf("slide %s card %s accent: %w", s.ID, c.ID, err)
			}
		}
		for _, f := range cardFields(c) {
			r := cardBlockBounds(c, f.role)
			_, _, _, ix, iy := cardStyle(f.role)
			if !validRect(r) || r.Width <= 2*ix || r.Height <= 2*iy {
				return fmt.Errorf("slide %s card %s %s slot has no usable text area", s.ID, c.ID, f.role)
			}
		}
		// Cards are root components. They cannot silently cover other components.
		if strings.TrimSpace(s.Title) != "" && overlap(c.Bounds, s.TitleBounds) {
			return fmt.Errorf("slide %s card %s overlaps slide title", s.ID, c.ID)
		}
		for _, other := range s.Cards {
			if other.ID != c.ID && overlap(c.Bounds, other.Bounds) {
				return fmt.Errorf("slide %s cards %s and %s overlap", s.ID, c.ID, other.ID)
			}
		}
		for _, p := range s.Pods {
			if overlap(c.Bounds, Rect{X: p.Bounds.X, Y: p.Bounds.Y, Width: p.Bounds.Width, Height: maxHeight(p)}) {
				return fmt.Errorf("slide %s card %s overlaps pod %s", s.ID, c.ID, p.ID)
			}
		}
		for _, r := range s.Roles {
			if overlap(c.Bounds, r.Bounds) {
				return fmt.Errorf("slide %s card %s overlaps role %s", s.ID, c.ID, r.ID)
			}
		}
		for _, p := range s.Phases {
			if overlap(c.Bounds, p.Bounds) {
				return fmt.Errorf("slide %s card %s overlaps phase %s", s.ID, c.ID, p.ID)
			}
		}
		if s.Legend != nil && overlap(c.Bounds, s.Legend.Bounds) {
			return fmt.Errorf("slide %s card %s overlaps staffing legend", s.ID, c.ID)
		}
		for _, el := range s.Canvas {
			declared := false
			for _, peer := range el.AllowOverlap {
				if peer == c.ID {
					declared = true
					break
				}
			}
			if overlap(c.Bounds, el.Bounds) && !declared {
				return fmt.Errorf("slide %s card %s overlaps canvas %s without an explicit canvas relationship", s.ID, c.ID, el.ID)
			}
		}
	}
	return nil
}

func cardProbes(s SlideSpec) ([]ProbeRequest, error) {
	var out []ProbeRequest
	for _, raw := range s.Cards {
		c, err := effectiveCard(raw)
		if err != nil {
			return nil, fmt.Errorf("slide %s card %s: %w", s.ID, raw.ID, err)
		}
		bg, err := resolveColor(c.Surface)
		if err != nil {
			return nil, fmt.Errorf("slide %s card %s surface: %w", s.ID, c.ID, err)
		}
		for _, f := range cardFields(c) {
			r := cardBlockBounds(c, f.role)
			size, bold, align, ix, iy := cardStyle(f.role)
			resolved, err := foreground(c.Foreground, bg)
			if err != nil {
				return nil, fmt.Errorf("slide %s card %s foreground: %w", s.ID, c.ID, err)
			}
			out = append(out, ProbeRequest{ID: requestID(s.ID, "card", c.ID, f.role), SlideID: s.ID, RoleID: c.ID + ":" + f.role, Kind: "card_" + f.role, Text: f.text, TextWidthPt: r.Width - 2*ix, HorizontalInsetPt: ix, VerticalInsetPt: iy, FontFace: "Arial", FontSizePt: size, Bold: bold, Foreground: resolved, Background: bg, Align: align})
		}
	}
	return out, nil
}

func planCards(s SlideSpec, out *PlannedSlide, m Measurements) error {
	for _, raw := range s.Cards {
		c, err := effectiveCard(raw)
		if err != nil {
			return fmt.Errorf("slide %s card %s: %w", s.ID, raw.ID, err)
		}
		surface, _ := resolveColor(c.Surface)
		accent := ""
		var accentBounds Rect
		if c.Kind == "numbered" {
			accent = c.Accent
			if accent == "" {
				accent = cardBlue
			}
			accent, _ = resolveColor(accent)
			accentBounds = Rect{X: c.Bounds.X, Y: c.Bounds.Y, Width: 5, Height: c.Bounds.Height}
		}
		fg, err := foreground(c.Foreground, c.Surface)
		if err != nil {
			return fmt.Errorf("slide %s card %s foreground: %w", s.ID, c.ID, err)
		}
		p := PlannedCard{ID: c.ID, Kind: c.Kind, Bounds: c.Bounds, Surface: surface, Accent: accent, Foreground: fg, AccentBounds: accentBounds}
		for _, f := range cardFields(c) {
			id := requestID(s.ID, "card", c.ID, f.role)
			q := m.ByRequestID[id]
			r := cardBlockBounds(c, f.role)
			size, bold, align, ix, iy := cardStyle(f.role)
			if !positive(q.RenderedWidthPt) || !positive(q.RenderedHeightPt) {
				return fmt.Errorf("slide %s card %s %s has invalid text measurement", s.ID, c.ID, f.role)
			}
			if q.RenderedWidthPt > r.Width-2*ix+1e-6 || q.RenderedHeightPt > r.Height-2*iy+1e-6 {
				return fmt.Errorf("slide %s card %s %s exceeds its fixed measured text area; enlarge the card or shorten copy", s.ID, c.ID, f.role)
			}
			p.Blocks = append(p.Blocks, PlannedCardBlock{Role: f.role, Text: f.text, Bounds: r, FontFace: "Arial", FontSizePt: size, Bold: bold, Foreground: fg, Align: align, HorizontalInsetPt: ix, VerticalInsetPt: iy, MeasurementID: id})
		}
		out.Cards = append(out.Cards, p)
	}
	return nil
}
