# Proposal copy and cardinality stress set

This isolated stress set copies all 13 current proposal value files into `values/` and keeps the narrative binding to `../narrative.json`. Only six dense pages are expanded: situation/friction, needs/responses, journey workflow, parallel paths, measures/release decisions, and team/accountability. Each addition is grounded in the pinned fictional qualification brief (`../brief.md`) and the matching narrative details. The additions clarify existing ownership, evidence, control, readiness, and transition statements; they introduce no new quantitative or outcome claims.

No contract, generator, normal proposal value, geometry, or font input was changed. Every populated string remains within its contract character cap, and all exact narrative required details and qualifications remain present. The two parallel-path arrays use six steps each, within the contract cardinality of 4–6.

## Copy increase

Counts below sum characters across populated value strings and array labels in each page. The increase compares the copied source values with the stress values; it is a content-volume check, not a measured fit result.

| Page | Source chars | Stress chars | Increase |
|---|---:|---:|---:|
| `situation-and-friction` | 1961 | 2333 | 19.0% |
| `needs-and-responses` | 1592 | 1881 | 18.2% |
| `journey-workflow` | 2116 | 2487 | 17.5% |
| `parallel-delivery-paths` | 1261 | 1486 | 17.8% |
| `measures-and-release-decision` | 2047 | 2404 | 17.4% |
| `team-and-accountability` | 1683 | 1942 | 15.4% |

All six expanded pages fall within the requested approximate 10–20% copy increase.

## Changed slots and source rationale

### `situation-and-friction`

- `workflow-matrix/workflow-rows/workflow-0-column-2/content/text`: 97 → 177 characters — Added current transaction-ownership boundary, matching first-release plan.
- `workflow-matrix/workflow-rows/workflow-1-column-2/content/text`: 124 → 203 characters — Clarified comparison between spreadsheet and processing records.
- `workflow-matrix/workflow-rows/workflow-2-column-2/content/text`: 124 → 223 characters — Distinguished policy-required transfers from potentially avoidable rework.
- `workflow-matrix/workflow-rows/workflow-3-column-2/content/text`: 123 → 205 characters — Separated incomplete submissions from waiting decisions and genuine demand.
- `workflow-matrix/workflow-rows/workflow-4-column-2/content/text`: 110 → 151 characters — Restated continued-system availability in the investment-scope context.

### `needs-and-responses`

- `enablecomp-response-fixture/row-copy-1/p-body-1/r-body-1/text`: 240 → 295 characters — Clarified consistency of intake measures across alternate routes.
- `enablecomp-response-fixture/row-copy-2/p-body-2/r-body-2/text`: 139 → 190 characters — Made reassignment reasons explicit and retained no-inbox-knowledge boundary.
- `enablecomp-response-fixture/row-copy-3/p-body-3/r-body-3/text`: 171 → 235 characters — Added legacy event-update feasibility validation.
- `enablecomp-response-fixture/row-copy-4/p-body-4/r-body-4/text`: 187 → 251 characters — Clarified distinct operational views for exception reasons.
- `enablecomp-response-fixture/row-copy-5/p-body-5/r-body-5/text`: 165 → 220 characters — Placed counterpart ownership before production release.

### `journey-workflow`

- `workflow-matrix/workflow-rows/workflow-0-column-1/content/text`: 103 → 169 characters — Clarified transaction-state origin and change authority.
- `workflow-matrix/workflow-rows/workflow-1-column-1/content/text`: 100 → 193 characters — Made the transition alternate-route control explicit.
- `workflow-matrix/workflow-rows/workflow-2-column-1/content/text`: 93 → 189 characters — Added ownership and reassignment evidence at the decision point.
- `workflow-matrix/workflow-rows/workflow-4-column-2/content/text`: 113 → 229 characters — Connected exception categories to next actions, escalation owner and resumption evidence.

### `parallel-delivery-paths`

- `process-changed-content/case-controls/text`: 240 → 330 characters — Linked case history to identifier for review of ownership and reassignment.
- `process-changed-content/release-controls/text`: 260 → 362 characters — Expanded explicit recording of control findings, owners and recovery actions.
- `process-changed-content/request/steps`: 77 → 99 characters — Added a sixth step to record decision reason.
- `process-changed-content/change/steps`: 70 → 81 characters — Expanded five steps into six: separated test evidence and operations readiness before steering gate.

### `measures-and-release-decision`

- `workflow-matrix/workflow-rows/workflow-0-column-1/content/text`: 57 → 152 characters — Added sampling-window, definition, exclusion and owner documentation.
- `workflow-matrix/workflow-rows/workflow-1-column-1/content/text`: 106 → 190 characters — Clarified stable measure definition and population across baseline and pilot.
- `workflow-matrix/workflow-rows/workflow-2-column-2/content/text`: 120 → 221 characters — Added evidence owner and decision-date tracking for unresolved readiness items.
- `workflow-matrix/workflow-rows/workflow-4-column-1/content/text`: 69 → 146 characters — Added common denominator and data limitations for steering materials.

### `team-and-accountability`

- `dense-delivery-relationships/client-b-0/text`: 62 → 88 characters — Clarified sponsor/steering ownership of proceed and hold scope gates.
- `dense-delivery-relationships/client-b-1/text`: 49 → 64 characters — Added program lead escalation of dependency blockers.
- `dense-delivery-relationships/client-b-2/text`: 92 → 136 characters — Made review with client architect at gates explicit.
- `dense-delivery-relationships/client-b-3/text`: 164 → 243 characters — Clarified shared specialists participate at gates; role and allocation categories are illustrative, not committed staffing levels.
- `dense-delivery-relationships/staffing-body/text`: 107 → 158 characters — Clarified that the staffing colors follow the legend and allocation categories remain illustrative.
- `dense-delivery-relationships/staffing-boundary/content/detail-1/text`: 114 → 166 characters — Clarified that the 21-role diagram represents responsibilities, not 21 full-time people, and that client decision rights are confirmed during mobilization.

## CLI qualification

Run the repeatability and negative qualification checks with:

```sh
python3 scripts/check-proposal-stress.py
```

The runner calls the CLI executable at `/tmp/pptxlib-wave4-integration` by default and uses `samples/proposal-authoring/capacity-v6.sqlite`. It assembles all 13 pages twice in a new directory under `/tmp`, checks both the combined deck spec and all 13 individual `spec.json` SHA-256 hashes match, and records timings and complete diagnostics in `diagnostics.json`. It then verifies rejection of a seven-step array, a string one character over its slot cap, an unknown slot, and an unknown style; each rejected assembly must leave no output directory. Use `--cli`, `--index`, and `--out` to select alternatives; the output directory must not already exist.

The latest root run produced 13 matching individual specs and an identical combined spec across both assemblies (1.412 s and 1.362 s), and passed all four negative checks. Diagnostics are preserved at `samples/proposal-stress-qualification-v2/diagnostics.json`. A deliberate missing-CLI run also retained a failed-status report and the exact error. These are file-only CLI/semantic checks. Specs remain `unmeasured_changed_copy`; there was no PowerPoint, native fit measurement, or visual review, so this does not establish fit or layout quality.
