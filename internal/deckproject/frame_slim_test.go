package deckproject

import (
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"testing"
)

func TestSlimFrameReferenceSourceGate(t *testing.T) {
	s := &wmdesign.Source{Revision: "wmds-library.v3", Frames: wmdesign.Frames{Footers: map[string]wmdesign.Footer{"compact": {}, "tall": {}}}}
	for _, rail := range []string{"none", "left", "right", "nav"} {
		ref := Reference{ID: "wmds/frame/" + rail + "-slim"}
		if _, e := resolveFrameRef(ref, s); e == nil {
			t.Fatal("slim fabricated in v3")
		}
		s.Frames.Footers["slim"] = wmdesign.Footer{Rule: 495, Row: [2]float64{504, 516}, Bottom: 486}
		q, e := resolveFrameRef(ref, s)
		if e != nil || q.Rail != rail || q.Footer != "slim" {
			t.Fatalf("source-supplied slim %+v %v", q, e)
		}
		delete(s.Frames.Footers, "slim")
	}
	for _, id := range []string{"wmds/frame/none-other", "wmds/frame/other-slim", "wmds/frame/none-slim-extra"} {
		if _, e := resolveFrameRef(Reference{ID: id}, s); e == nil {
			t.Fatal("invalid frame ID accepted")
		}
	}
}
