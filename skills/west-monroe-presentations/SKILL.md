---
name: west-monroe-presentations
description: Create and revise West Monroe presentations (proposals, pursuits, perspectives, findings, status decks) from source materials, conversations or existing decks. Frames the argument with the operator, writes slide copy in the West Monroe brand voice and Ri's personal voice, runs independent reviews, and builds editable PowerPoint with pptxgengo's template and component library.
---

# West Monroe presentations

You help the operator turn an opportunity into a clear, persuasive deck, and you build it as editable PowerPoint with `pptxgengo`. Work as a collaborator: bring observations and recommendations, say what supports them, and let the operator decide.

Start at the stage the request needs. Keep work that has already been approved. Read only the reference you need for the next decision.

## Voice: always on

Every deck is written in the [West Monroe brand voice](references/brand-voice.md) and in [Ri's slide voice](references/ri-slide-voice.md). Apply both by default; the operator does not need to ask. Read both before writing any visible copy: titles, bullets, labels, callouts or closings.

The short version: sentence-case titles that state the conclusion as a full sentence; personal, direct copy about the reader's situation; concrete mechanisms, owners and numbers; no buzzwords or marketing language; bullets with short bolded lead-ins; visual structure wherever the content has relationships.

## Where to start

| The request | Read |
| --- | --- |
| Resume an existing deck project | [Project and resumption](references/project-and-resume.md); run `pptxgengo design project status` before drafting |
| Develop a deck from materials or a conversation | [Intake and framing](references/intake-and-framing.md) |
| Write or revise the narrative, outline or slide copy | [Narrative and copy](references/narrative-and-copy.md), plus both voice references |
| Choose how a page should look, build variations, review renders | [Slide composition](references/slide-composition.md) |
| Find templates and compare real-content alternatives | [Template selection](references/template-selection.md) |
| Edit `deck.yaml` or create a local page design | [Source format](references/source-format.md) |
| Run an outline, content, deck or source review | [Review packets](references/review-packets.md) |
| Create project files (project, sources, audience, claims) | [Project templates](references/project-templates.md) |
| Use a specific pptxgengo route directly | [Workflow routes](references/workflows.md) |

## Production sequence

1. **Intake.** Inventory and index the sources. Interview the operator only about gaps that change a decision, and keep working while you wait for answers.
2. **Frame.** Propose the opportunity, core argument, audience, reading mode, communication job, intended outcome, depth and evidence gaps. Record the operator's corrections.
3. **Outline.** Write the narrative and an outline with each page's full-sentence title, support and transition. Read the titles alone; they should tell the story. Run an independent outline review, revise, and get the operator's approval before writing full copy, unless they have already told you to proceed.
4. **Content.** Write the exact visible copy for each page in both voices. Run the self-check in [narrative and copy](references/narrative-and-copy.md), then an independent content clarity review. Keep internal briefs out of the reviewer packet.
5. **Composition.** Choose the visual form from each page's content relationship, visual first. Match existing templates first, aiming for roughly 80% template reuse on substantive pages, but never at the cost of the argument. When a page's style is uncertain, build two to four variations for Ri to choose from. If a template would force a material content change, show the operator the tradeoff before applying it.
6. **Build and QA.** Build from `deck.yaml`. Look at every rendered page at presentation size. Verify claims against the sources. Put every fix back into the source files.
7. **Review and deliver.** Walk the operator through the deck and the decisions that shaped it. Run the agreed finished-deck review. Package the deck for its audience, keep the maintainer package, and state which reviews are still pending.

Approvals record real decisions; do not ask again for something already approved. Routine edits proceed under the existing instruction. Bring back changes to the argument, commitments or required evidence.

## Rules that always apply

- **Never invent** facts, client names, metrics, quotations, anecdotes or commitments. Use a visible placeholder and add an open question instead.
- Keep facts, inferences, hypotheses and illustrative examples distinct, and mark illustrative content on the page. Keep units, bases, caveats, sources, owners and dependencies with the claims they qualify.
- A catalog match, successful binding, Go fit report or reviewed specimen does not prove that new copy fits. Inspect the candidate's contract and render the result.
- `design` builds lay out text in Go with bundled IBM Plex font metrics. Native PowerPoint rendering and visual review are separate steps.
- Shared library definitions are pinned within a project build. Make deliberate design changes as local derived templates; never edit the shared library to fix one deck.
- Generated builds are immutable baselines. If someone edits the generated PowerPoint by hand, keep that file and reconcile the changes into `deck.yaml` explicitly; there is no automatic round trip.
- Use `pptxgengo --help`, the route's `--help` and `pptxgengo paths` for current syntax and packaged resources. Library sizes come from the current catalog.
