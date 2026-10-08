package wmdesign

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"
)

// nativeProfileList folds measured source bullet rows into native paragraph
// bullets. It accepts simple strings and lead/body rows, but keeps nested or
// otherwise structurally complex lists as their original objects.
func (r *renderer) nativeProfileList(p *scenePlan, raw json.RawMessage) *scenePlan {
	candidate, _ := r.nativeProfileListResult(p, raw)
	return candidate
}

func (r *renderer) nativeProfileListResult(p *scenePlan, raw json.RawMessage) (*scenePlan, string) {
	var source struct {
		X, Y, W float64
		Size    string            `json:"size"`
		Items   []json.RawMessage `json:"items"`
	}
	if json.Unmarshal(raw, &source) != nil || len(source.Items) == 0 || len(source.Items) > 100 {
		return nil, "list source is invalid or empty"
	}
	if nativeProfileListHasNestedItems(source.Items) {
		return nil, "nested bullet markers use source-specific outline geometry and cannot be represented by the solid native paragraph marker"
	}
	if len(p.Items) != len(source.Items)*2 {
		return nil, "nested bullets or extra objects prevent a one-row-per-paragraph conversion"
	}
	type parsedItem struct {
		plain  bool
		markup bool
		source string
		text   string
	}
	items := make([]parsedItem, len(source.Items))
	for i, rawItem := range source.Items {
		if json.Unmarshal(rawItem, &items[i].text) == nil {
			items[i].plain = true
			items[i].source = items[i].text
			items[i].markup = strings.Contains(items[i].text, "[[") || strings.Contains(items[i].text, "]]")
			if strings.Contains(items[i].text, "[^") {
				return nil, "footnote bullets need their source footnote objects and cannot be flattened"
			}
			continue
		}
		var rich struct {
			Lead string            `json:"lead"`
			Text string            `json:"text"`
			Sub  []json.RawMessage `json:"sub"`
		}
		if json.Unmarshal(rawItem, &rich) != nil || len(rich.Sub) != 0 || rich.Lead == "" || rich.Text == "" {
			if len(rich.Sub) != 0 {
				return nil, "nested bullet markers use source-specific outline geometry and cannot be represented by the solid native paragraph marker"
			}
			return nil, "list contains a nested, empty, or unsupported rich item"
		}
		items[i].text = rich.Lead + " " + rich.Text
		items[i].source = items[i].text
	}
	var records []TextRecord
	var paragraphs []RichParagraphLayout
	for i, item := range items {
		marker, tr := p.Items[2*i].Shape, p.Items[2*i+1].Text
		if marker == nil || tr == nil {
			return nil, "list row is not a marker followed by one text record"
		}
		if tr.Align != "left" || tr.Rotation != 0 || tr.NativeShape != nil {
			return nil, "list item is transformed, aligned, rotated, or differs from its measured source copy"
		}
		if item.markup {
			normalized := strings.ReplaceAll(strings.ReplaceAll(item.source, "[[", ""), "]]", "")
			if tr.Rich == nil || len(tr.Rich.Paragraphs) != 1 || tr.Layout.Original != normalized || tr.Layout.Displayed != tr.Rich.Paragraphs[0].Displayed {
				return nil, "inline markup did not produce source-matching measured rich runs"
			}
		} else if tr.Layout.Displayed != item.text || tr.Layout.Original != tr.Layout.Displayed {
			return nil, "list item differs from its measured source copy"
		}
		if marker.Record.Rect.W <= 0 || math.Abs(marker.Record.Rect.W-marker.Record.Rect.H) > .02 {
			return nil, "list marker is not a measured square"
		}
		if marker.Props.Fill == nil || marker.Props.Fill.Type == "none" || marker.Props.Line != nil && marker.Props.Line.Type != "none" {
			return nil, "outlined source bullet marker differs from the solid native paragraph marker"
		}
		para, ok := nativeProfileListParagraph(tr, strings.TrimSuffix(strings.TrimPrefix(tr.ID, p.ID+"."), ".text"), item.plain && !item.markup)
		if !ok {
			return nil, "list item contains unsupported rich runs or paragraph breaks"
		}
		para.Bullet = true
		para.BulletIndentPt = tr.Rect.X - source.X
		para.BulletMarkerPt = marker.Record.Rect.W
		para.BulletColor = marker.Record.Color
		if para.BulletColor == "" && marker.Props.Fill != nil {
			para.BulletColor = marker.Props.Fill.Color
		}
		if para.BulletColor == "" {
			return nil, "list marker has no measured fill color"
		}
		if para.BulletIndentPt <= 0 || math.Abs(tr.Rect.X+tr.Rect.W-source.X-source.W) > .02 {
			return nil, "list text geometry has incompatible measured indent or width"
		}
		if i > 0 && (tr.Color != records[0].Color || tr.Layout.Style != records[0].Layout.Style || !reflect.DeepEqual(tr.Layout.Font, records[0].Layout.Font) || math.Abs(tr.Rect.X-records[0].Rect.X) > .02) {
			return nil, "list rows do not share compatible measured typography and alignment"
		}
		records = append(records, *tr)
		paragraphs = append(paragraphs, para)
	}
	tr := records[0]
	tr.ID = p.ID
	tr.Rect.X = source.X
	tr.Rect.W = source.W
	tr.Layout.Lines = nil
	bottom := tr.Rect.Y
	for i, rec := range records {
		offset := rec.Rect.Y - tr.Rect.Y
		paragraphs[i].FirstLine = len(tr.Layout.Lines)
		for _, line := range rec.Layout.Lines {
			line.Baseline += offset
			tr.Layout.Lines = append(tr.Layout.Lines, line)
		}
		gap := 6.
		suffix := ""
		if source.Size == "small" {
			gap, suffix = 3, "-small"
		}
		if r.source.Tokens.Density != nil {
			gap = r.listMetric("list-gap"+suffix, gap)
		}
		if i+1 < len(records) {
			gap = records[i+1].Rect.Y - rec.Rect.Y - rec.Layout.AllocationHeight
		}
		if gap < -.02 {
			return nil, "measured list rows overlap"
		}
		paragraphs[i].ParagraphGapAfter = math.Max(0, gap)
		bottom = math.Max(bottom, rec.Rect.Y+math.Max(rec.Layout.AllocationHeight, rec.Layout.OccupiedTop+rec.Layout.EstimatedOccupiedHeight))
	}
	if len(records) > 1 {
		paragraphs[len(paragraphs)-1].ParagraphGapAfter = paragraphs[len(paragraphs)-2].ParagraphGapAfter
	}
	tr.Rect.H = bottom - tr.Rect.Y
	originals := make([]string, len(records))
	displayed := make([]string, len(records))
	for i := range records {
		originals[i] = records[i].Layout.Original
		displayed[i] = records[i].Layout.Displayed
	}
	tr.Layout.Original = strings.Join(originals, "\n")
	tr.Layout.Displayed = strings.Join(displayed, "\n")
	last := records[len(records)-1]
	offset := last.Rect.Y - tr.Rect.Y
	tr.Layout.AllocationHeight = offset + last.Layout.AllocationHeight
	tr.Layout.EstimatedOccupiedHeight = offset + last.Layout.OccupiedTop + last.Layout.EstimatedOccupiedHeight - tr.Layout.OccupiedTop
	tr.Layout.VerticalPolicy = "native-v1 preserves source-measured list paragraph positions; resize one box for added copy"
	tr.Rich = &RichTextLayout{Contract: RichTextContract, Paragraphs: paragraphs}
	tr.NativeParagraphContract = EditableListContract
	out := *p
	out.Groups = nil
	out.Bounds = tr.Rect
	out.Items = []sceneItem{{Text: &tr}}
	out.Warnings = append(append([]string{}, p.Warnings...), "Native square bullets require the Wingdings desktop font (not redistributed). Initial height retains measured source allocation; added copy may require resizing. Bullet identity/source adoption is manual.")
	return &out, ""
}

func nativeProfileListHasNestedItems(items []json.RawMessage) bool {
	for _, raw := range items {
		var item struct {
			Sub []json.RawMessage `json:"sub"`
		}
		if json.Unmarshal(raw, &item) == nil && len(item.Sub) > 0 {
			return true
		}
	}
	return false
}

func nativeProfileListParagraph(tr *TextRecord, key string, plain bool) (RichParagraphLayout, bool) {
	if plain {
		return plainProfileParagraph(*tr, key)
	}
	if tr.Rich == nil || len(tr.Rich.Paragraphs) != 1 || len(tr.Rich.Paragraphs[0].LineBreaks) != 0 || len(tr.Rich.Paragraphs[0].Runs) < 2 {
		return RichParagraphLayout{}, false
	}
	para := tr.Rich.Paragraphs[0]
	if para.Displayed != tr.Layout.Displayed {
		return RichParagraphLayout{}, false
	}
	for _, run := range para.Runs {
		if run.Displayed == "" || strings.ContainsAny(run.Displayed, "\r\n") || run.BaselineShift != 0 {
			return RichParagraphLayout{}, false
		}
	}
	para.Key = key
	return para, true
}
