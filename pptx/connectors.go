package pptx

import (
	"fmt"
	"math"
)

func validConnectorBounds(p PositionProps, rectangle bool) bool {
	w, h := 1.0, 1.0
	for _, c := range []*Coord{p.X, p.Y, p.W, p.H} {
		if c != nil && (math.IsNaN(c.Val) || math.IsInf(c.Val, 0) || math.Abs(c.Val) > 1e6) {
			return false
		}
	}
	if p.W != nil {
		w = p.W.Val
	}
	if p.H != nil {
		h = p.H.Val
	}
	if rectangle {
		return w > 0 && h > 0
	}
	return w >= 0 && h >= 0 && (w > 0 || h > 0)
}

func validConnectorConnection(c ConnectorConnection) bool {
	return c.Begin.ObjectName != "" && c.End.ObjectName != "" && c.Begin.ObjectName != c.End.ObjectName && c.Begin.Site >= 0 && c.Begin.Site <= 3 && c.End.Site >= 0 && c.End.Site <= 3
}

func connectorTargetID(slide *SlideBaseProps, endpoint ConnectorEndpoint) int {
	for index, object := range slide.SlideObjects {
		if object.Options != nil && object.Options.ObjectName == encodeXmlEntities(endpoint.ObjectName) {
			return index + 2
		}
	}
	return 0 // Write rejects unresolved/ambiguous names before serialization.
}

func validateNativeConnectors(slide *PresSlide) error {
	// Only construct the name index for slides using this opt-in API.
	needed := false
	for _, object := range slide.SlideObjects {
		needed = needed || object.Options != nil && object.Options.NativeConnection != nil
	}
	if !needed {
		return nil
	}
	targets := map[string][]SlideObject{}
	for _, object := range slide.SlideObjects {
		if object.Options != nil {
			targets[object.Options.ObjectName] = append(targets[object.Options.ObjectName], object)
		}
	}
	for _, object := range slide.SlideObjects {
		if object.Options == nil || object.Options.NativeConnection == nil {
			continue
		}
		connection := *object.Options.NativeConnection
		if object.Type != SlideObjectTypeText || object.Shape != ShapeTypeLine || !validConnectorConnection(connection) || !validConnectorBounds(object.Options.PositionProps, false) {
			return fmt.Errorf("invalid native connector")
		}
		for _, endpoint := range []ConnectorEndpoint{connection.Begin, connection.End} {
			matches := targets[encodeXmlEntities(endpoint.ObjectName)]
			if len(matches) != 1 {
				return fmt.Errorf("native connector target missing or ambiguous: %s", endpoint.ObjectName)
			}
			target := matches[0]
			if target.Type != SlideObjectTypeText || target.Shape != ShapeTypeRect || target.Options.NativeConnection != nil || target.Options.Placeholder != "" || target.Options.Rotate != 0 || boolDeref(target.Options.FlipH) || boolDeref(target.Options.FlipV) || !validConnectorBounds(target.Options.PositionProps, true) {
				return fmt.Errorf("native connector target must be an unrotated named rectangle with positive finite bounds: %s", endpoint.ObjectName)
			}
		}
	}
	return nil
}
