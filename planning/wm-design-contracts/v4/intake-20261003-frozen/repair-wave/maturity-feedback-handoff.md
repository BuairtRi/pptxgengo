# Maturity visual feedback handoff

Checkpoint: 2026-10-03. Agent: authoring_contract_research. Work stopped at the user's request before a fresh session.

## Requested correction

The user observed maturity-circle numbers sitting toward the top and some stage labels sitting too close to, or outside, the upper rule. The latest reference deck is the frozen 522-template build, approximately 33.7 MB. That build predates these visual-feedback corrections.

## Work performed this turn

Read the current maturity adapter, its tests, the final frozen renderer, typography candidate measurement/calibration, scene text emission, and frame-zone selection. **No maturity implementation or test files were modified. No native operations or new tests were run during this checkpoint turn.** This document is the only new file from the maturity-feedback investigation.

The prior card repair work remains in the shared uncommitted checkout:

- `internal/wmdesign/intake_repairs_cardfit.go`
- `internal/wmdesign/intake_repairs_cardfit_test.go`

Those repairs passed all 43 recorded rejected card/cardrow nodes and 14 complete native builds. Copy, styles, assets, revision gating, idempotence and atomic rejection were checked. Independent review subsequently found the Seven-R width guard allowed126pt at an originally102pt position; that is fixed and covered by an atomic rejection regression. The final focused card/scene-missing-chart/PPTX-missing-chart run passed (`internal/wmdesign`:1.294s; `pptx`:0.467s). Independent text/table/card/compact checks passed1.655s. These are Go/XML checks, not PowerPoint visual qualification.

## Concrete centering mechanism to investigate next

`internal/wmdesign/scene_intake_maturity.go`, function `maturityMarker`:

1. The circle's visible envelope is24pt centered at `(q.x,q.y)`; the native ellipse is22pt with a centered2pt stroke, inset1pt.
2. Numbers use IBM Plex Mono600,11pt for up to two characters and9pt for longer strings, exact leading equal to font size, no case transformation/tracking. These agree with the final frozen renderer's24pt flex-centered marker.
3. The current native text goes through `diagramStyledText(... Rect{q.x-10,q.y-12,20,24}, ... middle=true)`.
4. `diagramStyledText` centers the **conservative allocation height**, replaces the text rectangle with that height, then sends a top-aligned text record to `scene_render.go`. It does not center actual numeral ink.
5. The candidate calibration contains no exact IBM Plex Mono SemiBold11/11 or9/9 anchor. The face hash is `b1a936f96266fbe036b112f659155f7c32222af44a9a20f3bb011f51050669d5`. Existing anchors for this face are48/48,36/36,18/24 and10/12.
6. Consequently the fallback uses first baseline `.75*leading` and terminal allocation `max(leading,1.5*size)`. For11/11 this is baseline8.25pt and height16.5pt. Centering16.5pt puts the estimated first baseline exactly at the circle center; numeral ink lies mainly above its baseline. The9/9 fallback has the same issue (baseline6.75pt,height13.5pt).

This is a concrete mismatch between centering a padded allocation and centering numeral ink. It is consistent with the user's observation. It is **not** a measured PowerPoint baseline result for these exact sizes.

Next: choose and test a narrowly scoped marker-centering implementation. Options include a full24pt native text box with explicit vertical centering (requires a scene-text alignment field/emission hook), or a Go placement based on numeral glyph ink extents and an appropriate measured/native baseline anchor. Preserve the existing font sizes/copy and existing revision behavior; do not solve this through arbitrary font shrinking. Add regressions for single-digit, two-digit and fractional branch numbers, active/inactive circles, and native XML text alignment. Existing tests assert text-box centers, which is insufficient to detect the reported ink-centering issue.

## Label geometry and top rule

The frozen authority is:

- `source/explorations/components.src.html`, `case "maturity"` in the final frozen522 snapshot.
- `source/templates/library/maturity.json`, plus maturity nodes in other family files.

The adapter places labels above ticks:

- Base `tipY = q.y -30`.
- Final steep stages use an additional40pt lift (last two, or last three for at least six stages in the incoming revision).
- `maturityStack(... y=tipY-4, above=true)` measures the label+detail stack and subtracts its total height. The label token and small detail token are retained; the inter-item gap is4pt.
- `headroom`, `at`, `shape`, `labelW`, diagram position/height and source text wrapping all affect whether the resulting stack approaches the title rule.

There is no stage-label-specific upper clearance guard inside the maturity planner. For ordinary nonsplit scenes, `render.go` currently constructs the scene zone with `Y=0` and bottom at the body reservation. This permits some scene text above the normal body top, so a successful whole-slide Go build alone does not establish clearance from the title rule. Split tall scenes use a different zone; do not apply the ordinary title rule to their header-free tall column indiscriminately.

No offending page/template was conclusively identified in this checkpoint. Do not claim a label correction is complete. The next session should correlate the user's observed pages with the current deck/template mapping, inspect each resulting label rectangle against its actual header/title-rule reservation, then adjust specific layout/headroom/label positions while preserving copy, type sizes, curve/stage correspondence and footer clearance. Record any source amendment explicitly. A frame-aware label clearance contract is preferable to silently moving all labels or allowing overflow.

## Verification and native limitation

PowerPoint native access is blocked by the current automation bootstrap. The capture packet/export hung at approximately6:37; the user was advised to interrupt the shell process. No app termination, registry mutation, or font-environment changes were performed here.

No changes for this visual feedback have been compiled into the latest522 decks. After a centering/label fix: run focused normal/race tests and the full frozen522 build, regenerate the reference decks, then inspect native PowerPoint/PDF output in the fresh session. Preserve the source freeze and all current uncommitted work.
