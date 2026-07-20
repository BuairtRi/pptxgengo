// tables.go ports src/gen-tables.ts: the table auto-paging engine.
//
// SKIPPED: genTableToSlides() (src/gen-tables.ts ~L523-749) is not ported. It
// scrapes an HTML <table> element via `document.getElementById`,
// `window.getComputedStyle`, etc. — a browser-only DOM path with no Go
// equivalent, per PORTING.md "Skip browser-only paths". Everything it needs
// (column-width/margin/style extraction from a live DOM table) has no
// server-side analogue; a Go caller is expected to build []TableRow directly
// and call GetSlidesForTableRows itself.
//
// Ported:
//   - GetSlidesForTableRows (TS `getSlidesForTableRows`): the auto-paging
//     engine that splits table rows across slides based on estimated text
//     height.
//   - parseTextToLines: cell-text-to-lines heuristic (character-width based
//     line wrap estimate used only for pagination, not for final XML
//     rendering).
//
// Fidelity notes (see also the final task report for the full list):
//   - All `console.log` verbose-mode output is dropped (TableProps.Verbose is
//     accepted but has no effect); it never influenced computation in the TS
//     source either.
//   - `verbose` was a plain bool parameter to parseTextToLines with zero
//     effect on its return value (console.log only) — dropped entirely.
//   - Several odd-looking conditionals below (documented inline) are
//     faithful reproductions of quirks/bugs in the original TS, preserved
//     deliberately for byte-identical row-split behavior.
//   - DEVIATION: a zero-length input row is skipped in STEP 6, where upstream
//     TS crashes (it indexes row[0]/reduces over the row). This is a
//     deliberate safety choice, locked by
//     TestGetSlidesForTableRows_EmptyRowSkipped.
package pptx

import (
	"math"
	"strings"
)

// boolVal reports whether p is non-nil and true.
func boolVal(p *bool) bool {
	return p != nil && *p
}

// coordOrZero dereferences an optional Coord, returning the zero Coord
// (which getSmartParseNumber resolves to EMU 0) when c is nil — mirrors the
// TS `undefined` case for an optional x/y/w/h.
func coordOrZero(c *Coord) Coord {
	if c == nil {
		return Coord{}
	}
	return *c
}

// marginTRBLTop returns the "top" component of a Margin value (index 0 for
// both the uniform len==1 and TRBL len==4 forms); 0 when unset.
func marginTRBLTop(m Margin) float64 {
	if len(m) == 1 || len(m) == 4 {
		return m[0]
	}
	return 0
}

// marginTRBLBottom returns the "bottom" component of a Margin value (m[0]
// for the uniform len==1 form since all sides are equal, m[2] for TRBL);
// 0 when unset.
func marginTRBLBottom(m Margin) float64 {
	switch len(m) {
	case 1:
		return m[0]
	case 4:
		return m[2]
	default:
		return 0
	}
}

// expandMarginTRBL expands a Margin into a 4-element [top,right,bottom,left]
// array, returning fallback unchanged for nil/unrecognized lengths.
func expandMarginTRBL(m Margin, fallback [4]float64) [4]float64 {
	switch len(m) {
	case 1:
		v := m[0]
		return [4]float64{v, v, v, v}
	case 4:
		return [4]float64{m[0], m[1], m[2], m[3]}
	default:
		return fallback
	}
}

// colWAt safely indexes a column-width slice, returning 0 for an
// out-of-range index instead of panicking.
//
// TS DEVIATION: `tableProps.colW[iCell]` on a too-short JS array yields
// `undefined`, which propagates as NaN through parseTextToLines' CPL
// formula, making every char-overflow comparison `NaN > x` evaluate false —
// i.e. the cell's text is never wrapped (stays one line). Go has no
// `undefined`; returning 0 instead makes CPL == 0, which wraps every single
// word onto its own line (many lines) instead of zero lines. Both are
// degenerate/unintended fallbacks for a caller error (colW shorter than the
// column count) — this makes the mismatch a safe, testable non-panic instead
// of matching the exact JS NaN-propagation quirk.
func colWAt(colW []float64, idx int) float64 {
	if idx < 0 || idx >= len(colW) {
		return 0
	}
	return colW[idx]
}

// cloneCellProps returns a shallow copy of opts, or a fresh zero-value
// TableCellProps when opts is nil. Mirrors the TS `{ ...cell.options }`
// object spread, which always yields a defined (possibly empty) object even
// when `cell.options` is undefined.
func cloneCellProps(opts *TableCellProps) *TableCellProps {
	if opts == nil {
		return &TableCellProps{}
	}
	c := *opts
	return &c
}

// resolveFontSize resolves the effective font size for a table cell: the
// cell's own FontSize, else the table's FontSize, else DEF_FONT_SIZE.
func resolveFontSize(cellOpts *TableCellProps, tableProps *TableToSlidesProps) float64 {
	if cellOpts != nil && cellOpts.FontSize != 0 {
		return cellOpts.FontSize
	}
	if tableProps.FontSize != 0 {
		return tableProps.FontSize
	}
	return float64(DEF_FONT_SIZE)
}

// tokenizeWords splits text on spaces into word TableCells, each carrying a
// shallow clone of opts. When opts has BreakLine set, only the *last* word
// keeps BreakLine=true (mirrors the TS "only apply breakLine to the very
// last word" rule); words append a trailing space except the final word.
func tokenizeWords(text string, opts *TableCellProps) []TableCell {
	words := strings.Split(text, " ")
	out := make([]TableCell, len(words))
	for idx, w := range words {
		wordOpts := cloneCellProps(opts)
		isLast := idx+1 == len(words)
		if wordOpts.BreakLine != nil && *wordOpts.BreakLine {
			wordOpts.BreakLine = ptr(isLast)
		}
		txt := w
		if idx+1 < len(words) {
			txt += " "
		}
		out[idx] = TableCell{Type: SlideObjectTypeTablecell, Text: txt, Options: wordOpts}
	}
	return out
}

// parseTextToLines breaks a cell's text into lines based on the table
// column width (character-width heuristic used only for auto-paging height
// estimation — NOT the final line-wrap used when rendering XML).
//
// Ports gen-tables.ts `parseTextToLines`. The `verbose` parameter was
// dropped: the TS call site (inside getSlidesForTableRows) always passes
// `false`, and verbose only ever gated console.log statements with no effect
// on the returned lines.
func parseTextToLines(cell TableCell, colWidth float64) [][]TableCell {
	cellOpts := cell.Options

	autoPageCharWeight := 0.0
	fontSize := float64(DEF_FONT_SIZE)
	if cellOpts != nil {
		autoPageCharWeight = cellOpts.AutoPageCharWeight
		if cellOpts.FontSize != 0 {
			fontSize = cellOpts.FontSize
		}
	}
	foco := 2.3 + autoPageCharWeight                                              // Character Constant
	cpl := math.Floor((colWidth/float64(ONEPT))*float64(EMU)) / (fontSize / foco) // Chars-Per-Line

	// STEP 1: Ensure inputCells is a slice of TableCells.
	var inputCells []TableCell
	if len(cell.TextCells) > 0 {
		inputCells = cell.TextCells
	} else {
		trimmed := strings.TrimSpace(cell.Text)
		if cell.Text != "" && trimmed == "" {
			// Allow a single space/whitespace as cell text (user-requested feature).
			inputCells = []TableCell{{Text: " "}}
		} else {
			inputCells = []TableCell{{Text: trimmed}}
		}
	}

	// STEP 2+3 (combined): group into "fragments" — one fragment per input
	// cell, or one fragment per "\n"-split segment for multi-line text — then
	// tokenize each fragment's text into words.
	//
	// NOTE: the TS source builds an intermediate `inputLines1` (grouping
	// sibling non-breakLine cells onto one "line") and then a further
	// `inputLines2`, but STEP 3 explodes inputLines1 back into one
	// inputLines2 entry *per cell* regardless of that grouping (it iterates
	// `line.forEach(cell => ...)`), and STEP 4 resets its line-accumulator
	// at the start of every inputLines2 entry. Net effect: the STEP 2
	// grouping has zero influence on the final `parsedLines` — every
	// original cell (or "\n" fragment) always starts a fresh physical line.
	// This flattens the three steps into one without changing output.
	var inputLines2 [][]TableCell
	for _, c := range inputCells {
		if strings.Contains(c.Text, "\n") {
			for _, part := range strings.Split(c.Text, "\n") {
				fragOpts := cloneCellProps(c.Options)
				fragOpts.BreakLine = ptr(true)
				inputLines2 = append(inputLines2, tokenizeWords(part, fragOpts))
			}
		} else {
			inputLines2 = append(inputLines2, tokenizeWords(strings.TrimSpace(c.Text), c.Options))
		}
	}

	// STEP 4: Group cells/words into lines based upon space consumed by word letters.
	var parsedLines [][]TableCell
	for _, line := range inputLines2 {
		var lineCells []TableCell
		strCurrLineLen := 0
		for _, word := range line {
			// A: create new line when horizontal space is exhausted.
			if float64(strCurrLineLen+len(word.Text)) > cpl {
				parsedLines = append(parsedLines, lineCells)
				lineCells = nil
				strCurrLineLen = 0
			}
			// B: add current word to line cells.
			lineCells = append(lineCells, word)
			// C: track line's char length.
			strCurrLineLen += len(word.Text)
		}
		// Flush buffer: only create a line when there's text (avoid empty row).
		if len(lineCells) > 0 {
			parsedLines = append(parsedLines, lineCells)
		}
	}

	return parsedLines
}

// GetSlidesForTableRows takes an array of table rows and breaks it into an
// array of slides, each holding the rows that fit within the estimated
// available vertical space.
//
// Ports gen-tables.ts `getSlidesForTableRows`.
func GetSlidesForTableRows(tableRows []TableRow, tableProps *TableToSlidesProps, presLayout PresLayout, masterSlide *SlideLayout) []TableRowSlide {
	if tableProps == nil {
		tableProps = &TableToSlidesProps{}
	}

	arrInchMargins := DEF_SLIDE_MARGIN_IN
	var emuSlideTabW float64 = EMU
	var emuSlideTabH float64 = EMU
	var emuTabCurrH float64
	numCols := 0
	var tableRowSlides []TableRowSlide

	tablePropX := float64(getSmartParseNumber(coordOrZero(tableProps.X), "X", presLayout))
	tablePropY := float64(getSmartParseNumber(coordOrZero(tableProps.Y), "Y", presLayout))
	tablePropW := float64(getSmartParseNumber(coordOrZero(tableProps.W), "X", presLayout))
	tablePropH := float64(getSmartParseNumber(coordOrZero(tableProps.H), "Y", presLayout))
	tableCalcW := tablePropW

	// calcSlideTabH recomputes the available table height (emuSlideTabH) for
	// the *current* slide (tableRowSlides is read for its current length —
	// closures capture by reference, so appends elsewhere are visible here).
	calcSlideTabH := func() {
		var emuStartY float64
		if len(tableRowSlides) == 0 {
			if tablePropY != 0 {
				emuStartY = tablePropY
			} else {
				emuStartY = float64(inch2Emu(arrInchMargins[0]))
			}
		}
		if len(tableRowSlides) > 0 {
			// TS `autoPageSlideStartY || newSlideStartY || arrInchMargins[0]`
			// (gen-tables.ts:197): JS `||`, so nil OR explicit 0 falls through.
			startYIn := arrInchMargins[0]
			if tableProps.AutoPageSlideStartY != nil && *tableProps.AutoPageSlideStartY != 0 {
				startYIn = *tableProps.AutoPageSlideStartY
			} else if tableProps.NewSlideStartY != nil && *tableProps.NewSlideStartY != 0 { // @deprecated v3.3.0
				startYIn = *tableProps.NewSlideStartY
			}
			emuStartY = float64(inch2Emu(startYIn))
		}

		tabHOrLayout := tablePropH
		if tabHOrLayout == 0 {
			tabHOrLayout = float64(presLayout.Height)
		}
		emuSlideTabH = tabHOrLayout - emuStartY - float64(inch2Emu(arrInchMargins[2]))

		if len(tableRowSlides) > 1 {
			// RULE: Use margins for starting point after the initial Slide, not
			// `opt.y` (ISSUE #43, ISSUE #47, ISSUE #48).
			// TS `typeof === 'number'` (gen-tables.ts:203-205): an explicit 0 IS
			// an override (non-nil), distinct from the nil/unset fall-through.
			if tableProps.AutoPageSlideStartY != nil {
				emuSlideTabH = tabHOrLayout - float64(inch2Emu(*tableProps.AutoPageSlideStartY+arrInchMargins[2]))
			} else if tableProps.NewSlideStartY != nil { // @deprecated v3.3.0
				emuSlideTabH = tabHOrLayout - float64(inch2Emu(*tableProps.NewSlideStartY+arrInchMargins[2]))
			} else if tablePropY != 0 {
				base := arrInchMargins[0]
				if tablePropY/float64(EMU) < arrInchMargins[0] {
					base = tablePropY / float64(EMU)
				}
				emuSlideTabH = tabHOrLayout - float64(inch2Emu(base+arrInchMargins[2]))
				// Use whichever is greater: area between margins or the table H
				// provided (don't shrink usable area).
				if emuSlideTabH < tablePropH {
					emuSlideTabH = tablePropH
				}
			}
		}
	}

	// STEP 1: Calculate margins.
	{
		slideMargin := tableProps.SlideMargin
		if slideMargin == nil {
			slideMargin = Margin{DEF_SLIDE_MARGIN_IN[0]}
		}
		if masterSlide != nil && masterSlide.Margin != nil {
			arrInchMargins = expandMarginTRBL(masterSlide.Margin, arrInchMargins)
		} else if slideMargin != nil {
			arrInchMargins = expandMarginTRBL(slideMargin, arrInchMargins)
		}
	}

	// STEP 2: Calculate number of columns.
	// NOTE: cells may have a colspan, so the length of row[0] alone isn't
	// sufficient to determine column count.
	{
		var firstRow TableRow
		if len(tableRows) > 0 {
			firstRow = tableRows[0]
		}
		for _, cell := range firstRow {
			colspan := 1
			if cell.Options != nil && cell.Options.Colspan != 0 {
				colspan = cell.Options.Colspan
			}
			numCols += colspan
		}
	}

	// STEP 3: Calculate width using tableProps.ColW if possible.
	// M3: len==1 is the TS scalar `colW` shorthand (uniform width per column),
	// so total = colW * numCols; len>1 sums the explicit per-column widths.
	if tablePropW == 0 && len(tableProps.ColW) > 0 {
		if len(tableProps.ColW) == 1 {
			tableCalcW = tableProps.ColW[0] * float64(numCols) * float64(EMU)
		} else {
			sum := 0.0
			for _, w := range tableProps.ColW {
				sum += w
			}
			tableCalcW = sum * float64(EMU)
		}
	}

	// STEP 4: Calculate usable width now that total usable space is known.
	{
		if tableCalcW != 0 {
			emuSlideTabW = tableCalcW
		} else {
			fallbackX := arrInchMargins[1]
			if tablePropX != 0 {
				fallbackX = tablePropX / float64(EMU)
			}
			emuSlideTabW = float64(inch2Emu(fallbackX + arrInchMargins[3]))
		}
	}

	// STEP 5: Calculate column widths if not provided (distribute evenly), or
	// expand the M3 scalar shorthand (len==1) to a uniform per-column slice.
	if len(tableProps.ColW) == 0 {
		tableProps.ColW = make([]float64, numCols)
		for i := range tableProps.ColW {
			tableProps.ColW[i] = emuSlideTabW / float64(EMU) / float64(numCols)
		}
	} else if len(tableProps.ColW) == 1 && numCols > 1 {
		w := tableProps.ColW[0]
		tableProps.ColW = make([]float64, numCols)
		for i := range tableProps.ColW {
			tableProps.ColW[i] = w
		}
	}

	// STEP 6: **MAIN** Iterate over rows, add table content, create new
	// slides as rows overflow.
	newTableRowSlide := TableRowSlide{}
	for _, row := range tableRows {
		// DEVIATION (documented in the tables.go header): an empty row is
		// skipped here. Upstream TS indexes `row[0]`/reduces over the row and
		// crashes on a zero-length row; skipping is the safer Go behavior and is
		// locked by TestGetSlidesForTableRows_EmptyRowSkipped.
		if len(row) == 0 {
			continue
		}

		// A/B: Row variables + resolved (nil-safe) per-column options. The
		// same *TableCellProps pointer is reused for both `currTableRow` and
		// `rowCellLines` below, matching the TS aliasing where
		// `newCell.options = cell.options` shares the same object (so later
		// in-place mutation, e.g. AutoPageCharWeight, is visible from both).
		cellOptsList := make([]*TableCellProps, len(row))
		for i, cell := range row {
			o := cell.Options
			if o == nil {
				o = &TableCellProps{}
			}
			cellOptsList[i] = o
		}

		buildRow := func() TableRow {
			r := make(TableRow, len(row))
			for i := range row {
				r[i] = TableCell{Type: SlideObjectTypeTablecell, Options: cellOptsList[i]}
			}
			return r
		}

		var maxCellMarTopEmu, maxCellMarBtmEmu float64
		currTableRow := buildRow()
		for i := range row {
			o := cellOptsList[i]
			cellMargin := o.Margin
			pointsMode := marginTRBLTop(cellMargin) >= 1

			top := marginTRBLTop(cellMargin)
			tableTop := marginTRBLTop(tableProps.Margin)
			bottom := marginTRBLBottom(cellMargin)
			tableBottom := marginTRBLBottom(tableProps.Margin)

			if pointsMode {
				if cellMargin != nil && top != 0 {
					if v := float64(valToPts(top)); v > maxCellMarTopEmu {
						maxCellMarTopEmu = v
					}
				} else if tableProps.Margin != nil && tableTop != 0 {
					if v := float64(valToPts(tableTop)); v > maxCellMarTopEmu {
						maxCellMarTopEmu = v
					}
				}
				if cellMargin != nil && bottom != 0 {
					if v := float64(valToPts(bottom)); v > maxCellMarBtmEmu {
						maxCellMarBtmEmu = v
					}
				} else if tableProps.Margin != nil && tableBottom != 0 {
					if v := float64(valToPts(tableBottom)); v > maxCellMarBtmEmu {
						maxCellMarBtmEmu = v
					}
				}
			} else {
				if cellMargin != nil && top != 0 {
					if v := float64(inch2Emu(top)); v > maxCellMarTopEmu {
						maxCellMarTopEmu = v
					}
				} else if tableProps.Margin != nil && tableTop != 0 {
					if v := float64(inch2Emu(tableTop)); v > maxCellMarTopEmu {
						maxCellMarTopEmu = v
					}
				}
				if cellMargin != nil && bottom != 0 {
					if v := float64(inch2Emu(bottom)); v > maxCellMarBtmEmu {
						maxCellMarBtmEmu = v
					}
				} else if tableProps.Margin != nil && tableBottom != 0 {
					if v := float64(inch2Emu(tableBottom)); v > maxCellMarBtmEmu {
						maxCellMarBtmEmu = v
					}
				}
			}
		}

		// C: Calc usable vertical space/table height. Default first, adjusted
		// below inside the paging loop when necessary.
		calcSlideTabH()
		emuTabCurrH += maxCellMarTopEmu + maxCellMarBtmEmu

		// D: --==[[ BUILD DATA SET ]]==-- (split each cell's text into
		// lines[], compute lineHeight).
		rowCellLines := make([]TableCell, len(row))
		for i, cell := range row {
			o := cellOptsList[i]
			fontSize := resolveFontSize(o, tableProps)
			lineHeight := float64(inch2Emu((fontSize * (LINEH_MODIFIER + tableProps.AutoPageLineWeight)) / 100))

			// E-1: Exempt rowspan cells from increasing lineHeight (or a new
			// slide could be created unnecessarily).
			if o.Rowspan != 0 {
				lineHeight = 0
			}

			// E-2: parseTextToLines uses autoPageCharWeight — inherit from
			// table options (this OVERWRITES any cell-level value, mutating
			// the shared *TableCellProps in place, same as the TS source).
			o.AutoPageCharWeight = tableProps.AutoPageCharWeight

			// E-3: **MAIN** total column width for this cell (handles colspan).
			totalColW := colWAt(tableProps.ColW, i)
			if o.Colspan != 0 && len(tableProps.ColW) > 0 {
				// TS DEVIATION (preserved bug, not fixed): the original filter
				// is `idx >= iCell && idx < idx + cell.options.colspan`, whose
				// upper bound (`idx < idx+colspan`) is a tautology for any
				// colspan > 0 (a number is always less than itself plus a
				// positive value) — so it actually sums colW[iCell:] (every
				// remaining column), not just `colspan` many. Preserved here
				// verbatim for byte-identical row-split math.
				sum := 0.0
				for idx := i; idx < len(tableProps.ColW); idx++ {
					sum += tableProps.ColW[idx]
				}
				totalColW = sum
			}

			// E-4: parse into lines using the resolved (mutated) options.
			cellForParse := cell
			cellForParse.Options = o
			lines := parseTextToLines(cellForParse, totalColW)

			rowCellLines[i] = TableCell{Type: SlideObjectTypeTablecell, Lines: lines, LineHeight: lineHeight, Options: o}
		}

		// E: --==[[ PAGE DATA SET ]]==--
		// Add text one-line-at-a-time to this row's cells until lines are
		// exhausted or the table height limit is hit.
		currCellIdx := 0
		var emuLineMaxH float64
		for {
			srcCell := &rowCellLines[currCellIdx]
			tgtCell := &currTableRow[currCellIdx]

			// 1: calc emuLineMaxH (running max across the whole while loop;
			// recomputed every iteration in the TS source, though cell
			// LineHeight values never change, so this converges immediately).
			for _, c := range rowCellLines {
				if c.LineHeight >= emuLineMaxH {
					emuLineMaxH = c.LineHeight
				}
			}

			// 2: create a new slide if there is insufficient room for the
			// current row.
			if emuTabCurrH+emuLineMaxH > emuSlideTabH {
				// A: add current row slide or it will be lost (only if it has
				// rows and text).
				hasContent := false
				for _, c := range currTableRow {
					if len(c.TextCells) > 0 {
						hasContent = true
						break
					}
				}
				if len(currTableRow) > 0 && hasContent {
					newTableRowSlide.Rows = append(newTableRowSlide.Rows, currTableRow)
				}

				// B: add current slide to Slides slice.
				tableRowSlides = append(tableRowSlides, newTableRowSlide)

				// C: reset working/curr slide.
				newTableRowSlide = TableRowSlide{}

				// D: reset working/curr row.
				currTableRow = buildRow()

				// E: recalc usable vertical space now (tableRowSlides grew).
				calcSlideTabH()
				emuTabCurrH += maxCellMarTopEmu + maxCellMarBtmEmu // no-op: overwritten by F below (preserved for fidelity)

				// F: reset current table height for this new Slide.
				emuTabCurrH = 0

				// G: handle repeat headers option / continue current lines.
				if (boolVal(tableProps.AddHeaderToEach) || boolVal(tableProps.AutoPageRepeatHeader)) && len(tableProps.ArrObjTabHeadRows) > 0 {
					for _, headRow := range tableProps.ArrObjTabHeadRows {
						newHeadRow := make(TableRow, 0, len(headRow))
						var maxLineHeight float64
						for _, c := range headRow {
							newHeadRow = append(newHeadRow, c)
							if c.LineHeight > maxLineHeight {
								maxLineHeight = c.LineHeight
							}
						}
						newTableRowSlide.Rows = append(newTableRowSlide.Rows, newHeadRow)
						emuTabCurrH += maxLineHeight
					}
				}

				tgtCell = &currTableRow[currCellIdx]
			}

			// 3: set slice of words that comprise this line (TS `.shift()`).
			var currLine []TableCell
			if len(srcCell.Lines) > 0 {
				currLine = srcCell.Lines[0]
				srcCell.Lines = srcCell.Lines[1:]
			}

			// 4: append the line's words, or an empty placeholder cell if
			// there's no more content and the target has none yet (avoids a
			// "needs repair" issue from null cell content).
			if currLine != nil {
				tgtCell.TextCells = append(tgtCell.TextCells, currLine...)
			} else if len(tgtCell.TextCells) == 0 {
				tgtCell.TextCells = append(tgtCell.TextCells, TableCell{Type: SlideObjectTypeTablecell, Text: ""})
			}

			// 5: increase table height by the curr line height (once per
			// full row-of-columns, i.e. on the last column).
			if currCellIdx == len(rowCellLines)-1 {
				emuTabCurrH += emuLineMaxH
			}

			// 6: advance column/cell index (circle back to continue adding lines).
			if currCellIdx < len(rowCellLines)-1 {
				currCellIdx++
			} else {
				currCellIdx = 0
			}

			// 7: done when every cell's lines are exhausted.
			remaining := 0
			for _, c := range rowCellLines {
				remaining += len(c.Lines)
			}
			if remaining == 0 {
				break
			}
		}

		// F: Flush/capture row buffer before it resets at the top of this loop.
		if len(currTableRow) > 0 {
			newTableRowSlide.Rows = append(newTableRowSlide.Rows, currTableRow)
		}
	}

	// STEP 7: Flush buffer / add final slide.
	tableRowSlides = append(tableRowSlides, newTableRowSlide)

	return tableRowSlides
}
