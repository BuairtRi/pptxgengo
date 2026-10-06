# V11 completed release checkpoint — 2026-10-06

## Current production

Released and globally installed: **0.1.0-local.20 / V11**.
Canonical upstream commit: `3c56d842ba3abb5f24eb33cf0082be7a7a67f116`.
Bundle SHA: `eb7dbb02d78b0a32b8bba552ba54bfd96f60e462985d7829f6fe91b4d00ff633`.
Inventory SHA: `5fd96a039055e8d281c95fbb55fc472ca816308dfde34c62a0b869f223cacba1`.

649 templates: 648 active, one deprecated, 29 workshops.
`library/wm-design-system/v11` is the sole production bundle. Historical intakes
remain immutable. Existing deck projects retain their explicit source locks.
Global command and presentation-skill links select installed local.20.

## Core CLI completed

- Whole-slide body Comfortable / Compact / Dense; independent header density
  defaults to Comfortable. Roles follow the latest designer tokens.
- Automatic body fitting is on by default for V11, reports every adjustment,
  and never rewrites the requested YAML tier. `auto_density: false` opts out.
- Ten source-owned Comfortable-only limits stop fitting and reject explicitly
  prohibited tiers. Compiled metadata cannot relax stock limits. See
  `density-limited-template-keys.json` and the installed CLI probe under
  `release-qualification/`.
- Project/YAML/schema support, refreshed capacity/discovery, `measure-style`,
  and `--template-keys-file` are implemented and tested.
- Modern typography uses 312 exact native controls; older density sources keep
  their separate 306-control behavior. Fixed diagram numeral exceptions remain.
- Native publication compares embedded workbooks structurally, ignoring only
  ZIP packaging and valid created/modified timestamps. Data, styles, relationships,
  and other metadata remain strict. Negative tests pass.

## Template inventory and native acceptance

215 density migrations and later pillar/color changes produce 231 raw changed
compositions. Of 418 unchanged source compositions, paired rendering finds
121 additional visual changes and 297 identical eligible previews.

**352 new full-size native reviews + 297 verified inherited previews = 649.**
All new specimens are accepted. Retained packets: `native-review-set-a/` (230),
`native-review-set-b/` (119), and `native-review-set-c/` (3), with sibling PPTX,
native PDFs, full-size PNGs, export receipts and reviewer ledgers. Publication
rebuilt against final source and verified exact visible dependencies.
The final three approved repairs are Blue maturity inflection labels and
Grounded inactive road-fork numerals, retaining gray outlines.

The native contrast audit inventories all 1,947 template/tier combinations:
1,927 supported, 20 source-prohibited. It checks 62,578 text objects with zero
failures and zero coverage gaps. Stock acceptance does not qualify arbitrary
replacement content; edited decks require fit and native review.

## Tests, publication and installation

- All 649 source and 648 bound builds; 231 changed round trips and individual
  builds: PASS (exhaustive sweep 170.493 s).
- 418 paired rendering: PASS 83.72 s; 297 inheritance closure: PASS 59.25 s.
- `make test`: PASS, slowest package 122.337 s. After production cleanup:
  PASS, slowest package 151.928 s while checks shared machine resources.
- Maintained `make test-race`: PASS, largest package 104.581 s. Additional
  focused density race checks pass. No 30-minute broad race run is required.
- Gallery/SQLite published: 649 templates, 1,204 asset variants, all 521 original
  photography records. Matched documentation frozen against exact 3c56 source:
  649 templates, 29 families, 23 hashed assets.
- Installer PASS; installed manifest audits **5,151 files with zero drift**.
  Version, defaults, discovery, measurement, immutable limit probe, documentation
  source and linked skill were checked outside the repository.

Canonical record: [local.20 qualification](../../../../release/qualification-local20.json).
Logs are retained under `release-qualification/`.

## Cleanup and ownership

About 199.2 MiB of superseded comparison material was retired/compacted; historical
hashes and failure findings remain in `historical-audits/artifact-simplification.json`.
V10 production was deleted after V11 installation/integrity verification.

Preserve unrelated staged voice-test deletions, DentalXChange slide 041 and the
other agent's presentation-skill changes. Owned engineering skill additions are
`references/typography-density.md` and the documentation reference update.
No native export, long test or documentation server remains running.

For future source changes, freeze a new pin and repeat affected build, contrast,
render inheritance and native acceptance checks before publication.
