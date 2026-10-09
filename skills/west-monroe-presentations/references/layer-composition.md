# Compose architecture layer stacks

Choose the number and meaning of architectural layers for the message. A stack
usually means an ordered responsibility/dependency abstraction, not a reporting
hierarchy. The foundation is separate; cross-cutting controls apply to explicitly
selected layers. More layers are not automatically a more mature architecture.

## Select native source objects

Detach the catalog slide, inspect source objects and frame allocations, and read
installed `project layer` help:

```sh
pptxgengo design project diagram inspect --project PROJECT --slide SLIDE
pptxgengo design project layer inspect --project PROJECT --slide SLIDE
pptxgengo design project layer inspect --project PROJECT --slide SLIDE --selection layers.json
```

Ordinary numbered `layerrow` objects auto-discover in their authored unique
positive integer `n` order, resolving source bindings when present. After a
reorder, a later inspection/layout uses those authored numbers; incidental paint
array order and native movement do not establish semantic order. Duplicate,
noninteger, nonpositive or unresolved numbering refuses automatic selection.
Supply a reviewed explicit selection array when numbering is ambiguous or absent;
its array order declares the intended semantic order. For icon rows, 3-D planes and
compound layer-map rows, supply an explicit JSON array:

```json
[
  {"key":"experience","members":["experience-row"],"label_node":"experience-row","text_node":"experience-row","number_node":"experience-row"},
  {"key":"data","members":["data-number","data-label","data-description"],"label_node":"data-label","text_node":"data-description","number_node":"data-number"}
]
```

Use stable IDs from diagram inspection. Each member belongs to one layer. Map the
objects that actually own its label, description and optional number. A layer can
retain one `layerrow`, icon `node`, `plane`, or explicitly mapped text/block
members. Cross-cutting cards, foundation and external/system planes are excluded
from membership. Geometry does not infer membership.

## Customize and fit

Schema `pptxgengo.layer-patch.v1` requires `expected_source_sha256`, `actor`,
`reason`, explicit `layers` selections, and `operations`:

```sh
pptxgengo design project layer patch --project PROJECT --slide SLIDE --patch change.yaml
pptxgengo design project layer patch --project PROJECT --slide SLIDE --patch change.yaml --apply
```

- `add/layer`: new `key`, existing `prototype`, required `label`, optional `text`
  and `surface`. It clones the prototype's native source members with fresh IDs;
  it does not alter unrelated objects. Inspect the resulting IDs after apply.
- `set/layer`: existing `key` and explicit label/text/surface updates.
- `remove/layer`: existing `key`. Attached incident edges require explicit
  `cascade: true`, or revise the edges first.
- `reorder/layer`: every current key exactly once. References remain stable.
- `set/foundation` or `set/controls`: `node_id` and complete source `arguments`.
  Keep the object's visual type; use native card body blocks such as
  `body: [{bullets: [Security, Audit]}]`, not a string or invented text block.
- Every patch requires `set/layout` with a complete reviewed stack allocation.

```yaml
- action: set
  entity: layout
  layout:
    rect: {x_pt: 216, y_pt: 36, width_pt: 450, height_pt: 216}
    gap_pt: 0
    overlap_pt: 0
    palette: sequence
    foundation_node: foundation
    foundation_gap_pt: 18
    controls_node: shared-controls
    controls_mode: span
```

Use the rectangle field names shown by inspection (the example is points relative
to the frame zone). `gap_pt` and `overlap_pt` cannot both be positive. Overlap is
for deliberately layered 3-D planes; planes paint bottom-first. Flat stacks use
zero overlap and shared edges. Compound member geometry scales proportionally
inside each row; text is remeasured and never shrunk silently to force a fit.

`palette: preserve` retains each layer's explicit surface meaning.
`palette: sequence` redistributes design-system inverse/strong/subtle/outline tones
across the ordered count with contrast checks. With six layers some tones repeat;
sequence is ordinal, not six invented assessment categories. Do not use it when
color identifies a system/domain. Foundation retains its own surface.

The foundation's Y follows the stack plus `foundation_gap_pt`; its width and
height remain explicit. Controls get an updated layer-range label. `span` updates
the card's Y/H to span the stack; `label_only` preserves a side-column card's
independent allocation. Use the latter for split layouts where spanning would
collide with side copy.

Begin with `materialize/source` only when intentionally converting selected
resolved bindings to local constants. Controls/foundation source updates also
require this acknowledgment when bound. Shared/pinned templates, stale hashes,
and active native geometry/order overrides are refused. Keep predecessor and
raw-native evidence before changing source.

## Catalog-specific selection

| Catalog variants | Select and preserve |
|---|---|
| `architecture/layers`, `layers-left`, `layers-nav`, `layers-split` | Numbered main rows; explicit foundation/control IDs. Split controls use `label_only` |
| `architecture/layers-icons` | Main icon nodes; label maps to text, description to sub; foundation icon node is separate |
| `architecture/layers-3d`, `layers-3d-systems` | Main three co-linear planes only, visually ordered top-down; exclude external/system planes and blocks; explicit overlap and painter order |
| `architecture/today-vs-target` | Select the four numbered Target rows only; preserve both containing frames and all Today objects, raw connector paths, transition arrow and explanatory copy. Do not pack the whole before/after diagram as a single stack |
| `architecture/layer-map` | Each number block + label + description forms one explicit compound layer. Right-hand matrix is a separate relation model |

The layer-map matrix, companion descriptions and legacy raw connector point lists
are not inferred from row positions. Inspect/update their intended mappings and
connector attachments explicitly after topology changes, using diagram routing
and chart/table operations as appropriate. Native movement remains layout until a
reviewed semantic decision establishes order/membership. Verify 3/5/6 counts with
long labels, full frame clearance, controls, foundation and remaining content.
Use immutable builds, separate native edit copies, source reconciliation, numbered
snapshots and portable project archives.
