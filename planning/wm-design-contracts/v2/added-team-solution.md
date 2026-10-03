# Slice 3: added Team and Solution templates

Frozen source: `wmds-library.v2`, commit
`7bdcaee5030a12275a1f881a8542f4d302d207df`.

This family slice covers all **19 additions**: 10 Team and 9 Solution.
Each has a source example and a changed-content example, yielding **38 slides**.
Generation succeeded for all 19 pairs. Native visual review is complete for these 19 pairs through the integrated
PowerPoint reference review. Generation does not qualify arbitrary
replacement content. No tests were added or run.

## Closed content contracts

| Template | Required slots | Fixed array key sets |
| --- | ---: | ---: |
| architecture/layers-nav | 26 | 2 |
| architecture/layers-split | 26 | 2 |
| bio-full/portrait-nav | 14 | 1 |
| bios/leadership-specialists-nav | 52 | 8 |
| bios/three-nav | 30 | 6 |
| process/current-future | 26 | 0 |
| process/hierarchy | 54 | 5 |
| process/metrics | 34 | 0 |
| process/metrics-nav | 34 | 0 |
| process/sipoc | 35 | 13 |
| readiness/six-criteria-split | 30 | 6 |
| roles/by-phase-nav | 43 | 3 |
| team/org-chart-nav | 20 | 4 |
| team/org-roles-nav | 40 | 9 |
| team/pods-nav | 23 | 3 |
| team/pods-pairs-nav | 36 | 5 |
| team/roster-nav | 48 | 4 |
| team/roster-split | 45 | 2 |
| workstreams/five-narrative | 32 | 5 |

The exact slot names, kinds, source pointers and key cardinalities come from
`library-catalog`. Counts exclude the separate navigation contract. Slots include
header/source copy, role and name fields, process labels, responsibilities,
criteria, biographies and metrics. Array cardinality is fixed; callers supply
stable keys for every row, nested bullet list, org-chart branch and legend.
Geometry, diagram edges, layout, table column keys/styles, photos' positions and
surfaces remain fixed structure. Table-group `from`/`to` column indices are
structural and excluded from v2 slots; timeline date/domain values retain their
content semantics. Photo slots select a registered asset key.

Navigation variants additionally require 2–6 caller `{key,label}` items and an
active key naming one item. Source specimens use the source navigation; changed
specimens use four tabs with changed labels and family-specific active keys.
Split architecture, readiness, bio and roster variants use the implemented v2
frame contracts. Readiness retains its source three-line title capability.
Org charts and process maps preserve their authored relationship structure;
changed names/copy do not invent relationships or imply count-driven adaptation.

## Changed-content specimens

All examples are explicitly synthetic. Changed Team specimens replace fictional
names, keep initials consistent, vary experience points, capacity or operative
responsibility copy, and bind every source slot/key explicitly. Biography portraits
remain registered example assets. Solution specimens vary architecture services,
shared controls, process scope, cycle times/rework, readiness criteria and platform
workstream risk copy. They exercise at least two changed body slots per template,
in addition to any header/nav changes. Copy is kept within each source capacity.

Invoice baseline alternates total 12 days: 1 + 2.5 + 6.5 + 1 + 1. Approval is
54% after rounding; handoff counts remain four of nine. These are specimen values,
not client findings. The source specimens remain unchanged, including their source
synthetic claims and qualifications.

## Named amendment

`wmds.org-roles-nav-footer-capacity.v2` changes the responsibilities table's
`node25.rowH` from 48 to 46 pt. Its source position is y144, with a 28 pt dense
heading and six rows: the source ends at 460 pt, beyond the 450 pt reserved footer
boundary. With 46 pt rows it ends at 448 pt. Measured two-bullet cell allocations
fit at the same font sizes, without truncation. The table and its heading retain
their authored position; source JSON and v1 are unchanged. This amendment is
applied through `applyTeamSolutionLibraryRefinements` in the shared dispatcher.

The first changed roles/by-phase specimen exceeded its 40 pt row allocation.
Its synthetic sentence was shortened before the accepted generation. No source
geometry or font size was changed to accommodate that specimen.

## Reproduce and artifacts

```sh
python3 scripts/wmds-refresh-team-solution.py \
  --cli /tmp/wmds-team-solution-pptxdesign \
  --out /tmp/wmds-team-solution-new
```

The CLI must be freshly compiled with the family refinement dispatcher. The output
directory must be new. The script selects these exact 19 `added` inventory entries,
generates source references, creates exact closed slots/keys/navigation, binds
paired slides and emits per-template receipts. It invokes no native automation.

Family artifacts are under
`samples/wmds-refresh-slice3-20261002/work/team-solution/`:

- `combined.foundation.json`: all 38 compiled source/changed pairs for integration.
- `bound-content.json`: all exact caller content, including navigation and keys.
- `generation-receipts.json`: per-template source/paired generation results.
- Each template directory: paired content/foundation, bindings, layout reports and
  an editable paired deck with a generation receipt.

The candidate Go typography engine is unchanged. Generation receipts retain `generated_native_review_pending`; the separate
family native review receipt records accepted visual review of these 19 pairs.
The root review owns PowerPoint opening, local-printing PDF export, artifact
hashes and integrated final receipts.

## Integrated native review

All 38 integrated pages 79–116 were visually inspected from the PowerPoint local
printing PDF. Full-size checks covered biography badges, table rows and owner
tags, optional roles, dense readiness copy and five-workstream risks.

Two corrections were applied and natively rechecked: the alternate
process/hierarchy systems field now uses single-line specimen copy
(`ERP receivables`), and shared v2 outline-surface blocks now paint the source
1 pt inset border on optional org-chart roles and the architecture source-system
layer. All nine changed pages (89, 90, 93, 94, 99, 100, 101, 102 and 106) were
reviewed again at full size. The remaining family slide XML is unchanged from
the first accepted visual inspection. The family receipt
`work/team-solution/native-review.json` records both image sets, resolved findings
and final **accepted** status. No arbitrary-content
qualification is inferred.
