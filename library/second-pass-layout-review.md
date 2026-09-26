# Resolved geometry and second layout review

2026-09-25. Root reviewed each of the 40 new native PNG previews individually;
38 informed layout decisions and UHG 43/44 informed component boundaries.
The original 51 reviewed previews remain valid. Source decks were unchanged.

## Results

| Measure | First pass | Current |
|---|---:|---:|
| Source slides retained | 369 | 369 |
| Source previews in render manifest | 51 | 91 |
| Occurrences with layout decisions | 51 | 89 |
| Shared visual families | 17 | 24 |
| Reviewed distinct items | 1 | 2 |
| Classification work units | 336 | 306 |
| Repeated classification units avoided | 33 | 63 |
| Unreviewed singleton slides | 318 | 280 |

These are classification work units, not a certified count of unique reusable
layouts. Every copy, style and geometry variant remains addressable. Design
preference, content approval, and technical reuse readiness remain separate.

## Visual decisions

| Slides (source indices) | Decision and retained differences |
|---|---|
| UHG 67–85 | One dense bio family: portrait/identity, central biography, two right expertise cards. Retain different card heights, type sizes, portrait crops and copy density. UHG 67 is the user-nominated representative. |
| UHG 24, 28, 34 | One phase-detail family. Preserve phase-panel widths, row dimensions, exit-check count, active-arrow colors, and deliverable images. |
| UHG 41, 47 | Extend the section-divider family to seven members. Different photos, two-line title and optional magenta subtitle remain variants. |
| EnableComp 11–14 | One approach-detail family. Slide 12 has outlined boxes; slide 14 adds the NEXT PHASE headline treatment. |
| Modernization 44, 45 | One input/engine/output explainer family. Technical/business wording, panel height, typography and fourth benefit-row density differ. |
| Graphics and Layouts 14, 15 | One six-block grid family with filled versus outlined/colored header variants. |
| Graphics and Layouts 69, 78 | One half-year timeline family; retain row fills, header colors and main-idea callout differences. |
| Graphics and Layouts 82 | Extend the existing 74/75 six-week timeline family. Preserve filled navy week header, slate bars and adjusted row positions. |
| Graphics and Layouts 119, 120 | One annotated office-map family; round markers and pointer labels require different component and attachment rules. |
| Graphics and Layouts 121 | Keep the unannotated map separate. Shared background artwork does not supply the office-marker content roles. |

The expanded retrieval pass uses resolved leaf frames across group encodings.
Its initial 61-pair queue produced 45 reviewed pair decisions in this pass;
family-wide reviews also covered bio variants that were absent from that queue.
Recomputing against the accepted families leaves 17 geometry candidate pairs.
A separate [metadata-only shortlist](broader-dedup-candidates.md) has ten further
pairs. Neither list authorizes automatic merges.

## Geometry and component evidence

The geometry catalog resolves 9,669 of 10,102 object frames, including all 315
native groups. It recovers 579 placeholder transforms from layouts/masters.
The remaining 433 frames remain unresolved. Frame corners include ancestor
scaling, rotation and flips; text ink, font inheritance and clipping are separate.

The component catalog includes 315 native-group occurrences plus 21 curated
compositions across seven proposed families and 37 named slot candidates.
The 315 groups occupy 314 conservative structural buckets. That is not a claim
of 314 distinct semantic components. Native-group and curated records can overlap.

Root compared selected geometry records with native slide previews:

- UHG 43 pod 1: container at approximately (773,661), 205×202 pixels on the
  1920×1080 preview; two role boxes begin at y=699 and y=770. Pods 2/3 retain
  their third boxes at y=841. The shared PHASE 2 ONLY backdrop remains slide-owned.
- UHG 44: the first identity card is at (72.5,296), 345.6×72 pixels. Its 72×72
  portrait overlays the card's left portion. The paired names and portraits
  were checked against the source; they are source content, not reusable staffing claims.
- UHG 67: portrait and inherited identity frame share x=72.5 and width=252;
  identity starts at y=484.6. Inherited geometry is retained as such.
- UHG 24/28: the progress group spans x≈72.5–478.8, y≈342.3–387.1; its three
  child arrows have unequal widths. Preserve group scaling and source styling.

These are selected frame checks, not a pixel-perfect geometry certification.
The [local component gallery](../samples/catalog-next/component-review-v2.html)
includes unchanged source PNGs and toggleable SVG frame overlays. Its inputs,
preview hashes and member references were validated. Browser policy blocked
opening the local HTML in the automated browser, so the gallery UI itself has
not been visually verified. Direct source previews and coordinate records were reviewed.

## Source quality findings

- **QA-24 / UHG 73:** the final bio line reaches into the footer area. Family
  membership does not make this a fit-approved example.
- **QA-25 / Modernization 45:** a short underline floats below the intro after
  wording changed; on slide 44 it emphasizes “successful.” This is a useful
  future regression fixture for phrase anchoring and detached emphasis detection.

No source slide was repaired or rewritten during this inventory pass. These
issues must be distinguished from errors introduced by our own build pipeline.

## Next work

1. Review the remaining candidate pairs once per source occurrence, then classify
   and rate the remaining representative work units.
2. Resolve typography and text insets for the seven component families, measure
   rendered text, and exercise replacement/overflow behavior.
3. Keep semantic slot count separate from observed text length; do not invent
   character limits from the samples.
4. Connect dependent media, theme resources, anchors, and intentional overlaps
   before promoting any component to supported authoring.
