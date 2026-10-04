# Candidate v4 local composition

Three synthetic slides demonstrate reusable `local_templates` in one maintained `deck.yaml`:

| Slide | Local definition | Executable source adapter |
| --- | --- | --- |
| `align-three-lenses` | `three-lens-pilot` | `wmds/component/venn` |
| `assess-capability` | `four-stage-capability` | `wmds/component/maturity` |
| `sequence-three-gates` | `three-gate-road` | `wmds/component/road` |

**Candidate v4 is unreleased; qualification is pending.** Installed local.7 does not contain these three adapters. This example requires the repository candidate compiler and the frozen `library/wm-design-system/v4` bundle. It does not adopt the later 587-template upstream inventory.

## Maintain the source

All visible copy and data live in `slides[].values`, including frame titles, eyebrows, source caveats, diagram labels, circle numbers and illustrative milestone dates. The reusable definitions retain geometry, typography options, exact topology, typed binding constraints and stable item keys. Venn set order is People/Process/Data; maturity active positions and Venn region indices are zero-based. Road milestone `at` and `side` values explicitly place the authored milestone on the winding path.

Each diagram has a declared body allocation inside a one-line-title compact frame with a reserved source line. Character/item limits constrain the example; they do not guarantee that changed content will fit. Edit a local definition deliberately when its geometry needs to change. Recheck and inspect the resulting pages.

## Pin and build a working copy

Keep generated locks, state and immutable build artifacts outside this example directory. Copy this folder into a new working directory; use absolute compiler and bundle paths:

```sh
cp -R examples/v4-local-composition /tmp/wmds-v4-local-example
/tmp/pptxdesign-v4-feedback3 project init --project /tmp/wmds-v4-local-example --bundle /Users/rscott/Projects/pptxgengo/library/wm-design-system/v4 --engine wmds-go-foundation.v2
/tmp/pptxdesign-v4-feedback3 project check --project /tmp/wmds-v4-local-example --bundle /Users/rscott/Projects/pptxgengo/library/wm-design-system/v4
/tmp/pptxdesign-v4-feedback3 project build --project /tmp/wmds-v4-local-example --bundle /Users/rscott/Projects/pptxgengo/library/wm-design-system/v4
```

Use a fresh destination and the exact candidate compiler selected for qualification. Successful Go compilation or layout is technical evidence. Export through native PowerPoint and inspect every page before recording visual acceptance; the source example currently carries no native review receipt.
