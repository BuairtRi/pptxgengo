package wmdesign

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/buairtri/pptxgengo/pptx"
)

func sceneTableHasReference(raw json.RawMessage) bool {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil {
		return false
	}
	_, ref := fields["ref"]
	_, active := fields["refActive"]
	return ref || active
}

// Badges are editable text plus a native rectangle, never a rasterized label.
// Their fixed authored coordinates are covered by the table adornment warning.
func (r *renderer) sceneTableReferenceCell(p *scenePlan, id string, raw json.RawMessage, st Style, b Rect, surface string, ctx SceneContext) (pptx.TableCell, TextRecord, error) {
	if !isV6OrLaterLibrary(r.source.Revision) {
		return pptx.TableCell{}, TextRecord{}, fmt.Errorf("scene.table_reference_requires_v6")
	}
	var n struct {
		Text   *string `json:"text"`
		Sub    string  `json:"sub,omitempty"`
		Ref    *string `json:"ref,omitempty"`
		Active bool    `json:"refActive,omitempty"`
	}
	if err := sceneDecode(raw, &n); err != nil {
		return pptx.TableCell{}, TextRecord{}, err
	}
	if n.Text == nil || n.Ref == nil || strings.TrimSpace(*n.Ref) == "" {
		return pptx.TableCell{}, TextRecord{}, fmt.Errorf("scene.table_reference_requires_text_and_ref")
	}
	badgeStyle, err := r.sceneStyle("label")
	if err != nil {
		return pptx.TableCell{}, TextRecord{}, err
	}
	badgeStyle.Family = "IBM Plex Mono"
	badgeStyle.Size = 9
	badgeStyle.Leading = 13
	badgeStyle.Weight = 600
	badgeStyle.Tracking = "0.04em"
	badgeStyle.TrackingPt = .36
	badgeStyle.Case = ""
	badge, err := r.typeEngine.Measure(*n.Ref, badgeStyle, b.W-24)
	if err != nil {
		return pptx.TableCell{}, TextRecord{}, err
	}
	if len(badge.Lines) != 1 {
		return pptx.TableCell{}, TextRecord{}, fmt.Errorf("scene.table_reference_badge_wrap")
	}
	// An exact shaped advance wraps in native PowerPoint even when the Go
	// measurement is one line. Reserve the existing inline native allowance
	// inside the badge, preserving its authored padding, font and copy.
	badgeW := sequenceInlineWidth(badge.Lines[0].Advance, badgeStyle) + 6
	if badgeW+6 >= b.W-24 {
		return pptx.TableCell{}, TextRecord{}, fmt.Errorf("scene.table_reference_width")
	}
	// The source object cell gives the first paragraph a 1.2 line height.
	// Reserve the inline badge beside that paragraph; subtitle uses the full
	// cell width as an independent, editable source paragraph below it.
	st.Leading = st.Size * 1.2
	labelBox := b
	labelBox.X += badgeW + 6
	labelBox.W -= badgeW + 6
	cell, tr, err := r.sceneNativeCell(id, *n.Text, st, labelBox, surface, "primary", "left", 12, 12, ctx)
	if err != nil {
		return cell, tr, err
	}
	cell.Options.Margin[3] += (badgeW + 6) / 72
	need := math.Max(tr.Layout.AllocationHeight, tr.Layout.OccupiedTop+tr.Layout.EstimatedOccupiedHeight)
	bodyNeed := need
	var subStyle Style
	var subNeed float64
	if n.Sub != "" {
		subStyle, err = r.sceneStyle("small")
		if err != nil {
			return cell, tr, err
		}
		subStyle.Weight = 400
		subStyle.Leading = subStyle.Size * 1.2
		subLayout, e := r.typeEngine.Measure(n.Sub, subStyle, b.W-24)
		if e != nil {
			return cell, tr, e
		}
		subNeed = math.Max(subLayout.AllocationHeight, subLayout.OccupiedTop+subLayout.EstimatedOccupiedHeight)
		need += 2 + subNeed
	}
	badgeNeed := math.Max(badge.AllocationHeight, badge.OccupiedTop+badge.EstimatedOccupiedHeight)
	if badgeNeed > b.H+.02 || need > b.H+.02 {
		return cell, tr, fmt.Errorf("scene.table_reference_height: badge%.3fpt text%.3fpt capacity%.3fpt", badgeNeed, need, b.H)
	}
	// Native paragraph copy remains vertically centered, matching the source.
	badgeY := b.Y + (b.H-need)/2
	if n.Sub != "" {
		cell.Options.Valign = "top"
		cell.Options.Margin[0] = (badgeY - b.Y) / 72
		cell.Options.Margin[2] = (b.H - (badgeY - b.Y) - bodyNeed) / 72
		if _, err = r.sceneDataText(p, id+".sub", n.Sub, subStyle, Rect{b.X + 12, badgeY + bodyNeed + 2, b.W - 24, subNeed}, surface, "secondary", "left", ctx); err != nil {
			return cell, tr, err
		}
	}
	box := Rect{b.X + 12, badgeY, badgeW, 13}
	strong, err := r.sceneColor(surface, "strong")
	if err != nil {
		return cell, tr, err
	}
	primary, err := r.sceneColor(surface, "primary")
	if err != nil {
		return cell, tr, err
	}
	fill, ink := primary, primary
	if n.Active {
		fill = "070154"
		ink = "FFFFFF"
	}
	r.sceneDataShape(p, id+".ref.fill", box, pptx.ShapeTypeRect, fill, &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: strong}, Width: .75})
	if !n.Active {
		p.Items[len(p.Items)-1].Shape.Props.Fill = &pptx.ShapeFillProps{Type: "none"}
	}
	// Preserve the 13pt source badge while reserving the font engine's measured
	// editable text height (including its native vertical safety allowance).
	if err = r.sequenceLiteralText(p, id+".ref", *n.Ref, badgeStyle, Rect{box.X + 3, box.Y + (box.H-badgeNeed)/2, box.W - 6, badgeNeed}, ink, "center", false); err != nil {
		return cell, tr, err
	}
	return cell, tr, nil
}

func (r *renderer) sceneTablePriorityCell(p *scenePlan, id string, raw json.RawMessage, st Style, b Rect, surface string, ctx SceneContext) (pptx.TableCell, TextRecord, error) {
	if !isV6OrLaterLibrary(r.source.Revision) {
		return pptx.TableCell{}, TextRecord{}, fmt.Errorf("scene.table_priority_requires_v6")
	}
	chip, err := sceneTablePriority(raw)
	if err != nil {
		return pptx.TableCell{}, TextRecord{}, err
	}
	chipStyle, err := r.sceneStyle("label")
	if err != nil {
		return pptx.TableCell{}, TextRecord{}, err
	}
	chipStyle.Weight = 600
	label, err := r.typeEngine.Measure(chip.Label, chipStyle, b.W-24-2*chip.PaddingX)
	if err != nil {
		return pptx.TableCell{}, TextRecord{}, err
	}
	if len(label.Lines) != 1 {
		return pptx.TableCell{}, TextRecord{}, fmt.Errorf("scene.table_priority_wrap")
	}
	// Padding must surround a native-safe text allocation rather than the raw
	// advance; labels such as LATER otherwise break onto a second native line.
	width, err := chip.WidthForLabel(sequenceInlineWidth(label.Lines[0].Advance, chipStyle))
	if err != nil {
		return pptx.TableCell{}, TextRecord{}, err
	}
	height := math.Max(label.AllocationHeight, label.OccupiedTop+label.EstimatedOccupiedHeight) + 2*chip.PaddingY
	if width > b.W-24+.02 || height > b.H+.02 {
		return pptx.TableCell{}, TextRecord{}, fmt.Errorf("scene.table_priority_overflow")
	}
	cell, tr, err := r.sceneNativeCell(id, "", st, b, surface, "primary", "center", 12, 12, ctx)
	if err != nil {
		return cell, tr, err
	}
	box := Rect{b.X + (b.W-width)/2, b.Y + (b.H-height)/2, width, height}
	bg, fg, border := strings.TrimPrefix(chip.Background, "#"), strings.TrimPrefix(chip.Foreground, "#"), strings.TrimPrefix(chip.Border, "#")
	r.sceneDataShape(p, id+".priority.fill", box, pptx.ShapeTypeRect, bg, &pptx.ShapeLineProps{ShapeFillProps: pptx.ShapeFillProps{Color: border}, Width: chip.BorderPt})
	if chip.Background == "transparent" {
		p.Items[len(p.Items)-1].Shape.Props.Fill = &pptx.ShapeFillProps{Type: "none"}
		p.Items[len(p.Items)-1].Shape.Record.Color = ""
	}
	textBox := Rect{box.X + chip.PaddingX, box.Y + chip.PaddingY, box.W - 2*chip.PaddingX, box.H - 2*chip.PaddingY}
	if err = r.sequenceLiteralText(p, id+".priority", chip.Label, chipStyle, textBox, fg, "center", false); err != nil {
		return cell, tr, err
	}
	return cell, tr, nil
}
