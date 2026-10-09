# Assessment composition

Use this for ordinal, discrete heat assessments: capability, readiness or gap
severity across sites, teams or other comparable dimensions. The matrix compares
scores; it does not calculate an average, prioritize investments or establish a
source-backed conclusion by itself.

This is newer than the promoted v4.2.1 CLI. Check installed `project assessment`
help and the project's toolchain pins. Upgrade deliberately before using it.

## Choose the meaning before the geometry

- Decide what rows assess and what columns compare. Give each a stable key;
  changing display order must not reassign scores.
- Define integer scores **0..max**, where `max` is 1..4, and provide one label
  for every score. The five-step heat palette supports up to five distinct scores.
- Preserve the original scale's meaning: `seq` shows ordered intensity; `risk`
  uses the catalog's risk palette. A capability-gap template's zero can mean
  "No gap", not "Absent capability". Do not reverse that interpretation.
- **Zero is a measured score.** Null or an absent cell means not assessed. The
  generated legend explicitly explains blank cells. Never substitute zero for
  missing observations or compare scores based on different rubrics as if equal.
- Keep supporting evidence, score definitions, dates and caveats in the slide's
  source/notes or project source records. These commands do not invent them.

## Supported source

`wmds/component/assessment` owns keyed `columns`, keyed `rows`, each row's
`scores` map keyed by column, the domain/legend and the matrix layout. It uses
native PowerPoint table cells and the existing table/heat/legend visual language.

It is distinct from arbitrary tables, continuous percentages, survey proportions,
weighted scorecards, grouped assessments, RACI or commercial calculations.
Legacy inline group/total/row-scale/ink/height overrides and other table
adornments also fall outside this uniform ordinal matrix scope. Those
need their own domain contracts. Native cell edits currently require review and
explicit authored score updates; they do not automatically update this model.

Use a detached, slide-owned local template. Shared or revision-pinned local
references and adopted native geometry/order block structural patches; fork or
explicitly revise/reset those states after preserving the previous version.

## Inspect, preview and apply

```sh
pptxgengo design project assessment inspect \
  --project PROJECT --slide capability --node heat
pptxgengo design project assessment patch \
  --project PROJECT --slide capability --patch assessment-patch.yaml
# Inspect the measured candidate, then apply that exact guarded patch:
pptxgengo design project assessment patch \
  --project PROJECT --slide capability --patch assessment-patch.yaml --apply
```

Every patch requires the current source SHA-256, actor, reason and node ID.
Preview does not write the project. Apply rechecks source inputs, retains their
predecessors and records a decision. Then build, inspect the layout report and
render through PowerPoint; a successful measured patch is not desktop QA.

```yaml
schema: pptxgengo.assessment-patch.v1
expected_source_sha256: REPLACE_WITH_CURRENT_PROJECT_SOURCE_SHA256
actor: presentation operator
reason: Add a site and retain known and unassessed capability scores
node_id: heat
operations:
  - action: set
    entity: column
    key: west
    label: West
  - action: set
    entity: score
    key: access
    column: west
    score: 0
  - action: set
    entity: score
    key: scheduling
    column: west
    missing: true
  - action: reorder
    entity: column
    order: [west, campus, lakeside, ridgeview, harbor]
```

The example assumes those row/column keys already exist. Use inspection's keys,
not their labels or current index. Column keys `assessment-label`, `group`,
`total`, `ink`, `scale` and `h` are reserved; choose distinct semantic keys. Unknown references and fractional/out-of-domain
scores are rejected. Numeric `score` and `missing: true` are mutually exclusive.

## Initialize an existing catalog derivative explicitly

Inspect the legacy table before initialization using the geometry route:

```sh
pptxgengo design project diagram inspect \
  --project PROJECT --slide capability > diagram.json
```

Use its `source_sha256` for the initialization envelope and the authored table
and companion legend node IDs (not child PowerPoint object names). Assessment
inspection requires an assessment component and intentionally refuses the
uninitialized legacy table. After applying initialization, use `assessment inspect`.

A catalog heat table and its independent legend have no assessment domain model.
Initialization is an explicit **complete model replacement**, not automatic
score extraction. Map every retained source score to its row/column key and
confirm its rubric. Keep the existing slide title, notes and source facts.

The target must be a leaf local `wmds/component/table`: light header, styled row
header, ungrouped, zero-based heat columns. Name the exact companion local legend
in the same frame zone. Initialization removes that legend only and combines its
vertical allocation with the table's existing allocation. It preserves source
cell density, row height, label width, score display and heat palette/domain.

```yaml
operations:
  - action: initialize
    entity: source
    legend_node: key
    cascade: true # acknowledge complete replacement of existing cell source
    model:
      row_label: Capability
      label_width_pt: 144
      row_height_pt: 36
      dense: false
      show_scores: false
      domain:
        max: 4
        scale: risk
        labels: [No gap, Small, Moderate, Large, Critical]
        missing_label: Not assessed
      columns:
        - {key: campus, label: Main campus}
        - {key: harbor, label: Harbor}
      rows:
        - key: access
          label: Patient access
          scores: {campus: 1, harbor: 3}
        - key: scheduling
          label: Scheduling
          scores: {campus: 2, harbor: null}
```

Wrap this operations list in the required patch envelope above. These style
values match `capability-heat/callouts`; other specimens may differ. Reducing a
stock example to these rows/sites is a deliberate adaptation, not evidence that
the omitted capabilities do not exist. Record it in the composition log.

If an assessment node has bindings, an explicit first `materialize source`
operation freezes its current resolved model before subsequent operations. Other
components' bindings remain intact. Prefer editing authored bindings when that
is the intended ownership.

## Operations and fit choices

| Operation | Effect |
| --- | --- |
| `set row/column`, with `key` and `label` | Add an axis with missing scores or rename its label; existing scores survive |
| `reorder row/column`, with complete `order` | Change display order; preserve keyed scores |
| `remove row/column`, with `key` | Delete that axis and incident score entries; populated axes require `cascade: true`, including numeric zero |
| `set score`, with row `key`, `column`, `score` or `missing: true` | Change one explicit observation |
| `set domain`, with complete `domain` | Replace rubric and legend together; every retained score must still fit |
| `set layout` | Set `label_width_pt`, `row_height_pt` and/or `show_scores` explicitly |

Score columns share the width left after the label column. Adding a column reduces
each column's width; adding a row increases table height. The legend can wrap as
its definitions grow. Nothing silently shrinks text or discards scores.

For long labels, preview a greater row height or label width. If that causes
narrower comparison columns or exceeds the frame allocation, widen/reallocate the
local component, choose another catalog frame/layout, or split the assessment.
Retain failed proposals so the fit decision remains visible. Do not enlarge the
body into title/footer zones or remove the missing-value legend to force a fit.
