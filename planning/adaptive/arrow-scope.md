# Hand-drawn arrow qualification scope

The bounded candidate is [`library/diagram-components/arrow-qualification.json`](../../library/diagram-components/arrow-qualification.json): ten illustrative slides. Nine propose automatic placement; the tenth stages an incompatible-direction arrow for manual placement. The source SVGs and derived PNG fallbacks are SHA-256 pinned in the spec. The current asset catalog retains **manual candidate** status for general use and now links native and optical evidence for these exact examples. Their proposed endpoint and transform envelopes are not generally qualified. No source slide identity or production placement envelope is claimed.

## Exact artwork in scope

| Artwork ID | Source SVG | Selected use |
| --- | --- | --- |
| `wm-handdrawn-single-arrow` | `samples/visual-wave3/arrow-assets/source/handdrawn-single-arrow.svg` | Straight single arrow, right-facing tail and tip. Its SVG supplies the description “Handdrawn single arrow annotation for pointing to a specific item or direction.” |
| `wm-handdrawn-dashed-arrow` | `samples/visual-wave3/arrow-assets/source/wm_handrawn_dashed_arrow_rgb_240912.svg` | Bent dashed arrow, bottom source to left target in candidates; also the manual fallback. |
| `wm-handdrawn-right-angle-arrow` | `samples/visual-wave3/arrow-assets/source/wm_handrawn_right_angle_arrow_rgb_240912.svg` | Right-angle arrow, bottom source to left target. |

The exact hosted inventory has no visual description for the dashed and right-angle SVGs; their direction and bend descriptions are observations from the local artwork. The source connecting-loop arrow is outside these nine automatic cases and remains a manual candidate.

## Placement cases

Each of the three artworks has the same three selected scale and rotation experiments: `0.75× / 0°`, `1.25× / −15°`, and `1× / +15°`. Straight-arrow cases connect a right source port to a left target port. Dashed and right-angle cases connect a bottom source port to a left target port. The exact labels, target frames, gaps, endpoint coordinates, alpha tiles, SHA-256 values, and fallback paths are in the ten-slide spec. These experiments use uniform scale and rotation; they do not stretch or flip the SVG.

`arrow-direction-manual` is the tenth case. Its requested top source port conflicts with the dashed artwork's departure tangent. The spec stages the pinned artwork and gives an explicit placement note rather than claiming an automatic attachment. A human must choose a compatible source edge and position the arrow.

Unsupported without separate measurement and visual qualification: arbitrary spans, sizes or rotations beyond the selected cases; different source/target ports or direction; flipped or nonuniformly scaled artwork; overlapping labels and graphics outside these layouts; an automatic return-loop placement; and reuse of these candidate endpoints as general production limits. Asset geometry, structural packaging, and native text fit do not establish optical attachment or visual quality by themselves.

## Accepted evidence

All ten v3 cases passed native text fit, final PowerPoint verification and optical review. The nine automatic cases are accepted only at their exact authored frames and transforms. The tenth demonstrates the manual fallback and remains manual. The 71 fixed text zones have zero overflow and the layout planner passed. Every regenerated v3 PNG is byte-identical to its visually reviewed v2 image; see `arrow-render-comparison.json`.

- Input: `library/diagram-components/arrow-qualification.json`
- Fit: `samples/adaptive/arrow-qualification-v3-fit.json`
- Deck: `samples/adaptive/arrow-qualification-v3-deck/arrow-qualification-v3-deck.pptx`
- Native verification: `samples/adaptive/arrow-qualification-v3-verification.json`
- Renders: `samples/adaptive/arrow-qualification-v3-render/`
- Exact identities: `library/adaptive/checkpoint.json`, under `accents.accent_arrow`

PowerPoint exposes some graphic-internal SVG strokes as `single line` with the unavailable/mixed width sentinel `-2147483648`. The verifier retains those raw values and accepts the anomaly only for a no-outline SVG graphic after independent package checks prove the picture has no `<a:ln>`. Ordinary line and explicit-outline checks remain strict. Fresh v3 measurements and a rebuilt deck were required after that adapter correction; old evidence was not relabeled. The source input retains its original candidate-provenance wording; the checkpoint records the later acceptance.

The three selected transforms per artwork do not establish a continuous production envelope. The connecting-loop arrow remains untested here.
