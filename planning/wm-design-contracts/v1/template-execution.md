# Bound template execution v1

Executable contract: `pptxgengo.wmds-template-bindings.v2`.
This is a narrow executable API for four frozen templates. Input schema:
`pptxgengo.wmds-template-document.v1`. It is additive to the foundation document
and requires the opt-in `wmds-go-foundation.v2` engine. It does not turn the
historical v1 planning contract's general binding-schema proposal into an
implemented binding engine, and does not redefine that contract or profile.

## Closed input

```json
{
  "schema": "pptxgengo.wmds-template-document.v1",
  "year": 2026,
  "slides": [{
    "id": "sample-approach", "template": "cards/3",
    "content_kind": "synthetic_example",
    "values": {
      "eyebrow": "Illustrative approach",
      "title": "Three moves for a sample close",
      "cards": [
        {"key": "diagnose", "title": "Diagnose", "body": "Map the sample calendar."},
        {"key": "design", "title": "Design", "body": "Agree the checks and owners."},
        {"key": "deploy", "title": "Deploy", "body": "Start with a bounded pilot."}
      ]
    }
  }]
}
```

`BoundDocument` holds `schema`, explicit legal `year`, and `slides`.
`BoundSlide` holds unique `id`, exact `template` key, explicit `content_kind`, and
`values` as a closed object. The built-in reference uses only `synthetic_example`;
no fictional source copy is promoted to an observed client claim. Callers must
classify it as `synthetic_example` or `supplied_content`.

All listed content fields are required. Omitted values do not fall back to
the frozen template's fictional examples. Empty required text, unknown fields,
duplicate JSON member names, invalid types, conflicting alternatives and trailing
JSON documents fail. Validation applies recursively to raw `values`, metric
format objects and rich text; a raw-message boundary does not permit arbitrary
extra fields. No coercion, guessed defaults or arbitrary formatting code occurs.

| Exact template key | Required `values` fields |
|---|---|
| `cards/3` | `eyebrow`, `title`, `cards`: exactly three `{key,title,body}` objects |
| `cards/4` | `eyebrow`, `title`, `cards`: exactly four `{key,title,body}` objects |
| `stats/four-metrics` | `eyebrow`, `title`, `metric1`, `metric2`, `metric3`, `metric4`, `support`, `source` |
| `takeaway-rail/metrics-rail` | `eyebrow`, `title`, `metric1`, `metric2`, `support`, `railEyebrow`, `railHeading`, `railBody`, `source` |

Camel-case rail keys are intentional and match this closed API exactly. Card
titles and bodies are strings. Each named metric is a closed object requiring
`label` and exactly one of `value` (authored string) or `format`
(`NumberFormatSpec`). Card arrays do not accept metric, band, edge, padding or
other component customization. Named metric slots do not accept targets,
comparisons, status, change or local source lines in this slice.

Only `support`, `railHeading` and `railBody` accept either a plain string or a
`RichTextSpec` object. Other slots remain uniform strings. No markup parser
converts `[[phrase]]` to emphasis; hand-drawn marks and `[^n]` footnotes are
rejected. Rich text uses explicitly keyed paragraphs/runs and the existing
mixed-text contract. Each slot inherits its immutable source typography token,
including family, size, leading, tracking and case. A run may choose a supported
static numeric weight and surface ink subject to the rich-text contrast rules;
it cannot choose a different font family, size or layout.

```json
"support": {
  "paragraphs": [{"key": "support", "runs": [
    {"key": "prefix", "text": "A "},
    {"key": "owner", "text": "named owner", "weight": 600, "ink": "emphasis"},
    {"key": "suffix", "text": " reviews each exception."}
  ]}]
}
```

`source` is a plain footer-source string. The caller supplies its full displayed
copy; this slot does not add the standalone metric component's `Source: ` prefix.

## Persistent identity and repetition

The compiler resolves an exact template key from the frozen catalog and validates
its supported source anatomy. Fixed source nodes have persistent named identities
associated with source pointers in the exact pinned template. A content edit
does not recompute identity from text or visually similar shapes. Missing targets,
changed source types, changed hashes and unsupported source features fail rather
than being dropped or matched approximately.

`cards` binds one declared card-row composite, not disconnected source slide
objects. Every card supplies a nonempty unique ASCII letter/digit/underscore/
hyphen key. Input order selects the column and automatic number; the key selects
the card's persistent identity. Reordering the array changes order/number without
renaming the keyed card or its owned parts. Count remains exactly three or four;
there is no prototype inference, extra repetition or continuation.

Metrics are separate named `metric1` through `metric4` slots, or `metric1` and
`metric2` in the rail template. They are intentionally not an array binding across
disconnected source nodes. Their fixed names control identity and position;
values and labels can change without generating new identities or inferring a
repeating group. This conforms to the earlier planning contract's restriction
that array binding operates on one declared composite data field.

## Frozen source and immutable layout

The source root is `library/wm-design-system/v1`. Its existing inventory and bundle
hash validation remains mandatory. Relevant template-family source hashes are:

| Source family | SHA-256 |
|---|---|
| `templates/library/openers.json` | `54c11bef39b933ead57bf994d1c2da4c06c453d6a9e43761b253557d639ef4b6` |
| `templates/library/evidence.json` | `51b1d2dc40596cacb971de4307ad106b3e31479b4a55fc0be5c9ab5775ee4b02` |
| `templates/library/argument.json` | `21b54f50ce5d8d13bda451aa94923d5e84cc9532f77ac3e13739cfcc31f5eba2` |

Inventory hash is
`efe3801e16c282d1ea1fcffefc7f360a6a0afea924b0f9fe45798ccf92d8a012`;
bundle manifest hash is
`02a975693995f9aca88602cd64e159e054da5a9771d8e3205d3beffd3d3ee45e`.
Source changes require adapter migration and explicit identity preservation.
Content cannot replace these pins or mutate the frozen source bundle.

Geometry, styles, surfaces, numbering mode, spacing, rules, frame/header/footer
and logo placement come from the source compiler. There are no content fields for
node coordinates, frame changes, font overrides, card styling, count changes or
density. Typography, overflow and the existing 0.25pt component wrap-boundary
exclusion continue to apply. Source character/word budgets are guidance; the
Go engine must fit the actual displayed copy into each fixed slot. It never
shrinks, clips, truncates, changes density or substitutes fonts to force a fit.

The fixed arrangements are:

- `cards/3`: three 270pt cards at x=57/345/633, y=144, h=216, 18pt gaps,
  outline body, inverse title band, 18pt padding and band numbering.
- `cards/4`: four 198pt cards at x=57/273/489/705, y=144, h=198, 18pt gaps,
  subtle surface, 18pt padding and inline numbering.
- `stats/four-metrics`: four 198pt metric slots at x=57/273/489/705, y=180,
  with fixed 162pt capacity, a source rule at y=162 and body support at y=342.
  No rail and a tall footer.
- `takeaway-rail/metrics-rail`: metrics at x=57/345, y=162, w=198, h=144;
  rule at y=306; support at x=57, y=324, w=414. The inverse right rail has
  eyebrow at x=705, y=126, w=198; heading at y=144 with 144pt fixed capacity;
  small body at y=288. The source compiler resolves the exact rail/footer frame.

These are content-bound arrangements, not arbitrary-content qualified templates.
Unsupported source anatomy must fail compiler validation; a future source field
cannot silently acquire rendering semantics merely because the underlying
component implementation happens to support something similar.

### Source sample copy and mark resolution

The stats reference binds `3` as the value and `rollout waves` as the label,
where the source example uses `3 waves` as its value. This is a supplied content
choice. It does not rewrite the source, change geometry or authorize a smaller
font. Every supplied value is measured at the unchanged token size and tracking;
character count alone does not establish overflow.

The source stats and rail titles carry exploratory `emphasis` settings associated
with `[[phrase]]` marks. Bound titles here are plain strings without marks. The
compiler keeps title typography and does not render an inactive emphasis mode as
a visible decoration. Markup-bearing bound titles are rejected; highlights and
underscores are not silently reinterpreted as bold runs.

Raw formatting reuses `FormatNumber` and the bounded
[standalone metric contract](data-metrics.md). Formatting is a declared value
projection with exact decimal rounding; no client calculations, currency
conversion, savings inference, target derivation or model evaluation is built in.

## Runtime and review boundary

Compilation, numeric formatting, text measurement, fitting and PPTX generation
run in Go. Original IBM Plex Sans and IBM Plex Mono names remain in the deck and
PDF. Existing static faces and calibration files are unchanged. Native capture
is not a permanent generation prerequisite; local PowerPoint export is a bounded
development review operation.

`TemplateReference(year)` contains eight `synthetic_example` slides: two for each
of the four supported keys. Card pairs demonstrate reordered persistent keys and
changed content in the same geometry. Stats demonstrate authored and formatted
values plus rich support. Rail pairs demonstrate bound source/rail copy and rich
support, heading and body. All claims are explicitly illustrative; the reference
contains no evidence of observed client outcomes.

Generic qualification remains false. Native review, visible-line comparisons,
artifact hashes and any exact-fixture qualification results belong in the parent
integration receipt after export. The exact fixture review is recorded below. Arbitrary replacement content and
native font-file selection remain unqualified. No tests were added or run.


## Reviewed reference (2026-10-02)

The [eight-slide reference](../../../samples/wmds-templates-20261002/README.md)
opened in PowerPoint without repair and was exported locally. Every page was
reviewed at 1920×1080. All 133 predicted visible lines matched the PDF; package
geometry/styles and all 112 text objects/paragraphs were preserved. There are
30 native groups, four source rules and four rich text boxes (11 runs).

A separate source-to-output inspection confirms all 70 assignments, source
hashes/pointers, content projection and four pairs of unchanged source geometry
and styling. Three rich runs lack uniform vertical anchors and retain conservative
height estimates. Both right rails have 3,252 opaque navy perimeter pixels and
no white edge. Rebuilding from the published content reproduces all PPTX parts
except the core created/modified timestamps.

The [review receipt](../../../samples/wmds-templates-20261002/final/native-review.json)
pins artifact/implementation hashes. General qualification and per-character
native capture remain false. The [structural input schema](template-content.schema.json)
provides editor guidance; the Go decoder and measurement enforce additional
identity, formatting, contrast, glyph and fit constraints.

The new `rule` foundation primitive is restricted to v2: decorative, horizontal,
0.75pt height, source `line` ink, native editable rectangle with no stroke. It
uses the outer x/y/width lattice with a named height exception, stays within its
zone and has no content/style fields. It does not imply connector or arbitrary
line-style support.
