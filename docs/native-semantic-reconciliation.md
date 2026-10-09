# Reviewed native semantic reconciliation

The v4.3.0 source work supports explicit semantic review for Gantt retiming,
pod membership, ordinal assessments and native chart values. These additions
are not included in the promoted v4.2.1 package. The operator entry point is
[native semantic review](../skills/west-monroe-presentations/references/native-semantic-review.md),
with specialized interpretation and source operations in each family runbook.

## Bounded contract

The command consumes a closed `reconcile propose --geometry` review packet.
It authenticates the immutable receipt/build inputs, rereads the closed packet,
replays text/geometry analysis, checks its report exactly, and requires current
source/toolchain to match the review and semantic baseline. The selected local
Gantt must already be explicitly materialized and have no active native layout
or order overrides. Shared/pinned local definitions retain the existing refusal
in the measured source transaction.

Every period-label anchor must retain its native transform. It establishes one
period of width; anchors must be uniformly spaced, share a parent and have all
group ancestors unchanged/unrotated/unflipped. Conversion uses that parent space,
not slide coordinates. Supported objects retain their exact receipt-backed
identity, parent, rotation/flips, vertical position and height. Only horizontal
position and width may change. No geometric inference creates dates.

| Object | Proposal | Limits |
| --- | --- | --- |
| Single solid interval task bar | New `from/to` | No hatching, progress or soft start/end segments; valid positive interval inside the timeline |
| Original gate vertical guide | New `at` | Zero-width guide; moving its chip/label alone is not sufficient |
| Segmented/event/phase/today geometry | No retiming proposal | Explicit family patch or subsequent narrower qualification |
| Copies, deletions, imports or style/topology changes | Existing manual/unresolved evidence | No inferred new task, lane or membership |

Coordinates are rounded to six decimal places in periods. There is no nearest
integer, date, weekday or label-based snapping. EMU quantization can influence
the last decimal; a subsequent explicit Gantt patch should establish intended
integer/calendar-derived positions when needed. This policy is included in the
semantic report hash.

A decision names semantic report SHA256, actor, reason and exact proposal IDs.
`retime` accepts only a `proposed` item; `keep_source` explicitly retains its
meaning. Unselected items and manual findings remain unresolved. Regeneration
updates coupled labels, packing, gate chips and guides. It does not also adopt
stale native transforms. Full fit measurement and source transaction checks
precede application. Exact original PPTX, geometry report, semantic report and
review-decision bytes are retained once in the shared content-addressed store;
the composition receipt includes their paths/hashes. Baseline build outputs and
lock bytes are checked again under the source mutation guard.

The rebuilt deck contains the selected semantic edits. Other native text,
formatting, geometry and topology edits remain evidence, not adopted content.
They are explicitly surfaced in unresolved IDs/manual findings. A successful
partial adoption is not full native-document synchronization; review the new
baseline before subsequent reconciliation.

## Team and reporting boundary

Pod crossing alone cannot distinguish reassignment from spacing, shared roles,
resized containers or transient dragging. `project team reconcile` proposes
membership only when the original role surface and text move together by the
same translation, every relevant pod/frame allocation is unchanged, and exactly
one target contains the moved role. Overlap, resizing, partial movement, changed
ownership and stale source remain unresolved. A reviewed `reassign` decision
changes membership and regenerates its layout; the source transaction retains
the native geometry as evidence rather than applying the movement twice.

Reporting changes require relationships rather than position. A stronger
signal could use authenticated native connector endpoint IDs, explicit edge
meaning and known node identity to propose a parent change. It must retain tree
root/cycle/multiple-parent guards. Native arrow movement, shape proximity and
untagged imports do not establish reporting authority. Current
[team commands](../skills/west-monroe-presentations/references/team-composition.md)
provide explicit reviewed membership and reparent operations.

## Assessment and chart facts

`project assessment reconcile` recognizes canonical visible integer scores
within the declared ordinal domain, or an empty cell as unassessed. Zero remains
a score. Header/row identities, formatting and table structure must retain the
authenticated baseline; a score decision regenerates dependent fills and legend.

`project quantitative reconcile` recognizes numeric facts in column, bar, line,
pie, doughnut and scatter charts only when their original keyed cache and
embedded workbook agree. The worksheet/reference/table shells, axes, labels,
units and styling retain their baseline. Native formatting, formulas, renamed
categories, copied charts or inconsistent cache/workbook edits remain manual.
Commercial values are derived outputs: edit explicit source inputs/formulas,
rather than adopting a manually typed monetary result as a new model input.

## Qualification

Focused Go tests edit tagged transforms in generated PPTX XML, propose fractional
task/gate periods, preview/apply guarded source facts, retain exact evidence and
rebuild. They cover vertical movement remaining manual, changed axes refusing
inference, stale source, strict decision identity and ambiguity. These are source
and XML fixture qualifications. They do not establish native PowerPoint GUI
editing/Save As, Windows qualification or universal catalog-family round trips.
