# Frozen WMDS intake: capabilities and repair evidence

> **Repair-wave update:** all 522 templates now compile, and all 6,307 scene
> nodes pass. The 27-case queue below is the historical pre-repair result and
> is now closed. Native visual review and library promotion remain pending.
> Current evidence: `planning/wm-design-contracts/v4/intake-20261003-frozen/repair-wave/`.

## Target and inventory

The user reported that upstream authoring had stopped for now. The clean source
snapshot observed at **2026-10-03 18:55:07 UTC** is commit
`8f9f16ab8a7e2a6fff45a97627308668f4a12753`: **522 committed and working templates
across 24 families**. Compared with the accepted 248-template v3 library: **274
additions, 16 revisions, no removals**; 16 existing definitions moved families.
Compared with the preceding Round 12 observation: 22 additions and five revisions.
These canonical identity comparisons do not count a family move as an addition.

The immutable snapshot and its 35 verified file hashes are in
[observation.json](../../planning/wm-design-contracts/v4/intake-20261003-frozen/observation.json).
Renderer SHA-256:
`72057c4f23f077f095e7b6f22705b35bbb538eee55d9d9da6cf752f6cc10d20c`.
Earlier observations remain historical evidence. This snapshot is the current
implementation target; it is not yet a registered or installed library revision.

## Implemented in this slice

| Capability | Behavior and qualification boundary |
| --- | --- |
| Funnel and revised pyramid | Editable native bands, source level identities, inside/side copy, values and leaders. Legacy pyramid syntax retains its existing adapter. |
| Road | Native sampled road with round joins, exact 7pt centerline dashes, milestone pins, alternating labels/dates and actual ink bounds. |
| Bracket | Four orientations, native strokes, measured optional labels and ink bounds. |
| Cycle | Editable elliptical nodes and directed arcs; active/center/numbered states and quadratic feedback loops. Feedback-loop dash spacing uses a native preset approximation to the source 5pt/4pt pattern. |
| Standalone gauge | Editable semicircle segments, needle/hub, endpoints and value/caption. Palette segments remain template styling; values and visible copy are content bindings. |
| Revised team curves | Future v4 defaults to 241 cumulative samples with edge-clamped Gaussian smoothing; explicit zero smoothing uses monotone cubics and explicit Catmull retains that mode. Frozen v1/v2/v3 defaults remain unchanged. |
| Revised Venn and maturity | Exclusive-region label anchors, source mono sizing, region nudges and bounded badges; dense maturity curves lift the last three labels. |
| Slim frames and tint | Slim footer geometry, up to four title lines where declared, vertically centered footer copy, palette-backed tint bands and tint-aware shared chrome caching. |
| Tables, statuses, callouts | Scalar numeric cells; source string/boolean row-header truthiness; additional work-item status colors; inverse callout surface. Previously implemented heat, subtitle and score-cell contracts remain available. |
| Value/investment line charts | Up to 60 samples for future line charts, sparse empty axis labels, and signed values with explicit finite yMin. Missing values still require a renderer enhancement. |
| Integration | Scene dispatch and local composition, typed copy/data bindings, fixed topology and structural discovery; optional native line-join enum with byte-neutral empty default. |

New planners reject invalid/nonfinite geometry, invalid indices and excessive
counts/copy. Source geometry and styling remain template-owned. Public build
entry points still enforce registered source pins; diagnostic source injection
is internal only. No arbitrary CSS or off-palette fallback was introduced.

## Repair results

**27 previously identified incoming repair cases are fixed**, including all
20 failures from the earlier 63-definition capability trial. Repairs cover Venn
insets, maturity stroke bounds, heat-table widths/row heights, metric/card fit,
empty callout eyebrows, cycle node heights and a source string-boolean amendment.
The unsupported dot-score glyph legend now uses editable native circles while
preserving its labels. Amendments are revision-gated, atomic and idempotent;
original content and the accepted v3 specimens are preserved.

The expanded frozen catalog introduces a **new** failure queue. These are not
reopened failures from that completed repair set.

## Whole-catalog and node diagnostics

- **522 attempted, 495 compiled, 27 rejected** in the full-slide diagnostic.
- **6,307 scene nodes audited, 63 node failures across 26 templates, zero
  unsupported scene types.** The node audit continues after an earlier failure,
  so it exposes multiple problems on a slide. The full-slide check also detects
  title/frame limits; its count is not interchangeable with the node count.
- Node failures: 43 card fit, 11 text fit, three footer collisions, two table-cell
  fit, and one each of table tag wrap, missing chart value, table width sum and
  table row fit. Complete failures are retained in `node-results.json`.
- The diagnostic combines frozen v3 fonts/assets with incoming source/frame
  semantics. 232 identical v3 specimens retain their qualified amendments.
  It is **not a complete accepted v4 source bundle**.

The 495-slide prototype is 33,540,928 bytes. ZIP CRC, 2,208 XML parts and all
3,339 internal relationship targets pass package checks. It contains 137 native
tables, 3,010 custom geometries and 15,528 text runs. Gauge XML is separately
covered because the scorecard's earlier table-width error excludes that slide
from the combined deck. Native PowerPoint visual acceptance remains pending.
Prototype binaries remain in temporary diagnostic directories; the current
sample/reference decks have not been replaced.

## Initial expanded-catalog repair queue (now closed)

Full-slide diagnostics stop at the first error; consult the node audit when
repairing a template to close every affected node.

| Template | First failure |
| --- | --- |
| `sequence/seven-r` | `scene.card_vertical_overflow: node05 content bottom426.000 capacity423.000` |
| `sequence/evidence-to-model` | `scene.card_vertical_overflow: node07 content bottom336.000 capacity333.000` |
| `stakeholders/quadrant-split` | `scene.row_text_overflow: node01.item-010.text` |
| `readiness/scorecard` | `scene.table_width_sum: node01` |
| `readiness/adoption-curve` | `scene.chart_native_missing_value_not_supported: Actual, wave 1/4` |
| `adoption/dashboard-split` | `scene.content_overlaps_reserved_footer: node05.native bottom 460.000pt capacity 450.000pt` |
| `key-message/stat` | `scene.card_vertical_overflow: node05 content bottom231.000 capacity225.000` |
| `key-message/stat-left` | `scene.card_vertical_overflow: node07 content bottom339.000 capacity333.000` |
| `key-message/stat-nav` | `scene.card_vertical_overflow: node05 content bottom231.000 capacity225.000` |
| `cards/narrative-2x2-badge` | `scene.text_overflow: node05.text needs 48.000pt, capacity 36.000pt` |
| `cards/narrative-2x3-badge` | `scene.text_overflow: node05.text needs 48.000pt, capacity 36.000pt` |
| `cards/narrative-2x3-rows` | `scene.card_vertical_overflow: node07 content bottom282.000 capacity270.000` |
| `cards/narrative-2x3-slim` | `scene.card_vertical_overflow: node05 content bottom264.000 capacity252.000` |
| `key-message/photo-icons` | `scene.card_vertical_overflow: node03 content bottom330.000 capacity312.000` |
| `key-message/icons-tint` | `scene.card_vertical_overflow: node03 content bottom168.000 capacity162.000` |
| `funnel/compare` | `scene.content_overlaps_reserved_footer: node06.text bottom 453.000pt capacity 450.000pt` |
| `road/cards` | `scene.card_vertical_overflow: node02.item-001 content bottom441.000 capacity438.000` |
| `timeline/vertical` | `title has 2 lines, capacity 1` |
| `timeline/swimlanes` | `scene.content_overlaps_reserved_footer: node21.text bottom 451.000pt capacity 450.000pt` |
| `phase/define` | `scene.card_vertical_overflow: node09 content bottom459.000 capacity456.000` |
| `phase/build` | `scene.card_vertical_overflow: node09 content bottom459.000 capacity456.000` |
| `activities/subtrack-timeline` | `scene.text_overflow: node40.text needs 36.000pt, capacity 30.000pt` |
| `activities/by-phase-table` | `scene.table_cell_overflow: node01.row.item-012.w needs36.000pt capacity26.000pt` |
| `activities/ownership-split` | `scene.table_tag_wrap` |
| `lifecycle/columns-outcomes` | `scene.card_vertical_overflow: node06 content bottom459.000 capacity456.000` |
| `roadmap/grid-columns` | `scene.card_vertical_overflow: node11 content bottom249.000 capacity243.000` |
| `value-tracking/planned-vs-realized` | `scene.table_cell_overflow: node06.row.item-001.p needs42.000pt capacity36.000pt` |

## Checks and next gates

Affected normal package checks and complete affected race checks pass: wmdesign
74.464s, deckproject 47.209s, pptxdesign 4.265s. Final focused gauge/line-join
race checks pass (pptx 1.388s, wmdesign 2.317s), including all six gauge tests. The
accepted **248-slide v3 source reference regenerates successfully**. Independent
reviews found and corrected tint-cache identity, direct-IR tint validation,
cycle rich-copy measurement and road stroke-join issues. These are code and
package checks, not native visual acceptance. Broader pre-existing `pptx` test
failures remain separate; no claim is made that the whole repository is green.
Exact receipts are in the frozen snapshot directory.

1. Close the 27 new full-slide cases and their 63 node failures, including native
   chart gaps. Preserve copy and address declared allocations instead of hiding
   overflow or replacing missing data with zero.
2. Native-review representative new capabilities and repaired specimens; refine
   any visual differences, including feedback-loop dash treatment.
3. Build a complete versioned bundle with source/assets, binding projections,
   SQLite catalog metadata and previews for all accepted templates. Structural
   repairs such as score legends need coherent user-facing bindings.
4. Regenerate the reference library and qualify the new release, then install it
   and update the skill/catalog together. Existing project pins remain explicit.

No new release was installed, no current reference deck was replaced, and no
commit or push was made in this intake slice.
