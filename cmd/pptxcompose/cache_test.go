package main

import (
	"encoding/base64"
	"github.com/buairtri/pptxgengo/internal/compose"
	"testing"
)

func TestCacheContractBinding(t *testing.T) {
	q := compose.ProbeRequest{ID: "a", SlideID: "one", Text: "A phrase", TextWidthPt: 120, FontFace: "Arial", FontSizePt: 12, Align: "left", Foreground: "#070154", Background: "#FFFFFF"}
	key := contractKey(q)
	r := q
	r.ID = "b"
	r.SlideID = "two"
	if contractKey(r) != key {
		t.Fatal("locator changes should reuse a contract")
	}
	changes := []func(*compose.ProbeRequest){func(v *compose.ProbeRequest) { v.Text += "!" }, func(v *compose.ProbeRequest) { v.TextWidthPt++ }, func(v *compose.ProbeRequest) { v.FontFace = "other" }, func(v *compose.ProbeRequest) { v.FontSizePt++ }, func(v *compose.ProbeRequest) { v.Bold = true }, func(v *compose.ProbeRequest) { v.Align = "center" }, func(v *compose.ProbeRequest) { v.Foreground = "#FFFFFF" }, func(v *compose.ProbeRequest) { v.Background = "#070154" }, func(v *compose.ProbeRequest) { v.HorizontalInsetPt++ }, func(v *compose.ProbeRequest) { v.VerticalInsetPt++ }}
	for i, change := range changes {
		r = q
		change(&r)
		if contractKey(r) == key {
			t.Fatalf("rendering field %d not bound", i)
		}
	}
}
func TestCacheMissingContractsDeduplicate(t *testing.T) {
	q := compose.ProbeRequest{ID: "a", Text: "Repeated", TextWidthPt: 100, FontFace: "Arial", FontSizePt: 12}
	r := q
	r.ID = "b"
	third := q
	third.ID = "c"
	third.Text = "Different"
	m, missing, uses, e := cacheLookup(t.TempDir(), &measurementEnvironment{Schema: "fixture"}, []compose.ProbeRequest{q, r, third})
	if e != nil || len(missing) != 2 || len(uses) != 0 || len(m.ByRequestID) != 0 {
		t.Fatalf("unexpected miss result %d/%d: %v", len(missing), len(uses), e)
	}
}
func TestEnvironmentInvalidates(t *testing.T) {
	env := measurementEnvironment{Schema: "fixture", OS: "os", PowerPointBuild: "one", AdapterSHA: "a", ExecutableSHA: "b", Fonts: []environmentFont{{SHA: "font"}}}
	a := environmentKey(&env)
	env.PowerPointBuild = "two"
	if environmentKey(&env) == a {
		t.Fatal("PowerPoint build not bound")
	}
	env.PowerPointBuild = "one"
	env.Fonts[0].SHA = "changed"
	if environmentKey(&env) == a {
		t.Fatal("font bytes not bound")
	}
}
func TestRecoverLayoutSlot(t *testing.T) {
	s := compose.SlideSpec{Layouts: []compose.ContainerSpec{{ID: "grid", Cells: []compose.CellSpec{{ID: "cell", Blocks: []compose.BlockSpec{{ID: "body", Text: "before", Marker: "•"}}}}}}}
	name := "canvas:" + base64.RawURLEncoding.EncodeToString([]byte("grid/cell/body/text"))
	if e := setRecoveredText(&s, name, "after"); e != nil {
		t.Fatal(e)
	}
	if s.Layouts[0].Cells[0].Blocks[0].Text != "after" {
		t.Fatal("slot not updated")
	}
	name = "canvas:" + base64.RawURLEncoding.EncodeToString([]byte("grid/cell/body/marker"))
	if e := setRecoveredText(&s, name, "changed"); e == nil {
		t.Fatal("derived marker edit must not change content silently")
	}
}

func TestPhaseBackgroundAndExplicitLayers(t *testing.T) {
	spec := compose.Spec{Slides: []compose.SlideSpec{{WidthPt: 960, HeightPt: 540}}}
	p := compose.PlanResult{Slides: []compose.PlannedSlide{{ID: "s", Canvas: []compose.PlannedCanvas{{CanvasSpec: compose.CanvasSpec{ID: "front", Kind: "text", Layer: 20}}, {CanvasSpec: compose.CanvasSpec{ID: "back", Kind: "surface", Layer: 10}}}, Phases: []compose.PlannedPhase{{ID: "phase"}}}}}
	elems := finalSlides(spec, p)[0].Elements
	positions := map[string]int{}
	for i, e := range elems {
		positions[e.Name] = i
	}
	name := func(id string) string { return "canvas:" + base64.RawURLEncoding.EncodeToString([]byte(id)) }
	phase := "phase:" + base64.RawURLEncoding.EncodeToString([]byte("phase")) + "-surface"
	if positions[phase] >= positions[name("back")] || positions[name("back")] >= positions[name("front")] {
		t.Fatal("background or explicit canvas layer covered foreground")
	}
}
