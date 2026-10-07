# Persisted native editing profile

Native editing components are available explicitly as `editable-list`,
`editable-card` and `editable-table`. The approved Mac comparison covers those
plain examples at all three densities. For existing source templates, the
optional `native-v1` editing profile applies a bounded conversion after the
ordinary source planner successfully resolves copy, geometry, fonts and colors.
The source bundle and shared template bindings are unchanged.

```sh
pptxgengo design project create --out ./client-deck --id client-deck \
  --title 'Client deck' --bundle v11 --template cards/3 \
  --editing-profile native-v1
```

For an existing project, deliberately add `editing_profile: native-v1` to
`deck.yaml`, then check/build a new baseline. Absent or `stock` preserves prior
rendering. Unknown profiles and use with the older compiler engine are refused.
The source hash, compiled scene, layout report and immutable receipt record the
setting, so rebuilds and portable snapshots carry the same intent.

## Eligible conversion

- Flat unordered lists containing up to 100 plain strings: one native text box
  with square bullet paragraphs. Source-resolved font, color, indent, measured
  wrap lines and paragraph positions are retained. The initial box uses the
  original measured content height; added copy may require resizing that box.
  Wingdings supplies the square bullet and is an Office desktop prerequisite,
  not a redistributed font.
- Cards with one plain title and one plain body paragraph, one rectangular fill
  and no extra paint objects: one filled native object. Resolved padding,
  title/body styles, colors, initial wrap positions and paragraph gap come from
  the original successful source plan, rather than the explicit component's
  fixed defaults. Eligible unnumbered card-row children can also convert; the
  parent row group remains available.
- Tables whose successful plan contains only one native table: remove its
  redundant component wrapper. Parent composition groups remain intentional.

Rich/nested/ordered lists, multiple body blocks, decorated/outlined cards,
small-body cards and tables with extra adornments retain their existing object
structure. Each considered scene records conversion or exclusion in
`slides[].scenes[].warnings`. Normal planner errors still fail the build;
conversion never masks invalid source or overflow.

## Evidence and review boundaries

Regression checks compare source and converted run styles, font identities,
colors, text and measured baselines across Comfortable, Compact and Dense.
Real project builds retain the profile in their evidence. These checks preserve
initial planner geometry; they do not prove pixel identity for every source
catalog template. Existing gallery previews qualify their original specimens.
New profile builds still need PowerPoint visual review.

The [Mac comparison](native-component-demo.md) records actual native list reflow
and insertion, whole-card movement and table cell editing, with saved identities
verified. Windows, resizing/alignment and exhaustive catalog qualification remain
pending. Stock card roles and bullet item identities are not inferred for source
adoption: retain edited copies and review them manually. The explicit editable
card's named role bindings have their separate reviewed adoption contract.
