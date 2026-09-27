# Wave 2 — rich typography and visual components

Historical v7 checkpoint at commit `7090db77`. See the
[follow-up status](WAVE2_FOLLOWUP.md) for newer code and its pending visual gate.

2026-09-26. Eight bounded fixtures passed native verification and two visual
reviews. This is not blanket approval of the Wave 2 library or a whole-slide
pixel-identity claim. Evidence is bound in
[`proof.json`](../library/visual-components/proof.json).

## Review artifacts

- [Eight-slide PowerPoint](../samples/visual-wave2/wave2-review/wave2-review.pptx)
- [PDF](../samples/visual-wave2/wave2-review.pdf)
- [Native verification](../samples/visual-wave2/wave2-verification.json)
- [Combined visual review](../samples/visual-wave2/visual-review.json)

## Implemented

- Rich Arial paragraphs/runs on canvas and layout blocks, including bold, italic,
  single underline, color, paragraph spacing and line-spacing multiples. The v7
  native adapter checks
  each character's style and each paragraph's alignment/spacing.
- PNG/JPEG contain, cover, focal-point and source-crop placement; explicit stretch
  for source geometry that already distorts an image. Original image bytes and
  hashes are preserved.
- Editable `homePlate`, `star5`, `rightArrow` and `rect` shapes, bounded adjustment
  guides, and `wdUpDiag` pattern fills. Text remains independently editable.
- Structural verification of presets, adjustments, patterns, picture frames,
  image relationships, media hashes and exact source-crop percentages.
- Full style/environment cache binding, including the four actual Arial font
  files. Old measurements are not silently reused after a CLI binary change.
- Reproducible source-derived fixtures and an updated repository skill. See
  [usage and limits](../library/visual-components/README.md).

## Native smoke evidence

The smoke probe and final smoke deck passed native verification for mixed bold,
italic, underline, color and paragraph spacing at two text widths. The smoke
deck also passed picture-frame checks for contain, cover and stretch. These
checks do not substitute for full fixture visual review.

- [Smoke spec](../samples/visual-wave2/smoke-spec.json)
- [Native probe evidence](../samples/visual-wave2/smoke-evidence.json)
- [Native final evidence](../samples/visual-wave2/smoke-verification.json)

Initial PDF export timed out and subsequent PowerPoint opens failed. The user
granted file access, after which measurement and PDF export succeeded. The smoke
render was visually reviewed: full-source contain, cropped cover and deliberate
stretch display as expected; bold/italic/underline/color are visible. No unsaved user presentation was closed to recover the application.

## Fixture qualification

The combined review spec is
[`library/visual-components/review.json`](../library/visual-components/review.json).
It contains eight pages: source/changed-content deliverable panels, source portrait
crops, a synthetic biography, a 19-role roster, two roadmap interval examples, and
a five-row needs/response page with source-derived icon previews. All 276 text
zones pass planning with zero overflow and zero layout failures. The separate
long-role negative fixture exceeds its 14.4pt role zone by 18pt at the fixed 9pt
font size; the build rejects it without producing a deck.

The initial native probe measured 133 contracts. Its fit report caught two small
caption overflows caused by omitted 90% source line spacing (QA80). Source font
sizes and frames are preserved while the missing paragraph property is added.
The historical smoke evidence is v6; the final fixture requires v7 evidence.

The first render exposed five icons hidden by later row surfaces (QA83). Explicit
image layers fixed the issue; pages 1–7 retained identical render hashes. Primary
and independent Luna visual reviews accept all eight corrected pages for fit,
crops, overlap and caption placement. Both record a non-blocking biography polish
issue: uneven sidebar bullet pitch and lower density than the UHG67 reference.
Final native verification passed all 443 objects: 276 text objects (28 rich),
25 pictures and 27 preset shapes, plus surfaces/lines. Maximum frame deviation is
0.000053875pt; maximum text-bound excursion is 0.005039215pt, inside the 0.15pt
text tolerance. Structural checks bind image bytes/crops and preset geometry.
The proof additionally rejects pictures fully covered by later opaque rectangular
surfaces, a regression check for the hidden-icon issue.

UHG28 region comparison records 96.21–97.25% exact pixels for the four screenshot
regions, with mean channel error 0.43–0.91 on a 0–255 scale. Difference images
show the material residuals at picture borders. Both caption regions differ by
at most one RGB level per channel. This is evidence of close component
reproduction, not 100% pixel identity or whole-slide reconstruction.

Changing one rich run color produces exactly one new native request and reuses
275 unchanged text bindings. Final review objects total 443.

## Remaining Wave 2 scope

- Full native verification is still slow: this 443-object, 10,474-character deck
  takes over ten minutes in the current per-character AppleEvent adapter. Cache
  reuse accelerates planning but does not eliminate final inspection. Batch
  character/font snapshots should be spiked before qualifying much larger decks.

- Native rich bullets and hanging indentation are not implemented. Uniform
  separate marker/text blocks remain available; that is a narrower contract.
- Picture outlines currently use four native lines in the deliverable fixture;
  their joins are not certified identical to native picture outlines.
- Source controls preserve selected geometry and artwork, not every inherited
  source-slide style. UHG44 middle anchoring and line spacing are now resolved;
  the control reproduces measured glyph starts using separate text boxes.
- EnableComp SVG originals are retained, but the current compose renderer uses
  traced 8× AppKit PNG previews. Native SVG image support remains future work.
- Source roadmap groups are flattened and its table is rectangles/text. A
  declared interval example does not establish a general scheduling component.
- No universal freeform resizing, automated aesthetic placement or rich-text
  recovery approval is implied. Component catalog promotion waits for proof.

## Repository housekeeping

The approved cleanup closed 23 saved obsolete verification decks and archived
27 obsolete roots (1,216 files, including 9 PPTX files). Archive members were
hashed and verified before the originals were removed. Two unsaved generated
decks were left open. The recoverable archive is
`samples/_archive/2026-09-26-wave1-superseded.zip`, SHA-256
`9569100b251c48c128fc5c75fe435132e5861cee26be462e61f2cb2ae2692dbc`.
It includes a manifest with original paths and restoration instructions.
