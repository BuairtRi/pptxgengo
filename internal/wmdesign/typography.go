package wmdesign

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/font"
	ot "github.com/go-text/typesetting/font/opentype"
	"github.com/go-text/typesetting/font/opentype/tables"
	"github.com/go-text/typesetting/language"
	"github.com/go-text/typesetting/shaping"
	"golang.org/x/image/math/fixed"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

type FontIdentity struct {
	Family            string             `json:"authored_family"`
	Weight            int                `json:"authored_weight"`
	Italic            bool               `json:"authored_italic"`
	File              string             `json:"file"`
	SHA256            string             `json:"sha256"`
	FaceIndex         int                `json:"face_index"`
	LegacyFamily      string             `json:"legacy_family"`
	LegacyStyle       string             `json:"legacy_subfamily"`
	TypographicFamily string             `json:"typographic_family"`
	TypographicStyle  string             `json:"typographic_subfamily"`
	PostScript        string             `json:"postscript_name"`
	Typeface          string             `json:"pptx_typeface"`
	Bold              bool               `json:"pptx_bold"`
	NativeItalic      bool               `json:"pptx_italic"`
	Static            bool               `json:"static"`
	Axes              map[string]float32 `json:"axes"`
}
type resolvedFace struct {
	Identity FontIdentity
	Face     *font.Face
}
type Typography struct {
	faces                map[string]resolvedFace
	shaper               shaping.HarfbuzzShaper
	segmenter            shaping.Segmenter
	wrapper              shaping.LineWrapper
	engine               string
	anchors              map[string]VerticalAnchor
	densityAnchors       map[string]bool
	densitySupplementSHA string
}

func fontKey(f string, w int, i bool) string { return fmt.Sprintf("%s/%d/%t", f, w, i) }

// NewTypography inspects actual name, OS/2, head and variation tables. Numeric
// weights map to real static faces; directory order cannot synthesize a weight.
func NewTypography(root string) (*Typography, error) {
	t := &Typography{faces: map[string]resolvedFace{}, engine: Engine}
	entries, e := os.ReadDir(root)
	if e != nil {
		return nil, e
	}
	for _, entry := range entries {
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".ttf" && ext != ".otf" && ext != ".ttc" {
			continue
		}
		path, e := filepath.Abs(filepath.Join(root, entry.Name()))
		if e != nil {
			return nil, e
		}
		b, e := os.ReadFile(path)
		if e != nil {
			return nil, e
		}
		loaders, e := ot.NewLoaders(bytes.NewReader(b))
		if e != nil {
			return nil, e
		}
		for idx, l := range loaders {
			if _, e := l.RawTable(ot.MustNewTag("fvar")); e == nil {
				continue
			}
			raw, e := l.RawTable(ot.MustNewTag("name"))
			if e != nil {
				return nil, e
			}
			names, _, e := tables.ParseName(raw)
			if e != nil {
				return nil, e
			}
			raw, e = l.RawTable(ot.MustNewTag("OS/2"))
			if e != nil {
				return nil, e
			}
			os2, _, e := tables.ParseOs2(raw)
			if e != nil {
				return nil, e
			}
			desc, _ := font.Describe(l, nil)
			fam, sty := names.Name(16), names.Name(17)
			if fam == "" {
				fam = names.Name(1)
			}
			if sty == "" {
				sty = names.Name(2)
			}
			id := FontIdentity{Family: fam, Weight: int(os2.USWeightClass), Italic: desc.Aspect.Style == font.StyleItalic, File: path, SHA256: fmt.Sprintf("%x", sha256.Sum256(b)), FaceIndex: idx, LegacyFamily: names.Name(1), LegacyStyle: names.Name(2), TypographicFamily: fam, TypographicStyle: sty, PostScript: names.Name(6), Static: true}
			id.Axes = map[string]float32{}
			if math.Abs(float64(desc.Aspect.Stretch)-1) > .001 {
				continue
			}
			id.Typeface = id.LegacyFamily
			id.Bold = os2.FsSelection&(1<<5) != 0
			id.NativeItalic = id.Italic
			key := fontKey(fam, id.Weight, id.Italic)
			if _, ok := t.faces[key]; ok {
				return nil, fmt.Errorf("font.ambiguous_face: %s", key)
			}
			f, e := font.NewFont(l)
			if e != nil {
				return nil, e
			}
			t.faces[key] = resolvedFace{id, font.NewFace(f)}
		}
	}
	return t, nil
}
func (t *Typography) Resolve(s Style) (FontIdentity, error) {
	if s.Weight != 400 && s.Weight != 500 && s.Weight != 600 && s.Weight != 700 {
		return FontIdentity{}, fmt.Errorf("font.unsupported_weight: %d", s.Weight)
	}
	f, ok := t.faces[fontKey(s.Family, s.Weight, s.Italic)]
	if !ok {
		return FontIdentity{}, fmt.Errorf("font.unavailable_static_face: %s weight %d", s.Family, s.Weight)
	}
	return f.Identity, nil
}
func (t *Typography) Fonts() []FontIdentity {
	a := []FontIdentity{}
	for _, f := range t.faces {
		a = append(a, f.Identity)
	}
	sort.Slice(a, func(i, j int) bool {
		return fontKey(a[i].Family, a[i].Weight, a[i].Italic) < fontKey(a[j].Family, a[j].Weight, a[j].Italic)
	})
	return a
}

type oneFace struct{ face *font.Face }

func (f oneFace) ResolveFace(rune) *font.Face { return f.face }
func fixedPt(v float64) fixed.Int26_6         { return fixed.Int26_6(math.Round(v * 64)) }

type TextLine struct {
	Text     string  `json:"text"`
	Advance  float64 `json:"advance_pt"`
	Baseline float64 `json:"baseline_pt"`
}
type TextLayout struct {
	Original                string         `json:"original"`
	Displayed               string         `json:"displayed"`
	Style                   Style          `json:"style"`
	Font                    FontIdentity   `json:"font"`
	Lines                   []TextLine     `json:"lines"`
	AllocationHeight        float64        `json:"allocation_height_pt"`
	EstimatedOccupiedHeight float64        `json:"estimated_occupied_height_pt"`
	Features                map[string]int `json:"features"`
	NativeQualified         bool           `json:"native_qualified"`
	OccupiedTop             float64        `json:"occupied_top_pt,omitempty"`
	VerticalPolicy          string         `json:"vertical_policy,omitempty"`
	CalibrationSHA256       string         `json:"calibration_sha256,omitempty"`
}

// Measure uses Latin shaping, exact serialized tracking, metric kerning and
// requested point leading. The first baseline remains explicitly unqualified.
func (t *Typography) Measure(text string, s Style, width float64) (TextLayout, error) {
	if t.engine == CandidateEngine {
		return t.measureCandidate(text, s, width)
	}
	if width <= 0 || math.IsNaN(width) || math.IsInf(width, 0) {
		return TextLayout{}, fmt.Errorf("text.invalid_width")
	}
	id, e := t.Resolve(s)
	if e != nil {
		return TextLayout{}, e
	}
	face := t.faces[fontKey(s.Family, s.Weight, s.Italic)].Face
	display := strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
	if s.Case == "upper" {
		display = strings.ToUpper(display)
	}
	result := TextLayout{Original: text, Displayed: display, Style: s, Font: id, Features: map[string]int{"kern": 1, "liga": 1, "clig": 1}}
	// Tracking is applied per cluster. Disable discretionary clustering from
	// common ligatures when tracking is nonzero, matching the declared profile.
	if s.TrackingPt != 0 {
		result.Features["liga"] = 0
		result.Features["clig"] = 0
	}
	for _, p := range strings.Split(display, "\n") {
		runes := []rune(p)
		empty := len(runes) == 0
		if empty {
			runes = []rune{' '}
		}
		for _, r := range runes {
			if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) || !unicode.In(r, unicode.Latin, unicode.Common, unicode.Inherited) {
				return result, fmt.Errorf("text.unsupported_character: U+%04X", r)
			}
			if g, ok := face.NominalGlyph(r); !ok || g == 0 {
				return result, fmt.Errorf("font.missing_glyph: %s U+%04X", id.PostScript, r)
			}
		}
		features := []shaping.FontFeature{}
		for _, k := range []string{"kern", "liga", "clig"} {
			features = append(features, shaping.FontFeature{Tag: ot.MustNewTag(k), Value: uint32(result.Features[k])})
		}
		input := shaping.Input{Text: runes, RunEnd: len(runes), Direction: di.DirectionLTR, Face: face, Size: fixedPt(s.Size * 64), Language: language.NewLanguage("en"), FontFeatures: features}
		var shaped []shaping.Output
		for _, part := range t.segmenter.Split(input, oneFace{face}) {
			if part.Direction != di.DirectionLTR {
				return result, fmt.Errorf("text.unsupported_bidi")
			}
			part.FontFeatures = features
			o := t.shaper.Shape(part)
			div := func(v fixed.Int26_6) fixed.Int26_6 { return fixed.Int26_6(math.Round(float64(v) / 64)) }
			o.Size = fixedPt(s.Size)
			for i := range o.Glyphs {
				g := &o.Glyphs[i]
				if g.GlyphID == 0 {
					return result, fmt.Errorf("font.missing_shaped_glyph")
				}
				g.Advance = div(g.Advance)
				g.XAdvance = div(g.XAdvance)
				g.Width = div(g.Width)
				g.Height = div(g.Height)
				g.XOffset = div(g.XOffset)
				g.YOffset = div(g.YOffset)
				g.XBearing = div(g.XBearing)
				g.YBearing = div(g.YBearing)
				if i == len(o.Glyphs)-1 || g.ClusterIndex != o.Glyphs[i+1].ClusterIndex {
					g.Advance += fixedPt(s.TrackingPt)
					g.XAdvance += fixedPt(s.TrackingPt)
				}
			}
			o.RecomputeAdvance()
			shaped = append(shaped, o)
		}
		lines, _ := t.wrapper.WrapParagraphF(shaping.WrapConfig{Direction: di.DirectionLTR, BreakPolicy: shaping.WhenNecessary}, fixedPt(width+s.TrackingPt), runes, shaping.NewSliceIterator(shaped))
		if len(lines) == 0 {
			return result, fmt.Errorf("text.no_lines")
		}
		for _, line := range lines {
			start, end := len(runes), 0
			advance := 0.0
			for _, o := range line {
				start = min(start, o.Runes.Offset)
				end = max(end, o.Runes.Offset+o.Runes.Count)
				advance += float64(o.Advance) / 64
			}
			advance -= s.TrackingPt
			txt := string(runes[start:end])
			if empty {
				txt = ""
				advance = 0
			}
			result.Lines = append(result.Lines, TextLine{txt, advance, .9*s.Size + float64(len(result.Lines))*s.Leading})
			if advance > width+.02 {
				return result, fmt.Errorf("text.horizontal_overflow: %.3f exceeds %.3f", advance, width)
			}
		}
	}
	result.AllocationHeight = float64(len(result.Lines)) * s.Leading
	result.EstimatedOccupiedHeight = 1.2*s.Size + float64(len(result.Lines)-1)*s.Leading
	return result, nil
}
