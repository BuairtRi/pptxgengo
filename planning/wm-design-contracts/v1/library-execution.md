# Full library candidate execution

Date: 2026-10-02. Engine: `wmds-go-foundation.v2`.

All 97 canonical variants have closed source-pinned content definitions and Go
scene adapters. The frozen source is unchanged. Material native layout
resolutions are versioned and recorded in layout reports and the component
contracts. Generation and font measurement run in Go. PowerPoint is used for
bounded development review, not as a generation prerequisite.

## Commands

```sh
pptxdesign library-catalog --engine wmds-go-foundation.v2 > catalog.json
pptxdesign library-reference --engine wmds-go-foundation.v2 --year 2026 --out NEW-DIR
pptxdesign library-source-reference --engine wmds-go-foundation.v2 --year 2026 --out NEW-DIR
pptxdesign library-sweep --engine wmds-go-foundation.v2 --year 2026 --out NEW-DIR
pptxdesign library-bound-sweep --engine wmds-go-foundation.v2 --year 2026 --out NEW-DIR
```

Reference and sweep commands accept `--family` to select one canonical family.
Catalog reports all eight families. Source reference keeps exact source example
copy. Bound reference exercises the public content compiler. The four earlier
APIs (`cards/3`, `cards/4`, `stats/four-metrics`, `takeaway-rail/metrics-rail`)
retain their typed values and eight-slide evidence; other variants use the
`pptxgengo.wmds-library-bindings.v1` values contract below.
The catalog exposes `value_schema` for each of those four earlier APIs, including
its Go type, fields and exact card counts; they do not use the new slots contract.

## Content API

The enclosing input remains `pptxgengo.wmds-template-document.v1` with explicit
`year` and slides containing unique `id`, exact `template`, `content_kind`,
and `values`. Classify content as `synthetic_example` or `supplied_content`.
For the 93 new bindings, values has exactly two required fields:

- `slots`: map from the exact named catalog slot to a string, finite number or boolean,
  according to its declared kind. Every slot is required. Empty strings are
  accepted only where the pinned definition declares `allow_empty`.
- `keys`: map from each declared array name to exactly its specified count of
  unique stable ASCII keys. Caller order determines presentation order.

`library-reference` writes `template-content.json`, a complete synthetic example
of the public input. Use the catalog's source pointers to understand a slot;
callers supply named slots, not arbitrary pointer mutations. Geometry, typography,
surfaces, style discriminators and connector topology remain source owned.
Declared content exceptions include registered icon IDs, metric status values,
stepper/vstepper step states and checklist `on` booleans. They remain constrained
by the renderer's registry or enum; caller values cannot introduce arbitrary
style or geometry fields.
There is no omitted-value fallback to example copy. Unknown fields, duplicate
members, missing slots/keys, wrong types/counts, unsupported fields and overflow
produce named errors. Array counts are fixed for each variant.

Generated charts retain native editable workbook data. Direct labels and totals
outside charts are editable slide text and update through regeneration; manual
PowerPoint Edit Data does not synchronize those external overlays. Quadrants are
native editable diagram objects. Registered photos/art remain pictures; chart
and diagram content is not flattened to images.

## Resources and evidence

The source bundle includes pinned original Plex static faces and WM logos.
Registered source media resolves against the local WM branding root
(`WMDS_BRANDING_ROOT`, default `~/Documents/branding`); missing or changed assets
fail. `UsedPrimitiveAssets` exposes paths, SHA-256 and crop provenance for only the
media used by a deck. The branding root is needed during generation; selected
media is embedded in the PPTX, so the generated presentation does not require
that external directory to open or share. Source and bound reference layout reports record owned
native parts, groups, text measurement, table cells, chart data and resolutions.

Sweep receipts record each variant independently; a failure cannot hide the
other variants. Successful generation is named `generated_native_review_pending`.
Catalog availability and generation are not native visual qualification or a
qualified arbitrary-content envelope. The current user review and correction
record is `samples/wmds-library-20261002/user-review-pass02.json`; its page numbers
refer to the 87-slide pass02 deck and must not be applied to later full decks.

A line-chart writer schema defect that blanked the team-handoff slide during
PowerPoint repair is corrected. The isolated corrected slide opens without
repair; evidence is `samples/wmds-library-20261002/line-writer-fix02/native-review.json`.
The corrected 97-slide source reference in
`samples/wmds-library-20261002/review04/library-reference.pptx` and the 97-slide
bound fixture in `bound-review04` both open natively without repair. The source
reference has a native PDF export and separate page-review receipts in `review04`;
the combined page-review receipts cover all 97 exact source fixtures with no
remaining visible defects observed. The bound fixture's
successful opening is package acceptance; it is not a visual review of every
bound page or a qualified arbitrary-content envelope. The optional notes-master
list compatibility resolution and retained notes relationships are documented in
`notes-master-native-compatibility.md`.
Versioned acceptance and user-correction receipts are in `reference-review/`.
Generated decks, PDF exports and previews remain local artifacts excluded from Git.
No tests were added or run in this implementation pass.
