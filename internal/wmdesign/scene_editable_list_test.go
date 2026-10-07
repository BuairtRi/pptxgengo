package wmdesign

import (
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"testing"
)

func TestNativeEditingEditableListOneTextBoxAndParagraphStyles(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source = densityTestSource(t)
	var e error
	r.typeEngine, e = NewSourceTypographyEngine(r.source, filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	for _, density := range []string{"comfortable", "compact", "dense"} {
		r.bodyDensity = density
		for _, size := range []string{"body", "small"} {
			raw, _ := json.Marshal(map[string]any{"type": "editable-list", "x": 40, "y": 100, "w": 260, "h": 230, "size": size, "items": []string{"Short first bullet", "This longer bullet should wrap naturally inside the same text box while the third bullet follows it without requiring an individual shape move.", "Final bullet"}})
			p, e := r.planEditableList("list", raw, SceneContext{Surface: "light", Keys: map[string][]string{"/items": {"first", "second", "third"}}})
			if e != nil {
				t.Fatal(density, size, e)
			}
			if len(p.Items) != 1 || len(p.Groups) != 0 || p.Items[0].Text == nil {
				t.Fatal("list is not one ungrouped text box", p)
			}
			tr := p.Items[0].Text
			if tr.NativeParagraphContract != EditableListContract || len(tr.Rich.Paragraphs) != 3 || tr.Rect != (Rect{40, 100, 260, 230}) {
				t.Fatal(tr)
			}
			if tr.Rich.Paragraphs[1].LineCount < 2 || tr.Rich.Paragraphs[2].FirstLine <= tr.Rich.Paragraphs[1].FirstLine+1 {
				t.Fatal("wrapped paragraph did not move following baseline", tr.Rich)
			}
			xml := string(editableListParagraphXML(*tr))
			if strings.Count(xml, "<a:p>") != 3 || strings.Count(xml, `<a:buChar char="■"/>`) != 3 || strings.Count(xml, "<a:buClr>") != 3 || strings.Contains(xml, "<a:br") {
				t.Fatal(xml)
			}
			for i, key := range []string{"first", "second", "third"} {
				if tr.Rich.Paragraphs[i].Key != key || !tr.Rich.Paragraphs[i].Bullet || tr.Rich.Paragraphs[i].BulletIndentPt <= 0 || tr.Rich.Paragraphs[i].BulletMarkerPt <= 0 {
					t.Fatal(tr.Rich)
				}
				para := tr.Rich.Paragraphs[i]
				properties := string(sceneTableParagraphProperties(*tr, para.ParagraphGapAfter, true, para))
				indent := int(math.Round(para.BulletIndentPt * 12700))
				for _, want := range []string{
					fmt.Sprintf(`marL="%d" indent="-%d"`, indent, indent),
					fmt.Sprintf(`<a:buSzPts val="%d"/>`, int(math.Round(para.BulletMarkerPt*100))),
					fmt.Sprintf(`<a:buFont typeface="%s"/>`, tr.Layout.Font.Typeface),
					fmt.Sprintf(`<a:spcAft><a:spcPts val="%d"/></a:spcAft>`, int(math.Round(para.ParagraphGapAfter*100))),
					fmt.Sprintf(`<a:lnSpc><a:spcPts val="%d"/></a:lnSpc>`, int(math.Round(tr.Layout.Style.Leading*100))),
				} {
					if !strings.Contains(properties, want) {
						t.Fatal("native list metrics changed", want, properties)
					}
				}
			}
		}
	}
}

func TestNativeEditingEditableListRefusesUnsupportedAndOverflow(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source = densityTestSource(t)
	var e error
	r.typeEngine, e = NewSourceTypographyEngine(r.source, filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	for _, raw := range []string{
		`{"type":"editable-list","w":300,"h":200,"items":[]}`,
		`{"type":"editable-list","w":300,"h":200,"items":[""]}`,
		`{"type":"editable-list","w":300,"h":200,"items":["Two\nparagraphs"]}`,
		`{"type":"editable-list","w":300,"h":200,"items":["[[Rich]]"]}`,
		`{"type":"editable-list","w":300,"h":200,"items":[{"lead":"Lead","text":"Body"}]}`,
		`{"type":"editable-list","w":300,"h":200,"size":"heading","items":["Copy"]}`,
		`{"type":"editable-list","w":300,"h":1,"items":["Copy"]}`,
		`{"type":"editable-list","w":5,"h":200,"items":["Copy"]}`,
		`{"type":"editable-list","w":300,"h":200,"ordered":true,"items":["Copy"]}`,
	} {
		if _, e := r.planEditableList("bad", json.RawMessage(raw), SceneContext{Surface: "light"}); e == nil {
			t.Fatal("unsupported list accepted", raw)
		}
	}
	raw := json.RawMessage(`{"type":"editable-list","w":300,"h":200,"items":["One","Two"]}`)
	if _, e := r.planEditableList("bad", raw, SceneContext{Keys: map[string][]string{"/items": {"same", "same"}}}); e == nil {
		t.Fatal("duplicate keys accepted")
	}
}
