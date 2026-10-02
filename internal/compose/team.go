package compose

import (
	"fmt"
)

type StandaloneRoleSpec struct {
	ID                string  `json:"id"`
	Label             string  `json:"label"`
	Background        string  `json:"background"`
	Foreground        string  `json:"foreground"`
	Bounds            Rect    `json:"bounds"`
	FontFace          string  `json:"font_face"`
	FontSizePt        float64 `json:"font_size_pt"`
	Bold              bool    `json:"bold"`
	HorizontalInsetPt float64 `json:"horizontal_inset_pt"`
	VerticalInsetPt   float64 `json:"vertical_inset_pt"`
}

type PhaseSpec struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Bounds        Rect     `json:"bounds"`
	Members       []string `json:"members"`
	PaddingPt     float64  `json:"padding_pt"`
	LabelHeightPt float64  `json:"label_height_pt"`
	FontFace      string   `json:"font_face"`
	FontSizePt    float64  `json:"font_size_pt"`
	Bold          bool     `json:"bold"`
	Surface       string   `json:"surface"`
	Foreground    string   `json:"foreground"`
}

type LegendSpec struct {
	Bounds       Rect    `json:"bounds"`
	FontFace     string  `json:"font_face"`
	FontSizePt   float64 `json:"font_size_pt"`
	Bold         bool    `json:"bold"`
	Foreground   string  `json:"foreground"`
	SwatchSizePt float64 `json:"swatch_size_pt"`
	GapPt        float64 `json:"gap_pt"`
	ItemGapPt    float64 `json:"item_gap_pt"`
}

type PlannedPhase struct {
	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	Bounds             Rect     `json:"bounds"`
	Surface            string   `json:"surface"`
	Foreground         string   `json:"foreground"`
	TitleRect          Rect     `json:"title_rect"`
	TitleFontFace      string   `json:"title_font_face"`
	TitleFontSizePt    float64  `json:"title_font_size_pt"`
	TitleBold          bool     `json:"title_bold"`
	TitleMeasurementID string   `json:"title_measurement_id"`
	Members            []string `json:"members"`
}

type PlannedLegend struct {
	Bounds  Rect                `json:"bounds"`
	Entries []PlannedLegendItem `json:"entries"`
}

type PlannedLegendItem struct {
	Token         string  `json:"token"`
	Label         string  `json:"label"`
	SwatchBounds  Rect    `json:"swatch_bounds"`
	LabelBounds   Rect    `json:"label_bounds"`
	Color         string  `json:"color"`
	Foreground    string  `json:"foreground"`
	FontFace      string  `json:"font_face"`
	FontSizePt    float64 `json:"font_size_pt"`
	Bold          bool    `json:"bold"`
	MeasurementID string  `json:"measurement_id"`
}

type staffingLegend struct{ token, label string }

var canonicalStaffingLegend = []staffingLegend{
	{token: "staffing.wm_full_time", label: "WM FULL TIME"},
	{token: "staffing.wm_part_time", label: "WM PART TIME"},
	{token: "staffing.client_part_time", label: "CLIENT PART TIME"},
}

func isStaffingToken(s string) bool {
	for _, x := range canonicalStaffingLegend {
		if s == x.token {
			return true
		}
	}
	return false
}

func staffingTokens(s SlideSpec) []staffingLegend {
	used := map[string]bool{}
	for _, p := range s.Pods {
		for _, r := range p.Roles {
			if isStaffingToken(r.Background) {
				used[r.Background] = true
			}
		}
	}
	for _, r := range s.Roles {
		if isStaffingToken(r.Background) {
			used[r.Background] = true
		}
	}
	var out []staffingLegend
	for _, x := range canonicalStaffingLegend {
		if used[x.token] {
			out = append(out, x)
		}
	}
	return out
}

func teamProbes(s SlideSpec) ([]ProbeRequest, error) {
	var out []ProbeRequest
	for _, r := range s.Roles {
		bg, _ := resolveColor(r.Background)
		fg, err := foreground(r.Foreground, bg)
		if err != nil {
			return nil, fmt.Errorf("slide %s standalone role %s: %w", s.ID, r.ID, err)
		}
		out = append(out, ProbeRequest{ID: requestID(s.ID, "standalone_role", r.ID), SlideID: s.ID, RoleID: r.ID, Kind: "standalone_role", Text: r.Label, TextWidthPt: r.Bounds.Width - 2*r.HorizontalInsetPt, HorizontalInsetPt: r.HorizontalInsetPt, VerticalInsetPt: r.VerticalInsetPt, FontFace: r.FontFace, FontSizePt: r.FontSizePt, Bold: r.Bold, Foreground: fg, Background: bg})
	}
	for _, p := range s.Phases {
		bg, _ := resolveColor(p.Surface)
		fg, err := foreground(p.Foreground, bg)
		if err != nil {
			return nil, fmt.Errorf("slide %s phase %s: %w", s.ID, p.ID, err)
		}
		out = append(out, ProbeRequest{ID: requestID(s.ID, "phase_title", p.ID), SlideID: s.ID, PhaseID: p.ID, Kind: "phase_title", Text: p.Title, TextWidthPt: p.Bounds.Width - 2*p.PaddingPt, FontFace: p.FontFace, FontSizePt: p.FontSizePt, Bold: p.Bold, Foreground: fg, Background: bg})
	}
	if s.Legend != nil {
		tokens := staffingTokens(s)
		if len(tokens) == 0 {
			return nil, fmt.Errorf("slide %s legend requested without semantic staffing tokens", s.ID)
		}
		labelW := (s.Legend.Bounds.Width - float64(len(tokens))*(s.Legend.SwatchSizePt+s.Legend.GapPt) - float64(len(tokens)-1)*s.Legend.ItemGapPt) / float64(len(tokens))
		fg, err := foreground(s.Legend.Foreground, white)
		if err != nil {
			return nil, fmt.Errorf("slide %s legend: %w", s.ID, err)
		}
		for _, t := range tokens {
			out = append(out, ProbeRequest{ID: requestID(s.ID, "legend", t.token), SlideID: s.ID, LegendToken: t.token, Kind: "legend_label", Text: t.label, TextWidthPt: labelW, FontFace: s.Legend.FontFace, FontSizePt: s.Legend.FontSizePt, Bold: s.Legend.Bold, Foreground: fg, Background: white})
		}
	}
	return out, nil
}

func validateTeam(s SlideSpec, ids map[string]bool) error {
	teamMode := len(s.Roles) > 0 || len(s.Phases) > 0 || len(s.Connections) > 0 || s.Legend != nil
	semantic := staffingTokens(s)
	if teamMode && len(semantic) > 0 && s.Legend == nil {
		return fmt.Errorf("slide %s semantic staffing colors in a team composition require an auto-generated legend", s.ID)
	}
	rootIDs := map[string]bool{}
	for _, p := range s.Pods {
		rootIDs[p.ID] = true
	}
	for _, r := range s.Roles {
		if !validID(r.ID) || ids[r.ID] {
			return fmt.Errorf("slide %s component ID is empty or duplicated: %q", s.ID, r.ID)
		}
		ids[r.ID] = true
		rootIDs[r.ID] = true
		if !validID(r.Label) || !validRect(r.Bounds) || !inside(r.Bounds, Rect{Width: s.WidthPt, Height: s.HeightPt}) || !ValidFontFace(r.FontFace) || !positive(r.FontSizePt) || !nonnegative(r.HorizontalInsetPt) || !nonnegative(r.VerticalInsetPt) || r.Bounds.Width <= 2*r.HorizontalInsetPt || r.Bounds.Height <= 2*r.VerticalInsetPt {
			return fmt.Errorf("slide %s standalone role %s has invalid label, bounds, font or insets", s.ID, r.ID)
		}
		if teamMode && !isStaffingToken(r.Background) {
			return fmt.Errorf("slide %s standalone role %s must use a staffing semantic background", s.ID, r.ID)
		}
		if _, e := resolveColor(r.Background); e != nil {
			return fmt.Errorf("slide %s standalone role %s background: %w", s.ID, r.ID, e)
		}
		if _, e := foreground(r.Foreground, r.Background); e != nil {
			return fmt.Errorf("slide %s standalone role %s: %w", s.ID, r.ID, e)
		}
	}
	for _, p := range s.Pods {
		if teamMode {
			for _, r := range p.Roles {
				if !isStaffingToken(r.Background) {
					return fmt.Errorf("slide %s pod role %s/%s must use a staffing semantic background in team mode", s.ID, p.ID, r.ID)
				}
			}
		}
	}
	phaseIDs := map[string]bool{}
	for _, p := range s.Phases {
		if !validID(p.ID) || ids[p.ID] || phaseIDs[p.ID] {
			return fmt.Errorf("slide %s phase ID is empty or duplicated: %q", s.ID, p.ID)
		}
		ids[p.ID] = true
		phaseIDs[p.ID] = true
		if !validID(p.Title) || !validRect(p.Bounds) || !inside(p.Bounds, Rect{Width: s.WidthPt, Height: s.HeightPt}) || !ValidFontFace(p.FontFace) || !positive(p.FontSizePt) || !nonnegative(p.PaddingPt) || !positive(p.LabelHeightPt) || p.Bounds.Width <= 2*p.PaddingPt || p.Bounds.Height <= 2*p.PaddingPt+p.LabelHeightPt {
			return fmt.Errorf("slide %s phase %s has invalid title, bounds, font or footer geometry", s.ID, p.ID)
		}
		if _, e := resolveColor(p.Surface); e != nil {
			return fmt.Errorf("slide %s phase %s surface: %w", s.ID, p.ID, e)
		}
		if _, e := foreground(p.Foreground, p.Surface); e != nil {
			return fmt.Errorf("slide %s phase %s: %w", s.ID, p.ID, e)
		}
		if len(p.Members) == 0 {
			return fmt.Errorf("slide %s phase %s requires members", s.ID, p.ID)
		}
		seen := map[string]bool{}
		for _, m := range p.Members {
			if !rootIDs[m] || seen[m] {
				return fmt.Errorf("slide %s phase %s has unknown or duplicate member %q", s.ID, p.ID, m)
			}
			seen[m] = true
		}
	}
	if s.Legend != nil {
		l := s.Legend
		if !validRect(l.Bounds) || !inside(l.Bounds, Rect{Width: s.WidthPt, Height: s.HeightPt}) || !ValidFontFace(l.FontFace) || !positive(l.FontSizePt) || !positive(l.SwatchSizePt) || !nonnegative(l.GapPt) || !nonnegative(l.ItemGapPt) || l.SwatchSizePt > l.Bounds.Height {
			return fmt.Errorf("slide %s legend has invalid bounds, font, or item geometry", s.ID)
		}
		if _, e := foreground(l.Foreground, white); e != nil {
			return fmt.Errorf("slide %s legend: %w", s.ID, e)
		}
		tokens := staffingTokens(s)
		if len(tokens) == 0 {
			return fmt.Errorf("slide %s legend requested without semantic staffing tokens", s.ID)
		}
		labelsW := l.Bounds.Width - float64(len(tokens))*(l.SwatchSizePt+l.GapPt) - float64(len(tokens)-1)*l.ItemGapPt
		if !finite(labelsW) || labelsW <= 0 {
			return fmt.Errorf("slide %s legend has no label width after swatches and gaps", s.ID)
		}
	}
	return nil
}

func planTeam(s SlideSpec, out *PlannedSlide, m Measurements) error {
	for _, r := range s.Roles {
		id := requestID(s.ID, "standalone_role", r.ID)
		q := m.ByRequestID[id]
		if q.RenderedWidthPt > r.Bounds.Width-2*r.HorizontalInsetPt+1e-6 || q.RenderedHeightPt > r.Bounds.Height-2*r.VerticalInsetPt+1e-6 {
			return fmt.Errorf("slide %s standalone role %s text does not fit fixed bounds", s.ID, r.ID)
		}
		bg, _ := resolveColor(r.Background)
		fg, _ := foreground(r.Foreground, bg)
		out.Roles = append(out.Roles, PlannedRole{ID: r.ID, Label: r.Label, Bounds: r.Bounds, Anchors: rectAnchors(r.Bounds), Background: bg, Foreground: fg, FontFace: r.FontFace, FontSizePt: r.FontSizePt, Bold: r.Bold, HorizontalInsetPt: r.HorizontalInsetPt, VerticalInsetPt: r.VerticalInsetPt, MeasurementID: id})
	}
	for _, p := range s.Phases {
		id := requestID(s.ID, "phase_title", p.ID)
		measure := m.ByRequestID[id]
		width := p.Bounds.Width - 2*p.PaddingPt
		if measure.RenderedWidthPt > width+1e-6 || measure.RenderedHeightPt > p.LabelHeightPt+1e-6 {
			return fmt.Errorf("slide %s phase %s footer label does not fit fixed bounds", s.ID, p.ID)
		}
		bg, _ := resolveColor(p.Surface)
		fg, _ := foreground(p.Foreground, bg)
		footer := Rect{X: p.Bounds.X + p.PaddingPt, Y: p.Bounds.Y + p.Bounds.Height - p.PaddingPt - p.LabelHeightPt, Width: width, Height: p.LabelHeightPt}
		out.Phases = append(out.Phases, PlannedPhase{ID: p.ID, Title: p.Title, Bounds: p.Bounds, Surface: bg, Foreground: fg, TitleRect: footer, TitleFontFace: p.FontFace, TitleFontSizePt: p.FontSizePt, TitleBold: p.Bold, TitleMeasurementID: id, Members: append([]string(nil), p.Members...)})
	}
	if s.Legend != nil {
		l := s.Legend
		tokens := staffingTokens(s)
		fg, _ := foreground(l.Foreground, white)
		available := (l.Bounds.Width - float64(len(tokens))*(l.SwatchSizePt+l.GapPt) - float64(len(tokens)-1)*l.ItemGapPt) / float64(len(tokens))
		entries := make([]PlannedLegendItem, 0, len(tokens))
		x := l.Bounds.X
		for _, t := range tokens {
			id := requestID(s.ID, "legend", t.token)
			meas := m.ByRequestID[id]
			if meas.RenderedWidthPt > available+1e-6 || meas.RenderedHeightPt > l.Bounds.Height+1e-6 {
				return fmt.Errorf("slide %s legend label %s does not fit", s.ID, t.token)
			}
			color, _ := resolveColor(t.token)
			entries = append(entries, PlannedLegendItem{Token: t.token, Label: t.label, SwatchBounds: Rect{X: x, Y: l.Bounds.Y + (l.Bounds.Height-l.SwatchSizePt)/2, Width: l.SwatchSizePt, Height: l.SwatchSizePt}, LabelBounds: Rect{X: x + l.SwatchSizePt + l.GapPt, Y: l.Bounds.Y + (l.Bounds.Height-meas.RenderedHeightPt)/2, Width: available, Height: meas.RenderedHeightPt}, Color: color, Foreground: fg, FontFace: l.FontFace, FontSizePt: l.FontSizePt, Bold: l.Bold, MeasurementID: id})
			x += l.SwatchSizePt + l.GapPt + available + l.ItemGapPt
		}
		out.Legend = &PlannedLegend{Bounds: l.Bounds, Entries: entries}
	}
	return validateTeamPlan(s, *out)
}

func validateTeamPlan(s SlideSpec, p PlannedSlide) error {
	components := make(map[string]Rect)
	for _, x := range p.Pods {
		components[x.ID] = x.Bounds
	}
	for _, x := range p.Roles {
		components[x.ID] = x.Bounds
	}
	for id, a := range components {
		for jd, b := range components {
			if id < jd && overlap(a, b) {
				return fmt.Errorf("slide %s components %s and %s overlap", s.ID, id, jd)
			}
		}
		if p.TitleMeasurementID != "" && overlap(a, p.TitleBounds) {
			return fmt.Errorf("slide %s title collides with component %s", s.ID, id)
		}
		if p.Legend != nil && overlap(a, p.Legend.Bounds) {
			return fmt.Errorf("slide %s legend collides with component %s", s.ID, id)
		}
	}
	if p.Legend != nil && p.TitleMeasurementID != "" && overlap(p.Legend.Bounds, p.TitleBounds) {
		return fmt.Errorf("slide %s legend collides with slide title", s.ID)
	}
	phasePaddingByID := map[string]float64{}
	for _, ph := range s.Phases {
		phasePaddingByID[ph.ID] = ph.PaddingPt
	}
	for i, a := range p.Phases {
		for _, b := range p.Phases[i+1:] {
			if overlap(a.Bounds, b.Bounds) {
				return fmt.Errorf("slide %s phases %s and %s overlap", s.ID, a.ID, b.ID)
			}
		}
		if p.TitleMeasurementID != "" && overlap(a.Bounds, p.TitleBounds) {
			return fmt.Errorf("slide %s phase %s collides with slide title", s.ID, a.ID)
		}
		if p.Legend != nil && overlap(a.Bounds, p.Legend.Bounds) {
			return fmt.Errorf("slide %s phase %s collides with legend", s.ID, a.ID)
		}
		memberSet := map[string]bool{}
		for _, id := range a.Members {
			memberSet[id] = true
		}
		pad := phasePaddingByID[a.ID]
		content := Rect{X: a.Bounds.X + pad, Y: a.Bounds.Y + pad, Width: a.Bounds.Width - 2*pad, Height: a.TitleRect.Y - pad - (a.Bounds.Y + pad)}
		for id, r := range components {
			if memberSet[id] {
				if !inside(r, content) {
					return fmt.Errorf("slide %s phase %s member %s is not contained above footer", s.ID, a.ID, id)
				}
			} else if overlap(r, a.Bounds) {
				return fmt.Errorf("slide %s phase %s overlaps nonmember component %s", s.ID, a.ID, id)
			}
		}
	}
	return nil
}
