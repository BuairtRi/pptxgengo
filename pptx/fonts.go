/*
 * fonts.go — embedded TrueType font support (net-new feature, not in PptxGenJS).
 *
 * PPTX font embedding per ECMA-376:
 *   - font binary data is stored as ppt/fonts/fontN.fntdata parts
 *   - [Content_Types].xml declares <Default Extension="fntdata" ContentType="application/x-fontdata"/>
 *   - presentation.xml carries embedTrueTypeFonts="1" and a <p:embeddedFontLst> with one
 *     <p:embeddedFont> per typeface referencing the parts by relationship id
 *   - ppt/_rels/presentation.xml.rels adds relationships of type .../relationships/font
 */

package pptx

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
)

// FontStyle identifies which style slot of a typeface a font file fills.
type FontStyle string

const (
	FontRegular    FontStyle = "regular"
	FontBold       FontStyle = "bold"
	FontItalic     FontStyle = "italic"
	FontBoldItalic FontStyle = "boldItalic"
)

// FontEmbedProps describes one typeface to embed. Typeface is required and must
// match the font face name used in text runs (e.g. "Lato"). At least one style
// variant must be provided, either as raw TTF/OTF bytes or as a file path
// (bytes win when both are set for the same style).
type FontEmbedProps struct {
	Typeface string

	Regular    []byte
	Bold       []byte
	Italic     []byte
	BoldItalic []byte

	RegularPath    string
	BoldPath       string
	ItalicPath     string
	BoldItalicPath string
}

// EmbeddedFont is the resolved, validated form stored on the Presentation.
type EmbeddedFont struct {
	Typeface string
	// Variants maps style → font bytes; only present styles have entries.
	Variants map[FontStyle][]byte
}

// sfnt magic numbers accepted for embedding.
var errNotAFont = errors.New("pptx: data is not a TTF/OTF font (bad sfnt magic)")

// validateFontData performs a cheap sanity check that data looks like an
// sfnt-housed font (TrueType 0x00010000 or 'true', OpenType/CFF 'OTTO').
// TTC collections ('ttcf') are rejected: a .fntdata part holds a single face.
func validateFontData(data []byte) error {
	if len(data) < 12 {
		return errNotAFont
	}
	magic := binary.BigEndian.Uint32(data[:4])
	switch magic {
	case 0x00010000, 0x74727565 /* 'true' */, 0x4F54544F /* 'OTTO' */ :
		return nil
	case 0x74746366 /* 'ttcf' */ :
		return errors.New("pptx: TrueType collections (.ttc) cannot be embedded; extract a single face")
	default:
		return errNotAFont
	}
}

// resolveVariant returns the bytes for one style, reading from path when no
// inline bytes were given. Returns nil bytes when the style is absent.
func resolveVariant(data []byte, path string) ([]byte, error) {
	if len(data) > 0 {
		return data, nil
	}
	if path == "" {
		return nil, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("pptx: reading font file: %w", err)
	}
	return b, nil
}

// newEmbeddedFont validates props and produces the resolved EmbeddedFont.
func newEmbeddedFont(props FontEmbedProps) (*EmbeddedFont, error) {
	if props.Typeface == "" {
		return nil, errors.New("pptx: FontEmbedProps.Typeface is required")
	}
	type slot struct {
		style FontStyle
		data  []byte
		path  string
	}
	slots := []slot{
		{FontRegular, props.Regular, props.RegularPath},
		{FontBold, props.Bold, props.BoldPath},
		{FontItalic, props.Italic, props.ItalicPath},
		{FontBoldItalic, props.BoldItalic, props.BoldItalicPath},
	}
	ef := &EmbeddedFont{Typeface: props.Typeface, Variants: map[FontStyle][]byte{}}
	for _, s := range slots {
		b, err := resolveVariant(s.data, s.path)
		if err != nil {
			return nil, fmt.Errorf("%s (%s)", err, s.style)
		}
		if b == nil {
			continue
		}
		if err := validateFontData(b); err != nil {
			return nil, fmt.Errorf("%w (typeface %q, style %s)", err, props.Typeface, s.style)
		}
		ef.Variants[s.style] = b
	}
	if len(ef.Variants) == 0 {
		return nil, fmt.Errorf("pptx: no font data provided for typeface %q (need at least one style)", props.Typeface)
	}
	return ef, nil
}

// fontStyleOrder is the emission order of variant elements inside
// <p:embeddedFont> — must match the ECMA-376 sequence: regular, bold, italic, boldItalic.
var fontStyleOrder = []FontStyle{FontRegular, FontBold, FontItalic, FontBoldItalic}

// embeddedFontXMLTags maps a style to its element name in presentation.xml.
var embeddedFontXMLTags = map[FontStyle]string{
	FontRegular:    "p:regular",
	FontBold:       "p:bold",
	FontItalic:     "p:italic",
	FontBoldItalic: "p:boldItalic",
}
