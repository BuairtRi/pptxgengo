# Import selected native text shapes

Use this route to bring explicitly selected content from a foreign PowerPoint
into a local project template. It is separate from receipt-backed reconciliation
of a working copy. Check installed help: `project native-import` requires the
new development CLI and is not in the promoted v4.2.1 release.

## Choose the route

- For edits to a project's built deck, preserve the baseline and use
  [reconciliation](editing-slides.md). Tagged transforms and supported copied
  blocks retain their existing contract; do not reimport them to bypass findings.
- For a foreign deck, use `source-inventory` when you need its overall content,
  images, tables or charts. That command inventories content; it does not create
  editable project source.
- Use `native-import` for selected **top-level, unrotated rectangular shapes with
  nonempty plain text**. A selection must have an explicit transform and unique
  native name within its slide part. Multiple plain paragraphs are allowed; rich
  runs, bullets, fields, hyperlinks, relationships, groups, pictures, connectors,
  tables, charts, SmartArt and complex geometry are refused by this slice.

The import policy is **`reauthor_with_design_style`**. Original font, fill,
outline, effects, padding and paragraph styling are replaced by the explicitly
selected design-system text role/alignment and optional editable-block surface.
The original package is retained unchanged as private evidence. This is a
content-and-geometry import, not a fidelity-preserving native-object transplant.
Do not choose it when preserving the original styling is required.

## Inspect and map

Inputs are bounded to a 32 MiB PPTX, a 1 MiB map and 1–100 selections. Large
photo-heavy decks need a separately prepared smaller source package; retain the
original and record the preparation. Do not silently drop content to bypass a limit.

```sh
pptxgengo design project native-import inspect --in ./source-deck.pptx > native-inventory.json
pptxgengo design project diagram inspect --project ./deck --slide architecture > destination.json
```

The inventory returns `pptx_sha256`, exact part/name addresses, geometry, text
and unsupported reasons. Part sorting is lexical address order, not presentation
order. Select addresses from this inventory; do not assume `slide1.xml` is the
first visible slide. Unsupported shapes remain in the original package.

Detach/fork the destination template first. Copy its current `source_sha256`
into a strict map and select a new stable node ID for every imported object:

```yaml
schema: pptxgengo.native-import-map.v1
expected_source_sha256: REPLACE_WITH_CURRENT_DESTINATION_HASH
expected_pptx_sha256: REPLACE_WITH_INVENTORIED_SOURCE_HASH
actor: deck-author
reason: Reauthor the sourced service label in the selected architecture
formatting_policy: reauthor_with_design_style
selections:
  - part: ppt/slides/slide1.xml
    name: Service label
    node_id: imported-service
    kind: editable-block
    style: body
    surface: subtle
    align: center
    rect: {x_pt: 240, y_pt: 120, width_pt: 180, height_pt: 54}
```

`kind` is `text` or `editable-block`; text must omit `surface`, while an editable
block requires it. `style` and `align` are required. Without `rect`, the importer translates the original slide-absolute position
into the destination body-relative coordinates, preserving its world position.
An explicit `rect` is relative to the destination frame's `body` zone, just like
diagram patch rectangles. Inspect the frame before mapping; do not supply an
absolute source coordinate as a body-relative override. An original position
outside the destination body must be repositioned explicitly to fit.

The importer adds independent local template nodes. It does not infer a role,
pod member, reporting relationship, task, date or gate from their appearance.
Use [team composition](team-composition.md) or [Gantt composition](gantt-composition.md)
to author that meaning explicitly.

## Preview, apply and review

```sh
pptxgengo design project native-import patch --project ./deck --slide architecture \
  --in ./source-deck.pptx --map ./native-import.yaml > import-preview.json
pptxgengo design project native-import patch --project ./deck --slide architecture \
  --in ./source-deck.pptx --map ./native-import.yaml --apply > import-applied.json
pptxgengo design project build --project ./deck
```

Preview changes no files. Apply checks both source hashes, measures fit, rejects
active native geometry/order overrides, and commits source plus the mapping and
raw source package in one guarded transaction. The raw PPTX is stored once under
`assets/objects/sha256/<hash>`; its map and selected-object observations are in
`decisions/native-import-*.json`. Keep these private when sharing a complete
project because they contain the original deck. They are not copied into every
numbered source version. Full project shares retain the shared asset store and
decisions; a generated deck alone does not carry this maintainer evidence.

Review the preview's copy, geometry and changed source ownership. Build/render
through PowerPoint and inspect every imported item for wrapping, clipping,
meaning and neighboring content. An imported string becomes local authored copy;
subsequent edits must use that local source or reviewed reconciliation. A
successful import does not qualify the rest of the foreign deck.
