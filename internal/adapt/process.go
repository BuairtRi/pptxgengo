package adapt

import (
	"encoding/json"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/compose"
	"strings"
)

type processInput struct {
	Steps []processStep `json:"steps"`
	GapPt float64       `json:"gap_pt,omitempty"`
}
type processStep struct {
	ID         string   `json:"id"`
	Label      string   `json:"label"`
	Summary    string   `json:"summary"`
	Activities []string `json:"activities"`
	Outputs    []string `json:"outputs"`
	State      string   `json:"state,omitempty"`
}

func buildProcess(c Context, raw json.RawMessage) (Content, error) {
	var in processInput
	if e := Decode(raw, &in); e != nil {
		return Content{}, e
	}
	n := len(in.Steps)
	if n < 2 || n > 6 {
		return Content{}, fmt.Errorf("process requires 2..6 steps")
	}
	gap := in.GapPt
	if gap == 0 {
		gap = 18
	}
	if gap < 12 || gap > 36 {
		return Content{}, fmt.Errorf("process gap_pt must be 12..36")
	}
	width := (c.Bounds.Width - gap*float64(n-1)) / float64(n)
	if width < 115 {
		return Content{}, fmt.Errorf("process step width %.1fpt is below 115pt; reduce step count or enlarge bounds", width)
	}
	out := Content{Controls: map[string]any{"step_count": n, "gap_pt": gap, "step_width_pt": width, "semantic_states": []string{"planned", "active", "complete"}}, Limitations: []string{"Horizontal sequence of 2–6 stage panels; content remains subject to native text fit.", "New editable composition inspired by the process family; does not restructure original source OOXML."}}
	seen := map[string]bool{}
	for i, step := range in.Steps {
		if strings.TrimSpace(step.ID) == "" || seen[step.ID] || strings.Contains(step.ID, "/") {
			return out, fmt.Errorf("steps require unique nonempty IDs without slashes")
		}
		seen[step.ID] = true
		if strings.TrimSpace(step.Label) == "" || strings.TrimSpace(step.Summary) == "" || len(step.Activities) < 1 || len(step.Activities) > 8 || len(step.Outputs) < 1 || len(step.Outputs) > 5 {
			return out, fmt.Errorf("step %s needs label, summary, 1..8 activities and 1..5 outputs", step.ID)
		}
		fill := c.Style.MutedSurface
		switch step.State {
		case "", "planned":
		case "active":
			fill = c.Style.Active
		case "complete":
			fill = c.Style.Navy
		default:
			return out, fmt.Errorf("step %s unknown state %q", step.ID, step.State)
		}
		id := "process/" + step.ID
		mk := func(id, text string, bold bool, fg string) compose.BlockSpec {
			return compose.BlockSpec{ID: id, Text: text, FontFace: c.Style.FontFace, FontSizePt: c.Style.BodyFontPt, Bold: bold, Foreground: fg, Align: "left", Valign: "top", GapBeforePt: 4}
		}
		label := mk("heading", fmt.Sprintf("%02d  %s", i+1, step.Label), true, c.Style.Ink(fill))
		label.FontSizePt = c.Style.LabelFontPt
		label.GapBeforePt = 0
		summary := mk("summary", step.Summary, false, c.Style.Ink(c.Style.Surface))
		summary.GapBeforePt = 0
		acts := []compose.BlockSpec{mk("activities-heading", "ACTIVITIES", true, c.Style.Ink(c.Style.Surface))}
		acts[0].GapBeforePt = 0
		for j, text := range step.Activities {
			if strings.TrimSpace(text) == "" {
				return out, fmt.Errorf("step %s has empty activity", step.ID)
			}
			b := mk(fmt.Sprintf("activity-%d", j+1), text, false, c.Style.Ink(c.Style.Surface))
			b.Marker = "•"
			b.MarkerWidthPt = 9
			acts = append(acts, b)
		}
		outputs := []compose.BlockSpec{mk("outputs-heading", "EVIDENCE / OUTPUT", true, c.Style.Ink(c.Style.MutedSurface))}
		outputs[0].GapBeforePt = 0
		for j, text := range step.Outputs {
			if strings.TrimSpace(text) == "" {
				return out, fmt.Errorf("step %s has empty output", step.ID)
			}
			outputs = append(outputs, mk(fmt.Sprintf("output-%d", j+1), text, false, c.Style.Ink(c.Style.MutedSurface)))
		}
		panel := compose.ContainerSpec{ID: id, Bounds: compose.Rect{X: c.Bounds.X + float64(i)*(width+gap), Y: c.Bounds.Y, Width: width, Height: c.Bounds.Height}, Columns: []compose.TrackSpec{{Weight: 1}}, Rows: []compose.TrackSpec{{FixedPt: 42}, {MinPt: 48}, {MinPt: 80}, {MinPt: 55}}, Surface: c.Style.Surface, RowRule: &compose.RuleSpec{Color: c.Style.Border, WidthPt: .6}, Cells: []compose.CellSpec{
			{ID: "heading", Row: 0, Column: 0, Padding: compose.Insets{Top: 8, Left: 8, Right: 8, Bottom: 8}, Background: fill, Blocks: []compose.BlockSpec{label}},
			{ID: "summary", Row: 1, Column: 0, Padding: compose.Insets{Top: 9, Left: 8, Right: 8, Bottom: 8}, Blocks: []compose.BlockSpec{summary}},
			{ID: "activities", Row: 2, Column: 0, Padding: compose.Insets{Top: 8, Left: 8, Right: 8, Bottom: 8}, Blocks: acts},
			{ID: "outputs", Row: 3, Column: 0, Padding: compose.Insets{Top: 8, Left: 8, Right: 8, Bottom: 8}, Background: c.Style.MutedSurface, Blocks: outputs},
		}}
		out.Layouts = append(out.Layouts, panel)
		if i > 0 {
			out.Canvas = append(out.Canvas, compose.CanvasSpec{ID: fmt.Sprintf("process/arrow-%d", i), Kind: "shape", Preset: "rightArrow", Bounds: compose.Rect{X: panel.Bounds.X - gap + 3, Y: c.Bounds.Y + 15, Width: gap - 6, Height: 12}, Background: c.Style.Navy})
		}
	}
	// One measured grid shares row heights across every phase. This keeps
	// activities and evidence bands aligned when summaries wrap differently.
	shared := out.Layouts[0]
	shared.ID = "process"
	shared.Bounds = c.Bounds
	shared.Surface = ""
	shared.Columns = make([]compose.TrackSpec, n)
	shared.ColumnGapPt = gap
	shared.Cells = nil
	for i, panel := range out.Layouts {
		shared.Columns[i] = compose.TrackSpec{Weight: 1}
		for _, cell := range panel.Cells {
			cell.ID = in.Steps[i].ID + "-" + cell.ID
			cell.Column = i
			if cell.Background == "" {
				cell.Background = c.Style.Surface
			}
			shared.Cells = append(shared.Cells, cell)
		}
	}
	out.Layouts = []compose.ContainerSpec{shared}
	return out, nil
}
