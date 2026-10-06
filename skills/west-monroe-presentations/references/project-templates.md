# Project file templates

Copy these shapes when you create the project's supporting files (see [project and resumption](project-and-resume.md) for where each file lives). Create each file when its stage needs it, and fill only what you know. Give every substantive field a status: **confirmed**, **working**, **inferred** or **open**.

## `project.md`: the running project file

```text
Project / deck name:
Current stage:
Next useful move:
Operator's instructions and constraints:

Opportunity
  Client's stated need or ambition:
  Operator's understanding:
  Our working interpretation and proposed response:
  Trigger / timing:
  Source IDs and unresolved conflicts:

Deck frame
  Core argument and current support:
  Reading mode: presented / emailed / hybrid
  Communication job: introduce a perspective / invite discovery / explain findings / compare options / propose work / report status
  What readers should understand, decide or do next:
  Priority audience and other material readers:
  Necessary detail and expected evidence:
  Main-deck length and expected reading time:
  Appendix policy / required RFP coverage:
  Frame version and operator feedback:

Value and differentiation
  Outcomes and metrics (baseline / target / illustrative):
  Internal win logic: keep in win-strategy.md (context.win_strategy), never here
  Client-facing reasons to choose West Monroe, with proof:
  Unresolved value or proof questions:
  Links to seller-skill work:

Style decisions
  Slide styles Ri chose from variations, and why:
  Voice corrections from Ri:

Production state
  Outline version and approval (version, date, decision, limits):
  Slide-content version and review status:
  Composition plan and material fit decisions:
  Deck path / build version and QA status:
  Agreed independent reviews and audience lenses:

Decision log / open questions
  Date; question or decision; status; reason; input needed from; affected page IDs; revisit when:
```

Record voice corrections and chosen slide styles here so later sessions apply them.

## `sources/index.md`: one entry per source

```text
Source ID / title:
Path or URL:
Type and origin: RFP / transcript / research / prior deck / seller output / ...
Author or participants; date; version:
What it contains:
Why it matters to this deck:
Useful pages, timestamps, headings or slide numbers:
Claims or pages it supports:
Authority, limitations or conflicts:
Internal-only or client-safe:
Underlying source, if this is a summary:
Last checked; changed / missing / superseded:
```

An index entry explains why to read a source, not just where it is. Example:

| ID | What it is and why it matters | Useful locations | Authority and use |
| --- | --- | --- | --- |
| S01 | Client RFP: scope, evaluation criteria, submission rules | pp. 4–7 scope; p. 11 evaluation | Client requirements; check for addenda before submission |
| S02 | Discovery call transcript: the sponsor's account of operational pain beyond the RFP wording | 12:40–20:10; 34:05–38:00 | Client discussion; verify quotes; separate stated needs from our interpretation |
| S03 | Internal pursuit call: possible win theme and competitor concern | 08:15–14:30 | Internal strategy; only supported, client-appropriate points reach the deck |

## `audience-context.md`

This file describes the readers' starting position. It is given to independent reviewers, so it must not contain our win strategy, our working history, or the conclusion we want them to reach.

```text
Audience context version:
Reading situation: emailed / presented / hybrid; likely reading time
Primary audience: role, seniority, responsibilities, decision rights
Other material readers: roles and influence
Expertise: business, technical, industry; vocabulary they use
What they already know, and how:
What we cannot assume they know:
What they care about: outcomes, constraints, risks, questions
Evidence they are likely to expect:
Likely concerns or objections:
History with us that affects tone (prior misses, open issues, relationship strength):
Confirmed vs. inferred assumptions:
What they would know about why they received this deck:
```

## `outline.md`

```text
Narrative: a short audience-facing statement of the case
Main-flow length / intended reading time:
For each page, in order:
  Page ID
  Title (full sentence)
  Intended point
  Planned support: evidence, mechanism, visual
  Connection to the previous and next page
Appendix pages: role and where the main flow points to them
Approved version:
```

## Visible slide content (for content review)

```text
Page ID and title:
On-slide copy, exactly as planned:
Planned chart / table / diagram: values, labels, relationships, captions
Visible source line and qualifications:
```

## `briefs/<page-id>.md`

```text
Page job and narrative role:
Sources and claim IDs:
Required content:
Open issues and private rationale:
```

## `claims.md`

The CLI checks every slide's `evidence_refs` against this file when `deck.yaml` links it as `context.claims`. Give each claim its own heading whose text is exactly the claim ID, or add `{#id}` to the heading. Use no other `##` headings in this file: any `## Word` heading becomes a claim ID.

```markdown
# Claims

## C03

- **Claim:** PMs spend about two days writing each PRD.
- **Type:** fact (client-reported)
- **Sources:** S02 discovery interview, 12:40–14:10
- **Basis:** VP of Product's estimate, September 2026; not measured
- **Visible qualification:** "about", attributed to the client
- **Validated by:** operator
```

IDs: letters, digits, `.`, `_`, `-`, starting with a letter or digit. A YAML registry (`schema: pptxgengo.claims.v1` with `claims: [{id, text, source}]`) also works but accepts only those three fields.

## `composition-log.yaml`

Required. Write an entry for every slide, including hidden ones, before building it. `project check` and `build` fail when a slide has no entry, an entry names a slide that no longer exists, or `chosen_template` doesn't match the slide. Keep it at the project root as `composition-log.yaml` (or link another path as `context.composition_log`).

```yaml
schema: pptxgengo.composition-log.v1
slides:
  phases:
    purpose: Show how the five-week assessment runs and who owns each step
    relationship: sequence
    candidates: [lifecycle/three-phases, phases/four, cards/4]
    chosen_template: phases/four
    rationale: |
      Inventory: four stages, each with an owner and one output; ~110 words.
      Story unit: the stages and their owners must stay together.
      Search: "assessment phases" (scenario); --structures sequence --items 4 (shape).
      lifecycle/three-phases: rejected, three phases; the work has four.
      cards/4: rejected, parallel cards hide the sequence.
      phases/four: selected, four stages with gates; used as is.
      Copy changes: none. Variations shown: phases/four vs. stepper composition; Ri chose phases/four for scanability.
      Render check 2026-10-04: squint, scan, neighbor, text-dump and balance pass.
    unresolved: []
```

- Fields: `purpose`, `chosen_template` (the template ID without scope, for example `cards/3` or a local template ID) and `rationale` are required; `relationship`, `candidates` and `unresolved` are optional. No other fields are accepted.
- Put the full decision record in `rationale`: the content inventory and story units, the searches run, every candidate with why it was rejected or chosen, the decision (as is, variation, edited copy, local derivative, new composition) and what changed, the ladder steps tried and wireframe for custom slides, variations shown and the operator's choice, and the render check.
- A rejection reason should be about the content's structure or the argument. "Text didn't fit" is acceptable only after a roomier variant, a frame change, tighter copy and a split were tried and recorded.
- When a page's design changes, rewrite its entry to describe the current decision and add the earlier one to `rationale` as history.
- Update the log in the same step as any `slide add`, `remove`, `swap` or template change.

## `reviews/` log entry

| Field | Record |
| --- | --- |
| Stage and packet version | What was reviewed, audience lens, scope, page IDs |
| Finding | Observed wording or visual and the reader's interpretation |
| Author assessment | Accept, partly accept, decline with reason, or ask the operator |
| Change | The specific copy, narrative, evidence or design repair |
| Operator input | Decision or approval, when needed |
| Resolution | Changed artifact and version, open issue, recheck needed |
