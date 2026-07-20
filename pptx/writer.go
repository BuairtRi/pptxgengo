// writer.go ports the export half of src/pptxgen.ts (exportPresentation +
// write/writeFile). It assembles the OOXML package as a ZIP in the exact part
// order PptxGenJS/JSZip uses, so the individual parts are byte-identical to the
// JS library's output.
//
// JSZip defaults to STORE (no compression); DEFLATE is used only when the
// caller opts in via WriteProps.Compression — matching JSZip semantics.
package pptx

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
)

// Write renders the presentation to an in-memory .pptx byte slice. STORE
// (uncompressed) by default; pass a WriteProps with Compression=true for
// DEFLATE. Ports TS write.
func (p *Presentation) Write(props ...*WriteProps) ([]byte, error) {
	return p.build(compressionFrom(props))
}

// WriteTo renders the presentation and writes it to w, returning the byte count.
// Ports the Node stream path of TS write/stream.
func (p *Presentation) WriteTo(w io.Writer) (int64, error) {
	data, err := p.build(false)
	if err != nil {
		return 0, err
	}
	n, err := w.Write(data)
	return int64(n), err
}

// WriteFile renders the presentation and writes it to path (adding a .pptx
// extension if missing). Ports the Node fs path of TS writeFile.
func (p *Presentation) WriteFile(path string) error {
	if !strings.HasSuffix(strings.ToLower(path), ".pptx") {
		path += ".pptx"
	}
	data, err := p.build(false)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// WriteFileProps-style variant: WriteFileWith writes to path honoring compression.
func (p *Presentation) WriteFileWith(path string, compression bool) error {
	if !strings.HasSuffix(strings.ToLower(path), ".pptx") {
		path += ".pptx"
	}
	data, err := p.build(compression)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func compressionFrom(props []*WriteProps) bool {
	if len(props) > 0 && props[0] != nil && props[0].Compression != nil {
		return *props[0].Compression
	}
	return false
}

// build produces the complete .pptx package bytes.
func (p *Presentation) build(compression bool) ([]byte, error) {
	// Per-build clock + uuid, threaded explicitly (no package-global state) so
	// concurrent writes of different presentations don't race (REVIEW C1).
	bc := p.newBuildContext()

	// STEP 1: Read/encode all media before assembly (TS encodeSlideMediaRels).
	for i := range p.slides {
		resolveSlideMediaRels(&p.slides[i].SlideBaseProps)
	}
	for i := range p.slideLayouts {
		resolveSlideMediaRels(&p.slideLayouts[i].SlideBaseProps)
	}
	resolveSlideMediaRels(&p.masterSlide.SlideBaseProps)

	// STEP 2A: Add empty placeholder objects to slides that lack them.
	for i := range p.slides {
		addPlaceholdersToSlideLayouts(p.slides[i])
	}

	// Snapshot the slide model as a value slice for the makeXml* functions that
	// take []PresSlide (stable pointers live in p.slides; slices reallocate).
	slidesVal := make([]PresSlide, len(p.slides))
	for i := range p.slides {
		slidesVal[i] = *p.slides[i]
	}

	pres := &IPresentationProps{
		PresentationProps: PresentationProps{
			Author:      p.Author,
			Company:     p.Company,
			Layout:      p.layoutName,
			MasterSlide: p.masterSlide,
			PresLayout:  p.presLayout,
			Revision:    p.Revision,
			RtlMode:     p.RTL,
			Subject:     p.Subject,
			Theme:       p.Theme,
			Title:       p.Title,
		},
		Sections:      p.sections,
		SlideLayouts:  p.slideLayouts,
		Slides:        slidesVal,
		EmbeddedFonts: p.embeddedFonts,
	}

	buf := &bytes.Buffer{}
	zw := zip.NewWriter(buf)

	method := zip.Store
	if compression {
		method = zip.Deflate
	}
	addFile := func(name string, content []byte) error {
		w, err := zw.CreateHeader(&zip.FileHeader{Name: name, Method: method})
		if err != nil {
			return err
		}
		_, err = w.Write(content)
		return err
	}
	addStr := func(name, content string) error { return addFile(name, []byte(content)) }

	// STEP 2B: core package parts (TS section B).
	if err := addStr("[Content_Types].xml", makeXmlContTypes(slidesVal, p.slideLayouts, p.masterSlide, p.embeddedFonts)); err != nil {
		return nil, err
	}
	if err := addStr("_rels/.rels", makeXmlRootRels()); err != nil {
		return nil, err
	}
	if err := addStr("docProps/app.xml", makeXmlApp(slidesVal, p.Company)); err != nil {
		return nil, err
	}
	if err := addStr("docProps/core.xml", makeXmlCore(bc, p.Title, p.Subject, p.Author, p.Revision)); err != nil {
		return nil, err
	}
	if err := addStr("ppt/_rels/presentation.xml.rels", makeXmlPresentationRels(slidesVal, p.embeddedFonts)); err != nil {
		return nil, err
	}
	if err := addStr("ppt/theme/theme1.xml", makeXmlTheme(pres)); err != nil {
		return nil, err
	}
	if err := addStr("ppt/presentation.xml", makeXmlPresentation(pres, bc)); err != nil {
		return nil, err
	}
	if err := addStr("ppt/presProps.xml", makeXmlPresProps()); err != nil {
		return nil, err
	}
	if err := addStr("ppt/tableStyles.xml", makeXmlTableStyles()); err != nil {
		return nil, err
	}
	if err := addStr("ppt/viewProps.xml", makeXmlViewProps()); err != nil {
		return nil, err
	}

	// STEP 2C: layouts + slides + notes.
	for idx := range p.slideLayouts {
		layout := &p.slideLayouts[idx]
		if err := addStr(fmt.Sprintf("ppt/slideLayouts/slideLayout%d.xml", idx+1), makeXmlLayout(layout)); err != nil {
			return nil, err
		}
		if err := addStr(fmt.Sprintf("ppt/slideLayouts/_rels/slideLayout%d.xml.rels", idx+1), makeXmlSlideLayoutRel(idx+1, p.slideLayouts)); err != nil {
			return nil, err
		}
	}
	for idx := range slidesVal {
		if err := addStr(fmt.Sprintf("ppt/slides/slide%d.xml", idx+1), makeXmlSlide(&slidesVal[idx])); err != nil {
			return nil, err
		}
		if err := addStr(fmt.Sprintf("ppt/slides/_rels/slide%d.xml.rels", idx+1), makeXmlSlideRel(slidesVal, p.slideLayouts, idx+1)); err != nil {
			return nil, err
		}
		if err := addStr(fmt.Sprintf("ppt/notesSlides/notesSlide%d.xml", idx+1), makeXmlNotesSlide(&slidesVal[idx])); err != nil {
			return nil, err
		}
		if err := addStr(fmt.Sprintf("ppt/notesSlides/_rels/notesSlide%d.xml.rels", idx+1), makeXmlNotesSlideRel(idx+1)); err != nil {
			return nil, err
		}
	}
	if err := addStr("ppt/slideMasters/slideMaster1.xml", makeXmlMaster(p.masterSlide, p.slideLayouts)); err != nil {
		return nil, err
	}
	if err := addStr("ppt/slideMasters/_rels/slideMaster1.xml.rels", makeXmlMasterRel(p.masterSlide, p.slideLayouts)); err != nil {
		return nil, err
	}
	if err := addStr("ppt/notesMasters/notesMaster1.xml", makeXmlNotesMaster()); err != nil {
		return nil, err
	}
	if err := addStr("ppt/notesMasters/_rels/notesMaster1.xml.rels", makeXmlNotesMasterRel()); err != nil {
		return nil, err
	}

	// STEP 2D: chart + media rels (layouts, slides, master — TS order).
	writeChartMedia := func(base *SlideBaseProps) error {
		for j := range base.RelsChart {
			rel := &base.RelsChart[j]
			xlsx, err := createExcelWorksheet(rel, bc)
			if err != nil {
				return err
			}
			if err := addFile(fmt.Sprintf("ppt/embeddings/Microsoft_Excel_Worksheet%d.xlsx", rel.GlobalID), xlsx); err != nil {
				return err
			}
			if err := addStr("ppt/charts/_rels/"+rel.FileName+".rels", chartRelsXML(rel.GlobalID)); err != nil {
				return err
			}
			if err := addStr("ppt/charts/"+rel.FileName, makeXmlCharts(rel)); err != nil {
				return err
			}
		}
		for j := range base.RelsMedia {
			rel := &base.RelsMedia[j]
			if rel.Type == "online" || rel.Type == "hyperlink" {
				continue
			}
			raw, err := decodeMediaData(dataString(rel.Data))
			if err != nil {
				return err
			}
			target := strings.Replace(rel.Target, "..", "ppt", 1)
			if err := addFile(target, raw); err != nil {
				return err
			}
		}
		return nil
	}
	for idx := range p.slideLayouts {
		if err := writeChartMedia(&p.slideLayouts[idx].SlideBaseProps); err != nil {
			return nil, err
		}
	}
	for idx := range p.slides {
		if err := writeChartMedia(&p.slides[idx].SlideBaseProps); err != nil {
			return nil, err
		}
	}
	if err := writeChartMedia(&p.masterSlide.SlideBaseProps); err != nil {
		return nil, err
	}

	// STEP 2E: embedded fonts (net-new). One ppt/fonts/fontN.fntdata per variant.
	for _, v := range embeddedFontVariants(len(slidesVal), p.embeddedFonts) {
		data := fontVariantBytes(p.embeddedFonts, v.typeface, v.style)
		if err := addFile(fmt.Sprintf("ppt/fonts/font%d.fntdata", v.fileNum), data); err != nil {
			return nil, err
		}
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// chartRelsXML builds ppt/charts/_rels/chartN.xml.rels (the embedded-workbook
// package relationship). Mirrors the string built in gen-charts.ts.
func chartRelsXML(globalID int) string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
		`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
		`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/package" ` +
		`Target="../embeddings/Microsoft_Excel_Worksheet` + itoa(globalID) + `.xlsx"/>` +
		`</Relationships>`
}

// decodeMediaData normalizes a media rel's data string (adding a mime prefix
// when missing, exactly like createChartMediaRels) then base64-decodes the
// payload after the comma. Mirrors the TS media write path.
func decodeMediaData(data string) ([]byte, error) {
	hasComma := strings.Contains(data, ",")
	hasSemi := strings.Contains(data, ";")
	switch {
	case !hasComma && !hasSemi:
		data = "image/png;base64," + data
	case !hasComma:
		data = "image/png;base64," + data
	case !hasSemi:
		data = "image/png;" + data
	}
	parts := strings.Split(data, ",")
	b64 := parts[len(parts)-1]
	return base64.StdEncoding.DecodeString(b64)
}

// fontVariantBytes returns the raw font bytes for a typeface+style pair.
func fontVariantBytes(fonts []*EmbeddedFont, typeface string, style FontStyle) []byte {
	for _, f := range fonts {
		if f != nil && f.Typeface == typeface {
			return f.Variants[style]
		}
	}
	return nil
}
