package deckproject

import (
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func xmlShape(t *testing.T, raw string) *xmlNode {
	t.Helper()
	tree, err := readXML(strings.NewReader(raw))
	if err != nil || len(tree.Children) != 1 {
		t.Fatal(tree, err)
	}
	return tree.Children[0]
}

func TestNativeParagraphRunAndBreakAddresses(t *testing.T) {
	shape := xmlShape(t, `<p:sp xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"><p:txBody><a:p><a:r><a:rPr b="1"/><a:t>Hello </a:t></a:r><a:r><a:t>world</a:t></a:r><a:br/><a:r><a:t>next line</a:t></a:r></a:p><a:p/><a:p><a:r><a:t>last</a:t></a:r></a:p></p:txBody></p:sp>`)
	paragraphs := nativeParagraphs(shape)
	if len(paragraphs) != 3 || paragraphs[0].Text != "Hello world\nnext line" || len(paragraphs[0].Runs) != 4 || paragraphs[0].Runs[2].Kind != "br" || paragraphs[1].Text != "" {
		t.Fatal(paragraphs)
	}
	if nativeParagraphText(paragraphs) != "Hello world\nnext line\n\nlast" {
		t.Fatal("paragraph boundaries lost")
	}
	if paragraphs[0].Address.Body != 0 || paragraphs[2].Address.Paragraph != 2 || paragraphs[0].Runs[3].Ordinal != 3 {
		t.Fatal("addresses drifted", paragraphs)
	}
	if paragraphs[0].Runs[0].PropertiesSHA256 == paragraphs[0].Runs[1].PropertiesSHA256 {
		t.Fatal("rich styles flattened")
	}
}

func TestNativeTableCellParagraphAddresses(t *testing.T) {
	shape := xmlShape(t, `<p:graphicFrame xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"><a:graphic><a:graphicData><a:tbl><a:tr><a:tc><a:txBody><a:p><a:r><a:t>A</a:t></a:r></a:p></a:txBody></a:tc><a:tc><a:txBody><a:p><a:r><a:t>B</a:t></a:r></a:p><a:p><a:r><a:t>second</a:t></a:r></a:p></a:txBody></a:tc></a:tr><a:tr><a:tc><a:txBody><a:p><a:r><a:t>C</a:t></a:r></a:p></a:txBody></a:tc></a:tr></a:tbl></a:graphicData></a:graphic></p:graphicFrame>`)
	paragraphs := nativeParagraphs(shape)
	if len(paragraphs) != 4 || *paragraphs[1].Address.TableRow != 0 || *paragraphs[1].Address.TableColumn != 1 || paragraphs[2].Address.Paragraph != 1 || *paragraphs[3].Address.TableRow != 1 || paragraphs[3].Address.Body != 2 {
		t.Fatal(paragraphs)
	}
}

func TestNativeDynamicAndBulletTextRemainReviewItems(t *testing.T) {
	shape := xmlShape(t, `<p:sp xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"><p:txBody><a:p><a:pPr><a:buChar char="•"/></a:pPr><a:fld id="dynamic"><a:t>2026</a:t></a:fld></a:p></p:txBody></p:sp>`)
	paragraphs := nativeParagraphs(shape)
	if len(paragraphs) != 1 || len(paragraphs[0].ReviewItems) != 2 || paragraphs[0].Runs[0].Kind != "fld" {
		t.Fatal(paragraphs)
	}
}

func TestNativeFieldRequiresUniqueExactSource(t *testing.T) {
	p := example(t)
	r := ObjectRecord{LogicalID: "fixture/title", SourceSlots: map[string]string{"/slides/0/values/title": "title"}, NativeKind: "sp", NativeText: p.Document.Slides[0].Values["title"].(string), SourcePointers: []string{"/slides/0/values/title"}, Paragraphs: []NativeParagraph{{Address: NativeTextAddress{}, Runs: []NativeTextRun{{Kind: "r"}}}}}
	exact := attachNativeSourceFields(p, r)
	if exact.TextMapping != "plain_text_baseline" || len(exact.Fields) != 1 || exact.Fields[0].SourceValueSHA256 != exact.Fields[0].BaselineNativeSHA256 {
		t.Fatal(exact)
	}
	r.SourcePointers = append(r.SourcePointers, "/slides/0/values/eyebrow")
	if ambiguous := attachNativeSourceFields(p, r); ambiguous.TextMapping != "manual_review" || len(ambiguous.Fields) != 0 {
		t.Fatal("ambiguous source inferred", ambiguous)
	}
	r.SourcePointers = r.SourcePointers[:1]
	r.NativeText = "Source\nwas visually wrapped"
	if wrapped := attachNativeSourceFields(p, r); wrapped.TextMapping != "manual_review" {
		t.Fatal("layout breaks silently normalized", wrapped)
	}
	r.NativeText = p.Document.Slides[0].Values["title"].(string)
	r.Paragraphs[0].Runs = append(r.Paragraphs[0].Runs, NativeTextRun{Kind: "r"})
	if rich := attachNativeSourceFields(p, r); rich.TextMapping != "manual_review" {
		t.Fatal("multiple rich runs treated as one source field", rich)
	}
}

func TestTypedCardNativeObjectsMapOneActualField(t *testing.T) {
	p := example(t)
	compiled, err := Compile(p, bundle(t), wmdesign.CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	native, _, err := wmdesign.BuildWithEngineAndAssets(bundle(t), "", compiled.Document, wmdesign.CandidateEngine, compiled.Assets)
	if err != nil {
		t.Fatal(err)
	}
	mapped, err := ObjectMap(p, compiled.Document, native)
	if err != nil {
		t.Fatal(err)
	}
	if mapped.TextModelSchema != NativeTextModelSchema {
		t.Fatal(mapped)
	}
	found := map[string]bool{}
	for _, object := range mapped.Objects {
		if object.SlideID != "maintain-the-source" {
			continue
		}
		if object.NativeName == "cards.items.source.title" || object.NativeName == "cards.items.source.body.copy" {
			t.Logf("%s %v text=%q status=%s", object.NativeName, object.SourcePointers, object.NativeText, object.TextMapping)
			if len(object.SourcePointers) != 1 || len(object.Fields) != 1 || object.ItemKey != "source" {
				t.Fatal("card field remained ambiguous", object)
			}
			expected := "/slides/0/values/cards/0/title"
			if strings.HasSuffix(object.NativeName, "body.copy") {
				expected = "/slides/0/values/cards/0/body"
			}
			if object.Fields[0].Identity == "" || object.Fields[0].SourceSlot == "" {
				t.Fatal("stable source identity missing", object)
			}
			if object.Fields[0].SourcePointer != expected {
				t.Fatal("wrong card field", object)
			}
			found[object.NativeName] = true
		}
	}
	if len(found) != 2 {
		t.Fatal("expected representative card title/body", found)
	}
}

func TestNativeStructureHashSeparatesTextAndGeometry(t *testing.T) {
	first := xmlShape(t, `<p:sp xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"><p:spPr><a:xfrm><a:off x="1" y="2"/></a:xfrm></p:spPr><p:txBody><a:p><a:r><a:t>Original</a:t></a:r></a:p></p:txBody></p:sp>`)
	second := xmlShape(t, `<p:sp xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"><p:spPr><a:xfrm><a:off x="1" y="2"/></a:xfrm></p:spPr><p:txBody><a:p><a:r><a:t>Edited</a:t></a:r></a:p></p:txBody></p:sp>`)
	if nativeStructureHash(first) != nativeStructureHash(second) {
		t.Fatal("text affected structural fingerprint")
	}
	changed := xmlShape(t, `<p:sp xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"><p:spPr><a:xfrm><a:off x="9" y="2"/></a:xfrm></p:spPr><p:txBody><a:p><a:r><a:t>Edited</a:t></a:r></a:p></p:txBody></p:sp>`)
	if nativeStructureHash(first) == nativeStructureHash(changed) {
		t.Fatal("geometry change was hidden")
	}
}
