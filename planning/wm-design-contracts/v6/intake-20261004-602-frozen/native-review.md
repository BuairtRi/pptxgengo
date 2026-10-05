# v6 native visual review

**Accepted: 16 of 16 changed source specimens, 2026-10-05.** Each 1920×1080
native page was inspected individually at original detail. Root reviewed pages
1–8; root and the independent semantic reviewer reviewed pages 9–16.

The final editable [deck](native-review.pptx), [local PowerPoint PDF](native-review/deck.pdf)
and [unsigned review manifest](native-review/review-manifest.json) identify the
accepted artifacts by SHA256. PowerPoint File → Export → PDF → **Best for
printing** produced the PDF locally; the online distribution option was not
selected. `internal/nativeexport/pdf.swift` used macOS PDFKit to rasterize it.

The first native iteration exposed midword TECHNOLOGY/GOVERNANCE headers,
wrapped LATER chips, and table references split across two lines. The repaired
deck reserves native inline width for badges and priority labels. Positive-min
heat headers reduce symmetric padding and, when needed, use 8pt instead of 9pt
Mono text. The final native pages show complete single-line headers, LATER and
A1–A9 labels; active A4–A6 digits remain inside the navy fills.

| Page | Template | Acceptance evidence |
|---|---|---|
| 1 | capability-heat/dense | Accept: headers, dense copy, heat cells and margins clear. |
| 2 | capability-heat/grouped | Accept: group rails, headers and row copy clear. |
| 3 | capability-heat/grouped-split | Accept: split clearance, group rails and copy clear. |
| 4 | heat-notes/full | Accept: whole-word heat headers and note rows clear. |
| 5 | heat-notes/full-continued | Accept: whole-word headers, continuation and note rows clear. |
| 6 | heat-notes/grouped | Accept: group rails, whole-word headers and notes clear. |
| 7 | heat-notes/now-next-later | Accept: NOW/NEXT/LATER chips and whole-word headers clear. |
| 8 | heat-notes/split | Accept: split copy and margins clear. |
| 9 | heat-notes/split-last | Accept: two-line headline, continuation, evidence and legend clear. |
| 10 | heat-notes/summary | Accept: TECHNOLOGY/GOVERNANCE single-line; P1–P4, evidence and margins clear. |
| 11 | heat-ref/cards-split | Accept: A7–A9 centered; card evidence and selected locator outline clear. |
| 12 | heat-ref/detail-continued | Accept: A6 centered; continuation copy and locator row A6 clear. |
| 13 | heat-ref/detail-corner | Accept: A4/A5 centered; evidence and complete ROWS A4–A5 caption clear. |
| 14 | heat-ref/locator-split | Accept: A1–A3 centered; locator outline, labels and evidence clear. |
| 15 | heat-ref/overview | Accept: A1–A9 single-line; active digits visible; heat headers and detail card clear. |
| 16 | heat-ref/overview-grouped | Accept: same badge/header repairs; group labels centered and clear. |

The earlier CLI export succeeded for source SHA256
`e0b88262dc689692fc60921c06f94b0116eda12eb9f801bb109d3396b44d7065`, before
the repairs. Its [original signed manifest](native-review/initial-cli-render-manifest.json)
is historical evidence, not acceptance of the final deck. Fresh automated
export of the final SHA256 `9b5ab2e0a675ba734f8e77279d521a77145c8eae73b6927d62201de7cf652376`
remained blocked by `-10827` under the restricted caller. This GUI visual pass
does not claim a doctor pass or final automated export success.

The review covers these source specimens and their current copy. Semantic
metadata remains inferred, arbitrary edits still require fit/review, and the
production default remains frozen v5.
