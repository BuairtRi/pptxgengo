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
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Write renders the presentation to an in-memory .pptx byte slice. STORE
// (uncompressed) by default; pass a *WriteProps with Compression=true for
// DEFLATE. Ports TS write.
//
// Write is variadic only for backward-compatible call sites (Write() and
// Write(opts)); it is not a way to pass multiple option sets. Calling it with
// more than one *WriteProps is a caller error and returns an error rather
// than silently using the first argument and discarding the rest — prefer
// Write(opts) (zero or one argument).
func (p *Presentation) Write(props ...*WriteProps) ([]byte, error) {
	if len(props) > 1 {
		return nil, fmt.Errorf("pptx: Write takes at most one *WriteProps, got %d", len(props))
	}
	return p.build(compressionFrom(props))
}

// WriteTo renders the presentation (STORE, uncompressed) and writes it to w,
// returning the byte count. Ports the Node stream path of TS write/stream.
// Use WriteToOpts to opt into DEFLATE compression.
func (p *Presentation) WriteTo(w io.Writer) (int64, error) {
	return p.WriteToOpts(w, nil)
}

// WriteToOpts renders the presentation honoring opts.Compression and writes
// it to w, returning the byte count. This is the compression-aware sibling of
// WriteTo (REVIEW M9: compression was previously unreachable through the
// io.Writer path).
func (p *Presentation) WriteToOpts(w io.Writer, opts *WriteProps) (int64, error) {
	data, err := p.build(compressionFrom([]*WriteProps{opts}))
	if err != nil {
		return 0, err
	}
	n, err := w.Write(data)
	return int64(n), err
}

// WriteFile renders the presentation (STORE, uncompressed) and writes it
// atomically to path (adding a .pptx extension if missing). Ports the Node fs
// path of TS writeFile. Use WriteFileOpts to opt into DEFLATE compression.
func (p *Presentation) WriteFile(path string) error {
	return p.WriteFileOpts(path, nil)
}

// WriteFileOpts renders the presentation honoring opts.Compression and writes
// it atomically to path (adding a .pptx extension if missing): the data is
// written to a sibling temp file in the same directory, then renamed into
// place, so a failure (including a media error from build) never leaves a
// partially-written or corrupt file at path. This is the compression-aware
// sibling of WriteFile (REVIEW M9).
func (p *Presentation) WriteFileOpts(path string, opts *WriteProps) error {
	if !strings.HasSuffix(strings.ToLower(path), ".pptx") {
		path += ".pptx"
	}
	data, err := p.build(compressionFrom([]*WriteProps{opts}))
	if err != nil {
		return err
	}
	return atomicWriteFile(path, data)
}

// WriteFileWith is a deprecated thin wrapper over WriteFileOpts, kept for
// source compatibility with earlier ports; prefer WriteFileOpts.
func (p *Presentation) WriteFileWith(path string, compression bool) error {
	return p.WriteFileOpts(path, &WriteProps{WriteBaseProps: WriteBaseProps{Compression: &compression}})
}

// atomicWriteFile writes data to a temp file beside path (same directory,
// so the final os.Rename is same-filesystem and atomic on POSIX), then
// renames it into place. On any failure the temp file is removed and no
// partial/corrupt file is left at path (minor fix: WriteFile truncate-in-place
// could previously leave a corrupt file on a mid-write failure).
func atomicWriteFile(path string, data []byte) (err error) {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() {
		if err != nil {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err = tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	if err = os.Chmod(tmpPath, 0o644); err != nil {
		return err
	}
	if err = os.Rename(tmpPath, path); err != nil {
		return err
	}
	return nil
}

// compressionFrom resolves the effective compression flag from an optional
// WriteProps slice, treating a nil slice, a nil *WriteProps, or a nil
// Compression field alike as "not requested" (STORE).
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
	//
	// REVIEW C3: resolveSlideMediaRels reports (but does not itself fail on)
	// read/fetch errors — it substitutes IMG_BROKEN into the rel and keeps
	// going, matching the *synchronous* effect of the TS fallback. But real
	// PptxGenJS's async pipeline awaits Promise.all(arrMediaPromises) BEFORE
	// building the zip (pptxgen.ts:494), and each encodeSlideMediaRels promise
	// *rejects* on failure (gen-media.ts:70,93,125) — so the overall write
	// rejects with the first media error rather than silently shipping
	// IMG_BROKEN. We aggregate every error (not just the first) since Go has
	// no short-circuiting Promise.all equivalent worth emulating here, and
	// errors.Join reports all of them instead of hiding N-1.
	var mediaErrs []error
	for i := range p.slides {
		mediaErrs = append(mediaErrs, resolveSlideMediaRels(&p.slides[i].SlideBaseProps)...)
	}
	for i := range p.slideLayouts {
		mediaErrs = append(mediaErrs, resolveSlideMediaRels(&p.slideLayouts[i].SlideBaseProps)...)
	}
	mediaErrs = append(mediaErrs, resolveSlideMediaRels(&p.masterSlide.SlideBaseProps)...)
	if len(mediaErrs) > 0 {
		// Fail the whole build before any zip assembly, so Write/WriteTo/
		// WriteFile all observe the error and WriteFile never creates a
		// partial output file.
		return nil, errors.Join(mediaErrs...)
	}

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
	if err := addStr("ppt/theme/theme1.xml", customizeThemeColors(makeXmlTheme(pres), pres.Theme.ColorScheme)); err != nil {
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
