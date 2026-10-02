# Completion work packages: approach, solution and team

## Audit boundary

This is a read-only implementation plan for **34 frozen templates**: approach 11,
solution 12 and team 11. The machine-readable companion is
[coverage-diagrams.json](coverage-diagrams.json). None of these 34 templates is
among the four currently bound templates. Every entry remains inventory-only and
has `qualified_envelope: false`. No renderer changes, builds, tests, decks or
native captures were made during this audit.

The JSON records each exact key, tier, whole-family source SHA, template pointer,
top-level node types, every nested source object's fields and array shape,
actual card anatomy/table cell types, source feature pointers, binding work and
source resolutions. **32 distinct top-level node types** appear in these families.
A table column with `type: bullets` or `type: tag` is a cell descriptor, not an
additional independent slide node. A `card` with `bio` or a title-only `body: []`
has different execution dependencies from an already implemented body card.

Frozen JSON and the frozen reference renderer are the authority. Live
`wm-design-system/docs/authoring-reference.md` and generated component docs were
read to clarify semantics and identify conflicts. Their statement that every
node maps 1:1 to native PowerPoint does not describe the current Go adapter's
actual coverage.

## Shared interfaces before parallel implementation

The integration owner extends the closed source compiler and template binding
registry. Each exact template continues to be bound to its source SHA and named
source pointers. Unknown fields and mismatched anatomy fail; fictional examples
are cleared before binding. Fixed templates first retain source counts/topology.
Variable repetition requires a declared keyed array and capacity envelope.

New component planners should accept a resolved rectangle, frame zone and surface,
return measured text/shape/child-group records, and draw only after successful
planning. Native object ownership must include intentional containment,
connector labels, overhangs and nested groups. Shared text measurement remains
the Go engine with the original IBM font names. The existing renderer and its
contracts must not gain unrestricted new fields just to accept all source JSON.

Parallel implementation owners create separate planner/compiler files. The
integration owner alone edits shared `Node` dispatch, source/IR structs, CLI
registry, package grouping and coverage status. Suitable planner boundaries are:

| Owner | Files/interfaces to add | Shared integration required |
|---|---|---|
| Foundation and compiler | Source-context geometry policy; frozen palette refs; exact-key source compiler/binder projections | Frame/Node dispatch, registry, source identity report |
| Diagram primitives | Blocks, chevrons, containers, layer rows, cylinders, icon nodes and source connectors | Shape/group reporting; registered asset resolution |
| Lists and data grids | Standalone/mixed lists, native tables and matrices | Text/cell records; native table serialization and overlays |
| Sequence and graphs | Stepper, phase columns, Gantt, swimlane, org tree and governance planners | Named graph/group ownership; rotated text/route serialization |
| People and media | Person tiles, portraits, bio cards, pullquotes and thumbnails | Pinned asset registry; focus/grayscale/crop writer behavior |

Some ownership can run concurrently after these interfaces are agreed. Table and
Gantt work must consume the same list/typography/palette contracts; three owners
should not independently invent cell fonts, organization colors or key semantics.

## Package1: source contexts, palette and typography

Derive appendix and two-line header allocations directly from source. The three
`phase-detail/rail-*` slides omit standard title/eyebrow and start main content at
y54; they require explicit absent-header mode. Other rail templates have a main
header and separate rail content. Preserve `on` surfaces; infer actual frame zone
where the source omitted an otherwise implied rail context.

Add source footer notes and source text as separate typed concepts, and an explicit
stamp annotation. `pathways/two-lanes` has notes only. Preserve one-mark limits and
typed highlighted phrases; existing plain replacement titles can deactivate
example marks only under a reported adapter interpretation. Small list/card text
is an explicit source dense context, never an overflow remedy.

Resolve categorical `series.N`, `deemph.N` and organization mappings from frozen
token data. `Source.Ink` currently resolves surface roles only; its typed dataviz
data retains KPI, not full categorical/deemphasis arrays. Generic adapt color maps
are not substitutes for this source resolver.

Record additional typography contexts independently of implementation progress:

- Semibold Sans body14/21 and small12/18 in person/role/node/pod/cylinder/container
  labels; bold Mono9/12 table, matrix, layer and member labels.
- Gantt lane titles12.5/18 weight600; period sublabels7.5 with source label leading12.
- Line chart axis/category Mono9pt and Sans12pt semibold end labels, with explicit
  overlay baselines/leading instead of browser SVG text baselines.
- Pullquote opening mark Sans60/30; rotated governance escalation label; swimlane
  pain marker Mono9/12 weight700.

The JSON lists contexts by actual source component. Existing calibrated anchors
do not qualify these overrides automatically. Candidate Go measurements can
support implementation and reference generation with provenance while additional
contexts receive a separate native development review. This does not make native
capture a permanent runtime requirement or justify renamed output fonts.

## Package2: headings, lists and card contexts

Add native `block`, `grouplabel`, `numhead` and source `chevron` planners. Chevrons
with widths300 extend into neighboring columns deliberately; their custom shape
geometry cannot be copied as an ordinary strict18pt outer box. Keep native text
separate and editable within owned groups.

Standalone lists must handle strings, `{lead,text}` segments and nested sublists,
with square markers and keyed ownership. Existing uniform card bullets and rich
text provide useful measurement code, but do not expose the full source list API.

Add specifically named source card contexts:

- 162pt narrow cards for rail detail; 144pt architecture sidebar.
- 154.8pt sanctioned five-up cards, including title-only numbered headers.
- Title-only cards with no body, `titleStyle: body`, source `bandNumber`, dense
  gap3 and mixed lead/text blocks where explicitly present.

Do not globally lower the198pt generic card minimum or replace missing body with
invented text. `context/build-sustain` has multiple rich body blocks at270pt,
contrary to the current generic multi-block414pt rule. A dedicated context or
versioned adapter resolution is necessary.

This package unlocks most of `phases/activities-outcomes`, `phases/summary-bands`,
`phase-gate/evidence-matrix`, `workstreams/five-with-risks`, `context/three-zones`,
`context/build-sustain`, `patterns/three-rows` and `readiness/six-criteria`.
Additional containers/connectors/source annotations still apply by exact entry.

## Package3: containers, connectors and architecture

Implement distinct `container` styles region/boundary/layer/frame/external with
labels, top-right label positions, source border/fill and native children. Add
`layerrow`, cylinder with text/subtitle and icon-bearing `node`. The writer has
native `can` and custom curved geometry primitives; reuse does not establish
that source cylinder ratios/text bounds are already implemented.

Connector API needs explicit point routes, horizontal/vertical elbows, start/end/
both/none open arrowheads, solid/dashed/dotted strokes and labels attached to the
longest segment. Source diagrams often deliberately overlap parent containers;
native ownership distinguishes that from a routed line crossing an unrelated node.
`internal/compose/routes.go` supplies obstacle/anchor precedent, but its relationship
schema and automatic layout are different from the source connector contract.

Templates: all five `architecture/*`, `context/three-zones`, the graph portions of
`team/org-roles` and `team/pods-pairs`, and container portions of readiness/phase
detail. Source `architecture/layers` includes a144x264 sidebar, violating current
generic card width and outer-height lattice. Its414x90 title-band card may also
fail conservative measured header/body occupancy. Declare an exact adapter
resolution; rounding, clipping or silently increasing card height changes source.

## Package4: native tables and matrices

Initial table cell vocabulary required here is plain text, keyed bullets and
tags. Implement source light/dark headers, dense28pt header, fixed row heights,
semibold row headers and spanning column groups. Native table cells must remain
editable; unsupported adornments can be associated editable overlays with measured
cell ownership. `pptx.AddTable` and generic measured containers are backend
capabilities, not ready WMDS table contracts. Their existing pagination/font
estimates must not bypass the Go measured source envelope.

Matrices need labels above or left, optional column labels, row-local equal cell
widths, varying counts between rows, explicit row/cell surfaces, and a single
Magenta outlined row. `architecture/layer-map` has4/4/5/4/5/3 cells, gap3 and
rowGap9. `cutover/wave-matrix` has left labels216pt and3x4 state cells36pt high
with18pt gutters. They are different presets under one top-level type.

Exact fixed table bounds already expose useful limits:

| Template | Source layout arithmetic |
|---|---|
| `phase-detail/rail-table` | y54 +28pt header +4x96pt rows =466; compact body468 |
| `phase-detail/rail-matrix` | y54 +28 +5x60 =382; closing bands start396 |
| `team/org-roles` | y144 +28 +6x48 =460; compact body468 |
| `roles/by-phase` | y108 +24pt group header +28 +7x42 =454 |

These calculations are source geometry evidence, not proof of native cell fit.
Every cell must be measured against its own padding and fixed row height.

## Package5: sequences, schedules and flow graphs

Implement source `stepper` states, `phasehead`, four-column `phases`, then Gantt
and swimlane graphs. Stable keys identify phases/lanes/items/steps; supplied
labels do not identify them. Native composite groups should expose meaningful
constituent parts while preserving semantic regeneration.

Gantt requires fractional period positions, measured label packing into tracks,
phase/gate bands, group/lane sidebar, events, soft hatch starts/ends and legend.
`plan/gantt` uses8 periods with sublabels and5 lanes. The frozen JS `tw()` uses
character-count width approximations. A read-only evaluation of that exact
arithmetic yields track counts2/1/1/2/2, body bottom498 and legend top512, beyond
compact body468. Go should measure labels and reject a failed envelope or apply
an explicit approved adapter resolution; it must not copy the approximation or
stretch time bars silently.

`roadmap/staggered-phases` uses13 periods including Later and fractional intervals.
The generic adaptive roadmap caps periods at12 and uses inclusive period IDs;
it cannot directly implement this source. Source arithmetic for its fixture gives
body390 and legend404. This also needs native review after implementation.

Swimlane uses4x6 lane/column coordinates,7 keyed steps and8 links, including
start/end, a decision, Yes/No/Exception branches and pain flags. Source docs claim
lines never cross a step, but renderer heuristic routes do not prove that.
Measured obstacle checking and full badge/note bounds are needed. Generic process
panels are a different layout and do not preserve this graph.

## Package6: people, portraits and team composites

Implement source person/role/pod/legend anatomy, then org tree and governance
stack. Existing generic team helpers offer keyed pods/roles, ports and legends
as precedent, but use different fonts, geometry, chrome and combined role fields.

Specific source conflicts must become explicit versioned decisions:

- Org-chart metadata says198pt tiles and36pt level spacing; frozen renderer uses
 162x54 tiles with54pt between levels. Preserve documented renderer geometry for
  the exact initial template, with a bounded tree ending before legend414.
- `team/pods` has3 roles per pod, ending300 before the client band318. Increasing
  role count changes that envelope. `team/pods-pairs` has2 roles, ending456;
  a third would end498 and cross the footer.
- Governance metadata allows3–5 tiers, but this fixed3-tier template uses78pt
  tier height/12pt gaps and a legend414. Do not advertise5-tier capacity from
  that component metadata. Measure wrapping member chips and rotated escalation.

Add photo-or-initials person tiles and `card.bio` with54pt avatar, separate
name/title/role fields, wrapped tags and proof bullets. At198pt leadership width
with12pt padding, name area is108pt. The current top-level `card` implementation
does not support bio content, and source sample text is not a general fit proof.

Portrait square/imageframe require pinned image bytes, explicit focus/crop and
consistent grayscale policy. Current WMDS runtime bundle has logos/fonts only,
so source headshots and icons need a separately versioned SHA asset registry.
No substitute photo or guessed icon may satisfy an unresolved source asset.
Pullquotes need measured mark/quote/attribution; thumbnails need page placeholders
or pinned assets plus stack overhang and caption bounds.

Templates: all11 team entries, plus phase-detail thumbnails. Full-page bios reuse
existing bounded textblocks, but their portraits, bullets, group labels and
pullquotes remain separate dependencies.

## Package7: native line chart and final template bindings

`phase-detail/team-handoff` is the only chart in these34 templates: line series
with fixed5 categories, dashed second series, `yMin:0` and frozen palette refs.
Use editable chart/workbook data and source-specific measured labels/axis styles.
`pptx.AddChart` provides a writer backend; the SVG reference chart does not meet
the native chart requirement by itself.

Bind templates by dependency batches, not by trusting a top-level node name:

1. Foundation contexts plus heading/list/card extensions: source text/band/five-up
   compositions with closed fixed counts.
2. Container/connector architecture and simple person/pod templates.
3. Native table/matrix and responsibility/phase-detail templates.
4. Source phase/stepper, Gantt, swimlane, org tree and governance composites.
5. Portrait/bio/pullquote/thumbnail variants and native chart handoff template.

Compiler work can run with planner work behind declared interfaces. Review decks
can be generated using candidate Go measurements while exact new font contexts
are qualified separately. Completion for a template requires its full source
feature inventory to be consumed, all declared replacement content preserved,
native editable ownership, measured failure diagnostics and a review receipt
for the exact reference artifact. Broader variable-content envelopes remain a
separate status; this planning audit confers none.
