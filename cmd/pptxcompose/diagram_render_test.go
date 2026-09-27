package main

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/compose"
)

func TestExtendedPhraseAndMotionFixtureContracts(t *testing.T) {
	t.Chdir("../..")
	b, err := os.ReadFile("library/diagram-components/accent-arrow-qualification.json")
	if err != nil {
		t.Fatal(err)
	}
	var spec compose.Spec
	if err := json.Unmarshal(b, &spec); err != nil {
		t.Fatal(err)
	}
	qs, err := compose.ProbeRequests(spec)
	if err != nil {
		t.Fatal(err)
	}
	claims := map[string]compose.ProbeRequest{}
	for _, q := range qs {
		if len(q.PhraseRequests) > 0 {
			claims[q.SlideID] = q
		}
	}
	for _, mode := range []string{"underline", "highlight"} {
		a, aok := claims[mode+"-source-copy"]
		b, bok := claims[mode+"-position-shift"]
		if !aok || !bok {
			t.Fatal("missing native translation probe pair")
		}
		a.ID = ""
		b.ID = ""
		a.SlideID = ""
		b.SlideID = ""
		if !reflect.DeepEqual(a, b) {
			t.Fatalf("%s translation changed text/width/style/selector contract", mode)
		}
	}
	q := claims["underline-unicode-range"]
	if len(q.PhraseRequests) != 1 {
		t.Fatal("missing Unicode selector")
	}
	a, z, err := compose.ResolvePhrase(q.Text, q.PhraseRequests[0])
	if err != nil || string([]rune(q.Text)[a:z]) != "pivotal moment" || a == strings.Index(q.Text, "pivotal moment") {
		t.Fatalf("Unicode range did not distinguish code points from bytes: %d..%d %v", a, z, err)
	}
	if len(claims["highlight-rich-runs"].Paragraphs) == 0 {
		t.Fatal("rich-text phrase has no runs")
	}
	// Tiny invented glyph bounds test geometry algebra and collision handling only.
	// They are never emitted as native measurement evidence or fit qualification.
	m := compose.Measurements{ByRequestID: map[string]compose.Measurement{}}
	for _, q := range qs {
		mm := compose.Measurement{RenderedWidthPt: 5, RenderedHeightPt: 5, PhraseBounds: map[string][]compose.Rect{}}
		for _, ph := range q.PhraseRequests {
			mm.PhraseBounds[ph.ID] = []compose.Rect{{X: 2, Y: 2, Width: 40, Height: 4}}
			if q.SlideID == "highlight-multiline-manual" {
				mm.PhraseBounds[ph.ID] = append(mm.PhraseBounds[ph.ID], compose.Rect{X: 2, Y: 20, Width: 40, Height: 4})
			}
		}
		m.ByRequestID[q.ID] = mm
	}
	plans := map[string]compose.PlannedSlide{}
	for _, s := range spec.Slides {
		one := compose.Spec{Schema: spec.Schema, Slides: []compose.SlideSpec{s}}
		requests, err := compose.ProbeRequests(one)
		if err != nil {
			t.Fatal(err)
		}
		mm := compose.Measurements{ByRequestID: map[string]compose.Measurement{}}
		for _, q := range requests {
			mm.ByRequestID[q.ID] = m.ByRequestID[q.ID]
		}
		p, err := compose.Plan(one, mm)
		if err != nil {
			t.Errorf("%s: %v", s.ID, err)
			continue
		}
		plans[s.ID] = p.Slides[0]
	}
	if t.Failed() {
		return
	}
	for _, pair := range []struct {
		a, b   string
		dx, dy float64
	}{
		{"underline-source-copy", "underline-position-shift", 12, 42},
		{"highlight-source-copy", "highlight-position-shift", 12, 42},
		{"dense-team-control-highlight", "dense-team-highlight-body-shift", 4, -6},
	} {
		a, b := plans[pair.a].Accents[0], plans[pair.b].Accents[0]
		for _, rs := range [][2]compose.Rect{{a.TargetBounds, b.TargetBounds}, {a.VisibleBounds, b.VisibleBounds}, {a.Bounds, b.Bounds}} {
			if !near(rs[1].X-rs[0].X, pair.dx) || !near(rs[1].Y-rs[0].Y, pair.dy) || !near(rs[1].Width, rs[0].Width) || !near(rs[1].Height, rs[0].Height) {
				t.Fatalf("%s accent translation changed geometry: %+v", pair.b, rs)
			}
		}
	}
	for _, id := range []string{"ambiguous-manual-stage", "highlight-multiline-manual", "arrow-direction-manual", "connecting-loop-manual-only"} {
		if len(plans[id].ManualRequired) != 1 {
			t.Fatalf("%s did not remain unfinished", id)
		}
	}
	control := plans["single-arrow-0.75-+0"].ArtworkArrows[0]
	moved := plans["single-arrow-translated-anchors"].ArtworkArrows[0]
	if control.Status != "placed" || moved.Status != "placed" {
		t.Fatal("motion pair unexpectedly staged")
	}
	for _, ps := range [][2]compose.Point{{control.TailPoint, moved.TailPoint}, {control.TipPoint, moved.TipPoint}} {
		if !near(ps[1].X-ps[0].X, 12) || !near(ps[1].Y-ps[0].Y, 20) {
			t.Fatalf("arrow endpoint failed translation: %+v", ps)
		}
	}
	if !near(control.Bounds.Width, moved.Bounds.Width) || !near(control.Bounds.Height, moved.Bounds.Height) || !near(control.RotationDeg, moved.RotationDeg) {
		t.Fatal("translated endpoints changed artwork scale/rotation")
	}
	changed := plans["single-arrow-target-moved"].ArtworkArrows[0]
	if !near(changed.TailPoint.X, control.TailPoint.X) || !near(changed.TailPoint.Y, control.TailPoint.Y) || !near(changed.TipPoint.X-control.TipPoint.X, 20) || !near(changed.TipPoint.Y-control.TipPoint.Y, 8) {
		t.Fatal("moving only target did not preserve source and recalculate tip")
	}
	if near(changed.RotationDeg, control.RotationDeg) {
		t.Fatal("changed endpoint slope did not rotate artwork")
	}
	var negative compose.Spec
	b, err = os.ReadFile("library/diagram-components/accent-ambiguous-negative.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &negative); err != nil {
		t.Fatal(err)
	}
	if _, err = compose.ProbeRequests(negative); err == nil || !strings.Contains(err.Error(), "choose an explicit occurrence or range") {
		t.Fatalf("ambiguous selection accepted: %v", err)
	}
}

// Synthetic bounds exercise fixed geometry, ownership, route clearance and
// package serialization only. Actual copy fit requires PowerPoint measurements.
func TestProposalCapacityGeometry(t *testing.T) {
	t.Chdir("../..")
	b, err := os.ReadFile("library/proposal/capacity-templates.json")
	if err != nil {
		t.Fatal(err)
	}
	var spec compose.Spec
	if err := json.Unmarshal(b, &spec); err != nil {
		t.Fatal(err)
	}
	for _, slide := range spec.Slides {
		t.Run(slide.ID, func(t *testing.T) {
			s := compose.Spec{Schema: spec.Schema, Slides: []compose.SlideSpec{slide}}
			qs, err := compose.ProbeRequests(s)
			if err != nil {
				t.Fatal(err)
			}
			m := compose.Measurements{ByRequestID: map[string]compose.Measurement{}}
			for _, q := range qs {
				m.ByRequestID[q.ID] = compose.Measurement{RenderedWidthPt: 5, RenderedHeightPt: 5}
			}
			p, err := compose.Plan(s, m)
			if err != nil {
				t.Fatal(err)
			}
			for _, c := range slide.Canvas {
				if c.AssetPath != "" {
					if _, err := os.Stat(c.AssetPath); os.IsNotExist(err) {
						t.Skip("geometry passed; pinned local assets are not included in a fresh checkout")
					}
				}
			}
			es := finalSlides(s, p)
			deck, err := render(es)
			if err != nil {
				t.Fatal(err)
			}
			if err := validateImageStructure(deck, es); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// Structural integration only: tiny synthetic text bounds exercise image
// serialization, not native text fit or visual qualification.
func TestWave3ArrowPackageStructure(t *testing.T) {
	t.Chdir("../..")
	b, err := os.ReadFile("library/diagram-components/arrow-qualification.json")
	if err != nil {
		t.Fatal(err)
	}
	var spec compose.Spec
	if err := json.Unmarshal(b, &spec); err != nil {
		t.Fatal(err)
	}
	for _, s := range spec.Slides {
		for _, a := range s.ArtworkArrows {
			if _, err := os.Stat(a.Artwork.AssetPath); os.IsNotExist(err) {
				t.Skip("local pinned arrow assets are not included in a fresh checkout")
			}
		}
	}
	qs, err := compose.ProbeRequests(spec)
	if err != nil {
		t.Fatal(err)
	}
	m := compose.Measurements{ByRequestID: map[string]compose.Measurement{}}
	for _, q := range qs {
		m.ByRequestID[q.ID] = compose.Measurement{RenderedWidthPt: 10, RenderedHeightPt: 10}
	}
	p, err := compose.Plan(spec, m)
	if err != nil {
		for _, s := range spec.Slides {
			one := compose.Spec{Schema: spec.Schema, Slides: []compose.SlideSpec{s}}
			qs, _ := compose.ProbeRequests(one)
			mm := compose.Measurements{ByRequestID: map[string]compose.Measurement{}}
			for _, q := range qs {
				mm.ByRequestID[q.ID] = m.ByRequestID[q.ID]
			}
			if _, oneErr := compose.Plan(one, mm); oneErr != nil {
				t.Logf("slide %s: %v", s.ID, oneErr)
			}
		}
		t.Fatal(err)
	}
	for i, slide := range p.Slides {
		if len(slide.ArtworkArrows) != 1 {
			t.Fatal("missing arrow")
		}
		if i == len(p.Slides)-1 {
			if slide.ArtworkArrows[0].Status != "manual_required" || len(slide.ManualRequired) != 1 {
				t.Fatal("fallback not marked unfinished")
			}
		} else if slide.ArtworkArrows[0].Status != "placed" {
			t.Fatal("candidate unexpectedly staged")
		}
	}
	slides := finalSlides(spec, p)
	deck, err := render(slides)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateImageStructure(deck, slides); err != nil {
		t.Fatal(err)
	}
}

func TestConnectionTreatmentsRemainDistinct(t *testing.T) {
	cs := []compose.PlannedConnection{
		{ID: "dependency", Relationship: "dependency", Color: "#070154", WidthPt: 1, LineDash: "solid", EndArrow: "triangle", Points: []compose.Point{{X: 30, Y: 20}, {X: 10, Y: 20}}},
		{ID: "advisory", Relationship: "advisory", Color: "#070154", WidthPt: 1, LineDash: "dash", Points: []compose.Point{{X: 10, Y: 30}, {X: 30, Y: 30}}},
		{ID: "annotation", Relationship: "annotation", Color: "#070154", WidthPt: 1, LineDash: "dot", Points: []compose.Point{{X: 10, Y: 40}, {X: 30, Y: 40}}},
		{ID: "flow", Relationship: "flow", Strategy: "direct_right_arrow", Color: "#070154", WidthPt: 1, Points: []compose.Point{{X: 10, Y: 50}, {X: 30, Y: 50}}},
	}
	es := connectorElements(cs)
	if len(es) != 3 {
		t.Fatalf("expected three line treatments, got %+v", es)
	}
	byID := map[string]element{}
	for _, e := range es {
		for _, id := range e.ConnectionIDs {
			byID[id] = e
		}
	}
	if byID["dependency"].BeginArrow != "triangle" || byID["dependency"].EndArrow != "" {
		t.Fatalf("left-facing dependency arrow wrong: %+v", byID["dependency"])
	}
	if byID["advisory"].LineDash != "dash" || byID["annotation"].LineDash != "dot" {
		t.Fatalf("line styles conflated: %+v", byID)
	}
}
