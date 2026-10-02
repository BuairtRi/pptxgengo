# Source data scene adapters — v1

This additive execution contract covers `scene_cards.go`, `scene_tables.go`,
`scene_charts.go`, and the table paragraph pass in `scene_table_xml.go`. It does
not broaden the independent `CardSpec`, metric, grid, typography, or four initial
template APIs. Frozen source files and bundled fonts remain unchanged.

## Inputs and identity

The source compiler supplies strictly decoded nodes, a surface, source pointer,
and exact array-pointer key overlays. Unknown fields, duplicate JSON names,
case aliases, invalid unions, unsupported cell/chart variants, and measured
overflow fail. A non-nil bound key overlay requires every consumed caller array
to have keys. Ordinal keys are permitted only when the fixture has no overlay.
Native cell identity is `{table_id,row_key,column_key}`; row order remains the
array order. Source table columns are immutable semantic descriptors, not a
repeated content array.

A card-row common `card` definition may contain shared default body arrays.
Their supplied keys are projected under each item unless that item replaces
the corresponding field. A metric's singleton secondary object is a named
slot, not an artificial caller array. Its generated identity is `secondary`.
Generated RACI legends and target-series values likewise use fixed identities.

## Cards and standalone values

The source adapter retains bounded card geometry and compound anatomy:
contextual narrow/title-only cards; header labels/numbers/bands/edges; typed
paragraphs, labels, bullets, nested bullets, checklists and body columns; metric
and metric-group content; authored changes/status/secondary comparisons; quote,
person, biography, case-result and fee sections; selected/deemphasized/placeholder
states; media, icons and overhanging badges. Media and artwork use the separate
pinned-asset adapter. Containers and editable subparts group under stable IDs.

`value` strings are authored copy. Optional typed `format` values use the existing
strict Go `FormatNumber` contract. Formatting does not calculate fee totals,
financial returns, changes, or targets. Nested card metrics reject direct-node
geometry/surface/circle/source/target fields rather than silently ignoring them.
Direct metrics and callouts have their own closed source node adapters.

Font family, size, leading, case and tracking come from source tokens. Explicit
source variants such as the 48pt opening quote and conditional 13pt badge value
retain their authored sizes; relative tracking is recalculated at changed size.
Unanchored style combinations are reported and remain unqualified. Source
`[[marks]]` use the primitive rich-text/underscore planner, preserving token
fonts and measured mixed runs. Footnotes require the supplied notes context.

## Fixed source spacing resolutions

Generation found authored source cards whose text and gaps exceed their inside
capacity. The following deterministic **source-layout.v1** resolutions preserve
outer geometry, all copy and font sizes. Each activated rule is recorded in the
scene warnings. These are fixed adapter presets, not copy-dependent fitting.

| Resolution ID | Exact scope | Executable decision |
| --- | --- | --- |
| `dense-card-gap-3` | `bodySize=small`, or fee cards containing a checklist, with 6pt stack gap | Use 3pt stack gap. Checklist item gaps and typography remain unchanged. |
| `title-only-visible-envelope` | A title with no band or following body/compound/placeholder content | Exclude the invisible trailing 6pt title margin from the occupied-fit check; retain text coordinates. |
| `compact-side-heading-pad-12` | 90pt tall card, 18pt padding, heading title and side icon | Use 12pt padding. |
| `compact-quote-pad-9` | 162pt tall quote card with 18pt padding | Use 9pt padding; retain the 48pt/36pt quote glyph and conservative Go allocation. |
| `compact-metric-pad-9` | 126×126pt metric card with 12pt padding | Use 9pt padding; retain stat-sm and body labels. |
| `compact-body-title-margin-3` | `bodySize=small`, `titleStyle=body`, card height ≤108pt, padding ≤12pt, with following body content | Use a fixed 3pt title margin instead of 6pt. Retain title/body font sizes and outer geometry. |

These resolve the generation conflicts observed in the initial library sweep,
including method, compact challenge/use-case, hub, pricing options, capability,
firm quote, five-column heading, and compact case-result cards. They do not
establish native visual acceptance. The earlier `3 waves` overflow claim remains
withdrawn; no new tracking-based fit evidence is asserted here.

## Native tables

A table is a native PowerPoint table with explicit column widths and row
heights, `AutoPage=false`, closed dynamic row fields, and real merged heading,
group and run-rate cells. Source light/dark headers, highlight column, total
rows, row headers, dense typography, continuation caption and RACI legend are
retained. Current source presets include fees, capacity, ratecard, requirements
and RACI. No spreadsheet table lookalike substitutes for the native table.

Native cell text supports text, authored or typed-formatted numbers, delta,
status, allocation, maturity, tag, bullets and icon labels. Rating, Harvey,
gauge, RACI and check semantics use editable native shapes anchored to measured
cell geometry; status/allocation/icons similarly have owned native adornments.
True/false check values, bounded indicator domains and enum spellings are
validated. Empty native cells have explicit normal Plex defaults too.
Native cell footnote runs preserve the supplied notes context. Hand-drawn
`[[marks]]` in native cells are rejected explicitly; current library table
fixtures do not require them. They are not silently removed from bound text.

Bullet cells use real native square-bullet paragraphs: 3pt square, 12pt indent,
18pt small-text leading, and 3pt gaps between paragraphs, with wrapping measured
in the remaining width. The paragraph pass writes explicit font/default/end
properties, exact leading/tracking, rich baseline shifts, and no-autofit. It
preserves cell margins, merges, fills and borders and forces the planned row
heights. Source points are converted to native writer inches where required.

`CellRecords.Column` indexes the original row's native cell array. Inserted
horizontal merge dummies have explicit blank formatting, do not consume a
caller cell record, and retain their merge flag. Table cell paragraph records
are not standalone text shapes. Manual cell resize or text edits do not
automatically reposition independent adornments; rerender bound input to
recalculate them.

## Native charts and qualitative quadrants

Column/bar/line/pie/doughnut/scatter charts create native chart objects and
embedded XLSX data. All writer text roles specify normal Plex typefaces.
Source single/clustered/stacked/percent column modes, palette references,
single-series category highlights, nice axis maximum, per-series line dash
styles and targets are retained. Target lines are native derived chart series.
Source chart title/units/source are separate editable text. Category display
case follows source label styles.

Stacked totals and direct end labels are explicitly owned native text overlays.
Binding and rerendering synchronizes them with native data. PowerPoint's manual
**Edit Data** does not automatically update these external annotations. No claim
of native plot/label pixel parity is made before review.

A source quadrant defaults to qualitative normalized positions and produces
editable native quadrants, hatching, state tags, numbered key or direct labels,
point marks and destination arrows. This is not a numerical scatter claim.
The additive `positionMode=data` adapter creates a native scatter/workbook for
normalized numerical positions; keyed numerical legends are currently rejected.
Every marked quadrant requires a state tag. Point count and coordinate domain
remain bounded.

Nullable native series values and native regression/trend lines fail named
unsupported errors; neither is coerced to zero or a visual chart substitute.
They are not present in the current 97 source templates. Number formatting
uses the strict Go formatter for authored values; chart-owned labels use
explicit Excel format codes, whose automatic scaling and native appearance
remain unqualified. Range chart labels are unsupported.

The initial aggregate native open required repair at the line-chart handoff
slide. ZIP inspection isolated the removed chart and workbook while the data
and workbook formulas matched the planned series. The shared writer was
corrected to follow the [official Open XML SDK chart schema](https://raw.githubusercontent.com/dotnet/Open-XML-SDK/main/data/schemas/schemas_openxmlformats_org_drawingml_2006_chart.json):
line grouping is required, line-series markers precede data labels,
`invertIfNegative` belongs to bar series, and a two-dimensional line chart has
two axis IDs. The scene adapter supplies the valid right-side line label
position before cloning per-series options. The isolated handoff slide
subsequently opened natively without repair; the full-library open remains
separate acceptance evidence. The repaired file is diagnostic evidence only.

## User review corrections v1

The pass02 native PNG review identified these adapter defects. Source JSON,
authored copy and font sizes are retained. New generated receipts record each
activated adapter rule; corrected native visual review is still required.

| Rule | Reviewed scope and cause | Correction |
| --- | --- | --- |
| `bio-tag-width-reserve` | Bios pages11/12; tag text boxes were exactly the Go shaped advance, causing native final-glyph wrapping. Initials were readable. | Add fixed2pt text width inside each tag, retaining6pt margins and authored9pt typography. |
| `featured-tag-width-reserve` | Pricing/options final page72; the RECOMMENDED label broke before its final D below the short featured tag. | Add fixed2pt native text width inside the right-anchored featured tag, retaining6pt margins and source font/copy. |
| `allocation-full-height` | Pricing/capacity page62; 0.75pt top/bottom inset left a visible white seam inside an8pt wrapper. | Fill the full8pt height and exact allocation fraction of60pt width; paint a transparent outline above it. |
| `native-tag-transparent` | Roles/by-phase page73; opaque tag adornments painted above the native table and hid Owner text. | Keep native cell content, make the outline transparent, and add fixed2pt reserve. |
| `quadrant-rectangular-plot` | Quadrant pages66/67; forcing a255pt square wasted authored horizontal space. Tag boxes also wrapped at exact advance. | Use the allocated rectangle:522×267pt strong and516×267pt keyed subtle; reserve24pt for axes and306pt for the enabled key; add2pt tag text width. Normalized positions retain their values. |
| `circled-metric-clearance` | Circled headline page76; tightly cropped irregular artwork intersected the value's upper edge. | Center the ring on the measured visible envelope; enclose content at62% of ring width and55% of ring height, with at least12pt clearance. Move the following label below the ring with6pt gap. Fixed-height overflow is rejected. |
| `stacked-zero-labels-hidden` | Capacity conversion page87; native zero-segment labels crowded the total100 values, and optional decimals displayed a trailing dot for integer data. | For unformatted integer source series use `#,##0`; stacked integer series use `#,##0;-#,##0;`, whose empty third section hides zero labels. Preserve the external totals and workbook values. Decimal data is not rounded by this rule. |

Quadrant outlines, hatch segments and movement arrows now use native custom
geometry. The earlier `line` preset ignored the supplied points, rendering a
diagonal perimeter and incorrect directed endpoints. Custom paths preserve the
authored points and open arrowheads. This is a geometry correction, not a
raster substitute.

The focused eight-slide corrected source document and artifact live under
`samples/wmds-library-20261002/user-review-data-correction01/`. They preserve
source hashes, bindings and synthetic content classifications. The phrase
"near line80" cannot be uniquely located in the page87 image; the observed
top-of-chart zero/total collision is addressed, and its native result remains
pending review.

## Reference and acceptance evidence

`SceneDataReference(year)` returns a nine-slide synthetic `Document`: contextual
cards; compound columns/quote/case; metric cards/groups; native fees; native
indicators; merged headings/bullets; native clustered columns; native dashed
line series; qualitative quadrant. This fixture contains illustrative values
and people, not client facts.

Compilation is an implementation check only. The parent library source/binding
sweeps supply generation receipts; PowerPoint capture, object/workbook inspection,
PDF rendering, table/chart text and visual review remain separate acceptance
evidence. Generic content qualification is false.
