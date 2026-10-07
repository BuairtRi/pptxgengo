$ErrorActionPreference = 'Stop'
$env:CGO_ENABLED = '0'
$env:WMDS_BRANDING_ROOT = Join-Path $env:CI_PROJECT_DIR '.cache\empty-branding'
$env:PPTXGENGO_NATIVE_LIVE_OUT = ''
$env:PPTXGENGO_INSTALL_PROCESS_COMMIT = $env:CI_COMMIT_SHA
$env:PPTXGENGO_SEARCH_BENCH_COMMIT = $env:CI_COMMIT_SHA
$env:PPTXGENGO_INSTALL_PROCESS_OUT = Join-Path $env:CI_PROJECT_DIR '.cache\installation-process'
$env:PPTXGENGO_EMBED_MODEL_DIR = Join-Path $env:CI_PROJECT_DIR '.cache\offline-model'
$env:PPTXGENGO_SEARCH_BENCH_OUT = Join-Path $env:CI_PROJECT_DIR '.cache\search-performance'
[void](New-Item -ItemType Directory -Path $env:WMDS_BRANDING_ROOT -Force)
function Run-Go {
    param([string[]]$CommandArgs)
    & go @CommandArgs
    if ($LASTEXITCODE -ne 0) { throw "Go command failed: $CommandArgs" }
}
if ((go env GOHOSTOS) -ne 'windows' -or (go env GOHOSTARCH) -ne $env:PPTXGENGO_WINDOWS_ARCH) { throw 'Declared native Windows architecture required.' }
foreach ($script in @('internal/nativeexport/powerpoint.ps1', 'internal/nativeexport/roundtrip.ps1', 'internal/releasepackage/install-windows.ps1', 'internal/releasepackage/smoke-test-windows.ps1')) {
    $tokens = $null; $errors = $null
    [void][System.Management.Automation.Language.Parser]::ParseFile((Join-Path $PWD $script), [ref]$tokens, [ref]$errors)
    if ($errors.Count -gt 0) { throw ($errors | Out-String) }
}
Run-Go -CommandArgs @('test','-count=1','-timeout=150s','-run','TestKeyword|TestDiscoveryRead','./internal/wmdesign')
Run-Go -CommandArgs @('build','-trimpath','-o','bin/','./cmd/pptxgengo','./cmd/pptxdesign','./cmd/wmdsdocs')
& .\bin\pptxgengo.exe --version
if ($LASTEXITCODE -ne 0) { throw 'CLI version failed.' }
& .\bin\pptxgengo.exe paths
if ($LASTEXITCODE -ne 0) { throw 'CLI paths failed.' }
Run-Go -CommandArgs @('test','-short','-count=1','-timeout=150s','./cmd/pptxgengo','./cmd/pptxdesign','./internal/deckproject','./internal/library','./internal/releasepackage','./internal/textlayout','./pptx','./scripts/cmd/search-benchmark')
Run-Go -CommandArgs @('test','-short','-count=1','-timeout=150s','-v','-run','^(TestWindows|TestSignedNativeReceiptRejectsHandmadeAndTamperedEvidence|TestConcurrentNative)','./internal/nativeexport')
Run-Go -CommandArgs @('test','-count=1','-timeout=150s','./internal/installstate','./internal/finishedslide')
Run-Go -CommandArgs @('test','-count=1','-timeout=8m','-run','^TestInstallationRealToolProcesses$','-v','./internal/installstate')
Copy-Item internal/releasepackage/smoke-test-windows.ps1 .
powershell -NoProfile -ExecutionPolicy Bypass -File .\smoke-test-windows.ps1 -Out "$env:CI_PROJECT_DIR\.cache\windows-smoke\path with spaces"
if ($LASTEXITCODE -ne 0) { throw 'Relocated CLI smoke failed.' }
Run-Go -CommandArgs @('run','./cmd/pptxdesign','library-model','--download','--out',$env:PPTXGENGO_EMBED_MODEL_DIR)
Run-Go -CommandArgs @('test','-count=1','-timeout=5m','./internal/localembed','./internal/modelpackage','./scripts/cmd/release-ci')
Run-Go -CommandArgs @('test','-count=1','-timeout=8m','-run','^TestPinnedLibraryEmbeddingsEndToEnd$','./internal/wmdesign')
Run-Go -CommandArgs @('test','-count=1','-timeout=7m','-run','^TestPinnedLibrarySearchPerformance$','-v','./internal/wmdesign')
