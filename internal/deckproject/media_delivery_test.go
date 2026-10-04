package deckproject

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"image"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestRegisteredImageDeliveryDerivativePreservesOriginal(t *testing.T) {
	p := example(t)
	photo := image.NewRGBA(image.Rect(0, 0, 2048, 1536))
	state := uint32(12345)
	for i := 0; i < len(photo.Pix); i += 4 {
		for channel := 0; channel < 3; channel++ {
			state = state*1664525 + 1013904223
			photo.Pix[i+channel] = byte(state >> 24)
		}
		photo.Pix[i+3] = 255
	}
	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, photo, &jpeg.Options{Quality: 96}); err != nil {
		t.Fatal(err)
	}
	original := append([]byte(nil), encoded.Bytes()...)
	registered, err := RegisterAsset(p, AssetRegistration{ID: "delivery-photo", Description: "Synthetic high-resolution delivery regression image", Data: original})
	if err != nil {
		t.Fatal(err)
	}
	p, err = Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	p = rewrite(t, p, "enum: [sample-image]", "enum: [sample-image, delivery-photo]")
	p = rewrite(t, p, "photo: sample-image", "photo: delivery-photo")
	pin(t, p)
	for _, resize := range []bool{true, false} {
		if !resize {
			p = rewrite(t, p, "year: 2026", "year: 2026\nmedia_optimization:\n  deduplicate: true\n  resize_jpeg: false\n  compression: true\n  pixels_per_inch: 220\n  jpeg_quality: 90\n  min_savings_percent: 10")
		}
		built, err := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine})
		if err != nil {
			t.Fatal(err)
		}
		root := filepath.Join(p.Root, "builds", built.BuildID)
		data, err := os.ReadFile(filepath.Join(root, "layout-report.json"))
		if err != nil {
			t.Fatal(err)
		}
		var report wmdesign.Report
		if err = json.Unmarshal(data, &report); err != nil {
			t.Fatal(err)
		}
		if report.MediaOptimization == nil || report.MediaOptimization.Policy.ResizeJPEG != resize {
			t.Fatal("project media policy was not propagated", report.MediaOptimization)
		}
		found := false
		for _, part := range report.MediaOptimization.Parts {
			if part.SourceSHA256 != registered.SHA256 {
				continue
			}
			found = true
			if part.Resized != resize {
				t.Fatalf("unexpected registered JPEG derivative: %+v", part)
			}
			if resize && (part.OutputBytes >= part.SourceBytes || part.OutputWidth >= part.SourceWidth || part.OutputSHA256 == part.SourceSHA256) {
				t.Fatalf("derivative did not reduce bytes/resolution: %+v", part)
			}
			deck, err := zip.OpenReader(filepath.Join(root, "deck.pptx"))
			if err != nil {
				t.Fatal(err)
			}
			embedded := false
			for _, file := range deck.File {
				if file.Name != part.OutputPart {
					continue
				}
				r, err := file.Open()
				if err != nil {
					t.Fatal(err)
				}
				payload, err := io.ReadAll(r)
				r.Close()
				if err != nil {
					t.Fatal(err)
				}
				embedded = true
				if digest(payload) != part.OutputSHA256 || (!resize && !bytes.Equal(payload, original)) {
					t.Fatal("embedded output does not match media receipt")
				}
			}
			deck.Close()
			if !embedded {
				t.Fatal("media receipt output part is missing")
			}
		}
		if !found {
			t.Fatal("registered original missing from media receipt")
		}
		retained, err := os.ReadFile(filepath.Join(p.Root, registered.Path))
		if err != nil || !bytes.Equal(retained, original) {
			t.Fatal("build changed registered original", err)
		}
	}
}
