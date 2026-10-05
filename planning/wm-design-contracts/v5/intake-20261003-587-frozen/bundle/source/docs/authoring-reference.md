# Authoring reference: slides, templates and node types

Everything on the reference board is drawn from JSON "nodes". A slide template is
one `slide` node with metadata. This file lists every node type and the slide
geometry, so templates can be composed from the existing library instead of
inventing new shapes.

The exact field shapes for every node are in the examples in
`components/v0/components.json` (search for `"type": "<node>"`), and the renderer
is `explorations/components.src.html` (search for `case "<node>"`).

## Slide geometry (points)

- **Slide:** 960 × 540. **Module:** 18. Text leading is a multiple of 3.
- **Margins:** 57 left and right, 36 top. **Content:** x 57–903 (846 wide).
- **12 columns** of 54 with 18 gutters. Column *i* (1–12) starts at `57 + 72·(i−1)`.
  - Starts: 57, 129, 201, 273, 345, 417, 489, 561, 633, 705, 777, 849.
  - Common spans: 3 columns = 198, 4 = 270, 6 = 414, 8 = 558, 12 = 846.
- **Lattice:** x = 3 + 18k, y = 18k. Place every container, card and block edge on it.
- **5-up grid (sanctioned exception):** five equal columns of 154.8 with 18 gutters across the full content width.
  - Column starts: 57, 229.8, 402.6, 575.4, 748.2. Spans: 1 column = 154.8, 2 = 327.6.
  - Content inside a 5-up column sits on the lattice measured from that column's start (e.g. 229.8 + 12 is fine).
  - Full-width slides only, never beside a rail. Vertical rhythm is unchanged.
  - A 5-up band shares only its outer edges with 12-column content above or below it: separate the bands
    (a rule or 18 pt or more), never stack a 12-column inner edge flush against a 5-up one. The checker flags overlaps.
  - Use it when content truly has five parts (stages, gates, pillars, workstreams); do not use it for four or six.
- **Header zone:**
  - Eyebrow y 36, title y 54. One-line title: rule at 108, body from 126.
  - Two-line title (`titleLines: 2`): rule at 144, body from 162.
  - Appendix density (`density: "appendix"`): heading-size title, rule at 90, body from 108.
- **Body bottom:**
  - Compact footer: 468. Tall footer: 450.
  - With a source line, the body must end at 450 (one line) or 432 (two lines: notes + source).
  - A source line wider than its column (measured at 9 pt) wraps onto a second line automatically and counts as two
    lines, so the body ends at 432. This mostly happens in the 270 pt short column of a split frame.
- **Frames:** `rail` is `none`, `nav`, `left` or `right`; `footer` is `compact`, `tall` or `slim`.
  - `slim`: a shallower footer close to the slide edge (rule 495, row 504–516); the body runs to 486 (468 with one
    source line, 450 with two).
  - `tint`: `{x, w, surface}` (or a list) paints a full-height panel behind the body, e.g. the wide side of a split in
    Light Gray (`surface: "subtle"`) with a photo on the white side. Chrome stays in place; text on the tint uses the
    light inks.
  - `titleLines: 4` allows a statement-length title: rule at 216, body from 234 (split narrow column: rule 198).
- **Split frames** (`split`, no rail or the nav rail): the title zone covers only the short column and the tall
  column runs from y 36 to the body bottom, up through the header band. Standard 18 pt gutter.
  - `tall-right`: short x 57–327, tall x 345–903 (tall is two-thirds, 558 wide).
  - `tall-left`: tall x 57–615, short x 633–903.
  - `tall-right-narrow`: short x 57–615, tall x 633–903 (tall is one-third, 270 wide).
  - `tall-left-narrow`: tall x 57–327, short x 345–903.
  - A 270-wide short column sets the title at heading size (24/30), up to `titleLines: 3`: rule at 108 / 126 / 162.
    A 558-wide short column keeps title size: rule at 108 / 144 / 180. Body starts 18 below the rule.
  - The header dot patch, stamp and source line follow the short column. A 2x2 fits as a 432 square in a tall column.
  - Split frames combine with `rail: "nav"` (the tabs sit at x 21–39, outside the content area).
- **Nav rail** (`rail: "nav"`): vertical section tabs at x 21–39 down the left edge, for long decks. The tabs are
  slots: `"nav": {"items": ["Diagnose", "Design", "Build", "Run"], "active": 1}` (2–6 sections, short names).
  Content area is unchanged (x 57–903). Works with either footer and with split frames.
  - Left panel: x 0–273, rail content x 57–255, main x 345–903.
  - Right panel: x 687–960, rail content x 705–903, main x 57–615.
  - `railSurface` is `inverse`, `deep` or `subtle`.
  - Content placed on a rail must use inks for that surface: pass `"on":"inverse"` on text nodes, or use components on their own surfaces.

## Required chrome (every slide, no exceptions)

- **Legal line:** "© {year} West Monroe Partners | Reproduction and/or distribution without West Monroe
  Partners’ prior consent is prohibited." The frame draws it on every slide, including covers, dividers and
  closing slides, with the current year. There is no way to turn it off; `noLegal` is rejected by the checker.
- **Whiteboard grid:** 3 pt dots on the 18 pt lattice. Every slide has at least one field.
  - Default (omit `whiteboard`, or `"whiteboard": "header"`): a patch 198 × 162 from (21, 18) behind the eyebrow
    and title, fading out to the right and bottom. Beside a left panel it starts at x 291; beside the nav rail, x 57.
  - Covers, dividers and full-media slides name their own fields instead:
    `"whiteboard": [{"x": 561, "y": 36, "w": 342, "h": 18, "fade": "none"}, …]`.
    `fade` is `corner` (default), `right`, `left`, `bottom` or `none`. `"above": true` draws the field over the body
    (dots on a panel laid over a photo); `"on": "inverse"` picks the dark-surface dots. Fields go on the slide, never in `body`.
  - Pattern from the brand decks: a Grounded panel on one side of a divider with the dot field filling the white
    side; on a cover, an L of dots framing the photo; on content slides, the header patch.
  - Dots are Medium Gray on light surfaces and Dark Gray on Grounded and Deep; the frame picks the ink.

## The slide node

```json
{ "type": "slide", "rail": "none", "footer": "compact", "railSurface": "inverse",
  "eyebrow": "Business case", "title": "The program funds itself by [[month 14]]", "emphasis": "highlight",
  "titleLines": 1, "density": "standard", "stamp": "Draft for discussion", "page": "4",
  "source": { "notes": ["Net of one-time cost."], "text": "West Monroe analysis, FY26." },
  "body": [ /* nodes in slide coordinates */ ] }
```

- `[[…]]` in a title marks the emphasized phrase. `emphasis` is `highlight`, `underscore`, `circle` or `spark`.
  - Highlight only on White or the lighter grays, 1–4 words, at most one per slide.
- `[^1]` in any text becomes a footnote marker; its note goes in `source.notes`.
- A slide without `title`/`eyebrow` (cover, divider, closing) places its display text as body nodes.
- `surface`: `light` (default) or `inverse` / `deep` for dark slides (dividers, closing). Chrome and body inks follow it automatically.
- `noPage`: drop the page number (covers only). The legal line cannot be dropped.
- `whiteboard`: see Required chrome.
- Rail content: give text nodes `"on":"inverse"` (or the rail surface) so they take that surface's inks.

## Node types

**Text and headings**
- `text` {style, ink, x, y, w, text, emphasis?, on?}
  - Styles: display, title, heading, subhead, lead, body, small, source, stat, stat-sm, number, eyebrow, label.
  - Inks: display, primary, secondary, emphasis, callout, or series.N.
- `rule` {x, y, w, weight?, ink?}
- `grouplabel` {x, y, w, text}
- `numhead` {x, y, w, n, text, numInk}
- `colhead` {x, y, w, rule, weight?, label, title}
- `textblock` {x, y, w, label?, title?, body}

**Lists**
- `bullets` {x, y, w, size?, items: [string | {lead, text} | {text, sub: []}]}
- `ol` {x, y, w, size?, items}
- `strongnum` {x, y, w, numInk, numStyle?, size?, items: [{title, text}]}
- `size: "small"` sets a list to 12/18 for dense zones (phase detail, side columns, appendix). Default is 14/21.
- `schedule` {x, y, w, keyW, rowHeight, keyInk, items: [{k, t}]}
- `list` {variant: "index", x, y, w, rowHeight, items: [{n, title, page}]}

**Cards** (see card.* examples)
- `card` {x, y, w, h, surface, bodySize? ("small" = 12/18 body), pad?, label?, title, titleInk?, titleStyle?, inlineNumber?, number?, body: [{p} | {bullets} | {columns} | {checklist}], edge?, band?, bandNumber?, icon?, badge?, media?, metric?, metricGroup?, person?, bio?, case?, quote?, state?, tag?, fee?}
- `cardrow` {x, y, w (each), h, gap, numbering: "inline" | "band" | "corner", card: {...shared}, items: [...]}

**Data points**
- `metric` {x, y, w, value | format, label, change?, secondary?, target?, status?, source?, circle?}
- `callout` {x, y, w, label?, value, text, secondary?}
- `pullquote` {x, y, w, text, by, markInk?}
- `legend` {layout, x, y, w, items, title?}
- `gauge` {x, y, w, segments: [palette refs], value, valueText, caption, ends}

**People and sequence**
- `person` {x, y, w, size, initials | photo, name, role, org?}
- `role` {x, y, w, h, title, meta, edge}
- `stepper` {x, y, w, steps: [{label, title, state}]}
- `vstepper` {x, y, w, steps: [{label, title, text, state}]}
- `phasehead` {x, y, w, n, title, duration, rule}
- `timeaxis` {x, y, w, periods, today, milestones}

**Shapes and diagrams**
- `block` {x, y, w, h, surface, text, style, align?}
- `chevron` {x, y, w, h, first?, surface, text, style?, number?, sub?}: with `number` or `sub`, the number, title and subtitle stack left-aligned inside the arrow.
- `textarrow` {x, y, w, h, dir?, surface, text, style}
- `connector` {points: [[x, y], …], label?, labelPos?, style?, head?, elbow?, ink?, startDot?}
  - `style`: `solid` (a working flow), `dashed` (planned or optional), `dotted` (informal or advisory, e.g. a liaison or dotted-line report). `dashed: true` still works.
  - `head`: `end` (default), `start`, `both` (two-way exchange) or `none`.
  - `elbow`: `"h"` or `"v"` turns two points into an orthogonal route (horizontal or vertical first). Prefer it to diagonal lines.
- `node` {x, y, w, h, surface, icon, text, sub?, layout?}
- `frame` {x, y, w, h, style: "solid" | "dashed" | "filled", label}
- `container` {style: "region" | "boundary" | "layer" | "frame" | "external", x, y, w, h, label?, labelPos?, bullets?}
- `layer` {x, y, w, h, label, cells}
- `layerrow` {x, y, w, h, n?, label, text, surface?}
- `pod` {x, y, w, title, band, roles}
- `cylinder`, `device`, `screen`, `plane`: see the architecture examples.

**Composites**
- `table` {x, y, w, header, cols: [{k, label, w, type?}], rows, rowHeader?, groups?, highlight?, dense?, rowH?, preset?, runRate?, continued?}
  - Dot scores: cell type `dots` takes a number or `{value, max?, ink?, text?}`; filled dots take the ink (default
    Grounded; any palette ref such as `emphasis`, `callout`, `series.2`), empty dots are de-emphasis gray. `text` (a
    string or bullet list) sits under the dots. Set `ink` per row (`row.ink`) to color a whole row, e.g. one color per
    party. `rating` and `harvey` cells also take `{value, ink}` or `row.ink`/`col.ink`.
  - Cell types: num, delta, status, check, rating, harvey, maturity, gauge, tag, raci, bullets, allocation, dots, checkbox (an empty box for walk-through checklists; `true` ticks it), icon (`{"icon": name, "text": …}` or a name; 36 pt at most).
- `matrix` {x, y, w, labels?: "above" | "left", labelW?, cols?, cellH?, gap?, rowGap?, style?, rows: [{n?, label, surface?, outline?, cells: [string | {text, surface}]}]}
  - Labels above make a layer map; labels left with `cols` make a state grid (legacy, coexist, modern). `outline: true` puts a Magenta outline round one row (one per slide).
- `chart` {kind: "column" | "bar" | "line" | "pie" | "doughnut" | "scatter" | "quadrant", …}: see the chart.* examples.
  - Quadrant `key`: `true` draws numbered circles with a built-in key beside the chart; `"markers"` draws the numbered
    circles only, for a template that sets its own numbered legend in another column.
- `gantt`: see advanced.gantt and roadmap.
- `horizons`, `phases`, `orgchart`, `governance`, `feesummary`, `swimlane`, `cycle`, `pyramid`, `beforeafter`, `annotation`.

**Maturity curves**
- `maturity` {x, y, w, h, stages: [{label, text?, n?}], at?, shape?, active?, axisLabel?, axis?, inflection?, inflectionLabel?, branch?: {from, n, label, text}, labelW?}:
  a smooth curve that stays flat and then rises steeply (`shape`, default 4.2; higher is a sharper bend), with
  numbered stage markers on it, a short tick from each marker to its stage label and text, and an arrowhead. `at`
  places stages along the curve (0–1). `active` fills one marker. `branch` draws a flatter second path leaving one
  stage (e.g. "4.5 Beyond product delivery"). `here: {stage?, label?}` adds a Magenta "You are here" tag under a marker (default the active one).
  `headroom` (pt) lowers the curve's top so the last stage's text clears the title rule. Give it the body width and
  300–340 pt of height.

**Venn diagrams** (labels are always mono caps: 13 pt set labels, 11 pt overlap labels, 13 pt centre; overlap
anchors are computed per region; any label takes `dx`/`dy` nudges)
- `venn` {x, y, w, h, sets: [{label, text?, bullets?, fill?}], regions?: [{in: [0, 1], label?, text?, bullets?}], points?: [{x, y, label}], textW?, opacity?}:
  two, three or four overlapping circles sized to the box (give it height: about 360–430 pt tall for three or four
  circles, so a split frame's tall column suits it). Each set's label and text sit in its own part; `regions` place
  text at the overlap of the listed sets (`in: [0, 1, 2]` is the centre of three); pairwise overlaps are placed inside
  their lens, and a region may set its own `w`. Keep overlap text to a short label or one line. `points` plot labelled items at
  0–1 positions in the box, like a 2x2: each point is a numbered marker (`n`, or its order) with an optional short
  mono badge (`label`) to its right, or left with `side: "left"`. `labelStyle: "mono"` sets the set labels as mono caps. Default fills step through light grays and a soft blue and pink at 70%
  opacity with Grounded outlines; `fill` takes a role or palette ref.
- Legend items take `swatch: "dot"` and `ink` (role or palette ref) to key dot scores.

**Arrows:** `mark: "arrow-straight"` is a level hand-drawn arrow pointing right (`flipX` for left). Use it where
the curled arrows add too much direction, e.g. from a panel to the cards it leads to.

**Funnels, pyramids, roads, cycles and brackets**
- `funnel` {x, y, w, h, stages: [{label, value?, text?, surface?, active?}], shapeW?, neck? (0.3), gap? (6),
  labelSide?, ramp? ("light"), dark?: false}: trapezoid bands narrowing to a neck, darkest at the bottom. The label and
  value sit inside each band. When `shapeW` < `w`, each stage's `text` sits in a column to the right with a dotted
  leader (`labelSide: true` moves the label there too). `active` fills a band Magenta.
- `pyramid`: the same fields with `levels`, widening down from an apex, darkest at the top. Use a funnel for
  conversion or narrowing and a pyramid for hierarchy or foundations.
- `road` {x, y, w, h, milestones: [{label, date?, text?, at?, side?, active?, n?}], direction? ("left"), amp?, waves?,
  roadW?, labelW?}: a winding road rising across the box with numbered pins. Labels alternate above and below.
- `cycle` {x, y, w, h, items: [{label, text?, n?, surface?}], active?, center?: {label, title, text}, loops?:
  [{from, to, label, side?, bend?, ink?}], start? (deg, -90 = top), closed?, ring?, nodeW? (162), nodeH? (72),
  numbered?, align?}: a non-linear lifecycle. Items sit around an ellipse joined clockwise by arrowed arcs. `loops`
  are dashed Blue curved arrows for loop-backs (e.g. Optimize → Define). `closed: false` drops the last arc. Use it
  for PDLC and SDLC, never chevrons.
- `bracket` {x, y, w | h, orient: down|up|left|right, label?, depth?, ink?}: a square bracket with a centre tick,
  for marking spans (e.g. the SDLC across part of the PDLC).

**Heat maps**
- Table column `type: "heat"`: the whole cell is shaded on a five-step ramp by value (0..`max`, default 4), with a
  thin white gap between tiles. Cell value is a number or `{value, text?, scale?}`; `showValue: true` prints the
  number (mono); `text` prints a short word instead. `scale: "seq"` (default) climbs light gray-blue → Blue → Grounded,
  with White text on the top two steps; `scale: "risk"` climbs neutral → Magenta (hot = bad), always Grounded text.
  `scale` can be set per column, per row or per cell.
- `block` with `heat` (and `heatMax`, `heatScale`) fills the block from the same ramp: use it for heat-tile
  capability maps and risk grids drawn from blocks.
- `legend` items with `heat: 0..4` (and `heatScale`) draw ramp swatches; give five items for the full scale.

**Status values:** on, risk, off (RAG) plus work-item states done, progress, notstarted, blocked, pass, fail. A
status legend shows on/risk/off unless it sets `statuses: [...]`. Relabel a status without changing its colour: a table
status column takes `labels: {on: "Covered", risk: "Partial"}`, a cell can be `{status, label}`. A `legend` item can be
`{status: "off", text?}`: the same 8 pt status square, with the status name as default text.

**Tables:** a plain cell can be `{text, sub}` to add a secondary line, e.g. a row header with a subtitle.

**Logos, maps and curves**
- `logoslot` {x, y, w, h, name, src?, caption?}: a partner, client or cloud-provider logo. With `src` (a data URI or URL)
  the logo fits inside the box (never cropped); without one it draws a hatched placeholder labelled `name`, which is
  what templates ship with. Logos in a row share one box size (e.g. 126 × 54).
- `dotmap` {x, y, w, pitch?, dotInk?, points: [{name, lon, lat, kind?: "office" | "hub", labelPos?: "r" | "l" | "t" | "b", label?: false}]}:
  a dot-matrix map of the continental US from real outlines (2.4 pt dots on a 6 pt pitch by default; `pitch: 9` gives the
  coarser whiteboard look; height about 0.54 × w). Offices are Grounded
  dots, talent hubs Blue rings, each labelled. Locations outside the US (London, Costa Rica) go in a list beside it.
  - `region: "americas-uk"`: the US, Mexico and Central America down to Costa Rica (left 74% of `w`; total height about 0.61 × `w`, so
    `w` 558 gives about 340), plus a framed Great Britain inset top-right. Points land in the
    panel that contains their lon/lat (London -0.13, 51.51; San José, Costa Rica -84.09, 9.93).
- `teamcurve` {x, y, w, h, phases: [{label, sub?, at?}], series: [{name, values: [...], fill?, style?: "area" | "line", dashed?, labelAt?}], at?, max?, tension?, phaseH?}:
  stacked smooth bands across phases, one value per point per series (the curve is edited as numbers, not as a
  drawn shape). Later series stack on earlier ones; `style: "line"` draws a line instead of a band. Default fills:
  Light Gray, Grounded, Blue. `labelAt` is the point index where a band's direct label sits. Phases divide the width
  evenly unless `at` (0–1) is given. Curves are eased by default: each band edge is resampled and Gaussian-smoothed
  (`smooth`, default 0.07 = sigma as a share of the width; higher is rounder, `smooth: 0` gives the plain monotone
  cubic through the values, `curve: "catmull"` the older spline). Smoothing spreads ramps, so keep in-curve text
  inside the plateau, not on a ramp. Phase labels sit below the plot (`phaseH`, default 54). When phase names already head
  text blocks above the chart, set `phaseLabels: false` and `phaseH: 0` so they are not repeated.

**Marks and media**
- Icons: any `icon` field takes a name from the brand icon library (222 icons): see `docs/icons.md`.
  Inks are Grounded, Magenta (one highlighted icon) or White on dark; sizes 36, 54 or 72 pt.
- `mark` {mark, x, y, w, ink?, flipX?, flipY?, rotate?}
- `whiteboard` {x, y, cols, rows, color}: 15 × 15 dots = 255 pt at 18 pt pitch.
- `thumbnail` {x, y, w, h, photo? | src?, kind?: text | table | chart | diagram, stack?, caption?}: a deliverable page with a hairline border; without an image it draws a placeholder page of that kind; `stack: true` adds two pages behind it. Sizes 126 × 72, 162 × 90, 198 × 108 (16:9) or 90 × 117 (portrait). Use it instead of `screen` for deliverables.
- `square` {x, y, size, photo? | surface, stat?, label?}
- `imageframe` {x, y, w, h, photo, focus}
- `logo` {x, y, w, variant: "pos" | "rev"}
- `art` {src: "tagline-rev", x, y, w}
- Photos available:
  - People, healthcare: `photo-clinical-team`, `photo-clinical-leaders`, `photo-clinician-data`, `photo-corridor`.
  - People, office: `photo-team-meeting`, `photo-working-session`, `photo-executive`.
  - Industry: `photo-technician`, `photo-warehouse`.
  - Portrait: `photo-headshot` (the sample headshot for full-page bios; `"focus": "50% 22%"`) and `photo-headshot-face`
    (a tight face crop of the same image for person tiles and bio cards).
  - Abstract (good for dividers and closing): `photo-abstract-cubes`, `photo-abstract-grid`, `photo-abstract-blocks`,
    `photo-abstract-led`, `photo-stethoscope`.
  - `"grayscale": true` on `square`, `imageframe`, `person`, a card's `bio` or `media` turns a photo greyscale. It is a design
    choice: apply it to every portrait on a slide or to none.
  - Media may break out of a panel: an `imageframe` or `square` placed across a rail edge or panel edge (half on the
    panel, half on White) is a brand move. Keep its edges on the lattice.

**Collaboration**
- `reviewnote` {x: 942, y: 0, status, owner, due, updated, notes}

## Seeing your work

Render any template to PNG (1920 × 1080) and look at it before calling it done:

```
python3 tools/render_template.py --out <folder> --file templates/library/<family>.json  # every template in a family
python3 tools/render_template.py --out <folder> <family>:<id>/<variant>                 # one template
python3 tools/render_template.py --out <folder> --grid <family>:<id>/<variant>          # with the column grid
```

Elements the renderer flags (contrast, off-palette, grid rules) show a small warning tag on the image. Then run
`python3 tools/check_templates.py` (every family) or name one file, and `python3 tools/build_catalog.py` to refresh the index.

## Where templates live

- `templates/library/<family>.json`: the templates, one file per slide family (openers, argument, evidence, solution,
  approach, commercials, team, proof). Schema `wmds.templates.v2`: each template has `id`, `variant`, `name`, `tier`
  (`core` or `working`), `purpose`, `uses`, `legacy`, `budget`, `slots` and `slide`.
- The key `id/variant` is unique across the library. Reuse an existing `id` for a new variant of the same slide type
  (all team slides are `team/…`, all architecture slides `architecture/…`).
- `templates/catalog.json` is generated by `tools/build_catalog.py`; never edit it by hand.
- `templates/library/_gaps.json` records needs the renderer cannot meet yet.

## Composing and varying templates

Templates are starting points, not fixed artwork. Three kinds of variation are always allowed without a new template:
- **Card treatment.** Any `card` in a template can switch treatment: plain, title band (`band`), icon (`icon` with
  `layout`), badge (`badge`: number, metric, icon or image), accent edge (`edge`) and state (`state` plus `tag` or
  `placeholder` text). A 2×2 of plain cards can become a 2×2 of icon cards or badge cards in place. Keep one treatment
  per row of cards.
- **Frame.** Most templates can move to the nav rail, the tall footer, a left or right panel, or a split frame. The
  `-nav`, `-tall`, `-left`, `-right` and `-split` variants in each family show how the content refits.
- **Count.** Card rows, column layouts and stepped lists flex between their stated minimum and maximum (see each
  template's `slots`): recompute widths on the 12-column grid, or use 5-up for exactly five.

A new template is justified only when the composition changes (a new arrangement of zones), not when a treatment,
frame or count changes.

## Template rules

- A template is a frame plus components. Never draw a new shape when a component exists.
- Keep to the type scale and the lattice. Body text no smaller than 14 pt (12 pt only for sources, roles, table cells and appendix density).
- One mark per slide (highlight, underscore, circle, spark or hand-drawn arrow). The whiteboard grid goes on White only.
- Budget words: simple content slides stay under about 70 words; covers and dividers under 20. Dense working
  slides (phase detail, solution architecture, team and RACI, commercial summaries) commonly carry 120–220 words
  at 12–14 pt; set the budget to match the slide's job.
- Sample content is realistic and consistent: a finance-transformation proposal for a fictional client, "Northfield Health System". Use no real client names.
