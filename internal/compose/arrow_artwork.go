package compose

import (
	"fmt"
	"math"
)

// ArrowArtwork preserves the original aspect ratio. Tail/tip and tangents are
// measured from visible SVG geometry, never inferred from the picture frame.
type ArrowArtwork struct {
	ID                   string  `json:"id"`
	AssetPath            string  `json:"asset_path"`
	AssetSHA256          string  `json:"asset_sha256"`
	FallbackAssetPath    string  `json:"fallback_asset_path,omitempty"`
	FallbackAssetSHA256  string  `json:"fallback_asset_sha256,omitempty"`
	IntrinsicWidth       float64 `json:"intrinsic_width"`
	IntrinsicHeight      float64 `json:"intrinsic_height"`
	AlphaBounds          Rect    `json:"alpha_bounds"`
	InkRegions           []Rect  `json:"ink_regions,omitempty"`
	InkRegionsProvenance string  `json:"ink_regions_provenance,omitempty"`
	Tail                 Point   `json:"tail"`
	Tip                  Point   `json:"tip"`
	TailTangent          Point   `json:"tail_tangent"`
	TipTangent           Point   `json:"tip_tangent"`
	MinSpanPt            float64 `json:"min_span_pt"`
	MaxSpanPt            float64 `json:"max_span_pt"`
	MaxRotationDeltaDeg  float64 `json:"max_rotation_delta_deg"`
	EndpointProvenance   string  `json:"endpoint_provenance"`
}
type ArtworkAnchor struct {
	Target    string         `json:"target"`
	Port      string         `json:"port"`
	GapPt     float64        `json:"gap_pt,omitempty"`
	OffsetXPt float64        `json:"offset_x_pt,omitempty"`
	OffsetYPt float64        `json:"offset_y_pt,omitempty"`
	Phrase    *PhraseRequest `json:"phrase,omitempty"`
}
type ArtworkArrowSpec struct {
	ID          string         `json:"id"`
	From        ArtworkAnchor  `json:"from"`
	To          ArtworkAnchor  `json:"to"`
	Artwork     ArrowArtwork   `json:"artwork"`
	ClearancePt float64        `json:"clearance_pt"`
	Staging     *AccentStaging `json:"staging,omitempty"`
}
type PlannedArtworkArrow struct {
	ArtworkArrowSpec
	Bounds         Rect    `json:"bounds"`
	VisibleBounds  Rect    `json:"visible_bounds"`
	VisibleRegions []Rect  `json:"visible_regions,omitempty"`
	RotationDeg    float64 `json:"rotation_deg"`
	TailPoint      Point   `json:"tail_point"`
	TipPoint       Point   `json:"tip_point"`
	Status         string  `json:"status"`
	Reason         string  `json:"reason,omitempty"`
}

func arrowPhraseRequests(s SlideSpec, target string) []PhraseRequest {
	var out []PhraseRequest
	for _, a := range s.ArtworkArrows {
		for i, b := range []ArtworkAnchor{a.From, a.To} {
			if b.Target == target && b.Phrase != nil {
				q := *b.Phrase
				q.ID = fmt.Sprintf("%s/anchor-%d", a.ID, i)
				out = append(out, q)
			}
		}
	}
	return out
}
func arrowNoteProbe(s SlideSpec, a ArtworkArrowSpec) ProbeRequest {
	return ProbeRequest{ID: requestID(s.ID, "arrow-note", a.ID), SlideID: s.ID, Kind: "arrow_note", Text: "MANUAL PLACEMENT: " + a.Staging.Note, TextWidthPt: a.Staging.NoteBounds.Width, FontFace: "Arial", FontSizePt: 11, Foreground: "#070154", Background: white, Align: "left"}
}
func validateArtworkArrows(s SlideSpec, ids map[string]bool) error {
	for _, a := range s.ArtworkArrows {
		if !validID(a.ID) || ids[a.ID] {
			return fmt.Errorf("duplicate/empty artwork arrow ID %s", a.ID)
		}
		ids[a.ID] = true
		art := a.Artwork
		if art.ID == "" || art.AssetPath == "" || len(art.AssetSHA256) != 64 || art.EndpointProvenance == "" || !positive(art.IntrinsicWidth) || !positive(art.IntrinsicHeight) || !validRect(art.AlphaBounds) || !inside(art.AlphaBounds, Rect{Width: 1, Height: 1}) {
			return fmt.Errorf("arrow %s requires pinned artwork and measured alpha/endpoint provenance", a.ID)
		}
		for _, v := range []Point{art.Tail, art.Tip} {
			if !finite(v.X) || !finite(v.Y) || v.X < 0 || v.X > 1 || v.Y < 0 || v.Y > 1 {
				return fmt.Errorf("arrow %s endpoint outside asset", a.ID)
			}
		}
		if len(art.InkRegions) > 4096 || (len(art.InkRegions) > 0 && art.InkRegionsProvenance == "") {
			return fmt.Errorf("arrow %s ink regions need bounded count and source provenance", a.ID)
		}
		for _, r := range art.InkRegions {
			if !validRect(r) || !inside(r, Rect{Width: 1, Height: 1}) {
				return fmt.Errorf("arrow %s ink region outside artwork", a.ID)
			}
		}
		if samePoint(art.Tail, art.Tip) || !positive(art.MinSpanPt) || !positive(art.MaxSpanPt) || art.MinSpanPt > art.MaxSpanPt || !nonnegative(art.MaxRotationDeltaDeg) || art.MaxRotationDeltaDeg > 180 || !nonnegative(a.ClearancePt) {
			return fmt.Errorf("arrow %s invalid span/rotation/clearance envelope", a.ID)
		}
		for _, v := range []Point{art.TailTangent, art.TipTangent} {
			if !finite(v.X) || !finite(v.Y) || math.Abs(math.Hypot(v.X, v.Y)-1) > .01 {
				return fmt.Errorf("arrow %s needs measured unit endpoint tangents", a.ID)
			}
		}
		for _, v := range []ArtworkAnchor{a.From, a.To} {
			if v.Target == "" || !nonnegative(v.GapPt) || !finite(v.OffsetXPt) || !finite(v.OffsetYPt) {
				return fmt.Errorf("arrow %s invalid anchor", a.ID)
			}
			if _, _, err := anchor(Rect{}, v.Port); err != nil {
				return err
			}
		}
		if a.Staging != nil {
			z := a.Staging
			slide := Rect{Width: s.WidthPt, Height: s.HeightPt}
			if !validRect(z.AssetBounds) || !validRect(z.NoteBounds) || !inside(z.AssetBounds, slide) || !inside(z.NoteBounds, slide) || overlap(z.AssetBounds, z.NoteBounds) || z.Note == "" {
				return fmt.Errorf("arrow %s requires separate visible staging asset/note zones", a.ID)
			}
		}
	}
	return nil
}
func artworkAnchorPoint(a ArtworkArrowSpec, index int, p *PlannedSlide, m Measurements) (Point, Point, error) {
	v := a.From
	if index == 1 {
		v = a.To
	}
	var b Rect
	found := false
	for _, c := range p.Canvas {
		if c.ID == v.Target {
			b = c.Bounds
			found = true
			if v.Phrase != nil {
				fragments := m.ByRequestID[c.MeasurementID].PhraseBounds[fmt.Sprintf("%s/anchor-%d", a.ID, index)]
				if len(fragments) != 1 {
					return Point{}, Point{}, fmt.Errorf("phrase anchor %s is ambiguous or multiline", v.Target)
				}
				b = fragments[0]
				b.X += c.Bounds.X + c.InsetX
				b.Y += c.Bounds.Y + c.InsetY
			}
		}
	}
	if v.Phrase == nil {
		for _, c := range p.LayoutPorts {
			if c.ID == v.Target {
				b = c.Bounds
				found = true
			}
		}
		for _, c := range p.Pods {
			if c.ID == v.Target {
				b = c.Bounds
				found = true
			}
		}
		for _, c := range p.Roles {
			if c.ID == v.Target {
				b = c.Bounds
				found = true
			}
		}
	}
	if !found {
		return Point{}, Point{}, fmt.Errorf("unknown arrow anchor %s", v.Target)
	}
	at, n, err := anchor(b, v.Port)
	if err != nil {
		return at, n, err
	}
	at.X += n.X*v.GapPt + v.OffsetXPt
	at.Y += n.Y*v.GapPt + v.OffsetYPt
	return at, n, nil
}

func solveArtworkArrow(a ArtworkArrowSpec, from, to, fromNormal, toNormal Point) (PlannedArtworkArrow, error) {
	out := PlannedArtworkArrow{ArtworkArrowSpec: a, TailPoint: from, TipPoint: to, Status: "placed"}
	art := a.Artwork
	dx, dy := to.X-from.X, to.Y-from.Y
	span := math.Hypot(dx, dy)
	if span < art.MinSpanPt || span > art.MaxSpanPt {
		return out, fmt.Errorf("span %.2fpt outside curated [%.2f,%.2f]pt", span, art.MinSpanPt, art.MaxSpanPt)
	}
	sx, sy := (art.Tip.X-art.Tail.X)*art.IntrinsicWidth, (art.Tip.Y-art.Tail.Y)*art.IntrinsicHeight
	scale := span / math.Hypot(sx, sy)
	angle := math.Remainder((math.Atan2(dy, dx)-math.Atan2(sy, sx))*180/math.Pi, 360)
	if math.Abs(angle) > art.MaxRotationDeltaDeg {
		return out, fmt.Errorf("rotation %.2f° exceeds curated ±%.2f°", angle, art.MaxRotationDeltaDeg)
	}
	theta := angle * math.Pi / 180
	c, s := math.Cos(theta), math.Sin(theta)
	rotate := func(v Point) Point { return Point{c*v.X - s*v.Y, s*v.X + c*v.Y} }
	tailTan, tipTan := rotate(art.TailTangent), rotate(art.TipTangent)
	if tailTan.X*fromNormal.X+tailTan.Y*fromNormal.Y < .15 || tipTan.X*(-toNormal.X)+tipTan.Y*(-toNormal.Y) < .15 {
		return out, fmt.Errorf("artwork tangents do not face away from source and into target; choose another variant/ports")
	}
	w, h := art.IntrinsicWidth*scale, art.IntrinsicHeight*scale
	offset := rotate(Point{(art.Tail.X - .5) * w, (art.Tail.Y - .5) * h})
	cx, cy := from.X-offset.X, from.Y-offset.Y
	out.Bounds = Rect{X: cx - w/2, Y: cy - h/2, Width: w, Height: h}
	out.RotationDeg = angle
	box := art.AlphaBounds
	minX, minY, maxX, maxY := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, q := range []Point{{box.X, box.Y}, {box.X + box.Width, box.Y}, {box.X, box.Y + box.Height}, {box.X + box.Width, box.Y + box.Height}} {
		v := rotate(Point{(q.X - .5) * w, (q.Y - .5) * h})
		v.X += cx
		v.Y += cy
		minX = math.Min(minX, v.X)
		minY = math.Min(minY, v.Y)
		maxX = math.Max(maxX, v.X)
		maxY = math.Max(maxY, v.Y)
	}
	out.VisibleBounds = Rect{X: minX, Y: minY, Width: maxX - minX, Height: maxY - minY}
	for _, region := range art.InkRegions {
		minX, minY, maxX, maxY := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
		for _, q := range []Point{{region.X, region.Y}, {region.X + region.Width, region.Y}, {region.X, region.Y + region.Height}, {region.X + region.Width, region.Y + region.Height}} {
			v := rotate(Point{(q.X - .5) * w, (q.Y - .5) * h})
			minX, minY = math.Min(minX, cx+v.X), math.Min(minY, cy+v.Y)
			maxX, maxY = math.Max(maxX, cx+v.X), math.Max(maxY, cy+v.Y)
		}
		out.VisibleRegions = append(out.VisibleRegions, Rect{X: minX, Y: minY, Width: maxX - minX, Height: maxY - minY})
	}
	return out, nil
}

func artworkInkRegions(a PlannedArtworkArrow) []Rect {
	if len(a.VisibleRegions) > 0 {
		return a.VisibleRegions
	}
	return []Rect{a.VisibleBounds}
}
func planArtworkArrows(s SlideSpec, p *PlannedSlide, m Measurements) error {
	for _, a := range s.ArtworkArrows {
		from, fn, err := artworkAnchorPoint(a, 0, p, m)
		to, tn, toErr := artworkAnchorPoint(a, 1, p, m)
		if err == nil {
			err = toErr
		}
		var out PlannedArtworkArrow
		if err == nil {
			out, err = solveArtworkArrow(a, from, to, fn, tn)
		}
		if err == nil && !inside(out.VisibleBounds, Rect{Width: s.WidthPt, Height: s.HeightPt}) {
			err = fmt.Errorf("visible artwork exceeds slide")
		}
		if err == nil {
			for _, region := range artworkInkRegions(out) {
				if err = checkAccentCollisions(AccentSpec{ID: a.ID}, expand(region, a.ClearancePt), p); err != nil {
					break
				}
			}
		}
		if err != nil {
			if a.Staging == nil {
				return fmt.Errorf("arrow %s manual_required: %w", a.ID, err)
			}
			q := arrowNoteProbe(s, a)
			z, ok := m.ByRequestID[q.ID]
			if !ok || z.RenderedWidthPt > a.Staging.NoteBounds.Width+.01 || z.RenderedHeightPt > a.Staging.NoteBounds.Height+.01 {
				return fmt.Errorf("arrow %s manual note needs fitting native measurement", a.ID)
			}
			for _, b := range []Rect{a.Staging.AssetBounds, a.Staging.NoteBounds} {
				if e := checkAccentCollisions(AccentSpec{ID: a.ID}, b, p); e != nil {
					return e
				}
			}
			out = PlannedArtworkArrow{ArtworkArrowSpec: a, Bounds: a.Staging.AssetBounds, VisibleBounds: a.Staging.AssetBounds, Status: "manual_required", Reason: err.Error()}
			p.ManualRequired = append(p.ManualRequired, a.Staging.Note+" ("+err.Error()+")")
			p.Canvas = append(p.Canvas, PlannedCanvas{CanvasSpec: CanvasSpec{ID: a.ID + "/manual-note", Kind: "text", Bounds: a.Staging.NoteBounds, Text: q.Text, FontFace: "Arial", FontSizePt: 11, Foreground: "#070154", Align: "left", Valign: "top", Layer: 100}, MeasurementID: q.ID})
		}
		p.ArtworkArrows = append(p.ArtworkArrows, out)
	}
	return nil
}

// VerifyArtworkAnchors rechecks planned endpoints against final native text
// measurements. A fitting text box alone does not prove its selected phrase
// stayed in the same place after rendering the complete deck.
func VerifyArtworkAnchors(p *PlannedSlide, m Measurements, tolerance float64) error {
	for _, a := range p.ArtworkArrows {
		if a.Status != "placed" {
			continue
		}
		for i, expected := range []Point{a.TailPoint, a.TipPoint} {
			actual, _, err := artworkAnchorPoint(a.ArtworkArrowSpec, i, p, m)
			if err != nil {
				return fmt.Errorf("native arrow %s: %w", a.ID, err)
			}
			if math.Abs(actual.X-expected.X) > tolerance || math.Abs(actual.Y-expected.Y) > tolerance {
				return fmt.Errorf("native arrow %s endpoint %d shifted: %+v expected %+v", a.ID, i, actual, expected)
			}
		}
	}
	return nil
}
