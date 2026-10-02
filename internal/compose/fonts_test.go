package compose

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestExplicitFontFamilies(t *testing.T) {
	for _, face := range []string{"Arial", "IBM Plex Sans", "IBM Plex Mono", "Helvetica Neue"} {
		if !ValidFontFace(face) {
			t.Fatalf("valid family %q rejected", face)
		}
	}
	for _, face := range []string{"", "   ", " Arial", "Arial ", "+mj-lt", "IBM\nPlex Sans", "Arial\x00", "bad\ufffd"} {
		if ValidFontFace(face) {
			t.Fatalf("invalid family %q accepted", face)
		}
	}
}

func replaceFixtureFonts(v any) {
	switch value := v.(type) {
	case map[string]any:
		for key, child := range value {
			if key == "font_face" || key == "title_font_face" {
				value[key] = "IBM Plex Sans"
			} else {
				replaceFixtureFonts(child)
			}
		}
		if cards, ok := value["cards"].([]any); ok {
			for _, card := range cards {
				card.(map[string]any)["font_face"] = "IBM Plex Sans"
			}
		}
	case []any:
		for _, child := range value {
			replaceFixtureFonts(child)
		}
	}
}

func TestCompositionFixturesAcceptOtherFamilies(t *testing.T) {
	for _, path := range []string{"dynamic-components/pods.json", "dynamic-components/team.json", "dynamic-components/cards.json", "dynamic-components/fonts.json", "layout-components/controls.json"} {
		t.Run(path, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("../../library", path))
			if err != nil {
				t.Fatal(err)
			}
			var fixture any
			if err := json.Unmarshal(data, &fixture); err != nil {
				t.Fatal(err)
			}
			replaceFixtureFonts(fixture)
			data, err = json.Marshal(fixture)
			if err != nil {
				t.Fatal(err)
			}
			var spec Spec
			if err := json.Unmarshal(data, &spec); err != nil {
				t.Fatal(err)
			}
			requests, err := ProbeRequests(spec)
			if err != nil {
				t.Fatal(err)
			}
			for _, q := range requests {
				if len(q.Paragraphs) == 0 && q.FontFace != "IBM Plex Sans" {
					t.Fatalf("%s silently used %q", q.ID, q.FontFace)
				}
				for _, p := range q.Paragraphs {
					for _, run := range p.Runs {
						if run.FontFace != "IBM Plex Sans" {
							t.Fatalf("%s rich text font lost", q.ID)
						}
					}
				}
			}
		})
	}
}

func TestCardFontSurvivesPlanningAndKeepsDefault(t *testing.T) {
	for _, face := range []string{"", "IBM Plex Sans"} {
		s := SlideSpec{ID: "s", WidthPt: 960, HeightPt: 540, TitleFontFace: "IBM Plex Sans", TitleFontSizePt: 26,
			Cards: []CardSpec{{ID: "card", Kind: "metric", Bounds: Rect{X: 30, Y: 100, Width: 300, Height: 200}, FontFace: face, Value: "42", Label: "Sample metric"}}}
		spec := Spec{Schema: SpecSchema, Slides: []SlideSpec{s}}
		requests, err := ProbeRequests(spec)
		if err != nil {
			t.Fatal(err)
		}
		measured := Measurements{ByRequestID: map[string]Measurement{}}
		for _, q := range requests {
			measured.ByRequestID[q.ID] = Measurement{RenderedWidthPt: 40, RenderedHeightPt: 15}
		}
		plan, err := Plan(spec, measured)
		if err != nil {
			t.Fatal(err)
		}
		want := face
		if want == "" {
			want = "Arial"
		}
		for _, block := range plan.Slides[0].Cards[0].Blocks {
			if block.FontFace != want {
				t.Fatalf("card font %q became %q", want, block.FontFace)
			}
		}
	}
}
