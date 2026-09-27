package compose

import "fmt"

// AccentSpec attaches artwork to a complete, single-line canvas text block.
// Alpha bounds are normalized source-image coordinates, not the image frame.
// Phrase fragments inside rich text still use pptxanchor's native phrase adapter.
type AccentSpec struct {
	Phrase              string         `json:"phrase,omitempty"`
	Occurrence          int            `json:"occurrence,omitempty"`
	StartRune           *int           `json:"start_rune,omitempty"`
	EndRune             *int           `json:"end_rune,omitempty"`
	Multiline           string         `json:"multiline,omitempty"`
	ContainerAspect     float64        `json:"container_aspect,omitempty"`
	RotationDeg         float64        `json:"rotation_deg,omitempty"`
	FallbackAssetPath   string         `json:"fallback_asset_path,omitempty"`
	FallbackAssetSHA256 string         `json:"fallback_asset_sha256,omitempty"`
	Staging             *AccentStaging `json:"staging,omitempty"`
	ID                  string         `json:"id"`
	Target              string         `json:"target"`
	Mode                string         `json:"mode"`
	AssetPath           string         `json:"asset_path"`
	AssetSHA256         string         `json:"asset_sha256"`
	AlphaBounds         Rect           `json:"alpha_bounds"`
	PaddingXPt          float64        `json:"padding_x_pt"`
	PaddingYPt          float64        `json:"padding_y_pt"`
	OffsetXPt           float64        `json:"offset_x_pt"`
	OffsetYPt           float64        `json:"offset_y_pt"`
	StrokeHeightPt      float64        `json:"stroke_height_pt,omitempty"`
}
type AccentStaging struct {
	AssetBounds Rect   `json:"asset_bounds"`
	NoteBounds  Rect   `json:"note_bounds"`
	Note        string `json:"note"`
}
type PlannedAccent struct {
	TargetMeasurementID string `json:"target_measurement_id,omitempty"`
	SourceAccentID      string `json:"source_accent_id,omitempty"`
	FragmentIndex       int    `json:"fragment_index,omitempty"`
	Status              string `json:"status"`
	Reason              string `json:"reason,omitempty"`
	AccentSpec
	Bounds        Rect `json:"bounds"`
	VisibleBounds Rect `json:"visible_bounds"`
	TargetBounds  Rect `json:"target_bounds"`
}

func validateAccents(s SlideSpec, ids map[string]bool) error {
	targets := map[string]CanvasSpec{}
	for _, c := range s.Canvas {
		targets[c.ID] = c
	}
	for _, a := range s.Accents {
		if !validID(a.ID) || ids[a.ID] {
			return fmt.Errorf("accent ID duplicated/empty: %s", a.ID)
		}
		ids[a.ID] = true
		t, ok := targets[a.Target]
		if !ok || t.Kind != "text" || t.Valign != "top" || t.Background != "" {
			return fmt.Errorf("accent %s requires a top-aligned canvas text target", a.ID)
		}
		if a.Mode != "underline" && a.Mode != "highlight" {
			return fmt.Errorf("accent %s mode must be underline or highlight", a.ID)
		}
		if a.AssetPath == "" || len(a.AssetSHA256) != 64 || !validRect(a.AlphaBounds) || !inside(a.AlphaBounds, Rect{Width: 1, Height: 1}) {
			return fmt.Errorf("accent %s requires pinned asset and normalized alpha bounds", a.ID)
		}
		if err := validatePhraseAccent(a, t, s); err != nil {
			return err
		}
		if !finite(a.PaddingXPt) || !finite(a.PaddingYPt) || !finite(a.OffsetXPt) || !finite(a.OffsetYPt) || (a.Mode == "underline" && !positive(a.StrokeHeightPt)) {
			return fmt.Errorf("accent %s invalid optical calibration", a.ID)
		}
	}
	return nil
}
func checkAccentCollisions(a AccentSpec, v Rect, p *PlannedSlide) error {
	for _, c := range p.Canvas {
		if c.ID != a.Target && (c.Kind == "text" || c.Kind == "image") && overlap(v, c.Bounds) {
			return fmt.Errorf("accent %s visible ink collides with %s", a.ID, c.ID)
		}
	}
	type item struct {
		id string
		r  Rect
	}
	var obstacles []item
	for _, c := range p.Accents {
		obstacles = append(obstacles, item{c.ID, c.VisibleBounds})
	}
	for _, c := range p.ArtworkArrows {
		for _, r := range artworkInkRegions(c) {
			obstacles = append(obstacles, item{c.ID, r})
		}
	}
	if p.TitleMeasurementID != "" {
		obstacles = append(obstacles, item{"slide-title", p.TitleBounds})
	}
	for _, c := range p.Cards {
		obstacles = append(obstacles, item{c.ID, c.Bounds})
	}
	for _, c := range p.Pods {
		obstacles = append(obstacles, item{c.ID, c.Bounds})
	}
	for _, c := range p.Roles {
		obstacles = append(obstacles, item{c.ID, c.Bounds})
	}
	for _, c := range p.Phases {
		obstacles = append(obstacles, item{c.ID, c.Bounds})
	}
	if p.Legend != nil {
		obstacles = append(obstacles, item{"staffing-legend", p.Legend.Bounds})
	}
	for _, o := range obstacles {
		if overlap(v, o.r) {
			return fmt.Errorf("accent %s visible ink collides with %s", a.ID, o.id)
		}
	}
	for _, c := range p.Connections {
		for i := 1; i < len(c.Points); i++ {
			if segmentHits(c.Points[i-1], c.Points[i], expand(v, c.WidthPt/2)) {
				return fmt.Errorf("accent %s collides with reporting line %s", a.ID, c.ID)
			}
		}
	}
	return nil
}
