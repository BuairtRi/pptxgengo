# Review native edits for source meaning

Keep generated builds immutable. Edit a separate copy in PowerPoint, Save As a
new `.pptx` under the deck project or an owned Documents staging folder, and
retain the exact file. Use the project’s installed CLI and locked bundle.

## Create authenticated evidence

```sh
pptxgengo design project reconcile propose --geometry --project ./deck-project \
  --edited ./working/edited.pptx --out ./reviews/native-01
```

This compares original ownership tags and the pinned build receipt; it does not
assign meaning to arbitrary imported objects. Fix source/baseline conflicts
before semantic adoption. Materialize bound arguments through the family command
and rebuild a fresh baseline before editing. Native layout overrides need an
explicit reset or separate review before a source composition change.

## Select the family interpretation

| Native edit | Proposal command | Accepted meaning |
| --- | --- | --- |
| Original role surface and text moved together fully into exactly one unchanged pod allocation | `project team reconcile --project ./deck-project --slide ID --packet ./reviews/native-01` | Explicit `reassign` membership decision; position does not establish reporting |
| Visible ordinal assessment score cell changed | `project assessment reconcile --project ./deck-project --slide ID --node NODE --packet ./reviews/native-01` | Explicit `set_score`; canonical integer inside declared domain, or blank meaning unassessed; zero remains a score |
| Native chart numeric data edited | `project quantitative reconcile --project ./deck-project --slide ID --node NODE --packet ./reviews/native-01` | Explicit `set_value` after chart cache and embedded workbook agree on the keyed fact |
| Tagged Gantt task or gate retimed | `project gantt reconcile --project ./deck-project --slide ID --node NODE --packet ./reviews/native-01` | Explicit `retime` in authored period space; see the Gantt runbook |

Prefix table commands with `pptxgengo design`. Read `report`, its
`report_sha256`, each proposal’s status and reason, and all unresolved IDs.
One changed numeric cell does not authorize adopting altered axes, labels,
units, scales, formulas or formatting. A copied/reparented object or ambiguous
identity requires explicit source interpretation. A commercial output is a
calculated view: change the model’s inputs/formulas, not its displayed result.

## Decide, preview and apply

Create a YAML/JSON decisions file naming the report’s exact hash, the actor,
reason, and selected proposal IDs. For example, team membership:

```yaml
schema: pptxgengo.team-semantic-decisions.v1
report_sha256: REPLACE_WITH_REPORT_HASH
actor: named operator
reason: Confirm the reviewed membership change
decisions:
  - proposal_id: REPLACE_WITH_PROPOSAL_ID
    action: reassign
    reason: This role belongs to the target pod
```

Assessment and quantitative schemas are respectively
`pptxgengo.assessment-semantic-decisions.v1` and
`pptxgengo.quantitative-semantic-decisions.v1`. Use the matching action from the
table, or `keep_source`. Pass `--decisions FILE` to the same command for a
measured preview; add `--apply` only after reviewing it. Unsupported proposals
cannot be forced into supported meaning by changing their decision action.

Rebuild and render after adoption. Assessments regenerate fills and legends;
charts regenerate the source numeric data; pods regenerate membership layout.
The source transaction retains the original PPTX, reports and exact decisions
under shared content-addressed assets. Other native edits remain retained and
unresolved. A partial adoption is not complete synchronization. The previous
packet becomes stale after the source changes; begin subsequent reviews from
the new build.

See [geometry](architecture-geometry.md), [teams](team-composition.md),
[assessments](assessment-composition.md), [charts](quantitative-charts.md), and
[Gantt](gantt-composition.md) for family-specific source operations.
