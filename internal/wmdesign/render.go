package wmdesign

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/buairtri/pptxgengo/pptx"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Node struct {
	ID         string          `json:"id"`
	Kind       string          `json:"kind"`
	Rect       Rect            `json:"rect"`
	Scope      string          `json:"scope,omitempty"`
	Text       string          `json:"text,omitempty"`
	Style      string          `json:"style,omitempty"`
	Ink        string          `json:"ink,omitempty"`
	Surface    string          `json:"surface,omitempty"`
	Align      string          `json:"align,omitempty"`
	Grid       string          `json:"grid,omitempty"`
	Start      int             `json:"start,omitempty"`
	Span       int             `json:"span,omitempty"`
	TextBlock  *TextBlockSpec  `json:"textblock,omitempty"`
	Card       *CardSpec       `json:"card,omitempty"`
	CardRow    *CardRowSpec    `json:"card_row,omitempty"`
	DataMetric *DataMetricSpec `json:"data_metric,omitempty"`
	RichText   *RichTextSpec   `json:"richtext,omitempty"`
	Scene      *SceneSpec      `json:"scene,omitempty"`
}
type SlideSpec struct {
	NativeGeometry  map[string]NativeGeometry `json:"native_geometry,omitempty"`
	NativeOrder     map[string][]string       `json:"native_order,omitempty"`
	ID              string                    `json:"id"`
	Density         string                    `json:"density,omitempty"`
	DensityLimit    string                    `json:"source_density_limit,omitempty"`
	AutoDensity     *bool                     `json:"auto_density,omitempty"`
	Hidden          bool                      `json:"hidden,omitempty"`
	Notes           string                    `json:"notes,omitempty"`
	DraftReview     *DraftReviewNote          `json:"draft_review,omitempty"`
	Frame           FrameRequest              `json:"frame"`
	Eyebrow         string                    `json:"eyebrow"`
	Title           string                    `json:"title"`
	Source          string                    `json:"source,omitempty"`
	Nodes           []Node                    `json:"nodes"`
	ContentKind     string                    `json:"content_kind,omitempty"`
	TemplateBinding *TemplateSlideRecord      `json:"template_binding,omitempty"`
	LibraryChrome   *LibraryChrome            `json:"library_chrome,omitempty"`
}
type Document struct {
	EditingProfile    string                         `json:"editing_profile,omitempty"`
	Title             string                         `json:"title,omitempty"`
	BuildIdentity     *BuildIdentity                 `json:"build_identity,omitempty"`
	Schema            string                         `json:"schema"`
	Year              int                            `json:"year"`
	Slides            []SlideSpec                    `json:"slides"`
	Sections          []SectionSpec                  `json:"sections,omitempty"`
	MediaOptimization *pptx.MediaOptimizationOptions `json:"media_optimization,omitempty"`
}
type BuildIdentity struct {
	Timestamp string `json:"timestamp"`
	Seed      string `json:"seed"`
}
type TextRecord struct {
	NativeParagraphContract string           `json:"native_paragraph_contract,omitempty"`
	NativeShape             *NativeTextShape `json:"native_shape,omitempty"`
	ID                      string           `json:"id"`
	Rect                    Rect             `json:"rect"`
	Color                   string           `json:"color"`
	Align                   string           `json:"align"`
	VerticalAlign           string           `json:"vertical_align,omitempty"`
	Layout                  TextLayout       `json:"layout"`
	Rich                    *RichTextLayout  `json:"rich,omitempty"`
	Rotation                float64          `json:"rotation_deg,omitempty"`
}

// NativeTextShape records the containing rectangle of the editable-block pilot.
// TextRecord.Rect remains the independently measured inner text allocation.
type NativeTextShape struct {
	Rect              Rect                 `json:"rect"`
	Fill              string               `json:"fill"`
	Line              *pptx.ShapeLineProps `json:"line,omitempty"`
	ParagraphContract string               `json:"paragraph_contract,omitempty"`
}
type NativeGeometryObservation struct {
	Name        string `json:"name"`
	WorldBounds Rect   `json:"world_bounds"`
}
type SlideReport struct {
	NativeGeometry  []NativeGeometryObservation `json:"native_geometry,omitempty"`
	ID              string                      `json:"id"`
	Density         *SlideDensityRecord         `json:"density,omitempty"`
	Hidden          bool                        `json:"hidden,omitempty"`
	Notes           string                      `json:"notes,omitempty"`
	DraftReview     *DraftReviewRecord          `json:"draft_review,omitempty"`
	Page            int                         `json:"page"`
	Frame           ResolvedFrame               `json:"frame"`
	Texts           []TextRecord                `json:"texts"`
	Nodes           []Node                      `json:"nodes"`
	Components      []ComponentRecord           `json:"components,omitempty"`
	CardRows        []CardRowRecord             `json:"card_rows,omitempty"`
	Shapes          []ShapeRecord               `json:"shapes,omitempty"`
	Scenes          []SceneRecord               `json:"scenes,omitempty"`
	Tables          []SceneTableRecord          `json:"tables,omitempty"`
	Charts          []SceneChartRecord          `json:"charts,omitempty"`
	TemplateBinding *TemplateSlideRecord        `json:"template_binding,omitempty"`
}
type Report struct {
	EditingProfile     string                        `json:"editing_profile,omitempty"`
	DensityAdjustments []SlideDensityAdjustment      `json:"density_adjustments,omitempty"`
	Schema             string                        `json:"schema"`
	Profile            string                        `json:"profile"`
	Engine             string                        `json:"engine"`
	SourceRevision     string                        `json:"source_revision"`
	SourceCommit       string                        `json:"source_commit,omitempty"`
	PowerPointVerified bool                          `json:"powerpoint_verified"`
	VisuallyReviewed   bool                          `json:"visually_reviewed"`
	Qualification      string                        `json:"qualification"`
	SourceFiles        []SourceFile                  `json:"source_files"`
	Fonts              []FontIdentity                `json:"fonts"`
	Assets             []SourceFile                  `json:"assets"`
	Slides             []SlideReport                 `json:"slides"`
	Sections           []SectionSpec                 `json:"sections,omitempty"`
	Warnings           []string                      `json:"warnings"`
	MeasurementPolicy  map[string]string             `json:"measurement_policy"`
	PPTXSHA256         string                        `json:"pptx_sha256"`
	MediaOptimization  *pptx.MediaOptimizationReport `json:"media_optimization,omitempty"`
}
type renderer struct {
	editingProfile      string
	contrastProbe       *contrastProbe
	bodyDensity         string
	headerDensity       string
	densityScope        string
	projectAssets       map[string]AssetData
	source              *Source
	typeEngine          *Typography
	bundle              string
	pres                *pptx.Presentation
	slide               *pptx.Slide
	master              *pptx.SlideMasterProps
	records             *[]TextRecord
	err                 error
	libraryChrome       *LibraryChrome
	sceneContext        SceneContext
	sceneTargets        map[string]annotationTarget
	editableTargets     map[string]Rect
	editableTargetNames map[string]string
}

func pos(r Rect) pptx.PositionProps {
	x, y, w, h := pptx.Inches(r.X/72), pptx.Inches(r.Y/72), pptx.Inches(r.W/72), pptx.Inches(r.H/72)
	return pptx.PositionProps{X: &x, Y: &y, W: &w, H: &h}
}
func zero() *float64 { v := 0.0; return &v }
func (r *renderer) ink(surface, role string) string {
	c, e := r.source.Ink(surface, role)
	if e != nil {
		r.err = e
	}
	return c
}
func (r *renderer) shape(id string, b Rect, color string, transparency float64) {
	if r.err != nil {
		return
	}
	p := &pptx.ShapeProps{PositionProps: pos(b), ObjectNameProps: pptx.ObjectNameProps{ObjectName: id}, Fill: &pptx.ShapeFillProps{Color: color, Transparency: transparency}, Line: &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: color, Transparency: 100}, Width: 0}}
	if r.typeEngine.engine == CandidateEngine {
		p.Line = &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Type: "none", Color: color}}
	}
	if r.master != nil {
		r.master.Objects = append(r.master.Objects, pptx.SlideMasterObject{Rect: p})
	} else {
		r.err = r.slide.AddShape(pptx.ShapeTypeRect, p)
	}
}
func (r *renderer) text(id, text string, style Style, b Rect, color, align string, maxLines int) {
	if r.err != nil {
		return
	}
	if strings.Contains(text, "[[") || strings.Contains(text, "]]") || strings.Contains(text, "[^") {
		r.err = fmt.Errorf("text.unsupported_emphasis_or_footnote: %s", id)
		return
	}
	if align == "" {
		align = "left"
	}
	if align != "left" && align != "right" && align != "center" {
		r.err = fmt.Errorf("text.unsupported_align: %s", align)
		return
	}
	l, e := r.measureText(text, style, b.W)
	if e != nil {
		r.err = fmt.Errorf("%s: %w", id, e)
		return
	}
	if maxLines > 0 && len(l.Lines) > maxLines {
		r.err = fmt.Errorf("text.line_allocation_exceeded: %s has %d lines, capacity %d", id, len(l.Lines), maxLines)
		return
	}
	// Allocation includes conservative terminal line height. No clipping, shrink,
	// substitution, or automatic density changes resolve capacity failures.
	need := math.Max(l.AllocationHeight, l.OccupiedTop+l.EstimatedOccupiedHeight)
	if b.H == 0 {
		b.H = need
	}
	if need > b.H+.02 {
		r.err = fmt.Errorf("text.vertical_overflow: %s needs %.3fpt, capacity %.3fpt", id, need, b.H)
		return
	}
	opts := &pptx.TextPropsOptions{PositionProps: pos(b), ObjectNameProps: pptx.ObjectNameProps{ObjectName: id}, TextBaseProps: pptx.TextBaseProps{FontFace: l.Font.Typeface, FontSize: style.Size, Bold: &l.Font.Bold, Italic: &l.Font.NativeItalic, Color: color, Align: pptx.HAlign(align)}, CharSpacing: style.TrackingPt, LineSpacing: style.Leading, ParaSpaceBefore: zero(), ParaSpaceAfter: zero(), Margin: pptx.Margin{0}, Fit: "none", Valign: pptx.VAlign("top")}
	p := pptx.TextProps{Text: l.Displayed, Options: opts}
	if r.master != nil {
		r.master.Objects = append(r.master.Objects, pptx.SlideMasterObject{Text: &p})
	} else {
		r.err = r.slide.AddText([]pptx.TextProps{{Text: l.Displayed}}, opts)
	}
	*r.records = append(*r.records, TextRecord{ID: id, Rect: b, Color: color, Align: align, Layout: l})
}

func contrast(a, b string) float64 {
	lum := func(s string) float64 {
		v, _ := strconv.ParseUint(s, 16, 32)
		channel := func(n uint64) float64 {
			x := float64(n) / 255
			if x <= .04045 {
				return x / 12.92
			}
			return math.Pow((x+.055)/1.055, 2.4)
		}
		return .2126*channel(v>>16&255) + .7152*channel(v>>8&255) + .0722*channel(v&255)
	}
	x, y := lum(a), lum(b)
	return (math.Max(x, y) + .05) / (math.Min(x, y) + .05)
}
func (r *renderer) logo(surface string, b Rect) {
	if r.err != nil {
		return
	}
	stem := "wm_h_pos_clr_rgb_august2024"
	if surface == "inverse" || surface == "deep" {
		stem = "wm_h_rev_wht_rgb_august2024"
	}
	svg, e := os.ReadFile(filepath.Join(r.bundle, "assets", stem+".svg"))
	if e != nil {
		r.err = e
		return
	}
	png, e := os.ReadFile(filepath.Join(r.bundle, "assets", stem+".png"))
	if e != nil {
		r.err = e
		return
	}
	// Read the original viewBox aspect ratio rather than fitting into a square.
	var dims [4]float64
	vb := regexp.MustCompile(`viewBox="([^"]+)"`).FindSubmatch(svg)
	if len(vb) != 2 {
		r.err = fmt.Errorf("logo.invalid_viewbox")
		return
	}
	if _, e = fmt.Sscanf(string(vb[1]), "%f %f %f %f", &dims[0], &dims[1], &dims[2], &dims[3]); e != nil {
		r.err = e
		return
	}
	b.W = b.H * dims[2] / dims[3]
	p := &pptx.ImageProps{PositionProps: pos(b), ObjectNameProps: pptx.ObjectNameProps{ObjectName: "wm.logo"}, DataOrPathProps: pptx.DataOrPathProps{Data: "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString(svg)}, SVGFallbackData: "data:image/png;base64," + base64.StdEncoding.EncodeToString(png), AltText: "West Monroe"}
	if r.master != nil {
		r.master.Objects = append(r.master.Objects, pptx.SlideMasterObject{Image: p})
	} else {
		r.err = r.slide.AddImage(p)
	}
}
func (r *renderer) chrome(f ResolvedFrame, year int) {
	q := f.Request
	if r.libraryChrome != nil {
		if len(r.libraryChrome.Tint) > 8 {
			r.err = fmt.Errorf("library.invalid_tint_count")
			return
		}
		for i, tint := range r.libraryChrome.Tint {
			if !intakeFinite(tint.X, tint.W) || tint.X < 0 || tint.W <= 0 || tint.X+tint.W > 960 {
				r.err = fmt.Errorf("library.invalid_tint_geometry")
				return
			}
			if tint.Surface == "" {
				tint.Surface = "subtle"
			}
			color, err := r.sceneColor(tint.Surface, "bg")
			if err != nil {
				r.err = err
				return
			}
			r.shape(fmt.Sprintf("wm.tint.%02d", i+1), Rect{tint.X, 0, tint.W, 540}, color, 0)
		}
	}
	if f.Panel.W > 0 {
		r.shape("wm.rail", f.Panel, r.ink(q.RailSurface, "bg"), 0)
	}
	if f.FooterBand.H > 0 {
		r.shape("wm.footer.band", f.FooterBand, r.ink("subtle", "bg"), 0)
	}
	if f.FooterRule > 0 {
		r.shape("wm.footer.rule", Rect{f.Body.X, f.FooterRule, f.Body.W, .75}, r.ink(q.Surface, "line"), 0)
	}
	// Editable square dots: named native variant of the whiteboard field. Each
	// opacity samples the renderer's radial mask at its center (not pixel parity).
	wb := f.Whiteboard
	dot := r.ink("light", "line")
	if q.Surface == "inverse" || q.Surface == "deep" {
		dot = r.ink("inverse", "line")
	}
	if r.libraryChrome == nil || !r.libraryChrome.CustomWhiteboard {
		for x := 0.0; x < wb.W; x += 18 {
			for y := 0.0; y < wb.H; y += 18 {
				distance := math.Hypot((x+1.5)/wb.W, (y+1.5)/wb.H)
				opacity := math.Min(1, math.Max(0, (1-distance)/.65))
				if opacity > 0 {
					r.shape(fmt.Sprintf("wm.whiteboard.%g.%g", x, y), Rect{wb.X + x, wb.Y + y, 3, 3}, dot, 100*(1-opacity))
				}
			}
		}
	}
	if f.TitleRule > 0 {
		r.shape("wm.header.rule", Rect{f.Header.X, f.TitleRule, f.Header.W, .75}, r.ink(q.Surface, "line"), 0)
	}
	fs := q.Surface
	if q.Footer == "tall" {
		fs = "subtle"
	}
	ls := fs
	if q.Rail == "left" {
		ls = q.RailSurface
	}
	logoY := f.FooterRow.Y + 3
	if q.Footer == "slim" {
		logoY = f.FooterRow.Y + (f.FooterRow.H-12)/2
	}
	r.logo(ls, Rect{57, logoY, 0, 12})
	legal := ""
	for _, c := range r.source.Frames.Chrome {
		if c.ID == "legal" {
			legal = strings.ReplaceAll(c.Text, "{year}", fmt.Sprint(year))
		}
	}
	st, _ := r.source.Style("footer")
	x := f.Body.X
	w := f.Body.W
	if q.Rail != "left" {
		x += 114
		w -= 114
	}
	if q.Rail != "right" && !q.NoPage {
		w -= 30
	}
	legalBox := Rect{x, f.FooterRow.Y, w, 18}
	if q.Footer == "slim" {
		legalBox = r.centerFooterText(legal, st, legalBox, f.FooterRow)
	}
	r.text("wm.legal", legal, st, legalBox, r.ink(fs, "secondary"), "right", 2)
}

func (r *renderer) centerFooterText(text string, style Style, box, row Rect) Rect {
	layout, err := r.measureText(text, style, box.W)
	if err != nil {
		r.err = err
		return box
	}
	box.H = math.Max(layout.AllocationHeight, layout.OccupiedTop+layout.EstimatedOccupiedHeight)
	box.Y = row.Y + (row.H-box.H)/2
	return box
}
func (r *renderer) nav(f ResolvedFrame) {
	q := f.Request
	if r.err != nil || q.Rail != "nav" {
		return
	}
	h := (f.NavBottom - 36 - float64(len(q.Nav)-1)*6) / float64(len(q.Nav))
	for i, tab := range q.Nav {
		surf := "subtle"
		if tab.ID == q.Active {
			surf = "inverse"
		}
		b := Rect{21, 36 + float64(i)*(h+6), 18, h}
		r.shape("nav."+tab.ID, b, r.ink(surf, "bg"), 0)
		st, _ := r.source.Style("label")
		labelWidth := h - 12
		if isModernLibrary(r.source.Revision) {
			// Source .gtab span is a distinct 8pt/600 Mono style with 0.1em
			// tracking, rather than the generic 9pt label token. It has no
			// CSS padding; reserve 2pt at each native end for terminal spacing.
			st.Size, st.Weight, st.Tracking, st.TrackingPt = 8, 600, "0.1em", .8
			labelWidth = h - 4
		}
		id, e := r.typeEngine.Resolve(st)
		if e != nil {
			r.err = e
			return
		}
		text := strings.ToUpper(tab.Label)
		l, e := r.measureText(text, st, labelWidth)
		if e != nil || len(l.Lines) != 1 {
			r.err = fmt.Errorf("frame.nav_label_does_not_fit: %s", tab.Label)
			return
		}
		color := r.ink(surf, "primary")
		if isModernLibrary(r.source.Revision) && tab.ID != q.Active {
			color = r.ink(surf, "secondary")
		}
		p := &pptx.TextPropsOptions{PositionProps: pos(Rect{b.X, b.Y, 18, h}), ObjectNameProps: pptx.ObjectNameProps{ObjectName: "nav.label." + tab.ID}, TextBaseProps: pptx.TextBaseProps{FontFace: id.Typeface, FontSize: st.Size, Bold: &id.Bold, Italic: &id.NativeItalic, Color: color, Align: pptx.HAlign("center")}, CharSpacing: st.TrackingPt, LineSpacing: st.Leading, Margin: pptx.Margin{0}, Valign: pptx.VAlign("mid"), Vert: "vert270", Fit: "none", ParaSpaceBefore: zero(), ParaSpaceAfter: zero()}
		r.err = r.slide.AddText([]pptx.TextProps{{Text: text}}, p)
		if r.typeEngine.engine == CandidateEngine {
			*r.records = append(*r.records, TextRecord{ID: "nav.label." + tab.ID, Rect: Rect{b.X, b.Y, 18, h}, Color: color, Align: "center", Layout: l})
		}
		if r.err != nil {
			return
		}
	}
}
func inside(b, zone Rect) bool {
	return b.X >= zone.X-.01 && b.Y >= zone.Y-.01 && b.X+b.W <= zone.X+zone.W+.01 && b.Y+b.H <= zone.Y+zone.H+.01
}

func Build(bundle, sourceOverride string, doc Document) ([]byte, Report, error) {
	return BuildWithEngine(bundle, sourceOverride, doc, Engine)
}

func BuildWithEngine(bundle, sourceOverride string, doc Document, engine string) ([]byte, Report, error) {
	return BuildWithEngineAndAssets(bundle, sourceOverride, doc, engine, nil)
}

// BuildWithEngineAndAssets resolves project-owned media without mutable global
// registries or environment changes. The caller retains canonical asset paths.
func BuildWithEngineAndAssets(bundle, sourceOverride string, doc Document, engine string, assets map[string]AssetData) ([]byte, Report, error) {
	var report Report
	s, e := Load(bundle, sourceOverride)
	if e != nil {
		return nil, report, e
	}
	return buildWithLoadedSource(bundle, s, doc, engine, assets)
}

// buildWithLoadedSource keeps prototype source observations internal. Public
// builds still require Load's registered, checksum-verified source bundle.
func buildWithLoadedSource(bundle string, s *Source, doc Document, engine string, assets map[string]AssetData) (output []byte, report Report, buildErr error) {
	t, err := NewSourceTypographyEngine(s, filepath.Join(bundle, "fonts"), engine)
	if err != nil {
		return nil, report, err
	}
	return buildWithSlideDensities(bundle, s, t, doc, engine, assets)
}

func buildWithTypography(bundle string, s *Source, t *Typography, doc Document, engine string, assets map[string]AssetData, layoutOnly bool) (output []byte, report Report, buildErr error) {
	activeBody := false
	activeSlideID := ""
	defer func() {
		buildErr = bodyDensityFitFailure(buildErr, activeBody)
		if buildErr != nil && activeSlideID != "" && !strings.HasPrefix(buildErr.Error(), "slide "+activeSlideID) {
			buildErr = fmt.Errorf("slide %s: %w", activeSlideID, buildErr)
		}
	}()
	var e error
	for _, st := range s.Tokens.Type {
		if _, e = t.Resolve(st); e != nil {
			return nil, report, e
		}
	}
	if doc.Schema != "pptxgengo.wmds-foundation.v1" || doc.Year < 2000 || doc.Year > 9999 || len(doc.Slides) == 0 {
		return nil, report, fmt.Errorf("document.invalid_schema_year_or_slides")
	}
	if err := ValidateEditingProfile(doc.EditingProfile); err != nil {
		return nil, report, err
	}
	if doc.EditingProfile == NativeEditingProfile && engine != CandidateEngine {
		return nil, report, fmt.Errorf("native editing profile requires v2")
	}
	slideIDs := make([]string, len(doc.Slides))
	for i, slide := range doc.Slides {
		if err := ValidateDraftReviewNote(slide.DraftReview); err != nil {
			return nil, report, fmt.Errorf("slide %s: %w", slide.ID, err)
		}
		if err := ValidateSpeakerNotes(slide.Notes); err != nil {
			return nil, report, fmt.Errorf("slide %s: %w", slide.ID, err)
		}
		slideIDs[i] = slide.ID
	}
	if err := ValidateSections(doc.Sections, slideIDs); err != nil {
		return nil, report, err
	}
	p := pptx.New()
	if doc.BuildIdentity != nil {
		ts, err := time.Parse(time.RFC3339, doc.BuildIdentity.Timestamp)
		if err != nil {
			return nil, report, fmt.Errorf("document.invalid_build_timestamp: %w", err)
		}
		if err = p.SetBuildIdentity(pptx.BuildIdentity{Timestamp: ts, Seed: doc.BuildIdentity.Seed}); err != nil {
			return nil, report, err
		}
	}
	p.DefineLayout("WMDS", 960.0/72, 540.0/72)
	if e = p.SetLayout("WMDS"); e != nil {
		return nil, report, e
	}
	p.Title = "WMDS foundation reference"
	if doc.Title != "" {
		p.Title = doc.Title
	}
	p.Author = "West Monroe"
	p.Theme = pptx.ThemeProps{HeadFontFace: "IBM Plex Sans SemiBold", BodyFontFace: "IBM Plex Sans"}
	report = Report{EditingProfile: doc.EditingProfile, Schema: "pptxgengo.wmds-layout.v1", Profile: ProfileForEngine(engine), Engine: engine, Qualification: "implemented_unqualified", SourceFiles: s.Files, Fonts: t.Fonts(), Warnings: []string{"Go first-baseline/occupied-height predictions require native calibration for the new exact-leading profile.", "Native font identity, wrapping, visual quality and overflow remain unqualified until PowerPoint capture/review.", "Whiteboard variant: editable dots with center-sampled radial opacity; no browser pixel identity claim.", "Source scene components and closed template content bindings require v2. Catalog availability does not establish successful source rendering or native visual qualification."}}
	report.SourceRevision, report.SourceCommit = s.Revision, s.Commit
	report.MeasurementPolicy = map[string]string{"leading": "exact authored points between predicted baselines", "first_baseline": "provisional 0.9 em; native calibration pending", "occupied_height": "provisional 1.2 em plus leading between lines; native calibration pending", "advance": "Harfbuzz shaped at 64x then quantized to 1/64 pt; no legacy 1/8 pt correction", "tracking": "serialized 0.01 pt then applied per cluster; terminal cluster tracking excluded from line width", "ligatures": "liga/clig enabled at zero tracking, disabled at nonzero tracking; native feature parity pending", "paragraphs": "CRLF/CR normalized to LF; hard breaks and empty paragraphs preserved"}
	if engine == CandidateEngine {
		report.Warnings[0] = "v2 passes the 139 declared native controls; arbitrary content and exact native font file identity remain unqualified."
		report.MeasurementPolicy = map[string]string{
			"leading":         "exact authored points between predicted baselines",
			"first_baseline":  "pinned v1 native anchors for exact source font/size/leading; otherwise provisional 0.75 leading",
			"occupied_height": "visible line union, adopted terminal height; empty paragraphs reserve leading without visible ink",
			"advance":         "Latin glyph advances rounded to 1/8 point before tracking; candidate hypothesis",
			"tracking":        "serialized 0.01 point per cluster, terminal tracking included; retained at 64x layout precision",
			"ligatures":       "kern disabled, liga/clig enabled regardless of tracking; candidate hypothesis",
			"paragraphs":      "CRLF/CR normalized to LF; explicit paragraph for each hard break including trailing/empty paragraphs, with complete end defaults",
			"font_names":      "original IBM family and PostScript names; no renamed fonts",
		}
		if len(t.densityAnchors) > 0 {
			report.MeasurementPolicy["density_calibration_sha256"] = densityCalibrationForSource(s)
			report.MeasurementPolicy["density_leading"] = "observed native line pitch for exact supplemental font/size/authored-leading keys; authored paragraph spacing unchanged"
			report.MeasurementPolicy["density_qualification"] = "same-environment native controls; arbitrary content and native font file identity remain unqualified"
		}
	}
	entries, e := os.ReadDir(filepath.Join(bundle, "assets"))
	if e != nil {
		return nil, report, e
	}
	for _, f := range entries {
		b, e := os.ReadFile(filepath.Join(bundle, "assets", f.Name()))
		if e != nil {
			return nil, report, e
		}
		report.Assets = append(report.Assets, SourceFile{"assets/" + f.Name(), fmt.Sprintf("%x", sha256.Sum256(b))})
	}
	r := renderer{editingProfile: doc.EditingProfile, source: s, typeEngine: t, bundle: bundle, pres: p, projectAssets: assets}
	type sharedFrame struct {
		name  string
		texts []TextRecord
	}
	frameLayouts := map[string]sharedFrame{}
	seen := map[string]bool{}
	sectionAnchors := map[string]string{}
	for _, section := range doc.Sections {
		p.AddSection(pptx.SectionProps{Title: section.Title})
		sectionAnchors[section.BeforeSlideID] = section.Title
	}
	report.Sections = append([]SectionSpec(nil), doc.Sections...)
	sectionTitle := ""
	for i, slide := range doc.Slides {
		activeSlideID = slide.ID
		activeBody = false
		density, err := slideDensity(s, slide)
		if err != nil {
			return nil, report, err
		}
		r.bodyDensity, r.headerDensity, r.densityScope = density.Resolved, density.Header, "body"
		if slide.ID == "" || seen[slide.ID] {
			return nil, report, fmt.Errorf("slide.invalid_or_duplicate_id")
		}
		seen[slide.ID] = true
		f, e := s.ResolveFrame(slide.Frame)
		if e != nil {
			return nil, report, e
		}
		if f.Request.NoHeader && (slide.Title != "" || slide.Eyebrow != "") {
			return nil, report, fmt.Errorf("frame.no_header_with_header_content")
		}
		if (slide.Source != "") != (f.Request.SourceLines > 0) {
			return nil, report, fmt.Errorf("frame.source_line_allocation_required")
		}
		if slide.ContentKind != "" && slide.ContentKind != "synthetic_example" && slide.ContentKind != "supplied_content" {
			return nil, report, fmt.Errorf("slide.invalid_content_kind: %s", slide.ID)
		}
		sr := SlideReport{ID: slide.ID, Hidden: slide.Hidden, Page: i + 1, Frame: f, TemplateBinding: slide.TemplateBinding}
		r.libraryChrome = slide.LibraryChrome
		if err := r.registerEditableTargets(slide); err != nil {
			return nil, report, err
		}
		if err := r.registerSceneTargets(slide, f); err != nil {
			return nil, report, err
		}
		r.records = &sr.Texts
		var tint []LibraryTint
		if slide.LibraryChrome != nil {
			tint = slide.LibraryChrome.Tint
		}
		frameKeyBytes, _ := json.Marshal(struct {
			Frame            ResolvedFrame
			Year             int
			CustomWhiteboard bool
			Tint             []LibraryTint `json:",omitempty"`
		}{f, doc.Year, slide.LibraryChrome != nil && slide.LibraryChrome.CustomWhiteboard, tint})
		frameKey := fmt.Sprintf("%x", sha256.Sum256(frameKeyBytes))
		shared, exists := frameLayouts[frameKey]
		if !exists {
			shared.name = "wmds.frame." + frameKey[:20]
			r.master = &pptx.SlideMasterProps{Title: shared.name, Background: &pptx.BackgroundProps{ShapeFillProps: pptx.ShapeFillProps{Color: r.ink(f.Request.Surface, "bg")}}, Margin: pptx.Margin{0}}
			r.chrome(f, doc.Year)
			if r.err != nil {
				return nil, report, r.err
			}
			if e = p.DefineSlideMaster(r.master); e != nil {
				return nil, report, e
			}
			shared.texts = append([]TextRecord(nil), sr.Texts...)
			frameLayouts[frameKey] = shared
		} else {
			sr.Texts = append(sr.Texts, shared.texts...)
		}
		r.master = nil
		if title, ok := sectionAnchors[slide.ID]; ok {
			sectionTitle = title
		}
		r.slide = p.AddSlide(&pptx.AddSlideProps{MasterName: shared.name, SectionTitle: sectionTitle})
		if slide.Hidden {
			hidden := true
			r.slide.PresSlide().Hidden = &hidden
		}
		r.slide.PresSlide().Name = slide.ID
		if slide.LibraryChrome != nil && slide.LibraryChrome.CustomWhiteboard {
			if err := r.drawLibraryWhiteboards(slide.LibraryChrome, f, false); err != nil {
				return nil, report, err
			}
		}
		hasComponents := false
		for _, n := range slide.Nodes {
			if n.Kind == "card" || n.Kind == "textblock" || n.Kind == "cardrow" || n.Kind == "metric" || n.Kind == "richtext" {
				hasComponents = true
			}
		}
		if slide.Notes != "" {
			r.slide.AddNotes(slide.Notes + "\n\n")
		}
		if slide.TemplateBinding != nil {
			r.slide.AddNotes(fmt.Sprintf("WMDS bound template. Content classification: %s. Frozen source and binding provenance are recorded in layout-report.json and binding-report.json.", slide.ContentKind))
		} else if hasComponents {
			r.slide.AddNotes("Synthetic WMDS component reference. Example engagement details are fictional. Design sources: frozen components/v0/components.json, tokens/v0/tokens.json and explorations/components.src.html; source hashes and resolved font identities are recorded in layout-report.json.")
		} else {
			r.slide.AddNotes("WMDS foundation reference; synthetic labels. Source files and exact font identities are in layout-report.json. This Go layout profile requires PowerPoint qualification.")
		}
		if !f.Request.NoHeader {
			st, _ := r.headerStyle("eyebrow")
			eyebrowWidth := f.Header.W
			if slide.LibraryChrome != nil && slide.LibraryChrome.Stamp != "" {
				eyebrowWidth -= 180
			}
			r.text("eyebrow", slide.Eyebrow, st, Rect{f.Header.X, 36, eyebrowWidth, 12}, r.ink(f.Request.Surface, "emphasis"), "left", 1)
			st, _ = r.headerStyle(f.TitleStyle)
			if slide.LibraryChrome != nil && (strings.Contains(slide.Title, "[[") || strings.Contains(slide.Title, "[^")) {
				plan := &scenePlan{ID: "library-header", Bounds: Rect{f.Header.X, 54, f.Header.W, f.TitleRule - 54}}
				ctx := SceneContext{Surface: f.Request.Surface, Zone: f.Header, Path: "/title", Notes: slide.LibraryChrome.Notes}
				if err := r.primitiveRichText(plan, "title", slide.Title, st, plan.Bounds, f.Request.Surface, "display", "left", slide.LibraryChrome.Emphasis, "", ctx); err != nil {
					return nil, report, err
				}
				if err := r.drawScene(plan, &sr, "/title"); err != nil {
					return nil, report, err
				}
			} else {
				r.text("title", slide.Title, st, Rect{f.Header.X, 54, f.Header.W, f.TitleRule - 54}, r.ink(f.Request.Surface, "display"), "left", f.Request.TitleLines)
			}
		}
		if slide.Source != "" {
			st, _ := s.Style("source")
			r.text("source", slide.Source, st, f.Source, r.ink(f.Request.Surface, "secondary"), "left", f.Request.SourceLines)
		}
		r.nav(f)
		if !f.Request.NoPage {
			st, _ := s.Style("footer")
			st.Weight = 600
			surf := f.Request.Surface
			if f.Request.Footer == "tall" {
				surf = "subtle"
			}
			if f.Request.Rail == "right" {
				surf = f.Request.RailSurface
			}
			pageBox := Rect{879, f.FooterRow.Y, 24, 18}
			if f.Request.Footer == "slim" {
				pageBox = r.centerFooterText(fmt.Sprint(i+1), st, pageBox, f.FooterRow)
			}
			r.text("wm.page", fmt.Sprint(i+1), st, pageBox, r.ink(surf, "primary"), "right", 1)
		}
		if r.err != nil {
			return nil, report, r.err
		}
		activeBody = true
		nodeIDs := map[string]bool{}
		for _, n := range slide.Nodes {
			if n.ID == "" || nodeIDs[n.ID] || n.ID == "title" || n.ID == "eyebrow" || n.ID == "source" || strings.HasPrefix(n.ID, "wm.") || strings.HasPrefix(n.ID, "nav.") {
				return nil, report, fmt.Errorf("node.invalid_or_duplicate_id")
			}
			nodeIDs[n.ID] = true
			if n.Kind != "textblock" && n.TextBlock != nil || n.Kind != "card" && n.Card != nil ||
				n.Kind != "cardrow" && n.CardRow != nil || n.Kind != "metric" && n.DataMetric != nil ||
				n.Kind != "richtext" && n.RichText != nil || n.Kind != "scene" && n.Scene != nil {
				return nil, report, fmt.Errorf("node.unexpected_component_payload: %s", n.ID)
			}
			zone := f.Body
			surf := f.Request.Surface
			if n.Scope == "rail" {
				zone = f.Rail
				surf = f.Request.RailSurface
			} else if n.Scope == "short" && f.Request.Split != "" {
				zone = f.ShortBody
			} else if n.Scope == "tall" && f.Request.Split != "" {
				zone = f.TallBody
			} else if n.Scope != "" && n.Scope != "body" {
				return nil, report, fmt.Errorf("node.unsupported_scope")
			}
			b := n.Rect
			if math.IsNaN(b.X+b.Y+b.W+b.H) || math.IsInf(b.X+b.Y+b.W+b.H, 0) || b.W < 0 || b.H < 0 {
				return nil, report, fmt.Errorf("node.invalid_geometry: %s", n.ID)
			}
			if n.Grid != "" {
				var span Rect
				var err error
				switch n.Grid {
				case "columns":
					span, err = s.Tokens.Grid.Span(n.Start, n.Span)
				case "five-up":
					span, err = s.Tokens.Grid.FiveSpan(n.Start, n.Span, f.Request.Rail)
				default:
					err = fmt.Errorf("grid.unknown: %s", n.Grid)
				}
				if err != nil {
					return nil, report, err
				}
				if b.X != 0 && math.Abs(b.X-span.X) > .01 || b.W != 0 && math.Abs(b.W-span.W) > .01 {
					return nil, report, fmt.Errorf("grid.conflicting_geometry: %s", n.ID)
				}
				b.X, b.W = span.X, span.W
			} else if n.Start != 0 || n.Span != 0 {
				return nil, report, fmt.Errorf("grid.span_without_grid: %s", n.ID)
			}
			if f.Request.Split != "" && n.Kind != "scene" && (n.Scope == "" || n.Scope == "body") {
				if b.X >= f.TallBody.X-.01 && b.X+b.W <= f.TallBody.X+f.TallBody.W+.01 {
					zone = f.TallBody
				} else {
					zone = f.ShortBody
				}
			}
			if n.Surface != "" {
				if (n.Kind == "text" || n.Kind == "textblock" || n.Kind == "metric" || n.Kind == "richtext" || n.Kind == "rule") && n.Surface != surf {
					return nil, report, fmt.Errorf("text.surface_context_requires_matching_zone: %s", n.ID)
				}
				surf = n.Surface
			}
			if (n.Kind == "card" || n.Kind == "cardrow") && n.Surface == "" {
				surf = "light"
			}
			role := n.Ink
			if role == "" {
				role = "primary"
			}
			switch n.Kind {
			case "scene":
				if n.Scene == nil || len(n.Scene.Node) == 0 {
					return nil, report, fmt.Errorf("scene.missing_source_node: %s", n.ID)
				}
				sceneZone := Rect{f.Body.X, 0, f.Body.W, f.Body.Y + f.Body.H}
				var sourceTag struct {
					Type string  `json:"type"`
					Kind string  `json:"kind"`
					X    float64 `json:"x"`
					Y    float64 `json:"y"`
					W    float64 `json:"w"`
				}
				if err := json.Unmarshal(n.Scene.Node, &sourceTag); err != nil {
					return nil, report, err
				}
				if isExpandedLibrary(s.Revision) && sourceTag.Type == "maturity" {
					sceneZone = f.Body
				}
				if f.Rail.W > 0 && sourceTag.X >= f.Rail.X-.02 && sourceTag.X < f.Rail.X+f.Rail.W {
					sceneZone = Rect{f.Rail.X, 0, f.Rail.W, f.Rail.Y + f.Rail.H}
					surf = f.Request.RailSurface
				}
				if f.Request.Split != "" {
					if sourceTag.Type == "connector" {
						sceneZone = f.Body
					} else if sourceTag.X >= f.TallBody.X-.02 && sourceTag.X+sourceTag.W <= f.TallBody.X+f.TallBody.W+.02 {
						sceneZone = f.TallBody
					} else {
						sceneZone = f.ShortBody
					}
					// The refreshed quadrant allocates its tall plot at y27; its
					// outer border deliberately bleeds 9pt above the ordinary zone.
					// Measured text still obeys the ordinary footer reservation.
					if isModernLibrary(s.Revision) && sourceTag.Type == "chart" && sourceTag.Kind == "quadrant" && sourceTag.Y == 27 && sceneZone == f.TallBody {
						sceneZone.Y -= 9
						sceneZone.H += 9
					}
				}
				switch sourceTag.Type {
				case "imageframe", "square", "logo", "art", "mark", "thumbnail":
					sceneZone = Rect{0, 0, 960, 540}
				}
				if n.Scene.Allocation != nil {
					// Local components plan within their own complete allocation,
					// including label headroom and footer clearance. The frame
					// body may be larger than this component's usable space.
					sceneZone = *n.Scene.Allocation
				}
				ctx := SceneContext{Surface: surf, Zone: sceneZone, Path: n.Scene.Path, Keys: n.Scene.Keys, Notes: n.Scene.Notes}
				plan, err := r.planSceneNode(n.ID, n.Scene.Node, ctx)
				if err != nil {
					return nil, report, fmt.Errorf("slide %s/%s: %w", slide.ID, n.ID, err)
				}
				if n.Scene.Allocation != nil && !inside(plan.Bounds, *n.Scene.Allocation) {
					return nil, report, fmt.Errorf("scene.component_exceeds_allocation: %s at %+v, allocation %+v", n.ID, plan.Bounds, *n.Scene.Allocation)
				}
				splitBoundsOK := inside(plan.Bounds, f.ShortBody) || inside(plan.Bounds, f.TallBody)
				if isModernLibrary(s.Revision) && sourceTag.Type == "chart" && sourceTag.Kind == "quadrant" && sourceTag.Y == 27 && sceneZone.Y == 27 {
					splitBoundsOK = inside(plan.Bounds, sceneZone)
				}
				if f.Request.Split != "" && sceneZone.W != 960 && !splitBoundsOK {
					err := fmt.Errorf("scene.outside_split_zone: %s/%s: %+v", slide.ID, n.ID, plan.Bounds)
					if densityTextOnlyBottomOverflow(plan, sceneZone) {
						return nil, report, &densityFitError{cause: err}
					}
					return nil, report, err
				}
				plan.Warnings = append(plan.Warnings, n.Scene.Resolutions...)
				if err = r.drawScene(plan, &sr, n.Scene.Path); err != nil {
					return nil, report, err
				}
				b = plan.Bounds
			case "rule":
				sh, err := r.planRule(n, b, zone, surf)
				if err != nil {
					return nil, report, err
				}
				r.shape(sh.ID, sh.Rect, sh.Color, 0)
				b = sh.Rect
				sr.Shapes = append(sr.Shapes, sh)
			case "textblock", "card":
				plan, err := r.planComponent(n, b, zone, surf)
				if err != nil {
					return nil, report, err
				}
				r.drawComponent(plan)
				b = plan.record.Rect
				sr.Components = append(sr.Components, plan.record)
			case "cardrow":
				plan, err := r.planCardRow(n, b, zone, surf)
				if err != nil {
					return nil, report, err
				}
				r.drawCardRow(plan)
				b = plan.record.Rect
				sr.CardRows = append(sr.CardRows, plan.record)
				for _, child := range plan.children {
					sr.Components = append(sr.Components, child.record)
				}
			case "metric":
				plan, err := r.planDataMetric(n, b, zone, surf)
				if err != nil {
					return nil, report, err
				}
				r.drawComponent(plan)
				b = plan.record.Rect
				sr.Components = append(sr.Components, plan.record)
			case "richtext":
				tr, err := r.planRich(n, b, zone, surf, n.RichText)
				if err != nil {
					return nil, report, err
				}
				r.drawRich(tr)
				b = tr.Rect
			case "text":
				st, e := r.bodyStyle(n.Style)
				if e != nil {
					return nil, report, e
				}
				minimum := 4.5
				if st.Size >= 18 || st.Size >= 14 && st.Weight >= 700 {
					minimum = 3
				}
				foreground, e := s.Ink(surf, role)
				if e != nil {
					return nil, report, e
				}
				background, e := s.Ink(surf, "bg")
				if e != nil {
					return nil, report, e
				}
				if contrast(foreground, background) < minimum {
					return nil, report, fmt.Errorf("text.insufficient_contrast: %s", n.ID)
				}
				if strings.Contains(n.Text, "[[") || strings.Contains(n.Text, "[^") {
					return nil, report, fmt.Errorf("text.unsupported_emphasis_or_footnote")
				}
				if b.H == 0 {
					l, e := t.Measure(n.Text, st, b.W)
					if e != nil {
						return nil, report, e
					}
					b.H = math.Max(l.AllocationHeight, l.OccupiedTop+l.EstimatedOccupiedHeight)
				}
				if !inside(b, zone) {
					return nil, report, fmt.Errorf("node.outside_zone: %s", n.ID)
				}
				r.text(n.ID, n.Text, st, b, r.ink(surf, role), n.Align, 0)
			case "box":
				if n.Text != "" || n.Style != "" || n.Ink != "" || n.Align != "" {
					return nil, report, fmt.Errorf("node.unsupported_box_fields: %s", n.ID)
				}
				if b.H <= 0 || !inside(b, zone) {
					return nil, report, fmt.Errorf("node.outside_zone: %s", n.ID)
				}
				if n.Grid == "five-up" {
					if math.Abs(b.Y/18-math.Round(b.Y/18)) > 1e-8 || math.Abs(b.H/18-math.Round(b.H/18)) > 1e-8 {
						return nil, report, fmt.Errorf("grid.five_up_off_rhythm: %s", n.ID)
					}
				} else {
					if e = s.Tokens.Grid.OuterBox(b); e != nil {
						return nil, report, e
					}
				}
				if surf == "outline" {
					inner := Rect{b.X + .5, b.Y + .5, b.W - 1, b.H - 1}
					props := &pptx.ShapeProps{PositionProps: pos(inner), ObjectNameProps: pptx.ObjectNameProps{ObjectName: n.ID}, Fill: &pptx.ShapeFillProps{Color: r.ink(surf, "bg")}, Line: &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: r.ink(surf, "line")}, Width: 1}}
					if r.err == nil {
						r.err = r.slide.AddShape(pptx.ShapeTypeRect, props)
					}
				} else {
					r.shape(n.ID, b, r.ink(surf, "bg"), 0)
				}
			default:
				return nil, report, fmt.Errorf("node.unsupported_kind: %s", n.Kind)
			}
			n.Rect = b
			sr.Nodes = append(sr.Nodes, n)
		}
		if slide.LibraryChrome != nil {
			if err := r.drawLibraryWhiteboards(slide.LibraryChrome, f, true); err != nil {
				return nil, report, err
			}
			if err := r.drawLibraryStamp(slide.LibraryChrome, f, &sr); err != nil {
				return nil, report, err
			}
		}
		if r.err != nil {
			return nil, report, r.err
		}
		activeBody = false
		if slide.DraftReview != nil {
			if err := r.drawDraftReview(slide.ID, *slide.DraftReview, &sr); err != nil {
				return nil, report, err
			}
		}
		if e = validateOwnedParts(sr); e != nil {
			return nil, report, e
		}
		report.Slides = append(report.Slides, sr)
	}
	activeSlideID = ""
	if layoutOnly {
		return nil, report, nil
	}
	raw, e := p.Write()
	if e != nil {
		return nil, report, e
	}
	if engine == CandidateEngine {
		records := map[int][]TextRecord{}
		for i, sr := range report.Slides {
			records[i+1] = sr.Texts
			if sr.DraftReview != nil {
				records[i+1] = append(append([]TextRecord(nil), sr.Texts...), sr.DraftReview.Texts...)
			}
			for _, tr := range sr.Texts {
				if tr.Rich != nil {
					for _, paragraph := range tr.Rich.Paragraphs {
						for _, run := range paragraph.Runs {
							if !run.VerticalAnchored {
								report.Warnings = append(report.Warnings, fmt.Sprintf("Uncalibrated mixed-run vertical combination: slide %d, %s/%s/%s (%s/%d, %.2fpt size, %.2fpt leading).", i+1, tr.ID, paragraph.Key, run.Key, run.Style.Family, run.Style.Weight, run.Style.Size, run.Style.Leading))
							}
						}
					}
				} else if tr.Layout.CalibrationSHA256 == "" {
					report.Warnings = append(report.Warnings, fmt.Sprintf("Uncalibrated vertical combination: slide %d, %s (%s/%d, %.2fpt size, %.2fpt leading).", i+1, tr.ID, tr.Layout.Style.Family, tr.Layout.Style.Weight, tr.Layout.Style.Size, tr.Layout.Style.Leading))
				}
			}
		}
		raw, e = candidateParagraphs(raw, records)
		if e != nil {
			return nil, report, e
		}
	}
	raw, e = explicitTypography(raw)
	if e != nil {
		return nil, report, e
	}
	raw, e = componentGroups(raw, report.Slides)
	if e != nil {
		return nil, report, e
	}
	raw, e = cardRowGroups(raw, report.Slides)
	if e != nil {
		return nil, report, e
	}
	raw, e = sceneTableTypography(raw, report.Slides)
	if e != nil {
		return nil, report, e
	}
	raw, e = sceneNativeGroups(raw, report.Slides)
	if e != nil {
		return nil, report, e
	}
	raw, e = draftReviewNativeMetadata(raw, report.Slides)
	if e != nil {
		return nil, report, e
	}
	mediaPolicy := pptx.DeliveryMediaOptions()
	if doc.MediaOptimization != nil {
		mediaPolicy = *doc.MediaOptimization
	}
	raw, mediaReport, e := pptx.OptimizeMedia(raw, mediaPolicy)
	if e != nil {
		return nil, report, e
	}
	report.MediaOptimization = &mediaReport
	report.PPTXSHA256 = fmt.Sprintf("%x", sha256.Sum256(raw))
	return raw, report, nil
}

// This profile-specific DrawingML pass supplies explicit paragraph defaults and
// zero kerning thresholds. It never changes the shared writer or legacy decks.
func explicitTypography(raw []byte) ([]byte, error) {
	z, e := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if e != nil {
		return nil, e
	}
	var out bytes.Buffer
	w := zip.NewWriter(&out)
	paragraphs := regexp.MustCompile(`(?s)<a:p>.*?</a:p>`)
	run := regexp.MustCompile(`(?s)<a:rPr\b[^>]*>.*?</a:rPr>`)
	end := regexp.MustCompile(`<a:endParaRPr\b[^>]*/>|(?s)<a:endParaRPr\b[^>]*>.*?</a:endParaRPr>`)
	shape := regexp.MustCompile(`(?s)<p:sp>.*?</p:sp>`)
	for _, f := range z.File {
		rc, e := f.Open()
		if e != nil {
			return nil, e
		}
		b, e := io.ReadAll(rc)
		rc.Close()
		if e != nil {
			return nil, e
		}
		if strings.HasPrefix(f.Name, "ppt/slides/slide") && strings.HasSuffix(f.Name, ".xml") || strings.HasPrefix(f.Name, "ppt/slideLayouts/") && strings.HasSuffix(f.Name, ".xml") {
			b = paragraphs.ReplaceAllFunc(b, func(p []byte) []byte {
				p = run.ReplaceAllFunc(p, func(r []byte) []byte {
					if !bytes.Contains(r, []byte(` kern=`)) {
						r = bytes.Replace(r, []byte(` dirty=`), []byte(` kern="0" spc="0" dirty=`), 1)
					}
					if !bytes.Contains(r, []byte(` b=`)) {
						r = bytes.Replace(r, []byte(` dirty=`), []byte(` b="0" dirty=`), 1)
					}
					if !bytes.Contains(r, []byte(` i=`)) {
						r = bytes.Replace(r, []byte(` dirty=`), []byte(` i="0" dirty=`), 1)
					}
					return r
				})
				if existing := end.Find(p); bytes.Contains(existing, []byte("<a:latin")) && bytes.Contains(existing, []byte(` sz=`)) && bytes.Contains(existing, []byte(` kern=`)) && bytes.Contains(existing, []byte(` spc=`)) && bytes.Contains(existing, []byte(` b=`)) && bytes.Contains(existing, []byte(` i=`)) {
					return p
				}
				base := run.Find(p)
				if base == nil {
					return p
				}
				base = bytes.ReplaceAll(base, []byte("a:rPr"), []byte("a:endParaRPr"))
				p = end.ReplaceAll(p, nil)
				return bytes.Replace(p, []byte("</a:p>"), append(base, []byte("</a:p>")...), 1)
			})
			b = shape.ReplaceAllFunc(b, func(sp []byte) []byte {
				if !bytes.Contains(sp, []byte(`name="wm.page"`)) {
					return sp
				}
				return bytes.ReplaceAll(bytes.ReplaceAll(sp, []byte("<a:r>"), []byte(`<a:fld id="{7DC738B1-48F4-45B7-954D-37669918DC70}" type="slidenum">`)), []byte("</a:r>"), []byte("</a:fld>"))
			})
		}
		h := f.FileHeader
		dst, e := w.CreateHeader(&h)
		if e != nil {
			return nil, e
		}
		if _, e = dst.Write(b); e != nil {
			return nil, e
		}
	}
	if e = w.Close(); e != nil {
		return nil, e
	}
	return out.Bytes(), nil
}
func WriteJSON(path string, value any) error {
	b, e := json.MarshalIndent(value, "", "  ")
	if e != nil {
		return e
	}
	return os.WriteFile(path, append(b, '\n'), 0644)
}
