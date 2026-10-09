# Edit native content components and collections

Use this workflow for editable source copy, lists, cards, phase sequences,
legends, sets and content collections whose meaning is authored directly. Use
the specialized commands for schedules, people/reporting, quantitative charts,
commercial calculations, processes, journeys, portfolios, cycles, maturity and
staffing. Generic operations cannot bypass those semantic contracts.

```sh
pptxgengo design project component inspect --project PROJECT --slide SLIDE --node NODE
pptxgengo design project component patch --project PROJECT --slide SLIDE --patch change.yaml
pptxgengo design project component patch --project PROJECT --slide SLIDE --patch change.yaml --apply
```

Inspect a detached or deliberately scaffolded local component. Inspection returns
resolved arguments, collections and stable keys, source hash, semantic
relationships/markers where supported, measured frame geometry and render errors.
It also identifies the specialized command or explicit materialization needed.

## Stable selectors and operations

`path` is an argument JSON pointer. At arrays use `@KEY`, never a numeric index.
For example `/body/@observations/bullets` addresses one named card body block,
independent of its position. `set/argument` changes an existing scalar field;
`set/item` inserts or replaces a complete collection member; `reorder/item`
lists every key exactly once. Use `set/layout` with `rect` for outer placement.
Source `type`, `id`, `x/y/w/h/_h` and `_source_geometry` are reserved; layout preserves frame allocations and captured source anchors.

```yaml
schema: pptxgengo.component-patch.v1
expected_source_sha256: COPY_FROM_INSPECT
actor: Operator
reason: Add an observation and reorder supporting blocks
semantic_review: Observations are independent; related claims and legends were reviewed
node_id: message-card
operations:
  - action: set
    entity: item
    path: /body/@observations/bullets
    key: owner
    value: Name the accountable owner
  - action: reorder
    entity: item
    path: /body
    order: [observations, support]
```

Keys in the source overlay survive insert, remove and reorder, including nested
arrays when their parent moves. New object members with their own authored `key`
must agree with the operation key. Preserve child identities using nested keyed
operations; a complete object replacement with a different nested count requires
those explicit child operations rather than guessing identity correspondence.

`remove/item` requires `cascade: true`: it explicitly acknowledges deleting the
member's nested content and reviewing any external references. It does not
silently repair other slides, totals, criteria tables or narrative. Every patch
requires `semantic_review` explaining how those dependencies were checked.
Bound arguments need first `materialize/source`, freezing resolved local values,
or an update to the original bound source. Preview performs full renderer fit;
apply preserves predecessors and rejects stale hashes/pins/native overrides.

## Semantic adapters

### Phase sequences

`phases` uses an older positional current index internally. Reordering/removing
phases keeps current phase identity; deleting it with cascade clears the marker.
Use `set/marker`, `path: /current`, `key: PHASE_KEY`, or `remove/marker` to set/clear
it. Direct numeric `set/argument /current` is refused. Steps/activities/
deliverables are still authored work, not calculated dates or a critical path.

### Venn sets and regions

Inspect `relationships` for each region's named set members. Reordering sets
rewrites internal numeric membership while preserving set identity. Deleting a
set with cascade removes incident regions, retaining unrelated regions. Author a
region with `set/item`, `path: /regions`, a stable region key and a value such as
`{members: [access, quality], label: Shared opportunity}`. Positional `in` edits
are refused. Set count, measured labels and source geometry still determine fit.
Do not use intersections as quantitative areas unless a source explicitly defines
that meaning; the template is a conceptual set diagram.

### Colored matrices

For `matrix`, column reordering rewrites every row's corresponding cells. A new
column value is `{label: LABEL, cells: {ROW_KEY: VALUE, ...}}`, with an explicit
cell for every stable row. This prevents assigning a cell to the wrong heading.
Row metadata and cell styling remain authored. Use native `project table` for
actual tabular data, and assessment for a declared ordinal scoring domain.

## Practical composition

Use concise headings, native paragraph lists and enough space for wrapped copy.
Change counts to support the message; templates are examples, not fixed quotas.
After a count or ordering change, inspect related captions, numbering, links,
legends, criteria and evidence references. Split at a meaningful boundary when
fit fails. Native movement changes layout only; review known text/geometry using
the reconciliation packet and explicitly persist semantic facts in source.

Card order can express a workflow, chronology or dependency. A numbered layout
does not make its items freely interchangeable. For example, change control reads
Request → Assess → Approve → Update; reversing that order contradicts the process.
The generic reorder command applies your explicit intent and cannot infer the
business sequence from captions. Preserve the sequence when only customizing
copy, and review the whole-slide claim, footer and numbered labels before apply.
Stock example headlines also need checking against the authored values: the
`maturity/ai-assessment` example has six two-stage gaps and one one-stage gap,
despite its stock headline claiming seven two-stage gaps. Correct the local
headline or supply a new explicit scenario; retain the pinned parent's history.
