# Native PowerPoint round-trip qualification fixture

This source-only harness exercises the handoff's Save As, three supported text
edits and slide reordering criteria against a disposable synthetic project. It
is not included in stable v4.1.0. It records actual package identity/text/order
and bounded adoption/rebuild evidence separately from native visual acceptance
and human editing ergonomics.

## Prepare a fixture on macOS or Windows

Run from a source checkout with Go 1.27.1+. No private branding or Office is
needed for preparation. Use a new directory; existing output is refused.

```sh
PPTXGENGO_ROUNDTRIP_PREPARE_OUT='/tmp/pptx roundtrip fixture' \
  make test-roundtrip-prepare
```

PowerShell equivalent:

```powershell
$env:PPTXGENGO_ROUNDTRIP_PREPARE_OUT = "$env:TEMP\pptx roundtrip fixture"
go test -count=1 -timeout=2m -run '^TestNativeRoundTripPrepare$' ./internal/deckproject
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
```

The folder retains `baseline.pptx`, `plan.json`, `instructions.md` and a complete
maintained project with receipt-pinned immutable build outputs. Keep those files
unchanged. Open only this fixture in desktop PowerPoint:

1. **Save As** a new `saved-as.pptx` before editing. Keep that first save intact.
2. Edit the three named Selection Pane objects in `instructions.md`, using the
   prescribed exact copy. Do not replace objects or flatten groups.
3. Move the original second slide into first position.
4. **Save As** another new `edited.pptx`. Close the fixture. Preserve both saves.

Object names are human fixture instructions. The verifier matches lineage tags
and source fields; names, current text, positions and geometry never establish
correspondence with source.

## Verify supplied saved files

```sh
PPTXGENGO_ROUNDTRIP_FIXTURE='/tmp/pptx roundtrip fixture' \
PPTXGENGO_ROUNDTRIP_SAVED_AS='/tmp/pptx roundtrip fixture/saved-as.pptx' \
PPTXGENGO_ROUNDTRIP_EDITED='/tmp/pptx roundtrip fixture/edited.pptx' \
PPTXGENGO_ROUNDTRIP_VERIFY_OUT='/tmp/pptx roundtrip evidence' \
  make test-roundtrip-verify
```

On PowerShell, set the same four environment variables and run:

```powershell
go test -count=1 -timeout=3m -run '^TestNativeRoundTripVerify$' ./internal/deckproject
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
```

The verifier reads regular bounded files, checks the fixture plan against the
receipt-pinned baseline, checks full deck/slide/shape tag survival in both saves,
checks unchanged text/order in the first save, checks the exact three changes
and reordered slide tokens, and requires three unambiguous native-only proposals.
It preserves all reported package/format/geometry differences as manual review
items. A mismatch fails without source adoption. Existing evidence directories
are never overwritten.

Only a passing prescribed synthetic fixture is adopted into its disposable
`project/` directory. The actor/reasons explicitly identify automated fixture
work, with no human acceptance. The harness verifies repeated adoption and
reconciliation, rebuilds headlessly, checks exact regenerated copy, and verifies
that the old baseline and edited bytes remain available. It does not change
the authored YAML slide order or adopt native geometry/formatting.

Evidence includes exact saved files, lineage inspections, the first-save
comparison, a closed review packet, fixture decisions, adoption receipt and
`evidence.json`. Inspect `manual_review` even when `status` is `pass`. A supplied
file's native application provenance is not independently verified: retain app/OS
versions and the operator's observations separately. The verifier's platform
field identifies the verification machine, not the deck's editing platform.

## Opt-in Windows COM execution

On a signed-in Windows desktop with PowerPoint, run:

```powershell
$env:PPTXGENGO_ROUNDTRIP_WINDOWS_OUT = "$env:TEMP\pptx roundtrip live new"
go test -count=1 -timeout=3m -run '^TestNativeRoundTripLiveWindows$' ./internal/deckproject
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
```

The harness creates its own new directory and input copy. COM opens that copy
writable, validates deck/build tags, retains an initial Save As, finds the three
shapes through slide/shape/build tags, checks their baseline text, edits them,
moves the declared tagged slide to first position and saves an edited copy.
It records PowerPoint's version, the host-reported OS version, execution timestamps
and hashes of the closed input/saved files, then runs the same
independent saved-file verification/adoption/rebuild sequence.

These actions use Microsoft's documented
[Presentation.SaveAs](https://learn.microsoft.com/en-us/office/vba/api/powerpoint.presentation.saveas),
[Open XML format value 24](https://learn.microsoft.com/en-us/office/vba/api/powerpoint.ppsaveasfiletype),
[TextRange.Text](https://learn.microsoft.com/en-us/office/vba/api/powerpoint.textrange.text) and
[Slide.MoveTo](https://learn.microsoft.com/en-us/office/vba/api/powerpoint.slide.moveto).

The helper has a 90-second deadline and a separate five-second exact-path
cleanup budget. It closes only its three task paths, never quits or kills
PowerPoint, and retains files and diagnostics when cleanup is unconfirmed.
Inspect PowerPoint for policy/Protected View prompts after a failure; resolve
the prompt and use a new fixture output path. Do not repeatedly automate a
blocked desktop session.

GitLab's protected-default-branch manual `windows-native` lane runs this
fixture after native export smoke. GitHub Actions is disabled. The opt-in private
GitLab Windows CLI lane parses the script and exercises simulated/helper/verifier failures without
opening Office. Headless Make lanes clear all opt-in fixture variables; short
tests skip explicit live/prepare/verify entry points. Normal short and selected
race coverage still exercises the complete hermetic verifier.

## Qualification status and remaining tasks

No real Windows desktop execution is claimed until the interactive runner
produces passing retained evidence. macOS currently uses the supplied-file
workflow: the earlier automated Save As attempt timed out without an output.
No Mac automation success is implied by hermetic package edits.

Even a passing real COM run leaves human acceptance `not_recorded`. Complete
the [editing pilot](native-editing-pilot.md) move/align/resize/table/diagram tasks
and visual review on both platforms, including relevant density tiers. Record
actual app/OS versions, screenshots, task outcomes and manual differences.
Duplicate/delete/ungroup behavior remains bounded ambiguity/manual review;
this two-slide fixture does not qualify an entire library or general reverse
compilation. Use a new fixture for each qualification attempt and keep evidence
private when it includes maintained projects or customer material.
