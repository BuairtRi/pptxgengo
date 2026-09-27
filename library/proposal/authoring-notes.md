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

## Validation performed

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

## Native qualification not performed

Assembly only checks semantic values, narrative bindings, slot caps, exact required-detail/qualification presence, and writes generated specs/traces. No PowerPoint operation was run. The six assembled isolated slides are still `unmeasured_changed_copy` and require native probe/measurement, fit report, build, verification, negative stress, and visual review. They do not qualify the contracts. The full 13-slide assembly is not currently possible under the exact-copy slot constraints listed above.
