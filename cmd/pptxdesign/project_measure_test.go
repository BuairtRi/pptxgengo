package main

import (
	"reflect"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestFilterMeasurementSlidesPreservesReportOrder(t *testing.T) {
	slides := []wmdesign.SlideReport{{ID: "opening", Page: 1}, {ID: "middle", Page: 2}, {ID: "closing", Page: 3}, {ID: "hidden", Page: 4}}
	got, err := filterMeasurementSlides(slides, "4,opening,2-3")
	if err != nil {
		t.Fatal(err)
	}
	ids := []string{}
	for _, slide := range got {
		ids = append(ids, slide.ID)
	}
	if want := []string{"opening", "middle", "closing", "hidden"}; !reflect.DeepEqual(ids, want) {
		t.Fatalf("selected slides %v; want report order %v", ids, want)
	}
}

func TestFilterMeasurementSlidesValidatesUnknownAndMalformedSelectors(t *testing.T) {
	slides := []wmdesign.SlideReport{{ID: "one"}, {ID: "two"}, {ID: "three"}}
	for _, selector := range []string{"missing", "0", "4", "3-2", "1-x", "1,,2", "1-2-3"} {
		if _, err := filterMeasurementSlides(slides, selector); err == nil {
			t.Errorf("accepted selector %q", selector)
		}
	}
}

func TestFilterMeasurementSlidesEmptyMeansAll(t *testing.T) {
	slides := []wmdesign.SlideReport{{ID: "one"}}
	got, err := filterMeasurementSlides(slides, "")
	if err != nil || len(got) != 1 || got[0].ID != "one" {
		t.Fatalf("got %v, %v", got, err)
	}
}

func TestFilterMeasurementSlidesHyphenatedIDs(t *testing.T) {
	slides := []wmdesign.SlideReport{{ID: "dxc-041", Page: 1}, {ID: "candidate-001", Page: 2}, {ID: "third-slide", Page: 3}}
	got, err := filterMeasurementSlides(slides, "candidate-001,1,third-slide")
	if err != nil || !reflect.DeepEqual(got, slides) {
		t.Fatalf("hyphenated identities must remain selectable: %v, %v", got, err)
	}
}

func TestFilterMeasurementSlidesNumericSelectorRemainsPage(t *testing.T) {
	slides := []wmdesign.SlideReport{{ID: "2", Page: 1}, {ID: "other", Page: 2}}
	got, err := filterMeasurementSlides(slides, "2")
	if err != nil || len(got) != 1 || got[0].ID != "other" {
		t.Fatalf("numeric selectors must retain page semantics: %v, %v", got, err)
	}
}
