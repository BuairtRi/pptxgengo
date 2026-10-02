# Complete the WMDS template library

Date: 2026-10-02. This document preserves the initial plan; current implementation
status is in [library execution](library-execution.md) and the coverage ledger.

## Scope and starting point

Implement all 97 canonical variants in the frozen WMDS library, including both
the 51 core and 46 working designs. Design tier is independent of implementation
and qualification status. The eight live family JSON files in
`~/Projects/wm-design-system/templates/library` matched the frozen files during
this planning pass.

Four variants currently have executable content bindings: `cards/3`, `cards/4`,
`stats/four-metrics`, and `takeaway-rail/metrics-rail`. Their eight reference
specimens have native and visual review evidence. That evidence does not qualify
arbitrary replacement content. There are 93 variants still to implement and bind.

| Family | Total | Bound | Remaining |
|---|---:|---:|---:|
| approach | 11 | 0 | 11 |
| argument | 18 | 1 | 17 |
| commercials | 4 | 0 | 4 |
| evidence | 10 | 1 | 9 |
| openers | 18 | 2 | 16 |
| proof | 13 | 0 | 13 |
| solution | 12 | 0 | 12 |
| team | 11 | 0 | 11 |
| **Total** | **97** | **4** | **93** |

The existing Go typography, tokens, frames, basic cards, card rows, standalone
metrics, rich direct text, rules, provenance and native group machinery are shared
foundations. Generation and content measurement remain Go-only. Native PowerPoint
review is development evidence; users do not capture every generated deck.

## First parallel pass: establish the complete backlog

Three agents inspect disjoint families and produce feature-level coverage records:

| Owner | Families | Variants | Owned artifacts |
|---|---|---:|---|
| Primitives and content agent | openers, argument | 36 | coverage-primitives.json / .md |
| Data and cards agent | evidence, proof, commercials | 27 | coverage-data.json / .md |
| Diagrams and structure agent | approach, solution, team | 34 | coverage-diagrams.json / .md |
| Integrator | all families | 97 | merged ledger, compiler/binding interfaces, integration plan |

These records inspect nested fields and presets, not just top-level node names.
Each variant needs an exact source hash, required capabilities, content bindings,
source resolutions, and an honest implementation state. A source node named
`card` or an existing generic compose recipe does not establish WMDS support.

The merged [97-template ledger](template-coverage.json) records all eight families
and verifies every key and whole-family file hash against the frozen source.
Reproduce it with `python3 scripts/build-wmds-template-coverage.py` from the repo.
Detailed work packages are in [primitives/content](coverage-primitives.md),
[tables/charts/cards](coverage-data.md) and [diagrams/structure](coverage-diagrams.md).

## Implementation ownership

Keep three workers active beside the integrator. Assign reusable modules first;
assign template adapter batches when their dependencies are available.

| Work package | Owner | Contents | Depends on |
|---|---|---|---|
| A. Compiler and binding integration | Integrator | Full catalog; strict source normalization; identity overlay; capability dispatch; typed content adapters; source provenance and coverage reporting | Coverage records |
| B. Text, primitives and media | Worker 1 | Rich card/bullet text; headings/lists/quotes; source emphasis and footnotes; square/block/frame/number anatomy; registered images, crops, grayscale, logo/icon/art/thumbnail assets | Frozen source semantics; shared IR interface |
| C. Cards and native tables | Worker 2 | Case/bio/quote/options/checklist/column/media card presets; narrow contextual cards; typed cells, row groups, presets and fee summaries; native editable table serialization | Shared IR interface; existing measurement/group APIs; coordination with B for rich text/media |
| D. Diagrams and sequences | Worker 3 | Coordinate paths and explicit attached ports; containers, chevrons, cylinders, layers and matrices; before/after, phases/steps/time axes/Gantt/swimlanes; people/pods/org/governance | Shared IR interface; existing measurement/group APIs |
| E. Native charts and data | Worker 2 after table foundation | Authored column/stacked/line/quadrant semantics; data, axes, labels and number formats; native chart/workbook or explicitly contracted native diagram representation | Native writer audit; C foundation; source chart semantics |
| F. Template adapter batches | Integrator and available workers | All remaining 93 exact variants; required named slots; stable keys; explicit bounds/count policies and synthetic fixtures | Relevant B–E capabilities |
| G. Review and packaging | Integrator, with artifact inspection delegated | Native exports and page review; changed-content fit envelopes; exact evidence receipts; packaged CLI/catalog/docs and release readiness | Each completed adapter batch |

The longest track is D; split its remaining independent sequence and team modules
between freed workers after primitives/tables are integrated. The final ownership
list follows the feature ledger rather than distributing equal numbers of slides.

## Shared interface to freeze before renderer edits

Workers own new module files. The integrator initially owns shared `Node`, report,
render dispatch, source compiler, CLI and grouping changes. Agents propose changes
to those shared files for integration rather than editing them concurrently.

Each renderer module supplies:

1. A closed typed spec and a strict source decoder listing consumed fields.
2. A plan operation that resolves WMDS styles, measures content, computes required
   bounds, and reports overflow or source conflicts before emitting objects.
3. A native draw operation and a manifest of owned parts, semantic keys, geometry,
   assets, data and any connector attachments.
4. Explicit supported features and contextual variants; unsupported required
   fields return named errors. No fallback to screenshots or unmeasured text.
5. Synthetic reference inputs covering authored modes and distinct content cases.
   Review evidence is recorded separately from implementation.

The source compiler preserves hierarchy and local coordinate contexts. Source
pointers seed stable identities only for the pinned hash. Template bindings expose
named content fields, with schemas and keyed repetition. They cannot expose
arbitrary JSON-pointer mutation or change template geometry/base typography.
Existing four binding APIs and their artifacts remain compatible.

Shared color resolution must also cover frozen categorical series, deemphasis,
status and organization-role palettes. The current surface-role/KPI subset does
not resolve every authored color reference. Workers consume one token resolver;
they must not introduce separate approximations of the brand palette.

The current typography and grouping postprocessors primarily handle native shape
text. Before integrating tables, chart frames or pictures, extend those passes
explicitly for native table-cell text and owned picture/graphic-frame parts. Retain
the existing shape/run behavior and enforce part uniqueness and ordering.

Media inputs resolve registered IDs to local pinned bytes and explicit crop/focal
metadata. Charts and tables retain editable native data. Coordinate-only source
connections remain coordinate connections; semantic attachment is explicit.

## Integration order and checkpoints

1. Merge all 97 coverage records and reconcile source ambiguities. Freeze module
   interfaces and assign exact owned files and capability IDs.
2. Run B, C and D concurrently while A expands source loading and catalog/status
   reporting. Integrate small complete modules as they arrive.
3. Bind and build template batches as their feature dependencies close. Start with
   text-led openers and then media/quote layouts; queue table-driven proof and
   commercials, phase-detail, and structure families against their module readiness.
4. Start E when Worker 2 completes the table foundation. Reassign free workers to
   independent remaining D modules or adapter families with no shared file edits.
5. Review each generated batch in PowerPoint and inspect every rendered page.
   Keep source, bound input, resolved plan, package inspection, native output and
   review receipt together. Address shared feature defects once, then regenerate
   all dependent variants.
6. Close every ledger row, publish complete catalog status and package the
   implemented library. Release installation is a separate final operation.

Compilation checks and artifact inspection do not confer native qualification.
Focused native character capture is needed only when a new typography behavior
requires evidence; complete calibration need not be repeated per template.
No tests are added or run during the current planning pass.

## Resolve current source gaps explicitly

The source `_gaps.json` records 18 historical renderer limitations and authored
workarounds, some with pre-consolidation keys. Reconcile each to the current
canonical variant. Implement the canonical workaround as authored unless a
versioned adapter resolution explicitly changes it. Desired design improvements
to the historical original require a source migration; they are not inferred
from the gap memo.

Separate source conflicts, missing renderer capabilities, missing content bindings
and missing review evidence in the ledger. Do not mark a variant implemented when
required source semantics remain unresolved. Do not omit working-tier templates
or silently drop marks, sources, images, rows or data to close a row.

The parallel audit already identifies shared contract work: derive explicit
headerless frames for source covers/dividers/left-rail layouts; preserve declared
media bleed; represent vertical callout bars separately from horizontal rules;
contract five-up/rail/title-only and compound card anatomy; measure Gantt packing
with real fonts; and register component-specific typography. Local panel surfaces
must be represented as nested contexts instead of requiring every child to use its
frame surface. Source arithmetic or browser clipping is not evidence of native fit.

## Definition of complete

For all 97 variants:

- Exact source identity and all required fields are accounted for, with named
  resolutions for material source differences.
- Generation emits the intended editable native objects, groups, assets and data.
- A closed content API replaces all intended example content, preserving keys and
  fixed template geometry. Unknown fields, missing content and overflow fail clearly.
- A source reference specimen and defined changed-content cases are generated,
  inspected and visually reviewed. The recorded reuse envelope states supported
  counts, density, text features and known native measurement limits.
- CLI discovery and generation expose truthful implementation/review status;
  packaged resources, binding documentation and evidence links are complete.

Completion of the canonical template library is bounded by these template
contracts. It does not assert universal font/layout parity or unrestricted editing
and regeneration of manually modified decks.

## Current deliverable

The parallel implementation now has source scene adapters and content bindings
for all 97 canonical variants. Corrected source `review04` and public-bound
`bound-review04` each generated 97/97 and opened in PowerPoint without repair.
The source reference has a 97-page native PDF export and page-review receipts.
The user corrections retain their original pass02 numbering in
`samples/wmds-library-20261002/user-review-pass02.json`, with explicit links to
corrected full-reference pages. Native review covers these exact fixtures;
longer replacement copy and reusable content envelopes remain unqualified.
The installed release has not been replaced.
