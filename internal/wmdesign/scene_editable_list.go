package wmdesign

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

const EditableListContract = "pptxgengo.editable-list.v1"

type editableListSource struct {
	Type  string   `json:"type"`
	X     float64  `json:"x"`
	Y     float64  `json:"y"`
	W     float64  `json:"w"`
	H     float64  `json:"h"`
	On    string   `json:"on,omitempty"`
	Size  string   `json:"size,omitempty"`
	Items []string `json:"items"`
}

// One native text box owns all bullet paragraphs. Measured wrap lines are
// observations; they are not emitted as hard breaks that obstruct native reflow.
func (r *renderer) planEditableList(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, error) {
	var n editableListSource
	if e := sceneDecode(raw, &n); e != nil {
		return nil, e
	}
	b := Rect{n.X, n.Y, n.W, n.H}
	if b.W <= 0 || b.H <= 0 || math.IsNaN(b.X+b.Y+b.W+b.H) || math.IsInf(b.X+b.Y+b.W+b.H, 0) || len(n.Items) == 0 || len(n.Items) > 100 {
		return nil, fmt.Errorf("scene.editable_list_invalid_allocation_or_items")
	}
	token := n.Size
	if token == "" {
		token = "body"
	}
	if token != "body" && token != "small" {
		return nil, fmt.Errorf("scene.editable_list_requires_body_or_small")
	}
	surface := n.On
	if surface == "" {
		surface = ctx.Surface
	}
	if surface == "" {
		surface = "light"
	}
	st, e := r.sceneStyle(token)
	if e != nil {
		return nil, e
	}
	suffix := ""
	indent, marker, gap := 15., 4., 6.
	if token == "small" {
		suffix = "-small"
		indent, marker, gap = 12, 3, 3
	}
	if r.source.Tokens.Density != nil {
		indent = r.listMetric("list-indent"+suffix, indent)
		marker = r.listMetric("list-marker"+suffix, marker)
		gap = r.listMetric("list-gap"+suffix, gap)
	}
	if b.W <= indent || indent <= 0 || marker <= 0 || gap < 0 {
		return nil, fmt.Errorf("scene.editable_list_invalid_metrics")
	}
	keys, e := primitiveArrayKeys(ctx, "/items", len(n.Items))
	if e != nil {
		return nil, e
	}
	var combined TextRecord
	var layout TextLayout
	rich := RichTextLayout{Contract: RichTextContract, NativeQualified: false}
	offset, lastOccupied := 0., 0.
	for i, copy := range n.Items {
		if strings.TrimSpace(copy) == "" || strings.ContainsAny(copy, "\r\n\t") || strings.Contains(copy, "[[") || strings.Contains(copy, "]]") || strings.Contains(copy, "[^") {
			return nil, fmt.Errorf("scene.editable_list_requires_plain_single_paragraph_items")
		}
		part, e := r.primitiveMeasureRuns(id, []primitiveRun{{Text: copy}}, st, Rect{X: b.X + indent, Y: b.Y, W: b.W - indent}, surface, "primary", "left")
		if e != nil {
			return nil, e
		}
		if part.Layout.Displayed != copy || part.Rich == nil || len(part.Rich.Paragraphs) != 1 || len(part.Rich.Paragraphs[0].Runs) != 1 {
			return nil, fmt.Errorf("scene.editable_list_nonplain_copy")
		}
		if i == 0 {
			combined = part
			layout = part.Layout
			layout.Lines = nil
		}
		paragraph := part.Rich.Paragraphs[0]
		paragraph.Key = keys[i]
		paragraph.Bullet = true
		paragraph.BulletIndentPt = indent
		paragraph.BulletMarkerPt = marker
		paragraph.LineBreaks = nil
		paragraph.FirstLine = len(layout.Lines)
		paragraph.LineCount = len(part.Layout.Lines)
		// Retain the same after-spacing on the last paragraph: PowerPoint
		// inherits it when Enter adds an item. It has no painted effect until
		// there is a following paragraph, so it is excluded from final fit.
		paragraph.ParagraphGapAfter = gap + math.Max(0, part.Layout.OccupiedTop+part.Layout.EstimatedOccupiedHeight-part.Layout.AllocationHeight)
		for _, line := range part.Layout.Lines {
			line.Baseline += offset
			layout.Lines = append(layout.Lines, line)
		}
		lastOccupied = offset + part.Layout.OccupiedTop + part.Layout.EstimatedOccupiedHeight
		offset += part.Layout.AllocationHeight
		if i < len(n.Items)-1 {
			offset += paragraph.ParagraphGapAfter
		}
		rich.Paragraphs = append(rich.Paragraphs, paragraph)
	}
	if math.Max(offset, lastOccupied) > b.H+.02 {
		return nil, fmt.Errorf("scene.editable_list_initial_overflow: %s requires %.3f pt; allocation is %.3f pt", id, math.Max(offset, lastOccupied), b.H)
	}
	layout.Original = strings.Join(n.Items, "\n")
	layout.Displayed = layout.Original
	layout.AllocationHeight = offset
	layout.EstimatedOccupiedHeight = lastOccupied - layout.OccupiedTop
	layout.VerticalPolicy = "Native unordered-list paragraphs; initial measured fit only, native reflow/resize parity unqualified"
	combined.ID = id
	combined.Rect = b
	combined.Layout = layout
	combined.Rich = &rich
	combined.NativeParagraphContract = EditableListContract
	return &scenePlan{Definition: "scene.editable-list", ID: id, Bounds: b, Items: []sceneItem{{Text: &combined}}, Warnings: []string{"Editable list uses one native text box and styled bullet paragraphs. Square bullets require the Wingdings desktop font; it is not redistributed in the bundle. Font size is fixed; added copy may require resizing the one box. Native reflow and visual qualification remain pending; bullet source adoption is manual."}}, nil
}

func editableListParagraphXML(tr TextRecord) []byte {
	var out []byte
	for _, paragraph := range tr.Rich.Paragraphs {
		one := tr
		one.Rich = &RichTextLayout{Contract: RichTextContract, Paragraphs: []RichParagraphLayout{paragraph}}
		properties := sceneTableParagraphProperties(one, paragraph.ParagraphGapAfter, true, paragraph)
		// IBM Plex has no U+25A0 glyph: leaving it as the bullet font makes
		// PowerPoint substitute a tiny fallback square. Wingdings' small
		// solid square has 592 font units of ink in a 2048-unit em. Size the
		// glyph to the source marker's actual ink dimensions, not its em.
		// Wingdings is an Office desktop prerequisite, never redistributed.
		fontSize := int(math.Round(paragraph.BulletMarkerPt * 2048 / 592 * 100))
		oldSize := int(math.Round(paragraph.BulletMarkerPt * 100))
		properties = bytes.Replace(properties, []byte(`<a:buSzPts val="`+strconv.Itoa(oldSize)+`"/>`), []byte(`<a:buSzPts val="`+strconv.Itoa(fontSize)+`"/>`), 1)
		properties = bytes.Replace(properties, []byte(`<a:buFont typeface="`+sceneTableXMLEscape(tr.Layout.Font.Typeface)+`"/><a:buChar char="■"/>`), []byte(`<a:buFont typeface="Wingdings" charset="2"/><a:buChar char=""/>`), 1)
		// Keep bullet ink explicit rather than inheriting a presentation theme.
		properties = bytes.Replace(properties, []byte("<a:buSzPts"), []byte(`<a:buClr><a:srgbClr val="`+sceneTableXMLEscape(tr.Color)+`"/></a:buClr><a:buSzPts`), 1)
		out = append(out, richParagraphXML(one, properties)...)
	}
	return out
}
