# Slice 1 completion — 2026-10-02

## Delivered

The refreshed round-4 source is frozen as `wmds-library.v2`, selected explicitly
with `--bundle library/wm-design-system/v2`. The default v1 bundle is unchanged.
The catalog preserves lifecycle metadata and replacement guidance: 166 active
entries, or 167 with `--include-deprecated`; the old bundle still lists 97.

Four split frames now resolve distinct short/tall content zones, short-column
header and source placement, full-width footer, narrow/wide typography and up to
three title lines. Source navigation binds 2–6 caller tabs with stable keys and a
caller-selected active key. Tab placement uses the footer body bottom independently
of source-line reservation. Detailed contracts are in
[frames-and-navigation.md](frames-and-navigation.md).

Revision handling keeps the old agenda photo and pillar connector amendments
on v1. The unchanged typed card APIs remain available in both revisions; changed
metric APIs are not silently reused for v2.

## Native reference review

Artifact directory: `samples/wmds-refresh-20261002/final/`.

- `WMDS-frames-nav-reference.pptx`: 18 slides, editable native text and shapes,
  registered picture objects, plus an editable native chart and workbook.
- `WMDS-frames-nav-reference.pdf`: PowerPoint export using local printing output.
- `foundation.json` and `layout-report.json`: compiled input and prediction report.

Slides 1–8 cover all four split allocations with compact/tall footers, declared
source lines, and one/two/three-line titles. Slides 9–12 cover two/six tabs,
both footers, ordinary frames and split frames. Slides 13–17 are bound examples:
`agenda/schedule-split`, `key-message/stat-split`, `team/roster-split`,
`case-study/exhibit-split`, `bio-full/portrait-nav`. Slide 18 changes the portrait
example to six caller labels and a different active key.

PowerPoint opened the deck without a repair prompt. All 18 exported pages were
visually reviewed for title and body fit, dots, source placement, legal footer,
split gutter, badges, chart, images and navigation alignment. No correction was
needed. The PDF contains IBM Plex Sans and IBM Plex Mono resource names; aliases
were not introduced.

The last code rebuild produces identical slide, master, layout, theme and chart
XML. Package differences are document core properties and the embedded workbook;
the workbook's only internal difference is its document core properties.
The final compatibility edit preserves the old v1 tab-ID acceptance and has no
effect on v2 reference geometry.

Compilation succeeded. No tests were added or run. Generator qualification flags
remain unchanged; manual specimen review is recorded separately in the
[machine-readable receipt](reference-review/slice1.json). This review qualifies
these specimens, not arbitrary content or all 167 catalog entries.

## Next slice

Implement checkbox cells, square quadrant geometry and numbered markers,
named-target annotations, and revised metric content contracts. Six catalog
entries report an explicit pending capability. Complete revision-specific binding
migration and paired source/changed-content examples for all 15 revised designs
before reviewing the remaining added families.
