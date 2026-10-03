# Integrated Approach, Openers and Commercials specimens — slice 4

This packet covers all 56 active v2 designs across Approach (27), Commercials (7), and Openers (22). Each has two explicit synthetic bound-content specimens, for 112 slides. No deprecated key belongs to these three families; `from-to/rows` belongs to Argument and is handled by that family’s integrator.

## Binding and content

All 23 added-design alternates reuse the reviewed slice 3 caller values. Equivalent non-nav designs reuse identical reviewed body-slot contracts where their fixed source content schema matches; navigation is removed only when the selected design has no nav contract. The two card-row designs reuse the reviewed typed source/alternate pairs. Every other design has explicit per-key caller values in the new generator; no generic string replacement, truncation, scaling or font shrinking is performed.

Every pair is classified `synthetic_example`. Illustrative prices, close durations, value cases, quotes and contacts are specimen assumptions rather than client facts. Alternate content changes at least one substantive body slot for every key; changing only headers is rejected by the generator. Headers, array keys and caller nav may also change.

The generic bindings remain exact `values.slots` and `values.keys` contracts; the card-row designs retain their typed APIs. Numeric allocations, states, RACI values, checkbox cells and chart coordinates retain their existing validators. Fixed source geometry, font sizes and source text are preserved.

## Thumbnail composition

The five generic thumbnail surfaces in `phase-detail/rail-cohort` and `phase-detail/rail-table` receive explicit native platform-flow or release-plan miniatures, using the same specimen approach as the reviewed prior reference. The blank surface is followed by editable text/rectangles. It is an authored reference composition, not fallback content inserted into the closed template contract. All miniature geometry stays inside its authored thumbnail with 6 pt padding. Captions and caller binding provenance remain intact.

The miniature resolution is `wmds.integrated-approach-illustrative-miniature.v2`. Pair directories retain the closed authoring input, binding report, bound compiled document, and composed `combined.foundation.json`; the global combined file contains the composed examples.

## Coverage

| Template | Alternate origin | Substantive changed fields |
|---|---|---:|
| `agenda/index` | explicit_per_template_alternate | 4 |
| `agenda/schedule` | reviewed_equivalent:agenda/schedule-nav | 4 |
| `agenda/schedule-nav` | reviewed_slice3_alternate | 4 |
| `agenda/schedule-split` | reviewed_slice3_alternate | 4 |
| `agenda/sessions` | explicit_per_template_alternate | 6 |
| `cards/3` | previously_reviewed_typed_card_pair | 9 |
| `cards/4` | previously_reviewed_typed_card_pair | 11 |
| `closing/tagline` | explicit_per_template_alternate | 3 |
| `cover/grid` | explicit_per_template_alternate | 4 |
| `cover/photo` | explicit_per_template_alternate | 7 |
| `cutover/wave-matrix` | explicit_per_template_alternate | 4 |
| `divider/full-photo` | explicit_per_template_alternate | 2 |
| `divider/inverse` | explicit_per_template_alternate | 2 |
| `divider/light` | explicit_per_template_alternate | 3 |
| `divider/panel-edge` | explicit_per_template_alternate | 2 |
| `divider/panel-photo` | explicit_per_template_alternate | 2 |
| `key-message/stat` | reviewed_equivalent:key-message/stat-nav | 4 |
| `key-message/stat-nav` | reviewed_slice3_alternate | 4 |
| `key-message/stat-split` | reviewed_slice3_alternate | 3 |
| `key-message/statement` | explicit_per_template_alternate | 3 |
| `narrative/lead-in` | explicit_per_template_alternate | 4 |
| `phase-detail/rail-cohort` | explicit_per_template_alternate | 6 |
| `phase-detail/rail-matrix` | explicit_per_template_alternate | 4 |
| `phase-detail/rail-table` | explicit_per_template_alternate | 6 |
| `phase-detail/team-handoff` | explicit_per_template_alternate | 6 |
| `phase-gate/evidence-matrix` | explicit_per_template_alternate | 5 |
| `phases/activities-outcomes` | explicit_per_template_alternate | 5 |
| `phases/four` | explicit_per_template_alternate | 4 |
| `phases/summary-bands` | explicit_per_template_alternate | 5 |
| `plan/gantt` | explicit_per_template_alternate | 5 |
| `pricing/capacity` | explicit_per_template_alternate | 5 |
| `pricing/capacity-split` | reviewed_slice3_alternate | 6 |
| `pricing/fixed-fee` | explicit_per_template_alternate | 13 |
| `pricing/options` | reviewed_equivalent:pricing/options-nav | 18 |
| `pricing/options-nav` | reviewed_slice3_alternate | 18 |
| `quote/inverse` | explicit_per_template_alternate | 3 |
| `quote/light` | explicit_per_template_alternate | 3 |
| `roadmap/staggered-phases` | reviewed_equivalent:roadmap/staggered-phases-nav | 10 |
| `roadmap/staggered-phases-nav` | reviewed_slice3_alternate | 10 |
| `runbook/cutover-timeline` | reviewed_slice3_alternate | 11 |
| `runbook/escalation` | reviewed_slice3_alternate | 2 |
| `runbook/escalation-flow` | reviewed_slice3_alternate | 1 |
| `runbook/escalation-split` | reviewed_slice3_alternate | 1 |
| `runbook/go-no-go` | reviewed_slice3_alternate | 20 |
| `runbook/overview` | reviewed_slice3_alternate | 18 |
| `runbook/roles` | reviewed_slice3_alternate | 19 |
| `runbook/roles-raci` | reviewed_slice3_alternate | 4 |
| `runbook/step-detail` | reviewed_slice3_alternate | 16 |
| `runbook/step-detail-data` | reviewed_slice3_alternate | 22 |
| `runbook/step-detail-gate` | reviewed_slice3_alternate | 21 |
| `runbook/step-list` | reviewed_slice3_alternate | 21 |
| `runbook/step-list-split` | reviewed_slice3_alternate | 21 |
| `runbook/step-table` | reviewed_slice3_alternate | 18 |
| `runbook/step-table-nav` | reviewed_slice3_alternate | 18 |
| `scope/service-value` | reviewed_equivalent:scope/service-value-nav | 22 |
| `scope/service-value-nav` | reviewed_slice3_alternate | 22 |

Exact old/new slot values are recorded in `changed-slot-receipts.json`; array keys are explicit in `bound-content.json` and binding reports.

## Generation and native review

Initial focused source sweep: 56/56 generated. Final bound pairs: 56/56 generated. Combined: 112 slides generated. No shared renderer changes or source-specific Go amendments were required in this slice. Existing slice 2/3 refinements apply through the current renderer.

Rebuild from the repository root:

```sh
python3 scripts/wmds-integrated-approach-openers-commercials.py --cli ./pptxdesign
```

Outputs are under `samples/wmds-refresh-slice4-20261002/work/approach-openers-commercials`. All 112 integrated native pages have been visually reviewed using contact sheets, with 18 focused full-resolution checks. The source and alternate inverse dividers retain the three grid step photo layering, and the thumbnail compositions, RACI badges and split runbook headings are clean.

One alternate specimen required a copy correction: `phase-gate/evidence-matrix` (integrated page 52). Four evidence lists gained one wrapped line, leaving about 3 pt before their next labels. These four synthetic bullets now use equivalent concise wording, recorded exactly in `changed-slot-receipts.json`; the source copy, geometry and fonts remain fixed. The 96 pt evidence band accommodates four single-line bullets with normal spacing. Longer lists must reserve extra vertical space explicitly or use a different design. Rebuilt generation and corrected native page 52 review are complete: full-resolution review confirms one-line bullets and normal label separation. All other family PNGs are byte-identical to the initially reviewed export. `native-review.json` records all 112 accepted paired specimens, the resolved finding history and exact final evidence hashes.

Successful generation alone does not qualify arbitrary caller content; no tests or negative controls were added or run.
