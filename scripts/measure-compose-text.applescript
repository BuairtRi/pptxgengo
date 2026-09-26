-- Read-only whole-text-range measurement for generated composition probes.
-- The probe presentation must already be open in PowerPoint. This script never
-- edits, saves, exports, or closes a presentation.
use scripting additions

property rotationEpsilon : 0.001

on run argv
	if (count of argv) is not 1 then error "Usage: osascript measure-compose-text.applescript 'open presentation name'" number 64
	set presentationName to item 1 of argv
	if presentationName is "" then error "Presentation name must not be empty" number 64

	tell application "Microsoft PowerPoint"
		set openNames to name of every presentation
		set matchCount to 0
		repeat with candidateName in openNames
			if (contents of candidateName as text) is presentationName then set matchCount to matchCount + 1
		end repeat
		if matchCount is not 1 then error "Expected exactly one open presentation named " & quoted form of presentationName number 65
		set p to presentation presentationName
		set slideCount to count of slides of p
		if slideCount < 1 then error "Presentation has no slides" number 66
		set outputRows to ""
		set visibleSlideCount to 0
		repeat with slideIndex from 1 to slideCount
			set sl to slide slideIndex of p
			set slideIsHidden to hidden of slide show transition of sl
			if slideIsHidden is false then
				set visibleSlideCount to visibleSlideCount + 1
				set shapeCount to count of shapes of sl
				set seenShapeNames to {}
				repeat with shapeIndex from 1 to shapeCount
					set sh to shape shapeIndex of sl
					if visible of sh then
						set shapeName to name of sh
						if shapeName is "" then error "Visible shape has no name on slide " & slideIndex & ", shape " & shapeIndex number 67
						if seenShapeNames contains shapeName then error "Visible shape name is duplicated on slide " & slideIndex & ": " & shapeName number 74
						set end of seenShapeNames to shapeName
						set frameLeft to left position of sh
						set frameTop to top of sh
						set frameWidth to width of sh
						set frameHeight to height of sh
						set shapeRotation to rotation of sh
						if frameWidth ≤ 0 or frameHeight ≤ 0 then error "Visible shape has nonpositive frame on slide " & slideIndex & ", shape " & shapeIndex number 68
						if my absoluteValue(shapeRotation) > rotationEpsilon then error "Rotated shape is outside this adapter's supported measurements: " & shapeName & " on slide " & slideIndex number 69
						set fillVisibleValue to false
						set fillRGBValue to missing value
						set fillTransparencyValue to missing value
						set shapeFill to fill format of sh
						set fillVisibleValue to visible of shapeFill
						if fillVisibleValue then
							set fillRGBValue to fore color of shapeFill
							if (count of fillRGBValue) is not 3 then error "Visible shape fill is not a resolved RGB color: " & shapeName number 79
							set fillTransparencyValue to transparency of shapeFill
						end if
						set fillJSON to "{\"visible\":" & my jsonBoolean(fillVisibleValue) & ",\"rgb\":" & my jsonRGB(fillRGBValue) & ",\"transparency\":" & my jsonNullableNumber(fillTransparencyValue) & "}"

						set textValue to missing value
						set textBoundsValue to missing value
						set fontNameValue to missing value
						set fontSizeValue to missing value
						set boldValue to missing value
						set textColorValue to missing value
						set marginsValue to missing value
						if has text frame of sh then
							set tr to text range of text frame of sh
							set tf to text frame of sh
							set marginLeftValue to margin left of tf
							set marginRightValue to margin right of tf
							set marginTopValue to margin top of tf
							set marginBottomValue to margin bottom of tf
							set marginsValue to "{\"left\":" & my jsonNumber(marginLeftValue) & ",\"right\":" & my jsonNumber(marginRightValue) & ",\"top\":" & my jsonNumber(marginTopValue) & ",\"bottom\":" & my jsonNumber(marginBottomValue) & "}"
							set textValue to content of tr
							if textValue is not "" then
								set textLeft to left bounds of tr
								set textTop to top bounds of tr
								set textWidth to bounds width of tr
								set textHeight to bounds height of tr
								if textWidth ≤ 0 or textHeight ≤ 0 then error "Text range has nonpositive bounds: " & shapeName & " on slide " & slideIndex number 71
								set textBoundsValue to "{\"left\":" & my jsonNumber(textLeft) & ",\"top\":" & my jsonNumber(textTop) & ",\"width\":" & my jsonNumber(textWidth) & ",\"height\":" & my jsonNumber(textHeight) & "}"
							-- Whole-range font/color can resolve to paragraph defaults. Read
							-- visible characters and require one uniform effective style.
							set uniformFontName to missing value
							set uniformFontSize to missing value
							set uniformBold to missing value
							set uniformTextColor to missing value
							set measuredCharacterCount to 0
							set fullTextLength to text length of tr
							repeat with charIndex from 1 to fullTextLength
								set charRange to character charIndex of tr
								set charText to content of charRange
								if charText is not return and charText is not linefeed then
									try
										set charFont to font of charRange
										set charFontName to font name of charFont
										set charFontSize to font size of charFont
										set charBold to bold of charFont
										set charFontColor to font color of charFont
									on error
										error "Font attributes unavailable for character " & charIndex & " in " & shapeName number 75
									end try
									if charFontName is missing value or charFontSize is missing value or charBold is missing value or charFontColor is missing value or (count of charFontColor) is not 3 then error "Mixed or unresolved character font/color in " & shapeName number 76
									if measuredCharacterCount is 0 then
										set uniformFontName to charFontName
										set uniformFontSize to charFontSize
										set uniformBold to charBold
										set uniformTextColor to charFontColor
									else if (charFontName as text) is not (uniformFontName as text) or charFontSize is not uniformFontSize or charBold is not uniformBold or charFontColor is not uniformTextColor then
										error "Mixed character fonts are outside this adapter's supported measurements: " & shapeName number 77
									end if
									set measuredCharacterCount to measuredCharacterCount + 1
								end if
							end repeat
							if measuredCharacterCount is 0 then error "Text range contains no measurable characters: " & shapeName number 78
							set fontNameValue to uniformFontName
							set fontSizeValue to uniformFontSize
							set boldValue to uniformBold
							set textColorValue to uniformTextColor
							end if
						end if

						set fontJSON to "null"
						if textValue is not missing value and textValue is not "" then set fontJSON to "{\"name\":" & my jsonNullableString(fontNameValue) & ",\"size_pt\":" & my jsonNullableNumber(fontSizeValue) & ",\"bold\":" & my jsonNullableBoolean(boldValue) & "}"
						set rowJSON to "{\"slide_index\":" & slideIndex & ",\"shape_name\":" & my jsonString(shapeName) & ",\"shape_index\":" & shapeIndex & ",\"text\":" & my jsonNullableContent(textValue) & ",\"shape_frame\":{\"left\":" & my jsonNumber(frameLeft) & ",\"top\":" & my jsonNumber(frameTop) & ",\"width\":" & my jsonNumber(frameWidth) & ",\"height\":" & my jsonNumber(frameHeight) & ",\"rotation_degrees\":" & my jsonNumber(shapeRotation) & "},\"text_bounds\":" & my jsonRawOrNull(textBoundsValue) & ",\"fill\":" & fillJSON & ",\"text_color\":" & my jsonRGB(textColorValue) & ",\"margins\":" & my jsonRawOrNull(marginsValue) & ",\"font\":" & fontJSON & "}"
						if outputRows is not "" then set outputRows to outputRows & ","
						set outputRows to outputRows & rowJSON
					end if
				end repeat
			end if
		end repeat
	end tell

	if visibleSlideCount is 0 then error "Presentation has no visible slides" number 72
	if outputRows is "" then error "Presentation has no visible shapes to measure" number 73
	return "{\"schema\":\"pptxgengo.compose-text-measurement.v2\",\"presentation\":" & my jsonString(presentationName) & ",\"visible_slide_count\":" & visibleSlideCount & ",\"coordinates\":\"raw PowerPoint scripting object units; AppleScript dictionary does not specify units\",\"text_bounds_source\":\"PowerPoint text range bounds for the complete text range; no per-character sampling\",\"color_components\":\"PowerPoint AppleScript RGB list order as returned; integer components\",\"rotation_handled\":false,\"measurements\":[" & outputRows & "]}"
end run

on absoluteValue(valueNumber)
	if valueNumber < 0 then return 0 - valueNumber
	return valueNumber
end absoluteValue

on jsonNumber(valueNumber)
	return valueNumber as text
end jsonNumber

on jsonNullableNumber(valueNumber)
	if valueNumber is missing value then return "null"
	return my jsonNumber(valueNumber)
end jsonNullableNumber

on jsonNullableBoolean(valueBoolean)
	if valueBoolean is missing value then return "null"
	return my jsonBoolean(valueBoolean)
end jsonNullableBoolean

on jsonBoolean(valueBoolean)
	if valueBoolean then return "true"
	return "false"
end jsonBoolean

on jsonRGB(rgbList)
	if rgbList is missing value then return "null"
	if (count of rgbList) is not 3 then error "Expected an RGB triplet" number 80
	return "[" & my jsonNumber(item 1 of rgbList) & "," & my jsonNumber(item 2 of rgbList) & "," & my jsonNumber(item 3 of rgbList) & "]"
end jsonRGB

on jsonNullableString(valueText)
	if valueText is missing value then return "null"
	return my jsonString(valueText as text)
end jsonNullableString

on jsonNullableContent(valueText)
	if valueText is missing value or valueText is "" then return "null"
	return my jsonString(valueText as text)
end jsonNullableContent

on jsonRawOrNull(rawValue)
	if rawValue is missing value then return "null"
	return rawValue
end jsonRawOrNull

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
