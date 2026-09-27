# Wave 2 visual components

These source-derived fixtures exercise rich typography, image placement and
editable roadmap shapes. See [qualification status](../../planning/WAVE2_CHECKPOINT.md)
before treating any fixture as accepted. Source UHG assets must be available
locally; generated binaries and render evidence remain ignored under `samples/`.

## Reproduce

1. Extract the pinned artwork with `scripts/extract-wave2-assets.py` if absent.
2. Run `python3 scripts/build-wave2-review-spec.py` to regenerate `review.json`.
3. Build one CLI binary, then use it throughout `probe --cache`, `measure --cache`,
   `fit-report --cache`, `build --cache`, and `verify`. Command details are in
   [the rich text guide](../../planning/WAVE2_RICH_TEXT_IMPLEMENTATION.md).
4. Export the final task copy through PowerPoint, render it with
   `scripts/render-pdf.swift`, and review every slide at full size.

For source-region pixel comparisons, install
`scripts/requirements-visual-qa.txt` in a Python virtual environment and run
`scripts/compare-wave2-regions.py <render-dir> <new-output-dir>`. The report keeps
raw pixel differences; it does not turn a similarity score into blanket approval.
EnableComp icon previews use `scripts/render-svg-preview.swift`; the source SVG
and derived PNG hashes remain linked in `samples/visual-wave2/response-assets.json`.

PowerPoint operations are serialized. Close unchanged task probe decks after
measurement. Do not close or overwrite unsaved user presentations. If file access
or export is blocked, resolve that state before claiming a native result.

## Text

Use `paragraphs` with stable paragraph/run IDs instead of the legacy `text` and
uniform font fields. Runs support Arial, size, bold, italic, single underline and
RGB colors. Paragraphs support left/center/right alignment and point spacing
before/after, plus `line_spacing_multiple` (0.5–4; default 1). Use the source’s
measured line spacing: UHG28 captions use 0.9. Newlines belong between paragraphs,
not inside runs. Font files and
every run/paragraph property bind the native measurement cache.

Mixed formatting is supported on canvas text and layout blocks. Native bullet
paragraphs support `bullet: {character: "•", margin_left_pt: 23.04, hanging_pt: 22.5}`
with glyphs `•`, `–`, and `▪`, left alignment and first-run Arial/color at 100% size.
Hanging indentation must reserve at least 0.75 times the first-run point size.
Numbering and custom bullet fonts are unsupported. Indentation is structurally
verified; native character bounds exclude the bullet glyph, so visual review is
required. Existing separate uniform marker/text blocks remain available. Do not flatten rich text through
`recover-text`: rich text edits need regeneration or the scene workflow.

## Pictures

Every image retains a local path, SHA-256, alt text and explicit frame in points.

| `image_fit` | Behavior |
| --- | --- |
| `preserve` (default) | Reject frame/source aspect mismatch over 0.5%. |
| `contain` | Preserve the full source using transparent extent in the frame. |
| `cover` | Fill the frame without distortion; optional `focal_x/y` in 0–1. |
| `source_crop` | Reuse explicit fractional left/top/right/bottom source edges; reject distortion. |
| `stretch` | Deliberate distortion, required for the distorted UHG28 source picture. |

PNG/JPEG bytes stay unchanged. A source reference with a real portrait must retain
its actual identity; synthetic people use initials/placeholders. UHG deliverable
screenshots are labeled illustrative source examples. They are not evidence of
completed work for a new proposal.

The original v7 deliverable fixture approximates picture borders with four native
lines. The follow-up uses `outline_color` plus `outline_width_pt` for a native
picture outline (width up to 6pt). Do not treat a structural pass as proof of
pixel identity. Caption/picture pairing is a spec transform, not a native group.

## Roadmaps

`kind: "shape"` supports `rect`, `homePlate`, `star5`, and `rightArrow`. Use a
concrete RGB `background` or `pattern: {preset: "wdUpDiag", foreground: ..., background: ...}`.
For homePlate/rightArrow, `adjustments: {adj: 39542}` illustrates the bounded
0–100000 adjustment range. Outlines and text inside a shape are unsupported;
declare a separate text element and explicit overlap/layer relationships.

The source roadmap groups are flattened into editable native objects. Its month
header is rectangles/text, not a table. Exact source bar geometry does not imply
exact source glyph placement or general date-driven scheduling. The fixture
generator demonstrates declared interval and milestone transforms only.

## Wave 2 follow-up

See [follow-up qualification](../../planning/WAVE2_FOLLOWUP.md) for the four-page
review covering bullets, biography spacing, native outlines, and SVG icons.
Regenerate with `python3 scripts/build-wave2-followup-spec.py`. The original
`proof.json` remains historical v7 evidence; `followup-proof.json` records the
new fixture evidence after native/visual qualification.

Static SVG pictures require `asset_path`/`asset_sha256` for the original plus
`fallback_asset_path`/`fallback_asset_sha256` for a reviewed PNG fallback. Both
media are embedded without changing their bytes. This keeps the vector artwork
as an SVG picture; it does not convert paths into editable PowerPoint shapes.
Unsupported external/font dependencies, active content and dimensional ambiguity
are rejected. The fallback's aspect ratio must match within 0.5%. Visual review
must confirm that the fallback depicts the same artwork.
