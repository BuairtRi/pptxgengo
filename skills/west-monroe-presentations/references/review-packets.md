# Review packets and reviewer prompts

An independent review is a fresh agent that has never seen the working conversation and receives only a defined packet. Rereading your own draft in a different tone is self-review, which is useful but not independent. When you delegate a review, start a new agent (not a fork of your context) and pass only the packet below. If you cannot run an independent review, say that the review was a self-review.

## What each reviewer gets

| Review | Give the reviewer | Withhold |
| --- | --- | --- |
| Outline | Audience context, narrative, outline with point/support/transition per page | Working conversation, framing rationale, win strategy, source analysis, the answer you hope for |
| Content clarity | Audience context and the exact visible copy for each page, in order, including planned visual labels | Internal briefs, hidden explanations, source ledger, earlier findings |
| Finished deck | Audience context and the rendered pages, in order | Notes for an emailed deck, briefs, author rationale, earlier findings |
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

## `design project review`

`pptxgengo design project review --project PATH --out NEW-ZIP` produces a technical reviewer archive: current deck, receipt, layout report, object map and visible slide values. It is not a narrative or content packet. Build those from the project source as described above. Keep `audience-context.md` linked in the project `context` so changes to it invalidate the related review approvals.

Go measurement is not a rendered review. Native render, visual QA, source accuracy and audience comprehension are separate states; state which ones are still pending when you deliver.
