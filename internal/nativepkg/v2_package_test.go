package nativepkg

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestV2ExtractBuildUsesBindingsWithoutSourceSlideXML(t *testing.T) {
	tmp := t.TempDir()
	source := filepath.Join(tmp, "source.pptx")
	writeFixturePPTX(t, source)
	project := filepath.Join(tmp, "project")
	if err := Extract(source, project, []int{1}); err != nil {
		t.Fatal(err)
	}

	var manifest Manifest
	if err := readJSON(filepath.Join(project, "manifest.json"), &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Schema != "pptxgengo.native-scene.v2" {
		t.Fatalf("schema = %q, want v2", manifest.Schema)
	}
	if len(manifest.Slides) != 2 {
		t.Fatalf("extracted slide count = %d, want selected slide plus linked destination", len(manifest.Slides))
	}
	if _, err := os.Stat(filepath.Join(project, "resources", "ppt", "slides", "slide1.xml")); !os.IsNotExist(err) {
		t.Fatalf("source slide XML must not be retained as a build resource: %v", err)
	}

	var slide Slide
	if err := readJSON(filepath.Join(project, "slides", "uhg-001.json"), &slide); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(slide.Scene.XML()), bindingSentinelPrefix) || len(slide.Bindings) == 0 {
		t.Fatal("extracted v2 slide has no binding sentinels")
	}
	changedText, changedX, changedFont := false, false, false
	for i := range slide.Bindings {
		b := &slide.Bindings[i]
		switch {
		case b.ObjectID == "2" && b.Property == "text" && b.Value == "Original title":
			b.Value = "Bound title"
			changedText = true
		case b.ObjectID == "2" && b.Property == "transform.off.x":
			b.Value = "91440"
			changedX = true
		case b.ObjectID == "2" && b.Property == "style.rPr.rPr.sz":
			b.Value = "1800"
			changedFont = true
		}
	}
	if !changedText || !changedX || !changedFont {
		t.Fatalf("fixture bindings not found: text=%t x=%t font=%t", changedText, changedX, changedFont)
	}
	if err := writeJSON(filepath.Join(project, "slides", "uhg-001.json"), slide); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(source); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(tmp, "rebuilt.pptx")
	if err := Build(project, out, nil, false); err != nil {
		t.Fatal(err)
	}
	parts, err := readZip(out)
	if err != nil {
		t.Fatal(err)
	}
	gotSlide := string(parts["ppt/slides/slide1.xml"])
	for _, want := range []string{">Bound title<", `x="91440"`, `sz="1800"`} {
		if !strings.Contains(gotSlide, want) {
			t.Errorf("rebuilt slide missing %s\n%s", want, gotSlide)
		}
	}
	if strings.Contains(gotSlide, bindingSentinelPrefix) {
		t.Errorf("rebuilt slide retains a binding sentinel: %s", gotSlide)
	}
	if got := string(parts["ppt/slides/slide2.xml"]); !strings.Contains(got, `show="0"`) {
		t.Errorf("linked reference-only destination is not hidden: %s", got)
	}
	if got := string(parts["ppt/slides/_rels/slide1.xml.rels"]); !strings.Contains(got, `Target="slide2.xml"`) {
		t.Errorf("source internal slide link is not retained: %s", got)
	}
	if got := string(parts["ppt/presentation.xml"]); strings.Count(got, "p:sldId") < 3 { // list plus two entries
		t.Errorf("rebuilt presentation does not contain both slide IDs: %s", got)
	}
	var report BuildReport
	if err := readJSON(out+".build.json", &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Slides) != 1 || report.Slides[0] != 1 {
		t.Errorf("visible output slides = %#v, want [1]", report.Slides)
	}
	if len(report.HiddenDependencies) != 1 || report.HiddenDependencies[0] != 2 {
		t.Errorf("hidden dependencies = %#v, want [2]", report.HiddenDependencies)
	}
}

func TestV2BuildBindingFailureLeavesNoOutput(t *testing.T) {
	tmp := t.TempDir()
	source := filepath.Join(tmp, "source.pptx")
	writeFixturePPTX(t, source)
	project := filepath.Join(tmp, "project")
	if err := Extract(source, project, []int{1}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(project, "slides", "uhg-001.json")
	var slide Slide
	if err := readJSON(path, &slide); err != nil {
		t.Fatal(err)
	}
	if len(slide.Bindings) < 2 {
		t.Fatalf("fixture needs multiple bindings, got %d", len(slide.Bindings))
	}
	slide.Bindings = slide.Bindings[1:]
	if err := writeJSON(path, slide); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(tmp, "must-not-exist.pptx")
	if err := Build(project, out, nil, false); err == nil || !strings.Contains(err.Error(), "missing binding") {
		t.Fatalf("Build error = %v, want missing binding", err)
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Errorf("failed build wrote output %s: %v", out, err)
	}
	if _, err := os.Stat(out + ".build.json"); !os.IsNotExist(err) {
		t.Errorf("failed build wrote report %s.build.json: %v", out, err)
	}
}

func writeFixturePPTX(t *testing.T, file string) {
	t.Helper()
	parts := map[string]string{
		"[Content_Types].xml":  `<?xml version="1.0"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="xml" ContentType="application/xml"/></Types>`,
		"_rels/.rels":          `<?xml version="1.0"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="ppt/presentation.xml"/></Relationships>`,
		"ppt/presentation.xml": `<?xml version="1.0"?><p:presentation xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><p:sldIdLst><p:sldId id="256" r:id="rId1"/><p:sldId id="257" r:id="rId2"/></p:sldIdLst></p:presentation>`,
		"ppt/_rels/presentation.xml.rels": relationships(
			`rId1|http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide|slides/slide1.xml`,
			`rId2|http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide|slides/slide2.xml`,
		),
		"ppt/slides/slide1.xml": fixtureSlide("Original title"),
		"ppt/slides/slide2.xml": fixtureSlide("Linked destination"),
		"ppt/slides/_rels/slide1.xml.rels": relationships(
			`rId1|http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout|../slideLayouts/slideLayout1.xml`,
			`rId2|http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide|slide2.xml`,
		),
		"ppt/slides/_rels/slide2.xml.rels": relationships(
			`rId1|http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout|../slideLayouts/slideLayout1.xml`,
		),
		"ppt/slideLayouts/slideLayout1.xml": `<?xml version="1.0"?><p:sldLayout xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"/>`,
		"ppt/slideLayouts/_rels/slideLayout1.xml.rels": relationships(
			`rId1|http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideMaster|../slideMasters/slideMaster1.xml`,
		),
		"ppt/slideMasters/slideMaster1.xml": `<?xml version="1.0"?><p:sldMaster xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"/>`,
		"ppt/slideMasters/_rels/slideMaster1.xml.rels": relationships(
			`rId1|http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme|../theme/theme1.xml`,
		),
		"ppt/theme/theme1.xml": `<?xml version="1.0"?><a:theme xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" name="fixture"/>`,
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, body := range parts {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, buf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
}

func fixtureSlide(text string) string {
	return `<?xml version="1.0"?><p:sld xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"><p:cSld><p:spTree><p:sp><p:nvSpPr><p:cNvPr id="2" name="Title"/></p:nvSpPr><p:spPr><a:xfrm><a:off x="100" y="200"/><a:ext cx="300" cy="400"/></a:xfrm></p:spPr><p:txBody><a:bodyPr/><a:p><a:r><a:rPr sz="1200"/><a:t>` + text + `</a:t></a:r></a:p></p:txBody></p:sp></p:spTree></p:cSld></p:sld>`
}

func relationships(entries ...string) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	for _, entry := range entries {
		parts := strings.Split(entry, "|")
		b.WriteString(`<Relationship Id="` + parts[0] + `" Type="` + parts[1] + `" Target="` + parts[2] + `"/>`)
	}
	b.WriteString(`</Relationships>`)
	return b.String()
}
