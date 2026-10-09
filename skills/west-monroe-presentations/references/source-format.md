# deck.yaml source format

`deck.yaml` and the files it references are the deck's source. Edit the source, run `project check`, then build. Errors report the file, line, column and field. To change slides, follow [editing slides](editing-slides.md).

YAML comments are allowed. Duplicate keys, anchors, aliases, merge keys, custom tags, implicit dates and unknown fields are rejected. Each file holds one YAML document.

## The index

```yaml
schema: pptxgengo.deck-document.v1
id: client-deck
title: Phase 2 proposal
year: 2026
editing_profile: native-v1
toolchain:
  lockfile: toolchain.lock.json
context:
  project: project.md
  audience: audience-context.md
  outline: outline.md
  sources: sources
  claims: claims.md
assets:
  client-logo:
    path: assets/objects/sha256/REPLACE_WITH_REGISTERED_SHA256
    description: Client logo supplied by the operator
local_templates:
  phases-four: slides/templates/phases-four.yaml
slides:
  - slides/001-cover.yaml
  - slides/002-situation.yaml
sections:
  - {id: opening, title: Opening, before_slide_id: cover}
  - {id: approach, title: Approach, before_slide_id: approach-divider}
```

- `slides` is the page order. Entries are slide file paths (after `project split`) or inline slide objects.
- `local_templates` maps IDs to template files or inline definitions.
- `context` keys are limited to `project`, `audience`, `outline`, `sources`, `claims`, `composition_log`, `decisions`, `win_strategy` and `state`. Linked files invalidate approvals when they change; a directory (such as `sources`) tracks every file in it. `state` names the generated `state.json`.
- `sections` group slides in PowerPoint; manage them with `project section` ([editing slides](editing-slides.md#sections-and-dividers)).
- Builds deduplicate media and shrink JPEGs to 220 ppi at quality 90 by default. To change that, set `media_optimization` with all six fields (`deduplicate`, `resize_jpeg`, `compression`, `pixels_per_inch`, `jpeg_quality`, `min_savings_percent`); `resize_jpeg: false` keeps original JPEG bytes. Vector and lossless images are never resized.
- Register assets with `project asset add` to write actual object paths and hashes; the example object path above is illustrative.
- All paths are project-relative. Absolute paths, URLs, `..` and symlinks are rejected.

## A slide

```yaml
id: situation
template: {scope: shared, id: cards/3}
content_kind: supplied_content
brief: briefs/situation.md
evidence_refs: [C03]
notes_file: notes/situation.md
hidden: false
density: comfortable
header_density: comfortable
auto_density: true
content:
  eyebrow: What we heard
  headline: Your PMs spend [[two days]] on each PRD while interviews go unsynthesized
  cards:
    - {key: prd, title: PRDs, body: Each takes about two days to write.}
    - {key: interviews, title: Interviews, body: Recordings pile up unsynthesized.}
    - {key: roadmap, title: Roadmap, body: Priorities go to whoever argues hardest.}
bindings:
  eyebrow: /eyebrow
  headline: /title
  cards: /cards
```

| Field | Meaning |
| --- | --- |
| `id` | Stable slide ID; survives reordering |
| `template` | `{scope: shared, id: KEY}` for a library template, or `{scope: local, id: ID}` for a local template. Optional `revision` pins a shared revision. |
| `content_kind` | `supplied_content` for real copy, `synthetic_example` for placeholders. It labels content; it doesn't verify claims. |
| `content` + `bindings` | Readable copy, and the map from each content key or pointer to the template slot it fills. Every content value must be bound; required slots and array counts still come from the template. `project scaffold --stock` writes a complete map. |
| `values` | Template slot values in the template's own shape. Used by local templates and for slots not covered by `bindings`; detached local templates can also use `content` and `bindings`. |
| `brief` | Project-relative path to the slide's internal brief |
| `evidence_refs` | Claim IDs from `claims.md` that the slide relies on. `project check` fails on an unknown ID when `context.claims` is linked. |
| `notes_file` or `notes` | Speaker notes (Markdown file, or inline text); not both |
| `draft_review` | Optional internal status tab and off-slide review card; see [Draft Review Notes](draft-review-notes.md) for fields and client-export removal |
| `hidden` | `true` hides the slide in PowerPoint |
| `density` | Entire slide body: `comfortable`, `compact` or `dense`. Omit to start at the template's authored tier. |
| `header_density` | Header typography tier, independently defaulting to `comfortable`. |
| `auto_density` | V11 defaults to `true`: try denser body tiers when needed, with warnings. `false` keeps the requested tier. |

- Get template keys from search or the catalog; never guess them.
- Missing values are never filled from template examples.
- `[[…]]` in a title applies the slide's highlight mark to one to four words, once per slide.
- Image fields take a project asset ID (`photo: client-logo`) or a library asset key from `design asset-catalog`. Use `project:client-logo` if the two collide.
- Density is slide metadata, outside `content`, `values` and `bindings`. Read
  [typography density](typography-density.md) for role scales, limits and warnings.
  These fields do not alter template geometry or fixed item counts. Edit them
  directly in slide YAML; the `project edit` patch format does not include them.

## Assets

`assets` maps your IDs to a project-relative `path` (or a library `registry_id`), with optional SHA256 and provenance. Keep originals inside the project; builds never rewrite them. A derived image (cropped, resized) needs a `pptxgengo.asset-derivation.v1` JSON receipt naming the source asset, both hashes and the operation.

## Local templates

Design custom pages with [custom slide design](custom-slide-design.md) first. Start from `project scaffold --bundle v11 --template KEY --reason R` when a shared template is close ([editing slides](editing-slides.md#change-a-slides-layout)); use the project's locked bundle for a historical project.

A local template declares:

- `frame`: `{scope: shared, id: wmds/frame/<rail>-<footer>}`, where rail is `none`, `left`, `right` or `nav` and footer is `compact`, `tall` or source-supported `slim`. Optional `frame_options` sets `title_lines`, `source_lines`, header density, nav and split composition; `frame_chrome` keeps custom whiteboard geometry and emphasis. Use only allocations supported by the pinned source; these local-template options do not automatically modify stock frames or expand them during fitting.
- `grid`: `{scope: shared, id: wmds/grid/12-columns}`.
- `zones`: typed content slots with `role`, `required` and a JSON `schema` (single type, properties, required, items, min/max, enum/const, description). Roles `slide-title`, `eyebrow`, `source` and `nav` fill the frame; every other zone needs a node binding.
- `nodes`: each with a stable `id`, a `kind`, and `placement` in the `body`, `rail`, `short_body` or `tall_body` zone by grid `span` (`{start, count, y_pt, height_pt}`) or `rect`.

Node kinds:

| Kind | Use |
| --- | --- |
| `component` or `composite` with `definition: {scope: shared, id: wmds/component/<type>}` | Any design-system component. Prefer these. |
| Typed `text.block`, `card`, `data.metric`, `richtext` | Labeled text block, single card, metric, rich text |
| `text`, `box`, `image`, `rule`, `group` | Simple elements. Never build a whole page from these. |

Component `<type>` is the scene node type, not the catalog name (`stepper`, not `seq.stepper`). Accepted types:

`text`, `textblock`, `bullets`, `ol`, `list`, `schedule`, `grouplabel`, `numhead`, `colhead`, `strongnum`, `pullquote`, `imageframe`, `logo`, `art`, `square`, `mark`, `thumbnail`, `table`, `chart`, `card`, `cardrow`, `metric`, `callout`, `feesummary`, `block`, `frame`, `chevron`, `textarrow`, `connector`, `container`, `cylinder`, `node`, `layerrow`, `matrix`, `beforeafter`, `stepper`, `vstepper`, `phasehead`, `phases`, `timeaxis`, `pyramid`, `funnel`, `cycle`, `road`, `roadfork`, `gauge`, `bracket`, `scorelegend`, `gantt`, `swimlane`, `legend`, `pod`, `role`, `person`, `orgchart`, `governance`, `logoslot`, `device`, `plane`, `dotmap`, `teamcurve`, `venn`, `maturity`.

- Component arguments take literal values or `{binding: zone}`. They cannot set `type`, `id`, `x`, `y`, `w` or `h`; the placement sets the allocation. List-like components (bullets, tables, steppers, gantt and similar) take their height from content.
- Diagram-internal geometry (stage positions, connector endpoints) is set explicitly in the arguments.
- Arrays with item identities need matching `keys` entries.
- The planned output must fit the allocation, or the build fails.
- Images: PNG, JPEG or simple self-contained SVG, with `cover` or `contain`.
- For each component's arguments, inspect a shared template that uses it (`library-inspect`) or `source/components/v0/components.json` under `design_system_default` from `pptxgengo paths`. Source-checkout examples are optional and are not included in every release archive.

## Persisted editing profile

Optional top-level `editing_profile` is `stock` (the absent setting in existing projects) or
`native-v1` (persisted by default in new v2 projects). It changes native structure for eligible source scenes using
the v2 engine, while retaining pinned source styles. The profile is part of the
authored source hash, compiled scene, layout report and build receipt. Preserve
existing baselines and review converted builds. It does not alter shared bundle
bytes or qualify every template. Complex variants remain in their stock form.

## Schema and template upgrades

The deck schema is independent of the engine and template bundle version.
Preserve explicit template revisions and local definitions until their target
contracts are reviewed. Follow [project upgrades](upgrading-projects.md) before
changing schema, bundle, executable or template pins.

## Reconciled native geometry (development build)

Reviewed transforms and paint order are source-owned slide fields:
`native_geometry`, `native_order`, and `native_geometry_template`. Transform
keys are receipt-recorded object names; each value declares native kind, parent,
point coordinates/extents, optional rotation/flips, and group `child_space`.
They do not use frame-relative placement coordinates. They are included in source
hashes, builds, numbered versions and portable project packages. Use
[the geometry workflow](architecture-geometry.md) to create/review them; a
changed template reference requires explicit reset and review.

Each persisted transform also pins its authored geometry basis with
`source_geometry_sha256`. Changes to the underlying node/child coordinates are
rejected instead of applying an old transform to a newly generated coordinate
space. Use an explicit layout reset and fresh baseline to review that change.
The geometry proposal renderer verifies the actual current executable, fonts and
bundle against the project lock before computing current YAML geometry.

Reviewed structural reconciliation writes removed/copied components into the
project-owned local template and adjusts the selected slide's bindings. Copied
blocks receive fresh node IDs and independent content keys; their exact native
transforms and paint order use the same pinned fields above. Use
`reconcile propose --geometry --structure-map FILE` for explicit copy ownership,
then review `structure` field IDs. Never inject native XML into authored YAML.

### Logical diagram containment (development build)

A local slide can persist `diagram_containment` keyed by exact native names:

```yaml
diagram_containment:
  node06:
    container: node05
    padding: {top_pt: 28, right_pt: 12, bottom_pt: 4, left_pt: 12}
```

Create/review these rules with `project diagram contain|uncontain`. Missing
objects, cycles and invalid padding fail validation. Padding is measured in the
container's placement axes before rotation/parent scaling; the member and its
descendants must fit. Rules also require `native_geometry_template` to match the
selected template. Native copies inherit source membership; removed members are
pruned. A layout reset preserves these rules. Numbered source snapshots and
complete project shares carry them with the slide YAML.

### Attached elbow routes (development build)

`wmds/component/attached-connector` arguments include endpoint node/site pairs,
`route: straight|horizontal|vertical` (default straight), and an optional elbow
`bend` fraction in `[0,1]` (default `.5`). Do not supply a bend for a straight
connector. Use `project diagram connect` to preview these options.

Reviewed native elbow edits persist an optional `native_geometry/<name>/route`:
`{preset: bentConnector3, adj1: 65000}`. `adj1` is a literal DrawingML guide in
100000ths of the native width, before flips/rotation. Native guides may extend
past the endpoint rectangle; calculated route allocations must still fit the
frame and declared containment. Unknown presets/formulas remain manual review.
Use reconciliation to write these overrides and their source basis pins.

## Family and import patch files

Keep reusable source in local templates under `slides/templates/`. Team patches
use `pptxgengo.team-patch.v1`; Gantt patches use `pptxgengo.gantt-patch.v1` with
`timebase: periods`; native import maps use `pptxgengo.native-import-map.v1`.
Each requires actor/reason and the current source SHA256; imports also require
the input PPTX SHA256 and explicit formatting policy. These are command inputs,
not replacement deck schemas. Read the [team](team-composition.md),
[Gantt](gantt-composition.md) or [import](native-imports.md) contract for operations.
Explicit materialization changes selected copy from bindings to local constants.
These commands require the newer development CLI.
