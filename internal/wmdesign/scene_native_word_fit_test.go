package wmdesign

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestV5NativeWordInsetsPreserveFontsAndBounds(t *testing.T) {
	r := intakeTestRenderer(t)
	for _, tc := range []struct {
		text, token string
		width       float64
	}{
		{"Pricing spreadsheet", "small", 90}, {"No.", "label", 36},
		{"10", "small", 36}, {"Ridgeview", "label", 72}, {"Influence", "label", 72},
	} {
		st, err := r.sceneStyle(tc.token)
		if err != nil {
			t.Fatal(err)
		}
		for _, rev := range []string{LibraryRevisionV1, LibraryRevisionV2, LibraryRevisionV3, LibraryRevisionV4} {
			r.source.Revision = rev
			left, right, err := r.v5WordInsets(tc.text, st, tc.width, 12, 12)
			if err != nil || left != 12 || right != 12 {
				t.Fatalf("earlier revision changed: %s", rev)
			}
		}
		r.source.Revision = LibraryRevisionV5
		left, right, err := r.v5WordInsets(tc.text, st, tc.width, 12, 12)
		if err != nil || left != 6 || right != 6 {
			t.Fatalf("%s: inset %v %v, %v", tc.text, left, right, err)
		}
		l, err := r.typeEngine.Measure(tc.text, st, tc.width-left-right)
		if err != nil {
			t.Fatal(err)
		}
		if tc.text != "Pricing spreadsheet" && len(l.Lines) != 1 {
			t.Fatalf("%s still breaks: %+v", tc.text, l.Lines)
		}
	}
}

func TestV5NativeChartFormatPreservesLiteralUnits(t *testing.T) {
	for _, tc := range [][2]string{
		{`#,##0.#`, `#,##0.0`},
		{`#,##0.##" 0.#"`, `#,##0.00" 0.#"`},
		{`"$"0.#,,"M"`, `"$"0.0,,"M"`},
		{`[<0.1]0.#%;0%`, `[<0.1]0.0%;0%`},
		{`#,##0;-#,##0;`, `#,##0;-#,##0;`},
	} {
		if got := v5NativeChartFormat(tc[0]); got != tc[1] {
			t.Fatalf("format %q: got %q, want %q", tc[0], got, tc[1])
		}
	}
}

func TestV5NativeLegendAndChartReserve(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source.Revision = LibraryRevisionV5
	entries := intakeRepairEntries(t, filepath.Join("..", "..", "planning", "wm-design-contracts", "v5", "intake-20261003-587-frozen", "bundle", "source", "templates", "library"))
	var heat, chart librarySlide
	if err := json.Unmarshal(entries["capability-heat/annotated"].Slide, &heat); err != nil {
		t.Fatal(err)
	}
	p, _, err := r.planPeopleScene("legend", heat.Body[1], SceneContext{Surface: "light"})
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, item := range p.Items {
		if item.Text != nil {
			count++
			if len(item.Text.Layout.Lines) != 1 {
				t.Fatalf("legend wraps: %+v", item.Text)
			}
		}
	}
	if count != 5 {
		t.Fatal("legend content lost")
	}
	if err := json.Unmarshal(entries["investment/cost-vs-value"].Slide, &chart); err != nil {
		t.Fatal(err)
	}
	p, _, err = r.planChartScene("chart", chart.Body[0], SceneContext{Surface: "light"})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, item := range p.Items {
		if item.Chart != nil {
			found = true
			if item.Chart.Options.DataLabelFormatCode != "#,##0.0" {
				t.Fatalf("native format: %+v", item.Chart.Options.DataLabelFormatCode)
			}
		}
	}
	if !found {
		t.Fatal("missing native chart")
	}
}

func TestV5QuadrantBadgeReservesMeasuredHeader(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source.Revision = LibraryRevisionV5
	entries := intakeRepairEntries(t, filepath.Join("..", "..", "planning", "wm-design-contracts", "v5", "intake-20261003-587-frozen", "bundle", "source", "templates", "library"))
	slide := intakeRepairSlide(t, entries["stakeholders/quadrant"])
	if err := applyLibraryRefinements("stakeholders/quadrant", LibraryRevisionV5, &slide); err != nil {
		t.Fatal(err)
	}
	p, _, err := r.planChartScene("stakeholders", slide.Nodes[0].Scene.Node, SceneContext{Surface: "light"})
	if err != nil {
		t.Fatal(err)
	}
	var name, badge Rect
	for _, item := range p.Items {
		if item.Text != nil && item.Text.ID == "stakeholders.name.tr" {
			name = item.Text.Rect
		}
		if item.Shape != nil && item.Shape.Record.ID == "stakeholders.tag-bg.tr" {
			badge = item.Shape.Record.Rect
		}
	}
	if name.W == 0 || badge.W == 0 || badge.Y < name.Y+name.H+3.99 {
		t.Fatalf("badge overlaps measured header: %+v %+v", name, badge)
	}
}
