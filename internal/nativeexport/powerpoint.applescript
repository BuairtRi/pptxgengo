-- Only the exact task-copy path is eligible for export or cleanup.
-- An absolute application path avoids ambiguous Launch Services name lookup.
on run argv
 set sourceFile to POSIX file (item 1 of argv)
 set pdfFile to POSIX file (item 2 of argv)
 set secondsAllowed to (item 4 of argv) as integer
 set closeOnly to false
 if (count of argv) > 4 then set closeOnly to (item 5 of argv is "close")
 set expectedPath to my normalizePath(sourceFile)
 with timeout of secondsAllowed seconds
  tell application "/Applications/Microsoft PowerPoint.app"
   if not closeOnly then open sourceFile
   set matchCount to 0
   set matchedIndex to 0
   repeat with presentationIndex from 1 to (count of presentations)
    set candidate to presentation presentationIndex
    if my normalizePath(full name of candidate) is expectedPath then
     set matchCount to matchCount + 1
     set matchedIndex to presentationIndex as integer
    end if
   end repeat
   if closeOnly and matchCount is 0 then return
   if matchCount is not 1 then error "Expected exactly one open presentation at the task-copy path" number 65
   set taskPresentation to presentation matchedIndex
   if closeOnly then
    close taskPresentation saving no
    return
   end if
   try
    save taskPresentation in pdfFile as save as PDF
   on error messageText number errorNumber
    try
     close taskPresentation saving no
    end try
    error messageText number errorNumber
   end try
   close taskPresentation saving no
  end tell
 end timeout
end run

on normalizePath(livePath)
 try
  return POSIX path of (livePath as alias)
 on error
  return livePath as text
 end try
end normalizePath
