package wmdesign

import (
	"fmt"
	"math"
	"strings"

	"github.com/buairtri/pptxgengo/pptx"
)

// drawLibraryWhiteboards retains the frozen source dot lattice and paint-order
// switch. Opacity samples the source CSS mask at each editable square's center;
// native square opacity does not claim pixel identity with a browser mask.
func (r *renderer) drawLibraryWhiteboards(chrome *LibraryChrome, f ResolvedFrame, above bool) error {
	if chrome == nil || !chrome.CustomWhiteboard {
		return nil
	}
	for index, field := range chrome.Whiteboard {
		if field.Above != above {
			continue
		}
		if field.W <= 0 || field.H <= 0 || math.IsNaN(field.X+field.Y+field.W+field.H) || math.IsInf(field.X+field.Y+field.W+field.H, 0) {
			return fmt.Errorf("library.whiteboard_geometry: field %d", index)
		}
		fade := field.Fade
		if fade == "" {
			fade = "corner"
		}
		switch fade {
		case "corner", "right", "left", "bottom", "none":
		default:
			return fmt.Errorf("library.whiteboard_fade: %s", fade)
		}
		surface := field.On
		if surface == "" {
			surface = f.Request.Surface
			if f.Request.Rail == "left" && field.X < 273 || f.Request.Rail == "right" && field.X >= 687 {
				surface = f.Request.RailSurface
			}
		}
		if _, err := r.sceneColor(surface, "bg"); err != nil {
			return err
		}
		color, err := r.sceneColor("light", "line")
		if err != nil {
			return err
		}
		if surface == "inverse" || surface == "deep" {
			color, err = r.sceneColor("inverse", "line")
			if err != nil {
				return err
			}
		}
		cols, rows := int(math.Floor(field.W/18))+1, int(math.Floor(field.H/18))+1
		width, height := float64(cols-1)*18+3, float64(rows-1)*18+3
		if field.X < 0 || field.Y < 0 || field.X+width > 960+.02 || field.Y+height > 540+.02 {
			return fmt.Errorf("library.whiteboard_outside_canvas: field %d", index)
		}
		for col := 0; col < cols; col++ {
			for row := 0; row < rows; row++ {
				x, y := float64(col)*18, float64(row)*18
				u, v := (x+1.5)/width, (y+1.5)/height
				opacity := 1.0
				switch fade {
				case "corner":
					opacity = (1 - math.Hypot(u, v)) / .65
				case "right":
					opacity = (1 - u) / .6
				case "left":
					opacity = u / .6
				case "bottom":
					opacity = (1 - v) / .6
				}
				opacity = math.Min(1, math.Max(0, opacity))
				if opacity == 0 {
					continue
				}
				r.shape(fmt.Sprintf("wm.library-whiteboard.%03d.%03d.%03d", index+1, col+1, row+1), Rect{field.X + x, field.Y + y, 3, 3}, color, 100*(1-opacity))
				if r.err != nil {
					return r.err
				}
			}
		}
	}
	return nil
}

// drawLibraryStamp creates the source's right-aligned, padded native stamp.
func (r *renderer) drawLibraryStamp(chrome *LibraryChrome, f ResolvedFrame, sr *SlideReport) error {
	if chrome == nil || chrome.Stamp == "" {
		return nil
	}
	text := strings.ToUpper(chrome.Stamp)
	if strings.ContainsAny(text, "\r\n") {
		return fmt.Errorf("library.stamp_requires_single_line")
	}
	style, err := r.sceneStyle("label")
	if err != nil {
		return err
	}
	style.Size = 8
	style.Weight = 600
	style.Leading = 12
	style.Tracking = "0.08em"
	style.TrackingPt = .64
	style.Case = ""
	need, width, err := r.sequenceNeed(text, style, 8191)
	if err != nil {
		return err
	}
	if need > 18+.02 {
		return fmt.Errorf("library.stamp_height: requires %.3fpt capacity 18pt", need)
	}
	// Native PowerPoint includes a terminal tracking allowance when fitting a
	// line. Reserve 2pt beyond the measured glyph advance so the final letter
	// stays inside the stamp; the wrapper retains its source right edge.
	width += 2
	if width+12 > f.Body.W+.02 {
		return fmt.Errorf("library.stamp_width: requires %.3fpt capacity %.3fpt", width+12, f.Body.W)
	}
	box := Rect{f.Body.X + f.Body.W - width - 12, 33, width + 12, 18}
	p := &scenePlan{ID: "wm.library-stamp", Bounds: box}
	outline, err := r.sceneColor(f.Request.Surface, "strong")
	if err != nil {
		return err
	}
	if err = r.diagramShape(p, p.ID+".outline", box, pptx.ShapeTypeRect, "", outline, .75, "solid", nil); err != nil {
		return err
	}
	if err = r.diagramStyledText(p, p.ID+".text", text, style, Rect{box.X + 6, box.Y, width, box.H}, f.Request.Surface, "primary", "left", true); err != nil {
		return err
	}
	return r.drawScene(diagramFinish(p, "scene.library-stamp"), sr, "/stamp")
}
