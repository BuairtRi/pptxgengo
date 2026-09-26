# Design review: composable components

2026-09-26. User feedback is preserved verbatim in
[reference-preferences.json](../library/reference-preferences.json). Its shortlist
hash and every reference/component pair were matched during import. The original
export remains under `samples/`. This document distinguishes the user's decisions
from proposed implementation details.

## Recorded decisions

| Reference | Selected preference | Effect on the work |
|---|---|---|
| M1 | Alternate | The note calls this a good primary metric card. Keep it as the first metric implementation while retaining the actual selected preference. |
| M2 | Alternate | Retain the contextual metric option. |
| M3 | Alternate | Preserve large-number emphasis for case studies/headline statistics. The divider belongs between components, owned by the enclosing composition. |
| N1 | Alternate | Retain the existing source edit fixture and alternate panel design. |
| N2 | Preferred | Make the explanation row the primary numbered-card implementation priority. |
| N3 | Alternate | Retain the evidence/measurement option. |
| N4 | Avoid | Exclude this reference from default recommendations; keep it inspectable. |
| N5 | Alternate | Preserve the segmented, multi-band layout. Rework typography, size and density; the user finds the current type cramped and too large for sufficient content. |
| P1 | Preferred, with architecture correction | Extract a reusable role tile and compose a pod from a variable list of role tiles. Support one or more columns and content-dependent height. |
| P2 | Unreviewed | Preserve its source evidence without inferring user approval. |

Preferences apply to these reviewed references. They do not automatically approve
every member of a family, every color variant, factual reuse, or technical quality.

## Product model required by the pod feedback

1. **Role tile:** an editable rectangle with a role name, background color and
   foreground color. Reuse the tile in pods and in broader team structures.
2. **Pod container:** a titled surface that contains a variable list of role
   tiles. Its layout can use one or several columns. Neither the number of roles
   nor labels such as lead/engineer are fixed in the reusable API.
3. **Enclosing team composition:** positions pods and individual roles, owns
   reporting relationships, shared phase bands and legends.

The existing P1/P2 contracts remain useful source reconstruction fixtures. Their
two/three-slot limits describe those frozen scenes, not the intended pod model.

## Proposed first implementation

These are engineering proposals, not additional user-approved design variants.

```yaml
kind: pod
id: platform-delivery
title: Platform delivery
bounds: {x: 0, y: 0, width: 300, max_height: 260, unit: pt}
layout:
  columns: auto                 # or an explicit positive integer
  order: row_major              # retain declared role order
  gap: 8
  padding: 10
roles:
  - id: platform-lead
    label: Platform lead
    style: {background: staffing.wm_full_time, foreground: auto}
  - id: data-engineer
    label: Data engineer
    style: {background: staffing.wm_full_time, foreground: auto}
```

The numerical bounds, gaps and padding above illustrate the proposed interface.
Final defaults must be measured from the source and checked in native renders.

### Ownership and layout

- Keep role identity separate from its label and native PowerPoint shape ID.
  Multiple people or allocations can have the same role label.
- Resolve effective font, line spacing, insets, surface and text colors before
  measuring content. Retain semantic staffing tokens across layout changes.
- Lay out role tiles inside the pod's content zone below the title. For each
  candidate column count, compute tile width and measure labels at that width.
  Row height is the largest required tile height in the row; total pod height
  includes title, padding and gaps.
- In auto mode, evaluate a bounded set of column counts that fit the available
  width, then choose a fitting arrangement deterministically. Expose the chosen
  column count and measured dimensions. Preserve input order.
- An explicit column count is a constraint. If the result exceeds the allowed
  width/height or supported typography range, return a fit failure with the
  affected labels and alternatives. Do not silently remove roles or change labels.
- Foreground `auto` selects a readable brand foreground against the resolved
  background. Explicit foreground overrides require a contrast check. Surface
  changes must preserve the declared staffing meaning and corresponding legend.
- Return anchor positions for tiles and pod bounds. Reporting connectors and
  shared phase decorations belong to the enclosing composition and are routed
  after geometry is established.
- A metric strip similarly owns its separators, shared surface, heading and
  source note. An individual metric tile owns its value and label.

### Work order

1. Implement the role tile with measured text fit, style resolution and explicit
   contrast validation. Extract source evidence from UHG43.
2. Implement the pod container and variable-list layout using the same role tile.
   Keep staffing semantics independent of tile placement and role cardinality.
3. Exercise native rendering on the proposed fixture matrix below; collect
   layout, text, contrast and placement failures in `planning/QA_ERRORS.md`.
4. Build a new team slide with pods of different sizes, then validate connectors,
   phase bands and legends. This is the first composition proof beyond editing a
   frozen source scene.
5. Extend numbered-card work to preferred N2. Retain M1 as the first metric
   fixture; use M3 to prove independently owned separators. Treat N5 typography
   as a separate design experiment after the fit pipeline exists.

### Proposed acceptance fixtures

- Role counts 1, 2, 3, 5 and 8; no inference that these are the only supported counts.
- One-column and explicit two-column layouts; auto mode under wide and narrow
  bounds; uneven final rows and repeated role names with distinct IDs.
- Short, wrapped and excessive role labels; a multiline pod title; content that
  cannot fit the available slide region.
- Navy, blue and magenta semantic role surfaces, automatic foreground selection,
  a valid explicit override, and an override that fails contrast validation.
- Native editable shapes and text; stable semantic IDs in the generated metadata;
  outer bounds, inner content zones and inter-component collision checks.
- Source-aligned two/three-role fixtures for continuity, plus new multi-column
  compositions to demonstrate the dynamic capability.

The user review approves a design direction. Implementation and rendering must
still establish which concrete compositions are technically ready.
