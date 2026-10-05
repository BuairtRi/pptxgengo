# Workshop native visual review — 2026-10-05

**Accepted: 14/14 final workshop specimens.** Source and editable bound compositions pass Go generation and content round trips. This review covers the new workshop family, not a fresh native review of every earlier template.

The deck was exported locally in Microsoft PowerPoint using **File → Export → PDF → Best for printing**. The existing `internal/nativeexport/pdf.swift` rasterizer produced 1920 × 1080 slide images. Final artifact hashes and page/template identities are in [the unsigned manifest](native-review/review-manifest.json).

## Repairs and review

Initial Go generation rejected five templates: schedule row heights, due-date column widths and grouped table width allocation. The candidate's six named geometry amendments resolve these while preserving the authored copy, fonts, values and immutable upstream source. Exact changes are in [qualification.json](qualification.json).

The first native pass accepted 11 pages and found three defects: slide 10's last Open questions bullet crowded the Next steps header; slides 13 and 14 split COMPLETE and SCHEDULED at their final character. Moving two slide-10 groups upward 18 pt and giving the status columns 18 pt more width resolved these. Both reviewers accepted the three repaired pages. The other eleven final native PNGs are byte-identical to the already accepted initial pages.

| Slide | Template | Final status |
| --- | --- | --- |
| 1 | workshop-overview/calm-exec | Accepted |
| 2 | workshop-overview/classic | Accepted |
| 3 | workshop-overview/dense-leave-with | Accepted |
| 4 | workshop-overview/nav-series | Accepted |
| 5 | workshop-overview/split-timeline | Accepted |
| 6 | workshop-plan/run-of-show | Accepted |
| 7 | workshop-plan/run-of-show-split | Accepted |
| 8 | workshop-readout/full | Accepted |
| 9 | workshop-readout/prioritized | Accepted |
| 10 | workshop-readout/split-learnings | Repaired and accepted |
| 11 | workshop-readout/voices | Accepted |
| 12 | workshop-series/cards | Accepted |
| 13 | workshop-series/phase-groups | Repaired and accepted |
| 14 | workshop-series/table | Repaired and accepted |

Review covered schedule alignment, agenda/pre-work clearance, readout columns and next steps, source notes, grouped rails, status labels, priority chips and footer clearance. No further visual defect was found in these final specimens. Actual replacement copy still needs fit review.

## Qualification limits

Fresh automated exports under the restricted caller fail dispatch `-10827`. The final GUI export therefore proves native appearance but does not qualify the automated doctor/render path. The production gallery/index and global CLI are not yet updated to 616.
