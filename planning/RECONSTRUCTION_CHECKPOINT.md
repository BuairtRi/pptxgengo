# Reconstruction checkpoint — 2026-09-25

This durable summary records the completed local experiment. Large originals,
generated decks, previews and detailed measurement JSON remain ignored under
`samples/reconstruction/`. This summary is not a substitute for those pixel inputs.

## Implemented

- Go 1.27.1 baseline; `pptxscene` extracts native scenes with typed bindings and
  rebuilds text, geometry, style, table and crop values while retaining source
  topology/resources. It preserves hidden internal-link destinations.
- `pptxdiff` emits literal PNG metrics, overlays and amplified differences.
- `pptxanchor` calculates underline/highlight placement using measured phrase
  bounds, visible artwork bounds, rotation and explicit optical calibration.
- Native PowerPoint text measurement/export, PDF rasterization, alpha bounds,
  package audits and rendering-context controls are available in `scripts/`.

## Results and limits

UHG slides 8, 11, 12, 14, 24, 28, 36, 38, 43 and 44 were reconstructed
sequentially. The combined delivery has zero differing pixels against untouched
source-slide controls in the same shortened-deck export context at 1920×1080.
Literal full-85-slide-export comparisons retain small background/footer
differences: the maximum slide mean RGB channel error is 0.084751 on a 0–255
scale. The exact internal renderer cause is unproven.

The structural audit matched ten normalized shape trees and text payloads,
333 objects, 59/59 media hashes and 30/30 layout/master/theme resources.
The ten slides expose 19,243 bindings. Bio 67 and accent fixtures 5–6 were also
exercised. Builds retain source topology and resources: independent design from
a brief, semantic layouts and general content adaptation remain future work.

Both slide 5 underline placements worked. The second intentionally changed the
title to move the phrase onto line two. It did not repair a failed baseline.
The same calibration followed measured phrase bounds. The SVG's visible stroke
occupies only about 8.04% of its image canvas height, so image-box anchoring alone
would be misleading.

Slide 6 changed title size from 24 to 32 pt with identical wording. The highlight
moved/resized behind the phrase using the same calibration. Its baseline image
rectangle was reproduced exactly in integer EMU. In the resized before/after
highlight comparison, all 18,282 solid title-ink pixels were unchanged; changed
pixels were confined to the accent region. Optical padding is source calibrated,
not a universal typography rule.

The slide 5 arrow's original placement was preserved. General curve selection,
endpoint semantics, collision avoidance and aesthetically pleasing arrow routing
are **not yet proven**. The manual fallback stages assets outside the slide with
specific labels and native speaker-note instructions.

Native measurement/rendering require macOS PowerPoint and Apple frameworks.
The Go scene/anchor/pixel tools are separate from that platform adapter.
Previous validation: `go test -race ./...` passed after highlight implementation.
No new test run is implied by this checkpoint document.

## Evidence retained locally

- `samples/reconstruction/REPORT.md`: detailed results and output links.
- `samples/reconstruction/qa/delivery-control/report.json`: same-context results.
- `samples/reconstruction/structural-audit.json`: object/resource comparisons.
- `samples/reconstruction/work/`: calibration, phrase and raster measurements.
- `samples/reconstruction/output/qa-review.html`: self-contained comparison viewer.

The [QA findings](QA_ERRORS.md) preserve observed defects and unresolved limits
in the repository. New experiments should record exact source hashes, renderer,
resolution, intended changes and the candidate/control paths before claiming
fidelity. Future portable regression fixtures should use the committed source
or synthetic content rather than depend on ignored private scene projects.
