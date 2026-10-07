# Testing lanes

Run `make test` for everyday development. It runs the repository suite with Go's
`-short` mode, `-count=1`, and a five-minute per-package timeout. It should fit
within a normal short development loop. Only specifically marked exhaustive or
resource-dependent integration tests skip in short mode; ordinary unit and
focused regression coverage still runs.

## Lanes

| Command | Coverage | Environment |
| --- | --- | --- |
| `make test` | Normal short suite | Go 1.27.1+; no PowerPoint required |
| `make test-pr` | Focused PR units plus compilation of every package | Empty branding root; no actual installers/model/Office |
| `make test-race` | Race checks for project source mutation guards, the full hermetic native-export package, library catalog cache, and concurrent PPTX serialization | CGO-enabled Go toolchain |
| `make test-integration` | Full normal suite, including exhaustive catalogs and final relocated projects | Registered private branding assets; see below |
| `make test-race-full` | Full suite under the race detector, including tests skipped by `-short` | Opt-in; CGO-enabled Go; 10-minute per-package ceiling |
| `make test-native` | Opt-in live PowerPoint smoke test | macOS desktop session, Microsoft PowerPoint, Swift/PDFKit, and a new `PPTXGENGO_NATIVE_LIVE_OUT` path |
| `make test-roundtrip-prepare` / `make test-roundtrip-verify` | Retained synthetic three-field/Save As/reorder fixture and independent verification/adoption/rebuild | Explicit fixture/saved-file/new-output paths; preparation and verification do not open Office |
| `make test-roundtrip-windows` | Live COM actions followed by saved-file verification | Interactive Windows desktop with PowerPoint; new `PPTXGENGO_ROUNDTRIP_WINDOWS_OUT` |

The timeout passed to `go test` applies to each package test process, not to the
entire Make target. `make test-integration` and `make test-race-full` are bounded
by their per-package ceilings; a timeout is a test failure and should be
investigated by rerunning the affected package. The exhaustive race lane is not
a routine developer check and should not be described as passing unless it has
actually completed.
The fast target uses a five-minute per-package ceiling; focused race checks
use 150 seconds per package;
exhaustive race checks use 10 minutes per package.

## CI policy

All project jobs run in private GitLab. GitHub Actions is disabled, and the
repository contains no GitHub workflows. GitHub remains the source review and
merge authority used by slotctl; it does not execute CLI operations or retain
CI artifacts. Slot preflight validates the local GitLab job graph with
`go run ./scripts/cmd/ci-lint`; GitLab server lint validates rules and matrices.

Ordinary branch pushes create no pipelines. Opening/updating a PR selects
focused `make test-pr` units, workflow validation, and a secret scan. Main merges
run the complete hermetic `make test` suite, workflow validation, and secret
scanning. Neither tier runs races, actual installers, downloaded models, or
native desktop qualification.

Protected release tags add actual installation, model closure/goldens, source
security, generated resources/decks, six-platform release builds, signing,
notarization, archive scanning, signature verification, attestation and private
publication. Tags never run races or performance diagnostics. Nightly protected
main runs long races, retrieval/performance, installer and cross-platform/native
qualification. Source/binary security gates remain mandatory for release.

Agents can select one subsystem job through `PPTXGENGO_CI_JOB` in a web/API
pipeline. All native Linux ARM64 jobs use the dedicated Mac mini Linux runner,
not the Pi pool. Protected Mac runners never execute ordinary PR/main jobs.
See [CI tiers and notifications](ci-cadence.md) for the complete job inventory,
automatic GitHub PR relay, schedule and cache ownership.

Full `test-integration` and `test-race-full` jobs remain isolated to a runner
with tag `pptxgengo-integration`, enabled on protected main with
`PPTXGENGO_FULL_TESTS=true` in scheduled or manual web pipelines. General jobs
never open Office. Native PowerPoint qualification requires a separate desktop
run, retained evidence and human inspection.

### Windows and Intel Mac qualification

The forthcoming Windows shell runner uses tags `windows` and
`pptxgengo-windows-native`. Set `PPTXGENGO_WINDOWS_CLI=true` in a web pipeline on
protected main and start manual `windows-cli`. It parses PowerShell scripts,
builds all three executables, runs portable/helper tests, exercises actual
installation and a relocated path containing spaces, and verifies pinned model
retrieval/performance. It does not open PowerPoint. Native Windows ARM64 requires
an additional `arm64` runner tag and `PPTXGENGO_WINDOWS_ARM64_CLI=true`.

Live Office qualification is separate: set `PPTXGENGO_WINDOWS_NATIVE=true` and
start `windows-native` in a protected-main web pipeline. The runner must run in
the signed-in desktop account with PowerPoint and bundled fonts installed;
Session 0 is ineligible. This lane retains native export smoke and the synthetic
Save As/three-field/reorder round-trip evidence. See
[native round-trip](native-roundtrip.md) for ownership, cleanup and human
acceptance boundaries. No Windows runtime or Office result is claimed until
the runner actually executes the lane.

Private Intel Mac execution is configured separately for tags `macos` and
`darwin-amd64`, with a preprovisioned Go 1.27.1 toolchain. Enable
`PPTXGENGO_MACOS_INTEL=true` in a protected-main web pipeline and start
`macos-intel-cli` and `macos-intel-model`. These lanes remain pending until the
runner is available. Historical GitHub reports remain historical evidence;
they are not current CI gates.

Generic installer/model diagnostics are private maintainer-access GitLab
artifacts retained for 14 days. Native Office artifacts also remain private.
The tester ZIP includes `smoke-test-windows.ps1`; use `-Native` only on a real
desktop. Cross-build/signing on Linux remains available while Windows native
execution is pending. See `internal/releasepackage/WINDOWS.md` for operator checks.

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

The contrast catalog audit, Gantt source-artwork audit, and V10 individual
source builds also require private originals. V6/V10/V11 round-trip tests keep
their binding assertions in short mode; their artwork-dependent rendering
assertions run in the integration lanes. Frozen gallery closure remains a
required hermetic check.

Catalog cache race regressions run in separate invocations with the same
150-second ceiling for each. The combined invocation exceeded that ceiling on
hosted macOS even though its individual checks completed. All five regressions
and their assertions remain in the focused race lane.

## Offline search performance diagnostics

`make test-search-performance` requires an existing pinned offline model and
an explicit new `PPTXGENGO_SEARCH_BENCH_OUT` directory. The standard corpus
contains 26 actual entities; three separate child processes retain keyword,
semantic and hybrid timing/memory JSON. This opens no Office application.
GitLab runs this on Kubernetes Linux and protected Mac ARM64; configured
Windows and Intel Mac lanes remain opt-in pending their private runners. Generic
diagnostics are retained for 14 days. See [measurement scope and commands](search-performance.md).
Ordinary headless Make targets clear this opt-in output variable.

## Actual installer process qualification

`make test-installation-process` requires a new
`PPTXGENGO_INSTALL_PROCESS_OUT` directory. It builds both fixture versions of
the actual three tools and exercises their installation lifecycle through new
processes in owned paths. Windows also executes the actual installer with
PATH/skill/font opt-outs and synthetic full-format resources. Ordinary headless
Make targets clear this opt-in; private originals and Office are not required.

Linux amd64/arm64 run in Kubernetes and protected macOS arm64 on the private
Mac runner. Windows amd64/arm64 and Intel Mac lanes require their private runners
and explicit protected-main web-pipeline flags. Passing cross-builds are not
native execution evidence. GitLab retains generic qualification JSON for
14 days, excluding unsigned executable fixtures and customer content.
See [installation evidence and scope](installation.md).
