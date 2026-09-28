# Requested source layout templates

## Implemented

The 44 explicitly selected slides are registered in the executable template
library: 11 Lab slides and 33 UHG slides. Eight exact UHG contracts already
existed; 36 source contracts were added. The repository now registers 101 source
variants. This count does not claim 101 unique designs or dynamically reflowable
layouts. Graphics and Layouts selections are still to be supplied by the user.

- Lab: 2, 3, 4, 6, 13, 14, 18, 19, 20, 21, 23.
- UHG: 5, 7, 8, 10, 12, 15, 16, 17, 18, 20, 22, 24–45.
- Inventory: `library/templates/requested-templates.json`.
- Components: `library/templates/requested-components.json` records 132 curated
  occurrences, with 124 executable text/style subcontracts and 8 retained visual
  groups. These are occurrences, not a deduplicated unique-design count.
- Gallery: `samples/requested-templates/gallery/index.html` includes every selected
  source render, content JSON, contract, component controls, caveats, and two
  reference decks. Arrows/accents appear within components rather than inflating
  the template count.

## Use from this checkout

The installed `0.1.0-local.3` binary and installed skill are unchanged. These
additions are available in this source checkout; no new release was installed.

```sh
go run ./cmd/pptxtemplate list
go run ./cmd/pptxtemplate inspect --id t066-ai-accelerator-002
go run ./cmd/pptxtemplate components --id t066-ai-accelerator-002
go run ./cmd/pptxtemplate values --id t066-ai-accelerator-002 --source-values > /tmp/lab-layout-values.json
# Edit the original content fields, preserving their source run order.
go run ./cmd/pptxtemplate build-review --id t066-ai-accelerator-002 --values /tmp/lab-layout-values.json --out /tmp/new-lab-slide
# Or build unmodified reference layouts from several exact IDs.
go run ./cmd/pptxtemplate build-review --ids t066-ai-accelerator-002,t077-uhg-005 --source-values --out /tmp/new-layout-references
```

Outputs must be new directories. Review bundles group slides by source deck;
this operation does not merge several source decks into an existing user deck.

The original text, typography, image crops, native groups, geometry and master
resources remain the default. Source values retain names, claims, metrics,
testimonials and source typos, which an author must review when repurposing a
slide. Existing line breaks and tabs now survive source-bound editing. A changed
run must preserve rich-run cardinality. It may use plain text without controls
or retain the original tab/line-break sequence; it cannot add or reorder those
controls. Preserve source breaks by default and review changed wrapping natively.

## Original gauge customization

`template apply-gauge` operates on the original T045 five-cell semicircular
freeforms. For each row choose highlighted cells, pointer cell and highlight
color (gray/navy/blue/pink). With one highlighted cell, an omitted pointer follows
it. Multiple highlights require an explicit pointer. The pointer can target a
nonhighlighted cell. The surrounding table and top scale bands stay intact.

Read `library/templates/rollout/evidence_people/t045-graphics-and-layouts-045/gauge-authoring.md`.
The reviewed example is in `samples/requested-templates/gallery/gauge/`.
The earlier linear/dial comparison composers are separate new designs and should
not be substituted when the author asks to customize an original gauge.

## Evidence and limitations

- `contract-inspection.json`: all 44 main templates pass binding/value loading.
- `component-inspection.json`: all 44 main and 124 executable component contracts
  pass source-values inspection/checking, with exact input hashes.
- `build-equivalence.json`: both source-values template decks are byte-identical
  to the independently reconstructed copies rendered in native PowerPoint.
- `source-render-comparison.json`: all 33 UHG content areas match prior native
  previews exactly. Four entire PNGs match; the other 29 differ only in the
  inherited footer-logo rasterization, maximum channel delta 11.
- Root inspected all 11 Lab renders; the UHG reviewer inspected all 33 source
  previews. `gauge-review.json` records the exact native gauge example and two
  corrected QA findings (visual row order and indistinguishable gray fill).
- No changed-content capacity, automatic row/phase/pod counts, arbitrary image
  replacement, or universal font/palette switch is claimed by these imports.
  Colors are controlled only where exact roles/profiles are declared. Native
  geometry and original counts remain fixed unless a specific component operation
  implements that change.
- Source imports retain opaque embedded objects and images; text labels around
  them do not make their internal data editable.

No Go test suites were added or run. CLI builds, real contract operations,
PowerPoint exports, image comparisons and visual inspections supplied the checks.

## Cleanup

`cleanup.json` records 20 superseded probe folders moved under
`samples/_archive/2026-09-27-superseded-adaptive-probes`, retaining old paths with
compatibility symlinks. All 36 archived files match their original hashes. Core
sources, working decks, accepted evidence, native-cache, and the approved native
workspace remain intact. Only the user's Lab deck was open in PowerPoint at the
cleanup check; it was untouched.
