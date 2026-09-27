package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/buairtri/pptxgengo/internal/compose"
)

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
