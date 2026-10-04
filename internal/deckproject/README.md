# Authored deck project runtime v1

This package implements the Wave 2 adapter for `pptxgengo.deck-document.v1`; the
original Wave 1 documents describe broader workflow contracts. It imports the
native WMDS engine. Existing foundation/closed semantic JSON commands stay separate.

## Source and compilation

`Load(path)` reads exactly one YAML document. It preserves hard/trailing breaks,
checks duplicate/unknown/case-sensitive fields and variant-specific node fields,
and reports exact YAML line/column plus JSON pointer for typed source errors.
Aliases, merges, custom tags, nonfinite numbers and implicit date values fail.
Maximum source size is 16 MiB and nesting is bounded to 100 levels.

`Compile(project, bundle, engine)` binds shared templates using caller values only.
It resolves local template zone schemas/bindings, frame/grid placement and scoped
`project:<asset-id>` payloads into `wmdesign.Document`. Its result is derived IR,
never an alternative editable source. `Check` verifies the exact executable/Go/
platform/engine/bundle tree pins and compilation; native scene planners enforce
source component field contracts, topology and measured fit during `Build`.

Supported local vocabulary:

- Shared grid `wmds/grid/12-columns`; frame references
  `wmds/frame/{none,left,right,nav}-{compact,tall}`. Optional `frame_options` uses
  the existing frame resolver for header/source rows, density, nav and split
  composition; rail/footer must match the reference. A zone with role `nav`
  binds `{items: [{key,label}], active}` into the frame. Custom whiteboard geometry
  and emphasis can be preserved through `frame_chrome`.
- Frame-relative `body`, `rail`, `short_body`, `tall_body` placements by rect or
  grid span; unavailable zones/overflow fail. `panel`, `whiteboard`, `source`
  placement are reserved schema vocabulary with explicit unsupported errors.
- Text, box, image, rule and paint-ordered groups. Groups retain stable node
  prefixes. Images support PNG/JPEG/self-contained SVG, cover/contain and native
  rotation. Custom SVG uses a bounded path/rect/circle/group grammar with finite
  unitless geometry; stylesheet elements, hidden presentation styling, unknown
  or namespaced attributes, resources/events/DTDs fail before embedding.
  Fill and stroke:none inline styles are supported; unsupported drawing
  commands are checked by the native raster fallback. Rules currently use the
  existing line ink and .75pt weight. Box borders use the outline surface.
- Typed `text.block`, `card`, `data.metric`, `richtext` arguments (richtext uses body
  typography), or `wmds/component/<source-type>` and `wmds/composite/<source-type>`
  for the native source scene vocabulary. Arguments use literal values or explicit
  `{binding: zone}` objects. Top-level allocation geometry/type/ID overrides in arguments fail. Nested
  diagram geometry remains explicit in the arguments (for example, stage
  positions or connector endpoints).
  Arrays rendered with item identities require matching `keys` entries (relative
  source-array paths). Source-specific fields are validated by existing planners,
  and the planned output must fit the declared allocation.
- Zone schema subset: single type, properties/required/additionalProperties:false,
  items, min/max items/length/value, enum/const, description. Other schema keywords
  fail. Capacity notes do not establish fit or native qualification.
- The frame consumes zone roles `slide-title`, `eyebrow`, `source` and `nav`.
  Other semantic roles remain extensible and require an explicit node binding.
  An unbound zone fails at load with its source location; `role: title` does not
  implicitly render a frame headline. Use `role: slide-title` for that purpose.

Original assets stay at safe project-relative paths. Paths with absolute/URL/
parent traversal or symlink steps fail. A derivative declaration requires an
immutable JSON receipt with schema `pptxgengo.asset-derivation.v1`, `source_asset`,
`source_sha256`, `result_sha256`, `operation` and optional `parameters`. The source
and result hashes must match current asset bytes. Receipts and frozen ancestry
snapshots participate in invalidation. Original files are never rewritten by build.

## Build, state and review

`Pin` creates an absent lockfile; it never replaces existing pins. `Build` obtains
an exclusive project guard, checks inputs, preserves the existing generated
baseline, writes new read-only artifacts and atomically updates `state.json`.
A build is aborted if source/lock/dependencies change while compiling. The fixed
native timestamp is `2000-01-01T00:00:00Z`; identity seed is canonical authored
source SHA256. Receipts separate native artifact hashes from execution time.

`ObjectMap` records stable slide/node/item/part IDs, physical native slide parts
and IDs, emitted text, source pointers and baseline values. Generated chrome and
unbound source parts are explicitly marked. It supplies reconciliation inputs;
it does not implement edited-PPTX reverse compilation or claim every geometry
change can be mapped back to YAML.

`Status` computes invalidations without mutating state. `Resume` persists the
current resumption point. `Approve` writes hash-bound approval records and scope.
Changed copy invalidates affected content/selection/review; local geometry/asset/
ancestor changes invalidate composition/render stages. Unaffected slide approvals
remain valid. A later same-stage/same-scope approval supersedes its predecessor
while retaining history. Review/build/delivery approvals require inputs equal to
the current generated baseline; a compiler success does not assert PowerPoint or
visual approval. Any change to a generated artifact/receipt blocks regeneration
and export. Restore state rather than guessing when build history exists but
`state.json` is missing. Context keys are closed to `project`, `audience`, `outline`, `sources`, `claims`,
`decisions`, `state`. Files hash their own bytes; directories hash their full safe
file tree. Use `context.sources: sources` to track new/changed meeting evidence
under that folder. `context.sources: sources/index.md` tracks that index only;
linked files are not traversed. `context.state: state.json` names generated state and is
excluded from authored dependency hashes to avoid invalidating itself.

## Mutations and exports

`Fork` copies a verified local template and optionally rebinds slides that already
use that parent. `Detach` converts a shared generic source-scene slide to a local
template, retaining supplied slot content in named values. Both preserve byte-exact
predecessor YAML, frozen ancestor JSON/hash/source-file pins and decision receipts.
The serializer rewrites formatting/comments. Unsupported legacy typed IR or
source-note/stamp chrome returns an explicit error; authoring source stays intact.

`Export` writes a new native Go ZIP with safe members, fixed ZIP timestamps and an
export manifest. Client exports contain the PPTX; reviewer packets add report/map/
receipt and draft bindings. Maintainer exports include private source/context/
evidence, original/derivative assets, decisions, approval history and the current
build. Offline exports additionally include the exact runtime, bundled fonts/
calibration/templates and used pinned registry assets. Offline execution is
restricted to the compiler's original OS/architecture; PowerPoint/PDF/font system
installation is not included. Outputs/receipts are retained separately from assets.

Not implemented: automatic PDF import/export, arbitrary edited-PPTX reconciliation,
legacy typed IR detachment, source stamp/footnote chrome detachment, browser/native
visual qualification. `library-fit` owns candidate alternatives using explicit
real-content bound documents; a reviewer packet never substitutes source specimens.

## Sections, dividers, hidden slides and speaker notes

Optional `sections` group contiguous slides in native PowerPoint:

```yaml
sections:
  - id: opening
    title: Opening
    before_slide_id: introduction
  - id: findings
    title: Findings
    before_slide_id: findings-divider
```

The first section anchors the first slide. Subsequent anchors must follow slide
order, with unique stable IDs, titles and anchors. Each group ends at the next
anchor. Omitting sections preserves the previous ungrouped output.

```sh
pptxdesign project section list --project PATH
pptxdesign project section add --project PATH --id findings --title Findings --before results
pptxdesign project section rename --project PATH --id findings --title Recommendations
pptxdesign project section remove --project PATH --id findings
```

Adding the first section midway through a deck creates an explicit `Opening`
group for preceding slides. The JSON receipt reports its ID. Rename changes the
native section name; it retains visible divider copy, reported separately as
`divider_title`. Remove retains all slides, including visible dividers. Removing
the first group absorbs its slides into the next group; removing the sole remaining
group leaves an ungrouped deck.

To insert a visible divider with the section:

```sh
pptxdesign project section add --project PATH --id findings --title Findings --before results \
  --divider divider/panel-edge --divider-photo photo-abstract-cubes
```

`panel-edge`, `panel-photo` and `full-photo` accept the caller's title and photo
asset and derive the section ordinal. They use the exact pinned library slots.
Other divider variants require `--divider-values FILE.json`, containing their
complete closed library binding. No source example copy or asset is selected
implicitly. The inserted slide ID defaults to `SECTION-divider`, overridable
with `--divider-slide-id`. A fit failure leaves the source unchanged. Group and
divider mutations preserve YAML comments and scalar styles, save exact source
predecessors, and write a decision receipt before atomically replacing YAML.

Slides may optionally carry `hidden: true` and `notes: |` speaker text. Hidden
slides remain in their native sections and in the deck, with native PowerPoint
visibility metadata. Notes preserve authored text and line breaks (the PPTX
writer normalizes newlines to CRLF), followed by existing provenance notes.
Neither a slide brief path nor visible slide copy is substituted for notes.
Notes, visibility and sections invalidate the relevant approval dependencies.
