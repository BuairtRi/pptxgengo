# Compose portfolio horizons and initiatives

A portfolio roadmap separates horizons, initiatives and dependencies. A horizon
is a planning category, not a calendar promise. Author its `meaning` explicitly:
for example, Now/Committed focus, Next/Planned focus, Later/Options, not dates.
Use Gantt for dated/period work; use journey for a road metaphor.

```sh
pptxgengo design project portfolio inspect --project PROJECT --slide SLIDE --node PORTFOLIO
pptxgengo design project portfolio patch --project PROJECT --slide SLIDE --patch change.yaml
pptxgengo design project portfolio patch --project PROJECT --slide SLIDE --patch change.yaml --apply
```

Inspect returns source hash, keyed model and measured fit. Operations use schema
`pptxgengo.portfolio-patch.v1`, `expected_source_sha256`, `actor`, `reason`,
`node_id`, and `operations`. Preview before apply; source transactions reject
stale source, shared/pinned templates and active native overrides.

## Model and operations

- `horizons`: stable `key`, `label`, explicit `meaning`.
- `statuses`: stable `key`, visible `label`, design-system `surface`. Keep the
  mapping consistent across cards; do not turn color into an unstated judgment.
- `initiatives`: stable `key`, `label`, `horizon` key, `owner`, `status` key,
  explicit `confidence` label, and nonnegative `slot` within that horizon.
- `dependencies`: stable `key`, `from`/`to` initiative keys, optional `label` and
  intermediate `route` waypoints; optional `label_position` centers the measured
  label patch. Coordinates are local to the initiative area below the 42 pt
  horizon header, matching the existing route coordinate system. Every waypoint
  and label anchor requires exactly two finite numbers.
- `card_height_pt`: measured native card allocation; count limits are not fit
  guarantees. Add height or split the view rather than shrink text.

`set` adds/replaces a complete `horizon`, `status`, `initiative`, or `dependency`.
`reorder` includes every key of the entity exactly once. Horizon reordering
changes columns while dependencies and initiative membership retain identity.
Initiative source order does not change an explicit `slot`.

Dense adjacent horizons may have room for an arrow but not its label. Preview
refuses label patches within 2 pt of cards, route segments through card bodies
or unrelated-card clearance, and labels outside the initiative allocation.
Preserve the intended horizon order and membership; do not move initiatives to
different horizons to avoid a fit failure. Use `set/dependency` with the same
key, endpoints and label to author an explicit clear detour and label position.
For a five-horizon 846×300 pt example with 72 pt cards, Access in Now and Data in
Next at slot 0 and another initiative at slot 1, the dependency can use
`route: [[159.6,9], [169.2,9], [169.2,64.5]]` and
`label_position: [169.2,9]`. This keeps its label above both cards; inspect and
remeasure those coordinates when allocation, slots or counts change. A label
position requires nonempty authored label text and changes geometry only.

```yaml
schema: pptxgengo.portfolio-patch.v1
expected_source_sha256: COPY_FROM_INSPECT
actor: Operator
reason: Transfer the data initiative after readiness review
node_id: portfolio
operations:
  - action: set
    entity: initiative
    key: data
    initiative:
      key: data
      label: Data foundation
      horizon: next
      owner: Technology
      status: planned
      confidence: Medium confidence
      slot: 1
```

Transferring an initiative keeps its incoming/outgoing dependencies. Duplicate
horizon/slot occupancy is refused. Removing an initiative with dependencies
requires explicit `cascade: true`, or prior dependency edits. Removing a
populated horizon requires explicit cascade or transferring all its initiatives
first. A status in use cannot be deleted until all initiatives are reassigned.
`set/layout` accepts `card_height_pt` and `rect`. Inspect all connectors after
changes; boundary endpoints are recalculated, but custom routing needs review.

## Starting from a catalog example

For a composed Now/Next/Later example, preserve the frame and surrounding
narrative. Explicitly add a typed `wmds/component/portfolio` local component via
the diagram source workflow, or replace a named local rectangular leaf using
first `initialize/source`, a complete `model`, and `cascade: true`. This declares
an authored redesign; no dates, confidence, status or owner are inferred from
stock placement or generic cells. Template pins and predecessors are retained.
A complete stock multi-node composition needs deliberate coordinated removal of
superseded nodes, not a destructive whole-slide import.

There may be 1–8 horizons, 1–80 initiatives and 0–160 dependencies; measured fit
is the actual acceptance gate. Branch decisions, runbook procedures, readiness
scores and quantitative metrics retain their distinct semantics. Bound arguments
need first `materialize/source` or updates to their original values. Native
movement does not reassign horizon, owner, confidence or dependency meaning;
adopt those deliberately through a source patch after geometry reconciliation.
