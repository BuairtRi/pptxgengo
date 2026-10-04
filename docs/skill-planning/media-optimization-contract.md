# Native media optimization contract

Status: Wave 1 implementation and package smoke checks, 2026-10-03. Native PowerPoint/PDF visual review remains pending. This implementation has not changed the retained sample decks. Detailed results: [Wave 1 review](wave1-review.md).

## Ownership and default behavior

The application owns media packaging. An agent or human author selects source assets; the Go build automatically reuses embedded payloads, derives suitable photographic JPEGs, and compresses the final package. Source assets remain unchanged on disk.

The production WMDS renderer calls `pptx.OptimizeMedia` after its typography and native-group passes. Its default is `pptx.DeliveryMediaOptions()`:

| Option | Delivery default | Meaning |
|---|---:|---|
| `deduplicate` | `true` | Reuse byte-identical media throughout the package |
| `resize_jpeg` | `true` | Derive photographic JPEGs when safe and worthwhile |
| `compression` | `true` | DEFLATE-compress ZIP parts |
| `pixels_per_inch` | `220` | Minimum intended resolution at the largest modeled placement |
| `jpeg_quality` | `90` | Go JPEG encoder quality for resized derivatives |
| `min_savings_percent` | `10` | Keep the original unless the derivative saves at least this much |

The public low-level `Presentation.Write`, `WriteTo`, and `WriteFile` defaults are unchanged. Callers that need this policy can call the exported optimization API on completed PPTX bytes. Its zero-value policy disables deduplication, resizing, and compression; delivery defaults are explicit rather than silently applied to legacy callers.

```go
optimized, receipt, err := pptx.OptimizeMedia(raw, pptx.DeliveryMediaOptions())
```

An optional complete `media_optimization` policy on a WMDS foundation or bound-template document replaces the default policy. Template binding carries it into the compiled foundation document. Boolean fields are explicit: omitted booleans in a supplied policy are false. Zero PPI/quality values resolve to 220/90. PPI must be 72–1200, quality 1–100, and minimum savings 0–99. Authors should start from the full policy when changing one option.

```json
"media_optimization": {
  "deduplicate": true,
  "resize_jpeg": true,
  "compression": true,
  "pixels_per_inch": 300,
  "jpeg_quality": 95,
  "min_savings_percent": 10
}
```

The proposed single-file YAML contract uses the same complete policy fields. If the policy is omitted, its adapter will resolve and record the delivery defaults. Named profiles or partial-override conveniences can be added later without changing the resolved renderer policy.

## Deduplication

1. Read the final package and hash each `ppt/media/` payload.
2. Group identical bytes with compatible extensions and declared content types. `.jpg` and `.jpeg` are compatible for `image/jpeg` and its `image/jpg` alias used by the existing writer.
3. Aggregate placement requirements across all uses of each original payload.
4. Produce a derivative if applicable, then also reuse any identical resulting payloads.
5. Retain the first compatible media part in package order and redirect relationship targets from all package relationship parts. Preserve relationship IDs and external targets.
6. Remove obsolete media parts and their explicit content-type Overrides. Default extension declarations may remain harmlessly unused.

This includes slide, layout, and master media. SVG representations and their raster fallbacks keep their relationship IDs and payloads; exact duplicates may share media parts. Audio and video may also share byte-identical parts with compatible content types. No pixel-equivalence or perceptual deduplication is attempted.

## Safe JPEG derivatives

Requirements come from picture extents in slides, layouts, and masters, adjusted for crop fractions and nested group scaling. Rotated groups and pictures use conservative maximum-axis bounds where needed. For a payload used several times, the largest requirement wins. Derivatives preserve aspect ratio, never enlarge, and must satisfy both required pixel dimensions.

A JPEG is resized only when all modeled uses are safe, the scale is below 95% of the original, and the encoded derivative reaches the byte-saving threshold. The Go `image/jpeg` decoder/encoder and `golang.org/x/image/draw` Catmull–Rom resampler operate in process. No AppleScript, image utilities, external services, or new module dependencies are involved.

Large images use the same cubic kernel through Transform when the separable Scale intermediate would exceed 16 MiB. This trades some CPU time for lower temporary memory without using a cheaper sampling kernel. STORE ZIP payloads use read-only views of the input with CRC validation; decompression validates declared sizes and checksums. Packages above 2 GiB of uncompressed parts and individual parts above 256 MiB are rejected before allocation. The optimizer still operates in memory rather than streaming.

Original bytes are preserved when:

- A reference has unknown geometry, including background/shape fills and other unmodeled usages.
- A picture is tiled, has a nontrivial stretch fill rectangle, or has an unsupported crop.
- It is a vector fallback or has a DrawingML extension on its blip.
- The JPEG is unreadable, exceeds the 64-million-pixel decoding guard, or already has sufficient resolution.
- Embedded ICC/Adobe color data, CMYK/YCCK components, or non-default/uncertain EXIF orientation makes re-encoding unsafe.
- The extension and declared content type do not identify a JPEG picture.
- Re-encoding fails or offers too little saving.

PNG, GIF, TIFF, BMP, SVG, and other formats are preserved byte-for-byte and only deduplicated. In particular, alpha channels, vector editing, lossless logos, and animated images are not flattened. JPEGs containing color profiles may therefore remain large; color-managed derivatives are a future extension. Asset classification and lossless resizing of photographic PNGs are also deferred.

The optimizer refuses signed packages rather than invalidating an existing signature silently. Malformed package XML and ambiguous duplicate part names fail the build. Slide XML, object names, geometry, text, and native group identity remain unchanged.

## Receipt and future collaboration

The layout report contains `media_optimization`, with schema `pptxgengo.media-optimization.v1`:

- The resolved policy and input/output package hashes and sizes.
- Input/output media counts and total uncompressed media bytes.
- Every original media part, retained output part, source/derivative hashes and sizes.
- Original/output pixel dimensions and modeled pixel requirements for JPEGs.
- Whether the part was resized or reused and why it was retained or transformed.

The existing source-asset records and picture provenance preserve canonical source file identities. Receipt `source_sha256` refers to the embedded bytes immediately before this optimizer; an earlier crop, grayscale, recolor, or raster fallback can already differ from the canonical source bytes. Associate those cases through the picture relationships and canonical provenance rather than assuming the two hashes always match. The future object/asset map must record both transformations. Rebuild at another quality from the original source, not from an already reduced image in an edited PPTX.

An identical input package and policy produce the same transformation decisions and derivatives. This does not make upstream PPTX metadata timestamps and generated IDs deterministic. The output hash in this receipt is the final WMDS PPTX hash.

## Remaining acceptance work

- [x] Explicitly requested checks of package relationship integrity, crop/rotation/group sizing, and preservation cases.
- [ ] Native PowerPoint review of representative photo, SVG fallback, transparency, and layout/master specimens.
- [x] Measure the optimized full-library deck without replacing the current reference deck.
- [ ] Confirm delivery quality and agree whether an additional high-quality/offline profile is needed.
- [ ] Connect YAML policy resolution and project derivative/baseline packaging.
- [x] Review large-input memory behavior, reduce temporary allocations, and add resource limits.
- [ ] Introduce streaming packaging and whole-build memory budgets as larger projects require.

The earlier 78.5 MB estimate covered deduplication plus ZIP compression without resizing. It is an estimate, not a measured result from this implementation.

The measured delivery result is 31,550,932 bytes from 196,841,070 bytes (560 → 111 media parts; eight unique JPEG derivatives). After memory fixes, optimizer time was 12.58–13.11 seconds and peak process resident memory was 650–746 MB across two runs on this Mac. ZIP/XML, internal relationship, unchanged-content, and receipt-hash checks passed. These measurements establish package correctness and resource behavior; native visual acceptance is still pending.
