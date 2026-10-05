package wmdesign

import (
	"encoding/json"
	"math"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestV5NativeWordInsetsBoundariesAndNoOps(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source.Revision = LibraryRevisionV5
	st, err := r.sceneStyle("label")
	if err != nil {
		t.Fatal(err)
	}
	before := st
	layout, err := r.typeEngine.Measure("No.", st, 960)
	if err != nil {
		t.Fatal(err)
	}
	need := sequenceInlineWidth(layout.Lines[0].Advance, st)
	for _, tc := range []struct {
		name, text                              string
		width, left, right, wantLeft, wantRight float64
	}{
		{"bounded fit", "No.", need + 12 + .01, 12, 12, 6, 6},
		{"cannot fit six point inset", "No.", need + 12 - .01, 12, 12, 12, 12},
		{"already fits twelve point inset", "No.", need + 24 + .01, 12, 12, 12, 12},
		{"wide box", "No.", 126.01, 12, 12, 12, 12},
		{"asymmetric source inset", "No.", need + 12 + .01, 12, 11, 12, 11},
		{"rich run", "[^No.]", need + 12 + .01, 12, 12, 12, 12},
		{"empty copy", "", 36, 12, 12, 12, 12},
		{"oversized word", strings.Repeat("W", 100), 72, 12, 12, 12, 12},
	} {
		t.Run(tc.name, func(t *testing.T) {
			left, right, err := r.v5WordInsets(tc.text, st, tc.width, tc.left, tc.right)
			if err != nil || left != tc.wantLeft || right != tc.wantRight {
				t.Fatalf("got %v,%v error %v; want %v,%v", left, right, err, tc.wantLeft, tc.wantRight)
			}
		})
	}
	if !reflect.DeepEqual(st, before) {
		t.Fatal("source style changed")
	}
	if _, _, err := r.sceneNativeCell("oversized", strings.Repeat("W", 100), st, Rect{57, 144, 72, 18}, "light", "primary", "left", 12, 12, SceneContext{}); err == nil {
		t.Fatal("inset policy concealed an overflowing native cell")
	}
}

func TestV5NativeLegendReserveIsBoundedAndHistoricalPolicyPreserved(t *testing.T) {
	r := intakeTestRenderer(t)
	st, err := r.sceneStyle("small")
	if err != nil {
		t.Fatal(err)
	}
	var historicalFootprint float64
	for _, tc := range []struct {
		revision string
		size     float64
		reserve  bool
	}{
		{LibraryRevisionV3, 0, false}, {LibraryRevisionV4, 0, false},
		{LibraryRevisionV3, 12, true}, {LibraryRevisionV4, 12, true},
		{LibraryRevisionV5, 0, true},
	} {
		r.source.Revision = tc.revision
		raw, _ := json.Marshal(map[string]any{"type": "legend", "x": 57, "y": 144, "w": 240, "size": tc.size, "layout": "horizontal", "items": []map[string]any{{"key": "a", "text": "Pricing"}, {"key": "b", "text": "Ridgeview"}}})
		p, _, err := r.planPeopleScene("review", raw, SceneContext{Surface: "light"})
		if err != nil {
			t.Fatal(err)
		}
		var texts []*TextRecord
		for _, item := range p.Items {
			if item.Text != nil {
				texts = append(texts, item.Text)
			}
		}
		if len(texts) != 2 {
			t.Fatal("legend copy lost")
		}
		firstX := 57. + 10 + 9
		if tc.revision == LibraryRevisionV5 && tc.size == 0 {
			firstX -= sequenceInlineWidth(1, st) - 1
		}
		if math.Abs(texts[0].Rect.X-firstX) > .001 {
			t.Fatalf("%s size%v legend swatch gap changed incorrectly: x%.3f want%.3f", tc.revision, tc.size, texts[0].Rect.X, firstX)
		}
		for i, copy := range []string{"Pricing", "Ridgeview"} {
			tr := texts[i]
			l, err := r.typeEngine.Measure(copy, st, 240-10-9)
			if err != nil {
				t.Fatal(err)
			}
			width := l.Lines[0].Advance
			if tc.reserve {
				width = sequenceInlineWidth(width, st)
			}
			if math.Abs(tr.Rect.W-width) > .001 || tr.Layout.Displayed != copy || tr.Layout.Font.Typeface != l.Font.Typeface {
				t.Fatalf("legend policy/copy/font changed for %s size%v: %+v", tc.revision, tc.size, tr)
			}
			if tr.Rect.X < 57 || tr.Rect.X+tr.Rect.W > 297+.02 {
				t.Fatal("legend reserve escaped its source width")
			}
		}
		if v5RectsShareX(texts[0].Rect, texts[1].Rect) && math.Abs(texts[0].Rect.Y-texts[1].Rect.Y) < .02 {
			t.Fatal("legend item text rectangles overlap")
		}
		footprint := texts[1].Rect.X + texts[1].Rect.W - 57
		if tc.revision == LibraryRevisionV4 && tc.size == 0 {
			historicalFootprint = footprint
		}
		if tc.revision == LibraryRevisionV5 && tc.size == 0 && math.Abs(footprint-historicalFootprint) > .001 {
			t.Fatalf("native legend reserve increased source packing footprint: %.3f versus %.3f", footprint, historicalFootprint)
		}
	}
	r.source.Revision = LibraryRevisionV5
	raw := json.RawMessage(`{"type":"legend","x":57,"y":144,"w":25,"items":[{"key":"a","text":"Pricing"}]}`)
	if _, _, err := r.planPeopleScene("narrow", raw, SceneContext{Surface: "light"}); err == nil {
		t.Fatal("native reserve concealed an insufficient legend width")
	}
}

func TestV5NativeChartFormatsAndObservationsRevisionIsolation(t *testing.T) {
	r := intakeTestRenderer(t)
	raw := json.RawMessage(`{"type":"chart","x":57,"y":144,"w":540,"h":252,"kind":"column","categories":["One","Two"],"series":[{"name":"Actual","values":[1.25,2.5]}],"format":{"kind":"number","unit":"units"}}`)
	for _, revision := range []string{LibraryRevisionV3, LibraryRevisionV4, LibraryRevisionV5} {
		r.source.Revision = revision
		p, _, err := r.planChartScene("chart", raw, SceneContext{Surface: "light"})
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, item := range p.Items {
			if item.Chart == nil {
				continue
			}
			found = true
			want := `#,##0.##" units"`
			if revision == LibraryRevisionV5 {
				want = `#,##0.00" units"`
			}
			if item.Chart.Options.DataLabelFormatCode != want {
				t.Fatalf("%s format %q", revision, item.Chart.Options.DataLabelFormatCode)
			}
			if !reflect.DeepEqual(item.Chart.Data[0].Values, []float64{1.25, 2.5}) || item.Chart.Data[0].Name != "Actual" {
				t.Fatal("native formatting changed observations or series copy")
			}
		}
		if !found {
			t.Fatal("chart lost")
		}
	}
	for _, tc := range [][2]string{
		{`[=1]#,##0.#"s";#,##0.#"ss"`, `[=1]#,##0.0"s";#,##0.0"ss"`},
		{`+0.##" pts";"−"0.##" pts"`, `+0.00" pts";"−"0.00" pts"`},
		{`"0.#"0.##"0.##"`, `"0.#"0.00"0.##"`},
	} {
		if got := v5NativeChartFormat(tc[0]); got != tc[1] {
			t.Fatalf("format %q became %q, want %q", tc[0], got, tc[1])
		}
		if got := v5NativeChartFormat(tc[1]); got != tc[1] {
			t.Fatal("native format is not idempotent")
		}
	}
}

func TestV5QuadrantBadgeIsolationPreservesHeaderAndPoints(t *testing.T) {
	r := intakeTestRenderer(t)
	entries := intakeRepairEntries(t, filepath.Join("..", "..", "planning", "wm-design-contracts", "v5", "intake-20261003-587-frozen", "bundle", "source", "templates", "library"))
	slide := intakeRepairSlide(t, entries["stakeholders/quadrant"])
	if err := applyLibraryRefinements("stakeholders/quadrant", LibraryRevisionV5, &slide); err != nil {
		t.Fatal(err)
	}
	var oldText, newText map[string]TextRecord
	var oldBadge, newBadge Rect
	for _, revision := range []string{LibraryRevisionV4, LibraryRevisionV5} {
		r.source.Revision = revision
		p, _, err := r.planChartScene("q", slide.Nodes[0].Scene.Node, SceneContext{Surface: "light"})
		if err != nil {
			t.Fatal(err)
		}
		texts := map[string]TextRecord{}
		var badge Rect
		for _, item := range p.Items {
			if item.Text != nil {
				texts[item.Text.ID] = *item.Text
			}
			if item.Shape != nil && item.Shape.Record.ID == "q.tag-bg.tr" {
				badge = item.Shape.Record.Rect
			}
		}
		if revision == LibraryRevisionV4 {
			oldText, oldBadge = texts, badge
		} else {
			newText, newBadge = texts, badge
		}
	}
	if math.Abs(oldBadge.Y-(oldText["q.name.tr"].Rect.Y+18)) > .001 {
		t.Fatal("historical top badge policy changed")
	}
	if newBadge.Y < newText["q.name.tr"].Rect.Y+newText["q.name.tr"].Rect.H+3.99 {
		t.Fatal("new badge crowds header")
	}
	if len(oldText) != len(newText) {
		t.Fatal("quadrant copy lost")
	}
	for id, old := range oldText {
		next, ok := newText[id]
		if !ok || old.Layout.Displayed != next.Layout.Displayed || !reflect.DeepEqual(old.Layout.Font, next.Layout.Font) {
			t.Fatalf("quadrant copy/font changed: %s", id)
		}
		if !strings.Contains(id, ".tag.") && old.Rect != next.Rect {
			t.Fatalf("header/point geometry changed: %s", id)
		}
	}
}

func TestV5NativeHoleSizeRejectsQuadrantEarlyReturn(t *testing.T) {
	r := intakeTestRenderer(t)
	entries := intakeRepairEntries(t, filepath.Join("..", "..", "planning", "wm-design-contracts", "v5", "intake-20261003-587-frozen", "bundle", "source", "templates", "library"))
	slide := intakeRepairSlide(t, entries["stakeholders/quadrant"])
	if err := applyLibraryRefinements("stakeholders/quadrant", LibraryRevisionV5, &slide); err != nil {
		t.Fatal(err)
	}
	var node map[string]json.RawMessage
	if err := json.Unmarshal(slide.Nodes[0].Scene.Node, &node); err != nil {
		t.Fatal(err)
	}
	node["holeSize"] = json.RawMessage("50")
	raw, err := json.Marshal(node)
	if err != nil {
		t.Fatal(err)
	}
	for _, revision := range []string{LibraryRevisionV3, LibraryRevisionV4, LibraryRevisionV5} {
		r.source.Revision = revision
		if _, _, err := r.planChartScene("quadrant", raw, SceneContext{Surface: "light"}); err == nil {
			t.Fatalf("%s quadrant silently accepted doughnut-only holeSize", revision)
		}
	}
}
