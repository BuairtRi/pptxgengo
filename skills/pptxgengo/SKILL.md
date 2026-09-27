---
name: pptxgengo
description: Build or adapt PowerPoint presentations with this repository's native scene, reviewed component, or experimental dynamic composition tools. Use when working in pptxgengo; this does not treat catalog examples as approved templates.
---

# pptxgengo

Use the narrowest workflow that fits the requested change:

- Use `pptxscene` to extract selected native slides into a source-bound scene project and rebuild a new PPTX.
- Use `pptxcomponent` to inspect or apply reviewed text/color contracts in an extracted scene.
- Use `pptxcompose` for measured, editable compositions from JSON. Pods/team, cards, canvas, measured grids/panels, and accents have bounded support; the newer cards/canvas/accent surface is alpha.
- Use `pptxlib` to find/inspect/preview library contracts, instantiate semantic slots, and assemble a deck from saved values and narrative. Read [library authoring](references/library-authoring.md) for this route. The initial thirteen portable contracts are candidates awaiting changed-content native qualification.
- Use `pptxtemplate` to list, inspect and build changed-content review decks from the 65 shortlisted source designs. Read [template adaptation](references/template-rollout.md). These fixed-geometry examples have separate native visual review states; they are not arbitrary-content qualified templates.
- Use `pptxcompose recover-text --text-only` for bounded text recovery from a generated deck with its original spec/bundle. Original styling is restored; changed geometry or unknown objects require the scene workflow.
- Use `pptxanchor` to calculate phrase-level raster/SVG accent placement from native measurement evidence. It emits placement JSON; it does not edit a deck.

Read [the workflow reference](references/workflow.md) for command forms, measurement and catalog guidance, and current limits. Keep outputs new and preserve source evidence. Catalog ratings and design preferences are not technical approval. Do not describe source patterns as approved reusable templates. The five dense proposal recipes have exact native verification and visual-review evidence in `library/showcase/dense-proof.json`. Keep the broader canvas/cards/accent surface alpha; those fixtures do not approve arbitrary layouts or styles.

For named grids/panels and incremental native measurement, read
`library/layout-components/README.md`. Reuse the cache only through the CLI's
contract/environment validation. Every final deck still needs native verification
and visual review. Parent ownership lives in the spec; native manual dragging does
not move related shapes as a group.

For mixed typography, image crops, and roadmap presets, read
`library/visual-components/README.md` and its qualification checkpoint. These are
bounded Wave 2 features: Arial paragraphs/runs and two native bullet glyphs,
pinned PNG/JPEG or static SVG with a pinned PNG fallback, picture outlines, and
four editable shape presets. Read the follow-up qualification status before
promoting new variants. Rich-text recovery is not supported.
Keep source geometry controls distinct from changed-content designs; fixture
generation and passing unit tests alone do not establish visual fidelity.
