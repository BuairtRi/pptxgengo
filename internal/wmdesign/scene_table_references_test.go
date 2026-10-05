package wmdesign

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"
)

func TestTableV6ReferenceBadgesAndPriorityEditable(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source.Revision = LibraryRevisionV6
	raw := `{"type":"table","x":57,"y":126,"w":480,"rowH":72,"cols":[{"k":"name","label":"Name","w":300},{"k":"priority","label":"Priority","w":180,"type":"priority"}],"rows":[{"name":{"text":"Platform","sub":"System of record","ref":"A1","refActive":true},"priority":{"value":"p1","label":"Critical"}},{"name":{"text":"Data","ref":"A2"},"priority":"Future"}]}`
	p := enhancementPlan(t, r, raw)
	tab := enhancementTable(t, p)
	if tab.Rows[1][0].Text != "Platform" || tab.Rows[2][0].Text != "Data" {
		t.Fatal("body copy lost")
	}
	active := enhancementShape(t, p, ".row.item-001.name.ref.fill")
	if active.Record.Color != "070154" || active.Record.Rect.H != 13 {
		t.Fatalf("active badge %+v", active.Record)
	}
	inactive := enhancementShape(t, p, ".row.item-002.name.ref.fill")
	if inactive.Props.Fill.Type != "none" {
		t.Fatal("inactive badge not outlined")
	}
	critical := enhancementShape(t, p, ".row.item-001.priority.priority.fill")
	future := enhancementShape(t, p, ".row.item-002.priority.priority.fill")
	if critical.Record.Color != "070154" || critical.Record.Rect.W < 40 || future.Props.Fill.Type != "none" {
		t.Fatal("priority palette/minimum/fallback")
	}
	want := map[string]string{".row.item-001.name.ref": "A1", ".row.item-002.name.ref": "A2", ".row.item-001.priority.priority": "Critical", ".row.item-002.priority.priority": "Future"}
	for _, it := range p.Items {
		if it.Text == nil {
			continue
		}
		for suffix, text := range want {
			if strings.HasSuffix(it.Text.ID, suffix) {
				if it.Text.Layout.Original != text {
					t.Fatalf("%s copy lost", suffix)
				}
				if suffix == ".row.item-001.name.ref" && it.Text.Color != "FFFFFF" {
					t.Fatal("active badge contrast")
				}
				delete(want, suffix)
			}
		}
	}
	if len(want) > 0 {
		t.Fatalf("editable text missing %v", want)
	}
	subFound := false
	for _, it := range p.Items {
		if it.Text != nil && strings.HasSuffix(it.Text.ID, ".row.item-001.name.sub") {
			subFound = true
			if it.Text.Layout.Original != "System of record" || it.Text.Rect.X != 69 || it.Text.Rect.W != 276 {
				t.Fatal("reference subtitle width or copy changed")
			}
		}
	}
	if !subFound {
		t.Fatal("editable subtitle missing")
	}
	doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{{ID: "references", Frame: FrameRequest{NoHeader: true}, Nodes: []Node{{ID: "table", Kind: "scene", Scene: &SceneSpec{Node: json.RawMessage(raw)}}}}}}
	data, _, err := BuildWithEngine(v6IntakeBundle(), "", doc, CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	var body []byte
	for _, f := range z.File {
		if f.Name == "ppt/slides/slide1.xml" {
			rd, e := f.Open()
			if e != nil {
				t.Fatal(e)
			}
			body, e = io.ReadAll(rd)
			rd.Close()
			if e != nil {
				t.Fatal(e)
			}
		}
	}
	for _, text := range []string{"<a:tbl>", "Platform", "System of record", "A1", "A2", "CRITICAL", "FUTURE"} {
		if !bytes.Contains(body, []byte(text)) {
			t.Errorf("editable output missing %s", text)
		}
	}
	if bytes.Contains(body, []byte("<p:pic>")) {
		t.Fatal("reference/priority rasterized")
	}
}

func TestTableV6ReferenceAndPriorityFailures(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source.Revision = LibraryRevisionV6
	st, _ := r.sceneStyle("body")
	for _, raw := range []string{`{"text":"A","refActive":true}`, `{"text":"A","ref":""}`, `{"ref":"A1"}`, `{"text":"A","ref":7}`, `{"text":"A","ref":"A1","refActive":"yes"}`, `{"text":"A","ref":"A1","typo":true}`} {
		if _, _, err := r.sceneTableReferenceCell(&scenePlan{}, "ref", json.RawMessage(raw), st, Rect{57, 126, 180, 36}, "light", SceneContext{Surface: "light"}); err == nil {
			t.Errorf("accepted %s", raw)
		}
	}
	if _, _, err := r.sceneTablePriorityCell(&scenePlan{}, "priority", json.RawMessage(`"A priority whose label cannot fit"`), st, Rect{57, 126, 70, 28}, "light", SceneContext{Surface: "light"}); err == nil {
		t.Fatal("priority overflow accepted")
	}
	r.source.Revision = LibraryRevisionV5
	if _, _, err := r.sceneTableReferenceCell(&scenePlan{}, "ref", json.RawMessage(`{"text":"A","ref":"A1"}`), st, Rect{57, 126, 180, 36}, "light", SceneContext{Surface: "light"}); err == nil {
		t.Fatal("v5 reference accepted")
	}
	if _, _, err := r.sceneTablePriorityCell(&scenePlan{}, "priority", json.RawMessage(`"P1"`), st, Rect{57, 126, 180, 36}, "light", SceneContext{Surface: "light"}); err == nil {
		t.Fatal("v5 priority accepted")
	}
}
