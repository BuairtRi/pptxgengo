# WMDS mixed text runs v1

Contract: `pptxgengo.wmds-rich-text.v1`. Requires opt-in
`wmds-go-foundation.v2`. This is an additive executable subset of the foundation
document contract; the frozen source bundle, font files and calibration remain
unchanged.

## Scope and source

`kind: richtext` supports inline numeric weights and semantic surface inks in
one editable native text box. `style` selects a frozen typography token. All
runs inherit that token's family, size, exact leading, tracking and case. A run
may override weight with an actual static face at 400, 500, 600 or 700 and may
override ink with a supported surface role. Omitted overrides inherit the node.

The frozen exploration renderer uses `.b600` for semibold bullet leads and
the token/surface maps for typography and ink. Its `[[phrase]]` syntax instead
selects hand-drawn highlight, underscore, circle or spark marks. Those marks
and `[^n]` footnotes remain deferred: this slice rejects their markup and does
not reinterpret it as bold text. Card and bullet adapters retain their previous
uniform-copy contracts; this slice exposes direct rich text nodes.

## Input

```json
{
  "id": "review-copy",
  "kind": "richtext",
  "grid": "columns", "start": 1, "span": 6,
  "rect": {"y_pt": 144, "height_pt": 0},
  "style": "body", "ink": "primary",
  "richtext": {
    "paragraphs": [
      {"key": "summary", "runs": [
        {"key": "prefix", "text": "Every exception has "},
        {"key": "owner", "text": "a named owner", "weight": 600, "ink": "emphasis"},
        {"key": "period", "text": "."}
      ]},
      {"key": "trailing", "runs": []}
    ]
  }
}
```

`richtext` and the existing text/card/textblock content are mutually exclusive.
There is no automatic markup parser. Every paragraph and run has a nonempty
ASCII letter/digit/underscore/hyphen key. Paragraph keys are unique in the text
object; run keys are unique within their paragraph. Keys survive reordering in
the saved specification and report; an OOXML text run has no object-name field.

Paragraph arrays define hard breaks exactly. `runs: []` is an empty paragraph,
including a leading or trailing paragraph. A run's text must be nonempty and
contain no CR/LF, control or formatting characters. Literal whitespace is
preserved, including spaces across run boundaries. At least one paragraph must
contain visible content. Runs cannot split a combining-character cluster.

The existing scope, matching frame-surface context, grid resolution, alignment,
usable zone and geometry rules apply. Alignment is left, center or right for the
whole text box. A zero height receives its measured allocation. Fixed height
must contain the complete allocation and conservative visible terminal extent.
There is no autofit, shrink, clipping, font substitution or density change.

## Measurement and serialization

Go resolves and validates the actual static font for each run. Numeric weight is
never inferred from a Boolean `bold`, rounded to a neighboring weight or
synthesized. PPTX uses the original IBM legacy family names with the actual
face's native bold/italic flags; the report includes font-file hash and full name
metadata for every run.

The Latin/Common/Inherited and missing-glyph restrictions from v2 apply to every
run. HarfBuzz shapes the paragraph spans with global rune offsets. The line
wrapper sees all styled spans together, with kerning disabled, standard
ligatures enabled, 1/8-point quantized glyph advances and serialized
0.01-point tracking per cluster, including terminal tracking. Equal-face spans
coalesce for Go shaping, even when ink changes; different faces form separate
shaping spans. This avoids treating each run as an independent text box or
breaking paragraph wrapping at every style boundary.

An allocation decision within 0.25 point of the width, either for its selected
line or for that line plus the next word, fails `rich.uncertain_wrap_boundary`.
The check uses the mixed faces. This guard is a development heuristic derived
from earlier native-reference advance differences, not a proved universal error
bound.

For vertical estimates, the text object uses the maximum involved face baseline
and terminal extent. A pinned exact font-hash/size/leading anchor is used where
available; otherwise the existing conservative candidate estimates apply:
baseline 0.75 leading, terminal `max(leading, 1.5 × size)`. Empty paragraphs
reserve one leading without visible ink. Unknown weight/size combinations are
reported per run. The report's aggregate calibration hash is populated only if
the base face and every run have known anchors; it is never a native mixed-run
qualification flag.

Serialization emits explicit paragraphs and complete run properties, including
font names, size, tracking, zero kerning threshold, real weight flags and ink.
Each paragraph ends with complete **base-token** defaults, even if its last run
has another weight or color. Paragraph before/after spacing is zero and line
spacing is the exact source leading. Empty and trailing paragraphs receive the
same defaults. Soft wraps remain native and editable; predicted wraps are not
converted to authored hard breaks.

Contrast is checked for base paragraph defaults and every run against the actual frame surface: 4.5:1 below
18 points, except real weight 700 at 14 points or above may use 3:1; 18-point and
larger text uses 3:1. Weight 600 does not receive the weight-700 exception.

## Evidence boundary

The prior 139 native controls calibrate uniform text. They do not independently
qualify mixed-run line breaks, mixed-weight baseline alignment, face selection,
ligatures across native run boundaries or arbitrary content. Every rich layout
retains `native_qualified: false`. Reference PDF review can support the specific
saved examples' wrapping and visual quality; it cannot establish native
per-character bounds or exact installed font-file identity.

`RichReference` provides six fictional reference slides: real weights, paragraph
wrapping across runs, same-face and different-face splits inside ligature words,
empty/trailing paragraphs, inverse-surface ink, Mono runs, small/source text and
tracked uppercase tokens. Review receipts remain separate from generic build
qualification.

## Deferred

Per-run family, point size, leading, tracking, case, alignment, italic, underline,
baseline shift, hyperlinks, automatic line-spacing, hand-drawn marks, footnote
binding, semantic bullets and card rich-content bindings are not accepted by
this slice. Their extension requires an explicit contract and measured layout
policy. Unknown JSON fields fail at the CLI boundary.


## Combined reference review (2026-10-02)

The [combined reference](../../../samples/wmds-parallel-slices-20261002/README.md)
contains this slice on slides 14–19: 15 rich text boxes and 47 runs. All 19 pages were
reviewed individually after local PowerPoint PDF export. Package geometry/styles,
paragraph preservation and all 255 predicted visible lines passed the combined
inspection. The [receipt](../../../samples/wmds-parallel-slices-20261002/final/native-review.json)
pins the artifacts and implementation. This review applies to the exact fixture;
per-character bounds, exact native font-file identity and a general content
envelope remain unqualified. No tests were added or run.
