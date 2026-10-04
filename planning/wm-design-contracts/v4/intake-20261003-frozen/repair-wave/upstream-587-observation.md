# Upstream 587-template observation

Observed at `2026-10-03T21:18:17.047277+00:00`. This is a read-only inventory observation; the frozen v4 source, candidate bundle, installed release, and repair scope remain unchanged.

## Verified provenance

- Local upstream: `/Users/rscott/Projects/wm-design-system`.
- Current upstream commit: `d83bd58a9f9de68ebd8d6b3c9b0272c16ed516cf`; clean working tree (`git status --porcelain` empty).
- Frozen v4 intake commit: `8f9f16ab8a7e2a6fff45a97627308668f4a12753`.
- Independently counted **587 unique template identities** in both HEAD and working files, confirming the user's reported total. No duplicate identities; committed and working definitions match.
- Current upstream has 27 template family files, compared with 24 in the frozen intake (the `_gaps.json` inventory file contains no templates).
- Compared canonical definitions using `scripts/snapshot-wmds-intake.py`'s `entries`, `delta`, and `capabilities` functions. Working source bytes, commit, and status were checked again before saving this receipt.
- Exact changed identities, observed source file hashes, and commit provenance: [upstream-587-observation.json](upstream-587-observation.json).

## Delta from the frozen 522 templates

**65 added; 0 removed; 3 revised; 0 moved.** The existing 519 other definitions are unchanged.

| New family | Added templates |
| --- | ---: |
| interviews | 29 |
| offers | 17 |
| understanding | 19 |

The new templates cover stakeholder interview planning, access, coverage, profiles and readouts; packaged offers and one-page proposals; and understanding/context, objectives, scope, statements and traceability.

Existing revisions:

- `case-study/exhibit-split`: source disclaimer now asks for current firm data; word budget 70 to 74; revision 1 to 2, revised date 2026-10-03.
- `value-bridge/levers-split`: source disclaimer now asks for current firm data; word budget 86 to 90; body nodes 24–37 moved up 18 points.
- `value-curve/break-even-split`: source disclaimer now asks for current firm data; word budget 82 to 86.

There are **no new scene node types or scene fields** relative to frozen v4. This inventory comparison does not establish that the 65 additions render correctly or fit after binding.

Upstream commits since freeze:

- `080c0b4`: Round 14 — our understanding, stakeholder interviews, one-page proposals; status labels; wrapping source lines.
- `d83bd58`: Round 15 — denser one-page proposals (what/how/cost), more understanding variants, interview detail with headshots.

## Changed observed source files

- `docs/authoring-reference.md` (revised).
- `explorations/components.src.html` (revised).
- `templates/catalog.json` (revised).
- `templates/change-notes.json` (revised).
- `templates/changelog.md` (revised).
- `templates/changes.json` (revised).
- `templates/library/interviews.json` (added).
- `templates/library/offers.json` (added).
- `templates/library/proof.json` (revised).
- `templates/library/understanding.json` (added).
- `templates/library/value.json` (revised).
- `tools/build_explorations.py` (revised).

## Next intake actions

1. Keep the current 522-template repair and qualification evidence pinned to the frozen commit.
2. Create a separate immutable intake snapshot at `d83bd58a9f9de68ebd8d6b3c9b0272c16ed516cf` using a new destination and the frozen v4 source as the previous observation. This note does not copy or adopt upstream files.
3. Qualify the 65 additions and three revised definitions, plus any rendering effects from changed source support files, with fresh source and bound decks, node checks, and native review.
4. Select the next bundle revision only after those checks; regenerate its catalog, previews, database, and distribution counts together. The verified 587 total is an upstream inventory count, not the current installed library total.
