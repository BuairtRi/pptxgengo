# Go text height and line spacing calibration (v3 history)

The following records the earlier v3 calibration. The source now uses v4,
following the [completed 393-case native capture](font-native-calibration.md).
Fresh native measurements supersede the v3 character-box leading model below;
the glyph baseline model and frozen evidence remain unchanged.

At this stage the source engine reported **`go-text-prototype.v3`**. It uses a shared
**1.2 em character line box** for fit height, with a **0.9 em baseline** and
separate line advance. These rules apply to every resolved font; there are no
family-specific height constants. Generation remains entirely in Go.

## Results and limits

| Comparison | Evidence | Before | v3 |
| --- | --- | ---: | ---: |
| Largest height difference | 10 saved native character-bound controls | 4.8375 pt | 0.0000023 pt |
| Average line-gap difference | 441 PowerPoint PDF baseline gaps | 1.5195 pt | 0.3535 pt |
| Largest line-gap difference | Same 441 gaps | 3.3788 pt | 1.2000 pt |
| Matching line breaks | Two frozen PDF corpora, 393 cases | 391 | 391 |

The ten native controls were **used to derive the height rule**. They cover
Arial and IBM Plex Sans at 12, 16, 18, 26 and 28 pt, single lines and two
three-line paragraphs, and Plex regular/bold/italic/bold italic runs. This is
calibration evidence, not independent validation or arbitrary-font parity.
Font families, styles and file hashes agree with the saved native environment.
The probe frames establish a 1:1 mapping from raw native coordinates to points.

The PDF comparison includes **IBM Plex Mono**, all three families, mixed sizes,
paragraph spacing, blank lines, bullets and fractional sizes. It compares 297
gaps in the original corpus and 144 in the additional corpus. The first-to-last
baseline span also has a maximum difference of 1.2 pt in each corpus.
Two known wrapping differences prevent one-to-one baseline comparison and are
explicitly skipped. Text, advances, clusters and wrap warnings are unchanged
for all 393 cases; 391 match PowerPoint's visible lines.

**PDF baselines do not establish native character-box height.** Mono native
character bounds, fractional-size character bounds, mixed-size character
bounds and nondefault paragraph spacing still require native capture. Fresh
AppleScript capture failed in this restricted session; no new TextRange capture
or native final-deck verification is claimed. The existing native evidence
and previously captured PowerPoint PDFs were reused, with their hashes checked.
The subsequent [capture diagnosis](native-capture-diagnostics.md) localizes the
failure to the sandboxed shell's application registry and GUI process access.

## What the engine measures

| Quantity | v3 meaning |
| --- | --- |
| `height_pt` | Character line box: 1.2 × the largest shaped point size on the line |
| `layout_advance_pt` | Line box height × authored line spacing multiple |
| `y_pt` | Line box top, including 75% of extra leading above the box |
| `baseline_pt` | Line box top + 0.9 × the largest point size |
| `ink_bounds` | Shaped glyph outlines placed on the predicted baseline; a separate diagnostic |
| `rendered_height_pt` | Union of occupied character line boxes, excluding trailing blank lines and paragraph spacing |

Font ascent, descent and gap describe font metrics. In v2 those metrics also
determined fit height, so Plex and Arial produced different heights at the
same point size. The saved native character boxes instead measure 19.2 pt at
16 pt and 57.6 pt for three lines in both families. v3 reproduces that contract.

Paragraph spacing affects the distance between rows. The PowerPoint probe
exports suppress space before the first paragraph; v3 does the same. Space
after the last paragraph does not enlarge the occupied character union.
Interior blank lines advance layout; trailing blank lines have no occupied
character box. The original hard breaks remain in the line details.

Extra leading is separated from the character box. At 16 pt, a 1.5 spacing
multiple advances a line by 28.8 pt, retaining a 19.2 pt character box. Its box
starts 7.2 pt below the line's allocation origin. This leading distribution is
supported by the PDF baseline references; its native character-bound contract
still needs capture. Baselines remain predictions, with observed PDF gap
differences up to 1.2 pt. Their cause has not been fully established.

Glyph ink no longer expands **fit height**. This prevents exported font
selection rectangles or outline bounds from being treated as native character
bounds. Ink remains in the report and retains the existing horizontal bounds
behavior. A vertical ink overhang generates a review warning.

## Review warnings

Each request has `height_warnings` when it uses a family beyond the native
Arial/Plex Sans controls, fractional sizes, mixed sizes, blank lines,
nondefault leading/paragraph spacing, or ink outside the predicted character
box. Top-level `warnings` repeat these messages. `fit-report` adds
`height_review_warning_count`, alongside the existing wrap warning count.

These warnings identify the limits of the available evidence. They do not
change planner pass/fail. Go reports continue to set
`powerpoint_verified=false`; native final verification remains separate.

## Evidence and reproduction

- [Complete comparison](font-height-calibration/comparison-complete.json).
- [Native control evidence](font-height-calibration/native-control-evidence.json)
  and [probe manifest](font-height-calibration/native-control-manifest.json).
- [Calibration receipt](font-height-calibration/proof.json).
- [Original PDF baseline origins](font-height-calibration/reference-baselines.json)
  and [additional PDF baseline origins](font-height-calibration/additional-baselines.json).
- [Original-corpus wrapping comparison](font-height-calibration/reference-wrap-comparison.json)
  and [additional-corpus comparison](font-height-calibration/additional-wrap-comparison.json).

The wrapping comparison files still include PDF selection-height differences
for diagnostics. Their `go_height_underprediction_count` does **not** measure
native fit: Plex PDF selection rectangles are larger than native character
boxes. Use the character-height and baseline contracts in the complete
comparison for this calibration.

`scripts/read-font-reference-baselines.swift` reads PDF text-show origins using
CoreGraphics, in PDF page-tree order, without resolving fonts through a font
service. It handles simple horizontal probe text and rejects XObjects,
rotated/skewed text and shorthand text operators. Origins are grouped into
rows only after checking against the saved PDF selection row count. Swift is
a one-time reference extractor; the Go generation path does not use it.

```sh
go build -o /tmp/pptxcompose-height ./cmd/pptxcompose
/tmp/pptxcompose-height measure --engine go \
  --spec library/dynamic-components/fonts.json --out /tmp/height-control.json
/tmp/pptxcompose-height measure --engine go \
  --spec library/dynamic-components/font-reference/corpus.json \
  --out /tmp/height-reference.json
/tmp/pptxcompose-height measure --engine go \
  --spec library/dynamic-components/font-wrap-calibration/heldout/corpus.json \
  --out /tmp/height-additional.json
python3 scripts/compare-font-height.py \
  --control-go /tmp/height-control.json \
  --reference-go /tmp/height-reference.json \
  --additional-go /tmp/height-additional.json --out /tmp/height-comparison.json
```

All outputs must be new. The CLI was compiled, the reference comparisons and
fit report were produced, and the Go example was built with v3. No unit test
suite was added or run. Original reference and v2 calibration artifact hashes
are unchanged. Rebuild the source CLI to use v3; the installed frozen release
has not been updated.

The next qualification step is fresh native character-bound capture for
Mono, fractional sizes, mixed-size text, spacing and trailing blank lines,
followed by native verification of a representative modern template deck.
