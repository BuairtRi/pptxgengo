# Component and table composition scope

The v4.3.0 operator adapters cover authored local typed content/layout and table
components. They are source transactions, not arbitrary Office imports. Generic
operations require a semantic-review statement; specialized family commands own
membership, dates, quantities, dependency graphs, scoring and computed values.

## Implemented contracts

`pptxgengo.component-patch.v1` inspects scalar arguments and recursively keyed
collections. Array addressing uses stable `@key` selectors. Adding, removing and
reordering retains identity overlays for nested members. Numeric array selectors
are refused. Geometry uses a separate measured layout operation; reserved renderer
geometry metadata is retained, not directly editable. Bindings must be edited at
their source or explicitly materialized before replacement.

Semantic wrappers add keyed phase current markers, Venn set/region memberships,
and matrix column/cell mapping. Venn region additions require named members;
set reorder rewrites source positional membership and deletion cascades incident
regions. Matrix column additions require an explicit value for every row; reorder
keeps cells attached to their columns. These operations do not derive new facts.

`pptxgengo.table-patch.v1` retains source column keys, typed definitions, keyed
rows, all cells and reserved row metadata. New rows/columns require explicit cells,
including null for deliberately missing observations. Row and column groups use
named members; source numeric spans are recomputed and must remain contiguous.
Deletion requires cascade acknowledgement. Styled first-column row headers cannot
silently move to another column. Heat domain changes validate every retained value
against explicit column/global bounds; the legacy renderer's visual clamping does
not silently authorize a new out-of-domain source value. Continuous values inside
the declared heat range remain valid.

Both adapters use immutable measured previews followed by source-hash guarded
apply. They preserve template pins, authored frames, predecessors and supported
native overrides through the shared composition transaction.

## Family applicability

The operator [communication family runbooks](../skills/west-monroe-presentations/references/communication-family-runbooks.md)
map all 29 catalog family jobs to their applicable models. Tags overlap: a
commercial family can contain explicit textual assumptions as well as a computed
fee model; a lifecycle family can contain a repeating cycle, independent phase
list or responsibility matrix. Inspect actual source before selecting a command.

The source catalog test retains each complete source frame and companion nodes,
then exercises a selected component/table from one actual example per family.
It is deliberately a nightly/on-demand test. Specialized adapters have their own
variable-count, reorder, deletion, stale-source and fit tests. This does not claim
that every catalog variant has passed interactive PowerPoint qualification.
Runbooks/team-curves specimens containing distribution-excluded decorative assets
must use the explicit template-placeholder policy, recording replaced asset names;
required authored content is never silently removed.

See the [named 29-family input and portability ledger](component-table-family-qualification-20261009.md) for exact source identities and results.

## Evidence and limits

- `component_composition_test.go`: nested identity, insertion/reorder/removal,
  immutable preview, explicit null, strict fields, stale and overflow refusal.
- `component_semantics_test.go`: phase current, Venn membership, matrix cell meaning.
- `table_composition_test.go`: cells/null/zero, row metadata, groups, domains,
  variable rows/columns, nested cell identity and first-column label meaning.
- `component_catalog_test.go`: actual full-source representative for all families;
  optional private JSON receipts identify exact input and selected component.
- `process_journey_qualification_test.go`: explicit private source exports for
  subsequent native qualification. Source generation is not a native-save result.

Unknown formatting/native objects remain retained pending explicit adoption.
Raw cell moves, copied cells and geometry never infer dates, reporting, approval,
unit changes, calculations or source membership. Windows PowerPoint qualification
is outside this release's qualification scope.
