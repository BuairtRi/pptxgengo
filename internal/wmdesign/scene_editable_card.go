package wmdesign

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// EditableCardContract is an explicit two-paragraph authoring pilot. It neither
// migrates stock cards nor qualifies native desktop selection or density parity.
const EditableCardContract = "pptxgengo.editable-card.v1"

type editableCardSource struct {
	Type    string  `json:"type"`
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	W       float64 `json:"w"`
	H       float64 `json:"h"`
	Surface string  `json:"surface,omitempty"`
	Title   string  `json:"title"`
	Body    string  `json:"body"`
}

func editableCardPlain(s string) bool {
	return strings.TrimSpace(s) != "" && !strings.ContainsAny(s, "\r\n\t") && !strings.Contains(s, "[[") && !strings.Contains(s, "]]") && !strings.Contains(s, "[^")
}

func (r *renderer) planEditableCard(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, error) {
	var n editableCardSource
	if err := sceneDecode(raw, &n); err != nil {
		return nil, err
	}
	b := Rect{n.X, n.Y, n.W, n.H}
	if b.W <= 24 || b.H <= 24 || math.IsNaN(b.X+b.Y+b.W+b.H) || math.IsInf(b.X+b.Y+b.W+b.H, 0) {
		return nil, fmt.Errorf("scene.editable_card_invalid_bounds")
	}
	if !editableCardPlain(n.Title) || !editableCardPlain(n.Body) {
		return nil, fmt.Errorf("scene.editable_card_requires_two_plain_single_paragraph_fields")
	}
	surface := n.Surface
	if surface == "" {
		surface = "light"
	}
	if surface == "outline" {
		return nil, fmt.Errorf("scene.editable_card_outline_unsupported")
	}
	p := &scenePlan{Definition: "scene.editable-card", ID: id, Bounds: b}
	if err := r.sceneRect(p, id, b, surface); err != nil {
		return nil, err
	}
	fill := p.Items[0].Shape.Props.Fill.Color
	inner := Rect{b.X + 12, b.Y + 12, b.W - 24, b.H - 24}
	records := []TextRecord{}
	y := inner.Y
	for i, copy := range []string{n.Title, n.Body} {
		role, ink := "subhead", "primary"
		if i == 1 {
			role, ink = "body", "secondary"
		}
		st, err := r.sceneDataToken(role, 0)
		if err != nil {
			return nil, err
		}
		rec, err := r.primitiveMeasureRuns(id, []primitiveRun{{Text: copy}}, st, Rect{X: inner.X, Y: y, W: inner.W}, surface, ink, "left")
		if err != nil {
			return nil, err
		}
		if rec.Layout.Displayed != copy || rec.Rich == nil || len(rec.Rich.Paragraphs) != 1 || len(rec.Rich.Paragraphs[0].Runs) != 1 {
			return nil, fmt.Errorf("scene.editable_card_nonplain_typography")
		}
		need := math.Max(rec.Layout.AllocationHeight, rec.Layout.EstimatedOccupiedHeight)
		if y+need > inner.Y+inner.H+.02 {
			return nil, fmt.Errorf("scene.editable_card_vertical_overflow: %s", id)
		}
		para := &rec.Rich.Paragraphs[0]
		para.Key = []string{"title", "body"}[i]
		if i == 0 {
			// Source card body-flow adds a 6pt paragraph top margin after the
			// card's 8pt title/body gap. Keep both in the native paragraph gap.
			para.ParagraphGapAfter = 14 + math.Max(0, need-rec.Layout.AllocationHeight)
		}
		records = append(records, rec)
		y += need + 14
	}
	title, body := records[0], records[1]
	tr := title
	tr.Rect = inner
	tr.Layout.Original = n.Title + "\n" + n.Body
	tr.Layout.Displayed = tr.Layout.Original
	tr.Layout.Lines = append([]TextLine{}, title.Layout.Lines...)
	offset := body.Rect.Y - title.Rect.Y
	for _, line := range body.Layout.Lines {
		line.Baseline += offset
		tr.Layout.Lines = append(tr.Layout.Lines, line)
	}
	tr.Layout.AllocationHeight = offset + body.Layout.AllocationHeight
	tr.Layout.EstimatedOccupiedHeight = offset + body.Layout.EstimatedOccupiedHeight
	tr.Layout.VerticalPolicy = "explicit two plain paragraphs: independently measured title/body roles; native desktop parity unqualified"
	body.Rich.Paragraphs[0].FirstLine = len(title.Layout.Lines)
	tr.Rich = &RichTextLayout{Contract: RichTextContract, Paragraphs: []RichParagraphLayout{title.Rich.Paragraphs[0], body.Rich.Paragraphs[0]}, NativeQualified: false}
	tr.NativeShape = &NativeTextShape{Rect: b, Fill: fill, ParagraphContract: EditableCardContract}
	p.Items = []sceneItem{{Text: &tr}}
	p.Warnings = []string{"Editable card pilot contains title/body paragraphs in one filled native rectangle; move/resize, desktop editing and density-wide visual qualification remain pending."}
	return p, nil
}

func editableCardParagraphXML(tr TextRecord) []byte {
	var out []byte
	for _, para := range tr.Rich.Paragraphs {
		one := tr
		one.Rich = &RichTextLayout{Contract: RichTextContract, Paragraphs: []RichParagraphLayout{para}}
		run := para.Runs[0]
		one.Layout.Style, one.Layout.Font, one.Color = run.Style, run.Font, run.Color
		out = append(out, richParagraphXML(one, sceneTableParagraphProperties(one, para.ParagraphGapAfter, false))...)
	}
	return out
}
