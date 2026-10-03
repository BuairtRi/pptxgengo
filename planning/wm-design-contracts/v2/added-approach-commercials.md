# Added Approach and Commercials contracts — slice 3

This slice covers all 16 added Approach templates and all three added Commercials templates from frozen WMDS commit `7bdcaee5030a12275a1f881a8542f4d302d207df`. The source bundle and v1 contracts remain unchanged.

## Closed content API

Each template uses `pptxgengo.wmds-library-bindings.v1`: supply every catalog `values.slots` entry and every `values.keys` array with exactly its fixed count. Stable caller keys change native item identities. Text is bounded by fixed source geometry; geometry and styles are not caller slots. Empty new-ledger cells remain explicitly supplied empty strings. Checkbox cells require booleans; status, RACI, timeline coordinates and allocation values retain their existing strict component contracts.

Navigation variants require caller-owned `values.nav`: 2–6 keyed labels, ordered by the caller, plus one active key. Split variants inherit the reviewed frame zones and authored title-line reservations.

| Template | Slots | Key arrays | Changed alternate slots |
|---|---:|---:|---:|
| `pricing/capacity-split` | 38 | 6 | 7 |
| `pricing/options-nav` | 39 | 6 | 19 |
| `roadmap/staggered-phases-nav` | 72 | 10 | 11 |
| `runbook/cutover-timeline` | 38 | 10 | 12 |
| `runbook/escalation` | 28 | 2 | 3 |
| `runbook/escalation-flow` | 25 | 0 | 2 |
| `runbook/escalation-split` | 16 | 8 | 2 |
| `runbook/go-no-go` | 43 | 2 | 21 |
| `runbook/overview` | 38 | 7 | 19 |
| `runbook/roles` | 37 | 2 | 20 |
| `runbook/roles-raci` | 52 | 2 | 5 |
| `runbook/step-detail` | 26 | 5 | 17 |
| `runbook/step-detail-data` | 45 | 7 | 23 |
| `runbook/step-detail-gate` | 39 | 7 | 22 |
| `runbook/step-list` | 30 | 4 | 22 |
| `runbook/step-list-split` | 31 | 4 | 22 |
| `runbook/step-table` | 56 | 2 | 19 |
| `runbook/step-table-nav` | 56 | 2 | 19 |
| `scope/service-value-nav` | 31 | 8 | 23 |

Every array count and source pointer is recorded in each pair packet’s `binding-report.json`; no arrays are omitted or auto-resized.

## Reference content

Each template has a source-example slide and a meaningful alternate slide. Alternate content describes a synthetic regional wave 2 ledger release for Lakeview: regional actions, evidence, role holders, weekend windows, and release terms replace the original examples. Commercial examples change team monthly run-rate, engagement fees, durations and allocations. These are illustrative assumptions, not client facts or approved pricing.

The data and gate checklists illustrate partially executed work: completed checks use native ticks, incomplete checks retain native empty squares, and new-ledger amounts are filled only for completed rows. The go/no-go alternate changes both evidence and status/go values. RACI preserves one accountable role per step while changing step wording and caller row identities.

Roadmap headings and end-of-timeline bar labels have fixed capacity. The alternate uses concise regional wording; it does not hide or delete a source stage. Its onboarding bar remains under the authored handover phase and ownership lane. Oversized caller labels continue to fail with an explicit layout error.

## Named source amendment

`wmds.runbook-overview-escalation-capacity.v2` applies only to v2 `runbook/overview`. The four escalation chevrons grow from 54 to 72 pt, using the remaining 18 pt grid step to the compact-footer body bottom at 468 pt. Each original time label and role is arranged vertically in the same chevron instead of consuming the role’s line width horizontally. Canonical number/body sizes, text and the intended semibold role treatment remain intact. The time and role are editable native text objects with stable derivative IDs; the source bundle remains pinned.

The shared RACI badge renderer separately centers its measured label allocation inside the authored 18 pt square. This removes the prior artificial 12 pt text capacity failure without resizing the badge or changing RACI semantics.

`wmds.runbook-step-list-two-line-title-reservation.v2` reserves two heading lines in each fixed 108 pt tall-column row of `runbook/step-list-split`. Headings move 3 pt up; the owner and window lines move 27 pt down. The last window ends within the 468 pt compact-footer body bottom. This corrects native overlap with two-line alternate step names without changing content, type or column widths.

## Generation and review

Rebuild with:

```sh
python3 scripts/wmds-refresh-approach-commercials.py --cli ./pptxdesign
```

The script selects the 19 additions from the frozen source audit, runs a focused `library-bound-sweep`, binds each explicit source/alternate pair, emits per-template receipts and combines all 38 compiled slides in `combined.foundation.json`. Each pair’s content, compiled foundation, binding report and layout report remain available for diagnosis.

Compilation and Go generation are structural evidence. Native PowerPoint review is complete for the 38 paired specimens (integrated pages 1–38): all pages were inspected in contact sheets, with full-size RACI/overview/checklist/scope checks. Final corrected pages 7–8 and 27–28 passed full-size recheck. The durable receipt is `samples/wmds-refresh-slice3-20261002/work/approach-commercials/native-review.json`; it records resolved outcome-outline and title/owner overlap findings. This acceptance is limited to the paired specimen content. No tests were added or run. Arbitrary-content fit, automatic continuation, font shrinking and changed array counts are not qualified.
