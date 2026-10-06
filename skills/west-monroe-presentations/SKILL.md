---
name: west-monroe-presentations
description: Create and revise West Monroe presentations (proposals, pursuits, perspectives, findings, status decks) from source materials, conversations or existing decks. Frames the argument with the operator, writes slide copy in the West Monroe brand voice and Ri's personal voice, runs independent reviews, and builds editable PowerPoint with pptxgengo's template and component library.
---

# West Monroe presentations

Turn the operator's opportunity into a clear, persuasive deck and build it as editable PowerPoint with `pptxgengo`. Work as a collaborator: bring observations and recommendations, say what supports them, and let the operator decide.

Start at the stage the request needs. Keep work that has already been approved. Read only the reference you need for the next decision.

## Voice: always on

Write every deck in the [West Monroe brand voice](references/brand-voice.md) and [Ri's slide voice](references/ri-slide-voice.md), without being asked. Read both before writing any visible copy.

In short: sentence-case titles that state the conclusion as a full sentence; personal, direct copy about the reader's situation; concrete mechanisms, owners and numbers; no buzzwords or marketing language; bullets with short bolded lead-ins; visual structure wherever the content has relationships.

## Layout: templates first, every decision logged

- Use the design library's templates and their variations (card treatment, frame, item count) first.
- Build a custom slide only when every template fails for a reason about the content's structure. Never go custom because copy didn't fit, and never build a title with one large text box.
- Before building any slide, write its `composition-log.yaml` entry: templates considered, why each was rejected or chosen, and any variation, fork or custom design with its rationale.

## Where to start

| The request | Read |
| --- | --- |
| Start or resume a deck project | [Project and resumption](references/project-and-resume.md); on resume, run `pptxgengo design project status` before drafting |
| Develop a deck from materials or a conversation | [Intake and framing](references/intake-and-framing.md) |
| Write or revise the narrative, outline or slide copy | [Narrative and copy](references/narrative-and-copy.md), plus both voice references |
| Choose how a page should look, build variations, review renders | [Slide composition](references/slide-composition.md) |
| Find templates, match page content to them, compare alternatives | [Template selection](references/template-selection.md) |
| Find or register photos, icons, graphics and logos | [Assets](references/assets.md) |
| Build a page no template fits, or fork a template | [Custom slide design](references/custom-slide-design.md) |
| Change, add, move, hide, remove or re-layout individual slides; manage sections; rebuild an existing PowerPoint | [Editing slides](references/editing-slides.md) |
| Look up `deck.yaml`, slide file or local template fields | [Source format](references/source-format.md) |
| Fit wordier copy, use Comfortable / Compact / Dense, or understand an automatic fitting warning | [Typography density](references/typography-density.md) |
| Upgrade an existing YAML deck to the current library or CLI | [CLI migration](references/cli-reference.md#move-a-project-to-a-new-cli-version) |
| Track slide drafting status, owners, due dates or internal review notes | [Draft Review Notes](references/draft-review-notes.md) |
| Run an outline, content, deck or source review | [Review packets](references/review-packets.md) |
| Create project files (project, sources, audience, claims, composition log) | [Project templates](references/project-templates.md) |
| Look up a command, flag or render through PowerPoint | [CLI reference](references/cli-reference.md) |
| Browse design-system foundations, components, frames and template patterns | [Design-system documentation](references/design-system-documentation.md) |
| Build one slide outside a project | [Design-system one-slide routes](references/design-system-authoring.md) |

## Production sequence

1. **Intake.** Inventory and index the sources. Interview the operator only about gaps that change a decision, and keep working while you wait for answers.
2. **Frame.** Propose the opportunity, core argument, audience, reading mode, communication job, intended outcome, depth and evidence gaps. Record the operator's corrections.
3. **Outline.** Write the narrative and an outline with each page's full-sentence title, support and transition. Read the titles alone; they should tell the story. Run an independent outline review, revise, and get the operator's approval before writing full copy, unless they have already told you to proceed.
4. **Content.** Write the exact visible copy for each page in both voices. Run the self-check in [narrative and copy](references/narrative-and-copy.md), then an independent content clarity review. Keep internal briefs out of the reviewer packet.
5. **Composition.** Choose the visual form from each page's content relationship, visual first. Search the template library and work down the ladder: a template as is, a template variation, tighter copy, a fork of the closest template, and only then a custom composition designed with [custom slide design](references/custom-slide-design.md). Aim for roughly 80% template reuse on substantive pages, never at the cost of the argument. Log every slide's decision in `composition-log.yaml`. When a page's style is uncertain, build two to four variations for Ri to choose from. If a template would force a material content change, show the operator the tradeoff before applying it.
6. **Build and QA.** Run `project check`, `project build` and `project measure --report BUILD/layout-report.json`, then `design render --png --contact-sheet` and look at every page at full size. Review any automatic density warning ([typography density](references/typography-density.md)); a successful step-down still needs a legibility check. Record what you saw with `project attach-render`. If render fails, run `design render-doctor` and ask the operator to clear what it reports (usually a one-time PowerPoint "Grant File Access" prompt); don't retry repeatedly. Verify claims against the sources. Make every fix in the slide files ([editing slides](references/editing-slides.md)), never in the generated PowerPoint.
7. **Review and deliver.** Walk the operator through the deck and the decisions that shaped it. Run the agreed finished-deck review. Package the deck for its audience, keep the maintainer package, and state which reviews are still pending.

Approvals record real decisions; do not ask again for something already approved. Routine edits proceed under the existing instruction. Bring back changes to the argument, commitments or required evidence.

## Rules that always apply

- **Never invent** facts, client names, metrics, quotations, anecdotes or commitments. Use a visible placeholder and add an open question instead.
- Keep facts, inferences, hypotheses and illustrative examples distinct, and mark illustrative content on the page. Keep units, bases, caveats, sources, owners and dependencies with the claims they qualify.
- **No slide is built without a composition-log entry.** Custom slides must use design-system components on the grid; no text dumps.
- A catalog match, successful binding, Go fit report or reviewed specimen does not prove that new copy fits. Inspect the candidate's contract and render the result.
- `design` builds lay out text in Go with bundled IBM Plex font metrics. Native PowerPoint rendering and visual review are separate steps.
- Shared library definitions are pinned within a project build. Make deliberate design changes as local derived templates; never edit the shared library to fix one deck.
- Use the slide's density presets for type changes; body roles move together and the header stays Comfortable unless explicitly changed. Preserve source-owned density limits. Do not shrink individual cards independently or expand geometry to silence overflow.
- Generated builds are immutable baselines. If someone edits the generated PowerPoint by hand, keep that file and copy the changes into the slide files by hand; there is no automatic round trip.
- Use `pptxgengo design` for deck authoring and `pptxgengo docs` to browse the design-system reference board. See the [CLI reference](references/cli-reference.md) and `pptxgengo paths` for commands and packaged resources.
