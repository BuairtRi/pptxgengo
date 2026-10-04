package nativepkg

import (
	"archive/zip"
	"bytes"
	"hash/crc32"
	"os"
	"path/filepath"
	"testing"
)

func TestInventoryReportsPresentationOrderAndCompactVisualTopology(t *testing.T) {
	source := filepath.Join(t.TempDir(), "source.pptx")
	writeInventoryFixture(t, source)
	before, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	report, err := Inventory(source)
	if err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("source PPTX changed during inventory")
	}
	if report.Schema != SourceInventorySchema || report.SourceSHA256 == "" || report.SlideCount != 2 {
		t.Fatalf("incomplete inventory: %+v", report)
	}
	first := report.Slides[0]
	if first.OriginalSlideIndex != 1 || first.Part != "ppt/slides/slide2.xml" || first.Title != "Source Page" || first.Hidden {
		t.Fatalf("source order/title/visibility incorrect: %+v", first)
	}
	if first.ShapeCount != 4 || first.TextShapeCount != 2 || first.PictureCount != 1 || first.TableCount != 1 || first.ChartCount != 1 {
		t.Fatalf("unexpected visual topology: %+v", first)
	}
	if len(first.Texts) != 2 || first.Texts[1].Text != "Body content" {
		t.Fatalf("shape text extraction failed: %+v", first.Texts)
	}
	if len(first.Tables) != 1 || len(first.Tables[0].Rows) != 2 || first.Tables[0].Rows[1][1] != "B2" {
		t.Fatalf("table contents not extracted: %+v", first.Tables)
	}
	if len(first.Charts) != 1 || first.Charts[0].Title != "Quarterly mix" || first.Charts[0].Series != 2 || len(first.Charts[0].Types) != 1 || first.Charts[0].Types[0] != "bar" {
		t.Fatalf("chart metadata not extracted: %+v", first.Charts)
	}
	if len(first.Images) != 1 || first.Images[0].Description != "Office team collaborating" || first.Images[0].Target != "ppt/media/image1.png" || first.Images[0].Bytes != len("image") {
		t.Fatalf("picture topology not extracted: %+v", first.Images)
	}
	if len(first.SpeakerNotes) != 1 || first.SpeakerNotes[0] != "Presenter note" {
		t.Fatalf("speaker notes not extracted: %+v", first.SpeakerNotes)
	}
	if second := report.Slides[1]; second.OriginalSlideIndex != 2 || !second.Hidden {
		t.Fatalf("hidden page missing or reordered: %+v", second)
	}
}

func TestInventoryRejectsMalformedPPTX(t *testing.T) {
	file := filepath.Join(t.TempDir(), "bad.pptx")
	if err := os.WriteFile(file, []byte("not a zip"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Inventory(file); err == nil {
		t.Fatal("accepted malformed package")
	}
}

func TestInventoryRejectsOversizedExpandedPartBeforeReadingIt(t *testing.T) {
	var output bytes.Buffer
	archive := zip.NewWriter(&output)
	content := []byte("x")
	entry, err := archive.CreateRaw(&zip.FileHeader{Name: "huge.xml", Method: zip.Store, CRC32: crc32.ChecksumIEEE(content), CompressedSize64: 1, UncompressedSize64: maxInventoryPartBytes + 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = entry.Write(content); err != nil {
		t.Fatal(err)
	}
	if err = archive.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = readInventoryParts(output.Bytes()); err == nil {
		t.Fatal("accepted oversized expanded package part")
	}
}

func writeInventoryFixture(t *testing.T, file string) {
	t.Helper()
	parts := map[string]string{
		"ppt/presentation.xml":             `<p:presentation xmlns:p="p" xmlns:r="r"><p:sldIdLst><p:sldId id="257" r:id="rId2"/><p:sldId id="256" r:id="rId1"/></p:sldIdLst></p:presentation>`,
		"ppt/_rels/presentation.xml.rels":  `<Relationships><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide1.xml"/><Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide2.xml"/></Relationships>`,
		"ppt/slides/slide1.xml":            `<p:sld xmlns:p="p" xmlns:a="a" show="0"><p:cSld><p:spTree><p:sp><p:nvSpPr><p:cNvPr id="1" name="Title"/><p:nvPr><p:ph type="title"/></p:nvPr></p:nvSpPr><p:spPr/><p:txBody><a:p><a:r><a:t>Hidden page</a:t></a:r></a:p></p:txBody></p:sp></p:spTree></p:cSld></p:sld>`,
		"ppt/slides/slide2.xml":            `<p:sld xmlns:p="p" xmlns:a="a" xmlns:r="r" xmlns:c="c"><p:cSld><p:spTree><p:sp><p:nvSpPr><p:cNvPr id="1" name="Title"/><p:nvPr><p:ph type="title"/></p:nvPr></p:nvSpPr><p:spPr><a:xfrm><a:off x="1" y="2"/><a:ext cx="3" cy="4"/></a:xfrm></p:spPr><p:txBody><a:p><a:r><a:t>Source Page</a:t></a:r></a:p></p:txBody></p:sp><p:sp><p:nvSpPr><p:cNvPr id="2" name="Text 1"/></p:nvSpPr><p:txBody><a:p><a:r><a:t>Body content</a:t></a:r></a:p></p:txBody></p:sp><p:pic><p:nvPicPr><p:cNvPr id="3" name="Photo" descr="Office team collaborating"/></p:nvPicPr><p:blipFill><a:blip r:embed="rImg"/></p:blipFill><p:spPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="8" cy="9"/></a:xfrm></p:spPr></p:pic><p:graphicFrame><p:nvGraphicFramePr><p:cNvPr id="4" name="Table"/></p:nvGraphicFramePr><p:xfrm><a:off x="0" y="0"/><a:ext cx="20" cy="20"/></p:xfrm><a:graphic><a:graphicData uri="table"><a:tbl><a:tr><a:tc><a:txBody><a:p><a:r><a:t>A1</a:t></a:r></a:p></a:txBody></a:tc><a:tc><a:txBody><a:p><a:r><a:t>B1</a:t></a:r></a:p></a:txBody></a:tc></a:tr><a:tr><a:tc><a:txBody><a:p><a:r><a:t>A2</a:t></a:r></a:p></a:txBody></a:tc><a:tc><a:txBody><a:p><a:r><a:t>B2</a:t></a:r></a:p></a:txBody></a:tc></a:tr></a:tbl></a:graphicData></a:graphic></p:graphicFrame><p:graphicFrame><p:nvGraphicFramePr><p:cNvPr id="5" name="Chart"/></p:nvGraphicFramePr><p:xfrm/><a:graphic><a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/chart"><c:chart r:id="rChart"/></a:graphicData></a:graphic></p:graphicFrame></p:spTree></p:cSld></p:sld>`,
		"ppt/slides/_rels/slide2.xml.rels": `<Relationships><Relationship Id="rImg" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="../media/image1.png"/><Relationship Id="rChart" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/chart" Target="../charts/chart1.xml"/><Relationship Id="rNotes" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/notesSlide" Target="../notesSlides/notesSlide1.xml"/></Relationships>`,
		"ppt/charts/chart1.xml":            `<c:chartSpace xmlns:c="c" xmlns:a="a"><c:chart><c:title><c:tx><c:rich><a:p><a:r><a:t>Quarterly mix</a:t></a:r></a:p></c:rich></c:tx></c:title><c:plotArea><c:barChart><c:ser/><c:ser/></c:barChart></c:plotArea></c:chart></c:chartSpace>`,
		"ppt/notesSlides/notesSlide1.xml":  `<p:notes xmlns:p="p" xmlns:a="a"><p:cSld><p:spTree><p:sp><p:nvSpPr><p:nvPr><p:ph type="sldNum"/></p:nvPr></p:nvSpPr><p:txBody><a:p><a:r><a:t>1</a:t></a:r></a:p></p:txBody></p:sp><p:sp><p:nvSpPr><p:nvPr><p:ph type="body"/></p:nvPr></p:nvSpPr><p:txBody><a:p><a:r><a:t>Presenter note</a:t></a:r></a:p></p:txBody></p:sp></p:spTree></p:cSld></p:notes>`,
		"ppt/media/image1.png":             "image",
	}
	var output bytes.Buffer
	archive := zip.NewWriter(&output)
	for name, content := range parts {
		entry, err := archive.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = entry.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, output.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
}
