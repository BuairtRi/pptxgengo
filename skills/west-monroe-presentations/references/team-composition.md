# Teams, pods and governance

Use this runbook to customize responsibility and reporting compositions. A
catalog diagram is a starting example: its people, pod count, leadership layers
and reporting structure are not requirements for the new deck. Keep the visual
language; author the relationships for the actual engagement.

These commands require a build containing `project team inspect/patch`. Check
`pptxgengo design project team inspect --help`; the promoted v4.2.1 does not
contain them. Use the [canonical project structure](project-structure.md).

## Choose the communication job

| Meaning to communicate | Use | Author explicitly |
| --- | --- | --- |
| Delivery units and responsibilities | Pods | Pod count, roles, assignments, shared leadership outside pods |
| Actual reporting relationships | Org chart | Root, parent of each role, solid/dotted reporting, organizations |
| Decision rights and escalation | Governance tiers | Tier order, members, cadence, decisions |
| Availability or staffing over time | Staffing curve or role-by-phase table | Counts, phases, allocation; this team command does not edit these models |

Separate a person from a role, and reporting from collaboration. A dotted org
chart line is a reporting qualifier in this renderer; do not use it to imply a
different collaboration graph without labeling that interpretation. A pod does
not require a program manager. Put shared sponsors/program leadership outside
pods when they serve multiple units; do not duplicate one person into every pod.
Use an org chart only when a tree conveys the actual meaning. For a sponsor on
the side of a pod composition, use a separate `role` component and explicitly
labeled connectors, rather than forcing that position into the org chart tree.

Do not invent names or reporting authority. Mark illustrative examples and
unknown roles. Log the catalog choice, composition changes and rationale in
`composition-log.yaml`; obtain approval when the change alters commitments or
authority beyond the operator's instruction.

## Inspect and select a supported route

```sh
pptxgengo design project detach --project ./deck --slide team \
  --as team-local --reason 'Adapt the pods to the delivery responsibilities'
pptxgengo design project team inspect --project ./deck --slide team
```

Inspection reports typed `pod`, `orgchart`, `governance` and `role` components,
authored arguments, resolved copy, stable array keys, component allocations,
the frame and measured text/native geometry. `bound_arguments: true` means copy
still resolves from slide `content`/`bindings` or `values`. Read
`resolved_arguments` to see the current actual copy. Record
`diagram.source_sha256` for the patch; it covers all authored source files.

Supported catalog examples include `team/pods` and `team/org-chart`; inspect the
selected variant's actual components. Some variants use loose blocks, text,
connectors or coupled tables and are not automatically converted to semantic
team models. On those, detach first and explicitly replace the selected
example pieces with typed components using `project diagram patch` or
`team patch`'s `add-component`; remove only the reviewed obsolete pieces.
For mixed responsibility tables, update and review the table separately.
The command does not synchronize a companion table or legend by inference.

Shared catalog templates remain immutable. Fork a local definition if multiple
slides use it. Revise a pinned local reference deliberately before changing it.
Native geometry/order overrides must first be reviewed and explicitly reset
with the [diagram workflow](architecture-geometry.md); team composition refuses
to silently overwrite them.

After changing composition, review the entire slide's title, captions, count
claims, legends and companion tables. A stock title describing three pods or
four roles becomes false if the adapted source has two pods and three roles.
Use `project edit --check-fit` for the reviewed narrative update and rebuild;
do not rewrite an earlier native baseline or reconciliation receipt. Semantic
adoption updates the reviewed membership, not every narrative claim by inference.

## Preview and apply one source transaction

```yaml
schema: pptxgengo.team-patch.v1
actor: deck-author
reason: Two unequal pods share quality responsibility without a program manager
expected_source_sha256: REPLACE_WITH_DIAGRAM_SOURCE_SHA256
operations:
  - action: materialize-component
    component: node01
  - action: materialize-component
    component: node02
  - action: add-role
    component: node01
    id: quality-lead
    role: {label: Quality lead}
  - action: remove-role
    component: node02
    id: REPLACE_WITH_AN_EXISTING_ROLE_KEY
  - action: reassign-role
    component: node01
    id: quality-lead
    target: node02
  - action: arrange-pods
    order: [node01, node02]
    layout:
      rect: {x_pt: 0, y_pt: 0, width_pt: 846, height_pt: 216}
      columns: 2
      gap_pt: 18
```

Replace component IDs, role keys, source hash and allocation using inspection;
they are examples, not universal catalog coordinates. Preview retains source
bytes. Apply repeats the complete validation and retains exact predecessors and
an actor/reason decision through the guarded source transaction:

```sh
pptxgengo design project team patch --project ./deck --slide team \
  --patch ./team-patch.yaml
pptxgengo design project team patch --project ./deck --slide team \
  --patch ./team-patch.yaml --apply
pptxgengo design project check --project ./deck
pptxgengo design project build --project ./deck
```

If source changes after inspection, refresh inspection and review a fresh patch.
Do not simply replace the hash on an outdated plan.

`materialize-component` is an explicit local conversion: resolves the selected
component's existing bindings into its local definition, then removes only now
unused binding zones through the source transaction. Copy becomes local
template constants; subsequent copy changes must edit that definition or be
deliberately rebound. Preserve the previous authored version and the rationale.
Scoped stock reporting keys that collide across branches receive unique report
keys during this conversion; inspect the preview before relying on the IDs.
Existing native layout overrides are refused so this transition cannot silently
invalidate adopted geometry. Never run this conversion merely to suppress a
binding error without understanding ownership of the copy.

## Operation contracts

Every action rejects irrelevant fields; do not pass fields expecting them to be
ignored. Entity IDs must be stable identifiers. Order lists name every entity
in that selected collection exactly once.

| Action | Fields beyond `action` | Behavior |
| --- | --- | --- |
| `materialize-component` | `component` | Explicitly resolve selected typed component bindings into local source |
| `add-component` | `id`, `node` | Add matching typed `pod`, `orgchart`, `governance` or `role` node |
| `remove-component` | `component`; optional `incident_edges: remove` | Remove one reviewed component; incident attached connectors otherwise block removal |
| `move-component` | `component`, `rect` | Set complete rectangle relative to the component's existing frame zone |
| `arrange-pods` | `order`, `layout: {rect, columns, gap_pt}` | Reflow selected same-zone rectangular pods in rows/columns; preserve role count and typography |
| `add-role`, `update-role` | `component`, `id`, `role: {label}` | Add/change one pod role |
| `remove-role` | `component`, `id` | Remove explicitly named role |
| `reorder-roles` | `component`, `order` | Carry role keys with their copy |
| `reassign-role` | `component`, `id`, `target` | Move one role to another pod, retaining its key; reject an existing target key collision |
| `rename-role` | `component`, `id`, `target` | Deliberately establish a new unused role key, useful before moving scoped stock slot keys |
| `add-report` | `component`, `id`, `parent`, `report` | Add an org-chart branch with matching `report.key` |
| `update-report` | `component`, `id`, `report` | Replace title/name/org/dotted flag; omit children to retain topology |
| `remove-report` | `component`, `id` | Remove a leaf; root and implicit subtree deletion refused |
| `reparent-report` | `component`, `id`, `parent` | Move a reporting subtree; cycles/root moves refused |
| `reorder-reports` | `component`, `parent`, `order` | Reorder the chosen parent's children |
| `add-tier` | `component`, `id`, `tier` | Add matching `tier.key` with name/cadence/members/decisions |
| `update-tier` | `component`, `id`, `tier` | Change matching tier name/cadence; omit members/decisions to retain their identities |
| `remove-tier` | `component`, `id` | Explicitly remove that tier and its contents |
| `reorder-tiers` | `component`, `order` | Move tier contents and their keys together |
| `add-member`, `update-member` | `component`, `parent`, `id`, `member: {label, org}` | Edit member in tier `parent`; organization is `wm`, `client` or `tbd` |
| `add-decision`, `update-decision` | `component`, `parent`, `id`, `decision` | Edit plain decision copy in tier `parent` |
| `remove-member`, `remove-decision` | `component`, `parent`, `id` | Explicitly remove selected item |
| `reorder-members`, `reorder-decisions` | `component`, `parent`, `order` | Retain keys and reorder selected tier collection |

`report` fields: `key`, `org`, `title`, `name`, optional `dotted` and `children`
(children only for an added branch). `tier.members` is an array of
`[label, organization]` pairs; `tier.decisions` is an array of bullet copy.
New collection keys are assigned once and persisted; inspect them before the
next operation. Changing a tier's order also changes its positional shading and
escalation direction. Review that semantic change instead of treating it as a
cosmetic reorder.

A component added by `add-component` uses the normal source node contract:

```yaml
action: add-component
id: shared-sponsor
node:
  id: shared-sponsor
  kind: component
  definition: {scope: shared, id: wmds/component/role}
  placement:
    zone: body
    rect: {x_pt: 666, y_pt: 0, width_pt: 180, height_pt: 72}
  arguments:
    title: Executive sponsor
    meta: Client leadership
    surface: subtle
    edge: series.4
```

Reserve a separate sponsor allocation when arranging pods. Add connectors using
the [geometry runbook](architecture-geometry.md) only when they express actual
reporting/collaboration and have a legend or label that makes that meaning clear.
The command does not invent reporting edges from adjacent positions.

## Fit and review

- Pod rows use the renderer's 42 pt role budget and `48 + 42 × role count`
  envelope. `arrange-pods` uses the tallest pod in each row, preserves fonts and
  rejects insufficient allocation. It does not stretch surrounding components.
- Org-chart width grows with siblings and depth; widen allocation, choose a
  different topology or split slides when the renderer rejects capacity.
  Org charts retain a top-root tree layout; they do not offer arbitrary graph
  routing or side-root layout.
- Governance tiers use the existing 90 pt cadence and measured member/decision
  regions. Extra tiers, long names, or decisions can overflow. Resolve capacity
  explicitly; do not shrink selected text or omit authority to make it fit.
- Preview's measured allocations are Go fit evidence. Review neighboring
  components, frame reservations, organization colors, count, labels and native
  PowerPoint renders. Allocation warnings are not proof of exact ink clearance.

## Native edits and imports

Keep immutable builds and edit a separate PowerPoint copy. Use receipt-backed
reconciliation to propose supported uniquely mapped text, transforms/order,
whole deletions and explicitly mapped supported copies. Review proposals before
adopting and rebuild afterward. See [editing slides](editing-slides.md) and
[PowerPoint recovery](powerpoint-recovery.md).

Moving a role next to another pod does **not** establish membership. Moving an
org-chart node does not establish a different reporting parent. A native copied
box is not automatically a new role or governance member. Keep those edits
visible and use explicit team operations to persist their meaning. Arbitrary
native objects/SmartArt/imported charts and rich-format semantic imports are
not supported by this slice. Do not describe generic geometry qualification as
a passed team-family semantic round trip.


### When native movement might change meaning

Review the operator's intent before selecting `reassign-role` or
`reparent-report`. A role moved entirely inside another pod can be evidence for
an explicit membership question, but spacing, overlapping containers and shared
leadership make proximity ambiguous. Use `project team reconcile` with a closed
geometry review packet to propose an original role's complete translation into
exactly one unchanged pod. An explicit `reassign` decision confirms membership;
partial movement, resizing, overlap and stale source remain unresolved. Follow
[native semantic review](native-semantic-review.md) for decisions, measured
preview and guarded application. Rebuild from the confirmed membership and
preserve the original geometry as evidence rather than applying the move twice.

For reporting, a verified connector endpoint reassignment would be stronger
evidence than placing one node beneath another. Connector topology is not
currently reconciled into the org-chart tree. Do not infer a reporting change
from arrow direction, relative position, or an imported untagged line. Confirm
role identities/parent and use `reparent-report`; preserve cycle/root guards.
The bounded [Gantt semantic review](gantt-composition.md#review-native-bars-as-possible-schedule-changes)
handles period-based task/gate proposals separately from these membership
contracts.
