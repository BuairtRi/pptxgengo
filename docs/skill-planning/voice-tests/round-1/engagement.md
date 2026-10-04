# How we'll work together: slide copy (voice test)

Section of a proposal for a 16-week data platform build. Readers: your VP of Data (sponsor) and your procurement lead, reading independently before the finalist meeting.

Convention: anything in [brackets] is a placeholder or a proposed value the operator has not confirmed. Items marked *Illustrative* are examples, not commitments.

---

## 1. Narrative

Over 16 weeks, three workstreams will build claims data ingestion, the member 360 data model and the reporting migration at the same time, so most of the risk sits in how decisions, issues and changes move between them. Your product owner and data steward work inside all three workstreams next to our leads. Each decision has a named owner, and only changes to cost, dates or workstream boundaries go to your monthly steering committee. Your previous vendor's status reports stayed green until the project was two months late, so our report shows a forecast date beside every baseline date and changes color by rules your product owner approves in week one. After the SOW is signed, you can swap scope inside the budget and the 16 weeks without a change order; anything that adds cost or time goes to the steering committee with a written estimate first. By week 16, your team runs the platform without us, because they take over operating it while we are still on the project. We're asking you to name your product owner, data steward and [platform engineers] before week one so they start with us on day one.

---

## 2. Titles only

1. Your product owner and data steward work inside all three workstreams, next to our workstream leads
2. Issues surface in workstream reviews first, so nothing reaches your monthly steering committee as a surprise
3. Your product owner decides what gets built next; changes to cost, dates or workstream boundaries go to the steering committee
4. Our status report turns amber when a milestone is at risk, not after it has slipped
5. We stay flexible after the SOW is signed, so you can swap scope inside the 16 weeks without a change order
6. Your team will run the platform without us after week 16, because they take over operating it while we're still here

---

## 3. Page copy

### EM-01

- **Eyebrow:** Team and roles
- **Title:** Your product owner and data steward work inside all three workstreams, next to our workstream leads
- **Lead-in:** Each workstream pairs one West Monroe lead with [number] of your engineers, so your team learns the platform while it is being built.

**Diagram labels (top to bottom)**

- **Steering committee (monthly).** Your VP of Data, [other client members] and [West Monroe executive sponsor]. Decides changes to cost, dates and workstream boundaries.
- **Your product owner.** Sets priorities across all three workstreams and accepts finished work.
- **Your data steward.** Approves member and claims data definitions before they go into the member 360 model.
- **[West Monroe engagement lead].** Runs the plan, the status report and the hand-offs between workstreams.

**Workstream columns**

| Claims data ingestion | Member 360 data model | Reporting migration |
| --- | --- | --- |
| [West Monroe lead] with [your engineers] | [West Monroe lead] with [your engineers] | [West Monroe lead] with [your engineers] |
| Loads claims data from [source systems] into the platform | Builds one record per member from claims and [other sources] | Moves [number] reports from [current reporting tool] to the platform |

**Callout:** **What we need from you.** Names for the product owner, data steward and [platform engineers] before week one.

**Visual form:** Team structure diagram: steering committee at top, product owner, data steward and engagement lead as a band spanning three workstream columns.

---

### EM-02

- **Eyebrow:** Governance cadence
- **Title:** Issues surface in workstream reviews first, so nothing reaches your monthly steering committee as a surprise

**Ladder (three tiers, bottom to top)**

| Meeting | Who attends | What it decides | What moves up |
| --- | --- | --- | --- |
| **Workstream review** [weekly] | Workstream lead, your engineers, product owner as needed | Next week's work and design questions inside the workstream | Anything that affects another workstream or a milestone date |
| **Program review** [weekly] | Product owner, data steward, engagement lead, all three workstream leads | Priorities across workstreams, data definitions and hand-offs between workstreams | Anything that changes cost, the 16-week timeline or workstream boundaries |
| **Steering committee** monthly | Your VP of Data, [other members], [West Monroe executive sponsor] | Changes to cost, dates and boundaries; acceptance of [phase milestones] | |

**Arrow labels (upward):** "affects another workstream or a milestone date" / "changes cost, dates or boundaries"

**Bullets beside the ladder**

- **Cross-workstream hand-offs.** The member 360 model needs claims data from ingestion, so the program review checks that hand-off every week.
- **Steering pre-read.** Your VP of Data gets each decision with options and our recommendation [three business days] before the meeting.
- **No waiting a month.** If a decision can't wait for the next steering committee, the engagement lead calls [an ad hoc session] within [two business days].

**Visual form:** Three-tier escalation ladder with labeled upward arrows, bullets in a side panel.

---

### EM-03

- **Eyebrow:** Decision rights
- **Title:** Your product owner decides what gets built next; changes to cost, dates or workstream boundaries go to the steering committee

| Decision | Who decides | Who's consulted | Where it's made |
| --- | --- | --- | --- |
| Order of work within the 16 weeks | Your product owner | Workstream leads | Program review |
| Which reports migrate first | Your product owner | Reporting migration lead, [report owners] | Program review |
| Member and claims data definitions | Your data steward | Product owner, member 360 lead | Program review |
| Technical design inside a workstream | Workstream lead | Your engineers, [your architecture lead] | Workstream review |
| Scope swaps that fit the budget and 16 weeks | Your product owner | Engagement lead | Program review |
| Changes to cost, end date or workstream boundaries | Steering committee | Product owner, engagement lead | Steering committee |
| Acceptance of finished work | Your product owner | Data steward, for data definitions | [Milestone review] |

**Footnote:** If a decision isn't on this list, the product owner and engagement lead name its owner within [two business days].

**Visual form:** Decision-rights table, with the steering committee row shaded to show it is the only one that waits for the monthly meeting.

---

### EM-04

- **Eyebrow:** Status reporting
- **Title:** Our status report turns amber when a milestone is at risk, not after it has slipped
- **Lead-in:** Your previous vendor's status reports stayed green until the project was two months late. Our report shows the forecast date for every milestone, so you see a slip the week the forecast moves.

**Bullets**

- **Forecast beside baseline.** Every milestone shows the date in the plan and the date we forecast today.
- **Color rules you approve.** In week one, your product owner approves the rules: for example, amber when a forecast moves [N] days and red when it moves [N] days or misses a steering committee date.
- **One report for everyone.** Your product owner, your VP of Data and our team read the same [weekly] report; the steering committee doesn't get a separately polished version.
- **Every amber or red item answers four questions.** What caused it, what we're already doing, who owns it and the date we'll know whether it worked.

**Sample report row** (labeled *Illustrative*)

| Milestone | Baseline | Forecast | Status | Cause and action |
| --- | --- | --- | --- | --- |
| [First full claims load] | Week [N] | Week [N+1] | Amber | [Cause]. [Action] by [date]; owner [name]. |

**Callouts on the sample:** "Forecast moves first, so amber shows up before the baseline date passes" (pointing to Forecast) / "Rule approved in week one" (pointing to Status)

**Visual form:** Bolded-lead bullets on the left; an annotated, illustrative status-report excerpt on the right.

---

### EM-05

- **Eyebrow:** Scope changes
- **Title:** We stay flexible after the SOW is signed, so you can swap scope inside the 16 weeks without a change order
- **Lead-in:** In our experience, projects get rigid once the contract is signed, which is about when your team starts learning what it actually needs from the platform.

**Flow diagram**

1. Box: "Your team finds a change it needs"
2. Decision diamond: "Fits inside the current budget and 16 weeks?"
3. Yes branch, box: "Your product owner swaps it in at the program review. We re-plan that week. No change order."
4. No branch, box: "We write the cost and date impact within [N] business days."
5. Box (after No): "Steering committee decides before we start the work."

**Arrow labels:** "same size or smaller: swap" / "adds cost or time: estimate first"

**Objection callout:** **Flexible doesn't mean open-ended.** A swap has to fit the same budget and 16 weeks. Anything that adds cost or time goes to your steering committee with a written estimate before we start it.

**Proof:** **[Number] payer data engagements.** On [engagement], we [changed what, in which week] without a change order.

**Visual form:** Two-branch decision flow with labeled arrows, objection callout and proof line below.

---

### EM-06

- **Eyebrow:** Knowledge transfer
- **Title:** Your team will run the platform without us after week 16, because they take over operating it while we're still here

Form is uncertain; full copy for both variations is in section 4. Recommended: Variation B.

---

## 4. Variations for EM-06

Both variations use the EM-06 eyebrow and title above.

### Variation A: diagram-led (handoff timeline)

**Lead-in:** Your engineers write the runbooks as they take over each task, so the runbooks describe how your team actually works.

**Three-stage band across weeks 1–16**

| Weeks 1–[N] | Weeks [N]–[N] | Weeks [N]–16 |
| --- | --- | --- |
| **We build, your engineers pair** | **Your engineers build, we review** | **Your team runs it, we stand by** |
| Our engineers build the pipelines, model and reports; your engineers pair on every change. | Your engineers take the next [pipelines, model changes and reports]; our leads review and pair when they're stuck. | Your team runs the daily loads, fixes failures and ships changes; we step in only when they ask. |

**Gate labels between stages:** "moves on when [your platform lead] says the team is ready"

**End-state callout:** **After week 16.** Your team runs claims ingestion, the member 360 model and the migrated reports without us.

**Visual form:** Horizontal timeline band with three stages, the share of work shifting from our engineers to yours, gates between stages.

### Variation B: text-dense (readiness table)

**Lead-in:** Your engineers build with us from week one and run the platform themselves for the last [N] weeks, with us standing by.

| Workstream | What your team runs after week 16 | How they learn it | How we'll both know they're ready |
| --- | --- | --- | --- |
| **Claims data ingestion** | The claims load pipelines, their schedules and failure alerts | Pair on every pipeline we build; write the runbook as they take over | Your team runs [a full month of claims loads] and fixes failures without us |
| **Member 360 data model** | Changes to the model, with your data steward approving definitions | Pair on the model build; lead [the last N model changes] | Your team adds [a new member attribute] end to end without us |
| **Reporting migration** | The migrated reports and their refresh schedule | Migrate reports with us; then migrate [the last N reports] alone | Your team migrates [the last N reports] and we only review |

**Footnote:** Your product owner confirms each readiness test before week 16.

**Visual form:** Readiness table, one row per workstream, with the readiness column highlighted.

### Tradeoff and recommendation

- **Tradeoff:** Variation A shows when control moves to your team; Variation B shows what your team will own and the test that proves they can run it, but loses the sense of the gradual handoff.
- **Recommendation:** Variation B. The deck is read independently by a sponsor who has seen optimistic reporting and by procurement, and the readiness tests give both of them something they can check at week 16. If space allows, add Variation A's three stage names as a thin strip above the table.

---

## 5. Placeholders and open questions

**Placeholders to fill**

- [Number] payer data engagements, plus one mid-stream scope swap made without a change order (engagement, what changed, when).
- West Monroe names: engagement lead, executive sponsor, three workstream leads.
- Client names and counts: [platform engineers] per workstream, [architecture lead], [report owners], [platform lead] who gates the knowledge transfer stages, other steering committee members.
- Scope facts: claims [source systems], member 360 [other sources], [number] of reports and the [current reporting tool].
- Timing values: workstream and program review frequency (proposed weekly), steering pre-read lead time, ad hoc escalation window, estimate turnaround for scope changes, knowledge transfer stage weeks.
- Status color thresholds ([N] days for amber and red).
- Readiness tests: [a full month of claims loads], [a new member attribute], [the last N reports].

**Open questions**

1. Can we name the previous vendor's two-month slip directly on EM-04 (as written), or should it be phrased without "your previous vendor" since procurement will read it too?
2. Is "swap scope without a change order" a commitment you want in the proposal, and is "fits the current budget and 16 weeks" the right limit?
3. Are the data steward's approval rights (member and claims definitions) and the product owner's acceptance rights what the client expects?
4. Does the member 360 model depend on claims ingestion as described on EM-02? This is our inference, not a fact from the brief.
5. Is there a real anecdote from a payer engagement that fits EM-04 or EM-05? There's a slot for one on each.
6. Does this section need a short closing ask (the role names before week one), or does the proposal's next-steps page cover it?
