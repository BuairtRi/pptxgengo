# Release 0.1: whole-library customization audit

This is an initial coverage audit of all 65 fixed-source templates. The machine-readable row-by-row detail is in `customization-audit.csv` and `customization-audit.json`; `deck-outlines.json` provides complete source XML text blocks and slide-level title heuristics for both user decks.

## Evidence and boundary

- `pptxtemplate list --root . --available` returned 65 implementations with `adaptation_qualified: false` and inspected bindings. This CLI status is technical readiness, not editorial acceptance or capacity qualification.
- The shared implementation checkpoint reports 65 contracts, 895 named slots, and 1,570 original text bindings; 60 native-export examples and five native-replay examples have individual visual review. See `planning/SHARED_GAPS_IMPLEMENTATION.md`.
- Thirteen portable `pptxlib` contracts are a separate set. The current library index reports 13 contracts; do not conflate that portable set with the 65 fixed-source pattern contracts.
- Current geometry/item-count behavior is fixed per source. Semantic text slots and declared color roles are usable; font roles, arbitrary bullet/count edits, dynamic table dimensions, free-form connector rewiring and data-driven indicators are not general template controls.
- Two bounded empty-shape overlay zones have been demonstrated (t030); they do not provide general repeated-item generation or arbitrary reflow.
- Measured picture accents have been demonstrated on selected phrases in five templates (t002, t004, t014, t048, t054); that does not make general artwork, diagram connector, or font geometry editable.

## What authors will likely ask for

The audit proposes controls per archetype: for example, roadmaps need editable period scales, workstreams, milestone positions, dependency tails, and active/inactive states; architecture maps need layer/component counts and routed relationships; team views need role/pod counts, reporting lines and responsibility axes; comparisons need criteria/option counts and usable gauge scales; metric views need value/unit/baseline/target fields and indicator scaling. Those are desired design controls, not current promises. The CSV makes that distinction explicit for all 65 rows.

## Component routes

- Fixed-source copy/color changes: inspect an existing implementation with `pptxgengo template inspect --id <id>`, then use `template build-review` with approved values in a new output directory. Contract-specific checks are `pptxgengo component check/apply`.
- New variable-count geometry: author a separate `pptxcompose` spec and use its `probe`, `fit-report`, and `build` commands; consult `planning/DENSE_LAYOUT_GAPS.md` for native table/connector limitations.
- Portable contract reuse: `pptxlib` indexes 13 contracts separately; it does not make the remaining fixed-source slides variable-layout components.

References: [template workflow](../../cmd/pptxtemplate/README.md), [component workflow](../../cmd/pptxcomponent/README.md), [dense-layout gaps](../DENSE_LAYOUT_GAPS.md), [shared implementation checkpoint](../SHARED_GAPS_IMPLEMENTATION.md).
