package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/buairtri/pptxgengo/internal/compose"
)

type manifest struct {
	Environment       *measurementEnvironment `json:"environment,omitempty"`
	CacheUses         []cacheUse              `json:"cache_uses,omitempty"`
	DeckFile          string                  `json:"deck_file"`
	Schema            string                  `json:"schema"`
	SpecSHA           string                  `json:"spec_sha256"`
	DeckSHA           string                  `json:"deck_sha256"`
	Slides            []renderSlide           `json:"slides"`
	Requests          []compose.ProbeRequest  `json:"requests,omitempty"`
	ReusedFromSpecSHA string                  `json:"reused_from_spec_sha256,omitempty"`
	Plan              *compose.PlanResult     `json:"plan,omitempty"`
}
type nativeRect struct {
	Left, Top, Width, Height float64
	Rotation                 float64 `json:"rotation_degrees"`
}
type nativeRow struct {
	Line struct {
		DashStyle    string  `json:"dash_style,omitempty"`
		Visible      bool    `json:"visible"`
		RGB          []int   `json:"rgb"`
		WidthPt      float64 `json:"width_pt"`
		Transparency float64 `json:"transparency"`
		BeginArrow   string  `json:"begin_arrow"`
		EndArrow     string  `json:"end_arrow"`
	} `json:"line"`
	Fill struct {
		Visible      bool    `json:"visible"`
		RGB          []int   `json:"rgb"`
		Transparency float64 `json:"transparency"`
	} `json:"fill"`
	TextColor []int                                       `json:"text_color"`
	Margins   *struct{ Left, Right, Top, Bottom float64 } `json:"margins"`
	Slide     int                                         `json:"slide_index"`
	Name      string                                      `json:"shape_name"`
	Text      *string                                     `json:"text"`
	Frame     nativeRect                                  `json:"shape_frame"`
	Bounds    *nativeRect                                 `json:"text_bounds"`
	Font      struct {
		Name *string  `json:"name"`
		Size *float64 `json:"size_pt"`
		Bold *bool    `json:"bold"`
	} `json:"font"`
	Characters []nativeCharacter `json:"characters,omitempty"`
	Paragraphs []nativeParagraph `json:"paragraphs,omitempty"`
}
type nativeCharacter struct {
	Bounds     *nativeRect `json:"bounds,omitempty"`
	Text       string      `json:"text"`
	FontName   string      `json:"font_name"`
	FontSizePt float64     `json:"font_size_pt"`
	Bold       bool        `json:"bold"`
	Italic     bool        `json:"italic"`
	Underline  string      `json:"underline"`
	Color      []int       `json:"color"`
}
type nativeParagraph struct {
	Alignment          string   `json:"alignment"`
	SpaceBeforePt      float64  `json:"space_before_pt"`
	SpaceAfterPt       float64  `json:"space_after_pt"`
	LineRuleWithin     bool     `json:"line_rule_within"`
	SpaceWithin        float64  `json:"space_within"`
	BulletVisible      *bool    `json:"bullet_visible"`
	BulletType         string   `json:"bullet_type"`
	BulletCharacter    string   `json:"bullet_character"`
	BulletRelativeSize *float64 `json:"bullet_relative_size"`
	BulletUseTextColor *bool    `json:"bullet_use_text_color"`
	BulletUseTextFont  *bool    `json:"bullet_use_text_font"`
}
type nativeResult struct {
	Schema string      `json:"schema"`
	Count  int         `json:"visible_slide_count"`
	Rows   []nativeRow `json:"measurements"`
}
type evidence struct {
	Environment  *measurementEnvironment `json:"environment,omitempty"`
	AdapterSHA   string                  `json:"adapter_sha256"`
	Schema       string                  `json:"schema"`
	SpecSHA      string                  `json:"spec_sha256"`
	DeckSHA      string                  `json:"deck_sha256"`
	ManifestSHA  string                  `json:"manifest_sha256"`
	MeasuredAt   string                  `json:"measured_at"`
	Measurements compose.Measurements    `json:"measurements"`
	Native       json.RawMessage         `json:"native"`
}

func hash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func readJSON(path string, v any) ([]byte, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return nil, e
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if e = d.Decode(v); e != nil {
		return nil, fmt.Errorf("%s: %w", path, e)
	}
	if e = d.Decode(new(any)); e != io.EOF {
		return nil, fmt.Errorf("%s: trailing JSON", path)
	}
	return b, nil
}
func jsonBytes(v any) []byte {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		panic(e)
	}
	return append(b, '\n')
}
func writeNew(path string, b []byte) error {
	f, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if e != nil {
		return e
	}
	_, e = f.Write(b)
	ce := f.Close()
	if e != nil {
		os.Remove(path)
		return e
	}
	return ce
}
func bundle(dir string, m manifest) (err error) {
	if _, e := os.Lstat(dir); !os.IsNotExist(e) {
		return fmt.Errorf("output must be a new directory: %s", dir)
	}
	parent := filepath.Dir(dir)
	if err = os.MkdirAll(parent, 0755); err != nil {
		return
	}
	tmp, e := os.MkdirTemp(parent, ".compose-")
	if e != nil {
		return e
	}
	defer os.RemoveAll(tmp)
	b, e := render(m.Slides)
	if e != nil {
		return e
	}
	if e = validateShapeStructure(b, m.Slides); e != nil {
		return e
	}
	if e = validateImageStructure(b, m.Slides); e != nil {
		return e
	}
	if e = validateRichBulletStructure(b, m.Slides); e != nil {
		return e
	}
	m.DeckSHA = hash(b)
	m.DeckFile = filepath.Base(dir) + ".pptx"
	if e = writeNew(filepath.Join(tmp, m.DeckFile), b); e != nil {
		return e
	}
	if e = writeNew(filepath.Join(tmp, "manifest.json"), jsonBytes(m)); e != nil {
		return e
	}
	return os.Rename(tmp, dir)
}
func color(s string) string     { return strings.TrimPrefix(s, "#") }
func rect(r compose.Rect) frame { return frame{r.X, r.Y, r.Width, r.Height} }
func finalSlides(spec compose.Spec, p compose.PlanResult) []renderSlide {
	var out []renderSlide
	for i, s := range p.Slides {
		rs := renderSlide{ID: s.ID, Width: spec.Slides[i].WidthPt, Height: spec.Slides[i].HeightPt, Notes: spec.Slides[i].Notes + manualPlacementNotes(s.ManualRequired)}
		for _, e := range canvasElements(s.Canvas) {
			if e.Kind == "surface" || e.Kind == "shape" {
				rs.Elements = append(rs.Elements, e)
			}
		}
		rs.Elements = append(rs.Elements, accentElements(s.Accents)...)
		rs.Elements = append(rs.Elements, artworkArrowElements(s.ArtworkArrows)...)
		for _, e := range canvasElements(s.Canvas) {
			if e.Kind != "surface" && e.Kind != "shape" {
				rs.Elements = append(rs.Elements, e)
			}
		}
		rs.Elements = append(rs.Elements, cardElements(s.Cards)...)

		for _, ph := range s.Phases {
			name := "phase:" + base64.RawURLEncoding.EncodeToString([]byte(ph.ID))
			rs.Elements = append(rs.Elements, element{Name: name + "-surface", Kind: "surface", Frame: rect(ph.Bounds), Background: color(ph.Surface)})
			rs.Elements = append(rs.Elements, element{Name: name + "-label", Kind: "text", Frame: rect(ph.TitleRect), Text: ph.Title, FontFace: ph.TitleFontFace, FontSize: ph.TitleFontSizePt, Bold: ph.TitleBold, Foreground: color(ph.Foreground), MeasurementID: ph.TitleMeasurementID})
		}
		rs.Elements = append(rs.Elements, connectorElements(s.Connections)...)
		if s.TitleMeasurementID != "" {
			rs.Elements = append(rs.Elements, element{Name: "slide-title", Kind: "text", Frame: rect(s.TitleBounds), Text: s.Title, FontFace: s.TitleFontFace, FontSize: s.TitleFontSizePt, Bold: s.TitleBold, Foreground: color(s.TitleForeground), Align: "left", MeasurementID: s.TitleMeasurementID})
		}
		for j, pod := range s.Pods {
			st := spec.Slides[i].Pods[j].Style
			name := "pod:" + base64.RawURLEncoding.EncodeToString([]byte(pod.ID))
			rs.Elements = append(rs.Elements, element{Name: name + "-surface", Kind: "surface", Frame: rect(pod.Bounds), Background: color(pod.Surface)})
			rs.Elements = append(rs.Elements, element{Name: name + "-title", Kind: "text", Frame: rect(pod.TitleRect), Text: pod.Title, FontFace: pod.TitleFontFace, FontSize: pod.TitleFontSizePt, Bold: pod.TitleBold, Foreground: color(pod.TitleForeground), MeasurementID: pod.TitleMeasurementID})
			for _, r := range pod.Roles {
				rs.Elements = append(rs.Elements, element{Name: "role:" + base64.RawURLEncoding.EncodeToString([]byte(r.ID)), Kind: "text", Frame: rect(r.Bounds), Text: r.Label, FontFace: r.FontFace, FontSize: r.FontSizePt, Bold: r.Bold, Foreground: color(r.Foreground), Background: color(r.Background), InsetX: st.HorizontalInsetPt, InsetY: st.VerticalInsetPt + st.ParagraphGapPt, Valign: "middle", MeasurementID: r.MeasurementID})
			}
		}
		for _, r := range s.Roles {
			rs.Elements = append(rs.Elements, roleElement(r))
		}
		if s.Legend != nil {
			for _, item := range s.Legend.Entries {
				name := "legend:" + base64.RawURLEncoding.EncodeToString([]byte(item.Token))
				rs.Elements = append(rs.Elements, element{Name: name + "-swatch", Kind: "surface", Frame: rect(item.SwatchBounds), Background: color(item.Color)})
				rs.Elements = append(rs.Elements, element{Name: name + "-label", Kind: "text", Frame: rect(item.LabelBounds), Text: item.Label, FontFace: item.FontFace, FontSize: item.FontSizePt, Bold: item.Bold, Foreground: color(item.Foreground), Align: "left", MeasurementID: item.MeasurementID})
			}
		}
		// Explicit canvas layers are ordered globally; within a layer surfaces sit
		// behind accents and text. Team phase backgrounds precede canvas annotations.
		type drawKey struct{ layer, rank int }
		keys := map[string]drawKey{}
		targetLayers := map[string]int{}
		for _, c := range s.Canvas {
			rank := 2
			if c.Kind == "surface" || c.Kind == "shape" {
				rank = 0
			}
			keys["canvas:"+base64.RawURLEncoding.EncodeToString([]byte(c.ID))] = drawKey{c.Layer, rank}
			targetLayers[c.ID] = c.Layer
		}
		for _, a := range s.Accents {
			keys["accent:"+base64.RawURLEncoding.EncodeToString([]byte(a.ID))] = drawKey{targetLayers[a.Target], 1}
		}
		phaseSurfaces := map[string]bool{}
		for _, ph := range s.Phases {
			phaseSurfaces["phase:"+base64.RawURLEncoding.EncodeToString([]byte(ph.ID))+"-surface"] = true
		}
		key := func(name string) drawKey {
			if k, ok := keys[name]; ok {
				return k
			}
			return drawKey{0, 3}
		}
		sort.SliceStable(rs.Elements, func(i, j int) bool {
			a, b := rs.Elements[i].Name, rs.Elements[j].Name
			if phaseSurfaces[a] != phaseSurfaces[b] {
				return phaseSurfaces[a]
			}
			ka, kb := key(a), key(b)
			if ka.layer != kb.layer {
				return ka.layer < kb.layer
			}
			return ka.rank < kb.rank
		})
		out = append(out, rs)
	}
	return out
}
func finite(v float64) bool  { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func near(a, b float64) bool { return finite(a) && finite(b) && math.Abs(a-b) <= 0.12 }

// Paragraph line spacing is a unitless multiplier reported by PowerPoint with
// only floating-point representation noise. Keep this tighter than geometric
// comparisons so materially different settings such as 0.9 and 1.0 cannot
// pass native verification.
func nearLineSpacing(a, b float64) bool {
	return finite(a) && finite(b) && math.Abs(a-b) <= 0.0001
}
func checkNative(m manifest, n nativeResult, fit bool) (compose.Measurements, error) {
	result := compose.Measurements{ByRequestID: map[string]compose.Measurement{}}
	if (n.Schema != "pptxgengo.compose-text-measurement.v2" && n.Schema != "pptxgengo.compose-text-measurement.v3" && n.Schema != "pptxgengo.compose-text-measurement.v4" && n.Schema != "pptxgengo.compose-text-measurement.v5" && n.Schema != "pptxgengo.compose-text-measurement.v6" && n.Schema != "pptxgengo.compose-text-measurement.v7" && n.Schema != "pptxgengo.compose-text-measurement.v8") || n.Count != len(m.Slides) {
		return result, fmt.Errorf("native schema or slide count mismatch")
	}
	rows := map[string]nativeRow{}
	for _, r := range n.Rows {
		k := fmt.Sprintf("%d:%s", r.Slide, r.Name)
		if _, ok := rows[k]; ok {
			return result, fmt.Errorf("duplicate native shape %s", k)
		}
		rows[k] = r
	}
	for i, s := range m.Slides {
		for _, e := range s.Elements {
			k := fmt.Sprintf("%d:%s", i+1, e.Name)
			r, ok := rows[k]
			if !ok {
				return result, fmt.Errorf("missing native shape %s", k)
			}
			delete(rows, k)
			if !near(r.Frame.Left, e.Frame.X) || !near(r.Frame.Top, e.Frame.Y) || !near(r.Frame.Width, e.Frame.Width) || !near(r.Frame.Height, e.Frame.Height) || !near(math.Remainder(r.Frame.Rotation-e.RotationDeg, 360), 0) {
				return result, fmt.Errorf("native frame mismatch %s: %+v expected %+v", k, r.Frame, e.Frame)
			}

			if e.Kind == "image" {
				if r.Text != nil && *r.Text != "" {
					return result, fmt.Errorf("unexpected image text %s", k)
				}
				if e.OutlineColor == "" {
					if r.Line.Visible {
						return result, fmt.Errorf("unexpected image outline %s", k)
					}
				} else if !r.Line.Visible || rgbHex(r.Line.RGB) != e.OutlineColor || !near(r.Line.WidthPt, e.OutlineWidthPt) || !finite(r.Line.Transparency) || math.Abs(r.Line.Transparency) > 1e-6 {
					return result, fmt.Errorf("native image outline mismatch %s: %+v", k, r.Line)
				}
				continue
			}
			if e.Kind == "shape" {
				if (n.Schema == "pptxgengo.compose-text-measurement.v3" || n.Schema == "pptxgengo.compose-text-measurement.v4" || n.Schema == "pptxgengo.compose-text-measurement.v5" || n.Schema == "pptxgengo.compose-text-measurement.v6" || n.Schema == "pptxgengo.compose-text-measurement.v7" || n.Schema == "pptxgengo.compose-text-measurement.v8") && r.Line.Visible {
					return result, fmt.Errorf("unexpected shape outline %s", k)
				}
				if e.Pattern == nil && (!r.Fill.Visible || !finite(r.Fill.Transparency) || math.Abs(r.Fill.Transparency) > 1e-6 || rgbHex(r.Fill.RGB) != e.Background) {
					return result, fmt.Errorf("native shape fill mismatch %s: %+v expected %s", k, r.Fill, e.Background)
				}
				if r.Text != nil && *r.Text != "" {
					return result, fmt.Errorf("unexpected shape text %s", k)
				}
				continue
			}
			if e.Kind == "line" {
				if (n.Schema != "pptxgengo.compose-text-measurement.v3" && n.Schema != "pptxgengo.compose-text-measurement.v4" && n.Schema != "pptxgengo.compose-text-measurement.v5" && n.Schema != "pptxgengo.compose-text-measurement.v6" && n.Schema != "pptxgengo.compose-text-measurement.v7" && n.Schema != "pptxgengo.compose-text-measurement.v8") || !r.Line.Visible || rgbHex(r.Line.RGB) != e.Foreground || !near(r.Line.WidthPt, e.LineWidth) || math.Abs(r.Line.Transparency) > 1e-6 || !finite(r.Line.Transparency) || r.Line.BeginArrow != nativeArrow(e.BeginArrow) || r.Line.EndArrow != nativeArrow(e.EndArrow) || (e.LineDash != "" && r.Line.DashStyle != nativeDash(e.LineDash)) {
					return result, fmt.Errorf("native line mismatch %s: %+v", k, r.Line)
				}
				if r.Text != nil && *r.Text != "" {
					return result, fmt.Errorf("unexpected connector text %s", k)
				}
				continue
			}
			if (n.Schema == "pptxgengo.compose-text-measurement.v3" || n.Schema == "pptxgengo.compose-text-measurement.v4" || n.Schema == "pptxgengo.compose-text-measurement.v5" || n.Schema == "pptxgengo.compose-text-measurement.v6" || n.Schema == "pptxgengo.compose-text-measurement.v7" || n.Schema == "pptxgengo.compose-text-measurement.v8") && r.Line.Visible {
				return result, fmt.Errorf("unexpected outline %s", k)
			}

			if e.Background != "" {
				if !r.Fill.Visible || (!finite(r.Fill.Transparency) || math.Abs(r.Fill.Transparency) > 1e-6) || rgbHex(r.Fill.RGB) != e.Background {
					return result, fmt.Errorf("native fill mismatch %s: %+v expected %s", k, r.Fill, e.Background)
				}
			} else if r.Fill.Visible {
				return result, fmt.Errorf("unexpected native fill %s", k)
			}
			if e.Kind == "surface" {
				if r.Text != nil && *r.Text != "" {
					return result, fmt.Errorf("unexpected surface text %s", k)
				}
				continue
			}
			norm := func(v string) string { return strings.ReplaceAll(strings.ReplaceAll(v, "\r\n", "\n"), "\r", "\n") }
			if r.Text == nil || norm(*r.Text) != norm(e.Text) {
				return result, fmt.Errorf("native text mismatch %s", k)
			}
			if len(e.Paragraphs) > 0 {
				if n.Schema != "pptxgengo.compose-text-measurement.v8" {
					return result, fmt.Errorf("rich text requires v8 native paragraph/style/bullet evidence: %s", k)
				}
				if err := checkNativeRichText(k, e.Paragraphs, r); err != nil {
					return result, err
				}
			} else {
				if r.Font.Name == nil || *r.Font.Name != e.FontFace || r.Font.Size == nil || !near(*r.Font.Size, e.FontSize) || r.Font.Bold == nil || *r.Font.Bold != e.Bold {
					return result, fmt.Errorf("native font mismatch %s: %+v", k, r.Font)
				}
				if rgbHex(r.TextColor) != e.Foreground {
					return result, fmt.Errorf("native text color mismatch %s: %v expected %s", k, r.TextColor, e.Foreground)
				}
			}
			if r.Margins == nil || !near(r.Margins.Left, e.InsetX) || !near(r.Margins.Right, e.InsetX) || !near(r.Margins.Top, e.InsetY) || !near(r.Margins.Bottom, e.InsetY) {
				return result, fmt.Errorf("native margins mismatch %s: %+v", k, r.Margins)
			}
			b := r.Bounds
			if b == nil || !finite(b.Left) || !finite(b.Top) || !finite(b.Width) || !finite(b.Height) || b.Width <= 0 || b.Height <= 0 {
				return result, fmt.Errorf("invalid native text bounds %s", k)
			}
			// Probe overflow is recorded for candidate rejection. Final text must fit its inner safe zone.
			if fit && (b.Left < e.Frame.X+e.InsetX-0.15 || b.Top < e.Frame.Y+e.InsetY-0.15 || b.Left+b.Width > e.Frame.X+e.Frame.Width-e.InsetX+0.15 || b.Top+b.Height > e.Frame.Y+e.Frame.Height-e.InsetY+0.15) {
				return result, fmt.Errorf("native text outside safe zone %s: %+v frame %+v insets %.2f,%.2f", k, *b, e.Frame, e.InsetX, e.InsetY)
			}
			phrases, err := nativePhraseBounds(e, r)
			if err != nil {
				return result, fmt.Errorf("native phrase %s: %w", k, err)
			}
			result.ByRequestID[e.MeasurementID] = compose.Measurement{PhraseBounds: phrases, RenderedWidthPt: b.Width, RenderedHeightPt: b.Height, OffsetXPt: b.Left - e.Frame.X - e.InsetX, OffsetYPt: b.Top - e.Frame.Y - e.InsetY}
		}
	}
	if fit && m.Plan != nil {
		frames := map[string]element{}
		for _, slide := range m.Slides {
			for _, e := range slide.Elements {
				if e.MeasurementID != "" {
					frames[e.MeasurementID] = e
				}
			}
		}
		for _, slide := range m.Plan.Slides {
			if err := compose.VerifyArtworkAnchors(&slide, result, .15); err != nil {
				return result, err
			}
			for _, a := range slide.Accents {
				if a.Status != "placed" || a.Phrase == "" {
					continue
				}
				e := frames[a.TargetMeasurementID]
				fragments := result.ByRequestID[a.TargetMeasurementID].PhraseBounds[a.SourceAccentID]
				if a.FragmentIndex < 0 || a.FragmentIndex >= len(fragments) {
					return result, fmt.Errorf("native accent %s fragment missing", a.ID)
				}
				r := fragments[a.FragmentIndex]
				r.X += e.Frame.X + e.InsetX
				r.Y += e.Frame.Y + e.InsetY
				if !near(r.X, a.TargetBounds.X) || !near(r.Y, a.TargetBounds.Y) || !near(r.Width, a.TargetBounds.Width) || !near(r.Height, a.TargetBounds.Height) {
					return result, fmt.Errorf("native accent %s target shifted after measurement: %+v expected %+v", a.ID, r, a.TargetBounds)
				}
			}
		}
	}
	if len(rows) != 0 {
		return result, fmt.Errorf("unexpected native shapes: %d", len(rows))
	}
	return result, nil
}
func checkNativeRichText(shape string, expected []compose.ParagraphSpec, row nativeRow) error {
	if len(row.Paragraphs) != len(expected) {
		return fmt.Errorf("native rich paragraph count mismatch %s: %d expected %d", shape, len(row.Paragraphs), len(expected))
	}
	align := map[string]string{"left": "paragraph align left", "center": "paragraph align center", "right": "paragraph align right"}
	var want []nativeCharacter
	for i, p := range expected {
		np := row.Paragraphs[i]
		lineSpacing := p.LineSpacingMultiple
		if lineSpacing == 0 {
			lineSpacing = 1
		}
		if np.Alignment != align[p.Align] || !near(np.SpaceBeforePt, p.SpaceBeforePt) || !near(np.SpaceAfterPt, p.SpaceAfterPt) || !np.LineRuleWithin || !nearLineSpacing(np.SpaceWithin, lineSpacing) {
			return fmt.Errorf("native paragraph style mismatch %s paragraph %s: %+v", shape, p.ID, np)
		}
		if np.BulletVisible == nil {
			return fmt.Errorf("native paragraph bullet evidence missing %s paragraph %s", shape, p.ID)
		}
		if p.Bullet == nil {
			if *np.BulletVisible || np.BulletType != "ppBulletNone" {
				return fmt.Errorf("native unexpected bullet %s paragraph %s: %+v", shape, p.ID, np)
			}
		} else if !*np.BulletVisible || np.BulletType != "ppBulletUnnumbered" || np.BulletCharacter != p.Bullet.Character || np.BulletRelativeSize == nil || !nearLineSpacing(*np.BulletRelativeSize, 1) || np.BulletUseTextColor == nil || !*np.BulletUseTextColor || np.BulletUseTextFont == nil || !*np.BulletUseTextFont {
			return fmt.Errorf("native bullet style mismatch %s paragraph %s: %+v", shape, p.ID, np)
		}
		for _, run := range p.Runs {
			for _, ch := range run.Text {
				if ch == '\r' || ch == '\n' {
					continue
				}
				u := "no underline"
				if run.Underline {
					u = "underline single line"
				}
				want = append(want, nativeCharacter{Text: string(ch), FontName: run.FontFace, FontSizePt: run.FontSizePt, Bold: run.Bold, Italic: run.Italic, Underline: u, Color: rgbInts(run.Foreground)})
			}
		}
	}
	if len(row.Characters) != len(want) {
		return fmt.Errorf("native rich character count mismatch %s: %d expected %d", shape, len(row.Characters), len(want))
	}
	for i, w := range want {
		g := row.Characters[i]
		if g.Text != w.Text || g.FontName != w.FontName || !near(g.FontSizePt, w.FontSizePt) || g.Bold != w.Bold || g.Italic != w.Italic || g.Underline != w.Underline || rgbHex(g.Color) != rgbHex(w.Color) {
			return fmt.Errorf("native rich character style mismatch %s character %d: %+v expected %+v", shape, i+1, g, w)
		}
	}
	return nil
}
func rgbInts(hex string) []int {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return nil
	}
	out := make([]int, 3)
	for i := 0; i < 3; i++ {
		v, e := strconv.ParseInt(hex[i*2:i*2+2], 16, 16)
		if e != nil {
			return nil
		}
		out[i] = int(v)
	}
	return out
}
func rgbHex(v []int) string {
	if len(v) != 3 {
		return ""
	}
	for _, c := range v {
		if c < 0 || c > 255 {
			return ""
		}
	}
	return fmt.Sprintf("%02X%02X%02X", v[0], v[1], v[2])
}
func deckPath(dir string, m manifest) (string, error) {
	if m.DeckFile == "" || m.DeckFile != filepath.Base(m.DeckFile) || filepath.Ext(m.DeckFile) != ".pptx" {
		return "", fmt.Errorf("invalid bundle deck filename")
	}
	p := filepath.Join(dir, m.DeckFile)
	st, e := os.Lstat(p)
	if e != nil {
		return "", e
	}
	if !st.Mode().IsRegular() {
		return "", fmt.Errorf("bundle deck must be a regular file")
	}
	return p, nil
}
func nativeMeasure(deck, script string) ([]byte, error) {
	abs, e := filepath.Abs(deck)
	if e != nil {
		return nil, e
	}
	// Never silently measure an unsaved, already-open copy of the same filename.
	openScript := `on run argv
set sourceFile to POSIX file (item 1 of argv)
set sourceName to name of (info for sourceFile)
tell application "Microsoft PowerPoint"
set openNames to name of every presentation
if openNames contains sourceName then error "Close the existing presentation named " & sourceName & " before measuring this file."
open sourceFile
end tell
end run`
	c := exec.Command("osascript", "-", abs)
	c.Stdin = strings.NewReader(openScript)
	b, e := c.CombinedOutput()
	if e != nil {
		return nil, fmt.Errorf("PowerPoint open: %w: %s", e, b)
	}
	c = exec.Command("osascript", script, filepath.Base(abs))
	b, e = c.CombinedOutput()
	if e != nil {
		return nil, fmt.Errorf("PowerPoint measure: %w: %s", e, b)
	}
	return b, nil
}

func probeSlides(requests []compose.ProbeRequest) []renderSlide {
	var slides []renderSlide
	for i, q := range requests {
		// One isolated text box per probe avoids layout interactions and keeps the
		// calibration frame in slide points. The tall frame does not autofit text.
		width := math.Max(960, q.TextWidthPt+72)
		slides = append(slides, renderSlide{ID: q.ID, Width: width, Height: 540, Elements: []element{{Name: fmt.Sprintf("probe-%04d", i+1), Kind: "text", Frame: frame{36, 36, q.TextWidthPt, 450}, Text: q.Text, Paragraphs: q.Paragraphs, PhraseRequests: q.PhraseRequests, FontFace: q.FontFace, FontSize: q.FontSizePt, Bold: q.Bold, Foreground: "070154", Align: q.Align, MeasurementID: q.ID}}})
	}
	// The writer requires one common page size.
	maxW := 960.0
	for _, s := range slides {
		maxW = math.Max(maxW, s.Width)
	}
	for i := range slides {
		slides[i].Width = maxW
	}
	return slides
}

func main() {
	if e := run(os.Args[1:]); e != nil {
		fmt.Fprintln(os.Stderr, "pptxcompose:", e)
		os.Exit(1)
	}
}
func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: pptxcompose probe|measure|fit-report|build|verify|recover-text|cache-import [flags]")
	}
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	cacheDir := fs.String("cache", "", "native measurement cache directory; missing or stale contracts are never guessed")
	returned := fs.String("returned", "", "colleague-edited generated PPTX for recover-text")
	textOnly := fs.Bool("text-only", false, "recover text while restoring original spec styling; geometry changes require scene extraction")
	specPath := fs.String("spec", "", "composition spec JSON")
	out := fs.String("out", "", "new output directory (probe/build) or JSON file (measure/verify)")
	dir := fs.String("bundle", "", "probe or output bundle directory")
	reuse := fs.Bool("reuse-measurements", false, "explicitly reuse native measurements when every text probe request is unchanged")
	evPath := fs.String("evidence", "", "native probe evidence JSON")
	script := fs.String("adapter", releaseScript("measure-compose-text.applescript"), "native PowerPoint measurement adapter")
	nativeWorkspace := fs.String("native-workspace", "", "existing directory for a staged PowerPoint measurement copy (measure/verify only)")
	if e := fs.Parse(args[1:]); e != nil {
		return e
	}
	if *nativeWorkspace != "" && args[0] != "measure" && args[0] != "verify" {
		return fmt.Errorf("--native-workspace is supported for measure/verify only")
	}
	if fs.NArg() != 0 || *out == "" {
		return fmt.Errorf("--out is required; positional arguments unsupported")
	}
	switch args[0] {
	case "cache-import":
		if *cacheDir == "" {
			return fmt.Errorf("--cache required")
		}
		env, e := currentEnvironment(*script)
		if e != nil {
			return e
		}
		count, e := cacheImport(*cacheDir, *dir, *evPath, env)
		if e != nil {
			return e
		}
		return writeNew(*out, jsonBytes(map[string]any{"imported": count, "environment": env}))
	case "recover-text":
		if *returned == "" {
			return fmt.Errorf("--returned required")
		}
		if e := recoverText(*specPath, *dir, *returned, *out, *textOnly); e != nil {
			return e
		}
		fmt.Println(*out)
		return nil
	case "probe", "build", "fit-report":
		var spec compose.Spec
		raw, e := readJSON(*specPath, &spec)
		if e != nil {
			return e
		}
		requests, e := compose.ProbeRequests(spec)
		if e != nil {
			return e
		}
		m := manifest{Schema: "pptxgengo.compose-bundle.v1", SpecSHA: hash(raw)}
		var cached compose.Measurements
		var missing []compose.ProbeRequest
		if *cacheDir != "" {
			if *dir != "" || *evPath != "" || *reuse {
				return fmt.Errorf("--cache is exclusive with --bundle, --evidence and --reuse-measurements for probe/build/fit-report")
			}
			env, err := currentEnvironment(*script)
			if err != nil {
				return err
			}
			m.Environment = env
			cached, missing, m.CacheUses, err = cacheLookup(*cacheDir, env, requests)
			if err != nil {
				return err
			}
		}
		if args[0] == "probe" {
			if *cacheDir != "" {
				if len(missing) == 0 {
					if _, err := os.Lstat(*out); !os.IsNotExist(err) {
						return fmt.Errorf("output must be a new directory")
					}
					if err := os.MkdirAll(*out, 0755); err != nil {
						return err
					}
					err := writeNew(filepath.Join(*out, "cache-report.json"), jsonBytes(map[string]any{"status": "all_requests_cached", "requests": len(requests), "environment": m.Environment, "cache_uses": m.CacheUses}))
					if err != nil {
						return err
					}
					fmt.Println(*out + " (all requests cached; no probe deck needed)")
					return nil
				}
				requests = missing
			}
			m.Requests = requests
			m.Slides = probeSlides(requests)
		} else {
			var measured compose.Measurements
			var ev evidence
			if *cacheDir != "" {
				if len(missing) > 0 {
					return fmt.Errorf("%d uncached text contracts; run probe --cache and measure --cache first", len(missing))
				}
				measured = cached
			} else {
				if _, e = readJSON(*evPath, &ev); e != nil {
					return e
				}
				var pm manifest
				mb, e := readJSON(filepath.Join(*dir, "manifest.json"), &pm)
				if e != nil {
					return e
				}
				dp, e := deckPath(*dir, pm)
				if e != nil {
					return e
				}
				db, e := os.ReadFile(dp)
				if e != nil {
					return e
				}
				if pm.Schema != "pptxgengo.compose-bundle.v1" || pm.Plan != nil || ev.Schema != "pptxgengo.compose-evidence.v1" || ev.SpecSHA != pm.SpecSHA || (!*reuse && pm.SpecSHA != hash(raw)) || ev.ManifestSHA != hash(mb) || ev.DeckSHA != hash(db) || pm.DeckSHA != hash(db) || len(pm.Requests) == 0 {
					return fmt.Errorf("stale or mismatched probe evidence")
				}
				if !bytes.Equal(jsonBytes(pm.Requests), jsonBytes(requests)) {
					return fmt.Errorf("probe requests do not match current spec")
				}
				var nr nativeResult
				if e = json.Unmarshal(ev.Native, &nr); e != nil {
					return e
				}
				for _, s := range spec.Slides {
					if len(s.Accents) > 0 && nr.Schema != "pptxgengo.compose-text-measurement.v4" && nr.Schema != "pptxgengo.compose-text-measurement.v5" && nr.Schema != "pptxgengo.compose-text-measurement.v6" && nr.Schema != "pptxgengo.compose-text-measurement.v7" && nr.Schema != "pptxgengo.compose-text-measurement.v8" {
						return fmt.Errorf("measured accents require v4/v5/v6/v7/v8 native character-bound evidence")
					}
				}
				measured, e = checkNative(pm, nr, false)
				if e != nil {
					return e
				}
				if ev.SpecSHA != hash(raw) {
					m.ReusedFromSpecSHA = ev.SpecSHA
				}
			}
			// Reconstruct dimensions from raw native evidence instead of trusting an editable summary.
			plan, e := compose.Plan(spec, measured)
			if args[0] == "fit-report" {
				rows := compose.FixedTextFitReport(spec, measured)
				layoutFailures := compose.LayoutFitReport(spec, measured)
				failures := 0
				for _, row := range rows {
					if !row.Fits {
						failures++
					}
				}
				planError := ""
				if e != nil {
					planError = e.Error()
				}
				report := map[string]any{"schema": "pptxgengo.fixed-text-fit.v1", "spec_sha256": hash(raw), "probe_spec_sha256": ev.SpecSHA, "evidence_deck_sha256": ev.DeckSHA, "fixed_zone_count": len(rows), "overflow_count": failures, "layout_failure_count": len(layoutFailures), "layout_failures": layoutFailures, "zones": rows, "planner_passed": e == nil, "planner_error": planError, "cache_uses": m.CacheUses, "environment": m.Environment, "scope": "All fixed title/canvas/role/card text zones. Dynamic pod/phase/legend layout, collisions and routes are covered by the planner result. Native final verification and visual review are still required."}
				if err := writeNew(*out, jsonBytes(report)); err != nil {
					return err
				}
				fmt.Println(*out)
				return nil
			}
			if e != nil {
				return e
			}
			m.Plan = &plan
			m.Slides = finalSlides(spec, plan)
		}
		if e = bundle(*out, m); e != nil {
			return e
		}
		fmt.Println(*out)
		return nil
	case "measure", "verify":
		if args[0] == "verify" && *cacheDir != "" {
			return fmt.Errorf("--cache is supported for probe measurements, not final verification")
		}
		if _, e := os.Lstat(*out); !os.IsNotExist(e) {
			return fmt.Errorf("output must be a new file: %s", *out)
		}
		var m manifest
		mb, e := readJSON(filepath.Join(*dir, "manifest.json"), &m)
		if e != nil {
			return e
		}
		dp, e := deckPath(*dir, m)
		if e != nil {
			return e
		}
		db, e := os.ReadFile(dp)
		if e != nil {
			return e
		}
		if m.Schema != "pptxgengo.compose-bundle.v1" || hash(db) != m.DeckSHA {
			return fmt.Errorf("bundle hash mismatch")
		}
		if e = validateShapeStructure(db, m.Slides); e != nil {
			return e
		}
		if e = validateImageStructure(db, m.Slides); e != nil {
			return e
		}
		if e = validateRichBulletStructure(db, m.Slides); e != nil {
			return e
		}
		isFinal := args[0] == "verify"
		if isFinal != (m.Plan != nil) {
			return fmt.Errorf("measure requires probe bundle; verify requires built bundle")
		}
		adapterBytes, e := os.ReadFile(*script)
		if e != nil {
			return e
		}
		env, e := currentEnvironment(*script)
		if e != nil {
			return e
		}
		if m.Environment != nil && environmentKey(m.Environment) != environmentKey(env) {
			return fmt.Errorf("native environment changed after probe/build; regenerate the bundle")
		}
		measurePath := dp
		if *nativeWorkspace != "" {
			measurePath, e = stageNativeDeck(*nativeWorkspace, dp, db)
			if e != nil {
				return e
			}
		}
		b, measureErr := nativeMeasure(measurePath, *script)
		if *nativeWorkspace != "" {
			if e = checkNativeDeckUnchanged(measurePath, m.DeckSHA); e != nil {
				return e
			}
		}
		if measureErr != nil {
			return measureErr
		}
		afterEnv, e := currentEnvironment(*script)
		if e != nil {
			return e
		}
		if environmentKey(env) != environmentKey(afterEnv) {
			return fmt.Errorf("native environment changed during measurement")
		}
		afterBytes, e := os.ReadFile(dp)
		if e != nil {
			return e
		}
		if hash(afterBytes) != m.DeckSHA {
			return fmt.Errorf("deck changed during native measurement")
		}
		var nr nativeResult
		if e = json.Unmarshal(b, &nr); e != nil {
			return fmt.Errorf("native JSON: %w: %s", e, b)
		}
		measured, e := checkNative(m, nr, isFinal)
		if e != nil {
			failure := map[string]any{"status": "failed", "reason": e.Error(), "native": json.RawMessage(b), "deck_sha256": m.DeckSHA}
			if we := writeNew(*out+".failed.json", jsonBytes(failure)); we != nil {
				return fmt.Errorf("%w; failed to retain native evidence: %v", e, we)
			}
			return e
		}
		ev := evidence{Environment: env, AdapterSHA: hash(adapterBytes), Schema: "pptxgengo.compose-evidence.v1", SpecSHA: m.SpecSHA, DeckSHA: m.DeckSHA, ManifestSHA: hash(mb), MeasuredAt: time.Now().UTC().Format(time.RFC3339), Measurements: measured, Native: b}
		if e = writeNew(*out, jsonBytes(ev)); e != nil {
			return e
		}
		if *cacheDir != "" {
			if isFinal {
				return fmt.Errorf("--cache imports probe observations only; final evidence was saved")
			}
			if _, e = cacheImport(*cacheDir, *dir, *out, env); e != nil {
				return fmt.Errorf("evidence saved but cache import failed: %w", e)
			}
		}
		fmt.Println(*out)
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func manualPlacementNotes(notes []string) string {
	if len(notes) == 0 {
		return ""
	}
	return "\nUNFINISHED — MANUAL PLACEMENT REQUIRED:\n" + strings.Join(notes, "\n")
}
