# WMDS keyed card rows v1

Contract: `pptxgengo.wmds-card-rows.v1`. Requires the opt-in
`wmds-go-foundation.v2` engine. This is a bounded composite of the existing
[cards](components.md) and [metric cards](metrics.md), with the same font,
contrast, fit and content restrictions.

## Source and interpretation

Frozen `source/components/v0/components.json` defines `card.numbered` as two to
four equal-width cards, with inline, band and corner numbering examples.
Frozen `source/explorations/components.src.html`, `case "cardrow"`, repeats
`card` with an 18pt example gutter and two-digit ordinal numbers. It treats `w`
as the **child** width. The foundation API below instead uses the node rectangle
as the **whole row** width, matching the existing grid and frame-zone contract.
This distinction is explicit; child width is derived rather than guessed.

The frozen browser renderer permits fixed heights that can clip content. This
implementation measures each child and rejects insufficient height. Its
numbering mode `inline` is ignored by the browser when a band is present; here
that combination is rejected. Use `band` to number a title band.

Corner numbering is deferred. Although the browser reserves 54pt on the right,
the supported card adapter does not expose that asymmetric padding or corner
number anchor. Other composites, wrapped rows, unequal widths/heights,
five-up cards, custom gutters, automatic continuation and arbitrary card count
are not implemented by this contract.

## Authoring API

```json
{
  "id": "steps",
  "kind": "cardrow",
  "grid": "columns",
  "start": 1,
  "span": 12,
  "rect": {"y_pt": 144, "height_pt": 0},
  "surface": "subtle",
  "card_row": {
    "numbering": "inline",
    "items": [
      {
        "key": "diagnose",
        "card": {
          "title": "Diagnose",
          "body": [{"key": "summary", "p": "Rank delays by days lost."}]
        }
      },
      {
        "key": "design",
        "surface": "outline",
        "card": {
          "title": "Design",
          "body": [{"key": "summary", "p": "Move review to the start."}]
        }
      }
    ]
  }
}
```

Each item requires a nonempty immutable ASCII letter/digit/underscore/hyphen
`key`, unique within the row. Array order controls position and generated
number; key controls identity. Equal labels do not merge cards. There is no
implicit template merge: each item supplies its complete `card` specification.
An item surface overrides the row surface; otherwise it inherits the row's
surface, whose default is `light`.

`numbering` is absent, `none`, `inline` or `band`:

- `none` leaves each card's own supported inline-number option intact.
- `inline` generates `01` through `04`, requires every card title, and excludes
  title bands. Card `number_ink` keeps its existing default/validation.
- `band` requires every card title and explicit `band`, generates the same
  ordinals, and uses the band surface's `emphasis` ink. A conflicting explicit
  number ink is rejected.

Explicit `card.inline_number` conflicts with row-generated numbering. Neither
mode silently supplies a missing band, title, surface override or body.
Primitive text/style/ink/alignment fields and other component payloads on a
card-row node are rejected. Unknown JSON fields remain errors at the strict
document boundary.

## Geometry and fit

There are two to four cards and a fixed 18pt gutter. For whole row width `W`
and count `n`, child width is `(W - 18*(n-1))/n`. Every child must be at least
198pt wide and must independently satisfy the foundation outer-box lattice.
For example a full 846pt row yields widths 414, 270 or 198pt for two, three or
four cards. A six-column 414pt row supports two 198pt cards; an eight-column
558pt row supports two 270pt cards. Fractions that do not meet the lattice fail.

The first pass measures each card at automatic minimum height. The row's
reported `required_height_pt` is the largest raw child requirement. With row
height zero, final equal height is that maximum rounded up to 18pt. With a
positive height, all children use the supplied height and the row rejects if
any requirement exceeds it. Each child is planned again at the shared height,
so a metric comparison/status footer uses the card's existing bottom anchoring.
All children share the row's top and bottom; the row must fit wholly within
one resolved frame zone.

Inherited card limits include title/label/value line caps, source font names,
the conservative 0.25pt wrap-boundary rejection, explicit density and padding,
contrast checks and occupied terminal-line bounds. No clipping, shrinking,
content deletion, density switching or typography changes resolve overflow.

## Native identity and report

The row is a genuine native PowerPoint group containing one native card group
per item. Child identity is `<row-id>.items.<key>`; card-owned part IDs extend
that prefix. Reordering a row with the same ID preserves those names and changes
coordinates/ordinals. Removing a key removes only its generated group. The row
and card groups use identity transforms initially, retaining native text and
shape editability.

`slides[].card_rows[]` records the contract, source definition (`card.presets`
for the unnumbered repetition adapter, `card.numbered` for numbered rows), final rectangle,
raw maximum required height, gutter, numbering mode, ordered keys and ordered
child group IDs. Child card details remain in `slides[].components[]` and
text details in `slides[].texts[]`. Group IDs and owned part IDs must be unique
within a slide; references in row `parts` point to existing child groups rather
than declaring duplicate objects.

Direct PowerPoint movement and resizing are ordinary native group edits.
Semantic regeneration remains driven by the input document. It does not import
manual edits from PowerPoint or preserve arbitrary reflow after native resizing.

## Qualification boundary

`CardRowReference` supplies seven fictional slides covering counts, fixed and
automatic height, mixed surfaces, inline/band numbering, metric footers, nested
metric groups and reordered keys. Successful generation establishes measured
fit only. Any PowerPoint/PDF review receipt applies to the exact reviewed
fixture and artifact hashes; arbitrary replacement content, font file selection
and general typography qualification remain unqualified.


## Combined reference review (2026-10-02)

The [combined reference](../../../samples/wmds-parallel-slices-20261002/README.md)
contains this slice on slides 1–7: 9 row groups and 28 nested cards. All 19 pages were
reviewed individually after local PowerPoint PDF export. Package geometry/styles,
paragraph preservation and all 255 predicted visible lines passed the combined
inspection. The [receipt](../../../samples/wmds-parallel-slices-20261002/final/native-review.json)
pins the artifacts and implementation. This review applies to the exact fixture;
per-character bounds, exact native font-file identity and a general content
envelope remain unqualified. No tests were added or run.
