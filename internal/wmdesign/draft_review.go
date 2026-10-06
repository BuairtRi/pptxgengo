package wmdesign

import (
	"fmt"
	"math"
	"strings"
	"unicode"

	"github.com/buairtri/pptxgengo/pptx"
)

const DraftReviewContract = "urn:pptxgengo:draft-review:v1"

// DraftReviewNote is slide-owned collaboration data, never template content.
// Dates are display strings so authors can use either ISO dates or short labels.
type DraftReviewNote struct {
	Status      string `json:"status"`
	StatusText  string `json:"status_text,omitempty"`
	StatusColor string `json:"status_color,omitempty"`
	Owner       string `json:"owner,omitempty"`
	Due         string `json:"due,omitempty"`
	Updated     string `json:"updated,omitempty"`
	Notes       string `json:"notes,omitempty"`
	Placement   string `json:"placement,omitempty"`
}

// DraftReviewRecord is deliberately separate from presentation content and its
// fit/collision measurements. Only the status tab intersects the slide canvas.
type DraftReviewRecord struct {
	Note       DraftReviewNote `json:"note"`
	Group      ComponentRecord `json:"group"`
	VisibleTab Rect            `json:"visible_tab"`
	Texts      []TextRecord    `json:"texts"`
	Shapes     []ShapeRecord   `json:"shapes"`
}

func ValidateDraftReviewNote(note *DraftReviewNote) error {
	if note == nil {
		return nil
	}
	switch note.Status {
	case "notstarted", "wip", "complete", "qa":
	default:
		return fmt.Errorf("draft_review.invalid_status: use notstarted, wip, complete or qa")
	}
	if note.Placement != "" && note.Placement != "edge" && note.Placement != "pasteboard" {
		return fmt.Errorf("draft_review.invalid_placement: use edge or pasteboard")
	}
	switch note.StatusColor {
	case "", "kpi.off", "kpi.risk", "kpi.on", "brand.blue":
	default:
		return fmt.Errorf("draft_review.invalid_status_color: use kpi.off, kpi.risk, kpi.on or brand.blue")
	}
	for _, field := range []struct {
		name, text string
		limit      int
	}{
		{"status_text", note.StatusText, 128}, {"owner", note.Owner, 256}, {"due", note.Due, 128}, {"updated", note.Updated, 128}, {"notes", note.Notes, 8192},
	} {
		if len(field.text) > field.limit {
			return fmt.Errorf("draft_review.field_too_long: %s", field.name)
		}
		for _, char := range field.text {
			if unicode.IsControl(char) && !(field.name == "notes" && (char == '\n' || char == '\r')) {
				return fmt.Errorf("draft_review.invalid_character: %s", field.name)
			}
		}
	}
	return nil
}

func (r *renderer) planDraftReview(id string, note DraftReviewNote) (*scenePlan, error) {
	if err := ValidateDraftReviewNote(&note); err != nil {
		return nil, err
	}
	x := 942.0
	if note.Placement == "pasteboard" {
		x = 960
	}
	p := &scenePlan{ID: "wm-review/" + id, Bounds: Rect{x, 0, 234, 144}}
	grounded, err := r.sceneColor("light", "brand.grounded")
	if err != nil {
		return nil, err
	}
	statuses := []struct{ key, label, ref string }{
		{"notstarted", "Not started", "kpi.off"}, {"wip", "In progress", "kpi.risk"},
		{"complete", "Complete", "kpi.on"}, {"qa", "QA'd", "brand.blue"},
	}
	colors := make([]string, len(statuses))
	current := ""
	statusText := ""
	for i, status := range statuses {
		colors[i], err = r.sceneColor("light", status.ref)
		if err != nil {
			return nil, err
		}
		if status.key == note.Status {
			current = colors[i]
			statusText = status.label
		}
	}
	if note.StatusText != "" {
		statusText = note.StatusText
	}
	if note.StatusColor != "" {
		current, err = r.sceneColor("light", note.StatusColor)
		if err != nil {
			return nil, err
		}
	}
	shape := func(suffix string, box Rect, fill, line string, width float64) error {
		return r.diagramShape(p, p.ID+suffix, box, pptx.ShapeTypeRect, fill, line, width, "solid", nil)
	}
	if err := shape(".body", Rect{x + 18.5, .5, 215, 143}, "FFFFFF", grounded, 1); err != nil {
		return nil, err
	}
	if err := shape(".tab", Rect{x, 0, 18, 144}, current, "", 0); err != nil {
		return nil, err
	}
	if err := shape(".header", Rect{x + 18, 0, 216, 27}, grounded, "", 0); err != nil {
		return nil, err
	}
	small, err := r.sceneStyle("small")
	if err != nil {
		return nil, err
	}
	label, err := r.sceneStyle("label")
	if err != nil {
		return nil, err
	}
	if !densityRoleCorrections(r.source) {
		label.Size, label.Leading, label.TrackingPt = 7, 10, .28
	}
	// Literal collaboration copy does not interpret presentation rich-text marks.
	text := func(suffix, copy string, style Style, box Rect, color, align string, middle, single bool) error {
		if copy == "" {
			return nil
		}
		layout, err := r.measureText(copy, style, box.W)
		if err != nil {
			return fmt.Errorf("draft_review.%s: %w", strings.TrimPrefix(suffix, "."), err)
		}
		need := math.Max(layout.AllocationHeight, layout.OccupiedTop+layout.EstimatedOccupiedHeight)
		if need > box.H+.02 || single && len(layout.Lines) != 1 {
			return fmt.Errorf("draft_review.text_overflow: %s needs %.2fpt; capacity %.2fpt. Shorten this field; the review note stays 234 x 144pt", strings.TrimPrefix(suffix, "."), need, box.H)
		}
		for _, line := range layout.Lines {
			if line.Advance > box.W+.02 {
				return fmt.Errorf("draft_review.text_overflow: %s exceeds %.2fpt width", strings.TrimPrefix(suffix, "."), box.W)
			}
		}
		tr := TextRecord{ID: p.ID + suffix, Rect: box, Color: color, Align: align, Layout: layout}
		if middle {
			tr.VerticalAlign = "middle"
		}
		p.Items = append(p.Items, sceneItem{Text: &tr})
		r.auditLiteralContrast(p, tr.ID, style, box, color)
		return nil
	}
	ink := func(background string) string {
		if contrast("FFFFFF", background) >= 4.5 {
			return "FFFFFF"
		}
		return grounded
	}
	statusStyle := label
	if densityRoleCorrections(r.source) {
		statusStyle.Size, statusStyle.Leading, statusStyle.TrackingPt = 8, 10, .8
	}
	if err := text(".status", statusText, statusStyle, Rect{x - 63, 63, 144, 18}, ink(current), "center", true, true); err != nil {
		return nil, err
	}
	p.Items[len(p.Items)-1].Text.Rotation = 270
	owner := small
	owner.Weight = 600
	if !densityRoleCorrections(r.source) {
		owner.Size, owner.Leading = 12, 18
	}
	ownerWidth := 198.0
	if note.Due != "" {
		due := "Due " + note.Due
		layout, err := r.measureText(due, label, 90)
		if err != nil || len(layout.Lines) != 1 {
			return nil, fmt.Errorf("draft_review.text_overflow: due must fit one line in 90pt")
		}
		width := layout.Lines[0].Advance + 2
		ownerWidth -= width + 9
		if err := text(".due", due, label, Rect{x + 225 - width, 0, width, 27}, "CED7E6", "right", true, true); err != nil {
			return nil, err
		}
	}
	if err := text(".owner", note.Owner, owner, Rect{x + 27, 0, ownerWidth, 27}, "FFFFFF", "left", true, true); err != nil {
		return nil, err
	}
	chip := label
	chip.Size, chip.Leading, chip.TrackingPt = 6.5, 9, 0
	for i, status := range statuses {
		box := Rect{x + 24 + float64(i)*51.5, 32, 49.5, 13}
		if err := shape(fmt.Sprintf(".legend-%d.fill", i), box, colors[i], "", 0); err != nil {
			return nil, err
		}
		if err := text(fmt.Sprintf(".legend-%d.text", i), status.label, chip, box, ink(colors[i]), "center", true, true); err != nil {
			return nil, err
		}
	}
	if err := text(".notes", note.Notes, small, Rect{x + 27, 51, 198, 72}, grounded, "left", false, false); err != nil {
		return nil, err
	}
	if note.Updated != "" {
		if err := text(".updated", "Updated "+note.Updated, label, Rect{x + 27, 126, 198, 12}, "50658E", "left", true, true); err != nil {
			return nil, err
		}
	}
	diagramFinish(p, "collab.status")
	p.Groups[len(p.Groups)-1].Contract = DraftReviewContract
	return p, nil
}

func (r *renderer) drawDraftReview(id string, note DraftReviewNote, sr *SlideReport) error {
	p, err := r.planDraftReview(id, note)
	if err != nil {
		return err
	}
	// A separate temporary report keeps draft copy out of visible slide content.
	private := SlideReport{}
	original := r.records
	r.records = &private.Texts
	err = r.drawScene(p, &private, "components/v0/components.json#collab.status")
	r.records = original
	if err != nil {
		return err
	}
	if note.Placement == "" {
		note.Placement = "edge"
	}
	tab := Rect{942, 0, 18, 144}
	if note.Placement == "pasteboard" {
		tab = Rect{}
	}
	sr.DraftReview = &DraftReviewRecord{Note: note, Group: p.Groups[len(p.Groups)-1], VisibleTab: tab, Texts: private.Texts, Shapes: private.Shapes}
	return nil
}
