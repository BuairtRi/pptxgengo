package pptx

import (
	"strings"
	"testing"
)

func TestHiddenSlideAppCount(t *testing.T) {
	visible, hidden := false, true
	slides := []PresSlide{{}, {Hidden: &visible}, {Hidden: &hidden}, {Hidden: &hidden}}
	if !strings.Contains(makeXmlApp(slides, ""), "<HiddenSlides>2</HiddenSlides>") {
		t.Fatal("hidden count is inaccurate")
	}
	if !strings.Contains(makeXmlApp(slides[:2], ""), "<HiddenSlides>0</HiddenSlides>") {
		t.Fatal("legacy visible output changed")
	}
}
