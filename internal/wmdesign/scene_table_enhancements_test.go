package wmdesign

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/pptx"
)

func enhancementPlan(t *testing.T, r *renderer, raw string) *scenePlan {
	t.Helper()
	p, e := r.planSceneNode("sample", json.RawMessage(raw), SceneContext{Surface: "light", Path: "/body/0"})
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func enhancementShape(t *testing.T, p *scenePlan, suffix string) *sceneShape {
	t.Helper()
	for _, it := range p.Items {
		if it.Shape != nil && strings.HasSuffix(it.Shape.Record.ID, suffix) {
			return it.Shape
		}
	}
	t.Fatalf("shape %s missing", suffix)
	return nil
}
func enhancementTable(t *testing.T, p *scenePlan) *sceneTable {
	t.Helper()
	for _, it := range p.Items {
		if it.Table != nil {
			return it.Table
		}
	}
	t.Fatal("native table missing")
	return nil
}

func TestTableHeatFrozenRampAndPrecedence(t *testing.T) {
	r := intakeTestRenderer(t)
	p := enhancementPlan(t, r, `{"type":"table","x":117,"y":126,"w":360,"rowH":48,"heatMax":8,"cols":[{"k":"a","label":"Scale","w":120,"type":"heat","showValue":true},{"k":"b","label":"Risk","w":120,"type":"heat","scale":"risk"},{"k":"c","label":"Word","w":120,"type":"heat","scale":"risk"}],"rows":[{"scale":"risk","a":{"value":6,"scale":"seq"},"b":8,"c":{"value":6,"text":"Ready","scale":"seq"}}]}`)
	tab := enhancementTable(t, p)
	if tab.Rows[1][0].Text != "6" || tab.Rows[1][2].Text != "Ready" {
		t.Fatal("heat labels lost")
	}
	if tab.Rows[1][0].Options.Fill.Transparency != 100 || tab.Rows[1][0].Options.Color != "FFFFFF" {
		t.Fatal("native text opaque or wrong contrast")
	}
	for _, tc := range tab.CellRecords {
		if tc.Text.ID == "sample.row.slot-001.a" && tc.Text.Layout.Font.Family != "IBM Plex Mono" {
			t.Fatal("showValue not mono")
		}
	}
	for _, want := range []struct{ suffix, color string }{{".a.heat", "0047FF"}, {".b.heat", "F900D3"}, {".c.heat", "0047FF"}} {
		sh := enhancementShape(t, p, want.suffix)
		if sh.Record.Color != want.color {
			t.Fatalf("%s=%s", want.suffix, sh.Record.Color)
		}
		if sh.Record.Rect.W != 117 || sh.Record.Rect.H != 45 {
			t.Fatal("heat inset not1.5pt")
		}
	}
	index := -1
	for i, it := range p.Items {
		if it.Table != nil {
			index = i
		}
		if it.Shape != nil && strings.HasSuffix(it.Shape.Record.ID, ".heat") && index >= 0 {
			t.Fatal("heat paints above native text")
		}
	}
}

func TestTablePlainSubAndScoreInks(t *testing.T) {
	r := intakeTestRenderer(t)
	p := enhancementPlan(t, r, `{"type":"table","x":117,"y":126,"w":600,"rowH":96,"rowHeader":true,"cols":[{"k":"name","label":"Name","w":180},{"k":"rating","label":"Rating","w":120,"type":"rating","ink":"series.2","max":5},{"k":"harvey","label":"Progress","w":120,"type":"harvey","ink":"series.2"},{"k":"dots","label":"Dots","w":180,"type":"dots","ink":"series.2"}],"rows":[{"ink":"callout","name":{"text":"Platform","sub":"System of record"},"rating":{"value":3,"ink":"emphasis"},"harvey":2,"dots":{"value":2,"max":4,"text":["Clear owner","Next action"]}}]}`)
	tab := enhancementTable(t, p)
	var plain, dots *TextRecord
	for i := range tab.CellRecords {
		tr := &tab.CellRecords[i].Text
		if strings.HasSuffix(tr.ID, ".name") && strings.Contains(tr.ID, ".row.") {
			plain = tr
		}
		if strings.HasSuffix(tr.ID, ".dots") && strings.Contains(tr.ID, ".row.") {
			dots = tr
		}
	}
	if plain == nil || plain.Rich == nil || len(plain.Rich.Paragraphs) != 2 {
		t.Fatal("plain cell is not two native paragraphs")
	}
	if plain.Rich.Paragraphs[0].Runs[0].Style.Weight != 600 || plain.Rich.Paragraphs[1].Runs[0].Style.Weight != 400 || plain.Rich.Paragraphs[0].ParagraphGapAfter != 2 {
		t.Fatal("subtitle style/gap lost")
	}
	if plain.Layout.Displayed != "Platform\nSystem of record" {
		t.Fatal("subtitle text lost")
	}
	emphasis, _ := r.sceneColor("light", "emphasis")
	callout, _ := r.sceneColor("light", "callout")
	if enhancementShape(t, p, ".rating-1").Record.Color != emphasis || enhancementShape(t, p, ".harvey-fill").Record.Color != callout || enhancementShape(t, p, ".dots-1").Record.Color != callout {
		t.Fatal("cell > row > column ink precedence lost")
	}
	if enhancementShape(t, p, ".dots-4").Record.Color != "CED7E6" {
		t.Fatal("empty dot color")
	}
	if dots == nil || dots.Rich == nil || len(dots.Rich.Paragraphs) != 2 || !dots.Rich.Paragraphs[0].Bullet {
		t.Fatal("dot-copy bullet paragraphs missing")
	}
	if dots.Rect.Y < enhancementShape(t, p, ".dots-1").Record.Rect.Y+14-.02 {
		t.Fatal("dot copy intersects marks")
	}
	// Explicit caller identities retain nested dot-copy keys.
	raw := json.RawMessage(`{"type":"table","x":117,"y":126,"w":180,"rowH":96,"cols":[{"k":"score","label":"Dots","w":180,"type":"dots"}],"rows":[{"score":{"value":2,"text":["One","Two"]}}]}`)
	_, e := r.planSceneNode("keyed", raw, SceneContext{Surface: "light", Path: "/body/0", Keys: map[string][]string{"/body/0/rows": {"person"}, "/body/0/rows/0/score/text": {"owner", "action"}}})
	if e != nil {
		t.Fatal(e)
	}
}

func TestHeatBlocksAndLegendSwatches(t *testing.T) {
	r := intakeTestRenderer(t)
	p := enhancementPlan(t, r, `{"type":"block","x":117,"y":126,"w":180,"h":54,"heat":4,"text":"High","style":"small"}`)
	if enhancementShape(t, p, ".surface").Record.Color != "070154" {
		t.Fatal("block heat fill")
	}
	for _, it := range p.Items {
		if it.Text != nil && it.Text.Color != "FFFFFF" {
			t.Fatal("block heat contrast")
		}
	}
	p = enhancementPlan(t, r, `{"type":"legend","x":117,"y":126,"w":500,"layout":"horizontal","items":[{"text":"Low","heat":0},{"text":"Hot","heat":4,"heatScale":"risk"},{"text":"Client","swatch":"dot","ink":"callout"}]}`)
	if enhancementShape(t, p, "slot-001.swatch").Record.Rect.W != 18 || enhancementShape(t, p, "slot-001.swatch").Record.Rect.H != 10 {
		t.Fatal("heat legend dimensions")
	}
	if enhancementShape(t, p, "slot-002.swatch").Record.Color != "F900D3" {
		t.Fatal("risk legend palette")
	}
	dot := enhancementShape(t, p, "slot-003.swatch")
	if dot.Type != pptx.ShapeTypeEllipse || dot.Record.Rect.W != 8 || dot.Props.Line.Width != 0 {
		t.Fatal("dot swatch not plain8pt circle")
	}
	// Existing compact label padding stays3pt on18pt badges.
	p = enhancementPlan(t, r, `{"type":"block","x":117,"y":126,"w":18,"h":18,"surface":"light","text":"1","style":"label"}`)
	for _, it := range p.Items {
		if it.Text != nil && it.Text.Rect.W != 12 {
			t.Fatal("compact badge padding changed")
		}
	}
}

func TestTableEnhancementValidation(t *testing.T) {
	r := intakeTestRenderer(t)
	for _, raw := range []string{
		`{"type":"table","x":117,"y":126,"w":180,"cols":[{"k":"x","label":"X","w":180,"type":"heat","max":0}],"rows":[{"x":1}]}`,
		`{"type":"table","x":117,"y":126,"w":180,"cols":[{"k":"x","label":"X","w":180,"type":"heat"}],"rows":[{"x":{"value":2,"scale":"invented"}}]}`,
		`{"type":"table","x":117,"y":126,"w":180,"cols":[{"k":"x","label":"X","w":180,"type":"dots"}],"rows":[{"x":{"value":1.5}}]}`,
		`{"type":"table","x":117,"y":126,"w":180,"cols":[{"k":"x","label":"X","w":180,"type":"rating"}],"rows":[{"x":{"value":1,"max":1000000}}]}`,
		`{"type":"table","x":117,"y":126,"w":180,"cols":[{"k":"x","label":"X","w":180,"type":"harvey"}],"rows":[{"x":{"value":1,"max":5}}]}`,
		`{"type":"table","x":117,"y":126,"w":180,"cols":[{"k":"x","label":"X","w":180}],"rows":[{"x":{"text":"Text","background":"inverse"}}]}`,
		`{"type":"table","x":117,"y":126,"w":180,"cols":[{"k":"x","label":"X","w":180}],"rows":[{"x":{"text":"Text","sub":"Subtitle"}}]}`,
		`{"type":"table","x":117,"y":126,"w":180,"cols":[{"k":"x","label":"X","w":180,"type":"dots"}],"rows":[{"x":{"value":1,"text":["Too tall","To fit"]}}]}`,
		`{"type":"block","x":117,"y":126,"w":180,"h":54,"heat":2,"heatScale":"invented","text":"X"}`,
		`{"type":"legend","x":117,"y":126,"w":180,"items":[{"text":"X","heat":2,"heatMax":0}]}`,
	} {
		if _, e := r.planSceneNode("invalid", json.RawMessage(raw), SceneContext{Surface: "light", Path: "/body/0"}); e == nil {
			t.Fatalf("accepted unsupported or overflowing node: %s", raw)
		}
	}
	for _, value := range []float64{math.NaN(), math.Inf(1)} {
		if _, _, e := sceneHeat(value, nil, ""); e == nil {
			t.Fatal("invalid heat accepted")
		}
	}
	for _, maximum := range []float64{math.NaN(), math.Inf(1), 0, -1} {
		if _, _, e := sceneHeat(0, &maximum, ""); e == nil {
			t.Fatal("invalid maximum accepted")
		}
	}
}

func TestTableLegacyNumericMarksPreserved(t *testing.T) {
	r := intakeTestRenderer(t)
	st, _ := r.sceneStyle("body")
	b := Rect{100, 100, 84, 36}
	for _, kind := range []string{"rating", "harvey"} {
		a, c := &scenePlan{}, &scenePlan{}
		if e := r.sceneTableNumericMark(a, "mark", kind, 2, b, "light"); e != nil {
			t.Fatal(e)
		}
		_, _, e := r.sceneTableScoreCell(c, "mark", sceneTableColumn{Type: kind}, json.RawMessage("2"), st, b, "light", SceneContext{}, "")
		if e != nil {
			t.Fatal(e)
		}
		aa, _ := json.Marshal(a.Items)
		cc, _ := json.Marshal(c.Items)
		if !bytes.Equal(aa, cc) {
			t.Fatalf("legacy %s geometry changed", kind)
		}
	}
}

func TestTableEnhancementNativeXML(t *testing.T) {
	raw := json.RawMessage(`{"type":"table","x":117,"y":126,"w":480,"rowH":72,"cols":[{"k":"name","label":"Name","w":240},{"k":"value","label":"Value","w":240,"type":"heat","showValue":true}],"rows":[{"name":{"text":"Platform","sub":"System of record"},"value":4}]}`)
	doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{{ID: "heat", Frame: FrameRequest{NoHeader: true}, Nodes: []Node{{ID: "table", Kind: "scene", Scene: &SceneSpec{Node: raw}}}}}}
	data, _, e := BuildWithEngine(filepath.Join("..", "..", "library", "wm-design-system", "v2"), "", doc, CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	z, e := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if e != nil {
		t.Fatal(e)
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
	d := xml.NewDecoder(bytes.NewReader(body))
	for {
		_, e := d.Token()
		if e == io.EOF {
			break
		}
		if e != nil {
			t.Fatal(e)
		}
	}
	if !bytes.Contains(body, []byte("<a:tbl>")) || !bytes.Contains(body, []byte("System of record")) || !bytes.Contains(body, []byte(`<a:alpha val="0"/>`)) {
		t.Fatal("editable table/subtitle/transparent heat cell missing")
	}
	if bytes.Index(body, []byte(".heat-gap")) > bytes.Index(body, []byte(".native")) {
		t.Fatal("heat background paints above table")
	}
}

func enhancementDenseHeatJSON() json.RawMessage {
	cols := make([]map[string]any, 12)
	rows := make([]map[string]any, 60)
	for i := range cols {
		cols[i] = map[string]any{"k": fmt.Sprintf("c%d", i), "label": "X", "w": 48, "type": "heat"}
	}
	for i := range rows {
		rows[i] = map[string]any{}
		for j := range cols {
			rows[i][fmt.Sprintf("c%d", j)] = (i + j) % 5
		}
	}
	raw, _ := json.Marshal(map[string]any{"type": "table", "x": 117, "y": 126, "w": 576, "rowH": 24, "cols": cols, "rows": rows})
	return raw
}
func TestTableHeatDenseLinearPaintAssembly(t *testing.T) {
	r := intakeTestRenderer(t)
	var n sceneTableSource
	if e := sceneDecode(enhancementDenseHeatJSON(), &n); e != nil {
		t.Fatal(e)
	}
	p, e := r.sceneTable("dense", n, SceneContext{Surface: "light", Path: "/body/0"})
	if e != nil {
		t.Fatal(e)
	}
	if len(p.Items) != 60*12*2+1 {
		t.Fatalf("items=%d", len(p.Items))
	}
	if p.Items[len(p.Items)-1].Table == nil {
		t.Fatal("heat underlays did not precede native table")
	}
	for i, it := range p.Items[:len(p.Items)-1] {
		if it.Shape == nil {
			t.Fatal("unexpected underlay")
		}
		row, col := i/2/12, i/2%12
		key := fmt.Sprintf(".row.item-%03d.c%d.", row+1, col)
		if !strings.Contains(it.Shape.Record.ID, key) {
			t.Fatalf("underlays changed authored order: %s wanted %s", it.Shape.Record.ID, key)
		}
	}
	heatWarnings := 0
	for _, warning := range p.Warnings {
		if strings.Contains(warning, SceneHeatContract) {
			heatWarnings++
		}
	}
	if heatWarnings != 1 {
		t.Fatal("heat warning duplicated percell")
	}
}
func BenchmarkTableHeatDense(b *testing.B) {
	bundle := filepath.Join("..", "..", "library", "wm-design-system", "v2")
	s, e := Load(bundle, "")
	if e != nil {
		b.Fatal(e)
	}
	engine, e := NewTypographyEngine(filepath.Join(bundle, "fonts"), CandidateEngine)
	if e != nil {
		b.Fatal(e)
	}
	r := &renderer{source: s, typeEngine: engine, bundle: bundle}
	raw := enhancementDenseHeatJSON()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var n sceneTableSource
		if e := sceneDecode(raw, &n); e != nil {
			b.Fatal(e)
		}
		if _, e := r.sceneTable("dense", n, SceneContext{Surface: "light", Path: "/body/0"}); e != nil {
			b.Fatal(e)
		}
	}
}

func TestTableHeatSaturationAndSurfaceGutters(t *testing.T) {
	for _, tc := range []struct {
		value, max float64
		color      string
	}{
		{-1, 4, "E8EEF8"}, {5, 4, "070154"}, {math.MaxFloat64, math.SmallestNonzeroFloat64, "070154"}, {0, math.SmallestNonzeroFloat64, "E8EEF8"},
	} {
		fill, _, err := sceneHeat(tc.value, &tc.max, "seq")
		if err != nil || fill != tc.color {
			t.Fatalf("heat(%g,%g)=%s,%v", tc.value, tc.max, fill, err)
		}
	}
	r := intakeTestRenderer(t)
	st, _ := r.sceneStyle("body")
	for _, surface := range []string{"light", "inverse", "deep"} {
		p := &scenePlan{}
		cell, _, err := r.sceneTableHeatCell(p, "heat", sceneTableColumn{ShowValue: true}, json.RawMessage("7"), st, Rect{100, 100, 120, 48}, surface, SceneContext{Surface: surface})
		if err != nil {
			t.Fatal(err)
		}
		background, _ := r.sceneColor(surface, "bg")
		if cell.Text != "7" || enhancementShape(t, p, ".heat").Record.Color != "070154" || enhancementShape(t, p, ".heat-gap").Record.Color != background {
			t.Fatal("saturation lost raw copy or surface gutters")
		}
	}
}

func TestTableScoreFootprintsBeforeLegacyShortcut(t *testing.T) {
	r := intakeTestRenderer(t)
	st, _ := r.sceneStyle("body")
	for _, tc := range []struct {
		kind string
		rect Rect
	}{
		{"rating", Rect{100, 100, 52, 36}}, {"dots", Rect{100, 100, 67, 36}}, {"harvey", Rect{100, 100, 36, 15}}, {"harvey", Rect{100, 100, 15, 36}},
	} {
		p := &scenePlan{}
		if _, _, err := r.sceneTableScoreCell(p, "score", sceneTableColumn{Type: tc.kind}, json.RawMessage("2"), st, tc.rect, "light", SceneContext{Surface: "light"}, ""); err == nil {
			t.Fatalf("accepted overflowing %s", tc.kind)
		}
		if len(p.Items) != 0 {
			t.Fatal("invalid score emitted partial marks")
		}
	}
	p := &scenePlan{}
	b := Rect{100, 100, 54, 36}
	if _, _, err := r.sceneTableScoreCell(p, "score", sceneTableColumn{Type: "rating"}, json.RawMessage("2"), st, b, "light", SceneContext{Surface: "light"}, ""); err != nil {
		t.Fatal(err)
	}
	for _, item := range p.Items {
		rect := item.Shape.Record.Rect
		if rect.X < b.X || rect.X+rect.W > b.X+b.W || rect.Y < b.Y || rect.Y+rect.H > b.Y+b.H {
			t.Fatal("mark spilled into adjacent cell")
		}
	}
}

func TestTablePlainSubtitleNativeParagraphLeading(t *testing.T) {
	r := intakeTestRenderer(t)
	st, _ := r.sceneStyle("body")
	// A caller-specified main paragraph leading must not leak into the small
	// subtitle. Frozen v2/v3 body defaults are21pt;24pt exercises this defect.
	st.Leading = 24
	_, tr, err := r.sceneTablePlainCell("plain", json.RawMessage(`{"text":"Platform","sub":"System of record"}`), st, Rect{100, 100, 240, 72}, "light", SceneContext{Surface: "light"})
	if err != nil {
		t.Fatal(err)
	}
	body, err := sceneTableCellTextXML(tr)
	if err != nil {
		t.Fatal(err)
	}
	var spacings []string
	d := xml.NewDecoder(bytes.NewReader(body))
	inLeading := false
	for {
		token, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		switch el := token.(type) {
		case xml.StartElement:
			if el.Name.Local == "lnSpc" {
				inLeading = true
			}
			if el.Name.Local == "spcPts" && inLeading {
				for _, attr := range el.Attr {
					if attr.Name.Local == "val" {
						spacings = append(spacings, attr.Value)
					}
				}
			}
		case xml.EndElement:
			if el.Name.Local == "lnSpc" {
				inLeading = false
			}
		}
	}
	if strings.Join(spacings, ",") != "2400,1800" {
		t.Fatalf("native main/subtitle leading=%v", spacings)
	}
}
