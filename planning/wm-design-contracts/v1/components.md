# Measured text blocks and basic cards

Contract: `pptxgengo.wmds-components.v1`, implemented 2026-10-01 as an additive
input to `pptxgengo.wmds-foundation.v1`. Requires `wmds-go-foundation.v2` explicitly.
The primitive v1 engine remains the command default. This is a bounded prototype,
not execution support for every source component or arbitrary native text.

## Source and scope

Frozen `components/v0/components.json` supplies `text.block`, the basic body
presets from `card.presets`, `card.band`, and the shared `cardSystem` anatomy.
`tokens/v0/tokens.json` supplies typography, surfaces and inks. The frozen
`explorations/components.src.html` supplies internal spacing, bullet geometry,
and baseline alignment defaults. Source pins and original IBM font names stay intact.

Supported: three text-block variants; cards with uniform paragraphs or square
bullets; inline numbers; title bands; four accent-edge sides; all seven surfaces;
explicit standard/dense bodies; measured or fixed heights. Paragraph breaks,
including leading, empty and trailing paragraphs in a nonempty body, are preserved.

Deferred: mixed runs/emphasis, nested lists, columns, person/media/quote
presets, state tags, icons, badges, component footers, card rows, composites and
slide-template binding. Unknown JSON fields are errors. The generated reference
is fictional fixture content; it is not a client narrative or performance claim.

## Node envelope

`kind` is `textblock` or `card`, with the corresponding typed `textblock` or
`card` payload. Use existing `id`, `scope`, `rect`, and `grid/start/span` geometry.
Do not combine these payloads with primitive `text/style/ink/align` fields.

- Outer rectangles follow the 18pt module and fit their resolved body/rail zone.
- Text blocks: 198–558pt width (3–8 columns). Cards: at least 198pt.
  More than one body block requires 414pt (6 columns).
- Narrow five-up/context variants are deferred. A 198pt rail card is supported.
- Explicit height is a capacity, not a request to shrink. Height 0/omitted
  computes the required occupied/allocation extent and rounds up to 18pt.
- Text-block surface inherits the zone. A supplied surface must match that zone.
  Cards default to `light`; an explicit card surface paints its own container.
- Labels allow one measured line, titles two, inline numbers one. The body grows
  only within the supplied height and frame capacity. Required content cannot be
  empty/whitespace-only. No character count acts as a fit certificate.
- All child names are owned by their component. Collisions with other nodes fail.

## Text-block payload

`body` is required. `title` is optional. `label` requires `title`.

| Part | Style | Ink | Gap after |
|---|---|---|---|
| label | label 9/12, uppercase | emphasis | 6pt |
| title | subhead 18/24 | display | 6pt |
| body | body 14/21 | secondary | none |

## Card payload

| Field | Accepted values/default |
|---|---|
| label/title | optional uniform strings |
| body | nonempty array of keyed `BodyBlock` values |
| padding_pt | 12, 18, 24; default 12 below 270pt width, otherwise 18 |
| title_style | subhead (default), heading |
| title_ink | role passing contrast; default display |
| inline_number | one-line string, requires title |
| number_ink | role passing contrast; default emphasis; requires number |
| body_size | body (default), small |
| dense | must be true for small; cannot be true with body |
| band.surface | one of the seven surfaces; requires title |
| edge | side left/top/right/bottom, weight_pt 3/6, explicit ink |

Every `BodyBlock` has `key` and exactly one of `p` or `bullets`.
Every bullet has `key` and `text`. Keys contain only ASCII letters, digits,
underscore or hyphen and are unique across the component's body. Child identities
include these keys, preserving identity on reordering. Text markup is unsupported.

Internal flow uses authored line allocation, with an independent occupied-bound
check. Labels have 6pt gaps. An unbanded title has 12pt before the body. Paragraph
blocks have 6pt gaps. Standard square bullets have a 4pt marker at y+8.5pt, 15pt
text inset and 6pt item gap. Dense bullets use 3pt markers at y+7pt, 12pt inset
and 3pt item gaps. Numbers align their first baseline with the title baseline.

An edge occupies the inside of its container and adds its weight to the relevant
content padding. A band measures its own header height and uses its own inks.
The body starts below the band with card padding. Minimum contrast is 4.5:1 for
small text, 3:1 for qualifying large text, and 3:1 for accent edges against every
surface they cross. Outline is a white fill with a 1pt inset Medium Gray border.
Filled shapes have an explicit DrawingML no-line setting, rather than a
transparent stroke or an inherited theme outline.

## Wrap boundary guard and evidence

The existing v2 native matrix passes 139 controls, but new component copy exposed
one exact-width boundary: Go predicted a 414pt line, and PowerPoint moved the
last word to the next line. The original failed fixture and PDF comparison remain
in `samples/wmds-components-20261001/final-v3`. No font data or calibration changed.

Components now reject selected lines or eligible next-word decisions within
0.25pt of the available width as `component.uncertain_wrap_boundary`. This is a
conservative prototype exclusion informed by the prior maximum native advance
error of 0.132pt. It is not a proven error bound for arbitrary text. Change the
content or choose another permitted span; the engine does not silently inset,
shrink text, insert hard breaks or change density to resolve this error.

[The reviewed reference](../../../samples/wmds-components-20261001/README.md)
contains 34 native groups on 12 slides. Its package preserves all 153 text objects
and 160 authored paragraphs; the PDF matches all 195 predicted visible lines.
The 22 rejection records cover overflow, uncertain boundaries and unsupported
inputs. All pages were visually reviewed. No new character capture or general
font-file/renderer envelope qualification is claimed.

## Additive metric-card slice

Cards also accept the separately versioned [metric contract](metrics.md), with
`metric` or `metric_group` in place of the paragraph/bullet `body`. This does not
qualify standalone data metrics or other card presets.

## Native editability

Each component is a real `p:grpSp` group, with native text and rectangular shapes
named by their part identities. An initial identity transform keeps the planned
point coordinates. PowerPoint can move/resize the group and edit its members.
Bullet markers are native square shapes next to native text, not font glyphs.
Regeneration writes a new deck and does not import manual edits.

Generation and measurement run entirely in Go. The reference PDF comes from a
separate, local PowerPoint export used for development review. Per-deck native
capture is not a runtime requirement.
