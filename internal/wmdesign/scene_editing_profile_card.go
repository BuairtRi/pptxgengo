package wmdesign

import (
	"encoding/json"
	"math"
	"strings"

	"github.com/buairtri/pptxgengo/pptx"
)

func nativeProfileCardResult(p *scenePlan, raw json.RawMessage) (*scenePlan, string) {
	if p == nil {
		return nil, "card scene plan is unavailable"
	}
	var source sceneCardSource
	if json.Unmarshal(raw, &source) != nil {
		return nil, "card source could not be read"
	}
	if source.State != "" && source.State != "deemph" {
		return nil, "featured and placeholder states add independent decoration or content"
	}
	if source.ContentBottom != 0 {
		return nil, "card reserves separate footer space"
	}
	if len(source.Body) == 0 {
		return nil, "card needs at least one paragraph"
	}
	for _, blockRaw := range source.Body {
		var block map[string]json.RawMessage
		if json.Unmarshal(blockRaw, &block) != nil || len(block) != 1 || block["p"] == nil {
			return nil, "card body includes a non-paragraph component"
		}
		var text string
		if json.Unmarshal(block["p"], &text) != nil {
			return nil, "card paragraph text is not a string"
		}
		if strings.Contains(text, "[[") || strings.Contains(text, "]]") {
			return nil, "inline emphasis artwork cannot share the filled editable card shape safely"
		}
	}
	decorativeFields := source.Number != "" || source.InlineNumber != "" || source.BandNumber != "" || source.CornerNumber != "" || source.NumTile != "" || source.Band != nil || source.Edge != nil || source.Icon != nil || source.Media != nil || source.Metric != nil || source.MetricGroup != nil || source.Quote != nil || source.Person != nil || source.Bio != nil || source.Case != nil || source.Fee != nil
	if decorativeFields {
		return nil, "card includes an independently positioned decorative or content object"
	}
	if p != nil && len(p.Items) != 1+len(source.Body)+boolInt(source.Title != "")+boolInt(source.Label != "") {
		return nil, "card plan contains additional native artwork objects"
	}
	if candidate := nativeProfileCardMulti(p, raw); candidate != nil {
		return candidate, ""
	}
	return nil, "card text paragraphs do not share compatible measured geometry or supported formatting"
}

// nativeProfileCardMulti folds supported card text fields into one native text
// shape. Each paragraph retains its measured run style and baseline; source
// geometry, card fill and any supported outline remain fixed.
func nativeProfileCardMulti(p *scenePlan, raw json.RawMessage) *scenePlan {
	var source sceneCardSource
	if p == nil || json.Unmarshal(raw, &source) != nil || len(source.Body) == 0 || source.State != "" && source.State != "deemph" || source.ContentBottom != 0 {
		return nil
	}
	if source.Number != "" || source.InlineNumber != "" || source.BandNumber != "" || source.CornerNumber != "" || source.NumTile != "" || source.Band != nil || source.Edge != nil || source.Icon != nil || source.Media != nil || source.Metric != nil || source.MetricGroup != nil || source.Quote != nil || source.Person != nil || source.Bio != nil || source.Case != nil || source.Fee != nil {
		return nil
	}
	wantItems := 1 + len(source.Body) + boolInt(source.Title != "") + boolInt(source.Label != "")
	if len(p.Items) != wantItems {
		return nil
	}
	shape := p.Items[0].Shape
	if shape == nil || shape.Type != pptx.ShapeTypeRect || shape.Props.Fill == nil || shape.Props.Line == nil || shape.Record.Rotation != 0 {
		return nil
	}
	if shape.Props.Line.Type != "none" && (source.Surface != "outline" && source.State != "deemph" || shape.Props.Line.Width <= 0 || shape.Props.Line.Color == "") {
		return nil
	}
	paragraphs := make([]RichParagraphLayout, 0, len(source.Body)+boolInt(source.Title != "")+boolInt(source.Label != ""))
	records := make([]TextRecord, 0, cap(paragraphs))
	textItems := map[string]*TextRecord{}
	bodyRecords := make([]TextRecord, 0, len(source.Body))
	for _, item := range p.Items[1:] {
		if item.Text == nil {
			return nil
		}
		textItems[item.Text.ID] = item.Text
		if strings.HasPrefix(item.Text.ID, p.ID+".body.") {
			bodyRecords = append(bodyRecords, *item.Text)
		}
	}
	if source.Label != "" {
		label := textItems[p.ID+".label"]
		if label == nil {
			return nil
		}
		paragraph, ok := nativeCardSourceParagraph(*label, source.Label, "label")
		if !ok || label.Align != "left" || label.Rotation != 0 || label.NativeShape != nil {
			return nil
		}
		paragraphs, records = append(paragraphs, paragraph), append(records, *label)
		delete(textItems, label.ID)
	}
	if source.Title != "" {
		title := textItems[p.ID+".title"]
		if title == nil {
			return nil
		}
		paragraph, ok := nativeCardSourceParagraph(*title, source.Title, "title")
		if !ok || title.Align != "left" || title.Rotation != 0 || title.NativeShape != nil {
			return nil
		}
		paragraphs, records = append(paragraphs, paragraph), append(records, *title)
		delete(textItems, title.ID)
	}
	for i, rawBlock := range source.Body {
		var block map[string]json.RawMessage
		if json.Unmarshal(rawBlock, &block) != nil || len(block) != 1 || block["p"] == nil {
			return nil
		}
		if i >= len(bodyRecords) {
			return nil
		}
		body := &bodyRecords[i]
		var text string
		if body == nil || json.Unmarshal(block["p"], &text) != nil || body.Align != "left" || body.Rotation != 0 || body.NativeShape != nil {
			return nil
		}
		if len(records) > 0 && (math.Abs(records[0].Rect.X-body.Rect.X) > .02 || math.Abs(records[0].Rect.W-body.Rect.W) > .02) {
			return nil
		}
		paragraph, ok := nativeCardSourceParagraph(*body, text, "body-"+strings.TrimPrefix(body.ID, p.ID+".body."))
		if !ok {
			return nil
		}
		paragraphs, records = append(paragraphs, paragraph), append(records, *body)
		delete(textItems, body.ID)
	}
	if len(textItems) != 0 || len(records) == 0 {
		return nil
	}
	for i := 0; i+1 < len(records); i++ {
		gap := records[i+1].Rect.Y - records[i].Rect.Y - records[i].Layout.AllocationHeight
		if gap < -.02 {
			return nil
		}
		paragraphs[i].ParagraphGapAfter = math.Max(0, gap)
	}
	tr := records[0]
	tr.ID = p.ID
	outer := shape.Record.Rect
	pad := records[0].Rect.X - outer.X
	bottom := outer.Y + outer.H - pad
	if source.ContentBottom > 0 {
		bottom = math.Min(bottom, source.ContentBottom)
	}
	tr.Rect.H = bottom - records[0].Rect.Y
	if tr.Rect.H <= 0 || !inside(tr.Rect, outer) {
		return nil
	}
	tr.Layout.Lines = nil
	tr.Layout.Original = ""
	tr.Layout.Displayed = ""
	for i, record := range records {
		offset := record.Rect.Y - records[0].Rect.Y
		paragraphs[i].FirstLine = len(tr.Layout.Lines)
		for _, line := range record.Layout.Lines {
			line.Baseline += offset
			tr.Layout.Lines = append(tr.Layout.Lines, line)
		}
		if i > 0 {
			tr.Layout.Original += "\n"
			tr.Layout.Displayed += "\n"
		}
		tr.Layout.Original += record.Layout.Original
		tr.Layout.Displayed += record.Layout.Displayed
		if record.Rect.Y+record.Rect.H > tr.Rect.Y+tr.Rect.H+.02 {
			return nil
		}
	}
	last := records[len(records)-1]
	lastOffset := last.Rect.Y - records[0].Rect.Y
	tr.Layout.AllocationHeight = lastOffset + last.Layout.AllocationHeight
	tr.Layout.EstimatedOccupiedHeight = lastOffset + last.Layout.OccupiedTop + last.Layout.EstimatedOccupiedHeight - tr.Layout.OccupiedTop
	tr.Layout.VerticalPolicy = "native-v1 preserves source-measured card paragraph positions; resize one box for added copy"
	tr.Rich = &RichTextLayout{Contract: RichTextContract, Paragraphs: paragraphs}
	var line *pptx.ShapeLineProps
	if shape.Props.Line.Type != "none" {
		copy := *shape.Props.Line
		line = &copy
	}
	tr.NativeShape = &NativeTextShape{Rect: outer, Fill: shape.Props.Fill.Color, Line: line, ParagraphContract: EditableCardContract}
	out := *p
	out.Groups = nil
	out.Items = []sceneItem{{Text: &tr}}
	out.Warnings = append([]string{}, p.Warnings...)
	out.Warnings = append(out.Warnings, "Stock native card role/source adoption is manual; shared template bindings remain unchanged.")
	return &out
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func nativeCardSourceParagraph(tr TextRecord, source, key string) (RichParagraphLayout, bool) {
	if tr.Rich == nil {
		if !editableCardPlain(source) || tr.Layout.Original != source {
			return RichParagraphLayout{}, false
		}
		displayed := source
		if tr.Layout.Style.Case == "upper" {
			displayed = strings.ToUpper(source)
		}
		if tr.Layout.Displayed != displayed {
			return RichParagraphLayout{}, false
		}
		return RichParagraphLayout{Key: key, Displayed: tr.Layout.Displayed, LineCount: len(tr.Layout.Lines), Runs: []RichRunLayout{{Key: "text", Original: source, Displayed: tr.Layout.Displayed, Style: tr.Layout.Style, Font: tr.Layout.Font, Color: tr.Color, Start: 0, End: len([]rune(tr.Layout.Displayed))}}}, true
	}
	if !strings.Contains(source, "[^") || strings.Contains(source, "[[") || strings.Contains(source, "]]") || tr.Layout.Original == "" || tr.Rich.Contract != RichTextContract || len(tr.Rich.Paragraphs) != 1 {
		return RichParagraphLayout{}, false
	}
	paragraph := tr.Rich.Paragraphs[0]
	if len(paragraph.Runs) == 0 || len(paragraph.LineBreaks) != 0 || paragraph.Displayed != tr.Layout.Displayed {
		return RichParagraphLayout{}, false
	}
	for _, run := range paragraph.Runs {
		if run.Displayed == "" || run.Start < 0 || run.End < run.Start || run.End > len([]rune(paragraph.Displayed)) {
			return RichParagraphLayout{}, false
		}
	}
	paragraph.Key = key
	return paragraph, true
}
