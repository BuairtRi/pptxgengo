package wmdesign

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

type PortfolioHorizon struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	Meaning string `json:"meaning"`
}
type PortfolioStatus struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	Surface string `json:"surface"`
}
type PortfolioInitiative struct {
	Key        string `json:"key"`
	Label      string `json:"label"`
	Horizon    string `json:"horizon"`
	Owner      string `json:"owner"`
	Status     string `json:"status"`
	Confidence string `json:"confidence"`
	Slot       int    `json:"slot"`
}
type PortfolioDependency struct {
	Key           string       `json:"key"`
	From          string       `json:"from"`
	To            string       `json:"to"`
	Label         string       `json:"label,omitempty"`
	Route         [][2]float64 `json:"route,omitempty"`
	LabelPosition *[2]float64  `json:"label_position,omitempty"`
}

func (dependency *PortfolioDependency) UnmarshalJSON(raw []byte) error {
	type plain PortfolioDependency
	var value plain
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&value); err != nil {
		return err
	}
	// Reuse the checked process coordinate contract without accepting process
	// outcome names or unrelated fields in the portfolio source schema.
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return err
	}
	coordinates := map[string]json.RawMessage{}
	for _, key := range []string{"route", "label_position"} {
		if v, present := fields[key]; present {
			coordinates[key] = v
		}
	}
	encoded, err := json.Marshal(coordinates)
	if err != nil {
		return err
	}
	var checked ProcessLink
	if err = json.Unmarshal(encoded, &checked); err != nil {
		return fmt.Errorf("portfolio dependency coordinates: %w", err)
	}
	*dependency = PortfolioDependency(value)
	return nil
}

type PortfolioSpec struct {
	Type         string                `json:"type,omitempty"`
	X            float64               `json:"x,omitempty"`
	Y            float64               `json:"y,omitempty"`
	W            float64               `json:"w,omitempty"`
	H            float64               `json:"h,omitempty"`
	Horizons     []PortfolioHorizon    `json:"horizons"`
	Statuses     []PortfolioStatus     `json:"statuses"`
	Initiatives  []PortfolioInitiative `json:"initiatives"`
	Dependencies []PortfolioDependency `json:"dependencies"`
	CardHeight   float64               `json:"card_height_pt"`
}

func ValidatePortfolio(s PortfolioSpec) error {
	if len(s.Horizons) < 1 || len(s.Horizons) > 8 || len(s.Statuses) < 1 || len(s.Statuses) > 12 || len(s.Initiatives) < 1 || len(s.Initiatives) > 80 || len(s.Dependencies) > 160 || s.CardHeight < 42 || !intakeFinite(s.CardHeight) {
		return fmt.Errorf("portfolio invalid counts/card height")
	}
	horizons, statuses, initiatives, slots, edges := map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, h := range s.Horizons {
		if !validPartKey(h.Key) || horizons[h.Key] || strings.TrimSpace(h.Label) == "" || strings.TrimSpace(h.Meaning) == "" {
			return fmt.Errorf("portfolio horizon needs unique key, label and explicit non-date meaning")
		}
		horizons[h.Key] = true
	}
	for _, v := range s.Statuses {
		if !validPartKey(v.Key) || statuses[v.Key] || strings.TrimSpace(v.Label) == "" || strings.TrimSpace(v.Surface) == "" {
			return fmt.Errorf("invalid portfolio status")
		}
		statuses[v.Key] = true
	}
	for _, v := range s.Initiatives {
		slot := fmt.Sprintf("%s/%d", v.Horizon, v.Slot)
		if !validPartKey(v.Key) || initiatives[v.Key] || !horizons[v.Horizon] || !statuses[v.Status] || v.Slot < 0 || v.Slot > 11 || slots[slot] || strings.TrimSpace(v.Label) == "" || strings.TrimSpace(v.Owner) == "" || strings.TrimSpace(v.Confidence) == "" {
			return fmt.Errorf("invalid portfolio initiative %s", v.Key)
		}
		initiatives[v.Key] = true
		slots[slot] = true
	}
	for _, e := range s.Dependencies {
		if !validPartKey(e.Key) || edges[e.Key] || !initiatives[e.From] || !initiatives[e.To] || e.From == e.To || len(e.Route) > 64 {
			return fmt.Errorf("invalid portfolio dependency %s", e.Key)
		}
		edges[e.Key] = true
		if e.LabelPosition != nil && (strings.TrimSpace(e.Label) == "" || !intakeFinite(e.LabelPosition[0], e.LabelPosition[1])) {
			return fmt.Errorf("portfolio dependency %s label_position requires a finite named label", e.Key)
		}
		for _, p := range e.Route {
			if !intakeFinite(p[0], p[1]) {
				return fmt.Errorf("nonfinite portfolio route")
			}
		}
	}
	return nil
}
func (r *renderer) planPortfolioScene(id string, raw json.RawMessage, ctx SceneContext) (*scenePlan, bool, error) {
	var tag struct {
		Type string `json:"type"`
	}
	if e := json.Unmarshal(raw, &tag); e != nil {
		return nil, false, e
	}
	if tag.Type != "portfolio" {
		return nil, false, nil
	}
	var s PortfolioSpec
	if e := sceneDecode(raw, &s); e != nil {
		return nil, true, e
	}
	if e := ValidatePortfolio(s); e != nil {
		return nil, true, e
	}
	if !intakeFinite(s.X, s.Y, s.W, s.H) || s.W < 120 || s.H < 100 {
		return nil, true, fmt.Errorf("invalid portfolio allocation")
	}
	header := 42.
	cw := s.W / float64(len(s.Horizons))
	maxSlot := 0
	for _, v := range s.Initiatives {
		if v.Slot > maxSlot {
			maxSlot = v.Slot
		}
	}
	pitch := (s.H - header - 12) / float64(maxSlot+1)
	if pitch < s.CardHeight+12 || cw < 72 {
		return nil, true, fmt.Errorf("portfolio allocation too small for horizons/cards")
	}
	p := &scenePlan{ID: id}
	surface := ctx.Surface
	if surface == "" {
		surface = "light"
	}
	hidx := map[string]int{}
	status := map[string]PortfolioStatus{}
	_, e := r.sceneColor(surface, "strong")
	if e != nil {
		return nil, true, e
	}
	line, _ := r.sceneColor(surface, "line")
	for _, v := range s.Statuses {
		status[v.Key] = v
		if _, e = r.sceneColor(v.Surface, "bg"); e != nil {
			return nil, true, e
		}
	}
	for i, h := range s.Horizons {
		hidx[h.Key] = i
		x := s.X + float64(i)*cw
		sf := "light"
		if i%2 == 1 {
			sf = "subtle"
		}
		if e = r.sceneRect(p, id+".horizons."+h.Key+".surface", Rect{x, s.Y, cw, s.H}, sf); e != nil {
			return nil, true, e
		}
		if e = r.diagramText(p, id+".horizons."+h.Key+".label", h.Label, "small", Rect{x + 8, s.Y + 2, cw - 16, 20}, surface, "display", "center", 600, true); e != nil {
			return nil, true, e
		}
		if e = r.diagramText(p, id+".horizons."+h.Key+".meaning", h.Meaning, "label", Rect{x + 8, s.Y + 22, cw - 16, 18}, surface, "secondary", "center", 0, true); e != nil {
			return nil, true, e
		}
		if e = r.diagramLine(p, id+".horizons."+h.Key+".rule", [2]float64{x, s.Y + header}, [2]float64{x + cw, s.Y + header}, line, .75, "solid"); e != nil {
			return nil, true, e
		}
	}
	// Lower only geometry, keeping horizon/confidence/status facts in the source.
	// Cards use the same keyed graph renderer's native shape and route construction.
	pm := ProcessSpec{Type: "process", X: s.X, Y: s.Y + header, W: s.W, H: s.H - header, Columns: len(s.Horizons), LabelWidth: 0, HideLaneLabels: true, StepHeight: s.CardHeight}
	for slot := 0; slot <= maxSlot; slot++ {
		pm.Lanes = append(pm.Lanes, ProcessLane{Key: fmt.Sprintf("slot-%d", slot), Label: fmt.Sprintf("Slot %d", slot+1)})
	}
	for _, v := range s.Initiatives {
		st := status[v.Status]
		label := v.Label + "\n" + v.Owner + " · " + st.Label + "\n" + v.Confidence
		pm.Steps = append(pm.Steps, ProcessStep{Key: v.Key, Label: label, Lane: fmt.Sprintf("slot-%d", v.Slot), Column: hidx[v.Horizon], Kind: "process", Surface: st.Surface})
	}
	for _, d := range s.Dependencies {
		pm.Links = append(pm.Links, ProcessLink{Key: d.Key, From: d.From, To: d.To, Outcome: d.Label, Route: d.Route, LabelPosition: d.LabelPosition})
	}
	rr, _ := json.Marshal(pm)
	child, _, e := r.planProcessScene(id+".initiatives", rr, ctx)
	if e != nil {
		return nil, true, e
	}
	diagramMerge(p, child)
	p = diagramFinish(p, "scene.portfolio")
	if p.Bounds.X < s.X-.01 || p.Bounds.Y < s.Y-.01 || p.Bounds.X+p.Bounds.W > s.X+s.W+.01 || p.Bounds.Y+p.Bounds.H > s.Y+s.H+.01 || math.IsNaN(p.Bounds.H) {
		return nil, true, fmt.Errorf("portfolio content exceeds allocation")
	}
	return p, true, nil
}
