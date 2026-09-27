package main

import (
	"archive/zip"
	"bytes"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/compose"
)

func structuralImageFixture(t *testing.T, mode string) ([]byte, []renderSlide) {
	t.Helper()
	var data bytes.Buffer
	if err := png.Encode(&data, image.NewRGBA(image.Rect(0, 0, 200, 100))); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), "fixture.png")
	if err := os.WriteFile(file, data.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	e := element{Name: "canvas:cGljdHVyZQ", Kind: "image", Frame: frame{X: 10, Y: 20, Width: 100, Height: 100}, AssetPath: file, AssetSHA256: hash(data.Bytes()), AltText: "fixture", ImageFit: mode}
	switch mode {
	case "source_crop":
		e.ImageCrop = &compose.ImageCropSpec{Left: .5}
	case "cover":
		e.FocalX = pointer(1.0)
	}
	slides := []renderSlide{{ID: "image", Width: 960, Height: 540, Elements: []element{e}}}
	deck, err := render(slides)
	if err != nil {
		t.Fatal(err)
	}
	return deck, slides
}

func TestImageStructureBindsDeclaredRotation(t *testing.T) {
	for _, rotation := range []float64{1, -15, 90} {
		t.Run(strconv.FormatFloat(rotation, 'f', -1, 64), func(t *testing.T) {
			_, slides := structuralImageFixture(t, "stretch")
			slides[0].Elements[0].RotationDeg = rotation
			deck, err := render(slides)
			if err != nil {
				t.Fatal(err)
			}
			if err := validateImageStructure(deck, slides); err != nil {
				t.Fatal(err)
			}
			slides[0].Elements[0].RotationDeg += 1
			if err := validateImageStructure(deck, slides); err == nil {
				t.Fatal("unexpected image rotation accepted")
			}
		})
	}
}

func TestImageStructureBindsCropModes(t *testing.T) {
	for _, mode := range []string{"source_crop", "contain", "cover"} {
		deck, slides := structuralImageFixture(t, mode)
		if err := validateImageStructure(deck, slides); err != nil {
			t.Fatalf("%s: %v", mode, err)
		}
	}
}

func TestImageStructureRejectsCropAndMediaCorruption(t *testing.T) {
	deck, slides := structuralImageFixture(t, "source_crop")
	tests := []func(string, []byte) []byte{
		func(name string, body []byte) []byte {
			if name == "ppt/slides/slide1.xml" {
				return bytes.Replace(body, []byte(`l="50000"`), []byte(`l="49999"`), 1)
			}
			return body
		},
		func(name string, body []byte) []byte {
			if name == "ppt/slides/slide1.xml" {
				return bytes.Replace(body, []byte(`<a:stretch/>`), []byte(`<a:stretch><a:fillRect/></a:stretch>`), 1)
			}
			return body
		},
		func(name string, body []byte) []byte {
			if strings.HasPrefix(name, "ppt/media/") && len(body) > 0 {
				out := append([]byte(nil), body...)
				out[len(out)-1] ^= 1
				return out
			}
			return body
		},
	}
	for i, mutate := range tests {
		changed := rewriteImageFixture(t, deck, mutate)
		if err := validateImageStructure(changed, slides); err == nil {
			t.Fatalf("corruption %d accepted", i)
		}
	}
}

func TestImageStructureRejectsTransformAndFillRectCorruption(t *testing.T) {
	uncropped, uncroppedSlides := structuralImageFixture(t, "stretch")
	cropped, croppedSlides := structuralImageFixture(t, "source_crop")
	tests := []struct {
		name   string
		deck   []byte
		slides []renderSlide
		mutate func(string, []byte) []byte
	}{
		{
			name: "nondefault fillRect", deck: uncropped, slides: uncroppedSlides,
			mutate: func(name string, body []byte) []byte {
				if name == "ppt/slides/slide1.xml" {
					return bytes.Replace(body, []byte(`<a:fillRect/>`), []byte(`<a:fillRect l="1"/>`), 1)
				}
				return body
			},
		},
		{
			name: "horizontal flip", deck: uncropped, slides: uncroppedSlides,
			mutate: func(name string, body []byte) []byte {
				if name == "ppt/slides/slide1.xml" {
					return mutatePictureTransform(body, ` flipH="1"`)
				}
				return body
			},
		},
		{
			name: "vertical flip", deck: uncropped, slides: uncroppedSlides,
			mutate: func(name string, body []byte) []byte {
				if name == "ppt/slides/slide1.xml" {
					return mutatePictureTransform(body, ` flipV="1"`)
				}
				return body
			},
		},
		{
			name: "rotation", deck: uncropped, slides: uncroppedSlides,
			mutate: func(name string, body []byte) []byte {
				if name == "ppt/slides/slide1.xml" {
					return mutatePictureTransform(body, ` rot="60000"`)
				}
				return body
			},
		},
		{
			name: "srcRect child", deck: cropped, slides: croppedSlides,
			mutate: func(name string, body []byte) []byte {
				if name == "ppt/slides/slide1.xml" {
					return mutateSourceRectChild(body)
				}
				return body
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			changed := rewriteImageFixture(t, test.deck, test.mutate)
			if err := validateImageStructure(changed, test.slides); err == nil {
				t.Fatal("corruption accepted")
			}
		})
	}
}

func mutatePictureTransform(body []byte, attrs string) []byte {
	start := bytes.Index(body, []byte(`name="canvas:cGljdHVyZQ"`))
	if start < 0 {
		return body
	}
	rel := bytes.Index(body[start:], []byte(`<a:xfrm>`))
	if rel < 0 {
		return body
	}
	pos := start + rel
	out := append([]byte(nil), body[:pos]...)
	out = append(out, []byte(`<a:xfrm`+attrs+`>`)...)
	out = append(out, body[pos+len(`<a:xfrm>`):]...)
	return out
}

func mutateSourceRectChild(body []byte) []byte {
	start := bytes.Index(body, []byte(`<a:srcRect `))
	if start < 0 {
		return body
	}
	end := bytes.Index(body[start:], []byte(`/>`))
	if end < 0 {
		return body
	}
	end += start
	out := append([]byte(nil), body[:end]...)
	out = append(out, []byte(`><a:extLst/></a:srcRect>`)...)
	out = append(out, body[end+2:]...)
	return out
}

func rewriteImageFixture(t *testing.T, deck []byte, mutate func(string, []byte) []byte) []byte {
	t.Helper()
	z, err := zip.NewReader(bytes.NewReader(deck), int64(len(deck)))
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	w := zip.NewWriter(&out)
	for _, f := range z.File {
		r, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(r)
		r.Close()
		if err != nil {
			t.Fatal(err)
		}
		entry, err := w.CreateHeader(&f.FileHeader)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write(mutate(f.Name, body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}
