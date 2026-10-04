package wmdesign

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestV5LabelOnlyTextblockPreservesSourceContent(t *testing.T) {
	r := intakeTestRenderer(t)
	raw := json.RawMessage("{\"type\":\"textblock\",\"x\":100,\"y\":160,\"w\":270,\"label\":\"Topic 1\",\"title\":\"Denials\"}")
	for _, revision := range []string{LibraryRevisionV1, LibraryRevisionV2, LibraryRevisionV3, LibraryRevisionV4} {
		r.source.Revision = revision
		if _, _, err := r.planPrimitiveScene("topic", raw, SceneContext{Surface: "light"}); err == nil {
			t.Fatalf("body requirement changed for %s", revision)
		}
	}
	r.source.Revision = LibraryRevisionV5
	p, handled, err := r.planPrimitiveScene("topic", raw, SceneContext{Surface: "light"})
	if err != nil || !handled {
		t.Fatal(handled, err)
	}
	texts := []string{}
	for _, item := range p.Items {
		if item.Text != nil {
			texts = append(texts, item.Text.Layout.Original)
		}
	}
	if strings.Join(texts, "/") != "Topic 1/Denials" {
		t.Fatalf("source content invented or lost: %v", texts)
	}
	for _, raw := range []json.RawMessage{
		json.RawMessage("{\"type\":\"textblock\",\"x\":100,\"y\":160,\"w\":270}"),
		json.RawMessage("{\"type\":\"textblock\",\"x\":100,\"y\":160,\"w\":270,\"text\":\"Unsupported alias\"}"),
	} {
		if _, _, err := r.planPrimitiveScene("empty", raw, SceneContext{Surface: "light"}); err == nil {
			t.Fatal("empty textblock or unsupported alias accepted")
		}
	}
}

func TestV5TenColumnCategoriesPreserveAllSourceValues(t *testing.T) {
	r := intakeTestRenderer(t)
	entries := intakeRepairEntries(t, filepath.Join("..", "..", "library", "wm-design-system", "v5", "source", "templates", "library"))
	var slide librarySlide
	if err := json.Unmarshal(entries["interviews-coverage/overview"].Slide, &slide); err != nil {
		t.Fatal(err)
	}
	raw := slide.Body[4]
	var source sceneChartSource
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	if len(source.Categories) != 10 || len(source.Series) != 1 || len(source.Series[0].Values) != 10 {
		t.Fatal("source specimen changed")
	}
	for _, revision := range []string{LibraryRevisionV1, LibraryRevisionV2, LibraryRevisionV3, LibraryRevisionV4} {
		r.source.Revision = revision
		if _, _, err := r.planChartScene("coverage", raw, SceneContext{Surface: "light"}); err == nil {
			t.Fatalf("column capacity changed for %s", revision)
		}
	}
	r.source.Revision = LibraryRevisionV5
	p, handled, err := r.planChartScene("coverage", raw, SceneContext{Surface: "light"})
	if err != nil || !handled {
		t.Fatal(handled, err)
	}
	found := false
	wantLabels := append([]string(nil), source.Categories...)
	for i := range wantLabels {
		wantLabels[i] = strings.ToUpper(wantLabels[i])
	}
	for _, item := range p.Items {
		if item.Chart == nil {
			continue
		}
		if len(item.Chart.Data) != 1 || !reflect.DeepEqual(item.Chart.Data[0].Labels, [][]string{wantLabels}) {
			t.Fatal("chart category labels dropped or changed")
		}
		for i, value := range item.Chart.Data[0].Values {
			if value != *source.Series[0].Values[i] {
				t.Fatalf("numeric observation %d changed", i)
			}
		}
		found = true
	}
	if !found {
		t.Fatal("native chart missing")
	}
	source.Categories = append(source.Categories, "Eleventh")
	v := 1.0
	source.Series[0].Values = append(source.Series[0].Values, &v)
	raw, _ = json.Marshal(source)
	if _, _, err := r.planChartScene("over-capacity", raw, SceneContext{Surface: "light"}); err == nil {
		t.Fatal("unbounded v5 column capacity")
	}
}

func TestV5ExplicitCardBulletSizeIsBounded(t *testing.T) {
	r := intakeTestRenderer(t)
	flow := func(raw string, revision string) (*scenePlan, error) {
		r.source.Revision = revision
		p := &scenePlan{}
		_, err := r.sceneBodyFlow(p, "card", []json.RawMessage{json.RawMessage(raw)}, SceneContext{Surface: "light"}, "body", Rect{100, 160, 270, 0}, "light", "", 6)
		return p, err
	}
	raw := "{\"bullets\":[\"Seven hospitals\",\"64 clinics\"],\"size\":\"small\"}"
	for _, revision := range []string{LibraryRevisionV1, LibraryRevisionV2, LibraryRevisionV3, LibraryRevisionV4} {
		if _, err := flow(raw, revision); err == nil {
			t.Fatalf("new size accepted by %s", revision)
		}
	}
	p, err := flow(raw, LibraryRevisionV5)
	if err != nil {
		t.Fatal(err)
	}
	seen := 0
	for _, item := range p.Items {
		if item.Text != nil && strings.HasSuffix(item.Text.ID, ".text") {
			if item.Text.Layout.Style.ID != "small" || item.Text.Layout.Style.Size != 12 {
				t.Fatal("authored bullet size ignored", item.Text.Layout.Style)
			}
			seen++
		}
	}
	if seen != 2 || len(p.Warnings) == 0 {
		t.Fatal("missing content or explicit extension provenance")
	}
	for _, invalid := range []string{
		"{\"bullets\":[\"One\"],\"size\":\"tiny\"}",
		"{\"p\":\"One\",\"size\":\"small\"}",
		"{\"bullets\":[\"One\"],\"size\":12}",
	} {
		if _, err := flow(invalid, LibraryRevisionV5); err == nil {
			t.Fatal("unsupported size accepted", invalid)
		}
	}
}

func TestV5StatusLegendPreservesLabelsAndSquares(t *testing.T) {
	r := intakeTestRenderer(t)
	raw := json.RawMessage("{\"type\":\"legend\",\"x\":57,\"y\":432,\"w\":846,\"layout\":\"horizontal\",\"items\":[{\"text\":\"Received\",\"status\":\"on\"},{\"text\":\"Pending, inside due date\",\"status\":\"risk\"},{\"text\":\"Overdue or blocked\",\"status\":\"off\"}]}")
	for _, revision := range []string{LibraryRevisionV1, LibraryRevisionV2, LibraryRevisionV3, LibraryRevisionV4} {
		r.source.Revision = revision
		if _, _, err := r.planPeopleScene("status", raw, SceneContext{Surface: "light"}); err == nil {
			t.Fatalf("status legend expanded %s", revision)
		}
	}
	r.source.Revision = LibraryRevisionV5
	p, handled, err := r.planPeopleScene("status", raw, SceneContext{Surface: "light"})
	if err != nil || !handled {
		t.Fatal(handled, err)
	}
	texts, fills := []string{}, []string{}
	for _, item := range p.Items {
		if item.Text != nil {
			texts = append(texts, item.Text.Layout.Original)
		}
		if item.Shape != nil && strings.HasSuffix(item.Shape.Record.ID, ".swatch") {
			if item.Shape.Record.Rect.W != 8 || item.Shape.Record.Rect.H != 8 || item.Shape.Props.Line.Width != .75 {
				t.Fatal("source status swatch geometry changed")
			}
			fills = append(fills, item.Shape.Record.Color)
		}
	}
	if !reflect.DeepEqual(texts, []string{"Received", "Pending, inside due date", "Overdue or blocked"}) || !reflect.DeepEqual(fills, []string{"1DD566", "FFC700", "F52C00"}) {
		t.Fatal("status content or colors changed", texts, fills)
	}
	raw = json.RawMessage("{\"type\":\"legend\",\"x\":57,\"y\":432,\"w\":846,\"items\":[{\"status\":\"unknown\"}]}")
	if _, _, err := r.planPeopleScene("invalid", raw, SceneContext{Surface: "light"}); err == nil {
		t.Fatal("unknown legend status accepted")
	}
}
