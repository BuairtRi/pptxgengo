Original authored notes
Sources: DXC production release calendar (Jan 2024 – Sep 2026, Core applications); working sessions with Core engineering, release management, new-platform engineering, Reconcile AI engineering, and QA leadership; engineering survey (n=44); Commercialization Framework (Concord). Finding 1: 30 release events in 2026 to Sep 1 (15 standard, 11 patches, 2 off-version, 2 rollbacks). Tickets exist (DXCM, CORE, ITSUPPORT), but we found no classification or trend of causes; confirm with release management before presenting. Core engineering separately reported 62 releases and 15 patches, counted per deployment. Finding 2: the Salesforce client password expired on 05/30 (blocked HS1 case creation) and 08/28; outbound Salesforce calls to Core were restored on 09/01. BMA, DIL, and SSO certificate updates also shipped as patches; the calendar does not say whether they had expired. Finding 3: release management described separate staging and production branches merged on release day, and flagged it as something that may need to change; late changes delay product owner sign-off within a Monday-to-Thursday window. A rebuild is inferred, not confirmed. Finding 4: QA leadership reports Claude substantially cuts test creation and diagnosis time; access is limited to some full-time staff and not yet available to contractors. Half of the Core team is offshore. Finding 5: the Rules Engine has a rough MVP estimate toward year-end, with platform and team placement undecided and no epic. New-platform engineering leadership named commitments made without engineering input as its top issue. Internal context only: the Portfolio Governance Board has no engineering seat, and the G0 effort tier is confirmed by the PjM.
QA session (senior QA leadership): when QA leadership arrived five to six months ago there was no QA process or automation framework. Core is the proving ground, now rolling out to the other eight or nine pods. Core went from 21 to 200+ automated tests; estimated coverage about 50%, with critical paths defined with Product and about 10–15% automated. Eligibility AI is estimated at 15–20% automated. QA is largely contractors, with two full-time staff; one leader oversees about 15 testers and asks for a QA lead. Code ships four days after dev done, leaving three to four days for QA and UAT; pre-live is not production-like. A hardening-sprint plan (release 9–10 days after dev done) has been presented to engineering leadership; environments, which run on VMs, are the main constraint.

Original visible source copy
Six engineering findings, each with a clear fix
From DXC's own release records and working sessions. Each finding points to a specific action and owner.
ENGINEERING FINDINGS
© 2026 West Monroe Partners | Reproduction and/or distribution without West Monroe Partners’ prior consent is prohibited.
21
FINDING
WHAT WE FOUND
SO WHAT
1
Unplanned, untracked
Half of 2026 Core release events were unplanned (15 of 30), and their causes are not classified or trended.
Without cause data, the same failures repeat. A two-week triage turns the patch count into a fix list.
2
Same failure, twice
The Salesforce client password expired in May and again in August, blocking HS1 case creation and calls into Core.
Predictable expiries are handled as incidents. An inventory with owners and alerts prevents the repeat.
3
Tested ≠ shipped
Code ships four days after dev done, pre-live is not production-like, and branches merge on release day.
Staging evidence may not match production. Add a hardening window, template environments, promote the tested build.
4
AI locked to few
Claude already cuts test-writing and diagnosis time, but only some full-time staff can use it. Contractors, most of QA, cannot.
AI gains stay with a few people. Extend sanctioned access, with the same guardrails, to everyone who builds.
5
Dates before design
The Rules Engine carries a year-end MVP estimate while its platform, team, and first epic are still undecided.
Predictability breaks at commitment. No customer date without an engineering estimate and a signed tech spec.
6
QA built, not staffed
In five months QA grew Core automation from 21 to 200+ tests and built Claude-based tools, with one leader over about 15 testers.
The foundation exists. It needs a production home for the tools, coverage of critical paths, and a QA lead.
Sources: DXC production release calendar (Jan 2024 – Sep 2026, Core applications); working sessions with engineering, release, and QA; engineering survey (n=44).

Source provenance
DXC_Engineering_Findings_and_Action_Plan_DRAFT 1.pptx · original page 2
SHA256 b6a045336db6bf2238f5458f12ae55d18a2086cb1d1a4cc5c9a092db68866f58
