-- Export through PowerPoint's document object model, never UI keystrokes.
-- Work on task copies; the source document is not saved or closed.
on run argv
	if (count argv) is not 2 then error "Usage: source-copy.pptx output.pdf"
	set sourcePath to item 1 of argv
	set pdfPath to item 2 of argv
	-- Keep previous evidence and prevent a stale PDF from passing this export.
	do shell script "if [ -e " & quoted form of pdfPath & " ]; then echo 'PDF output already exists; choose a new path' >&2; exit 86; fi"
	set sourceFile to POSIX file sourcePath
	set sourceName to name of (info for sourceFile)
	with timeout of 120 seconds
		tell application "Microsoft PowerPoint"
			open sourceFile
			set targetPresentation to presentation sourceName
			save targetPresentation in (POSIX file pdfPath) as save as PDF
		end tell
	end timeout
	-- PowerPoint can return successfully without producing a file after a
	-- denied or unresolved access prompt. Require a real PDF before reporting it.
	set pdfHeader to do shell script "/usr/bin/head -c 5 " & quoted form of pdfPath
	if pdfHeader is not "%PDF-" then error "PowerPoint did not create a valid PDF at " & pdfPath number 86
	return pdfPath
end run
