package nativeexport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testNativePDF(pages int) []byte {
	var b bytes.Buffer
	b.WriteString("%PDF-1.4\n")
	offsets := []int{0}
	write := func(object string) {
		offsets = append(offsets, b.Len())
		fmt.Fprintf(&b, "%d 0 obj\n%s\nendobj\n", len(offsets)-1, object)
	}
	write("<< /Type /Catalog /Pages 2 0 R >>")
	kids := ""
	for i := 0; i < pages; i++ {
		kids += fmt.Sprintf("%d 0 R ", i+3)
	}
	write(fmt.Sprintf("<< /Type /Pages /Count %d /Kids [%s] >>", pages, kids))
	for i := 0; i < pages; i++ {
		write("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 960 540] >>")
	}
	xref := b.Len()
	fmt.Fprintf(&b, "xref\n0 %d\n0000000000 65535 f \n", len(offsets))
	for _, offset := range offsets[1:] {
		fmt.Fprintf(&b, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&b, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(offsets), xref)
	return b.Bytes()
}

func fakeWindowsExport(t *testing.T, mismatch bool) runner {
	t.Helper()
	return func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if name != "powershell.exe" || len(args) != 10 || args[8] != "-Config" {
			t.Fatalf("unexpected PowerShell invocation: %s %q", name, args)
		}
		raw, err := os.ReadFile(args[len(args)-1])
		if err != nil {
			return nil, err
		}
		var request windowsExportRequest
		if err = json.Unmarshal(raw, &request); err != nil {
			return nil, err
		}
		if request.Action == "close" {
			return []byte(`{"closed":true}`), nil
		}
		original, err := os.ReadFile(request.PPTX)
		if err != nil {
			return nil, err
		}
		_, count, hidden, err := reviewCopy(original, false)
		if err != nil {
			return nil, err
		}
		count -= len(hidden)
		pdfPages := count
		if mismatch {
			pdfPages++
		}
		if err = os.WriteFile(request.PDF, testNativePDF(pdfPages), 0600); err != nil {
			return nil, err
		}
		if request.PNG {
			if err = os.MkdirAll(request.PagesDirectory, 0700); err != nil {
				return nil, err
			}
			for i := 1; i <= count; i++ {
				var pngData bytes.Buffer
				if err = png.Encode(&pngData, image.NewRGBA(image.Rect(0, 0, 160, 90))); err != nil {
					return nil, err
				}
				if err = os.WriteFile(filepath.Join(request.PagesDirectory, fmt.Sprintf("slide-%03d.png", i)), pngData.Bytes(), 0600); err != nil {
					return nil, err
				}
			}
		}
		return json.Marshal(windowsExportResult{Pages: count, Version: "16.0", MissingFonts: []string{}})
	}
}

func TestWindowsNativeRenderPreservesSourceAndSlideMappings(t *testing.T) {
	for _, include := range []bool{false, true} {
		t.Run(fmt.Sprint(include), func(t *testing.T) {
			root := t.TempDir()
			source := filepath.Join(root, "Unicode-é and spaces & source.pptx")
			original := fixture(t)
			if err := os.WriteFile(source, original, 0600); err != nil {
				t.Fatal(err)
			}
			out := filepath.Join(root, "native output")
			receipt, err := render(context.Background(), Options{PPTX: source, Out: out, PDF: true, PNG: true, ContactSheet: true, IncludeHidden: include, StagingRoot: filepath.Join(root, "stage"), Timeout: time.Minute}, fakeWindowsExport(t, false), "windows")
			if err != nil {
				t.Fatal(err)
			}
			expected := 1
			if include {
				expected = 2
			}
			if receipt.Pages != expected || len(receipt.PNGs) != expected || receipt.ContactSheet == nil || !strings.Contains(receipt.Renderer, "Windows PowerPoint COM") {
				t.Fatalf("bad receipt: %+v", receipt)
			}
			if !include && receipt.PageMappings[0].SourceSlide != 2 {
				t.Fatal(receipt.PageMappings)
			}
			after, _ := os.ReadFile(source)
			if !bytes.Equal(original, after) {
				t.Fatal("source modified")
			}
			raw, _ := os.ReadFile(filepath.Join(out, "render-manifest.json"))
			if _, err = VerifyReceipt(raw); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestWindowsNativePDFMismatchIssuesNoReceipt(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source.pptx")
	if err := os.WriteFile(source, fixture(t), 0600); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(root, "out")
	_, err := render(context.Background(), Options{PPTX: source, Out: out, PDF: true, StagingRoot: filepath.Join(root, "stage"), Timeout: time.Minute}, fakeWindowsExport(t, true), "windows")
	if err == nil || !strings.Contains(err.Error(), "page count") {
		t.Fatal(err)
	}
	if _, err = os.Stat(filepath.Join(out, "render-manifest.json")); !os.IsNotExist(err) {
		t.Fatal("success receipt issued")
	}
	if _, err = os.Stat(filepath.Join(out, "render-error.txt")); err != nil {
		t.Fatal(err)
	}
}
func TestWindowsDoctorUsesActualStagingAndNativeExports(t *testing.T) {
	checks := doctor(context.Background(), DoctorOptions{StagingRoot: filepath.Join(t.TempDir(), "stage"), Timeout: time.Minute}, fakeWindowsExport(t, false), "windows")
	for _, check := range checks {
		if check.Status != "pass" {
			t.Fatal(check)
		}
	}
}
func TestWindowsPDFInspectionRejectsMalformed(t *testing.T) {
	if n, err := windowsPDFPageCount(testNativePDF(3)); err != nil || n != 3 {
		t.Fatal(n, err)
	}
	if _, err := windowsPDFPageCount([]byte("%PDF-1.4\ninvalid")); err == nil {
		t.Fatal("malformed PDF accepted")
	}
	if path := os.Getenv("PPTXGENGO_TEST_POWERPOINT_PDF"); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if n, err := windowsPDFPageCount(data); err != nil || n < 1 {
			t.Fatal("PowerPoint-produced PDF inspection failed", n, err)
		}
	}
}
