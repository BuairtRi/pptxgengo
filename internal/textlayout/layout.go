package textlayout

import (
	"fmt"
	"math"
	"strings"
	"unicode"

	"github.com/buairtri/pptxgengo/internal/compose"
	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/language"
	"github.com/go-text/typesetting/shaping"
	"golang.org/x/image/math/fixed"
)

const Engine = "go-text-prototype.v4"
const Library = "github.com/go-text/typesetting@v0.3.5"
const advanceQuantumPt = 1.0 / 8
const lineBoxEm = 1.2
const baselineEm = .9

type Report struct {
	Schema             string                   `json:"schema"`
	Engine             string                   `json:"engine"`
	Library            string                   `json:"library"`
	SpecSHA256         string                   `json:"spec_sha256,omitempty"`
	PowerPointVerified bool                     `json:"powerpoint_verified"`
	HeightPolicy       string                   `json:"height_policy"`
	AdvancePolicy      string                   `json:"advance_policy"`
	Fonts              []FontRecord             `json:"fonts"`
	Warnings           []string                 `json:"warnings,omitempty"`
	Measurements       compose.Measurements     `json:"measurements"`
	Requests           map[string]RequestLayout `json:"requests"`
}

type RequestLayout struct {
	Lines            []Line            `json:"lines"`
	GlyphCount       int               `json:"glyph_count"`
	BoundaryWarnings []BoundaryWarning `json:"boundary_warnings,omitempty"`
	HeightWarnings   []string          `json:"height_warnings,omitempty"`
}

// BoundaryWarning identifies a line with little clearance relative to the
// accumulated per-glyph rounding quantum. This is a review heuristic, not a
// bound on PowerPoint's error or a reason to change the predicted line break.
type BoundaryWarning struct {
	LineIndex    int     `json:"line_index"`
	ClearancePt  float64 `json:"clearance_pt"`
	ReviewBandPt float64 `json:"review_band_pt"`
}
type Line struct {
	Text            string        `json:"text"`
	StartRune       int           `json:"start_rune"`
	EndRune         int           `json:"end_rune"`
	XPt             float64       `json:"x_pt"`
	YPt             float64       `json:"y_pt"`
	AdvancePt       float64       `json:"advance_pt"`
	HeightPt        float64       `json:"height_pt"`
	LayoutAdvancePt float64       `json:"layout_advance_pt"`
	FontSizePt      float64       `json:"max_font_size_pt"`
	BaselinePt      float64       `json:"baseline_pt"`
	InkBounds       *compose.Rect `json:"ink_bounds,omitempty"`
}

type EngineLayout struct {
	resolver  *Resolver
	shaper    shaping.HarfbuzzShaper
	segmenter shaping.Segmenter
	wrapper   shaping.LineWrapper
}

func New(roots []string) (*EngineLayout, error) {
	r, err := NewResolver(roots)
	if err != nil {
		return nil, err
	}
	return &EngineLayout{resolver: r}, nil
}

func (e *EngineLayout) Measure(requests []compose.ProbeRequest) (Report, error) {
	report := Report{Schema: "pptxgengo.go-layout.v1", Engine: Engine, Library: Library, HeightPolicy: "1.2 em line allocation scaled by line-spacing multiple; character boxes include paragraph spacing before/after between paragraphs; expanded terminal allocation is a conservative estimate requiring native review; baseline retains 75% extra leading above; nonblank box union excludes trailing blanks and final paragraph spacing; ink bounds are diagnostic", AdvancePolicy: "shaped glyph advances rounded to nearest 1/8 pt before wrapping", Measurements: compose.Measurements{ByRequestID: map[string]compose.Measurement{}}, Requests: map[string]RequestLayout{}}
	for _, q := range requests {
		if _, ok := report.Requests[q.ID]; ok {
			return Report{}, fmt.Errorf("duplicate text request %q", q.ID)
		}
		m, detail, err := e.measure(q)
		if err != nil {
			return Report{}, fmt.Errorf("request %s: %w", q.ID, err)
		}
		report.Measurements.ByRequestID[q.ID] = m
		report.Requests[q.ID] = detail
		for _, warning := range detail.HeightWarnings {
			report.Warnings = append(report.Warnings, fmt.Sprintf("request %s: %s", q.ID, warning))
		}
		for _, warning := range detail.BoundaryWarnings {
			report.Warnings = append(report.Warnings, fmt.Sprintf("request %s line %d is near a wrap boundary (%.4f pt clearance, %.4f pt review band); native wrapping may differ", q.ID, warning.LineIndex+1, warning.ClearancePt, warning.ReviewBandPt))
		}
	}
	report.Fonts = e.resolver.Records()
	report.Warnings = append(report.Warnings, e.resolver.Warnings...)
	return report, nil
}

type oneFace struct{ face *font.Face }

func (f oneFace) ResolveFace(rune) *font.Face { return f.face }

func toFixed(v float64) fixed.Int26_6 { return fixed.Int26_6(math.Round(v * 64)) }
func points(v fixed.Int26_6) float64  { return float64(v) / 64 }

// Shape at 64x scale, then return point metrics. The library rounds input sizes
// up to whole units internally; scaling preserves fractional authored sizes.
func (e *EngineLayout) shape(text []rune, start, end int, run compose.RunSpec) ([]shaping.Output, error) {
	if !compose.ValidFontFace(run.FontFace) || run.FontSizePt <= 0 || math.IsNaN(run.FontSizePt) || math.IsInf(run.FontSizePt, 0) || run.FontSizePt > 4096 {
		return nil, fmt.Errorf("invalid font family or point size (maximum 4096pt)")
	}
	f, err := e.resolver.resolve(run.FontFace, run.Bold, run.Italic)
	if err != nil {
		return nil, err
	}
	for _, r := range text[start:end] {
		if r == '\t' {
			return nil, fmt.Errorf("tab stops are not supported by the Go prototype")
		}
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) || !unicode.In(r, unicode.Latin, unicode.Common, unicode.Inherited) {
			return nil, fmt.Errorf("character U+%04X is outside the Go prototype's horizontal Latin scope", r)
		}
		if gid, ok := f.face.NominalGlyph(r); !ok || gid == 0 {
			return nil, fmt.Errorf("font %q style %s has no glyph for U+%04X (no glyph fallback)", run.FontFace, f.record.Style, r)
		}
	}
	input := shaping.Input{Text: text, RunStart: start, RunEnd: end, Direction: di.DirectionLTR, Face: f.face, Size: toFixed(run.FontSizePt * 64), Language: language.NewLanguage("en")}
	var out []shaping.Output
	for _, part := range e.segmenter.Split(input, oneFace{f.face}) {
		if part.Direction != di.DirectionLTR {
			return nil, fmt.Errorf("bidirectional layout is not supported by the Go prototype")
		}
		o := e.shaper.Shape(part)
		if len(o.Glyphs) == 0 && end > start {
			return nil, fmt.Errorf("shaper returned no glyphs")
		}
		div := func(v fixed.Int26_6) fixed.Int26_6 { return fixed.Int26_6(math.Round(float64(v) / 64)) }
		// PowerPoint's exported Latin glyph positions follow a 1/8 pt advance
		// grid. Round shaped advances on that grid before summing a line, rather
		// than rounding the final width or adding a family-specific allowance.
		advance := func(v fixed.Int26_6) fixed.Int26_6 {
			pt := float64(v) / (64 * 64)
			return toFixed(math.Round(pt/advanceQuantumPt) * advanceQuantumPt)
		}
		o.Size = toFixed(run.FontSizePt)
		o.Advance = div(o.Advance)
		o.LineBounds = shaping.Bounds{Ascent: div(o.LineBounds.Ascent), Descent: div(o.LineBounds.Descent), Gap: div(o.LineBounds.Gap)}
		o.GlyphBounds = shaping.Bounds{Ascent: div(o.GlyphBounds.Ascent), Descent: div(o.GlyphBounds.Descent)}
		for i := range o.Glyphs {
			g := &o.Glyphs[i]
			if g.GlyphID == 0 {
				return nil, fmt.Errorf("shaper returned missing glyph")
			}
			g.Width = div(g.Width)
			g.Height = div(g.Height)
			g.XBearing = div(g.XBearing)
			g.YBearing = div(g.YBearing)
			g.Advance = advance(g.Advance)
			g.XAdvance = advance(g.XAdvance)
			g.YAdvance = advance(g.YAdvance)
			g.XOffset = div(g.XOffset)
			g.YOffset = div(g.YOffset)
		}
		o.RecomputeAdvance()
		out = append(out, o)
	}
	return out, nil
}

func union(a, b compose.Rect) compose.Rect {
	x, y := math.Min(a.X, b.X), math.Min(a.Y, b.Y)
	return compose.Rect{X: x, Y: y, Width: math.Max(a.X+a.Width, b.X+b.Width) - x, Height: math.Max(a.Y+a.Height, b.Y+b.Height) - y}
}

func (e *EngineLayout) measure(q compose.ProbeRequest) (compose.Measurement, RequestLayout, error) {
	if q.TextWidthPt <= 0 || math.IsNaN(q.TextWidthPt) || math.IsInf(q.TextWidthPt, 0) || q.TextWidthPt > 1e6 {
		return compose.Measurement{}, RequestLayout{}, fmt.Errorf("invalid text width")
	}
	// Native character bounds and arbitrary phrase accents require a separate
	// cluster-to-character mapping; fail explicitly instead of estimating them.
	if len(q.PhraseRequests) > 0 {
		return compose.Measurement{}, RequestLayout{}, fmt.Errorf("phrase bounds and measured phrase accents are not supported by the Go prototype")
	}
	paragraphs := q.Paragraphs
	if len(paragraphs) == 0 {
		text := strings.ReplaceAll(strings.ReplaceAll(q.Text, "\r\n", "\n"), "\r", "\n")
		for _, line := range strings.Split(text, "\n") {
			paragraphs = append(paragraphs, compose.ParagraphSpec{Align: q.Align, Runs: []compose.RunSpec{{Text: line, FontFace: q.FontFace, FontSizePt: q.FontSizePt, Bold: q.Bold}}})
		}
	}
	var detail RequestLayout
	heightWarning := func(message string) {
		for _, warning := range detail.HeightWarnings {
			if warning == message {
				return
			}
		}
		detail.HeightWarnings = append(detail.HeightWarnings, message)
	}
	var total compose.Rect
	var occupiedTop, occupiedBottom float64
	occupied := false
	initialized := false
	add := func(r compose.Rect) {
		if !initialized {
			total = r
			initialized = true
		} else {
			total = union(total, r)
		}
	}
	y := 0.0
	runeOffset := 0
	for pi, p := range paragraphs {
		if len(p.Runs) == 0 {
			return compose.Measurement{}, detail, fmt.Errorf("paragraph requires runs")
		}
		align := p.Align
		if align == "" {
			align = "left"
		}
		if align != "left" && align != "center" && align != "right" {
			return compose.Measurement{}, detail, fmt.Errorf("unsupported alignment %q", align)
		}
		multiple := p.LineSpacingMultiple
		if multiple == 0 {
			multiple = 1
		}
		if multiple < .5 || multiple > 4 || math.IsNaN(multiple) {
			return compose.Measurement{}, detail, fmt.Errorf("invalid paragraph line spacing")
		}
		if p.SpaceBeforePt < 0 || p.SpaceAfterPt < 0 || math.IsNaN(p.SpaceBeforePt+p.SpaceAfterPt) || math.IsInf(p.SpaceBeforePt+p.SpaceAfterPt, 0) {
			return compose.Measurement{}, detail, fmt.Errorf("invalid paragraph spacing")
		}
		var text []rune
		for _, run := range p.Runs {
			if strings.ContainsAny(run.Text, "\r\n") {
				return compose.Measurement{}, detail, fmt.Errorf("rich runs cannot contain hard breaks")
			}
			text = append(text, []rune(run.Text)...)
			if !strings.EqualFold(run.FontFace, "Arial") && !strings.EqualFold(run.FontFace, "IBM Plex Sans") && !strings.EqualFold(run.FontFace, "IBM Plex Mono") {
				heightWarning("native character-height reference covers Arial, IBM Plex Sans and IBM Plex Mono; other families need native review")
			}
		}
		if multiple != 1 {
			heightWarning("nondefault line spacing uses a full terminal line allocation; captured expanded-spacing heights are conservatively overestimated, and other sizes/styles/multiples need native review")
		}
		var shaped []shaping.Output
		start := 0
		for _, run := range p.Runs {
			end := start + len([]rune(run.Text))
			if end > start {
				out, err := e.shape(text, start, end, run)
				if err != nil {
					return compose.Measurement{}, detail, err
				}
				shaped = append(shaped, out...)
			}
			start = end
		}
		// Empty hard-break paragraphs still occupy one font line.
		empty := len(text) == 0
		if empty {
			out, err := e.shape([]rune{' '}, 0, 1, p.Runs[0])
			if err != nil {
				return compose.Measurement{}, detail, err
			}
			shaped = out
		}
		indent := 0.0
		var bullet []shaping.Output
		if p.Bullet != nil {
			b := p.Bullet
			if align != "left" || b.MarginLeftPt <= 0 || b.HangingPt < 0 || b.HangingPt > b.MarginLeftPt || math.IsNaN(b.MarginLeftPt+b.HangingPt) {
				return compose.Measurement{}, detail, fmt.Errorf("invalid bullet indent")
			}
			indent = b.MarginLeftPt
			var err error
			bullet, err = e.shape([]rune(b.Character), 0, len([]rune(b.Character)), p.Runs[0])
			if err != nil {
				return compose.Measurement{}, detail, err
			}
			var width float64
			for _, o := range bullet {
				width += points(o.Advance)
			}
			if width > b.HangingPt {
				return compose.Measurement{}, detail, fmt.Errorf("bullet glyph exceeds hanging indent")
			}
		}
		if indent >= q.TextWidthPt {
			return compose.Measurement{}, detail, fmt.Errorf("bullet indent leaves no text width")
		}
		var lines []shaping.Line
		if empty {
			lines = []shaping.Line{shaped}
		} else {
			lines, _ = e.wrapper.WrapParagraphF(shaping.WrapConfig{Direction: di.DirectionLTR, BreakPolicy: shaping.WhenNecessary}, toFixed(q.TextWidthPt-indent), text, shaping.NewSliceIterator(shaped))
		}
		if len(lines) == 0 {
			return compose.Measurement{}, detail, fmt.Errorf("line wrapper returned no lines")
		}
		for li, line := range lines {
			advance, size := 0.0, 0.0
			glyphCount := 0
			a, b := len(text), 0
			for _, o := range line {
				glyphCount += len(o.Glyphs)
				advance += points(o.Advance)
				size = math.Max(size, points(o.Size))
				a = min(a, o.Runes.Offset)
				b = max(b, o.Runes.Offset+o.Runes.Count)
			}
			if empty {
				advance = 0
				a = 0
				b = 0
			}
			for _, o := range bullet {
				size = math.Max(size, points(o.Size))
			}
			// Fresh native character snapshots include leading and paragraph
			// spacing in their boxes. Glyph baselines describe a different contract.
			baseHeight := lineBoxEm * size
			layoutAdvance := baseHeight * multiple
			before, after := 0.0, 0.0
			if pi > 0 && li == 0 {
				before = p.SpaceBeforePt
			}
			if pi+1 < len(paragraphs) && li+1 == len(lines) {
				after = p.SpaceAfterPt
			}
			height := before + layoutAdvance + after
			lineTop := y
			if height <= 0 {
				return compose.Measurement{}, detail, fmt.Errorf("font has no usable line metrics")
			}
			x := indent
			available := q.TextWidthPt - indent
			clearance := available - advance
			band := float64(glyphCount) * advanceQuantumPt / 2
			if !empty && clearance >= 0 && clearance <= band {
				detail.BoundaryWarnings = append(detail.BoundaryWarnings, BoundaryWarning{LineIndex: len(detail.Lines), ClearancePt: clearance, ReviewBandPt: band})
			}
			if align == "center" {
				x += (available - advance) / 2
			}
			if align == "right" {
				x += available - advance
			}
			baseline := lineTop + before + .75*(layoutAdvance-baseHeight) + baselineEm*size
			row := Line{Text: string(text[a:b]), StartRune: runeOffset + a, EndRune: runeOffset + b, XPt: x, YPt: lineTop, AdvancePt: advance, HeightPt: height, LayoutAdvancePt: layoutAdvance, FontSizePt: size, BaselinePt: baseline}
			add(compose.Rect{X: x, Y: lineTop, Width: advance, Height: height})
			if strings.TrimSpace(row.Text) != "" || (p.Bullet != nil && li == 0) {
				if !occupied {
					occupiedTop, occupiedBottom, occupied = lineTop, lineTop+height, true
				} else {
					occupiedTop = math.Min(occupiedTop, lineTop)
					occupiedBottom = math.Max(occupiedBottom, lineTop+height)
				}
			}
			draw := func(outputs []shaping.Output, dot float64) {
				for _, o := range outputs {
					for _, g := range o.Glyphs {
						detail.GlyphCount++
						if g.GlyphID != font.EmptyGlyph && g.Width != 0 && g.Height != 0 {
							box := compose.Rect{X: dot + points(g.XOffset+g.XBearing), Y: baseline - points(g.YOffset+g.YBearing), Width: points(g.Width), Height: -points(g.Height)}
							if row.InkBounds == nil {
								row.InkBounds = &box
							} else {
								u := union(*row.InkBounds, box)
								row.InkBounds = &u
							}
							add(box)
						}
						dot += points(g.Advance)
					}
				}
			}
			if !empty {
				draw(line, x)
			}
			if p.Bullet != nil && li == 0 {
				bx := indent - p.Bullet.HangingPt
				draw(bullet, bx)
				var ba float64
				for _, o := range bullet {
					ba += points(o.Advance)
				}
				add(compose.Rect{X: bx, Y: lineTop, Width: ba, Height: height})
			}
			if row.InkBounds != nil && (row.InkBounds.Y < lineTop-1.0/64 || row.InkBounds.Y+row.InkBounds.Height > lineTop+height+1.0/64) {
				heightWarning("glyph ink extends beyond the predicted character box; inspect the native rendering")
			}
			detail.Lines = append(detail.Lines, row)
			y += height
		}
		runeOffset += len(text)
		if pi+1 < len(paragraphs) {
			runeOffset++
		}
	}
	// Match the native non-whitespace character-box contract: blank lines affect
	// the distance between occupied rows, but trailing blanks/spacing do not add
	// a visible box. Keep ink bounds separately; they are not TextRange bounds.
	if !initialized {
		return compose.Measurement{}, detail, fmt.Errorf("text has no layout")
	}
	if !occupied {
		occupiedTop, occupiedBottom = 0, y
	}
	return compose.Measurement{OffsetXPt: total.X, OffsetYPt: occupiedTop, RenderedWidthPt: math.Max(total.Width, 1.0/64), RenderedHeightPt: occupiedBottom - occupiedTop}, detail, nil
}
