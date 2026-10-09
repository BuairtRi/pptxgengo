# Compose processes, handoffs and decisions

A process describes work and its outcomes. Actor lanes describe responsibility;
columns describe visual progression, not elapsed time. A fork is a decision only
when outcomes have names. Multiple incoming links are a join. The position of a
shape does not establish a dependency, actor or decision outcome.

## Choose the source

Inspect installed help and the project lock first. Use `project process` for a
local `wmds/component/process` model. Use `project cycle` for a repeating cycle,
`project journey` for a road metaphor and `project gantt` for period schedules.
SIPOC, hierarchy, metrics and checklist templates retain their own content/table
meaning; do not convert them to graphs just because they share a family tag.

```sh
pptxgengo design project diagram inspect --project PROJECT --slide SLIDE
pptxgengo design project process inspect --project PROJECT --slide SLIDE --node FLOW
pptxgengo design project process patch --project PROJECT --slide SLIDE --patch change.yaml
pptxgengo design project process patch --project PROJECT --slide SLIDE --patch change.yaml --apply
```

Inspection returns the source hash, keyed model, measured geometry and any render
error. Preview must succeed before apply. The source transaction preserves its
predecessor and refuses stale hashes, pinned/shared templates and active native
geometry overrides. Resolve those through their own explicit workflows first.

## Author the graph

A model contains `lanes`, `steps`, `links`, `columns`, `label_width_pt`, and
`step_height_pt`. Give every lane, step and link a stable key. A step contains
`key`, `label`, `lane`, `column` and `kind`: `process`, `decision`, `join`, `start`
or `end`. Multiple branches can have unequal numbers of steps. Keep one step per
lane/column. More actors need more height; more columns need more width. Do not
shrink type to force a process onto one slide: split it at a meaningful handoff.

Links address `from` and `to` step keys; a decision's outgoing links require
unique nonempty `outcome` labels. Loops use explicit links back to previous
columns. `start`, `current` and the `end` key list are distinct markers. Markers
do not infer topology or change step kind. Stable links survive source array
reordering. Moving a step to another lane requires an explicit complete step
replacement so responsibility is intentional.

```yaml
schema: pptxgengo.process-patch.v1
expected_source_sha256: COPY_FROM_INSPECT
actor: Operator
reason: Add a second review outcome and join
node_id: flow
operations:
  - action: set
    entity: step
    key: revise
    step: {key: revise, label: Revise, lane: owner, column: 3, kind: process}
  - action: set
    entity: link
    key: needs-work
    link: {key: needs-work, from: review, to: revise, outcome: Needs revision}
  - action: set
    entity: link
    key: revised-join
    link: {key: revised-join, from: revise, to: merge}
  - {action: set, entity: current, key: revise}
```

Set operations add or replace a complete lane/step/link. `reorder` requires every
key of that entity exactly once. Lane order changes vertical layout; step array
order preserves explicit columns and relationships. `set/layout` accepts
`columns`, `label_width_pt`, `step_height_pt` and `rect`. Deleting an incident step
requires removing/reassigning its links and markers first, or an explicit
`cascade: true`. Populated lane deletion requires the same acknowledgement.

## Routing and geometry

Endpoints are recalculated at step boundaries from the current layout. Default
cross-lane routes use orthogonal bends. For a loop or congested diagram, set a
link's `route` to explicit component-local intermediate `[x, y]` points. Endpoints
are still derived from step geometry. Review all routes after changing lanes,
columns, counts or labels; default paths alone do not prove obstacle avoidance.
Preview refuses segments through a step's body or within 2 pt of an unrelated
step. Attached endpoints may touch their declared step edge; that does not allow
the rest of a route through that step. Refusal names the link, segment and
blocking step. Author an explicit detour through open space or rearrange steps,
then preview again. The guard checks route geometry; it does not choose a new
business dependency or silently replan your authored path.

Outcome labels include a measured background patch and must clear every step by
at least 2 pt. A short gap between adjacent steps may fit an arrow but not its
label. Preview refuses an overlapping label rather than letting step surfaces
cover it. Author a detour through open space and a component-local
`label_position: [x, y]` at the patch center, increase spacing, or shorten the
outcome. The anchor requires a named outcome and finite coordinates; the whole
patch must stay within the component allocation. For example, a YES route can
rise above both step boxes with its label centered on that upper segment. Review
the route and label together after any count, lane or geometry change.
Keep waypoints inside the component allocation, with clearance around steps,
labels and the framing device. Source fit rejects allocation or text overflow.

## Initialize an existing composed example

A stock swimlane uses positional lanes and an older source contract. For a
semantic graph, author the complete keyed model and explicitly replace a named
local rectangular leaf:

```yaml
operations:
  - action: initialize
    entity: source
    cascade: true
    model:
      columns: 3
      label_width_pt: 90
      step_height_pt: 48
      lanes: [{key: owner, label: Owner}]
      steps:
        - {key: start, label: Start, lane: owner, column: 0, kind: start}
        - {key: review, label: Review, lane: owner, column: 1, kind: process}
        - {key: done, label: Done, lane: owner, column: 2, kind: end}
      links:
        - {key: start-review, from: start, to: review}
        - {key: review-done, from: review, to: done}
      start: start
      current: review
      end: [done]
```

This is an explicit semantic redesign, not an automatic import. Keep the source
hash, selected node ID and existing rectangle; related narrative/metrics/table
nodes remain separate and must be updated deliberately. Model coordinates are
not accepted: use placement via `set/layout`. Bound arguments require first
`materialize/source` to freeze resolved values, or editing the original bound
values instead.

## Native edits

Edit an owned PowerPoint copy. Use the project's geometry/text reconciliation
packet, retaining the original package and unknown objects. Moving a step does
not change lane responsibility or links. Explicit process patches persist those
semantic decisions. Do not claim arbitrary copied shapes become process steps.
