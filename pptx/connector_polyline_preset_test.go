package pptx

import (
	"math"
	"reflect"
	"testing"
)

func connectorPresetReconstruction(s ShapeProps, count int) [][2]float64 {
	a := func(i int) float64 { return float64(s.Adjustments["adj"+string(rune('0'+i))]) / 100000 }
	var points [][2]float64
	switch count {
	case 3:
		points = [][2]float64{{0, 0}, {1, 0}, {1, 1}}
	case 4:
		points = [][2]float64{{0, 0}, {a(1), 0}, {a(1), 1}, {1, 1}}
	case 5:
		points = [][2]float64{{0, 0}, {a(1), 0}, {a(1), a(2)}, {1, a(2)}, {1, 1}}
	case 6:
		points = [][2]float64{{0, 0}, {a(1), 0}, {a(1), a(2)}, {a(3), a(2)}, {a(3), 1}, {1, 1}}
	}
	cx, cy := s.X.Val+s.W.Val/2, s.Y.Val+s.H.Val/2
	for i, p := range points {
		if boolDeref(s.FlipH) {
			p[0] = 1 - p[0]
		}
		if boolDeref(s.FlipV) {
			p[1] = 1 - p[1]
		}
		x, y := s.X.Val+p[0]*s.W.Val, s.Y.Val+p[1]*s.H.Val
		if s.Rotate == 90 {
			x, y = cx-(y-cy), cy+(x-cx)
		}
		points[i] = [2]float64{x, y}
	}
	return points
}
func TestNativePolylineConnectorPresetAllBendsAxesAndDirections(t *testing.T) {
	routes := [][][2]float64{
		{{0, 0}, {1, 0}, {1, 1}},
		{{0, 0}, {.25, 0}, {.25, 1}, {1, 1}},
		{{0, 0}, {.25, 0}, {.25, .5}, {1, .5}, {1, 1}},
		{{0, 0}, {.25, 0}, {.25, .5}, {.75, .5}, {.75, 1}, {1, 1}},
	}
	for _, route := range routes {
		for _, vertical := range []bool{false, true} {
			for _, flipX := range []bool{false, true} {
				for _, flipY := range []bool{false, true} {
					points := append([][2]float64(nil), route...)
					for i, p := range points {
						if vertical {
							p[0], p[1] = p[1], p[0]
						}
						if flipX {
							p[0] = 1 - p[0]
						}
						if flipY {
							p[1] = 1 - p[1]
						}
						points[i] = p
					}
					original := ShapeProps{PositionProps: PositionProps{X: ptr(Inches(2)), Y: ptr(Inches(3)), W: ptr(Inches(4)), H: ptr(Inches(7))}, ObjectNameProps: ObjectNameProps{ObjectName: "Owned route"}, Line: &ShapeLineProps{Width: 1}}
					before := original
					preset, out, e := NativePolylineConnectorPreset(original, points)
					if e != nil {
						t.Fatal(e)
					}
					if preset != ShapeType("bentConnector"+string(rune('0'+len(points)-1))) || len(out.Adjustments) != len(points)-3 || out.ObjectName != original.ObjectName || out.Line != original.Line || len(out.Points) != 0 {
						t.Fatalf("wrong output %v %+v", preset, out)
					}
					if !reflect.DeepEqual(before, original) {
						t.Fatal("conversion mutated caller")
					}
					actual := connectorPresetReconstruction(out, len(points))
					for i, p := range points {
						want := [2]float64{2 + p[0]*4, 3 + p[1]*7}
						if math.Abs(actual[i][0]-want[0]) > 1e-10 || math.Abs(actual[i][1]-want[1]) > 1e-10 {
							t.Fatalf("%d vertex vertical=%v flips=%v,%v point%d actual%v want%v", len(points), vertical, flipX, flipY, i, actual[i], want)
						}
					}
				}
			}
		}
	}
}
func TestNativePolylineConnectorPresetNegativeAndOutsideGuides(t *testing.T) {
	shape := ShapeProps{PositionProps: PositionProps{X: ptr(Inches(1)), Y: ptr(Inches(2)), W: ptr(Inches(5)), H: ptr(Inches(3))}}
	points := [][2]float64{{.25, .25}, {0, .25}, {0, .75}, {.75, .75}}
	_, out, e := NativePolylineConnectorPreset(shape, points)
	if e != nil {
		t.Fatal(e)
	}
	if out.Adjustments["adj1"] != -50000 {
		t.Fatalf("outside route clamped %+v", out.Adjustments)
	}
	for i, actual := range connectorPresetReconstruction(out, len(points)) {
		if math.Abs(actual[0]-(1+points[i][0]*5)) > 1e-10 || math.Abs(actual[1]-(2+points[i][1]*3)) > 1e-10 {
			t.Fatal("outside guide changed route")
		}
	}
}
func TestNativePolylineConnectorPresetExplicitUnsupportedBoundaries(t *testing.T) {
	shape := ShapeProps{PositionProps: PositionProps{W: ptr(Inches(4)), H: ptr(Inches(3))}}
	for _, points := range [][][2]float64{
		{{0, 0}, {1, 1}},
		{{0, 0}, {.5, .5}, {1, 1}},
		{{0, 0}, {0, 0}, {1, 1}},
		{{0, 0}, {.25, 0}, {.5, 0}, {.5, 1}, {1, 1}},
		{{0, 0}, {.5, 0}, {.5, 1}, {1, 1}, {1, 0}},
		{{0, 0}, {.2, 0}, {.2, .2}, {.4, .2}, {.4, .6}, {.7, .6}, {.7, 1}, {1, 1}},
		{{0, 0}, {math.NaN(), 0}, {1, 1}},
		{{0, 0}, {1.01, 0}, {1.01, 1}, {1, 1}},
		{{0, 0}, {1, 0}, {1, 1}, {1e-10, 1}},
	} {
		if _, _, e := NativePolylineConnectorPreset(shape, points); e == nil {
			t.Fatalf("unsupported route accepted %+v", points)
		}
	}
	shape.Rotate = 90
	if _, _, e := NativePolylineConnectorPreset(shape, [][2]float64{{0, 0}, {1, 0}, {1, 1}}); e == nil {
		t.Fatal("preexisting transform accepted")
	}
	shape.Rotate = 0
	shape.W = ptr(Percent(4))
	if _, _, e := NativePolylineConnectorPreset(shape, [][2]float64{{0, 0}, {1, 0}, {1, 1}}); e == nil {
		t.Fatal("percentage allocation accepted")
	}
}
