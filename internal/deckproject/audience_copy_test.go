package deckproject

import (
	"reflect"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestAudienceVisibleCopyIncludesNativeTableCells(t *testing.T) {
	text := func(original, displayed string) wmdesign.TextRecord {
		return wmdesign.TextRecord{Layout: wmdesign.TextLayout{Original: original, Displayed: displayed}}
	}
	slide := wmdesign.SlideReport{
		Texts: []wmdesign.TextRecord{text("heading", "HEADING"), text("not displayed", "")},
		Tables: []wmdesign.SceneTableRecord{{Cells: []wmdesign.SceneCellRecord{
			{Row: 0, Column: 0, Text: text("owner", "OWNER")},
			{Row: 1, Column: 0, Text: text("Alex", "Alex")},
		}}},
	}
	if got := reviewVisibleCopy(slide); !reflect.DeepEqual(got, []string{"HEADING", "OWNER", "Alex"}) {
		t.Fatalf("displayed shape/table copy mismatch: %#v", got)
	}
}
