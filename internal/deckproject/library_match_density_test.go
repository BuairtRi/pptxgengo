package deckproject

import (
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestPreferredSlotOverflowAllowsDensityBuildAndKeepsOptOut(t *testing.T) {
	v11 := "../../planning/wm-design-contracts/v11/intake-20261006-649-frozen/bundle"
	densitySource, err := wmdesign.Load(v11, "")
	if err != nil {
		t.Fatal(err)
	}
	v10 := "../../planning/wm-design-contracts/v10/intake-20261006-649-frozen/bundle"
	historicalSource, err := wmdesign.Load(v10, "")
	if err != nil {
		t.Fatal(err)
	}
	no := false
	yes := true
	for _, tc := range []struct {
		name   string
		source *wmdesign.Source
		auto   *bool
		want   bool
	}{
		{name: "density default permits renderer trial", source: densitySource, want: false},
		{name: "explicit density opt out keeps slot rejection", source: densitySource, auto: &no, want: true},
		{name: "explicit auto permits renderer trial", source: densitySource, auto: &yes, want: false},
		{name: "historical source keeps legacy rejection", source: historicalSource, want: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := preferredSlotOverflowBlocksBuild(true, tc.source, tc.auto); got != tc.want {
				t.Fatalf("slot overflow rejection = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestMapPageLetsRendererTryAutomaticDensityAfterPreferredOverflow(t *testing.T) {
	bundle := "../../planning/wm-design-contracts/v11/intake-20261006-649-frozen/bundle"
	catalog, err := wmdesign.LibraryCatalog(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	var def wmdesign.LibraryTemplate
	for _, candidate := range catalog {
		if candidate.Key == "cards/3" {
			def = candidate
			break
		}
	}
	if def.Key == "" {
		t.Fatal("v11 cards/3 definition missing")
	}
	found := false
	for words := 1; words <= 2; words++ {
		page := matchPage()
		page.Items[0].Text = strings.TrimSpace(strings.Repeat("The team coordinates priorities across owners and turns decisions into measurable delivery outcomes. ", words))
		candidate, err := MapPageToTemplate(page, def, bundle, wmdesign.CandidateEngine, "density-adjustment", 2026)
		if err != nil {
			t.Fatal(err)
		}
		if candidate.Status != "go_layout_succeeded_native_review_pending" || candidate.layout == nil {
			continue
		}
		for _, slide := range candidate.layout.Slides {
			if slide.Density != nil && slide.Density.Adjusted {
				found = true
				preferredOverflow := false
				for _, fit := range candidate.SlotFits {
					preferredOverflow = preferredOverflow || fit.Status == "overflow_estimate"
				}
				if !preferredOverflow {
					t.Fatal("automatic adjustment passed without preserving preferred-tier overflow evidence")
				}
				if slide.Density.Resolved == slide.Density.Requested {
					t.Fatalf("candidate reports automatic adjustment without a changed tier: %+v", slide.Density)
				}
				if slide.Density.Resolved != "compact" && slide.Density.Resolved != "dense" {
					t.Fatalf("unexpected adjusted tier: %+v", slide.Density)
				}
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Fatal("no tested card copy crossed preferred-tier capacity and fit after automatic density")
	}
}
