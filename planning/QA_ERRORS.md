# Reconstruction QA findings

This log records observed problems and limits. Pixel equality is always scoped
to the named renderer, resolution, and reference. It is not a claim about every
PowerPoint version or about generating new designs from prose.

| ID | Finding | Resolution / remaining work |
| --- | --- | --- |
| QA-01 | PowerPoint AppleScript export returned without creating a PDF when given a plain path string. | Pass a `POSIX file` to `save … in …`; require the PDF to exist before comparison. |
| QA-02 | Extracting/re-serializing source XML alone would overstate independent reconstruction. | Added v2 typed text, geometry, style, table and crop bindings, sentinel validation, mutation tests. Source object topology and master/layout resources are still retained. This is source-derived reconstruction, not inferred design. |
| QA-03 | A shortened deck changes automatic slide-number fields. | Explicit `--freeze-slide-numbers` converts them into editable text with source numbers. Report this intentional transformation. |
| QA-04 | A source slide rendered in a small deck can differ slightly from the same slide rendered in the full source deck. | Keep literal full-deck differences. Separately render an untouched source-slide control in the same small-deck context. Slide 11 control is exactly equal to its rebuilt candidate. This isolates the context effect without silently relaxing tolerance. Exact internal cause remains unproven. |
| QA-05 | Team slides link to omitted appendix slides. | Extract dependencies transitively and append destinations as hidden slides. Do not strip the links or silently redirect them. |
| QA-06 | The slide 5 underline SVG has about 46% transparent padding above and below its visible stroke. | Measure alpha bounds at high resolution; map the desired visible stroke to the full image rectangle. Image-box bounds are insufficient. |
| QA-07 | The slide 5 arrow has a source rotation of 345.7749167 degrees; its transformed visible extent is wider than its image canvas. | Include rotation and transform actual nontransparent raster-cell corners when measuring visible bounds. Unrotated bounding boxes are insufficient. |
| QA-08 | Earlier structural-only notes misidentified the slide 5 right panel and slide 8 imagery. | Native source rendering is mandatory. Metadata is supporting evidence, not a substitute for inspecting the rendered slide. |
| QA-09 | Detailed source slides include small text and dense placement; exact reproduction can preserve existing readability compromises. | Preserve source fidelity in this experiment; evaluate readability separately when adapting the layout to new content. |

## Review discipline

Each candidate is built, exported by native PowerPoint, rasterized at
1920×1080, compared, and visually inspected before moving to the next target.
Difference images exaggerate RGB differences four times; the JSON report
records literal differences, including very small background changes.
Source media remains source media: an existing architecture illustration is
retained as an image, while source-native text, tables, roles, connectors and
shapes remain native objects. No slide is replaced by a full-slide screenshot.

## Additional findings during final review

- **QA-10 — Coordinate schema mismatch, fixed:** native phrase bounds use
  `left/top`; asset alpha bounds use `x/y`. The initial anchor prototype decoded
  both as `x/y`. Root review caught this before applying the solver. Separate
  required-field types and a regression with the real measurement schema now
  prevent silent zero-origin placement.
- **QA-11 — Failed output cleanup, fixed:** extraction and build formerly could
  leave partial outputs on an ordinary error. Both now stage writes; tests cover
  broken relationships, missing bindings and pre-existing reports. The two-file
  PPTX/report commit is not guaranteed atomic across a process crash.
- **QA-12 — Unreliable preview claim, retracted:** one secondary visual review
  claimed missing content. Direct RGBA decoding and zero-difference controls
  contradicted it; the reviewer retracted the claim. Review conclusions must
  identify exact files and be reconciled with pixel/structural evidence.
- **QA-13 — Export differences persist in full-source context:** a second
  85-slide control made some targets exactly equal and changed faint background
  pattern rasterization on others. The underlying OOXML/resources are audited
  separately; no blanket full-reference pixel-perfect claim is made.
- **QA-14 — Experimental wording needed clearer labeling:** the slide 5
  underline-reflow fixture intentionally changes the title to force the phrase
  onto line two. The user noticed this during review. Original wording remains
  in `UHG-5-annotations-baseline.pptx`; the report labels the changed-title file
  as an experiment, not an exact recreation.

## Inventory review findings

- **QA-15 — Hidden-slide metadata location:** the modernization summary initially
  checked presentation slide IDs for `show` and reported zero hidden slides. Root
  review inspected each slide's `p:sld` root and found two hidden parts. Hidden
  status must come from the slide part and then be mapped into presentation order.
  The underlying raw inventory already retains these root attributes.
