# Go wrapping calibration

This records the v2 wrapping work. The subsequent
[v3 height calibration](font-height-calibration.md) retains these wrapping
results and separately calibrates character height and line spacing.

The source Go engine now reports **`go-text-prototype.v2`**. It rounds each
shaped glyph advance to the nearest **1/8 point before line wrapping**. The
rule applies to every resolved font. It contains no IBM Plex or Arial
exceptions and uses no native application or script during measurement.

## Results

| Corpus | Cases | v1 matching line breaks | v2 matching line breaks |
| --- | ---: | ---: | ---: |
| Frozen three-family reference | 276 | 252 | 275 |
| Additional reference text and sizes | 117 | 89 | 116 |
| Combined | 393 | 341 | 391 |

There are no line-break regressions in either corpus. Font file hashes and
resolved styles/axes match between candidates. All source text survives the
native PDF capture, and every expected font family appears in the PDF font
resources. On the original set, all 201 non-boundary cases now match; eight
of the nine boundary cases predicting one line where PowerPoint uses two
are corrected.

The additional corpus contains 39 cases per family. It uses new paragraph,
label and long-word text at 13, 17 and 26 pt. Regular, bold, italic and bold
italic are included. Label widths are 0.2 pt below and above v2's advance.
The corpus and predictions were prepared before capturing its PowerPoint
output. It is an additional calibration check, rather than a broad sample
of real presentations.

The CLI was compiled, both corpora were measured and compared, `fit-report`
warning counts were checked, and the prototype example was built with v2.
No unit test suite was run in this task.

## Why the rule changes wrapping

The previous engine shaped at 64 times the requested size, then reduced
every advance to a 1/64 point value before wrapping. The captured PowerPoint
positions are consistent with a coarser 1/8 point grid. Summing many glyphs
on different grids changes which words fit.

For example, a Mono glyph's nominal advance at 16 pt is 9.6 pt. On the new
grid it is 9.625 pt, making some lines slightly wider. At 24 pt the nominal
14.4 pt advance becomes 14.375 pt, making some lines slightly narrower.
The observed PDF positions approximate these values; PDF text-position
encoding adds small differences. A constant padding adjustment would not
describe both directions.

The shaper retains kerning and its default ligature behavior. Point sizes,
glyph ink metrics and the font ascent/descent/gap height policy retain their
existing logic. Total paragraph height can change because wrapping changes
the number of lines. Native height calibration remains separate work.

## Remaining differences and boundary warnings

Two cases still predict one line where the captured PowerPoint output uses two:

| Case | Available width | v2 advance | Clearance |
| --- | ---: | ---: | ---: |
| `plex-sans-boundary-16.5-regular-plus-0` | 174.578125 pt | 174.5 pt | 0.078125 pt |
| `arial-heldout-26-regular-boundary-0.2` | 285.825 pt | 285.625 pt | 0.2 pt |

The cause of those remaining positioning differences has not been established.
No family-specific correction was added to force a match.

`RequestLayout.boundary_warnings` flags any nonempty line whose clearance
is within **glyph count × 1/16 pt**. Each warning records a zero-based line
index, clearance and review band. The top-level report repeats these in
`warnings`; `fit-report` includes `wrap_boundary_warning_count`. Both
remaining differences are flagged. The original corpus has 55 flagged lines.

This band is a review heuristic based on half an advance quantum per glyph.
It is not a proven bound on PowerPoint's error. Warnings do not change line
breaks or planner pass/fail decisions, and an absence of warnings does not
establish native parity. Reports continue to set `powerpoint_verified=false`.

## Evidence and reproduction

The original [reference set](font-layout-reference.md) and its recorded
artifact hashes are unchanged. Candidate results are stored separately:

- [Final frozen-corpus comparison](font-wrap-calibration/comparison-final.json).
- [Final additional-corpus comparison](font-wrap-calibration/heldout/comparison-final.json).
- [Calibration receipt and hashes](font-wrap-calibration/proof.json).
- [Additional source corpus](font-wrap-calibration/heldout/corpus.json),
  [probe PPTX](font-wrap-calibration/heldout/font-reference-probes.pptx),
  [PowerPoint PDF](font-wrap-calibration/heldout/heldout-native.pdf) and
  [PDF measurements](font-wrap-calibration/heldout/native-measurements.json).

Final comparisons supersede the preliminary `comparison-calibrated.*`
reports in this calibration directory. The original v1 binary hash is
recorded in the receipt; it was used to measure the additional source
corpus for the before/after comparison.

The additional PDF was exported locally through PowerPoint 16.113.3 on
macOS. Its probe PPTX was unchanged after capture. PDF line selections
remain distinct from native TextRange geometry and do not qualify height
or final native fit.

Rebuild the source CLI to use v2; an installed frozen release is unchanged.
Compare a subsequent candidate without writing into the frozen dataset:

```sh
go build -o /tmp/pptxcompose-wrap ./cmd/pptxcompose
/tmp/pptxcompose-wrap measure --engine go \
  --spec library/dynamic-components/font-reference/corpus.json \
  --out /tmp/font-wrap-candidate.json
mkdir /tmp/font-wrap-comparison
python3 scripts/font-layout-reference.py compare \
  --out library/dynamic-components/font-reference \
  --native library/dynamic-components/font-reference/native-measurements.json \
  --go /tmp/font-wrap-candidate.json --reports /tmp/font-wrap-comparison \
  --name comparison-final
```

Repeat against `font-wrap-calibration/heldout` for the additional corpus.
Use new output filenames/directories. The next calibration task is native
TextRange height and line spacing, while retaining the two recorded width
boundary cases as explicit limitations.
