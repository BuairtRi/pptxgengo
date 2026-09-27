package main

import (
	"encoding/base64"
	"fmt"
	"math"
	"sort"

	"github.com/buairtri/pptxgengo/internal/compose"
)

// Merge overlapping collinear strokes. Shared reporting trunks are rendered
// once, while the manifest retains every relationship contributing to a stroke.
type routeSegment struct {
	vertical             bool
	axis, lo, hi, width  float64
	color                string
	dash                 string
	beginArrow, endArrow string
	relationship         string
	ids                  []string
}

func connectorElements(cs []compose.PlannedConnection) []element {
	var segs []routeSegment
	for _, c := range cs {
		if c.Strategy == "direct_right_arrow" {
			continue
		}
		for i := 1; i < len(c.Points); i++ {
			a, b := c.Points[i-1], c.Points[i]
			v := math.Abs(a.X-b.X) < 1e-7
			sg := routeSegment{vertical: v, width: c.WidthPt, color: color(c.Color), dash: c.LineDash, relationship: c.Relationship, ids: []string{c.ID}}
			if i == len(c.Points)-1 && c.EndArrow == "triangle" {
				if (!v && b.X > a.X) || (v && b.Y > a.Y) {
					sg.endArrow = "triangle"
				} else {
					sg.beginArrow = "triangle"
				}
			}
			if v {
				sg.axis = a.X
				sg.lo = math.Min(a.Y, b.Y)
				sg.hi = math.Max(a.Y, b.Y)
			} else {
				sg.axis = a.Y
				sg.lo = math.Min(a.X, b.X)
				sg.hi = math.Max(a.X, b.X)
			}
			if sg.hi-sg.lo > 1e-7 {
				segs = append(segs, sg)
			}
		}
	}
	sort.SliceStable(segs, func(i, j int) bool {
		a, b := segs[i], segs[j]
		if a.vertical != b.vertical {
			return !a.vertical
		}
		if a.axis != b.axis {
			return a.axis < b.axis
		}
		if a.color != b.color {
			return a.color < b.color
		}
		if a.width != b.width {
			return a.width < b.width
		}
		if a.relationship != b.relationship {
			return a.relationship < b.relationship
		}
		if a.dash != b.dash {
			return a.dash < b.dash
		}
		if a.beginArrow != b.beginArrow {
			return a.beginArrow < b.beginArrow
		}
		if a.endArrow != b.endArrow {
			return a.endArrow < b.endArrow
		}
		return a.lo < b.lo
	})
	var merged []routeSegment
	for _, s := range segs {
		if len(merged) > 0 {
			m := &merged[len(merged)-1]
			if m.relationship == "reporting" && s.relationship == "reporting" && m.beginArrow == "" && m.endArrow == "" && s.beginArrow == "" && s.endArrow == "" && m.vertical == s.vertical && math.Abs(m.axis-s.axis) < 1e-7 && m.color == s.color && m.width == s.width && m.dash == s.dash && s.lo <= m.hi+1e-7 {
				m.hi = math.Max(m.hi, s.hi)
				for _, id := range s.ids {
					found := false
					for _, old := range m.ids {
						if old == id {
							found = true
						}
					}
					if !found {
						m.ids = append(m.ids, id)
					}
				}
				continue
			}
		}
		merged = append(merged, s)
	}
	var result []element
	for i, s := range merged {
		sort.Strings(s.ids)
		f := frame{X: s.lo, Y: s.axis, Width: s.hi - s.lo, Height: 0}
		if s.vertical {
			f = frame{X: s.axis, Y: s.lo, Width: 0, Height: s.hi - s.lo}
		}
		result = append(result, element{Name: fmt.Sprintf("connection-segment-%03d", i+1), Kind: "line", Frame: f, Foreground: s.color, LineWidth: s.width, LineDash: s.dash, BeginArrow: s.beginArrow, EndArrow: s.endArrow, ConnectionIDs: s.ids})
	}
	return result
}
func roleElement(r compose.PlannedRole) element {
	return element{Name: "role:" + base64.RawURLEncoding.EncodeToString([]byte(r.ID)), Kind: "text", Frame: rect(r.Bounds), Text: r.Label, FontFace: r.FontFace, FontSize: r.FontSizePt, Bold: r.Bold, Foreground: color(r.Foreground), Background: color(r.Background), InsetX: r.HorizontalInsetPt, InsetY: r.VerticalInsetPt, Valign: "middle", MeasurementID: r.MeasurementID}
}
