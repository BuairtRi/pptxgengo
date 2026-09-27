# Shared fit checks for fixed source templates

## Decision

Use PowerPoint's measured text geometry to check whether visible text stays
inside its existing shape's usable frame. Treat the result as a geometric
preflight, not a measure of how much copy a shape can support and not visual
approval. `--strict` reports a failing exit status for definitive overflow. The template
builder does not yet enforce this checker automatically; callers must wire that
gate into their workflow. Copy is never shrunk automatically.

The first reusable piece is now available as
[`scripts/check-template-fit.py`](../scripts/check-template-fit.py). It reads
the JSON produced by [`scripts/measure-template-frames.applescript`](../scripts/measure-template-frames.applescript),
compares text bounds with the inner native shape frame and margins, and writes a
per-shape JSON finding. It defaults to reporting findings without a failing
process status. `--strict` returns nonzero when any measured text overflows;
invalid evidence returns nonzero regardless. It accepts `--native-deck` to pin
top-level shape order, name, full text, position, extents and rotation to the
candidate PPTX's OOXML. This selects a coordinate scale only when the native
measurement matches a recognized PowerPoint-point or EMU scale. The result
includes hashes for the frame report and candidate deck. `--scene` accepts a native
frame report captured from the untouched source deck and compares shape paths,
names, margins and geometry. Nested group/table shapes remain inconclusive
unless both reports carry an explicitly verified transform chain. The checker
does not run PowerPoint.

## Existing measurement and preflight

There are two related but separate systems:

1. The `pptxcompose` workflow in `cmd/pptxcompose/main.go` probes generated
   compositions in PowerPoint using
   [`scripts/measure-compose-text.applescript`](../scripts/measure-compose-text.applescript).
   The v8 adapter returns per-shape frames, text-range bounds, margins,
   paragraph data, character font attributes and each visible character's
   non-whitespace bounds. `checkNative` verifies the probe against the known
   generated manifest: shape identity, exact text, font, margins and frame.
   In final-fit mode it rejects text bounds that cross the frame's inner safe
   zone (0.15pt tolerance) and retains measured character bounds for phrase
   placement. `fit-report` emits fixed text-zone findings and dynamic layout
   failures; planning still checks panels, cards, roles, accents and routes.
   This path knows the geometry because it generated the deck from the spec.

2. The template review adapter
   [`scripts/measure-template-frames.applescript`](../scripts/measure-template-frames.applescript)
   already walks slide shapes, recursively walks group children and visits
   table cells. It reports the text range, shape left/top/width/height,
   rotation, margins, inner frame dimensions, and text-range left/top/width/
   height. It preserves per-object scripting errors in the report. Before this
   change it emitted `qualification: measurement_only`, but had no shared
   checker, did not emit a versioned schema, and did not itself distinguish
   top-level, grouped, and table-cell coordinate scopes. Its text bounds are
   PowerPoint text-range bounds, not character-union bounds or rendered raster
   ink bounds.

The current rollout contains 65 contracts (23 architecture/product, 21
approach/narrative, 21 evidence/people), with 895 semantic slots and 1,570
editable text bindings. The source binding guides show 168 bindings inside
grouped shapes and 171 inside table cells. Those nested bindings make source
geometry resolution a first-class requirement: treating every shape's left
and top as slide-relative can produce false passes or failures.

## Proposed complete fit contract (beyond the current checker)

1. **Pin the exact item.** Every preflight report carries the candidate deck
   hash, source presentation hash, optional scene hash, measurement-adapter
   version, PowerPoint version/platform, timestamp, coordinate units, and
   tolerance. A pinned untouched-source frame report is a geometry reference,
   not a replacement for matching the edited candidate to its source binding
   IDs. Require a unique slide/path/name identity; duplicate names alone are not
   identities.

2. **Measure the whole fixed frame.** For each visible text-bearing shape,
   capture shape path, name, text, shape frame (x/y/width/height/rotation),
   text-range bounds, non-whitespace character-bound union where PowerPoint
   exposes it, all four text margins, shape type, containing group path, and
   table row/column if applicable. Keep range bounds and character union
   separate. A missing property or AppleScript error is an `inconclusive`
   finding, never a pass. Empty shapes have no text-fit result.

3. **Resolve coordinates before comparison.** Top-level shapes can
   compare measured slide-coordinate bounds directly against
   `[left + margin_left, top + margin_top, right - margin_right, bottom -
   margin_bottom]`. Group children require the accumulated group transform
   (child offset/extent, parent child offset/extent, and parent rotation).
   Table cells require the native cell's own frame plus cell margins and the
   enclosing table transform. For rotated shapes, transform the text-bound
   rectangle corners into the shape's local frame; the checker uses that
   conservative rectangle test when the report gives a resolved frame.
   If report and pinned scene do not establish the same full path and exact
   geometry, label the case `inconclusive`. Identity pinning alone does not
   validate an unhandled transform.

4. **Use a small explicit tolerance.** The starter checker defaults to 0.15
   native coordinate units, matching the compose final-fit tolerance when the
   adapter reports points. Do not assume all versions of PowerPoint scripting
   use points: the current frame adapter labels its coordinates “raw native
   PowerPoint units” because the installed dictionary does not specify units.
   Until source-frame measurement confirms the unit scale, leave the tolerance
   explicit in the report and do not claim point-level accuracy.

5. **Bound the response.** A measured text overflow has status `overflow`,
   includes each edge's excess and the shape/slot identity, and causes strict
   build preflight to stop. Offer an actionable path: edit the copy, choose a
   reviewed shorter variant, or request a deliberate geometry/style change.
   Do not silently drop words, reduce font size, change line spacing, or resize
   the source shape. An optional approved typography profile may be a separate
   user-visible choice and must receive its own measurement and review.

6. **Do not overclaim.** `fits` means only that the reported text rectangle is
   contained by the measured usable frame within tolerance. It does not prove
   good line breaks, contrast, readability, semantic pairing, no collisions
   with neighboring objects, adequate optical padding, content capacity,
   image/chart/table legibility, or cross-platform equivalence. It does not
   qualify a template for arbitrary replacement content.

## Exact limits and extension points

- `scripts/measure-compose-text.applescript` is the strongest existing source
  for native character bounds and effective character styles, but its output
  is tied to generated compose probe shapes and its manifest. Reuse its
  per-character union and attribute collection patterns; do not reuse
  `checkNative` as a generic imported-template checker without a source scene
  resolver.
- `cmd/pptxcompose/main.go:checkNative`, `compose.FixedTextFitReport`,
  `compose.LayoutFitReport`, and `fit-report` already provide measured fit
  reporting for the composition DSL. Keep those checks as the compose path's
  preflight; the new frame checker is the narrower source-template path.
- `scripts/measure-template-frames.applescript` now emits a schema version,
  explicit `coordinate_scope`, table cell coordinates, complete `shape_frame`
  and `text_bounds` objects, margins and retained error records. It still does
  not output effective font/paragraph/run styles, character bounds, source
  binding IDs, visibility status or resolved group transform chains. Group and
  table rows remain inconclusive until a pinned report includes verified
  transforms.
- Bounds from PowerPoint APIs are not rasterized ink. Glyph overhang, antialias
  pixels, clipping, bullet glyphs, text effects, theme fallback fonts, variable
  font substitution, and some rich-text baseline/vertical alignment behavior
  can differ from these geometric rectangles. A screenshot remains necessary
  for final fit and quality review.
- Native PowerPoint measurement is platform/version dependent. Frame-report
  cache keys must include candidate/source hashes and measurement environment;
  never reuse values across changed text, frames, margins, fonts, or adapter
  revisions.
- Source template reports are separate from implementation reports. A lane's
  `completed` count means contracts and sample values exist and pass inspection
  and value checks. Native fit and visual approval live in the review ledger.

## t030 speech-bubble note

The source scene for `graphics-and-layouts:117` has three freeform speech
bubbles (objects 6, 7, 8), but only one existing text box (object 10) over the
left bubble. The two empty freeforms have no editable text bindings. The safe
extension is to keep all three bubble geometries and colors, then add two
separate transparent text-box overlays using object 10's white type,
paragraph treatment and margins as the style donor. Place text inside the
visible interior of each overlapping bubble and keep its tail clear. Treat
these as new objects with new fixed frames and explicit illustrative content;
record the component as partially source-bound and the overlays as authored
additions. Measure the actual overlay geometry in PowerPoint and review the
native render before claiming the three-voice composition fits. Do not convert
the empty bubbles to editable text areas by mutating their source freeform
shapes unless the serializer can preserve their original drawing geometry and
style exactly.
