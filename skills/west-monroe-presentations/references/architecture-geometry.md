# Architecture and geometry editing

These commands are available in the geometry development build, **not the
promoted v4.2.1 binary**. Check `pptxgengo design project diagram inspect --help`
before using them. Keep diagrams in the [canonical project layout](project-structure.md).

## Customize a catalog architecture

Shared templates are immutable. Detach the selected slide before changing its
geometry or element count:

```sh
pptxgengo design project detach --project ./deck --slide architecture \
  --as architecture-local --reason 'Adapt the service boundaries'
pptxgengo design project diagram inspect --project ./deck --slide architecture
```

If other slides use the same local definition, fork it for the selected slide
first. A pinned local revision must be deliberately revised before definition
changes. Definitions live under `slides/templates/`; copy stays in the slide's
`content`/`bindings` or `values`.

Inspection returns frame rectangles, source nodes, measured text and scene
bounds, final native world bounds, supported rectangle ports, and calculated
attached-arrow endpoints. Source placements use **points relative to the named
frame zone**. Native transforms retain their **parent coordinate space**; inspect
`final_native_geometry` for positions on the actual slide. Do not add a group's
offset to its children a second time.

## Preview and apply a patch

```yaml
schema: pptxgengo.diagram-patch.v1
actor: deck-author
reason: Rearrange services and remove the unused service
operations:
  - action: move
    id: node06
    dy_pt: 12
  - action: remove
    id: node08
    incident_edges: remove
  - action: add
    id: monitoring
    node:
      id: monitoring
      kind: component
      definition:
        scope: shared
        id: wmds/component/editable-block
      placement:
        zone: body
        rect: {x_pt: 288, y_pt: 270, width_pt: 180, height_pt: 36}
      arguments:
        text: Monitoring
        surface: light
        style: body
```

```sh
pptxgengo design project diagram patch --project ./deck --slide architecture \
  --patch geometry.yaml
# Review the measured result, then apply that patch:
pptxgengo design project diagram patch --project ./deck --slide architecture \
  --patch geometry.yaml --apply
```

`move` accepts deltas or a complete `rect`; `resize` requires a complete rect.
Both require a rect placement. Adding a node supports the existing typed local
node/component definitions and literal arguments. Removing a node with attached
incident edges requires `incident_edges: remove`; otherwise the patch fails.
Removed bindings and copy are retained in exact predecessor files. No catalog
file or generated baseline is rewritten. Patches reject invalid bindings,
source fit failures, and placement outside the selected frame zone.

Align and distribute explicit nodes, using the first as the alignment anchor:

```sh
pptxgengo design project diagram arrange --project ./deck --slide architecture \
  --nodes node06,node07,node08 --align top --distribute horizontal \
  --actor deck-author --reason 'Align the services'
```

Use `--apply` after reviewing the preview. Distribution retains the supplied
order and first/last positions; insufficient space is an error.

## Keep elements inside a logical container

Visual nesting is not automatically a containment rule. Use exact native names
from `final_native_geometry` to declare the relationship on the selected slide:

```sh
pptxgengo design project diagram contain --project ./deck --slide architecture \
  --nodes node06,node07,node08 --container node05 \
  --padding 12 --padding-top 28 --padding-bottom 4 \
  --actor deck-author --reason 'Keep services inside the Services box'
```

Review the preview, then repeat with `--apply`. Uniform padding defaults to 12
points; `--padding-top`, `--padding-right`, `--padding-bottom` and `--padding-left`
override individual edges. Padding uses the container's native placement axes,
before its own rotation and parent scaling. It is not necessarily slide-space
padding. Use a larger top value to reserve the container heading.

`inspect` reports padded clearances and allocation overlaps between declared
siblings. Overlap is a warning; review intentional layers and rotated envelopes
visually. Builds, patches and native reconciliation reject objects or descendants
outside their declared allocation **before source writes**. The rules persist in
slide YAML as `diagram_containment` and pin the local template reference. They do
not move objects, reparent them, reserve ink/stroke space, or infer relationships
from appearances.

A mapped native copy inherits its source component's membership. Deleting a
member removes its rule. Deleting a container with surviving members is refused;
first reassign those members or explicitly remove the obsolete relationship:

```sh
pptxgengo design project diagram uncontain --project ./deck --slide architecture \
  --nodes node06,node07 --actor deck-author --reason 'Remove the obsolete boundary'
```

Before changing the template reference, deliberately remove obsolete memberships
and declare/review new ones against the new template. A layout reset alone does
not remove their template pin.

Preview before `--apply`. `reset_native_layout` clears adopted transforms/order
while preserving containment rules. Review clearance again after a reset.

## Draw an attached arrow

```sh
pptxgengo design project diagram connect --project ./deck --slide architecture \
  --id service-edge --from node06 --from-site right --to node07 --to-site left \
  --actor deck-author --reason 'Show the dependency'
```

Review the preview; repeat with `--apply`. Arrows support stock `block` and
`editable-block` rectangle targets and top/left/bottom/right sites. Authored node
movement recalculates endpoints. `--head` accepts none/start/end/both;
`--style` accepts solid/dashed/dotted.

Use `--route horizontal` for a horizontal/vertical/horizontal elbow, or
`--route vertical` for a vertical/horizontal/vertical elbow. `--bend .35` places
the middle segment at 35% of the first axis from the start; it defaults to 50%.
Source bend fractions must be in `[0,1]`. Straight is the default and does not
accept `--bend`. `inspect` returns the final calculated path points in slide
coordinates, including adopted transforms and bend positions. These are single
native attached connectors, not separate line fragments.

A supported native elbow bend edit can be reviewed through `reconcile propose
--geometry`; adoption persists its `bentConnector3` preset and literal `adj1`
guide in `native_geometry/<name>/route`. Guides outside the endpoint rectangle
are permitted only when their calculated allocation still fits the frame and
any declared container. A source layout reset removes adopted bend overrides
along with transforms, returning to the source-authored route.

Obstacle avoidance, custom paths, curved/other elbow presets, connector labels,
and container connection sites are not implemented. Unsupported metadata or
attachment changes remain review items. Inspect crossings and inner-container
clearance visually; allocation checks do not establish stroke/arrowhead or native
Save As fidelity.

## Reconcile native geometry

Build a baseline, then edit a separate working copy under `working/`:

```sh
pptxgengo design project reconcile propose --project ./deck --geometry \
  --edited ./deck/working/architecture-edited.pptx --out ./deck/reviews/geometry-1
```

Review `geometry` and `structure` fields alongside text fields and every
`manual_review` item.
The supported native properties are transform placement/extent, group child
space, rotation, flips, and existing-object paint order. Matching uses receipt
lineage tokens; names in the report are the recorded source binding, not guesses
from the edited deck. Changes are compared against baseline and rendered current
YAML. A changed authored group coordinate space requires a fresh baseline.

Use the existing review-decisions schema and the exact geometry field IDs:

```yaml
schema: pptxgengo.text-review-decisions.v1
report_sha256: COPY_THE_EXACT_PACKET_HASH
actor: deck-author
decisions:
  - field_id: COPY_THE_REVIEWED_GEOMETRY_FIELD_ID
    action: use_native
    reason: Accept the reviewed geometry edit
```

Actions are `use_native` or
`keep_yaml`, with a reason. Select only `native_only` or `conflict` proposals.
Run `project reconcile adopt --packet DIR --decisions FILE`, then rebuild and
render. Geometry persists in the slide's `native_geometry`/`native_order`, pinned
to `native_geometry_template`. Exact edited bytes, decisions and predecessors
are retained. An adopted attached-arrow layout must retain its actual endpoint
attachments; include incident connector transforms when required.

Native group resizing preserves its scaling rather than reflowing source text.
Measured text remains in authored coordinates; review scaled typography in
PowerPoint. Whole local leaf deletions and explicitly mapped block copies are
supported as described below. Other additions, unresolved duplicate identities,
reparenting, changed routes, formatting and rich copy remain explicit review. Do not
report complete synchronization while any edit remains unresolved.

## Reconcile removed and copied components

For a whole-component deletion, `--geometry` proposes `structure` fields with
`action: remove_node`. Review the node ID and complete native-object list before
selecting its field ID with `use_native`. This supports top-level project-owned
local leaves, not deletion of a single composite subshape or a nested group.
Untracked or ambiguous identities prevent a deletion proposal. Source edits
become conflicts; a surviving attached arrow prevents deletion. If the component
and its incident arrows are all removed, select their separate proposals together.
Unused bindings are pruned and exact predecessor files retained.

For a new copy of an existing block or editable block, provide an explicit map:

```yaml
schema: pptxgengo.native-structure-map.v1
copies:
  - slide_id: architecture
    native_object: Monitoring copy
    source_node: node07
    new_node: monitoring
```

`native_object` is the unique name of the **new copy** in the edited PPTX
(available in PowerPoint's Selection Pane). `source_node` is its baseline YAML
component, and `new_node` is an unused stable ID. A diagnostic native name selects
that pinned input only; it does not become permanent source identity.

```sh
pptxgengo design project reconcile propose --project ./deck --geometry \
  --structure-map copies.yaml --edited ./deck/working/architecture-edited.pptx \
  --out ./deck/reviews/geometry-2
```

The map supports complete top-level `block`/`editable-block` components with
matching native structure and styling. Copies can move, resize, rotate, flip,
and change uniquely bound plain text. Each copy gets independent source bindings.
The copied source component must be unchanged in YAML since the baseline.
Literal text changes, rich formatting, new arbitrary Office objects, unsupported routes,
and reparenting need separate source authoring or further implementation.

PowerPoint may duplicate identity tags along with objects. The explicit map
identifies the new copy; analysis ignores only its inherited tags and requires
exactly one original instance to remain. Exact edited bytes are preserved.
Do not map the original as a new copy or erase tags in the original build.

Select reviewed `structure` field IDs using the same review-decisions schema.
Copy proposals carry the edited root order; related supported additions/deletions
on that slide must be accepted together. Missing mappings remain review items.
After adoption, rebuild, render every changed slide, and check the hierarchy and
clearance visually. Declared containment enforces allocation padding; frame-box
checks alone do not infer visual container relationships or detect every overlap.

A reconciled native layout takes precedence over source placement. Before a
later source placement/topology change, explicitly start the patch with
`action: reset_native_layout`, `id: <slide-id>`; this clears the native transform
and order overrides and retains predecessors. Then preview the new source
layout. Never silently clear overrides during a schema/template upgrade.

Each persisted transform also pins its authored geometry basis with
`source_geometry_sha256`. Changes to the underlying node/child coordinates are
rejected instead of applying an old transform to a newly generated coordinate
space. Use an explicit layout reset and fresh baseline to review that change.
The geometry proposal renderer verifies the actual current executable, fonts and
bundle against the project lock before computing current YAML geometry.
