# Compose roads and branching journeys

A road illustrates progression; it does not promise dates or dependencies.
A decision fork shows mutually named options. A parallel fork shows concurrent
streams. Keep the current milestone separate from the chosen option.

## Inspect and patch

Use local typed `wmds/component/road` or `wmds/component/roadfork` nodes after
materializing/detaching the selected catalog example. The six road and seven
road-fork catalog variants use these contracts, including split and criteria
variants. Related cards, criteria tables and side narratives stay independent.

```sh
pptxgengo design project journey inspect --project PROJECT --slide SLIDE --node ROAD
pptxgengo design project journey patch --project PROJECT --slide SLIDE --patch change.yaml
pptxgengo design project journey patch --project PROJECT --slide SLIDE --patch change.yaml --apply
```

Use the returned `source_sha256`. Inspection resolves the existing source into
keyed milestones and branches, keeping source copy, geometry options, numbering
and current/chosen meaning. A measured preview precedes guarded apply. Stale
source, shared/pinned templates and active native overrides are refused.

## Change counts and options

Milestones have stable `key`, `label`, optional `date`, `text` and `n`. `branch`
selects the destination branch for `set/milestone`; omit it for the trunk. A
branch has `key`, `title`, `text`, and a complete `milestones` list. Branch counts
are 2–3; fork trunk counts are 0–12 and each branch has 1–6 milestones. Roads have
1–24 milestones. These are safety limits, not a guarantee that every count fits.
Short copy and wider allocations are important; reject measured overflow and
split a long journey instead of compressing the fonts.

```yaml
schema: pptxgengo.journey-patch.v1
expected_source_sha256: COPY_FROM_INSPECT
actor: Operator
reason: Add a preparation step before the decision
node_id: road
operations:
  - action: set
    entity: milestone
    key: prepare
    milestone: {key: prepare, label: Prepare, date: Week 1, text: Confirm scope}
  - action: reorder
    entity: milestone
    order: [prepare, EXISTING_TRUNK_KEY]
  - {action: set, entity: current, key: prepare}
  - {action: set, entity: chosen, key: OPTION_KEY}
```

`set/chosen` applies only to decision forks. Changing branch order atomically
rewrites the old positional chosen index; the chosen option keeps its identity.
`set/current` can select a milestone before or after the fork. It likewise
survives milestone and branch reordering. `remove/current` and `remove/chosen`
clear those distinct facts. Removing the current milestone requires explicit
cascade or reassignment first; populated branch removal requires cascade and
clears affected current/chosen markers. Reordering requires all keys once.

Set a complete `fork` using `label`, `text`, `tag`, `at` (fraction of road width).
`set/layout` replaces layout options, preserving the rectangle when omitted.
Road options: `direction`, `roadW`, `amp`, `waves`, `labelW`; milestone `at` is a
0–1 position and `side` is `above`/`below`. Fork options: `roadW`, `forkW`, `small`;
branch milestone positions come from order, not authored road `at` or `side`.
Copy current layout from inspection before replacing it to retain its options.
A road's current marker uses the stock marker; fork milestones may set
`hereLabel`. Do not apply geometry fields from the other kind.

## Preserve intent and surrounding content

Unequal branch lengths are valid. An option's title is its meaning, not merely
an A/B label. Change criteria and detail cards explicitly when options change.
Neither a relocated pin nor changed road curvature establishes a deadline.
Use `project gantt` for schedules and `project process` for actual branching
work, joins, ownership and explicit dependencies. Portfolio horizons and
initiative dependencies are a different model; do not overload road milestones.

Arguments with bindings require first `materialize/source`, or edit the bound
values. Preview validates the complete proposed source and all companion nodes.
Edit a separate native copy and reconcile supported geometry/text explicitly;
PowerPoint position alone does not infer current/chosen/branch membership.
