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

## Review native bars as possible schedule changes

Use a separate PowerPoint working copy of an immutable build. Prepare the local
component first: materialize bound source explicitly, inspect its stable keys,
and build a fresh baseline. Use `project gantt reconcile --help` to confirm the
installed CLI includes this additional development command.

```sh
pptxgengo design project reconcile propose --geometry --project ./deck-project \
  --edited ./working-copy.pptx --out ./native-review
pptxgengo design project gantt reconcile --project ./deck-project --slide plan \
  --node node01 --packet ./native-review > schedule-review.json
```

The second command replays the receipt-backed packet and reports **proposals**.
Select the complete original keyed task group when moving work horizontally:
bars, event markers and their captions travel together. For a single solid task,
a horizontal group move can propose new `from/to` periods while retaining duration.
An individual original bar handle can resize the interval; its caption remains
source copy and is regenerated on explicit retime adoption. A horizontal move/resize
of that original single bar can also propose new `from/to` periods. A horizontal move of the original vertical gate guide can
propose a new `at`. A gate chip or label alone does not establish its period.
All tagged period anchors and their ancestor coordinate transforms must be unchanged.
The original direct keyed task group may translate horizontally; scaling, vertical
movement, rotation, flips or reparenting require explicit source interpretation.
PowerPoint may recompute a group envelope without changing its child coordinate
map; that envelope normalization remains retained native evidence.
Select the authored node ID discovered by inspection; `native_node_id` in the
semantic report retains its qualified namespace through nested local groups.
Coordinates are rebased from the authenticated task group into that common axis
parent space and are converted to fractional
periods rounded to six decimal places; no week/day/integer snapping is applied.
DrawingML quantization can affect the last decimal. Confirm the proposed
coordinates against the intended schedule; use an explicit keyed patch when a
clean integer/date-derived coordinate is required.

The command does not infer calendar dates, dependencies, lane membership,
progress, soft starts/ends or task creation/deletion. Rotated, vertically moved,
segmented, hatched or progress bars remain manual or unresolved findings. It
refuses concurrent authored changes, active native overrides or bound source;
review those independently and rebuild a baseline first. Do not reset native
layout merely to suppress that refusal.

Write a decisions YAML using `report_sha256` and proposal IDs from the returned
report. `retime` confirms a proposed meaning; `keep_source` records that the
movement does not change the schedule. A manual finding cannot be retimed.

```yaml
schema: pptxgengo.gantt-semantic-decisions.v1
report_sha256: REPLACE_WITH_SEMANTIC_REPORT_HASH
actor: Named operator
reason: Confirm the updated work plan after native editing
decisions:
  - proposal_id: REPLACE_WITH_PROPOSAL_ID
    action: retime
    reason: The task now starts one quarter-period later
```

```sh
pptxgengo design project gantt reconcile --project ./deck-project --slide plan \
  --node node01 --packet ./native-review --decisions ./schedule-decisions.yaml
pptxgengo design project gantt reconcile --project ./deck-project --slide plan \
  --node node01 --packet ./native-review --decisions ./schedule-decisions.yaml --apply
pptxgengo design project build --project ./deck-project
```

Preview measures the complete regenerated model without changing source. Apply
rechecks source, toolchain and immutable build evidence, persists only accepted
schedule facts and retains the exact edited PPTX, semantic report, geometry
report and decisions in the shared content-addressed asset store. Associated
labels/chips/packing are regenerated from those facts; the movement is not also
stored as a native layout override.

Inspect `manual_review`, `unresolved_geometry_ids`, `unresolved_text_ids` and
`unresolved_structure_ids`, plus unselected proposals. This command does **not**
adopt those changes. Preserve the original native copy and review remaining
copy/format/layout separately against the new baseline. The old packet becomes
stale after a source change. Do not apply the old generic geometry decisions
on top of semantic adoption or imply the rebuilt deck contains all native edits.

Added untagged bars are not imported as tasks; deleted subshapes are not task
deletion instructions. Use keyed Gantt patches for those confirmed changes.
Build and review the new deck through PowerPoint; the source/XML round-trip
qualification does not establish a GUI Save As or Windows qualification. See
[architecture and geometry](architecture-geometry.md),
[editing slides](editing-slides.md) and
[PowerPoint recovery](powerpoint-recovery.md).

## Keep gate labels readable

Gate chips occupy measured horizontal space above the timeline. Duplicate dates
or nearby gates can make their chips overlap even when the guide lines fit.
The compiler refuses overlapping measured gate labels. Explicitly combine gates
whose decision really is shared, remove a superseded gate by stable key, rename a
label where its meaning permits, or move its authored period. Do not silently
shift a decision date just to make labels fit. Preview the complete schedule after
retiming; an adopted date must also pass this readability guard.
