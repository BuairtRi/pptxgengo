# Slice 3 — added template families

Implemented all 70 added templates from frozen source commit
`7bdcaee5030a12275a1f881a8542f4d302d207df`, using four concurrent work groups.
Each has a source specimen and a meaningful changed-content specimen. The combined
reference has 140 slides; alternate examples change body content, not just titles.

| Work group | Templates | Integrated pages |
|---|---:|---:|
| Approach + Commercials | 19 | 1–38 |
| Proof + Evidence | 20 | 39–78 |
| Team + Solution | 19 | 79–116 |
| Argument + Openers | 12 | 117–140 |

## Contracts and amendments

Family notes describe exact content fields and bounded capacities:
[Approach/Commercials](added-approach-commercials.md),
[Proof/Evidence](added-proof-evidence.md),
[Team/Solution](added-team-solution.md),
[Argument/Openers](added-argument-openers.md).

Shared corrections preserve source semantics:

- Navigation uses the source's dedicated Plex Mono 8pt/600 style and 0.8pt tracking,
  with two-point native end guards. Generic label typography is not substituted.
- Table group row boundaries are structural; visible group labels remain content.
- Compound RACI badges allocate their measured text height and center it in the
  original box; no font shrinking is introduced.
- Scene text permits fixed source weights 400/500/600/700, preserving semibold
  runbook roles when title/time labels are separated.
- Fixed case-study text blocks reserve 57pt/75pt row heights and 9pt separator
  clearance; excess text returns `scene.textblock_overflow`.
- V2 outline surfaces have the source's 1pt inset `#CED7E6` border; outer box bounds
  and text placement remain fixed.

Named family amendments provide taller overview chevrons with stacked time/role
labels, two-line step titles above owner/window labels, a quote card with enough
vertical capacity, chart/source clearance, and a shorter six-row org-role table.
Source JSON and v1 are unchanged. Explicit editable thumbnail illustrations are
composed specimen content; placeholders remain placeholders in the source contract.

## Review and scope

Microsoft PowerPoint for macOS opened the combined reference without repair and
exported a local “Best for printing” PDF. All 140 pages were visually inspected
across the four work groups. Corrections address an alternate runbook title/owner
collision, missing outline borders, a long synthetic systems label, and inconsistent
synthetic metric/day/chart copy. The first 19 changed pages were re-exported for
focused review; the last case-study capacity amendment was re-exported and
reviewed on pages 51–54. The other 136 slide XML files were unchanged in that last
round. In total, 119 slides were unchanged from the initial native review.

[Review receipt](reference-review/slice3.json) pins the final artifacts and scope.
This evidence covers these source/alternate pairs. General caller text still uses
measured fit and fails when capacity is exceeded. Full 167-design integrated review,
unchanged-design regression review and release packaging remain slice 4. The
installed release and default v1 selection have not changed. No tests were added
or run; no commits or push were performed in this slice.

## Reproduce

Compile the repository `cmd/pptxdesign` binary, then run each family script with
`--cli PATH --out NEW-DIR`:

- `scripts/wmds-refresh-approach-commercials.py`
- `scripts/wmds-refresh-proof-evidence.py`
- `scripts/wmds-refresh-team-solution.py`
- `scripts/wmds-refresh-argument-openers.py`

Canonical group documents live under
`samples/wmds-refresh-slice3-20261002/work/`. Integrate with:

```sh
python3 scripts/integrate-wmds-slice3.py --cli ./pptxdesign --out /tmp/wmds-added-reference
```

The integrator requires exactly two specimens for every added inventory key and
writes the 140-slide foundation document plus page manifest. Native PDF export is
a development review step, not part of normal Go deck generation. PNGs and contact
sheets can be reproduced using `scripts/render-pdf.swift` and
`scripts/render-contact-sheet.swift` after native export.
