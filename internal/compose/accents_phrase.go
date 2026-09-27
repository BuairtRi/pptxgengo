package compose

import (
	"fmt"
	"math"
	"strings"
)

// PhraseRequest selects Unicode code points, not bytes. A range is zero-based,
// end-exclusive. An occurrence is one-based and mandatory when text repeats.
// ID binds measured fragments to the accent that requested them.
type PhraseRequest struct {
	ID         string `json:"id"`
	Phrase     string `json:"phrase"`
	Occurrence int    `json:"occurrence,omitempty"`
	StartRune  *int   `json:"start_rune,omitempty"`
	EndRune    *int   `json:"end_rune,omitempty"`
}

func ResolvePhrase(text string, q PhraseRequest) (int, int, error) {
	text = strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
	rr, wanted := []rune(text), []rune(q.Phrase)
	if q.ID == "" || len(wanted) == 0 {
		return 0, 0, fmt.Errorf("phrase needs an ID and nonempty text")
	}
	if q.StartRune != nil || q.EndRune != nil {
		if q.StartRune == nil || q.EndRune == nil || q.Occurrence != 0 {
			return 0, 0, fmt.Errorf("phrase range requires both ends and no occurrence")
		}
		a, b := *q.StartRune, *q.EndRune
		if a < 0 || b <= a || b > len(rr) || string(rr[a:b]) != q.Phrase {
			return 0, 0, fmt.Errorf("phrase range does not match exact text")
		}
		return a, b, nil
	}
	if q.Occurrence < 0 {
		return 0, 0, fmt.Errorf("phrase occurrence must be positive")
	}
	var starts []int
	for i := 0; i+len(wanted) <= len(rr); i++ {
		if string(rr[i:i+len(wanted)]) == q.Phrase {
			starts = append(starts, i)
		}
	}
	n := q.Occurrence
	if n == 0 {
		if len(starts) != 1 {
			return 0, 0, fmt.Errorf("phrase has %d matches; choose an explicit occurrence or range", len(starts))
		}
		n = 1
	}
	if n > len(starts) {
		return 0, 0, fmt.Errorf("phrase occurrence %d missing (%d matches)", n, len(starts))
	}
	return starts[n-1], starts[n-1] + len(wanted), nil
}

func accentPhraseRequests(s SlideSpec, target string) []PhraseRequest {
	var out []PhraseRequest
	for _, a := range s.Accents {
		if a.Target == target && a.Phrase != "" {
			out = append(out, PhraseRequest{ID: a.ID, Phrase: a.Phrase, Occurrence: a.Occurrence, StartRune: a.StartRune, EndRune: a.EndRune})
		}
	}
	return append(out, arrowPhraseRequests(s, target)...)
}

func validatePhraseAccent(a AccentSpec, t CanvasSpec, s SlideSpec) error {
	if a.Multiline != "" && a.Multiline != "per_line" && a.Multiline != "manual" {
		return fmt.Errorf("accent %s multiline must be per_line or manual", a.ID)
	}
	if !finite(a.ContainerAspect) || a.ContainerAspect < 0 || !finite(a.RotationDeg) || a.RotationDeg < -15 || a.RotationDeg > 15 {
		return fmt.Errorf("accent %s invalid aspect/rotation", a.ID)
	}
	if a.Mode == "underline" && a.RotationDeg != 0 {
		return fmt.Errorf("accent %s underline rotation is unsupported", a.ID)
	}
	if a.Phrase != "" {
		text := t.Text
		if len(t.Paragraphs) > 0 {
			text = richText(t.Paragraphs)
		}
		_, _, err := ResolvePhrase(text, PhraseRequest{ID: a.ID, Phrase: a.Phrase, Occurrence: a.Occurrence, StartRune: a.StartRune, EndRune: a.EndRune})
		if err != nil && a.Staging == nil {
			return fmt.Errorf("accent %s manual_required: %w", a.ID, err)
		}
	} else if a.Occurrence != 0 || a.StartRune != nil || a.EndRune != nil {
		return fmt.Errorf("accent %s selector requires phrase", a.ID)
	}
	if a.Staging != nil {
		z := a.Staging
		slide := Rect{Width: s.WidthPt, Height: s.HeightPt}
		if !validRect(z.AssetBounds) || !validRect(z.NoteBounds) || !inside(z.AssetBounds, slide) || !inside(z.NoteBounds, slide) || overlap(z.AssetBounds, z.NoteBounds) || strings.TrimSpace(z.Note) == "" {
			return fmt.Errorf("accent %s staging needs separate on-slide asset/note frames and precise placement note", a.ID)
		}
	}
	return nil
}

func accentNoteProbe(s SlideSpec, a AccentSpec) ProbeRequest {
	return ProbeRequest{ID: requestID(s.ID, "accent-note", a.ID), SlideID: s.ID, Kind: "accent_note", Text: "MANUAL PLACEMENT: " + a.Staging.Note, TextWidthPt: a.Staging.NoteBounds.Width, FontFace: "Arial", FontSizePt: 11, Foreground: "#070154", Background: white, Align: "left"}
}

func planAccents(s SlideSpec, p *PlannedSlide, m Measurements) error {
	for _, a := range s.Accents {
		var t PlannedCanvas
		for _, c := range p.Canvas {
			if c.ID == a.Target {
				t = c
				break
			}
		}
		z := m.ByRequestID[t.MeasurementID]
		var fragments []Rect
		reason := ""
		if a.Phrase != "" {
			text := t.Text
			if len(t.Paragraphs) > 0 {
				text = richText(t.Paragraphs)
			}
			_, _, err := ResolvePhrase(text, PhraseRequest{ID: a.ID, Phrase: a.Phrase, Occurrence: a.Occurrence, StartRune: a.StartRune, EndRune: a.EndRune})
			if err != nil {
				reason = err.Error()
			} else {
				fragments = z.PhraseBounds[a.ID]
				if len(fragments) == 0 {
					return fmt.Errorf("accent %s lacks native phrase geometry; remeasure", a.ID)
				}
			}
		} else {
			font := t.FontSizePt
			if len(t.Paragraphs) > 0 {
				font = t.Paragraphs[0].Runs[0].FontSizePt
			}
			if z.RenderedHeightPt > font*1.5 {
				reason = "target is multiline; select an exact phrase or use explicit per-line phrase placement"
			} else {
				fragments = []Rect{{X: z.OffsetXPt, Y: z.OffsetYPt, Width: z.RenderedWidthPt, Height: z.RenderedHeightPt}}
			}
		}
		if len(fragments) > 1 && a.Multiline != "per_line" {
			reason = fmt.Sprintf("selected phrase spans %d native lines", len(fragments))
		}
		if reason != "" {
			if a.Staging == nil {
				return fmt.Errorf("accent %s manual_required: %s", a.ID, reason)
			}
			q := accentNoteProbe(s, a)
			nm, ok := m.ByRequestID[q.ID]
			if !ok || nm.RenderedWidthPt > a.Staging.NoteBounds.Width+.01 || nm.RenderedHeightPt > a.Staging.NoteBounds.Height+.01 {
				return fmt.Errorf("accent %s staging note lacks fitting native measurement", a.ID)
			}
			if err := checkAccentCollisions(AccentSpec{ID: a.ID}, a.Staging.AssetBounds, p); err != nil {
				return err
			}
			if err := checkAccentCollisions(AccentSpec{ID: a.ID}, a.Staging.NoteBounds, p); err != nil {
				return err
			}
			p.Canvas = append(p.Canvas, PlannedCanvas{CanvasSpec: CanvasSpec{ID: a.ID + "/manual-note", Kind: "text", Bounds: a.Staging.NoteBounds, Text: q.Text, FontFace: "Arial", FontSizePt: 11, Foreground: "#070154", Align: "left", Valign: "top", Layer: 100}, MeasurementID: q.ID})
			p.Accents = append(p.Accents, PlannedAccent{AccentSpec: a, Bounds: a.Staging.AssetBounds, VisibleBounds: a.Staging.AssetBounds, Status: "manual_required", Reason: reason})
			p.ManualRequired = append(p.ManualRequired, "Manual accent placement: "+a.Staging.Note+" ("+reason+")")
			continue
		}
		for i, r := range fragments {
			if !validRect(r) {
				return fmt.Errorf("accent %s invalid native fragment", a.ID)
			}
			target := Rect{X: t.Bounds.X + t.InsetX + r.X, Y: t.Bounds.Y + t.InsetY + r.Y, Width: r.Width, Height: r.Height}
			v := Rect{X: target.X - a.PaddingXPt + a.OffsetXPt, Y: target.Y - a.PaddingYPt + a.OffsetYPt, Width: target.Width + 2*a.PaddingXPt, Height: target.Height + 2*a.PaddingYPt}
			if a.Mode == "underline" {
				v.Y = target.Y + target.Height + a.OffsetYPt
				v.Height = a.StrokeHeightPt
				if a.ContainerAspect > 0 {
					v.Height = v.Width / a.AlphaBounds.Width / a.ContainerAspect * a.AlphaBounds.Height
				}
			}
			if !validRect(v) || !inside(v, Rect{Width: s.WidthPt, Height: s.HeightPt}) {
				return fmt.Errorf("accent %s visible bounds exceed slide or collapse", a.ID)
			}
			if err := checkAccentCollisions(a, v, p); err != nil {
				return err
			}
			b, err := accentImageFrame(a, v)
			if err != nil {
				return fmt.Errorf("accent %s: %w", a.ID, err)
			}
			placed := a
			if len(fragments) > 1 {
				placed.ID = fmt.Sprintf("%s/line-%d", a.ID, i+1)
			}
			p.Accents = append(p.Accents, PlannedAccent{AccentSpec: placed, Bounds: b, VisibleBounds: v, TargetBounds: target, Status: "placed", TargetMeasurementID: t.MeasurementID, SourceAccentID: a.ID, FragmentIndex: i})
		}
	}
	return nil
}

// Solve the rotated alpha-rectangle envelope, matching pptxanchor highlight math.
func accentImageFrame(a AccentSpec, v Rect) (Rect, error) {
	if a.Mode == "underline" {
		w := v.Width / a.AlphaBounds.Width
		h := v.Height / a.AlphaBounds.Height
		if a.ContainerAspect > 0 {
			h = w / a.ContainerAspect
		}
		return Rect{X: v.X - a.AlphaBounds.X*w, Y: v.Y - a.AlphaBounds.Y*h, Width: w, Height: h}, nil
	}
	theta := a.RotationDeg * math.Pi / 180
	c, s := math.Abs(math.Cos(theta)), math.Abs(math.Sin(theta))
	det := c*c - s*s
	vw, vh := (c*v.Width-s*v.Height)/det, (c*v.Height-s*v.Width)/det
	if !positive(vw) || !positive(vh) {
		return Rect{}, fmt.Errorf("rotation cannot fit phrase envelope; choose another asset or manual staging")
	}
	w, h := vw/a.AlphaBounds.Width, vh/a.AlphaBounds.Height
	dx := (a.AlphaBounds.X + a.AlphaBounds.Width/2 - .5) * w
	dy := (a.AlphaBounds.Y + a.AlphaBounds.Height/2 - .5) * h
	cx := v.X + v.Width/2 - (math.Cos(theta)*dx - math.Sin(theta)*dy)
	cy := v.Y + v.Height/2 - (math.Sin(theta)*dx + math.Cos(theta)*dy)
	return Rect{X: cx - w/2, Y: cy - h/2, Width: w, Height: h}, nil
}
