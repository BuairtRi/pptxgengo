package pptx

import (
	"strings"
	"testing"
)

func TestShapeLineJoinEnumAndNativeXML(t *testing.T) {
	for _, join := range []string{"", "round", "bevel", "miter"} {
		slide := newTestSlide()
		props := &ShapeProps{Line: &ShapeLineProps{ShapeFillProps: ShapeFillProps{Color: "070154"}, Width: 22, LineJoin: join}}
		if err := addShapeDefinition(slide, ShapeTypeCustGeom, props); err != nil {
			t.Fatal(err)
		}
		if slide.SlideObjects[0].Options.Line.LineJoin != join {
			t.Fatal("line normalization discarded join")
		}
		body := slideObjectToXml(&slide.SlideBaseProps, nil)
		if join == "" {
			if strings.Contains(body, "<a:round") || strings.Contains(body, "<a:bevel") || strings.Contains(body, "<a:miter") {
				t.Fatal("empty join changed legacy XML")
			}
		} else if !strings.Contains(body, "<a:"+join+"/>") {
			t.Fatal("native join element missing")
		}
	}
	for _, join := range []string{"rounded", "ROUND", `round\"/><a:bevel/>`} {
		slide := newTestSlide()
		if err := addShapeDefinition(slide, ShapeTypeLine, &ShapeProps{Line: &ShapeLineProps{LineJoin: join}}); err == nil || len(slide.SlideObjects) != 0 {
			t.Fatal("invalid enum accepted or partial shape emitted")
		}
	}
}
func TestShapeEmptyLineJoinByteNeutral(t *testing.T) {
	a, b := newTestSlide(), newTestSlide()
	if err := addShapeDefinition(a, ShapeTypeLine, &ShapeProps{Line: &ShapeLineProps{ShapeFillProps: ShapeFillProps{Color: "070154"}, Width: 2}}); err != nil {
		t.Fatal(err)
	}
	if err := addShapeDefinition(b, ShapeTypeLine, &ShapeProps{Line: &ShapeLineProps{ShapeFillProps: ShapeFillProps{Color: "070154"}, Width: 2, LineJoin: ""}}); err != nil {
		t.Fatal(err)
	}
	body := slideObjectToXml(&a.SlideBaseProps, nil)
	if body != slideObjectToXml(&b.SlideBaseProps, nil) || strings.Contains(body, "<a:round") || strings.Contains(body, "<a:miter") || strings.Contains(body, "<a:bevel") {
		t.Fatal("legacy empty join is not byte neutral")
	}
}

func TestShapeLineJoinPublicTextAndImageValidation(t *testing.T) {
	slide := &Slide{ps: newTestSlide()}
	if err := slide.AddText([]TextProps{{Text: "Example"}}, &TextPropsOptions{Line: &ShapeLineProps{LineJoin: "rounded"}}); err == nil || len(slide.ps.SlideObjects) != 0 {
		t.Fatal("AddText accepted invalidjoin or emitted partial object")
	}
	if err := slide.AddImage(&ImageProps{Line: &ShapeLineProps{LineJoin: "rounded"}}); err == nil || len(slide.ps.SlideObjects) != 0 {
		t.Fatal("AddImage accepted invalidjoin or emitted partial object")
	}
	for _, join := range []string{"", "round", "bevel", "miter"} {
		if err := slide.AddText([]TextProps{{Text: "Example"}}, &TextPropsOptions{Line: &ShapeLineProps{ShapeFillProps: ShapeFillProps{Color: "070154"}, Width: 2, LineJoin: join}}); err != nil {
			t.Fatal(err)
		}
	}
}
