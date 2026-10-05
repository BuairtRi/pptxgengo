package nativeexport

import "github.com/buairtri/pptxgengo/pptx"

// The operational probe is self-contained. Explicit theme and shape geometry
// avoid relying on the writer's unset text defaults, which PowerPoint may reject.
func doctorPresentation() ([]byte, error) {
	p := pptx.New()
	if err := p.SetLayout("LAYOUT_WIDE"); err != nil {
		return nil, err
	}
	p.Author = "pptxgengo"
	p.Company = "pptxgengo"
	p.Title = "Native rendering file-access diagnostic"
	p.Subject = "Temporary one-slide PowerPoint open and PDF-write probe"
	p.Theme = pptx.ThemeProps{HeadFontFace: "Arial", BodyFontFace: "Arial"}
	slide := p.AddSlide()
	slide.PresSlide().SlideBaseProps.Background = &pptx.BackgroundProps{ShapeFillProps: pptx.ShapeFillProps{Color: "FFFFFF"}}
	x, y, w, h := pptx.Inches(1), pptx.Inches(1), pptx.Inches(10), pptx.Inches(1)
	options := &pptx.TextPropsOptions{PositionProps: pptx.PositionProps{X: &x, Y: &y, W: &w, H: &h}, TextBaseProps: pptx.TextBaseProps{FontFace: "Arial", FontSize: 24, Color: "000000", Lang: "en-US", Align: "left", Valign: "top"}, ObjectNameProps: pptx.ObjectNameProps{ObjectName: "native-file-access-diagnostic"}, Fit: "none", Margin: pptx.Margin{0}, LineSpacing: 28}
	if err := slide.AddText([]pptx.TextProps{{Text: "PowerPoint staging file-access diagnostic"}}, options); err != nil {
		return nil, err
	}
	return p.Write()
}
