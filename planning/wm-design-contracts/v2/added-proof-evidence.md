# Added v2 Proof and Evidence templates

## Scope and evidence

This slice implements all 12 added Proof and 8 added Evidence templates from the
frozen source commit `7bdcaee5030a12275a1f881a8542f4d302d207df`. Source definitions,
source hashes and the v1 bundle remain unchanged. Closed bindings require exact
scalar slots and exact-count stable key arrays; navigation takes 2–6 caller keyed
sections and an active key. Named annotation targets remain structural IDs.

Each template has a source-content specimen and a changed-content specimen.
All numeric claims, client aliases, quotations, dates and miniature exhibits are
synthetic illustrations. Source examples are retained as reference copy, not
endorsed client evidence. Alternate source notes explicitly identify illustration.

Go compilation and all 40 specimen generations succeeded. All 40 paired specimens on integrated pages 39–78 passed native visual review.
The initial complete contact-sheet review was followed by full-size inspection
and rechecks of corrected pages 52/54/64/66/72/76, then final fixed-row pages 51–54.
No tests were added or run. Acceptance applies only to these paired specimens;
see the family `native-review.json` receipt and final artifact hashes.
The generation flags remain `powerpoint_verified:false` and
`visually_reviewed:false`, separate from any subsequent review receipt.

## Named geometry amendments

- `wmds.quote-outcome-quote-capacity.v2`: `case-study/quote-outcome` node06
  moves from y324 to y288, increases height126→162 and sets padding 9. It retains
  the original width 270, fonts, quote mark and caller copy; its bottom stays 450.
  The authored 126 pt card overflowed by 42 pt. The preceding bullet list ends 255,
  leaving 33 pt before the revised card.
- `wmds.column-full-split-source-clearance.v2`: `chart/column-full-split`
  node01 height432→414 retains y36 and the source note reservation; chart bottom
  becomes450. Chart categories, series, values, title and units stay editable.

Shared renderer work by the integrating agent restores the v2 navigation tab's
source font8pt/600 Mono and its bounded horizontal clearance. Shared binding work
keeps table group `from`/`to` indices structural, while group labels stay content.
Those changes resolve source navigation and alternate table-generation failures.

## Fixed row text capacities

`wmds.what-we-did-row-clearance.v2` applies to both
`case-study/what-we-did` and `case-study/what-we-did-split`. Textblocks02/03
at y132 and05/06 at y204 have fixed 57 pt height. Textblocks08/09 at y276
have fixed 75 pt height. Each allocation reserves 9 pt above the following separator
or Results label, while preserving original positions, widths, styles and copy.

The shared textblock renderer checks every measured text part against the
explicit outer height and returns `scene.textblock_overflow` for excess content.
The first two left rows therefore fit a one-line subhead plus one-line body;
the last left row can fit a one-line subhead plus two-line body at the original
styles. Label/body rows use the same height limit and their own token measures.
These limits are enforced on source and bound inputs, not just recorded in docs.
No shrink, ellipsis or specimen fallback resolves unsafe copy. Negative controls
were not run; code inspection and successful source/alternate generation provide
the current implementation evidence.

The alternate matching sentence was edited to preserve its 12 busiest accounts
and automated-matching meaning within the one-line allocation. This caller-copy
choice does not alter the source example or replace its contract with a fallback.

## Explicit specimen illustrations

The bound templates keep the source thumbnail topology and do not acquire fixed
example fallbacks. The composed reference documents replace silhouette thumbnails
with blank page surfaces and add editable native miniature charts, close-step
rows, flow stages or pilot-plan labels. These overlays are identified by
`wmds.proof-evidence-illustrative-preview.v2` and `/specimens/...` pointers.

`deliverables/annotated` reuses the reviewed native scorecard specimen: headline
readiness and pilot decision, close-step baseline, then accountable actions and
dates. Zones1 and2 receive heights72 and90 solely in the explicit composed
specimen. The caller contract remains the source's three named frame labels and
six annotation label/body slots.

Alternate bindings exercise chart values/categories, quadrant coordinates,
metric values, close-process explanations, table rows, checkbox feature flags,
quotes, attribution, risk owners/dates and deliverable descriptions. Intentional
empty sample cells, WM asset IDs and status enums remain valid. The records list
exact changed body slots; successful generation establishes these samples only,
not arbitrary-copy qualification.

## Content contracts

| Template | Required slots | Exact key-array counts | Navigation | Changed body slots |
|---|---:|---|---|---:|
| `case-studies/matrix` | 33 | node01.cols=5, node01.rows=5 | None | 17 |
| `case-study/exhibit` | 30 | node01.categories=4, node01.series.item01.values=4, node01.series.item02.values=4, node01.series=2, node04.items=3 | None | 13 |
| `case-study/exhibit-split` | 28 | node01.categories=4, node01.series.item01.values=4, node01.series.item02.values=4, node01.series=2, node04.items=3 | None | 12 |
| `case-study/quote-nav` | 16 | None | 2–6 keyed sections | 8 |
| `case-study/quote-outcome` | 22 | node05.items=3 | None | 8 |
| `case-study/timeline` | 27 | node05.items=4 | None | 6 |
| `case-study/what-we-did` | 23 | None | None | 10 |
| `case-study/what-we-did-split` | 23 | None | None | 9 |
| `chart/column-full-split` | 23 | node01.categories=6, node01.series.item01.values=6, node01.series=1, node03.items=2 | None | 9 |
| `deliverables/annotated` | 12 | None | None | 6 |
| `deliverables/index` | 28 | node01.cols=5, node01.rows=4 | None | 8 |
| `deliverables/sample-grid-nav` | 23 | None | 2–6 keyed sections | 8 |
| `deliverables/walkthrough` | 16 | None | None | 5 |
| `image-text/frame-right-split` | 11 | node04.items=2 | None | 1 |
| `quadrant/numbered-legend` | 35 | node01.items=6, node02.items=6 | None | 12 |
| `quadrant/subtle-left` | 44 | node01.items=8 | None | 16 |
| `status/classic` | 61 | node01.cols=1, node01.rows=1, node02.cols=3, node02.groups=1, node02.rows=4, node04.items=4, node06.items=4, node07.cols=5, node07.groups=1, node07.rows=3, node09.items=4 | None | 7 |
| `status/milestones-workstreams` | 55 | node01.cols=5, node01.groups=1, node01.rows=5, node03.items=3, node05.steps=4 | None | 5 |
| `status/raid-log` | 63 | node01.cols=6, node01.rows=9 | 2–6 keyed sections | 11 |
| `status/steering-update` | 38 | node03.items=3, node04.body.item01.bullets=3, node04.body=1, node05.cols=4, node05.groups=1, node05.rows=3, node07.items=3 | 2–6 keyed sections | 10 |

## Reproduction and artifacts

Run from repository root with a newly built CLI and a new output directory:

```sh
python3 scripts/wmds-refresh-proof-evidence.py \
  --cli /tmp/wmds-proof-evidence-pptxdesign \
  --out /tmp/wmds-proof-evidence-new
```

Canonical family artifacts live at
`samples/wmds-refresh-slice3-20261002/work/proof-evidence/`:

- `combined.foundation.json`: 40 interleaved source/alternate composed slides.
- `bound-content.json`: all20 exact caller bindings for changed content.
- `contracts.json`: the closed catalog definitions used by the generator.
- `generation-receipts.json` and `source-generation-receipts.json`.
- `reference.pptx` and `layout-report.json`: generated family review packet.
- Per-template directories contain two-slide documents, bound content, receipts,
  editable PowerPoint files and layout reports.

No native capture, measurement alias, font calibration update, source bundle edit
or commit is required by this generation workflow.
