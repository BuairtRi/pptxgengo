# Wave 3 accent audit: phrase anchors and arrow assets

## Implementation update after the initial audit

The observations below describe the initial audit, not the current compose API.
The compose path now supports occurrence/range selectors, rich-run phrases,
explicit `multiline: "per_line"` underline/highlight placement, moved anchors,
and pinned arrow endpoint contracts. `manual_only` arrows always stage artwork
without requiring automatic calibration. The 29-case qualification set at
`library/diagram-components/accent-arrow-qualification.json` covers these paths.
Geometry tests and structural probes pass; native visual qualification of this
expanded set remains pending. Earlier source-specific visual evidence does not
qualify these new variations. See the diagram-components README for current inputs.

## Scope and evidence

Read-only audit of the current phrase-anchor adapter, UHG5/6 experiments, and
the UHG5/6/14/36/43 source slides. Source deck is
`samples/UHG Fabric Platforming RFP Response - July 2026.pptx`, SHA-256
`b0f254ed7739768d0f345257689264393049d06cdd3849a177fc764f3d348d99`.
Use the source PNG controls under `samples/reconstruction/reference/png/` and
the structural sidecars under `samples/inspection/uhg/sidecars/` as provenance.
The exact source renderer previews are `slide-005.png`, `slide-006.png`,
`slide-014.png`, `slide-036.png`, and `slide-043.png`.

## What can be reused now

### Phrase anchors (`cmd/pptxanchor`)

`cmd/pptxanchor` is a JSON-only placement calculator; it never edits a deck.
Its native PowerPoint measurement input carries presentation, slide, shape
index, complete text, target phrase, character range, point bounds, rotation,
and per-line ranges. Asset input carries `viewBox` and visible alpha bounds.
Underline mode takes an explicit container aspect, signed stroke padding, and
optical x/y offsets. Highlight mode requires an explicit source rotation and
separate nonnegative x/y padding and offsets; its output includes the required
`behind_target_text` z-order intent. Both reject missing/nonzero phrase rotation
and multiline ranges with `manual_required` and no placement coordinates.
Malformed bounds and transparent bounds outside the viewBox are rejected.

The existing implementation does **not** identify the phrase within the slide
or remeasure it after an arbitrary edit: callers must obtain a fresh native
measurement and bind the intended phrase. It also does not write/update slides,
apply z-order, detect duplicate phrase matches, calculate arrow geometry, or
guarantee native render quality. `internal/compose/accents.go` is a different,
whole-canvas-text accent path: it requires a top-aligned plain canvas text box,
checks collisions, and returns manual-required for multiline; do not treat it as
the rich-text phrase adapter.

### Proven phrase/asset calibrations

| Case | Reusable evidence | Limits |
| --- | --- | --- |
| UHG5 underline, `pivotal moment` | `cmd/pptxanchor/README.md`; bounds in `samples/reconstruction/work/pivotal-moment-bounds-v2.json`; alpha report `work/underline-visible-bounds.json`; calibration/output `work/baseline-underline-placement.json`; original and moved-phrase previews `output/UHG-5-annotations-baseline.pptx`, `output/UHG-5-underline-reflow.pptx`. Explicit observed calibration: container aspect `1.64819110185159`, signed padding `-1.2317322059691946pt`, x offset `+0.11945163576578466pt`, y offset `-3.227455397478323pt`. Baseline output geometry `x=2101165,y=283028,cx=2219853,cy=1346842 EMU`. | These values are for this source underline/style and optical relationship, not defaults. Reflow experiment changes the title so the target words are on line 2; it shows the measurement can follow the moved phrase, but source slide 5 title in the original is one line. No arbitrary phrase search or duplicate phrase test is recorded. Multiline span is deliberately unsupported. |
| UHG6 highlight, `West Monroe` | Asset is source `Picture5` (`ppt/media/image90.png`); alpha report `work/highlight-visible-bounds.json`; explicit calibration `work/highlight-calibration.json`; baseline and resized placement JSONs; structural and raster checks. Rotation `1°`, horizontal/vertical padding `5.058313518049971/5.220756405694651pt`, offsets `1.6051474964713748/0.18744265638906654pt`. Source exact placement `2355376,531327,2028186,463041 EMU`; 32pt title experiment resizes target while retaining calibration. | Values reproduce only this source highlight and optical treatment. `output/EXPERIMENT-UHG-6-highlight-32pt.pptx` intentionally changes title size, not wording. Verify source alpha-envelope behavior; visible ink has ragged corners. |
| UHG5 arrow, `Graphic 20` | `samples/reconstruction/work/arrow-visible-bounds.json` and `arrow-source.svg.png` are available for inspection. `QA_ERRORS.md` records source rotation `345.7749167°` and that transformed visible extent is wider than image canvas. Manual fallback instructions are in `scripts/stage-slide5-accents.py` and `work/manual-placement/placement-todo.json`. | No endpoint calibration, tail/tip semantic points, or validated direction-flip/short/long stretch contract exists. The script stages it off-slide and tells a human to place it curving down/right toward the panel. This is deliberately not an endpoint solver. |

The source package confirms three concrete reusable source assets (hashes are
SHA-256 of the media part bytes, not of a separately re-exported file):

| Source asset | Source part and hash | Existing description/provenance |
| --- | --- | --- |
| Navy underline | `ppt/media/image77.svg`; `2926b957ab364340e4f1f2e5b9bad1ff9ef1258aef8241d731bf6551b422cef3` | UHG5 title underline; reconstruction report describes roughly 46% transparent padding above and below its visible stroke. Alpha bounds are separately measured in `work/underline-visible-bounds.json`. |
| Navy dashed arrow | `ppt/media/image86.svg`; `a0902a239b555ed25b977ac65ba87749aad2cb3474450c1194a262f65673cea7` | UHG5 `Graphic 20`, and UHG14 `Graphic 17` references this same part. Existing intent: in the UHG5 whitespace, curve down/right toward OUR UNDERSTANDING. Do not infer a general-purpose editable connector from the UHG14 use. |
| West Monroe highlight | `ppt/media/image90.png`; `b0048aa39744978d339a9910f5cc1482814a8a0eae4c5a63e1f80d6a2885b314` | UHG6 `Picture5`, transparent 1200×165 source per `REPORT.md`, used as the behind-text highlight on “West Monroe.” |

`image123.png` (UHG14 small preview; SHA-256
`098e8b0b818ecde2cbfc8ebbdfa20fd0fe1b53ffe77c263121a1d62eedc47fda`) and
`image124.png` (architecture board; SHA-256
`993899406c12a39bee60c1e1034b4df0a2de12176883c3d65d9fce2585d22d2d`) are
also pinned raster controls, not phrase accents. The media part IDs and hashes
are grounded in the PPTX package plus `planning/WAVE3_FIRST_SLICE.md`; preserve
these identities in asset records and retain existing wording verbatim.

The exact descriptive copy in existing reconstruction notes should be kept:
the underline is “navy underline”; the slide 5 arrow is “navy dashed arrow.”
Do not upgrade these to a broader asset taxonomy without visual review.

### Source-slide context

- **UHG5:** accent-bearing executive framing page. The underline targets
  `pivotal moment`; curved dashed arrow occupies whitespace between the title
  and OUR UNDERSTANDING panel and points down/right toward that panel. See
  sidecar `slide-005.md`, reference preview and `REPORT.md`.
- **UHG6:** highlighter is behind phrase “West Monroe.” The phrase has been
  proven at source 24pt and changed 32pt title size. The evidence tests phrase
  movement and size changes, not deck-wide arbitrary repositioning.
- **UHG14:** `Graphic 17` is `ppt/media/image86.svg` at approximately
  `[852.18,132.80,70.29,70.29] pt`; `Picture15` is the board image
  `image124.png`; `Picture11` is `image123.png`. The board labels are baked into
  raster artwork. This slide is a pinned-art/container control only: neither
  semantic arrow endpoints nor editable internal diagram relationships can be
  inferred. `planning/WAVE3_FIRST_SLICE.md` contains the image hashes and exact
  placement context.
- **UHG36:** two native, grouped, five-node pathways with four preset
  right-arrow shapes per row; zero pictures. These are editable process
  connectors with explicit node gaps, not hand-drawn arrow assets. Keep this
  as a source-control for visible connector attachment/gap behavior, not as
  evidence of reusable SVG endpoint calibration. See
  `planning/WAVE3_FIRST_SLICE.md` for measured node frames and source caveat.
- **UHG43:** people/team structure and text-heavy annotations, no pictures.
  Useful for testing accents around repeated role labels and crowding/collision
  with a dense slide; it does not currently supply a curated hand-drawn arrow
  family. Preserve source role and allocation meanings from `slide-043.md`.

## Known tests and remaining gaps

`cmd/pptxanchor/main_test.go` covers underline baseline reconstruction,
invalid/missing geometry, missing rotation, nonzero rotation, multiline manual
fallback, highlight baseline and resized placement, highlight rotation/alpha
geometry, and manual highlight cases. UHG5/6 reports document native rendered
visual review and source geometry comparison. `samples/reconstruction/QA_ERRORS.md`
records the initial alpha-padding issue, UHG5 arrow rotation/extent issue, and
phrase schema mismatch already fixed. These checks do not cover:

1. phrase movement after parent/container translation while preserving
   character-range identity;
2. repeated target phrases (must bind exact occurrence/range, never first
   textual match silently);
3. a phrase whose wrapping changes between 1, 2, and more lines, including a
   target partially moving to a new line; multi-line remains manual until a
   defensible per-line rule exists;
4. phrase edits that change punctuation, Unicode apostrophes, or run boundaries;
5. arrow tail/tip and tangent visible endpoints in asset coordinates;
6. arrow aspect preservation and supported min/max visible span, short/long
   fit, and direction reversal with optical endpoint calibration;
7. arrow endpoint movement after connected text/container moves, overlap and
   renderer visual checks for those variants.

No endpoint calibration should be added by measuring only an image rectangle.
For hand-drawn artwork, inspect its alpha artwork and previews first, define
tail/tip **and local tangents** in normalized source coordinates, preserve the
source aspect/stretch/flip constraints, and calibrate each variant against an
observed native placement with visible endpoint evidence. If the source does
not establish a semantic endpoint, keep it decorative and manually placed.

## Recommended Wave3 fixture matrix

Use immutable copied fixture projects/outputs; retain source deck hash, source
slide/object IDs, media part, media SHA-256, measurement JSON, alpha report,
calibration provenance and native-render preview alongside each fixture. For
any edit fixture, store the changed text/geometry delta and do not label it an
exact source reproduction.

| Fixture | Starting point and change | Expected contract/evidence |
| --- | --- | --- |
| A1 source underline | UHG5 original phrase and object; use existing native measurement + visible-bounds reports and explicit README calibration | Source geometry match and preview. Pin source slide/object/media identity and unchanged original wording. |
| A2 moved phrase | Existing `work/underline-reflow` and `output/UHG-5-underline-reflow.pptx`; phrase shifted by title reflow | Re-measure target by exact range; underline visible stroke follows the words. Verify it is placed against the actual line bounds and does not touch adjacent line glyphs. |
| A3 repeat phrase | New test title with two exact occurrences of the same phrase; identify intended occurrence by shape ID plus start/end character offsets | Must anchor selected occurrence; missing/ambiguous range yields actionable error/manual status. Never choose first match by default. |
| A4 wrap transition | UHG5 title widths/copy creating one-line phrase, phrase entirely on line two, then phrase crossing a line break | First two can place if their target range has a single measured line fragment; crossing line break returns `manual_required`, no coordinates. Preserve wording per fixture metadata. |
| A5 translated container | Translate title/container and anchor inputs equally with unchanged phrase | Anchor-to-phrase relative geometry remains invariant; resulting slide-space placement translates by same delta. Native render verifies. |
| A6 UHG6 highlight | Original and 32pt title fixtures | Reuse explicit source calibration, confirm phrase range and behind-text ordering, compare expected source/resized placements and render. |
| A7 arrow source inventory | UHG5 Graphic 20 (`image86.svg`) plus UHG14 Graphic 17 only if visual inspection confirms same asset/variant; record exact source rel/media hash and original bounds/rotation | Build curated entry only after inspecting previews and defining visible endpoints/tangents. Current evidence labels UHG5 as navy dashed, down/right source intent. UHG14 placement alone adds no endpoint semantics. |
| A8 arrow short/long | Each approved curated arrow variant at shortest and longest permitted visible endpoint span | No nonuniform distortion; endpoints land at declared anchors; bounds/aspect fit rules report pass or `manual_required`. Establish limits empirically from previews, do not invent numeric ranges. |
| A9 arrow direction | Each variant in source direction and reversed direction, plus alternate curve direction only where a separate inspected variant supports it | Preserve hand-drawn character and correct head/tail/tangents. Flip only if explicitly allowed; do not assume flipping changes semantic tip correctly. |
| A10 dense/repeated visual | Add approved phrase accents to copies of UHG43 dense team layout at repeated labels and moved parent location | Bind exact occurrence; collision/overlap checks and native preview catch accents crossing glyphs, roles, or connectors. Existing UHG43 source itself has no accent. |
| A11 manual fallback | UHG5 source underline and arrow with deliberately multiline target, missing measurement, unsupported rotation, or uncalibrated arrow | Follow the pattern in `scripts/stage-slide5-accents.py`: stage assets off-slide, place native labels/instructions and speaker notes, emit `manual_required` with target phrase, desired relation, asset, endpoint/visibility constraints and visual review cue. Never show guessed placement on-slide. |

## Implementation handoff

Keep current `pptxanchor` formula/calibrations as named source-specific presets;
add explicit target identity/range resolution and tests before generalizing. The
caller should remeasure after the text/layout pass, then compute placement,
apply z-order in the owning slide writer, and render for review. Expand curated
arrow metadata and endpoint solving only after asset-by-asset preview inspection
and observed source placement calibration. Until those prerequisites exist,
the Wave3 arrow behavior should stage assets and request manual placement.
