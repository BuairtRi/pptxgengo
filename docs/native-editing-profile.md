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
Paragraph fonts, colors, baselines and gaps remain source-resolved. Complex
independently positioned content and artwork retain their existing structure
where a combined object would change paint order or appearance.

Tables are native PowerPoint tables. Eligible tables lose a redundant outer
component group, retaining the original cells, grid, formatting and geometry.
Intentional parent composition groups and separate adornments remain.

Every considered source scene records conversion or a retention reason in
`slides[].scenes[].warnings`. The browsing generator emits
`native-editing-coverage.json`, with an entry for every retained template and
counts of actual native list boxes, card shapes and tables. This is rendered
catalog coverage, not a count inferred from source JSON. Normal source errors
still fail the build; conversion never hides invalid source or overflow.

## Evidence

Regression checks compare source and converted run styles, fonts, colors,
paragraph gaps and measured baselines at Comfortable, Compact and Dense. XML
checks cover native paragraph formatting and marker ink. Existing gallery
previews describe the original source specimens; they do not qualify all native
adaptations.

The approved [Mac comparison](native-component-demo.md) demonstrates list reflow
and insertion, whole-card movement and direct table cell editing. Catalog-wide
PowerPoint visual review, Windows qualification and exhaustive resizing and
alignment checks remain pending. Stock bullet/card source adoption still requires
manual review; their new object structure does not invent role mappings.
