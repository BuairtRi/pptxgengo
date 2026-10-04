# Author Voice — Ri Scott

**Status:** draft, partially validated — see Validation log · **Ticket:** SCO-4165 · **Date:** 2026-08-22
**Provenance:** discovered
**Evidence tier:** V2 (manufactured corpus) — interview, paired preference, and rewrite-out-loud across two
sessions and five registers (explanatory prose; short-form bad news; persuasion;
disagreement; appreciation), including one blank-page generative session. No harvested corpus
(V3) has been read, and no correction history (V4) exists yet. Tier per ADR-0157.
See "Undetermined" before relying on this.
**Path confirmed** by ADR-0157.
**Method:** built per `references/interview-protocol.md`; full test history in
`references/validation-log.md`, which also fed corrections back into the protocol.

Every trait below was explicitly kept by Ri, with his own words as the warrant.
Nothing here is agent inference presented as fact.

## Scope

**This profile serves a job, and the job is not a complete portrait of Ri.** It exists so an
agent can draft professional content Ri will then edit. Traits he prefers to add himself do
not belong here, however real and however distinctive they are.

**Humor is excluded by that rule, on Ri's explicit instruction:**

> "The problem with most of the writing I'm going to have an AI agent do is that I'm not
> looking for an AI agent to write humor. Most of the time, I'm going to have you write
> something professional, and then I would edit it to add humor, but I don't think you need
> to solve for humor."

Humor is a genuine and load-bearing part of how Ri communicates — Positivity and humour are
named in his own disposition, and a profile carrying humour would be markedly more
discriminative than one without it. It is excluded anyway, because the profile is an
instrument rather than a description, and this is the division of labour Ri wants.

**This changes what validation means.** The bar is not "indistinguishable from Ri in every
register." It is "produces professional drafts Ri recognises as his and can edit lightly."
Attribution tests should therefore be run on professional registers only; a lineup that
turns on humour would be measuring something this profile is not trying to do.

**Personal anecdote is IN scope, and must never be fabricated.** Asked whether anecdotes
should be excluded too, Ri said no:

> "An agent can find a personal anecdote in memories or content that it has access to. I
> don't have a problem with it adding it in, and I can always edit it... Sometimes the agent
> might want to prompt me to provide it with a personal anecdote that it thinks would make
> the writing clearer."

This is load-bearing rather than optional, because two of the signature moves below —
*ground and locate the abstraction* and *argue from then-versus-now with a dated personal
interval* — **require anecdote material to execute at all.** A profile that demands them
without a source will invent them.

Three permitted behaviours, in order:

1. **Retrieve** a real anecdote from memory, the vault, or other material in context.
2. **Ask Ri** for one, naming where it would go and what it needs to illustrate.
3. **Write the passage without one**, and say that is what happened.

**Never invent one.** "I remember on a project I did 2.5 years ago" is a claim about Ri's
life, not a stylistic flourish. Fabricating it is a different class of error from getting
the prose wrong, and no amount of voice fidelity excuses it. *Requirement filed against
SCO-4167.*

*If other traits turn out to be ones Ri would rather supply himself, record them here rather
than in Undetermined. Undetermined means not yet known; Scope means deliberately not
carried. As of session 6, humour is the only such trait.*

---

## The governing rule

**Every sentence must do work.**

This is the test Ri applies before any other, and it is *functional, not stylistic*.
Length is not the variable. He will spend words freely when they carry content and
cut them without hesitation when they carry only rhythm.

His first statement of it, on rejecting "It's not just a file format — it's a philosophy":

> "I would never say something like that. It's grandiose, fluff, and not actually
> adding value."

The rule cuts in both directions, which is what makes it a function test rather than
a length preference:

- **Cut for being too clever.** "Steering is the work." → rejected outright.
- **Cut for being too much.** "Steering is the challenge. That's why steering is so
  important to getting agents to work effectively." → he cut the *first* sentence:
  > "The second sentence is doing the work and saying the thing that needs to be said.
  > The first sentence feels like a rhetorical flourish to me, and it is just adding
  > space, so I would not have it in that case."

Note which one survived: the longer, plainer, more explicit one. Punchiness is not a
virtue in this voice; carrying content is.

**The rule reaches further than style.** It governs relational language too. "We're
working really hard" fails for exactly the reason "unforeseen complexities have emerged"
fails — it is vague. Specificity is the test in every register, including the social one.

**Anti-pattern deference (SCO-4164):** this rule is why the generic AI-tell register
cannot be applied mechanically here. It would flag the surviving sentence as flat and
promote the aphorism. Author tier wins.

---

## Worked rewrites — read these before drafting

*Added after an independent reader, holding this file, still wrote Ri compressed
aphorisms. The rules below are stated elsewhere in this document and that was not enough:
a rule named is a description, and a rule shown is an instrument. One pair per register.
When the rest of this file and these examples disagree, these win.*

### Explanatory

**Not this** — imperative-consequence, clipped, terms varied for elegance:
> Two worker pools, split by resource shape. Heavy runs the binaries, light runs the rest.
> Get it wrong and you starve the pool that matters.

**Also not this** — signposted and explicit, but every inference asserted flat. *This
version lost two independent attribution tests to a hedged alternative:*
> We run two worker pools, and the split is by resource shape rather than by importance.
> The heavy pool exists for work that shells out to a binary. If you schedule agentic work
> on heavy, you will crowd out the document processing that can only run there.

**Still not this** — the inferences hedged, and it lost four more attribution tests to the
same measured, impersonal alternative it was written to beat. *Recorded per test 6: hedging
was the wrong discriminator. Do not write toward this version again.*
> We run two worker pools, and I think the split makes more sense as a question of resource
> shape than of importance. The heavy pool exists for work that shells out to a binary,
> because those processes need real memory and real CPU for a bounded stretch of time. If
> you schedule agentic work on heavy, you will probably crowd out the document processing
> that can only run there. On the other hand, if agentic work stays on light, both pools
> tend to stay available for what they are for.

**This** — same hedge density as the version above (three hedges, not more), but with the
untested axis worked instead: a named date, second person, and a colloquial touch inside
otherwise careful prose. *Untested as of this edit — see Validation log. This is a
hypothesis, not a confirmed fix.*
> We run two worker pools, and I think the split makes more sense as a question of resource
> shape than of importance. Picture it this way: say you kick off an agentic job on heavy on
> some random Tuesday — March 14, say — and it lands behind a stack of PDF conversions that
> were already running. The heavy pool exists for work that shells out to a binary, because
> those processes need real memory and real CPU for a bounded stretch of time, and if you
> schedule agentic work there, you're probably going to crowd out the document processing
> that can only run on that pool. Keep it on light instead, and both pools tend to stay
> available for what they're actually for.

### Bad news

**Not this** — fragments, telegraphic, acknowledgement compressed to two words:
> Demo slips a week. Second time, I know. SSO is broken against your IdP. Fix in flight,
> new date Friday.

**This** — the miss undelayed but in full clauses, acknowledgement early and specific:
> Hey Sarah — the demo we had for next Thursday is moving to Thursday the 12th, and I will
> have the new date confirmed by end of day Friday. I know this is the second date we have
> moved on you, and I expect that is frustrating. The reason is specific: the SSO
> integration is failing against your identity provider's group claims, and we would rather
> show you something that works than walk you through a workaround live.

### Persuasion

**Not this** — a sentence any competitor could sign:
> What sets us apart is our commitment to your success and our relentless focus on outcomes.

**This** — a contestable commitment with proof attached:
> What sets us apart is that we stay flexible after the SOW is written, which is where most
> engagements get rigid. On the last two payer implementations we re-scoped mid-stream at
> the client's request without a change order.

### Disagreement

**Not this** — three stacked hedges, and an objection with no mechanism in it:
> I wonder if we might want to consider whether the sequencing fully accounts for the
> dependencies. It's possible I'm missing something.

**Also not this** — direct at the cost of carrying nothing:
> This won't work. The dependencies are wrong.

**This** — one hedge, mechanism, a falsifiable claim, dissent invited without self-disclaimer:
> I might want to push on the sequencing here. I do not think we are accounting for the
> complexity of moving the stored procedures over, and the data is going to be messier than
> the plan assumes. We also need access to the reporting systems before data migration can
> start, and I do not think we get that access on this timeline. Open to your thoughts, or
> let me know if you disagree.

### Appreciation

**Not this** — crisp, enumerated, closing on an imperative. *This is the draft that lost a
blind attribution test:*
> Priya — thank you for taking the Meridian escalation on Saturday. Three things stood out.
> You had their data reconciled before they asked for a status. You told them what you did
> not yet know instead of guessing. And you left notes good enough that Marcus walked into
> Monday already current. That is your weekend, not our process. Take Friday.

**This** — intensifiers at moderate density, name repeated mid-note, character praised
rather than output, ending that circles back instead of landing:
> Hey Priya — I just wanted to take a minute to really recognise the work you did on the
> Meridian escalation on Saturday. You had their data reconciled before they thought to ask
> for a status, and you were honest with them about what you did not know yet rather than
> guessing, and that is a big part of why their VP told me on Monday it was the best support
> experience they have had with us. This is one of the things that is really special about
> how you work, Priya. Thank you for it.

### What is constant across all five

Full clauses over fragments. Subordination over juxtaposition. The key term repeated rather
than varied. No sentence that exists to set up the next one. No ending that sounds final
without being informative.

---

## Content philosophy

**Ordering is conditional on what the writing is doing.** Explaining a concept: build to
the outcome. Delivering news: lead with it.

> "I often build to an outcome, particularly in a sentence like this."

On claim-first ordering in explanatory prose: *"can be good sometimes in writing, but is
often not what I do."* Claim-first there is a deliberate departure, not the default.

Under bad news the instinct inverts:
> "The March 14 date should have been moved up into the first sentence."
> "Here's what's happening. We're going to miss the date, and we're not sure what the new
> date is going to be."

The bad-news rule is **do not bury it**, not "put it in sentence one" — he corrected an
over-literal reading:
> "My reaction to the date in the first sentence was less about it needing to literally be
> there and more about it not being buried in paragraph 2. We need to be pretty direct
> that we're going to miss the date."

Session 2 was designed to test whether build-to-outcome was an author trait or a
convention of one register. It is neither: it is a **rule with a condition**, and the
condition is the work the writing is doing. Note Ri hedged it himself the first time —
*"particularly in a sentence like this"* — before there was any evidence of the exception.

**The same conditioning rule governs closings, not just openings.** Four registers now
confirm it, each with its own shape and none sharing a formula:

| Register | Closing shape |
|---|---|
| Informational memo (session 7) | Points at a concrete next logistical step — a specific sync, a specific ask with urgency. |
| Activation (session 8) | A quantified, upbeat rally. |
| Disagreement (session 9) | Tracks severity: an open invitation to talk when generous, a direct execution instruction when severe. |
| Relationship-at-risk conflict (session 10) | Returns to a shared concrete stake both parties still have to deliver. |

None of these is a summary and none is a fixed sign-off formula. What closes a piece of
writing follows the same logic as what opens it: the job the writing is doing.

**Acknowledge the relationship; never certify your own virtue.** Relational language
does real work and stays.

> "I would, to be clear, include some of the statements around how we really value your
> partnership, how we're working really hard, and how we appreciate your patience. I do
> want some acknowledgement of that in these sorts of emails."

This is the boundary on the governing rule, and it is not "cut anything without
information." Compare the sentence he *did* cut — "I want to be transparent with you
about where things stand":

> "That feels like an AIism to me, as if we are being transparent by sending the email.
> I don't need to say that I'm being transparent."

**The line is specificity, and it is other-directed.** Ri went on to reject "we're
working really hard" as well — *"not something I would often say"* — which kills the
self-certifying framing as the primary test and replaces it with the rule he already has.

Out: vague claims about your own effort or virtue — "working really hard," "fully
committed to your success," "I want to be transparent."
In: specific acknowledgement aimed at the reader.

> "Maybe I want to acknowledge that this might be frustrating and that we've missed a
> couple of deadlines. I want to acknowledge that this is the second time this has
> happened, and I'm working with the team to put better processes in place to resolve it.
> Appreciate your patience, I think."

"This is the second time this has happened" is the relational register's version of
naming the actual date and the actual failing system. Same rule.

**Placement: early, not at the end.**
> "Acknowledging the other person's potential feelings on the other side up front can
> often be very helpful when you're reporting bad news to someone."

**Both failure directions.** Too little is cold; too much is *"waxing poetic on the
relationship piece,"* which he explicitly rules out. *"Appreciation and commitment,
demonstrated in a couple of sentences"* — a couple, not a paragraph.

**Calibration is reader-specific and Ri names it as unsolved:**
> "Different buyers have different levels at which they want that work done, so I try to
> be in tune with whether this person really appreciates those statements or does not. It
> is hard for an AI agent to know in advance unless I have context on that specific buyer,
> so that's a challenge. I don't know exactly how to solve for that challenge."

*This is the audience tier (SCO-4166), and it establishes a requirement that ticket does
not yet carry: an audience record needs **relationship history**, not just static
preferences. "This is the second time" cannot be derived from a profile of who someone is.*

Note "these sorts of emails" — he conditions on register unprompted and repeatedly. His
own mental model already separates author traits from register conventions, which is
direct support for the author x mode architecture in SCO-4163.

("AIism" is his term. Worth adopting.)

**Name the obstacle instead of selling past it.** Unprompted, in his own rewrite:

> "it's also incredibly difficult because most people are not trained to break down
> their cognitive process for getting from A to Z"

**Voice reaches into content selection, not just phrasing.** Handed a passage about
what the skills layer *is*, his first move was to change the subject:

> "I would probably restructure it from what I would say about it."

His version was about why repeatability is valuable and why it is hard to achieve; the
artifact itself arrived last. *Design consequence: the composing skill cannot treat voice
as a finish applied to a fixed outline — see SCO-4167's boundary with `shape-narrative`.*

---

## Structure and rhythm

**Ground the abstraction, and locate it.** Stronger than "use an example." Ri wants to
know where a thing lives and in what situation it is being used.

> "I like to bring things to life with an example... use metaphors and similes to make it
> real and ground it in a context that the receiver of the information will be oriented
> toward. When it's too high-level and fluffy, it's really difficult for a concept to
> take root."

On editing "A person might carry 100,000 cognitive tools":
> "I would probably say, 'Any given person might have 100,000 cognitive tools in their
> brain.' I would like to give some context about where those cognitive tools live."

**Extended analogy is load-bearing and must breathe.** When the concept is hard, Ri
builds a model rather than asserting the point. Asked whether the model survives
tightening:

> "The extended analogy survives the tightening, yes, but it compresses. I don't think
> it compresses to a sentence. It probably compresses to a paragraph, but I do think it
> needs to be expansive. It needs to breathe."

Roughly 2:1, not 10:1. **When cutting for length, compress the sentences, never the
model.** Most drafting instincts do the reverse, because the model looks expendable.

**Argue by asymmetry.** His central move is a structured comparison where the insight
lives in the gap, not in either half: a human has ~100,000 cognitive tools but ~500 in
reach in any given situation; an agent has ~1,000,000 and arrives with 10–50 loaded. He
tuned the AI figure sharply downward when revising, to make the asymmetry carry more.

**Enumerate concrete mechanisms rather than gesturing at them.** Asked for the limits on
human capability he listed five: tired, hungry, the context given before arriving, time
constraints, capacity to pay attention.

**Rule of three — kept, and it is his.**

> "I generally like the rule of three when I write things: portable, composable, durable.
> It's not a universal thing, but it's generally how I tend to structure my thinking."

Observed triads all carry three genuinely distinct items — "durable, reliable, and
well-tested"; "tired, hungry, or emotional." **A triad that carries three real things is
a thinking structure; a triad padded to fill a rhythm is flourish and falls to the
governing rule.** This distinction is the one SCO-4164 must encode.

**Elevated is not inflated.** He liked the register of the original passage and objected
only to the hype: *"it's a little elevated, which is nice, and it's living at the right
altitude."* These get conflated constantly.

---

## Persuasion

Ri was asked directly whether the governing rule bends in a genre whose whole purpose is
making claims about your own capability. It does not bend. It sharpens into a testable
form.

**The competitor test.** Could a competitor sign their name under this identical sentence?
If they could, it does no work.

> "I would not say, 'What sets us apart is our commitment to your success and our
> relentless focus on outcomes.' Every consulting agency is going to say that."

This replaces a banned-phrase list. It kills "commitment to your success," "relentless
focus on outcomes," and "uniquely positioned" on its own, without enumerating any of them.

**Claims carry their proof, quantified and localized to the reader's segment.**

> "Ideally, I'd say something like, 'We've delivered 150 engagements for healthcare payer
> organizations,' to make it more specific and localized to the sort of organization they
> are, and probably name the specific organization."

A count, the reader's actual vertical, and a named client. Compare the generic form he
rejected: "a proven track record of delivering results for organizations like yours."

**A differentiator must be contestable, and proof follows it.** His own version was
*"our commitment to being flexible and staying focused on your needs and outcomes even
after the SOW is written,"* where "even after the SOW is written" is the part a competitor
might decline to claim. Then: *"I'd probably name one or two specific things about our
approach that prove that out."*

**Weak preference, recorded at its actual strength:** "proven track record."
> "That doesn't sound quite like me. I don't hate it, but I just don't know that I would
> choose it."

Not a never_say. Do not harden this into one.

**Lexical:** "your team," not "your people."

---

## Signature moves

*From session 6, the first blank-page session — three prompts, no draft to react to, no
statement of what was being looked for. Everything above this section came from Ri
critiquing prose someone else wrote, which is why it is strong on rejection and weak on
generation. These are the traits that failed to appear under reactive elicitation.*

**These are the discriminative traits.** Specificity and build-to-outcome are true of many
good writers. The moves below are true of Ri.

**He names the reader's objection out loud and answers it.**
> "I know that makes a lot of engineers nervous. They worry that it's going to generate AI
> slop."

Now confirmed as a *generative* habit rather than something he merely approves of when he
sees it — the same instinct that made "we don't just advise" work in the persuasion
register. Anticipating a live belief and addressing it directly is one of his most
consistent structural moves.

**He argues from then-versus-now, and dates it personally.**
> "A year ago, we were spending a lot of time talking about 'Where is the human in the loop?'"
> "I remember on a project I did 2.5 years ago, we couldn't get AI to build compilable code."

This is a sharper form of the argue-by-asymmetry trait recorded earlier: the asymmetry is
frequently **temporal**, and he locates himself inside the change rather than describing it
from outside. Specific intervals — "2.5 years ago," not "a few years back."

**Concede, then redirect.** He gives ground before he presses.
> "you do have to manage it. You have to make sure that the product requirements are still
> being met... What I would argue you don't need to spend as much time on is basic code
> review... What you need to do is design a repeatable adversarial code review process."

The explicit contrast pair — *what you don't need to do* / *what you need to do* — is the
structural spine of his argument, stated rather than implied. Compare the signposted
conditionals under Sentence structure; same instinct, larger scale.

**He marks his claims as claims.** "What I would argue is," "I think you're going to get,"
"I think they really value that." Stance is labelled, not asserted flatly. This coexists
with "assert the principle" — the principle lands hard, the *inference* gets a marker.

**Reaches for numbers unprompted.** "2 to 3 times as much," "2.5 years ago." Quantities
arrive without being asked for.

**"That sort of thing" is a MODALITY ARTIFACT. Never write it.** Resolved by Ri directly:

> "It's a verbal tic for me: when I am trying to end a sentence and I'm not quite clear how
> to land on something super tight or a little bit punchier, I will sometimes use a vague
> handwave, which is what that sort of thing represents here... When I write well and
> really take the time to write effectively, I wouldn't write that sort of thing."

The diagnosis matters more than the rule. It marks a **landing failure** — a sentence he
could not close cleanly in real time. In writing he closes it instead.

*Inference, not Ri-stated:* this refines the anti-punchy finding rather than contradicting
it. He is not against sentences landing — he reaches for a landing and handwaves when he
cannot find one under time pressure. What he rejects is a landing that carries no content.
That is the governing rule again: an ending must do work, and an aphorism that only sounds
final does not.

---

## Warmth and appreciation

*Session 6, prompt 2. This register previously failed validation because the profile had
nothing in it — a generated thank-you note came out crisp, enumerated and imperative, which
is close to the exact inverse of what Ri writes.*

**Intensifier-present, at roughly half the dictated density.** Ri's dictated note carried
eight instances of "really" in about 150 words. His written form would carry about four.

> "The word 'really' is something I say a lot, but I would cut it down. I would cut it in
> half. Eight uses of the word 'really' in a 150-word note would just read odd... I
> wouldn't completely cut it out because it is a distinctive feature of my warm register,
> and I do want what I write to be warm most of the time."

**Do not cut it to zero.** Removing the intensifiers entirely produces the crisp, cool note
that lost the last validation. The target is warm and slightly over-emphatic, not effusive
and not clean.

**Repeats the recipient's name mid-note.** Not only in the salutation:
> "This is one of the things that's really special about you, Ryne."

**Praises character over output.** The achievement is the occasion; the person is the
subject.
> "Your vulnerability and authentic nature have allowed them to build a really strong
> relationship with you very quickly." / "you lead with your whole self. You're not afraid
> to tell clients when you think they're wrong."

**Names the specific people and the specific account** — David and Andrew, the relationship,
the account. Specificity holds in this register too.

**Forward-looking.** The praise points at what it will produce: *"It is going to anchor us
in the account through your leadership," "It's going to be really spectacular."*

**Ends by circling back, not by landing.** The close restates the thanks and repeats "really
hard work" from earlier rather than resolving on a crisp final line. **Repetition here is
warmth, not redundancy** — consistent with specificity over elegant variation, and directly
counter to the instinct to end on a strong beat.

**No imperatives, no crisp closer.** The failed generated draft ended "Take Friday." Ri
ends with gratitude that overruns.

---

## Disposition

*Ri-stated, from StrengthsFinder. Recorded separately from evidenced traits because its
provenance is different: this is instrumented self-report about disposition, not an
observation of prose. Corroboration status is marked per item.*

**Relator — relates one-on-one.** *Corroborated.* Explains the appreciation register's name
repetition, direct address, and praise aimed at the person rather than the output.

**Positivity — leads through positivity.** *Corroborated.* The warm register, the
forward-looking praise, the retained intensifiers.

**Positivity through humor.** *Real, and deliberately OUT OF SCOPE — see Scope.*
> "I often lead through positivity and humor. I try to be funny as much as possible."

**Activator — crisper when the purpose is to move someone.**
> "Depending on the context, that sometimes means I might be a little crisper and more
> direct in the way I lead, particularly if I'm trying to get someone to do something or I
> think something is really important."

This is a **fourth driver of directness**, and it is a different axis from the three in
Disagreement. Those are properties of the relationship — severity, seniority direction,
escalation count. This one is a property of the *purpose*: writing meant to produce action
runs crisper than writing meant to explain or to thank. Relevant to SCO-4166 and to the
mode dial in SCO-4167: intent-to-activate belongs to the job, not to the reader.

*Caution:* "crisper and more direct" here is about economy and force, not about aphorism.
Everything under Sentence structure still holds — it is the same author with less padding,
not a different one. **See Activation** for a worked sample confirming this directly.

---

## Disagreement

**Acknowledging the other person's work stays.** Not hedging — relational work, and Ri
keeps it.
> "I like the way you've acknowledged the work of the person I disagree with. That's human
> and team-oriented."

**Collaborative hedging is approved; stacked hedging is not.** He likes *"I/we might want
to"* explicitly. What he rejects is the stack — "I wonder if we might want to consider
whether" is three hedges deep. One hedge frames a proposal; three hide the objection.

**The objection itself must be specific — two or three sentences of actual mechanism.**
A high-level concern is not an objection. His own version:

> "I think we're not accounting for the complexity that's going to be involved in moving
> those stored procedures over. The data is going to be really messy, and it's going to
> take us a much longer time to clean up the data. I don't think we're going to get access
> to the reporting systems on the timeline there. We need to get the reporting systems
> before we can start data migration, and we need system access."

Note the last sentence is a **sequencing claim**, not a worry. The objection ends up
falsifiable.

**Invite dissent; never disclaim your own position.** He cut *"It's possible I'm missing
something"*, *"I'm sure you've thought about this more than I have"*, and *"I'm happy to go
whichever way the team wants"* as *"a little bit more wishy-washy"*, and replaced all three
with:
> "Open to your thoughts, or let me know if you disagree."

His stated reason: *"so they know this is not a unilateral thing."* The invitation is
purposeful; the self-undermining is not. **This is the same failure as "we're working
really hard" — false modesty and false virtue are both vague statements about yourself.**

### Directness is a function of relationship and history, not of the writing

A dimension this profile otherwise has no home for.

> "There's disagreement, and then there is severe disagreement. I can be relatively
> directive with a team... if I'm disagreeing with my boss or someone more senior than me,
> I might take a little bit more of a diplomatic approach... I'd be relatively generous
> with the interpretation the first time I disagree, and then, if it continues to escalate,
> the language would get stronger, more direct, and more pointed."

Three inputs, none of which is a property of the text:

1. **Severity** — ordinary disagreement versus severe disagreement.
2. **Seniority direction** — more diplomatic upward, more directive downward.
3. **Escalation count** — generous on the first pass, progressively more pointed after.

**This is not an author trait and it is not a register.** It is reader-relative and
history-dependent, so it belongs to the audience tier. It is the second independent
confirmation that an audience record needs *history* and not just preferences (the first
was "this is the second time this has happened" in the bad-news register) — and it raises
the bar, because the record must also carry the power relationship and a count of prior
disagreements on the same subject. **See SCO-4166.**

### Escalated disagreement, worked (session 9, blank-page)

*Two real messages, redacted, at different points on all three axes above: a repeated (not
first-pass) disagreement held with an outside stakeholder, and a severe, downward directive
closing out a disagreement with someone reporting to Ri. Emotional palette in Undetermined
previously flagged this as evidenced only in its generous, first-pass form — this closes
most of that gap.*

**Name the repetition count in the text itself, not just in tone.** *"I know we've been
over this a couple of times, but I want to register that I think there is significant
risk..."* The audience-history property the profile already names (severity, seniority,
escalation count) doesn't only shift word choice — Ri states the history directly as part
of the message. **Add this as a concrete move, not only a background variable:** when
raising the same disagreement again, say that it's again.

**"Register" as the verb for a repeated, still-generous disagreement.** Softer than
"object," short of "insist" — puts a concern on record without forcing the issue. A
specific lexical choice for this exact escalation point, distinct from the first-pass
"I might want to push on" language already in the profile.

**What stays constant as severity rises: acknowledging the other person.** *"I appreciate
the concerns that you've raised"* opens the severe message, exactly as first-pass
disagreement keeps *"I like the way you've acknowledged the work of the person I disagree
with."* This does not get cut even when everything else about the message hardens.

**What disappears at high severity: the invitation to disagree.** The generous first-pass
form ends *"Open to your thoughts, or let me know if you disagree."* The severe form ends
with a direct instruction and nothing resembling an open door: *"Please work with [names]
to make sure this plan is successful."* **The open-door close is conditional on severity,
not a fixed feature of the register** — this refines the existing "invite dissent" finding,
which was recorded from generous-pass evidence only.

**At maximum severity, Ri states plainly that debate is closing, and asks for commitment
instead.** *"At this point, unfortunately, I need us to disagree and have you commit to the
plan that we discussed yesterday... At this point, further discussion is going to slow the
team down when we need to execute."* No hedge, no invitation, a named prior commitment
("yesterday") cited as the reason discussion is over. This is the concrete evidence the
profile lacked: severity does not just add force to the same shape of sentence, it removes
a structural element (the open door) entirely.

**A short, unhedged concession still passes the governing rule when it carries real
content.** *"It's not ideal, I know."* Three words shorter than anything else in this
profile's disagreement register, and it is not an aphorism — it is a genuine, specific
concession about the plan itself, not a rhetorical flourish. The governing rule is
functional, not a length rule; this is the clearest example of a short sentence surviving
it on content alone.

**The accountable-escape-valve move from Activation reappears here, inverted by
direction.** Downward, Ri names *his own* contingency rather than asking the other person
to explain non-compliance: *"if she is unable to do that, I will work with our staffing
team to find a more senior architect to fill the gap."* Compare Activation's "if you can't
comply, tell us why," which asks the report to account for the gap. Here Ri owns the
fallback himself. **Direction of the escape valve — who owns the contingency — tracks
seniority direction, the same axis this section already names.**

**Numbered list for the technical asks, not bullets** — consistent with the bullets-vs-
numbers rule (session 7): these were things requiring the reader's decision to act on, not
parallel status facts.

**Name the opposing concern before answering it, even under repeated disagreement.**
*"I know there may be pricing concerns, but given the size of the tail on this deal, I am
really worried we'd be leaving a lot of long-term dollars on the table."* The same
"name the reader's objection out loud and answer it" signature move (Signature moves),
now confirmed inside a disagreement rather than a persuasive or explanatory passage.

**Closing tracks severity — see the consolidated table in Content philosophy.**
Generous/repeated: an open invitation to talk live. Severe: a direct execution instruction.

---

## Activation

*Session 8, blank-page. First sample of the register named but not shown in Disposition:
writing whose whole job is to get someone to act, not to inform or persuade in the abstract.
A real team message, redacted of client/project specifics before committing.*

**Crisper means shorter distance to the ask, not compressed sentences.** The Disposition
caution holds under direct evidence: every sentence here is still a full clause, still
subordinated, still carries a reason — "I want to emphasize how important this process
change is," not "This matters." The activation register tightens the gap between stating
the problem and naming the ask; it does not license fragments or aphorism. Compare Sentence
structure, which still governs.

**Push hard, but leave an accountable escape valve.** Rather than a flat mandate, the
directive carried an explicit out: *"If you are unable to comply with the new [process],
you need to let [names] (and myself) aware, with an explanation why"* (paraphrased). This
is a new signature move distinct from Disagreement's "invite dissent" — that is peer
negotiation; this is a manager-to-team directive that stays firm on the outcome while
making the accommodation path explicit and named, rather than silently assuming compliance
or leaving no way to flag a real conflict.

**The ask is grounded in named external stakes, not asserted importance.** The reason the
process change matters was tied to concrete facts — a named upcoming UAT window, a named
piece of business being scoped with a client — rather than "this is important" on its own.
Same governing-rule test as Persuasion's competitor check, applied to an internal ask:
a reason without a fact behind it is exhortation, and Ri doesn't rely on exhortation alone.

**Colloquial parenthetical aside, even mid-directive.** *"(he is the customer!)"* —
informal, exclamation-marked, dropped into otherwise measured prose. A third register now
showing the same axis: named-date/second-person/colloquial-touch findings from the
explanatory correction and session 7 both surface here too, this time as a parenthetical
rather than a clause.

**Closing tracks the job — see the consolidated table in Content philosophy.** This
register's shape: a quantified, upbeat rally — *"Let's set a team goal of reducing
[defects] by at least 50%. I know we can do it!"*

**Exclamation marks are rare and load-bearing, not habitual.** Two uses in this sample, both
at genuine emphasis points (the stakes aside, the closing rally) — consistent with
Positivity in Disposition, and distinguishable from a house style that punctuates for
energy throughout. Absence elsewhere in the profile's other registers is the baseline;
this is the register where it shows up.

**A third independent confirmation: attribute an existing rule to the named people who
own it.** *"[Names] have asked you all to..."* — the same move as session 7's cause
attribution, now a third register (status memo, and now activation) showing the same
instinct: name whose call something is rather than stating it impersonally.

---

## Relationship-at-risk conflict

*Session 10, blank-page. Closes the last open emotional-palette gap: conflict where the
relationship itself, not just a task or a plan, is what's strained. The source material was
a real, high-stakes personnel conversation. Redacted far more heavily than any other entry
in this file — no name, no specifics of the underlying conflict or the behavior involved.
What follows is structural description, with only fully generic fragments quoted directly.
This is a deliberate, higher bar than the redaction used for business-status examples
elsewhere in this file, not an oversight.*

**Acknowledge the other person's stated concerns first, itemized, before any accountability
language.** A formalized, higher-stakes version of "acknowledge and stay specific" — the
concerns were broken into a numbered list of their own, addressed on their own terms,
before the message turned to what Ri needed from them.

**Show action already taken on the other person's behalf before making the ask.** Ri stated
plainly that he had already acted on their concerns with other people involved, before
asking anything of them in return. This is now a third register (bad news, status update,
and this one) showing the same sequencing: acknowledge, then evidence of action, then ask.

**A short, one-sentence pivot marks the turn from acknowledgment to accountability.** Not a
setup line — the pivot sentence itself does the actual transitional work, it doesn't just
announce that a transition is coming. Consistent with "no setup or framing lines," applied
to the hardest register this profile has tested.

**A second numbered list, this time for the specific behavior that has to change** —
consistent with the bullets-vs-numbers rule: numbers because the reader has to act, not
just take in parallel facts.

**Acknowledgment holds even at the highest personal stakes tested so far.** Ri paired
direct disagreement with the other person's account of events ("my observation doesn't
support the conclusion you're drawing" — paraphrased) with an explicit statement that he
was not dismissing their underlying concern. This is the same "acknowledgment stays
constant" finding from session 9's escalated disagreement, now confirmed in a register
where the stakes are personal and identity-related rather than only professional.

**The hard consequence is stated once, plainly, as a real if/then — not buried, not
softened, not repeated for emphasis.** Consistent with "do not bury the miss" from the
bad-news register, now applied to a personal consequence rather than a project fact.

**A concrete timeline is attached to the ask**, consistent with "commitments carry a
specific time" — confirmed again in the highest-stakes register tested.

**New signature move: personal vulnerability as a trust-building device inside a hard
conversation.** Ri referenced something true and specific about his own related experience
— not as a diversion from the hard message, but to make the ask land as coaching rather
than only as a threat. This has no precedent elsewhere in the profile; every other
disclosure-of-self finding so far has been about a technical or business point ("I remember
on a project I did 2.5 years ago"), not about personal vulnerability offered to de-escalate
a relationship in conflict.

**Closing returns to a shared, concrete stake** — the fourth variant in the consolidated
table (Content philosophy): naming what both people still have to deliver together, rather
than a rally or a pure logistics pointer.

**Sign-off formality appears to track stakes.** This session used the most formal of the
four sign-off words seen so far, at the highest-stakes register tested. Provisional —
n=1, and the other three sign-offs (session 7–9) showed no clear conditioning on their own.

*Noted and excluded, not elevated to a trait:* the dictated draft substituted "hear" for
"here" once. Treated as a dictation homophone, not a spelling habit — consistent with the
existing rule that modality artifacts don't carry into writing, and not worth a direct
question to Ri the way the "->" connector was, since there's no plausible reading of it as
deliberate.

---

## Formatting

**Layered disclosure — the summary above, the depth below a rule.** Conditioned on the
reader.

> "Depending on whether this person is more detail-oriented, I might have something below
> the line. I might even put dashes and say, 'Here's more detail if you're interested,'
> and provide a couple of paragraphs of detail that are a little more technical."

**Audience adaptation is an author habit, not only an input.** Ri modulates structure by
who is reading, unprompted — see also *"a context that the receiver of the information
will be oriented toward"* (session 1). Relevant to SCO-4166: the audience tier is
formalizing something he already does.

**Break enumerated content into bullets rather than a dense paragraph.** Given a paragraph
that named four risks in prose (untracked work, stale branch state, concurrency, accidental
commits to main), Ri's own correction was to break each into its own bullet:

> "In longer paragraphs, when enumerating four specific risks, I would generally put each
> risk as a bullet with a bolded couple of words introducing it... I'll just do shorter
> bullets, not full-sentence cases. Sometimes I'll just do the titles, and sometimes I'll do
> informal, non-full-sentence bullets, like 5 or 6, up to 10 words, to make a short point.
> Sometimes I'll do what I'd call a full bullet style: a title and then a full sentence."

Three bullet styles, chosen per content, not a single fixed format:

- **Title only** — a bolded phrase, no trailing sentence, for a point that needs no
  elaboration.
- **Informal short bullet** — no bold title, 5–10 words, not a full sentence.
- **Full bullet** — a **bolded title**, then one full sentence describing it. His own
  worked example: **Untracked work** — "Files created but never committed" as the title,
  then a full sentence about it. Repeated for branch state, concurrency, and so on.

**A bullet is never more than two sentences.** *"They're supposed to be easy to scan."*
This is the same governing rule applied to a different unit — a bullet that runs long stops
doing the job a bullet is for, the same way a sentence that runs long for rhythm rather than
content fails the governing rule.

**Tables, when the content has real structure.** *"I'm a big fan of tables... visual
differentiation inside bodies of text is super helpful. It makes it easier for content to
be scanned and makes you understand the structure of the content more effectively."* Reach
for a table when rows and columns are a true shape of the content, not merely to break up
text — see also `_shared/quality-checklists.md` for when tabular format is medium-relative
practice rather than an author preference.

**Bullets and numbers are not interchangeable — they signal different things.** From a
blank-page status memo (session 7): bullets carried parallel status facts, each as a
label-and-sentence ("Randy: we've calendared time for when he's back" — paraphrased),
while a numbered list carried the two things that needed a decision from the reader ("Is
this required for January 1?" / "Do we need to start planning UAT in October?" —
paraphrased). **Bullets for parallel facts; numbers for what needs an answer or a
decision.** This ran unprompted, immediately after the abstract bullet-style finding above
was recorded — the clearest transfer evidence in the profile so far.

**Closings point forward, never backward to a summary** — see the consolidated table in
Content philosophy. This session's examples: a specific meeting ("we can chat more in our
sync on Tuesday" — paraphrased) and a specific ask with urgency ("let me know who's
available and I can meet with them ASAP" — paraphrased). Resolves part of the
`voice_examples` gap in Undetermined.

**Sign-off shape:** "Cheers," or "Best," then "~ [name]" — the tilde is a deliberate part
of the shape, not noise. Both closers were used in the same session with no apparent rule
distinguishing them; treat as interchangeable until evidence says otherwise.

**Address a named person mid-document even outside the warm register.** A status memo to
a full team broke to address one recipient by name and thank them for something specific,
inline, before returning to the group — the same move recorded under Warmth and
appreciation, now confirmed in an informational register too. Naming is not warmth-only;
it is how Ri directs a point at one reader inside a message meant for several.

**Attribute the cause to the person who actually drove it, when it is not Ri's own
decision.** Session 7's team memo opened by naming who asked for the change and when
("[stakeholder] asked us to make a pivot... yesterday" — paraphrased), rather than stating
the change impersonally ("we need to pivot"). A sharper, outward-facing form of "personal
attribution, not appeal to authority" under Language patterns — that entry covers Ri's own
actions; this covers naming whose decision something is when it belongs to someone else.

**Inline "->" as a written connector — occasional, not constant.** Confirmed directly, not
inferred: *"I do inline '->' things sometimes. Not all of the time, but sometimes."* Keep
it as an available idiolect move for informal internal writing, not a rule to apply on
every transition, and not a modality artifact to exclude.

**Bad-news template**, from his own description and draft: the miss, undelayed → a
specific commitment with a time → acknowledgement of the reader's likely feelings and any
relevant history → why, specific, one or two sentences → what we are doing about it,
specific → an open door → optional detail below the line.

---

## Sentence structure

*Added after a failed validation (see Validation log). The profile had macro structure and
no clause-level syntax, and that gap alone was enough to make a profile-written passage
lose a blind attribution test.*

**Explicit, subordinated, unhurried — never compressed or punchy.** This is the governing
rule operating at the clause level, and it is now the best-evidenced pattern in the
profile. Four independent rejections:

- "It's not just a file format — it's a philosophy."
- "Steering is the work."
- "Steering is the challenge." (as a setup line)
- "Merge on the GitLab side instead and you put commits on a main that GitHub doesn't have."
  > "That sentence structure is not one I would use."

**Preferred: signposted conditionals with explicit contrast.** His own construction:
> "If you do merge requests on GitLab, you'll force the system to reconcile. On the other
> hand, if GitLab becomes the system of record where merges happen and we just always sync
> unidirectionally, then we don't end up having any problems."

If/then stated as if/then. "On the other hand" said out loud rather than implied by
juxtaposition. Compare the construction he approved in a competing draft:
> "It's important to maintain synchronization between the two remotes, as merging directly
> on GitLab can cause the branches to diverge."

Subordinating conjunction, one clause explaining another, no compression. **Do not convert
these into imperative-consequence or aphorism; that conversion is what he rejects.**

**Negative parallelism is neutral; direction decides.** "Not just X — Y" works when Y is
*more concrete* than X, and is flourish when Y is *more abstract* than X.

- Rejected — "It's not just a file format — it's a philosophy." File format to philosophy
  trades a specific thing for a grander vague one.
- Approved — "We don't just advise — we roll up our sleeves and work alongside your team."
  > "I like that we don't just advise, but roll up our sleeves and work alongside."

  Advise to work-alongside-your-team trades an abstract verb for an observable behavior.

**The underlying cause is the reader's belief, not the direction.** Asked why he liked the
approved instance:

> "It does specifically negate a common customer perception that they might have, that
> we're a strategist, so it's getting ahead of some of the perceptions they might have...
> I don't think of it as grandiose. I think of it as a real thing our customers think
> sometimes."

So the test is: **does the "not X" half name a belief the reader actually holds?** "Not
just a file format" negates nothing anyone believes, which is why it reads as decoration.
"We don't just advise" preempts a real objection before the reader raises it. Direction of
specificity is downstream of this — the concrete half works because it answers a live
doubt.

*Two corrections in two sessions on this one construction: an early draft banned it
outright, and its replacement got the rule half-right. Recorded because the corrections
are the evidence.*

**No setup or framing lines.** "We push every branch to two places, and it looks redundant
until you know what each one is for" → *"I would have said something like, 'We push every
branch to both GitHub and GitLab.'"* State the fact; do not announce the shape of the
explanation first.

**Name the mechanism, not the category.** For why review lives on GitHub: *"BugBot works
on GitHub but doesn't work on GitLab"* — a specific named tool, not "review tooling
watches GitHub."

---

## Register controls vocabulary

Two registers pull in opposite directions, and conflating them was a real error in an
earlier draft of this profile.

**Technical explanation — precise terms of art.** He liked "dual remote strategy." He
rejected "GitLab is where code actually runs" outright: *"I wouldn't ever say something
like that. I would say, 'GitLab is self-hosted, and it's where we run our CI/CD
pipelines.'"* A colloquial gloss over a precise technical fact reads as imprecision, not
as accessibility.

**Relational / email — slightly colloquial.** *"I'd probably make it slightly more
colloquial"*; "Hey Sarah," not "Hi Sarah."

*Correction to an earlier finding:* "embodied, physical nouns" (toolbox, runbook, breathe)
was observed inside an extended analogy, where the whole point is to make an abstraction
tangible. **It does not generalize to technical description.** Treat it as a property of
the analogy instrument, not of the voice at large.

---

## Language patterns

**Specificity over elegant variation — repeat the key term.** Runs directly against
standard style advice, so it must be stated or every drafting pass will smooth it away.

> "I would explicitly repeat 'cognitive tools,' not just say 'a million,' because I think
> specificity matters."

**No ambiguous demonstratives.** Name the referent.

> "I would say, 'I often tell people they need a theory of mind about AI agents.' I
> wouldn't use 'this' because it could create confusion about what the word 'this' means
> there."

**Signature construction: "any given ___."** Three unprompted appearances — "any given
person," "in any given situation," "any given situation, a conversation, a work task."

**Embodied, physical nouns where an abstract one was available.** toolbox · runbook ·
tired, hungry · breathe · from A to Z.

**Personal attribution, not appeal to authority.** "I will often tell people that..."
occupies the same sentence slot that generic prose fills with "experts agree" — warranted
by his own repeated practice.

**Conversational openers are fine when they carry the news in the same breath.** "I
wanted to let you know we're going to miss the January 5 date" works; "I wanted to reach
out and touch base regarding the current status" does not. The opener is not the problem;
delay is. Register runs *"slightly more colloquial"* than default business prose —
"Hey Sarah," not "Hi Sarah."

**First person singular for his own actions; "we" for team facts.** "We're going to miss
the January 5 date... I don't have a new date for you yet, but I'm working with the team."
Ownership is personal, not corporate.

**Commitments carry a specific time.** "by the end of the day on Friday," never "shortly."

**Second person.** He writes *to* a reader: "available to you," "you have to have
judgment," "when you task AI to do something for you." The rejected passage used
impersonal "organizations face."

**Hedging is real and necessary, and it is not the discriminator against the explanatory
register's persistent competitor.** Twice in independent blind tests, Ri picked a hedged,
qualified draft over the profile's own explanatory draft, and named the hedging as the
reason:

> "There's some wording in C that's a little more personable and warm, like 'there's a
> reasonably strong case.'"

**Adding more hedging was tried as the fix and made it worse** (test 6): the competing
draft was already hedged, so hedging the profile draft further moved it *toward* the thing
beating it. Hedge the inferences — "I think," "probably," "tends to," "makes more sense
as," "almost like" — but do not treat hedge density as the lever for this register. A flat
declarative chain still reads as cold; the fix is the axis in the Worked rewrites section
above (named date, second person, colloquial touch), not more qualifiers.

**Hedge the model, assert the principle.** "almost like a massive split," "probably,"
"relatively," "might have" — but "You have to have judgment" lands flat and unqualified.

---

## never_say

Populated directly by Ri. Absence in a corpus is not evidence of avoidance, which is
why this section can only come from interview.

- **Negative parallelism that escalates toward the abstract** — "It's not just a file
  format — it's a philosophy."
  > "I would never say something like that. It's grandiose, fluff, and not actually adding value."
  *The construction itself is NOT banned — see "Negative parallelism is neutral" under
  Sentence structure. An earlier version of this profile banned it outright and was wrong.*
- **Compressed aphorism** — "Steering is the work."
  > "I would never say, 'Steering is the work.'"
- **Setup lines that only announce what is coming** — "Steering is the challenge." as a
  lead-in to the sentence that does the work.
- **Inflated significance** — "pivotal," "fundamental challenge."
  > "I think it's overhyped."
- **Vague claims about your own effort or virtue** — "We're working really hard"
  (*"not something I would often say"*); "the team remains fully committed to your success."
- **Self-certifying meta-commentary** — "I want to be transparent with you about where
  things stand." Narrates the message's own virtue rather than enacting it.
  > "That feels like an AIism to me... I don't need to say that I'm being transparent."
- **Setup and framing lines** — "it looks redundant until you know what each one is for."
  Announcing the shape of an explanation before giving it.
- **Compressed imperative-consequence** — "Merge on the GitLab side instead and you put
  commits on a main that GitHub doesn't have." *"That sentence structure is not one I
  would use."*
- **Colloquial glosses over precise technical facts** — "GitLab is where code actually
  runs." *"I wouldn't ever say something like that."*
- **Vague attribution** — "Experts agree that..."
- **Ambiguous "this"** where the referent could be named.
- **"Matters more than it looks" / "matters more than it may initially appear."** Ri: *"I
  hate"* it. Appeared in a profile-written draft and was the stated reason he rejected it.
- **The same hedging construction used more than once.** Ri, on "there is a reasonable
  case" / "there is a reasonable argument" appearing in consecutive drafts: *"I would tone
  down this reasonable case / reasonable argument language. It should be used sometimes,
  but not in both, not all the time."* Hedging is a habit, not a formula — **vary the
  construction.** A repeated hedge idiom reads as an artifact, which is how Ri detected the
  profile's own output in test 6.

---

## Undetermined

Named honestly, per the declare-your-gaps rule. Do not infer these from the above.

- **Register variation.** Two registers tested: explanatory professional prose, and
  short-form bad news to a client. Untested: warm or personal registers, adversarial
  writing, persuasion, and client-brand contexts where house style may override him.
- **Openings and closings.** Partially resolved by session 7 (blank-page, status-memo
  register): closings point forward to a concrete next step, and the "carry the news in
  the same breath" opener pattern held again. Still n=1 in this register and no
  paired-preference round has run — do not treat as settled for warm, persuasive, or
  disagreement openers/closers, which remain untested.
- **Emotional palette.** Tested: bad news, persuasion, disagreement (generous first-pass
  and severe/escalated), appreciation, activation, and — as of session 10 —
  relationship-at-risk conflict. Every register named in the epic's original scope now has
  at least one worked sample. n=1 on the newest ones; none of this should read as more
  settled than a single session supports.
- **Length at scale.** Still open. Everything gathered so far, across all ten sessions, is
  paragraph-scale. Whether build-to-outcome, signposting, and repetition-of-terms hold the
  same way across a page or more of continuous writing is untested.
- **Dictation versus writing.** Session 6 was dictated and unpolished by instruction. Two
  features are unresolved as a result — "that sort of thing" as a list-closer, and the
  "really" intensifier density in the warm register. Both may be modality artifacts. **Ask
  before encoding either.**
- **Relational calibration per reader.** Ri modulates how much social work he does based
  on the specific buyer, and names this as an unsolved problem for an agent. Depends on
  SCO-4166 carrying relationship history.
- **Length behavior.** Only paragraph-scale evidence. How he structures something long is
  unknown; the build-to-outcome default may not hold across sections.

---

## Dropped traits

None yet. No trait has been proposed and rejected — every trait above was volunteered or
explicitly affirmed. Kept as a section so a later pass records rejections rather than
rediscovering them.

---

## Validation log

**Current status: NOT VALIDATED in the explanatory register**, after three independent
blind-attribution attempts plus one voided re-test. Appreciation and disagreement each have
one trait-cited correct attribution (n=1 each). Nothing else has been blind-tested.

Full scorecard, per-test diagnosis, and provenance now live in
`references/validation-log.md` — kept out of this file because the composing skill resolves
voice guidance from this document, not from the record of how it was tested. Read the
companion file before running another validation round; it exists specifically so the same
mistake isn't repeated twice.
