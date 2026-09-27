# Wave 2 image follow-up: SVG fallback and picture outline (preimplementation audit)

## Observed implementation

The compose renderer in `cmd/pptxcompose/image_render.go` hashes one image asset, then calls Go `image.DecodeConfig` to obtain dimensions. Its registered decoders are PNG and JPEG. The image spec has one `asset_path`/`asset_sha256` pair and no fallback pair, so a direct SVG image is rejected before composition.

The lower-level PowerPoint writer already recognizes an SVG data URI. `pptx.addImageDefinition` registers two relationships and `slideObjectImageToXml` writes an `asvg:svgBlip` extension for the vector. However, it currently assigns the same SVG bytes to the PNG-labeled fallback relationship. `pptx/media.go` documents this path-based duplication as a reproduced upstream quirk. The current image structure checker also expects a single, childless `<a:blip>` and does not validate the SVG extension or a separate fallback relationship. That leaves no verified SVG plus valid PNG fallback path through the compose pipeline.

`ImageProps` has rounding and shadow options, but no line. `ObjectOptions` already has a `Line` field for shapes; the image writer's `<p:spPr>` emits transform and geometry, then shadow, but no `<a:ln>`. Wave 2's thumbnail fixture therefore draws borders as four independent line objects.

## Source-derived fixtures

| Case | Source evidence | Use |
| --- | --- | --- |
| SVG primary + explicit fallback PNG | EnableComp slide 5 icons: five source SVG masters and 8× AppKit PNG previews, listed as source/derived hash pairs in `samples/visual-wave2/response-assets.json`. Source SVGs have no PNG fallback relationship in their original package. | Exercise dual-media packaging with a known SVG and an explicit PNG derived from that exact SVG. Label it as a raster fallback preview; do not imply it is editable or an original package fallback. |
| Existing bitmap compatibility | UHG slide 28 thumbnail parts `image146.png`–`image149.png`, with original hashes in `library/component-contracts/visual-wave2-candidates.json`. | Confirm PNG sizing, crop/contain handling and one-media relationship stay unchanged. |
| Picture outline | UHG slide 28 pictures IDs 8, 11, 15, 19 (parts `image146.png`–`image149.png`). Each has rectangular `a:prstGeom`; the picture's own `p:spPr` contains `a:ln` → `a:solidFill` → `a:schemeClr val="accent3"`. The slide's resolved theme1 accent3 is `#CED7E6`. No explicit line width or dash appears in the picture XML. | Replace the four-line mock border with a single outline on the picture object. Treat omitted line width as source default/inherited; do not claim a measured point width from XML alone. |

The enablecomp source slide uses SVG extensions but no PNG relationship fallback. The AppKit previews in its asset manifest are derived fallback candidates, not original source fallback bytes. The UHG28 images provide a separate pinned bitmap control and exact source picture-outline evidence.

## Smallest reliable API

Add optional `fallback_asset_path` and `fallback_asset_sha256` to compose image elements. Require both fields together and require them when the primary asset is SVG. Verify the SVG hash, verify that the supplied fallback hash matches, and run `image.DecodeConfig` on the fallback. Derive fit geometry from explicit SVG root `width`/`height` when both are usable CSS px/unitless values; otherwise require a declared intrinsic pixel size or use the fallback dimensions with an explicit source such as `fallback_png`. Preserve SVG bytes in the vector relationship and PNG bytes in the fallback relationship. Never copy SVG bytes into a PNG relationship or silently rasterize during ordinary composition.

At the writer boundary, add a PNG fallback data/path field to image options and use it for the first media relationship while retaining the current SVG relationship and `<asvg:svgBlip>` extension. Extend image structure validation to check the PNG relationship's content type and pinned hash, the SVG relationship's content type and pinned hash, the `asvg:svgBlip` relationship reference, and the direct `<a:blip>` PNG fallback reference. Keep the existing strict PNG/JPEG path unchanged.

For borders, expose an optional bounded `line` object on image elements (color, width in points, dash). Carry it to `ImageProps.Line` and serialize `<a:ln>` inside the picture's `<p:spPr>`, next to geometry. For the source UHG28 case, a resolved RGB color `#CED7E6` and no explicit width/dash match the observed color and absent attributes; this resolves the source theme color statically rather than preserving theme binding. Validate the outline in the picture's own `spPr`, not through four separate decorative shapes.

## Offline proof checklist

- [ ] Render an SVG with an explicit pinned PNG fallback and verify output media parts retain the exact respective input hashes and content types.
- [ ] Verify slide XML references the PNG fallback directly and references the SVG through `asvg:svgBlip`; no external media links.
- [ ] Verify invalid/missing fallback, mismatched hashes, malformed SVG dimensions and unsupported SVG structure fail with actionable errors.
- [ ] Confirm source PNG/JPEG, crop/contain/stretch, alt text, transparency and existing image specs remain unchanged.
- [ ] Compare outline XML against UHG28: one rectangular picture geometry and one inline picture line; check color and explicit width/dash behavior.
- [ ] Add structure tests for both image forms and re-run the complete nonnative probe suite.
- [ ] Only after the code and structure checks pass, open a generated fixture in PowerPoint, save/export it natively, and visually compare SVG rendering, fallback display, and outline joins against the source reference.

This document records the preimplementation audit. No PowerPoint automation was used during that audit.
