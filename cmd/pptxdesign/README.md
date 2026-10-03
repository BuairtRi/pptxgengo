# WMDS foundation and component prototype

`pptxdesign` implements the first WMDS integration slice in Go. It loads a
reviewed source snapshot, resolves static fonts by numeric weight, resolves grid
and frame geometry, and writes editable PowerPoint objects and a layout report.
Generation has no AppleScript, PowerPoint, browser, Python, or network dependency.

This repository command is separate from the installed local release. Build from
the repository root with Go 1.27.1 or newer:

```sh
go build -o /tmp/pptxdesign ./cmd/pptxdesign
/tmp/pptxdesign inspect --bundle library/wm-design-system/v1
/tmp/pptxdesign reference --bundle library/wm-design-system/v1 \
  --year 2026 --out /tmp/wmds-reference
/tmp/pptxdesign build --bundle library/wm-design-system/v1 \
  --spec samples/wmds-foundation-20261001/final/foundation.json \
  --out /tmp/wmds-custom
/tmp/pptxdesign typography-probes --bundle library/wm-design-system/v1 \
  --out /tmp/wmds-typography
```

The output directory must be new. `reference.pptx`, `foundation.json` and
`layout-report.json` are written after successful resolution. The default year
comes from the clock; use `--year` to reproduce the reference fixture. For `build`,
the document's explicit `year` controls the legal line.

`--source /path/to/wm-design-system` selects a local source tree instead of the
bundled snapshot. Every inventoried file must match its pinned hash. This is an
override for the same snapshot, not automatic migration to newer definitions.
The snapshot manifest and bundled fonts/logos are pinned too. Missing files,
unreviewed changes and unavailable static faces fail with a specific diagnostic.

## Implemented behavior

- All 14 token styles; real static 400/500/600/700 faces; explicit legacy and
  typographic names, PostScript names, file hashes, face indexes and native flags.
- Exact point leading, em tracking serialized once to 0.01 pt, explicit `kern=0`
  at every size, zero paragraph spacing and explicit paragraph-end base styles.
- Original and displayed text, including uppercase transforms, remain in the
  report. CR/CRLF normalize to LF; other content is preserved.
- Twelve-column spans and full-width five-up spans, preserving fractional points
  until EMU serialization. Five-up is rejected on every rail, including nav.
- None/nav/left/right rails with compact/tall footers; one/two-line titles,
  appendix header, one/two-line source zones and no-header/optional-page cover.
- Original positive/reverse WM SVG logos with PNG fallback, legal text, footer
  chrome, whiteboard field and dynamic native slide-number fields. Static chrome
  lives in native slide layouts through the writer's master API; titles, sources,
  navigation and content remain slide objects.
- Whiteboard variant **editable dots**, using 3 pt squares on an 18 pt pitch and
  center-sampled radial opacity. Its quantized extent is 201 × 165 pt. This is a
  declared native variant; no browser pixel identity is claimed.
- Strict JSON input fields; node IDs; grid/zone capacities; outer box lattice;
  selected text role contrast; explicit text overflow errors. No shrink, crop,
  ellipsis, font substitution, or automatic density changes resolve errors.

The candidate v2 engine implements source scene adapters and closed content
bindings for all 97 templates. Native visual review and reusable content
qualification are tracked separately from implementation. See
[library execution](../../planning/wm-design-contracts/v1/library-execution.md)
for catalog, source and bound references, generation receipts and media resources.
Arbitrary font resolution
is supported by the typography code; this WMDS bundle intentionally uses the
families/styles authored by WMDS.

## Foundation document input

The schema is `pptxgengo.wmds-foundation.v1`. See the generated
[18-slide fixture](../../samples/wmds-foundation-20261001/final/foundation.json)
for complete examples. All geometry is in points; column indexes are one-based.

```json
{
  "schema": "pptxgengo.wmds-foundation.v1",
  "year": 2026,
  "slides": [{
    "id": "example",
    "frame": {"rail": "none", "footer": "compact"},
    "eyebrow": "Foundation example",
    "title": "One consistent starting point",
    "nodes": [{
      "id": "message",
      "kind": "text",
      "style": "body",
      "text": "Content stays editable and uses the authored body style.",
      "grid": "columns",
      "start": 1,
      "span": 8,
      "rect": {"y_pt": 144, "height_pt": 72}
    }]
  }]
}
```

Frame defaults: light surface, inverse rail surface, standard density, one title
line, no source lines, required header and visible page number. Panel rail
surfaces are inverse/deep/subtle. Navigation requires explicit `nav` ID/label
entries and an `active` ID; the initial renderer uses 18 pt-wide tabs. `source`
requires explicit `source_lines` (1 or 2). `no_header` excludes title/eyebrow.

Nodes are `text` or `box`. `scope` is body (default) or rail. A text node uses a
token `style`, a surface ink role (default primary), optional left/center/right
alignment and explicit geometry or a grid span. Height may be omitted: the
prototype estimates it, then enforces the assigned zone. Text surface context
must match its zone. Boxes use a named surface, explicit height and optional grid
span; outline uses an inside 1 pt Medium Gray border. Their dimensions obey the
18 pt rhythm, with the sanctioned five-up width exception. `title`, `eyebrow`,
`source`, `wm.*` and `nav.*` IDs are reserved.

This section describes the original foundation authoring API. The candidate v2
adds components and source scene adapters, including native tables/charts,
diagrams, mixed runs, footnotes and hand-drawn marks. Unsupported fields and
unimplemented combinations still fail explicitly.

## Evidence and limits

The [reference deck](../../samples/wmds-foundation-20261001/final/reference.pptx)
was opened in PowerPoint and exported locally as a PDF. All 18 pages were visually
reviewed, every reported text string is present in the PDF, and rendered font
resource names are Plex Sans Regular/Semibold and Plex Mono Regular/Medium/Semibold.
See [native-review.json](../../samples/wmds-foundation-20261001/final/native-review.json).
Those observations apply to this exact reference fixture.

The Go engine is **`wmds-go-foundation.v1`**, distinct from existing
`go-text-prototype.v4`. Its first baseline (0.9 em) and occupied-height estimate
(1.2 em plus requested leading between lines) are provisional. Shaping uses
metric kerning, 1/64 pt advances, per-cluster tracking and excludes terminal
tracking from line width. `liga/clig` are disabled at nonzero tracking; their
native parity is unqualified. No family-specific correction was fitted.

The generated report always sets `powerpoint_verified:false` and
`visually_reviewed:false`; independent review evidence does not rewrite the Go
prediction. Exact native file/instance identity, TextRange bounds and general
wrapping envelopes still need controls. Sans Regular is currently installed as a
variable file sharing the static Regular PostScript name, so the PDF name alone
cannot prove identical font bytes. Only the missing static Sans Medium/Semibold
aliases were installed during this slice; other installed fonts were preserved.

Build and reference generation succeeded. No unit test suite was added or run.

## Typography qualification slice

`typography-probes` writes `typography-probes.pptx` and `probes.json` in a new
directory. It reuses the production text renderer and emits 75 controls for all
14 styles and eight static faces, plus 18 observed rejection records. The controls
are uniform Latin, horizontal, left aligned, with exact leading and no margins.

Prepare a capture packet from the repository root:

```sh
python3 scripts/prepare-wmds-typography.py --packet /tmp/wmds-typography
```

Native capture is a separate read-only calibration operation. Open the exact
packet deck without editing it, then run its `capture.sh` with the packet directory
argument from a Terminal where PowerPoint automation works. Export PDF locally
with **Best for printing** and use the pinned Swift readers for selection/baseline
evidence. `pdf-export.json` must associate the exact saved deck and PDF hashes;
it is an explicit review receipt, not inferred by the analyzer. See the completed
[control packet and instructions](../../samples/wmds-typography-20261001/README.md).

`scripts/compare-wmds-typography.py --packet PACKET --out NEW.json` compares PDF
line breaks/baselines and keeps width comparisons diagnostic. Add `--capture-dir`
for receipt-verified native frames, margins, effective character styles, paragraph
leading, line breaks, occupied height and overflow. `--require-parity` exits with
status 2 while parity remains failed or unqualified. It never grants a broad
qualified envelope from this fixture or changes authoring flags.

The native PDF comparison found **68/75 matching line breaks**, seven boundary
failures, first-baseline errors up to **7.2 pt**, and matched multiline pitches
within **0.12 pt**. All expected Plex PostScript names were rendered. The current
engine therefore remains **unqualified**; native character capture and exact
font file/instance evidence are still required before a new calibrated engine.
Package inspection also found one empty middle paragraph with incomplete
paragraph-end font defaults. The next engine must repair that case explicitly;
the original v1 control packet is retained unchanged.

The [complete native capture](../../samples/wmds-typography-20261001/NATIVE_RESULTS.md)
now supplies all 75 character-bound observations. Frames, margins, effective
character attributes and paragraph spacing match in every control. Seven wrap
boundaries and the occupied-height model fail. A trailing hard break is also
serialized as a literal newline in one run, losing the intended second paragraph.
The recovered receipt retains that failure; diagnostic capture completion is
separate from qualification. No v1 recapture is needed before correction.

## Opt-in v2 candidate

The corrected typography candidate is selected explicitly:

```sh
/tmp/pptxdesign reference --engine wmds-go-foundation.v2 \
  --year 2026 --out /tmp/wmds-v2-reference
/tmp/pptxdesign typography-probes --engine wmds-go-foundation.v2 \
  --out /tmp/wmds-v2-controls
```

`inspect` and `build` accept the same `--engine` flag. V1 remains the default.
V2 includes terminal tracking in fitting, retains tracking precision, rounds
Latin glyph advances to 1/8 pt, disables metric kerning and enables standard
ligatures regardless of tracking. Explicit paragraphs preserve trailing and
empty breaks with complete end defaults. Baseline/terminal-height estimates use
17 pinned v1 anchors; unknown font/size/leading combinations are clearly reported
as uncalibrated. Native qualification flags remain false.

[New v2 PDF evidence](../../samples/wmds-typography-v2-20261001/README.md)
matches 139/139 line-break controls and all visible baselines within 0.12 pt.
Original v1 native development replay matches 75/75 wraps and 74/74 comparable
heights after these corrections; that fitted evidence is not independent validation.
The [independent v2 character capture](../../samples/wmds-typography-v2-20261001/NATIVE_RESULTS.md)
now passes all 139 controls. Occupied-height error is below 0.0001 pt and native
line-advance error is at most 0.132 pt. Exact font file/instance access and a
general typography envelope remain unqualified. See the
[calculation contract](../../library/wm-design-system/typography-v2-candidate/README.md)
and [reviewed foundation candidate](../../samples/wmds-foundation-v2-20261001/README.md).

Normal IBM Plex names are retained in PPTX and PDF. Font hashes and file/instance
metadata distinguish measurement identities internally; no prototype aliases
are introduced. The current adapter records independent CoreText family and
PostScript resolution plus installed file/variation metadata without changing
fonts. Actual PowerPoint file access remains unobserved.

The WMDS-only v9 capture adapter includes empty paragraphs, paragraph font defaults
and paragraph contents; null native bounds mean no measurable visible characters.
Raw content and bounds are always retained. The v2 comparison uses a visible
non-whitespace character union and excludes empty terminal snapshots; this does
not rewrite the preserved v1 adapter or evidence.

**Runtime:** text measurement, shaping, wrapping and generation execute in Go.
PowerPoint/AppleScript capture is a development qualification operation. New
fonts/features or a changed rendering environment can require reference capture;
normal deck generation does not depend on it.

### Captured terminal paragraph collection semantics

The v9 capture preserves all authored hard breaks in native range content. In
this environment, its native paragraph collection omits the final zero-length
entry for the four controls ending in a return. The current repository analyzer
checks that behavior against exact native text, concatenated paragraph snapshots
and the independently inspected saved XML. It retains direct authored/native
count differences as diagnostics and never trims or fabricates content. End
font defaults for the unenumerated terminal entry remain package evidence.

The prepared v2 packet's original analyzer and initial four-count failure remain
immutable. Its corrected `comparison-v2.json` was produced by the repository
analyzer and includes its own source hash. New prepared packets will carry the
corrected analyzer. Native automation remains a development qualification step.


## Measured text blocks and basic cards

This slice is opt-in with `--engine wmds-go-foundation.v2`. Generate the 12-slide
component reference with:

```sh
/tmp/pptxdesign component-reference --engine wmds-go-foundation.v2 \
  --year 2026 --out /tmp/wmds-components
/tmp/pptxdesign build --engine wmds-go-foundation.v2 \
  --spec samples/wmds-components-20261001/final-v4/components.json \
  --out /tmp/wmds-components-custom
```

`component-reference` writes `component-reference.pptx`, `components.json` and
`layout-report.json`. `build` retains its usual `reference.pptx` and
`foundation.json` filenames. PowerPoint/PDF review evidence is a separate receipt.

Add a typed component node to the existing document's `nodes` array:

```json
{
  "id": "message-card",
  "kind": "card",
  "grid": "columns", "start": 1, "span": 4,
  "rect": {"y_pt": 144, "height_pt": 216},
  "surface": "subtle",
  "card": {
    "label": "Review process",
    "title": "Assigned owners",
    "body": [{"key": "copy", "p": "Every exception has an owner and a review date."}]
  }
}
```

For a text block, use `kind: "textblock"` and
`textblock: {"title": "Assigned owners", "body": "Every exception has an owner."}`.
A label is optional and requires a title. Omit height or use zero for measured
height rounded up to the 18pt module. Fixed height is a hard capacity.

Go callers use `Document`, `Node`, `TextBlockSpec`, `CardSpec`, `BodyBlock`,
`BulletItem`, `CardBand`, `CardEdge`, and `BuildWithEngine`. The layout report
records resolved component bounds, required height, density, owned parts and all
text measurements. Each component becomes a native editable group.

See the [complete component contract](../../planning/wm-design-contracts/v1/components.md)
for supported fields, defaults, spacing, widths and limits. Unknown presets and
fields fail. The prototype explicitly rejects uncertain line breaks within
0.25pt of a wrap boundary. No automatic shrinking or dense-body fallback occurs.

[Reviewed PPTX/PDF and evidence](../../samples/wmds-components-20261001/README.md)
cover 34 component instances, all seven surfaces, bands, edges, bullets, numbers,
hard/empty paragraphs, density and measured heights. The scope remains bounded;
person/media/quote cards and template bindings remain future work. The later
parallel slice below adds keyed card rows, standalone metrics and direct mixed runs. The installed release has not changed.


## Measured metric cards

The next additive slice uses the same Go renderer and native grouping:

```sh
/tmp/pptxdesign metric-reference --engine wmds-go-foundation.v2 \
  --year 2026 --out /tmp/wmds-metrics
/tmp/pptxdesign build --engine wmds-go-foundation.v2 \
  --spec samples/wmds-metrics-20261002/final/metrics.json \
  --out /tmp/wmds-metrics-custom
```

`metric-reference` writes `metric-reference.pptx`, `metrics.json`, and
`layout-report.json`. Choose exactly one card content kind: `body`, `metric`, or
`metric_group`. Single metrics require display-string `value` and `label`, with
optional `secondary: {value, label}`, `change` string, and `status: on|risk|off`.
Groups require `primary: {value, label}` and one or two uniquely keyed secondary
objects `{key, value, label}`. Values remain authored strings; formatting and
calculated changes are deferred.

Go callers use `MetricSpec`, `MetricValue`, `MetricGroupSpec`, and
`KeyedMetricValue` inside `CardSpec`. Each metric record declares
`pptxgengo.wmds-metrics.v1`. All existing card surfaces, headers, bands, edges and
padding rules remain available. Metric-specific text sizes come from the frozen
tokens; `body_size` and `dense` are unsupported for metric cards.

The [metric contract](../../planning/wm-design-contracts/v1/metrics.md) defines
footer anchoring, comparison baseline alignment, grouped columns, status colors,
line limits, source fit conflicts, and deferred fields. The
[reviewed reference](../../samples/wmds-metrics-20261002/README.md) covers 28 groups
on nine slides, with 158 preserved text objects and 163 matching visible PDF lines.
This evidence applies to that fixture; generic build qualification flags stay false.
No new font capture or calibration is required to generate metric cards.


## Card rows, standalone metrics and mixed text runs (2026-10-02)

Three additive contracts now execute in the opt-in v2 Go renderer:

| Node kind / payload | Go types | Supported scope |
|---|---|---|
| `cardrow` / `card_row` | `CardRowSpec`, `CardRowItem` | 2–4 keyed equal cards, 18pt gutters, common measured height, inline/band numbering, native nested groups |
| `metric` / `data_metric` | `DataMetricSpec`, `NumberFormatSpec`, `DataMetricChange` | Standalone values, authored changes/targets, one/two keyed comparisons, KPI status, source, decimal formatting |
| `richtext` / `richtext` | `RichTextSpec`, `RichParagraphSpec`, `RichRunSpec` | Keyed paragraphs/runs, real 400/500/600/700 weights, per-run surface inks, combined wrapping, explicit editable paragraph defaults |

```sh
/tmp/pptxdesign parallel-reference --engine wmds-go-foundation.v2 \
  --year 2026 --out /tmp/wmds-parallel
/tmp/pptxdesign build --engine wmds-go-foundation.v2 \
  --spec samples/wmds-parallel-slices-20261002/final/slices.json \
  --out /tmp/wmds-parallel-custom
```

`parallel-reference` writes `parallel-reference.pptx`, `slices.json` and
`layout-report.json` (19 slides). Individual commands are `card-row-reference`
(7 slides; `card-rows.json`), `data-metric-reference` (6; `data-metrics.json`), and
`rich-reference` (6; `rich-text.json`); each writes its correspondingly named PPTX
and the shared layout report. All require explicit v2 selection.

The contracts specify the JSON examples, source resolutions and limits:
[card rows](../../planning/wm-design-contracts/v1/card-rows.md),
[standalone metrics](../../planning/wm-design-contracts/v1/data-metrics.md), and
[mixed runs](../../planning/wm-design-contracts/v1/rich-text.md). `FormatNumber` is
also available to Go callers; existing metric-card values remain display strings.

The [combined reference and review](../../samples/wmds-parallel-slices-20261002/README.md)
contains 9 row groups, 28 nested cards, 18 standalone metric groups and 15 rich
text objects. Local PowerPoint PDF review matched all 255 predicted visible lines
and preserved 242 paragraphs, including blank/trailing entries. All 19 pages were
visually reviewed; the navy rail perimeter has no white edge.

These results apply to the saved fixture. Mixed runs retain `native_qualified: false`;
14 run/style combinations in this fixture use conservative uncalibrated vertical
estimates. Direct rich text is supported; rich card/bullet bindings, per-run size
and family changes, footnotes and hand-drawn marks remain deferred. General native
font-file identity and arbitrary-content parity remain unqualified. Go-only
generation needs no per-deck capture. Source snapshots, calibration and the installed
release are unchanged. No tests were added or run for these slices.


## Bound WM slide templates

Four source-pinned templates now accept content through closed named slots:
`cards/3`, `cards/4`, `stats/four-metrics`, and `takeaway-rail/metrics-rail`.
The source compiler resolves their exact frozen variants and preserves geometry,
styles, surfaces, numbering and chrome. All visible content is supplied explicitly.

```sh
/tmp/pptxdesign templates --engine wmds-go-foundation.v2
/tmp/pptxdesign template-reference --engine wmds-go-foundation.v2 \
  --year 2026 --out /tmp/wmds-templates
/tmp/pptxdesign template --engine wmds-go-foundation.v2 \
  --spec samples/wmds-templates-20261002/final/template-content.json \
  --out /tmp/wmds-bound-deck
```

`templates` lists only the four executable adapters, with source hashes, persistent
node identities and advisory character guidance. `template` requires a
`pptxgengo.wmds-template-document.v1` input with explicit legal year, unique slide
IDs, template keys, content classification (`synthetic_example` or `supplied_content`),
and `values`. See the [contract and example](../../planning/wm-design-contracts/v1/template-execution.md)
and [structural schema](../../planning/wm-design-contracts/v1/template-content.schema.json).

`template` writes `bound-templates.pptx`; `template-reference` writes
`template-reference.pptx` (eight slides). Both write `template-content.json`,
`compiled-document.json`, `binding-report.json`, and `layout-report.json`.
The bound content file is the authoring input; the compiled document is the
resolved foundation IR for inspection, not a template customization API.

Go callers use `BoundDocument`, `BoundSlide`, `DecodeBoundDocument`,
`TemplateCatalog`, and `BindTemplates`, then `BuildWithEngine` with v2. Typed
content includes `BoundCardsContent`, `BoundStatsContent`, `BoundRailContent`,
`BoundMetric` and `TextContent`. Duplicate properties, wrong-case/unknown names,
missing/empty content, invalid unions and unsupported fields fail. Card arrays
require exactly three/four unique keys; metrics are fixed named slots. Body text
slots allow strings or explicit rich paragraphs/runs. Geometry/style mutation,
markup, automatic continuation, extra cards and missing-source fallback are
unsupported. Measured fixed capacity applies after binding; no font shrinking.

[The reviewed reference](../../samples/wmds-templates-20261002/README.md) has 70 slot
assignments, 30 editable groups and 133 matching native PDF lines. All eight
pages were reviewed individually. Pair comparisons retain fixed source geometry
and styles; both right-rail perimeters are opaque navy. Binding provenance stays
unqualified for arbitrary content and exact native font-file selection. Generation
is Go-only; no per-deck character capture is required. No tests were added or run.

## Refreshed library: split frames and caller navigation

Select round-4 source explicitly with `--bundle library/wm-design-system/v2`.
The original v1 bundle remains the default. Source revision and typography engine
are separate; both revisions use the existing candidate Go font engine for scenes.

```sh
pptxdesign library-catalog --bundle library/wm-design-system/v2 \
  --engine wmds-go-foundation.v2
pptxdesign frame-reference --bundle library/wm-design-system/v2 \
  --engine wmds-go-foundation.v2 --out /tmp/wmds-frame-reference
pptxdesign library-reference --bundle library/wm-design-system/v2 \
  --engine wmds-go-foundation.v2 \
  --template-keys agenda/schedule-split,key-message/stat-split,team/roster-split,case-study/exhibit-split,bio-full/portrait-nav \
  --out /tmp/wmds-split-templates
```

The catalog reports 166 active definitions; `--include-deprecated` also includes
`from-to/rows`. Metadata records source revision, template revision, lifecycle,
replacement and pending capabilities. Definitions are not evidence that all 167
fixtures render or that new content fits. The focused `--template-keys` option is
also supported by source references and sweeps; it is exclusive with `--family`.

Nav templates require `values.nav` with 2–6 `{key,label}` items and an `active` key.
Callers choose labels, ordering, tab count and active selection. Keys remain stable
native identities; active selection must name an item. Geometry/style changes are
not content slots. See [the frame/navigation contract](../../planning/wm-design-contracts/v2/frames-and-navigation.md).

The v2 first two slices implement split/nav infrastructure, the new component
semantics and revised bindings. The two old typed card-row APIs remain valid
in v2; the four historical typed APIs remain available through the v1 bundle.

Slice 2 implements checkbox cells, square quadrant/numbered marker modes and
named-target annotations. Revised stats and rail templates require the v2 library
slots/keys contract; the earlier typed APIs remain on the v1 bundle.
See [component contracts and native review](../../planning/wm-design-contracts/v2/components-and-revised-bindings.md).

Slice 3 covers all 70 added designs with source/meaningful-content pairs in a
140-slide native reference. Family reproduction scripts and the integration
command are in [the completion note](../../planning/wm-design-contracts/v2/slice3-completion.md).
V2 navigation uses the dedicated 8pt semibold source style; table group boundaries
remain structural, and outline surfaces retain their inset border. Native review
of these pairs does not qualify arbitrary input.

Slice 4 completes full-catalog regression review: all 167 designs have accepted
source/meaningful-alternate specimens (334 slides). The final library deck has
one source example per design. See [integrated completion](../../planning/wm-design-contracts/v2/slice4-completion.md)
for the native receipts, named content-capacity amendments and portable packaging.

The local.6 package exposes `pptxgengo design`, defaulting to bundled v2 source
and the candidate Go engine, and `pptxgengo catalog --design-system`. The gallery
provides native previews and editable one-slide values/foundation documents.
Original font names remain intact. Normal builds use packaged font/artwork bytes
in Go; PowerPoint capture is calibration/review evidence, not a build dependency.
Standalone `pptxdesign` keeps its earlier defaults. `--bundle v1` / `--bundle v2`
aliases resolve relative to the installed release (or repository working directory
for standalone use). `asset-catalog` lists the complete hash-pinned artwork registry.
