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
apply a move. Rebuild and review a new baseline first. Added/deleted/duplicated
or reparented native objects, connector route/attachment edits and unsupported
formatting remain visible; they do not become guessed imports or deletions.
Legacy text proposals now also report paint-order changes.

## Bounded evidence

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
general card/image/table acceptance, structural adoption, bent arrows and the
other target-scope items remain pending. No release qualification or complete
geometry synchronization is claimed.

Each persisted transform also pins its authored geometry basis with
`source_geometry_sha256`. Changes to the underlying node/child coordinates are
rejected instead of applying an old transform to a newly generated coordinate
space. Use an explicit layout reset and fresh baseline to review that change.
The geometry proposal renderer verifies the actual current executable, fonts and
bundle against the project lock before computing current YAML geometry.
