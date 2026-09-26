# Wave 2 response-row fixture

## Observed source facts

EnableComp slide 5 is pinned to source deck SHA-256 `22c39b27bd97fece99008fd456cdde377ec0750eb296100d8570ee83f767c77b`. Its five need tiles align vertically with five response rows. Each response row contains a numbered bold lead and a regular explanatory paragraph, plus one source SVG icon and a vertical divider. Source preview: `samples/catalog/render/enablecomp/png/slide-005.png`; structural sidecar: `samples/inspection/enablecomp/sidecars/slide-005.md`; source mapping and media hashes: `library/component-contracts/visual-wave2-candidates.json`.

The original icons are package parts `image21.svg`, `image20.svg`, `image22.svg`, `image24.svg`, and `image23.svg`, mapped in row order to slide picture IDs 91, 90, 94, 97, and 95. The candidate registry pins their bytes. The source package has no PNG fallback relation for these icons. The renderer currently accepts PNG/JPEG but not SVG, so the five exact SVG masters were rendered through `scripts/render-svg-preview.swift` using AppKit NSImage at 8×. The derived 608×608 PNGs are **raster previews, not editable vectors**; `samples/visual-wave2/response-assets.json` binds each preview hash to its SVG source hash and rendering recipe.

## Fixture transforms

`library/visual-components/response-review.json` is a changed-copy composition fixture with five paired need/response rows. It retains a close approximation of row pitch and content zones, includes icon, separator, numbered bold title, and body copy, and uses editable text/surfaces. The source slide uses a rotated triangle behind the needs; the fixture substitutes an editable `rightArrow` preset as a declared source-like treatment. Source-magenta section labels are darkened in the fixture to pass the composition validator’s contrast rule. Source response paragraphs use 14 pt/11 pt Arial; the fixture uses 13.5 pt/9.5 pt to leave working room for changed copy in the explicit row frame.

The fixture was accepted by the nonnative compose probe at `samples/visual-wave2/response-probe`. This validates spec expansion only; it is not PowerPoint measurement or visual-fit evidence. Reordering, alternate row counts, different copy lengths, and native SVG editability remain unproven.
