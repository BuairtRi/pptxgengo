package wmdesign

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/buairtri/pptxgengo/pptx"
)

// The expanded V12 catalog introduces editable dot-leader contents and source
// preset markers. Keep these semantics behind the explicit frozen revision.
func (r *renderer) planV12CatalogScene(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, bool, error) {
	var head struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &head); err != nil {
		return nil, false, err
	}
	if r.source.Revision != LibraryRevisionV12 || (head.Type != "toclist" && head.Type != "preset") {
		return nil, false, nil
	}
	p := &scenePlan{ID: id}
	if head.Type == "preset" {
		var n struct {
			Type    string  `json:"type"`
			Shape   string  `json:"shape"`
			Surface string  `json:"surface"`
			X       float64 `json:"x"`
			Y       float64 `json:"y"`
			W       float64 `json:"w"`
			H       float64 `json:"h"`
		}
		if err := sceneDecode(raw, &n); err != nil {
			return nil, true, err
		}
		if !intakeFinite(n.X, n.Y, n.W, n.H) || n.W <= 0 || n.H <= 0 || (n.Shape != "diamond" && n.Shape != "rect") {
			return nil, true, fmt.Errorf("scene.v12_invalid_preset: %s", id)
		}
		fill, err := r.sceneColor(n.Surface, "bg")
		if err != nil {
			return nil, true, err
		}
		kind := pptx.ShapeTypeDiamond
		if n.Shape == "rect" {
			kind = pptx.ShapeTypeRect
		}
		if err = r.diagramShape(p, id+".marker", Rect{n.X, n.Y, n.W, n.H}, kind, fill, "", 0, "", nil); err != nil {
			return nil, true, err
		}
		return diagramFinish(p, "scene.preset"), true, nil
	}
	var n struct {
		Type  string   `json:"type"`
		X     float64  `json:"x"`
		Y     float64  `json:"y"`
		W     float64  `json:"w"`
		Gap   *float64 `json:"gap"`
		NumW  float64  `json:"numW"`
		Rules bool     `json:"rules"`
		Items []struct {
			N     string `json:"n"`
			Text  string `json:"text"`
			Page  string `json:"page"`
			Level int    `json:"level"`
		} `json:"items"`
	}
	if err := sceneDecode(raw, &n); err != nil {
		return nil, true, err
	}
	gap, numW := 6., n.NumW
	if n.Gap != nil {
		gap = *n.Gap
	}
	if numW == 0 {
		numW = 36
	}
	if !intakeFinite(n.X, n.Y, n.W, gap, numW) || n.W <= numW+45 || gap < 0 || numW < 0 || len(n.Items) == 0 || len(n.Items) > 100 {
		return nil, true, fmt.Errorf("scene.v12_invalid_toc: %s", id)
	}
	keys, err := primitiveArrayKeys(ctx, "/items", len(n.Items))
	if err != nil {
		return nil, true, err
	}
	y := n.Y
	for i, it := range n.Items {
		level := it.Level
		if level == 0 {
			level = 1
		}
		if level != 1 && level != 2 {
			return nil, true, fmt.Errorf("scene.v12_invalid_toc_level: %s", id)
		}
		pre := id + ".item." + keys[i]
		if level == 1 && i > 0 && n.Rules {
			line, e := r.sceneColor(ctx.Surface, "line")
			if e != nil {
				return nil, true, e
			}
			y += 6
			if e = r.diagramLine(p, pre+".rule", [2]float64{n.X, y}, [2]float64{n.X + n.W, y}, line, .75, "solid"); e != nil {
				return nil, true, e
			}
			y += 12
		}
		token, ink, numToken, numInk := "body", "secondary", "body", "secondary"
		if level == 1 {
			token, ink, numToken, numInk = "subhead", "display", "number", "emphasis"
		}
		st, e := r.sceneStyle(token)
		if e != nil {
			return nil, true, e
		}
		ns, e := r.sceneStyle(numToken)
		if e != nil {
			return nil, true, e
		}
		mono, e := r.sceneStyle("label")
		if e != nil {
			return nil, true, e
		}
		ns.Family = mono.Family
		ns.Weight = 600
		layout, e := r.measureText(it.Text, st, 1920)
		if e != nil {
			return nil, true, e
		}
		tw := 0.
		for _, l := range layout.Lines {
			tw = math.Max(tw, sequenceInlineWidth(l.Advance, st))
		}
		pl, e := r.measureText(it.Page, ns, 1920)
		if e != nil {
			return nil, true, e
		}
		pw := 27.
		for _, l := range pl.Lines {
			pw = math.Max(pw, sequenceInlineWidth(l.Advance, ns))
		}
		available := n.W - numW - pw - 36
		if tw > available+.02 {
			return nil, true, fmt.Errorf("scene.toc_horizontal_overflow: %s needs %.2fpt, capacity %.2fpt", pre, tw, available)
		}
		h, _, e := r.sequenceNeed(it.Text, st, tw+.1)
		if e != nil {
			return nil, true, e
		}
		nh, _, e := r.sequenceNeed(it.Page, ns, pw)
		if e != nil {
			return nil, true, e
		}
		h = math.Max(h, nh)
		if e = r.diagramStyledText(p, pre+".number", it.N, ns, Rect{n.X, y, numW, h}, ctx.Surface, "emphasis", "left", false); e != nil {
			return nil, true, e
		}
		if e = r.diagramStyledText(p, pre+".text", it.Text, st, Rect{n.X + numW, y, tw + .1, h}, ctx.Surface, ink, "left", false); e != nil {
			return nil, true, e
		}
		if e = r.diagramStyledText(p, pre+".page", it.Page, ns, Rect{n.X + n.W - pw, y, pw, h}, ctx.Surface, numInk, "right", false); e != nil {
			return nil, true, e
		}
		color, e := r.sceneColor(ctx.Surface, "secondary")
		if e != nil {
			return nil, true, e
		}
		if e = r.diagramLine(p, pre+".leader", [2]float64{n.X + numW + tw + 9, y + h - 4}, [2]float64{n.X + n.W - pw - 9, y + h - 4}, color, 1.5, "dot"); e != nil {
			return nil, true, e
		}
		y += h + gap
	}
	return diagramFinish(p, "scene.toclist"), true, nil
}
