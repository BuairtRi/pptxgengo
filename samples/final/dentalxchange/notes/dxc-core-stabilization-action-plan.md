Original authored notes
Action plan for move 1. Owners are DXC roles; names to confirm with DXC. Weeks follow the two engineering workstreams in the proposal.

Original visible source copy
Stabilize Core SDLC: action plan
Fixes findings 1, 2, and 3. Delivered through Testing & release controls, with the process map in Reconcile AI engineering across Core.
MOVE 1 · STABILIZE CORE SDLC
© 2026 West Monroe Partners | Reproduction and/or distribution without West Monroe Partners’ prior consent is prohibited.
21
Action
What, exactly
DXC owner
West Monroe role
Weeks
Done when
Patch triage
Classify the 15 unplanned 2026 changes by cause: code, configuration, data, credential, or unknown.
Release management
QE Architect
1–2
Every unplanned change has a cause and an owner.
Cert inventory
List every credential and certificate with expiry date and owner; alert before expiry.
Release management + DevOps
QE Architect
1–3
100% of credentials and certificates have an owner and an alert.
Hardening window
Adopt QA's plan to release 9–10 days after dev done, and settle the branch strategy.
Engineering leadership + QA lead + Release
Executive Engineering Leader
1–3
First Core release ships on the new window.
Templated environments
Template the VM-based QA and pre-live environments so two releases can be in flight.
DevOps + application architects
QE Architect
2–6
Environments start from a template; two releases run without conflict.
Same-build promotion
Release the build that passed pre-live instead of a new merge on release day.
DevOps + Release management
QE Architect
3–6
Pre-live and production build IDs match for every release.
Hotfix lane
Each patch or rollback gets a cause code and a short review within the sprint.
Release management + Core engineering lead
QE Architect
2–6
No patch closes without a recorded root cause.
Decision needed:  Name the triage owner, approve the hardening window and same-build promotion, and fund environment templating by week 2.
Measure:  Unplanned release events (2026: 15 of 30)  ·  Repeat causes: zero  ·  Test window: 3–4 days → 9–10  ·  Matching builds: 100%

Source provenance
DXC_Engineering_Findings_and_Action_Plan_DRAFT 1.pptx · original page 9
SHA256 b6a045336db6bf2238f5458f12ae55d18a2086cb1d1a4cc5c9a092db68866f58
