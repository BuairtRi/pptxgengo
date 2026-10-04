Original authored notes
Risks come from the working sessions: Claude access limited to some full-time staff; QA tools not yet hosted in DXC's environment; VM-based environments as the main constraint for the hardening window; QA and Core leadership capacity; the branch strategy debate (feature branches versus master); CAB friction with switched-off code; and the undecided Rules Engine platform.

Original visible source copy
Risks and dependencies: most are decisions, not unknowns
Each risk has a mitigation, an owner, and the week it is needed by.
RISKS AND DEPENDENCIES
© 2026 West Monroe Partners | Reproduction and/or distribution without West Monroe Partners’ prior consent is prohibited.
21
Risk or dependency
If it slips
Mitigation
Owner
Needed by
Claude access and PHI rule
AI moves stall for contractors
Decide at week 2; start with full-time pilot users
Engineering leadership + Security
Week 2
Security review to host QA tools in AWS
Test generator and PR bot stay manual
Start the review in week 1 on the existing Claude subscription
Security + DevOps
Week 3
Environment templating capacity
Hardening window cannot run two releases
Template pre-live first; stagger releases
DevOps + application architects
Week 2
Load on QA and Core leads
Actions slip; leads burn out
West Monroe QE engineers take execution; name a QA lead
Engineering leadership
Week 2
Branch strategy not settled
Hardening window blocked
Options paper and decision at the week-2 gate
Core lead + QA lead + Release
Week 2
CAB rule for switched-off code
Release friction continues
Written rule as part of the release gate
Release management + CAB
Week 3
Data quality for baselines
Metrics not trusted
Start with calendar and build IDs; add Jira cause codes
Engineering leadership
Week 3
Rules Engine platform undecided
Spec-first pilot delayed
Spec is platform-neutral; decide at week 2
Product + Engineering leadership
Week 2
Five of the eight are decisions at the week-2 gate. The gate is where the plan is won or lost.

Source provenance
DXC_Engineering_Findings_and_Action_Plan_DRAFT 1.pptx · original page 17
SHA256 b6a045336db6bf2238f5458f12ae55d18a2086cb1d1a4cc5c9a092db68866f58
