# Library review corrections — 2026-10-03

These amendments follow the user's review of the 167-slide Slice 4 library.
The frozen v2 source files and content bindings retain their existing hashes.
Geometry changes are recorded as named Go adapter refinements after binding.

| Library slide | Template/component | Change |
| --- | --- | --- |
| 28 | `phases/four` / phases | Paint the axis before the gate diamonds. |
| 29 | `phases/summary-bands` | All three chevrons are 270pt wide, matching their column cards. |
| 30 | `plan/gantt` | Use 24pt activity tracks inside 42pt minimum lanes and 60pt double-track lanes. The full-width 8pt legend sits below the table. Body type and footer chrome are unchanged. |
| 38–39 | `roadmap/staggered-phases`, nav variant | Full-width, single-row 8pt legend with proportionately smaller swatches and gaps. |
| 43 | `runbook/escalation-split` | Three straight hand-drawn arrows rotate upward and occupy the centers of the 36pt gaps, leaving 6pt above and below. |
| 44, 115–119 | Shared table/card status square | One filled native rectangle with a uniform 1pt stroke, inset by half its width to preserve the original outside bounds. |
| 87 | `frameworks/layer-table` | Header rules start 12pt after each measured label and end 12pt before the next column. A rule is omitted when the label leaves no positive rule width. |
| 88 | `from-to/cards` | Rotate the dashed arrow to 335° and center it horizontally between the cards. |
| 89, 121 | `from-to/rows`, `transformation/before-after` | Replace thick arrow shapes with the supplied straight hand-drawn SVG at its natural aspect ratio, centered in each row gutter. |

The new registered asset is `arrow-straight`, sourced from
`~/Documents/branding/assets/graphics/arrows/left-facing-arrow-navy.svg`.
Despite that filename, the SVG's visible arrow points right. Its canonical hash
is `115a8f5079b158c346e9f9ca6c6eeecb7567bbab70f4cd058c365f09ed10290e`.
The existing release assembler packages registered payloads from the branding
root; this addition brings the registry to 697 keys and 696 distinct paths.

## Generation and review state

The updated Go command builds successfully. Ten affected template source/alternate
pairs generate through closed content binding. The regenerated 167-slide deck
retains the prior library's illustrative custom content. Exactly 16 source slide
XML files change; the other 151 are byte-identical to the prior deliverable.
The three changed Gantt legends all allocate their labels on one row at 8pt.
No tests were added or run.

Current deliverable:
`samples/wmds-feedback-20261003/latest/WMDS-template-library.pptx`.
Composition, layout report and generation receipt are retained in the same packet.

**Native review is pending.** PowerPoint opened an intermediate regenerated deck,
but later CUA inputs repeatedly returned “The user changed Microsoft PowerPoint,”
including after fresh AX reads, app reconnection and runtime reset. Screenshot
capture also reported unavailable. No connected document-control session was
available. No native PDF was exported for the final corrected deck, and old native
acceptance is not transferred to these changed slides. The staged local.6 release
and its accepted gallery remain the prior Slice 4 snapshot.

Next: export the current deliverable using local PowerPoint's **Best for printing**,
inspect all 16 changed pages, then refresh the affected paired review receipts,
gallery and staged release. This is visual artifact review, not a font calibration
round or a dependency on native measurement during generation.
