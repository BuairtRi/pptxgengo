# Fresh native font calibration

The 393-case PowerPoint capture completed on October 1, 2026, with full native
validation and provenance. It covers **Arial, IBM Plex Sans and IBM Plex Mono**,
using matching font files, real styles and variable axes. The source engine now
reports **`go-text-prototype.v4`**. Font selection remains general; measurement
code contains no family-specific size or spacing constants.

The subsequent [45-case focused capture and native example verification](font-native-focused-results.md)
also completed successfully. All 14 text shapes in the exact two-slide example
match native line breaks and fit their safe zones. Expanded terminal sizing and
tight wrapping remain documented review conditions; no further capture is
required before a template pilot.

## Results

| Comparison | v3 | v4 |
| --- | ---: | ---: |
| Matching visible line breaks | 391/393 | 391/393 |
| Heights within 0.01 pt of native character union | 385/393 | 385/393 |
| Matching visible line tops, where wrapping agrees | 815/832 | 832/832 |
| Largest height error in the six expanded-spacing cases | 6.460 pt too short | 3.225 pt too tall |
| Height review warnings, original 276 cases | 160 | 6 |

The height match count is unchanged because the six expanded-spacing cases
still need a terminal-box rule, and two cases have different wrapping. v4
corrects the line allocation contract and removes the observed height
underprediction in the expanded-spacing cases. It intentionally retains a
conservative full final allocation for nondefault spacing, with a review
warning. This is not a proven upper bound for other fonts, sizes, styles or
spacing multiples.

Default-spacing character heights agree for fractional sizes, mixed sizes,
all four styles, blank lines and all three reference families. All 832 visible
line tops agree within 0.01 pt where the native and Go wrapping matches; those
832 rows span 391 cases. The engine's glyph baselines are unchanged from v3,
so the previous 441-gap PDF comparison remains unchanged (average absolute gap
error 0.3535 pt, maximum 1.2 pt). PDF glyph baselines and native character boxes
are separate contracts.

The 393 cases are bounded reference evidence. They do not establish arbitrary
font parity or final native fit for a generated deck. Some observations were
used to revise the engine, so this entire corpus is not an independent final
validation set. Go reports continue to set `powerpoint_verified=false`.

## Corrected measurement contract

PowerPoint's non-whitespace character boxes include leading and paragraph
spacing allocations. Their top is not the top of glyph ink:

- Default line allocation is 1.2 times the largest shaped point size.
- Line spacing multiplies the allocation. The character box starts at the
  allocation origin; the predicted glyph baseline retains 75% of extra leading
  above it and a 0.9 em baseline within the default allocation.
- Space before belongs to the first line's box, except in the first paragraph,
  where PowerPoint suppresses it. Space after belongs to the previous
  paragraph's last line box when another paragraph follows.
- The last paragraph's space after is excluded. Interior blank lines advance
  layout; trailing blank lines have no occupied character box.
- Native expanded-spacing **terminal** boxes are shorter than a full
  allocation. v4 currently uses the full allocation and warns.

At 16 pt and 1.5 spacing, an ordinary allocation is 28.8 pt. A following
paragraph's 4 pt space before is included in its first character box. The
captured final box is 25.66 pt for Plex Sans/Mono and 25.57504 pt for Arial,
before adding space before. The existing corpus only measures this terminal
behavior at 16 pt regular. A family-specific correction would not establish a
general layout rule.

Horizontal geometry is unchanged. Go fit includes shaped ink and bullet
glyphs; the native character union excludes bullet glyphs and measures advance
boxes. The 12 bullet requests consequently have different horizontal unions
(by up to 16.036 pt). Elsewhere, matching-wrap widths differ by up to 1.789 pt,
partly from italic ink overhang and shaping/position differences. These contracts
must be considered when interpreting the comparison.

The original fit report still has three Go overflows: two Plex Mono 24 pt
paragraphs genuinely exceed the 400 pt height (403.2 pt in both Go and native);
the Arial long-word case includes a Go ink overhang (120.1875 pt versus a
119.9950 pt native character union in a 120 pt box). An accepted native probe
can overflow its source zone; probes intentionally use a 450 pt high frame.

## Focused capture (completed)

A frozen [45-control packet](../../samples/font-native-focused-capture-20261001/README.md)
completed successfully. It isolates spacing at 16/24 pt, regular/bold/italic, single/wrapped
paragraphs, and space before/after independently. It also samples both narrow
wrapping boundaries and captures a wide label for each:

1. Plex Sans regular 16.5 pt, `A clear typography plan`, width 174.578125 pt:
   Go predicts one line, PowerPoint two.
2. Arial regular 26 pt, `Roadmap AV office 2027`, width 285.825 pt:
   Go predicts one line, PowerPoint two.

Per-character positions reveal differences from rounded Go shaping around
kerning pairs. Their complete cause and PowerPoint's boundary decision are not
yet established; the engine retains its existing boundary warnings and makes
no family-specific width patch.

The completed packet was run from normal local Terminal, which has demonstrated live PowerPoint access:

```sh
bash /Users/rscott/Projects/pptxgengo/samples/font-native-focused-capture-20261001/capture.sh
```

The run captured 45 controls and **separately verified the two-slide example
built with v4**. The focused receipt is saved before example verification, so
failure of the final example check cannot hide the calibration capture. The
restricted agent shell still cannot reach the live app. The packet avoided
rerunning the full 393 cases. The example passed final native verification;
see the [focused results and remaining limits](font-native-focused-results.md).

## Evidence and reproduction

- [Full comparison](font-native-calibration/comparison-v4.json).
- [Calibration receipt](font-native-calibration/proof.json).
- [Successful capture receipt](font-native-calibration/capture-receipt.json).
- [276-case native evidence](font-native-calibration/reference/native.json) and
  [117-case native evidence](font-native-calibration/additional/native.json).
- [Prior controls and PDF baseline comparison](font-native-calibration/pdf-and-control-comparison-v4.json).
- Archived specs, manifests, PPTX probes, v3 predictions and v4 predictions are
  alongside each native evidence file. The three trailing-break preflight
  repeats are archived separately and excluded from the 393 unique-case count.

The comparison checks evidence/receipt/spec/manifest/deck hashes, request order,
source text, coordinate scale and converted bounds. Native validation checked
frames, text, effective styles, paragraph properties, margins and bounds during
capture. Native font-file hashes and styles agree with Go. CoreText omits
variable axes at their defaults; the comparison reads the selected font's
`fvar` defaults to check those omissions explicitly.

```sh
python3 scripts/compare-font-native.py \
  --reference-go library/dynamic-components/font-native-calibration/reference/v4-go.json \
  --additional-go library/dynamic-components/font-native-calibration/additional/v4-go.json \
  --out /tmp/native-font-comparison.json
```

Outputs must be new. The CLI was compiled, reference comparisons and fit report
were produced, and the existing Go example was built. No unit tests were added
or run. Earlier reference/calibration artifacts remain frozen. The installed
release has not been updated; rebuild source to use v4.
