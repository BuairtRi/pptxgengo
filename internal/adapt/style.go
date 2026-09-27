package adapt

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func ResolveStyle(o StyleOptions) (Style, error) {
	s := Style{Profile: "neutral", FontFace: "Arial", TitleFontPt: 26, BodyFontPt: 11, LabelFontPt: 11, Text: "#070154", Secondary: "#50658E", Surface: "#FFFFFF", MutedSurface: "#E8EEF8", Accent: "#0047FF", Active: "#0047FF", Inactive: "#CED7E6", Border: "#CED7E6", Navy: "#070154", White: "#FFFFFF"}
	if o.Profile != "" {
		s.Profile = o.Profile
	}
	switch s.Profile {
	case "neutral":
	case "subtle":
		s.Surface = "#F4F6FA"
	case "inverse":
		s.Surface = "#070154"
		s.Text = "#FFFFFF"
		s.Secondary = "#CED7E6"
		s.MutedSurface = "#28245C"
	default:
		return s, fmt.Errorf("unknown style profile %q", s.Profile)
	}
	if o.FontFace != "" && o.FontFace != "Arial" {
		return s, fmt.Errorf("font_face %q unsupported: native measurement currently supports Arial", o.FontFace)
	}
	for _, x := range []struct {
		v        float64
		p        *float64
		name     string
		min, max float64
	}{{o.TitleFontPt, &s.TitleFontPt, "title_font_pt", 18, 36}, {o.BodyFontPt, &s.BodyFontPt, "body_font_pt", 9, 18}, {o.LabelFontPt, &s.LabelFontPt, "label_font_pt", 9, 18}} {
		if x.v != 0 {
			if math.IsNaN(x.v) || math.IsInf(x.v, 0) || x.v < x.min || x.v > x.max {
				return s, fmt.Errorf("%s must be in [%g,%g]", x.name, x.min, x.max)
			}
			*x.p = x.v
		}
	}
	roles := map[string]*string{"text.primary": &s.Text, "text.secondary": &s.Secondary, "surface": &s.Surface, "surface.muted": &s.MutedSurface, "accent": &s.Accent, "state.active": &s.Active, "state.inactive": &s.Inactive, "border": &s.Border}
	for role, color := range o.Colors {
		p, ok := roles[role]
		if !ok {
			return s, fmt.Errorf("unknown semantic color role %q", role)
		}
		if !validColor(color) {
			return s, fmt.Errorf("role %s requires #RRGGBB", role)
		}
		*p = strings.ToUpper(color)
	}
	if contrast(s.Text, s.Surface) < 4.5 {
		return s, fmt.Errorf("text.primary contrast against surface is below 4.5:1")
	}
	return s, nil
}
func validColor(c string) bool {
	if len(c) != 7 || c[0] != '#' {
		return false
	}
	_, e := strconv.ParseUint(c[1:], 16, 24)
	return e == nil
}
func luminance(c string) float64 {
	n, _ := strconv.ParseUint(strings.TrimPrefix(c, "#"), 16, 24)
	v := []float64{float64((n>>16)&255) / 255, float64((n>>8)&255) / 255, float64(n&255) / 255}
	for i, x := range v {
		if x <= 0.04045 {
			v[i] = x / 12.92
		} else {
			v[i] = math.Pow((x+0.055)/1.055, 2.4)
		}
	}
	return .2126*v[0] + .7152*v[1] + .0722*v[2]
}
func contrast(a, b string) float64 {
	x, y := luminance(a), luminance(b)
	if x < y {
		x, y = y, x
	}
	return (x + .05) / (y + .05)
}

// TextOn resolves a readable brand ink for painted component surfaces.
func TextOn(fill string) string {
	if contrast("#070154", fill) >= contrast("#FFFFFF", fill) {
		return "#070154"
	}
	return "#FFFFFF"
}

func styleColor(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

// Ink honors the requested primary text color when it is readable on a
// component surface, then falls back to the higher-contrast brand ink.
func (s Style) Ink(fill string) string {
	if contrast(s.Text, fill) >= 4.5 {
		return s.Text
	}
	return TextOn(fill)
}
