# Text and media sizing survey

## Coverage

Across **101 registered templates**, the survey records **1,357 text-bearing source objects** and **278 picture/media candidates**, all 278 with placed frame bounds. Text records join semantic slot names to source object IDs. One table object may contain many cells; these counts are not counts of fully resolved independent text zones. Only objects exposed through the main template contracts contribute text records.

**185 records have low-confidence drafting estimates.** Remaining records retain source character counts and available geometry/typography. Missing estimates are deliberate: mixed or inherited fonts, missing explicit insets, table containers, media placeholders, rotation, unresolved geometry or insufficient height make a simple box estimate unreliable. Native vector icons and grouped artwork are not exhaustively classified by the picture/name-based media pass.

The per-template records are in [sizing.json](sizing.json); regenerate with `python3 scripts/survey-template-sizing.py` from the repository root and the local source corpus available.

## Concrete examples

Characters include spaces in the recorded source runs, without adding paragraph separators. The drafting ranges below are planning hints, not maximum capacities.

| Template / zone | Placed frame | Source copy | Drafting guidance |
|---|---|---:|---|
| T077 / UHG 5 title | 5.045 × 2.070 in; observed 24pt | 132 characters | Keep the observed density as a reference; no automatic range because insets are unresolved. |
| T077 / UHG 5 partner takeaway | See object 36 in JSON | 203 characters | Roughly 116–210 characters under the heuristic assumptions. |
| T068 / Lab 4 person name | 2 × 0.306 in; observed 14pt | 8 characters | Roughly 9–16 characters as an initial drafting target. Longer names require actual wrapping review. |
| T068 / Lab 4 biography | See object 12 in JSON | 185 characters | Roughly 82–173 characters; the source exceeds this conservative range. This is evidence to calibrate the estimate, not proof of overflow. |
| T068 / Lab 4 portrait placeholder | 2 × 1.833 in | “PHOTO” placeholder | Media frame, not a five-character writing zone. |
| T045 / evidence comparison table | 12.323 × 3.969 in outer frame | Multiple cells | No whole-table character budget. Derive separate cell bounds and typography before setting cell budgets. |

## Estimation method and limits

Geometry comes from transformed OOXML source frames via the existing geometry exporter for registered decks, supplemented by direct, top-level binding-guide frames for the Lab source. Text counts aggregate source runs per object. Explicit font sizes/faces are observed properties, not a complete resolution of theme, master, paragraph and run inheritance.

Where eligible, the heuristic subtracts explicit text insets and spans mean glyph advances of 0.45–0.60em and line heights of 1.15–1.35em, retaining 55–75% of the resulting geometric character count. These assumptions are deliberately low confidence: actual glyph widths, bullet indents, paragraph spacing, font substitution and wrapping are not measured. A uniform observed font candidate does not establish identical effective formatting for every run. Records flag when source copy exceeds the estimated range. No new native rendering was performed for this survey.

Before exposing these as agent drafting guidance, calibrate family-specific ranges against representative native renderings. Show the source count alongside each range, and preserve an explicit unknown rather than presenting an unsupported hard limit. Dense variants deserve their own calibration rather than a universal padding or character policy.

## Media sizing and consolidation

Repeated picture-frame sizes include **0.50 × 0.50 in (37 occurrences)** and **0.75 × 0.75 in (10 occurrences)**. A **1.74 × 1.74 in frame occurs 13 times**. These are observed dimensions rounded to hundredths, not approved universal icon sizes. Repeated logos and reused artwork bias the counts; square geometry alone does not identify an icon.

A useful media-zone contract should distinguish:

- **Frame:** position, width, height, aspect ratio and relationship to neighboring text.
- **Asset:** intrinsic dimensions/aspect ratio and intended role (photo, portrait, icon, logo, illustration or artifact thumbnail).
- **Placement:** contain versus cover, crop/focal point, transparency/visible artwork bounds and clear space.
- **Confidence:** measured frame, classified asset, reviewed crop, or unresolved.

The current export resolves frames. It does **not** yet resolve picture crop/fit, intrinsic raster dimensions, transparent padding or vector optical bounds. Photo placeholders are identified separately from source pictures and naming-based media candidates where the source allows it.

## Recommended next increment

1. Attach source counts, frames and confidence to template/component discovery and the galleries.
2. Resolve effective fonts and text insets for frequent family variants; split table records into cell zones.
3. Calibrate ordinary and dense text ranges by family, keeping explicit line/bullet counts alongside characters.
4. Resolve picture crop/asset identity and classify native vector groups before standardizing media sizes.

These are survey artifacts. No template contracts, runtime defaults or installed release were changed.
