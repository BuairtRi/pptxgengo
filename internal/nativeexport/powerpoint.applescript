-- Only the exact task-copy file identity is eligible for export or cleanup.
-- Do not use the active document, a display name, or a guessed presentation.
use framework "Foundation"
use scripting additions
on run argv
 set sourceFile to POSIX file (item 1 of argv)
 set pdfFile to POSIX file (item 2 of argv)
 set secondsAllowed to (item 4 of argv) as integer
 set closeOnly to false
 set probeOnly to false
 if (count of argv) > 4 then set closeOnly to (item 5 of argv is "close")
 if (count of argv) > 4 then set probeOnly to (item 5 of argv is "probe")
 set expectedIdentity to my fileIdentity(sourceFile)
 if (count of argv) > 5 then
  set retainedIdentity to my fileIdentity(POSIX file (item 6 of argv))
  if retainedIdentity is "" or retainedIdentity is not expectedIdentity then error "identity_changed: staged task inode differs from retained task identity" number 68
  set expectedIdentity to retainedIdentity
 end if
 if expectedIdentity is "" then error "file_access_denied: cannot stat task copy" number 66
 with timeout of secondsAllowed seconds
  tell application "/Applications/Microsoft PowerPoint.app"
   if not closeOnly then
    try
     with timeout of 12 seconds
      open sourceFile
     end timeout
    on error messageText number errorNumber
     if errorNumber is -1712 then error "open_identity_timeout: PowerPoint open did not answer; inspect PowerPoint for a blocking dialog or file-access prompt" number 67
     error messageText number errorNumber
    end try
   end if
   set matchedIndex to 0
   repeat with attempt from 1 to 40
    set matchCount to 0
    set matchedIndex to 0
    with timeout of 3 seconds
     repeat with presentationIndex from 1 to (count of presentations)
      set candidate to presentation presentationIndex
      if my fileIdentity(full name of candidate) is expectedIdentity then
       set matchCount to matchCount + 1
       set matchedIndex to presentationIndex as integer
      end if
     end repeat
    end timeout
    if matchCount > 1 then error "identity_ambiguous: more than one open presentation identifies the exact task-copy file" number 65
    if matchCount is 1 then exit repeat
    if closeOnly and matchCount is 0 then return
    delay 0.25
   end repeat
   if matchCount is not 1 then error "open_identity_timeout: no presentation identified the exact task-copy file within 10 seconds" number 67
   set taskPresentation to presentation matchedIndex
   if closeOnly or probeOnly then
    if my fileIdentity(full name of taskPresentation) is not expectedIdentity then error "identity_changed: task presentation index changed before cleanup" number 68
    close taskPresentation saving no
    return
   end if
   try
    if my fileIdentity(full name of taskPresentation) is not expectedIdentity then error "identity_changed: task presentation index changed before export" number 68
    save taskPresentation in pdfFile as save as PDF
   on error messageText number errorNumber
    try
     if my fileIdentity(full name of taskPresentation) is not expectedIdentity then error "identity_changed: task presentation index changed before error cleanup" number 68
     close taskPresentation saving no
    end try
    error messageText number errorNumber
   end try
   if my fileIdentity(full name of taskPresentation) is not expectedIdentity then error "identity_changed: task presentation index changed before close" number 68
   close taskPresentation saving no
  end tell
 end timeout
end run

on fileIdentity(livePath)
 try
  set filePath to my normalizePath(livePath)
  -- Resolve symlinks, including /tmp versus /private/tmp, then compare device/inode.
  set resolvedPath to ((current application's NSString's stringWithString:filePath)'s stringByResolvingSymlinksInPath()) as text
  set attrs to current application's NSFileManager's defaultManager()'s attributesOfItemAtPath:resolvedPath |error|:(missing value)
  if attrs is missing value then return ""
  set deviceNumber to attrs's objectForKey:(current application's NSFileSystemNumber)
  set inodeNumber to attrs's objectForKey:(current application's NSFileSystemFileNumber)
  return ((deviceNumber's stringValue()) as text) & ":" & ((inodeNumber's stringValue()) as text)
 on error
  return ""
 end try
end fileIdentity

on normalizePath(livePath)
 try
  return POSIX path of (livePath as alias)
 on error
  return livePath as text
 end try
end normalizePath
