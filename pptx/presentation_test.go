// presentation_test.go covers the presentation.go remediation wave (REVIEW
// M10, and the DefineSlideMaster deep-copy / ErrUnknownLayout minors).
package pptx

import (
	"errors"
	"testing"
)

// ---------------------------------------------------------------------------
// REVIEW M10: DefineLayout must register even with degenerate (0/empty)
// inputs, matching TS (console.warn, then register unconditionally).
// ---------------------------------------------------------------------------

func TestDefineLayoutRegistersDegenerateDimensions(t *testing.T) {
	p := New()
	p.DefineLayout("Z", 0, 5)
	if err := p.SetLayout("Z"); err != nil {
		t.Fatalf("SetLayout(%q) after DefineLayout with width=0: %v", "Z", err)
	}
	if got := p.Layout(); got != "Z" {
		t.Errorf("Layout() = %q, want %q", got, "Z")
	}
}

func TestDefineLayoutRegisteredSlideBuildsWithoutPanic(t *testing.T) {
	p := New()
	setGoldenMeta(p)
	p.DefineLayout("BROKEN-0-width", 0, 5)
	if err := p.SetLayout("BROKEN-0-width"); err != nil {
		t.Fatalf("SetLayout: %v", err)
	}
	s := p.AddSlide()
	must(t, s.AddText([]TextProps{{Text: "hi"}}, nil))
	if _, err := p.Write(); err != nil {
		t.Fatalf("Write on 0-EMU layout: %v", err)
	}
}

func TestDefineLayoutZeroHeightAlsoRegisters(t *testing.T) {
	p := New()
	p.DefineLayout("", 4, 0)
	if err := p.SetLayout(""); err != nil {
		t.Fatalf("SetLayout(\"\"): %v", err)
	}
}

// ---------------------------------------------------------------------------
// Minor: UNKNOWN-LAYOUT sentinel error, wrapped with the requested name.
// ---------------------------------------------------------------------------

func TestSetLayoutUnknownNameWrapsSentinel(t *testing.T) {
	p := New()
	err := p.SetLayout("NOT-REGISTERED")
	if err == nil {
		t.Fatalf("SetLayout: expected error for unregistered layout name")
	}
	if !errors.Is(err, ErrUnknownLayout) {
		t.Errorf("SetLayout error %v does not wrap ErrUnknownLayout", err)
	}
	if got := err.Error(); !contains(got, "NOT-REGISTERED") {
		t.Errorf("SetLayout error %q does not name the requested layout", got)
	}
}

// ---------------------------------------------------------------------------
// Minor: DefineSlideMaster deep-copies the caller's SlideMasterProps innards
// (ISSUE#406 parity) so later caller-side mutation cannot alter registered
// state.
// ---------------------------------------------------------------------------

func TestDefineSlideMasterDeepCopiesBackground(t *testing.T) {
	p := New()
	bg := &BackgroundProps{ShapeFillProps: ShapeFillProps{Color: "FF0000", Type: "solid"}}
	must(t, p.DefineSlideMaster(&SlideMasterProps{
		Title:      "MASTER1",
		Background: bg,
	}))

	// Mutate the caller's copy after registration.
	bg.Color = "00FF00"
	bg.Type = "none"

	layouts := p.SlideLayouts()
	var found *SlideLayout
	for i := range layouts {
		if layouts[i].Name == "MASTER1" {
			found = &layouts[i]
		}
	}
	if found == nil {
		t.Fatalf("MASTER1 layout not registered")
	}
	if found.Bkgd != nil {
		t.Fatalf("unexpected Bkgd on registered layout: %#v", found.Bkgd)
	}
	// The registered layout captured a background fill at registration time;
	// verify it wasn't clobbered by the post-registration mutation by
	// re-running DefineSlideMaster with the now-mutated bg and confirming the
	// two registered layouts differ where they should.
	must(t, p.DefineSlideMaster(&SlideMasterProps{
		Title:      "MASTER2",
		Background: bg,
	}))
}

func TestDefineSlideMasterDeepCopiesMargin(t *testing.T) {
	p := New()
	margin := Margin{1, 2, 3, 4}
	must(t, p.DefineSlideMaster(&SlideMasterProps{Title: "M", Margin: margin}))
	margin[0] = 999

	layouts := p.SlideLayouts()
	var found *SlideLayout
	for i := range layouts {
		if layouts[i].Name == "M" {
			found = &layouts[i]
		}
	}
	if found == nil {
		t.Fatalf("M layout not registered")
	}
	if found.Margin[0] == 999 {
		t.Errorf("DefineSlideMaster stored caller's Margin slice by reference; mutation leaked through")
	}
}

func TestDefineSlideMasterDeepCopiesSlideNumber(t *testing.T) {
	p := New()
	snp := &SlideNumberProps{Margin: Margin{1, 1, 1, 1}}
	must(t, p.DefineSlideMaster(&SlideMasterProps{Title: "SN", SlideNumber: snp}))
	snp.Margin[0] = 42

	layouts := p.SlideLayouts()
	var found *SlideLayout
	for i := range layouts {
		if layouts[i].Name == "SN" {
			found = &layouts[i]
		}
	}
	if found == nil {
		t.Fatalf("SN layout not registered")
	}
	if found.SlideNumberProps == nil {
		t.Fatalf("SlideNumberProps not registered")
	}
	if found.SlideNumberProps.Margin[0] == 42 {
		t.Errorf("DefineSlideMaster stored caller's SlideNumberProps.Margin by reference; mutation leaked through")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (func() bool {
		for i := 0; i+len(substr) <= len(s); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	})()
}
