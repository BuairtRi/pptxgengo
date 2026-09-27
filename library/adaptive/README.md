# Adaptive capability declarations

`capabilities.json` is the generated, machine-readable inventory for the initial semantic adaptation surface. It keeps each source-bound template contract separate from a taxonomy hint for further review. A category hint is not a builder-selection or structural-suitability decision. It does not copy the source's visual design, preserve its artwork, or claim source fidelity. In a future installed release that provides the `adapt` route, read the packaged inventory with `pptxgengo adapt capabilities`. The frozen 0.1.0-local.3 release does not provide that route.

In a **source checkout**, regenerate the inventory from local rollout assignments, contracts, illustrative values, source scenes, and native review checkpoints:

```sh
python3 scripts/build-adaptive-capabilities.py
```

Use `--root PATH` for another checkout, `--out PATH` for a different JSON destination, and `--checkpoint PATH` to select the optional adaptive-example checkpoint. The default checkpoint is `library/adaptive/checkpoint.json`; use `--no-checkpoint` to omit it. The default output is `library/adaptive/capabilities.json`. The generator preserves contract slot cardinalities/binding IDs, source color roles and style profiles, constraints, fixed/opaque areas, retained source content, source scene image relationships and any actual source alt text. When the source material does not provide a visual description, the record says so instead of inventing one.

## Record structure

- Top-level `qualification_summary` is the current evidence boundary. `families_with_reviewed_examples` and `reviewed_adaptive_example_count` count only examples explicitly marked `visually_reviewed` in the adaptive checkpoint. The exact controls, paths, verification evidence, hashes, and limitations are preserved under each family's `reviewed_examples`. These counts do not qualify arbitrary changed-content capacity, every supported count combination, or source-template adaptation.
- `family_builders` lists the five semantic family builders implemented in `internal/adapt/`: `roadmap`, `architecture`, `process`, `team`, and `comparison`. Each record includes accepted input fields and bounded cardinality/geometry. Its `schema_path` links to the corresponding implementation file for inspection; packaged copies of those selected files are references, not a complete compilable source checkout. `bounded_examples_visually_reviewed` records scoped example evidence only; `native_qualification` remains `pending` and `source_fidelity` remains `not_claimed` until the family-level qualification scope is explicitly reviewed.
- Each `items[]` record contains the original fixed-source evidence, the semantic recommendation, source/asset metadata links, and review state separately. A source `template_id` supplied to the adapter is report-only provenance and does not select or import the referenced slide. Required `role` and `takeaway` metadata are placed in speaker notes, not visible slide content. The top-level `adaptive_style_controls` lists Arial-only typography, three profiles, semantic color roles, and contrast behavior; family-specific readable-ink overrides may take precedence.
- `fixed_source.source_bound_slots` preserves each declared slot's ordered binding IDs and illustrative value cardinality. `source_bound_color_roles` and `source_bound_style_profiles` are copied directly from the contract. `contract_constraints`, `fixed_areas`, and `retained_source_content` preserve declared boundaries and disclosures.
- `adaptive_family_recommendation.status` is `category_candidate_requires_review` when the source taxonomy is a discovery hint for a family. This is not a suitability assessment or routing decision: structural fit must be reviewed because a category can include forms the current builder does not implement. Other items are `unmapped` by taxonomy, which does not rule out future redesign. The five reference patterns are illustrative; neither category matches nor reviewed examples claim the other source templates are adaptable.
- `adaptive_family_qualification` stays pending for every item. `current_bounded_engine_capabilities` describes existing composer features, not validated support for the new adapter families.

Initial generated totals: 65 fixed-source contracts, 895 named text slots and 1,570 text bindings; 36 taxonomy category candidates requiring structural review and 29 unmapped category records. These 36 are discovery hints, not structural adaptation recommendations. Exact reviewed examples may be reported from the optional adaptive checkpoint, while `qualification_summary.adaptive_family_native_qualified` remains 0. The counts are generated from checked-in inputs and change when those inputs change.

## Asset metadata

`source_assets` points to the exact image relationship and package target in each retained scene, preserving source object name and alt text where supplied. Source presentation imagery often has no actual description in the rollout record; no semantic subject is guessed. `asset_description_metadata.catalog_reference` points to the curated `library/showcase/assets.json` descriptions and restrictions where those exact assets are used. The hosted WM inventory is filename/path metadata, not rich visual descriptions. See [the audit](../../planning/adaptive/CATALOG_AUDIT.md) before recommending assets.

Treat every adaptive family item as a new composition requiring its own source-grounded content, native fit, and visual review. Do not use this catalog to claim that an arbitrary source slide is adaptively qualified.

## Review gallery

In a **source checkout**, build a portable, locally browsable review gallery from the prepared review manifest. The generator and its review manifest are development inputs and are not part of the installed command set:

```sh
python3 scripts/build-adaptive-gallery.py \
  --review planning/adaptive/review.json \
  --out /path/to/new-gallery-directory
```

The input uses `{ "slides": [...] }` with `id`, `family`, `variant`, optional source template/image links, native render path, spec/values links, status, findings, controls and limitations. Review and capability JSON paths may be absolute or relative to the checkout. The gallery copies referenced images, input JSON, and existing evidence files into its own `assets/` directory; evidence files get local links while original provenance paths and hashes remain in `index.json`. Capability cards show a lazy-loaded source preview when the preview file is available, clearly labeled as fixed-source reference only. It writes `index.html` and `index.json` and refuses an existing output directory. Its two views keep variant review separate from all 65 fixed-source capability records. It compares structural controls with a normal variant when both share a source template and family. Missing render images downgrade native review display; missing source or render images prevent a visual-review claim. A supplied artifact hash mismatch downgrades the case to `needs_revision` and identifies the changed file. Taxonomy classifications are discovery hints only.
