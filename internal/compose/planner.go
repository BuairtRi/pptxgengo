// Package compose validates dynamic component specifications and plans native
// geometry from caller-supplied text measurements. It never estimates text.
package compose

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

const SpecSchema = "pptxgengo.compose-spec.v1"
const PlanSchema = "pptxgengo.compose-plan.v1"

// Rect uses points. Input pod bounds use PodBounds because their height is a
// maximum; output Rect heights are the measured, planned heights.
type Rect struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// AnchorSet exposes stable edge-midpoint and center locations for a parent
// compositor to route connectors after pod geometry is fixed.
type AnchorSet struct {
	Top    Point `json:"top"`
	Bottom Point `json:"bottom"`
	Left   Point `json:"left"`
	Right  Point `json:"right"`
	Center Point `json:"center"`
}

type PodBounds struct {
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Width       float64 `json:"width"`
	MaxHeightPt float64 `json:"max_height_pt"`
}

type Spec struct {
	Schema string      `json:"schema"`
	Slides []SlideSpec `json:"slides"`
}

type SlideSpec struct {
	Accents             []AccentSpec         `json:"accents,omitempty"`
	ArtworkArrows       []ArtworkArrowSpec   `json:"artwork_arrows,omitempty"`
	Role                string               `json:"role,omitempty"`
	Takeaway            string               `json:"takeaway,omitempty"`
	Notes               string               `json:"notes,omitempty"`
	Canvas              []CanvasSpec         `json:"canvas,omitempty"`
	Cards               []CardSpec           `json:"cards,omitempty"`
	ID                  string               `json:"id"`
	Title               string               `json:"title"`
	WidthPt             float64              `json:"width_pt"`
	HeightPt            float64              `json:"height_pt"`
	TitleBounds         Rect                 `json:"title_bounds"`
	TitleFontFace       string               `json:"title_font_face"`
	TitleFontSizePt     float64              `json:"title_font_size_pt"`
	TitleBold           bool                 `json:"title_bold"`
	TitleForeground     string               `json:"title_foreground"`
	Pods                []PodSpec            `json:"pods"`
	Roles               []StandaloneRoleSpec `json:"roles,omitempty"`
	Phases              []PhaseSpec          `json:"phases,omitempty"`
	Legend              *LegendSpec          `json:"legend,omitempty"`
	Connections         []ConnectionSpec     `json:"connections,omitempty"`
	Layouts             []ContainerSpec      `json:"layouts,omitempty"`
	Paths               []ProcessPathSpec    `json:"paths,omitempty"`
	resolvedLayoutPorts []ResolvedLayoutPort
}

type PodSpec struct {
	ID     string     `json:"id"`
	Title  string     `json:"title"`
	Bounds PodBounds  `json:"bounds"`
	Layout LayoutSpec `json:"layout"`
	Style  PodStyle   `json:"style"`
	Roles  []RoleSpec `json:"roles"`
}

type LayoutSpec struct {
	Columns        int     `json:"columns"` // 0 selects automatically
	GapPt          float64 `json:"gap_pt"`
	PaddingPt      float64 `json:"padding_pt"`
	TitleGapPt     float64 `json:"title_gap_pt"`
	MinTileWidthPt float64 `json:"min_tile_width_pt"`
}

type PodStyle struct {
	FontFace          string  `json:"font_face"`
	FontSizePt        float64 `json:"font_size_pt"`
	Bold              bool    `json:"bold"`
	TitleFontFace     string  `json:"title_font_face"`
	TitleFontSizePt   float64 `json:"title_font_size_pt"`
	TitleBold         bool    `json:"title_bold"`
	TileMinHeightPt   float64 `json:"tile_min_height_pt"`
	HorizontalInsetPt float64 `json:"horizontal_inset_pt"`
	VerticalInsetPt   float64 `json:"vertical_inset_pt"`
	ParagraphGapPt    float64 `json:"paragraph_gap_pt"`
	Surface           string  `json:"surface"`
	Foreground        string  `json:"foreground"`
	TitleForeground   string  `json:"title_foreground"`
}

type RoleSpec struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	Background string `json:"background"`
	Foreground string `json:"foreground"` // auto or explicit #RRGGBB
}

// ProbeRequest describes one text-measurement operation. TextWidthPt is the
// usable inner width; the caller should measure the text with zero textbox
// margins. Insets are returned separately for tile geometry and auditability.
type ProbeRequest struct {
	PhraseRequests    []PhraseRequest `json:"phrase_requests,omitempty"`
	Align             string          `json:"align,omitempty"`
	ID                string          `json:"id"`
	SlideID           string          `json:"slide_id"`
	PodID             string          `json:"pod_id,omitempty"`
	RoleID            string          `json:"role_id,omitempty"`
	PhaseID           string          `json:"phase_id,omitempty"`
	LegendToken       string          `json:"legend_token,omitempty"`
	Kind              string          `json:"kind"` // slide_title, pod_title, role
	Columns           int             `json:"columns"`
	Text              string          `json:"text"`
	Paragraphs        []ParagraphSpec `json:"paragraphs,omitempty"`
	TextWidthPt       float64         `json:"text_width_pt"`
	HorizontalInsetPt float64         `json:"horizontal_inset_pt"`
	VerticalInsetPt   float64         `json:"vertical_inset_pt"`
	FontFace          string          `json:"font_face"`
	FontSizePt        float64         `json:"font_size_pt"`
	Bold              bool            `json:"bold"`
	Foreground        string          `json:"foreground"`
	Background        string          `json:"background"`
}

type Measurement struct {
	PhraseBounds     map[string][]Rect `json:"phrase_bounds,omitempty"`
	OffsetXPt        float64           `json:"offset_x_pt,omitempty"`
	OffsetYPt        float64           `json:"offset_y_pt,omitempty"`
	RenderedWidthPt  float64           `json:"rendered_width_pt"`
	RenderedHeightPt float64           `json:"rendered_height_pt"`
}

type Measurements struct {
	ByRequestID map[string]Measurement `json:"by_request_id"`
}

type PlanResult struct {
	Schema string         `json:"schema"`
	Slides []PlannedSlide `json:"slides"`
}

type PlannedSlide struct {
	Accents            []PlannedAccent       `json:"accents,omitempty"`
	ArtworkArrows      []PlannedArtworkArrow `json:"artwork_arrows,omitempty"`
	Canvas             []PlannedCanvas       `json:"canvas,omitempty"`
	Cards              []PlannedCard         `json:"cards,omitempty"`
	ID                 string                `json:"id"`
	Title              string                `json:"title"`
	TitleFontFace      string                `json:"title_font_face"`
	TitleFontSizePt    float64               `json:"title_font_size_pt"`
	TitleBold          bool                  `json:"title_bold"`
	TitleForeground    string                `json:"title_foreground"`
	TitleBounds        Rect                  `json:"title_bounds"`
	WidthPt            float64               `json:"width_pt"`
	HeightPt           float64               `json:"height_pt"`
	TitleMeasurementID string                `json:"title_measurement_id,omitempty"`
	Pods               []PlannedPod          `json:"pods"`
	Roles              []PlannedRole         `json:"roles,omitempty"`
	Phases             []PlannedPhase        `json:"phases,omitempty"`
	Legend             *PlannedLegend        `json:"legend,omitempty"`
	Connections        []PlannedConnection   `json:"connections,omitempty"`
	LayoutPorts        []ResolvedLayoutPort  `json:"layout_ports,omitempty"`
	ManualRequired     []string              `json:"manual_required,omitempty"`
}

type PlannedPod struct {
	ID                 string        `json:"id"`
	Title              string        `json:"title"`
	Columns            int           `json:"columns"`
	Bounds             Rect          `json:"bounds"`
	Anchors            AnchorSet     `json:"anchors"`
	Surface            string        `json:"surface"`
	Foreground         string        `json:"foreground"`
	TitleForeground    string        `json:"title_foreground"`
	TitleFontFace      string        `json:"title_font_face"`
	TitleFontSizePt    float64       `json:"title_font_size_pt"`
	TitleBold          bool          `json:"title_bold"`
	TitleRect          Rect          `json:"title_rect"`
	TitleMeasurementID string        `json:"title_measurement_id,omitempty"`
	Roles              []PlannedRole `json:"roles"`
}

type PlannedRole struct {
	ID                string    `json:"id"`
	Label             string    `json:"label"`
	Bounds            Rect      `json:"bounds"`
	Anchors           AnchorSet `json:"anchors"`
	Background        string    `json:"background"`
	Foreground        string    `json:"foreground"`
	FontFace          string    `json:"font_face"`
	FontSizePt        float64   `json:"font_size_pt"`
	Bold              bool      `json:"bold"`
	HorizontalInsetPt float64   `json:"horizontal_inset_pt"`
	VerticalInsetPt   float64   `json:"vertical_inset_pt"`
	MeasurementID     string    `json:"measurement_id"`
}

const (
	navy    = "#070154"
	blue    = "#0047FF"
	magenta = "#F900D3"
	white   = "#FFFFFF"
	light   = "#E8EEF8"
)

func finite(v float64) bool      { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func positive(v float64) bool    { return finite(v) && v > 0 }
func nonnegative(v float64) bool { return finite(v) && v >= 0 }
func validID(s string) bool      { return strings.TrimSpace(s) != "" }

func requestID(parts ...any) string {
	b, _ := json.Marshal(parts)
	return "measure:" + base64.RawURLEncoding.EncodeToString(b)
}

func resolveColor(s string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "staffing.wm_full_time":
		return navy, nil
	case "staffing.wm_part_time":
		return blue, nil
	case "staffing.client_part_time":
		return magenta, nil
	case "surface.light":
		return light, nil
	case "surface.white":
		return white, nil
	}
	if len(s) != 7 || s[0] != '#' {
		return "", fmt.Errorf("invalid color %q: expected #RRGGBB or a supported semantic token", s)
	}
	for _, r := range s[1:] {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return "", fmt.Errorf("invalid color %q: expected #RRGGBB", s)
		}
	}
	return strings.ToUpper(s), nil
}

func luminance(hex string) float64 {
	var vals [3]float64
	for i := 0; i < 3; i++ {
		u, _ := strconv.ParseUint(hex[1+i*2:3+i*2], 16, 8)
		v := float64(u)
		v /= 255
		if v <= 0.04045 {
			vals[i] = v / 12.92
		} else {
			vals[i] = math.Pow((v+0.055)/1.055, 2.4)
		}
	}
	return 0.2126*vals[0] + 0.7152*vals[1] + 0.0722*vals[2]
}

// ContrastRatio implements WCAG relative-luminance contrast.
func ContrastRatio(a, b string) (float64, error) {
	ca, err := resolveColor(a)
	if err != nil {
		return 0, err
	}
	cb, err := resolveColor(b)
	if err != nil {
		return 0, err
	}
	la, lb := luminance(ca), luminance(cb)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05), nil
}

func foreground(requested, bg string) (string, error) {
	return foregroundWithThreshold(requested, bg, 4.5)
}

// WCAG large text permits 3:1 at >=18pt regular or >=14pt bold. This only
// applies to a uniform text shape; mixed runs keep their own stricter checks.
func foregroundAtSize(requested, bg string, size float64, bold bool) (string, error) {
	threshold := 4.5
	if size >= 18 || (bold && size >= 14) {
		threshold = 3
	}
	return foregroundWithThreshold(requested, bg, threshold)
}

func foregroundWithThreshold(requested, bg string, threshold float64) (string, error) {
	b, err := resolveColor(bg)
	if err != nil {
		return "", err
	}
	if strings.EqualFold(strings.TrimSpace(requested), "auto") {
		navyRatio, _ := ContrastRatio(navy, b)
		whiteRatio, _ := ContrastRatio(white, b)
		fg, ratio := navy, navyRatio
		if whiteRatio > navyRatio {
			fg, ratio = white, whiteRatio
		}
		if ratio < threshold {
			return "", fmt.Errorf("no automatic brand foreground reaches %.1f:1 contrast", threshold)
		}
		return fg, nil
	}
	f, err := resolveColor(requested)
	if err != nil {
		return "", err
	}
	ratio, _ := ContrastRatio(f, b)
	if ratio < threshold {
		return "", fmt.Errorf("foreground %s on %s has %.2f:1 contrast; minimum is %.1f:1", f, b, ratio, threshold)
	}
	return f, nil
}

func maxHeight(p PodSpec) float64 { return p.Bounds.MaxHeightPt }

func titleProbe(s SlideSpec) ProbeRequest {
	return ProbeRequest{ID: requestID(s.ID, "slide_title"), SlideID: s.ID, Kind: "slide_title", Text: s.Title, TextWidthPt: s.TitleBounds.Width, FontFace: s.TitleFontFace, FontSizePt: s.TitleFontSizePt, Bold: s.TitleBold, Foreground: s.TitleForeground, Background: white}
}

// ProbeRequests returns all text measurements needed to evaluate every allowed
// column count. The caller must key its measurements by each returned ID.
func ProbeRequests(spec Spec) ([]ProbeRequest, error) {
	expanded, err := ExpandLayoutsForProbes(spec)
	if err != nil {
		return nil, err
	}
	return probeRequestsExpanded(expanded)
}

// probeRequestsExpanded operates on a spec whose declarative layouts have
// already been lowered. Keeping this separate prevents recursive expansion.
func probeRequestsExpanded(spec Spec) ([]ProbeRequest, error) {
	if err := validateSpec(spec); err != nil {
		return nil, err
	}
	var out []ProbeRequest
	for _, s := range spec.Slides {
		if strings.TrimSpace(s.Title) != "" {
			out = append(out, titleProbe(s))
		}
		for _, p := range s.Pods {
			titleColor, err := foreground(p.Style.TitleForeground, p.Style.Surface)
			if err != nil {
				return nil, fmt.Errorf("slide %s pod %s title: %w", s.ID, p.ID, err)
			}
			if strings.TrimSpace(p.Title) != "" {
				id := requestID(s.ID, p.ID, "pod_title")
				out = append(out, ProbeRequest{ID: id, SlideID: s.ID, PodID: p.ID, Kind: "pod_title", Text: p.Title, TextWidthPt: p.Bounds.Width - 2*p.Layout.PaddingPt, FontFace: p.Style.TitleFontFace, FontSizePt: p.Style.TitleFontSizePt, Bold: p.Style.TitleBold, Foreground: titleColor, Background: p.Style.Surface})
			}
			for _, cols := range candidateColumns(p) {
				tileW := tileWidth(p, cols)
				innerW := tileW - 2*p.Style.HorizontalInsetPt
				for _, role := range p.Roles {
					bg, _ := resolveColor(role.Background)
					fg, err := foreground(role.Foreground, bg)
					if err != nil {
						return nil, fmt.Errorf("slide %s pod %s role %s: %w", s.ID, p.ID, role.ID, err)
					}
					out = append(out, ProbeRequest{ID: requestID(s.ID, p.ID, cols, role.ID), SlideID: s.ID, PodID: p.ID, RoleID: role.ID, Kind: "role", Columns: cols, Text: role.Label, TextWidthPt: innerW, HorizontalInsetPt: p.Style.HorizontalInsetPt, VerticalInsetPt: p.Style.VerticalInsetPt, FontFace: p.Style.FontFace, FontSizePt: p.Style.FontSizePt, Bold: p.Style.Bold, Foreground: fg, Background: bg})
				}
			}
		}
		team, err := teamProbes(s)
		if err != nil {
			return nil, err
		}
		out = append(out, team...)
		cards, err := cardProbes(s)
		if err != nil {
			return nil, err
		}
		out = append(out, cards...)
		out = append(out, canvasProbes(s)...)
	}
	for _, q := range out {
		for _, r := range q.Text {
			if (r < 0x20 && r != '\t' && r != '\n' && r != '\r') || r == 0xFFFE || r == 0xFFFF {
				return nil, fmt.Errorf("measurement %s contains an unsupported XML text character", q.ID)
			}
		}
	}
	return out, nil
}

func candidateColumns(p PodSpec) []int {
	if p.Layout.Columns > 0 {
		if tileWidth(p, p.Layout.Columns) >= p.Layout.MinTileWidthPt && tileWidth(p, p.Layout.Columns) > 2*p.Style.HorizontalInsetPt {
			return []int{p.Layout.Columns}
		}
		return nil
	}
	avail := p.Bounds.Width - 2*p.Layout.PaddingPt
	maxCols := int(math.Floor((avail + p.Layout.GapPt) / (p.Layout.MinTileWidthPt + p.Layout.GapPt)))
	if maxCols < 1 {
		return nil
	}
	if maxCols > len(p.Roles) {
		maxCols = len(p.Roles)
	}
	out := make([]int, maxCols)
	n := 0
	for cols := 1; cols <= maxCols; cols++ {
		w := tileWidth(p, cols)
		if w >= p.Layout.MinTileWidthPt && w > 2*p.Style.HorizontalInsetPt {
			out[n] = cols
			n++
		}
	}
	return out[:n]
}

func tileWidth(p PodSpec, cols int) float64 {
	return (p.Bounds.Width - 2*p.Layout.PaddingPt - float64(cols-1)*p.Layout.GapPt) / float64(cols)
}

func validateSpec(spec Spec) error {
	if spec.Schema != SpecSchema {
		return fmt.Errorf("schema must be %q", SpecSchema)
	}
	if len(spec.Slides) == 0 {
		return fmt.Errorf("spec must contain at least one slide")
	}
	slideIDs := map[string]bool{}
	for _, s := range spec.Slides {
		if !validID(s.ID) || slideIDs[s.ID] {
			return fmt.Errorf("slide IDs must be nonempty and unique: %q", s.ID)
		}
		slideIDs[s.ID] = true
		if !positive(s.WidthPt) || !positive(s.HeightPt) {
			return fmt.Errorf("slide %s dimensions must be positive finite points", s.ID)
		}
		if !ValidFontFace(s.TitleFontFace) || !positive(s.TitleFontSizePt) {
			return fmt.Errorf("slide %s requires an explicit title font family and positive point size", s.ID)
		}
		if strings.TrimSpace(s.Title) != "" {
			if !validRect(s.TitleBounds) || !inside(s.TitleBounds, Rect{Width: s.WidthPt, Height: s.HeightPt}) {
				return fmt.Errorf("slide %s title bounds are invalid or outside slide", s.ID)
			}
		}
		if strings.TrimSpace(s.Title) != "" {
			if _, e := foreground(s.TitleForeground, white); e != nil {
				return fmt.Errorf("slide %s title: %w", s.ID, e)
			}
		}
		podIDs := map[string]bool{}
		componentIDs := map[string]bool{}
		for _, p := range s.Pods {
			if !validID(p.ID) || podIDs[p.ID] {
				return fmt.Errorf("slide %s pod IDs must be nonempty and unique: %q", s.ID, p.ID)
			}
			podIDs[p.ID] = true
			if componentIDs[p.ID] {
				return fmt.Errorf("slide %s component ID is duplicated: %q", s.ID, p.ID)
			}
			componentIDs[p.ID] = true
			if !validID(p.Title) {
				return fmt.Errorf("slide %s pod %s requires a title", s.ID, p.ID)
			}
			mh := maxHeight(p)
			if !positive(p.Bounds.Width) || !positive(mh) || !nonnegative(p.Bounds.X) || !nonnegative(p.Bounds.Y) {
				return fmt.Errorf("slide %s pod %s bounds must be finite with positive width/height and nonnegative origin", s.ID, p.ID)
			}
			if p.Bounds.X+p.Bounds.Width > s.WidthPt+1e-6 || p.Bounds.Y+mh > s.HeightPt+1e-6 {
				return fmt.Errorf("slide %s pod %s bounds exceed slide", s.ID, p.ID)
			}
			l := p.Layout
			if l.Columns < 0 || !nonnegative(l.GapPt) || !nonnegative(l.PaddingPt) || !nonnegative(l.TitleGapPt) || !positive(l.MinTileWidthPt) {
				return fmt.Errorf("slide %s pod %s has invalid layout values", s.ID, p.ID)
			}
			if p.Bounds.Width-2*l.PaddingPt <= 0 {
				return fmt.Errorf("slide %s pod %s has no usable title/content width after padding", s.ID, p.ID)
			}
			st := p.Style
			if !ValidFontFace(st.FontFace) || !ValidFontFace(st.TitleFontFace) || !positive(st.FontSizePt) || !positive(st.TitleFontSizePt) || !positive(st.TileMinHeightPt) || !nonnegative(st.HorizontalInsetPt) || !nonnegative(st.VerticalInsetPt) || st.ParagraphGapPt != 0 {
				return fmt.Errorf("slide %s pod %s requires explicit font families, zero paragraph_gap_pt and valid tile dimensions", s.ID, p.ID)
			}
			if tileWidth(p, 1)-2*st.HorizontalInsetPt <= 0 {
				return fmt.Errorf("slide %s pod %s has no usable role text width", s.ID, p.ID)
			}
			if _, e := resolveColor(st.Surface); e != nil {
				return fmt.Errorf("slide %s pod %s surface: %w", s.ID, p.ID, e)
			}
			if _, e := foreground(st.Foreground, st.Surface); e != nil {
				return fmt.Errorf("slide %s pod %s foreground: %w", s.ID, p.ID, e)
			}
			if _, e := foreground(st.TitleForeground, st.Surface); e != nil {
				return fmt.Errorf("slide %s pod %s title: %w", s.ID, p.ID, e)
			}
			if len(p.Roles) == 0 {
				return fmt.Errorf("slide %s pod %s requires at least one role", s.ID, p.ID)
			}
			for _, r := range p.Roles {
				if !validID(r.ID) || componentIDs[r.ID] {
					return fmt.Errorf("slide %s role IDs must be nonempty and unique across slide: %q", s.ID, r.ID)
				}
				componentIDs[r.ID] = true
				if !validID(r.Label) {
					return fmt.Errorf("slide %s pod %s role %s requires a nonempty label", s.ID, p.ID, r.ID)
				}
				if _, e := resolveColor(r.Background); e != nil {
					return fmt.Errorf("slide %s pod %s role %s background: %w", s.ID, p.ID, r.ID, e)
				}
				if _, e := foreground(r.Foreground, r.Background); e != nil {
					return fmt.Errorf("slide %s pod %s role %s: %w", s.ID, p.ID, r.ID, e)
				}
			}
			if l.Columns > 0 && l.Columns > len(p.Roles) {
				return fmt.Errorf("slide %s pod %s explicit columns exceed role count", s.ID, p.ID)
			}
			if len(candidateColumns(p)) == 0 || tileWidth(p, candidateColumns(p)[0])-2*st.HorizontalInsetPt <= 0 {
				return fmt.Errorf("slide %s pod %s cannot fit a role tile", s.ID, p.ID)
			}
		}
		if err := validateTeam(s, componentIDs); err != nil {
			return err
		}
		if err := validateCards(s, componentIDs); err != nil {
			return err
		}
		if err := validateAccents(s, componentIDs); err != nil {
			return err
		}
		if err := validateArtworkArrows(s, componentIDs); err != nil {
			return err
		}
		if err := validateCanvas(s, componentIDs); err != nil {
			return err
		}
		if err := validateConnectionSpecs(s); err != nil {
			return err
		}
	}
	return nil
}

func validRect(r Rect) bool {
	return nonnegative(r.X) && nonnegative(r.Y) && positive(r.Width) && positive(r.Height)
}
func inside(a, b Rect) bool {
	return a.X >= b.X && a.Y >= b.Y && a.X+a.Width <= b.X+b.Width+1e-6 && a.Y+a.Height <= b.Y+b.Height+1e-6
}

func rectAnchors(r Rect) AnchorSet {
	cx, cy := r.X+r.Width/2, r.Y+r.Height/2
	return AnchorSet{
		Top: Point{X: cx, Y: r.Y}, Bottom: Point{X: cx, Y: r.Y + r.Height},
		Left: Point{X: r.X, Y: cy}, Right: Point{X: r.X + r.Width, Y: cy}, Center: Point{X: cx, Y: cy},
	}
}
func overlap(a, b Rect) bool {
	return a.X < b.X+b.Width-1e-6 && b.X < a.X+a.Width-1e-6 && a.Y < b.Y+b.Height-1e-6 && b.Y < a.Y+a.Height-1e-6
}

type candidatePlan struct {
	cols   int
	height float64
	roles  []PlannedRole
	rows   []float64
}

// Plan plans pods from exact caller-supplied text dimensions. Missing or
// malformed evidence is an error; candidate-specific width or height failures
// simply disqualify that auto-column candidate.
func Plan(spec Spec, measurements Measurements) (PlanResult, error) {
	probeSpec, err := ExpandLayoutsForProbes(spec)
	if err != nil {
		return PlanResult{}, err
	}
	requests, err := probeRequestsExpanded(probeSpec)
	if err != nil {
		return PlanResult{}, err
	}
	known := map[string]ProbeRequest{}
	for _, q := range requests {
		known[q.ID] = q
	}
	for id, m := range measurements.ByRequestID {
		if _, ok := known[id]; !ok {
			return PlanResult{}, fmt.Errorf("unexpected measurement %q", id)
		}
		if !positive(m.RenderedWidthPt) || !positive(m.RenderedHeightPt) {
			return PlanResult{}, fmt.Errorf("measurement %s must have positive finite dimensions", id)
		}
	}
	for id := range known {
		if _, ok := measurements.ByRequestID[id]; !ok {
			return PlanResult{}, fmt.Errorf("missing measurement %s", id)
		}
	}
	spec, err = ExpandLayouts(spec, measurements)
	if err != nil {
		return PlanResult{}, err
	}
	result := PlanResult{Schema: PlanSchema}
	for _, s := range spec.Slides {
		titleFg := s.TitleForeground
		if strings.TrimSpace(s.Title) != "" {
			titleFg, _ = foreground(s.TitleForeground, white)
		}
		ps := PlannedSlide{ID: s.ID, Title: s.Title, TitleFontFace: s.TitleFontFace, TitleFontSizePt: s.TitleFontSizePt, TitleBold: s.TitleBold, TitleForeground: titleFg, TitleBounds: s.TitleBounds, WidthPt: s.WidthPt, HeightPt: s.HeightPt}
		ps.LayoutPorts = append(ps.LayoutPorts, s.resolvedLayoutPorts...)
		if strings.TrimSpace(s.Title) != "" {
			id := titleProbe(s).ID
			m := measurements.ByRequestID[id]
			q := known[id]
			if m.RenderedWidthPt > q.TextWidthPt+1e-6 || m.RenderedHeightPt > s.TitleBounds.Height+1e-6 {
				return PlanResult{}, fmt.Errorf("slide %s title does not fit its supplied bounds", s.ID)
			}
			ps.TitleMeasurementID = id
		}
		for _, p := range s.Pods {
			candidates := []candidatePlan{}
			for _, cols := range candidateColumns(p) {
				w := tileWidth(p, cols)
				if w+1e-6 < p.Layout.MinTileWidthPt || w-2*p.Style.HorizontalInsetPt <= 0 {
					continue
				}
				var titleH float64
				titleID := ""
				if strings.TrimSpace(p.Title) != "" {
					titleID = requestID(s.ID, p.ID, "pod_title")
					m := measurements.ByRequestID[titleID]
					q := known[titleID]
					if m.RenderedWidthPt > q.TextWidthPt+1e-6 {
						continue
					}
					titleH = m.RenderedHeightPt
				}
				rows := (len(p.Roles) + cols - 1) / cols
				heights := make([]float64, rows)
				roles := make([]PlannedRole, len(p.Roles))
				fits := true
				for i, r := range p.Roles {
					rid := requestID(s.ID, p.ID, cols, r.ID)
					m := measurements.ByRequestID[rid]
					q := known[rid]
					if m.RenderedWidthPt > q.TextWidthPt+1e-6 {
						fits = false
						break
					}
					h := math.Max(p.Style.TileMinHeightPt, m.RenderedHeightPt+2*p.Style.VerticalInsetPt+2*p.Style.ParagraphGapPt)
					row := i / cols
					if h > heights[row] {
						heights[row] = h
					}
					bg, _ := resolveColor(r.Background)
					fg, _ := foreground(r.Foreground, bg)
					roles[i] = PlannedRole{ID: r.ID, Label: r.Label, Background: bg, Foreground: fg, FontFace: p.Style.FontFace, FontSizePt: p.Style.FontSizePt, Bold: p.Style.Bold, HorizontalInsetPt: p.Style.HorizontalInsetPt, VerticalInsetPt: p.Style.VerticalInsetPt, MeasurementID: rid}
				}
				if !fits {
					continue
				}
				total := 2*p.Layout.PaddingPt + titleH
				if titleH > 0 {
					total += p.Layout.TitleGapPt
				}
				for _, h := range heights {
					total += h
				}
				if rows > 1 {
					total += float64(rows-1) * p.Layout.GapPt
				}
				if total > maxHeight(p)+1e-6 {
					continue
				}
				candidates = append(candidates, candidatePlan{cols: cols, height: total, roles: roles, rows: heights})
			}
			if len(candidates) == 0 {
				return PlanResult{}, fmt.Errorf("slide %s pod %s has no fitting column layout for %d roles within %.2f × %.2fpt at %.2fpt text; increase bounds, allow more columns, or revise labels", s.ID, p.ID, len(p.Roles), p.Bounds.Width, maxHeight(p), p.Style.FontSizePt)
			}
			sort.SliceStable(candidates, func(i, j int) bool {
				if candidates[i].cols != candidates[j].cols {
					return candidates[i].cols < candidates[j].cols
				}
				return candidates[i].height < candidates[j].height
			})
			chosen := candidates[0]
			actual := Rect{X: p.Bounds.X, Y: p.Bounds.Y, Width: p.Bounds.Width, Height: chosen.height}
			if !inside(actual, Rect{Width: s.WidthPt, Height: s.HeightPt}) {
				return PlanResult{}, fmt.Errorf("slide %s pod %s planned bounds exceed slide", s.ID, p.ID)
			}
			podFg, _ := foreground(p.Style.Foreground, p.Style.Surface)
			titleFg, _ := foreground(p.Style.TitleForeground, p.Style.Surface)
			surface, _ := resolveColor(p.Style.Surface)
			pp := PlannedPod{ID: p.ID, Title: p.Title, Columns: chosen.cols, Bounds: actual, Anchors: rectAnchors(actual), Surface: surface, Foreground: podFg, TitleForeground: titleFg, TitleFontFace: p.Style.TitleFontFace, TitleFontSizePt: p.Style.TitleFontSizePt, TitleBold: p.Style.TitleBold, Roles: chosen.roles}
			y := p.Bounds.Y + p.Layout.PaddingPt
			if strings.TrimSpace(p.Title) != "" {
				id := requestID(s.ID, p.ID, "pod_title")
				m := measurements.ByRequestID[id]
				pp.TitleRect = Rect{X: p.Bounds.X + p.Layout.PaddingPt, Y: y, Width: p.Bounds.Width - 2*p.Layout.PaddingPt, Height: m.RenderedHeightPt}
				pp.TitleMeasurementID = id
				y += m.RenderedHeightPt + p.Layout.TitleGapPt
			}
			tileW := tileWidth(p, chosen.cols)
			rowY := y
			currentRow := 0
			for i := range pp.Roles {
				row, col := i/chosen.cols, i%chosen.cols
				if row != currentRow {
					rowY += chosen.rows[currentRow] + p.Layout.GapPt
					currentRow = row
				}
				rh := chosen.rows[row]
				pp.Roles[i].Bounds = Rect{X: p.Bounds.X + p.Layout.PaddingPt + float64(col)*(tileW+p.Layout.GapPt), Y: rowY, Width: tileW, Height: rh}
				pp.Roles[i].Anchors = rectAnchors(pp.Roles[i].Bounds)
			}
			// Check that the supplied text extents fit each computed role tile.
			for i := range pp.Roles {
				r := &pp.Roles[i]
				m := measurements.ByRequestID[r.MeasurementID]
				innerH := r.Bounds.Height - 2*p.Style.VerticalInsetPt - 2*p.Style.ParagraphGapPt
				if m.RenderedHeightPt > innerH+1e-6 {
					return PlanResult{}, fmt.Errorf("slide %s pod %s role %s text height exceeds tile", s.ID, p.ID, r.ID)
				}
			}
			ps.Pods = append(ps.Pods, pp)
		}
		for i, a := range ps.Pods {
			for _, b := range ps.Pods[i+1:] {
				if overlap(a.Bounds, b.Bounds) {
					return PlanResult{}, fmt.Errorf("slide %s pods %s and %s overlap", s.ID, a.ID, b.ID)
				}
			}
			if ps.TitleMeasurementID != "" && overlap(a.Bounds, ps.TitleBounds) {
				return PlanResult{}, fmt.Errorf("slide %s title and pod %s overlap", s.ID, a.ID)
			}
		}
		if err := planTeam(s, &ps, measurements); err != nil {
			return PlanResult{}, err
		}
		if err := planCards(s, &ps, measurements); err != nil {
			return PlanResult{}, err
		}
		if err := planCanvas(s, &ps, measurements); err != nil {
			return PlanResult{}, err
		}
		if err := planConnections(s, &ps); err != nil {
			return PlanResult{}, err
		}
		if err := planAccents(s, &ps, measurements); err != nil {
			return PlanResult{}, err
		}
		if err := planArtworkArrows(s, &ps, measurements); err != nil {
			return PlanResult{}, err
		}
		result.Slides = append(result.Slides, ps)
	}
	return result, nil
}
