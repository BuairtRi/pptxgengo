# Continuing intake: capability implementation and evidence

> Current target: [frozen 522-template intake and updated capability evidence](intake-capabilities-20261003-frozen.md). This document preserves the earlier observation.

## Scope and source

This slice observes the actively edited West Monroe design system and implements
ready renderer contracts. It does not publish the working drafts or claim that
all incoming templates fit. The installed local.7 library stays on its reviewed
248-definition v3 snapshot.

The immutable observation at **2026-10-03 17:50:15 UTC** has 301 committed and
369 working definitions. Against v3: 121 additions, 13 revisions, no removals.
There are 68 additions beyond current HEAD. Counts are observations, not a final
catalog total. The source manifest is
`planning/wm-design-contracts/v4/intake-20261003/observation.json`.

New scene types are identified only in `template.slide.body`. A table column
whose semantic type is `maturity` is not a diagram scene. An earlier inventory
looked at the wrong body location; its empty new-type result is superseded.

The frozen renderer is `a7f1f046d38723cb86d6e0dbacd97886c183784df1e4949192322e74d6a33cf6`.
Venn family: `a0c26d8d4a804b6d8016e8e1b9adf1fce6cedb59d28f18aeb5922ebddcd809fa`.
Maturity family: `63eca0027e5797a19be2b1236986772960889f2bb4ad1bb18a93a08416145c5b`.
Fixtures and review target those bytes. Upstream authoring has continued since.

## Implemented contracts

| Capability | Implemented behavior | Boundary |
| --- | --- | --- |
| Venn | 2–4 editable circles; set/intersection copy and bullets; width overrides, lens anchors, mono caps, point numbering and left/right badges; opacity | Finite required allocation; 64 points, 11 distinct regions, 128 bullets and 65,536 copy runes per node. Shared palette policy; arbitrary CSS colors unsupported. |
| Maturity | Exponential curve with 61 samples; stage markers/ticks; active marker, axis, inflection, branch, headroom and “you are here” | Eight frozen scene examples; finite required geometry, valid indices/positions, bounded copy. Source stage positions and branch anchor remain fixed in content bindings. |
| Team curves | Monotone Fritsch–Carlson paths, reversible band edges, area/line series, dashed lines, actual cubic bounds | Explicit `curve: monotone` works now. Frozen v1/v2/v3 keep Catmull defaults. Only the future v4 revision default is changed; accepting/registering that revision remains a separate task. Nonfinite generated controls fail cleanly. |
| Plain table subtitle | `{text, sub}` as two editable native paragraphs, secondary subtitle token and 2pt gap | Per-paragraph leading is serialized from its effective token; combined height must fit the declared row. |
| Score cells | Dots with optional string/bullets; object rating/Harvey forms; cell, row, column ink precedence | Integral scores and maximum 1–12; marker footprints must fit the cell. Scalar geometry remains unchanged for valid existing specimens. |
| Heat | Table heat columns, cell text/value/scale, row/column scale, heat maximum, block heat and legend swatches | Five-step sequential/risk ramps. Finite out-of-scale values saturate color while the original number remains visible. Invalid maximum/scale/nonfinite data fail. |
| Straight arrow | Existing pinned hand-drawn asset; horizontal flip and rotation | Actual intrinsic height participates in bounds. Complete turns are normalized before native integer angle conversion; excessive/nonfinite input fails. |

Table heat colors use editable underlay shapes and transparent native table
cells to preserve the authored inset gutters and native text. Underlays are
collected and assembled once. No quadratic growing-tail copies remain.
Ordinary cell background fields are not invented: the observed contract defines
heat shading, not an arbitrary per-cell CSS background API.

Local composition recognizes Venn and maturity. Binding projection exposes
visible copy/data and fixed-count identities, including bare numbered points;
style/geometry/scale/max remain template-owned. Structural discovery includes
Venn, maturity, and heat-map affordances. An incoming template enters the actual
SQLite catalog only after its source revision is imported and accepted.

## Independent review and fixes

Three capability agents implemented separate files; reviews crossed ownership
boundaries. The review receipt is
`planning/wm-design-contracts/v4/intake-20261003/independent-review.json`.

Resolved findings include:

- Missing Venn identities when a point has implicit numbering but no label.
- Rotated marks reporting a zero/inaccurate intrinsic height; multi-turn angles
  overflowing native integer conversion.
- Maturity `at` values being accidentally exposed as content rather than geometry.
- Missing/null origin coordinates being interpreted as zero.
- Missing maturity headroom and bounded auxiliary copy.
- Node-wide Venn bullet/copy limits being applied separately to each region.
- Tiny curve sample intervals producing nonfinite cubic controls.
- Dense heat tables repeatedly copying a growing item array.
- Subtitle paragraphs inheriting main-copy line spacing in native XML.
- Legacy scalar score paths bypassing footprint checks.
- Heat colors rejecting intentional overcapacity data and white gutters on inverse slides.

## Checks and prototype

Normal checks passed for `internal/wmdesign`, `internal/deckproject`,
`cmd/pptxdesign`, and `cmd/pptxgengo`. Full affected race checks passed:
62.648s, 61.650s, and 6.400s for the three tested packages. Earlier broad `pptx`
baseline failures remain separately documented; this is not a claim that the
entire repository test suite is green.

All twelve frozen Venn nodes and eight maturity scene nodes pass planning and
native editable XML checks. Curve checks cover ten frozen v3 specimens plus
monotone/reverse/dashed/bounded-output cases. The existing 248-slide v3 source
reference regenerated after the fixes.

The full-slide intake diagnostic attempts 63 definitions from Venn, maturity,
heat maps, vendor comparisons, score roles, and team curves. **43 compiled; 20
were rejected.** A completed diagnostic is not an assertion that all sources fit.
Results are recorded per key in the frozen snapshot's `smoke-results.json`.

The unreleased prototype is
`/tmp/wmds-intake-capabilities-20261003-final/capability-reference.pptx`:
43 slides, 395,211 bytes, SHA256
`8761055b62b9e6507a56891e16670cdb1bd916dc40f11050e9271494a0e343b7`.
Its 300 internal relationships resolve; XML contains 28 native tables and 95
custom geometry parts. Incoming team curves are explicitly lowered to monotone
in this prototype; the installed v3 source defaults are not migrated.

Reproduce the diagnostic with a **new output directory**:

```sh
WMDS_INTAKE_SMOKE_OUT=/tmp/wmds-intake-new-output \
  go test ./internal/wmdesign -run '^TestIntakeManualSourceSmoke$' -v -count=1
```

Native visual review is pending. PowerPoint remains on the operator's reference
review; no review session was interrupted. No incoming release was installed,
and no commits/push were made.

## Remaining template and frame work

The 20 rejections are a concrete intake repair list:

- **12 diagram split cases:** eleven Venn and one maturity specimen have actual
  ink bounds beyond split reservations. Distinguish intentional stroke/circle
  bleed from content overflow, then adopt an explicit bounded frame rule or a
  reviewed composition amendment. Do not hide actual bounds or relax every scene.
- **One heat split case:** left bullets begin above the frame's short-body start.
- **Two table fit cases:** multiline cell/subtitle copy exceeds a 36pt row.
- **Three card fit cases:** measured copy exceeds the authored card height by 3–6pt.
- **One table width case:** columns do not sum to the declared table width.
- **One glyph case:** vendor legend uses U+25CF, unavailable in the pinned Sans
  face; adopt a native marker/legend treatment or an explicit qualified fallback.

The frame snapshot adds a **slim footer**: rule495, row504–516, body bottom486,
source bottom489. Tokens are unchanged. The next frame slice must review source
reservation, logo/legal/page placement, project frame references, and catalog
metadata before accepting it. Existing frames stay pinned.

Next, import ready committed template families, resolve their fit/frame issues,
natively review source specimens, generate previews and catalog evidence, and
publish a coherent runtime/library/skill release. Reinventory at every boundary
as upstream agents finish more drafts. Accepting a v4 bundle requires new source
hashes/counts and modern revision semantics; this observation directory is not a
registered bundle.
