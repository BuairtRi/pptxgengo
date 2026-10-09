# Assessment composition scope

The first slice introduces an explicit ordinal assessment domain instead of
presenting arbitrary editable tables as domain-aware assessments. It uses the
current project's pinned source bundle and the existing native table/heat/legend
planners; no modified catalog bundle or invented provenance is required.

## Implemented contract

- `project assessment inspect/patch` targets `wmds/component/assessment`.
- Stable row/column keys keep score maps attached through rename/reorder.
- Discrete integer 0..max, max 1..4, with complete labels and seq/risk palette.
- Absent/null scores are unassessed, distinct from numeric zero; blank cells have
  an explicit neutral legend entry. Numeric zero participates in deletion guards.
- Explicit complete initialization from a selected ungrouped light/row-header
  local heat table; the named companion legend is replaced, not guessed.
  Initialization requires `cascade: true` and refuses inline group/total/row-scale/ink/height metadata or unsupported table adornments. It preserves density, cell height,
  label width, palette/domain and score display treatment. It does not extract
  source scores automatically or qualify arbitrary tables as assessments.
- Add/remove/reorder axes, change scores/rubric/layout, explicit binding
  materialization, source-hash guards, measured preview and atomic apply use the
  shared composition transaction. No weighting, totals or silent text shrink.

See the [operator runbook](../skills/west-monroe-presentations/references/assessment-composition.md)
for actual command schemas and fit decisions. Generic geometry/native cell
changes do not imply score changes. Native score reconciliation remains pending.

## Catalog exceptions

`capability-heat/callouts` is an ungrouped risk-gap specimen, not a maturity score.
Its catalog numeric zero means "No gap"; translating it into "Absent" reverses
its meaning. Dense/grouped variants need a grouping model before domain adoption.
Risk likelihood-impact grids, continuous capacity/proportion/survey heat maps and
coupled annotations are separate interpretation contracts. This implementation
leaves installed catalog sources immutable.

## Adjacent work with independent scope

| Workstream | Concrete model/operation needs |
| --- | --- |
| Weighted assessments | Dimension weights, explicit missing policy, comparable rubric, score denominator and traceable aggregation; no default averaging |
| Quantitative charts | Keyed categories/series, units and measurement basis, source records, missing-vs-zero points, chart/table synchronization, explicit axis/scale changes; existing chart primitives do not establish those coupled domain operations |
| Commercial models | Rates, quantity units, period/currency conventions, phase identity, formulas, totals, rounding and authoritative assumptions; editable fee/rate tables alone do not establish validated calculation semantics |
| Grouped heat assessments | Keyed groups and membership, contiguous-vs-free ordering policy, legend synchronization and caption/annotation dependencies |
| Assessment native adoption | Receipt-backed cell identity and explicit score proposals; no inference from cell fill or geometry alone |

These workstreams can proceed independently of architecture/process composition.

## Qualification

Focused source tests cover initialization, key-preserving reorder/additions,
zero/null, deletion acknowledgement, out-of-domain/fractional scores, rubric
reduction, long-label fit/refusal, preview immutability, stale source, shared/
pinned/native-layout boundaries and the companion-legend-before-table order.
Renderer tests check numeric-zero text and heat identity, absent-cell blank
identity and missing legend separately. Source rebuild checks are not Office
Save As or Windows qualification. Live desktop evidence is recorded separately
by the integrating operator with exact source/binary/bundle/build hashes.
