# Shape and layout execution detail

Proposed implementation, integrated by the coordinating architect. The existing
structural inventory provides object trees and local transforms; the following
semantic contracts and geometry resolution remain to be implemented.

## Catalog boundaries

| Record | Identity and purpose |
| --- | --- |
| Source occurrence | Source hash + part + object locator; preserves every object and native group, including duplicates |
| Component | Stable catalog ID + version; curated shape or composition with slots, anchors, styles and allowed adaptations |
| Designed layout | Full-slide arrangement with zones, cardinality, grid/constraints and dependencies |
| Native layout/master | Original package part and inherited design resources; tracked separately from designed layouts |
| Polished slide | Finished content example with independent design preference and factual reuse status |

Native groups are candidate component boundaries. Useful components can also
span ungrouped objects. Record membership and rationale; never infer semantic
grouping solely from proximity. Preserve occurrence lineage when consolidating
families. Hash equality can establish exact asset duplication; a similar design
or alternate colorway requires an explicit variant relationship.

## Geometry and style contract

- Canonical geometry remains integer EMU. New arrangements can use logical
  columns, gutters, spacing steps, safe regions and optional snapping.
- Preserve group-local coordinates and compute slide-space transforms through
  nested offsets, extents/child extents, rotation and flips. Reject degenerate
  transforms with a specific diagnostic.
- Resolve typography and styling through explicit properties, placeholders,
  layouts, masters and themes. Retain provenance and unresolved values.
- Expose container bounds, visible artwork bounds, text insets, baseline and
  alignment anchors, connector attachment points, permitted resizing and layering.
- Express containment, equal spacing, alignment and minimum readable text size
  as constraints. Declare intentional overlaps such as highlights behind text.
- Store observed character/line counts separately from tested content capacity.
  Native measurement and rendering determine whether new content fits.

## First bounded batch

1. Freeze the source/item manifest schema and source allowlist.
2. Deduplicate across all 369 source slides using canonical dependency-aware
   fingerprints and structural comparison. Review uncertain candidates with native
   previews; retain reversible membership, variants and unresolved singletons.
3. Render/classify the first 10–15 canonical layouts/variants covering cards,
   columns, tables and diagrams. Process each confirmed design once; retain all
   source occurrences and their distinct content/fit review requirements.
4. Propose 25–40 component candidates with source locators and in-context previews.
   Review grouping, duplicate variants and design preference before promotion.
5. Bind one phase-detail and one pod/team layout. Exercise content lengths and
   item counts, then compose one new arrangement from approved components.
6. Preserve a handoff baseline, edit text in PowerPoint, re-import and rebuild.

Luna workers own independent classification, description and fixture manifests.
Sol/Terra own transform/style resolution, indexing, binding and composition code.
The architect integrates schemas, serializes native rendering, reviews evidence
and commits. Asset description work proceeds independently. See the full
[execution waves](../IMPLEMENTATION_PLAN.md#7-delegation-dependencies-and-execution-waves).

## Promotion evidence

Each reusable item needs baseline and adapted previews, named-slot mappings,
effective styles, dependency hashes, supported operations, explicit fit limits,
known failures and independent design/content/readiness states. Require both a
successful typical case and boundary cases with honest failure/fallback results.
Raw inventory completion is tracked separately from these promotion gates.


## Component expansion checkpoint

The small 21-example seed set has expanded to 244 retained source examples from
74 source patterns, simplified into 31 semantic retrieval families. Five reviewed
duplicate examples remain as aliases. Six proposed color profiles use explicit
presentation roles; source data/actor/status encodings remain separate. See the
[component report](../library/component-expansion-report.md) and
[style contract](../library/component-styling.md).

Next: develop editable contracts for metric panels, numbered cards and delivery
pods. Bind object/run roles and shared dependencies, resolve typography, exercise
boundary content and color variants, then promote only successful native renders.
Continue the stock-diagram backlog without counting whole layouts as components.
