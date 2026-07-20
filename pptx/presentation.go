// presentation.go ports src/pptxgen.ts: the public Presentation type (PptxGenJS
// class) — the top-level entry point that owns the slide/layout/section model
// and the metadata that feeds docProps. Writer logic lives in writer.go; the
// per-slide user API lives in slide.go.
package pptx

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Version is the PptxGenJS library version this port tracks.
const Version = "4.0.1"

// Presentation is the top-level object that builds a .pptx. Create one with
// New, add slides/masters/sections/fonts, then Write / WriteTo / WriteFile.
//
// Ports the PptxGenJS class. Metadata that TS exposed via getters/setters are
// plain exported fields here (Author, Company, Revision, Subject, Title, RTL,
// Theme); the layout is a validated setter (SetLayout) plus a Layout getter.
type Presentation struct {
	// Core document metadata (docProps/core.xml + docProps/app.xml).
	Author   string
	Company  string
	Revision string // whole number only, per OOXML
	Subject  string
	Title    string
	RTL      bool
	Theme    ThemeProps

	layoutName string
	presLayout PresLayout
	layouts    map[string]PresLayout

	masterSlide   *PresSlide
	slides        []*PresSlide
	slideLayouts  []SlideLayout
	sections      []SectionProps
	embeddedFonts []*EmbeddedFont

	// nowFunc, when non-nil, supplies the timestamp threaded through each build
	// (docProps/core.xml and the embedded workbook's core.xml), making output
	// deterministic in tests. Nil = wall-clock time. It is read into a per-build
	// buildContext (never a package global), so concurrent writes of different
	// presentations cannot clobber each other's clock (REVIEW C1).
	nowFunc func() time.Time

	// uuidFunc, when non-nil, supplies the section GUID emitted in
	// ppt/presentation.xml (mirrors PORTING.md's promised uuid hook). Nil =
	// getUuid (random). Tests may pin it to make section GUIDs deterministic.
	uuidFunc func(format string) string

	// chartCtr assigns each chart its 1-based part index (chart1.xml, ...).
	// Per-Presentation and concurrency-safe — see the chartCounter doc comment.
	chartCtr chartCounter
}

// buildContext carries the per-build injectable dependencies — a clock and a
// UUID generator — so the XML/worksheet generators never read package-global
// mutable state. It is created once per build() from the Presentation's fields.
//
// REVIEW C1: PptxGenJS's Go port previously swapped package-level xmlNowFunc /
// excelNowFunc for the duration of build(); two goroutines writing two
// presentations raced (and a deferred restore could reinstate the other
// writer's clock). Threading the clock explicitly removes that shared state.
type buildContext struct {
	now  func() time.Time
	uuid func(format string) string
}

// newBuildContext builds the context from the Presentation's optional hooks,
// falling back to wall-clock time and the real getUuid.
func (p *Presentation) newBuildContext() *buildContext {
	now := p.nowFunc
	if now == nil {
		now = time.Now
	}
	uuid := p.uuidFunc
	if uuid == nil {
		uuid = getUuid
	}
	return &buildContext{now: now, uuid: uuid}
}

// chartCounter is a per-Presentation, concurrency-safe counter that assigns each
// chart its 1-based part index (chart1.xml, chart2.xml, ...).
//
// DEVIATION (REVIEW C2/M3): PptxGenJS uses a module-level `let _chartCounter`
// that is never reset, so a second presentation built in the same process
// numbers its charts chart2+/Microsoft_Excel_Worksheet2+ — output that leaks
// across presentations and is not reproducible without a manual reset. The Go
// port makes the counter per-Presentation: every Presentation numbers its charts
// from 1, deterministic regardless of process history, and safe under concurrent
// AddChart both across presentations and on the same presentation.
type chartCounter struct {
	mu sync.Mutex
	n  int
}

// next returns the next 1-based chart index.
func (c *chartCounter) next() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n++
	return c.n
}

// New creates a Presentation with the same defaults as the PptxGenJS
// constructor: 16x9 layout, "PptxGenJS" author/company/title/subject strings,
// revision "1", one DEFAULT slide layout, and an empty slide master.
func New() *Presentation {
	p := &Presentation{
		Author:   "PptxGenJS",
		Company:  "PptxGenJS",
		Revision: "1",
		Subject:  "PptxGenJS Presentation",
		Title:    "PptxGenJS Presentation",
	}

	p.layouts = map[string]PresLayout{
		"LAYOUT_4x3":   {Name: "screen4x3", Width: 9144000, Height: 6858000, SizeW: 9144000, SizeH: 6858000},
		"LAYOUT_16x9":  {Name: "screen16x9", Width: 9144000, Height: 5143500, SizeW: 9144000, SizeH: 5143500},
		"LAYOUT_16x10": {Name: "screen16x10", Width: 9144000, Height: 5715000, SizeW: 9144000, SizeH: 5715000},
		"LAYOUT_WIDE":  {Name: "custom", Width: 12192000, Height: 6858000, SizeW: 12192000, SizeH: 6858000},
	}

	def := p.layouts[DEF_PRES_LAYOUT]
	p.layoutName = DEF_PRES_LAYOUT
	p.presLayout = PresLayout{
		Name:   def.Name,
		Width:  def.Width,
		Height: def.Height,
		SizeW:  def.Width,
		SizeH:  def.Height,
	}

	p.slideLayouts = []SlideLayout{
		{SlideBaseProps: SlideBaseProps{
			Margin:     append(Margin{}, DEF_SLIDE_MARGIN_IN[:]...),
			Name:       DEF_PRES_LAYOUT_NAME,
			PresLayout: p.presLayout,
			SlideNum:   1000,
		}},
	}
	p.slides = []*PresSlide{}
	p.sections = []SectionProps{}
	p.masterSlide = &PresSlide{SlideBaseProps: SlideBaseProps{PresLayout: p.presLayout}}

	return p
}

// Layout returns the current layout key (e.g. "LAYOUT_16x9").
func (p *Presentation) Layout() string { return p.layoutName }

// SetLayout selects a standard or custom layout by name (as registered by
// DefineLayout). Ports the TS `layout` setter, which throws UNKNOWN-LAYOUT.
func (p *Presentation) SetLayout(name string) error {
	newLayout, ok := p.layouts[name]
	if !ok {
		return errors.New("UNKNOWN-LAYOUT")
	}
	p.layoutName = name
	p.presLayout = PresLayout{
		Name:   newLayout.Name,
		Width:  newLayout.Width,
		Height: newLayout.Height,
		SizeW:  newLayout.Width,
		SizeH:  newLayout.Height,
	}
	return nil
}

// PresLayout returns the resolved presentation layout (size in EMU).
func (p *Presentation) PresLayout() PresLayout { return p.presLayout }

// Slides returns the presentation's slides (pointers into internal storage).
func (p *Presentation) Slides() []*PresSlide { return p.slides }

// Sections returns the presentation's sections.
func (p *Presentation) Sections() []SectionProps { return p.sections }

// SlideLayouts returns the presentation's slide layouts.
func (p *Presentation) SlideLayouts() []SlideLayout { return p.slideLayouts }

// EmbeddedFonts returns the registered embedded fonts.
func (p *Presentation) EmbeddedFonts() []*EmbeddedFont { return p.embeddedFonts }

// DefineLayout registers a custom layout sized in inches. Ports TS defineLayout
// (inches → EMU via Math.round). The new layout can then be selected with
// SetLayout(name).
func (p *Presentation) DefineLayout(name string, width, height float64) {
	if name == "" || width == 0 || height == 0 {
		return
	}
	w := int(jsRound(width * EMU))
	h := int(jsRound(height * EMU))
	p.layouts[name] = PresLayout{Name: name, Width: w, Height: h, SizeW: w, SizeH: h}
}

// AddSection adds a named user section. Ports TS addSection.
func (p *Presentation) AddSection(props SectionProps) {
	if props.Title == "" {
		return
	}
	newSection := SectionProps{Type: "user", Slides: []*PresSlide{}, Title: props.Title}
	if props.Order > 0 && props.Order <= len(p.sections) {
		// insert at index props.Order (mirrors Array.splice(order, 0, …))
		p.sections = append(p.sections, SectionProps{})
		copy(p.sections[props.Order+1:], p.sections[props.Order:])
		p.sections[props.Order] = newSection
	} else {
		p.sections = append(p.sections, newSection)
	}
}

// AddSlide adds a new slide (optionally with a master layout and/or section).
// Returns the Slide wrapper for adding content. Ports TS addSlide.
func (p *Presentation) AddSlide(props ...*AddSlideProps) *Slide {
	var opts *AddSlideProps
	if len(props) > 0 {
		opts = props[0]
	}
	return p.addSlideInternal(opts)
}

func (p *Presentation) addSlideInternal(options *AddSlideProps) *Slide {
	masterName := ""
	if options != nil {
		masterName = options.MasterName
	}

	// Default (fresh) layout for slides without a named master.
	slideLayout := SlideLayout{SlideBaseProps: SlideBaseProps{
		Name:       p.presLayout.Name,
		PresLayout: p.presLayout,
		SlideNum:   len(p.slides) + 1,
	}}
	if masterName != "" {
		for i := range p.slideLayouts {
			if p.slideLayouts[i].Name == masterName {
				slideLayout = p.slideLayouts[i]
				break
			}
		}
	}

	newSlide := &PresSlide{
		SlideBaseProps: SlideBaseProps{
			Name:       fmt.Sprintf("Slide %d", len(p.slides)+1),
			PresLayout: p.presLayout,
			SlideNum:   len(p.slides) + 1,
		},
		RID:         len(p.slides) + 2,
		SlideID:     len(p.slides) + 256,
		SlideLayout: slideLayout,
	}
	// Slide numbers must live in master/layout/slide; inherit from the layout.
	if slideLayout.SlideNumberProps != nil {
		newSlide.SlideNumberProps = slideLayout.SlideNumberProps
	}

	// A: add slide to pres
	p.slides = append(p.slides, newSlide)

	// B: section placement
	sectionTitle := ""
	if options != nil {
		sectionTitle = options.SectionTitle
	}
	if sectionTitle != "" {
		found := false
		for i := range p.sections {
			if p.sections[i].Title == sectionTitle {
				p.sections[i].Slides = append(p.sections[i].Slides, newSlide)
				found = true
				break
			}
		}
		_ = found // TS only warns when not found
	} else if len(p.sections) > 0 {
		last := &p.sections[len(p.sections)-1]
		if last.Type == "default" {
			last.Slides = append(last.Slides, newSlide)
		} else {
			defCount := 0
			for i := range p.sections {
				if p.sections[i].Type == "default" {
					defCount++
				}
			}
			p.sections = append(p.sections, SectionProps{
				Title:  fmt.Sprintf("Default-%d", defCount+1),
				Type:   "default",
				Slides: []*PresSlide{newSlide},
			})
		}
	}

	return &Slide{ps: newSlide, pres: p}
}

// addNewSlide is the auto-paging callback handed to addTableDefinition. It
// keeps continued slides in the same section as the slide being paged. Ports
// the TS `addNewSlide` closure.
func (p *Presentation) addNewSlide(options *AddSlideProps) *PresSlide {
	if options == nil {
		options = &AddSlideProps{}
	}
	sectAlreadyInUse := false
	if len(p.sections) > 0 && len(p.slides) > 0 {
		lastSlideNum := p.slides[len(p.slides)-1].SlideNum
		lastSect := p.sections[len(p.sections)-1]
		for _, s := range lastSect.Slides {
			if s != nil && s.SlideNum == lastSlideNum {
				sectAlreadyInUse = true
				break
			}
		}
	}
	if sectAlreadyInUse {
		options.SectionTitle = p.sections[len(p.sections)-1].Title
	} else {
		options.SectionTitle = ""
	}
	return p.addSlideInternal(options).ps
}

// getSlide returns the slide with the given slide number (auto-paging callback).
func (p *Presentation) getSlide(slideNum int) *PresSlide {
	for _, s := range p.slides {
		if s.SlideNum == slideNum {
			return s
		}
	}
	return nil
}

// setSlideNumber records slide-number config on the master and DEFAULT layout,
// so slide numbers appear in all three of master/layout/slide. Ports the TS
// `setSlideNumber` closure.
func (p *Presentation) setSlideNumber(snp *SlideNumberProps) {
	p.masterSlide.SlideNumberProps = snp
	for i := range p.slideLayouts {
		if p.slideLayouts[i].Name == DEF_PRES_LAYOUT_NAME {
			p.slideLayouts[i].SlideNumberProps = snp
		}
	}
}

// DefineSlideMaster registers a new slide master/layout. Ports TS
// defineSlideMaster. Requires a Title.
func (p *Presentation) DefineSlideMaster(props *SlideMasterProps) error {
	if props == nil || props.Title == "" {
		return errors.New("defineSlideMaster() object argument requires a `title` value. (https://gitbrent.github.io/PptxGenJS/docs/masters.html)")
	}

	margin := props.Margin
	if margin == nil {
		margin = append(Margin{}, DEF_SLIDE_MARGIN_IN[:]...)
	}

	newLayout := SlideLayout{SlideBaseProps: SlideBaseProps{
		Margin:           margin,
		Name:             props.Title,
		PresLayout:       p.presLayout,
		SlideNum:         1000 + len(p.slideLayouts) + 1,
		SlideNumberProps: props.SlideNumber,
		Background:       props.Background,
		Bkgd:             props.Bkgd,
	}}

	// STEP 1: build the master/layout objects.
	if err := createSlideMaster(props, &p.chartCtr, &newLayout); err != nil {
		return err
	}

	// STEP 2: register it.
	p.slideLayouts = append(p.slideLayouts, newLayout)
	added := &p.slideLayouts[len(p.slideLayouts)-1]

	// STEP 3: background (image data/path captured before write).
	if props.Background != nil || props.Bkgd != nil {
		addBackgroundDefinition(props.Background, &added.SlideBaseProps)
	}

	// STEP 4: propagate slide number to the master slide (if any).
	if added.SlideNumberProps != nil && p.masterSlide.SlideNumberProps == nil {
		p.masterSlide.SlideNumberProps = added.SlideNumberProps
	}

	return nil
}

// EmbedFont registers a TrueType/OpenType font for embedding (net-new vs
// PptxGenJS). Duplicate typefaces are rejected. Ports the fonts.go design.
func (p *Presentation) EmbedFont(props FontEmbedProps) error {
	for _, ef := range p.embeddedFonts {
		if ef != nil && ef.Typeface == props.Typeface {
			return fmt.Errorf("pptx: font typeface %q already embedded", props.Typeface)
		}
	}
	ef, err := newEmbeddedFont(props)
	if err != nil {
		return err
	}
	p.embeddedFonts = append(p.embeddedFonts, ef)
	return nil
}
