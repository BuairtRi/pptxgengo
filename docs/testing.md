# Testing lanes

Run `make test` for everyday development. It runs the repository suite with Go's
`-short` mode, `-count=1`, and a 150-second per-package timeout. It should fit
within a normal short development loop. Only specifically marked exhaustive or
resource-dependent integration tests skip in short mode; ordinary unit and
focused regression coverage still runs.

## Lanes

| Command | Coverage | Environment |
| --- | --- | --- |
| `make test` | Normal short suite | Go 1.27.1+; no PowerPoint required |
| `make test-race` | Race checks for project source mutation guards, the full hermetic native-export package, library catalog cache, and concurrent PPTX serialization | CGO-enabled Go toolchain |
| `make test-integration` | Full normal suite, including exhaustive catalogs and final relocated projects | Registered private branding assets; see below |
| `make test-race-full` | Full suite under the race detector, including tests skipped by `-short` | Opt-in; CGO-enabled Go; 10-minute per-package ceiling |
| `make test-native` | Opt-in live PowerPoint smoke test | macOS desktop session, Microsoft PowerPoint, Swift/PDFKit, and a new `PPTXGENGO_NATIVE_LIVE_OUT` path |

The timeout passed to `go test` applies to each package test process, not to the
entire Make target. `make test-integration` and `make test-race-full` are bounded
by their per-package ceilings; a timeout is a test failure and should be
investigated by rerunning the affected package. The exhaustive race lane is not
a routine developer check and should not be described as passing unless it has
actually completed.
The fast and focused-race targets use a 150-second per-package ceiling;
exhaustive race checks use 10 minutes per package.

## CI policy

GitHub Actions runs `make test` and `make test-race` on hosted macOS for pushes
and pull requests, plus its scheduled run. It sets an empty branding root and
clears `PPTXGENGO_NATIVE_LIVE_OUT`, so these checks are headless, require no
private artwork, and do not open PowerPoint. The native-export package still
tests worker startup/cancellation and its hermetic export harness.

GitLab runs the same targets in a Go 1.27.1 Linux container. It is also
headless and sets an empty branding root. Full `test-integration` and
`test-race-full` jobs are isolated to a runner tagged
`pptxgengo-integration`. GitLab enables them only on the protected default
branch when `PPTXGENGO_FULL_TESTS=true`, for scheduled pipelines or as a manual
job in a web pipeline. The GitHub exhaustive matrix is similarly opt-in through
the repository variable `PPTXGENGO_FULL_TESTS=true` on the default branch for
scheduled or manually dispatched workflows and requires a self-hosted runner
tagged `pptxgengo-integration`.

Do not set the live native output environment variable in general CI jobs.
Native PowerPoint qualification remains a separate `make test-native` run.

### Windows preview

`.github/workflows/windows-tests.yml` builds the three Windows executables,
parses the PowerShell scripts, runs portable Go regression tests and a package-style
smoke test in a path containing spaces. It uses hosted `windows-latest`, does not
open PowerPoint and needs no private photographs. Windows COM export and NTFS
receipt permissions also have simulated regression coverage.

Actual Windows PowerPoint qualification is a separate opt-in workflow job. Register
a self-hosted runner with labels `Windows` and `pptxgengo-windows-native`, install
desktop PowerPoint and the bundled IBM Plex fonts, and start the runner's `run.cmd`
from the signed-in desktop account. A Windows service/session-zero runner is not
an eligible Office automation session. Set repository variable
`PPTXGENGO_WINDOWS_NATIVE=true`, then manually dispatch the workflow on the default
branch. It retains PDFs, PNGs, receipts and logs; a human must still inspect the
full-size images. No native job is enabled for pull requests.

GitLab has the matching opt-in `windows-native` job for a PowerShell shell runner
tagged `windows` and `pptxgengo-windows-native`. Enable
`PPTXGENGO_WINDOWS_NATIVE=true` and run its manual job in a web pipeline on the
protected default branch. It also requires an interactive desktop runner and
retains the smoke-test output as an artifact.

The tester ZIP includes `smoke-test-windows.ps1`; use `-Native` only on a real
Windows desktop with PowerPoint. Its manifest explicitly reports that a
cross-compiled build has not yet received Windows runtime qualification. See
`internal/releasepackage/WINDOWS.md` for installation and operator checks.

## Short-mode skips

The following tests retain their full assertions in `make test-integration` and
`make test-race-full`, but skip under `make test`:

- `internal/deckproject.TestStockEditableEntireBindableCatalogRoundTrips` and
  `internal/wmdesign.TestAuthoringMetadataAll587PinnedTemplates`: exhaustive
  source catalog sweeps;
- `internal/deckproject.TestFinalRelocatedProjectsCompileAndPreserveResourcePins`:
  rebuilds three final projects from relocated copies;
- `internal/deckproject.TestDetachSourceScenePreservesAuthoredCopy`,
  `TestDividerClosedContractAndRemoval`,
  `TestSplitSectionDividerAndDetachPreserveSourceFiles`, and the artwork-backed
  `lifecycle/three-phases` and `plan/gantt` subtests of
  `TestScaffoldLoadsBuildsAndPinsActualParent`; these build slides using
  registered client/brand images. The negative validation checks in
  `TestScaffoldRequiresReasonAndLeavesTypedRecipesExplicit` still run in short
  mode before its artwork-backed positive build is skipped;
- `cmd/pptxdesign.TestAssetGalleryVerifiedOriginals` and the renderer
  integration tests below: read registered
  private artwork;
- `internal/wmdesign.TestCompactArchitectureLabels`,
  `TestIntakeLocalDiagramsAndStraightArrow`, `TestIntakeCardFitAllFrozenFailures`,
  `TestIntakeRepairsFrozenTwentyRejections`,
  `TestIntakeRepairsRound12IncomingTwentyTwo`,
  `TestIntakeRepairsFinalFrozenTwentyTwo`, `TestV5DeltaSourceQualification`,
  `TestV5NativeOfferSectionsHavePrimitiveClearance`,
  `TestIntakeArchitectureDeviceNativeGroup`, and
  `TestIntakeArchitectureAndGeographyAllFrozenV3Nodes`: frozen-layout and
  renderer integration sweeps that require the registered artwork tree.

These skips keep the fast lane independent of private artwork and expensive
whole-catalog or relocated-project sweeps. Synthetic asset, preview, layout,
serialization, and validation tests remain in the fast lane. `TestNativeLiveSmoke`
runs only when its explicit output-path environment variable is set; all
non-native Make targets clear that variable. `make test-native` requires an
output path that does not already exist before running the test.

## Private branding assets

Some renderer and gallery integration tests read artwork named in
`internal/wmdesign/scene_primitive_reference.go`. Those image bytes are not
stored in this repository. Set `WMDS_BRANDING_ROOT` to the directory whose
relative paths match the registry before running `make test-integration` or
`make test-race-full`. Without that private asset tree, the named short-mode
tests are skipped in `make test`; running the full lanes without it will fail
when a test actually renders a registered asset. The actual native PowerPoint
smoke test does not replace these Go renderer integration checks.

## Local commands

```sh
make test
make test-race
make test-integration
# Deliberately expensive; run only when an exhaustive race pass is needed.
make test-race-full
```

To reproduce one failure with its package timeout:

```sh
go test -count=1 -timeout=10m ./internal/deckproject
```
