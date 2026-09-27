package main

import (
	"encoding/base64"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/compose"
	"github.com/buairtri/pptxgengo/pptx"
)

type frame struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}
type element struct {
	AssetMode           string                  `json:"asset_mode,omitempty"`
	AssetPath           string                  `json:"asset_path,omitempty"`
	AssetSHA256         string                  `json:"asset_sha256,omitempty"`
	FallbackAssetPath   string                  `json:"fallback_asset_path,omitempty"`
	FallbackAssetSHA256 string                  `json:"fallback_asset_sha256,omitempty"`
	AltText             string                  `json:"alt_text,omitempty"`
	OutlineColor        string                  `json:"outline_color,omitempty"`
	OutlineWidthPt      float64                 `json:"outline_width_pt,omitempty"`
	ImageFit            string                  `json:"image_fit,omitempty"`
	ImageCrop           *compose.ImageCropSpec  `json:"image_crop,omitempty"`
	FocalX              *float64                `json:"focal_x,omitempty"`
	FocalY              *float64                `json:"focal_y,omitempty"`
	Preset              string                  `json:"preset,omitempty"`
	Adjustments         map[string]int          `json:"adjustments,omitempty"`
	Pattern             *compose.PatternSpec    `json:"pattern,omitempty"`
	Name                string                  `json:"name"`
	Kind                string                  `json:"kind"`
	Frame               frame                   `json:"frame"`
	Text                string                  `json:"text,omitempty"`
	Paragraphs          []compose.ParagraphSpec `json:"paragraphs,omitempty"`
	FontFace            string                  `json:"font_face,omitempty"`
	FontSize            float64                 `json:"font_size,omitempty"`
	Bold                bool                    `json:"bold,omitempty"`
	Foreground          string                  `json:"foreground,omitempty"`
	Background          string                  `json:"background,omitempty"`
	InsetX              float64                 `json:"inset_x"`
	InsetY              float64                 `json:"inset_y"`
	Align               string                  `json:"align,omitempty"`
	Valign              string                  `json:"valign,omitempty"`
	MeasurementID       string                  `json:"measurement_id,omitempty"`
	LineWidth           float64                 `json:"line_width,omitempty"`
	ConnectionIDs       []string                `json:"connection_ids,omitempty"`
}
type renderSlide struct {
	Notes    string    `json:"notes,omitempty"`
	ID       string    `json:"id"`
	Width    float64   `json:"width"`
	Height   float64   `json:"height"`
	Elements []element `json:"elements"`
}

func pointer[T any](v T) *T { return &v }
func pos(f frame) pptx.PositionProps {
	return pptx.PositionProps{X: pointer(pptx.Inches(f.X / 72)), Y: pointer(pptx.Inches(f.Y / 72)), W: pointer(pptx.Inches(f.Width / 72)), H: pointer(pptx.Inches(f.Height / 72))}
}

// Native rectangles and text stay editable. All geometry is in points at this
// boundary. Probe and output text use identical paragraph/font settings.
func render(slides []renderSlide) ([]byte, error) {
	if len(slides) == 0 {
		return nil, fmt.Errorf("empty render")
	}
	p := pptx.New()
	p.DefineLayout("COMPOSE", slides[0].Width/72, slides[0].Height/72)
	if err := p.SetLayout("COMPOSE"); err != nil {
		return nil, err
	}
	p.Author = "pptxgengo"
	p.Company = "West Monroe"
	p.Title = "Editable PowerPoint compositions"
	p.Subject = "Illustrative proposal and component examples"
	p.Theme = pptx.ThemeProps{HeadFontFace: "Arial", BodyFontFace: "Arial"}
	for _, rs := range slides {
		if rs.Width != slides[0].Width || rs.Height != slides[0].Height {
			return nil, fmt.Errorf("mixed slide sizes unsupported")
		}
		s := p.AddSlide()
		s.PresSlide().Name = "compose:" + base64.RawURLEncoding.EncodeToString([]byte(rs.ID))
		s.Background(&pptx.BackgroundProps{Color: "FFFFFF"})
		if rs.Notes != "" {
			s.AddNotes(rs.Notes)
		}
		for _, e := range rs.Elements {
			if e.Kind == "image" {
				if err := renderImage(s, e); err != nil {
					return nil, err
				}
				continue
			}
			if e.Kind == "line" {
				if err := s.AddShape(pptx.ShapeTypeLine, &pptx.ShapeProps{PositionProps: pos(e.Frame), ObjectNameProps: pptx.ObjectNameProps{ObjectName: e.Name}, Fill: &pptx.ShapeFillProps{Type: "none"}, Line: &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: e.Foreground}, Width: e.LineWidth, BeginArrowType: "none", EndArrowType: "none"}}); err != nil {
					return nil, err
				}
				continue
			}

			if e.Kind == "surface" {
				if err := s.AddShape(pptx.ShapeTypeRect, &pptx.ShapeProps{PositionProps: pos(e.Frame), ObjectNameProps: pptx.ObjectNameProps{ObjectName: e.Name}, Fill: &pptx.ShapeFillProps{Color: e.Background}, Line: &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Type: "none"}}}); err != nil {
					return nil, err
				}
				continue
			}
			if e.Kind == "shape" {
				fill := &pptx.ShapeFillProps{Color: e.Background}
				if e.Pattern != nil {
					fill = &pptx.ShapeFillProps{Pattern: &pptx.ShapePatternFillProps{Preset: pptx.PatternType(e.Pattern.Preset), Foreground: color(e.Pattern.Foreground), Background: color(e.Pattern.Background)}}
				}
				if err := s.AddShape(pptx.ShapeType(e.Preset), &pptx.ShapeProps{PositionProps: pos(e.Frame), ObjectNameProps: pptx.ObjectNameProps{ObjectName: e.Name}, Adjustments: e.Adjustments, Fill: fill, Line: &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Type: "none"}}}); err != nil {
					return nil, err
				}
				continue
			}
			if e.Kind != "text" {
				return nil, fmt.Errorf("unsupported render element %s", e.Kind)
			}
			align, valign := e.Align, e.Valign
			if align == "" {
				align = "center"
			}
			if valign == "" {
				valign = "top"
			}
			o := &pptx.TextPropsOptions{PositionProps: pos(e.Frame), ObjectNameProps: pptx.ObjectNameProps{ObjectName: e.Name}, TextBaseProps: pptx.TextBaseProps{FontFace: e.FontFace, FontSize: e.FontSize, Bold: pointer(e.Bold), Color: e.Foreground, Align: align}, Valign: valign, Fit: "none", Wrap: pointer(true), ParaSpaceBefore: pointer(0.0), ParaSpaceAfter: pointer(0.0), LineSpacingMultiple: 1}
			// The current text writer maps array margins as left,right,bottom,top.
			// Emit explicit zeros as well, so the native renderer adds no defaults.
			o.Margin = pptx.Margin{e.InsetX, e.InsetX, e.InsetY, e.InsetY}
			o.Line = &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Type: "none"}}
			if e.Background != "" {
				o.Fill = &pptx.ShapeFillProps{Color: e.Background}
				o.Shape = pptx.ShapeTypeRect
			}
			var textRuns []pptx.TextProps
			if len(e.Paragraphs) == 0 {
				textRuns = []pptx.TextProps{{Text: e.Text}}
			} else {
				for pi, paragraph := range e.Paragraphs {
					for ri, run := range paragraph.Runs {
						breakLine := pi < len(e.Paragraphs)-1 && ri == len(paragraph.Runs)-1
						lineSpacing := paragraph.LineSpacingMultiple
						if lineSpacing == 0 {
							lineSpacing = 1
						}
						runOpts := &pptx.TextPropsOptions{TextBaseProps: pptx.TextBaseProps{FontFace: run.FontFace, FontSize: run.FontSizePt, Bold: pointer(run.Bold), Italic: pointer(run.Italic), Color: color(run.Foreground), Align: pptx.HAlign(paragraph.Align), BreakLine: pointer(breakLine)}, ParagraphContinuation: ri > 0, ParaSpaceBefore: pointer(paragraph.SpaceBeforePt), ParaSpaceAfter: pointer(paragraph.SpaceAfterPt), LineSpacingMultiple: lineSpacing}
						if ri == 0 && paragraph.Bullet != nil {
							runes := []rune(paragraph.Bullet.Character)
							runOpts.Bullet = &pptx.BulletProps{CharacterCode: fmt.Sprintf("%04X", runes[0]), MarginLeftPt: paragraph.Bullet.MarginLeftPt, HangingPt: paragraph.Bullet.HangingPt}
						}
						if run.Underline {
							runOpts.Underline = &pptx.UnderlineProps{Style: "sng"}
						}
						textRuns = append(textRuns, pptx.TextProps{Text: run.Text, Options: runOpts})
					}
				}
			}
			if err := s.AddText(textRuns, o); err != nil {
				return nil, err
			}
		}
	}
	return p.Write()
}
