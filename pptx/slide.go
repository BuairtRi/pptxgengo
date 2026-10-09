// slide.go ports src/slide.ts: the per-slide user API. A Slide is a thin
// wrapper around a *PresSlide (the data model in types.go) plus a back-pointer
// to its Presentation (needed for table auto-paging callbacks). Every method
// delegates to the gen-objects package functions, mirroring slide.ts exactly.
package pptx

import (
	"errors"
	"fmt"
)

// Slide is the user-facing handle for a single slide. Obtain one from
// Presentation.AddSlide. Ports the TS Slide class.
type Slide struct {
	ps   *PresSlide
	pres *Presentation

	// newAutoPagedSlides holds slides created by table auto-paging on the most
	// recent AddTable call (mirrors TS `_newAutoPagedSlides`).
	newAutoPagedSlides []*PresSlide
}

// PresSlide returns the underlying slide data model.
func (s *Slide) PresSlide() *PresSlide { return s.ps }

// NewAutoPagedSlides returns any slides created by the last AddTable auto-page.
func (s *Slide) NewAutoPagedSlides() []*PresSlide { return s.newAutoPagedSlides }

// AddText adds a text box (one or more runs) to the slide. The per-run styling
// lives on each TextProps.Options; box-level layout/styling lives on opts.
// Ports TS addText (array form).
func (s *Slide) AddText(text []TextProps, opts *TextPropsOptions) error {
	if opts != nil && opts.Line != nil && !validShapeLineJoin(opts.Line.LineJoin) {
		return errors.New("shape line join must be empty, round, bevel or miter")
	}
	addTextDefinition(s.ps, text, objectOptionsFromTextPropsOptions(opts), false)
	return nil
}

// AddShape adds a shape to the slide. Ports TS addShape.
func (s *Slide) AddShape(shapeName ShapeType, opts *ShapeProps) error {
	return addShapeDefinition(s.ps, shapeName, opts)
}

// AddConnector adds an attached native p:cxnSp connector. Targets can be added later;
// Write validates their unique names, rectangle kinds and connection sites.
func (s *Slide) AddConnector(opts *ConnectorProps) error {
	if opts == nil || !validConnectorConnection(opts.Connection) {
		return errors.New("connector requires distinct named endpoints and rectangle sites in [0,3]")
	}
	shape := opts.ShapeProps
	if !validConnectorBounds(shape.PositionProps, false) {
		return errors.New("native connector requires finite nonnegative bounds and distinct points")
	}
	if shape.Fill != nil || shape.Hyperlink != nil || len(shape.Points) != 0 || shape.RectRadius != 0 || shape.AngleRange != nil || len(shape.Adjustments) != 0 {
		return errors.New("native connector does not support fill, links or custom geometry")
	}
	preset := ShapeTypeLine
	if opts.Route != nil {
		if opts.Route.Preset == "polyline" {
			var e error
			preset, shape, e = NativePolylineConnectorPreset(shape, opts.Route.Points)
			if e != nil {
				return e
			}
		} else {
			if len(opts.Route.Points) != 0 || (shape.Rotate != 0 && shape.Rotate != 90) {
				return errors.New("unsupported native connector route")
			}
			switch opts.Route.Preset {
			case "bentConnector2", "bentConnector3", "bentConnector4", "bentConnector5":
				preset = ShapeType(opts.Route.Preset)
			default:
				return errors.New("unsupported native connector preset")
			}
			shape.Adjustments = map[string]int{}
			count := int(opts.Route.Preset[len(opts.Route.Preset)-1] - '2')
			for i, value := range []int{opts.Route.Adjustment, opts.Route.Adjustment2, opts.Route.Adjustment3} {
				if i < count {
					if value < -2147483647 || value > 2147483647 {
						return errors.New("invalid native connector adjustment")
					}
					shape.Adjustments[fmt.Sprintf("adj%d", i+1)] = value
				} else if value != 0 {
					return errors.New("unused native connector adjustment")
				}
			}
		}
	} else if shape.Rotate != 0 {
		return errors.New("native straight connector does not support rotation")
	}
	if err := addShapeDefinition(s.ps, preset, &shape); err != nil {
		return err
	}
	connection := opts.Connection
	s.ps.SlideObjects[len(s.ps.SlideObjects)-1].Options.NativeConnection = &connection
	return nil
}

// AddImage adds an image (by path or base64 data) to the slide. Ports TS addImage.
func (s *Slide) AddImage(opts *ImageProps) error {
	return addImageDefinition(s.ps, opts)
}

// AddMedia adds audio/video/online media to the slide. Ports TS addMedia.
func (s *Slide) AddMedia(opts *MediaProps) error {
	return addMediaDefinition(s.ps, opts)
}

// AddNotes adds speaker notes to the slide. Ports TS addNotes.
func (s *Slide) AddNotes(notes string) {
	addNotesDefinition(s.ps, notes)
}

// AddTable adds a table, auto-paging onto new slides when needed. Ports TS
// addTable. Newly created slides are also available via NewAutoPagedSlides.
func (s *Slide) AddTable(rows []TableRow, opts *TableProps) error {
	newSlides, err := addTableDefinition(
		s.ps, rows, opts,
		&s.ps.SlideLayout, s.ps.PresLayout,
		s.pres.addNewSlide, s.pres.getSlide,
	)
	if err != nil {
		return err
	}
	s.newAutoPagedSlides = newSlides
	return nil
}

// AddChart adds a single-type chart to the slide. Ports TS addChart (single
// type). For combo charts use AddMultiChart.
func (s *Slide) AddChart(chartType ChartType, data []ChartData, opts *ChartOptions) error {
	_, err := addChartDefinition(s.ps, &s.pres.chartCtr, chartType, nil, data, opts)
	return err
}

// AddMultiChart adds a multi-type (combo) chart to the slide. Ports the
// IChartMulti[] form of TS addChart.
func (s *Slide) AddMultiChart(multi []IChartMulti, opts *ChartOptions) error {
	_, err := addChartDefinition(s.ps, &s.pres.chartCtr, "", multi, nil, opts)
	return err
}

// Background sets the slide background (color, fill, or image). Image data/path
// is captured now (before Write). Ports the TS `background` setter.
func (s *Slide) Background(props *BackgroundProps) {
	s.ps.Background = props
	if props != nil {
		addBackgroundDefinition(props, &s.ps.SlideBaseProps)
	}
}

// SlideNumber enables/positions the slide number on this slide (and registers
// it on the master/DEFAULT layout). Ports the TS `slideNumber` setter.
func (s *Slide) SlideNumber(props *SlideNumberProps) {
	s.ps.SlideNumberProps = props
	s.pres.setSlideNumber(props)
}

// GetTextContent returns the concatenated plain text of all text objects on the
// slide (convenience accessor; not part of PptxGenJS).
func (s *Slide) GetTextContent() string {
	var out string
	for i := range s.ps.SlideObjects {
		o := &s.ps.SlideObjects[i]
		if o.Type == SlideObjectTypeText || o.Type == SlideObjectTypePlaceholder {
			for _, tp := range o.Text {
				out += tp.Text
			}
		}
	}
	return out
}
