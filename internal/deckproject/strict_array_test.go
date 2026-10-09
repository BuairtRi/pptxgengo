package deckproject

import (
	"reflect"
	"strings"
	"testing"
)

func TestStrictFixedGeometryArrays(t *testing.T) {
	p := &Project{}
	typ := reflect.TypeOf([][2]float64{})
	if err := p.shapeType([]any{[]any{float64(0), float64(1)}}, typ, "/route/points"); err != nil {
		t.Fatal(err)
	}
	for _, value := range []any{
		[]any{[]any{float64(0)}},
		[]any{[]any{float64(0), float64(1), float64(2)}},
		[]any{[]any{float64(0), "1"}},
		[]any{map[string]any{"x": float64(0), "y": float64(1)}},
	} {
		if err := p.shapeType(value, typ, "/route/points"); err == nil || !strings.Contains(err.Error(), "/route/points/0") {
			t.Fatalf("malformed fixed point accepted or unlocated: %v: %v", value, err)
		}
	}
}
