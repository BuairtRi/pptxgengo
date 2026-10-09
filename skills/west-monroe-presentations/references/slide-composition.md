# Slide composition and visual review

Use this when choosing how a page should look and when reviewing the rendered deck. For finding templates, use [template selection](template-selection.md).

## What Ri wants a slide to look like

- **Dense is fine; hard to navigate is not.** A page can carry a lot of content when the reader can scan its structure at a glance: a clear title, visible groups, bolded lead-ins, a table or diagram that shows the shape of the content.
- **Visual first.** When the content has relationships, show them: a process flow, a diagram, a chart, a roadmap, a matrix. Use text-only layouts for content that really is a list of parallel points.
- **The title carries the message.** A sentence-case full-sentence title on every content page.
- **Bullets with bolded lead-ins** and short paragraphs inside the visual structure.

## How to compose a page

1. **Name the relationship** the page has to show, using the table below.
2. **Search the template library** for that shape and the page's scenario ([template selection](template-selection.md)). Templates are the default.
3. **Work down the ladder:** a template as is, then a template with a variation (card treatment, frame, count), then tighter copy, then a fork of the closest template, and only then a new composition ([custom slide design](custom-slide-design.md)).
4. **Log the decision** in `composition-log.yaml` before building: the candidates considered, why each was rejected or chosen, and any changes ([format](project-templates.md#composition-logyaml)). This is required for every slide.

Never build a custom slide only because copy didn't fit a template, and never build a custom slide as a title plus one large text box.

## Choose the form from the content relationship

| The content is… | Try first | Also consider | Library families to search |
| --- | --- | --- | --- |
| A sequence of steps or stages | Process flow with labeled handoffs | Numbered steps; chevron row | approach (phases, sequence), argument (flow, goal-steps), solution (process), lifecycle |
| Activities over time | Roadmap or timeline with milestones and decision gates | Phase table | approach (roadmap, plan, phase-gate), diagrams (road, timeline), software (roadmap), offers (gantt) |
| Options compared on criteria | Table or matrix | Comparison with scores; side-by-side columns | decisions (options, decision-request), argument (comparison, decision), commercials (pricing-options) |
| Before and after, or a gap | Two-column comparison with the gap called out | Bar chart of the two states | argument (from-to), value (value-bridge) |
| Quantities, trends, shares | Chart with a titled takeaway | Big-number callouts with basis | evidence (chart, stats, quadrant), value |
| Actors and who does what | Team structure, swimlane or RACI table | Role pods | team (roles, governance, team), approach (roles), lifecycle (lifecycle-roles) |
| Layers or components of a system | Architecture diagram | Layered table | architecture, solution (context, patterns), software (capability-map) |
| A cycle or feedback loop | Cycle diagram | Numbered loop with a return arrow | lifecycle (pdlc, sdlc, pdlc-sdlc), diagrams |
| Capability or readiness assessment | Capability map or heat map | Maturity scale | argument (capability-map, capability-table), heatmaps, maturity |
| Parallel ideas of equal weight | Bolded-lead bullets or cards | Icon row | core (cards, text), argument (pillars) |
| One big point with evidence | Statement title plus proof panel | Big number with supporting bullets | core (key-message, quote), evidence (stats) |
| An objection and its answer | Callout or two-part panel | Concede/redirect two-column | argument (problem, comparison), core (key-message) |
| What we heard / the client's situation | Situation or "what we heard" page | Interview readout | understanding, interviews |
| Status and progress | Status report with RAG and next steps | Decision log | status, decisions (decision-log) |

Use the smallest structure that makes the relationship clear. Keep distinctions between outputs and outcomes, owners and contributors, and dependencies and sequence. Do not put unequal ideas in equal boxes; do not invent a third card to fill a three-card layout.

## Build variations

When the right form is not obvious, or Ri has not yet seen this kind of page in the current deck, build **two to four variations** of the same page with the same approved copy. Make them meaningfully different:

- one diagram- or chart-led, one text-dense (bolded-lead bullets or a table);
- different templates for the same content shape;
- a template version and a fork or custom composition.

Present them side by side with a one-line note on each: what it emphasizes, what it sacrifices, and which one you recommend. Record Ri's choice and the reason in the composition log and `project.md`, and apply it as a preference to similar pages later in the deck. Do not build variations for every page once the deck's style is established.

## Density

- Put the main point and essential support in the main flow; optional depth goes in a referenced appendix.
- A dense proposal page keeps meaningful comparison, sequence, ownership or acceptance detail rather than flattening it into short generic bullets.
- Use a table only when rows and columns genuinely compare like with like.
- Leave enough space that the reader sees the title first, the structure second and the detail third.
- Use supported [whole-slide density presets](typography-density.md) when they
  preserve reading quality; review automatic fitting warnings. Do not shrink
  individual elements independently or delete material qualifications to fit.
  If the supported tiers still fail, change the form, split the page or edit copy.

## West Monroe visual rules that affect composition

- White and Grounded Blue (`#070154`) dominate. Accent colors have assigned jobs: Highlight Blue (`#0047FF`) for small emphasis, Magenta (`#F900D3`) sparingly for callouts, grays for quiet surfaces.
- The yellow marker highlight goes behind one to four words of a main headline, on covers and pivotal pages only, never on every page.
- IBM Plex Sans for titles and body; IBM Plex Mono for eyebrows, labels, numbers and footers.
- Square corners, flat surfaces, no shadows or gradients.
- Charts: label directly instead of using a detached legend where possible, and do not rely on color alone to carry meaning.
- Use verified registered originals available on this machine or operator-supplied images registered in the project. Match crop and placement to the page. Photography should show real people doing meaningful work.

## Accents

Accents are supporting emphasis attached to a specific phrase or component. Pick the phrase from the argument. Supported accent geometry covers specific artwork, modes and single-line phrases; when the tool reports an ambiguous or unsupported target, leave a clear area for manual placement and tell the operator.

## Visual review of the rendered deck

Render with `pptxgengo design render --png --contact-sheet` (use `--slides` for pages you changed) and look at every page at full size, in order, including covers, closings and repeated components. Use the contact sheet to compare variations and check consistency across pages. Record verdicts with `project attach-render` ([review packets](review-packets.md#record-native-review)). Check:

1. **Titles.** Each states the conclusion; read in order, they tell the story.
2. **Scanability.** The structure is visible at a glance; bolded lead-ins, groups and diagrams guide the eye.
3. **Fit.** No clipped, overflowing or cramped text; type is readable at presentation size.
4. **Relationships.** Diagram labels, arrows, alignment and grouping show the relationships the copy describes, and do not suggest relationships that are not true.
5. **Numbers.** Units, bases, dates and owners are consistent across pages; illustrative values are marked.
6. **Color and contrast.** Colors carry their assigned meaning; text is legible on its surface.
7. **Retained material.** Images, chart data and text inherited from a source template are relevant and true for this deck.
8. **Flow.** Each page moves the argument forward; transitions and decisions are explicit.
9. **Family resemblance.** Custom and forked pages look like they belong with the shared templates: components rather than plain boxes, no large text dumps, aligned to the grid.
10. **Log.** Every page has a current composition-log entry that matches what was built.

Report what remains unresolved plainly. A successful build, a Go fit report and native measurement are each different kinds of evidence; none of them alone is visual approval.

## Customize family topology

Choose the actual structure before placing it. For variable roles, pods, reporting
and governance, use [team composition](team-composition.md); for task intervals,
workstreams, phases and gates, use [Gantt composition](gantt-composition.md).
The family operators preserve the frame, preview fit and require explicit source
ownership changes. A catalog example does not define the new engagement's count
or relationships. These commands require the newer development CLI.
