# Designing a custom slide

Read this before building any slide that is not a shared template used as is.

- Do not build a slide as a title plus one large text box.
- Do not build a grid of identical boxes filled with paragraphs.
- Do not go custom because copy didn't fit a template.
- A custom slide must look like it belongs in the shared library. If it would look out of place next to the shared templates, it isn't finished.

## Before you go custom

Work down this ladder and stop at the first step that works:

1. **A template as is.** The content's shape matches the template's shape.
2. **A supported template variation.** Look for sibling card treatments, frames
   and item-count variants in the library. Use only options exposed by the chosen
   contract; stock slots and geometry are fixed. A supported sibling is still
   template reuse. For typography alone, try the slide density presets.
3. **Edited copy in a template.** Tighten the copy (see [narrative and copy](narrative-and-copy.md)), or split the idea across two pages that each use a template. If the edit changes the meaning, ask the operator.
4. **A fork of the closest template.** When a template is about 80% right, scaffold a local derivative of it (`project scaffold --bundle v11 --template KEY --reason R`, matching the project's lock) and change only what is missing ([editing slides](editing-slides.md#change-a-slides-layout)). Record what the original got right and what you changed. For copy density alone, use the [slide presets](typography-density.md) first.
5. **A new composition.** Only when every candidate fails for a reason about the content's structure or the argument, not only because the text was too long.

"The text didn't fit" is not on its own a reason to go custom. First try a roomier variant, a different frame, edited copy, or a split. Every step you try goes in the [composition log](#record-every-decision).

## Think like an information architect first

Do this on paper, in the composition log, before placing a single node.

### 1. Name the slide's job

- What question does the reader bring to this page?
- What is the one-sentence answer? That is the title.
- What should the reader do or believe after reading it?

### 2. Inventory the content

List every item that will appear and label its role:

| Role | Examples |
| --- | --- |
| Headline | The title |
| Key message | A lead-in or callout that carries the "so what" |
| Groups | Three workstreams, four risks, two options |
| Items within groups | Bullets, steps, roles, metrics |
| Relationships | Sequence, flow, ownership, comparison, hierarchy, cause |
| Evidence | Numbers, quotes, examples, sources |
| Qualifications | Caveats, assumptions, illustrative markers |
| Ask | A decision, next step, owner, date |

Count the items and the words. Note which items are equal in weight and which are not.

### 3. Decide what must stay together

Some content is one story and loses its meaning when split. Keep these on the same page:

- a claim and the evidence that supports it;
- a claim and the caveat that limits it;
- both sides of a comparison;
- every step of a sequence the reader needs to see whole;
- a decision, its options and the recommendation;
- a number and its basis.

Other content can split cleanly: a list of parallel items that can become two pages with their own conclusions, or detail that can move to an appendix with a pointer. Write down the story units before deciding the page count.

### 4. Choose the structure from the relationships

Use the table in [slide composition](slide-composition.md#choose-the-form-from-the-content-relationship): a sequence becomes a process flow or stepper, options become a table or matrix, ownership becomes a swimlane or RACI, quantities become a chart. If no relationship exists and the items really are parallel, use bolded-lead bullets or a card row.

### 5. Set the hierarchy

Decide what the eye reads first, second and third. Usually: title, then the structure (group headings, diagram labels, the big number), then detail. At most three levels. If everything has the same visual weight, the hierarchy is missing.

### 6. Set the density budget

Match density to the page's job:

- **Simple content pages:** about 70 words or fewer.
- **Covers and dividers:** under 20 words.
- **Dense working pages** (phase detail, solution architecture, team and RACI, commercial summaries): commonly 120–220 words at 12–14 pt.

Use the pinned source's role scales and [whole-slide density presets](typography-density.md):
standard body is 14 / 12 / 11 pt, small/card copy 12 / 11 / 10 pt, and small
table cells can reach 8 pt. Keep corresponding roles consistent across the slide.
Do not shrink individual elements below their roles to force a fit. When supported
tiers still exceed the allocation, change the structure or split the page.

## Then design it like a designer

### Start from the grid

- The slide is 960 × 540 points with a 12-column grid (columns of 54, gutters of 18, content from x 57 to 903). Place every edge on the 18-point lattice.
- Common column splits: 8 + 4 (main content and a side panel), 6 + 6 (comparison), 4 + 4 + 4 (three groups), 3 × 4 (four groups). Use the 5-up grid only when there are exactly five parts.
- Pick the frame on purpose: no rail for most pages, a left or right panel for a supporting photo or summary, a split frame for a tall diagram or image, the nav rail for long sectioned decks.

### Use components, not boxes

Reference design-system components in local templates as `wmds/component/<type>`, using the scene type names listed in [source format](source-format.md#local-templates). Never draw a box and fill it with text when a component exists:

| Content | Component |
| --- | --- |
| Bullets with bolded lead-ins | `bullets` with `{lead, text}` items |
| Ordered points with titles | `strongnum`, `ol` |
| Parallel ideas | `cardrow`, `card` (choose one treatment per row) |
| Steps or stages | `stepper`, `vstepper`, `chevron`, `phases`, `phasehead` |
| Time | `gantt`, `timeaxis`, `road` |
| Comparison | `table`, `matrix`, `beforeafter` |
| Who does what | `swimlane`, `governance`, `orgchart`, `pod`, `role`, `person` |
| System structure | `layerrow`, `container`, `node`, `connector`, `cylinder`, `device` |
| Numbers | `metric`, `callout`, `chart`, `gauge` |
| Loops and maturity | `cycle`, `maturity`, `funnel`, `pyramid`, `venn` |
| A quote | `pullquote` |
| Labeled text section | `textblock`, `colhead`, `grouplabel` |

If `project check` rejects a component, use the closest one it accepts and note the gap in the composition log. For working examples see `examples/local-composition` under the release `root`; inspect `source/components/v0/components.json` under `design_system_default` (from `pptxgengo paths`) and the [design documentation board](design-system-documentation.md) for component fields and geometry rules.

### Design rules

- **One focal point.** One thing the eye lands on first: the title, a key number, the diagram.
- **One mark per slide** at most (highlight, underscore, circle, spark or hand-drawn arrow).
- **No text dumps.** No single text element holds more than about three sentences or 60 words. Break longer content into labeled groups, bolded-lead bullets or a table.
- **Alignment.** Shared edges line up; groups of equal weight have equal widths; spacing between groups is larger than spacing within them.
- **White space is structure.** Leave clear space between groups; do not fill every region.
- **Consistency across the deck.** Use the same component for the same kind of content on every page, so the reader learns the pattern once.
- **Brand.** White and Grounded Blue dominate; colors have assigned meanings (see [slide composition](slide-composition.md#west-monroe-visual-rules-that-affect-composition)).

## Sketch, build, critique

1. **Sketch.** In the composition log, write the wireframe before building: the frame, then each zone with its column span, component and content. For example:
   ```text
   frame: none rail, compact footer, one-line title
   cols 1–8:  stepper, 4 steps (label + one-line text each)
   cols 9–12: callout "9 days → [target]" with basis; 2 bolded-lead bullets below
   footer:    source line
   ```
2. **Build** it as a local template (see [source format](source-format.md)), scaffolded from the closest shared template where one exists.
3. **Render and critique** at presentation size:
   - **Squint test.** Blur your eyes: is there one clear focal point and a visible structure?
   - **Scan test.** Read only the title and the bold text: do you get the page's message?
   - **Neighbor test.** Put it next to two shared templates: does it look like the same family?
   - **Text-dump test.** Is any text element longer than about three sentences?
   - **Balance.** Are there awkward empty regions or crowded corners?
4. **Revise** until all five pass. Record what you changed.

## Record every decision

Every slide, template or custom, gets an entry in `composition-log.yaml` before it is built (format in [project templates](project-templates.md#composition-logyaml)). For a custom slide, the entry must show:

- the templates considered and why each was rejected;
- which steps of the ladder above you tried;
- the content inventory, story units and density budget;
- the wireframe;
- what you changed after the render critique.

The operator should be able to read the log and understand every choice without asking.
