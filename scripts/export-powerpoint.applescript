-- Export through PowerPoint's document object model, never UI keystrokes.
-- Work on task copies; the source document is not saved or closed.
on run argv
	if (count argv) is not 2 then error "Usage: source-copy.pptx output.pdf"
	set sourcePath to item 1 of argv
	set pdfPath to item 2 of argv
	set sourceFile to POSIX file sourcePath
	set sourceName to name of (info for sourceFile)
	with timeout of 120 seconds
		tell application "Microsoft PowerPoint"
			open sourceFile
			set targetPresentation to presentation sourceName
			save targetPresentation in (POSIX file pdfPath) as save as PDF
		end tell
	end timeout
	return pdfPath
end run
