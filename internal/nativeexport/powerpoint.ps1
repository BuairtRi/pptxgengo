param([Parameter(Mandatory=$true)][string]$Config)
$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
[Console]::OutputEncoding = New-Object System.Text.UTF8Encoding($false)
$request = Get-Content -LiteralPath $Config -Raw -Encoding UTF8 | ConvertFrom-Json
$app = $null
$deck = $null
$printRange = $null
$failure = $null
$result = $null
function Same-Path([string]$a, [string]$b) {
    return [StringComparer]::OrdinalIgnoreCase.Equals([IO.Path]::GetFullPath($a), [IO.Path]::GetFullPath($b))
}
try {
    if (-not [Environment]::UserInteractive) { throw 'Run from a signed-in Windows desktop, not a service.' }
    $app = New-Object -ComObject PowerPoint.Application
    if ($request.Action -eq 'close') {
        # Never close by name or quit the application. Only this exact task path.
        for ($i = $app.Presentations.Count; $i -ge 1; $i--) {
            $candidate = $app.Presentations.Item($i)
            try {
                if (Same-Path $candidate.FullName $request.PPTX) {
                    $candidate.Saved = -1
                    $candidate.Close()
                }
            } finally { [void][Runtime.InteropServices.Marshal]::ReleaseComObject($candidate) }
        }
        for ($i = $app.ProtectedViewWindows.Count; $i -ge 1; $i--) {
            $view = $app.ProtectedViewWindows.Item($i)
            try {
                if (Same-Path $view.Presentation.FullName $request.PPTX) { $view.Close() }
            } finally { [void][Runtime.InteropServices.Marshal]::ReleaseComObject($view) }
        }
        $result = @{ closed = $true }
    } else {
        Add-Type -AssemblyName System.Drawing
        $fonts = New-Object System.Drawing.Text.InstalledFontCollection
        try {
            $names = @($fonts.Families | ForEach-Object { $_.Name })
            $missing = @('IBM Plex Sans', 'IBM Plex Mono' | Where-Object { $_ -notin $names })
        } finally { $fonts.Dispose() }
        if ($request.RequireFonts -and $missing.Count -gt 0) {
            throw ('Missing fonts: ' + ($missing -join ', ') + '. Install the bundled fonts, restart PowerPoint, and rerun.')
        }
        # msoTrue read-only; retain the exact task filename; hide its window.
        $deck = $app.Presentations.Open($request.PPTX, -1, 0, 0)
        if (-not (Same-Path $deck.FullName $request.PPTX)) { throw 'PowerPoint opened an unexpected presentation path.' }
        # PDF, print intent, slide pages, omit hidden slides, all visible slides.
        # Supply a real COM PrintRange rather than relying on PowerShell's
        # marshaling of Nothing. ppPrintAll (1) still exports every visible slide.
        # Preserve IRM and accessibility tags.
        $printRange = $deck.PrintOptions.Ranges.Add(1, $deck.Slides.Count)
        $deck.ExportAsFixedFormat($request.PDF, 2, 2, 0, 1, 1, 0, $printRange, 1, '', $true, $true, $true, $true, $false)
        $page = 0
        $height = [Math]::Max(1, [int][Math]::Round(1600 * $deck.PageSetup.SlideHeight / $deck.PageSetup.SlideWidth))
        if ($request.PNG) { [void][IO.Directory]::CreateDirectory($request.PagesDirectory) }
        for ($i = 1; $i -le $deck.Slides.Count; $i++) {
            $slide = $deck.Slides.Item($i)
            try {
                if ($slide.SlideShowTransition.Hidden -ne -1) {
                    $page++
                    if ($request.PNG) {
                        $path = Join-Path $request.PagesDirectory ('slide-{0:D3}.png' -f $page)
                        $slide.Export($path, 'PNG', 1600, $height)
                    }
                }
            } finally { [void][Runtime.InteropServices.Marshal]::ReleaseComObject($slide) }
        }
        $result = @{ pages = $page; powerpoint_version = [string]$app.Version; missing_fonts = @($missing) }
    }
} catch { $failure = $_ }
finally {
    if ($null -ne $printRange) { [void][Runtime.InteropServices.Marshal]::ReleaseComObject($printRange) }
    if ($null -ne $deck) {
        try {
            if (-not (Same-Path $deck.FullName $request.PPTX)) { throw 'Task presentation identity changed; close is unconfirmed.' }
            $deck.Saved = -1
            $deck.Close()
        } catch { if ($null -eq $failure) { $failure = $_ } }
        [void][Runtime.InteropServices.Marshal]::ReleaseComObject($deck)
    }
    if ($null -ne $app) { [void][Runtime.InteropServices.Marshal]::ReleaseComObject($app) }
    # No Application.Quit and no process-name kill; unrelated decks stay open.
}
if ($null -ne $failure) { [Console]::Error.WriteLine($failure.ToString()); exit 1 }
$result | ConvertTo-Json -Compress -Depth 4
