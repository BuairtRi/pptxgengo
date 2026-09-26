# Wave 1 — measured layout implementation

Status: six-slide fixture accepted after native verification and two visual
reviews. Exact evidence and artifact hashes are recorded in
[the fixture proof](../library/layout-components/proof.json).

## Changes

- `internal/compose/layout.go` lowers named containers/cells/blocks into native
  canvas shapes in two passes: widths/probe requests, then measured block and row
  heights. It preserves the input spec, explicit parent padding, named content
  slots, bounded tracks, nonuniform gutters, flow gaps and marker/text pairs.
- Canvas layers are ordered in final rendering. Phase backgrounds precede canvas
  annotations; panel/cell surfaces precede their content. Transparent text has a
  separate effective background for contrast checks.
- `LayoutFitReport` records capacity failures, including every cell exceeding its
  bounded row. Existing fixed-zone reports never mistake 1pt probe placeholders
  for actual dynamic capacity.
- Native observations can be cached by complete rendering contract and the local
  environment. Cache entries retain and revalidate original observations rather
  than copied dimension summaries. Changed content is measured independently.
- Known-deck text recovery supports named layout block slots; derived marker edits
  remain unsupported and explicit.

## Fixtures

The controls preserve the accepted dense deck's first three slides: five-phase
approach, phase detail and five-row workflow matrix. This is a conversion control
against `proposal-complexity-v2`, not a new claim of exact UHG reconstruction.
The controls retain all 174 text blocks. The full six-slide spec contains 298
text blocks and 423 editable objects, with zero fixed-zone or layout failures. The five output labels changed from blue to navy
because inherited-background validation found 4.33:1 contrast on the gray surface.

The review deck pairs these controls with:

1. Four workflow rows, including a longer paragraph that requires row growth.
2. Six workflow rows with new illustrative content.
3. A nested five-phase panel moved four points by changing only its parent origin.

A separate deliberately overfull matrix must reject generation and identify both
affected cells. Native probe/cache evidence, final deck verification, render review,
and geometry comparisons are local artifacts referenced by the final proof.

## Structural regression coverage

Permanent Go tests cover non-mutating expansion; tallest-cell row sizing; parent
translation; all affected capacity failures; cycle rejection; padded nesting;
transparent inherited contrast; cache-contract field coverage; identical-contract
miss deduplication; environment invalidation; semantic-slot text recovery; and
phase-background/explicit-layer order. Unit fixtures use supplied dimensions and
are not substitutes for native font measurements or visual review.

## Parallel work

A coding subagent implemented the bounded layout engine. The lead integrated CLI
rendering, cache provenance, named-slot recovery, regression fixtures and the
proposal examples. A Luna agent reviewed cache invalidation and prepared the
Wave 2 visual asset contracts. The lead caught a roster-count discrepancy in that
inventory; the corrected record contains all 19 portrait/card pairs.

## Deliberate limits

- Blocks are still uniform-style Arial text; rich paragraphs/runs remain Wave 2.
- Grid cells are editable native shapes, not PowerPoint table objects.
- Parent relationships are in the composition spec; manually dragging one native
  shape in PowerPoint does not move its semantic descendants automatically.
- Explicit bounds/minimum heights encode the supported design. General page
  splitting, aesthetic layout selection and arbitrary freeform resizing remain
  later work.
- Runtime style tokens are resolved by the fixture helper, not a general CLI token
  registry. Weighted columns share available width; rows grow to content and do
  not distribute unused height by weight.
- Cache compatibility includes the full CLI binary hash. A changed build forces
  fresh observations; repeated authoring under the same build benefits from reuse.

## Completed integration checks

- Shared-height growth: the longer first matrix cell grows its row from 62pt to
  86.60000038147pt; all four columns share that height.
- Six rows fit within the declared 479pt bottom boundary (actual bottom 477pt).
- Parent translation moves 75 descendants down exactly 4pt with unchanged size.
- Two deliberately oversized cells are reported together; build rejects without
  leaving an output directory.
- All 298 review text contracts are cache hits. One changed control string creates
  one new probe while reusing 173 unchanged bindings; that probe was measured in
  native PowerPoint and imported into the cache.
- Six integrity fixtures passed: modified contract, changed evidence bytes,
  missing environment, old native schema, altered request/deck binding, and edited
  summary dimensions. The final case reconstructs the original plan from native
  rows instead of trusting summary dimensions.
- All six full-resolution rendered PNGs were reviewed by the primary agent and an
  independent Luna reviewer. A partial-preview false alarm was resolved against
  the full file; see QA74. Final native object acceptance is recorded in proof.json: maximum frame
  deviation 0.0000106791pt; maximum measured text-bound excursion 0.00503922pt
  (within the 0.15pt tolerance). All 174 control glyph bounds exactly match the
  earlier dense deck.

Final code review found no issue affecting these fixtures. Broader layer validation,
row-span allocation, implicit parent-cell contrast and concurrent cache import are
not qualified; see the layout README and QA75. Per-contract cache provenance is
present even though legacy single-probe fields in cached fit reports are empty.

## Deliverable

[Six-slide PowerPoint review deck](../samples/component-adaptation/dynamic-pods/wave1-layout-review/wave1-layout-review.pptx)
and [PDF](../samples/component-adaptation/dynamic-pods/wave1-layout-review.pdf).
Pages 1–3 are dense controls; pages 4–6 exercise content growth, row cardinality
and parent movement. These files remain local and ignored, while the source specs,
implementation, QA scripts and proof are tracked. Wave 2 source extraction is also
prepared: 33 pinned assets, with component adaptation still pending.
