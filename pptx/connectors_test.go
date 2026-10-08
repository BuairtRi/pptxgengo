package pptx

import (
	"encoding/xml"
	"io"
	"math"
	"strings"
	"testing"
)

func nativeConnectorOptions() *ConnectorProps {
	return &ConnectorProps{ShapeProps: ShapeProps{PositionProps: PositionProps{X: ptr(Inches(2)), Y: ptr(Inches(2)), W: ptr(Inches(3)), H: ptr(Inches(0))}, ObjectNameProps: ObjectNameProps{ObjectName: "edge"}, Line: &ShapeLineProps{ShapeFillProps: ShapeFillProps{Color: "070154"}, EndArrowType: "triangle"}}, Connection: ConnectorConnection{Begin: ConnectorEndpoint{ObjectName: "input & source", Site: 3}, End: ConnectorEndpoint{ObjectName: "output", Site: 1}}}
}
func connectorRect(t *testing.T, s *Slide, name string, shape ShapeType) {
	t.Helper()
	if e := s.AddShape(shape, &ShapeProps{ObjectNameProps: ObjectNameProps{ObjectName: name}}); e != nil {
		t.Fatal(e)
	}
}

func TestNativeConnectorForwardReferencesAndConnectionCopy(t *testing.T) {
	p := New()
	s := p.AddSlide()
	opts := nativeConnectorOptions()
	if e := s.AddConnector(opts); e != nil {
		t.Fatal(e)
	}
	// Caller edits cannot redirect the retained connection.
	opts.Connection.Begin.ObjectName, opts.Connection.Begin.Site = "absent", 0
	connectorRect(t, s, "input & source", ShapeTypeRect)
	connectorRect(t, s, "output", ShapeTypeRect)
	raw := unzipParts(t, mustWrite(t, p))["ppt/slides/slide1.xml"]
	start, end := strings.Index(raw, "<p:cxnSp>"), strings.Index(raw, "</p:cxnSp>")
	if start < 0 || end < start {
		t.Fatal("missing native connector", raw)
	}
	edge := raw[start:end]
	for _, want := range []string{`<p:cNvPr id="2" name="edge">`, `<a:stCxn id="3" idx="3"/>`, `<a:endCxn id="4" idx="1"/>`, `<a:prstGeom prst="line">`, `<a:tailEnd type="triangle"/>`} {
		if !strings.Contains(edge, want) {
			t.Fatalf("missing %s in %s", want, edge)
		}
	}
	if strings.Contains(edge, "txBody") || strings.Contains(edge, "nvSpPr>") {
		t.Fatal("connector serialized as a text shape", edge)
	}
	decoder := xml.NewDecoder(strings.NewReader(raw))
	for {
		if _, e := decoder.Token(); e == io.EOF {
			break
		} else if e != nil {
			t.Fatal(e)
		}
	}
}

func TestNativeConnectorRejectsInvalidInputWithoutAppending(t *testing.T) {
	for name, mutate := range map[string]func(*ConnectorProps){
		"same target":     func(o *ConnectorProps) { o.Connection.End.ObjectName = o.Connection.Begin.ObjectName },
		"empty target":    func(o *ConnectorProps) { o.Connection.Begin.ObjectName = "" },
		"negative site":   func(o *ConnectorProps) { o.Connection.Begin.Site = -1 },
		"unknown site":    func(o *ConnectorProps) { o.Connection.End.Site = 4 },
		"fill":            func(o *ConnectorProps) { o.Fill = &ShapeFillProps{Color: "FFFFFF"} },
		"rotation":        func(o *ConnectorProps) { o.Rotate = 90 },
		"custom geometry": func(o *ConnectorProps) { o.RectRadius = 1 },
		"infinite":        func(o *ConnectorProps) { o.X = ptr(Inches(math.Inf(1))) },
		"negative extent": func(o *ConnectorProps) { o.W = ptr(Inches(-1)) },
		"coincident":      func(o *ConnectorProps) { o.W = ptr(Inches(0)) },
	} {
		t.Run(name, func(t *testing.T) {
			s := New().AddSlide()
			o := nativeConnectorOptions()
			mutate(o)
			if e := s.AddConnector(o); e == nil || len(s.ps.SlideObjects) != 0 {
				t.Fatal("accepted or partially appended invalid connector", e)
			}
		})
	}
	if e := New().AddSlide().AddConnector(nil); e == nil {
		t.Fatal("nil accepted")
	}
}

func TestNativeConnectorRejectsMissingAmbiguousAndUnsupportedTargets(t *testing.T) {
	for _, kind := range []string{"missing", "duplicate", "ellipse", "rotated", "flipped", "zero width", "nonfinite", "changed connection"} {
		t.Run(kind, func(t *testing.T) {
			p := New()
			s := p.AddSlide()
			if e := s.AddConnector(nativeConnectorOptions()); e != nil {
				t.Fatal(e)
			}
			connectorRect(t, s, "input & source", ShapeTypeRect)
			if kind != "missing" {
				connectorRect(t, s, "output", ShapeTypeRect)
			}
			if kind == "duplicate" {
				connectorRect(t, s, "output", ShapeTypeRect)
			}
			switch kind {
			case "ellipse":
				s.ps.SlideObjects[1].Shape = ShapeTypeEllipse
			case "rotated":
				s.ps.SlideObjects[1].Options.Rotate = 90
			case "flipped":
				s.ps.SlideObjects[1].Options.FlipH = ptr(true)
			case "zero width":
				s.ps.SlideObjects[1].Options.W = ptr(Inches(0))
			case "nonfinite":
				s.ps.SlideObjects[1].Options.X = ptr(Inches(math.NaN()))
			case "changed connection":
				s.ps.SlideObjects[0].Options.NativeConnection.End.Site = 8
			}
			if data, e := p.Write(); e == nil || data != nil {
				t.Fatal("invalid target produced a package", e)
			}
		})
	}
}

func TestNativeConnectorDoesNotChangeLegacyLineSerialization(t *testing.T) {
	p := New()
	s := p.AddSlide()
	opts := nativeConnectorOptions()
	if e := s.AddShape(ShapeTypeLine, &opts.ShapeProps); e != nil {
		t.Fatal(e)
	}
	raw := unzipParts(t, mustWrite(t, p))["ppt/slides/slide1.xml"]
	if strings.Contains(raw, "cxnSp") || strings.Contains(raw, "stCxn") || !strings.Contains(raw, "<p:sp>") {
		t.Fatal("legacy line changed")
	}
}

func TestNativeConnectorElbowMetadataAndRouteCopy(t *testing.T) {
	p := New()
	s := p.AddSlide()
	opts := nativeConnectorOptions()
	opts.Route = &ConnectorRoute{Preset: "bentConnector3", Adjustment: 35000}
	opts.Rotate = 90
	if e := s.AddConnector(opts); e != nil {
		t.Fatal(e)
	}
	opts.Route.Adjustment = 90000
	connectorRect(t, s, "input & source", ShapeTypeRect)
	connectorRect(t, s, "output", ShapeTypeRect)
	raw := unzipParts(t, mustWrite(t, p))["ppt/slides/slide1.xml"]
	for _, want := range []string{`<p:cxnSp>`, `prst="bentConnector3"`, `name="adj1" fmla="val 35000"`, `rot="5400000"`, `<a:stCxn`, `<a:endCxn`} {
		if !strings.Contains(raw, want) {
			t.Fatal("missing elbow metadata", want)
		}
	}
	for _, route := range []*ConnectorRoute{{Preset: "curvedConnector3", Adjustment: 50000}, {Preset: "bentConnector3", Adjustment: 2147483648}} {
		opts := nativeConnectorOptions()
		opts.Route = route
		slide := New().AddSlide()
		if e := slide.AddConnector(opts); e == nil || len(slide.ps.SlideObjects) != 0 {
			t.Fatal("invalid route appended")
		}
	}
}
