# Geometry editing implementation and qualification

Development work dated 2026-10-08. The promoted release is still v4.2.1.
[Operator instructions](../skills/west-monroe-presentations/references/architecture-geometry.md)
cover the usable commands. [Target scope](geometry-editing-scope.md) retains the
complete requested feature and unimplemented parts.

## Source and rendering contract

Diagram authoring changes local definitions and selected slide content through
`commitSourceChanges`. It rejects shared catalog mutations, multi-slide local
definitions without an explicit fork, implicit incident-edge deletion, invalid
bindings, and failed renderer fit. Preview loads and renders source overrides;
it writes neither authored files nor project state. Exact predecessor bytes and
an actor/reason decision accompany application.

Connector scaffolds/detach use a one-EMU allocation on a zero axis and explicit
`arguments.point_space: allocation` normalized points. Their initial endpoints
remain unchanged; moving their placement now moves their points. Older absolute
connector arguments keep their existing behavior.

A slide's optional `native_geometry` maps receipt-recorded source names to typed
DrawingML transforms: native kind/parent, offset/extent in points, rotation,
flips, and group child coordinates. `native_order` stores complete existing-object
paint orders by parent. `native_geometry_template` prevents silent reuse after a
template-reference change. Editing source placement/topology after adoption
requires an explicit native-layout reset; history is retained.

After ordinary layout and group assembly, builds apply those transforms by
bounded XML span replacement, preserving other native payloads. Affine parent
matrices resolve scales, rotations and flips into final world bounds. Overrides
must resolve uniquely to the recorded kind and parent. Frame-zone checks include
transformed descendants; attached straight connectors must retain their actual
endpoint/site geometry. Native orders require an exact existing-object set and
cannot discard interleaved non-object XML.

The layout report retains text measured in authored coordinates and separately
records final native world bounds. Transform scaling is not a claim about native
text reflow or typography acceptance. Frame boxes are checked; full ink extent,
inner-container padding, obstacles and crossing diagnostics remain open.

## Reconciliation protections

`reconcile propose --geometry` adds geometry fields to the existing receipt-backed
closed packet. It renders current source, compares original/current/edited
transforms independently of plain text, and matches edited objects by lineage
tokens. Native renaming or numeric-ID changes do not become inferred ownership.
The exact report is replayed from retained inputs before adoption. Decisions are
explicit, drift is rejected, exact bytes are retained, and adoption is idempotent.

An authored group child-coordinate-space change is deliberately unresolved,
because applying the old native transform to newly rebased children could double
apply a move. Rebuild and review a new baseline first. Whole local leaves and
explicitly mapped block copies can now be reconciled. Partial deletions, unmapped
additions, ambiguous identities, reparenting, connector route/attachment edits
and unsupported formatting remain visible.
Legacy text proposals now also report paint-order changes.

## First-slice evidence (PR #31)

The asset-free `architecture/nested` catalog specimen was detached and built
using a fixed development binary and the installed V11 bundle. Task files are
under `~/Documents/pptxgengo-qualification/geometry-v1-20261008/`; original builds
and working copies are separate. No private artifact is checked into the source
repository.

- Controlled XML cases: no-op, move, resize, connector move, paint reorder,
  addition, deletion, supported combined edits, and combined structural edits.
- Supported combined case: four adopted transform/order proposals; zero unresolved
  fields or manual items; all 41 native object transforms and all paint orders
  equal after rebuild.
- Native addition/deletion cases retain untagged/missing-object review items.
- PowerPoint 16.113.4 opened/exported original, edited and rebuilt decks from the
  same explicitly qualified Documents staging folder. The edited and rebuilt
  1920×1080 PNGs are byte-identical, SHA-256
  `20dd306c48c8762de99f5d3d14a9c8a42542cbbcfea372d552f09c7fdab7514e`.
- Every exported page was visually inspected. The intended moved block, taller
  block and translated connector persist without changing the frame or copy.
- A complete numbered version, private ZIP share and extraction retain native
  geometry, paint order and template identity.
- Focused Go coverage includes rotation/flips, unchanged child coordinates,
  three-way conflicts, source drift, idempotent adoption, frame rejection,
  explicit group-basis review, template pinning, source add/remove, preview
  safety, distribution, recalculated arrow endpoints and incident-edge decisions.

These are controlled XML edits followed by actual PowerPoint rendering, not
PowerPoint UI editing/Save As geometry qualification. Windows desktop execution,
general card/image/table acceptance, broader structural adoption, bent arrows and the
other target-scope items remain pending. No release qualification or complete
geometry synchronization is claimed.

Each persisted transform also pins its authored geometry basis with
`source_geometry_sha256`. Changes to the underlying node/child coordinates are
rejected instead of applying an old transform to a newly generated coordinate
space. Use an explicit layout reset and fresh baseline to review that change.
The geometry proposal renderer verifies the actual current executable, fonts and
bundle against the project lock before computing current YAML geometry.

## Reviewed structure contract

`structure` proposals cover complete top-level project-owned local leaves.
All receipt-bound native parts must be missing; partial deletion stays manual.
Ambiguous/untracked slide identities block deletion. A source-node fingerprint
includes bindings, values, frame/profile and retained native transforms, so
simultaneous YAML changes require a conflict decision. Surviving attached edges
block deletion; missing edges require separate decisions. Shared local definitions
must be forked and revision-pinned definitions deliberately revised first.

`--structure-map` adds a closed `structure-map.json` input to the review packet.
Explicit edited-object/source-node/new-node mappings support stock block and
editable-block payloads. Native structure/style must match the verified source;
only transforms and maintained plain string content vary. New nodes and content
bindings are independent. The actual renderer supplies source-basis hashes for
the generated copy before typed native transforms are persisted.

Copied identity tags are ignored only on explicitly selected new subtrees during
analysis. Every inherited identity must leave exactly one original instance
outside all mapped copies. Untagged copies need no tag normalization. Exact input
PPTX bytes are retained, with their original hash in reports/receipts; deterministic
normalization and mapping are replayed before any source write. No native XML is
imported into authored YAML. Copy root order couples supported structural changes;
partial acceptance is refused when that order depends on another proposal.

Text, geometry and structural decisions share one authored AST and one guarded
transaction. Unused values/zones and deleted-node overrides/order entries are
pruned. Reviewed text, other transforms and source comments are retained. Final
frame validation precedes writes; idempotence and input/source drift checks apply.

Qualification uses `scripts/qualify-geometry-roundtrip.py --structures --native`
with fixed development binaries and private output under Documents. It covers
combined move/resize/connector translation/reorder, whole service deletion, and
an explicitly mapped monitoring copy. The final comparison checks every transform,
paint order, source topology, numbered version and private ZIP extraction, followed
by actual PowerPoint exports. Controlled XML edits do not establish GUI editing/
Save As qualification. Inner-container/ink clearance and bent routes remain open.

Second-slice private evidence is under
`~/Documents/pptxgengo-qualification/geometry-structure-20261008/`:

- `qualification-01`: untagged copied block plus deletion and transforms; all
  41 transforms and orders equal after rebuild, numbered snapshot/ZIP extraction
  passed. Edited/rebuilt PowerPoint PNG bytes matched (the deliberately displaced
  copy also illustrated that inner-container overlap checks are not implemented).
- The revised combined fixture places the monitoring copy in the removed service's
  slot. `qualification-03` uses the final fixed development binary and covers
  both untagged and inherited-tag copies. Combined inherited-tag edits produce
  three transform proposals and two structural proposals, with no unresolved
  items; all transforms and orders, source topology, repeat adoption and private
  version/share/extraction match. This run does not request native rendering.
- `qualification-02` passed reconciliation, rebuilding and sharing but the second
  PowerPoint export timed out. A subsequent file-access probe returned error
  `-9074`; the cause was not established. GUI inspection was unavailable. The
  final revised/inherited-tag specimen therefore has no native rendering signoff.
- Focused tests additionally cover independent copy text, copies with inherited
  tags plus simultaneous original movement, refusal to map an original as a new
  copy, style/ownership rejection, lost tags, partial deletion, related structural
  decisions, source conflicts/drift, off-frame refusal before source writes and
  repeated idempotent adoption. Full GUI/Save As and Windows qualification remain
  separate acceptance work.
