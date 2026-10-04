Original authored notes
Target flow keeps the Core team, Java stack, and Scrum cadence. Gates are process changes plus AI assistance; same-build promotion is the only change that touches the release pipeline. The PR gate reduces dependence on a small senior review group without removing senior review of high-risk changes.
The hardening window adopts the plan QA leadership has already presented: release 9–10 days after dev done, which requires templated environments so two releases can be in flight, and a settled branch strategy.

Original visible source copy
Stabilized Core SDLC: seven gates on the flow Core already runs
Same team, same Java stack, same tools. Each gate has an owner and makes a weak point visible.
CORE SDLC · STABILIZED
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
Gate we add
Ready gate: acceptance criteria, test cases, and tech notes before a story enters the sprint.
Run-lane allowance: capacity reserved for support and patches.
Claude Code pairing: code, unit tests, and docs before the PR.
PR gate: Claude PR bot in the pipeline plus a rotating second reviewer.
Quality gate: defined critical paths automated; evidence linked to the release.
Hardening window: release 9–10 days after dev done; the build that passed pre-live ships.
Hotfix lane: root cause within the sprint; credential and certificate alerts.
Where Claude helps
Drafts acceptance criteria and test cases.
Sizes slices from the spec.
Writes code, unit tests, and docs.
Pre-reviews every PR in the pipeline.
Agents draft tests from the backlog.
Drafts release notes from delivered work.
Drafts the root-cause summary.
DXC owner
PMs + QA lead
Scrum master
Core engineers
Core engineering lead
QA lead
Release management + DevOps
Release management
No re-platforming: we add gates, evidence, and AI assistance to the existing flow, and back QA's hardening plan.

Source provenance
DXC_Engineering_Findings_and_Action_Plan_DRAFT 1.pptx · original page 4
SHA256 b6a045336db6bf2238f5458f12ae55d18a2086cb1d1a4cc5c9a092db68866f58
