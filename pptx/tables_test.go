package pptx

import "testing"

// Fixture derivation (hand-executed against src/gen-tables.ts algorithm):
//
// Layout: testLayout (16x9) = 12192000 x 6858000 EMU (see utils_test.go).
// Default slide margins: DEF_SLIDE_MARGIN_IN = [0.5, 0.5, 0.5, 0.5] (TRBL, inches).
// Table: no x/y/w/h given, no colW given, default fontSize (12), single column.
//
//  1. calcSlideTabH() for the FIRST slide (tableRowSlides empty):
//     emuStartY    = inch2Emu(0.5)                    = jsRound(914400*0.5)   = 457200
//     emuSlideTabH = presLayout.height - emuStartY - inch2Emu(0.5)
//     = 6858000 - 457200 - 457200 = 5943600 EMU (6.5 in)
//     Because tableProps.AutoPageSlideStartY/NewSlideStartY/tablePropY are all
//     unset (0), the "tableRowSlides.length>1" override branch never fires, so
//     EVERY slide (1st, 2nd, 3rd, ...) gets this same 5943600 EMU budget.
//
//  2. Column width (STEP 3-5): tablePropW==0 and colW unset => tableCalcW==0,
//     so emuSlideTabW falls back to:
//     inch2Emu((tablePropX==0 ? arrInchMargins[1] : ...) + arrInchMargins[3])
//     = inch2Emu(0.5 + 0.5) = inch2Emu(1.0) = 914400 EMU (1 in)
//     Distributed evenly over numCols=1: colW = [914400/914400/1] = [1.0] (inch).
//
//  3. Line height (single line per row, default fontSize=12, autoPageLineWeight=0):
//     lineHeight = inch2Emu((12 * (1.67 + 0)) / 100)
//     = inch2Emu(0.2004) = jsRound(914400 * 0.2004) = 183246 EMU
//
//  4. Each row's cell text ("Row N") is short (<=6 chars, 2 "words": "Row " and
//     "N"). CPL for colWidth=1.0in, fontSize=12, autoPageCharWeight=0:
//     CPL = floor((1.0/12700)*914400) / (12/2.3) = floor(72.0) / 5.217... = 13.8
//     "Row " (4 chars) + "N" (<=2 chars) never exceeds CPL=13.8, so
//     parseTextToLines always returns exactly 1 line per row => each row
//     consumes exactly one lineHeight (183246 EMU) of vertical space, with no
//     per-cell margins (none configured) contributing anything extra.
//
//  5. Rows-per-slide = floor(5943600 / 183246) = 32 (32*183246 = 5863872 <=
//     5943600; 33*183246 = 6047118 > 5943600 => row 33 overflows to a new slide).
//     For a 40-row table: slide 1 gets rows 1-32 (32 rows), slide 2 gets rows
//     33-40 (8 rows) => distribution [32, 8].
const (
	fixtureLineHeightEMU     = 183246
	fixtureRowsPerSlide      = 32
	fixtureSlideTabHEMU      = 5943600
	fixtureColWDefaultInches = 1.0
)

// shortRow returns a single-column TableRow with plain text "Row N" (no
// wrapping at the fixture's CPL of 13.8 chars).
func shortRow(n string) TableRow {
	return TableRow{{Type: SlideObjectTypeTablecell, Text: "Row " + n}}
}

func manyShortRows(count int) []TableRow {
	digits := "0123456789"
	rows := make([]TableRow, count)
	for i := 0; i < count; i++ {
		n := i + 1
		s := ""
		for n > 0 {
			s = string(digits[n%10]) + s
			n /= 10
		}
		if s == "" {
			s = "0"
		}
		rows[i] = shortRow(s)
	}
	return rows
}

func totalRows(slides []TableRowSlide) int {
	n := 0
	for _, s := range slides {
		n += len(s.Rows)
	}
	return n
}

func TestGetSlidesForTableRows_SingleShortTable(t *testing.T) {
	rows := manyShortRows(3)
	slides := GetSlidesForTableRows(rows, &TableToSlidesProps{}, testLayout, nil)

	if len(slides) != 1 {
		t.Fatalf("expected 1 slide, got %d", len(slides))
	}
	if len(slides[0].Rows) != 3 {
		t.Fatalf("expected 3 rows on the single slide, got %d", len(slides[0].Rows))
	}
}

func TestGetSlidesForTableRows_TallTable_HandDerivedSplit(t *testing.T) {
	rows := manyShortRows(40)
	slides := GetSlidesForTableRows(rows, &TableToSlidesProps{}, testLayout, nil)

	if len(slides) != 2 {
		t.Fatalf("expected 2 slides, got %d", len(slides))
	}
	if got := len(slides[0].Rows); got != 32 {
		t.Errorf("slide 1: expected 32 rows, got %d", got)
	}
	if got := len(slides[1].Rows); got != 8 {
		t.Errorf("slide 2: expected 8 rows, got %d", got)
	}
	if got := totalRows(slides); got != 40 {
		t.Errorf("expected 40 total rows preserved across slides, got %d", got)
	}

	// Spot-check row content/order is preserved: first row of slide 1 is
	// "Row 1", last row of slide 2 is "Row 40".
	first := slides[0].Rows[0][0]
	if len(first.TextCells) == 0 {
		t.Fatal("expected first row's cell to have TextCells (word fragments)")
	}
	last := slides[1].Rows[len(slides[1].Rows)-1][0]
	if len(last.TextCells) == 0 {
		t.Fatal("expected last row's cell to have TextCells (word fragments)")
	}
}

func TestGetSlidesForTableRows_ColW_NoneGiven_EvenDistribution(t *testing.T) {
	// 2 columns, no colW: emuSlideTabW falls back to inch2Emu(0.5+0.5)=914400
	// (see fixture derivation #2), distributed evenly => [0.5, 0.5] inches.
	rows := []TableRow{
		{
			{Type: SlideObjectTypeTablecell, Text: "A"},
			{Type: SlideObjectTypeTablecell, Text: "B"},
		},
	}
	props := &TableToSlidesProps{}
	GetSlidesForTableRows(rows, props, testLayout, nil)

	if len(props.ColW) != 2 {
		t.Fatalf("expected 2 colW entries, got %d (%v)", len(props.ColW), props.ColW)
	}
	for i, w := range props.ColW {
		if w != 0.5 {
			t.Errorf("colW[%d] = %v, want 0.5", i, w)
		}
	}
}

func TestGetSlidesForTableRows_ColW_ExplicitArray(t *testing.T) {
	rows := []TableRow{
		{
			{Type: SlideObjectTypeTablecell, Text: "A"},
			{Type: SlideObjectTypeTablecell, Text: "B"},
		},
	}
	props := &TableToSlidesProps{TableProps: TableProps{ColW: []float64{3.0, 4.0}}}
	GetSlidesForTableRows(rows, props, testLayout, nil)

	if len(props.ColW) != 2 || props.ColW[0] != 3.0 || props.ColW[1] != 4.0 {
		t.Fatalf("explicit colW should be used as-is unchanged, got %v", props.ColW)
	}
}

func TestGetSlidesForTableRows_ColW_MismatchedCount_NoPanic(t *testing.T) {
	// 2 columns but only 1 colW entry provided: colWAt(colW, 1) safely
	// returns 0 for the missing column instead of panicking (see colWAt's
	// doc comment for the TS-vs-Go fidelity tradeoff here). CPL==0 for that
	// column forces every word of its text onto its own line, which we can
	// observe as a longer row.
	rows := []TableRow{
		{
			{Type: SlideObjectTypeTablecell, Text: "A"},
			{Type: SlideObjectTypeTablecell, Text: "one two three four"},
		},
	}
	props := &TableToSlidesProps{TableProps: TableProps{ColW: []float64{3.0}}}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("GetSlidesForTableRows panicked on mismatched colW: %v", r)
		}
	}()
	slides := GetSlidesForTableRows(rows, props, testLayout, nil)
	if len(slides) == 0 || len(slides[0].Rows) == 0 {
		t.Fatal("expected at least one row on one slide")
	}
}

func TestGetSlidesForTableRows_RepeatHeadRow(t *testing.T) {
	// 40 body rows normally split [32, 8] (see fixture derivation). With
	// AutoPageRepeatHeader + ArrObjTabHeadRows, every slide after the first
	// gets the header row injected first, consuming one lineHeight of budget
	// too: [32, 1(header)+8].
	headerRow := TableRow{
		{Type: SlideObjectTypeTablecell, Text: "Header", LineHeight: fixtureLineHeightEMU},
	}
	rows := manyShortRows(40)
	props := &TableToSlidesProps{
		TableProps: TableProps{
			ArrObjTabHeadRows: []TableRow{headerRow},
		},
		AutoPageRepeatHeader: ptr(true),
	}
	slides := GetSlidesForTableRows(rows, props, testLayout, nil)

	if len(slides) != 2 {
		t.Fatalf("expected 2 slides, got %d", len(slides))
	}
	if got := len(slides[0].Rows); got != 32 {
		t.Errorf("slide 1: expected 32 rows (no header), got %d", got)
	}
	if got := len(slides[1].Rows); got != 9 {
		t.Errorf("slide 2: expected 9 rows (1 header + 8 body), got %d", got)
	}
	// The first row of slide 2 should be the repeated header.
	if len(slides[1].Rows) > 0 {
		headerCell := slides[1].Rows[0][0]
		if headerCell.Text != "Header" {
			t.Errorf("slide 2 row 0 should be the repeated header, got Text=%q", headerCell.Text)
		}
	}
	// Slide 1 must NOT contain the header (it's only injected on overflow).
	for _, r := range slides[0].Rows {
		if r[0].Text == "Header" {
			t.Fatal("slide 1 should not contain the repeated header row")
		}
	}
}

func TestGetSlidesForTableRows_NewSlideStartY_HonoredOnPage2Plus(t *testing.T) {
	// Baseline (no override): every page after page 1 reuses the default
	// margin-based budget (5943600 EMU => 32 rows/page). With
	// AutoPageSlideStartY=1.0in, page 2+ budget shrinks to:
	//   6858000 - inch2Emu(1.0) - inch2Emu(0.5) = 6858000-914400-457200 = 5486400
	//   => floor(5486400/183246) = 29 rows/page.
	// 72 total rows (32 fits on page 1 either way):
	//   baseline:  [32, 32, 8]   (page2 uses default 32/page budget)
	//   override:  [32, 29, 11]  (page2+ uses the shrunk budget)
	rows := manyShortRows(72)

	baseline := GetSlidesForTableRows(rows, &TableToSlidesProps{}, testLayout, nil)
	if len(baseline) != 3 {
		t.Fatalf("baseline: expected 3 slides, got %d", len(baseline))
	}
	wantBaseline := []int{32, 32, 8}
	for i, s := range baseline {
		if got := len(s.Rows); got != wantBaseline[i] {
			t.Errorf("baseline slide %d: got %d rows, want %d", i, got, wantBaseline[i])
		}
	}

	override := GetSlidesForTableRows(rows, &TableToSlidesProps{
		TableProps: TableProps{AutoPageSlideStartY: ptr(1.0)},
	}, testLayout, nil)
	if len(override) != 3 {
		t.Fatalf("override: expected 3 slides, got %d", len(override))
	}
	wantOverride := []int{32, 29, 11}
	for i, s := range override {
		if got := len(s.Rows); got != wantOverride[i] {
			t.Errorf("override slide %d: got %d rows, want %d", i, got, wantOverride[i])
		}
	}
}

func TestGetSlidesForTableRows_AutoPageCharWeight_AdjustsCharsPerLine(t *testing.T) {
	// Direct unit test of parseTextToLines: colWidth=1.0in, fontSize=12.
	//   CPL(weight=0)   = floor(72)/(12/2.3) = 72/5.217... = 13.8
	//   CPL(weight=0.5) = floor(72)/(12/2.8) = 72/4.2857... = 16.8
	// Text "Hello World Test" tokenizes to "Hello "(6) + "World "(6) + "Test"(4):
	//   weight=0:   6, 6+6=12<=13.8, 12+4=16>13.8   => 2 lines: ["Hello ","World "],["Test"]
	//   weight=0.5: 6, 12<=16.8, 12+4=16<=16.8       => 1 line:  ["Hello ","World ","Test"]
	cellDefault := TableCell{Text: "Hello World Test", Options: &TableCellProps{}}
	linesDefault := parseTextToLines(cellDefault, 1.0)
	if len(linesDefault) != 2 {
		t.Fatalf("default autoPageCharWeight: expected 2 lines, got %d (%v)", len(linesDefault), linesDefault)
	}

	cellWide := TableCell{Text: "Hello World Test", Options: &TableCellProps{AutoPageCharWeight: 0.5}}
	linesWide := parseTextToLines(cellWide, 1.0)
	if len(linesWide) != 1 {
		t.Fatalf("autoPageCharWeight=0.5: expected 1 line, got %d (%v)", len(linesWide), linesWide)
	}
}

func TestParseTextToLines_MultiLineEmbeddedNewline(t *testing.T) {
	// Each "\n"-delimited segment always starts its own physical line
	// (independent of CPL), so 3 embedded newlines => 3 lines minimum,
	// regardless of how short each segment is.
	cell := TableCell{Text: "Line1\nLine2\nLine3"}
	lines := parseTextToLines(cell, 1.0)
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines for embedded \\n text, got %d (%v)", len(lines), lines)
	}
	want := []string{"Line1", "Line2", "Line3"}
	for i, line := range lines {
		if len(line) != 1 {
			t.Fatalf("line %d: expected 1 word-cell, got %d", i, len(line))
		}
		if line[0].Text != want[i] {
			t.Errorf("line %d: got Text=%q, want %q", i, line[0].Text, want[i])
		}
		if line[0].Options == nil || line[0].Options.BreakLine == nil || !*line[0].Options.BreakLine {
			t.Errorf("line %d: expected BreakLine=true (injected by \\n split)", i)
		}
	}
}

func TestParseTextToLines_TextCellsRuns(t *testing.T) {
	// A cell whose text is a TableCell[] (mixed-formatting runs): each run is
	// its own fragment and therefore starts its own physical line for
	// pagination purposes (see the STEP 2+3 fidelity note in tables.go) —
	// this is a faithful reproduction of the upstream TS algorithm's
	// behavior, not a Go-specific choice.
	cell := TableCell{
		TextCells: []TableCell{
			{Text: "Bold", Options: &TableCellProps{TextBaseProps: TextBaseProps{Bold: ptr(true)}}},
			{Text: "Normal"},
		},
	}
	lines := parseTextToLines(cell, 1.0)
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines (one per text run), got %d (%v)", len(lines), lines)
	}
	if lines[0][0].Text != "Bold" {
		t.Errorf("line 0 word 0: got Text=%q, want %q", lines[0][0].Text, "Bold")
	}
	if lines[0][0].Options == nil || lines[0][0].Options.Bold == nil || !*lines[0][0].Options.Bold {
		t.Error("line 0 word 0: expected Bold option preserved")
	}
	if lines[1][0].Text != "Normal" {
		t.Errorf("line 1 word 0: got Text=%q, want %q", lines[1][0].Text, "Normal")
	}
}

func TestGetSlidesForTableRows_NilPropsDefensive(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("GetSlidesForTableRows(nil tableProps) panicked: %v", r)
		}
	}()
	rows := manyShortRows(2)
	slides := GetSlidesForTableRows(rows, nil, testLayout, nil)
	if len(slides) != 1 {
		t.Fatalf("expected 1 slide, got %d", len(slides))
	}
}
