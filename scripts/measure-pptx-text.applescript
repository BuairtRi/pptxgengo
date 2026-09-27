use scripting additions

property lineTopEpsilon : 0.5

on run argv
	if (count of argv) is not 4 then error "Usage: osascript measure-pptx-text.applescript 'presentation name' slide-number shape-index phrase" number 64
	set presentationName to item 1 of argv
	try
		set slideNumber to (item 2 of argv) as integer
		set shapeIndex to (item 3 of argv) as integer
	on error
		error "Slide number and shape index must be integers" number 64
	end try
	set requestedPhrase to item 4 of argv
	if slideNumber < 1 or shapeIndex < 1 then error "Slide number and shape index must be positive" number 64
	if requestedPhrase is "" then error "Phrase must not be empty" number 64

	tell application "Microsoft PowerPoint"
		set matchingPresentations to every presentation whose name is presentationName
		if (count of matchingPresentations) is not 1 then error "Expected exactly one open presentation named " & quoted form of presentationName number 65
		set p to item 1 of matchingPresentations
		if slideNumber > (count of slides of p) then error "Slide number is outside the presentation" number 66
		set sl to slide slideNumber of p
		if shapeIndex > (count of shapes of sl) then error "Shape index is outside the slide" number 67
		set sh to shape shapeIndex of sl
		if (has text frame of sh) is false then error "Selected shape has no text frame" number 68
		set shapeRotation to rotation of sh
		set shapeName to name of sh
		set shapeLeft to left position of sh
		set shapeTop to top of sh
		set shapeWidth to width of sh
		set shapeHeight to height of sh
		set tr to text range of text frame of sh
		set fullText to content of tr
		set textLength to text length of tr
	end tell

	-- Match exact literal text and reject duplicates so the caller must refine
	-- its phrase instead of silently measuring the wrong occurrence.
	set phraseLength to count of characters of requestedPhrase
	set occurrenceStarts to {}
	if phraseLength > textLength then error "Phrase was not found in the selected shape" number 69
	repeat with candidateStart from 1 to (textLength - phraseLength + 1)
		set candidateEnd to candidateStart + phraseLength - 1
		if (text candidateStart thru candidateEnd of fullText) is requestedPhrase then set end of occurrenceStarts to candidateStart
	end repeat
	if (count of occurrenceStarts) is 0 then error "Phrase was not found in the selected shape" number 69
	if (count of occurrenceStarts) > 1 then error "Phrase occurs more than once; pass a more specific phrase" number 70
	set matchStart to item 1 of occurrenceStarts
	set matchEnd to matchStart + phraseLength - 1

	-- Keep a literal match, but omit any trailing whitespace from the measured
	-- character range, as trailing spaces have no visible glyph bounds.
	repeat while matchEnd >= matchStart
		set tailCharacter to character matchEnd of fullText
		if tailCharacter is " " or tailCharacter is tab or tailCharacter is return or tailCharacter is linefeed then
			set matchEnd to matchEnd - 1
		else
			exit repeat
		end if
	end repeat
	if matchEnd < matchStart then error "Phrase contains no measurable characters after trailing whitespace is excluded" number 71

	set characterRows to {}
	set lineRows to {}
	set unionLeft to missing value
	set unionTop to missing value
	set unionRight to missing value
	set unionBottom to missing value
	tell application "Microsoft PowerPoint"
		repeat with charIndex from matchStart to matchEnd
			set charRange to character charIndex of tr
			set charText to content of charRange
			set charLeft to left bounds of charRange
			set charTop to top bounds of charRange
			set charWidth to bounds width of charRange
			set charHeight to bounds height of charRange
			set charRight to charLeft + charWidth
			set charBottom to charTop + charHeight
			set end of characterRows to {charIndex, charText, charLeft, charTop, charWidth, charHeight}
			if unionLeft is missing value then
				set unionLeft to charLeft
				set unionTop to charTop
				set unionRight to charRight
				set unionBottom to charBottom
			else
				if charLeft < unionLeft then set unionLeft to charLeft
				if charTop < unionTop then set unionTop to charTop
				if charRight > unionRight then set unionRight to charRight
				if charBottom > unionBottom then set unionBottom to charBottom
			end if
			set matchingLine to 0
			repeat with lineIndex from 1 to (count of lineRows)
				set rowTop to item 1 of item lineIndex of lineRows
				if my absoluteValue(charTop - rowTop) ≤ lineTopEpsilon then
					set matchingLine to lineIndex
					exit repeat
				end if
			end repeat
			if matchingLine is 0 then
				set end of lineRows to {charTop, charIndex, charIndex, charLeft, charTop, charRight, charBottom}
			else
				set oldRow to item matchingLine of lineRows
				set rowStart to item 2 of oldRow
				set rowEnd to item 3 of oldRow
				set rowLeft to item 4 of oldRow
				set rowTop to item 5 of oldRow
				set rowRight to item 6 of oldRow
				set rowBottom to item 7 of oldRow
				if charIndex < rowStart then set rowStart to charIndex
				if charIndex > rowEnd then set rowEnd to charIndex
				if charLeft < rowLeft then set rowLeft to charLeft
				if charTop < rowTop then set rowTop to charTop
				if charRight > rowRight then set rowRight to charRight
				if charBottom > rowBottom then set rowBottom to charBottom
				set item matchingLine of lineRows to {item 1 of oldRow, rowStart, rowEnd, rowLeft, rowTop, rowRight, rowBottom}
			end if
		end repeat
	end tell

	set outputJSON to "{\"shape_name\":" & my jsonString(shapeName) & ",\"shape_bounds\":{\"left\":" & my jsonNumber(shapeLeft) & ",\"top\":" & my jsonNumber(shapeTop) & ",\"width\":" & my jsonNumber(shapeWidth) & ",\"height\":" & my jsonNumber(shapeHeight) & "},\"presentation\":" & my jsonString(presentationName) & ",\"slide\":" & slideNumber & ",\"shape_index\":" & shapeIndex & ",\"phrase\":" & my jsonString(requestedPhrase) & ",\"text\":" & my jsonString(fullText) & ",\"start_character\":" & matchStart & ",\"end_character\":" & matchEnd & ",\"coordinates\":\"raw PowerPoint text-range units; scripting dictionary does not specify units\",\"bounds\":{\"left\":" & my jsonNumber(unionLeft) & ",\"top\":" & my jsonNumber(unionTop) & ",\"width\":" & my jsonNumber(unionRight - unionLeft) & ",\"height\":" & my jsonNumber(unionBottom - unionTop) & "},\"line_top_epsilon\":" & my jsonNumber(lineTopEpsilon) & ",\"rotation_handled\":false,\"rotation_degrees\":" & my jsonNumber(shapeRotation) & ",\"lines\":["
	repeat with lineIndex from 1 to (count of lineRows)
		if lineIndex > 1 then set outputJSON to outputJSON & ","
		set lineRow to item lineIndex of lineRows
		set outputJSON to outputJSON & "{\"start_character\":" & item 2 of lineRow & ",\"end_character\":" & item 3 of lineRow & ",\"bounds\":{\"left\":" & my jsonNumber(item 4 of lineRow) & ",\"top\":" & my jsonNumber(item 5 of lineRow) & ",\"width\":" & my jsonNumber((item 6 of lineRow) - (item 4 of lineRow)) & ",\"height\":" & my jsonNumber((item 7 of lineRow) - (item 5 of lineRow)) & "}}"
	end repeat
	set outputJSON to outputJSON & "],\"characters\":["
	repeat with rowIndex from 1 to (count of characterRows)
		if rowIndex > 1 then set outputJSON to outputJSON & ","
		set charRow to item rowIndex of characterRows
		set outputJSON to outputJSON & "{\"index\":" & item 1 of charRow & ",\"text\":" & my jsonString(item 2 of charRow) & ",\"left\":" & my jsonNumber(item 3 of charRow) & ",\"top\":" & my jsonNumber(item 4 of charRow) & ",\"width\":" & my jsonNumber(item 5 of charRow) & ",\"height\":" & my jsonNumber(item 6 of charRow) & "}"
	end repeat
	set outputJSON to outputJSON & "]}"
	return outputJSON
end run

on absoluteValue(valueNumber)
	if valueNumber < 0 then return 0 - valueNumber
	return valueNumber
end absoluteValue

on jsonNumber(valueNumber)
	if valueNumber is missing value then return "null"
	return valueNumber as text
end jsonNumber

on jsonString(sourceText)
	set slash to ASCII character 92
	set quoteMark to ASCII character 34
	set outputText to quoteMark
	repeat with sourceCharacter in characters of sourceText
		set sourceCharacter to contents of sourceCharacter
		set codePoint to id of sourceCharacter
		if codePoint is 34 then
			set outputText to outputText & slash & quoteMark
		else if codePoint is 92 then
			set outputText to outputText & slash & slash
		else if codePoint is 8 then
			set outputText to outputText & slash & "b"
		else if codePoint is 9 then
			set outputText to outputText & slash & "t"
		else if codePoint is 10 then
			set outputText to outputText & slash & "n"
		else if codePoint is 12 then
			set outputText to outputText & slash & "f"
		else if codePoint is 13 then
			set outputText to outputText & slash & "r"
		else if codePoint < 32 then
			set outputText to outputText & slash & "u00" & my hexByte(codePoint)
		else
			set outputText to outputText & sourceCharacter
		end if
	end repeat
	return outputText & quoteMark
end jsonString

on hexByte(valueNumber)
	set digits to "0123456789abcdef"
	set highDigit to (valueNumber div 16) + 1
	set lowDigit to (valueNumber mod 16) + 1
	return character highDigit of digits & character lowDigit of digits
end hexByte
