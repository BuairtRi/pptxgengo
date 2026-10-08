# Native editing across the template catalog

New v2 projects and CLI template/reference generation use `native-v1` by default.
The packaged template browsing deck uses the same profile. Rendering adapts each
eligible component after the source planner resolves its content, geometry, fonts
and colors; frozen source bundles and their checksums remain unchanged.

```sh
pptxgengo design project create --out ./client-deck --id client-deck \
  --title 'Client deck' --bundle v11 --template cards/3
pptxgengo design browsing-library --kind templates --bundle v11 \
  --as-of 2026-10-07 --out ./template-library
```

New projects persist `editing_profile: native-v1` in `deck.yaml`. Existing
projects retain their recorded setting: absent or `stock` preserves prior
rendering. Add `editing_profile: native-v1` deliberately to an existing project
and build a new baseline. `--editing-profile stock` requests the original
structure for new projects or template generation. Unknown profiles and native-v1
with the older compiler are refused. The source hash, compiled scene, layout
report and immutable receipt carry the editing setting into portable snapshots.

## Native objects

Eligible unordered lists use one native text box with styled bullet paragraphs,
including measured lead/body runs and supported inline formatting. Fonts, colors,
marker color, indents, initial wrap positions and paragraph gaps come from the
source plan. Wingdings supplies the square marker; it is an Office desktop
prerequisite and is not redistributed. The box initially retains the measured
content height; added copy may require resizing one box.

Eligible cards put their measured text paragraphs in the card's native shape.
Titleless body cards and cards with label, title and body paragraphs are eligible
when their resolved styles and positions fit the combined shape. Plain and
supported rich inline paragraph runs retain their source-resolved fonts, colors,
baselines and gaps. Simple outline and deemphasis surfaces retain the resolved
fill and border. Cards with featured/stateful decoration, unsupported content,
or incompatible placement retain their existing structure with an explicit
reason. Independently positioned content and artwork retain their existing
structure where a combined object would change paint order or appearance.

Tables are native PowerPoint AddTable objects. Eligible scenes remove only a
redundant outer component group around the table, retaining its cells, grid,
formatting and geometry. No cell-to-shape conversion occurs. Intentional parent
composition groups and separate adornments remain.

Every considered source scene records conversion or a retention reason in
`slides[].scenes[].warnings`. The browsing generator emits
`native-editing-coverage.json`, with an entry for every retained template and
counts of actual native list boxes, card shapes and tables. The v4.2.1 full-catalog
run emitted 586 native list boxes, 96 native card shapes and 206 native tables.
Converted list/card objects (682 total) matched source displayed text and
absolute line positions in that source-plan comparison. Counts describe rendered
outputs, not source JSON incidence; this does not qualify PowerPoint pixels or
Windows behavior. Normal source errors still fail the build; conversion never
hides invalid source or overflow. Retained scenes include explicit reasons.

## Evidence

Regression checks compare source and converted run styles, fonts, colors,
paragraph gaps and measured baselines at Comfortable, Compact and Dense. XML
checks cover native paragraph formatting and marker ink. Existing gallery
previews describe the original source specimens; they do not qualify all native
adaptations. Frozen v11 source content and pins are unchanged.

The approved [Mac comparison](native-component-demo.md) demonstrates list reflow
and insertion, whole-card movement and direct table cell editing. Catalog-wide
PowerPoint visual review, Windows qualification and exhaustive resizing and
alignment checks remain pending. Source-to-field adoption still requires manual
review; the native profile does not invent role mappings. Existing unprofiled
projects retain their recorded rendering, and adoption into them is deliberate.
