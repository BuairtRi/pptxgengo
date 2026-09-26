# Team composition

The [team spec](team.json) extends the measured role/pod API with standalone roles,
phase backgrounds, an automatic staffing legend and orthogonal reporting lines.
The two slides are illustrative engineering examples, with editable native objects.
They are not copies of UHG43 and do not establish a staffing commitment.

## Composition ownership

| Object | Owns |
|---|---|
| Role tile | Label, staffing meaning, font, fill and foreground, insets, measured fit |
| Pod | Variable roles, columns, row heights, title, padding, local anchors |
| Phase | Explicit member IDs, shared background and bottom caption, padded containment |
| Team slide | Standalone roles, pod positions, staffing legend and reporting relationships |

A phase cannot cover an undeclared role/pod, intersect another phase, or cover the
legend/title. Its members must remain inside the padded content zone above the
caption. The phase fill `#CCDAFF` and bottom-caption pattern come from UHG43;
PowerPoint reported RGB `(204,218,255)` for its `Rectangle 103`. Pod sizes, content,
fonts and arrangement in this example are new.

## JSON additions

Use the existing `pptxgengo.compose-spec.v1` schema with optional slide fields:

- `roles[]`: standalone role fields `id`, `label`, `background`, `foreground`,
  `bounds` (x/y/width/height), `font_face`, `font_size_pt`, `bold`,
  `horizontal_inset_pt`, `vertical_inset_pt`. Fixed bounds must fit measured text.
- `phases[]`: `id`, `title`, `members[]` (root role/pod IDs), `bounds`, `padding_pt`,
  `label_height_pt`, `font_face`, `font_size_pt`, `bold`, `surface`, `foreground`.
  Caption geometry is relative to the bottom of the phase bounds.
- `legend`: `bounds`, `font_face`, `font_size_pt`, `bold`, `foreground`,
  `swatch_size_pt`, `gap_pt`, `item_gap_pt`. Entries are derived from used staffing
  tokens, in canonical order. Each label keeps the same width used in its probe.
- `connections[]`: `id`, `from`, `to`, `relationship: "reporting"`, `from_anchor`,
  `to_anchor`, `color`, `width_pt`, `clearance_pt`. Anchor sides are `top`, `bottom`,
  `left`, `right`; endpoints are standalone role IDs or complete pod IDs.

Team roles require semantic staffing tokens. Navy means WM full time, blue means
WM part time and magenta means client part time. Using raw hex colors to stand in
for these meanings is rejected. The legend uses `CLIENT PART TIME`; it does not
assume a particular client name. Legacy pods-only specs still support their
previous colors without requiring a team legend.

## Routing and its limits

The planner reserves endpoint stubs and expands component/title/legend/phase-caption
obstacles by the requested clearance plus half the stroke width. It first tries
straight and centered elbow paths. If those are blocked, a bounded visibility-grid
search minimizes path length with a bend penalty. Routes must stay within the
slide. Unknown endpoints, duplicate IDs, reporting cycles, blocked approach stubs,
and routes without a supported clear path are errors.

Common-source relationships may share a continuous trunk, then diverge. They
cannot cross or rejoin after divergence. Collinear strokes are merged for output,
with contributing relationship IDs retained in the manifest. Other relationships
cannot overlap or cross; shared-endpoint contact must be a point. Unrelated
nonintersecting strokes also have a minimum ink separation check. Route order is
the declared connection order. A route can fail even when another ordering or a
more elaborate layout could succeed; the CLI does not move roles automatically.

`clearance_pt` controls distance from content obstacles. It is not a general
uniform separation guarantee between all branches of a shared tree. Native and
visual QA remain required. Line contrast is checked against white and any crossed
phase surface using a 3:1 engineering threshold.

The output uses editable orthogonal line segments. They are **not PowerPoint-glued
connectors**: manually moving a role does not automatically reroute them. Rebuild
from the spec after geometry changes. Arrowheads, peer/collaboration relationships,
nested-role routing and hand-drawn arrow artwork need separate implementations.

## Native verification

Run the workflow in [README.md](README.md), substituting `team.json`. The v4 native
adapter records character-bound unions, whole-range bounds for diagnosis, uniform
font/color properties, fill, insets, and line color/weight/transparency/arrowheads.
It accepts a zero width or height for a straight line, while rejecting a point
with both dimensions zero. Line visibility is derived from PowerPoint's documented
line-style enum (`line style unset` means no line in the observed fixtures).

Character-bound unions avoid a native whole-range bounding-box anomaly found on
“Quality engineer.” These are native character advance bounds, not a raster-ink
measurement. All final slide renders still require direct visual inspection.

See [team proof](team-proof.json) for the exact artifacts, routing strategies,
negative cases and scope of the native checks.
