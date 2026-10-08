package deckproject

import (
	"fmt"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

// CheckSlideFit checks selected slides with the build renderer. Native visual
// acceptance is separate; no generated output or source is persisted here.
func CheckSlideFit(p *Project, ids []string, bundle, engine string) error {
	c, err := Compile(p, bundle, engine)
	if err != nil {
		return err
	}
	selected := map[string]bool{}
	for _, id := range ids {
		selected[id] = true
	}
	var slides []wmdesign.SlideSpec
	for _, slide := range c.Document.Slides {
		if selected[slide.ID] {
			slides = append(slides, slide)
			delete(selected, slide.ID)
		}
	}
	if len(selected) != 0 || len(slides) == 0 {
		return fmt.Errorf("fit check requires known slide IDs")
	}
	c.Document.Slides, c.Document.Sections = slides, nil
	native, report, err := wmdesign.BuildWithEngineAndAssets(bundle, "", c.Document, engine, c.Assets)
	if err != nil {
		return fmt.Errorf("slide fit check failed; source unchanged: %w", err)
	}
	if _, err = applyNativeGeometry(native, c.Document, &report); err != nil {
		return fmt.Errorf("native geometry fit check failed; source unchanged: %w", err)
	}
	return nil
}
