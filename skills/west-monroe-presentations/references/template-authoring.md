# Fixed-source template authoring

Use this route when a source slide's retained structure and fixed layout already fit the story. It produces review copies; it does not merge native slides from several source decks into an existing proposal.

## Locate and inspect

Open the packaged template gallery, choose an exact source layout, then inspect its contract and controls:

```sh
pptxgengo paths
pptxgengo catalog --templates
pptxgengo template inspect --id t053-graphics-and-layouts-049
pptxgengo template components --id t053-graphics-and-layouts-049
```

`paths` returns JSON keys `root`, `library`, `scripts`, `catalog`,
`catalog_templates`, `catalog_components`, and `skill`. Open a gallery with
`pptxgengo catalog --templates` or `pptxgengo catalog --components`; add
`--open` to open it or `--print` to print its packaged path. The current package
catalog contains 101 source contracts, 124 executable source component groups,
8 visual-only groups, and 132 component occurrences. Occurrences are repeated
uses, not unique designs; totals can change with package releases. The inspected
contract and values are under `library/templates/rollout/<lane>/<id>/` beneath
the installed root. `inspect` reports slot names, each slot's ordered source
binding IDs, profiles, and source identity. `components --id` reports grouped
controls for that source slide. A listed group may be visual-only; use its
contract to confirm executable status, exact bindings, and bounds. Choose based
on purpose and geometry, not only visual resemblance.

To retrieve baseline values, use `pptxgengo template values --id ID` for the
registered illustrative example, or add `--source-values` for the original
source-run contents. `build-review --ids ID1,ID2 --source-values` creates a
source-content reference bundle for selected contracts. That bundle is useful
for preserving and comparing the original design; it is not custom client copy.

To inspect the record's exact editable field count and illustrative values, open `contract.json` and `example-values.json` from the reported package path. The values shape is:

```json
{
  "slots": {
    "title": ["Four-step learning cycle for modernization"],
    "diagram_step_2": ["STEP 2"],
    "step_2_explanation": ["Build and test a thin capability slice."],
    "cycle_center": ["PILOT", "LEARNING"]
  }
}
```

The example is from `t053-graphics-and-layouts-049`; it is a shape illustration, not a claim that this slide suits a specific pursuit. Each array length must match the contract's exact binding/run cardinality and order. Preserve every slot you are not changing by copying the complete `example-values.json`, then edit only required values. Use `pptxgengo template inspect --id ID` for the actual contract; do not reuse the example slot set for another ID. Unknown slots, reordered run segments, stale source scenes, and incompatible lengths are rejected. Content is never spread across runs automatically.

## Build a review bundle

```sh
pptxgengo template build-review --id t053-graphics-and-layouts-049 \
  --values /path/to/custom-values.json \
  --out /path/to/new-review-bundle
```

For an example-value review, omit `--values`; to build a set use `--lane` or `--category` after inspecting candidates. `--out` must be a new directory. The result snapshots its inputs and contains source-preserving review decks, edited scene projects, change logs, source/resource hashes, and pending fit/render/review status.

After build, render each page in PowerPoint using the packaged native review workflow exposed by `pptxgengo paths` (`scripts`), then compare every changed page with its source. A changed-content build can still overflow, collide with artwork, retain irrelevant facts/logos/charts, or misstate content. Review retained content and source facts before client use.

## Boundaries

Contracts have named text slots, selected rich text/run structure, and sometimes explicit RGB/theme color roles or text zones. They retain fixed source geometry and do not qualify variable item count, automatic reflow, universal font/style behavior, or arbitrary content capacity. A corrected illustrative example is still evidence for that example only. Use the packaged release checkpoint for the current per-item open/export and accent evidence; those checks do not establish arbitrary-content capacity.

When the ask requires substantially different hierarchy, added rows, or a combined story from different source decks, use `compose` for new component-built slides or `scene` for source-preserving edits. If slides from several source decks must be inserted into an existing proposal, the current CLI does not perform a cross-source native merge; create and inspect the source-specific outputs, then use PowerPoint's copy/import flow as a separate reviewed authoring step.

## Native render and optional phrase accents

Resolve `release_root` from `pptxgengo paths` (JSON field `root`), then use the
existing approved native workspace. For this installation:

```sh
release_root="$(pptxgengo paths | python3 -c 'import json,sys; print(json.load(sys.stdin)["root"])')"
python3 "$release_root/scripts/render-template-review.py" /path/to/new-review-bundle \
  --name unique-review-version \
  --native-workspace /Users/rscott/Projects/pptxgengo/samples/visual-wave3
```

The output's `native-render.json` links exact native artifacts and PNGs. Render
success still requires visual review. Native exports are serial and can require
PowerPoint access to the established workspace.

For T045's bounded native gauge styling, read the contract's packaged
`gauge-authoring.md`. `apply-gauge` changes per-gauge cell highlights and a
discrete source pointer while preserving the original gauge artwork. It is a
separate source-specific component operation, not a template-wide palette. Use
only the installed route and values schema shown in that guide.

For supported phrase accents, copy and edit the intent file at
`library/templates/rollout/accent-intents.json`. Its shipped phrases target the
illustrative example copy; select the exact phrase for the new narrative.

```sh
pptxgengo template adapt-accents --bundle /path/to/new-review-bundle \
  --intents /path/to/edited-intents.json --out /path/to/new-accent-bundle \
  --name unique-accent-version \
  --anchor-bin "$release_root/bin/pptxanchor" \
  --scene-bin "$release_root/bin/pptxscene" \
  --native-workspace /Users/rscott/Projects/pptxgengo/samples/visual-wave3
```

This measures the phrase, solves supported artwork placement, and rebuilds a new
bundle. Render that output and review it. It does not choose semantic emphasis,
place every arrow family, or guarantee a useful result for unsupported geometry.
The historical geometry-replay recovery script is a task-specific diagnostic,
not the normal authoring/render workflow.
