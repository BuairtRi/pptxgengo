# Measured proposal layouts — Wave 1

This slice replaces individually positioned text cells with bounded grids and
panels. It is a proposal-layout foundation, not a general automatic designer.

## Accepted fixture

The [proof](proof.json) records six native-verified, visually reviewed slides:
423 objects, 298 text blocks, no fit/planner failures, and unchanged glyph bounds
for all 174 control text blocks. The exact supported variants are qualified;
this is not broad approval for arbitrary content or layouts.

## Components and specimens

- `controls.json`: five-phase approach, phase-detail page and workflow matrix,
  using the accepted dense deck's exact copy and typography.
- `variants.json`: four-row matrix with longer copy, six-row matrix with new
  illustrative content, and a nested parent-panel movement.
- `review.json`: six-slide controls/variants comparison with sequential page numbers.
- `overflow.json`: two deliberately overfull cells; must report both and reject build.
- `tokens.json`: named WM colors/type/spacing resolved by the fixture authoring
  helper into explicit point/RGB values. General runtime token inheritance is
  not implemented.
- `scripts/build-layout-wave1-spec.py` instantiates the saved examples. The Go CLI
  computes block positions and measured shared row heights.

The approach control intentionally changes the five “Evidence / output” labels
from blue to navy. Checking the inherited gray surface revealed only 4.33:1
contrast for the old blue. Copy and font sizes remain unchanged. Source-faithful
comparison reports must identify this correction rather than claim pixel identity.

## Contract

`SlideSpec.layouts[]` contains bounded `ContainerSpec` panels with row/column
tracks and named cells. Each cell has independent four-side padding and a flow
of named text blocks. Blocks support measured height, minimum/maximum height,
gap, alignment, and a paired bullet marker with hanging indent. Width is resolved
before probing; row height is the tallest cell's required height. All emitted
objects are ordinary editable native text/shapes.

A nested panel's bounds are relative to its parent's **padded content origin**.
It must remain inside that content area. Moving the parent moves its descendants.
A panel's `layer` is relative to its parent; cell surfaces precede their text.
Native row rules follow actual resolved row boundaries. Explicit column gutters
may be uniform or an array of per-boundary values. Body rows grow to measured
requirements within min/max bounds; unused vertical space is not automatically
stretched into rows. Row weights do not distribute spare height.

The author must declare supported maximum slide/panel sizes. The planner does not
shrink type, cut content, invent rows, or spill overflow into a second slide.
`fit-report` returns fixed-zone results and separate `layout_failures` with named
cells. Dynamic placeholders used during probing are never reported as real capacity.

Color contrast resolves through block fill, cell fill and ancestor surfaces.
`contrast_background` carries the effective color for transparent text; it does
not paint a rectangle. This is structural inheritance, not a pixel-background
sampler for arbitrary overlapping artwork.

## Native measurement cache

```sh
go build -o /tmp/pptxcompose ./cmd/pptxcompose
/tmp/pptxcompose probe --spec library/layout-components/controls.json --cache samples/layout-cache --out samples/layout-control-probes
/tmp/pptxcompose measure --bundle samples/layout-control-probes --cache samples/layout-cache --out samples/layout-control-evidence.json
/tmp/pptxcompose fit-report --spec library/layout-components/controls.json --cache samples/layout-cache --out samples/layout-control-fit.json
/tmp/pptxcompose build --spec library/layout-components/controls.json --cache samples/layout-cache --out samples/layout-control-deck
/tmp/pptxcompose verify --bundle samples/layout-control-deck --out samples/layout-control-verification.json
```

Use fresh output names. Native automation is serial. A probe with all requests
cached produces `cache-report.json` and no PPTX; skip `measure` in that case.
Changed variants follow the same sequence and probe only missing unique contracts.
`cache-import` can register a previously measured probe bundle when its evidence
contains the current environment fingerprint; legacy evidence cannot be upgraded
by inference.

A cache key binds exact text, width, font/style/alignment and color/inset contract,
plus macOS/PowerPoint versions, resolved Arial font file hashes, adapter and
inspector hashes, and the compiled CLI hash. The binary hash is deliberately
conservative: recompiling changed CLI code invalidates reuse even when a change
was unrelated to typography. Copy edits under one tool version benefit from reuse.
IDs locate slots but are excluded from the visual contract, enabling reuse across
slides. Cache records retain original bundle/evidence/request IDs and hashes;
lookup revalidates original native v5 observations and never trusts edited summary
dimensions. Final deck verification and per-slide visual review remain mandatory.

The fingerprint captures the configured local renderer, not every possible
PowerPoint preference or a portable cross-machine font guarantee. Unsupported
fonts/styles require extending both the probe contract and environment audit.

## Revision boundary

Known-deck `recover-text --text-only` can map generated layout block text back to
its named slot. Derived bullet-marker edits are rejected. Layout or formatting
changes remain on the explicit scene/reconciliation path. Recovered copy requires
new measurements; unchanged contracts can use the verified cache.

## Reproduce acceptance checks

Generate the specimens with `python3 scripts/build-layout-wave1-spec.py`. Probe
and measure `samples/layout-wave1/stress-probes-spec.json` with one fixed CLI build
and the cache. It includes the review deck and deliberately overfull specimen.
Build `controls.json` and `review.json`; run `fit-report` for review and overflow.
The overflow build must fail without leaving an output bundle.

The local proof workflow also uses:

- `scripts/check-layout-wave1-geometry.py`: shared row heights, parent movement,
  row count and original-copy preservation against the earlier dense deck.
- `scripts/check-layout-wave1-cache.py CLI`: evidence/contract corruption and
  legacy-schema rejection, plus reconstruction from native rows. Its local output
  directories must be new; preserve the original measured binary.
- `scripts/check-layout-wave1-copy.py`: compare native glyph bounds for the three
  control slides against the previous dense deck after final verification.
- `scripts/report-layout-wave1.py`: bind final deck, specs, native verification,
  render review, cache checks and geometry checks into the tracked proof.

These integration helpers expect the named local artifacts under
`samples/component-adaptation/dynamic-pods` and `samples/layout-wave1`. Binary
PowerPoint/PDF/render artifacts are intentionally ignored by Git. The proof records
their hashes; it does not make them available in a fresh checkout.

## Additional implementation limits

Final review found that direct canvas/row-rule layers and accumulated nested layer
sums do not yet share complete range validation. Use the tested explicit layer
values; broad layer validation remains a release task. Multi-row spans allocate
extra growth to the final spanned row and can reject a layout for which another
allocation would fit. Nested container contrast follows parent surfaces, not an
implicitly intersected parent-cell fill. The current fixtures avoid these cases.
Cache-backed fit reports retain per-contract provenance in `cache_uses`; legacy
single-probe hash fields are empty because a cache may combine several sources.
Keep cache writes serialized; concurrent import is not qualified.
