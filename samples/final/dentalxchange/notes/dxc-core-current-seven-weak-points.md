Original authored notes
Core SDLC as described in the Core engineering, release management, and QA sessions: grooming with two PMs; refinement with 10 engineers, PMs, and a scrum master, estimating by vote; two-week sprints of about 65 story points; work assigned by skill, with research-heavy items kept onshore because offshore has no live data access; Copilot in IntelliJ since January 20, 2026, Claude for some; documentation written after coding; senior review of every Core PR before a QA build; staging on Monday, product owner sign-off, production on Thursday, with a branch merge on release day; patches outside the cadence. Half of the Core team is offshore.
QA details from the QA session: Playwright framework; Claude-based Playwright agents with human-in-the-loop points; a Claude PR bot triggered manually; critical paths defined with Product, 10–15% automated; no performance testing yet; pre-live and QA are the two lower environments, and pre-live is not production-like.

Original visible source copy
Core SDLC today: a working flow with seven weak points
From the Core engineering, release, and QA sessions and the 2026 release calendar.
CORE SDLC · TODAY
© 2026 West Monroe Partners | Reproduction and/or distribution without West Monroe Partners’ prior consent is prohibited.
21
Core SDLC stage
1  Refine
▶
2  Plan
▶
3  Build
▶
4  Review
▶
5  QA
▶
6  Release
▶
7  Patch
How it works
PMs and 10 engineers estimate stories by vote; a scrum master runs the sprint.
Two-week sprints, about 65 story points; the lead assigns work by skill.
Copilot in IntelliJ since January 2026; Claude for some.
Senior review for every Core PR; QA's Claude PR bot runs on demand.
Playwright framework; Core went from 21 to 200+ automated tests in five months.
Code ships four days after dev done: pre-live Monday, production Thursday.
Fixes ship outside the cadence as patches.
Where it breaks
QA and technical design are not in refinement.
Patches and support land on the same people as sprint work.
Docs written after coding; offshore has no live data access.
Review rests on a small senior group.
Most testing still manual; critical paths 10–15% automated; no performance tests.
3–4 days for QA and UAT; pre-live is not production-like; merge on release day.
15 unplanned changes in 2026; causes not trended; same expiry twice.
Only one delivery number exists today: story points. Nobody can see lead time, failure rate, or where work waits.

Source provenance
DXC_Engineering_Findings_and_Action_Plan_DRAFT 1.pptx · original page 8
SHA256 b6a045336db6bf2238f5458f12ae55d18a2086cb1d1a4cc5c9a092db68866f58
