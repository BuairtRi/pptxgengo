# Data, proof and commercials completion packages

Read-only coverage audit, 2026-10-02. Scope: **27 templates** in `evidence` (10),
`proof` (13) and `commercials` (4). The current typed execution list includes
`stats/four-metrics`; the other **26 remain inventory only**. This is a work plan,
not implementation or qualification evidence. No renderer, source bundle, font,
calibration, test, build or native-capture changes were made during this audit.

The machine ledger is [coverage-data.json](coverage-data.json). It preserves exact
template keys, family source hashes, renderable body node types, authored field
signatures, nested card/table/chart/media anatomy, dependencies, binding work and
source resolutions. Root frame fields are separate from renderable node types;
table-column descriptors such as `num`, `status`, `allocation` and `bullets` are
cell semantics, not independent slide nodes. An authored field's presence does
not imply the Go adapter consumes it.

## Inventory

| Family / template | Completion work |
|---|---|
| evidence: `stats/four-metrics` | Existing bound adapter; full source title-mark parity remains separate |
| `stats/circled-headline` | Measured circle mark + two comparison metric cards |
| `chart/column-full` | Native column chart, workbook, keyed categories, point highlights |
| `chart/column-rail` | Native clustered chart, legend, rail copy |
| `quadrant/subtle` | Native positional diagram, quadrant states/hatch, numbered point/key composite |
| `quadrant/strong` | Strong quadrant fills, moving-point arrows, note block |
| `image-text/square-left` | Pinned square cover crop + bullets with semibold leads |
| `image-text/frame-right` | Pinned focal crop, alt text, caption |
| `status/four-panel` | Three metric-status cards + native numeric/status table |
| `value/capacity-conversion` | Native stacked chart, numbered steps, text arrow, two-line title |
| proof: `case-study/rail` | Image, label/body blocks, three named inverse-rail metrics |
| `case-study/quote` | Image, label/body blocks, metrics, pull quote/attribution |
| `case-study/metrics-exhibits` | Pyramid, narrow metric cards, six question/answer connectors |
| `case-studies/three` | Three keyed case cards, badges, results footer |
| `case-studies/two` | Two keyed case cards, badges, results footer |
| `case-studies/cards-quotes` | Compound narrative + metric-group proof cards, survey cards, quotes |
| `rfp-map/appendix` | Native requirements table, cell bullets, continuation caption/state |
| `confidential/notice` | Header stamp, lead and three note blocks |
| `history/axis` | Period/milestone axis, three label/body notes, callout |
| `outcomes/three-cases` | Explicit three-outcome composite owning panels/metrics/value copy |
| `offerings/three-panels` | Alliance chips, offering bullets, capability panels/local surfaces |
| `deliverables/sample-grid` | Four-deliverable composite, image exhibits or explicit synthetic thumbnails |
| `about/glance` | Quote/icon cards, nested columns, narrow panel metrics, title highlight |
| commercials: `pricing/options` | Checklist cards, recommended state/tag, duration, anchored fee footer |
| `pricing/fixed-fee` | Native fee table/total row + fee summary |
| `pricing/capacity` | Native role table, allocation marks/run-rate footer + fee summary |
| `scope/service-value` | Native dense table, row headers and measured cell bullet lists |

Source-family hashes (whole JSON file, relative to the frozen source root):

- `templates/library/evidence.json`:
  `51b1d2dc40596cacb971de4307ad106b3e31479b4a55fc0be5c9ab5775ee4b02`.
- `templates/library/proof.json`:
  `618ca5821646b797c180bd396de48846711afe336cea7672b77e59d8e497d8c2`.
- `templates/library/commercials.json`:
  `131e69a9e83a819bc26a6e399c1b00e183e89c05134c3f0d736c37e4aafca547`.

## Reuse and the native-object boundary

The WMDS renderer already supplies frozen source loading, grid/frame resolution,
Go typography, native plain/rich text, basic/metric cards, keyed card rows,
standalone metrics, decimal formatting and four typed template adapters. Reuse
those implementations and their precise limits. Existing capability does not
automatically qualify a new source arrangement or narrower frame.

The lower-level writer already has real native capabilities:

- `pptx.Slide.AddTable` builds an `a:tbl` inside `p:graphicFrame`. `TableProps`
  supports explicit column/row sizes, cell fill/borders/margins and spanning;
  `TableCell.TextCells` carries styled text parts. It is a real table, not a
  collection of rectangles. Its existing auto-pagination uses character-width
  heuristics; do not use that as WMDS measured-fit evidence or silently invoke it.
- `pptx.Slide.AddChart` builds a `c:chart`, chart relationships and embedded XLSX.
  `ChartOptions` includes vertical bars (`BarDir=col`), clustered/stacked grouping,
  data labels, fonts, axes, legend and point colors. Reuse this for column data
  charts. An SVG or rectangle reconstruction does not satisfy editable chart
  completion, even if it looks identical in PDF.
- `pptx.Slide.AddImage` supports native pictures; `internal/compose` already has
  pinned local assets, focal cover crops and OOXML `srcRect` semantics. Its canvas
  contracts also provide named overlap/layer handling and bounded pattern fills.
- `pptx.ShapeProps` supports editable custom paths, presets, outlines, arrows and
  patterns. Reuse it for quadrants, pyramids, connectors and the deliberate
  synthetic thumbnail placeholders.

The WMDS integration is still missing for tables/charts/media. Its component
grouping currently matches `p:sp` objects; it does not group `p:pic` or native
table/chart `p:graphicFrame`. Its exact-leading/end-paragraph/font postprocessors
operate on `p:sp` text, not `a:tc` table text or chart-owned text. Its existing
PDF inspection primarily sees standalone text records. Generalize owned-part
inventory, serialization and inspection for these object kinds before claiming
native editability, complete font defaults, or complete text coverage.

## Shared prerequisites

1. **`shared.binding-identity`:** retain exact source-file SHA and persistent
   source-pointer identity overlays. Extend closed typed source structs only for
   declared fields; `titleLines`, `density`, `stamp`, image focus, nested body
   unions and heterogeneous slot `count` metadata must be consumed deliberately.
   Reject unknown fields and duplicate JSON members at every raw-message boundary.
2. **`shared.source-owned-geometry`:** allow declared internal geometry from a
   pinned composite rather than broadening author-controlled grid rules. Frozen
   source uses 16pt chips, 90×117pt thumbnails, y=226/276 and 126/162pt metrics.
   Every exception needs a named owner, source pointer, bounds and explicit fit
   policy. If the source must migrate, record old/new geometry and identity.
3. **`shared.local-surface-context`:** introduce owned panels/regions. Main-body
   text with `on:inverse` must be attached to its actual inverse panel for contrast
   and containment. Matching only the frame surface or pretending the panel is a
   rail is insufficient. Register intentional panel/child overlaps and z-order.
4. **`shared.native-grouping`:** carry pictures and graphic frames as owned
   native objects, including relationships, transforms and IDs. Cells have no
   independent PowerPoint object name; stable `{row_key,column_key}` identities
   live in the adapter/report and map to native row/cell positions.
5. **`shared.inline-marks`:** measured circle, highlight and underscore adornments
   preserve editable text and registered mark assets/geometry. Existing plain
   replacement titles keep unused source emphasis inactive; this is distinct
   from complete source-example mark parity. Do not parse mark syntax as bold.

## Complete work packages and proposed interfaces

These names correspond to dependency IDs in the ledger. Interfaces are proposals
for the next implementation wave, not current APIs. Each package should expose
an explicit measured plan plus native object records; the parent integrates
shared dispatch/source/binding/report changes centrally.

### Native tables and indicators

**`data.native-table`**: `TableSpec` with immutable source column definitions,
`TableRowSpec {key, kind, cells}`, closed cell unions keyed by column, explicit
row/header heights, borders, density and preset. Data row `kind`, `total`, group
rows and spanning run-rate footer cannot collide. Plan every cell's text at its
effective width/margin before `AddTable`; preserve empty cells and numeric right
alignment. Never replace it with native shape grids as the completion path.

Owned templates need these cells: plain text, numeric display/format, enum status,
allocation and keyed bullets. Keep a shared extensible cell union compatible with
other template families' future check/Harvey/rating/maturity/RACI/gauge/icon/tag
cells, but do not implement their semantics by accepting unknown shapes silently.

Source-native typography requires explicit preset variants:

| Cell / part | Source effective typography | Gap from current WMDS tokens |
|---|---|---|
| Standard body | Sans 14/21, 400 | Existing body token reusable |
| Dense body | Sans 12/18, 400 | Existing small token reusable |
| Row header / total body | Same size/leading, weight 600 | Explicit face override and occupied envelope |
| Header | Mono 9/12, weight 600 | Base label token is weight 500 |
| Standard numeric cell | Mono 14/21, weight 600 | Existing number token is 18/24 |
| Dense numeric cell | Mono 12/18, weight 600 | No matching existing number token |
| Status text | Sans small 12/18 alongside square | Mixed cell parts must retain independent style |
| Allocation text | Mono label 9/12, weight 600 alongside 60×8 bar | Explicit cell-part plan, not a gradient string |

Current candidate anchors are keyed by exact font-file SHA, point size and
leading. New header/numeric variants cannot inherit unrelated anchors. Existing
engine fallback must stay explicitly unanchored/conservative until calibrated;
no silent token substitution or typography-file rewrite is justified by this
plan. Table-cell serialization must write actual font names, exact leading,
paragraph defaults, native margins and appropriate end-paragraph properties
inside `a:tc/a:txBody`, outside current shape-only postprocessors.

**`data.cell-indicators`**: typed `StatusCell` and `AllocationCell`, planned native
marks plus explicit text. Use semantic KPI colors and the normative contained
Grounded boundary; validate surface contrast. Allocation is finite in [0,1]; bar
length and rounded percentage are declared computations from that input. Native
tables cannot contain independent slide shapes inside a cell: any owned
decoration must be positioned from exact cell geometry and reported. Keep
decorations synchronized when regenerating; native edit review must state how
manual table resizing affects anchored decorations.

**`data.table-continuation`**: explicit caption/state and declared row-boundary
continuation. Source `continued` is merely a caption; it is not a paging contract.
Requirements rows preserve keys, order, numbering and repeated headers. Start
with fixed-capacity overflow errors; implementing extra-page flow requires an
explicit versioned contract and deterministic source/footer behavior.

### Charts and positional diagrams

**`data.native-column-chart`**: `ColumnChartSpec` with keyed categories and series,
finite raw numeric values, exact alignment/count validation, fixed source mode,
chart title/units and formatting. Single-series category highlights bind category
keys. Source single, clustered and stacked modes map to native column charts with
embedded numeric workbook data; preserve explicit source series colors/outlines,
legend, labels and plot bounds. Chart-owned labels/axes/title have separate native
font fields with writer defaults that often use Arial: set every relevant field
to the original Plex face explicitly.

Frozen `chart.column` explicitly requires a native barChart with embedded
workbook, at most eight categories/four series, 64% slot fill, one value axis and
1pt Grounded mark outlines. Multiple series need direct labels or a legend;
color alone is not sufficient. These component rules remain part of the package
even where a single source template does not exercise every limit.

The exploratory renderer produces SVG and uses positive `niceMax`, raw totals,
category direct labels, last-stack direct series labels and custom 14pt label
separation. Specify native resolutions for these differences. Do not overlay
static labels that become wrong after data editing without disclosing their
ownership/regeneration behavior. If writer extensions are needed for direct
series labels or stacked totals, extend native chart semantics. Keep caches,
series ranges and embedded workbook values synchronized. Default chart formatting
uses one decimal and optional value suffix; it is not automatically the standalone
`FormatNumber` precision. Explicit currency formatting needs matching XLSX/chart
number-format codes and a fixed declared scale, or per-point format support.

**`data.quadrant-diagram`**: `QuadrantSpec` with four fixed named quadrants,
axis titles, keyed normalized [0,1] points, optional destinations and a linked
numbered-key composite. This is intentionally a native positional diagram, not
a statistical chart pretending to have a meaningful workbook. Preserve desirable
fill, undesirable editable hatch, strong inverse state, tags, axes, point marker
contrast and dashed movement arrows. Reject zero-length arrow normalization and
labels that collide; use deterministic measured key spacing. **`data.native-legend`**
shares swatch/label/count planning and semantic series identity.

Frozen `chart.quadrant` explicitly permits drawn shapes for qualitative positions
and requires native scatter when positions represent real data. The current two
owned examples have normalized coordinates and qualitative axis labels. Their
closed binding must declare that semantic mode; it cannot silently convert actual
data into an unlinked shape diagram. If measured-data mode is exposed, extend the
native chart/workbook package with scatter and quadrant background ownership.
The component limit is twelve items; marked desirable/undesirable quadrants carry
explicit text tags. These are separate semantic guarantees from screenshot parity.

### Media, lists and source text variants

**`data.media`**: registered `AssetRef {id,path,sha256,alt}` with explicit crop/focal
point. Preserve `object-fit:cover` through native crop coordinates. The frozen
bundle contains the WM logo files, not `photo-technician` or `photo-warehouse`
bytes; missing photos are a real asset dependency. Supply a pinned local asset
catalog and hash receipt. No network fetching or unrequested photo substitution
at build. An asset-classified synthetic fixture can use an explicitly authored
image; real case studies need their real source assets.

**`data.rich-lists`**: keyed bullets with optional semibold lead, ordered strong
numbers, checklist Boolean items, labeled body blocks and nested columns. Preserve
uniform/in-line typography and inherited roles rather than flattening to strings.
`strongnum` source overrides number to 14/21 (or small 12/18); name that variant.
**`data.label-textblock`** adds source label + body with no title as a versioned
anatomy; current `TextBlockSpec` label_requires_title is not source-complete.

**`data.pullquote`**: authored text/attribution and measured opening quote mark.
Pullquote opening mark uses 60pt/30 leading; quote-card mark uses 48pt/36 leading.
Name and review these variants, including occupied ink outside nominal leading.
Quotation attribution and engagement results must be bound/classified; the
frozen source's sample statements are not evidence of client outcomes.

### Cards, panels and commercial summaries

**`data.advanced-cards`** supplies distinct source anatomies rather than relaxing
the existing card union globally:

- Case card with client/challenge, keyed actions, two-result footer, asymmetric
  right padding and metric badge overhang.
- Compound proof card containing narrative plus metric group and platform label.
- Quote/icon cards with inline icon layout and keyed nested body columns.
- Commercial option card with checklist, labeled duration, recommended state/tag
  and anchored fee footer.
- Explicit outcome-card composite owning existing disconnected overlays.
- Source-specific narrow metric variants, with no blanket minimum-width change.

All badge/mark overhangs belong to the group's declared extent. Badge source
shortens its font when text exceeds five characters; replace that silent rule
with explicit supported badge variants and measured overflow. Pricing fees can
remain authored compound strings; neither fee nor total calculations are inferred.
**`data.icons`** consumes registered source SVG/icon identity and native fallback
policy, with original ink role and explicit sizing.

**`data.fee-summary`**: model tag, total value/label, keyed label/value line pairs
and keyed term bullets in an inverse native group. Its line values use Mono
12/18 semibold, not number18/24. Header/total/terms gaps and divider follow source;
measure label/value tracks and total height. No billing, cumulative, percentage
or financial-model arithmetic occurs unless a future explicit contract asks for
it. The capacity run-rate footer is a native table span with owned label/value.

### Shapes, axes, stamps and exhibits

**`data.shape-diagrams`**: native filled blocks/text arrows, pyramids, group labels,
column heads, connectors and alliance chips, all with named owned text frames.
Pyramid bands use trapezoid custom paths, native ramp colors, label/description
tracks and leader lines. Text-arrow capacity is its shaft rectangle, not its full
outer bounding box. Native connectors preserve path, arrowheads and semantics.

**`data.timeline`**: keyed periods/milestones, normalized positions, label bounds
and native diamond marks. Period data and headline assertions are independent
authored copy: the history source says eight years while listing seven year
labels, so bound content must be consistent without automatic factual repair.
**`data.callout`** preserves source callout padding and number/label roles.

**`data.header-stamp`**: closed source enum with Mono 8pt/.08em, native outline,
y=33 and eyebrow width reduced by180pt. This is an explicit chrome variant,
not overlay decoration that can collide with the unchanged header.

**`data.thumbnails`**: declared four-deliverable composite, each owning a column
head, scope and two exhibits/captions. Real-image mode requires pinned asset bytes.
Synthetic mode preserves the source placeholder bars/rows/diagram/text strokes
and is labeled synthetic. Placeholder `kind:chart`/`kind:table` is not a native
chart/table and must never count as their completion or as deliverable evidence.

## Source resolutions requiring deliberate decisions

Established structural gaps are different from suspected visual fit problems.
The ledger marks fit candidates for later measurement rather than claiming new
Go/native evidence. In particular:

- The previous `3 waves` overflow claim was withdrawn because it ignored authored
  negative tracking. Supplying `3`/`rollout waves` in a reference is a content
  choice, not a source-fit resolution. This audit performed no new fit measurement.
- Existing minimum198pt metric/414pt metric-group guards reject source126pt
  cards,162pt standalone metrics and306pt compound proof groups. Complete these
  through specific declared contexts/variants or explicit source migrations.
- `case-studies/cards-quotes` mixes narrative and metric-group content; its source
  requires a compound anatomy, not an ignored body field.
- Metric starts90pt apart, short status-card heights, dense bullet-cell rows,
  quote marks and long labels need actual occupied-bound/wrap/collision review.
  Source guidance lengths never establish fit. The fixed-fee table's declared
  column widths do sum correctly to558pt.
- Source frame sources use ellipsis/clipping; Go completion must retain full
  bound source copy and reject overflow. No silent ellipsis inheritance.
- Generic source metadata has heterogeneous count hints such as `2 per card`,
  image notes and counts. Extend metadata validation without confusing guidance
  strings with executable repeated-item cardinality.

## Parallel implementation order and acceptance evidence

Start with shared identity/owned-geometry/surface/object infrastructure. Then
independent workers can build native tables, native column charts and media/list
variants in separate files/contracts. Advance to advanced cards/fees and quadrant/
diagram/timeline/exhibit packages after their primitives are integrated. Wire
typed template adapters in family batches; there is no need to hand-author27
independent low-level renderers.

Every completed template needs two bound-content variants, explicit synthetic
classification where applicable, strict rejection cases, source/identity/geometry
receipts and a native review artifact. Counts must distinguish inventory,
implemented bound adapter, source fixture parity, reviewed fixture and qualified
content envelope. Do not graduate a template merely because the Go builder emits
a PPTX or a visual stand-in reproduces a screenshot.

Acceptance evidence includes:

- Exact declared fields, required slots, duplicate rejection, persistent keys
  after reorder and source hash checks; no dropped unsupported nested content.
- Saved native table/chart/picture shapes and relationships. Charts need workbook
  cell values/number formats, series caches/ranges and a PowerPoint Edit Data
  review. Tables need PowerPoint table editing, row/cell styles and spans.
- Font-name/face/style/default inspection inside shapes, table cells and charts;
  explicit unanchored style records where exact calibration is missing.
- Full source/paragraph/run text retention, measured visible-line/occupied-bound
  comparison, fixed frame/panel/group geometry and native visual review of every
  page. Chart axis/legend/direct labels and table cell decorations need dedicated
  records; existing standalone-text PDF checks are insufficient alone.
- Actual asset bytes/hash/focal crop/alt metadata, mark/shape editability, owned
  overlap behavior and a declared receipt for source conflicts or migrations.
- Full normal IBM Plex names in PPTX/PDF. Go performs measurement/generation;
  native development review is not a permanent content-build dependency.

No generic qualified envelope is inferred from this inventory or work plan.
