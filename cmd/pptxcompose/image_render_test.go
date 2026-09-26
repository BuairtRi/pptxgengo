package main

import (
	"archive/zip"
	"bytes"
	"image"
	stdcolor "image/color"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/compose"
)

func TestImageCropModes(t *testing.T) {
	e := element{Name: "sample", Frame: frame{Width: 100, Height: 100}}
	if _, err := imageCrop(e, 200, 100); err == nil {
		t.Fatal("default allowed distortion")
	}
	e.ImageFit = "contain"
	c, err := imageCrop(e, 200, 100)
	if err != nil || c.Top != -.5 || c.Bottom != -.5 || c.Left != 0 {
		t.Fatalf("contain: %+v %v", c, err)
	}
	e.ImageFit = "cover"
	e.FocalX = pointer(1.0)
	c, err = imageCrop(e, 200, 100)
	if err != nil || c.Left != .5 || c.Right != 0 {
		t.Fatalf("right cover: %+v %v", c, err)
	}
	e.FocalX = nil
	e.ImageFit = "source_crop"
	e.ImageCrop = &compose.ImageCropSpec{Left: .5}
	if _, err = imageCrop(e, 200, 100); err != nil {
		t.Fatal(err)
	}
	e.ImageCrop = &compose.ImageCropSpec{Left: .2}
	if _, err = imageCrop(e, 200, 100); err == nil {
		t.Fatal("source crop allowed distortion")
	}
}

func TestImageCropNativePackage(t *testing.T) {
	var buf bytes.Buffer
	im := image.NewRGBA(image.Rect(0, 0, 200, 100))
	im.Set(0, 0, stdcolor.RGBA{R: 255, A: 255})
	if err := png.Encode(&buf, im); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), "fixture.png")
	if err := os.WriteFile(p, buf.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	e := element{Name: "picture", Kind: "image", Frame: frame{X: 10, Y: 20, Width: 100, Height: 100}, AssetPath: p, AssetSHA256: hash(buf.Bytes()), AltText: "test", ImageFit: "cover", FocalX: pointer(1.0)}
	b, err := render([]renderSlide{{ID: "test", Width: 960, Height: 540, Elements: []element{e}}})
	if err != nil {
		t.Fatal(err)
	}
	z, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range z.File {
		if f.Name != "ppt/slides/slide1.xml" {
			continue
		}
		r, _ := f.Open()
		raw, _ := io.ReadAll(r)
		r.Close()
		s := string(raw)
		if !strings.Contains(s, `<a:srcRect l="50000" r="0" t="0" b="0"/>`) {
			t.Fatalf("wrong crop: %s", s)
		}
		if !strings.Contains(s, `<a:ext cx="1270000" cy="1270000"/>`) {
			t.Fatalf("wrong picture frame: %s", s)
		}
		return
	}
	t.Fatal("missing slide")
}

func TestImageContainPreservesRatio(t *testing.T) {
	for _, wh := range [][2]int{{200, 100}, {100, 300}, {512, 512}} {
		e := element{Frame: frame{Width: 230, Height: 145}, ImageFit: "contain"}
		c, err := imageCrop(e, wh[0], wh[1])
		if err != nil {
			t.Fatal(err)
		}
		displayRatio := (e.Frame.Width / (1 - c.Left - c.Right)) / (e.Frame.Height / (1 - c.Top - c.Bottom))
		if math.Abs(displayRatio-float64(wh[0])/float64(wh[1])) > 1e-9 {
			t.Fatal("contain distorted source")
		}
	}
}
