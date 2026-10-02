# Standalone metrics and decimal formatting

Contract: `pptxgengo.wmds-data-metrics.v1`. Additive component execution in the
foundation document; requires `wmds-go-foundation.v2`. This contract implements
the bounded `data.metric`, `data.metric+` and `prim.number` formatting subset of
the frozen WMDS bundle. It does not change the separate metric-card contract.

## Inputs

```json
{
  "id": "savings", "kind": "metric", "grid": "columns", "start": 1, "span": 3,
  "rect": {"y_pt": 144, "height_pt": 0},
  "data_metric": {
    "format": {"value": 4250000, "kind": "currency"},
    "label": "run-rate savings", "source": "Illustrative planning scenario"
  }
}
```

`DataMetricSpec` requires `label` and exactly one of authored `value` or `format`.
Both inputs fail rather than allowing the source renderer's format override.
Strings retain their authored text; number formatting applies only to raw input.
Optional fields:

- `change`: `{ "dir": "up" | "down", "text": "authored change" }`. Direction is
  authored, independently of numeric sign or KPI status; no arithmetic inference.
- `secondary`: one or two `{key, value, label}` objects. Values are authored
  strings. Keys are unique ASCII letters/digits/underscore/hyphen, nonempty.
- `target`: authored text, including any desired “Target” prefix.
- `status`: `on`, `risk`, or `off`, with explicit On track / At risk / Off track
  labels and the pinned KPI colors.
- `source`: authored source text. Rendering adds `Source: ` exactly once; callers
  supply the source without that prefix.

Unknown fields fail at strict JSON decoding. Primitive text/style/ink/alignment,
card/textblock payloads, circled values, footnote markup, images and badges are
unsupported. Standalone metrics inherit their frame-zone surface and paint no
background. An explicit node surface must match that zone. A differently colored
panel uses a card or a frame surface, not an invisible surface declaration.

## Number formats

`NumberFormatSpec` mirrors the frozen format object: `value` is a JSON numeric
scalar, or a two-element numeric array for a range. Optional `currency`, `scale`,
`unit`, and `of` are accepted only for their applicable kind.

| Kind | Meaning |
|---|---|
| currency | USD default; explicit USD/EUR/GBP; `$`, `€`, `£`; optional K/M/B scale |
| percent | Fractional input times 100; one decimal below absolute 10%, otherwise whole |
| number | en-US grouping; at most two decimals; optional word unit |
| delta | Explicit `+` or U+2212; at most three decimals; `pts` multiplies fraction by 100 using percent precision |
| range | Ordered two endpoints; `of` is percent or currency; unit appears once at the end |

Currency auto-scale uses absolute raw magnitude: B at 1e9, M at 1e6, K at 1e3,
otherwise none. M/B figures below absolute 10 retain at most one decimal;
remaining currency figures are whole. Explicit K/M/B fixes the scale; empty scale
means automatic. No additional currency or locale is guessed. Word units contain
only ASCII letters and spaces, are trimmed, and have at most 24 characters.

Rounding uses exact decimal rational arithmetic with ties away from zero, followed
by removal of trailing decimal zeros. Negative scalar figures use U+2212. Source
examples therefore include `$4.3M`, `6.5%`, `1,240`, `0.5 FTE`, `−7 days`, and
`−4 pts`. Currency ranges use one shared scale chosen from their largest endpoint,
or the caller's explicit scale. Negative currency ranges are deferred; reversed
endpoints fail. Percent ranges can include negative endpoints. No estimates,
currency conversion, percentage-change calculations or target inference occur.

Numeric input is bounded to absolute 9,007,199,254,740,991, numeric spelling at
most 64 characters, and an explicit exponent between −12 and 15. Invalid JSON,
null, strings, arrays outside ranges, nonfinite values and irrelevant options
fail. This envelope bounds computation and avoids implying arbitrary precision
interop with the exploratory JavaScript. The Go formatter itself preserves the
accepted decimal spelling exactly before rounding.

### Named source resolutions

Pinned `tokens/v0/tokens.json:numberFormats` is normative; the exploratory
`fmtNum` function supplies mechanics where those rules are silent. Three conflicts
are resolved explicitly:

1. K figures below 10 are whole per the token rule (the HTML also gives K one
   decimal). For example, 4,250 USD becomes `$4K`.
2. Negative unscaled numbers use U+2212 per tokens (HTML locale formatting emits
   an ASCII hyphen). All accepted numeric kinds follow the same minus rule.
3. Currency ranges share their scale, selected from the maximum endpoint. HTML's
   first-endpoint selection can display mixed-scale ranges with only one suffix,
   changing the meaning. For example `[500, 10000]` is `$1–10K`, with explicitly
   visible currency rounding, rather than `$500–10K`.

Source tokens, frozen files, font names and typography calibration are unchanged.
Chart workbook number-format codes and bindings are deferred.

## Measured layout and native identity

Outer bounds use the 18pt lattice, minimum 198pt width, and the selected frame
zone. Two secondaries require at least 414pt. Five-up is deferred. Height zero
computes the measured minimum and rounds up to 18pt; fixed heights strictly reject
overflow. Extra height leaves space below the metric; this component has no card
footer or fill. One-line values never shrink to fit.

| Part | Token / role | Line cap |
|---|---|---|
| Primary value | stat 48/48, display | 1 |
| Primary label | body 14/21, secondary | 3 |
| Comparison value | number 18/24, display | 1 |
| Comparison label | small 12/18, secondary | 3 |
| Change text | label 9/12, emphasis | 1 |
| Status text | label 9/12, primary | 1 |
| Target | label 9/12, secondary | 1 |
| Source | source 9/12, secondary | 2 |

Standalone primary values use `stat` even at three columns, matching the source.
Cards' narrower `stat-sm` choice does not apply. The flow reserves the greater
of text allocation and predicted occupied extent, with 6pt gaps. Source lines
add their source 3pt top margin. Comparison rows add the source 6pt top margin,
.75pt divider and 9pt padding, then baseline-align each value and label. One or
two equal-width tracks have 18pt gutters; each value takes its measured width
and a 9pt gap before its label. This bounded track rule replaces the source's
unbounded intrinsic flex-item sizing; insufficient remaining label width fails.

Change direction is a native editable 8×6pt triangle, inside the exploratory
8×8pt slot, with 6pt following gap. Downward change rotates the triangle 180°.
The semantic shape and text independently satisfy contrast. Status/target share
a row with 9pt gaps. The 8×8pt status square has a contained 1pt Grounded outline
per the normative data-mark token (the exploratory outline is .75pt outside).
The square uses the same boundary contrast policy as metric cards. Status text
gets its measured width with 1pt headroom, then the target gets remaining width.
If a target cannot fit on one line, it fails rather than silently moving below.

Every metric is a native group. Names include `savings.value`, `savings.label`,
`savings.change.marker`, `savings.secondary.cost.value`, `savings.status.mark`
and `savings.source`. Reordering secondaries retains their keyed names. Parts
and group bounds appear in the layout report, including triangle geometry and
rotation. Normal IBM Plex Mono and Sans typeface names remain in output.

## Qualification

`DataMetricReference(year)` contains six slides with fictional values, authored
changes, one/two comparisons, decimal formats, ranges, all three statuses,
targets, source lines and a navy rail. It is a review fixture, not a general
qualified text envelope. Go measurement, numeric formatting and generation have
no permanent native-capture dependency. Native export/inspection supplies bounded
development review; no additional font-file identity claim is made.

Native review results and artifact hashes belong in the parent integration
receipt after export. No unit tests were added or run for this slice.


## Combined reference review (2026-10-02)

The [combined reference](../../../samples/wmds-parallel-slices-20261002/README.md)
contains this slice on slides 8–13: 18 standalone metric groups. All 19 pages were
reviewed individually after local PowerPoint PDF export. Package geometry/styles,
paragraph preservation and all 255 predicted visible lines passed the combined
inspection. The [receipt](../../../samples/wmds-parallel-slices-20261002/final/native-review.json)
pins the artifacts and implementation. This review applies to the exact fixture;
per-character bounds, exact native font-file identity and a general content
envelope remain unqualified. No tests were added or run.
