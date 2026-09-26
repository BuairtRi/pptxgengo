package pptx

import "testing"

func TestPatternFillXMLIsBoundedAndDeterministic(t *testing.T) {
	got := genXmlColorSelection(&ShapeFillProps{Pattern: &ShapePatternFillProps{Preset: PatternWdUpDiag, Foreground: "CED7E6", Background: "FFFFFF"}})
	want := `<a:pattFill prst="wdUpDiag"><a:fgClr><a:srgbClr val="CED7E6"/></a:fgClr><a:bgClr><a:srgbClr val="FFFFFF"/></a:bgClr></a:pattFill>`
	if got != want {
		t.Fatalf("pattern XML = %q, want %q", got, want)
	}
}

func TestShapeOptionsRejectUnboundedXMLValues(t *testing.T) {
	p := New()
	s := p.AddSlide()
	badPattern := &ShapeProps{Fill: &ShapeFillProps{Pattern: &ShapePatternFillProps{Preset: PatternType(`wdUpDiag\" injected="1`), Foreground: "CED7E6", Background: "FFFFFF"}}}
	if err := s.AddShape(ShapeTypeHomePlate, badPattern); err == nil {
		t.Fatal("unrecognized pattern accepted")
	}
	if err := s.AddShape(ShapeTypeHomePlate, &ShapeProps{Adjustments: map[string]int{`adj\" injected="1`: 1}}); err == nil {
		t.Fatal("unrecognized adjustment name accepted")
	}
	if err := s.AddShape(ShapeTypeHomePlate, &ShapeProps{Adjustments: map[string]int{"adj": 100001}}); err == nil {
		t.Fatal("out-of-range adjustment accepted")
	}
}
