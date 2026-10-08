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

## Draw an attached arrow

```sh
pptxgengo design project diagram connect --project ./deck --slide architecture \
  --id service-edge --from node06 --from-site right --to node07 --to-site left \
  --actor deck-author --reason 'Show the dependency'
```

Review the preview; repeat with `--apply`. Straight arrows support stock `block`
and `editable-block` rectangle targets and top/left/bottom/right sites. Authored
node movement recalculates endpoints. `--head` accepts none/start/end/both;
`--style` accepts solid/dashed/dotted. Elbows, obstacle avoidance, custom bends,
and container connection sites are not implemented by this command. Inspect
crossings and inner-container clearance; frame-box checks do not establish
visual or full stroke/arrowhead clearance.

## Reconcile native geometry

Build a baseline, then edit a separate working copy under `working/`:

```sh
pptxgengo design project reconcile propose --project ./deck --geometry \
  --edited ./deck/working/architecture-edited.pptx --out ./deck/reviews/geometry-1
```

Review `geometry` fields alongside text fields and every `manual_review` item.
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
PowerPoint. Added/deleted/duplicated/reparented native objects, changed routes,
formatting, rich copy and ambiguous mappings remain explicit review. Do not
report complete synchronization while any edit remains unresolved.

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
