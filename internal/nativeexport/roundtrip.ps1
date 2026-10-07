param([Parameter(Mandatory=$true)][string]$Config)
$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
[Console]::OutputEncoding = New-Object System.Text.UTF8Encoding($false)
$request = Get-Content -LiteralPath $Config -Raw -Encoding UTF8 | ConvertFrom-Json
$app = $null; $deck = $null; $failure = $null; $result = $null
$ownedPaths = @($request.Input, $request.SavedAs, $request.Edited)
function Same-Path([string]$a, [string]$b) {
    return [StringComparer]::OrdinalIgnoreCase.Equals([IO.Path]::GetFullPath($a), [IO.Path]::GetFullPath($b))
}
function Owned-Path([string]$path) {
    foreach ($owned in $ownedPaths) { if (Same-Path $path $owned) { return $true } }
    return $false
}
function Release-Com($object) {
    if ($null -ne $object -and [Runtime.InteropServices.Marshal]::IsComObject($object)) { [void][Runtime.InteropServices.Marshal]::ReleaseComObject($object) }
}
function Tag($object, [string]$name) {
    $tags = $object.Tags
    try { return [string]$tags.Item($name) } finally { Release-Com $tags }
}
# These actions are selected through tags, never shape names, current text,
# slide indices or geometry. Names are only written to the human fixture plan.
function Find-Shapes($shapes, [string]$token, $matches, [int]$depth) {
    if ($depth -gt 16) { throw 'Fixture group depth exceeds qualification bound.' }
    if ($shapes.Count -gt 10000) { throw 'Fixture shape count exceeds qualification bound.' }
    for ($i = 1; $i -le $shapes.Count; $i++) {
        $shape = $shapes.Item($i)
        $retained = $false
        try {
            if ((Tag $shape 'PPTXGENGO_SHAPE') -ceq $token) { [void]$matches.Add($shape); $retained = $true }
            if ($shape.Type -eq 6) {
                $children = $shape.GroupItems
                try { Find-Shapes $children $token $matches ($depth + 1) } finally { Release-Com $children }
            }
        } finally { if (-not $retained) { Release-Com $shape } }
    }
}
function Find-Slide([string]$token) {
    $matches = New-Object 'System.Collections.Generic.List[object]'
    for ($i = 1; $i -le $deck.Slides.Count; $i++) {
        $slide = $deck.Slides.Item($i)
        if ((Tag $slide 'PPTXGENGO_SLIDE') -ceq $token) { [void]$matches.Add($slide) } else { Release-Com $slide }
    }
    if ($matches.Count -ne 1) {
        foreach ($slide in $matches) { Release-Com $slide }
        throw 'Missing or duplicated fixture slide tag.'
    }
    return $matches[0]
}
function Close-Own($presentation) {
    if (-not (Owned-Path $presentation.FullName)) { throw 'Fixture presentation changed identity; close unconfirmed.' }
    $presentation.Saved = -1
    $presentation.Close()
}
try {
    if (-not [Environment]::UserInteractive) { throw 'Run qualification from a signed-in Windows desktop, not a service.' }
    $app = New-Object -ComObject PowerPoint.Application
    if ($request.Action -eq 'close') {
        for ($i = $app.Presentations.Count; $i -ge 1; $i--) {
            $candidate = $app.Presentations.Item($i)
            try { if (Owned-Path $candidate.FullName) { Close-Own $candidate } } finally { Release-Com $candidate }
        }
        for ($i = $app.ProtectedViewWindows.Count; $i -ge 1; $i--) {
            $view = $app.ProtectedViewWindows.Item($i)
            try { if (Owned-Path $view.Presentation.FullName) { $view.Close() } } finally { Release-Com $view }
        }
        $result = @{ closed = $true }
    } elseif ($request.Action -eq 'edit') {
        foreach ($path in @($request.SavedAs,$request.Edited)) { if (Test-Path -LiteralPath $path) { throw 'Fixture Save As destination already exists.' } }
        for ($i = 1; $i -le $app.Presentations.Count; $i++) {
            $candidate = $app.Presentations.Item($i)
            try { if (Owned-Path $candidate.FullName) { throw 'Fixture is already open; refusing to reuse a presentation.' } } finally { Release-Com $candidate }
        }
        # Writable, retain filename, visible window in the interactive desktop.
        $deck = $app.Presentations.Open($request.Input, 0, 0, -1)
        if (-not (Same-Path $deck.FullName $request.Input)) { throw 'PowerPoint opened an unexpected fixture path.' }
        if ((Tag $deck 'PPTXGENGO_SCHEMA') -cne '1' -or (Tag $deck 'PPTXGENGO_DECK') -cne $request.deck_token -or (Tag $deck 'PPTXGENGO_BUILD') -cne $request.build_token) { throw 'Fixture deck/build tags do not match.' }
        # ppSaveAsOpenXMLPresentation = 24. Preserve the first native Save As
        # separately so tag survival is checked before text edits/reordering.
        $deck.SaveAs($request.SavedAs, 24)
        if (-not (Same-Path $deck.FullName $request.SavedAs)) { throw 'Save As did not select the requested new filename.' }
        foreach ($edit in $request.edits) {
            $slide = Find-Slide $edit.slide_token
            $matches = New-Object 'System.Collections.Generic.List[object]'
            try {
                if ((Tag $slide 'PPTXGENGO_BUILD') -cne $request.build_token) { throw 'Fixture slide build tag differs.' }
                $shapes = $slide.Shapes
                try { Find-Shapes $shapes $edit.shape_token $matches 0 } finally { Release-Com $shapes }
                if ($matches.Count -ne 1) { throw 'Missing or duplicated fixture shape tag.' }
                $shape = $matches[0]
                if ((Tag $shape 'PPTXGENGO_BUILD') -cne $request.build_token -or $shape.HasTextFrame -ne -1) { throw 'Fixture shape is not the expected tagged text object.' }
                $frame = $shape.TextFrame
                try {
                    $range = $frame.TextRange
                    try {
                        if ([string]$range.Text -cne $edit.before) { throw 'Fixture text differs from the receipt-pinned baseline.' }
                        $range.Text = $edit.after
                        if ([string]$range.Text -cne $edit.after) { throw 'PowerPoint did not retain exact edited text.' }
                    } finally { Release-Com $range }
                } finally { Release-Com $frame }
            } finally { foreach ($shape in $matches) { Release-Com $shape }; Release-Com $slide }
        }
        $move = Find-Slide $request.move_slide_token
        try {
            if ($move.SlideIndex -eq 1) { throw 'Fixture reorder action would be a no-op.' }
            $move.MoveTo(1)
        } finally { Release-Com $move }
        $deck.SaveAs($request.Edited,24)
        if (-not (Same-Path $deck.FullName $request.Edited)) { throw 'Edited Save As did not select the requested filename.' }
        $result = @{ schema = 'pptxgengo.windows-roundtrip-execution.v1'; powerpoint_version = [string]$app.Version; os_version = [Environment]::OSVersion.VersionString; saved_as = [string]$request.SavedAs; edited = [string]$request.Edited; closed = $true }
    } else { throw 'Unknown fixture action.' }
} catch { $failure = $_ }
finally {
    if ($null -ne $deck) {
        try { Close-Own $deck } catch { if ($null -eq $failure) { $failure = $_ } }
        Release-Com $deck
    }
    Release-Com $app
    # Never quit PowerPoint or kill it by process name; unrelated decks stay open.
}
if ($null -ne $failure) { [Console]::Error.WriteLine($failure.ToString()); exit 1 }
$result | ConvertTo-Json -Compress -Depth 4
