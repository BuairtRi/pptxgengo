package nativeexport

import (
	"archive/zip"
	"bytes"
	"reflect"
	"regexp"
	"sort"
	"testing"
)

func packParts(t *testing.T, source map[string][]byte) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	names := []string{}
	for name := range source {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		part, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write(source[name]); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func TestReviewCopyFollowsPresentationOrderAndIgnoresOrphanSlides(t *testing.T) {
	source := parts(t, fixture(t))
	const renamed = "ppt/slides/customer-hidden.xml"
	source["ppt/slides/slide999.xml"] = append([]byte(nil), source["ppt/slides/slide1.xml"]...)
	source[renamed] = source["ppt/slides/slide1.xml"]
	delete(source, "ppt/slides/slide1.xml")
	source["ppt/slides/_rels/customer-hidden.xml.rels"] = source["ppt/slides/_rels/slide1.xml.rels"]
	delete(source, "ppt/slides/_rels/slide1.xml.rels")
	source["[Content_Types].xml"] = bytes.ReplaceAll(source["[Content_Types].xml"], []byte("/ppt/slides/slide1.xml"), []byte("/"+renamed))
	// An unreferenced, typed slide part is a valid package part, not an output page.
	source["[Content_Types].xml"] = bytes.Replace(source["[Content_Types].xml"], []byte("</Types>"), []byte(`<Override PartName="/ppt/slides/slide999.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slide+xml"/></Types>`), 1)
	source["ppt/_rels/presentation.xml.rels"] = bytes.ReplaceAll(source["ppt/_rels/presentation.xml.rels"], []byte("slides/slide1.xml"), []byte("slides/customer-hidden.xml"))
	// Mark both referenced pages hidden, then reverse presentation order. ZIP
	// order and numeric part names must not control hidden page accounting.
	source["ppt/slides/slide2.xml"] = bytes.Replace(source["ppt/slides/slide2.xml"], []byte("<p:sld "), []byte(`<p:sld show="false" `), 1)
	ids := regexp.MustCompile(`<p:sldId\b[^>]*/>`).FindAll(source["ppt/presentation.xml"], -1)
	if len(ids) != 2 {
		t.Fatal("fixture does not have two actual presentation slide IDs")
	}
	list := regexp.MustCompile(`(?s)<p:sldIdLst>.*?</p:sldIdLst>`)
	newList := append(append(append([]byte("<p:sldIdLst>"), ids[1]...), ids[0]...), []byte("</p:sldIdLst>")...)
	source["ppt/presentation.xml"] = list.ReplaceAllLiteral(source["ppt/presentation.xml"], newList)
	original := packParts(t, source)
	unchanged, count, hidden, err := reviewCopy(original, false)
	if err != nil || !bytes.Equal(unchanged, original) || count != 2 || !reflect.DeepEqual(hidden, []string{"ppt/slides/slide2.xml", renamed}) {
		t.Fatalf("wrong actual order/count/hidden/source: %d %+v %v", count, hidden, err)
	}
	visible, count, hidden, err := reviewCopy(original, true)
	if err != nil || count != 2 || !reflect.DeepEqual(hidden, []string{"ppt/slides/slide2.xml", renamed}) {
		t.Fatalf("review copy has wrong actual pages: %d %+v %v", count, hidden, err)
	}
	result := parts(t, visible)
	for name, data := range source {
		if name == renamed || name == "ppt/slides/slide2.xml" {
			if bytes.Contains(result[name], []byte("show=")) {
				t.Errorf("referenced hidden slide was not made visible: %s", name)
			}
		} else if !bytes.Equal(data, result[name]) {
			t.Errorf("non-slide/orphan part changed: %s", name)
		}
	}
}

func TestReviewCopyRejectsMissingPresentationSlideRelationship(t *testing.T) {
	source := parts(t, fixture(t))
	delete(source, "ppt/_rels/presentation.xml.rels")
	if _, _, _, err := reviewCopy(packParts(t, source), true); err == nil {
		t.Fatal("missing actual slide relationships accepted")
	}
}
