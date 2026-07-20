// slide.go ports src/slide.ts: the per-slide user API. A Slide is a thin
// wrapper around a *PresSlide (the data model in types.go) plus a back-pointer
// to its Presentation (needed for table auto-paging callbacks). Every method
// delegates to the gen-objects package functions, mirroring slide.ts exactly.
package pptx

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
	addTextDefinition(s.ps, text, objectOptionsFromTextPropsOptions(opts), false)
	return nil
}

// AddShape adds a shape to the slide. Ports TS addShape.
func (s *Slide) AddShape(shapeName ShapeType, opts *ShapeProps) error {
	return addShapeDefinition(s.ps, shapeName, opts)
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
	_, err := addChartDefinition(s.ps, chartType, nil, data, opts)
	return err
}

// AddMultiChart adds a multi-type (combo) chart to the slide. Ports the
// IChartMulti[] form of TS addChart.
func (s *Slide) AddMultiChart(multi []IChartMulti, opts *ChartOptions) error {
	_, err := addChartDefinition(s.ps, "", multi, nil, opts)
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
