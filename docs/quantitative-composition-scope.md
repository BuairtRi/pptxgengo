# Quantitative composition scope and evidence

## Implemented contract

`project quantitative inspect|patch|reconcile` handles actual local chart source nodes. The quantitative profile spans 36 V11 variants: 29 contain 30 chart nodes; seven are metric/content/media/assessment compositions and use their actual source-family commands. Profile membership does not turn a metric or a photo into a chart.

Keyed categories, series, scatter observations and quadrant items support add, replacement, removal and complete reorder. Numeric edits address category + series identities. Counts, nested key overlays, value/category pairing, highlights, labels and explicit colors survive reorder. Missing observations remain null, distinct from zero. Column/bar gaps require explicit `allowMissing: true`; stacked or all-missing series remain invalid. Scatter coordinates must be observed. Full source-kind migration uses reviewed `replace/source` with complete literal data, exact collection keys and explicit cascade acknowledgement. Placement, identity and other nodes remain intact.

Optional finite x/y limits control supported native charts without changing frozen defaults or clipping observations. Signed data needs suitable limits. Direct annotations use measured line heights and cannot overrun the allocated plot. Labels, source, units, independent chart-center claims and qualitative axes remain authored facts.

Native numerical adoption is available for authored column, bar, line, pie, doughnut and scatter charts. It authenticates the closed receipt-backed geometry packet, exact source/lock/build baseline, chart ownership and unchanged native geometry; verifies every cache value against its embedded workbook cell; verifies unchanged chart type/axis identities, category/series identities, formula references, referenced table ranges and nonnumeric workbook cells. Office transport renaming and one-level category-cache flattening are supported after independent reference authentication. Formatting, hidden-axis styles, workbook view/revision metadata and removed orphan table parts remain manual findings, retained in the exact edited package; source formatting is unchanged and visual equivalence is not claimed. Only explicit `set_value` decisions enter source. Cached values without workbook agreement, external references, formulas, extra/remapped workbook cells, changed counts/axis identities and incomplete scatter observations are refused. Non-scatter baselines require `preserveWorkbookZeros: true`. Qualitative quadrants and progress rings require deliberate source interpretation; they do not become numeric evidence by moving shapes. Unknown edits and the original native package are retained by the semantic transaction.

## Exact V11 applicability

| Catalog variant | Actual source types | Operator route |
|---|---|---|
| `stats/four-metrics` | metric (4), rule (1), text (3) | actual metric/content/media/assessment source; no native chart cache |
| `stats/four-metrics-right` | metric (4), rule (1), text (3) | actual metric/content/media/assessment source; no native chart cache |
| `stats/circled-headline` | card (3), metric (1) | actual metric/content/media/assessment source; no native chart cache |
| `chart/column-full` | chart (1) | quantitative chart source + surrounding component content |
| `chart/column-full-split` | bullets (1), chart (1), text (1) | quantitative chart source + surrounding component content |
| `chart/column-rail` | chart (1), text (3) | quantitative chart source + surrounding component content |
| `chart/column-left` | chart (1), text (3) | quantitative chart source + surrounding component content |
| `quadrant/subtle` | chart (1), textblock (3) | quantitative chart source + surrounding component content |
| `quadrant/subtle-left` | chart (1), textblock (3) | quantitative chart source + surrounding component content |
| `quadrant/numbered-legend` | chart (1), strongnum (1) | quantitative chart source + surrounding component content |
| `quadrant/strong` | chart (1), textblock (3) | quantitative chart source + surrounding component content |
| `image-text/square-left` | bullets (1), square (1), text (2) | actual metric/content/media/assessment source; no native chart cache |
| `image-text/frame-right` | imageframe (1), text (3) | actual metric/content/media/assessment source; no native chart cache |
| `image-text/frame-right-split` | bullets (1), imageframe (1), text (3) | actual metric/content/media/assessment source; no native chart cache |
| `value/capacity-conversion` | chart (1), strongnum (1), textarrow (1) | quantitative chart source + surrounding component content |
| `phase-detail/team-handoff` | block (3), card (1), chart (1), chevron (2), stepper (1), text (6) | quantitative chart source + surrounding component content |
| `interviews-coverage/overview` | chart (2), metric (4) | quantitative chart source + surrounding component content |
| `stakeholders/quadrant` | chart (1), strongnum (1) | quantitative chart source + surrounding component content |
| `stakeholders/quadrant-split` | chart (1), schedule (1), text (1) | quantitative chart source + surrounding component content |
| `readiness/scorecard` | dots (1), gauge (1), metric (1), num (1), status (1), table (1) | actual metric/content/media/assessment source; no native chart cache |
| `readiness/adoption-curve` | chart (1), metric (3), text (1) | quantitative chart source + surrounding component content |
| `adoption/dashboard` | chart (1), metric (4), num (1), rating (1), status (1), table (1) | quantitative chart source + surrounding component content |
| `adoption/dashboard-split` | chart (1), metric (3), num (1), rating (1), status (1), table (1) | quantitative chart source + surrounding component content |
| `capacity/ramp` | chart (1), metric (1), rule (1), strongnum (1) | quantitative chart source + surrounding component content |
| `investment/roi-payback` | bullets (1), chart (1), metric (4), text (1) | quantitative chart source + surrounding component content |
| `investment/cost-vs-value` | chart (1), rule (1), text (5) | quantitative chart source + surrounding component content |
| `value-summary/left-panel` | chart (1), metric (3), rule (1), text (6) | quantitative chart source + surrounding component content |
| `value-curve/break-even` | chart (1), connector (1), text (8) | quantitative chart source + surrounding component content |
| `value-curve/break-even-split` | bullets (1), chart (1), connector (1), rule (4), text (11) | quantitative chart source + surrounding component content |
| `value-curve/scenarios` | chart (1), connector (1), rule (3), text (10) | quantitative chart source + surrounding component content |
| `value-types/hard-soft-right` | bullets (1), chart (1), metric (2), rule (2), text (10) | quantitative chart source + surrounding component content |
| `inaction/cost-of-waiting` | bullets (1), chart (1), rule (3), text (9) | quantitative chart source + surrounding component content |
| `value-tracking/planned-vs-realized` | chart (1), metric (4), num (2), status (1), table (1) | quantitative chart source + surrounding component content |
| `case-study/exhibit` | bullets (1), chart (1), grouplabel (2), metric (2), textblock (1) | quantitative chart source + surrounding component content |
| `case-study/exhibit-split` | bullets (1), chart (1), grouplabel (2), metric (1), textblock (1) | quantitative chart source + surrounding component content |
| `case-study/exhibit-left` | bullets (1), chart (1), grouplabel (2), rule (2), text (12) | quantitative chart source + surrounding component content |

## Focused evidence and runnable fixtures

Tests are in `internal/deckproject/quantitative_composition_test.go`, `quantitative_catalog_test.go`, `chart_reconcile_test.go`, and `internal/wmdesign/scene_quantitative_axes_test.go`. They exercise full actual stock scaffolds, per-chart bound-source materialization and reversed keyed order, real catalog count growth, null/zero, scale bounds, malformed/stale/strict patches, receipt/hash guards, column/bar/line/pie/doughnut cache+workbook edits, scatter coordinate edits, missing line/column values, source-unit retention and rebuild. A separate four-quadrant regression compares the published source geometry to the compiled local scaffold exactly. These XML/package tests do not claim arbitrary PowerPoint objects or Windows Office qualification.

Run only these focused cases in the managed slot:

```sh
slotctl --repository pptxgengo --slot pptx-v430-composition exec -- go test ./internal/deckproject ./internal/wmdesign ./pptx ./cmd/pptxdesign -run 'TestQuantitative|TestChartMissingValues|TestProjectQuantitative' -count=1
```

Set `PPTXGENGO_COMPOSITION_FIXTURES` for fixture export in the catalog tests. Private fixture directories include `chart-column-rail`, `chart-column-rail-growth`, and `value-curve-scenarios`, each with `deck.yaml` and the current-hash `patch.json`. Pin that project with the same V11 bundle, preview/apply its quantitative patch against `catalog-slide`, then build. The fixture patch enables explicit workbook zeros for native adoption. Keep the baseline build immutable, edit a separate native copy, generate the geometry review packet, then run quantitative reconcile with explicit decisions. Verify the rebuilt source units and numeric facts before taking a snapshot and portable ZIP.

The integration qualification ledger records final run outcomes and desktop evidence. Operator instructions live in [the quantitative runbook](../skills/west-monroe-presentations/references/quantitative-charts.md).

## CI cadence for source inventory evidence

The exhaustive source catalog audits run without `-short` in the private GitLab
`composition-catalog` nightly/on-demand lane. Normal main and release-tag unit
checks retain representative variable-count, source-preservation, geometry and
semantic regressions, while `testing.Short()` skips the two full inventory sweeps.
A retained complete source audit remains required before cutting the v4.3.0 tag;
the protected release resource job independently regenerates the full template
browsing deck. No race detector or desktop runner is added to release tags.

## macOS Edit Data evidence

The owned PowerPoint/Excel GUI exercise on 2026-10-09 first saved workbook B2=15 while chart1 cached and displayed 14. The initial save omitted explicit Excel Command-S; that contradictory artifact remains intact and is refused. The opt-in private `chart_office_qualification_test.go` reads the exact historical ownership/cache/workbook facts and a separately labeled XML-coherent copy without mutating the GUI artifact or fabricating toolchain pins.

The controlled fresh CLI projects exercised both explicit `autoUpdateWorkbook:true` and the legacy omitted/false setting. In both cases, **Edit Data in Excel → edit B2 → explicit Excel Save → close the owned workbook → PowerPoint Save As** produced saved cache/workbook15. The false-setting chart still displayed14 immediately before PowerPoint Save As, so that intermediate display did not prove the final package was stale. Reopen or export the saved copy and independently inspect the native cache/workbook before adoption. No causal claim that true is required follows from these results; legacy defaults remain unchanged.

The actual true-setting GUI package was receipt-authenticated and proposed exactly one source-keyed14→15 fact. Explicit numeric-only preview/apply retained30 manual findings plus raw edited package, source styles and units (`Days to close`). Rebuild `build-20261009T152406-eb6141024f011c05` persisted source15 and independently verified cache/workbook15. The legacy false-setting actual GUI package independently passed the same explicit proposal/preview/apply workflow, retaining30 manual findings and units. Rebuild `build-20261009T152605-22501dcb5993573d` verified source/cache/workbook15 (`09718075d52aa960638b80b05fda353c`). This is a named macOS native-column exercise, not qualification of arbitrary Office objects or Windows PowerPoint.

`autoUpdateWorkbook` is an explicit boolean source setting propagated through CLI patches, scene planning and native `c:autoUpdate`. Omitted/false retains legacy serialization; true requests automatic updates. Focused tests prove both flag values preserve internal embedded relationships, series/category formula ranges, zero versus missing and numeric cache/workbook facts. The completed four-package source/contract test run (`51b3eb604e3fe1a0d13555159fafeb57`) retained the exhaustive91-assignment catalog proof; final native guard/flag/private historical evidence regression passed (`1beab2136832a66fca9c13c6d534815b`).
