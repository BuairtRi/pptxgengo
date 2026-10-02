package main

import (
	"reflect"
	"testing"

	"github.com/buairtri/pptxgengo/internal/compose"
)

func TestFontRequirementsIncludeMixedRunsAndPartialCacheHits(t *testing.T) {
	requests := []compose.ProbeRequest{
		{FontFace: "IBM Plex Sans", Bold: true},
		{Paragraphs: []compose.ParagraphSpec{{Runs: []compose.RunSpec{
			{FontFace: "IBM Plex Sans"},
			{FontFace: "IBM Plex Sans", Italic: true},
			{FontFace: "IBM Plex Sans", Bold: true, Italic: true},
			{FontFace: "Arial"},
		}}}},
		{FontFace: "IBM Plex Sans", Bold: true},
	}
	want := []fontRequirement{{"Arial", "regular"}, {"IBM Plex Sans", "bold"}, {"IBM Plex Sans", "bold_italic"}, {"IBM Plex Sans", "italic"}, {"IBM Plex Sans", "regular"}}
	if got := requestFonts(requests); !reflect.DeepEqual(got, want) {
		t.Fatalf("font inventory: %+v, want %+v", got, want)
	}
	// A partial cache probe must fingerprint fonts used by cached requests too.
	m := manifest{Requests: requests[:1], Environment: &measurementEnvironment{}}
	for _, font := range want {
		m.Environment.Fonts = append(m.Environment.Fonts, environmentFont{Family: font.Family, Style: font.Style})
	}
	if got := manifestFonts(m); !reflect.DeepEqual(got, want) {
		t.Fatalf("partial probe lost font requirements: %+v", got)
	}
	m.Slides = []renderSlide{{Elements: []element{{Kind: "text", FontFace: "IBM Plex Mono"}}}}
	if len(manifestFonts(m)) != len(want)+1 {
		t.Fatal("final rendered text font was omitted")
	}
}

func TestVariableFontCoordinatesInvalidateEnvironment(t *testing.T) {
	env := &measurementEnvironment{Fonts: []environmentFont{{Family: "IBM Plex Sans", SHA: "same-file", Variations: map[string]float64{"2003265652": 400}}}}
	before := environmentKey(env)
	env.Fonts[0].Variations["2003265652"] = 700
	if environmentKey(env) == before {
		t.Fatal("variable font weight did not invalidate the environment")
	}
}

func TestPlainFontVerificationRejectsLaterCharacterSubstitution(t *testing.T) {
	e := element{FontFace: "IBM Plex Sans", FontSize: 16, Bold: true}
	row := nativeRow{Characters: []nativeCharacter{{FontName: e.FontFace, FontSizePt: 16, Bold: true}, {FontName: e.FontFace, FontSizePt: 16, Bold: true}}}
	if err := checkNativePlainFonts(e, row, true); err != nil {
		t.Fatal(err)
	}
	row.Characters[1].FontName = "Arial"
	if err := checkNativePlainFonts(e, row, true); err == nil {
		t.Fatal("substituted later character font was accepted")
	}
	row.Characters = nil
	if err := checkNativePlainFonts(e, row, true); err == nil {
		t.Fatal("missing v8 character evidence was accepted")
	}
}
