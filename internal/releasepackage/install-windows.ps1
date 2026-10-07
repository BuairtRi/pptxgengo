param([string]$Destination, [switch]$SkipSkill, [switch]$SkipFonts, [switch]$NoPath, [switch]$StageOnly)
$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
# Hashes establish package consistency. Authenticate the outer signed release
# manifest and archive before invoking this script; these hashes are not a signature.
function Assert-Package([string]$Directory) {
    $root = [IO.Path]::GetFullPath($Directory).TrimEnd('\') + '\'
    $manifestPath = Join-Path $Directory 'release-manifest.json'
    if ((Get-Item -LiteralPath $Directory -Force).Attributes -band [IO.FileAttributes]::ReparsePoint) { throw 'Package root must not be a reparse point.' }
    # Enumerate one directory at a time so junctions are rejected before traversal.
    $pending = New-Object System.Collections.Generic.Queue[string]
    $pending.Enqueue($Directory)
    $files = @{}
    while ($pending.Count -gt 0) {
        foreach ($item in Get-ChildItem -LiteralPath ($pending.Dequeue()) -Force) {
            if ($item.Attributes -band [IO.FileAttributes]::ReparsePoint) { throw ('Package reparse point: ' + $item.FullName) }
            if ($item.PSIsContainer) { $pending.Enqueue($item.FullName); continue }
            $relative = $item.FullName.Substring($root.Length).Replace('\', '/')
            if ($files.ContainsKey($relative)) { throw ('Duplicate package path: ' + $relative) }
            $files[$relative] = $item.FullName
        }
    }
    $manifest = Get-Content -LiteralPath $manifestPath -Raw -Encoding UTF8 | ConvertFrom-Json
    if ($manifest.schema -ne 'pptxgengo.local-release-manifest.v1' -or $manifest.target_os -ne 'windows') { throw 'Not a Windows presentation package.' }
    if ($manifest.version -notmatch '^[A-Za-z0-9][A-Za-z0-9._-]*$') { throw 'Invalid package version.' }
    if ($manifest.target_arch -notin @('amd64', 'arm64')) { throw 'Unsupported package architecture.' }
    $nativeArchitecture = if ($env:PROCESSOR_ARCHITEW6432) { $env:PROCESSOR_ARCHITEW6432 } else { $env:PROCESSOR_ARCHITECTURE }
    $expectedArchitecture = if ($manifest.target_arch -eq 'amd64') { 'AMD64' } else { 'ARM64' }
    if ($nativeArchitecture -ne $expectedArchitecture) { throw ('Use a package for the native PC architecture: ' + $nativeArchitecture + '; selected ' + $manifest.target_arch) }
    $expected = @{}
    foreach ($entry in $manifest.files_sha256.PSObject.Properties) {
        $relative = $entry.Name
        if ($relative -match '[\\:*?"<>|]' -or $relative.StartsWith('/') -or $relative -eq 'release-manifest.json') { throw ('Unsafe package path: ' + $relative) }
        foreach ($part in $relative.Split('/')) {
            if (-not $part -or $part -in @('.', '..') -or $part.EndsWith('.') -or $part.EndsWith(' ') -or $part -match '^(CON|PRN|AUX|NUL|COM[1-9]|LPT[1-9])(\.|$)') { throw ('Unsafe package path: ' + $relative) }
        }
        if ($expected.ContainsKey($relative) -or $entry.Value -notmatch '^[a-f0-9]{64}$') { throw ('Invalid package hash entry: ' + $relative) }
        if (-not $files.ContainsKey($relative)) { throw ('Missing package file: ' + $relative) }
        if ((Get-FileHash -LiteralPath $files[$relative] -Algorithm SHA256).Hash.ToLowerInvariant() -ne $entry.Value) { throw ('Package hash mismatch: ' + $relative) }
        $expected[$relative] = $true
    }
    if ($expected.Count -ne $manifest.file_count -or $files.Count -ne ($expected.Count + 1)) { throw 'Package file inventory mismatch; extract into a clean directory.' }
    if ($manifest.selected_bundle -notmatch '^v[1-9][0-9]*$') { throw 'Invalid selected library bundle.' }
    $bundlePrefix = 'library/wm-design-system/' + $manifest.selected_bundle + '/'
    foreach ($required in @('VERSION', 'release/VERSION', 'bin/pptxgengo.exe', 'bin/pptxdesign.exe', 'bin/wmdsdocs.exe', 'wmds-docs/site/SOURCE.json', ($bundlePrefix + 'bundle.json'), ($bundlePrefix + 'library.sqlite'), ($bundlePrefix + 'catalog/design-system.html'))) {
        if (-not $expected.ContainsKey($required)) { throw ('Required package resource missing: ' + $required) }
    }
    if (-not $SkipSkill -and -not $expected.ContainsKey('skills/west-monroe-presentations/SKILL.md')) { throw 'Packaged skill missing; use -SkipSkill only if intentional.' }
    $fontPrefix = 'library/wm-design-system/' + $manifest.selected_bundle + '/fonts/'
    if (-not $SkipFonts -and @($expected.Keys | Where-Object { $_.StartsWith($fontPrefix) -and $_.EndsWith('.ttf') }).Count -eq 0) { throw 'Packaged fonts missing; use -SkipFonts only if intentional.' }
    foreach ($versionFile in @('VERSION', 'release/VERSION')) {
        if ((Get-Content -LiteralPath (Join-Path $Directory $versionFile) -Raw).Trim() -ne $manifest.version) { throw ('Package version mismatch: ' + $versionFile) }
    }
    return $manifest
}
function Assert-ToolsStart([string]$Directory, [string]$Version) {
    foreach ($tool in @('pptxgengo', 'pptxdesign', 'wmdsdocs')) {
        $process = New-Object System.Diagnostics.Process
        $process.StartInfo.FileName = Join-Path $Directory ('bin\' + $tool + '.exe')
        $process.StartInfo.Arguments = '--version'
        $process.StartInfo.UseShellExecute = $false
        $process.StartInfo.CreateNoWindow = $true
        $process.StartInfo.RedirectStandardOutput = $true
        $process.StartInfo.RedirectStandardError = $true
        try {
            [void]$process.Start()
            $stdout = $process.StandardOutput.ReadToEndAsync()
            $stderr = $process.StandardError.ReadToEndAsync()
            if (-not $process.WaitForExit(30000)) {
                $process.Kill()
                throw ('Packaged tool version check timed out: ' + $tool)
            }
            $reported = $stdout.GetAwaiter().GetResult()
            [void]$stderr.GetAwaiter().GetResult()
            if ($process.ExitCode -ne 0 -or $reported.Trim() -ne $Version) { throw ('Packaged tool did not report the expected version: ' + $tool + '. Confirm architecture and enterprise application policy.') }
        } finally { $process.Dispose() }
    }
}
$source = [IO.Path]::GetFullPath($PSScriptRoot)
Write-Host 'Checking source package...'
$sourceManifestHash = (Get-FileHash -LiteralPath (Join-Path $source 'release-manifest.json') -Algorithm SHA256).Hash
$manifest = Assert-Package $source
if ((Get-FileHash -LiteralPath (Join-Path $source 'release-manifest.json') -Algorithm SHA256).Hash -ne $sourceManifestHash) { throw 'Source manifest changed during verification.' }
if (-not $Destination) { $Destination = Join-Path $env:LOCALAPPDATA ('pptxgengo\releases\' + $manifest.version) }
$Destination = [IO.Path]::GetFullPath($Destination)
# A destination below the source would recursively copy the stage into itself.
if ($Destination.StartsWith($source.TrimEnd('\') + '\', [StringComparison]::OrdinalIgnoreCase)) { throw 'Destination must be outside the extracted package.' }
if (Test-Path -LiteralPath $Destination) { throw 'Destination already exists. Choose a new -Destination; existing releases are never overwritten.' }
$parent = Split-Path -Parent $Destination
[void][IO.Directory]::CreateDirectory($parent)
$stage = Join-Path $parent ('.pptxgengo-stage-' + [Guid]::NewGuid().ToString('N'))
try {
    Write-Host ('Staging verified package for ' + $Destination)
    Copy-Item -LiteralPath $source -Destination $stage -Recurse -Force
    $stagedManifest = Assert-Package $stage
    if ($stagedManifest.version -ne $manifest.version -or $stagedManifest.target_arch -ne $manifest.target_arch -or (Get-FileHash -LiteralPath (Join-Path $stage 'release-manifest.json') -Algorithm SHA256).Hash -ne $sourceManifestHash) { throw 'Source package changed while staging.' }
    Assert-ToolsStart $stage $manifest.version
    # Same-parent rename commits only a completely copied, checked release.
    # Directory.Move refuses to replace a concurrently created destination.
    [IO.Directory]::Move($stage, $Destination)
} finally {
    if (Test-Path -LiteralPath $stage) { Remove-Item -LiteralPath $stage -Recurse -Force }
}
if ($StageOnly) {
    Write-Host ('Verified release staged at ' + $Destination + '. PATH, skill and fonts have not been activated.')
    return
}
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
Write-Host ('CLI verified: ' + $manifest.version + ' (' + $manifest.target_arch + ')')
Write-Host 'Installation complete. Restart Codex to load the skill; use a new terminal to pick up PATH.'
Write-Host ('Getting started: open "' + (Join-Path $Destination 'guides\00-start.html') + '" in your browser.')
Write-Host ('Smoke test: powershell -NoProfile -ExecutionPolicy Bypass -File "' + (Join-Path $Destination 'smoke-test-windows.ps1') + '"')
