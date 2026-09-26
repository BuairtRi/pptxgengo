# pptxanchor

This read-only CLI computes an OOXML image placement (`x`, `y`, `cx`, `cy` in EMU) that aligns the visible alpha bounds in a `scripts/svg-visible-bounds.swift` report to native PowerPoint text-range bounds. It does not edit a deck.

The examples below use local, ignored UHG measurement fixtures. Generate measurements
with `scripts/measure-pptx-text.applescript` and alpha bounds with
`scripts/svg-visible-bounds.swift` for another deck/asset. These fixture paths
are not present in a clean checkout.

```sh
go run ./cmd/pptxanchor \
  --measurement samples/reconstruction/work/pivotal-moment-bounds-v2.json \
  --asset samples/reconstruction/work/underline-visible-bounds.json \
  --units points \
  --container-aspect 1.64819110185159 \
  --stroke-padding-pt -1.2317322059691946 \
  --horizontal-offset-pt 0.11945163576578466 \
  --vertical-offset-pt -3.227455397478323
```

The placement formula derives normalized visible alpha bounds directly from the report's `viewBox` and `visibleBounds`. It sizes the image so visible width equals phrase width plus twice the signed stroke padding, fits image-container height using the explicit source container aspect ratio, and positions the visible top at phrase bottom plus the explicit vertical offset. The explicit horizontal offset allows a source optical adjustment. PowerPoint source geometry is in points for this adapter, with `12,700 EMU/point`; `--units points` is mandatory so raw bounds are not silently treated as points.

The parameter values above calibrate slide 5's `pivotal moment` underline to the supplied source shape geometry: `x=2101165`, `y=283028`, `cx=2219853`, `cy=1346842` EMU. They derive from the measured phrase bounds, the image77 alpha bounds, and that known source placement. The negative padding is a symmetric 1.2317322 pt visible-width inset; the horizontal offset is +0.1194516 pt; the vertical offset is -3.2274554 pt relative to the phrase's measured bottom. These are explicit calibration values, not inferred defaults. Reuse them on later phrase reflows only if the same underline styling and optical relation are intended.

The CLI emits `manual_required` without placement coordinates when phrase rotation is missing/nonzero or the phrase spans multiple line fragments. It rejects malformed or nonpositive dimensions.

## Highlight mode

Highlight mode fits the rotated alpha-bounds rectangle around the phrase's measured bounds. Horizontal and vertical padding expand the target rectangle independently; optical offsets shift its top-left. The source picture rotation is explicit and preserved in both degrees and PowerPoint's 60,000-units-per-degree `rot` value. The inverse extent solve scales image width and height independently so the rotated alpha rectangle's axis-aligned envelope matches the target rectangle. Missing/nonzero text rotation and multiline phrases require manual placement. The output carries `z_order_intent: behind_target_text`; the caller remains responsible for applying the object order in the slide XML.

For the UHG slide 6 highlight example:

```sh
go run ./cmd/pptxanchor \
  --mode highlight \
  --measurement samples/reconstruction/work/west-monroe-bounds.json \
  --asset samples/reconstruction/work/highlight-visible-bounds.json \
  --units points \
  --asset-rotation-deg 1 \
  --horizontal-padding-pt 5.058313518049971 \
  --vertical-padding-pt 5.220756405694651 \
  --horizontal-offset-pt 1.6051474964713748 \
  --vertical-offset-pt 0.18744265638906654
```

Those values were derived from the source Picture5 container and its rotated source alpha bounding rectangle, then reproduce the source placement exactly: `x=2355376`, `y=531327`, `cx=2028186`, `cy=463041` EMU at 1°. The same explicit values follow the phrase when it is reflowed: the resized measurement yields `x=2998599`, `y=536828`, `cx=2659504`, `cy=573960` EMU. The source PNG has ragged corners. Its pixel report's tight rotated-ink bounds are smaller than the rotated rectangle envelope; the solver fits the alpha bounding rectangle, which is the correct input for reconstructing the source image container. Native rendering still needs visual review.
