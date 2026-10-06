# Renderer density checkpoint — 2026-10-06

Candidate source: `4fce3cd7ecdfaf8daadfeb14bf6e4b596d731c9a`. Default/production remains V10. No release or production pin change is authorized by this checkpoint.

## Completed runtime

- Whole-slide Comfortable/Compact/Dense body roles, independent explicitly chosen header role tier (default Comfortable), per-slide automatic body fit retries, requested/resolved/step reasons in reports. Preferred input and cached sources remain immutable; body-only layout preflight serializes the final deck once. Old bundles retain existing behavior.
- Latest designer role corrections, number-long, exact floor exceptions, preserved strongnum Mono family with body/small metrics, and fixed diagram/quote exceptions.
- Rich/plain/table measurement uses observed pitch only for supplemental calibrated pairs; base-anchor historical behavior preserved.
- Latest draft-review Due/Updated LABEL and Owner SMALL; fixed status8/10 and chip6.5/9; legacy source unchanged.
- Final `numTile:"callout"`: fixed27×27 Magenta, fixed Mono60014/14 Grounded centered, titleX+36, centered on first title leading, no row-flow expansion. Generic strict source field and native XML coverage added. Temporary700 policy fully removed.

## Bounded tests completed

Focused runtime suite passed1.120s:

```sh
go test ./internal/wmdesign -run '^(TestCardNumberTile|TestDraftReviewDensity|TestAutoDensity|TestOldBundleDensity|TestTypographyDensity|TestDensityNative|TestDensitySplit|TestDensityReviewed|TestContrastProbe)' -count=1
```

Audit-agent final649 source/648 bound/full218 source and bound build sweeps passed. Final431 retained-source collector found334 byte-equivalent and97 changed renders after native calibration; these97 require native qualification, not inheritance. Audit agent owns final partition/evidence files.

## Complete contrast audit (open design gate)

Historical `contrast-audit-4fce.json` now retains the original full-report SHA-256, counts and every failing check; passing object records were retired after final publication. Current qualification is `contrast-audit-3c56-final.json` (zero failures). Summary: **1,947 specimens,62,850 checked text/font/color objects,286 low-contrast records across46 templates, zero coverage/traversal gaps**. Comfortable66 failures, Compact110, Dense110. These are color/actual-size/actual-weight checks independent of wrapping/capacity and do not qualify layout fit. All compiled scene nodes, emitted manual diagram text, rich runs, table text and native chart axis/data font options are traversed. Textless inline artwork skips rasterization only inside this private audit; normal source image behavior is unchanged.

The all-tier contrast regression intentionally fails while these findings remain unresolved:

```sh
WMDS_CONTRAST_AUDIT_OUT=/private/tmp/contrast-resume go test ./internal/wmdesign -run '^TestLibraryDensityAllTierContrastAudit$' -count=1 -v
```

Do not call Comfortable-profile failures proven historical V10 violations without matching V10 evidence. Examples: small White-on-Magenta; Magenta12/11pt agenda keys; Grounded/white/subtle contrasts at smaller role tiers. The final6 added checks are fixed de-emphasized road-fork/decision-chosen pin numerals (2pins×3tiers),97A4BAonWhite10/600; those are existing fixed diagram semantics, not density scaling. User/design approval is pending for broader color repairs. No broad recoloring or weight override has been applied.

## Native controls

- Fixed-tile3tiers: `/private/tmp/pptxgengo-numtile-controls/numtile.pptx`, compiled and report sidecars. Reproducible via `WMDS_NUMTILE_CONTROL_OUT=<path> go test ./internal/wmdesign -run '^TestWriteCardNumberTileNativeFixture$' -count=1`.
- Original250 bare font controls and final56 extra pair controls remain separately frozen under `/private/tmp/pptxgengo-density-controls` and `/private/tmp/pptxgengo-density-extra-controls-r2`. Supplement306 anchors owned by root/Luna. Do not regenerate original controls against a newer source.
- Root owns final density-demo/tile/native stock review. Native qualification/publication/default switch still pending.

## Owned implementation files

Primary: `density.go`, `density_test.go`, `density_controls_test.go`, `contrast_audit.go`, `contrast_audit_test.go`, `scene_card_numtile_test.go`. Shared owned hunks: `source.go` token density parsing (audit owns revision/pins), `geometry.go` header field, `library_templates.go` raw density/header parsing and fixed numTile metadata (audit owns cardinality/gates), `library_refinements.go` V11 inheritance, `render.go` density/report/preflight fields (root owns constructor/nav fix), `rich_text.go`, `scene.go`, `scene_cards.go`, `draft_review.go`, `components.go`, `data_metrics.go`, `metrics.go`, and body/cell/list role/measurement call sites in scene primitives, sequences, tables/XML/references/rowgroups, diagrams, media, charts, roadfork, and intake scenes. Capacity/project/YAML/BoundSlide/CLI fields belong to Luna. Calibration loader/plain candidate and publication compatibility belong to root. Frozen source/tests/partition belong to audit agent.

## Resume order

1. Read root durable RESUME and current final pin/partition; preserve other dirty work.
2. Resolve operator/design preference for broad contrast findings; any source changes require a new audited candidate pin and rerun systemic contrast tests.
3. Finish root native qualification for changed render set (218 source changes plus97 calibration/role drifts); retain334 exact inheritance evidence.
4. Check final project/CLI/capacity warnings/tests, full suite (contrast gate remains explicit), then publication/release only after native/design gates.
