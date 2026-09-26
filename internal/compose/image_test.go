package compose

import (
	"math"
	"testing"
)

func TestImagePlacementValidation(t *testing.T) {
	tests := []CanvasSpec{
		{ID: "a", ImageFit: "guess"},
		{ID: "a", ImageFit: "source_crop"},
		{ID: "a", ImageFit: "source_crop", ImageCrop: &ImageCropSpec{Left: .8, Right: .3}},
		{ID: "a", ImageFit: "source_crop", ImageCrop: &ImageCropSpec{Top: math.NaN()}},
		{ID: "a", ImageFit: "contain", FocalX: func() *float64 { x := .5; return &x }()},
	}
	for _, c := range tests {
		if validateImagePlacement(c) == nil {
			t.Fatalf("accepted invalid %+v", c)
		}
	}
	for _, mode := range []string{"", "preserve", "cover", "contain", "stretch"} {
		if err := validateImagePlacement(CanvasSpec{ID: "ok", ImageFit: mode}); err != nil {
			t.Fatal(err)
		}
	}
}
