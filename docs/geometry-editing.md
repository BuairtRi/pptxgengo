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
transformed descendants; attached supported connectors must retain their actual
endpoint/site geometry. Native orders require an exact existing-object set and
cannot discard interleaved non-object XML.

The layout report retains text measured in authored coordinates and separately
records final native world bounds. Transform scaling is not a claim about native
text reflow or typography acceptance. Frame boxes and explicitly declared inner-container allocations are checked;
full ink extent, obstacles and crossing diagnostics remain open.

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
additions, ambiguous identities, reparenting, unsupported connector route/attachment edits
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
general card/image/table acceptance, broader structural adoption, other arrow presets and the
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
Save As qualification. Full ink clearance and broader routing remain open.

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

## Logical containment contract

`diagram_containment` stores per-member native container names and nonnegative
per-edge padding. The affine inverse of the container's native outer transform
projects every member/descendant allocation into container placement axes. Parent
scaling, rotation, flips and group child coordinate systems are included. A
surviving member referencing a deleted container fails final fit. Explicit
membership cycles and relationships contradicting native group ancestry fail.
This does not reparent objects or infer semantic containers from visual overlap.

The check runs after all native transforms/order and endpoint validation during
build, preview and guarded adoption. Source member removal prunes membership;
complete native deletion prunes exact owned names. Mapped copies inherit their
source member rules. Related membership enters structural source fingerprints,
so edits to those rules after a baseline require conflict review. Rules retain
template identity pins even without transform overrides. Layout reset preserves
rules, and uncontain is a separate actor/reason transaction with predecessors.

Inspection returns clearances and overlap envelopes for direct declared siblings.
Overlap warnings are diagnostic and conservative for rotated allocations; they
are not exact ink collisions. Container headings need authored top padding, and
stroke/arrowhead extent is still separate visual acceptance work.

Focused tests cover source/native escape refusal before writes, scaled/rotated
nested groups and descendants, cycles, missing references, padding, overlaps,
copy inheritance/idempotence, source/native member removal, container deletion,
reset preservation and canonical receipt retention. Qualification uses
`scripts/qualify-geometry-roundtrip.py --structures --containment` with private
Documents output; it additionally checks exact native rebuild, numbered snapshots
and share/extraction. PowerPoint rendering remains a separate acceptance step.

Third-slice private evidence is under
`~/Documents/pptxgengo-qualification/geometry-containment-20261008/`:

- `qualification-03` uses fixed CLI SHA-256
  `dbc6146d85bd2ab7e7599970be3be5ed738489e3c557e01dde05dbe46f24818f`.
  Authored and native moves escaping Services while staying inside the slide are
  refused before writes. The combined inherited-tag specimen has three native
  geometry and two structural proposals, zero manual items, all 41 transforms
  and paint orders equal after rebuild, no declared sibling overlap, inherited
  monitoring membership, deleted-member pruning and version/share/extraction.
- The overlap diagnostic exposed a 2 pt sibling collision in the earlier copy
  fixture. The corrected specimen moves the wider copy 168 pt horizontally,
  preserving a 6 pt sibling gap and 12 pt container right padding.
- A fresh Documents-based `render-doctor` probe reaches PowerPoint 16.113.4 and
  passes automation/staging checks, but PDF export again returns `-9074` with
  file access marked unknown. Its cause is unconfirmed; no native visual signoff
  is claimed for this specimen. Controlled XML reconciliation qualification
  passed independently. Full GUI/Save As and Windows checks remain pending.

## Attached elbow contract

Source `attached-connector` supports straight and horizontal/vertical first
orthogonal routes. Elbows serialize as one attached native `p:cxnSp` using
`bentConnector3` and literal `adj1`; vertical first uses a 90-degree transform
with endpoint-relative extents/flips. The low-level connector API copies route
metadata and continues to validate explicit unique rectangle/site attachments.
Source movement recalculates the path. Inspection resolves final path points
through native parent coordinates and orientation, including adopted bends.

Supported native preset/guide metadata enters the typed native geometry, source
basis and three-way comparison. Exact validated route spans are replaced with
transforms, preserving attachments and other native payloads. Geometry-only
comparison ignores preset XML only after complete supported metadata validation;
unknown presets, extra guides, formulas, custom paths and attachment edits remain
manual. Default empty-guide bent presets normalize to `adj1=50000`. Literal
signed guides are bounded; their actual path allocation enters frame and
container checks, including bends outside the endpoint rectangle. Ink/stroke
extent and automatic obstacle routing remain separate work.

Focused tests cover both orientations, forward/reverse and all direction
quadrants, aligned endpoints, strict route inputs, route snapshot isolation,
35%-to-65% native bend adoption, exact rebuild, idempotence and refused frame/
container escapes before writes. Private catalog CLI qualification is performed
with `scripts/qualify-routed-connectors.py`; it includes preview safety, source
node movement, calculated endpoints, controlled bend edits, snapshots and ZIP
sharing. It does not establish native PowerPoint rendering or GUI/Save As fidelity.

The preset definition was checked against the
[LibreOffice source copy of the DrawingML preset definitions](https://raw.githubusercontent.com/LibreOffice/core/master/oox/source/drawingml/customshapes/presetShapeDefinitions.xml).
Attachment semantics follow the
[Microsoft Open XML connection-shape reference](https://learn.microsoft.com/en-us/dotnet/api/documentformat.openxml.drawing.connectionshape?view=openxml-3.0.1).

Fourth-slice private evidence is under
`~/Documents/pptxgengo-qualification/geometry-elbows-20261008/qualification-01/`.
Fixed CLI SHA-256 is
`3b796c48104050f6b5a31e3cfb9fa61f58af552883ae9c3487f2ff38000c1cab`.
Both catalog-derived projects pass preview/no-write, source movement with attached
endpoints, single-proposal bend adoption, changed calculated path, exact guide
rebuild, repeat adoption and version/share/extraction. The horizontal specimen
also refuses inner-container and frame escape before source writes. Native
rendering/GUI Save As was not run; the earlier `-9074` probe remains unresolved.
