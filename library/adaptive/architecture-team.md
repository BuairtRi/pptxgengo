# Architecture and team adapters

These semantic adapters create new editable compositions through the measured
`compose` planner. They do not rewrite source OOXML, and their fixture examples
are synthetic. In a future release with the `adapt` route, set `LIBRARY` to the
`library` path reported by `pptxgengo paths`, then compile a single example into
a new directory:

```sh
pptxgengo adapt compile --spec "$LIBRARY/adaptive/architecture-standard.json" --out /path/to/new-architecture-compilation
```

The frozen 0.1.0-local.3 release does not provide `adapt`. From this source
checkout, the development equivalent is `go run ./cmd/pptxadapt compile` with
the same flags and a repository-relative fixture path.

Use `team-standard.json` for a team example. The other single-slide fixtures
exercise fewer or more components, longer copy, and style choices. Consult
`checkpoint.json` for exact native and visual evidence; compilation alone is
not qualification. Current architecture layers allocate their full cell height
and reserve a 16pt gap for visible connector routes. Team matrices use measured
rows without painting unused space below them. Native fit and visual checks are
still required for changed content.

## Architecture input

Each slide uses `family: "architecture"` and a `content` object:

```json
{
  "left_rail_title": "DESIGN GUARDRAILS",
  "left_rail": [{"label": "Identity", "detail": "authentication policy"}],
  "layers": [
    {"id": "services", "heading": "DOMAIN SERVICES", "components": [{
      "id": "journey-api", "label": "Journey API",
      "detail": "bounded service contract", "fill": "accent",
      "width_weight": 1.2
    }]},
    {"id": "persistence", "heading": "PERSISTENCE", "components": [{
      "id": "repository", "label": "Repository", "fill": "muted"
    }]}
  ],
  "relations": [{"from": "journey-api", "to": "repository", "relationship": "dependency"}]
}
```

`layers` accepts 2–8 entries. Each layer requires 1–8 independently editable
component cells, each with a unique lowercase `id` and nonempty `label`. Optional
`detail` adds a second, separately measured text block. `width_weight` is
optional from 0–4 (default 1) and allocates relative width within that layer's
component row. The left control rail is explicitly scoped to all layers, with the supplied
`left_rail_title` shown as a separate subtitle above content-sized concern rows.
A vertical separator sits in a 20pt gutter. The main grid labels its columns
`LAYER` and `COMPONENTS`; each navy layer heading is centered beside its
component cells. Component cells fill each layer band, and the fixed 16pt
inter-layer lane provides vertical connections routing space. `fill` is one of `surface`,
`muted`, `accent`, or `active`.
The optional left rail has 0–5 labeled items; `left_rail_title` requires at least
one item. Relationships reference component IDs and use compose semantics
`reporting`, `dependency`, `advisory`, or `annotation`. Same-layer links use
facing left/right cell ports; links between layers use bottom/top ports according
to layer order. Connector ink is selected only when it reaches at least 3:1
contrast against white and both endpoint fills. The shared planner still rejects
obscured routes, cycles, or a color that fails its local contrast checks. There
is no adapter-level maximum relation count; a dense or cyclic graph can fail.

## Team input

Each slide uses `family: "team"` and a `content` object:

```json
{
  "pods": [
    {"id": "sponsor-pod", "title": "CLIENT SPONSORSHIP", "roles": [
      {"id": "sponsor", "label": "Business sponsor", "staffing": "staffing.client_part_time"}
    ]},
    {"id": "journey-pod", "title": "JOURNEY DELIVERY", "roles": [
      {"id": "engineer", "label": "Application engineer", "staffing": "staffing.wm_full_time"}
    ]}
  ],
  "reporting_lines": [{"from": "sponsor-pod", "to": "journey-pod", "relationship": "reporting"}],
  "responsibilities": {
    "columns": ["Product", "Engineering"],
    "rows": [{"label": "Journey priority", "values": ["A/R", "C"]}]
  },
  "matrix_position": "right"
}
```

Teams support 1–6 pods and 1–8 roles per pod. One pod uses one grid column;
two through four use two columns; five or six use three. Role IDs must be unique
across the slide. Optional staffing values are the compose tokens
`staffing.wm_full_time`, `staffing.wm_part_time`, and
`staffing.client_part_time`; using them adds the matching staffing legend.
Reporting connections refer to pod IDs and use `reporting`, `dependency`,
`advisory`, or `annotation`; at most 12 are accepted. An optional responsibility
matrix has 1–6 rows. Its default `matrix_position` is `right` and supports 2–4
responsibility columns, preserving vertical pod space; `bottom` supports 2–6
columns and reserves a lower panel at least 34% of content height (or large
enough for minimum row tracks). Same-row pod links use facing left/right ports;
cross-row links use top/bottom ports. Matrix cells retain their individual fills
and white gutters; the matrix does not draw a background border container.
Matrix values are caller-defined: provide a footer legend whenever values use
notation such as R/A/C/I. The shared planner validates port clearance and
relationship cycles.

Team roles are measured single-label tiles with an optional staffing category;
the adapter has no role subfields or pod-internal relationship endpoints.
Responsibility matrices have a fixed declared row/column count and text-only
cells. `Style.Ink` honors `text.primary` when it meets 4.5:1 contrast on a
component fill, otherwise it chooses a high-contrast fallback. Style color roles
are limited to `text.primary`, `text.secondary`, `surface`, `surface.muted`,
`accent`, `state.active`, `state.inactive`, and `border`. Typography is Arial
only; title size is 18–36pt and body/label sizes are 9–18pt.

Both adapters use `adapt.Decode`, reject unknown JSON properties and validate
known IDs, counts, fields, and relationship enums. Native text fitting remains
measurement driven. The maximum accepted counts define input bounds, not a
promise that every combination or arbitrary wording will fit. Dense and long
copy fixtures deliberately exercise that boundary.

## Fixtures

`architecture-team-v2.json` is the original combined example. `team-v4.json`
combines the current team fixtures with a caller-supplied RACI legend on all
matrix examples. The current architecture examples are combined in
`architecture-v4.json`.

- `fewer`: two-layer architecture and two-pod team.
- `standard`: three-layer boundary and four-pod team with a matrix.
- `dense` and `more`: additional layers, component cells, roles, pods and matrix entries.
- `longer`: intentionally long labels/details for measured fit review.
- `inverse`: inverse-profile examples for contrast checks.

The individual inputs (`architecture-*.json`, `team-*.json`) make each case
easy to inspect or compile independently.
