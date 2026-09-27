# Proposal library authoring notes

## Inputs and scope

The source is `narrative.json`, with `brief.md` as its sole pinned content source (SHA-256 `073118fd9caa3bdbbf5761af65b671887084eeb1ed5147c79213bec22aaafa66`). All claims remain synthetic, unverified, or hypotheses as stated. No changes were made to the brief or narrative. Values use only contract-owned semantic slots and the `source` style variant where the example specifies one. The opening page uses `wm/dense-argument` and retains the bound assertion title, role, takeaway, and a phrase accent on “observable end-to-end journey.”

`assembly.json` orders all 13 narrative slides and points to individual values files. The team and accountability page remains separate from the 19-role roster; the biography remains a separate synthetic role page. The parallel-process candidate still has two paths, and the architecture page uses the three-layer/nine-node contract without adding an external-systems node.

## Selection reasons

| Page | Contract | Reason |
|---|---|---|
| decision-and-thesis | `wm/dense-argument` | Three-part understand/prove/decide argument supports the opening decision request and puts authorization conditions beside the thesis. |
| situation-and-friction | `wm/workflow-matrix` | Relates current constraints, proposed response, and measures to validate. |
| needs-and-responses | `wm/needs-response` | Five paired operating needs and response mechanisms. |
| journey-workflow | `wm/workflow-matrix` | Journey rows can connect current friction, owner/control, and validation evidence. |
| five-phase-sequence | `wm/five-phase` | Five ordered stages pair activities and evidence outputs. |
| discovery-detail | `wm/phase-detail` | Objective, activities, outputs, assumptions, and attributed sample-artifact zone. |
| architecture-and-controls | `wm/layered-architecture` | Three responsibilities layers, nine nodes, and governance/security bands. |
| parallel-delivery-paths | `wm/parallel-process` | Exactly two paths for case decisions and change/release review. |
| roadmap-and-gates | `wm/release-roadmap` | Phase timing, dependency markers, and proceed/hold/adjust gates. |
| team-and-accountability | `wm/delivery-team` | Sponsorship, program, architecture, three pods, shared roles, and client relationships. |
| role-roster | `wm/dense-roster` | Nineteen role-category tiles with core/specialist separation. |
| synthetic-platform-lead-profile | `wm/biography` | Separate synthetic identity, role focus, working approach, work areas, and responsibilities. |
| measures-and-release-decision | `wm/workflow-matrix` | Evidence, owner, measure, and release-decision rows. |

These are candidate contracts. Selection is not qualification or a claim of preference.

## Earlier authoring attempt (superseded)

Built the public CLI after the repository owner refreshed contract metadata and rebuilt an index into ignored `samples/proposal-authoring/index-final2.sqlite` (13 contracts, 17,456 inventory items). Commands used:

```sh
go build -o /tmp/pptxlib-proposal ./cmd/pptxlib
/tmp/pptxlib-proposal index --out samples/proposal-authoring/index-final2.sqlite
/tmp/pptxlib-proposal assemble --index samples/proposal-authoring/index-final2.sqlite --config library/proposal/assembly.json --out samples/proposal-authoring/assembled --allow-unqualified
```

Whole-deck assembly stops at `five-phase-sequence`: required narrative sentence “Build and prove, weeks 5–10: implement the first journey, interface contract, access controls, telemetry, and test evidence in a nonproduction environment.” is not present verbatim in the contract's supplied slot copy. All slide assertion titles, roles, and takeaways were kept exact; no narrative was shortened to evade a binding.

For diagnostic isolation, one-slide configs were written in ignored `samples/proposal-authoring/` and run through the same public `assemble --allow-unqualified` command. These six isolated slides assembled: decision-and-thesis, situation-and-friction, needs-and-responses, journey-workflow, architecture-and-controls, measures-and-release-decision. The remaining seven fail exact required-detail/qualification validation:

- five-phase-sequence: first missing sentence is cited above; two 155-character phase statements are also longer than the largest single body slot cap of 152.
- discovery-detail: “Produce validated baseline and sampling notes; journey and ownership map; interface and dependency register; prioritized thin-release backlog; and agreed control and acceptance checklist.” is 187 characters. Although the objective slot has a larger individual cap (408), other required detail and qualifications consume the available semantic slot copy; current complete copy cannot satisfy all clauses.
- parallel-delivery-paths: 96-character architecture lead responsibility exceeds the largest single string slot cap (95). This contract's total non-binding string-slot caps sum to 264 characters while its required details plus qualifications total 807 characters.
- roadmap-and-gates: five required sentences are 111–174 characters, exceeding the largest single text-slot cap (95). The complete requirements total 944 characters; non-binding slot caps sum to 1,019 before any headings, context, or copy needed for the visual.
- team-and-accountability: the 21-role diagram caveat is 114 characters and the responsibility-model qualification is 85; current remaining slot capacity cannot hold all required statements verbatim.
- role-roster: two required caveats are 112 and 123 characters, while the largest single body slot cap is 95.
- synthetic-platform-lead-profile: the 111-character preparation statement and 101-character scope qualification are absent from current supplied slot copy; the full set totals 966 characters against 2,781 characters of nominal slot caps, but body copy plus contract-owned static slot text leaves less available room than raw caps suggest.

The phrases above are validation blockers, not proposed shortened copy. The contract authoring caps are editorial bounds; this check does not establish actual rendered capacity. A new semantic slot, a broader bounded content structure, or a narrative change would be needed where no existing slot can accept required prose. Per repository scope, this task did not alter contracts or the narrative to bypass those limits.

## Earlier native qualification status

Assembly only checks semantic values, narrative bindings, slot caps, exact required-detail/qualification presence, and writes generated specs/traces. No PowerPoint operation was run. The six assembled isolated slides are still `unmeasured_changed_copy` and require native probe/measurement, fit report, build, verification, negative stress, and visual review. They do not qualify the contracts. At that checkpoint the full 13-slide assembly was not possible under the exact-copy slot constraints listed above; the capacity revision below supersedes this state.


## Capacity revision — current candidate

The seven exact-copy blockers above are resolved in semantic assembly. Eight new
candidate variants are generated by `scripts/build-proposal-capacity-templates.py`
into `capacity-templates.json`; their contracts are now version 0.2.0. The earlier
Wave 1/2 fixtures and native proof remain unchanged. Their visual acceptance is
not inherited by these variants.

- **Five phases:** each actual phase owns its complete required activity sentence,
  timing, progression condition and evidence output. Week ranges agree with the
  narrative (pre-engagement, 1–4, 5–10, 11–14, 15–20). Qualifications have a separate
  bottom zone rather than displacing a phase output.
- **Discovery detail:** a full-width assertion introduces the phase; objective and
  complete proposed outputs have a sidebar; four activity/output pairs cover
  journey, baseline, systems and release definition. The original sample images
  retain their pinned assets and have a distinct, explicit attribution zone.
- **Parallel paths:** preserve two editable five-step lanes; case-state and release
  control prose sit below the correct path. Ownership and discovery boundaries
  occupy two dedicated panels in the previously unused left column.
- **Roadmap:** four native editable bars map to a true 20-week axis, replacing the
  unrelated month-based fixture. Mobilization, pilot and expansion gate prose
  stays separate from bar labels and the noncommitted-date qualification.
- **Team:** retain 21 role boxes, three pods and typed relationships; move each
  responsibility under its correct heading and give staffing/client readiness
  caveats separate zones. Structural routing caught a note crossing the architect
  route and the note was moved out of that corridor.
- **Roster:** retain all 19 tiles and the 10/9 category split. Role-category labels
  replace fictional person labels. Staffing agreement and non-additive-count
  language appear together outside the small tiles.
- **Architecture:** layer, governance and node copy now describes the correct responsibility. A dedicated qualification band preserves transaction-ownership, vendor-neutrality and client-review boundaries. The nine nodes, three layers and dependency connectors remain.
- **Needs/response:** five paired needs match the fictional brief; inherited RCM/provider-marketing labels are removed. Longer response frames and a separate qualification zone preserve all required clauses.
- **Biography:** existing rich-text slots hold all required detail and caveats;
  fixed run-boundary punctuation, removed a duplicated caveat from job duties,
  and changed an unsupported baseline-establishment claim to supporting validation.
- **Three matrix pages:** rewritten column and row meanings for situation, journey
  responsibilities, and release decisions. Required sentences no longer get
  appended to unrelated inherited template copy. Required detail was retained.

The public CLI now assembles the complete 13-page proposal. Root structural
preparation succeeds and synthetic-bound geometry tests exercise all eight changed
layouts, including connector clearance and package serialization. Those tests do
not measure actual text fit. An ignored AppKit authoring estimate helped identify
oversized inherited matrix copy; it is not PowerPoint evidence, skips rich text,
and cannot qualify a slide or substitute for the native gate.

Current artifacts and exact hashes are recorded in
`planning/WAVES_3_4_IMPLEMENTATION_CHECKPOINT.json`. Native access was recovered
by staging directly in the approved `samples/visual-wave3` folder. The complete
13-page candidate is natively rendered and visually reviewed. All 437 fixed text
zones pass native fit; final verification and expanded/stress qualification remain
pending. See `planning/NATIVE_PROPOSAL_REVIEW_2026-09-27.md`.

Reusable template corrections retain the content and font sizes: team text
frames enlarged by 2pt, response row backgrounds expanded with their text zones,
and the architecture parent surface moved behind its six connectors. Explicit
asset bindings replace mismatched response icons with vetted WM artwork. Asset
files remain local/ignored; the binding manifest records official download URLs
and exact hashes for new artwork. Existing inventory assets must be hydrated
before a fresh clone can build the proposal.
