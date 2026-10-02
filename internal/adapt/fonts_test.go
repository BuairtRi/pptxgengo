package adapt

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/buairtri/pptxgengo/internal/compose"
)

func TestAdaptiveFamiliesPropagateSelectedFont(t *testing.T) {
	for _, name := range []string{"process-normal", "roadmap-normal", "architecture-v4", "team-v4", "comparison-dial"} {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("../../library/adaptive", name+".json"))
			if err != nil {
				t.Fatal(err)
			}
			var input Spec
			if err := Decode(data, &input); err != nil {
				t.Fatal(err)
			}
			for i := range input.Slides {
				input.Slides[i].Style.FontFace = "IBM Plex Sans"
				input.Slides[i].Footer = "Font selection also applies to the footer"
			}
			output, _, err := Compile(input)
			if err != nil {
				t.Fatal(err)
			}
			requests, err := compose.ProbeRequests(output)
			if err != nil {
				t.Fatal(err)
			}
			if len(requests) == 0 {
				t.Fatal("no text requests")
			}
			for _, q := range requests {
				if q.FontFace != "IBM Plex Sans" {
					t.Fatalf("%s silently used %q", q.ID, q.FontFace)
				}
			}
		})
	}
}

func TestStyleFontDefaultAndExplicitFamily(t *testing.T) {
	style, err := ResolveStyle(StyleOptions{})
	if err != nil || style.FontFace != "Arial" {
		t.Fatalf("default changed: %+v %v", style, err)
	}
	for _, face := range []string{"IBM Plex Sans", "Helvetica", "IBM Plex Mono"} {
		style, err := ResolveStyle(StyleOptions{FontFace: face})
		if err != nil || style.FontFace != face {
			t.Fatalf("explicit family lost: %+v %v", style, err)
		}
	}
	if _, err := ResolveStyle(StyleOptions{FontFace: "+mn-lt"}); err == nil {
		t.Fatal("theme font alias accepted")
	}
}
