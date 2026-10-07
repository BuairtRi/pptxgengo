# Native text field mappings

This source change supplies a shared baseline contract for native editing and
bounded reconciliation. It does not apply edited PowerPoint changes to YAML or
qualify direct editing on a desktop.

`object-map.json` retains the deck/object-map v1 envelope and declares the
additive `pptxgengo.native-text-model.v1` model. Each native object records kind,
paragraphs, ordered runs, exact text, formatting hashes, native structure hash,
and explicit source bindings. Existing logical, slide, item, native part/ID/name
and source/template baseline information remains available.

## Addresses and exact text

Addresses are zero-based body and paragraph positions. Table text also records
row and column, with separate bodies for each cell. Runs retain ordinal, kind
(`r`, `br`, `fld`), text and properties hash. Text concatenates runs within a
paragraph; explicit breaks and paragraph boundaries remain newlines. Empty
paragraphs are retained. Run boundaries do not become invented paragraph breaks.

Paragraph and end properties are hashed separately. Dynamic fields, bullet or
numbered paragraphs and unknown content remain explicit review items. Table
addresses identify native cells; they do not establish source-cell adoption.

The native structure hash excludes only DrawingML text-leaf contents. Geometry,
names, run structure, formatting and other shape payload remain included.
This is a conservative change detector, not a renderer or equivalence claim.
Serialization changes during Save As may require manual review until qualified.

## Supported baseline field mapping

A plain-text baseline requires exactly one explicit string source field, a stable
source slot, exact source/native text equality, a text shape and no unsupported
paragraph features or multiple rich runs. It records a stable field identity,
source slot/pointer, source/native hashes and paragraph addresses. Baseline
support does not establish edited-file identity survival or update permission.

Typed card title and body objects now select their specific assignment, rather
than inheriting every card field from the containing row. The keyed slot retains
item identity when an authored array is reordered. Frame fields use the actual
source slot, including raw `values.slots`, instead of guessing a direct value.
Ambiguous objects, rich text, cells, source/native break differences and missing
stable bindings remain `manual_review`; source text is never guessed by similarity.
Literal shared-template copy is not made mutable source content by this contract.

## Qualification and next steps

Targeted fixtures cover mixed runs, explicit and empty paragraph breaks, cell
coordinates, dynamic/bullet review flags, unique exact fields, typed card title/
body bindings and separate text/geometry fingerprints. A headless generated
PowerPoint supplies the actual representative card objects. Desktop behavior is
still pending.

Next: tie edited objects to proven deck/build lineage, implement three-way text
proposals and explicit conflicts, preserve unsupported structure/format reports,
and adopt reviewed proposals through guarded source mutation. Evaluate Save As,
reorder, duplicate/delete/ungroup and editing on Mac and Windows before extending
supported identities. Grouping/geometry pilot changes must preserve this mapping
contract and receive visual/edit-task review.
