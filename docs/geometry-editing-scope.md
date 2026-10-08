# Geometry editing and architecture reconciliation: next slice

Requested by Ri on 2026-10-08. **Target scope; not all implemented or released.**

Main contains the first geometry slice: line-safe catalog
detach, typed diagram inspect/patch/connect/arrange with previews, straight
attached block connectors, and receipt-backed native transform/paint-order
adoption. [Operator guidance](../skills/west-monroe-presentations/references/architecture-geometry.md)
describes the available surface and exclusions. The promoted v4.2.1 release does
not include these changes. A second slice adds reviewed whole local leaf
deletions and explicitly mapped block copies with independent text bindings,
including copied identity tags. Arbitrary native imports, partial deletions and reparenting,
custom paths/other connector presets, obstacle routing, full ink bounds and desktop
qualification remain open. A third slice adds explicit logical containment with
container-axis padding, descendant checks, sibling allocation overlap warnings,
and rule inheritance/pruning during structural adoption. Attached horizontal/vertical
elbows now expose calculated paths and adopt supported native bend guides. Source group coordinate-space changes require a
fresh baseline; transforms do not claim native typography reflow qualification.
The v4.2.1 baseline already has explicit local-template placements, resolved frame
zones, native object identity tags, low-level connector support and reviewed
plain-text adoption. It does not expose a complete diagram-editing CLI or adopt
native geometry. See [text reconciliation](text-reconciliation.md) for the
current boundary and [project structure](../skills/west-monroe-presentations/references/project-structure.md)
for the active source/working-copy/history layout.

## Required outcome

An agent customizes an architecture template's content and geometry without
editing shared library definitions. Nodes can be added, removed, moved and
resized; connectors express real relationships and stay attached when nodes
move. The complete diagram remains within its selected framing device. A human
then changes geometry in PowerPoint, and reviewed changes persist in YAML and
survive rebuilding, numbered snapshots, ZIP sharing and another machine.

The geometry model must serve ordinary slide objects as well as architecture
slides. Architecture is the first complete authoring and desktop demo; geometry
reconciliation is not permanently restricted to that template family.

## 1. Source-owned geometry and identities

- Represent diagram nodes, containers, groups, ports and edges by stable IDs.
  Their count is variable; the stock template's item count is a starting layout,
  not a limit on a project-owned diagram.
- Keep semantic relationships separate from positions. Edges refer to node IDs
  and named connection sites, not just numeric endpoints or object names.
- Persist instance-specific geometry in slide YAML or an explicitly linked
  project-owned local definition under slides/templates. Use a local derivative
  when customization changes topology. Preserve shared template/source pins;
  never alter the installed catalog to repair one slide.
- Declare coordinate space and units, parent/group transforms, bounding boxes,
  rotations, flips and paint order. Translate native coordinates into the same
  source space before comparison; applying a group move must not double-apply
  child offsets. Geometry overrides must be validated after layout and included
  in source hashes, object maps, receipts and exported source.
- Bind every editable native object to its source-owned geometry property and
  logical owner. Composite subshapes need explicit ownership rather than guessed
  correspondence from copy, position or physical numeric IDs.

## 2. Agent-facing CLI operations

Design one diagram command family with machine-readable results. The concrete
names/flags are to be settled with the source contract before implementation.
It must support:

| Operation | Required assistance |
| --- | --- |
| Inspect | Nodes, edges, group hierarchy, writable properties, stable IDs, bounds and frame allocation. |
| Add/remove | Typed nodes and containers with variable counts; explicit incident-edge treatment when deleting a node. |
| Move/resize | Absolute placement and deltas, with affected edges and measured text allocation. |
| Arrange | Align, distribute, snap and bounded layout proposals while retaining the operator's intended topology. |
| Measure | Node/text bounds, minimum legible allocation, overlaps, clearance and off-frame content. |
| Connect/route | Attached endpoints, arrowheads, labels, straight/elbow routes and explicit bends. |
| Preview/apply | A deterministic patch and diagnostics before one guarded source transaction. |

Resolve the allowed diagram rectangle from the selected frame and any inner
container padding. Header, footer, source text and rails are reserved regions;
split frames require their actual zones, not an enclosing rectangle that crosses
reserved space. Account for rotated objects, stroke/arrowhead extent, labels and
connector bends when checking bounds. Report overlap/crossing separately from
out-of-bounds errors because some intentional overlaps express containment.

Endpoint calculation uses actual object/port geometry. Offer obstacle-aware
routing and clearance diagnostics; do not route across labels or outside the
frame silently. Preserve authored/manual bend points unless rerouting is
requested. If no valid route or fit exists, retain the proposal and identify the
constraint; do not silently shrink typography or change the architecture.

## 3. Native geometry reconciliation

Extend the existing receipt-pinned propose/adopt workflow. Compare three states:
original generated baseline, current authored YAML and edited PowerPoint.

- Extract typed geometry independently of text and style. PowerPoint Save As
  serialization differences must not invent semantic geometry changes.
- Match tagged object identity and explicitly recorded source bindings.
  Resolve transforms through parent groups and map the result into the source
  coordinate space. Record the actual connector endpoint attachment/site and
  route, including deliberate bends, instead of only its envelope.
- Classify unchanged, YAML-only, native-only, matching and conflicting changes
  for each supported property or atomic geometry unit. Preserve current YAML
  changes when the native copy still matches the baseline.
- A reviewed native change writes persistent source geometry, retains exact
  predecessor bytes, invalidates affected approvals and rebuilds a new baseline.
  The original build and edited file remain immutable evidence.
- Added/deleted/duplicated/reparented objects are explicit structural changes.
  Imported nodes receive new stable IDs. Duplicated inherited tags are ambiguous
  until deliberately mapped; missing objects do not silently imply deletion.
  Reconcile incident edges and source-owned content with topology changes.
- Changes that violate bounds, fit or supported ownership stay visible in a
  retained packet with an actionable resolution. Never claim full synchronization
  if any edit remains unresolved or unrepresentable. Do not clip or discard the
  human's original edited geometry to make an adoption pass.
- Keep text and geometry proposals independent where their source bindings
  permit it; moving a text box must not hide its legitimate copy edit. Preserve
  the existing packet hash/replay, source-drift and guarded mutation protections.

Begin with translation/resize and attached straight/elbow connectors for uniquely
mapped source objects, then close group transforms, rotation/flips and structural
imports. The full requested feature requires completing those cases or clearly
reporting the remaining exclusions; a translation-only demo is not completion.

## 4. Project repair and operator skill

The skill includes one canonical project tree, ownership rules and the current
layout repair workflow. Active slides and local definitions are normalized via
`project layout`; supporting context/notes/media repairs remain explicit.
Generated builds, numbered versions and predecessor receipts are not moved or
rewritten. The next diagram commands and reconciliation support must be added to
the skill only when implemented, including conflict handling, frame constraints
and persisted geometry location. CLI help, source schema and operator guidance
must agree.

## Completion demo and qualification

Use a receipt-backed architecture project with a frame/container, real named
nodes and attached connectors. In the CLI, add/remove elements, change arrangement
and demonstrate routing/fit diagnostics. In PowerPoint, move and resize a node,
move a group and reroute a connector. Reconcile, rebuild and compare the rendered
geometry with the edited copy. Demonstrate a simultaneous YAML/native conflict
without losing either edit. Show that adding/removing an element and changing
connections persists after rebuild and complete ZIP/share extraction.

The general geometry path also needs representative card, text, image and table
placement cases, grouped transforms and object identity ambiguities. Verify
actual Mac and Windows Save As behavior separately; preserve task files in an
owned PowerPoint-accessible Documents folder and keep exact baseline/runtime
pins. Desktop evidence and unresolved cases must be reported honestly.

Keep PR/main pipelines light. Heavy catalog-wide and native qualification belongs
in the established nightly/on-demand tiers; private GitLab remains the only CI
and artifact system. Do not add broad new gates to every push for this slice.
