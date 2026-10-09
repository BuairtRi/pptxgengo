# Edit native tables without losing cell meaning

Use `project table` for a local typed `table` or `editable-table`. It keeps stable
row identities, source column keys, cell values, formatting domains and reserved
row metadata together. Use commercial, assessment or quantitative commands when
calculations or explicit scoring/measurement models are intended.

```sh
pptxgengo design project table inspect --project PROJECT --slide SLIDE --node TABLE
pptxgengo design project table patch --project PROJECT --slide SLIDE --patch change.yaml
pptxgengo design project table patch --project PROJECT --slide SLIDE --patch change.yaml --apply
```

Inspection returns the source hash, complete column definitions, keyed rows,
reserved metadata, keyed group memberships and source style arguments. It also
returns measured fit and render errors. Preview before guarded apply.

`inline_sections` exposes full-width source `group` header keys, labels and member
row keys. It differs from the side-label `row_groups` spans. Reorder members
within their original section, or move a complete header-and-members block with
one complete row-order list. A flat reverse that leaves headers after their
members is refused. To deliberately reassign an ordinary row, use `move/row`
with its `key` and target `section` header key; record that changed association in
`semantic_review`. New ordinary rows in these tables also require `section`.
Removing a populated inline header is refused even with cascade: remove its
members first, or explicitly move each retained member to another declared
section, then remove the empty header. Do not silently erase group meaning.

`sparse_leading_labels` names rows whose first text-column label is blank while
other rows have labels. Blanks, colors and proximity do not declare phase
membership. A roadmap with `Phase` shown only at the first row of each phase
must retain row order, or explicitly set each row's phase from authoritative
source before reordering. The CLI refuses order changes until that ambiguity is
resolved; it does not fill labels down or infer the intended business grouping.

## Row and cell edits

Use `set/row` with a stable row key and `row` values. Existing rows merge supplied
cells/metadata, preserving unspecified values. New rows require an explicit value
for every column: use `null` where the declared cell type supports missing data.
Full-width header rows instead declare `row: {group: LABEL}`; they are not
ordinary rows with missing cells. Their label and stable header key define the
explicit inline section that follows.
`set/cell` addresses a row `key`, source `column` key and explicit `value`.
Zero is a value; null is deliberately missing, not implicitly zero.

```yaml
schema: pptxgengo.table-patch.v1
expected_source_sha256: COPY_FROM_INSPECT
actor: Operator
reason: Add a validation case
semantic_review: Status labels and evidence references remain consistent; no totals inferred
node_id: validation
operations:
  - action: set
    entity: row
    key: access
    row: {case: Access review, owner: Operations, status: notstarted}
  - action: set
    entity: cell
    key: access
    column: status
    value: {status: progress, label: Review underway}
```

Keep status colors tied to their canonical status meaning: column `labels` only
changes visible labels. Source status keys are `on`, `risk`, `off`, `done`,
`progress`, `notstarted`, `blocked`, `pass`, and `fail`. Preserve explicit heat
min/max/scale and score labels. Table source patches validate retained heat values against explicit column bounds (falling back to global bounds, then 0..4). Continuous values inside the range are supported; missing stays missing. Typed rendering additionally validates cell options and measured fit.

## Columns and ordering

`set/column` supplies a complete `column_value` with matching `k`, `label`, `w`
and type/domain options. A new column also requires `cells`, mapping every stable
row key to an explicit value. Existing column replacement retains cells; update
cells separately if their domain changes. Reordering rows/columns includes all
keys once and keeps values associated with their column and row identities.

Styled `rowHeader` refers to the first column—even if the old source uses a
string. Reordering must retain that column first; deleting it also needs explicit header reassignment. To deliberately reassign it,
explicitly disable `rowHeader`, reorder, then enable it again. Do not let styling
silently change the row-label meaning.

Deleting a row or column requires `cascade: true`, acknowledging its cells,
metadata and group relationships. This includes zero-valued cells. Removed
highlight columns clear the coupled highlight. Unknown cell keys are refused.
Existing metadata `group`, `total`, `ink`, `scale`, and `h` is retained; a new
column using those reserved names is refused to avoid changing metadata meaning.
Legacy source columns with those names retain their original interpretation.

## Merged and side groups

Inspection exposes `row_groups` and `column_groups` with stable keys, labels and
`members`. Source inclusive numeric spans are rewritten after reorder. Members
must remain contiguous and groups must not overlap. If a reorder splits a group,
explicitly revise its membership first rather than trusting a stale span.
`set/row_group` or `set/column_group` supplies a complete `group`; `remove` and
`reorder` address those keys. Row-group `fill` is a color reference, not a surface
name. Deleting the last member with cascade removes its group.

`set/layout` accepts `rect` and `options`: header, dense, rowH, rowHeader, preset,
highlight, continued, deltaUnit, heatMax/heatMin, groupW and runRate. Keep column
width totals coordinated with the actual table allocation and any side group
rail. Add room or split a table when new rows/columns no longer fit.

## Source and evidence boundaries

Patches require `semantic_review` explaining relationships to totals, legends,
criteria, owners, citations and other slides. Generic table changes do not
calculate or infer those facts. Native table cells remain editable, but raw Office
cell moves/copies do not automatically become row/column semantic patches. Use
reviewed reconciliation evidence and persist the intended facts explicitly.
Bound source needs first `materialize/source` or changes to its original values.
Predecessors, frame allocations, native overrides and template pins are guarded.
