# Our approach: AI-assisted SDLC (voice test)

Section of a proposal for the CIO and three engineering directors at [bank name]. Presented live, then left behind. Seven pages.

Bracketed text is a visible placeholder. Every mechanism described is our proposed design, not something we have already agreed with you. Open questions 9 and 10 cover it.

---

## 1. Narrative

Your 140 engineers ship a release every six weeks, and two things set that pace: nine days of manual regression testing in every release, and a merge queue that waits on the two senior engineers who approve most changes. If agents write more of your code and the process around them stays the same, both bottlenecks will likely get worse, because more changes arrive at the same two approvers and the same nine-day test cycle. A year ago the question was where the human sits in the loop. Now the question is how you design the loop. We propose one loop for your teams. Your engineers write the product requirements and design the review. Agents write the code and the regression tests. A repeatable adversarial review checks every change before an engineer approves the merge, and every step leaves the record your examiners will ask for. We'd prove the loop with one pilot team, and we're asking you to choose that team, name a compliance contact, and open access to your delivery tools.

---

## 2. Titles only

1. Nine days of manual regression testing and two merge approvers set the pace of your six-week release.
2. A year ago the question was where the human sits in the loop; now the question is how your team designs the loop.
3. Your engineers write the requirements and design the review; agents write the code and the regression tests.
4. Adversarial review moves your senior engineers from approving most merges to designing the checks every merge must pass.
5. We'd move regression testing from a nine-day manual cycle to generated tests that run on every merge.
6. Your examiners will ask who approved each change and why; the loop records the answer for every merge.
7. We'd prove the loop with one pilot team over [number] releases, starting with three decisions from you.

---

## 3. Pages

### AP-01

**Eyebrow:** Where you are today

**Title:** Nine days of manual regression testing and two merge approvers set the pace of your six-week release.

**Visual (left): release timeline**
- Bar label: "One release: six weeks"
- Highlighted block: "Manual regression testing: 9 days"
- Caption under the bar: "Every release, before it ships." (Block position is set after open question 1 is answered.)

**Visual (right): merge funnel**
- Top: "140 engineers open merge requests"
- Narrow point: "2 senior engineers approve most of them"

**Bullets**
- **Nine days of re-testing.** Manual regression testing takes nine days in every release, so testing by hand, not building, sets the release calendar.
- **Two approvers.** Two senior engineers approve most merges, so changes from 140 engineers wait in one queue for two people.
- **What agents add.** If agents write more of the code and the process stays the same, more changes arrive at the same two approvers and the same nine-day test cycle.

**Source line:** Engineer count, release cadence, regression duration and approver count: [bank name] engineering, [date provided].

**Form:** Two-part page: a six-week timeline bar with the nine-day block called out, beside a two-level funnel. Three bolded-lead bullets underneath.

---

### AP-02

**Eyebrow:** Our point of view

**Title:** A year ago the question was where the human sits in the loop; now the question is how your team designs the loop.

**Lead-in:** When agents write the basic code, your senior engineers' time is worth more on what to build and how to check it than on reading each line.

**Left column header:** Less of your engineers' time

- **Line-by-line review of basic code.** Agents write it; the adversarial review and the regression tests check it.
- **Approving most merges personally.** Two people can't read every change from 140 engineers and their agents.

**Right column header:** More of your engineers' time

- **Product requirements.** The requirement is what the agent builds from and what the adversarial review checks against, so a vague requirement produces vague code.
- **Designing the adversarial review.** Your senior engineers decide what the review looks for and which changes need a senior approver.
- **Judging the exceptions.** Changes the adversarial review flags come to an engineer with the finding and the reason attached.

**Form:** Concede-and-redirect two-column comparison. The left column is visually quieter (gray). The right column carries the emphasis in Grounded Blue.

---

### AP-03

**Eyebrow:** The approach

**Title:** Your engineers write the requirements and design the review; agents write the code and the regression tests.

**Diagram: five-stage loop (stage label / who acts / what it produces)**

1. **Requirement.** Your product owner and engineer write it with acceptance criteria.
2. **Code.** An agent writes the change and proposes regression tests from the acceptance criteria.
3. **Adversarial review.** A separate reviewer agent, briefed to find what is wrong, checks the change against the checks your senior engineers wrote.
4. **Regression tests.** The full generated suite runs against the change.
5. **Merge approval.** An engineer approves using the review findings and test results; changes in a high-risk tier go to a senior engineer.

**Arrow labels**
- 1 → 2: "approved requirement"
- 2 → 3: "change + proposed tests"
- 3 → 4: "no open findings"
- 4 → 5: "all tests pass"
- 3 → 2 and 4 → 2 (return arrow): "finding or failed test, sent back with the reason"
- 5 → out: "merged to the release branch"

**Band under the diagram:** **Evidence log.** Each stage records who or what acted, on which version, and with what result.

**Caption:** The loop advances only when a stage's condition is met. Nothing merges on an agent's approval alone.

**Form:** Left-to-right process flow with labeled arrows and one return arrow, plus a full-width evidence band underneath that ties to AP-06.

---

### AP-04

**Eyebrow:** Adversarial review

**Title:** Adversarial review moves your senior engineers from approving most merges to designing the checks every merge must pass.

Form is uncertain; two variations in section 4.

---

### AP-05

**Eyebrow:** Regression testing

**Title:** We'd move regression testing from a nine-day manual cycle to generated tests that run on every merge.

**Visual: two release timelines, stacked**
- Top bar, "Today": six-week release with a block labeled "Manual regression: 9 days".
- Bottom bar, "Proposed": tick marks along the whole bar labeled "Generated regression tests run on each merge", plus a short block labeled "Release check: [target] days".
- Gap callout between the two blocks: "9 days today → [target] days. Target set during the pilot."

**Bullets**
- **Tests from requirements.** The agent proposes regression tests from each requirement's acceptance criteria, so the suite grows with every change instead of in a separate test project.
- **Your testers approve the suite.** [Team that owns regression testing today] decides which generated tests enter the suite and retires the ones that no longer apply, instead of re-running scripts by hand each release.
- **Run before merge approval.** A failing change goes back to its author before it merges, instead of turning up in the release-end test cycle.
- **Manual testing keeps a role.** Exploratory testing and [areas your team must test by hand] stay with your testers in a shorter release check.

**Callout (cost):** We expect the first [number] releases to be slower, because your testers review more generated tests up front than they will once the suite is established.

**Form:** Before-and-after timeline comparison with the gap called out, four bolded-lead bullets to the right, and a small cost callout at the bottom.

---

### AP-06

**Eyebrow:** Governance

**Title:** Your examiners will ask who approved each change and why; the loop records the answer for every merge.

**Table**

| What an examiner may ask | Where the loop records the answer |
| --- | --- |
| Who asked for this change? | The requirement, its author and the person who approved it |
| What wrote the code? | The agent, model version and instructions used for that change |
| How was it checked? | Adversarial review findings, how each one was resolved, and regression test results |
| Who approved the merge? | The named engineer, plus the senior engineer for high-risk tiers |
| Who changed the rules? | Version history of the review checks, risk tiers and test suite, with the approver of each change |

**Callout (objection):** Your teams may expect regulators to object to agent-written code. Before the pilot starts, we'd map the loop to [your change-management and model risk policies] with [your compliance or model risk contact], so the record answers the questions your examiners already ask.

**Bullet under the table**
- **No agent-only merges.** Every merge has a named human approver, and the evidence log shows what that person saw.

**Form:** Two-column question-and-answer table as the main element. Objection callout in a magenta-edged panel to the right.

---

### AP-07

**Eyebrow:** Next steps

**Title:** We'd prove the loop with one pilot team over [number] releases, starting with three decisions from you.

**Left panel: proof**
- **[Number] comparable engagements.** [Segment, for example AI-assisted delivery at regulated financial institutions], since [year].
- **[Named client or result, if approved for use].**

**Middle panel: what the pilot measures against today**
- **Regression time.** Nine days of manual testing per release today.
- **Merge wait.** Time from merge request to approval. [Baseline to measure in week one]
- **Approver load.** Share of merges approved by the two senior engineers. [Baseline to measure]
- **Escaped defects.** Defects found after merge. [Baseline from your defect tracker]

**Right panel: decisions (numbered)**
1. **Choose the pilot team.** Your engineering directors pick one team and its codebase by [date].
2. **Name a compliance contact.** Your CIO names the [compliance or model risk] lead who reviews the loop design by [date].
3. **Open access.** Your platform team gives us access to [source control and CI/CD tools] by [date].

**Closing line:** In week one, we write the first requirements and adversarial review checks with your senior engineers.

**Form:** Three-panel page. Proof on the left, measures in the middle, numbered decisions on the right. The decisions panel gets the visual weight.

---

## 4. Variations for AP-04

Both variations share the title:

**Adversarial review moves your senior engineers from approving most merges to designing the checks every merge must pass.**

### Variation A: diagram-led

**Diagram: review routing**
- Input box: "Change + proposed tests"
- Three parallel reviewer boxes, each with one brief:
  - **Meets the requirement.** Checks the change against its acceptance criteria.
  - **Tries to break it.** Hunts for edge cases, error handling gaps and [your known failure patterns].
  - **Security and data.** Checks customer-data handling and dependencies, alongside [your existing scanners].
- Merge point: "Findings, each with the line and the reason"
- Router: "Risk tier set by your senior engineers"
  - Arrow up: "Low tier, no open findings → any engineer approves"
  - Arrow down: "High tier or open finding → senior engineer approves"

**Side panel, "Your senior engineers own":**
- The review briefs
- The risk tiers
- A new check after every escaped defect

**Callout (objection):** Your engineers may worry that agents will produce "AI slop." Adversarial review is how you catch it: every change faces a reviewer whose only job is to find what is wrong, using checks your senior engineers wrote.

**Form:** Fan-out/fan-in routing diagram, a narrow ownership panel on the right, and the objection callout along the bottom.

### Variation B: text-dense

**Table**

| Review check | Who runs it | What it catches | When an engineer sees it |
| --- | --- | --- | --- |
| Meets the requirement | Reviewer agent | Code that misses an acceptance criterion | Every miss, with the criterion named |
| Tries to break it | Reviewer agent briefed to find failures | Edge cases, error handling gaps, [your known failure patterns] | Every finding, with the line and the reason |
| Security and data | Reviewer agent plus [your existing scanners] | Customer-data exposure, unsafe dependencies | Every finding |
| Risk tier | Rules your senior engineers set | Changes to [high-risk areas, for example payments or customer data] | Every high-tier change goes to a senior engineer |
| New checks | Your senior engineers | Defects that got past the review | Each escaped defect becomes a new check |

**Bullets under the table**
- **Senior time moves upstream.** Your two senior engineers write and tune the checks instead of approving most merges one by one.
- **Same checks every time.** Every change gets the same adversarial review, so review quality doesn't depend on who has time that afternoon.

**Callout (objection):** Your engineers may worry that agents will produce "AI slop." Adversarial review is how you catch it: every change faces a reviewer whose only job is to find what is wrong, using checks your senior engineers wrote.

**Form:** Four-column table as the main element, two bolded-lead bullets underneath, objection callout on the right.

### Tradeoff and recommendation

**Tradeoff:** A shows routing and who decides at a glance, which works well when presented live, but it drops the "what it catches" detail. B carries that detail for every check but reads as a list when presented.

**Recommendation:** B. Skeptical engineering directors will judge this page on mechanism, and as a leave-behind it is read without a presenter. If Ri wants A for the live session, put B's table in the appendix and reference it from A.

---

## 5. Placeholders and open questions

**Placeholders**
- [bank name]; [date provided] for the source line on AP-01.
- [Number] comparable engagements, their [segment], [year], and any [named client or result] approved for use (AP-07).
- [target] days for the release check, and [number] slower releases at the start (AP-05).
- [number] pilot releases (AP-07 title); three decision [dates] (AP-07).
- [Team that owns regression testing today]; [areas your team must test by hand] (AP-05).
- [your known failure patterns]; [your existing scanners]; [high-risk areas] (AP-04).
- [your change-management and model risk policies]; [compliance or model risk contact] (AP-06, AP-07).
- [source control and CI/CD tools] (AP-07).
- Baselines for merge wait, approver load and escaped defects (AP-07).

**Open questions for Ri**
1. Are the nine days of regression testing working or calendar days, and do they fall at the end of the six-week cycle? The timeline on AP-01 and AP-05 depends on the answer.
2. Do "two senior engineers approve most merges" cover all 140 engineers, or one group? Is there a share (for example, [x]% of merges) we can cite?
3. Who owns regression testing today: a separate QA team or the product teams?
4. Which regulators and internal policies apply (for example, your change-management standard, model risk policy or third-party risk review of the agent tooling)? We have not named any on the pages.
5. Has anyone at the bank raised "AI slop" in those words? If not, AP-04's callout should use the readers' own phrase.
6. Do we have a real example from a comparable engagement for AP-04 or AP-07, such as a defect the adversarial review caught that line-by-line review missed? There is a slot for one; none was written.
7. Is a pilot of [number] releases with one team the commercial shape of this proposal, or does the approach section need to show a broader rollout?
8. Is seven pages acceptable? The brief said about six. AP-01 could merge into AP-02 if the proposal already has a situation section before "Our approach".
9. The loop's mechanisms (separate reviewer agents, risk tiers, evidence log, no agent-only merges) are our proposed design. Do they match what West Monroe would commit to, and are they consistent with the comparable engagements?
10. Is "nothing merges on an agent's approval alone" a commitment we want to make in writing?

---

## Self-check notes

- **Titles.** All seven are full sentences stating a conclusion, and read in order they tell the story: the bottleneck, the point of view, the loop, review, testing, governance, the pilot.
- **Competitor test.** The point of view ("design the loop"), the client's own numbers (140, six weeks, nine days, two approvers) and the named mechanisms keep the pages specific to this client and this proposal. No generic expertise claims.
- **Proof.** The only proof claim is the [number] comparable engagements placeholder. Targets are shown as placeholders, never as commitments.
- **Hedges.** "Likely" (narrative), "We expect" (AP-05) and "may" in the two objection callouts. None are stacked.
- **Key terms held constant.** "Adversarial review", "regression tests", "merge approval", "the loop", "risk tier", "evidence log".
- **Invention.** No client names, regulations, quotations, anecdotes or outcome numbers were added.
