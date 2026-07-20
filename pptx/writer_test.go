// writer_test.go covers the writer.go remediation wave (REVIEW C3, M9, and the
// atomic-WriteFile minor): media errors surfaced through Write/WriteTo/WriteFile,
// a coherent compression surface across all three output methods, and an
// atomic (rename-based) WriteFile that never leaves a partial file behind.
package pptx

import (
	"archive/zip"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// REVIEW C3: media failures must surface from Write/WriteTo/WriteFile instead
// of being silently swallowed (with IMG_BROKEN substituted into the output).
// ---------------------------------------------------------------------------

func brokenImageDeck(t *testing.T) (*Presentation, string) {
	t.Helper()
	p := New()
	setGoldenMeta(p)
	missing := filepath.Join(t.TempDir(), "does-not-exist.png")
	s := p.AddSlide()
	if err := s.AddImage(&ImageProps{
		PositionProps:   PositionProps{X: cp(Inches(1)), Y: cp(Inches(1)), W: cp(Inches(2)), H: cp(Inches(2))},
		DataOrPathProps: DataOrPathProps{Path: missing},
	}); err != nil {
		t.Fatalf("AddImage: %v", err)
	}
	return p, missing
}

func TestWriteSurfacesMediaError(t *testing.T) {
	p, missing := brokenImageDeck(t)
	data, err := p.Write()
	if err == nil {
		t.Fatalf("Write: expected error for nonexistent media path, got nil (data len %d)", len(data))
	}
	if !strings.Contains(err.Error(), missing) {
		t.Errorf("Write error %q does not name the missing path %q", err.Error(), missing)
	}
	if data != nil {
		t.Errorf("Write: expected nil data on error, got %d bytes", len(data))
	}
}

func TestWriteToSurfacesMediaError(t *testing.T) {
	p, missing := brokenImageDeck(t)
	buf := &bytes.Buffer{}
	n, err := p.WriteTo(buf)
	if err == nil {
		t.Fatalf("WriteTo: expected error for nonexistent media path")
	}
	if !strings.Contains(err.Error(), missing) {
		t.Errorf("WriteTo error %q does not name the missing path %q", err.Error(), missing)
	}
	if n != 0 {
		t.Errorf("WriteTo: expected 0 bytes written on error, got %d", n)
	}
}

func TestWriteFileSurfacesMediaErrorAndLeavesNoPartialFile(t *testing.T) {
	p, missing := brokenImageDeck(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "broken.pptx")
	err := p.WriteFile(path)
	if err == nil {
		t.Fatalf("WriteFile: expected error for nonexistent media path")
	}
	if !strings.Contains(err.Error(), missing) {
		t.Errorf("WriteFile error %q does not name the missing path %q", err.Error(), missing)
	}
	entries, rerr := os.ReadDir(dir)
	if rerr != nil {
		t.Fatalf("ReadDir: %v", rerr)
	}
	if len(entries) != 0 {
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.Name()
		}
		t.Errorf("WriteFile left stray file(s) in output dir after error: %v", names)
	}
}

// A deck with only valid media must still write successfully (the error path
// must not regress the happy path).
func TestWriteValidMediaStillSucceeds(t *testing.T) {
	p := buildCase01(t)
	if _, err := p.Write(); err != nil {
		t.Fatalf("Write: unexpected error for valid deck: %v", err)
	}
}

func TestReadMediaFileRejectsNonRegularAndOversizedFiles(t *testing.T) {
	if _, err := readMediaFile(t.TempDir()); err == nil || !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("directory media error = %v", err)
	}
	path := filepath.Join(t.TempDir(), "oversized.png")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(maxMediaBytes + 1); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := readMediaFile(path); err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("oversized media error = %v", err)
	}
}

// ---------------------------------------------------------------------------
// REVIEW M9: compression must be reachable through Write/WriteTo/WriteFile.
// ---------------------------------------------------------------------------

// Write must reject more than one *WriteProps rather than silently using the
// first and discarding the rest.
func TestWriteRejectsMultipleProps(t *testing.T) {
	p := buildCase01(t)
	_, err := p.Write(&WriteProps{}, &WriteProps{})
	if err == nil {
		t.Fatalf("Write: expected error when given >1 *WriteProps")
	}
}

func TestWriteToOptsCompression(t *testing.T) {
	p := buildCase01(t)
	buf := &bytes.Buffer{}
	n, err := p.WriteToOpts(buf, &WriteProps{WriteBaseProps: WriteBaseProps{Compression: ptr(true)}})
	if err != nil {
		t.Fatalf("WriteToOpts: %v", err)
	}
	if n != int64(buf.Len()) {
		t.Errorf("WriteToOpts returned n=%d, buf has %d bytes", n, buf.Len())
	}
	assertAllDeflate(t, buf.Bytes())
	compareAgainstGolden(t, "01-basic", buf.Bytes())
}

func TestWriteFileOptsCompression(t *testing.T) {
	p := buildCase01(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "deck.pptx")
	if err := p.WriteFileOpts(path, &WriteProps{WriteBaseProps: WriteBaseProps{Compression: ptr(true)}}); err != nil {
		t.Fatalf("WriteFileOpts: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	assertAllDeflate(t, data)
	compareAgainstGolden(t, "01-basic", data)
}

// WriteFileWith remains a thin deprecated wrapper honoring compression.
func TestWriteFileWithStillHonorsCompression(t *testing.T) {
	p := buildCase01(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "deck.pptx")
	if err := p.WriteFileWith(path, true); err != nil {
		t.Fatalf("WriteFileWith: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	assertAllDeflate(t, data)
}

// assertAllDeflate parses a zip and confirms every non-directory entry uses
// DEFLATE, and that the decompressed parts still golden-match (checked by the
// caller via compareAgainstGolden — zip.Reader transparently inflates).
func assertAllDeflate(t *testing.T, data []byte) {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("zip open: %v", err)
	}
	saw := false
	for _, f := range zr.File {
		if strings.HasSuffix(f.Name, "/") {
			continue
		}
		if f.Method != zip.Deflate {
			t.Errorf("part %q not deflate-compressed (method %d)", f.Name, f.Method)
		} else {
			saw = true
		}
	}
	if !saw {
		t.Errorf("no parts found in archive")
	}
}

// ---------------------------------------------------------------------------
// Minor: atomic WriteFile — a successful write replaces an existing file, and
// a failing write (unwritable target dir) leaves no stray temp file.
// ---------------------------------------------------------------------------

func TestWriteFileAtomicReplacesExisting(t *testing.T) {
	p := buildCase01(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "deck.pptx")
	if err := os.WriteFile(path, []byte("stale placeholder content"), 0o644); err != nil {
		t.Fatalf("seed existing file: %v", err)
	}
	if err := p.WriteFile(path); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	compareAgainstGolden(t, "01-basic", data)

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 1 {
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.Name()
		}
		t.Errorf("expected exactly 1 file after atomic replace, found %v", names)
	}
}

func TestWriteFileFailureLeavesNoStrayTempFile(t *testing.T) {
	p := buildCase01(t)
	dir := t.TempDir()
	// A path inside a nonexistent subdirectory: os.Rename (and the temp file
	// create) must fail, and no ".tmp-*" file should be left behind anywhere
	// findable (specifically not in the parent dir we can inspect).
	badPath := filepath.Join(dir, "nosuchsubdir", "deck.pptx")
	err := p.WriteFile(badPath)
	if err == nil {
		t.Fatalf("WriteFile: expected error writing into nonexistent subdir")
	}
	entries, rerr := os.ReadDir(dir)
	if rerr != nil {
		t.Fatalf("ReadDir: %v", rerr)
	}
	if len(entries) != 0 {
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.Name()
		}
		t.Errorf("WriteFile left stray file(s) after failure: %v", names)
	}
}

// errors.Join must be usable to unwrap multiple aggregated media errors.
func TestBuildErrorIsJoinedAndUnwrappable(t *testing.T) {
	p := New()
	setGoldenMeta(p)
	s := p.AddSlide()
	missingA := filepath.Join(t.TempDir(), "a-missing.png")
	missingB := filepath.Join(t.TempDir(), "b-missing.png")
	must(t, s.AddImage(&ImageProps{
		PositionProps:   PositionProps{X: cp(Inches(0)), Y: cp(Inches(0)), W: cp(Inches(1)), H: cp(Inches(1))},
		DataOrPathProps: DataOrPathProps{Path: missingA},
	}))
	must(t, s.AddImage(&ImageProps{
		PositionProps:   PositionProps{X: cp(Inches(2)), Y: cp(Inches(2)), W: cp(Inches(1)), H: cp(Inches(1))},
		DataOrPathProps: DataOrPathProps{Path: missingB},
	}))
	_, err := p.Write()
	if err == nil {
		t.Fatalf("expected error")
	}
	var unwrapped []error
	if u, ok := err.(interface{ Unwrap() []error }); ok {
		unwrapped = u.Unwrap()
	}
	if len(unwrapped) != 2 {
		t.Fatalf("expected 2 joined errors, got %d: %v", len(unwrapped), err)
	}
	if !strings.Contains(err.Error(), missingA) || !strings.Contains(err.Error(), missingB) {
		t.Errorf("joined error %q missing one of the paths", err.Error())
	}
	_ = errors.Join // sanity: package uses errors.Join too
}
