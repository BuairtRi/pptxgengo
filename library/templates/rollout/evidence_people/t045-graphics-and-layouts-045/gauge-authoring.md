# T045 source gauge controls

The source slide contains five semicircular gauges. Each gauge has five native
freeform cells and a separate native freeform pointer. `apply-gauge` changes
their fills and pointer transforms in a copied scene project; it preserves the
five-cell arcs, table, text, and other source objects. The two three-segment
scale bands above the table are retained source artwork and are not controlled
by this command.

The input scene must retain the pinned native tree and all source geometry,
including group and pointer transforms. Earlier component edits may change
bound text and fill colors; transform or other non-color binding changes are
rejected before gauge controls are applied.

`gauge-values.json` is an illustrative example. The five `gauges` entries map
to the five comparison rows from top to bottom. `highlight_cells` is a nonempty
list of distinct integers 1–5, ordered from the left end of the arc to the
right. A single highlighted cell sets the default `pointer_cell`; multiple
highlights require an explicit pointer target. `pointer_cell` may differ from
the highlighted cells. `highlight_color` is `pink` (default), `blue`, `navy`, or
`gray`; unhighlighted cells use source gray `#CED7E6`. Selected gray is the
visibly darker `#7F7F7F`. The other exact colors are source pink `#F900D3`,
source navy `#070154`, and standard blue `#0047FF`.

In local release `0.1.0-local.4`, find the frozen resource root with
`pptxgengo paths`. Create a text-applied project, then a gauge-applied project,
then build a new deck. Use new output paths for each command:

```sh
release_root="$(pptxgengo paths | python3 -c 'import json,sys; print(json.load(sys.stdin)["root"])')"
pptxgengo component apply \
  --project "$release_root/samples/template-expansion/rollout/sources/graphics-and-layouts" \
  --contract "$release_root/library/templates/rollout/evidence_people/t045-graphics-and-layouts-045/contract.json" \
  --values /absolute/path/edited-text-values.json \
  --out /absolute/path/new-text-project
pptxgengo template apply-gauge \
  --project /absolute/path/new-text-project \
  --reference "$release_root/samples/template-expansion/rollout/sources/graphics-and-layouts/slides/uhg-045.json" \
  --values /absolute/path/edited-gauge-values.json \
  --out /absolute/path/new-gauge-project
pptxgengo scene build --project /absolute/path/new-gauge-project --slides 45 --out /absolute/path/new-gauge-deck.pptx
```

Start the text file from `pptxgengo template values --id
t045-graphics-and-layouts-045 --source-values` and the gauge file from the
packaged `gauge-values.json` in this directory. To change only gauges, pass the
packaged source project directly to `apply-gauge` and omit the text-edit step.
The previous `0.1.0-local.3` release does not include this operation.

The pointer targets are discrete source exemplar positions 1–5. This command
does not calculate continuous scores, change the number or shape of cells,
edit the upper scale bands, or establish arbitrary-content fit. Changed copy,
multi-cell fills, and pointer/cell disagreement require native PowerPoint open,
fit, and visual review before client use. The output project includes
`gauge-application.json` with pinned input/output hashes and selected controls.
