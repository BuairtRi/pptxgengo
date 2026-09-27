// Package adapt lowers semantic slide-family inputs into measured compose specs.
// It does not rewrite imported native templates or estimate text capacity.
package adapt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/buairtri/pptxgengo/internal/compose"
)

const Schema = "pptxgengo.adaptive-deck.v1"

type Spec struct {
	Schema string  `json:"schema"`
	Slides []Slide `json:"slides"`
}
type Slide struct {
	ID         string          `json:"id"`
	Family     string          `json:"family"`
	TemplateID string          `json:"template_id,omitempty"`
	Title      string          `json:"title"`
	Footer     string          `json:"footer,omitempty"`
	Role       string          `json:"role"`
	Takeaway   string          `json:"takeaway"`
	Bounds     *compose.Rect   `json:"bounds,omitempty"`
	Style      StyleOptions    `json:"style,omitempty"`
	Content    json.RawMessage `json:"content"`
}
type StyleOptions struct {
	Profile     string            `json:"profile,omitempty"`
	FontFace    string            `json:"font_face,omitempty"`
	TitleFontPt float64           `json:"title_font_pt,omitempty"`
	BodyFontPt  float64           `json:"body_font_pt,omitempty"`
	LabelFontPt float64           `json:"label_font_pt,omitempty"`
	Colors      map[string]string `json:"colors,omitempty"`
}
type Style struct {
	Profile, FontFace                                                                     string
	TitleFontPt, BodyFontPt, LabelFontPt                                                  float64
	Text, Secondary, Surface, MutedSurface, Accent, Active, Inactive, Border, Navy, White string
}
type Context struct {
	Slide  Slide
	Bounds compose.Rect
	Style  Style
}
type Content struct {
	Canvas      []compose.CanvasSpec
	Layouts     []compose.ContainerSpec
	Pods        []compose.PodSpec
	Roles       []compose.StandaloneRoleSpec
	Connections []compose.ConnectionSpec
	Paths       []compose.ProcessPathSpec
	Legend      *compose.LegendSpec
	Controls    map[string]any
	Limitations []string
}
type SlideReport struct {
	ID            string         `json:"id"`
	Family        string         `json:"family"`
	TemplateID    string         `json:"template_id,omitempty"`
	Controls      map[string]any `json:"controls"`
	Limitations   []string       `json:"limitations"`
	Qualification string         `json:"qualification"`
}
type Report struct {
	Schema string        `json:"schema"`
	Slides []SlideReport `json:"slides"`
}

func Decode(data []byte, into any) error {
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(into); err != nil {
		return err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("trailing JSON input")
	}
	return nil
}
