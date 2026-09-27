# Wave 2 follow-up — native bullets and picture fidelity

The four bounded fixtures passed native verification and primary plus independent
visual review. This closes this follow-up, not the entire Wave 2 fidelity gate.
Historical eight-slide proof remains tied to commit `7090db77` and is unchanged.

## Review artifacts

- [Editable PowerPoint](../samples/visual-wave2-followup/followup-review-v2/followup-review-v2.pptx)
- [Four-page PDF](../samples/visual-wave2-followup/followup-review-v2.pdf)
- [Tracked proof with artifact/code hashes](../library/visual-components/followup-proof.json)
- [Source specification](../library/visual-components/followup-review.json)

The pages cover UHG67 source/changed-content rich bullets, a dense synthetic
biography, UHG28 thumbnail/caption controls, and an EnableComp needs/response
composition with five original SVG icons. These are component qualification
fixtures, not a complete proposal narrative.

## Implemented and verified

- Native round and dash bullets with explicit hanging/text indentation, mixed
  runs, paragraph spacing, cache invalidation and structural/native checks.
- Original static SVG pictures with explicit pinned PNG fallback media.
- Native picture outlines replacing four independent border lines.
- Consistent biography sidebar spacing and five responsibilities per panel.
- Read-only native adapter v8; paragraph properties use a single snapshot.

Final deck: **101 editable objects**, **63 text objects**
(17 rich), 13 pictures (5 SVG),
4 native picture outlines and 8 native bullet paragraphs.
All 63 text zones fit, with zero layout failures. Maximum observed frame delta:
**0.00004980 pt**; maximum text-bound excursion:
**0.00503922 pt**.

The deliberate 30 pt bullet zone requires 131.2 pt: 101.2 pt overflow. Both its
fit report and attempted build fail; no output bundle is created. Full Go tests,
AppleScript compilation, Python syntax checks and `git diff --check` pass.

## Visual evidence and corrections

The first follow-up passed native checks for all 101 objects, yet its square
bullet was invisible in the PowerPoint PDF. Both visual reviewers rejected it.
The API now rejects that glyph; a regression test preserves the restriction.
Round `•` and dash `–` are qualified. Initial spec, render, review and native
verification remain in `samples/visual-wave2-followup/` as rejected evidence.
A stale reference to a nonexistent following slide was also corrected.

All four corrected 1920×1080 pages were inspected individually by the primary
agent and an independent Luna reviewer. Slides 2 and 4 retain identical PNG
bytes to the initial accepted pages. Reviews bind to exact rendered image hashes.

Source/control comparisons use identical native render dimensions without
resampling:

| Region | Result |
| --- | --- |
| UHG28 four thumbnail crops | Two pixel-identical; two over 99.996% exact pixels; maximum channel difference 1/255. |
| UHG28 two caption crops | Every pixel within one RGB level per channel. |
| UHG67 industries panel | 99.7268% exact pixels; 99.9052% within one RGB level; maximum channel difference 6/255. |

Native outlines improve the earlier four-line border approximation. These metrics
apply only to the named crops and do not establish whole-slide pixel identity.

A separate SVG-selection control pairs a green SVG with a deliberately magenta PNG
fallback. Its native PDF contains 65,536 green-dominant pixels and no magenta pixels
in the interior test region, proving primary SVG selection for that control.
PDF color management changes RGB values; this is not an exact-color or SVG/PNG
equivalence test.

## Reproduction and environment

Generator: `scripts/build-wave2-followup-spec.py`. The qualification spec includes
the deliberate negative fixture and is not a deliverable deck. Keep one frozen
CLI binary throughout probe, measure, fit, build and verification; changed binaries
invalidate the environment-bound measurement cache.

Qualified binary: `/tmp/pptxcompose-wave2-v8c`, SHA-256
`865fd42696f0b39f5394f895ae468d80afa0a946f51d85d8adcbb204c6b6523a`.
Final probe measurement: 51 distinct text contracts, 271.73 seconds. Native final
verification remains approximately five minutes because exact character inspection
is retained. See [the verifier spike](WAVE2_VERIFIER_SPIKE.md).

Evidence paths are under `samples/visual-wave2-followup/`: `qualification-v8c`,
`qualification-v8c-evidence.json`, `v8c-cache`, `fit-v2.json`,
`negative-fit-v2.json`, `negative-build-v2-check.json`, `verification-v2.json`,
`render-v2`, `source-region-comparison-v2`, `bullet-region-comparison-v2`, and
`visual-review.json`. `scripts/report-wave2-followup.py` binds the accepted artifacts
and current implementation into the tracked proof. Generated binaries remain ignored.

The file-access blockage was resolved when the user granted access. The PDF export
helper now rejects an existing destination and requires a PDF header; rasterization
and visual review are separate gates. PowerPoint operations stayed serialized,
and unsaved user decks were preserved.

## Limits and next slice

Character bounds exclude bullet glyph ink, so indentation checks do not replace
visual review. UHG67 uses an explicit Arial bullet font; this fixture follows the
first Arial text run. Native SVG pictures preserve vectors, not editable paths;
fallback hashes prove identity, not equivalence. Native groups/tables remain
separate editable shapes in these fixtures. Rich-text recovery and automatic
layout selection remain unsupported.

Next: [measured process paths with container anchors](WAVE3_FIRST_SLICE.md), using
UHG36 source controls, changed content, 4/6-node stress cases and parent translation.
Reuse the existing editable right-arrow renderer; add anchor and gap calculations.
Broader diagram routing and accent placement remain separate qualification work.
