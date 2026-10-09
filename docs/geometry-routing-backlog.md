# Diagram routing backlog and first diagnostics

Requested on 2026-10-08. The promoted release remains v4.2.1. This document
separates the implemented diagnostic slice from future routing capabilities.
See [geometry editing](geometry-editing.md) for source ownership, reconciliation,
and existing qualification boundaries.

## Intent

An architecture, team, governance, or process template demonstrates a visual
language and a plausible arrangement. Its stock nodes and connectors are not a
fixed topology. Project-owned diagrams need different counts, relationships,
positions, and routes while retaining typography, framing, and readable visual
hierarchy. Layer diagrams need count-aware spacing and color treatment; composed
diagrams need semantic nodes and relationships independent of example geometry.

Routing must serve that customization. An agent should be able to explain why a
connector is placed, inspect its attachments and final path, identify a collision,
and deliberately choose a correction. Automatic layout must not change the
meaning of a relationship, silently remove a manual bend, shrink typography, or
modify the installed catalog to fit one project.

## Implemented diagnostic slice

`project diagram inspect` and diagram mutation previews/applications return an
optional `routing` object when the slide has supported attached connectors.
It contains:

- Actual final native route centerlines after group transforms and adopted bends.
- A sorted inventory of supported block and nonempty native text allocations.
- `obstacle_intersection` and `obstacle_clearance` observations with the nearest
  segment and distance in points. The current 6 pt minimum is an advisory
  diagnostic threshold, not an authored policy or proof of visual legibility.
- `route_crossing` and `route_overlap` observations with route IDs, segment
  indices, and a world-coordinate intersection point. A shared attachment-site
  endpoint is an intentional join; overlapping segments still produce warnings.

Diagnostics are warnings. They neither reject a valid source transaction nor
change a route. A preview writes no authored files. Moving source objects or
adopting supported native transforms updates the next inspection.

Endpoint blocks and their text are excluded from their own connection's obstacle
checks. Explicitly declared enclosing blocks are excluded when they contain an
endpoint or the connector. Separate container-heading text remains an obstacle.
Enclosing geometry is not inferred from visual overlap. An undeclared surrounding
block can therefore produce a conservative warning until its ownership is made
explicit or the operator reviews it as intentional.

### Diagnostic limits

- The path is a centerline. Obstacle bounds are axis-aligned world allocation
  envelopes; rotation can produce conservative warnings.
- Stroke thickness, arrowheads, font ink, shadows, and exact rotated shape
  silhouettes are not measured. Text allocations are not exact glyph bounds.
- The first obstacle set covers the currently supported rectangular port blocks
  and native text. Images, arbitrary polygons, decorative shapes, legacy stock
  path connectors, and unowned Office additions require further capability work.
- This does not introduce connector labels, obstacle avoidance, clearance policy,
  or a route-planning command. An unchanged diagram with zero warnings is not a
  complete visual qualification.
- Existing frame-zone and explicit padded-container checks remain hard fit
  constraints. Routing warnings supplement those checks; a warning cannot excuse
  an off-frame or escaped-container route.

## Prioritized work

### 1. Explicit obstacle and clearance policy

Make obstacle ownership and reserved label/header regions authored data. Allow
agents to inspect, preview, and set minimum centerline clearance, with a future
ink-aware margin kept separate. Expose excluded obstacles and their reasons.

Acceptance:

- An image, label, and supported node can each be declared as an obstacle.
- Padding and transformed geometry are respected without counting a containing
  frame as a solid obstacle.
- Intentional overlaps have explicit reviewable exclusions; unresolved objects
  are visible rather than silently omitted.
- Allocation and ink measurements are identified separately in JSON and help.

### 2. Obstacle-aware route proposals

Propose an orthogonal route between stable node/site endpoints inside the actual
frame zone and applicable padded containers. Preserve existing bends until the
operator requests rerouting. First support a bounded simple elbow search; more
complex routes require an explicit persistent bend model rather than pretending
every path is `bentConnector3`.

Acceptance:

- Move a blocker into a straight route; preview at least one valid alternative
  with clearance measurements, then apply one explicit actor/reason transaction.
- Correctly handle reversed endpoints, all direction quadrants, aligned endpoints,
  rotated/scaled parent groups, and two distinct frame zones.
- Return an actionable no-route result when the corridor is blocked; preserve
  source, manual bends, font sizes, and topology without partial writes.
- Repeat the same proposal with identical inputs and obtain the same candidates.

### 3. Native route model and human round trip

Extend typed route representation before accepting additional native presets,
multiple bends, curved routes, connection sites, or endpoint attachment edits.
Define source mappings for every supported control. Keep unsupported paths in
manual review with retained original bytes.

Acceptance:

- Actual macOS and Windows PowerPoint edits move a node, drag a bend handle, and
  save a separate copy; reconciliation preserves attachments and intentional
  bends through rebuild, numbered snapshot, ZIP extraction, and another machine.
- Simultaneous YAML/native route edits are explicit conflicts; native formatting
  or attachment changes do not silently become inferred relationships.
- Unknown presets/formulas/custom geometry stay unresolved until represented.

### 4. Relationship labels and crossing readability

Define a stable label identity, text binding, route anchor/offset, and measurable
label allocation. Offer crossing diagnostics, clear labeling, and intentional
intersection/junction semantics. No crossing implies a junction automatically.

Acceptance:

- Add, remove, and reposition a label without changing endpoint meaning.
- Moving an endpoint recomputes an automatic label placement; an explicit manual
  label override survives until deliberately reset.
- Text wrapping and allocation are measured; crossings/overlapping labels and
  reserved frame regions are reported with readable object IDs.
- Human text/geometry edits reconcile independently, with ambiguity retained.

### 5. Additional object sites and family runbooks

Support ports on more typed nodes and containers, and link routing guidance to
family-specific semantic composition. Team and governance diagrams need reporting,
collaboration, and sponsorship relationships. Process diagrams need forks,
joins, starting points, and current-position markers. Architecture layers need
count-aware composition and deliberate cross-layer dependencies. These are
family design choices, not generic inferred relationships from shape positions.

Acceptance:

- Representative examples change count and topology rather than merely replacing
  stock text, using only documented CLI/YAML operations.
- The runbook describes permitted customization, relationship notation, layout
  tradeoffs, fit failure resolution, and human reconciliation boundaries.
- Agents can measure a proposed arrangement and explain its choices without
  editing catalog definitions or treating the example topology as mandatory.

## Verification and CI placement

Focused unit tests cover segment intersection/distance, reversed overlaps,
degenerate segments, bend-intersection deduplication, shared endpoint exclusions,
block/text clearances, logical containing-block exclusions, catalog preview
source safety, and diagnostic recalculation after movement. Current-route native
geometry tests remain responsible for endpoint/guide persistence and fit refusal.

Every routing slice needs targeted tests and a private receipt-backed catalog
specimen. Desktop rendering/Save As, broad catalog stress cases, and large routing
search qualification belong to nightly or on-demand GitLab jobs. Do not add race,
desktop, or catalog-wide jobs to normal branch or main pipelines for diagnostics.

The first slice does not establish new PowerPoint GUI or Windows qualification.
