param([string]$Destination, [switch]$SkipSkill, [switch]$SkipFonts, [switch]$NoPath)
$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
$source = $PSScriptRoot
$manifest = Get-Content -LiteralPath (Join-Path $source 'release-manifest.json') -Raw -Encoding UTF8 | ConvertFrom-Json
if ($manifest.schema -ne 'pptxgengo.local-release-manifest.v1' -or $manifest.target_os -ne 'windows') { throw 'Not a Windows pptxgengo package.' }
if ($manifest.version -notmatch '^[A-Za-z0-9._-]+$') { throw 'Invalid package version.' }
Write-Host 'Checking packaged files...'
$root = [IO.Path]::GetFullPath($source).TrimEnd('\') + '\'
$count = 0
foreach ($entry in $manifest.files_sha256.PSObject.Properties) {
    $path = [IO.Path]::GetFullPath((Join-Path $source $entry.Name))
    if (-not $path.StartsWith($root, [StringComparison]::OrdinalIgnoreCase)) { throw ('Unsafe package path: ' + $entry.Name) }
    if ((Get-FileHash -LiteralPath $path -Algorithm SHA256).Hash.ToLowerInvariant() -ne $entry.Value) { throw ('Package hash mismatch: ' + $entry.Name) }
    $count++
}
if ($count -ne $manifest.file_count) { throw 'Package file count mismatch.' }
if (-not $Destination) { $Destination = Join-Path $env:LOCALAPPDATA ('pptxgengo\releases\' + $manifest.version) }
if (Test-Path -LiteralPath $Destination) { throw 'Destination already exists. Choose a new -Destination; existing releases are never overwritten.' }
[void][IO.Directory]::CreateDirectory((Split-Path -Parent $Destination))
Write-Host ('Installing to ' + $Destination)
Copy-Item -LiteralPath $source -Destination $Destination -Recurse
$bin = Join-Path $Destination 'bin'
if (-not $NoPath) {
    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    $parts = @($userPath -split ';' | Where-Object { $_ })
    if ($bin -notin $parts) { [Environment]::SetEnvironmentVariable('Path', (($parts + $bin) -join ';'), 'User') }
    $env:PATH = $bin + ';' + $env:PATH
}
if (-not $SkipSkill) {
    $codexRoot = if ($env:CODEX_HOME) { $env:CODEX_HOME } else { Join-Path $env:USERPROFILE '.codex' }
    $skillRoot = Join-Path $codexRoot 'skills'
    [void][IO.Directory]::CreateDirectory($skillRoot)
    $skill = Join-Path $skillRoot 'west-monroe-presentations'
    if (Test-Path -LiteralPath $skill) {
        $backup = $skill + '.backup-' + (Get-Date -Format 'yyyyMMdd-HHmmss-fff')
        Move-Item -LiteralPath $skill -Destination $backup
        Write-Host ('Previous skill preserved at ' + $backup)
    }
    Copy-Item -LiteralPath (Join-Path $Destination 'skills\west-monroe-presentations') -Destination $skill -Recurse
}
if (-not $SkipFonts) {
    try {
        $fontFolder = Join-Path $env:LOCALAPPDATA 'Microsoft\Windows\Fonts'
        [void][IO.Directory]::CreateDirectory($fontFolder)
        $registry = 'HKCU:\Software\Microsoft\Windows NT\CurrentVersion\Fonts'
        [void](New-Item -Path $registry -Force)
        Add-Type -TypeDefinition @'
using System;
using System.Runtime.InteropServices;
public static class PptxgengoFontInstall {
 [DllImport("gdi32.dll", CharSet=CharSet.Unicode)] public static extern int AddFontResourceEx(string name, uint flags, IntPtr reserved);
 [DllImport("user32.dll", CharSet=CharSet.Unicode)] public static extern IntPtr SendMessageTimeout(IntPtr window, uint message, IntPtr wParam, IntPtr lParam, uint flags, uint timeout, out IntPtr result);
}
'@
        $fontSource = Join-Path $Destination ('library\wm-design-system\' + $manifest.selected_bundle + '\fonts')
        foreach ($font in Get-ChildItem -LiteralPath $fontSource -Filter '*.ttf') {
            $target = Join-Path $fontFolder $font.Name
            Copy-Item -LiteralPath $font.FullName -Destination $target -Force
            [void](New-ItemProperty -Path $registry -Name ($font.BaseName + ' (TrueType)') -Value $target -PropertyType String -Force)
            [void][PptxgengoFontInstall]::AddFontResourceEx($target, 0, [IntPtr]::Zero)
        }
        $notify = [IntPtr]::Zero
        [void][PptxgengoFontInstall]::SendMessageTimeout([IntPtr]0xffff, 0x001d, [IntPtr]::Zero, [IntPtr]::Zero, 2, 2000, [ref]$notify)
        Write-Host 'IBM Plex fonts installed for this user. Restart PowerPoint to load them.'
    } catch { Write-Warning ('CLI installed, but font installation needs attention: ' + $_.Exception.Message + '. Install the bundled TTFs manually before native review.') }
}
& (Join-Path $bin 'pptxgengo.exe') --version
if ($LASTEXITCODE -ne 0) { throw 'Installed CLI did not start. Confirm the PC architecture and enterprise application policy.' }
Write-Host 'Installation complete. Restart Codex to load the skill; use a new terminal to pick up PATH.'
Write-Host ('Smoke test: powershell -NoProfile -ExecutionPolicy Bypass -File "' + (Join-Path $Destination 'smoke-test-windows.ps1') + '"')
