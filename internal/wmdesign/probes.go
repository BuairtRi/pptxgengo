package wmdesign

import (
	"crypto/sha256"
	"fmt"
	"math"
	"path/filepath"
	"strings"

	"github.com/buairtri/pptxgengo/pptx"
)

// TypographyProbes is a qualification fixture, not an authoring document. Each
// slide contains three horizontal text objects using the production renderer.
// There are no labels, logos or master text to contaminate native observations.
type TypographyProbe struct {
	ID     string     `json:"id"`
	Case   string     `json:"case"`
	Slide  int        `json:"slide"`
	Record TextRecord `json:"record"`
}
type RejectionControl struct {
	ID       string `json:"id"`
	Expected string `json:"expected_error"`
	Observed string `json:"observed_error"`
}
type ProbeManifest struct {
	Schema        string             `json:"schema"`
	Profile       string             `json:"profile"`
	Engine        string             `json:"engine"`
	DeckSHA256    string             `json:"deck_sha256"`
	Sources       []SourceFile       `json:"sources"`
	Fonts         []FontIdentity     `json:"fonts"`
	Probes        []TypographyProbe  `json:"probes"`
	Rejections    []RejectionControl `json:"rejections"`
	Qualification string             `json:"qualification"`
}

func TypographyProbes(bundle, override string) ([]byte, ProbeManifest, error) {
	return TypographyProbesWithEngine(bundle, override, Engine)
}

func TypographyProbesWithEngine(bundle, override, engine string) ([]byte, ProbeManifest, error) {
	m := ProbeManifest{Schema: "pptxgengo.wmds-typography-probes.v1", Profile: ProfileForEngine(engine), Engine: engine, Qualification: "capture_pending"}
	s, e := Load(bundle, override)
	if e != nil {
		return nil, m, e
	}
	t, e := NewTypographyEngine(filepath.Join(bundle, "fonts"), engine)
	if e != nil {
		return nil, m, e
	}
	m.Sources, m.Fonts = s.Files, t.Fonts()
	p := pptx.New()
	p.DefineLayout("WMDS", 960.0/72, 540.0/72)
	if e = p.SetLayout("WMDS"); e != nil {
		return nil, m, e
	}
	p.Title = "WMDS native typography controls"
	p.Author = "West Monroe"
	p.Theme = pptx.ThemeProps{HeadFontFace: "IBM Plex Sans SemiBold", BodyFontFace: "IBM Plex Sans"}
	r := renderer{source: s, typeEngine: t, bundle: bundle, pres: p}
	addFrame := func(st Style, kind, text string, width, height float64) error {
		i := len(m.Probes)
		if i%3 == 0 {
			r.slide = p.AddSlide()
			r.slide.AddNotes("Qualification controls: synthetic Latin text. Match named objects to probes.json; no native qualification is implied by generation.")
		}
		id := fmt.Sprintf("probe-%03d.%s.%s", i+1, st.ID, kind)
		var records []TextRecord
		r.records = &records
		r.text(id, text, st, Rect{60, 36 + 168*float64(i%3), width, height}, "000000", "left", 0)
		if r.err != nil {
			return r.err
		}
		m.Probes = append(m.Probes, TypographyProbe{id, kind, i/3 + 1, records[0]})
		return nil
	}
	add := func(st Style, kind, text string, width float64) error { return addFrame(st, kind, text, width, 144) }
	// Retain all original v1 frame widths to make replay comparisons meaningful.
	boundaryEngine := t
	if engine == CandidateEngine {
		boundaryEngine, e = NewTypography(filepath.Join(bundle, "fonts"))
		if e != nil {
			return nil, m, e
		}
	}
	for _, st := range s.Tokens.Type {
		for _, c := range []struct{ kind, text string }{{"plain", "Agpq AV 012"}, {"hard-two", "Agpq AV\nType 012"}, {"ligature-kerning", "office affine AVATAR"}} {
			if e = add(st, c.kind, c.text, 840); e != nil {
				return nil, m, e
			}
		}
		boundary, e := boundaryEngine.Measure("AV office", st, 840)
		if e != nil {
			return nil, m, e
		}
		for _, delta := range []float64{-.75, .75} {
			kind := "boundary-below"
			if delta > 0 {
				kind = "boundary-above"
			}
			if e = add(st, kind, "AV office", boundary.Lines[0].Advance+delta); e != nil {
				return nil, m, e
			}
		}
		// Exercise the same renderer capacity check as authored content. Keep
		// negative fixtures out of the deck: deliberately clipped text is not
		// a qualified positive control.
		l, e := t.Measure("Agpq\nType", st, 840)
		if e != nil {
			return nil, m, e
		}
		var records []TextRecord
		negative := renderer{source: s, typeEngine: t, pres: p, slide: r.slide, records: &records}
		negative.text("negative", "Agpq\nType", st, Rect{60, 36, 840, math.Max(l.AllocationHeight, l.EstimatedOccupiedHeight) - 1}, "000000", "left", 0)
		if negative.err == nil || !strings.Contains(negative.err.Error(), "text.vertical_overflow") {
			return nil, m, fmt.Errorf("probe.expected_vertical_rejection: %s", st.ID)
		}
		m.Rejections = append(m.Rejections, RejectionControl{st.ID + ".height-minus-one", "text.vertical_overflow", negative.err.Error()})
	}
	body, _ := s.Style("body")
	for _, c := range []struct{ kind, text string }{{"empty-paragraph", "Agpq\n\nType"}, {"trailing-break", "Agpq\n"}} {
		if e = add(body, c.kind, c.text, 840); e != nil {
			return nil, m, e
		}
	}
	// The styles use five faces. Cover the three remaining bundled faces too.
	for _, c := range []struct {
		family string
		weight int
	}{{"IBM Plex Sans", 500}, {"IBM Plex Sans", 700}, {"IBM Plex Mono", 700}} {
		st := body
		st.Family = c.family
		st.Weight = c.weight
		st.ID = fmt.Sprintf("identity-%s-%d", strings.ReplaceAll(c.family, " ", "-"), c.weight)
		if e = add(st, "face-identity", "Agpq AV 012", 840); e != nil {
			return nil, m, e
		}
	}
	if engine == CandidateEngine {
		for _, st := range s.Tokens.Type {
			boundary, e := t.Measure("AV office", st, 840)
			if e != nil {
				return nil, m, e
			}
			for _, delta := range []float64{-.125, .125} {
				kind := "v2-boundary-below"
				if delta > 0 {
					kind = "v2-boundary-above"
				}
				if e = add(st, kind, "AV office", boundary.Lines[0].Advance+delta); e != nil {
					return nil, m, e
				}
			}
			heldout := "HÉÅgj Çy 89"
			if e = add(st, "heldout", heldout, 840); e != nil {
				return nil, m, e
			}
			l, e := t.Measure(heldout, st, 840)
			if e != nil {
				return nil, m, e
			}
			if e = addFrame(st, "heldout-tight-height", heldout, 840, math.Max(l.AllocationHeight, l.OccupiedTop+l.EstimatedOccupiedHeight)); e != nil {
				return nil, m, e
			}
		}
		for _, c := range []struct{ kind, text string }{
			{"leading-blank", "\nAgpq"}, {"two-leading-blanks", "\n\nAgpq"}, {"two-trailing-blanks", "Agpq\n\n"},
			{"blanks-only", "\n"}, {"trailing-spaces", "Agpq  "}, {"punctuation-latin", "AV To fi ffi — naïve café"},
			{"mixed-blanks", "\nAgpq\n\nType\n"},
		} {
			if e = add(body, c.kind, c.text, 840); e != nil {
				return nil, m, e
			}
		}
		eyebrow, _ := s.Style("eyebrow")
		if e = add(eyebrow, "long-tracking", "AV0123456789 AV0123456789 AV0123456789", 120); e != nil {
			return nil, m, e
		}
	}
	for _, c := range []struct{ kind, text, expected string }{{"bidi", "مرحبا", "text.unsupported_character"}, {"tab", "A\tB", "text.unsupported_character"}, {"emphasis", "[[word]]", "text.unsupported_emphasis_or_footnote"}, {"footnote", "word[^1]", "text.unsupported_emphasis_or_footnote"}} {
		var records []TextRecord
		negative := renderer{source: s, typeEngine: t, pres: p, slide: r.slide, records: &records}
		negative.text("negative", c.text, body, Rect{60, 36, 840, 144}, "000000", "left", 0)
		if negative.err == nil || !strings.Contains(negative.err.Error(), c.expected) {
			return nil, m, fmt.Errorf("probe.expected_rejection: %s", c.kind)
		}
		m.Rejections = append(m.Rejections, RejectionControl{c.kind, c.expected, negative.err.Error()})
	}
	raw, e := p.Write()
	if e != nil {
		return nil, m, e
	}
	if engine == CandidateEngine {
		records := map[int][]TextRecord{}
		for _, q := range m.Probes {
			records[q.Slide] = append(records[q.Slide], q.Record)
		}
		raw, e = candidateParagraphs(raw, records)
		if e != nil {
			return nil, m, e
		}
	}
	raw, e = explicitTypography(raw)
	if e != nil {
		return nil, m, e
	}
	m.DeckSHA256 = fmt.Sprintf("%x", sha256.Sum256(raw))
	return raw, m, nil
}
