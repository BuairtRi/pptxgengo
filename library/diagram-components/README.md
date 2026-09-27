# Wave 3 diagram qualification fixtures

Run `python3 scripts/build-wave3-diagram-spec.py` from the repository root to regenerate `review.json`, `overflow.json`, `art-translation.json`, and `source-evidence.json`. The generator pins the original UHG PPTX SHA-256 and extracts UHG14 artwork from that source. It uses `sips` to make the SVG's PNG fallback for native Office rendering.

`review.json` has eight pages: UHG36 source-control process geometry and wording, two synthetic five-step paths, four- and six-node path stresses, a translated parent, a synthetic architecture relation smoke, a six-role team relation smoke, and the UHG14 pinned-art placement control. The architecture and team pages are unqualified mechanics checks; `dense-review.json` provides the substantive layered architecture and 21-role delivery organization candidates described below. `overflow.json` deliberately exceeds a measured node width. `art-translation.json` shifts the two source rasters and one SVG together by +8 x/+6 y points while retaining their hashes and relative offsets.

The source UHG36 grouped node frames were resolved into slide coordinates because the original group uses unequal x/y scaling. `#F900D3` is the exact source decorative number color at 16 pt bold. The source-control page preserves the two pathway labels, ten node texts, result text, node frames, and arrow gaps; it is a bounded reconstruction, not a full slide extraction. The UHG14 board's internal labels and arrows remain raster content. Its separate SVG artwork is pinned visually and has no inferred semantic endpoints.

The Go compose tests and `pptxcompose probe` validate structure, route contracts, declared assets, and synthetic geometry. Native text measurement, final build, source render comparison, and visual review remain required before approving a reusable component. In particular, the source-control page uses the measured grid's fixed Arial text style and does not assert pixel-perfect reproduction of the original grouped shape typography or every decorative element.

## Dense and phrase candidates

`build-wave3-dense-spec.py` now supplies `dense-review.json`: a nine-node,
three-layer architecture with governance/security bands and the 21-role delivery
organization with three pods and distinct advisory/dependency/annotation links.
These replace the sparse mechanics pages as delivery candidates. Their native
fit and visual review remain pending; they are not approved by the earlier smoke.

`build-wave3-accent-spec.py` supplies source-copy, changed-width, repeated-phrase,
per-line underline, rotated highlight, and ambiguous-phrase manual staging cases.
Phrase selection uses exact text plus occurrence or a Unicode code-point range.
Native character bounds supply each wrapped fragment. Final verification checks
that fragments have not moved from the measured positions. A manual placement
note and selected artwork remain visible, and the deck is marked unfinished.

## Arrow candidates and collision geometry

`build-wave3-arrow-catalog.py` records original WM descriptions, original SVG
hashes, visible endpoints/tangents, and annotated calibration previews. The source
connecting loop remains a manual candidate: its real tail and tip are close
while the path extends far away. Its rectangular image frame is not an endpoint.

`build-wave3-arrow-spec.py` creates nine candidate placements (three arrow styles,
short/long and rotated cases) plus an incompatible-direction manual fallback.
The solver uses uniform scale and rotation; it does not stretch or flip artwork.
Declared source/target ports and phrase anchors remain explicit. These proposed
experimental envelopes are not production limits until native visual acceptance.

For a curved arrow, a rotated bounding rectangle includes large empty corners.
The fixture generator derives conservative 32×32 occupied alpha tiles from a
720×720 raster of the exact SVG, dilated by one raster pixel. The plan transforms
each tile with the artwork and checks it against labels, images, other accents,
and connector strokes. Geometry without tiles retains conservative bounding-box
checks. The source SVG, raster hash, and method are recorded. Tiles and endpoints
are vetted asset-contract metadata; native frame verification alone does not
independently prove their optical accuracy. Native rendering and review remain
mandatory.

Several stock PNG counterparts have different canvas proportions from their SVGs.
The candidate generator therefore derives a pinned fallback from the exact SVG
at its original aspect ratio. Original SVG bytes are retained in the PPTX. Native
SVG and fallback picture relationships, bytes, frame, crop, and rotation receive
structural checks; the fallback is not silently stretched to fit another canvas.

## Current evidence boundary

All Go tests pass at the implementation checkpoint. The nine-arrow package test
uses explicitly synthetic tiny text bounds only to exercise geometry and package
serialization; it is **not** native text fit or visual approval. The phrase spike
has native probe measurements. Final export/visual qualification remains pending
because PowerPoint PDF export is timing out. The application is not restarted,
and unsaved user decks are preserved.

## Expanded qualification inputs

Run `python3 scripts/build-wave3-expanded-review.py --proposal-bundle samples/proposal-authoring/full-capacity-v9`
after assembling the current normal proposal. The generator checks assembly,
narrative, values, contract and template hashes before deriving review inputs.

- `expanded-review.json`: 15 review pages covering seven dense patterns, phrase
  emphasis, three arrow styles and two explicit unfinished examples.
- `source-controls.json`: three bounded process/art-placement controls.
- `accent-arrow-qualification.json`: 29 cases, including unchanged-copy
  translations, second-occurrence selection, Unicode ranges, rich runs,
  per-line highlights, changed arrow endpoints and a dense team accent.
- `expanded-review-plan.json`: source hashes, intended comparisons and rejection
  expectations. Every native acceptance gate remains pending.
- `accent-ambiguous-negative.json`: repeated phrase without an occurrence/range
  or staging; public `probe` must reject it.

The new `artwork_arrows[].manual_only` option always stages the original pinned
artwork with a visible target note. It requires valid source/target identities,
a reserved staging area and measured note fit, but does not require invented
endpoint calibration or automatic span limits. `loop-manual.json` exercises this
policy with the WM return-loop asset. Even a calibrated asset stays manual when
this option is set. The result remains unfinished until placement is reviewed.

Structural tests use synthetic glyph bounds exclusively for geometry algebra;
they do not establish typography, native fit or optical quality. Current probe
bundles are `samples/visual-wave3/extended-accents-v3-probe`,
`source-controls-v2-probe` and `expanded-review-v2-probe`. They contain measurement
inputs, not finished presentation deliverables.
