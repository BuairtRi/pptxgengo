# New slide composition

Use `compose` when new slide content should be assembled from editable native text and shapes instead of editing fixed source geometry. This is component composition with bounded schemas, not prose-to-deck generation.

## Minimal valid spec

`pptxcompose` accepts `pptxgengo.compose-spec.v1`; geometry uses points from the slide's top-left. A one-slide spec with a title and two canvas text blocks can start like this:

```json
{
  "schema": "pptxgengo.compose-spec.v1",
  "slides": [{
    "id": "decision",
    "title": "A phased pilot can validate value before expansion",
    "width_pt": 960,
    "height_pt": 540,
    "title_bounds": {"x": 48, "y": 34, "width": 864, "height": 72},
    "title_font_face": "Arial",
    "title_font_size_pt": 26,
    "title_bold": true,
    "title_foreground": "#070154",
    "pods": [],
    "canvas": [
      {"id": "pilot", "kind": "text", "bounds": {"x": 64, "y": 150, "width": 380, "height": 120}, "text": "Pilot one bounded workflow", "font_face": "Arial", "font_size_pt": 18, "bold": true, "foreground": "#070154", "background": "#E8EEF8", "align": "left", "valign": "top", "inset_x": 12, "inset_y": 10},
      {"id": "gate", "kind": "text", "bounds": {"x": 500, "y": 150, "width": 380, "height": 120}, "text": "Expand after agreed evidence and acceptance checks", "font_face": "Arial", "font_size_pt": 18, "bold": false, "foreground": "#070154", "contrast_background": "#FFFFFF", "align": "left", "valign": "top", "inset_x": 12, "inset_y": 10}
    ]
  }]
}
```

Every canvas ID must be unique, and text frames must remain within slide bounds. Overlapping elements require intentional `allow_overlap` declarations. If multiple text objects overlap a surface, declare the overlap on the surface with each text ID. Keep colors explicit; supported WM tokens are documented with the component types, not inherited automatically by every freeform object. For slide-specific dense grids, complex cards, diagrams, team structures, or images, inspect packaged recipes instead of growing this starter ad hoc.

## Discover packaged recipes/components

```sh
pptxgengo paths
pptxgengo catalog --components
pptxgengo catalog --components --print
```

Open the component gallery to browse previews and source provenance; use
`--print` when you need the gallery's installed path. `paths` returns
`catalog_templates` and `catalog_components` as well as the installed `root`,
`library`, and `scripts`. Use `root` as the base for packaged resources; do not
depend on a particular frozen release directory. The gallery distinguishes
fixed-source groups from new-slide composition patterns. A visual-only source
group is a reference to the source artwork, not an executable component.

The supported new-slide paths fall into five semantic families through `adapt`
and the following bounded component recipe areas through `compose`:

- Adaptive family specs cover roadmaps, architecture, processes, teams, and
  comparisons. They create new semantic compositions; they do not structurally
  adapt every source template classified with the same category hint.
- `library/dynamic-components/` provides variable role pods and team structures,
  numbered and metric cards, and explicitly positioned canvas primitives.
- `library/layout-components/controls.json` contains measured grid and panel
  patterns, including phase, phase-detail, and workflow-matrix examples.
- `library/diagram-components/` describes bounded diagram/artwork-arrow cases;
  some arrow placement remains manual and must be explicitly authored.
- `library/showcase/dense-deck.json` is an exact-copy, visually reviewed set of
  five dense proposal specimens. It is a reviewed example set, not a reusable
  generic template contract.

Use each area's README/schema and the current component gallery to confirm
which operation is executable and what evidence/qualification exists. Do not
assume a catalog grouping shares one schema, or that a reviewed specimen
qualifies changed copy.

Additional packaged recipes include:

- `library/dynamic-components/pods.json` and `team.json`: variable role pods and team structures.
- `library/dynamic-components/cards.json` and `cards.md`: a working two-slide compose spec with three numbered rows and three illustrative metric cards. It uses Arial, fixed 960 × 540 point slides, `surface.light`, explicit `#0047FF` side rules, editable text blocks, and measured fixed slots. Metric values are explicitly sample values, not client results.
- `library/dynamic-components/canvas.md`: explicitly positioned canvas elements and accents.
- `library/layout-components/controls.json` and `README.md`: measured grids/panels for bounded dense layouts.
- `library/visual-components/README.md`: rich text, image crop and shape-preset limits.
- `library/diagram-components/README.md`: diagram and artwork-arrow limits.
- `library/showcase/dense-deck.json` and `README.md`: five dense proposal patterns with exact-copy native review evidence.
- `library/showcase/assets.json`: local asset hashes, provenance, dimensions, and use notes.

To start from the packaged working cards recipe, make a copy of `library/dynamic-components/cards.json` at a writable task path, edit its titles/numbered card fields/metric labels and values, then use:

```sh
pptxgengo paths
pptxgengo compose probe --spec /path/to/copied-cards.json --out /tmp/wm-probes
pptxgengo compose measure --bundle /tmp/wm-probes --native-workspace /Users/rscott/Projects/pptxgengo/samples/visual-wave3 --out /tmp/wm-evidence.json
pptxgengo compose fit-report --spec /path/to/copied-cards.json --bundle /tmp/wm-probes \
  --evidence /tmp/wm-evidence.json --out /tmp/wm-fit.json
pptxgengo compose build --spec /path/to/copied-cards.json --bundle /tmp/wm-probes \
  --evidence /tmp/wm-evidence.json --out /tmp/wm-built
pptxgengo compose verify --bundle /tmp/wm-built --native-workspace /Users/rscott/Projects/pptxgengo/samples/visual-wave3 --out /tmp/wm-verification.json
```

`pptxgengo paths` supplies the installed package root; the source file is `<root>/library/dynamic-components/cards.json`. Use unique new output paths. The recipe contains two slides in a single spec, so probe/measure/build/verify the deck together.

The relevant component checkpoint reports six native-verified, visually reviewed Wave 1 grid/panel slides. Five dense proposal recipes have exact native verification/visual-review records. These are bounded accepted specimens; neither proves arbitrary layouts or changed-content reflow. Dynamic pods/teams support variable roles within their own contract. Components include numbered and metric cards; explicit canvas text/surface/line/image/shape types; measured layouts; phases/legend; connectors; and supported accent specs. Images must use local pinned PNG/JPEG bytes and SHA-256. Rich-text/font support is bounded; check the packaged schema and docs for the particular component before authoring.

For broader semantic library contracts, first use `pptxgengo lib find --help` and then inspect an exact record with `pptxgengo lib inspect --id ID`. The initial thirteen portable contracts are candidates; the qualified default search may be empty. Exploratory inventory or fixture state is not approval. Candidate instantiation requires explicit experimental override and the same native QA below.

## Native measurement and build loop

Save the spec as a new file and use fresh output paths. On macOS with PowerPoint installed:

```sh
pptxgengo compose probe --spec /path/to/spec.json --out /path/to/probes
pptxgengo compose measure --bundle /path/to/probes --out /path/to/evidence.json
pptxgengo compose fit-report --spec /path/to/spec.json --bundle /path/to/probes \
  --evidence /path/to/evidence.json --out /path/to/fit-report.json
pptxgengo compose build --spec /path/to/spec.json --bundle /path/to/probes \
  --evidence /path/to/evidence.json --out /path/to/built
pptxgengo compose verify --bundle /path/to/built --out /path/to/verification.json
```

Measurement and final verification use native PowerPoint. These calls run serially. Use a stable installed binary for the qualification run; rebuilding can invalidate native evidence/cache environment. Never measure a same-named open presentation with unsaved work. Probe/evidence must match the exact spec and deck hashes; changed copy needs new measurements. `fit-report` reveals fixed-zone overflow and planner/layout failures separately. A successful command exit on the report means only that the report was written.

The approved native working directory for this checkout is `/Users/rscott/Projects/pptxgengo/samples/visual-wave3`. Pass `--native-workspace /Users/rscott/Projects/pptxgengo/samples/visual-wave3` explicitly to both `measure` and `verify`, as in the commands above. Native PowerPoint operations run serially. Reuse this established folder; do not create a new native workspace. In a different environment where it is unavailable, use that environment's existing approved workspace.

After verify, export/render every page and inspect at presentation size. Measurement validates declared text/font/frame/color/bounds; it cannot determine whether the argument is persuasive, hierarchy is effective, a retained asset is appropriate, or every collision is visually acceptable. Resolve overflow with better geometry, a different pattern, divided argument, or carefully revised copy while retaining essential evidence and qualifications.
