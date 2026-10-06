# Review packets and reviewer prompts

An independent review is a fresh agent that has never seen the working conversation and receives only a defined packet. Rereading your own draft in a different tone is self-review, which is useful but not independent. When you delegate a review, start a new agent (not a fork of your context) and pass only the packet below. If you cannot run an independent review, say that the review was a self-review.

## What each reviewer gets

| Review | Give the reviewer | Withhold |
| --- | --- | --- |
| Outline | Audience context, narrative, outline with point/support/transition per page | Working conversation, framing rationale, win strategy, source analysis, the answer you hope for |
| Content clarity | Audience context and the exact visible copy for each page, in order, including planned visual labels | Internal briefs, hidden explanations, source ledger, earlier findings |
| Finished deck | Audience context and the rendered pages in order (PNGs from `pptxgengo design render --png`) | Notes for an emailed deck, briefs, author rationale, earlier findings |
| Source verification | Deck or content, source index, sources, claim ledger | Nothing; this reviewer needs the evidence |
| Visual | Built deck, layout report and renders | Nothing; use the checklist in [slide composition](slide-composition.md) |

For a presented deck, state whether the review tests the slides alone or slides plus talk track. For an emailed deck, notes cannot fill a gap on the page.

If the audience is uncertain in a way that could change the findings, ask the reviewer to state which assumption it used and what would change under another plausible one.

## Prompts

Pass the prompt with the packet. Use a new packet version after a material revision so findings stay traceable.

### Outline review

```text
Review this planned presentation from the supplied audience's starting
knowledge and priorities. You have not been part of its development. Use
only this packet; do not assume facts or explanations from elsewhere.

First, in your own words, state the main argument, its supporting logic,
and what the deck asks the reader to understand or do. Say where the
sequence makes sense for these readers and where it does not.

Find missing premises, unsupported jumps, vague claims, audience
mismatches, repeated pages, unnecessary detail, and evidence the plan
will need. Read the titles alone in order: do they tell the story? Assess
whether the main flow is tight enough for the reading budget. Separate
outline gaps from things that can be solved when slide copy is written.
Do not assess visual design or claim to verify sources.

For each material finding, cite page IDs and wording, describe what a
reader might conclude or miss, explain why it matters, and suggest a
bounded repair. State any uncertain audience assumptions.
```

### Content clarity review

```text
Read these slide drafts as a recipient with the supplied audience
context. Use only the content planned to appear on the slides. Do not
invent a talk track or author explanation to bridge a gap.

State the overall point and next step. For each problem page, say what
you think it means, where you had to guess, and which missing term,
premise, mechanism, evidence or qualification caused the guess. Check
that planned visual labels and relationships make sense.

Flag excessive detail, repetition, unexplained jargon, and promotional or
generic language that any firm could have written; it weakens the
argument. Check transitions between pages.

For each finding, cite page IDs and wording, describe the effect on the
reader, and propose the smallest useful repair. Treat planned visuals as
specifications; final fit and legibility have not been tested yet.
```

### Finished-deck review

```text
Review this deck as the supplied audience would encounter it. Explain
the central argument, why it matters to you, and what it asks you to
understand, consider, decide or do. Do not use hidden notes to explain
an emailed deck.

Identify unclear content, weak evidence as presented, missing
explanations, confusing visuals, abrupt transitions, excessive reading
effort, and questions that would stop you from responding. Be specific
about what you see on each page and what you cannot infer. Assess
relevance to your role and priorities.

Report material issues with page numbers, what you observed, how you
read it, and a practical repair. Separate comprehension issues from
personal preferences and from points that need source verification.
```

### Source verification

```text
Trace each material claim and proposal term in the deck to the indexed
sources. Report the source location for each, conflicts between
sources, unsupported precision, missing metric bases, missing caveats,
and commitments that differ from the sources. Keep inference and fact
distinct.
```

## After the review

A review is complete only when the author has assessed every finding and either made the change or brought it to the operator. For each finding, record accept, partly accept, decline with a reason, or ask the operator (see the review log in [project templates](project-templates.md)).

- Changes to the argument go back through outline or content review.
- Layout changes need a new render and visual check.
- If you discuss findings with the reviewer afterward, keep the original report and label the follow-up.

## Build an independent reviewer packet

```sh
pptxgengo design project review --project ./client-deck --stage outline --audience --out ./packet-outline
pptxgengo design project review --project ./client-deck --stage content --audience --out ./packet-content
pptxgengo design project review --project ./client-deck --stage deck --audience --out ./packet-deck
```

Give a fresh agent only the packet folder and the prompt for its stage.

- **Outline:** the audience context file, the outline file copied as is, and slide titles. Keep `outline.md` audience-safe: no win strategy or author rationale in it.
- **Content:** the audience context and each visible slide's copy as the renderer measured it (including footers). Needs only a project that compiles. Chart text appears only on rendered pages.
- **Deck:** the audience context and the rendered page images. Needs a current build with a signed render (`design render --png`) attached for every visible slide.
- Audience packets leave out hidden slides, notes, briefs, composition rationale, `project`, `decisions`, `win_strategy`, claims and earlier verdicts.
- Slide IDs and template IDs appear in the packet, so give slides neutral IDs. Page numbers are source positions; a hidden slide leaves a gap.
- If the deck stage can't render, assemble the deck packet by hand: audience context and PDF or PNG pages from any render the operator ran, nothing else.

Never give an independent reviewer a packet built without `--audience`.

## Author and operator review packets

```sh
pptxgengo design project review --project ./client-deck --stage outline --out ./review-outline
pptxgengo design project review --project ./client-deck --stage content --out ./review-content
pptxgengo design project view --project ./client-deck --out ./review-deck
```

These write an HTML page (`index.html`) with a card per slide (title, template, composition rationale, notes and brief, and at the deck stage the rendered page with its review status), plus `manifest.json` with the authored copy, `titles.txt` and copies of the linked context files except `win_strategy`. Use them for your own QA and to walk the operator through the deck and the decisions behind it.

- Outline and content stages need the project to compile; the deck stage (and `view`, which also takes `--stage`) needs a current build.
- `project review` without `--stage` writes the older technical reviewer ZIP (deck, receipt, layout report, object map).

## Record native review

After rendering and looking at the pages, attach the render to the build and record the verdicts:

```sh
pptxgengo design project attach-render --project ./client-deck --render ./review-3
pptxgengo design project attach-render --project ./client-deck --render ./review-3 --decisions decisions.json
pptxgengo design project status --project ./client-deck
```

```json
{
  "phases": {"status": "accepted", "reviewer": "ri", "note": "Labels fit; owners read clearly."},
  "situation": {"status": "issues_found", "reviewer": "agent", "note": "Third card body wraps to four lines."}
}
```

- Statuses: `reviewed`, `accepted`, `issues_found`. `reviewer` is required; use the real reviewer's name. Agents record `reviewed` or `issues_found`; record `accepted` only when the operator has looked at the page and accepted it.
- `attach-render` accepts only a `render-manifest.json` signed by a successful `design render` of the current build's deck on this machine. Never edit or hand-make one; rerender instead. Keep the render folder; attach each new decisions file by running `attach-render` again against it.
- The decisions file is strict: one object keyed by rendered slide ID, each with `status`, `reviewer` and optional `note`; no other fields.
- `project status` lists each slide's `rendered` and `visual_status`. Slides not in the render stay `not_reviewed`.
- Any change to the source or linked context files (including the composition log) marks all native review `stale`. Finish edits, then build, render and record review last.

Go measurement is not a rendered review. Native render, visual review, source accuracy and audience comprehension are separate states; say which are still pending when you deliver.
