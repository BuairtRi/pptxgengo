Original authored notes
Metrics use data DXC already has: Jira (including the DXCM, CORE, and ITSUPPORT projects referenced in the release calendar), the release calendar, the in-house build system's build IDs, Git, and support tickets. Delivery metrics follow the DORA definitions. Story points remain a planning tool. The Reconcile AI team has already proposed cycle time by stage and flow charts, which this set builds on.

Original visible source copy
Measure what matters: ten metrics from data DXC already has
Today the only delivery number is story points. These cover speed, stability, flow, quality, and AI adoption and cost.
MOVE 4 · MEASURE WHAT MATTERS
© 2026 West Monroe Partners | Reproduction and/or distribution without West Monroe Partners’ prior consent is prohibited.
21
Group
Metric
Definition
Source (exists today)
Cadence
Delivery
Deployment frequency
Production releases per month
Release calendar, build system
Monthly
Lead time for changes
Merge to production, in days
Git + build system
Monthly
Change failure rate
Share of releases that need a patch or rollback
Release calendar + support tickets
Monthly
Time to restore
Incident opened to service restored
Support tickets
Monthly
Flow
Cycle time by stage
Days a story spends in each Jira status
Jira
Every sprint
Unplanned work share
Share of sprint capacity spent on support and patches
Jira labels
Every sprint
Commitment reliability
Stories done versus committed
Jira
Every sprint
Quality
Escaped defects
Production defects per release, by cause
Support tickets + Jira
Every release
Critical-path automation
Share of defined critical paths with automated tests
Test suite
Every release
AI
Adoption and cost
Builders with sanctioned access; model cost per run, as QA already tracks
Access lists, QA cost log
Monthly
How we start:  Weeks 1–2: baseline from data that already exists  ·  Weeks 3–4: agree definitions, add Jira cause codes and labels  ·  Weeks 5–6: one dashboard, first review in retro and with leadership
Story points stay as a planning tool, not a performance measure. The same view covers Core, the new platform, and the vendor team.

Source provenance
DXC_Engineering_Findings_and_Action_Plan_DRAFT 1.pptx · original page 14
SHA256 b6a045336db6bf2238f5458f12ae55d18a2086cb1d1a4cc5c9a092db68866f58
