# Native editing inventory and family pilot

## Inventory an actual build

Source builds after v4.1.0 add:

```sh
pptxgengo design project editability --project ./deck > native-editability.json
```

The command reads the state-pinned current build through the verified receipt.
For a historical build, supply both `--build BUILD_ID` and a trusted
`--receipt-sha256 SHA256`. It changes neither source nor build files.

The report records build/receipt/PPTX/generation pins and every actual native
object's Selection Pane name, native kind, stable token, source fields,
paragraph/cell count, immediate group parent, top-level owner and nesting depth.
Compiled component definitions are listed where available; family categories
such as table/image/group are native structures, not a semantic template sweep.
Names are display labels. Generation/slide/shape tags remain correspondence.

`text_and_paint_in_same_shape` counts native text shapes with an explicit fill.
It does not imply correct text positioning, suitable resize behavior or an
intuitive selection target. Group presence and nesting counts are measurements,
not desktop qualification. Manual field mappings and absent source bindings stay
visible. Table cell coordinates are inventoried; table/chart adoption is outside
the plain-text reconciliation scope.

## Bounded representative pilot

Headless generated fixtures include a keyed card row with title/body fields,
a list, a native two-column table and a two-node diagram with a connector. They
establish actual package ownership, source-field visibility and cell counts.
No frozen template or shared library definition is rewritten by the inventory.
This is the starting baseline before selecting serialization changes. Prepare
an owned standalone synthetic fixture for later desktop tasks:

```sh
PPTXGENGO_EDITABILITY_PILOT_OUT=/absolute/NEW-pilot-project \
  go test ./internal/deckproject \
  -run '^TestNativeEditabilityPilotListTableAndDiagram$' -v -count=1 -timeout=3m
```

The new fixture retains the source, exact immutable build, native inventory and
file hashes. It is not a full portable runtime package: build/adoption still
requires the recorded bundle. Copy its `builds/<build-id>/deck.pptx` to a separate
working file before desktop editing. Preparation opens no PowerPoint app and
records no qualification.

On **both** Mac and Windows desktop PowerPoint, use a separate edited copy:

| Family | Task | Evidence to retain |
| --- | --- | --- |
| Cards/pillars | Edit one title and one body; select and move a whole card; align two cards; resize a card. | Selection Pane target, before/after geometry, copy, wrapping and ownership. |
| Lists | Edit one item; select/move the intended list unit; inspect bullets and line breaks. | Exact copy/paragraphs, selected unit and nesting. |
| Table | Edit one cell and adjust a column. | Cell address/copy, row/column geometry and native visual result. |
| Diagram | Edit the node copy; move/align the unit; inspect any connector adjustment. | Group ownership, z-order, endpoint behavior and text. |
| Identity | Save As; reorder; duplicate/delete/ungroup task copies. | Saved PPTX, lineage inspection and explicit ambiguity reports. |

Repeat relevant copy and visual tasks at Comfortable, Compact and Dense tiers
using an exact source/bundle pin that supports those tiers. Record actual app/OS
versions, reviewer, timestamps, tasks, saved deck hashes, full-size screenshots
and each verdict. Count the families/density cases exercised; never extrapolate
these pages to the complete gallery.

Compare edited copies with `project reconcile propose`. Supported text should
retain stable source correspondence; rich copy, geometry, formatting and changed
ownership must stay explicit. Rebuild reviewed source and inspect the actual
native output. Reconciliation alone does not qualify editing ergonomics.

## Current evidence and remaining work

Structural/CLI tests cover card parent/row hierarchy, exact title/body fields,
repeat inventory, protected metadata, an actual native table's six cell bodies,
and list/diagram inclusion. This establishes inventory capability only.

One generated synthetic fixture has 68 native objects, ten groups, maximum group
depth two, twelve plain baseline fields and a six-cell native table. Its diagram
connector is a group of three ordinary shapes (a segment and two arrowhead
parts), with zero native `cxnSp` connectors. The report discloses that endpoint
attachment is not established. This is a concrete pilot gap: moving a node must
not be assumed to move or reconnect this arrow. These counts describe this exact
fixture, not the library.

The earlier Mac PowerPoint Save As trial timed out without a saved output.
Mac and Windows task/visual evidence, density coverage, meaningful serializer
changes and family rollout decisions remain pending. The report states
`desktop_qualification: not_recorded`. Stable v4.1.0 contains neither this
inventory nor the newer reconciliation commands.
