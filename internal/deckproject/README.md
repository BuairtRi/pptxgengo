# Authored deck project runtime v1

This package implements the Wave 2 adapter for `pptxgengo.deck-document.v1`; the
original Wave 1 documents describe broader workflow contracts. It imports the
native WMDS engine. Existing foundation/closed semantic JSON commands stay separate.

## Source and compilation

`Load(path)` reads the main YAML and its explicit slide/template/notes references.
Each YAML file contains exactly one document. It preserves hard/trailing breaks,
checks duplicate/unknown/case-sensitive fields and variant-specific node fields,
and reports exact YAML line/column plus JSON pointer for typed source errors.
Aliases, merges, custom tags, nonfinite numbers and implicit date values fail.
Each YAML source is at most 16 MiB; Markdown notes are at most 1 MiB; the combined
authored source limit is 64 MiB. Nesting is bounded to 100 levels.

### Editable multi-file projects

Inline documents remain supported. `project split` converts a project in place
into a short index, ordered slide files, optional template files, and Markdown
notes. With `--bundle v5`, shared stock slides gain human-readable content names.
The expanded canonical source and PowerPoint identity stay unchanged; the
existing toolchain lock is preserved. Conversion refuses occupied destinations
and retains the original source tree under `decisions/sources/`.
Legacy JSON-compatible flow collections become indented block YAML in the main,
slide and template files; scalar quoting, copy, comments and note bytes remain
preserved. Repeating the conversion is byte-idempotent.

```yaml
# deck.yaml
schema: pptxgengo.deck-document.v1
id: example-deck
title: Client discussion
year: 2026
toolchain: {lockfile: toolchain.lock.json}
slides:
  - slides/001-interviews.yaml
  - slides/002-recommendations.yaml
# Include local definitions only when needed:
local_templates:
  recommendation-layout: templates/recommendation-layout.yaml
```

Each slide file is a slide object. All referenced paths are relative to the
project root, including assets and briefs inside a slide or template file.
References do not recursively include other YAML. Unsafe paths, symlinks,
duplicate slide/template references and unknown fields fail with the source
file, line, column and document pointer.

```yaml
# slides/001-interviews.yaml
id: interviews
template: {scope: shared, id: interviews-readout/full}
content_kind: supplied_content
notes_file: notes/interviews.md
content:
  headline: Who we interviewed
  interviews:
    - name: Sam
      role: Director
bindings:
  headline: /slots/title
  /interviews/0/name: /slots/node01.rows.item01.n
  /interviews/0/role: /slots/node01.rows.item01.r
```

The binding names in this illustrative excerpt must match the actual selected
template's closed contract. Stock scaffolding supplies that complete map.
`content` contains authored copy; `bindings` maps its aliases or nested JSON
pointers to JSON pointers inside the original `values`. Unmapped technical
arrays/navigation can stay in `values`. Every content leaf must be consumed,
source/target pointers cannot overlap, and targets cannot collide with explicit
values. Required slots and array counts are still checked by the shared contract.
`notes_file` and inline `notes` are mutually exclusive; Markdown bytes become
speaker notes exactly, without parsing or trimming.

Slide edits, forks and section operations retain external files. Editing a slide
writes its own YAML; unselected slide files remain byte-identical. Comments on
untouched YAML nodes are retained. A mutation validates the candidate source
before replacing files, checks all predecessors under one guard, and saves a
recoverable preimage tree. Multi-file replacement uses individual atomic renames
with rollback on an I/O error; readers should reload after the operation completes.
External source bytes participate in build receipts, approval invalidation and
portable exports. Generated builds include an `authored/` source snapshot.

`Compile(project, bundle, engine)` binds shared templates using caller values only.
Shared source-declared image fields accept declared project asset IDs as well as
library registry keys. Only media `photo`/`src` slots are resolved, including
card media, bio and badge fields; business strings and authored values are
unchanged. Registry keys retain their meaning if a project ID collides; an
explicit `project:<asset-id>` selects the project asset.
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

## Draft Review Notes

A slide can carry an optional `draft_review` overlay independently of its shared
or local template. Normal builds include the native editable status tab and note
body. The default `edge` placement keeps the tab at the slide's top-right edge
and the note body outside the canvas; `pasteboard` puts the full component outside
the canvas. The component is 234 × 144 pt with an 18 pt status tab.

```yaml
draft_review:
  status: wip
  status_text: Work in progress
  status_color: kpi.risk
  owner: Ri
  due: "10/20"
  updated: "10/05"
  notes: |
    Confirm the source for this claim.
  placement: edge
```

Status is one of `notstarted`, `wip`, `complete`, or `qa`. The tab displays the
selected status label by default. Optional `status_text` overrides that label;
optional `status_color` independently overrides its color using `kpi.off`,
`kpi.risk`, `kpi.on`, or `brand.blue`. Custom status text is a single line, at most
128 bytes, and must fit the tab. Empty status text/color clears the override
and restores the selected status default. Dates are authored text and must be
quoted when YAML would otherwise interpret them as dates. Owner, due,
updated and notes are optional strings. Placement defaults to `edge`. Draft notes
are internal metadata and are excluded from audience copy review. Client export
removes the entire native component and its metadata; reviewer, maintainer and
offline exports retain it. The immutable draft build is preserved.

```sh
pptxdesign project slide draft-review set --project PATH --id findings \
  --status wip --status-text 'Work in progress' --status-color kpi.risk --owner Ri --due 10/20 --updated 10/05 --notes 'Confirm source.'
pptxdesign project slide draft-review set --project PATH --id findings --status qa
pptxdesign project slide draft-review show --project PATH --id findings
pptxdesign project slide draft-review clear --project PATH --id findings
```

`set` preserves omitted fields and initializes a new note with `notstarted`.
Explicit empty strings clear optional text fields. Repeating an identical `set`
or clearing an absent note preserves source bytes and produces no decision receipt.
`clear` removes the entire
property; `show` emits its current JSON value. Mutations preserve source comments
and unrelated slide metadata, use the same atomic mutation guard and preimage
backup as other slide operations, and invalidate the affected build/review inputs.
Both inline and split source projects are supported. No template fork is needed.

### Native text model

New object maps declare `pptxgengo.native-text-model.v1` and retain paragraph/run/
cell addresses, exact text, source-slot identities and formatting/structure
hashes. Typed card title/body mappings select the specific field; ambiguous or
unsupported text remains manual review. This is baseline infrastructure, not an
edited-PPTX adoption engine or desktop qualification. See
[native field mapping](../../docs/native-field-mapping.md).
