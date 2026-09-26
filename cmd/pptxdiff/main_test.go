package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompareReportsRGBAndRGBAChanges(t *testing.T) {
	ref := image.NewNRGBA(image.Rect(0, 0, 3, 2))
	cand := image.NewNRGBA(image.Rect(0, 0, 3, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 3; x++ {
			ref.SetNRGBA(x, y, color.NRGBA{R: 100, G: 100, B: 100, A: 255})
			cand.SetNRGBA(x, y, color.NRGBA{R: 100, G: 100, B: 100, A: 255})
		}
	}
	cand.SetNRGBA(1, 0, color.NRGBA{R: 101, G: 103, B: 110, A: 255}) // max RGB delta 10
	cand.SetNRGBA(2, 1, color.NRGBA{R: 100, G: 100, B: 100, A: 254}) // alpha-only exact mismatch
	r := compare(ref, cand, sha256.Sum256([]byte("r")), sha256.Sum256([]byte("c"))).report
	if r.ExactMismatchCount != 2 || r.ExactMismatchFraction != 1.0/3 {
		t.Fatalf("unexpected exact mismatch metrics: %+v", r)
	}
	if r.MeanAbsoluteRGBError != 14.0/18 {
		t.Fatalf("mean RGB error = %v, want %v", r.MeanAbsoluteRGBError, 14.0/18)
	}
	if r.MaxDelta != 10 || r.PixelsOver1 != 1 || r.PixelsOver3 != 1 || r.PixelsOver10 != 0 {
		t.Fatalf("unexpected RGB thresholds: %+v", r)
	}
	if r.MismatchBounds == nil || *r.MismatchBounds != (boundingBox{X: 1, Y: 0, Width: 2, Height: 2}) {
		t.Fatalf("unexpected mismatch bounds: %+v", r.MismatchBounds)
	}
}

func TestCompareIdenticalHasNoBoundingBox(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	r := compare(img, img, [32]byte{}, [32]byte{}).report
	if r.ExactMismatchCount != 0 || r.ExactMismatchFraction != 0 || r.MismatchBounds != nil {
		t.Fatalf("unexpected identical-image report: %+v", r)
	}
}

func TestExecuteWritesArtifactsAndRefusesClobber(t *testing.T) {
	dir := t.TempDir()
	refPath, candPath, out := filepath.Join(dir, "ref.png"), filepath.Join(dir, "candidate.png"), filepath.Join(dir, "out")
	writeTestPNG(t, refPath, image.NewNRGBA(image.Rect(0, 0, 2, 1)))
	candidate := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	candidate.SetNRGBA(1, 0, color.NRGBA{R: 255, A: 255})
	writeTestPNG(t, candPath, candidate)
	if err := execute(refPath, candPath, out, false); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"report.json", "difference.png", "overlay.png"} {
		if _, err := os.Stat(filepath.Join(out, name)); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
	data, err := os.ReadFile(filepath.Join(out, "report.json"))
	if err != nil {
		t.Fatal(err)
	}
	var r report
	if err := json.Unmarshal(data, &r); err != nil {
		t.Fatal(err)
	}
	if r.ExactMismatchCount != 1 {
		t.Fatalf("report mismatch count = %d, want 1", r.ExactMismatchCount)
	}
	if err := execute(refPath, candPath, out, false); err == nil || !strings.Contains(err.Error(), "--overwrite") {
		t.Fatalf("expected actionable clobber error, got %v", err)
	}
	if err := execute(refPath, candPath, out, true); err != nil {
		t.Fatalf("overwrite should succeed: %v", err)
	}
}

func TestExecuteRejectsDimensionMismatch(t *testing.T) {
	dir := t.TempDir()
	a, b := filepath.Join(dir, "a.png"), filepath.Join(dir, "b.png")
	writeTestPNG(t, a, image.NewNRGBA(image.Rect(0, 0, 1, 1)))
	writeTestPNG(t, b, image.NewNRGBA(image.Rect(0, 0, 2, 1)))
	err := execute(a, b, filepath.Join(dir, "out"), false)
	if err == nil || !strings.Contains(err.Error(), "same pixel dimensions") {
		t.Fatalf("expected dimension guidance, got %v", err)
	}
}

func TestRunErrorsGoToStderr(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"--reference", "missing.png", "--candidate", "also-missing.png", "--out", "out"}, &stdout, &stderr)
	if code != 1 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "pptxdiff:") {
		t.Fatalf("unexpected run result: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func writeTestPNG(t *testing.T, path string, img image.Image) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}
