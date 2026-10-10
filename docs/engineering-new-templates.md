# Upstream template intake

## Current V12 intake — 2026-10-10

The frozen upstream commit is `36132d5637abdbdeb70945ad650b795cabea05ef`:
735 source / 734 active templates. This expands the earlier 677-template V12
candidate by 58 additions and two revisions. Against V11, there are 86 additions
and two revisions; all prior identities remain.

All source and active bound specimens build. The gallery publisher accepts
88 fresh PowerPoint reviews and 647 retained native previews verified by exact
composition and paired rendered dependencies. V12 is the new-project and release
default. Historical project pins and the original V12 intake are preserved.
See [current intake evidence](../planning/wm-design-contracts/v12/intake-20261010-735-frozen/README.md).

## Historical V12 implementation — 2026-10-08

The then-current upstream commit `efec671fe40d2145d14780dc39c3128bc9c65308` adds
28 templates: ten product backlogs, nine product strategy roadmaps and nine
resource plans. They are implemented as the separate pinned V12 candidate
(677 source / 676 active), with editable content bindings, discovery index and
native PowerPoint review specimens. Full source/bound generation and the Go
regression checks pass; primary-agent visual review passed all 28 additions.
Use `--bundle v12` with the repository-built CLI. Independent publication
qualification remains pending; the published default remains V11. See the
[V12 intake](../planning/wm-design-contracts/v12/intake-20261008-677-frozen/README.md)
for four explicit native allocations, browser metadata behavior and validation.

## Historical intake records

This file retains earlier intake/qualification snapshots. Its production and
upstream counts below describe their dated v10-era checkpoint, not the current
package. The V11 checkpoint had 649 source templates (648 active);
see the [frozen v11 intake record](../planning/wm-design-contracts/v11/intake-20261006-649-frozen/README.md).
Native render and visual qualification remain pending; see [current release
status](release-status.md).

## V10 intake qualification — 2026-10-06

The qualified production library is **v10 / 649 templates** (648 active), pinned
to `c14fb286fb38e15800a6fd476a1ed67956f1165f`. The 18 additions comprise five
pillar layouts, seven branching roadmaps and six narrative roadmaps. All 18 have
accepted native PowerPoint previews and independent visual review. All 631
retained specimens passed paired rendering inheritance checks. The complete
649-source / 648-active-bound build sweeps passed. The workshop count remains 29.

The intake adds an editable branching-roadmap renderer and content bindings,
with six named allocation corrections. Four defects found during the first
native review were repaired and reviewed in a second native export. No remaining
visual issues were found in the final 18 specimens. See the
[v10 evidence](../planning/wm-design-contracts/v10/intake-20261006-649-frozen/README.md)
and [local.19 qualification](../release/qualification-local19.json).

The next intake covers the designer's 215 slide-level density migrations and
subsequent pillar visual revisions. These changes require new implementation
and visual qualification; they are not claimed as part of the v10 release.

## Earlier implementation status — 2026-10-05

Upstream now contains **616 templates**, including **14 workshop additions** at
`c788cefeb5bb409118ac217adb53216d8156eec3`. They are implemented in Go as the
pinned v7 production library. The repository default and published gallery are
**v7 / 616**. The global CLI is now **0.1.0-local.16 / v7**. Installation used
`--cli-only`, preserving the presentation skill link. The release commit excludes
the other agent's presentation skill and DentalXChange slide edits.

- [Workshop candidate and final deck](../planning/wm-design-contracts/v7/intake-20261005-616-frozen/README.md):
  **14/14 native visual acceptance**, exact stock source/bound content round trips,
  inheritance of all 602 earlier definitions, and six named geometry amendments.
  Five initial Go allocation failures and three native defects were repaired.
  Final review covers schedule alignment, agenda clearance, readout tables,
  source notes, grouped series rails and status labels.
- [Heat-map candidate and final deck](../planning/wm-design-contracts/v6/intake-20261004-602-frozen/README.md):
  **16/16 native visual acceptance** after repairing A1–A9 badge wrapping, LATER
  chip wrapping and single-word heat column headers. The header repair uses
  reduced inset padding and a bounded 9→8 pt reduction only when required by
  the existing narrow score column. Legacy v5 geometry is unchanged.
- Native acceptance used local PowerPoint PDF export and the existing Swift
  rasterizer. An initial installed CLI export succeeded after reboot for all 16
  heat-map specimens. Fresh automated final exports subsequently failed
  operational Apple-event dispatch `-10827` under the restricted caller.
  Final GUI acceptance does not establish a fresh doctor/automated render pass.
- The doctor probe now has explicit geometry, background, Arial theme and text
  sizing instead of relying on nil/default deck options. XML/package and focused
  race checks pass; the revised probe still needs a successful live CLI check.
- The production gallery contains all **616** specimens: **586 inherited** from
  unchanged, hash-verified authored compositions and **30 accepted native**
  specimens from the heat-map/workshop intake. Paired renderer comparisons pass
  for all 586 inherited specimens. The Go publisher verifies native deck XML and
  visible dependencies for the other 30. The single production SQLite index
  includes 616 templates; photography coverage now includes all 521 originals in the requested branding
  folder (up from 20), with 1,204 asset variants overall. Historical source pins remain
  under planning, without another production gallery or SQLite index.
- Final repository tests passed in **120.61s**. Full 616-source/615-active-bound
  Go builds passed in **89.106s**, skipped by the everyday short test lane.
  Focused race checks passed for native doctor, reference/priority widths and
  shared source resolution. The exhaustive project race lane remains out of
  band; no new full race pass is claimed.
- The capacity-comment and shared staging fixes are retained in installed local.16.
  The installed package now includes all 616 previews, 1,847 artifact links and
  all 521 registered photos. Global version and photo discovery were checked
  from outside the repository; all packaged file hashes were verified.
- Semantic metadata remains inferred. Generic item matching cannot express
  complete matrices and linked detail locators. Stock specimen acceptance does
  not guarantee that replacement copy fits; edits still require fit review.

The remaining sections preserve the initial heat-map intake analysis and estimates.

## Initial comparison

Read-only comparison on 2026-10-04. Confirmed source: `/Users/rscott/Projects/wm-design-system`; catalog count 602. This observation preceded the candidate implementation above.

- Current Go pin: `d83bd58a9f9de68ebd8d6b3c9b0272c16ed516cf`.
- Observed upstream HEAD: `c355c881d543dceccb82dbe12d7129bca9b1fcac`.
- Only `templates/library/heatmaps.json` changes among library source files. Heat-map inventory grows from 22 to 37: **15 additions, one existing revision, zero removals**. The latest checkout includes changes after the originally described ten-template round.
- Existing `capability-heat/dense` changes from revision 1 to 2: adds three `rowGroups` with `groupW: 24`, reduces the row-header column width from 198 to 168 pt, and updates purpose/budget. Every other existing heat-map template is structurally identical JSON.

## Added keys

| Key | Source topology |
| --- | --- |
| `capability-heat/grouped` | 12 rows, four row groups, six site scores and average |
| `capability-heat/grouped-split` | 14 rows, four row groups, four site scores; insight bullets and callout |
| `heat-notes/full` | Four rows with five scores, variable row heights and evidence bullets |
| `heat-notes/full-continued` | Six rows, two bullets per row, continuation label |
| `heat-notes/grouped` | Six rows in three groups, five scores and evidence bullets |
| `heat-notes/now-next-later` | Grouped six-row notes table with horizon priority chips and key |
| `heat-notes/summary` | Five-row notes table with P1–P4 priority chips and key |
| `heat-notes/split` | Seven-row notes table beside an insight column |
| `heat-notes/split-last` | Four-row final continuation table beside an insight column |
| `heat-ref/overview` | Nine rows, reference badges in row labels, five dimension scores and overall score |
| `heat-ref/overview-grouped` | Reference overview with three row groups |
| `heat-ref/detail-corner` | Primitive locator heat grid plus detail for A4–A5 |
| `heat-ref/detail-continued` | Primitive locator, A6 detail and cross-row reading |
| `heat-ref/locator-split` | Locator in tall column and A1–A3 detail; latest commit adjusts inset padding |
| `heat-ref/cards-split` | A7–A9 detail panels beside a locator |

## Go gaps and intake order

1. **Table row groups and width.** `sceneTableSource` supports column `groups`, not upstream `rowGroups`/`groupW`. The new contract is `rowGroups: [{label, from, to, fill?}]` with inclusive zero-based body-row indices. `groupW` defaults to 24; columns move right by `groupW + 6`, reducing usable width. Add separate measured group labels, custom fills/contrast and validation of integral ranges, bounds, overlap and effective table width. Do not reinterpret existing column groups or shift the frozen dense template.
2. **Per-row heights and heat minimum.** Table rows currently reject `h`; table columns lack upstream `min` and per-column `size`. The source authors `size`, but the upstream renderer ignores it and renders bullet cells as `small`; accepting the field is schema compatibility, not a typography change. The frozen heat ramp normalizes from zero. Upstream notes/reference tables and locator blocks/legends use a 1–5 domain (`min` / `heatMin`), requiring consistent cell, block and legend normalization under a new source revision. Keep the current 0–4 ramp unchanged for the pin.
3. **Priority chips.** Column type `priority` is absent from the supported table type set. Cells accept a string or `{value, label?}`. Implement case-insensitive P1–P4 and now/next/later styles, native editable labels and measured chip padding. Unknown upstream values receive an outline fallback; rejecting them in Go would be a deliberate policy difference rather than source parity. Preserve the author's key line.
4. **Reference badges and table labels.** Plain cells accept `{text, sub?, ref?, refActive?}` upstream; Go currently accepts only text/sub. Active badges use Grounded fill and White text. Add measured reference badges, preserve A1–A9 identity and active state, and budget row-label width separately from numeric cells. Ordinary native bullet cells and continuation labels already exist; qualify their new row heights and dense layouts rather than replacing them.
5. **Editable projection and semantic authoring.** `libraryTableCellWalk` currently omits priority/reference fields. Explicitly project priority value/label, reference ID/activity, group labels and evidence bullets; otherwise rendered specimen values would remain frozen. Keep row-group `from`/`to` as structural geometry rather than business values. Extend aliases for the new fields. A generic `PageSpec` item cannot represent these complete matrices, legends and linked detail locators; keep matcher gaps explicit until a dedicated table/heat-map spec exists. New metadata remains inferred until reviewed.

## Required qualification before migration

- Pin the new source as a separate revision and inventory all 15 additions plus the changed dense template; retain the old revision and its original values/contract.
- Tests for row-group ranges and overlap, width sums, variable heights, 1–5 endpoint/midpoint colors shared by tables/blocks/legends, chip vocabularies, badge placement, and continuation labels.
- Full editable stock round trips: exact numeric ratings, every evidence bullet, reference identity, priority value, grouping label and source line; no synthetic business-copy fallback.
- Go layout checks and native PowerPoint export for all added templates, with focused visual review of grouped split layouts, four-bullet rows, reference badges, locator outlines, band padding and priority-chip keys.
- Rebuild discovery/authoring indexes for the new revision and report unsupported capacities honestly. Original frozen qualification must still pass without changed canonical values.

## Follow-up sequence and estimate

Independent Sol 6.1 review confirmed the intake and identified the editable
projection and source-parity details above. Estimated effort assumes a working
native export environment:

| Stage | Scope | Engineer days |
| --- | --- | ---: |
| 1 | Explicit contracts, fixtures, revision gates | 0.5–1 |
| 2 | Heat minimum across all renderers and per-row heights | 1.5–2.5 |
| 3 | Row groups, width offsets, fills and reference badges | 2–3 |
| 4 | Priority chips, editable projection and authoring aliases | 1–2 |
| 5 | New bundle revision, indexes and CLI wiring | 1–2 |
| 6 | All 16 changed-template round trips, native review and v5 regressions | 2–3 |

Total: approximately **8–14 engineer days**, excluding native environment
troubleshooting. Stages 1–4 can remain behind revision gates without changing the
frozen v5 default. The new revision must explicitly retain reviewed v5 amendments
for unchanged templates. A dedicated semantic matrix/detail matching spec is a
separate estimated **3–6 day** effort.
