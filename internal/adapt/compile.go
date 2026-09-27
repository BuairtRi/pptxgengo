package adapt

import (
	"fmt"
	"math"
	"strings"

	"github.com/buairtri/pptxgengo/internal/compose"
)

func Compile(input Spec) (compose.Spec, Report, error) {
	out := compose.Spec{Schema: compose.SpecSchema}
	report := Report{Schema: "pptxgengo.adaptive-compilation.v1"}
	if input.Schema != Schema || len(input.Slides) == 0 {
		return out, report, fmt.Errorf("expected %s with at least one slide", Schema)
	}
	ids := map[string]bool{}
	for _, slide := range input.Slides {
		if strings.TrimSpace(slide.ID) == "" || ids[slide.ID] || strings.Contains(slide.ID, "/") {
			return out, report, fmt.Errorf("slides require unique nonempty IDs without slashes")
		}
		ids[slide.ID] = true
		if strings.TrimSpace(slide.Title) == "" || strings.TrimSpace(slide.Role) == "" || strings.TrimSpace(slide.Takeaway) == "" {
			return out, report, fmt.Errorf("slide %s requires title, role and takeaway", slide.ID)
		}
		style, err := ResolveStyle(slide.Style)
		if err != nil {
			return out, report, fmt.Errorf("slide %s: %w", slide.ID, err)
		}
		bounds := compose.Rect{X: 48, Y: 125, Width: 864, Height: 360}
		if slide.Bounds != nil {
			bounds = *slide.Bounds
		}
		for _, v := range []float64{bounds.X, bounds.Y, bounds.Width, bounds.Height} {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return out, report, fmt.Errorf("slide %s nonfinite content bounds", slide.ID)
			}
		}
		if bounds.X < 40 || bounds.Y < 115 || bounds.Width <= 0 || bounds.Height <= 0 || bounds.X+bounds.Width > 920 || bounds.Y+bounds.Height > 485 {
			return out, report, fmt.Errorf("slide %s content bounds must stay within x40..920,y115..485", slide.ID)
		}
		c := Context{Slide: slide, Bounds: bounds, Style: style}
		var content Content
		switch slide.Family {
		case "process":
			content, err = buildProcess(c, slide.Content)
		case "roadmap":
			content, err = buildRoadmap(c, slide.Content)
		case "comparison":
			content, err = buildComparison(c, slide.Content)
		case "architecture":
			content, err = buildArchitecture(c, slide.Content)
		case "team":
			content, err = buildTeam(c, slide.Content)
		default:
			err = fmt.Errorf("unsupported adaptive family %q", slide.Family)
		}
		if err != nil {
			return out, report, fmt.Errorf("slide %s: %w", slide.ID, err)
		}
		s := compose.SlideSpec{ID: slide.ID, Title: slide.Title, Role: slide.Role, Takeaway: slide.Takeaway, Notes: "Adaptive family: " + slide.Family + ". " + slide.Role + " Takeaway: " + slide.Takeaway, WidthPt: 960, HeightPt: 540, TitleBounds: compose.Rect{X: 48, Y: 32, Width: 864, Height: 66}, TitleFontFace: style.FontFace, TitleFontSizePt: style.TitleFontPt, TitleBold: true, TitleForeground: style.Navy, Canvas: content.Canvas, Layouts: content.Layouts, Pods: content.Pods, Roles: content.Roles, Connections: content.Connections, Paths: content.Paths, Legend: content.Legend}
		s.Canvas = append(s.Canvas, compose.CanvasSpec{ID: "chrome/title-rule", Kind: "line", Bounds: compose.Rect{X: 48, Y: 105, Width: 64, Height: 0}, Foreground: style.Accent, LineWidthPt: 3}, compose.CanvasSpec{ID: "chrome/logo", Kind: "image", Bounds: compose.Rect{X: 48, Y: 500, Width: 108, Height: 108 * 300.0 / 1436.0}, AssetPath: "samples/showcase/assets/wm_h_pos_clr_rgb_august2024.png", AssetSHA256: "47a4eaa1be957777d3e106f5c1cc9f67dbfcda79d74fe7a7f2bcbb88b13fa341", AltText: "West Monroe logo", ImageFit: "preserve"}, compose.CanvasSpec{ID: "chrome/footer", Kind: "text", Bounds: compose.Rect{X: 185, Y: 502, Width: 670, Height: 18}, Text: slide.Footer, FontFace: "Arial", FontSizePt: 9, Foreground: style.Secondary, ContrastBackground: style.White, Align: "right", Valign: "middle"}, compose.CanvasSpec{ID: "chrome/page", Kind: "text", Bounds: compose.Rect{X: 875, Y: 502, Width: 37, Height: 18}, Text: fmt.Sprint(len(out.Slides) + 1), FontFace: "Arial", FontSizePt: 9, Foreground: style.Navy, Align: "right", Valign: "middle"})
		// Footer is on the fixed white chrome even when body components use inverse styling.
		s.Canvas[len(s.Canvas)-2].Foreground = "#50658E"
		if strings.TrimSpace(slide.Footer) == "" {
			s.Canvas = append(s.Canvas[:len(s.Canvas)-2], s.Canvas[len(s.Canvas)-1])
		}
		out.Slides = append(out.Slides, s)
		report.Slides = append(report.Slides, SlideReport{ID: slide.ID, Family: slide.Family, TemplateID: slide.TemplateID, Controls: content.Controls, Limitations: append(content.Limitations, "Text capacity requires native measurement and final visual review; source template pixel identity is not claimed."), Qualification: "compiled_pending_native_measurement"})
	}
	if _, err := compose.ProbeRequests(out); err != nil {
		return out, report, fmt.Errorf("lowered composition invalid: %w", err)
	}
	return out, report, nil
}
