package wmdesign

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
)

func adornmentText(t *testing.T, p *scenePlan, suffix string) *TextRecord {
	t.Helper()
	for _, item := range p.Items {
		if item.Text != nil && strings.HasSuffix(item.Text.ID, suffix) {
			return item.Text
		}
	}
	t.Fatal("editable adornment text missing", suffix)
	return nil
}

func TestTableV6ReferenceBadgesReserveNativeInlineWidth(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source.Revision = LibraryRevisionV6
	st, err := r.sceneStyle("body")
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 9; i++ {
		for _, active := range []bool{false, true} {
			t.Run(fmt.Sprintf("A%d/active=%t", i, active), func(t *testing.T) {
				ref := fmt.Sprintf("A%d", i)
				raw := json.RawMessage(fmt.Sprintf(`{"text":"Capability","ref":%q,"refActive":%t}`, ref, active))
				before := string(raw)
				p := &scenePlan{}
				b := Rect{57, 156, 198, 30}
				cell, record, err := r.sceneTableReferenceCell(p, "name", raw, st, b, "light", SceneContext{Surface: "light"})
				if err != nil {
					t.Fatal(err)
				}
				text := adornmentText(t, p, "name.ref")
				shape := enhancementShape(t, p, "name.ref.fill")
				style := text.Layout.Style
				advance := text.Layout.Lines[0].Advance
				want := sequenceInlineWidth(advance, style)
				if len(text.Layout.Lines) != 1 || text.Layout.Original != ref || text.Layout.Displayed != ref || text.Rect.W < want-.001 || text.Rect.W <= advance {
					t.Fatalf("native guard/copy missing: %+v advance%.3f guarded%.3f", text.Rect, advance, want)
				}
				if style.Family != "IBM Plex Mono" || style.Size != 9 || style.Weight != 600 || style.TrackingPt != .36 || shape.Record.Rect.H != 13 {
					t.Fatal("font or source badge height changed")
				}
				if math.Abs(shape.Record.Rect.W-(text.Rect.W+6)) > .001 || math.Abs((text.Rect.X+text.Rect.W/2)-(shape.Record.Rect.X+shape.Record.Rect.W/2)) > .001 {
					t.Fatal("padding or horizontal centering changed")
				}
				if math.Abs((text.Rect.Y+text.Rect.H/2)-(shape.Record.Rect.Y+shape.Record.Rect.H/2)) > .001 {
					t.Fatal("vertical centering changed")
				}
				if active && (text.Color != "FFFFFF" || shape.Record.Color != "070154") || !active && (text.Color != "070154" || shape.Props.Fill.Type != "none") {
					t.Fatal("active reference palette changed")
				}
				if cell.Text != "Capability" || record.Layout.Original != "Capability" || record.Rect.X < shape.Record.Rect.X+shape.Record.Rect.W+6-.001 {
					t.Fatal("body copy or badge/body clearance changed")
				}
				if string(raw) != before {
					t.Fatal("source JSON mutated")
				}
			})
		}
	}
}

func TestTableV6PriorityLabelsReserveNativeInlineWidth(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source.Revision = LibraryRevisionV6
	st, err := r.sceneStyle("body")
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"P1", "P2", "P3", "P4", "now", "next", "later", "Later", "LATER", "Hold"} {
		t.Run(value, func(t *testing.T) {
			raw, _ := json.Marshal(value)
			chip, err := sceneTablePriority(raw)
			if err != nil {
				t.Fatal(err)
			}
			p := &scenePlan{}
			b := Rect{400, 156, 72, 30}
			_, _, err = r.sceneTablePriorityCell(p, "priority", raw, st, b, "light", SceneContext{Surface: "light"})
			if err != nil {
				t.Fatal(err)
			}
			text := adornmentText(t, p, "priority.priority")
			shape := enhancementShape(t, p, "priority.priority.fill")
			advance := text.Layout.Lines[0].Advance
			guarded := sequenceInlineWidth(advance, text.Layout.Style)
			if len(text.Layout.Lines) != 1 || text.Layout.Original != chip.Label || text.Rect.W < guarded-.001 || text.Rect.W <= advance {
				t.Fatalf("native width/copy missing for %s: rect%+v advance%.3f guarded%.3f", value, text.Rect, advance, guarded)
			}
			if text.Layout.Style.Family != "IBM Plex Mono" || text.Layout.Style.Size != 9 || text.Layout.Style.Weight != 600 {
				t.Fatal("priority font changed")
			}
			if shape.Record.Rect.W < b.W-24 && math.Abs((shape.Record.Rect.X+shape.Record.Rect.W/2)-(b.X+b.W/2)) > .001 {
				t.Fatal("chip off center")
			}
			if math.Abs((text.Rect.X+text.Rect.W/2)-(shape.Record.Rect.X+shape.Record.Rect.W/2)) > .001 || text.Rect.X < shape.Record.Rect.X+chip.PaddingX-.001 || text.Rect.X+text.Rect.W > shape.Record.Rect.X+shape.Record.Rect.W-chip.PaddingX+.001 {
				t.Fatal("priority padding/centering changed")
			}
			if text.Color != strings.TrimPrefix(chip.Foreground, "#") || shape.Record.Rect.W > b.W-24+.001 || shape.Record.Rect.H > b.H+.001 {
				t.Fatal("palette or cell bounds changed")
			}
		})
	}
}

func TestTableV7InheritsReferenceAndPriorityNativeWidth(t *testing.T) {
	r := intakeTestRenderer(t)
	st, err := r.sceneStyle("body")
	if err != nil {
		t.Fatal(err)
	}
	for _, priority := range []bool{false, true} {
		var previous *scenePlan
		for _, revision := range []string{LibraryRevisionV6, LibraryRevisionV7} {
			r.source.Revision = revision
			p := &scenePlan{}
			if priority {
				_, _, err = r.sceneTablePriorityCell(p, "priority", json.RawMessage(`"later"`), st, Rect{57, 156, 72, 30}, "light", SceneContext{Surface: "light"})
			} else {
				_, _, err = r.sceneTableReferenceCell(p, "ref", json.RawMessage(`{"text":"Capability","ref":"A4","refActive":true}`), st, Rect{57, 156, 198, 30}, "light", SceneContext{Surface: "light"})
			}
			if err != nil {
				t.Fatal(revision, err)
			}
			if previous != nil && !reflect.DeepEqual(previous, p) {
				t.Fatal("later revision changed inherited badge/priority treatment")
			}
			previous = p
		}
	}
}
