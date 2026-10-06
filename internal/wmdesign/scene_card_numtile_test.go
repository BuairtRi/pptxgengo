package wmdesign

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func tileTestRenderer(t *testing.T, level string) renderer {
	t.Helper()
	s := densityTestSource(t)
	ty, err := NewSourceTypographyEngine(s, filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	return renderer{source: s, typeEngine: ty, bundle: densityTestBundle(), bodyDensity: level, headerDensity: "comfortable", densityScope: "body"}
}
func tileTestCard() sceneCardSource {
	return sceneCardSource{Type: "card", X: 57, Y: 180, W: 270, H: 180, Pad: 18, Gap: 6, Surface: "subtle", Title: "Pillar title", InlineNumber: "1", NumTile: "callout", NumInk: "unused-invalid-ignored", Body: []json.RawMessage{json.RawMessage(`{"p":"A short body."}`)}}
}
func planText(t *testing.T, p *scenePlan, suffix string) TextRecord {
	t.Helper()
	for _, item := range p.Items {
		if item.Text != nil && strings.HasSuffix(item.Text.ID, suffix) {
			return *item.Text
		}
	}
	t.Fatalf("missing text %s", suffix)
	return TextRecord{}
}
func TestCardNumberTileFixedExceptionAllDensities(t *testing.T) {
	for _, level := range []string{"comfortable", "compact", "dense"} {
		t.Run(level, func(t *testing.T) {
			r := tileTestRenderer(t, level)
			n := tileTestCard()
			before, _ := json.Marshal(n)
			p, err := r.sceneCard("tile", n, SceneContext{Surface: "light", Zone: Rect{0, 0, 960, 450}})
			if err != nil {
				t.Fatal(err)
			}
			after, _ := json.Marshal(n)
			if string(before) != string(after) {
				t.Fatal("tile changed input")
			}
			title := planText(t, p, ".title")
			num := planText(t, p, ".number")
			body := planText(t, p, ".paragraph")
			wantTile := Rect{75, 198 + (title.Layout.Style.Leading-27)/2, 27, 27}
			var tile *sceneShape
			for _, item := range p.Items {
				if item.Shape != nil && item.Shape.Record.ID == "tile.number-tile" {
					tile = item.Shape
				}
			}
			if tile == nil || tile.Record.Rect != wantTile || tile.Record.Color != "F900D3" || tile.Props.Line.Type != "none" {
				t.Fatalf("tile contract: %+v", tile)
			}
			if p.Bounds != (Rect{57, 180, 270, 180}) || title.Rect.X != 111 || title.Rect.W != 198 || num.Rect != wantTile {
				t.Fatal("outer geometry/title gap changed")
			}
			ns := num.Layout.Style
			if ns.Size != 14 || ns.Leading != 14 || ns.Family != "IBM Plex Mono" || ns.Weight != 600 || num.Color != "070154" || num.Align != "center" || num.VerticalAlign != "middle" || len(num.Layout.Lines) != 1 {
				t.Fatalf("fixed number style: %+v", num)
			}
			if body.Rect.Y != title.Rect.Y+title.Rect.H+12 {
				t.Fatalf("tile changed body flow: body%v title%v", body.Rect, title.Rect)
			}
		})
	}
}
func TestCardNumberTileClosedSourceField(t *testing.T) {
	r := tileTestRenderer(t, "compact")
	for _, mutate := range []func(*sceneCardSource){func(n *sceneCardSource) { n.NumTile = "light" }, func(n *sceneCardSource) { n.InlineNumber = "" }, func(n *sceneCardSource) { n.Title = "" }} {
		n := tileTestCard()
		mutate(&n)
		if _, err := r.sceneCard("invalid", n, SceneContext{Surface: "light"}); err == nil || !strings.Contains(err.Error(), "unsupported_num_tile") {
			t.Fatalf("invalid tile accepted: %v", err)
		}
	}
	old, err := Load(v10IntakeBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	r.source = old
	if _, err := r.sceneCard("old", tileTestCard(), SceneContext{Surface: "light"}); err == nil {
		t.Fatal("legacy source accepted new tile")
	}
	if !libraryFixedString["numTile"] {
		t.Fatal("tile exposed as editable string")
	}
}
func TestDraftReviewDensityRolesAndLegacyPreservation(t *testing.T) {
	for _, level := range []string{"comfortable", "compact", "dense"} {
		r := tileTestRenderer(t, level)
		p, err := r.planDraftReview("density", DraftReviewNote{Status: "wip", Owner: "Morgan", Due: "Fri", Updated: "Today", Notes: "Ready"})
		if err != nil {
			t.Fatal(err)
		}
		label, _ := r.bodyStyle("label")
		small, _ := r.bodyStyle("small")
		for _, suffix := range []string{".due", ".updated"} {
			if st := planText(t, p, suffix).Layout.Style; st.Size != label.Size || st.Leading != label.Leading {
				t.Fatalf("%s %s bypassed label density", level, suffix)
			}
		}
		owner := planText(t, p, ".owner").Layout.Style
		if owner.Size != small.Size || owner.Leading != small.Leading || owner.Weight != 600 {
			t.Fatal("owner bypassed small role")
		}
		status := planText(t, p, ".status").Layout.Style
		if status.Size != 8 || status.Leading != 10 {
			t.Fatal("status fixed exception changed")
		}
		chip := planText(t, p, ".legend-0.text").Layout.Style
		if chip.Size != 6.5 || chip.Leading != 9 {
			t.Fatal("chip exception changed")
		}
	}
	old := intakeTestRenderer(t)
	p, err := old.planDraftReview("legacy", DraftReviewNote{Status: "wip", Owner: "Morgan", Due: "Fri", Updated: "Today"})
	if err != nil {
		t.Fatal(err)
	}
	if planText(t, p, ".due").Layout.Style.Size != 7 || planText(t, p, ".status").Layout.Style.Size != 7 || planText(t, p, ".owner").Layout.Style.Size != 12 {
		t.Fatal("legacy note changed")
	}
}
func TestWriteCardNumberTileNativeFixture(t *testing.T) {
	out := os.Getenv("WMDS_NUMTILE_CONTROL_OUT")
	if out == "" {
		t.Skip("set WMDS_NUMTILE_CONTROL_OUT for native fixture")
	}
	doc := numberTileFixtureDocument()
	before, _ := json.Marshal(doc)
	deck, report, err := BuildWithEngine(densityTestBundle(), "", doc, CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	after, _ := json.Marshal(doc)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("fixture source mutated")
	}
	if err = os.MkdirAll(out, 0755); err != nil {
		t.Fatal(err)
	}
	for name, value := range map[string]any{"numtile.compiled.json": doc, "numtile.layout-report.json": report} {
		b, _ := json.MarshalIndent(value, "", "  ")
		if err = os.WriteFile(filepath.Join(out, name), b, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err = os.WriteFile(filepath.Join(out, "numtile.pptx"), deck, 0644); err != nil {
		t.Fatal(err)
	}
}

func numberTileFixtureDocument() Document {
	doc := densityProbeDoc()
	doc.Slides = nil
	auto := false
	for _, level := range []string{"comfortable", "compact", "dense"} {
		n := tileTestCard()
		n.NumInk = "callout"
		raw, _ := json.Marshal(n)
		doc.Slides = append(doc.Slides, SlideSpec{ID: "numtile-" + level, Title: "Fixed numeral tile", Eyebrow: "Density " + level, Density: level, AutoDensity: &auto, Frame: FrameRequest{Footer: "compact", Surface: "light", TitleLines: 1}, Nodes: []Node{{ID: "card", Kind: "scene", Scene: &SceneSpec{Node: raw}}}})
	}
	return doc
}

func TestCardNumberTileNativeXML(t *testing.T) {
	deck, _, err := BuildWithEngine(densityTestBundle(), "", numberTileFixtureDocument(), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(deck), int64(len(deck)))
	if err != nil {
		t.Fatal(err)
	}
	pages := 0
	for _, f := range zr.File {
		if !strings.HasPrefix(f.Name, "ppt/slides/slide") || !strings.HasSuffix(f.Name, ".xml") {
			continue
		}
		rd, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(rd)
		rd.Close()
		if err != nil {
			t.Fatal(err)
		}
		xml := string(data)
		for _, want := range []string{`name="card.number-tile"`, `cx="342900" cy="342900"`, `val="F900D3"`, `val="070154"`, `typeface="IBM Plex Mono SemiBold"`, `sz="1400"`, `anchor="ctr"`} {
			if !strings.Contains(xml, want) {
				t.Errorf("%s missing %s", f.Name, want)
			}
		}
		pages++
	}
	if pages != 3 {
		t.Fatalf("native tier pages %d", pages)
	}
}
