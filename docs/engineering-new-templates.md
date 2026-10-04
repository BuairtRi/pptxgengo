# Upstream heat-map intake

Read-only comparison on 2026-10-04. Confirmed source: `/Users/rscott/Projects/wm-design-system`; current catalog count is 602. No bundle, source pin, index or frozen qualification was migrated.

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
