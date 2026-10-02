package wmdesign

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"

	"github.com/buairtri/pptxgengo/pptx"
	"github.com/go-text/typesetting/di"
	ot "github.com/go-text/typesetting/font/opentype"
	"github.com/go-text/typesetting/language"
	"github.com/go-text/typesetting/shaping"
	"golang.org/x/image/math/fixed"
)

const RichTextContract = "pptxgengo.wmds-rich-text.v1"

// RichTextSpec is an explicit alternative to Node.Text. Empty paragraphs use
// an empty runs array; run text cannot contain paragraph separators or markup.
// The source typography token supplies family, size, leading, tracking and case.
type RichTextSpec struct {
	Paragraphs []RichParagraphSpec `json:"paragraphs"`
}
type RichParagraphSpec struct {
	Key  string        `json:"key"`
	Runs []RichRunSpec `json:"runs"`
}
type RichRunSpec struct {
	Key    string `json:"key"`
	Text   string `json:"text"`
	Weight int    `json:"weight,omitempty"`
	Ink    string `json:"ink,omitempty"`
}
type RichRunLayout struct {
	Key              string       `json:"key"`
	Original         string       `json:"original"`
	Displayed        string       `json:"displayed"`
	Style            Style        `json:"style"`
	Font             FontIdentity `json:"font"`
	Color            string       `json:"color"`
	Start            int          `json:"start_rune"`
	End              int          `json:"end_rune"`
	VerticalAnchored bool         `json:"vertical_anchored"`
	BaselineShift    float64      `json:"baseline_shift_pt,omitempty"`
}
type RichParagraphLayout struct {
	Key               string          `json:"key"`
	Runs              []RichRunLayout `json:"runs"`
	Displayed         string          `json:"displayed"`
	FirstLine         int             `json:"first_line"`
	LineCount         int             `json:"line_count"`
	Bullet            bool            `json:"bullet,omitempty"`
	ParagraphGapAfter float64         `json:"paragraph_gap_after_pt,omitempty"`
	LineBreaks        []int           `json:"explicit_line_break_runes,omitempty"`
}
type RichTextLayout struct {
	Contract        string                `json:"contract"`
	Paragraphs      []RichParagraphLayout `json:"paragraphs"`
	NativeQualified bool                  `json:"native_qualified"`
}

// richParagraphLines shapes all style spans into a single paragraph iterator.
// Wrapping sees the combined advances and global rune offsets, never the sum of
// independently wrapped strings. Equal-face spans coalesce for shaping so an
// ink-only split does not introduce an artificial ligature boundary in Go.
func (t *Typography) richParagraphLines(p RichParagraphLayout, width float64) ([]TextLine, error) {
	runes := []rune(p.Displayed)
	if len(runes) == 0 {
		return []TextLine{{}}, nil
	}
	features := []shaping.FontFeature{{Tag: ot.MustNewTag("kern"), Value: 0}, {Tag: ot.MustNewTag("liga"), Value: 1}, {Tag: ot.MustNewTag("clig"), Value: 1}}
	var shaped []shaping.Output
	for i := 0; i < len(p.Runs); {
		r := p.Runs[i]
		end, next := r.End, i+1
		for next < len(p.Runs) && p.Runs[next].Font.SHA256 == r.Font.SHA256 && p.Runs[next].Font.FaceIndex == r.Font.FaceIndex && p.Runs[next].Style.Weight == r.Style.Weight && p.Runs[next].Style.Size == r.Style.Size && p.Runs[next].Style.TrackingPt == r.Style.TrackingPt && p.Runs[next].BaselineShift == r.BaselineShift {
			end = p.Runs[next].End
			next++
		}
		face := t.faces[fontKey(r.Style.Family, r.Style.Weight, r.Style.Italic)].Face
		input := shaping.Input{Text: runes, RunStart: r.Start, RunEnd: end, Direction: di.DirectionLTR, Face: face, Size: fixedPt(r.Style.Size * 64), Language: language.NewLanguage("en"), FontFeatures: features}
		for _, part := range t.segmenter.Split(input, oneFace{face}) {
			if part.Direction != di.DirectionLTR {
				return nil, fmt.Errorf("text.unsupported_bidi")
			}
			part.FontFeatures = features
			o := t.shaper.Shape(part)
			quantize := func(v fixed.Int26_6) fixed.Int26_6 { return fixedPt(math.Round(float64(v)/4096*8) / 8 * 64) }
			for j := range o.Glyphs {
				g := &o.Glyphs[j]
				if g.GlyphID == 0 {
					return nil, fmt.Errorf("font.missing_shaped_glyph")
				}
				g.Advance, g.XAdvance = quantize(g.Advance), quantize(g.XAdvance)
				if j == len(o.Glyphs)-1 || g.ClusterIndex != o.Glyphs[j+1].ClusterIndex {
					g.Advance += fixedPt(r.Style.TrackingPt * 64)
					g.XAdvance += fixedPt(r.Style.TrackingPt * 64)
				}
			}
			o.RecomputeAdvance()
			shaped = append(shaped, o)
		}
		i = next
	}
	wrapped, _ := t.wrapper.WrapParagraphF(shaping.WrapConfig{Direction: di.DirectionLTR, BreakPolicy: shaping.WhenNecessary}, fixedPt(width*64), runes, shaping.NewSliceIterator(shaped))
	if len(wrapped) == 0 {
		return nil, fmt.Errorf("text.no_lines")
	}
	var lines []TextLine
	for _, line := range wrapped {
		start, end, advance := len(runes), 0, 0.0
		for _, o := range line {
			start, end = min(start, o.Runes.Offset), max(end, o.Runes.Offset+o.Runes.Count)
			advance += float64(o.Advance) / 4096
		}
		if advance > width+.02 {
			return nil, fmt.Errorf("text.horizontal_overflow: %.3f exceeds %.3f", advance, width)
		}
		lines = append(lines, TextLine{Text: string(runes[start:end]), Advance: advance})
	}
	return lines, nil
}

func richPrefix(p RichParagraphLayout, count int) RichParagraphLayout {
	q := RichParagraphLayout{Key: p.Key, Displayed: string([]rune(p.Displayed)[:count])}
	for _, r := range p.Runs {
		if r.Start >= count {
			break
		}
		r.End = min(r.End, count)
		r.Displayed = string([]rune(r.Displayed)[:r.End-r.Start])
		q.Runs = append(q.Runs, r)
	}
	return q
}

// planRich applies the existing source/zone/contrast rules, using a conservative
// common baseline and terminal envelope for the involved real static faces.
// Native mixed-run baseline and wrapping parity remain independently unqualified.
func (r *renderer) planRich(n Node, b, zone Rect, surface string, spec *RichTextSpec) (TextRecord, error) {
	var out TextRecord
	if r.typeEngine.engine != CandidateEngine {
		return out, fmt.Errorf("rich.requires_v2")
	}
	if spec == nil || len(spec.Paragraphs) == 0 || n.Text != "" || n.TextBlock != nil || n.Card != nil {
		return out, fmt.Errorf("rich.invalid_content_union: %s", n.ID)
	}
	if b.W <= 0 || b.W > 8191 || math.IsNaN(b.X+b.Y+b.W+b.H) || math.IsInf(b.X+b.Y+b.W+b.H, 0) || b.H < 0 {
		return out, fmt.Errorf("rich.invalid_geometry: %s", n.ID)
	}
	st, err := r.source.Style(n.Style)
	if err != nil {
		return out, err
	}
	baseID, err := r.typeEngine.Resolve(st)
	if err != nil {
		return out, err
	}
	st.TrackingPt = math.Round(st.TrackingPt*100) / 100
	role := n.Ink
	if role == "" {
		role = "primary"
	}
	baseColor, err := r.source.Ink(surface, role)
	if err != nil {
		return out, err
	}
	bg, err := r.source.Ink(surface, "bg")
	if err != nil {
		return out, err
	}
	baseMinimum := 4.5
	if st.Size >= 18 || st.Size >= 14 && st.Weight >= 700 {
		baseMinimum = 3
	}
	if contrast(baseColor, bg) < baseMinimum {
		return out, fmt.Errorf("text.insufficient_contrast: %s base defaults", n.ID)
	}
	align := n.Align
	if align == "" {
		align = "left"
	}
	if align != "left" && align != "center" && align != "right" {
		return out, fmt.Errorf("text.unsupported_align: %s", align)
	}
	l := TextLayout{Style: st, Font: baseID, Features: map[string]int{"kern": 0, "liga": 1, "clig": 1}, VerticalPolicy: "mixed-run candidate: common maximum face baseline; conservative maximum terminal height; native mixed-run parity unqualified"}
	rich := RichTextLayout{Contract: RichTextContract}
	baseline, terminal, allKnown := 0.0, 0.0, true
	anchor := func(id FontIdentity) bool {
		a, known := r.typeEngine.anchors[anchorKey(id.SHA256, st.Size, st.Leading)]
		bl, ht := .75*st.Leading, math.Max(st.Leading, 1.5*st.Size)
		if known {
			bl, ht = a.Baseline, a.TerminalHeight
		}
		baseline, terminal = math.Max(baseline, bl), math.Max(terminal, ht)
		allKnown = allKnown && known
		return known
	}
	anchor(baseID)
	keys := map[string]bool{}
	var originals, displays []string
	for _, p := range spec.Paragraphs {
		if !validPartKey(p.Key) || keys[p.Key] {
			return out, fmt.Errorf("rich.invalid_or_duplicate_paragraph_key: %s", p.Key)
		}
		keys[p.Key] = true
		q := RichParagraphLayout{Key: p.Key, Runs: []RichRunLayout{}}
		runKeys := map[string]bool{}
		original := ""
		for _, run := range p.Runs {
			if !validPartKey(run.Key) || runKeys[run.Key] || run.Text == "" {
				return out, fmt.Errorf("rich.invalid_or_duplicate_run_key_or_empty_text: %s/%s", p.Key, run.Key)
			}
			runKeys[run.Key] = true
			if strings.ContainsAny(run.Text, "\r\n") || strings.Contains(run.Text, "[[") || strings.Contains(run.Text, "]]") || strings.Contains(run.Text, "[^") {
				return out, fmt.Errorf("rich.run_break_or_markup: %s/%s", p.Key, run.Key)
			}
			rs := st
			if run.Weight != 0 {
				rs.Weight = run.Weight
			}
			id, err := r.typeEngine.Resolve(rs)
			if err != nil {
				return out, err
			}
			ink := role
			if run.Ink != "" {
				ink = run.Ink
			}
			color, err := r.source.Ink(surface, ink)
			if err != nil {
				return out, err
			}
			minimum := 4.5
			if rs.Size >= 18 || rs.Size >= 14 && rs.Weight >= 700 {
				minimum = 3
			}
			if contrast(color, bg) < minimum {
				return out, fmt.Errorf("text.insufficient_contrast: %s/%s/%s", n.ID, p.Key, run.Key)
			}
			display := run.Text
			if st.Case == "upper" {
				display = strings.ToUpper(display)
			}
			runes := []rune(display)
			if len(q.Runs) > 0 && (unicode.Is(unicode.Mn, runes[0]) || unicode.Is(unicode.Mc, runes[0]) || unicode.Is(unicode.Me, runes[0])) {
				return out, fmt.Errorf("rich.run_splits_combining_cluster: %s/%s", p.Key, run.Key)
			}
			face := r.typeEngine.faces[fontKey(rs.Family, rs.Weight, rs.Italic)].Face
			for _, ch := range runes {
				if unicode.IsControl(ch) || unicode.Is(unicode.Cf, ch) || !unicode.In(ch, unicode.Latin, unicode.Common, unicode.Inherited) {
					return out, fmt.Errorf("text.unsupported_character: U+%04X", ch)
				}
				if g, ok := face.NominalGlyph(ch); !ok || g == 0 {
					return out, fmt.Errorf("font.missing_glyph: %s U+%04X", id.PostScript, ch)
				}
			}
			start := len([]rune(q.Displayed))
			q.Runs = append(q.Runs, RichRunLayout{Key: run.Key, Original: run.Text, Displayed: display, Style: rs, Font: id, Color: color, Start: start, End: start + len(runes), VerticalAnchored: anchor(id)})
			q.Displayed += display
			original += run.Text
		}
		if strings.Contains(original, "[[") || strings.Contains(original, "]]") || strings.Contains(original, "[^") {
			return out, fmt.Errorf("rich.run_break_or_markup: %s/%s concatenated runs", n.ID, p.Key)
		}
		lines, err := r.typeEngine.richParagraphLines(q, b.W)
		if err != nil {
			return out, fmt.Errorf("%s/%s: %w", n.ID, p.Key, err)
		}
		// Guard both the chosen line and the next word's fit. This uses the
		// mixed faces, including a next word that crosses an authored run.
		consumed := 0
		for i, line := range lines {
			if math.Abs(b.W-line.Advance) <= .25 {
				return out, fmt.Errorf("rich.uncertain_wrap_boundary: %s/%s line %d", n.ID, p.Key, i+1)
			}
			if i+1 < len(lines) {
				next := []rune(lines[i+1].Text)
				j := 0
				for j < len(next) && unicode.IsSpace(next[j]) {
					j++
				}
				for j < len(next) && !unicode.IsSpace(next[j]) {
					j++
				}
				// Shape the line plus the first next word using local offsets.
				end := consumed + len([]rune(line.Text)) + j
				prefix := richPrefix(q, end)
				prefix.Displayed = string([]rune(prefix.Displayed)[consumed:])
				local := prefix.Runs[:0]
				for _, rr := range prefix.Runs {
					if rr.End <= consumed {
						continue
					}
					if rr.Start < consumed {
						rr.Displayed = string([]rune(rr.Displayed)[consumed-rr.Start:])
						rr.Start = consumed
					}
					rr.Start, rr.End = rr.Start-consumed, rr.End-consumed
					local = append(local, rr)
				}
				prefix.Runs = local
				candidate, e := r.typeEngine.richParagraphLines(prefix, 8191)
				if e != nil {
					return out, e
				}
				if len(candidate) == 1 && math.Abs(b.W-candidate[0].Advance) <= .25 {
					return out, fmt.Errorf("rich.uncertain_wrap_boundary: %s/%s next word after line %d", n.ID, p.Key, i+1)
				}
			}
			consumed += len([]rune(line.Text))
		}
		q.FirstLine, q.LineCount = len(l.Lines), len(lines)
		l.Lines = append(l.Lines, lines...)
		rich.Paragraphs = append(rich.Paragraphs, q)
		originals, displays = append(originals, original), append(displays, q.Displayed)
	}
	l.Original, l.Displayed = strings.Join(originals, "\n"), strings.Join(displays, "\n")
	if strings.TrimSpace(l.Displayed) == "" {
		return out, fmt.Errorf("rich.empty_visible_content: %s", n.ID)
	}
	if allKnown {
		l.CalibrationSHA256 = CandidateCalibrationSHA
	}
	first, last := -1, -1
	for i := range l.Lines {
		l.Lines[i].Baseline = baseline + float64(i)*st.Leading
		if strings.TrimSpace(l.Lines[i].Text) != "" {
			if first < 0 {
				first = i
			}
			last = i
		}
	}
	l.AllocationHeight = float64(len(l.Lines)) * st.Leading
	l.OccupiedTop = float64(first) * st.Leading
	lastHeight := terminal
	if last < len(l.Lines)-1 {
		lastHeight = st.Leading
	}
	l.EstimatedOccupiedHeight = float64(last-first)*st.Leading + lastHeight
	need := math.Max(l.AllocationHeight, l.OccupiedTop+l.EstimatedOccupiedHeight)
	if b.H == 0 {
		b.H = need
	}
	if need > b.H+.02 {
		return out, fmt.Errorf("text.vertical_overflow: %s needs %.3fpt, capacity %.3fpt", n.ID, need, b.H)
	}
	if !inside(b, zone) {
		return out, fmt.Errorf("node.outside_zone: %s", n.ID)
	}
	out = TextRecord{ID: n.ID, Rect: b, Color: baseColor, Align: align, Layout: l, Rich: &rich}
	return out, nil
}

// drawRich creates one editable native text box; candidateParagraphs replaces
// the temporary uniform text with the explicit resolved run/paragraph contract.
func (r *renderer) drawRich(tr TextRecord) {
	if r.err != nil {
		return
	}
	s, id := tr.Layout.Style, tr.Layout.Font
	opts := &pptx.TextPropsOptions{PositionProps: pos(tr.Rect), ObjectNameProps: pptx.ObjectNameProps{ObjectName: tr.ID}, TextBaseProps: pptx.TextBaseProps{FontFace: id.Typeface, FontSize: s.Size, Bold: &id.Bold, Italic: &id.NativeItalic, Color: tr.Color, Align: pptx.HAlign(tr.Align)}, CharSpacing: s.TrackingPt, LineSpacing: s.Leading, ParaSpaceBefore: zero(), ParaSpaceAfter: zero(), Margin: pptx.Margin{0}, Fit: "none", Valign: pptx.VAlign("top")}
	r.err = r.slide.AddText([]pptx.TextProps{{Text: tr.Layout.Displayed}}, opts)
	*r.records = append(*r.records, tr)
}

// richParagraphXML shares the inherited exact-leading paragraph properties but
// writes explicit run properties and complete base defaults for empty/trailing
// paragraphs. Source-owned footnote runs may carry an explicit baseline shift.
func richParagraphXML(tr TextRecord, pp []byte) []byte {
	escape := func(s string) string { var b bytes.Buffer; xml.EscapeText(&b, []byte(s)); return b.String() }
	boolean := func(v bool) string {
		if v {
			return "1"
		}
		return "0"
	}
	properties := func(s Style, id FontIdentity, color string) (string, string) {
		attrs := ` lang="en-US" sz="` + strconv.Itoa(int(math.Round(s.Size*100))) + `" spc="` + strconv.Itoa(int(math.Round(s.TrackingPt*100))) + `" kern="0" b="` + boolean(id.Bold) + `" i="` + boolean(id.NativeItalic) + `" dirty="0"`
		children := `<a:solidFill><a:srgbClr val="` + escape(color) + `"/></a:solidFill><a:latin typeface="` + escape(id.Typeface) + `"/><a:ea typeface="` + escape(id.Typeface) + `"/><a:cs typeface="` + escape(id.Typeface) + `"/>`
		return attrs, children
	}
	var out strings.Builder
	for _, p := range tr.Rich.Paragraphs {
		out.WriteString("<a:p>")
		out.Write(pp)
		position, nextBreak := 0, 0
		for _, run := range p.Runs {
			attrs, children := properties(run.Style, run.Font, run.Color)
			if run.BaselineShift != 0 {
				attrs += ` baseline="` + strconv.Itoa(int(math.Round(100000*run.BaselineShift/run.Style.Size))) + `"`
			}
			runes := []rune(run.Displayed)
			for start := 0; start < len(runes); {
				if nextBreak < len(p.LineBreaks) && p.LineBreaks[nextBreak] == position {
					out.WriteString("<a:br><a:rPr" + attrs + ">" + children + "</a:rPr></a:br>")
					nextBreak++
				}
				end := len(runes)
				if nextBreak < len(p.LineBreaks) {
					end = min(end, start+p.LineBreaks[nextBreak]-position)
				}
				out.WriteString("<a:r><a:rPr" + attrs + ">" + children + "</a:rPr><a:t xml:space=\"preserve\">" + escape(string(runes[start:end])) + "</a:t></a:r>")
				position += end - start
				start = end
			}
		}
		attrs, children := properties(tr.Layout.Style, tr.Layout.Font, tr.Color)
		out.WriteString("<a:endParaRPr" + attrs + ">" + children + "</a:endParaRPr></a:p>")
	}
	return []byte(out.String())
}
