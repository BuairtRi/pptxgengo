# Wave 2 follow-up — native bullets and picture fidelity

Implementation is complete for this bounded slice; final native/visual qualification is blocked. Historical eight-slide checkpoint
and its proof remain tied to commit `7090db77`; new evidence is recorded separately.

## Scope

- Native three-glyph rich bullets with explicit hanging/text indentation, mixed
  runs, paragraph spacing, cache invalidation and structural/native checks.
- Original static SVG icons embedded alongside explicit pinned PNG fallbacks.
- Native picture outlines replacing four independent lines in the source control.
- Consistent biography sidebar item spacing and five responsibilities per panel.
- Read-only native adapter v8; paragraph properties use a single snapshot.

## Review fixtures

1. UHG67 source industries panel and changed-content bullets with wrapped emphasis.
2. Polished illustrative biography.
3. UHG28 thumbnail/caption control with native picture borders.
4. EnableComp needs/response page with five original SVG icons.

Generator: `scripts/build-wave2-followup-spec.py`.
Specs: `library/visual-components/followup-review.json` and a deliberate
30pt-height bullet overflow fixture. `followup-qualification.json` adds the
negative fixture for measurement; it is not a deliverable deck.

## Evidence and current blocker

- `go test ./...`, AppleScript compilation, Swift type checking, Python generator
  syntax checks, and `git diff --check` pass.
- Preliminary native bullet smoke: three text zones, all three glyphs, zero fit
  failures; 41.26s. Six-object adapter regression preserved every historical
  character/style/bounds observation exactly.
- Preflight measured 51 distinct text contracts in 272.25s. The four-page build
  has 101 editable objects, 63 text objects (17 rich), 13 pictures (five SVG),
  four native picture outlines, and eight native bullet paragraphs. All 63 text
  zones fit; no layout failures.
- The negative bullet fixture needs 131.2pt in a 30pt zone: 101.2pt overflow.
  Build rejects it without creating an output bundle.
- Preflight artifacts: `samples/visual-wave2-followup/preflight-review/`,
  `qualification-evidence.json`, `preflight-fit.json`, and
  `preflight-negative-fit.json`.

These preflight observations use the frozen v8 binary before the final SVG
validation hardening. The final binary is `/tmp/pptxcompose-wave2-v8b`, SHA-256
`c7c804e83ae0b887061f55d714f168b3c83bbed0c321d1c6f744aaeace8445c4`.
Its final probe bundle is already prepared. Environment binding correctly rejected
using the older preflight bundle under the new binary; older cache entries were
not bypassed or imported as current proof.

PowerPoint PDF export timed out (-1712); later opens fail with -9074. A retry
returned without error but produced no PDF; that is recorded as a failed export,
not visual evidence. Desktop inspection is unavailable (CUA pipe startup failure,
no display capture, and System Events timeout). The user has been asked to inspect
any file-access/repair/export dialog. Only the saved task preflight deck was closed;
unsaved user decks were preserved. No PDF, render, final native verification, or
visual approval is claimed. `followup-proof.json` must not be written until those
gates pass.

## Resume qualification

Use the same frozen binary throughout. The generated final probe exists; do not
regenerate over it or reuse the v8 cache. After PowerPoint access is restored:

```sh
/tmp/pptxcompose-wave2-v8b measure --bundle samples/visual-wave2-followup/final-qualification-probe --cache samples/visual-wave2-followup/v8b-cache --out samples/visual-wave2-followup/final-qualification-evidence.json
/tmp/pptxcompose-wave2-v8b fit-report --spec library/visual-components/followup-review.json --cache samples/visual-wave2-followup/v8b-cache --out samples/visual-wave2-followup/fit.json
/tmp/pptxcompose-wave2-v8b fit-report --spec library/visual-components/followup-negative.json --cache samples/visual-wave2-followup/v8b-cache --out samples/visual-wave2-followup/negative-fit.json
/tmp/pptxcompose-wave2-v8b build --spec library/visual-components/followup-review.json --cache samples/visual-wave2-followup/v8b-cache --out samples/visual-wave2-followup/followup-review
/tmp/pptxcompose-wave2-v8b verify --bundle samples/visual-wave2-followup/followup-review --out samples/visual-wave2-followup/verification.json
```

Close only unchanged task probe decks before verification. Export the final task
copy with `scripts/export-powerpoint.applescript`, render all four pages with
`scripts/render-pdf.swift`, and obtain primary plus independent visual reviews.
Compare UHG28 regions using `scripts/compare-wave2-regions.py --spec
library/visual-components/followup-review.json <render-dir> <new-comparison-dir>`.
The optional QA-only SVG selection control (`scripts/build-svg-selection-control.go`)
has a green primary SVG and magenta PNG fallback: native rendering should show
green. Its preflight PPTX exists locally but has not been rendered.

Once evidence is complete, `scripts/report-wave2-followup.py` binds the artifacts
and code into the tracked follow-up proof. Keep the historical v7 proof intact.

## Limits

Native character bounds exclude bullet glyphs. A minimum hanging indent reserves
space, structural checks bind exact paragraph geometry, and visual review checks
bullet separation and wrapping. The UHG67 reference uses explicit Arial for its
bullet; this fixture follows the first Arial text run. It preserves visible
geometry without claiming identical source markup or whole-slide identity.

SVG pictures retain vector media, not editable individual paths. The PNG fallback
must be independently visually reviewed; hashes prove identity, not equivalence
between two different formats. Unsupported SVG dependencies fail explicitly.

Final native verification remains slow. PowerPoint's bulk character query returns
an aggregate range, so per-character inspection is retained; see
[the spike](WAVE2_VERIFIER_SPIKE.md). Rich-text recovery, arbitrary diagram routing
and whole-slide reconstruction remain outside this follow-up.
