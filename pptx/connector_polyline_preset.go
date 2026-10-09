package pptx

import (
	"fmt"
	"math"
)

// NativePolylineConnectorPreset lowers alternating orthogonal routes to native
// elbow presets, preserving endpoint attachments and editable bend handles.
// Points are normalized to the supplied positive allocation. Extra bends,
// diagonal segments and zero endpoint extents have no equivalent in this
// finite-adjustment contract and are refused rather than silently flattened.
// Literal guides use DrawingML's 100000-based adjustment precision.
func NativePolylineConnectorPreset(shape ShapeProps, points [][2]float64) (ShapeType, ShapeProps, error) {
	fail := func(reason string) (ShapeType, ShapeProps, error) {
		return "", ShapeProps{}, fmt.Errorf("native polyline preset: %s", reason)
	}
	if len(points) < 3 || len(points) > 6 {
		return fail("requires 3..6 alternating orthogonal vertices")
	}
	if shape.W == nil || shape.H == nil || shape.W.IsPct || shape.H.IsPct || shape.W.Val <= 0 || shape.H.Val <= 0 || !validConnectorBounds(shape.PositionProps, false) {
		return fail("requires finite positive absolute allocation")
	}
	if shape.Rotate != 0 || boolDeref(shape.FlipH) || boolDeref(shape.FlipV) || len(shape.Points) != 0 || len(shape.Adjustments) != 0 {
		return fail("requires untransformed allocation without existing geometry")
	}
	for _, c := range []*Coord{shape.X, shape.Y} {
		if c != nil && c.IsPct {
			return fail("percentage origins are unsupported")
		}
	}
	horizontal := false
	for i, p := range points {
		if math.IsNaN(p[0]) || math.IsNaN(p[1]) || math.IsInf(p[0], 0) || math.IsInf(p[1], 0) || p[0] < 0 || p[0] > 1 || p[1] < 0 || p[1] > 1 {
			return fail("points require finite normalized coordinates")
		}
		if i == 0 {
			continue
		}
		prior := points[i-1]
		isHorizontal := prior[1] == p[1] && prior[0] != p[0]
		isVertical := prior[0] == p[0] && prior[1] != p[1]
		if !isHorizontal && !isVertical {
			return fail("diagonal or duplicate segment is unsupported")
		}
		if i == 1 {
			horizontal = isHorizontal
		}
		expected := horizontal == ((i-1)%2 == 0)
		if isHorizontal != expected {
			return fail("collinear consecutive segments require explicit simplification")
		}
	}
	a, z := points[0], points[len(points)-1]
	dx, dy := z[0]-a[0], z[1]-a[1]
	if dx == 0 || dy == 0 {
		return fail("zero endpoint extent cannot retain an orthogonal detour")
	}
	x, y := 0.0, 0.0
	if shape.X != nil {
		x = shape.X.Val
	}
	if shape.Y != nil {
		y = shape.Y.Val
	}
	w, h := shape.W.Val, shape.H.Val
	ax, ay, zx, zy := x+a[0]*w, y+a[1]*h, x+z[0]*w, y+z[1]*h
	out := shape
	out.Points = nil
	out.Adjustments = map[string]int{}
	out.Rotate = 0
	if horizontal {
		out.X, out.Y = ptr(Inches(math.Min(ax, zx))), ptr(Inches(math.Min(ay, zy)))
		out.W, out.H = ptr(Inches(math.Abs(zx-ax))), ptr(Inches(math.Abs(zy-ay)))
		out.FlipH, out.FlipV = ptr(dx < 0), ptr(dy < 0)
	} else {
		// Rotation is around the shape center; simply swapping W/H at the old
		// origin would move every point. The local y-axis runs opposite slide x.
		cx, cy := (ax+zx)/2, (ay+zy)/2
		out.W, out.H = ptr(Inches(math.Abs(zy-ay))), ptr(Inches(math.Abs(zx-ax)))
		out.X, out.Y = ptr(Inches(cx-out.W.Val/2)), ptr(Inches(cy-out.H.Val/2))
		out.FlipH, out.FlipV = ptr(dy < 0), ptr(dx > 0)
		out.Rotate = 90
	}
	if math.Round(out.W.Val*EMU) == 0 || math.Round(out.H.Val*EMU) == 0 {
		return fail("zero native EMU endpoint extent cannot retain a detour")
	}
	for i := 1; i <= len(points)-3; i++ {
		axis := (i - 1) % 2
		if !horizontal {
			axis = 1 - axis
		}
		denom := z[axis] - a[axis]
		guide := math.Round((points[i][axis] - a[axis]) / denom * 100000)
		if math.IsNaN(guide) || math.IsInf(guide, 0) || guide < -2147483647 || guide > 2147483647 {
			return fail("bend guide exceeds signed DrawingML integer range")
		}
		out.Adjustments[fmt.Sprintf("adj%d", i)] = int(guide)
	}
	return ShapeType(fmt.Sprintf("bentConnector%d", len(points)-1)), out, nil
}
