package nativeexport

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/buairtri/pptxgengo/pptx"
)

func fixture(t *testing.T) []byte {
	t.Helper()
	// Build a real OPC presentation, including slide IDs/relationships, content
	// types, masters and layouts. Loose slide XML parts are not presentation pages.
	p := pptx.New()
	hidden := true
	first := p.AddSlide()
	first.PresSlide().Hidden = &hidden
	if err := first.AddText([]pptx.TextProps{{Text: "Keep exact text"}}, nil); err != nil {
		t.Fatal(err)
	}
	p.AddSlide()
	data, err := p.Write()
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func parts(t *testing.T, data []byte) map[string][]byte {
	t.Helper()
	z, e := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if e != nil {
		t.Fatal(e)
	}
	result := map[string][]byte{}
	for _, f := range z.File {
		r, e := f.Open()
		if e != nil {
			t.Fatal(e)
		}
		b, e := io.ReadAll(r)
		r.Close()
		if e != nil {
			t.Fatal(e)
		}
		result[f.Name] = b
	}
	return result
}
func TestReviewCopy(t *testing.T) {
	original := fixture(t)
	same, total, hidden, e := reviewCopy(original, false)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(same, original) || total != 2 || len(hidden) != 1 {
		t.Fatal("default copy changed source", total, hidden)
	}
	copy, total, hidden, e := reviewCopy(original, true)
	if e != nil {
		t.Fatal(e)
	}
	if total != 2 || len(hidden) != 1 {
		t.Fatal(total, hidden)
	}
	before, after := parts(t, original), parts(t, copy)
	for name, value := range before {
		expected := value
		if name == "ppt/slides/slide1.xml" {
			expected = bytes.Replace(value, []byte(` show="0"`), nil, 1)
		}
		if !bytes.Equal(expected, after[name]) {
			t.Fatalf("unexpected part edit %s", name)
		}
	}
	// Idempotent and source-independent; no second change is required.
	again, _, newHidden, e := reviewCopy(copy, true)
	if e != nil || len(newHidden) != 0 || !bytes.Equal(again, copy) {
		t.Fatal("not idempotent", e)
	}
}

func TestRemoveShowPreservesQuotedContent(t *testing.T) {
	input := []byte(`<p:sld description="retain show='0' text" x:show="keep" show = 'false' xmlns:p="p">`)
	want := []byte(`<p:sld description="retain show='0' text" x:show="keep" xmlns:p="p">`)
	if got := removeShow(input); !bytes.Equal(got, want) {
		t.Fatalf("%s", got)
	}
}
func TestRenderHermetic(t *testing.T) {
	dir := t.TempDir()
	original := fixture(t)
	source := filepath.Join(dir, "original with spaces.pptx")
	os.WriteFile(source, original, 0600)
	out := filepath.Join(dir, "new output")
	calls := 0
	fake := func(ctx context.Context, name string, args ...string) ([]byte, error) {
		calls++
		if e := ctx.Err(); e != nil {
			return nil, e
		}
		switch name {
		case "/usr/bin/osascript":
			review, e := os.ReadFile(args[1])
			if e != nil {
				return nil, e
			}
			if bytes.Contains(parts(t, review)["ppt/slides/slide1.xml"], []byte(`show="0"`)) {
				t.Fatal("hidden copy not visible")
			}
			if args[1] == source {
				t.Fatal("opened original")
			}
			return nil, os.WriteFile(args[2], []byte("%PDF-1.7\nnative fixture"), 0600)
		case "/usr/bin/swift":
			os.MkdirAll(args[2], 0755)
			for _, name := range []string{"slide-001.png", "slide-002.png"} {
				f, e := os.Create(filepath.Join(args[2], name))
				if e != nil {
					return nil, e
				}
				e = png.Encode(f, image.NewRGBA(image.Rect(0, 0, 32, 18)))
				f.Close()
				if e != nil {
					return nil, e
				}
			}
			return []byte(`{"pages":2}`), nil
		}
		t.Fatal(name)
		return nil, nil
	}
	receipt, e := render(context.Background(), Options{PPTX: source, Out: out, PDF: true, PNG: true, IncludeHidden: true, Timeout: time.Minute, StagingRoot: filepath.Join(dir, "stage")}, fake, "darwin")
	if e != nil {
		t.Fatal(e)
	}
	if calls != 2 || receipt.Pages != 2 || len(receipt.PNGs) != 2 || receipt.PNGs[0].Width != 32 || receipt.PDF == nil {
		t.Fatal(receipt, calls)
	}
	if receipt.Source.SHA256 != hash(original) || receipt.ReviewCopySHA256 == hash(original) {
		t.Fatal("wrong provenance")
	}
	current, _ := os.ReadFile(source)
	if !bytes.Equal(original, current) {
		t.Fatal("original changed")
	}
	names, _ := os.ReadDir(out)
	for _, n := range names {
		if strings.HasPrefix(n.Name(), ".native-work") {
			t.Fatal("temporary copy retained")
		}
	}
	if _, e = os.Stat(filepath.Join(out, "render-manifest.json")); e != nil {
		t.Fatal(e)
	}
	issued, e := os.ReadFile(filepath.Join(out, "render-manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = VerifyReceipt(issued); e != nil {
		t.Fatal("successful export did not issue verifiable provenance", e)
	}
	if _, e = render(context.Background(), Options{PPTX: source, Out: out, PDF: true, Timeout: time.Minute}, fake, "darwin"); e == nil {
		t.Fatal("overwrote existing output")
	}
}
func TestRenderFailures(t *testing.T) {
	source := filepath.Join(t.TempDir(), "original.pptx")
	os.WriteFile(source, fixture(t), 0600)
	tests := []struct {
		name, platform string
		run            runner
		want           string
	}{
		{"unsupported", "linux", nil, "requires macOS"},
		{"native-error", "darwin", func(context.Context, string, ...string) ([]byte, error) { return nil, errors.New("automation denied") }, "PowerPoint PDF export failed"},
		{"missing-pdf", "darwin", func(context.Context, string, ...string) ([]byte, error) { return nil, nil }, "without a PDF"},
		{"bad-pdf", "darwin", func(_ context.Context, _ string, args ...string) ([]byte, error) {
			return nil, os.WriteFile(args[2], []byte("bad"), 0600)
		}, "not a PDF"},
		{"timeout", "darwin", func(ctx context.Context, _ string, _ ...string) ([]byte, error) { <-ctx.Done(); return nil, ctx.Err() }, "deadline exceeded"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), "out")
			_, e := render(context.Background(), Options{PPTX: source, Out: out, PDF: true, Timeout: time.Millisecond, StagingRoot: filepath.Join(t.TempDir(), "stage")}, test.run, test.platform)
			if e == nil || !strings.Contains(e.Error(), test.want) {
				t.Fatal(e)
			}
			if _, e = os.Stat(filepath.Join(out, "render-manifest.json")); !os.IsNotExist(e) {
				t.Fatal("failure published success")
			}
		})
	}
}

func TestNativeCleanupUsesExactTaskPath(t *testing.T) {
	source := filepath.Join(t.TempDir(), "original.pptx")
	original := fixture(t)
	if err := os.WriteFile(source, original, 0600); err != nil {
		t.Fatal(err)
	}
	var exportArgs []string
	calls := 0
	fake := func(_ context.Context, name string, args ...string) ([]byte, error) {
		calls++
		if name != "/usr/bin/osascript" {
			t.Fatalf("unexpected executable %s", name)
		}
		if calls == 1 {
			exportArgs = append([]string(nil), args...)
			if len(args) != 5 || args[1] == source {
				t.Fatalf("invalid export arguments %v", args)
			}
			return nil, errors.New("automation denied")
		}
		if calls != 2 || len(args) != 6 || args[5] != "close" || args[4] != "3" {
			t.Fatalf("invalid cleanup arguments %v", args)
		}
		for i := 0; i < 4; i++ {
			if args[i] != exportArgs[i] {
				t.Fatalf("cleanup selected a different task: %v", args)
			}
		}
		return nil, nil
	}
	_, err := render(context.Background(), Options{PPTX: source, Out: filepath.Join(t.TempDir(), "out"), PDF: true, Timeout: time.Minute, StagingRoot: filepath.Join(t.TempDir(), "stage")}, fake, "darwin")
	if err == nil || calls != 2 {
		t.Fatalf("expected failed export and exact-path cleanup: %v, %d calls", err, calls)
	}
	after, err := os.ReadFile(source)
	if err != nil || !bytes.Equal(original, after) {
		t.Fatal("source changed", err)
	}
	// Both export and best-effort cleanup share the path guard. PowerPoint's
	// display-name extension and active-document state cannot select a deck.
	script := string(exportScript)
	for _, required := range []string{"my fileIdentity(sourceFile)", "my fileIdentity(full name of candidate) is expectedIdentity", "if matchCount is not 1 then error", "if closeOnly and matchCount is 0 then return", "NSFileSystemFileNumber", "NSFileSystemNumber", "stringByResolvingSymlinksInPath", "repeat with attempt from 1 to 40", "with timeout of 12 seconds"} {
		if !strings.Contains(script, required) {
			t.Fatalf("missing identity guard %q", required)
		}
	}
	for _, unsafe := range []string{"active presentation", "presentation taskName"} {
		if strings.Contains(script, unsafe) {
			t.Fatalf("unsafe selection %q", unsafe)
		}
	}
}
