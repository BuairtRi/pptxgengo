package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEngineeringCommandsValidateFlags(t *testing.T) {
	for name, run := range map[string]func([]string) error{"match": runLibraryMatch, "authoring": runLibraryAuthoring, "gallery": runAssetGallery, "swap": runProjectSwap} {
		if err := run([]string{"--unknown"}); err == nil {
			t.Fatal(name, "accepted unknown flag")
		}
	}
	if err := runLibraryIndex("library-find", []string{"--asset-kind", "photo"}); err == nil || !strings.Contains(err.Error(), "requires") {
		t.Fatal(err)
	}
}

func TestAssetGalleryVerifiedOriginals(t *testing.T) {
	if testing.Short() {
		t.Skip("gallery integration reads private registered brand artwork; run make test-integration with WMDS_BRANDING_ROOT")
	}
	root := filepath.Join(t.TempDir(), "gallery")
	if err := runAssetGallery([]string{"--out", root, "--kind", "icon", "--query", "risk", "--limit", "2"}); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"index.html", "assets.json"} {
		if data, err := os.ReadFile(filepath.Join(root, name)); err != nil || len(data) == 0 {
			t.Fatal(name, err)
		}
	}
	entries, err := os.ReadDir(filepath.Join(root, "assets"))
	if err != nil || len(entries) != 6 {
		t.Fatal("missing color variants", len(entries), err)
	}
	if err := runAssetGallery([]string{"--out", root, "--kind", "icon"}); err == nil {
		t.Fatal("overwrote gallery")
	}
}

func TestLibraryMatchGapExitAndReport(t *testing.T) {
	root := t.TempDir()
	page := filepath.Join(root, "page.yaml")
	out := filepath.Join(root, "candidates")
	if err := os.WriteFile(page, []byte("title: A supplied claim\nrelationship: parallel\n"), 0644); err != nil {
		t.Fatal(err)
	}
	sink, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer sink.Close()
	previous := os.Stdout
	os.Stdout = sink
	defer func() { os.Stdout = previous }()
	bundle, err := filepath.Abs("../../library/wm-design-system/v5")
	if err != nil {
		t.Fatal(err)
	}
	err = runLibraryMatch([]string{"--page", page, "--templates", "cards/3", "--out", out, "--bundle", bundle})
	if err == nil || !strings.Contains(err.Error(), "no_complete_candidate") {
		t.Fatal("gap-only command passed", err)
	}
	report, err := os.ReadFile(filepath.Join(out, "match-report.json"))
	if err != nil || !strings.Contains(string(report), `"passed": 0`) {
		t.Fatal("missing gap report", err)
	}
	if _, err = os.Stat(filepath.Join(out, "candidates.pptx")); !os.IsNotExist(err) {
		t.Fatal("created synthetic candidate", err)
	}
}

func TestAssetGalleryRasterPreviewPreservesAspectAndTransparency(t *testing.T) {
	source := image.NewNRGBA(image.Rect(0, 0, 1200, 600))
	source.SetNRGBA(600, 300, color.NRGBA{R: 255, A: 128})
	var input bytes.Buffer
	if err := png.Encode(&input, source); err != nil {
		t.Fatal(err)
	}
	before := append([]byte(nil), input.Bytes()...)
	preview, extension, derived, err := assetGalleryPreview(input.Bytes(), "sample.png")
	if err != nil {
		t.Fatal(err)
	}
	result, err := png.Decode(bytes.NewReader(preview))
	if err != nil {
		t.Fatal(err)
	}
	if extension != ".png" || !derived || result.Bounds().Dx() != 420 || result.Bounds().Dy() != 210 {
		t.Fatal("invalid preview", extension, derived, result.Bounds())
	}
	_, _, _, alpha := result.At(0, 0).RGBA()
	if alpha != 0 || !bytes.Equal(before, input.Bytes()) {
		t.Fatal("changed transparency or original")
	}
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10"></svg>`)
	vector, extension, derived, err := assetGalleryPreview(svg, "vector.svg")
	if err != nil || derived || extension != ".svg" || !bytes.Equal(vector, svg) {
		t.Fatal("altered registered vector", err)
	}
}
