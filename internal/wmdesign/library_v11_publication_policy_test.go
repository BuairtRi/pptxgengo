package wmdesign

import "testing"

func TestLibraryV11PublicationDensityIntroductionIsBounded(t *testing.T) {
	previous, err := Load(v10IntakeBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	current, err := Load("../../planning/wm-design-contracts/v11/intake-20261006-649-frozen/bundle", "")
	if err != nil {
		t.Fatal(err)
	}
	if !compatiblePublicationStyle(previous, current) {
		t.Fatal("added density policy blocked the subsequent exact rendering comparisons")
	}
	for _, tc := range []struct {
		name   string
		change func(*Source, *Source)
	}{
		{"older-revision", func(a, b *Source) { a.Revision = LibraryRevisionV9 }},
		{"reverse-introduction", func(a, b *Source) { a.Revision, b.Revision = LibraryRevisionV11, LibraryRevisionV10 }},
		{"missing-density", func(a, b *Source) { b.Tokens.Density = nil; b.Tokens.Grid.Gutter++ }},
		{"changed-grid", func(a, b *Source) { b.Tokens.Grid.Gutter++ }},
		{"existing-density-changed", func(a, b *Source) { a.Tokens.Density = b.Tokens.Density; b.Tokens.Density = nil }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, b := *previous, *current
			tc.change(&a, &b)
			if compatiblePublicationStyle(&a, &b) {
				t.Fatal("unqualified dependency change accepted")
			}
		})
	}
	if !compatiblePublicationStyle(previous, current) {
		t.Fatal("compatibility check mutated the cached sources")
	}
}

func TestPublicationCompositionRetainsDensitySettings(t *testing.T) {
	a := SlideSpec{ID: "slide", Density: "comfortable", Frame: FrameRequest{HeaderDensity: "comfortable"}}
	b := a
	b.Density = "compact"
	if samePublicationComposition(a, b) {
		t.Fatal("body density change inherited an old preview")
	}
	b = a
	b.Frame.HeaderDensity = "dense"
	if samePublicationComposition(a, b) {
		t.Fatal("header density change inherited an old preview")
	}
	b = a
	disabled := false
	b.AutoDensity = &disabled
	if samePublicationComposition(a, b) {
		t.Fatal("automatic fit policy change discarded during composition comparison")
	}
}
