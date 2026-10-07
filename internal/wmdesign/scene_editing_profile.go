package wmdesign

import (
	"encoding/json"
	"fmt"
	"github.com/buairtri/pptxgengo/pptx"
	"math"
	"reflect"
	"strings"
)

// NativeEditingProfile changes native object structure only for eligible source
// scenes. It is persisted in authored source; frozen bundles remain unchanged.
const NativeEditingProfile = "native-v1"

func ValidateEditingProfile(profile string) error {
	if profile != "" && profile != "stock" && profile != NativeEditingProfile {
		return fmt.Errorf("unknown editing profile: %s", profile)
	}
	return nil
}

func (r *renderer) applyNativeEditingProfile(p *scenePlan, raw json.RawMessage, ctx SceneContext) *scenePlan {
	if r.editingProfile != NativeEditingProfile || p == nil {
		return p
	}
	var tag struct {
		Type string `json:"type"`
	}
	if json.Unmarshal(raw, &tag) != nil {
		return p
	}
	var candidate *scenePlan
	switch tag.Type {
	case "bullets":
		candidate = r.nativeProfileList(p, raw)
	case "card":
		candidate = nativeProfileCard(p, raw)
	case "table":
		if len(p.Items) == 1 && p.Items[0].Table != nil {
			copy := *p
			copy.Groups = nil
			candidate = &copy
		}
	default:
		return p
	}
	if candidate != nil {
		if err := sceneTextEnvelope(candidate, ctx); err != nil {
			p.Warnings = append(p.Warnings, "native-v1 retained source object structure: "+p.ID+" (converted text allocation crosses the declared content envelope).")
			return p
		}
	}
	if candidate == nil {
		p.Warnings = append(p.Warnings, "native-v1 retained source object structure: "+p.ID+" (requires a simple plain component with compatible measured geometry).")
		return p
	}
	candidate.Warnings = append(candidate.Warnings, "native-v1 converted source component: "+p.ID+"; source-resolved geometry, fonts and colors retained. Native edits require review; catalog previews do not qualify this editing profile.")
	return candidate
}

func plainProfileParagraph(tr TextRecord, key string) (RichParagraphLayout, bool) {
	if !editableCardPlain(tr.Layout.Displayed) || tr.Layout.Original != tr.Layout.Displayed || tr.Align != "left" || tr.Rotation != 0 || tr.NativeShape != nil {
		return RichParagraphLayout{}, false
	}
	if tr.Rich != nil {
		if len(tr.Rich.Paragraphs) != 1 || len(tr.Rich.Paragraphs[0].Runs) != 1 || len(tr.Rich.Paragraphs[0].LineBreaks) != 0 {
			return RichParagraphLayout{}, false
		}
		para := tr.Rich.Paragraphs[0]
		para.Key = key
		return para, true
	}
	return RichParagraphLayout{Key: key, Displayed: tr.Layout.Displayed, LineCount: len(tr.Layout.Lines), Runs: []RichRunLayout{{Key: "text", Original: tr.Layout.Original, Displayed: tr.Layout.Displayed, Style: tr.Layout.Style, Font: tr.Layout.Font, Color: tr.Color, Start: 0, End: len([]rune(tr.Layout.Displayed))}}}, true
}

func (r *renderer) nativeProfileList(p *scenePlan, raw json.RawMessage) *scenePlan {
	var source struct {
		X, Y, W float64
		Size    string
		Items   []string
	}
	if json.Unmarshal(raw, &source) != nil || len(source.Items) == 0 || len(source.Items) > 100 || len(p.Items) != len(source.Items)*2 {
		return nil
	}
	var records []TextRecord
	var paragraphs []RichParagraphLayout
	for i, text := range source.Items {
		marker, tr := p.Items[2*i].Shape, p.Items[2*i+1].Text
		if marker == nil || tr == nil || tr.Layout.Displayed != text || marker.Record.Rect.W <= 0 || math.Abs(marker.Record.Rect.W-marker.Record.Rect.H) > .02 {
			return nil
		}
		para, ok := plainProfileParagraph(*tr, strings.TrimSuffix(strings.TrimPrefix(tr.ID, p.ID+"."), ".text"))
		if !ok {
			return nil
		}
		para.Bullet = true
		para.BulletIndentPt = tr.Rect.X - source.X
		para.BulletMarkerPt = marker.Record.Rect.W
		if para.BulletIndentPt <= 0 || math.Abs(tr.Rect.X+tr.Rect.W-source.X-source.W) > .02 {
			return nil
		}
		if i > 0 && (tr.Color != records[0].Color || tr.Layout.Style != records[0].Layout.Style || !reflect.DeepEqual(tr.Layout.Font, records[0].Layout.Font) || math.Abs(tr.Rect.X-records[0].Rect.X) > .02) {
			return nil
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
			gap = 3
			suffix = "-small"
		}
		if r.source.Tokens.Density != nil {
			gap = r.listMetric("list-gap"+suffix, gap)
		}
		if i+1 < len(records) {
			gap = records[i+1].Rect.Y - rec.Rect.Y - rec.Layout.AllocationHeight
		}
		if gap < -.02 {
			return nil
		}
		paragraphs[i].ParagraphGapAfter = math.Max(0, gap)
		bottom = math.Max(bottom, rec.Rect.Y+math.Max(rec.Layout.AllocationHeight, rec.Layout.OccupiedTop+rec.Layout.EstimatedOccupiedHeight))
	}
	if len(records) > 1 {
		paragraphs[len(records)-1].ParagraphGapAfter = paragraphs[len(records)-2].ParagraphGapAfter
	}
	tr.Rect.H = bottom - tr.Rect.Y
	tr.Layout.Original = strings.Join(source.Items, "\n")
	tr.Layout.Displayed = tr.Layout.Original
	last := records[len(records)-1]
	offset := last.Rect.Y - tr.Rect.Y
	tr.Layout.AllocationHeight = offset + last.Layout.AllocationHeight
	tr.Layout.EstimatedOccupiedHeight = offset + last.Layout.OccupiedTop + last.Layout.EstimatedOccupiedHeight - tr.Layout.OccupiedTop
	tr.Layout.VerticalPolicy = "native-v1 preserves source-measured flat list paragraph positions; resize one box for added copy"
	tr.Rich = &RichTextLayout{Contract: RichTextContract, Paragraphs: paragraphs}
	tr.NativeParagraphContract = EditableListContract
	out := *p
	out.Groups = nil
	out.Bounds = tr.Rect
	out.Items = []sceneItem{{Text: &tr}}
	out.Warnings = append(append([]string{}, p.Warnings...), "Native square bullets require the Wingdings desktop font (not redistributed). Initial height retains measured source allocation; added copy may require resizing. Bullet identity/source adoption is manual.")
	return &out
}

func nativeProfileCard(p *scenePlan, raw json.RawMessage) *scenePlan {
	var source sceneCardSource
	if json.Unmarshal(raw, &source) != nil || len(source.Body) != 1 || source.BodySize != "" || source.State != "" || source.ContentBottom != 0 || source.Surface == "outline" || len(p.Items) != 3 {
		return nil
	}
	shape, title, body := p.Items[0].Shape, p.Items[1].Text, p.Items[2].Text
	if shape == nil || title == nil || body == nil || shape.Type != pptx.ShapeTypeRect || shape.Props.Fill == nil || shape.Props.Line == nil || shape.Props.Line.Type != "none" || shape.Record.Rotation != 0 || title.ID != p.ID+".title" || !strings.HasPrefix(body.ID, p.ID+".body.") {
		return nil
	}
	var block map[string]json.RawMessage
	if json.Unmarshal(source.Body[0], &block) != nil || len(block) != 1 || block["p"] == nil {
		return nil
	}
	if math.Abs(title.Rect.X-body.Rect.X) > .02 || math.Abs(title.Rect.W-body.Rect.W) > .02 {
		return nil
	}
	a, ok := plainProfileParagraph(*title, "title")
	if !ok {
		return nil
	}
	b, ok := plainProfileParagraph(*body, "body")
	if !ok {
		return nil
	}
	gap := body.Rect.Y - title.Rect.Y - title.Layout.AllocationHeight
	if gap < -.02 {
		return nil
	}
	a.ParagraphGapAfter = math.Max(0, gap)
	tr := *title
	tr.ID = p.ID
	outer := shape.Record.Rect
	pad := title.Rect.X - outer.X
	tr.Rect.H = outer.Y + outer.H - title.Rect.Y - pad
	if tr.Rect.H <= 0 || !inside(tr.Rect, outer) || body.Rect.Y+body.Rect.H > tr.Rect.Y+tr.Rect.H+.02 {
		return nil
	}
	tr.Layout.Lines = append([]TextLine{}, title.Layout.Lines...)
	offset := body.Rect.Y - title.Rect.Y
	for _, line := range body.Layout.Lines {
		line.Baseline += offset
		tr.Layout.Lines = append(tr.Layout.Lines, line)
	}
	b.FirstLine = len(title.Layout.Lines)
	tr.Layout.Original = title.Layout.Original + "\n" + body.Layout.Original
	tr.Layout.Displayed = tr.Layout.Original
	tr.Layout.AllocationHeight = offset + body.Layout.AllocationHeight
	tr.Layout.EstimatedOccupiedHeight = offset + body.Layout.OccupiedTop + body.Layout.EstimatedOccupiedHeight - tr.Layout.OccupiedTop
	tr.Rich = &RichTextLayout{Contract: RichTextContract, Paragraphs: []RichParagraphLayout{a, b}}
	tr.NativeShape = &NativeTextShape{Rect: outer, Fill: shape.Props.Fill.Color, ParagraphContract: EditableCardContract}
	out := *p
	out.Groups = nil
	out.Items = []sceneItem{{Text: &tr}}
	out.Warnings = append([]string{}, p.Warnings...)
	out.Warnings = append(out.Warnings, "Stock native card role/source adoption is manual; shared template bindings remain unchanged.")
	return &out
}
