package nativeexport

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"strings"
	"testing"
)

func TestDoctorPresentationHasExplicitNativeGeometryAndFonts(t *testing.T) {
	data, err := doctorPresentation()
	if err != nil {
		t.Fatal(err)
	}
	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	slides, err := presentationSlides(archive)
	if err != nil || len(slides) != 1 {
		t.Fatal("diagnostic is not one presentation page", slides, err)
	}
	all := parts(t, data)
	for name, content := range all {
		if strings.HasSuffix(name, ".xml") || strings.HasSuffix(name, ".rels") {
			decoder := xml.NewDecoder(bytes.NewReader(content))
			for {
				_, err := decoder.Token()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatalf("invalid diagnostic XML %s: %v", name, err)
				}
			}
		}
	}
	for _, part := range []string{"[Content_Types].xml", "ppt/presentation.xml", "ppt/_rels/presentation.xml.rels", "ppt/slides/slide1.xml", "ppt/slides/_rels/slide1.xml.rels", "ppt/slideLayouts/slideLayout1.xml", "ppt/slideMasters/slideMaster1.xml", "ppt/theme/theme1.xml"} {
		if len(all[part]) == 0 {
			t.Errorf("diagnostic missing self-contained part %s", part)
		}
	}
	for _, want := range []string{`cx="12192000"`, `cy="6858000"`} {
		if !bytes.Contains(all["ppt/presentation.xml"], []byte(want)) {
			t.Errorf("diagnostic size missing %s", want)
		}
	}
	slide := all[slides[0]]
	for _, want := range []string{`x="914400"`, `y="914400"`, `cx="9144000"`, `cy="914400"`, `typeface="Arial"`, `sz="2400"`, `val="2800"`, `val="FFFFFF"`, `native-file-access-diagnostic`, `PowerPoint staging file-access diagnostic`} {
		if !bytes.Contains(slide, []byte(want)) {
			t.Errorf("diagnostic text geometry/style missing %s", want)
		}
	}
	if bytes.Count(all["ppt/theme/theme1.xml"], []byte(`<a:latin typeface="Arial"/>`)) != 2 {
		t.Fatal("diagnostic theme does not explicitly select Arial for heading and body")
	}
	for name := range all {
		if strings.HasPrefix(name, "ppt/media/") || strings.HasPrefix(name, "ppt/fonts/") {
			t.Fatal("diagnostic unexpectedly depends on media/embedded font assets", name)
		}
	}
	_, total, hidden, err := reviewCopy(data, false)
	if err != nil || total != 1 || len(hidden) != 0 {
		t.Fatal("diagnostic not a single visible slide", total, hidden, err)
	}
}
