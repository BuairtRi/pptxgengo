package main

import (
	"fmt"
	"github.com/buairtri/pptxgengo/pptx"
)

type frame struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}
type element struct {
	Name          string   `json:"name"`
	Kind          string   `json:"kind"`
	Frame         frame    `json:"frame"`
	Text          string   `json:"text,omitempty"`
	FontFace      string   `json:"font_face,omitempty"`
	FontSize      float64  `json:"font_size,omitempty"`
	Bold          bool     `json:"bold,omitempty"`
	Foreground    string   `json:"foreground,omitempty"`
	Background    string   `json:"background,omitempty"`
	InsetX        float64  `json:"inset_x"`
	InsetY        float64  `json:"inset_y"`
	Align         string   `json:"align,omitempty"`
	Valign        string   `json:"valign,omitempty"`
	MeasurementID string   `json:"measurement_id,omitempty"`
	LineWidth     float64  `json:"line_width,omitempty"`
	ConnectionIDs []string `json:"connection_ids,omitempty"`
}
type renderSlide struct {
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
	p.Title = "Editable role and pod compositions"
	p.Subject = "Illustrative component fixtures; no staffing commitments"
	p.Theme = pptx.ThemeProps{HeadFontFace: "Arial", BodyFontFace: "Arial"}
	for _, rs := range slides {
		if rs.Width != slides[0].Width || rs.Height != slides[0].Height {
			return nil, fmt.Errorf("mixed slide sizes unsupported")
		}
		s := p.AddSlide()
		s.Background(&pptx.BackgroundProps{Color: "FFFFFF"})
		s.AddNotes("Illustrative roles for component development. Semantic IDs, measurements and fit status are recorded in the adjacent composition report. These roles are not a client staffing proposal.")
		for _, e := range rs.Elements {
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
			if err := s.AddText([]pptx.TextProps{{Text: e.Text}}, o); err != nil {
				return nil, err
			}
		}
	}
	return p.Write()
}
