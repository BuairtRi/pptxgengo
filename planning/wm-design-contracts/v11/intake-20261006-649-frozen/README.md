# WMDS v11 committed intake

Frozen candidate upstream commit: `3c56d842ba3abb5f24eb33cf0082be7a7a67f116`.
This candidate has completed the final source/runtime audit; publication still requires the exact native packets and inheritance receipts.
Baseline: qualified v10 source `c14fb286fb38e15800a6fd476a1ed67956f1165f`.
The catalog still contains 649 templates: 648 active, one deprecated and 29
workshop templates. No template keys are added or removed.

Every source file was acquired with `git show` at the exact committed source.
The snapshot has 40 source paths: 38 changed physical files and two verified
links into immutable v10. Font, license, asset and typography evidence files
retain v10 identities; the 15 manifest entries are unchanged and verified.

- Bundle SHA-256: `eb7dbb02d78b0a32b8bba552ba54bfd96f60e462985d7829f6fe91b4d00ff633`
- Inventory SHA-256: `5fd96a039055e8d281c95fbb55fc472ca816308dfde34c62a0b869f223cacba1`

[Intake audit](intake-audit.json) records committed Git blob identities, byte
counts, hashes, storage relationships and every source-slide change.

The earlier unqualified candidate at `d7027329522724b7c7225b0ca8309385cb7d0d42`
was replaced with the explicitly approved committed browser density fixes and
exception documentation. It was never qualified or published. The
[superseded candidate audit](superseded-candidate-audit.json) preserves its
source identities, original density policy and the exact six-file refresh
delta. Its executable pin was removed. The role refresh changed no template compositions. The subsequent final color/tile
refresh adds only two Grounded title ink fields and twelve tile declarations, all
already within the 218 changed-source set. Qualified v10 and prior pins remain intact.

## Delta and ordered keys

There are 231 changed compositions and 418 unchanged compositions versus v10.
The original 218 changes comprise 215 density migrations and seven pillar
revisions (four overlap). The later contrast policy adds thirteen previously
unchanged compositions: ten comfortable-only density limits and seven Blue ink
fields across three templates. No copy or geometry changes occur in those thirteen.

| Set | Count | Ordered JSON | Comma-separated CLI keys |
| --- | ---: | --- | --- |
| Changed source | 231 | [JSON](changed-template-keys.json) | [Text](changed-template-keys.txt) |
| Density | 215 | [JSON](density-template-keys.json) | [Text](density-template-keys.txt) |
| Pillars | 7 | [JSON](pillar-template-keys.json) | [Text](pillar-template-keys.txt) |
| Unchanged source | 418 | [JSON](unchanged-template-keys.json) | [Text](unchanged-template-keys.txt) |
| Additional rendered differences | 121 | [JSON](additional-render-review-template-keys.json) | [Text](additional-render-review-template-keys.txt) |
| New native review | 352 | [JSON](native-review-template-keys.json) | [Text](native-review-template-keys.txt) |
| Render inheritance | 297 | [JSON](inherit-eligible-template-keys.json) | [Text](inherit-eligible-template-keys.txt) |

All lists follow the current upstream catalog order, which is unchanged from
v10. The changed set contains active templates only. The unchanged set includes
the deprecated `from-to/rows` source specimen.

### Density migration

The density-only edit adds `density: "compact"` to 215 slides and removes 659
component overrides: 316 `bodySize: "small"` and 343 `size: "small"`. Every other
source-slide field remains unchanged outside the separately identified seven pillar
revisions, four exact color/tile amendments and thirteen final contrast-policy
compositions: copy, values, geometry, IDs, arrays and topology are preserved.

Current source body policies: 215 compact, 429 omitted/default and five legacy
appendix. All 649 source slides omit `headerDensity`, whose source default is
comfortable. Typography adds comfortable/compact/dense token tables, while
every prior token field remains unchanged. Frame changes rename the standard
policy to comfortable and document compact/dense; frame geometry is unchanged.

The final source fixes the density roles for Gantt lane titles, numbered item
labels, numeric display tables, Venn point tags and table reference badges.
Strong numbered-list numerals follow the list's body/small role. Long card
metric badges use `number-long`: 13/12/11 pt with 17/16/14 pt leading.
Fixed pin/dot/decision glyphs retain their sizes. Quote marks follow the later source-owned density formulas with a 40 pt floor.
Two documented exceptions stay below the 8 pt floor: Gantt period sub-labels
at 7.5/9 pt mono and review-note status chips at 6.5/9 pt mono caps.

The final committed source includes Grounded titles on the Magenta callout cards in
`capability-heat/annotated` and `risk-heat/annotated`, and `numTile: "callout"` on
all twelve cards in `pillars/two-categories-six-magenta` and
`pillars/two-categories-six-stacked-magenta`. Copy and card geometry are unchanged.
The fixed 27 × 27 pt tile uses Grounded IBM Plex Mono Semibold at 14/14 pt at
every density. Its centre follows the current title's first-line leading; its
left is the original numeral's left and the title begins 36 pt later. The tile
does not enlarge the title row or move following body text. `numTile` ignores
`numInk`. The previous uniform weight-700 proposal is superseded and must not
be published.

[Color and tile refresh evidence](color-tile-source-refresh-audit.json) preserves
the unqualified 8c and 440 candidates and their hashes; their executable pins
were removed. The final source implements browser rendering and contrast checks
at actual rendered point size/weight. Its committed bounded regression is
`tools/check_density_contrast.js`: four repaired templates at three densities,
plus a 600-weight NUMBER check that passes 3:1 at 18 pt and needs 4.5:1 at 16/14 pt.
The twelve browser specimens passed. Exact issued native specimens qualify visual results separately.

In the original density migration, two hundred one templates bump their source
revision metadata and fourteen migrated slides retain the same revision value; source and render hashes determine
qualification rather than the revision number alone.

### Pillar revisions

The seven keys are `pillars/four-why-matters`, `pillars/four-why-matters-nav`,
`pillars/four-why-overlap`, `pillars/three-why-matters`,
`pillars/three-why-overlap`, `pillars/three-why-overlap-dense` and
`pillars/three-why-overlap-nav`.

Thirteen cards narrow by 18 pt, thirteen right-angle arrow marks are appended,
and eleven existing arrows shift right by 9 or 18 pt. These use the existing
editable mark primitive. Four keys contain intentional upstream copy edits:
five string replacements and one removed bullet. In particular, the second
pillar of `four-why-overlap` changes from five bullets to four. Its closed
binding contract must follow the current source cardinality.

## Engineering recipe

Run from the pptxgengo repository with the explicit candidate bundle. Each
reference output directory must be new. Before and after use the same ordered
231 keys, allowing native page comparisons by position and template key.

```sh
go test ./internal/wmdesign -run 'TestLibraryV11(FrozenSnapshotClosure|PinnedDensityAndPillarDelta)$' -count=1
go test ./internal/wmdesign -run TestLibraryV11ChangedTemplatesRoundTripAndBuild -count=1
go test ./internal/wmdesign -run TestLibraryV11ChangedIndividualSourceBuilds -count=1
go test ./internal/wmdesign -run TestLibraryV11AllTemplateBuilds -count=1
go test ./internal/wmdesign -run TestLibraryV11UnchangedPublicationRenderInheritance -count=1
go test ./internal/wmdesign -run TestLibraryV11UnchangedSourceRenderAudit -count=1

go build -o /private/tmp/pptxdesign-v11 ./cmd/pptxdesign
intake=planning/wm-design-contracts/v11/intake-20261006-649-frozen
previous=planning/wm-design-contracts/v10/intake-20261006-649-frozen/bundle
keys=$(cat "$intake/changed-template-keys.txt")
/private/tmp/pptxdesign-v11 library-source-reference --bundle "$previous" --template-keys "$keys" --year 2026 --out /private/tmp/wmds-v11-before231-source
/private/tmp/pptxdesign-v11 library-source-reference --bundle "$intake/bundle" --template-keys "$keys" --year 2026 --out /private/tmp/wmds-v11-after231-source
/private/tmp/pptxdesign-v11 library-reference --bundle "$intake/bundle" --template-keys "$keys" --year 2026 --out /private/tmp/wmds-v11-after231-bound
```

The current closure, density/pillar delta, historical role/color/tile refresh
and final contrast-policy/historical-pin tests passed together in 3.269 seconds. They verify the
exact Git acquisition, font/asset manifest, prior token fields, key sets,
removed aliases, unchanged density-only content contracts and executable
inherited amendment guards for all 649 source compositions.

## Latest committed renderer refresh

Exact `03fad66` adds source-driven quote glyph density (40 pt floor), Grounded
non-display shape text on Magenta, and surface-emphasis chevron numbers.
`66a469b` commits the earlier performance/lazy-render changes; upstream is clean.
No template, catalog, token or frame source changes occur versus `4fce3cd`.
[Refresh evidence](renderer-density-source-refresh-audit.json) records five
changed source paths and the retained exact historical
[4fce bundle](historical-source/4fce3cd/bundle/bundle.json). Its executable pin,
contrast evidence and package comparison remain historical evidence.

The historical 03fad metadata-only browser contrast sweep reports 78 failures
across 13 templates (39 compact and 39 dense, none comfortable).
[Failures](browser-contrast-audit-03fad66.json) and
[summary](browser-contrast-summary-03fad66.log) are supplemental; SVG text and
other unannotated elements are not covered. Full Go contrast remains the gate.
The latest engine uses the source-aware 312-control native supplement. Historical
4fce retains its immutable 306-control supplement and 315/334 partition under
`historical-source/4fce3cd`. Current ordered lists describe the final source.

## Final contrast policy and browser guard refresh

Exact `2fc6c29` adds `densityLimit: "comfortable"` to five agenda/schedule
and five key-message/lead-questions variants. Seven Blue ink fields replace
Magenta in `guide/deck-overview`, `status/steering-update` and `venn/four-text`.
Exact `ab8b065` adds the browser's missing `numTile` negative guards without
changing compositions. Twelve valid tile specimens and four negative cases
(surface, inline-number, title and band) passed the focused browser regression.
[Committed policy evidence](contrast-policy-source-refresh-audit.json) records
these exact fields, keys, commit identities and source acquisition.

There are **639 templates supporting all three body tiers**, plus **ten
comfortable-only templates**: **1,927 supported configurations and 20 explicit
unsupported configurations** out of 1,947 combinations. Source/default-auto
must respect each source-owned limit. Unsupported tiers are reported explicitly
and cannot be counted as contrast passes. The browser contrast tool filters
limits; browser stage/preview enforcement remains absent. Go enforces limits.

Exact `3c56d842` closes the last genuine native contrast defects: two maturity
inflection labels now use Highlight Blue and two inactive road pin numerals use
Grounded while their Gray outlines remain unchanged. No raw template, copy,
font, token, frame or geometry edits occur. The browser checks now instrument
these labels and all road pin numerals at their actual rendered size/weight.
[Narrow source and browser regression evidence](native-contrast-source-refresh/audit.json)
records the exact commit/tree, two changed snapshot paths, owned checker and
nine additional specimens across all three tiers. The Go full-plan regression
confirms exactly four text-color edits across these three keys, including
historical V10 isolation.

The final authoritative Go contrast sweep covers 1,947 configurations:
1,927 supported, 20 explicitly unsupported, 62,578 text checks, zero failures
and zero planner gaps. All 649 source and 648 active bound stock builds,
231 source/binding roundtrips and 231 individual changed source builds pass.
The final 418-specimen paired audit passes in 83.72 seconds: 121 additional
render differences, 297 fully equivalent visible dependency graphs and zero
automatic density adjustments. The actual final partition is **352 requiring
new native review** (351 active plus deprecated `from-to/rows`) and **297
eligible for inheritance**. The independent publication inheritance verifier
passes all 297 in 59.25 seconds.

[Final per-key audit](final-unchanged-render-audit.json) and
[artifact hashes, recipe and scope](final-render-inheritance-audit/manifest.json)
retain the complete outcome and temporary rebuildable PPTX/report paths.
[Per-key object differences](final-render-inheritance-audit/render-difference-details.json)
record the XML hashes, visible text and transforms of 293 changed top-level
slide objects across the 121 additional specimens.
The durable packet keeps summaries, hashes and logs instead of another 165 MB
copy of paired decks and reports. The ordered sets follow the pinned catalog;
Set B's 119 keys are exactly the additional 121 minus the two pending color
specimens, now covered by the final Set C.

Native set A contains 230 exact changed-source specimens, excluding
`maturity/insights`. Pages 61–135 were independently viewed at full size and
accepted, with each template/image identity recorded in
[native review evidence](native-review-set-a/review-sol-audit-061-135.json).
Five V10 comparison PNGs are preserved adjacent to that evidence. Acceptance
covers these issued PNG/PPTX bytes; later source refreshes must prove byte
parity before reusing it.

## Native qualification

The final measured partition is 352 new native specimens and 297 inherited
specimens. Raw equality alone does not establish render equality: 121 of the
418 unchanged source compositions have changed visible PowerPoint dependencies.
The final collector compares all visible dependency graphs, including shapes,
images, charts, workbooks, themes and layouts. The publication verifier repeats
the comparison for all 297 inherited specimens.

Native sets A (230), B (119) and C (3) cover all 352 new specimens. This audit
independently accepted every full-size assigned page in Set A (61–135) and
Set B (41–80), with per-page hashes and notes in their adjacent signed-source
review packets. The three source-color fixes were excluded from A/B and are
reviewed in C. Source refreshes require package byte equivalence before these
receipts can be reused. Numeric contrast compliance is a separate gate.

The historical 4fce audit measured 97 geometry differences among 431 unchanged
compositions and a 315/334 partition. Its source pin, source acquisition,
summary/hashes and initial comparison remain historical evidence. The old
stock build audit recorded three automatic Compact→Dense adjustments; the
final full-library audit determines the current stock warnings. All stock
headers remain Comfortable.

Generation tests establish stock source/binding execution parity. Native
review qualifies exact specimens and does not qualify arbitrary authored copy.
This intake does not promote a production bundle or change release defaults.

## Historical artifact retirement

After final publication succeeded (649 entries, 352 new native previews and
297 inherited previews), the superseded 4fce paired decks and full reports were
retired. Their original hashes plus the retained per-key comparison and object
details remain in [historical audit evidence](historical-audits/4fce-render-inheritance/retirement.json).
The historical contrast files for 4fce and ab8b are concise summaries retaining
original full-report hashes, counts and every failing check; they omit passing
objects. The final 3c56 contrast inventory and native A/B/C receipts are intact.

[Retirement inventory](historical-audits/artifact-simplification.json) records
208,909,670 bytes retired or compacted (about 199.2 MiB), including the two
unused provisional 220-key lists. Current ordered 352/297 sets remain intact.
Historical exact source pins, calibration controls and the before218 visual
packet are preserved. This retirement does not change renderer or release code.
