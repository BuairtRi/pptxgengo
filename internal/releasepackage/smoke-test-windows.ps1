param([string]$Out, [switch]$Native)
$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
# Load only the required Windows PowerShell modules from this host's PSHOME.
# Explicit paths avoid global module discovery and PowerShell 7/user modules.
Import-Module ([IO.Path]::Combine($PSHOME, 'Modules\Microsoft.PowerShell.Management\Microsoft.PowerShell.Management.psd1')) -ErrorAction Stop
Import-Module ([IO.Path]::Combine($PSHOME, 'Modules\Microsoft.PowerShell.Utility\Microsoft.PowerShell.Utility.psd1')) -ErrorAction Stop
$cli = Join-Path $PSScriptRoot 'bin\pptxgengo.exe'
if (-not $Out) { $Out = Join-Path $env:TEMP ('pptxgengo Windows smoke ' + (Get-Date -Format 'yyyyMMdd-HHmmss-fff')) }
if (Test-Path -LiteralPath $Out) { throw 'Smoke-test -Out must be a new directory.' }
[void][IO.Directory]::CreateDirectory($Out)
$utf8 = New-Object System.Text.UTF8Encoding($false)
$steps = New-Object System.Collections.Generic.List[object]
$script:index = 0
function Invoke-Cli([string[]]$Arguments) {
    $script:index++
    $log = Join-Path $Out ('{0:D2}.json' -f $script:index)
    $errorLog = $log + '.stderr.txt'
    # Windows PowerShell 5.1 turns native stderr into ErrorRecords. Density and
    # preview warnings must be logged without turning an exit-0 command into a
    # terminating PowerShell error. Use the native exit status as the result.
    try {
        $ErrorActionPreference = 'Continue'
        $output = & $cli @Arguments 2> $errorLog
        $exitCode = $LASTEXITCODE
    } finally { $ErrorActionPreference = 'Stop' }
    $text = ($output -join [Environment]::NewLine)
    [IO.File]::WriteAllText($log, $text, $utf8)
    $steps.Add(@{ command = @($Arguments); exit_code = $exitCode; output = [IO.Path]::GetFileName($log); stderr = [IO.Path]::GetFileName($errorLog) })
    if ($exitCode -ne 0) { throw ('CLI failed: ' + ($Arguments -join ' ') + '; see ' + $errorLog) }
    return $text
}
$core = 'not_started'
$nativeStatus = 'not_requested'
$failure = $null
$project = Join-Path $Out 'editable deck with spaces'
try {
    [void](Invoke-Cli @('--version'))
    [void](Invoke-Cli @('paths'))
    [void](Invoke-Cli @('design', 'library-find', '--query', 'pillars', '--kinds', 'template', '--summary'))
    [void](Invoke-Cli @('design', 'library-inspect', '--id', 'lifecycle/three-phases', '--summary'))
    [void](Invoke-Cli @('design', 'library-preview', '--id', 'lifecycle/three-phases'))
    Copy-Item -LiteralPath (Join-Path $PSScriptRoot 'examples\deck-project') -Destination $project -Recurse
    [void](Invoke-Cli @('design', 'project', 'init', '--project', $project))
    [void](Invoke-Cli @('design', 'project', 'check', '--project', $project))
    [void](Invoke-Cli @('design', 'project', 'split', '--project', $project))
    [void](Invoke-Cli @('design', 'project', 'section', 'add', '--project', $project, '--id', 'test-section', '--title', 'Windows testing', '--before', 'local-composition'))
    $build = (Invoke-Cli @('design', 'project', 'build', '--project', $project)) | ConvertFrom-Json
    $buildDir = Join-Path $project ('builds\' + $build.build_id)
    [void](Invoke-Cli @('design', 'project', 'measure', '--report', (Join-Path $buildDir 'layout-report.json')))
    [void](Invoke-Cli @('design', 'project', 'migrate', '--project', $project, '--dry-run'))
    [void](Invoke-Cli @('design', 'project', 'export', '--project', $project, '--mode', 'client', '--out', (Join-Path $Out 'client.zip')))
    $core = 'passed'
    Write-Host ('Core checks passed. Open the generated PowerPoint: ' + (Join-Path $buildDir 'deck.pptx'))
    if ($Native) {
        $nativeStatus = 'running'
        [void](Invoke-Cli @('design', 'render-doctor', '--json', '--timeout', '60s'))
        $renderOut = Join-Path $Out 'native review'
        [void](Invoke-Cli @('design', 'render', '--pptx', (Join-Path $buildDir 'deck.pptx'), '--out', $renderOut, '--pdf', '--png', '--contact-sheet', '--timeout', '2m'))
        [void](Invoke-Cli @('design', 'project', 'attach-render', '--project', $project, '--render', $renderOut))
        $nativeStatus = 'passed_export_not_visually_reviewed'
        Write-Host ('Native export passed. Visually review the full-size PNGs in ' + $renderOut)
    }
} catch {
    $failure = $_.Exception.Message
    if ($core -ne 'passed') { $core = 'failed' } else { $nativeStatus = 'failed' }
} finally {
    $summary = @{ schema = 'pptxgengo.windows-smoke.v1'; core = $core; native = $nativeStatus; error = $failure; steps = @($steps.ToArray()); project = $project; os = [Environment]::OSVersion.VersionString }
    [IO.File]::WriteAllText((Join-Path $Out 'summary.json'), ($summary | ConvertTo-Json -Depth 10), $utf8)
    Write-Host ('Results: ' + $Out + '\summary.json')
}
if ($null -ne $failure) { throw $failure }
