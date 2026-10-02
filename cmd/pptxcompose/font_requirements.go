package main

import (
	"fmt"
	"sort"

	"github.com/buairtri/pptxgengo/internal/compose"
)

// Font requirements identify only the styles actually used by native text.
// Keep the full set on cache-backed bundles even when only misses are probed.
type fontRequirement struct {
	Family string `json:"family"`
	Style  string `json:"style"`
}

func fontStyle(bold, italic bool) string {
	if bold && italic {
		return "bold_italic"
	}
	if bold {
		return "bold"
	}
	if italic {
		return "italic"
	}
	return "regular"
}

func textFonts(face string, bold bool, paragraphs []compose.ParagraphSpec) []fontRequirement {
	if len(paragraphs) == 0 {
		return []fontRequirement{{Family: face, Style: fontStyle(bold, false)}}
	}
	var out []fontRequirement
	for _, paragraph := range paragraphs {
		for _, run := range paragraph.Runs {
			out = append(out, fontRequirement{Family: run.FontFace, Style: fontStyle(run.Bold, run.Italic)})
		}
	}
	return out
}

func canonicalFonts(fonts []fontRequirement) []fontRequirement {
	seen := map[fontRequirement]bool{}
	out := make([]fontRequirement, 0, len(fonts))
	for _, font := range fonts {
		if !seen[font] {
			seen[font] = true
			out = append(out, font)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Family != out[j].Family {
			return out[i].Family < out[j].Family
		}
		return out[i].Style < out[j].Style
	})
	return out
}

func requestFonts(requests []compose.ProbeRequest) []fontRequirement {
	var out []fontRequirement
	for _, request := range requests {
		out = append(out, textFonts(request.FontFace, request.Bold, request.Paragraphs)...)
	}
	return canonicalFonts(out)
}

func manifestFonts(m manifest) []fontRequirement {
	out := requestFonts(m.Requests)
	for _, slide := range m.Slides {
		for _, e := range slide.Elements {
			if e.Kind == "text" {
				out = append(out, textFonts(e.FontFace, e.Bold, e.Paragraphs)...)
			}
		}
	}
	// Cached hits may not have a shape in a partial probe deck. Fingerprint
	// their font files too so measure/cache-import agree with the probe.
	if m.Environment != nil {
		for _, font := range m.Environment.Fonts {
			out = append(out, fontRequirement{Family: font.Family, Style: font.Style})
		}
	}
	return canonicalFonts(out)
}

func checkNativePlainFonts(e element, row nativeRow, requireCharacters bool) error {
	if requireCharacters && len(row.Characters) == 0 {
		return fmt.Errorf("missing native character font evidence")
	}
	for i, ch := range row.Characters {
		if ch.FontName != e.FontFace || !near(ch.FontSizePt, e.FontSize) || ch.Bold != e.Bold || ch.Italic {
			return fmt.Errorf("character %d has %q %.2fpt bold=%t italic=%t; expected %q %.2fpt bold=%t italic=false", i+1, ch.FontName, ch.FontSizePt, ch.Bold, ch.Italic, e.FontFace, e.FontSize, e.Bold)
		}
	}
	return nil
}
