package compose

import (
	"fmt"
	"strings"
)

// CanvasSpec is an explicitly positioned native element. Overlaps require named
// peers; this makes backgrounds/accents intentional rather than silent collisions.
// Asset bytes are verified by the CLI against SHA256 before rendering.
type CanvasSpec struct {
	ID         string          `json:"id"`
	Layer      int             `json:"layer,omitempty"`
	Kind       string          `json:"kind"` // text, surface, shape, line, image
	Bounds     Rect            `json:"bounds"`
	Text       string          `json:"text,omitempty"`
	Paragraphs []ParagraphSpec `json:"paragraphs,omitempty"`
	FontFace   string          `json:"font_face,omitempty"`
	FontSizePt float64         `json:"font_size_pt,omitempty"`
	Bold       bool            `json:"bold,omitempty"`
	Foreground string          `json:"foreground,omitempty"`
	Background string          `json:"background,omitempty"`
	// ContrastBackground supplies the effective inherited surface used for
	// color validation and measurement without painting a text-box fill.
	ContrastBackground  string         `json:"contrast_background,omitempty"`
	InsetX              float64        `json:"inset_x,omitempty"`
	InsetY              float64        `json:"inset_y,omitempty"`
	Align               string         `json:"align,omitempty"`
	Valign              string         `json:"valign,omitempty"`
	LineWidthPt         float64        `json:"line_width_pt,omitempty"`
	AssetPath           string         `json:"asset_path,omitempty"`
	AssetSHA256         string         `json:"asset_sha256,omitempty"`
	FallbackAssetPath   string         `json:"fallback_asset_path,omitempty"`
	FallbackAssetSHA256 string         `json:"fallback_asset_sha256,omitempty"`
	AltText             string         `json:"alt_text,omitempty"`
	OutlineColor        string         `json:"outline_color,omitempty"`
	OutlineWidthPt      float64        `json:"outline_width_pt,omitempty"`
	ImageFit            string         `json:"image_fit,omitempty"`
	ImageCrop           *ImageCropSpec `json:"image_crop,omitempty"`
	FocalX              *float64       `json:"focal_x,omitempty"`
	FocalY              *float64       `json:"focal_y,omitempty"`
	Preset              string         `json:"preset,omitempty"`
	Adjustments         map[string]int `json:"adjustments,omitempty"`
	RotationDeg         float64        `json:"rotation_deg,omitempty"`
	ArcStartDeg         float64        `json:"arc_start_deg,omitempty"`
	ArcEndDeg           float64        `json:"arc_end_deg,omitempty"`
	ArcThicknessRatio   float64        `json:"arc_thickness_ratio,omitempty"`
	Pattern             *PatternSpec   `json:"pattern,omitempty"`
	AllowOverlap        []string       `json:"allow_overlap,omitempty"`
}
type PlannedCanvas struct {
	CanvasSpec
	MeasurementID  string          `json:"measurement_id,omitempty"`
	PhraseRequests []PhraseRequest `json:"phrase_requests,omitempty"`
}

func validateCanvas(s SlideSpec, ids map[string]bool) error {
	for _, c := range s.Canvas {
		if !validID(c.ID) || ids[c.ID] || !validLayer(c.Layer) {
			return fmt.Errorf("slide %s duplicate/empty canvas ID %q", s.ID, c.ID)
		}
		ids[c.ID] = true
		if c.Kind == "line" {
			if !nonnegative(c.Bounds.X) || !nonnegative(c.Bounds.Y) || !nonnegative(c.Bounds.Width) || !nonnegative(c.Bounds.Height) || (c.Bounds.Width == 0) == (c.Bounds.Height == 0) || !positive(c.LineWidthPt) || c.LineWidthPt > 6 {
				return fmt.Errorf("canvas line %s requires one positive axis and width in (0,6]", c.ID)
			}
		} else if !validRect(c.Bounds) {
			return fmt.Errorf("canvas %s invalid bounds", c.ID)
		}
		if !inside(c.Bounds, Rect{Width: s.WidthPt, Height: s.HeightPt}) {
			return fmt.Errorf("canvas %s exceeds slide", c.ID)
		}
		if c.Kind != "shape" && (c.Preset != "" || len(c.Adjustments) != 0 || c.Pattern != nil || c.RotationDeg != 0 || c.ArcStartDeg != 0 || c.ArcEndDeg != 0 || c.ArcThicknessRatio != 0) {
			return fmt.Errorf("canvas %s shape fields require kind shape", c.ID)
		}
		switch c.Kind {
		case "text":
			rich := len(c.Paragraphs) > 0
			if (!rich && (!validID(c.Text) || !ValidFontFace(c.FontFace) || !positive(c.FontSizePt))) || (rich && (c.Text != "" || c.FontFace != "" || c.FontSizePt != 0 || c.Bold || c.Foreground != "")) || !nonnegative(c.InsetX) || !nonnegative(c.InsetY) || c.Bounds.Width <= 2*c.InsetX || c.Bounds.Height <= 2*c.InsetY {
				return fmt.Errorf("canvas text %s requires explicit font family typography, content and usable frame", c.ID)
			}
			if c.Align != "left" && c.Align != "center" && c.Align != "right" {
				return fmt.Errorf("canvas text %s requires explicit left/center/right alignment", c.ID)
			}
			if c.Valign != "top" && c.Valign != "middle" {
				return fmt.Errorf("canvas text %s requires top/middle valign", c.ID)
			}
			bg := c.Background
			if bg == "" {
				bg = c.ContrastBackground
			}
			if bg == "" {
				bg = white
			}
			if rich {
				if e := validateRichText(c.Paragraphs, bg); e != nil {
					return fmt.Errorf("canvas %s: %w", c.ID, e)
				}
				if e := validateRichBulletWidth(c.Paragraphs, c.Bounds.Width-2*c.InsetX); e != nil {
					return fmt.Errorf("canvas %s: %w", c.ID, e)
				}
			} else if _, e := foregroundAtSize(c.Foreground, bg, c.FontSizePt, c.Bold); e != nil {
				return fmt.Errorf("canvas %s: %w", c.ID, e)
			}
		case "surface":
			if _, e := resolveColor(c.Background); e != nil {
				return e
			}
		case "shape":
			if e := validateShape(c); e != nil {
				return e
			}
		case "line":
			if _, e := resolveColor(c.Foreground); e != nil {
				return e
			}
		case "image":
			if c.AssetPath == "" || len(c.AssetSHA256) != 64 || !validID(c.AltText) {
				return fmt.Errorf("canvas image %s requires local asset_path, SHA256 and alt_text", c.ID)
			}
			if strings.Contains(c.AssetPath, "://") {
				return fmt.Errorf("canvas image %s requires a local pinned file", c.ID)
			}
			isSVG := strings.HasSuffix(strings.ToLower(c.AssetPath), ".svg")
			if isSVG != (c.FallbackAssetPath != "" || c.FallbackAssetSHA256 != "") {
				return fmt.Errorf("canvas image %s requires a pinned PNG fallback exactly for SVG", c.ID)
			}
			if isSVG && (len(c.FallbackAssetSHA256) != 64 || strings.Contains(c.FallbackAssetPath, "://") || !strings.HasSuffix(strings.ToLower(c.FallbackAssetPath), ".png")) {
				return fmt.Errorf("canvas image %s requires local PNG fallback path and SHA256", c.ID)
			}
			if (c.OutlineColor == "") != (c.OutlineWidthPt == 0) || c.OutlineWidthPt < 0 || c.OutlineWidthPt > 6 {
				return fmt.Errorf("canvas image %s outline requires color and width in (0,6]", c.ID)
			}
			if c.OutlineColor != "" {
				if _, err := resolveColor(c.OutlineColor); err != nil {
					return fmt.Errorf("canvas image %s outline: %w", c.ID, err)
				}
			}
			if err := validateImagePlacement(c); err != nil {
				return err
			}
		default:
			return fmt.Errorf("canvas %s unknown kind %q", c.ID, c.Kind)
		}
	}
	for _, c := range s.Canvas {
		for _, a := range c.AllowOverlap {
			if a != "slide-title" && !ids[a] {
				return fmt.Errorf("canvas %s unknown overlap peer %s", c.ID, a)
			}
			if a == c.ID {
				return fmt.Errorf("canvas %s self overlap declaration", c.ID)
			}
		}
	}
	return nil
}
func canvasProbes(s SlideSpec) []ProbeRequest {
	var q []ProbeRequest
	for _, c := range s.Canvas {
		if c.Kind == "text" {
			bg := c.Background
			if bg == "" {
				bg = c.ContrastBackground
			}
			if bg == "" {
				bg = white
			}
			fg := ""
			paragraphs := c.Paragraphs
			text := c.Text
			if len(paragraphs) > 0 {
				paragraphs = resolveRichText(paragraphs, bg)
				text = richText(paragraphs)
			} else {
				fg, _ = foregroundAtSize(c.Foreground, bg, c.FontSizePt, c.Bold)
			}
			q = append(q, ProbeRequest{ID: requestID(s.ID, "canvas", c.ID), SlideID: s.ID, Kind: "canvas_text", PhraseRequests: accentPhraseRequests(s, c.ID), Text: text, Paragraphs: paragraphs, TextWidthPt: c.Bounds.Width - 2*c.InsetX, FontFace: c.FontFace, FontSizePt: c.FontSizePt, Bold: c.Bold, Foreground: fg, Background: bg, Align: c.Align})
		}
	}
	for _, a := range s.Accents {
		if a.Staging != nil {
			q = append(q, accentNoteProbe(s, a))
		}
	}
	for _, a := range s.ArtworkArrows {
		if a.Staging != nil {
			q = append(q, arrowNoteProbe(s, a))
		}
	}
	return q
}
func planCanvas(s SlideSpec, p *PlannedSlide, m Measurements) error {
	for _, c := range s.Canvas {
		pc := PlannedCanvas{CanvasSpec: c, PhraseRequests: accentPhraseRequests(s, c.ID)}
		if len(c.Adjustments) != 0 {
			pc.Adjustments = make(map[string]int, len(c.Adjustments))
			for name, value := range c.Adjustments {
				pc.Adjustments[name] = value
			}
		}
		if c.Background != "" {
			pc.Background, _ = resolveColor(c.Background)
		}
		pc.Pattern = resolvePattern(c.Pattern)
		if c.ContrastBackground != "" {
			pc.ContrastBackground, _ = resolveColor(c.ContrastBackground)
		}
		if c.Foreground != "" {
			if c.Kind == "text" {
				bg := c.Background
				if bg == "" {
					bg = c.ContrastBackground
				}
				if bg == "" {
					bg = white
				}
				pc.Foreground, _ = foregroundAtSize(c.Foreground, bg, c.FontSizePt, c.Bold)
			} else {
				pc.Foreground, _ = resolveColor(c.Foreground)
			}
		}
		if len(c.Paragraphs) > 0 {
			bg := c.Background
			if bg == "" {
				bg = c.ContrastBackground
			}
			if bg == "" {
				bg = white
			}
			pc.Paragraphs = resolveRichText(c.Paragraphs, bg)
			pc.Text = richText(pc.Paragraphs)
		}
		if c.Kind == "text" {
			pc.MeasurementID = requestID(s.ID, "canvas", c.ID)
			z := m.ByRequestID[pc.MeasurementID]
			if z.RenderedWidthPt > c.Bounds.Width-2*c.InsetX+0.01 || z.RenderedHeightPt > c.Bounds.Height-2*c.InsetY+0.01 {
				if len(c.Paragraphs) > 0 {
					return fmt.Errorf("canvas text %s does not fit %.2f x %.2fpt with configured rich text styles", c.ID, c.Bounds.Width-2*c.InsetX, c.Bounds.Height-2*c.InsetY)
				}
				return fmt.Errorf("canvas text %s does not fit %.2f x %.2fpt at %.2fpt", c.ID, c.Bounds.Width-2*c.InsetX, c.Bounds.Height-2*c.InsetY, c.FontSizePt)
			}
		}
		p.Canvas = append(p.Canvas, pc)
	}
	type box struct {
		id string
		r  Rect
	}
	var boxes []box
	if p.TitleMeasurementID != "" {
		boxes = append(boxes, box{"slide-title", p.TitleBounds})
	}
	for _, q := range p.Pods {
		boxes = append(boxes, box{q.ID, q.Bounds})
	}
	for _, q := range p.Roles {
		boxes = append(boxes, box{q.ID, q.Bounds})
	}
	for _, q := range p.Phases {
		boxes = append(boxes, box{q.ID, q.Bounds})
	}
	for _, q := range p.Cards {
		boxes = append(boxes, box{q.ID, q.Bounds})
	}
	if p.Legend != nil {
		boxes = append(boxes, box{"staffing-legend", p.Legend.Bounds})
	}
	allowed := func(a CanvasSpec, id string) bool {
		for _, b := range a.AllowOverlap {
			if b == id {
				return true
			}
		}
		return false
	}
	for i, c := range p.Canvas {
		// Stroke bounds include ink, including zero-extent straight lines.
		r := c.Bounds
		if c.Kind == "line" {
			r = expand(r, c.LineWidthPt/2)
			if !inside(r, Rect{Width: s.WidthPt, Height: s.HeightPt}) {
				return fmt.Errorf("canvas line %s ink exceeds slide", c.ID)
			}
		}
		for _, b := range boxes {
			if overlap(r, b.r) && !allowed(c.CanvasSpec, b.id) {
				return fmt.Errorf("canvas %s overlaps %s without a declared relationship", c.ID, b.id)
			}
		}
		for _, b := range p.Canvas[:i] {
			br := b.Bounds
			if b.Kind == "line" {
				br = expand(br, b.LineWidthPt/2)
			}
			if overlap(r, br) && !allowed(c.CanvasSpec, b.ID) && !allowed(b.CanvasSpec, c.ID) {
				return fmt.Errorf("canvas %s overlaps %s without a declared relationship", c.ID, b.ID)
			}
		}
	}
	return nil
}
