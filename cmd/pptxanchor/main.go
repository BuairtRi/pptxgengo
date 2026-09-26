// Command pptxanchor computes the image-container placement needed to anchor
// visible SVG artwork to a native PowerPoint text phrase. It only emits JSON;
// it never edits a presentation.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
)

const emuPerPoint = 12700.0

type rect struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// phraseRect matches PowerPoint's native text measurement schema, which uses
// left/top while pixel_qa reports use x/y.
type phraseRect struct {
	Left   float64 `json:"left"`
	Top    float64 `json:"top"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

func (r *phraseRect) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for _, key := range []string{"left", "top", "width", "height"} {
		if _, ok := fields[key]; !ok {
			return fmt.Errorf("phrase bounds missing required %q field", key)
		}
	}
	var parsed struct {
		Left   float64 `json:"left"`
		Top    float64 `json:"top"`
		Width  float64 `json:"width"`
		Height float64 `json:"height"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}
	*r = phraseRect(parsed)
	return nil
}

type phraseMeasurement struct {
	Presentation string     `json:"presentation"`
	Slide        int        `json:"slide"`
	ShapeIndex   int        `json:"shape_index"`
	Phrase       string     `json:"phrase"`
	Text         string     `json:"text"`
	StartChar    int        `json:"start_character"`
	EndChar      int        `json:"end_character"`
	Coordinates  string     `json:"coordinates"`
	Bounds       phraseRect `json:"bounds"`
	Rotation     *float64   `json:"rotation_degrees"`
	Lines        []struct {
		Start  int        `json:"start_character"`
		End    int        `json:"end_character"`
		Bounds phraseRect `json:"bounds"`
	} `json:"lines"`
}

type assetBounds struct {
	File          string    `json:"file"`
	ViewBox       []float64 `json:"viewBox"`
	VisibleBounds rect      `json:"visibleBounds"`
}

type placementEMU struct {
	X      int64 `json:"x"`
	Y      int64 `json:"y"`
	Width  int64 `json:"cx"`
	Height int64 `json:"cy"`
}

type result struct {
	Status          string        `json:"status"`
	Mode            string        `json:"mode"`
	Intent          string        `json:"intent"`
	Phrase          string        `json:"phrase"`
	Slide           int           `json:"slide,omitempty"`
	ShapeIndex      int           `json:"shape_index,omitempty"`
	PhraseBounds    *rect         `json:"phrase_bounds_points,omitempty"`
	Asset           string        `json:"asset,omitempty"`
	VisibleFraction *rect         `json:"visible_fraction,omitempty"`
	TargetVisible   *rect         `json:"target_visible_bounds_points,omitempty"`
	Placement       *placementEMU `json:"placement_emu,omitempty"`
	RotationDegrees *float64      `json:"rotation_degrees,omitempty"`
	RotationOOXML   *int64        `json:"rotation_ooxml_60000,omitempty"`
	ZOrderIntent    string        `json:"z_order_intent,omitempty"`
	OperatorNote    string        `json:"operator_note"`
}

func main() {
	measurementPath := flag.String("measurement", "", "native phrase measurement JSON")
	assetPath := flag.String("asset", "", "pixel_qa visible-bounds JSON")
	mode := flag.String("mode", "underline", "placement mode: underline or highlight (default underline)")
	units := flag.String("units", "", "explicit coordinate acknowledgment; must be points")
	padding := flag.Float64("stroke-padding-pt", 0, "extra visible-art padding on each side of the phrase, in points")
	horizontalPadding := flag.Float64("horizontal-padding-pt", 0, "highlight visible-art padding on left and right, in points")
	verticalPadding := flag.Float64("vertical-padding-pt", 0, "highlight visible-art padding above and below, in points")
	horizontalOffset := flag.Float64("horizontal-offset-pt", 0, "explicit optical x offset applied to the image container, in points")
	verticalOffset := flag.Float64("vertical-offset-pt", 0, "visible artwork top offset below the phrase bottom, in points")
	containerAspect := flag.Float64("container-aspect", 0, "PowerPoint image-container width/height ratio")
	assetRotation := flag.Float64("asset-rotation-deg", 0, "highlight source image rotation in degrees, preserved in output")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Underline: pptxanchor --mode underline --measurement phrase.json --asset visible-bounds.json --units points --container-aspect W/H --stroke-padding-pt N --horizontal-offset-pt N --vertical-offset-pt N\nHighlight: pptxanchor --mode highlight --measurement phrase.json --asset visible-bounds.json --units points --asset-rotation-deg N --horizontal-padding-pt N --vertical-padding-pt N --horizontal-offset-pt N --vertical-offset-pt N\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if *measurementPath == "" || *assetPath == "" || *units != "points" || flag.NArg() != 0 || (*mode != "underline" && *mode != "highlight") || (*mode == "underline" && *containerAspect == 0) {
		flag.Usage()
		os.Exit(2)
	}
	if !finite(*padding) || !finite(*horizontalPadding) || !finite(*verticalPadding) || !finite(*horizontalOffset) || !finite(*verticalOffset) || !finite(*assetRotation) || (*mode == "underline" && (!finite(*containerAspect) || *containerAspect <= 0)) {
		fail(errors.New("padding, offsets and rotation must be finite; underline container aspect must be finite and positive"))
	}
	if *mode == "highlight" {
		rotationWasSet := false
		flag.Visit(func(f *flag.Flag) {
			if f.Name == "asset-rotation-deg" {
				rotationWasSet = true
			}
		})
		if !rotationWasSet {
			fail(errors.New("highlight mode requires explicit --asset-rotation-deg"))
		}
		if *horizontalPadding < 0 || *verticalPadding < 0 {
			fail(errors.New("highlight padding must be nonnegative"))
		}
	}
	measurement, err := readJSON[phraseMeasurement](*measurementPath)
	if err != nil {
		fail(err)
	}
	asset, err := readJSON[assetBounds](*assetPath)
	if err != nil {
		fail(err)
	}
	var out result
	if *mode == "highlight" {
		out, err = calculateHighlight(measurement, asset, *horizontalPadding, *verticalPadding, *horizontalOffset, *verticalOffset, *assetRotation)
	} else {
		out, err = calculate(measurement, asset, *padding, *horizontalOffset, *verticalOffset, *containerAspect)
	}
	if err != nil {
		fail(err)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(out); err != nil {
		fail(err)
	}
}

func calculate(m phraseMeasurement, a assetBounds, padding, horizontalOffset, verticalOffset, aspect float64) (result, error) {
	if m.Phrase == "" {
		return result{}, errors.New("measurement has an empty phrase")
	}
	if !finite(m.Bounds.Left) || !finite(m.Bounds.Top) || !finite(m.Bounds.Width) || !finite(m.Bounds.Height) || m.Bounds.Width <= 0 || m.Bounds.Height <= 0 {
		return result{}, errors.New("phrase bounds must have finite coordinates and positive width and height")
	}
	if len(a.ViewBox) != 4 || !allFinite(a.ViewBox) || a.ViewBox[2] <= 0 || a.ViewBox[3] <= 0 {
		return result{}, errors.New("asset viewBox must contain finite x,y,width,height with positive dimensions")
	}
	if !finite(a.VisibleBounds.X) || !finite(a.VisibleBounds.Y) || !finite(a.VisibleBounds.Width) || !finite(a.VisibleBounds.Height) || a.VisibleBounds.Width <= 0 || a.VisibleBounds.Height <= 0 {
		return result{}, errors.New("asset visibleBounds must have finite coordinates and positive width and height")
	}
	vbX, vbY, vbW, vbH := a.ViewBox[0], a.ViewBox[1], a.ViewBox[2], a.ViewBox[3]
	visible := rect{
		X:      (a.VisibleBounds.X - vbX) / vbW,
		Y:      (a.VisibleBounds.Y - vbY) / vbH,
		Width:  a.VisibleBounds.Width / vbW,
		Height: a.VisibleBounds.Height / vbH,
	}
	if visible.X < 0 || visible.Y < 0 || visible.Width <= 0 || visible.Height <= 0 || visible.X+visible.Width > 1+1e-9 || visible.Y+visible.Height > 1+1e-9 {
		return result{}, errors.New("asset visibleBounds must be contained within its viewBox")
	}

	base := result{
		Mode:            "underline",
		Intent:          "place visible asset artwork below the measured phrase",
		Phrase:          m.Phrase,
		Slide:           m.Slide,
		ShapeIndex:      m.ShapeIndex,
		Asset:           a.File,
		PhraseBounds:    &rect{X: m.Bounds.Left, Y: m.Bounds.Top, Width: m.Bounds.Width, Height: m.Bounds.Height},
		VisibleFraction: &visible,
	}
	if m.Rotation == nil {
		base.Status = "manual_required"
		base.OperatorNote = "Phrase rotation is absent from the measurement; inspect and provide rotation_degrees before placing. No image coordinates were guessed."
		return base, nil
	}
	if !finite(*m.Rotation) {
		return result{}, errors.New("phrase rotation must be finite")
	}
	if math.Abs(math.Remainder(*m.Rotation, 360)) > 1e-7 {
		base.Status = "manual_required"
		base.OperatorNote = fmt.Sprintf("Phrase rotation is %.6g degrees; rotated anchoring is not implemented. Place this asset manually.", *m.Rotation)
		return base, nil
	}
	if len(m.Lines) != 1 {
		base.Status = "manual_required"
		base.OperatorNote = fmt.Sprintf("Phrase spans %d measured line fragments; no guessed union placement was produced. Place the asset manually.", len(m.Lines))
		return base, nil
	}

	visibleTargetWidthPt := m.Bounds.Width + 2*padding
	if !finite(visibleTargetWidthPt) || visibleTargetWidthPt <= 0 {
		return result{}, errors.New("phrase width plus twice the stroke padding must be positive")
	}
	imageWidthPt := visibleTargetWidthPt / visible.Width
	imageHeightPt := imageWidthPt / aspect
	imageXPt := m.Bounds.Left - padding + horizontalOffset - visible.X*imageWidthPt
	visibleTopPt := m.Bounds.Top + m.Bounds.Height + verticalOffset
	imageYPt := visibleTopPt - visible.Y*imageHeightPt
	if !finite(imageWidthPt) || !finite(imageHeightPt) || imageWidthPt <= 0 || imageHeightPt <= 0 {
		return result{}, errors.New("computed image dimensions must be finite and positive")
	}
	placement := placementEMU{
		X:      int64(math.Round(imageXPt * emuPerPoint)),
		Y:      int64(math.Round(imageYPt * emuPerPoint)),
		Width:  int64(math.Round(imageWidthPt * emuPerPoint)),
		Height: int64(math.Round(imageHeightPt * emuPerPoint)),
	}
	base.Status = "placed"
	base.Placement = &placement
	base.OperatorNote = fmt.Sprintf("Container aspect %.9g (width/height); alpha-visible bounds aligned with %.6g pt padding per side, %.6g pt horizontal offset, and %.6g pt vertical offset. The geometry is in PowerPoint points, converted at 12,700 EMU/point. Rotation is not applied.", aspect, padding, horizontalOffset, verticalOffset)
	return base, nil
}

func calculateHighlight(m phraseMeasurement, a assetBounds, horizontalPadding, verticalPadding, horizontalOffset, verticalOffset, rotationDegrees float64) (result, error) {
	visible, err := normalizedVisibleBounds(a)
	if err != nil {
		return result{}, err
	}
	if m.Phrase == "" {
		return result{}, errors.New("measurement has an empty phrase")
	}
	if !finite(m.Bounds.Left) || !finite(m.Bounds.Top) || !finite(m.Bounds.Width) || !finite(m.Bounds.Height) || m.Bounds.Width <= 0 || m.Bounds.Height <= 0 {
		return result{}, errors.New("phrase bounds must have finite coordinates and positive width and height")
	}
	if !finite(horizontalPadding) || !finite(verticalPadding) || horizontalPadding < 0 || verticalPadding < 0 || !finite(horizontalOffset) || !finite(verticalOffset) || !finite(rotationDegrees) {
		return result{}, errors.New("highlight padding must be nonnegative and offsets/rotation finite")
	}
	base := result{
		Mode:            "highlight",
		Intent:          "fit visible highlight artwork around the measured phrase bounds",
		Phrase:          m.Phrase,
		Slide:           m.Slide,
		ShapeIndex:      m.ShapeIndex,
		Asset:           a.File,
		PhraseBounds:    &rect{X: m.Bounds.Left, Y: m.Bounds.Top, Width: m.Bounds.Width, Height: m.Bounds.Height},
		VisibleFraction: &visible,
		ZOrderIntent:    "behind_target_text",
	}
	if m.Rotation == nil {
		base.Status = "manual_required"
		base.OperatorNote = "Phrase rotation is absent from the measurement; inspect and provide rotation_degrees before placing. No image coordinates were guessed."
		return base, nil
	}
	if !finite(*m.Rotation) {
		return result{}, errors.New("phrase rotation must be finite")
	}
	if math.Abs(math.Remainder(*m.Rotation, 360)) > 1e-7 {
		base.Status = "manual_required"
		base.OperatorNote = fmt.Sprintf("Phrase rotation is %.6g degrees; highlight anchoring assumes a horizontal phrase. Place this asset manually.", *m.Rotation)
		return base, nil
	}
	if len(m.Lines) != 1 {
		base.Status = "manual_required"
		base.OperatorNote = fmt.Sprintf("Phrase spans %d measured line fragments; no guessed union placement was produced. Place the asset manually.", len(m.Lines))
		return base, nil
	}

	target := rect{
		X:      m.Bounds.Left - horizontalPadding + horizontalOffset,
		Y:      m.Bounds.Top - verticalPadding + verticalOffset,
		Width:  m.Bounds.Width + 2*horizontalPadding,
		Height: m.Bounds.Height + 2*verticalPadding,
	}
	if !finite(target.X) || !finite(target.Y) || !finite(target.Width) || !finite(target.Height) || target.Width <= 0 || target.Height <= 0 {
		return result{}, errors.New("target visible highlight dimensions must be finite and positive")
	}
	angle := rotationDegrees * math.Pi / 180
	c, s := math.Abs(math.Cos(angle)), math.Abs(math.Sin(angle))
	determinant := c*c - s*s
	if math.Abs(determinant) < 1e-10 {
		base.Status = "manual_required"
		base.OperatorNote = "Asset rotation makes axis-aligned extent fitting singular or unstable; place this asset manually."
		return base, nil
	}
	// The rotated alpha rectangle has AABB dimensions:
	// W = |cos|*alphaWidth + |sin|*alphaHeight, and
	// H = |sin|*alphaWidth + |cos|*alphaHeight. Solve that system for the
	// unrotated visible rectangle, then divide by normalized alpha fractions.
	alphaWidth := (c*target.Width - s*target.Height) / determinant
	alphaHeight := (c*target.Height - s*target.Width) / determinant
	if !finite(alphaWidth) || !finite(alphaHeight) || alphaWidth <= 0 || alphaHeight <= 0 {
		return result{}, errors.New("rotated target bounds imply nonpositive source visible dimensions")
	}
	imageWidthPt := alphaWidth / visible.Width
	imageHeightPt := alphaHeight / visible.Height
	if !finite(imageWidthPt) || !finite(imageHeightPt) || imageWidthPt <= 0 || imageHeightPt <= 0 {
		return result{}, errors.New("computed image dimensions must be finite and positive")
	}
	// Preserve the supplied art rotation while centering its rotated alpha box
	// on the target phrase-plus-padding rectangle.
	alphaCenterDX := (visible.X + visible.Width/2 - 0.5) * imageWidthPt
	alphaCenterDY := (visible.Y + visible.Height/2 - 0.5) * imageHeightPt
	rotatedDX := alphaCenterDX*math.Cos(angle) - alphaCenterDY*math.Sin(angle)
	rotatedDY := alphaCenterDX*math.Sin(angle) + alphaCenterDY*math.Cos(angle)
	targetCenterX, targetCenterY := target.X+target.Width/2, target.Y+target.Height/2
	imageXPt := targetCenterX - rotatedDX - imageWidthPt/2
	imageYPt := targetCenterY - rotatedDY - imageHeightPt/2
	placement := placementEMU{
		X:      int64(math.Round(imageXPt * emuPerPoint)),
		Y:      int64(math.Round(imageYPt * emuPerPoint)),
		Width:  int64(math.Round(imageWidthPt * emuPerPoint)),
		Height: int64(math.Round(imageHeightPt * emuPerPoint)),
	}
	rotationOOXML := int64(math.Round(rotationDegrees * 60000))
	base.Status = "placed"
	base.TargetVisible = &target
	base.Placement = &placement
	base.RotationDegrees = &rotationDegrees
	base.RotationOOXML = &rotationOOXML
	base.OperatorNote = fmt.Sprintf("Visible alpha AABB fits phrase bounds with %.6g pt horizontal and %.6g pt vertical padding plus explicit optical offsets %.6g/%.6g pt. Preserve %.6g degree image rotation; keep image behind the target text. Image width and height were solved independently from normalized alpha fractions.", horizontalPadding, verticalPadding, horizontalOffset, verticalOffset, rotationDegrees)
	return base, nil
}

func normalizedVisibleBounds(a assetBounds) (rect, error) {
	if len(a.ViewBox) != 4 || !allFinite(a.ViewBox) || a.ViewBox[2] <= 0 || a.ViewBox[3] <= 0 {
		return rect{}, errors.New("asset viewBox must contain finite x,y,width,height with positive dimensions")
	}
	if !finite(a.VisibleBounds.X) || !finite(a.VisibleBounds.Y) || !finite(a.VisibleBounds.Width) || !finite(a.VisibleBounds.Height) || a.VisibleBounds.Width <= 0 || a.VisibleBounds.Height <= 0 {
		return rect{}, errors.New("asset visibleBounds must have finite coordinates and positive width and height")
	}
	vbX, vbY, vbW, vbH := a.ViewBox[0], a.ViewBox[1], a.ViewBox[2], a.ViewBox[3]
	visible := rect{
		X:      (a.VisibleBounds.X - vbX) / vbW,
		Y:      (a.VisibleBounds.Y - vbY) / vbH,
		Width:  a.VisibleBounds.Width / vbW,
		Height: a.VisibleBounds.Height / vbH,
	}
	if visible.X < 0 || visible.Y < 0 || visible.Width <= 0 || visible.Height <= 0 || visible.X+visible.Width > 1+1e-9 || visible.Y+visible.Height > 1+1e-9 {
		return rect{}, errors.New("asset visibleBounds must be contained within its viewBox")
	}
	return visible, nil
}

func readJSON[T any](path string) (T, error) {
	var value T
	file, err := os.Open(path)
	if err != nil {
		return value, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&value); err != nil {
		return value, fmt.Errorf("decode %s: %w", path, err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return value, fmt.Errorf("decode %s: expected one JSON value", path)
	}
	return value, nil
}

func allFinite(values []float64) bool {
	for _, value := range values {
		if !finite(value) {
			return false
		}
	}
	return true
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

func fail(err error) {
	fmt.Fprintln(os.Stderr, "pptxanchor:", err)
	os.Exit(2)
}
