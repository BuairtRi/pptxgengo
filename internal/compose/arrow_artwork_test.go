package compose

import (
	"math"
	"strings"
	"testing"
)

func arrowFixture() ArtworkArrowSpec {
	return ArtworkArrowSpec{ID: "a", Artwork: ArrowArtwork{ID: "right", IntrinsicWidth: 100, IntrinsicHeight: 40, AlphaBounds: Rect{X: .1, Y: .1, Width: .8, Height: .8}, Tail: Point{.1, .5}, Tip: Point{.9, .5}, TailTangent: Point{1, 0}, TipTangent: Point{1, 0}, MinSpanPt: 40, MaxSpanPt: 200, MaxRotationDeltaDeg: 180}}
}

func TestArtworkArrowFinalPhraseGeometry(t *testing.T) {
	a := arrowFixture()
	a.From = ArtworkAnchor{Target: "label", Port: "right", GapPt: 4, Phrase: &PhraseRequest{Phrase: "key"}}
	a.To = ArtworkAnchor{Target: "target", Port: "left", GapPt: 4}
	p := &PlannedSlide{Canvas: []PlannedCanvas{
		{CanvasSpec: CanvasSpec{ID: "label", Kind: "text", Bounds: Rect{X: 20, Y: 30, Width: 100, Height: 40}}, MeasurementID: "m"},
		{CanvasSpec: CanvasSpec{ID: "target", Kind: "text", Bounds: Rect{X: 200, Y: 30, Width: 100, Height: 40}}},
	}}
	m := Measurements{ByRequestID: map[string]Measurement{"m": {PhraseBounds: map[string][]Rect{"a/anchor-0": {{X: 10, Y: 5, Width: 30, Height: 20}}}}}}
	from, _, err := artworkAnchorPoint(a, 0, p, m)
	if err != nil {
		t.Fatal(err)
	}
	to, _, err := artworkAnchorPoint(a, 1, p, m)
	if err != nil {
		t.Fatal(err)
	}
	p.ArtworkArrows = []PlannedArtworkArrow{{ArtworkArrowSpec: a, Status: "placed", TailPoint: from, TipPoint: to}}
	if err := VerifyArtworkAnchors(p, m, .15); err != nil {
		t.Fatal(err)
	}
	m.ByRequestID["m"].PhraseBounds["a/anchor-0"][0].X += 2
	if err := VerifyArtworkAnchors(p, m, .15); err == nil || !strings.Contains(err.Error(), "shifted") {
		t.Fatalf("shift accepted: %v", err)
	}
	delete(m.ByRequestID["m"].PhraseBounds, "a/anchor-0")
	if err := VerifyArtworkAnchors(p, m, .15); err == nil {
		t.Fatal("missing phrase accepted")
	}
}

func TestArtworkCollisionIncludesAccentsAndPriorArrows(t *testing.T) {
	b := Rect{X: 100, Y: 100, Width: 50, Height: 20}
	for _, p := range []*PlannedSlide{
		{Accents: []PlannedAccent{{AccentSpec: AccentSpec{ID: "underline"}, VisibleBounds: b}}},
		{ArtworkArrows: []PlannedArtworkArrow{{ArtworkArrowSpec: ArtworkArrowSpec{ID: "staged-arrow"}, VisibleBounds: b, Status: "manual_required"}}},
	} {
		if err := checkAccentCollisions(AccentSpec{ID: "new"}, b, p); err == nil {
			t.Fatal("artwork collision accepted")
		}
	}
}

func TestArtworkInkRegionsKeepEmptyCornersAvailable(t *testing.T) {
	p := &PlannedSlide{ArtworkArrows: []PlannedArtworkArrow{{
		ArtworkArrowSpec: ArtworkArrowSpec{ID: "elbow"},
		VisibleBounds:    Rect{X: 100, Y: 100, Width: 100, Height: 100},
		VisibleRegions:   []Rect{{X: 100, Y: 100, Width: 8, Height: 100}, {X: 100, Y: 192, Width: 100, Height: 8}},
	}}}
	if err := checkAccentCollisions(AccentSpec{ID: "label"}, Rect{X: 140, Y: 120, Width: 40, Height: 30}, p); err != nil {
		t.Fatalf("empty curved-arrow corner treated as ink: %v", err)
	}
	if err := checkAccentCollisions(AccentSpec{ID: "label"}, Rect{X: 102, Y: 120, Width: 40, Height: 30}, p); err == nil {
		t.Fatal("real visible stroke overlap accepted")
	}
	a := arrowFixture()
	a.Artwork.InkRegions = []Rect{{X: .1, Y: .4, Width: .8, Height: .2}}
	r, err := solveArtworkArrow(a, Point{100, 100}, Point{100, 180}, Point{0, 1}, Point{0, -1})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.VisibleRegions) != 1 || math.Abs(r.VisibleRegions[0].Width-8) > .001 || math.Abs(r.VisibleRegions[0].Height-80) > .001 {
		t.Fatalf("ink regions did not inherit arrow rotation: %+v", r.VisibleRegions)
	}
}
func TestArtworkArrowSimilarityAndTranslation(t *testing.T) {
	a := arrowFixture()
	r, err := solveArtworkArrow(a, Point{100, 100}, Point{180, 100}, Point{1, 0}, Point{-1, 0})
	if err != nil {
		t.Fatal(err)
	}
	if r.Bounds != (Rect{X: 90, Y: 80, Width: 100, Height: 40}) || r.RotationDeg != 0 {
		t.Fatalf("wrong similarity: %+v", r)
	}
	moved, err := solveArtworkArrow(a, Point{120, 130}, Point{200, 130}, Point{1, 0}, Point{-1, 0})
	if err != nil {
		t.Fatal(err)
	}
	if moved.Bounds.X-r.Bounds.X != 20 || moved.Bounds.Y-r.Bounds.Y != 30 {
		t.Fatal("does not follow parents")
	}
	rotated, err := solveArtworkArrow(a, Point{100, 100}, Point{100, 180}, Point{0, 1}, Point{0, -1})
	if err != nil || math.Abs(rotated.RotationDeg-90) > .001 {
		t.Fatalf("rotation: %+v %v", rotated, err)
	}
	if math.Abs(rotated.Bounds.Width/rotated.Bounds.Height-2.5) > .001 {
		t.Fatal("distorted arrow aspect")
	}
}
func TestArtworkArrowRejectsUnqualifiedPlacement(t *testing.T) {
	a := arrowFixture()
	if _, err := solveArtworkArrow(a, Point{100, 100}, Point{120, 100}, Point{1, 0}, Point{-1, 0}); err == nil {
		t.Fatal("short span accepted")
	}
	a.Artwork.MaxRotationDeltaDeg = 10
	if _, err := solveArtworkArrow(a, Point{100, 100}, Point{100, 180}, Point{0, 1}, Point{0, -1}); err == nil {
		t.Fatal("rotation outside contract accepted")
	}
	if _, err := solveArtworkArrow(a, Point{100, 100}, Point{180, 100}, Point{-1, 0}, Point{-1, 0}); err == nil {
		t.Fatal("reverse-facing tail accepted")
	}
}

func manualOnlyArrowFixture() (SlideSpec, *PlannedSlide, Measurements) {
	a := arrowFixture()
	a.ManualOnly = true
	a.Artwork.AssetPath = "loop.svg"
	a.Artwork.AssetSHA256 = strings.Repeat("a", 64)
	a.Artwork.MinSpanPt = 0
	a.Artwork.MaxSpanPt = 0
	a.Artwork.Tail = Point{}
	a.Artwork.Tip = Point{}
	a.Artwork.TailTangent = Point{}
	a.Artwork.TipTangent = Point{}
	a.From = ArtworkAnchor{Target: "source", Port: "right"}
	a.To = ArtworkAnchor{Target: "destination", Port: "left"}
	a.Staging = &AccentStaging{
		AssetBounds: Rect{X: 400, Y: 300, Width: 120, Height: 80},
		NoteBounds:  Rect{X: 540, Y: 300, Width: 280, Height: 60},
		Note:        "Place loop artwork manually after reviewing endpoints.",
	}
	s := SlideSpec{ID: "s", WidthPt: 960, HeightPt: 540, ArtworkArrows: []ArtworkArrowSpec{a}}
	p := &PlannedSlide{Canvas: []PlannedCanvas{
		{CanvasSpec: CanvasSpec{ID: "source", Bounds: Rect{X: 40, Y: 80, Width: 100, Height: 80}}},
		{CanvasSpec: CanvasSpec{ID: "destination", Bounds: Rect{X: 250, Y: 80, Width: 100, Height: 80}}},
	}}
	m := Measurements{ByRequestID: map[string]Measurement{
		arrowNoteProbe(s, a).ID: {RenderedWidthPt: 200, RenderedHeightPt: 30},
	}}
	return s, p, m
}

func TestManualOnlyArtworkArrowStagesWithoutCalibration(t *testing.T) {
	s, p, m := manualOnlyArrowFixture()
	if err := validateArtworkArrows(s, map[string]bool{}); err != nil {
		t.Fatal(err)
	}
	if err := planArtworkArrows(s, p, m); err != nil {
		t.Fatal(err)
	}
	if len(p.ArtworkArrows) != 1 || p.ArtworkArrows[0].Status != "manual_required" ||
		p.ArtworkArrows[0].Bounds != s.ArtworkArrows[0].Staging.AssetBounds ||
		!strings.Contains(p.ArtworkArrows[0].Reason, "manual_only policy disables automatic endpoint placement") ||
		len(p.ManualRequired) != 1 || len(p.Canvas) != 3 {
		t.Fatalf("uncalibrated arrow was not staged: %+v", p)
	}
}

func TestManualOnlyArtworkArrowRequiresStagingAndValidTargets(t *testing.T) {
	s, p, m := manualOnlyArrowFixture()
	s.ArtworkArrows[0].Staging = nil
	if err := validateArtworkArrows(s, map[string]bool{}); err == nil || !strings.Contains(err.Error(), "requires staging") {
		t.Fatalf("missing staging accepted: %v", err)
	}
	s, p, m = manualOnlyArrowFixture()
	s.ArtworkArrows[0].To.Target = "missing"
	if err := planArtworkArrows(s, p, m); err == nil || !strings.Contains(err.Error(), "unknown arrow anchor") {
		t.Fatalf("unknown target accepted: %v", err)
	}
}

func TestAutomaticArtworkArrowStillRequiresSpanContract(t *testing.T) {
	s, _, _ := manualOnlyArrowFixture()
	a := &s.ArtworkArrows[0]
	a.ManualOnly = false
	a.Artwork.EndpointProvenance = "measured"
	a.Artwork.Tail = Point{.1, .5}
	a.Artwork.Tip = Point{.9, .5}
	a.Artwork.TailTangent = Point{1, 0}
	a.Artwork.TipTangent = Point{1, 0}
	if err := validateArtworkArrows(s, map[string]bool{}); err == nil || !strings.Contains(err.Error(), "invalid span") {
		t.Fatalf("automatic arrow without span contract accepted: %v", err)
	}
}

func TestManualOnlyArtworkArrowNeverAutoPlaces(t *testing.T) {
	s, p, m := manualOnlyArrowFixture()
	// Even valid endpoint calibration cannot override an explicit manual choice.
	art := arrowFixture().Artwork
	art.AssetPath = "loop.svg"
	art.AssetSHA256 = strings.Repeat("a", 64)
	art.EndpointProvenance = "test calibration"
	s.ArtworkArrows[0].Artwork = art
	if err := validateArtworkArrows(s, map[string]bool{}); err != nil {
		t.Fatal(err)
	}
	if err := planArtworkArrows(s, p, m); err != nil {
		t.Fatal(err)
	}
	if p.ArtworkArrows[0].Status != "manual_required" {
		t.Fatal("manual choice was overridden by calibration")
	}
}

func TestManualOnlyArtworkArrowStillChecksStaging(t *testing.T) {
	for _, mode := range []string{"collision", "note-overflow", "missing-measurement"} {
		t.Run(mode, func(t *testing.T) {
			s, p, m := manualOnlyArrowFixture()
			a := s.ArtworkArrows[0]
			switch mode {
			case "collision":
				p.Canvas[0].Kind = "text"
				p.Canvas[0].Bounds = a.Staging.AssetBounds
			case "note-overflow":
				m.ByRequestID[arrowNoteProbe(s, a).ID] = Measurement{RenderedWidthPt: 281, RenderedHeightPt: 30}
			case "missing-measurement":
				delete(m.ByRequestID, arrowNoteProbe(s, a).ID)
			}
			if err := planArtworkArrows(s, p, m); err == nil {
				t.Fatal("invalid manual staging accepted")
			}
		})
	}
}
