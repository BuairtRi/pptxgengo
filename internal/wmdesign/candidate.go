package wmdesign

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/go-text/typesetting/di"
	ot "github.com/go-text/typesetting/font/opentype"
	"github.com/go-text/typesetting/language"
	"github.com/go-text/typesetting/shaping"
	"golang.org/x/image/math/fixed"
)

const CandidateEngine = "wmds-go-foundation.v2"
const CandidateProfile = "wmds-native.v2-candidate"

// The anchors are estimates adopted from the v1 controls on one PowerPoint
// environment, not proof of file identity or out-of-sample native parity.
const CandidateCalibrationSHA = "fd9b1b8704e3ce26404e1acf8af2c58b4790a1ff1f025d3ee4c15b2868d77f0a"

type VerticalAnchor struct {
	FontSHA        string  `json:"font_sha256"`
	Size           float64 `json:"size_pt"`
	Leading        float64 `json:"leading_pt"`
	Baseline       float64 `json:"first_baseline_pt"`
	TerminalHeight float64 `json:"terminal_character_height_pt"`
	ProbeID        string  `json:"source_probe_id"`
}
type CandidateCalibration struct {
	Schema  string           `json:"schema"`
	Anchors []VerticalAnchor `json:"anchors"`
}

func anchorKey(sha string, size, leading float64) string {
	return fmt.Sprintf("%s/%g/%g", sha, size, leading)
}
func NewTypographyEngine(root, engine string) (*Typography, error) {
	if engine != Engine && engine != CandidateEngine {
		return nil, fmt.Errorf("text.unsupported_engine: %s", engine)
	}
	t, e := NewTypography(root)
	if e != nil {
		return nil, e
	}
	if engine == Engine {
		return t, nil
	}
	path := filepath.Join(root, "..", "..", "typography-v2-candidate", "calibration.json")
	data, e := os.ReadFile(path)
	if e != nil {
		return nil, e
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != CandidateCalibrationSHA {
		return nil, fmt.Errorf("text.calibration_drift")
	}
	var c CandidateCalibration
	if e = json.Unmarshal(data, &c); e != nil {
		return nil, e
	}
	if c.Schema != "pptxgengo.wmds-vertical-candidate.v1" {
		return nil, fmt.Errorf("text.invalid_calibration_schema")
	}
	t.engine = engine
	t.anchors = map[string]VerticalAnchor{}
	for _, a := range c.Anchors {
		if a.Size <= 0 || a.Leading <= 0 || a.Baseline <= 0 || a.TerminalHeight <= 0 {
			return nil, fmt.Errorf("text.invalid_vertical_anchor")
		}
		key := anchorKey(a.FontSHA, a.Size, a.Leading)
		if _, ok := t.anchors[key]; ok {
			return nil, fmt.Errorf("text.duplicate_vertical_anchor")
		}
		t.anchors[key] = a
	}
	return t, nil
}
func ProfileForEngine(engine string) string {
	if engine == CandidateEngine {
		return CandidateProfile
	}
	return Profile
}

// measureCandidate models the captured Latin advances at 1/8 pt, with kerning
// disabled and standard ligatures enabled. It retains 64x layout units for decimal point
// tracking so a 0.54pt spacing does not become 0.546875pt in every character.
// Terminal tracking participates in line fitting and is retained in advances.
func (t *Typography) measureCandidate(text string, s Style, width float64) (TextLayout, error) {
	var result TextLayout
	if width <= 0 || math.IsNaN(width) || math.IsInf(width, 0) {
		return result, fmt.Errorf("text.invalid_width")
	}
	id, e := t.Resolve(s)
	if e != nil {
		return result, e
	}
	if s.Size <= 0 || s.Leading <= 0 || math.IsNaN(s.Size+s.Leading+s.TrackingPt) || math.IsInf(s.Size+s.Leading+s.TrackingPt, 0) || s.Size > 4096 || s.Leading > 8191 || math.Abs(s.TrackingPt) > 4096 || width > 8191 {
		return result, fmt.Errorf("text.invalid_candidate_style_or_width")
	}
	s.TrackingPt = math.Round(s.TrackingPt*100) / 100
	face := t.faces[fontKey(s.Family, s.Weight, s.Italic)].Face
	display := strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
	if s.Case == "upper" {
		display = strings.ToUpper(display)
	}
	result = TextLayout{Original: text, Displayed: display, Style: s, Font: id, Features: map[string]int{"kern": 0, "liga": 1, "clig": 1}}
	anchor, known := t.anchors[anchorKey(id.SHA256, s.Size, s.Leading)]
	baseline, terminal := .75*s.Leading, math.Max(s.Leading, 1.5*s.Size)
	result.VerticalPolicy = "uncalibrated: 0.75 leading baseline and conservative max(leading, 1.5 em) terminal allocation; requires native review"
	if known {
		baseline, terminal = anchor.Baseline, anchor.TerminalHeight
		result.VerticalPolicy = "v1 native-control anchor adopted as v2 candidate estimate; exact source font/size/leading key; native v2 validation pending"
		result.CalibrationSHA256 = CandidateCalibrationSHA
	}
	features := []shaping.FontFeature{{Tag: ot.MustNewTag("kern"), Value: 0}, {Tag: ot.MustNewTag("liga"), Value: 1}, {Tag: ot.MustNewTag("clig"), Value: 1}}
	for _, paragraph := range strings.Split(display, "\n") {
		runes := []rune(paragraph)
		if len(runes) == 0 {
			result.Lines = append(result.Lines, TextLine{"", 0, baseline + float64(len(result.Lines))*s.Leading})
			continue
		}
		for _, r := range runes {
			if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) || !unicode.In(r, unicode.Latin, unicode.Common, unicode.Inherited) {
				return result, fmt.Errorf("text.unsupported_character: U+%04X", r)
			}
			if g, ok := face.NominalGlyph(r); !ok || g == 0 {
				return result, fmt.Errorf("font.missing_glyph: %s U+%04X", id.PostScript, r)
			}
		}
		input := shaping.Input{Text: runes, RunEnd: len(runes), Direction: di.DirectionLTR, Face: face, Size: fixedPt(s.Size * 64), Language: language.NewLanguage("en"), FontFeatures: features}
		var shaped []shaping.Output
		for _, part := range t.segmenter.Split(input, oneFace{face}) {
			if part.Direction != di.DirectionLTR {
				return result, fmt.Errorf("text.unsupported_bidi")
			}
			part.FontFeatures = features
			o := t.shaper.Shape(part)
			quantize := func(v fixed.Int26_6) fixed.Int26_6 { return fixedPt(math.Round(float64(v)/4096*8) / 8 * 64) }
			for i := range o.Glyphs {
				g := &o.Glyphs[i]
				if g.GlyphID == 0 {
					return result, fmt.Errorf("font.missing_shaped_glyph")
				}
				g.Advance = quantize(g.Advance)
				g.XAdvance = quantize(g.XAdvance)
				if i == len(o.Glyphs)-1 || g.ClusterIndex != o.Glyphs[i+1].ClusterIndex {
					g.Advance += fixedPt(s.TrackingPt * 64)
					g.XAdvance += fixedPt(s.TrackingPt * 64)
				}
			}
			o.RecomputeAdvance()
			shaped = append(shaped, o)
		}
		lines, _ := t.wrapper.WrapParagraphF(shaping.WrapConfig{Direction: di.DirectionLTR, BreakPolicy: shaping.WhenNecessary}, fixedPt(width*64), runes, shaping.NewSliceIterator(shaped))
		if len(lines) == 0 {
			return result, fmt.Errorf("text.no_lines")
		}
		for _, line := range lines {
			start, end := len(runes), 0
			advance := 0.0
			for _, o := range line {
				start = min(start, o.Runes.Offset)
				end = max(end, o.Runes.Offset+o.Runes.Count)
				advance += float64(o.Advance) / 4096
			}
			if advance > width+.02 {
				return result, fmt.Errorf("text.horizontal_overflow: %.3f exceeds %.3f", advance, width)
			}
			result.Lines = append(result.Lines, TextLine{string(runes[start:end]), advance, baseline + float64(len(result.Lines))*s.Leading})
		}
	}
	result.AllocationHeight = float64(len(result.Lines)) * s.Leading
	first, last := -1, -1
	for i, l := range result.Lines {
		if strings.TrimSpace(l.Text) != "" {
			if first < 0 {
				first = i
			}
			last = i
		}
	}
	if first >= 0 {
		result.OccupiedTop = float64(first) * s.Leading
		lastHeight := terminal
		if last < len(result.Lines)-1 {
			lastHeight = s.Leading
		}
		result.EstimatedOccupiedHeight = float64(last-first)*s.Leading + lastHeight
	}
	return result, nil
}
