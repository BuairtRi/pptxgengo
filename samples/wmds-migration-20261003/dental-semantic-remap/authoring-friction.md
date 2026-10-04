# Dental57 semantic remap: authoring and review

The final preferred source is `project-v2/deck.yaml`. It retains57 original slide identities,15 hidden slides,7 original native section names/memberships, all original notes, and full original visible wording appended to notes. There are no continuation slides. Visible copy is reshaped for the selected semantic layout while factual qualifications remain intact. Originals and the prior conversion are unchanged.

## Semantic selection

The unified Go SQLite catalog was used across all587 accepted templates.32 relevant candidate previews were individually opened;57 original native source pages were individually opened. The57 authored mappings use25 catalog families, with cardinality or geometry adaptations recorded as local derivatives with real shared-parent hashes. `authored-mapping-v2.json` records purpose, selected zones, parent, reason and local ID.

Three alternatives are meaningful: source10/13/17 use a left working-input narrative and three stacked deliverable cards from `cards/narrative-2x3-slim`, instead of the preferred three horizontal cards. The visible values remain identical and each page highlights a different current deliverable. No unrelated six-pillar alternative is claimed.

## Native findings

Root individually reviewed all57 v1 native pages:50 accepted,7 repair-pending. The repair batch changed3/8/20/35/54/56/57. Legend descriptions on3/8/35 did not match actual drawn colors/shapes;20 clipped its PHI qualification;54/56/57 exposed author editing instructions. All seven revised pages were individually reopened and accepted. Root's per-PNG receipt and exact50-page native-part carry are in `native-qualification-v2.json`.

The layout measurement report showed no estimated collisions/spills before native review. This did not detect20's text overflow inside an editable table: table cell interiors are outside that estimator's scope. Native review was necessary.

## Practical CLI/YAML friction

- `library-find --summary` now exposes verified native screenshot paths; `library-inspect --summary` exposes named zone bindings. Scenario ranking no longer saturates at two matching terms. These supported actual semantic comparison instead of specimen-copy substitution.
- `project scaffold` now carries legitimate parent provenance and explicit synthetic source values. Synthetic values were replaced by source content. Array keys still need deliberate stable IDs during cardinality adaptations.
- Some source layouts cross currently supported composition zones (`cover/photo`, `agenda/index`, `divider/panel-edge`), so bounded truthful derivatives were needed. Typed cardrow recipes cannot safely be borrowed as an unbound body; the preferred three-card pages are honest native component derivatives.
- Source rules needed a native rule node; `wmds/component/rule` is not supported. Stroke placement must reserve the actual ink width.
- Title-line-allocation and invalid/duplicate-node errors would be more useful with the stable slide ID and offending node in the message. Reserved node names include `title`, `eyebrow` and `source`.
- `project edit --patch` atomically preserves notes, hidden states and groups. Its YAML flow output is YAML, not necessarily JSON; use `yq -o=json` before `jq`, and `yq -P` for readable maintained YAML.
- Table cell text fit, actual legend/color association and visible author comments still need native visual and semantic review.

No new Python helpers were authored for this remap. Maintained deck authoring used YAML, `yq`/`jq`, `apply_patch` and the frozen Go CLI. The frozen runtime's pre-existing Python source artifact remains part of its pinned bundle.

## Rebuild

Use frozen `/tmp/pptxdesign-semantic-remap-v4`, SHA256 `370c08c471eb0dc13f6bcb9e57ca6c9d199a1aa467bd831ec9eaf849ff3b2e20`, with `library/wm-design-system/v5`.

```sh
/tmp/pptxdesign-semantic-remap-v4 project check --project samples/wmds-migration-20261003/dental-semantic-remap/project-v2 --bundle library/wm-design-system/v5
/tmp/pptxdesign-semantic-remap-v4 project build --project samples/wmds-migration-20261003/dental-semantic-remap/project-v2 --bundle library/wm-design-system/v5
```

The exported preferred offline ZIP was unzipped into a new location and rebuilt using only its packaged Go executable, bundle, fonts and calibration. Its generated PPTX hash exactly matched the accepted preferred build.

## Minimal retained delivery

Retain preferred/alternative PPTX outputs, the final maintained project(s), corresponding toolchain lockfiles, referenced assets and provenance snapshots, build receipts, mapping and native qualification receipts, and offline/maintainer/client packages. Archive inspection projects, older versions, exploratory search receipts and temporary proofs together. Keep the original source deck unchanged. The new maintained project has no one-off authoring scripts.
