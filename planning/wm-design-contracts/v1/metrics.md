# Measured metric cards

Contract: `pptxgengo.wmds-metrics.v1`, implemented 2026-10-02 as an additive
card content option in `pptxgengo.wmds-foundation.v1`. Requires the explicit
`wmds-go-foundation.v2` engine. The existing body-card contract remains available.

## Source and supported scope

Frozen `components/v0/components.json` supplies the single and grouped metric
presets under `card.presets`. Frozen `explorations/components.src.html` supplies
the anatomy, gaps, comparison baseline alignment and column proportions.
`tokens/v0/tokens.json` supplies font styles, surface roles and KPI colors.
Source snapshots, bundle pins, fonts and calibration anchors have not changed.

Supported: a primary value/label; an optional comparison value/label; an authored
change string; a status square/text; a grouped primary with one or two keyed
secondaries. These use the existing container, label/title, inline number, band,
padding, edge, surface, frame-zone and native-group machinery. All seven surfaces
remain available, subject to the existing contrast and fit rules.

Deferred: standalone `data.metric`/`data.metric+`, raw-value formatting, automatic
change calculations, targets, ranges, sources inside metrics, person/media/quote
cards, state tags, icons, badges, arbitrary metric counts, composites and slide
bindings. A metric status is a part of this card, not a general status component.
Unknown fields fail. Fixture metrics are fictional, not engagement claims.

## Card content union

Exactly one of these must contain content:

- `body`: the existing nonempty keyed paragraph/bullet array.
- `metric`: a single-metric object.
- `metric_group`: a group object.

Combining them fails as `card.exactly_one_content_kind_required`. Metric typography
comes from its own tokens; supplying `body_size` or `dense` with a metric fails.
Existing title and band requirements still apply. JSON uses snake_case consistently
with the Go adapter, so the frozen renderer's `metricGroup` maps to `metric_group`.

```json
{
  "id": "close-card", "kind": "card", "surface": "outline",
  "grid": "columns", "start": 1, "span": 3,
  "rect": {"y_pt": 144, "height_pt": 180},
  "card": {
    "metric": {
      "value": "5 days", "label": "to close",
      "secondary": {"value": "12", "label": "days today"},
      "status": "on"
    }
  }
}
```

`MetricSpec` requires `value` and `label`. Optional fields are `secondary`
(`MetricValue` with required `value`/`label`), `change` (uniform text string), and
`status` (`on`, `risk`, `off`). Empty/whitespace-only required text fails. Strings
are authored display values, including units and punctuation. No formatter or
arithmetic inference modifies them.

`MetricGroupSpec` requires `primary` (`MetricValue`) and `secondary` (one or two
`KeyedMetricValue` objects, each with `key`, `value`, `label`). Keys follow the
existing ASCII identifier policy and are unique within the group. Reordering
moves each keyed part while preserving its native name. The primary has fixed
identity; it is not another item in the secondary array.

## Typography and content limits

| Part | Style | Surface role | Maximum measured lines |
|---|---|---|---|
| Single primary value, card width below 270pt | stat-sm 36/36 | display | 1 |
| Single primary value, card width at least 270pt | stat 48/48 | display | 1 |
| Group primary value | stat 48/48 | display | 1 |
| Primary label | body 14/21 | secondary | 3 |
| Comparison/secondary value | number 18/24 | display | 1 |
| Comparison/secondary label | small 12/18 | secondary | 3 |
| Change string | label 9/12, uppercase | emphasis | 1 |
| Status text | label 9/12, uppercase | primary | 1 |

These are hard line limits in addition to measured width/height and zone checks.
Numeric values do not wrap. A value that requires wrapping fails; no automatic
font-size reduction, truncation, inserted breaks or fallback font resolves it.
Authored paragraph breaks, including blanks, are preserved and count against the
line limit. The 0.25pt component wrap-boundary exclusion still applies, with the
existing exception for the `number` token's tightly measured boxes.

## Geometry and named native resolutions

Outer cards use the existing 18pt lattice and minimum 198pt width. Groups require
at least 414pt; that width is a minimum, not a promise that every value fits.
Existing padding defaults and accent-edge insets apply before metric layout.

1. **Vertical flow:** each metric text part reserves the greater of authored line
   allocation and predicted occupied extent. A primary label starts 6pt after its
   value. This prevents terminal occupied bounds from colliding with following
   content and does not change the font model. Ordinary body-card flow is unchanged.
2. **Comparison/status footer:** optional comparison, change and status parts stay
   together in that order at the padded card bottom. Their minimum gap below the
   primary label is 12pt (the source's two 6pt gaps around its flex filler). Extra
   fixed-height space expands this gap. Height zero computes the minimum then
   rounds up to 18pt; the footer anchors to that resolved bottom too. A bare
   primary stays at the top and does not acquire an empty footer.
3. **Comparison row:** a .75pt top divider and 9pt padding precede the row. The
   value gets its measured one-line width plus .01pt serialization headroom. A
   9pt gap leaves the remaining width to the small label. Their first baselines
   align using the existing Go anchors; the label may occupy up to three lines.
   The next footer part starts 6pt after the larger row extent.
4. **Group columns:** after a 6pt top gap, the remaining content width divides
   into `1.3fr` for the primary and `1fr` per secondary with 18pt gutters. Each
   secondary's .75pt left divider and 12pt left padding live inside its track.
   Secondary text starts 6pt below the group top; value-to-label gap is 6pt.
   Divider height follows that secondary's content, rather than the full card.
5. **Status mark:** `on` = `#1DD566` / “On track”; `risk` = `#FFC700` / “At risk”;
   `off` = `#F52C00` / “Off track”. Colors come from the pinned KPI tokens.
   The native mark occupies 8×8pt with a contained 1pt Grounded outline and 6×6pt
   color fill. The token requirement for a 1pt data-mark outline takes precedence
   over the exploratory CSS's .75pt outside outline. Its boundary must reach 3:1
   against the card surface through the outline or semantic fill. Hue is never
   recolored automatically. A 6pt gap separates the mark and explicit status text.
   The small status text independently satisfies 4.5:1.

Dividers are decorative separators using the surface `line` role, not status
marks. Their stroke is represented by thin native rectangles. All filled shapes
retain explicit DrawingML no-line settings. Each card is one native editable group,
with owned names such as `close-card.metric.value`,
`close-card.metric.footer.secondary.label`, and
`adoption.metric.secondary.live.divider`.

## Source fit conflicts

The source's first 198×162pt journal metric card has a predicted minimum height
of **169.970pt** when the comparison label receives its remaining row width.
The reviewed reference uses 198pt equal-height cards for that row. A card requesting
162pt with that same content fails; automatic shrinking does not copy a browser
flexbox overflow into the native deck.

The source's 414pt group with two secondaries leaves **134.727pt** for its primary
value after 18pt padding/gutters and `1.3:1:1` division. `$4.2M` at the source
48pt stat token measures **138.950pt**. That content/width is invalid. The reference
uses seven columns (486pt), which leaves 163.091pt. A six-column group with `92%`
or a single secondary remains supported. Source files are retained unchanged;
these are explicit adapter fit decisions, not source migrations.

## Review boundary

The [nine-slide reference](../../../samples/wmds-metrics-20261002/README.md) contains
28 native groups. All 158 text objects and 160 authored paragraphs survive the
package. All 163 predicted visible lines match the local PowerPoint PDF. All pages
were visually reviewed. XML geometry matches within 0.01pt, including group
transforms, fractional track coordinates, colors and text font/style defaults.
The rail also fills its outer edges in PDF and native Slide Show.

This is exact-fixture review, not a general qualified font/content envelope.
No new character capture or font-file identity observation occurred. Measurement
and generation remain Go-only; native export is a development review operation.
No test suite was added or run. The installed release is unchanged.
