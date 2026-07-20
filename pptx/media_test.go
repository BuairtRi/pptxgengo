package pptx

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// tinyPNG is the 5x5 red PNG fixture used across the Go port's tests
// (see scripts/gen-golden.mjs).
const tinyPNGBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAIAAAACDbGyAAAAEUlEQVR42mP4z8CAjBgo5AMA/XwY6DI3DH0AAAAASUVORK5CYII="

func tinyPNGBytes(t *testing.T) []byte {
	t.Helper()
	b, err := base64.StdEncoding.DecodeString(tinyPNGBase64)
	if err != nil {
		t.Fatalf("failed to decode tinyPNGBase64 fixture: %v", err)
	}
	return b
}

// ---------------------------------------------------------------------------
// resolveSlideMediaRels
// ---------------------------------------------------------------------------

func TestResolveSlideMediaRels_Base64Passthrough(t *testing.T) {
	layout := &SlideBaseProps{
		RelsMedia: []SlideRelMedia{
			{Type: "image/png", Path: "/does/not/exist/should-not-be-read.png", Data: "image/png;base64,ALREADYSET", RID: 1, Target: "../media/image-1-1.png"},
		},
	}
	errs := resolveSlideMediaRels(layout)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
	if got := layout.RelsMedia[0].Data.(string); got != "image/png;base64,ALREADYSET" {
		t.Errorf("expected pre-set data to pass through unchanged, got %q", got)
	}
}

func TestResolveSlideMediaRels_LocalFile(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "pic.png")
	content := tinyPNGBytes(t)
	if err := os.WriteFile(fp, content, 0o644); err != nil {
		t.Fatalf("failed to write fixture file: %v", err)
	}

	layout := &SlideBaseProps{
		RelsMedia: []SlideRelMedia{
			{Type: "image/png", Path: fp, RID: 1, Target: "../media/image-1-1.png"},
		},
	}
	errs := resolveSlideMediaRels(layout)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
	want := base64.StdEncoding.EncodeToString(content)
	got, ok := layout.RelsMedia[0].Data.(string)
	if !ok || got != want {
		t.Errorf("Data = %v, want base64 %q", layout.RelsMedia[0].Data, want)
	}
}

func TestResolveSlideMediaRels_MissingFile(t *testing.T) {
	layout := &SlideBaseProps{
		RelsMedia: []SlideRelMedia{
			{Type: "image/png", Path: "/definitely/does/not/exist/nope.png", RID: 1, Target: "../media/image-1-1.png"},
		},
	}
	errs := resolveSlideMediaRels(layout)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	got, ok := layout.RelsMedia[0].Data.(string)
	if !ok || got != IMG_BROKEN {
		t.Errorf("expected Data == IMG_BROKEN for missing file, got %v", layout.RelsMedia[0].Data)
	}
}

func TestResolveSlideMediaRels_HTTP(t *testing.T) {
	content := tinyPNGBytes(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(content)
	}))
	defer srv.Close()

	layout := &SlideBaseProps{
		RelsMedia: []SlideRelMedia{
			{Type: "image/png", Path: srv.URL + "/pic.png", RID: 1, Target: "../media/image-1-1.png"},
		},
	}
	errs := resolveSlideMediaRels(layout)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
	want := base64.StdEncoding.EncodeToString(content)
	got, ok := layout.RelsMedia[0].Data.(string)
	if !ok || got != want {
		t.Errorf("Data = %v, want base64 %q", layout.RelsMedia[0].Data, want)
	}
}

func TestResolveSlideMediaRels_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	layout := &SlideBaseProps{
		RelsMedia: []SlideRelMedia{
			{Type: "image/png", Path: srv.URL + "/missing.png", RID: 1, Target: "../media/image-1-1.png"},
		},
	}
	errs := resolveSlideMediaRels(layout)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error for HTTP 404, got %d: %v", len(errs), errs)
	}
	got, ok := layout.RelsMedia[0].Data.(string)
	if !ok || got != IMG_BROKEN {
		t.Errorf("expected Data == IMG_BROKEN for HTTP 404, got %v", layout.RelsMedia[0].Data)
	}
}

func TestResolveSlideMediaRels_DedupeSamePath(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "pic.png")
	content := tinyPNGBytes(t)
	if err := os.WriteFile(fp, content, 0o644); err != nil {
		t.Fatalf("failed to write fixture file: %v", err)
	}

	layout := &SlideBaseProps{
		RelsMedia: []SlideRelMedia{
			{Type: "image/png", Path: fp, RID: 1, Target: "../media/image-1-1.png"},
			{Type: "image/png", Path: fp, RID: 2, Target: "../media/image-1-1.png"},
			{Type: "image/png", Path: fp, RID: 3, Target: "../media/image-1-1.png"},
		},
	}
	errs := resolveSlideMediaRels(layout)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
	want := base64.StdEncoding.EncodeToString(content)
	for i, rel := range layout.RelsMedia {
		got, ok := rel.Data.(string)
		if !ok || got != want {
			t.Errorf("rel[%d].Data = %v, want %q", i, rel.Data, want)
		}
	}
	if layout.RelsMedia[0].IsDuplicate == nil || *layout.RelsMedia[0].IsDuplicate {
		t.Errorf("rel[0] (first occurrence) should not be marked duplicate")
	}
	if layout.RelsMedia[1].IsDuplicate == nil || !*layout.RelsMedia[1].IsDuplicate {
		t.Errorf("rel[1] (repeat path) should be marked duplicate")
	}
	if layout.RelsMedia[2].IsDuplicate == nil || !*layout.RelsMedia[2].IsDuplicate {
		t.Errorf("rel[2] (repeat path) should be marked duplicate")
	}
}

func TestResolveSlideMediaRels_DedupeMissingFileOnlyErrorsOnce(t *testing.T) {
	layout := &SlideBaseProps{
		RelsMedia: []SlideRelMedia{
			{Type: "image/png", Path: "/nope/nope.png", RID: 1, Target: "t1"},
			{Type: "image/png", Path: "/nope/nope.png", RID: 2, Target: "t1"},
		},
	}
	errs := resolveSlideMediaRels(layout)
	if len(errs) != 1 {
		t.Fatalf("expected exactly 1 error (dedupe should avoid a second failed read), got %d: %v", len(errs), errs)
	}
}

func TestResolveSlideMediaRels_OnlineTypeSkipped(t *testing.T) {
	layout := &SlideBaseProps{
		RelsMedia: []SlideRelMedia{
			{Type: "online", Path: "https://www.youtube.com/embed/xyz", Data: "dummy", RID: 1, Target: "https://www.youtube.com/embed/xyz"},
		},
	}
	errs := resolveSlideMediaRels(layout)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
	if got := layout.RelsMedia[0].Data.(string); got != "dummy" {
		t.Errorf("expected online rel Data to stay 'dummy', got %q", got)
	}
}

func TestResolveSlideMediaRels_PreencodedPathSkipped(t *testing.T) {
	// A cover-image rel with no data and a "preencoded" path should be left
	// alone (matches TS candidateRels filter excluding paths containing
	// "preencoded").
	layout := &SlideBaseProps{
		RelsMedia: []SlideRelMedia{
			{Type: "image/png", Path: "preencoded.png", Data: "", RID: 1, Target: "../media/image-1-1.png"},
		},
	}
	errs := resolveSlideMediaRels(layout)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
	got, _ := layout.RelsMedia[0].Data.(string)
	if got != "" {
		t.Errorf("expected preencoded-path rel to be left untouched, got %q", got)
	}
}

func TestResolveSlideMediaRels_SvgPngPreencoded_BecomesImgBroken(t *testing.T) {
	// Mirrors the real Node behavior: encodeSlideMediaRels' STEP-5 filter
	// (rel.isSvgPng && rel.data) runs before STEP 1-4's async file reads
	// resolve, so it only ever catches rels that ALREADY carried base64 data
	// when the function was invoked (i.e. caller-supplied preencoded SVGs).
	svgData := "image/svg+xml;base64,PHN2Zy8+"
	layout := &SlideBaseProps{
		RelsMedia: []SlideRelMedia{
			{Type: "image/png", Path: "preencoded.svgpng", Data: svgData, IsSvgPng: ptr(true), RID: 1, Target: "../media/image-1-1.png"},
			{Type: "image/svg+xml", Path: "preencoded.svg", Data: svgData, RID: 2, Target: "../media/image-1-2.svg"},
		},
	}
	errs := resolveSlideMediaRels(layout)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
	if got := layout.RelsMedia[0].Data.(string); got != IMG_BROKEN {
		t.Errorf("expected isSvgPng companion Data == IMG_BROKEN, got %q", got)
	}
	if got := layout.RelsMedia[1].Data.(string); got != svgData {
		t.Errorf("expected the real SVG rel's Data to stay untouched, got %q", got)
	}
}

func TestResolveSlideMediaRels_SvgPngFromPath_NotBroken(t *testing.T) {
	// Path-based SVGs (no preencoded data) never trip the Node "SVG not
	// supported" fallback in real PptxGenJS output, because the STEP-5
	// filter runs before the async read that would populate rel.data. The
	// dedupe copy in STEP 4 (both rels share the same source `path`) means
	// the PNG companion actually ends up holding base64 of the *raw SVG
	// bytes* — a latent upstream quirk we intentionally reproduce.
	dir := t.TempDir()
	fp := filepath.Join(dir, "shape.svg")
	svgBytes := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="10" height="20"></svg>`)
	if err := os.WriteFile(fp, svgBytes, 0o644); err != nil {
		t.Fatalf("failed to write fixture file: %v", err)
	}

	layout := &SlideBaseProps{
		RelsMedia: []SlideRelMedia{
			// PNG companion is pushed first by addImageDefinition.
			{Type: "image/png", Path: fp, IsSvgPng: ptr(true), SvgSize: &SlideRelMediaSize{W: 100, H: 200}, RID: 1, Target: "../media/image-1-1.png"},
			{Type: "image/svg+xml", Path: fp, Extn: "svg", RID: 2, Target: "../media/image-1-2.svg"},
		},
	}
	errs := resolveSlideMediaRels(layout)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
	want := base64.StdEncoding.EncodeToString(svgBytes)
	pngCompanion, ok := layout.RelsMedia[0].Data.(string)
	if !ok || pngCompanion != want {
		t.Errorf("png companion Data = %v, want raw-svg base64 %q (not IMG_BROKEN)", layout.RelsMedia[0].Data, want)
	}
	if pngCompanion == IMG_BROKEN {
		t.Errorf("path-based SVG png companion must NOT become IMG_BROKEN (see comment)")
	}
	svgRel, ok := layout.RelsMedia[1].Data.(string)
	if !ok || svgRel != want {
		t.Errorf("svg rel Data = %v, want %q", layout.RelsMedia[1].Data, want)
	}
}

// ---------------------------------------------------------------------------
// getSizeFromImage
// ---------------------------------------------------------------------------

func TestGetSizeFromImage_PNG(t *testing.T) {
	size, err := getSizeFromImage(tinyPNGBytes(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if size.Width != 5 || size.Height != 5 {
		t.Errorf("size = %+v, want 5x5", size)
	}
}

// tinyJPEGBytes is a minimal, hand-built baseline JPEG (SOI, APP0, a single
// SOF0 segment declaring 4x3, EOI). It's not a *displayable* JPEG (no huffman
// tables/scan data) but it exercises the SOF0 dimension-sniffing path exactly
// like a real decoder would, without requiring image/jpeg's encoder (which
// won't let us assert exact byte-level marker parsing).
func tinyJPEGBytes() []byte {
	b := []byte{0xFF, 0xD8} // SOI
	// APP0 (JFIF) - not required for our sniffer but present in real files.
	b = append(b, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00, 0x01, 0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00)
	// SOF0: marker, length(17), precision(1), height(2), width(2), components(1), then 3 bytes per component.
	b = append(b, 0xFF, 0xC0, 0x00, 0x11, 0x08, 0x00, 0x03, 0x00, 0x04, 0x03,
		0x01, 0x11, 0x00, 0x02, 0x11, 0x01, 0x03, 0x11, 0x01)
	b = append(b, 0xFF, 0xD9) // EOI
	return b
}

func TestGetSizeFromImage_JPEG(t *testing.T) {
	size, err := getSizeFromImage(tinyJPEGBytes())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if size.Width != 4 || size.Height != 3 {
		t.Errorf("size = %+v, want 4x3", size)
	}
}

func tinyGIFBytes() []byte {
	// GIF89a header, 6x7 logical screen, minimal rest-of-file (not a fully
	// valid renderable GIF, but the header is all the sniffer reads).
	b := []byte("GIF89a")
	b = append(b, 6, 0, 7, 0) // width=6, height=7 (little-endian uint16 each)
	b = append(b, 0x00, 0x00, 0x00)
	b = append(b, 0x3B) // trailer
	return b
}

func TestGetSizeFromImage_GIF(t *testing.T) {
	size, err := getSizeFromImage(tinyGIFBytes())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if size.Width != 6 || size.Height != 7 {
		t.Errorf("size = %+v, want 6x7", size)
	}
}

func tinyBMPBytes() []byte {
	// Minimal 40-byte BITMAPINFOHEADER BMP: width=8, height=9.
	b := make([]byte, 54)
	b[0], b[1] = 'B', 'M'
	// width at offset 18, height at offset 22 (little-endian int32).
	b[18], b[19], b[20], b[21] = 8, 0, 0, 0
	b[22], b[23], b[24], b[25] = 9, 0, 0, 0
	return b
}

func TestGetSizeFromImage_BMP(t *testing.T) {
	size, err := getSizeFromImage(tinyBMPBytes())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if size.Width != 8 || size.Height != 9 {
		t.Errorf("size = %+v, want 8x9", size)
	}
}

func TestGetSizeFromImage_BMP_TopDownNegativeHeight(t *testing.T) {
	b := tinyBMPBytes()
	// height = -9 (top-down BMP): 0xFFFFFFF7 little-endian.
	b[22], b[23], b[24], b[25] = 0xF7, 0xFF, 0xFF, 0xFF
	size, err := getSizeFromImage(b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if size.Width != 8 || size.Height != 9 {
		t.Errorf("size = %+v, want 8x9 (abs of negative height)", size)
	}
}

func TestGetSizeFromImage_SVG_WidthHeightAttrs(t *testing.T) {
	svg := []byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg" width="120" height="80" viewBox="0 0 240 160"></svg>`)
	size, err := getSizeFromImage(svg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if size.Width != 120 || size.Height != 80 {
		t.Errorf("size = %+v, want 120x80 (width/height attrs take priority over viewBox)", size)
	}
}

func TestGetSizeFromImage_SVG_ViewBoxFallback(t *testing.T) {
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 300 150"></svg>`)
	size, err := getSizeFromImage(svg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if size.Width != 300 || size.Height != 150 {
		t.Errorf("size = %+v, want 300x150 from viewBox", size)
	}
}

func TestGetSizeFromImage_SVG_PercentageRejected(t *testing.T) {
	// width/height="100%" must NOT be treated as a pixel dimension; the
	// sniffer should fall back to viewBox instead.
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="100%" height="100%" viewBox="0 0 64 32"></svg>`)
	size, err := getSizeFromImage(svg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if size.Width != 64 || size.Height != 32 {
		t.Errorf("size = %+v, want 64x32 from viewBox fallback (percentages rejected)", size)
	}
}

func TestGetSizeFromImage_SVG_PercentageNoViewBox_Errors(t *testing.T) {
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="100%" height="50%"></svg>`)
	_, err := getSizeFromImage(svg)
	if err == nil {
		t.Errorf("expected an error when width/height are percentages and no viewBox is present")
	}
}

func TestGetSizeFromImage_UnknownFormat_Errors(t *testing.T) {
	_, err := getSizeFromImage([]byte("not an image"))
	if err == nil {
		t.Errorf("expected an error for unrecognized image data")
	}
}

func TestGetSizeFromImage_TooShort_Errors(t *testing.T) {
	_, err := getSizeFromImage([]byte{0x89, 0x50})
	if err == nil {
		t.Errorf("expected an error for truncated data")
	}
}

// Sanity: make sure we don't accidentally treat arbitrary text containing
// the substring "svg" as SVG data.
func TestGetSizeFromImage_PlainTextNotMistakenForSVG(t *testing.T) {
	_, err := getSizeFromImage([]byte("this data mentions svg but has no tag"))
	if err == nil {
		t.Errorf("expected an error; input has no real <svg> tag")
	}
}
