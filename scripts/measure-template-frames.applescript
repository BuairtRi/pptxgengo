-- Read-only native text bounds for an already-open review deck.
-- Reports every text shape, including group children and table cells.
-- This is geometry evidence, not an optical or semantic acceptance decision.
use scripting additions
property outputRows : {}

on run argv
	if (count argv) < 1 or (count argv) > 2 then error "Usage: measure-template-frames.applescript presentation-name [source-scene-sha256]" number 64
	set outputRows to {}
	set presentationName to item 1 of argv
	set sceneSHA to ""
	if (count argv) is 2 then set sceneSHA to item 2 of argv
	with timeout of 120 seconds
		tell application "Microsoft PowerPoint"
			set candidates to every presentation whose name is presentationName
			if (count candidates) is not 1 then error "Expected one open review presentation" number 65
			set p to item 1 of candidates
			set slideCount to count of slides of p
			repeat with slideNumber from 1 to slideCount
				set sl to slide slideNumber of p
				repeat with shapeNumber from 1 to (count of shapes of sl)
					my inspectShape(shape shapeNumber of sl, slideNumber, shapeNumber as text, "slide")
				end repeat
			end repeat
		end tell
	end timeout
	set savedDelimiters to AppleScript's text item delimiters
	set AppleScript's text item delimiters to ","
	set body to outputRows as text
	set AppleScript's text item delimiters to savedDelimiters
	return "{\"schema\":\"pptxgengo.template-native-frames.v1\",\"presentation\":" & my jsonString(presentationName) & ",\"source_scene_sha256\":" & my jsonString(sceneSHA) & ",\"coordinates\":\"raw native PowerPoint units; convert only after confirming the PowerPoint scripting API unit\",\"qualification\":\"measurement_only\",\"shapes\":[" & body & "]}"
end run

on inspectShape(sh, slideNumber, shapePath, coordinateScope)
	try
		tell application "Microsoft PowerPoint"
			set sp to properties of sh
			if shape type of sp is shape type group then
				repeat with k from 1 to (count of shapes of sh)
					my inspectShape(shape k of sh, slideNumber, shapePath & "." & k, "group_local")
				end repeat
			else if has table of sp then
				set tb to table object of sh
				repeat with r from 1 to (count of rows of tb)
					repeat with c from 1 to (count of columns of tb)
						set cellRef to get cell from tb row r column c
						my inspectShape(shape of cellRef, slideNumber, shapePath & ".r" & r & "c" & c, "table_cell")
					end repeat
				end repeat
			else if has text frame of sp then
				set tf to text frame of sh
				set fp to properties of tf
				if has text of fp then
					set rp to properties of text range of tf
					set txt to content of rp
					set w to width of sp
					set h to height of sp
					set ml to margin left of fp
					set mr to margin right of fp
					set mt to margin top of fp
					set mb to margin bottom of fp
					set bw to bounds width of rp
					set bh to bounds height of rp
					set angle to rotation of sh
					set shapeLeft to left position of sp
					set shapeTop to top of sp
					set textLeft to left bounds of rp
					set textTop to top bounds of rp
					set shapeTypeName to shape type of sp as text
					set recordJSON to "{\"slide\":" & slideNumber & ",\"shape_path\":" & my jsonString(shapePath) & ",\"coordinate_scope\":" & my jsonString(coordinateScope) & ",\"shape_type\":" & my jsonString(shapeTypeName) & ",\"name\":" & my jsonString(name of sp) & ",\"text\":" & my jsonString(txt) & ",\"shape_frame\":{\"left\":" & shapeLeft & ",\"top\":" & shapeTop & ",\"width\":" & w & ",\"height\":" & h & ",\"rotation_degrees\":" & my jsonNumber(angle) & "},\"text_bounds\":{\"left\":" & textLeft & ",\"top\":" & textTop & ",\"width\":" & bw & ",\"height\":" & bh & "},\"margins\":{\"left\":" & ml & ",\"right\":" & mr & ",\"top\":" & mt & ",\"bottom\":" & mb & "},\"frame_width\":" & w & ",\"frame_height\":" & h & ",\"inner_width\":" & (w - ml - mr) & ",\"inner_height\":" & (h - mt - mb) & ",\"text_width\":" & bw & ",\"text_height\":" & bh & ",\"shape_left\":" & shapeLeft & ",\"shape_top\":" & shapeTop & ",\"text_left\":" & textLeft & ",\"text_top\":" & textTop & "}"
					set end of outputRows to recordJSON
				end if
			end if
		end tell
	on error messageText number errorNumber
		set end of outputRows to "{\"slide\":" & slideNumber & ",\"shape_path\":" & my jsonString(shapePath) & ",\"coordinate_scope\":" & my jsonString(coordinateScope) & ",\"error\":" & my jsonString(messageText) & ",\"error_number\":" & errorNumber & "}"
	end try
end inspectShape

on jsonString(sourceText)
	set slash to ASCII character 92
	set quoteMark to ASCII character 34
	set outputText to quoteMark
	repeat with sourceCharacter in characters of (sourceText as text)
		set sourceCharacter to contents of sourceCharacter
		set codePoint to id of sourceCharacter
		if codePoint is 34 then
			set outputText to outputText & slash & quoteMark
		else if codePoint is 92 then
			set outputText to outputText & slash & slash
		else if codePoint is 9 then
			set outputText to outputText & slash & "t"
		else if codePoint is 10 then
			set outputText to outputText & slash & "n"
		else if codePoint is 13 then
			set outputText to outputText & slash & "r"
		else if codePoint < 32 then
			set digits to "0123456789abcdef"
			set outputText to outputText & slash & "u00" & character ((codePoint div 16) + 1) of digits & character ((codePoint mod 16) + 1) of digits
		else
			set outputText to outputText & sourceCharacter
		end if
	end repeat
	return outputText & quoteMark
end jsonString

on jsonNumber(v)
	if v is missing value then return "null"
	return v as text
end jsonNumber
