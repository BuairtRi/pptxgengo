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
This is the original baseline for comparison with the explicit components below. Prepare
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

The earlier automated Mac Save As trial timed out; the retained native UI
Documents-folder round trip subsequently passed (see [round-trip evidence](native-roundtrip.md)).
The combined-diagram task below adds bounded Mac movement/endpoint evidence.
Other family tasks, both-platform density qualification and rollout decisions
remain pending. The original structural inventory states
`desktop_qualification: not_recorded`. Stable v4.1.0 contains neither this
inventory nor the newer reconciliation commands.

## Explicit combined blocks and attached connectors

Maintained source adds two authoring components:

- `wmds/component/editable-block`: one filled native rectangle containing one
  plain text field. It keeps the original `block` text measurement, inner text
  allocation, font/style, alignment and fill. It has no extra surface/text group.
- `wmds/component/attached-connector`: a native straight `p:cxnSp` with declared
  same-slide rectangle endpoints. Endpoint references use the compiled authored
  node IDs; no endpoint is inferred from positions, names in an edited deck or copy.

For example, a local template can bind `input_copy` and `output_copy` string zones
and add these nodes (the connector allocation must contain both endpoint sites):

```yaml
nodes:
  - id: input-output
    kind: component
    placement: {zone: body, span: {start: 1, count: 12, y: 36, h: 144}}
    definition: {scope: shared, id: wmds/component/attached-connector}
    arguments:
      from: {node: input, site: right}
      to: {node: output, site: left}
      head: end
  - id: input
    kind: component
    placement: {zone: body, span: {start: 1, count: 6, y: 36, h: 144}}
    definition: {scope: shared, id: wmds/component/editable-block}
    arguments: {surface: subtle, text: {binding: input_copy}, style: body}
  - id: output
    kind: component
    placement: {zone: body, span: {start: 7, count: 6, y: 36, h: 144}}
    definition: {scope: shared, id: wmds/component/editable-block}
    arguments: {surface: subtle, text: {binding: output_copy}, style: body}
```

Forward references are supported. Sites are `top`, `left`, `bottom`, `right`
(DrawingML zero-based 0, 1, 2, 3). Connector `style` accepts `solid`, `dashed`,
`dotted`; `head` accepts `none`, `start`, `end`, `both` (default `end`). `ink`
uses the existing scene color references (default `strong`). Missing, ambiguous,
coincident, same-object and unsupported endpoints fail the build. Only declared
`editable-block` targets are accepted by this component; rotated/flipped targets,
custom routes, other shape kinds and cross-slide links are outside this pilot.

Combined blocks require positive dimensions, width greater than 24 points,
measured fitting plain copy and a simple fill without a separate border.
Interpreted rich/footnote markup and outlined surfaces are refused. Existing
shared `block` definitions and static `connector` paths retain their original
serialization. This authoring choice does not migrate stock templates.

`project editability` records declared endpoint tokens/sites from the actual
native package and the explicit pilot definitions. This establishes the written
attachment references. It does not establish PowerPoint rerouting or move/resize
behavior. The standard [native connector contract](https://learn.microsoft.com/en-us/dotnet/api/documentformat.openxml.presentation.connectionshape?view=openxml-3.0.1)
and [PowerPoint connection API](https://learn.microsoft.com/en-us/office/vba/api/powerpoint.connectorformat.beginconnect)
describe the intended native relationship; both desktop platforms still require
the tasks above.

Prepare the separate synthetic combined-shape/attached-connector fixture:

```sh
PPTXGENGO_NATIVE_DIAGRAM_PILOT_OUT=/absolute/NEW-native-diagram \
  go test ./internal/deckproject \
  -run '^TestNativeEditabilityCombinedBlocksAndAttachedConnector$' -v -count=1 -timeout=3m
```

Headless tests verify one filled/text object per node, exact source fields,
serialized inner text insets, endpoint IDs/sites, forward/reversed geometry and
bounded refusal paths. Original versus combined block plans match at three
V11 density tiers for body, small and label roles. Three reviewed fields can be
adopted and rebuilt while retaining the connector and combined native units;
connector/geometry changes remain manual review. This is limited structural and
source round-trip evidence, with no desktop visual or whole-library verdict.

## Mac combined-diagram task — 2026-10-07

Native UI actions in PowerPoint 16.113.4 (16.113.26100421), on macOS ARM64,
opened an owned copy under `~/Documents/pptxgengo-qualification/`. Selecting the
filled input rectangle selected its text and paint together. Two keyboard nudges
and a downward drag moved that native unit. The attached straight connector
followed its right-center endpoint while the output node stayed fixed. Editing
the input text produced exactly `Reviewed input`; a separate Save As retained
all six baseline identities. The UI visibly showed the rerouted connector and
combined node after the move. This records agent-operated fixture actions,
not a human ergonomics or whole-family acceptance decision.

Independent receipt-pinned supplied-file verification passed at
2026-10-07T13:46:38Z. Input y changed from 2,514,600 to 3,200,401 EMU
(685,801 EMU, approximately 54 points); x, width and height stayed unchanged.
The saved connector retained the exact tagged input/output references and
right/left connection sites. Its actual transform reaches both current
rectangle centers with at most one EMU endpoint rounding. It remains a straight
native connector, and both nodes retain native rectangle text/fill. Original
baseline files and receipts remain unchanged.

Private retained directory:
`~/Documents/pptxgengo-qualification/native-editing-macos-20261007/`.
It contains `attached-diagram/`, `attached-diagram-owned.pptx`,
`attached-diagram-moved-edited.pptx`, a native local-printing PDF, and
`diagram-supplied-verification/evidence.json`. Pins:

- Baseline PPTX: `0d627adfa70c9bdc364f452477883c680eee2a6dcfa9ea9300e4d90c8ae10344`.
- Edited PPTX: `6ae720acc38b945f81ebd038357b1c468f56d619398fe22af30b9c08fb61ee7d`.
- Native PDF: `b112393a2caa7de97c9026d318ff8737bd59f8aac3ce609b509ccb15dea18615`.
- Baseline receipt: `b9dc1086d54d80afd4ad735fa1a25e995db0bb46ff68501d013e749ff0e3f274`.

The supplied-file reader does not independently prove application provenance.
The native UI observations above are separate from the on-disk verifier.
`human_acceptance` remains `not_recorded`. This task does not establish node
resize/align, cards/lists/tables, other density tiers, Windows execution, copied
slides, or geometry adoption into YAML.

### Repeat the bounded supplied-file task

Use the combined-shape fixture prepared above, work on a separate copy, move
only the input node vertically down without resizing/flipping it, and replace
its entire text with exactly `Reviewed input`. Leave the output geometry and
other text unchanged, then save a new `.pptx`. This harness is deliberately tied
to that synthetic authored fixture rather than guessing arbitrary deck objects.

```sh
PPTXGENGO_NATIVE_DIAGRAM_FIXTURE=/absolute/retained-native-diagram \
PPTXGENGO_NATIVE_DIAGRAM_EDITED=/absolute/edited-copy.pptx \
PPTXGENGO_NATIVE_DIAGRAM_VERIFY_OUT=/absolute/NEW-evidence-directory \
  make test-diagram-verify
```

It follows generation/shape tokens, verifies every immutable receipt output,
compares exact expected copy, native kinds and endpoint IDs/sites/geometry, and
retains a read-only copy plus evidence. It refuses missing/duplicate identities,
wrong copy, absent movement, changed node geometry, flipped/rotated nodes, wrong
connector sites, curved/custom connectors and detached geometry. Supplied-file
verification does not migrate the fixture's compiler or authorize source edits.
Ordinary headless Make targets clear these opt-in paths. Generated positive and
negative geometry fixtures run in short tests and the selected editability race
lane; explicit supplied-file execution is skipped without all paths.
