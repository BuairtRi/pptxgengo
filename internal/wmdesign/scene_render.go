package wmdesign

import (
	"encoding/json"
	"fmt"
	"github.com/buairtri/pptxgengo/pptx"
)

type SceneRecord struct {
	Definition    string            `json:"definition,omitempty"`
	ID            string            `json:"id"`
	SourcePointer string            `json:"source_pointer"`
	Bounds        Rect              `json:"rect"`
	Parts         []string          `json:"parts"`
	Groups        []ComponentRecord `json:"groups,omitempty"`
	Warnings      []string          `json:"warnings,omitempty"`
}

type SceneTableRecord struct {
	ID      string            `json:"id"`
	Rect    Rect              `json:"rect"`
	Rows    int               `json:"rows"`
	Columns int               `json:"columns"`
	Cells   []SceneCellRecord `json:"cells"`
}

type SceneCellRecord struct {
	Row    int        `json:"row"`
	Column int        `json:"column"`
	Text   TextRecord `json:"text"`
}

type SceneChartRecord struct {
	ID             string           `json:"id"`
	Rect           Rect             `json:"rect"`
	Type           pptx.ChartType   `json:"type"`
	Data           []pptx.ChartData `json:"data"`
	NativeEditable bool             `json:"native_editable"`
}

func (r *renderer) drawScene(p *scenePlan, sr *SlideReport, path string) error {
	if p == nil {
		return fmt.Errorf("scene.empty_plan")
	}
	record := SceneRecord{Definition: p.Definition, ID: p.ID, SourcePointer: path, Bounds: p.Bounds, Groups: p.Groups, Warnings: p.Warnings}
	for _, item := range p.Items {
		count := 0
		for _, present := range []bool{item.Shape != nil, item.Text != nil, item.Image != nil, item.Table != nil, item.Chart != nil} {
			if present {
				count++
			}
		}
		if count != 1 {
			return fmt.Errorf("scene.invalid_item_union: %s", p.ID)
		}
		switch {
		case item.Shape != nil:
			sh := item.Shape
			if sh.Props.ObjectName == "" {
				sh.Props.ObjectName = sh.Record.ID
			}
			if sh.Connection != nil {
				if err := r.slide.AddConnector(&pptx.ConnectorProps{ShapeProps: sh.Props, Connection: *sh.Connection}); err != nil {
					return err
				}
			} else if err := r.slide.AddShape(sh.Type, &sh.Props); err != nil {
				return err
			}
			sr.Shapes = append(sr.Shapes, sh.Record)
			record.Parts = append(record.Parts, sh.Props.ObjectName)
		case item.Text != nil:
			tr := *item.Text
			s, id := tr.Layout.Style, tr.Layout.Font
			valign := pptx.VAlign("top")
			if tr.VerticalAlign != "" {
				valign = pptx.VAlign(tr.VerticalAlign)
			}
			opts := &pptx.TextPropsOptions{PositionProps: pos(tr.Rect), ObjectNameProps: pptx.ObjectNameProps{ObjectName: tr.ID}, TextBaseProps: pptx.TextBaseProps{FontFace: id.Typeface, FontSize: s.Size, Bold: &id.Bold, Italic: &id.NativeItalic, Color: tr.Color, Align: pptx.HAlign(tr.Align)}, CharSpacing: s.TrackingPt, LineSpacing: s.Leading, ParaSpaceBefore: zero(), ParaSpaceAfter: zero(), Margin: pptx.Margin{0}, Fit: "none", Valign: valign, Rotate: tr.Rotation}
			if tr.NativeShape != nil {
				outer := tr.NativeShape.Rect
				if !inside(tr.Rect, outer) || (tr.Rich != nil && tr.NativeShape.ParagraphContract != EditableCardContract) || tr.Rotation != 0 {
					return fmt.Errorf("scene.invalid_combined_text_shape: %s", tr.ID)
				}
				opts.PositionProps = pos(outer)
				opts.Fill = &pptx.ShapeFillProps{Color: tr.NativeShape.Fill}
				opts.Line = &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Type: "none"}}
				opts.Shape = pptx.ShapeTypeRect
				isTextBox := false
				opts.IsTextBox = &isTextBox
				// The writer's Margin order is left/right/bottom/top, in points.
				opts.Margin = pptx.Margin{tr.Rect.X - outer.X, outer.X + outer.W - tr.Rect.X - tr.Rect.W, outer.Y + outer.H - tr.Rect.Y - tr.Rect.H, tr.Rect.Y - outer.Y}
			}
			if err := r.slide.AddText([]pptx.TextProps{{Text: tr.Layout.Displayed}}, opts); err != nil {
				return err
			}
			*r.records = append(*r.records, tr)
			record.Parts = append(record.Parts, tr.ID)
		case item.Image != nil:
			if err := r.slide.AddImage(item.Image); err != nil {
				return err
			}
			record.Parts = append(record.Parts, item.Image.ObjectName)
		case item.Table != nil:
			t := item.Table
			t.Options.PositionProps = pos(t.Rect)
			t.Options.ObjectName = t.ID
			if err := r.slide.AddTable(t.Rows, &t.Options); err != nil {
				return err
			}
			if len(r.slide.NewAutoPagedSlides()) != 0 {
				return fmt.Errorf("scene.table_unexpected_pagination: %s", t.ID)
			}
			tr := SceneTableRecord{ID: t.ID, Rect: t.Rect, Rows: len(t.Rows)}
			tr.Columns = len(t.Options.ColW)
			for _, cell := range t.CellRecords {
				tr.Cells = append(tr.Cells, SceneCellRecord{Row: cell.Row, Column: cell.Column, Text: cell.Text})
			}
			sr.Tables = append(sr.Tables, tr)
			record.Parts = append(record.Parts, t.ID)
		case item.Chart != nil:
			c := item.Chart
			c.Options.PositionProps = pos(c.Rect)
			c.Options.ObjectName = c.ID
			var err error
			if len(c.Options.MultiTypes) > 0 {
				err = r.slide.AddMultiChart(c.Options.MultiTypes, &c.Options)
			} else {
				err = r.slide.AddChart(c.Type, c.Data, &c.Options)
			}
			if err != nil {
				return err
			}
			sr.Charts = append(sr.Charts, SceneChartRecord{ID: c.ID, Rect: c.Rect, Type: c.Type, Data: c.Data, NativeEditable: true})
			record.Parts = append(record.Parts, c.ID)
		}
	}
	sr.Scenes = append(sr.Scenes, record)
	return nil
}

func (r *renderer) planSourceRule(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, bool, error) {
	var tag struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &tag); err != nil {
		return nil, false, err
	}
	if tag.Type != "rule" {
		return nil, false, nil
	}
	var n struct {
		Type   string  `json:"type"`
		X      float64 `json:"x"`
		Y      float64 `json:"y"`
		W      float64 `json:"w"`
		Weight float64 `json:"weight"`
		Ink    string  `json:"ink"`
		On     string  `json:"on"`
	}
	if err := sceneDecode(raw, &n); err != nil {
		return nil, true, err
	}
	if n.Weight == 0 {
		n.Weight = .75
	}
	if n.Ink == "" {
		n.Ink = "line"
	}
	if n.On != "" {
		ctx.Surface = n.On
	}
	color, err := r.sceneColor(ctx.Surface, n.Ink)
	if err != nil {
		return nil, true, err
	}
	b := Rect{n.X, n.Y, n.W, n.Weight}
	p := &scenePlan{ID: id, Bounds: b}
	props := pptx.ShapeProps{PositionProps: pos(b), ObjectNameProps: pptx.ObjectNameProps{ObjectName: id}, Fill: &pptx.ShapeFillProps{Color: color}, Line: &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Type: "none"}}}
	p.Items = append(p.Items, sceneItem{Shape: &sceneShape{Type: pptx.ShapeTypeRect, Props: props, Record: ShapeRecord{ID: id, Rect: b, Color: color}}})
	return p, true, nil
}
