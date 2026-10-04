Original authored notes
QA layers. Integration checks respond to the 2026 Salesforce password expiries, which broke HS1 case creation and calls into Core in production. The capacity practice comes from QA leadership: time-box a problem for 15–20 minutes before asking Claude, use it to review work before a PR, and never to do the whole job. The added QE capacity is a deliverable of the Testing & release controls workstream.
QA tools already built: Claude-based Playwright agents (built for HIPAA compliance, with human review points), a Claude PR bot with per-run cost tracking (not yet in the Bitbucket pipeline), and a Jira story-to-test prototype that is not yet hosted in DXC's environment. Claude is also used to check that tests follow the framework conventions. The test library holds 3,000+ atomic scenarios after Eligibility AI was imported, with a small share automated.

Original visible source copy
Robust QA: scale what QA has built, layer by layer
Seven layers, each with an owner. The gaps are coverage, performance, environments, and QA leadership capacity.
MOVE 3 · ROBUST QA
© 2026 West Monroe Partners | Reproduction and/or distribution without West Monroe Partners’ prior consent is prohibited.
21
Layer
Who
What changes
Today
Refinement
QA lead + PMs
Test cases written from acceptance criteria; QA's Jira test generator drafts them.
QA not in the room
Unit
Core engineers + Claude Code
Changed code ships with unit tests, drafted by Claude and reviewed.
Engineering only; not measured
API and UI
QA (Playwright, Postman)
Claude-based agents draft tests from the backlog daily; the PR bot reviews them.
200+ automated on Core; most manual
Critical paths
QA + Product
Paths already defined with Product, automated release by release.
10–15% automated
Performance
QA + DevOps
Baseline load tests on the highest-volume Core flows.
Not done
UAT
Product owners
Realistic scenarios in pre-live, inside the hardening window.
3–4 day window
Production checks
Release management + DevOps
Smoke tests and post-checkout evidence linked to the release.
In place: smoke tests, post-checkout
Capacity:  one QA leader covers about 15 testers, two of them full-time  ·  name a QA lead  ·  Claude for every tester, contractors included  ·  one sprint of added QE capacity
QA built the framework and the AI tools in five months. The next phase gives them coverage, a production home, and a lead.

Source provenance
DXC_Engineering_Findings_and_Action_Plan_DRAFT 1.pptx · original page 12
SHA256 b6a045336db6bf2238f5458f12ae55d18a2086cb1d1a4cc5c9a092db68866f58
