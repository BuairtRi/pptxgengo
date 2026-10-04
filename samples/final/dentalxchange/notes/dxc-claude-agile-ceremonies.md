Original authored notes
Adapted from Reconcile AI's agentic spec-driven delivery (spec, technical stories, agent implementation, verification, delivery report). On Core, the pattern is applied inside the existing Scrum sprint and Java codebase; it is not a modernization. Claude Code access requires the sanctioned-tool and PHI decisions at week 2.

Original visible source copy
Agile with Claude Code: what changes in each ceremony
Reconcile AI's spec-driven pattern and QA's tools, adapted to Core's Scrum. People keep every decision; Claude does the drafting.
MOVE 2 · AGILE WITH CLAUDE CODE
© 2026 West Monroe Partners | Reproduction and/or distribution without West Monroe Partners’ prior consent is prohibited.
21
Ceremony
Today
With Claude Code
Human gate
Refinement
PMs and engineers estimate stories by vote.
Claude drafts acceptance criteria and test cases from the spec.
PM and QA lead approve the story as Ready.
Planning
About 65 points; the lead assigns work by skill.
Smaller slices; capacity reserved for support and patches.
Scrum master and lead confirm capacity.
Build
Copilot autocomplete; docs written after coding.
Claude Code implements, writes unit tests, and drafts docs in the existing Java codebase.
The engineer owns every line merged.
Review
Senior review for every PR.
QA's Claude PR bot, moved into the pipeline, pre-reviews every Core PR.
Rotating reviewer approves; senior review for high-risk PRs.
Retrospective
Held every sprint, without delivery data.
Cycle time by stage, escaped defects, and unplanned work on screen.
The team commits to one change per sprint.
Guardrails:  sanctioned access for every engineer, contractors included  ·  tools run in DXC's accounts  ·  no PHI in prompts  ·  cost tracked per run
Not a modernization of Core: same backlog, same sprints, same codebase. Pilot on the Rules Engine and one Core team, then scale at the week-6 gate.

Source provenance
DXC_Engineering_Findings_and_Action_Plan_DRAFT 1.pptx · original page 10
SHA256 b6a045336db6bf2238f5458f12ae55d18a2086cb1d1a4cc5c9a092db68866f58
