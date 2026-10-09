# Quantitative charts

Use charts to compare measured facts. State the unit, denominator, period, source, and whether figures are observed, estimated, or illustrative before changing the example. Templates show a visual treatment; category/series counts are editable. A media exhibit, headline metric, or ordinal assessment listed in the audit's quantitative profile is not automatically a numerical chart.

## Start from an actual template

Inspect installed help and the project's toolchain lock. On versions exposing these commands:

```sh
pptxgengo design project diagram inspect --project PROJECT --slide SLIDE
pptxgengo design project quantitative inspect --project PROJECT --slide SLIDE --node CHART
pptxgengo design project quantitative patch --project PROJECT --slide SLIDE --patch patch.yaml
pptxgengo design project quantitative patch --project PROJECT --slide SLIDE --patch patch.yaml --apply
```

A shared template must first become a deliberately scaffolded/forked local template; use the [architecture composition runbook](architecture-geometry.md) for local-template/geometry setup. Select an actual `wmds/component/chart`, retain its frame and allocation, and inspect its source and keys. This command handles native column, bar, line, pie, doughnut, scatter and quadrant source treatments. It preserves unrelated source nodes, chrome, template provenance, and bindings outside the selected component. Bound chart arguments require an explicit first `materialize/source` operation; editing the original binding values is the alternative.

## Edit with stable identities

Patches use `pptxgengo.quantitative-patch.v1`, `expected_source_sha256`, `actor`, `reason`, `node_id`, and `operations`. Copy the current hash and keys from inspection. Operations:

| Action/entity | Fields and effect |
|---|---|
| `replace/source` | Complete literal `data`, exact top-level collection `keys`, and `cascade: true`; reviewed chart-kind/source-union migration retaining this node’s placement and all other nodes |
| `materialize/source` | First operation only; resolves selected bindings into reviewed literal arguments |
| `set/category` | `key`, `label`; upserts a category; a new category creates missing observations to fill explicitly |
| `set/series` | `key`, `data`; complete renderer source series (`name`, aligned `values`, optional `color`/`dashed`) |
| `set/value` | Category `key`, `series` key, and numeric `value` **or** `missing: true` |
| `set/point` | Scatter `key`, `point: [x,y,label]` (label optional); both coordinates observed |
| `set/item` | Quadrant `key`, `data: {x,y,label,to}` (`to` optional); explicit original chart coordinate semantics |
| `remove/category`, `remove/series`, `remove/point`, `remove/item` | Existing `key`, `cascade: true`; explicit acknowledgement of lost observations |
| `reorder/category`, `reorder/series`, `reorder/point`, `reorder/item` | `order` containing every key exactly once |
| `set/setting` | `field`, `setting`; change a supported source setting; null removes that optional field |

Example (replace the hash and existing keys):

```yaml
schema: pptxgengo.quantitative-patch.v1
expected_source_sha256: CURRENT_SOURCE_HASH
actor: Operator
reason: Compare supplied capacity scenarios in FTE
node_id: chart
operations:
  - {action: materialize, entity: source}
  - {action: set, entity: category, key: pilot, label: Pilot}
  - {action: set, entity: value, key: pilot, series: capacity, value: 4.5}
  - {action: set, entity: setting, field: units, setting: FTE}
  - {action: set, entity: setting, field: source, setting: 'Approved capacity plan, September 2026'}
  - {action: set, entity: setting, field: preserveWorkbookZeros, setting: true}
```

Category reorder moves every series' observations, highlights, and pie/doughnut colors with their original keys. Series reorder carries colors with the series. Removing an item does not delete other slides, update deck prose, or change a commercial calculation. Review those linked claims explicitly.

## Units, scales, labels and missing data

Supported settings: `title`, `units`, `source`, `kind`, `mode`, `format`, `valueSuffix`, `yMin`, `yMax`, `xMin`, `xMax`, `allowMissing`, `xTitle`, `yTitle`, `target`, `progress`, `center`, `trend`, `quadrants`, `key`, `style`, `fill`, `strongState`, `positionMode`, `holeSize`, `preserveCategories`, `preserveWorkbookZeros`, `colors`, and `highlight`. Their underlying renderer unions and limits remain authoritative; listing a setting does not make it valid for every kind. `trend` remains unsupported for native scatter trendlines.

- Values stay in their authored units. `$M` is not interchangeable with USD; use source values and number formatting consistently. Changing a unit label does not convert numbers. For exact decimal conversions/calculations use [commercial models](commercial-models.md).
- `format` describes display. `valueSuffix` and `format` are alternatives. Progress doughnuts use an explicit `progress` fraction 0..1; remove it deliberately before switching to independent series observations.
- `yMin`/`yMax` control numerical vertical/value axes. `xMin`/`xMax` are scatter-only. Explicit bounds must be finite, ordered, and contain observations/stack totals and targets; clipping is refused. Signed line/column/bar/scatter data requires a suitable explicit minimum.
- Null means no observation; **zero is observed**. Lines support missing values. Column/bar require `allowMissing: true`; missing stacked/percent observations are refused. Pie/doughnut and scatter require observed values. All-missing series are refused.
- Set `preserveWorkbookZeros: true` before a fresh baseline if native data will be reconciled. It writes observed zeros explicitly in the embedded workbook.
- `autoUpdateWorkbook` accepts an explicit boolean and writes the native automatic-update request. Omitted/false retains legacy behavior. Both explicit true and legacy false produced a coherent saved chart in the named macOS column-chart exercise after explicit Excel Save plus PowerPoint Save As. The flag is not required by that qualified save workflow; cache/workbook agreement is still mandatory.
- Lines join across absent observations according to the retained source contract; column/bar missing observations leave gaps. Explain this meaning in the slide rather than replacing nulls with zero.
- Direct line labels are measured and packed inside the chart allocation. Long labels or excessive counts can still fail fit. Resize within the frame, simplify labels, or split the slide; do not shrink into illegibility.
- Pie totals, doughnut centers, annotations, titles, headline metrics, and accompanying claims remain independently authored unless explicitly mapped to a commercial calculation. Update/review them whenever data changes.
- Quadrant `positionMode: qualitative` is a design judgement, not a measured axis. Retain its declared state/color meanings; coordinate movement does not create measured facts.

## Review native numerical edits

For supported native charts, use this macOS workflow: **Edit Data in Excel → commit the cell edit → save the owned Excel workbook explicitly (Command-S) → close that workbook → Save As a separate PowerPoint copy → reopen or export that saved copy and verify its visible values**. A visible label before PowerPoint Save As can still be stale. Package reconciliation must independently verify the cache and embedded workbook; neither the Excel save nor the pre-save display alone proves agreement. Only close the workbook opened for this chart.

Edit a **copy** of an immutable build. Make a closed geometry-enabled review packet using `project reconcile propose --geometry`, then:

```sh
pptxgengo design project quantitative reconcile --project PROJECT --slide SLIDE --node CHART --packet PACKET
pptxgengo design project quantitative reconcile --project PROJECT --slide SLIDE --node CHART --packet PACKET --decisions decisions.yaml
pptxgengo design project quantitative reconcile --project PROJECT --slide SLIDE --node CHART --packet PACKET --decisions decisions.yaml --apply
```

Decisions use `pptxgengo.quantitative-semantic-decisions.v1`, the exact returned `report_sha256`, `actor`, `reason`, and `decisions: [{proposal_id, action: set_value|keep_source, reason}]`.

Supported adoption reads the authenticated native chart and its embedded workbook; both must agree exactly. It keeps unchanged source keys, category/series labels, units, chart type and axis identities, formula ranges, referenced workbook table ranges, and the original owned shape and geometry. Office can rename internal chart/workbook relationships, flatten one-level category caches, and rewrite styles or workbook views during Save As. Those formatting changes remain manual findings; numerical adoption retains the exact edited package and source formatting. It does not claim visual equivalence. It supports series facts in column/bar/line/pie/doughnut and observed scatter coordinates. It never estimates numbers from bar height or marker movement. Source or lock changes after the baseline require a fresh build. Qualitative quadrants, progress union charts, remapped ranges, changed axes/series counts, formulas, external workbooks, ambiguous identities or cache/workbook contradictions require explicit source interpretation.

Review the full report, including manual findings, before using its exact hash in decisions. State that the decision adopts the numeric fact only and leaves formatting findings unresolved. If Excel saved the new value but the plotted cache or visible label still shows the old value, stop: this is a cache/workbook contradiction, not a permission to prefer either value. Keep that failed save as evidence; a source `quantitative patch` can explicitly implement the intended fact while GUI refresh is diagnosed. Never rewrite the edited cache merely to make reconciliation pass.

Unselected changes remain in the report and retained edited package. Accepted numeric changes rebuild the supported chart; this is partial semantic adoption, not synchronization of all formatting, annotations, geometry, or arbitrary imported charts. Verify the rebuilt deck in PowerPoint, then save a numbered source-and-deck snapshot and share the complete project according to [project structure](project-structure.md).

For a different chart kind, use `replace/source` with the complete intended source model and explicit keys for its categories, series, points or items. This is a reviewed model replacement: include source, units, labels and any explicit axes you want to retain. Identity and placement remain fixed. Prefer keyed edits within the existing source model when its meaning stays the same.
