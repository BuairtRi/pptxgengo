package main

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

func measurement(rotation float64) phraseMeasurement {
	return phraseMeasurement{
		Phrase:   "anchor phrase",
		Bounds:   phraseRect{Left: 100, Top: 20, Width: 80, Height: 10},
		Rotation: &rotation,
		Lines: []struct {
			Start  int        `json:"start_character"`
			End    int        `json:"end_character"`
			Bounds phraseRect `json:"bounds"`
		}{{Start: 1, End: 13, Bounds: phraseRect{Left: 100, Top: 20, Width: 80, Height: 10}}},
	}
}

func asset() assetBounds {
	return assetBounds{
		File:          "underline.svg",
		ViewBox:       []float64{0, 0, 72, 72},
		VisibleBounds: rect{X: 0, Y: 36, Width: 72, Height: 6},
	}
}

func TestCalculateAccountsForTransparentCanvas(t *testing.T) {
	m := measurement(0)
	got, err := calculate(m, asset(), 4, 0, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "placed" || got.Placement == nil {
		t.Fatalf("expected placement, got %#v", got)
	}
	// Visible width is 88 pt (phrase plus 4 pt on each side). The visible alpha
	// occupies half the square SVG vertically and starts at phrase bottom + 2 pt.
	want := placementEMU{X: int64((100 - 4) * emuPerPoint), Y: int64(10 * emuPerPoint), Width: int64(88 * emuPerPoint), Height: int64(44 * emuPerPoint)}
	if *got.Placement != want {
		t.Fatalf("placement = %#v, want %#v", *got.Placement, want)
	}
	if got.VisibleFraction == nil || got.VisibleFraction.X != 0 || got.VisibleFraction.Y != .5 || got.VisibleFraction.Height != 1.0/12 {
		t.Fatalf("visible fraction = %#v", got.VisibleFraction)
	}
}

func TestMultilineRequestsManualPlacement(t *testing.T) {
	m := measurement(0)
	m.Lines = append(m.Lines, m.Lines[0])
	got, err := calculate(m, asset(), 0, 0, 0, 2)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "manual_required" || got.Placement != nil {
		t.Fatalf("expected manual fallback, got %#v", got)
	}
	if !strings.Contains(got.OperatorNote, "2 measured line fragments") {
		t.Fatalf("unexpected note: %s", got.OperatorNote)
	}
}

func TestRotatedPhraseRequestsManualPlacement(t *testing.T) {
	got, err := calculate(measurement(12), asset(), 0, 0, 0, 2)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "manual_required" || got.Placement != nil {
		t.Fatalf("expected manual fallback, got %#v", got)
	}
}

func TestMissingRotationRequestsManualPlacement(t *testing.T) {
	m := measurement(0)
	m.Rotation = nil
	got, err := calculate(m, asset(), 0, 0, 0, 2)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "manual_required" || got.Placement != nil {
		t.Fatalf("expected manual fallback, got %#v", got)
	}
}

func TestRejectsNonpositiveDimensions(t *testing.T) {
	m := measurement(0)
	m.Bounds.Width = 0
	if _, err := calculate(m, asset(), 0, 0, 0, 2); err == nil {
		t.Fatal("expected zero phrase width rejection")
	}
	a := asset()
	a.VisibleBounds.Height = 0
	if _, err := calculate(measurement(0), a, 0, 0, 0, 2); err == nil {
		t.Fatal("expected zero visible asset height rejection")
	}
}

func TestSourceFiveUnderlineCalibration(t *testing.T) {
	rotation := 0.0
	m := phraseMeasurement{
		Phrase:   "pivotal moment",
		Bounds:   phraseRect{Left: 164.094879150391, Top: 45.479290008545, Width: 177.255039215088, Height: 28.8},
		Rotation: &rotation,
		Lines: []struct {
			Start  int        `json:"start_character"`
			End    int        `json:"end_character"`
			Bounds phraseRect `json:"bounds"`
		}{{Start: 13, End: 26, Bounds: phraseRect{Left: 164.094879150391, Top: 45.479290008545, Width: 177.255039215088, Height: 28.8}}},
	}
	a := assetBounds{
		File:          "image77.svg",
		ViewBox:       []float64{0, 0, 72, 72},
		VisibleBounds: rect{X: 0, Y: 33.1083984375, Width: 72, Height: 5.7919921875},
	}
	// These explicit offsets reproduce the supplied source placement. They are
	// calibration inputs, not inferred defaults for future phrase reflow.
	padding := -1.2317322059691946
	horizontalOffset := 0.11945163576578466
	verticalOffset := -3.227455397478323
	containerAspect := 2219853.0 / 1346842.0
	got, err := calculate(m, a, padding, horizontalOffset, verticalOffset, containerAspect)
	if err != nil {
		t.Fatal(err)
	}
	want := placementEMU{X: 2101165, Y: 283028, Width: 2219853, Height: 1346842}
	if got.Placement == nil || *got.Placement != want {
		t.Fatalf("source calibration placement = %#v, want %#v", got.Placement, want)
	}
}

func TestNativeMeasurementLeftTopSchema(t *testing.T) {
	const input = `{"presentation":"uhg-reference.pptx","slide":5,"shape_index":5,"phrase":"pivotal moment","text":"You’re at a pivotal moment ","start_character":13,"end_character":26,"coordinates":"raw PowerPoint text-range units","rotation_degrees":0,"bounds":{"left":164.094879150391,"top":45.479290008545,"width":177.255039215088,"height":28.8},"lines":[{"start_character":13,"end_character":26,"bounds":{"left":164.094879150391,"top":45.479290008545,"width":177.255039215088,"height":28.8}}]}`
	var m phraseMeasurement
	if err := json.Unmarshal([]byte(input), &m); err != nil {
		t.Fatal(err)
	}
	if m.Bounds.Left != 164.094879150391 || m.Bounds.Top != 45.479290008545 {
		t.Fatalf("left/top bounds decoded as %#v", m.Bounds)
	}
	got, err := calculate(m, asset(), 0, 0, 0, 2)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "placed" || got.Placement == nil || got.Placement.X == 0 || got.Placement.Y == 0 {
		t.Fatalf("expected nonzero coordinates from native left/top schema; got %#v", got)
	}
}

func TestHighlightRotationFitsTargetVisibleBounds(t *testing.T) {
	m := measurement(0)
	m.Bounds = phraseRect{Left: 40, Top: 30, Width: 140, Height: 28}
	a := assetBounds{
		File:          "background.png",
		ViewBox:       []float64{0, 0, 100, 50},
		VisibleBounds: rect{X: 10, Y: 10, Width: 80, Height: 30},
	}
	got, err := calculateHighlight(m, a, 5, 4, 2, -1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "placed" || got.Placement == nil || got.RotationOOXML == nil || *got.RotationOOXML != 60000 {
		t.Fatalf("unexpected highlight result: %#v", got)
	}
	if got.ZOrderIntent != "behind_target_text" {
		t.Fatalf("z-order intent = %q", got.ZOrderIntent)
	}
	// Independently rotate the output image's alpha bbox and confirm its AABB
	// matches phrase bounds plus padding/offset within EMU rounding tolerance.
	container := got.Placement
	x, y := float64(container.X)/emuPerPoint, float64(container.Y)/emuPerPoint
	w, h := float64(container.Width)/emuPerPoint, float64(container.Height)/emuPerPoint
	v := *got.VisibleFraction
	cx, cy := x+w/2, y+h/2
	alphaX0, alphaY0 := x+v.X*w, y+v.Y*h
	alphaX1, alphaY1 := alphaX0+v.Width*w, alphaY0+v.Height*h
	angle := math.Pi / 180
	minX, minY, maxX, maxY := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, corner := range [][2]float64{{alphaX0, alphaY0}, {alphaX1, alphaY0}, {alphaX0, alphaY1}, {alphaX1, alphaY1}} {
		dx, dy := corner[0]-cx, corner[1]-cy
		rx := cx + dx*math.Cos(angle) - dy*math.Sin(angle)
		ry := cy + dx*math.Sin(angle) + dy*math.Cos(angle)
		minX, maxX = math.Min(minX, rx), math.Max(maxX, rx)
		minY, maxY = math.Min(minY, ry), math.Max(maxY, ry)
	}
	want := *got.TargetVisible
	checks := [][2]float64{{minX, want.X}, {minY, want.Y}, {maxX - minX, want.Width}, {maxY - minY, want.Height}}
	for _, check := range checks {
		if math.Abs(check[0]-check[1]) > 0.0001 {
			t.Fatalf("rotated alpha bounds differ: got %v want %v", check[0], check[1])
		}
	}
}

func TestHighlightReflowsWithPhraseLocationAndWidth(t *testing.T) {
	m := measurement(0)
	a := assetBounds{File: "highlight.png", ViewBox: []float64{0, 0, 100, 20}, VisibleBounds: rect{X: 0, Y: 0, Width: 100, Height: 20}}
	first, err := calculateHighlight(m, a, 2, 2, 0, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	m.Phrase = "WestMonroe"
	m.Bounds = phraseRect{Left: 239.184875, Top: 45.47929, Width: 200.049992, Height: 38.400002}
	second, err := calculateHighlight(m, a, 2, 2, 0, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if first.Placement == nil || second.Placement == nil || *first.Placement == *second.Placement {
		t.Fatalf("expected changed phrase geometry to change placement: first=%#v second=%#v", first.Placement, second.Placement)
	}
}

func TestHighlightMultilineManualFallback(t *testing.T) {
	m := measurement(0)
	m.Lines = append(m.Lines, m.Lines[0])
	got, err := calculateHighlight(m, asset(), 4, 3, 0, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "manual_required" || got.Placement != nil || got.ZOrderIntent != "behind_target_text" {
		t.Fatalf("expected manual highlight fallback, got %#v", got)
	}
}

func westMonroeMeasurement(resized bool) phraseMeasurement {
	rotation := 0.0
	phrase := phraseMeasurement{
		Presentation: "uhg-reference.pptx",
		Slide:        6,
		ShapeIndex:   5,
		Phrase:       "West Monroe",
		Rotation:     &rotation,
	}
	if resized {
		phrase.Bounds = phraseRect{Left: 239.184875488281, Top: 45.479290008545, Width: 200.049991607666, Height: 38.400001525879}
	} else {
		phrase.Bounds = phraseRect{Left: 188.609848022461, Top: 45.479290008545, Width: 150.195048332214, Height: 28.799999237061}
	}
	phrase.Lines = []struct {
		Start  int        `json:"start_character"`
		End    int        `json:"end_character"`
		Bounds phraseRect `json:"bounds"`
	}{{Start: 14, End: 24, Bounds: phrase.Bounds}}
	return phrase
}

func TestHighlightCalibrationMatchesOriginalPicture5(t *testing.T) {
	a := assetBounds{File: "image90.png", ViewBox: []float64{0, 0, 1200, 165}, VisibleBounds: rect{X: 0, Y: 0, Width: 1200, Height: 165}}
	got, err := calculateHighlight(westMonroeMeasurement(false), a, 5.058313518049971, 5.220756405694651, 1.6051474964713748, 0.18744265638906654, 1)
	if err != nil {
		t.Fatal(err)
	}
	want := placementEMU{X: 2355376, Y: 531327, Width: 2028186, Height: 463041}
	if got.Placement == nil || *got.Placement != want {
		t.Fatalf("baseline highlight placement = %#v, want %#v", got.Placement, want)
	}
	if got.RotationOOXML == nil || *got.RotationOOXML != 60000 || got.ZOrderIntent != "behind_target_text" {
		t.Fatalf("rotation/z-order metadata is wrong: %#v", got)
	}
}

func TestHighlightCalibrationReflowsResizedWestMonroe(t *testing.T) {
	a := assetBounds{File: "image90.png", ViewBox: []float64{0, 0, 1200, 165}, VisibleBounds: rect{X: 0, Y: 0, Width: 1200, Height: 165}}
	got, err := calculateHighlight(westMonroeMeasurement(true), a, 5.058313518049971, 5.220756405694651, 1.6051474964713748, 0.18744265638906654, 1)
	if err != nil {
		t.Fatal(err)
	}
	want := placementEMU{X: 2998599, Y: 536828, Width: 2659504, Height: 573960}
	if got.Placement == nil || *got.Placement != want {
		t.Fatalf("resized highlight placement = %#v, want %#v", got.Placement, want)
	}
}
