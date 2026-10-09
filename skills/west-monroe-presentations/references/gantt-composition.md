# Compose Gantt schedules and phase gates

Use this runbook for a dated or period-based schedule: variable workstreams,
lanes, tasks, phases, workshops, milestones and decision gates. A catalog plan is
an illustrative starting composition. Keep its visual language and frame while
changing the work that the operator actually needs to show.

These commands require the development CLI containing `project gantt`; the
promoted v4.2.1 binary does not contain them. Confirm with
`pptxgengo design project gantt inspect --help` before beginning. See
[installation](installation.md) to locate and deliberately select a newer CLI.

## Decide what the timeline means

Record the communication job, source of dates/owners, granularity and assumptions
in the composition log. Choose a Gantt when intervals and parallel work matter;
use a process diagram for dependencies/branching and an ordinal roadmap when
Now/Next/Later is the actual claim. A bar's proximity to another bar does not
establish a dependency. This renderer is not a critical-path scheduler.

The supported timebase is **periods**. For N labels, position 0 is the first
period's left boundary and N the last period's right boundary. `from: 1.25` and
`to: 3.5` represent period coordinates. Sublabels such as “Oct 12” are display
copy: they do not establish timezone, month length, weekends or business-day
arithmetic. Translate sourced dates into coordinates explicitly and record the
mapping. Do not accept a `dates` timebase or infer dates from labels.

Keep task identity, lane membership, time interval and native placement distinct.
Keep phase names and gate criteria consistent with supporting slides/tables.
Migration state colors retain their authored state meanings.

## Inspect the actual component

Work inside the standard portable project structure described in
[project structure](project-structure.md). Preserve immutable builds and keep
authored slide YAML under `slides/`, local template YAML under
`slides/templates/`, and shared deck assets under `assets/`.

```sh
pptxgengo design project detach --project ./deck-project --slide plan --as plan-local --reason 'Adapt the schedule to the supplied work plan'
pptxgengo design project diagram inspect --project ./deck-project --slide plan > diagram.json
pptxgengo design project gantt inspect --project ./deck-project --slide plan --node node01 > schedule.json
```

Use the node ID discovered in the diagram inspection, not the example `node01`.
Only a local semantic `wmds/component/gantt` node is supported. A timeline built
from loose bars/shapes needs a deliberate local derivative; this command does
not guess that those shapes are tasks. Local templates shared with other slides
must be forked first. Pinned local revisions need an explicit revision change.

Inspection provides the current source hash and keyed schedule. When the source
renders, it also provides frame/allocations, measured text, final native geometry
and warnings. A `render_error` means those measurements failed; it is not a
qualified render. You can still inspect the records and prepare a recovery patch,
which must successfully measure before application. Generated identities from the
stock specimen are frozen into keyed records before reordering. Prefer meaningful
new keys such as `risk-review` rather than array positions. Keys are scoped by
their group/lane; changing a task's parent changes its full native object path
even though its task key survives.

## Compose one reviewed transaction

Write a strict YAML/JSON patch. Copy `source_sha256` from the latest inspection
into `expected_source_sha256`. Every operation is ordered and the entire result
is measured before any source files are changed.

```yaml
schema: pptxgengo.gantt-patch.v1
expected_source_sha256: REPLACE_WITH_CURRENT_64_CHARACTER_HASH
actor: Named operator or agent
reason: Apply the sourced phase plan and approval checkpoint
node_id: node01
timebase: periods
operations:
  - action: materialize
    entity: source
  - action: set
    entity: task
    group: delivery
    lane: implementation
    key: build
    task:
      key: build
      kind: core
      label: Implement
      from: 1
      to: 4.5
  - action: set
    entity: gate
    key: approval
    gate:
      key: approval
      at: 4.5
      label: Approve
      callout: true
  - action: set
    entity: today
    at: 2.25
```

Replace example group/lane/node keys with the inspected identities. `kind` and
`event` must refer to the component's existing style/legend definitions. A task
record can use interval `from/to`, `progress`, `softStart/softEnd`, or an `at`
event with `milestone`, `event`, and `labelSide` where appropriate. Do not turn
interval tasks into events by leaving contradictory fields populated.

**Materialization is explicit.** If the selected component contains bindings,
the first `materialize/source` operation resolves them into local template
constants. Its copy will then be edited in that local template rather than the
previous bound slide value. Other components' bindings remain in place; only
bindings no longer used by any node are pruned. Omit materialization for an
already literal component, or edit its existing authored bound values directly
if retaining their location is preferable. Inspect the preview before adopting
this authoring change.

| Operation | Payload and effect |
| --- | --- |
| `set/group` | `key`, complete `group_value` with matching key, label/fill/lanes; adds a new group or replaces an existing one |
| `set/lane` | Parent `group`, `key`, complete `lane_value` with matching key, title/sub/icon/items |
| `set/task` | Parent `group/lane`, `key`, complete matching `task`; add, relabel or retime a task |
| `remove/group` or `remove/lane` | Parent if needed, `key`; nonempty entities require `cascade: true` |
| `remove/task` | Parent `group/lane`, `key` |
| `move/lane` | Current `group`, `key`, `to_group` |
| `move/task` | Current `group/lane`, `key`, `to_group/to_lane`; preserves the task record |
| `reorder/group`, `reorder/lane`, `reorder/task` | Parent if needed, `order` listing every existing key exactly once |
| `set/phase` | `key`, complete matching `phase` with label/from/to/rule |
| `set/gate` | `key`, complete matching string-keyed `gate` with label/at/optional callout |
| `remove/phase` or `remove/gate` | `key` |
| `reorder/phase` or `reorder/gate` | Complete `order` |
| `set/periods` | Nonempty `labels`, optional same-length `sublabels`; existing intervals must still fit |
| `set/today` or `remove/today` | Explicit numeric `at` for set; no payload for remove |
| `set/layout` | `track_pitch_pt` (24–60), `group_width_pt`, `lane_width_pt`, `legend_full_width`, `legend_size_pt`; node allocation and frame stay fixed |

`set` replaces the complete entity; it is not a partial merge. Copy an inspected
record and change the intended fields to preserve other content. Replacing an
entire group/lane also replaces its children; dropping retained descendant keys
requires `cascade: true` even on `set`. Prefer task/lane operations for isolated
edits. Legacy boolean gate keys are preserved as emphasis using
`callout` while the gate receives a stable string key.
Lane `icon: none` explicitly omits an optional graphic while keeping the heading
allocation. An empty icon preserves the renderer's legacy target-icon default.

```sh
pptxgengo design project gantt patch --project ./deck-project --slide plan --patch schedule-patch.yaml > schedule-preview.json
pptxgengo design project gantt patch --project ./deck-project --slide plan --patch schedule-patch.yaml --apply > schedule-applied.json
pptxgengo design project check --project ./deck-project
pptxgengo design project build --project ./deck-project
```

Apply requires the same source hash and retains a decision receipt/source
predecessor through the guarded source transaction. A changed source, invalid
interval, unknown style, label overflow or frame allocation failure leaves the
authored files unchanged. Native geometry/order overrides require an explicit
reviewed `reset_native_layout` via the diagram patch route before composition;
never silently discard a person's native edits.

## Fit and native review

The renderer packs overlapping intervals into tracks and measures lane labels,
task/event labels, gates and legends. More tracks use more vertical allocation;
adding lanes can require another slide. Resolve capacity by shortening copy with
approval, selecting an appropriate allocation/template, or splitting the
schedule. Do not change dates, omit work or shrink type to make a preview pass.

The stock eight-week plan has a renderer exception that places its legend in the
90–108 pt header-adjacent band. A detached local component may correctly reject
that placement because it lies outside its body allocation. An explicit compact
track pitch and full-width legend can reflow the legend inside the body; inspect
the result rather than relaxing the frame check or claiming stock geometry was
preserved exactly. The catalog's frozen example remains unchanged.

Build and render through PowerPoint using the normal skill QA workflow. Examine
every label, phase boundary, gate, today marker and legend. Verify 3 versus 8
lanes, overlapping tasks, fractional gates, start/end milestones, long right-edge
labels and an intentionally oversized schedule for the actual presentation.
Automated source tests are not desktop qualification.

Use a separate PowerPoint working copy. Receipt-backed reconciliation can report
known text/geometry changes, but **moving or resizing a native bar does not
automatically retime a task**. Added untagged bars are not imported as tasks;
deleted subshapes are not task deletion instructions. Review the native edits,
confirm their semantic meaning with the operator, then use a keyed Gantt patch
to change the task/gate/phase source. Preserve the native copy as evidence and
rebuild a fresh baseline before further reconciliation. See
[architecture and geometry](architecture-geometry.md),
[editing slides](editing-slides.md) and
[PowerPoint recovery](powerpoint-recovery.md).
