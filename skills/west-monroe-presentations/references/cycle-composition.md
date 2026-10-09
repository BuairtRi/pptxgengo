# Compose lifecycle cycles and feedback loops

Use a cycle when the primary message is an ordered, repeating set of stages, with
optional named feedback paths. The stock lifecycle stages are examples: keep the
visual language while authoring the stages and feedback relationships the operator
needs. For independent branching decisions or arbitrary joins, use a different
form; this command does not compose general flowcharts or forked roadmaps.

These commands require a CLI whose installed `design project cycle --help`
includes `inspect` and `patch`. Check the project's toolchain pins before using a
newer binary; follow [Project upgrades](upgrading-projects.md) for a deliberate
upgrade. The promoted v4.2.1 release predates these commands.

## Choose the meaning before the geometry

- `steps` have stable keys and authored labels/detail. Their order defines the
  primary path around the ring. Reordering changes sequence, not just spacing.
- `loops` are extra directed feedback paths. Each has its own stable key and
  explicit `from`/`to` step keys. Spatial proximity does not create a relationship.
- `active` identifies the current stage by key. Keep this separate from a stage's
  explicit `surface` styling: both can affect emphasis, but only `active` records
  the current stage.
- `closed: false` omits the last primary connection to the first stage. `ring:
  false` hides primary ring paths entirely; explicit feedback loops remain.
- `center` provides optional context, not another stage or relationship endpoint.
- `n` is a displayed number/string override. `numbered` supplies ordinal labels
  when no override is present. If numbering has business meaning, update it
  explicitly rather than assuming a reorder changes that meaning.

The renderer accepts 2–12 stages and up to 32 feedback loops. Those are source
limits, not a fit promise. Label lengths, stage dimensions, ring size, feedback
paths and the chosen frame still need measured preview and native visual review.

## Inspect the local composition

```sh
pptxgengo design project cycle inspect --project PATH --slide lifecycle --node cycle
```

Start with a catalog template such as `phase/define`, `sdlc/cycle` or
`pdlc/loop-backs`. Use [Editing slides](editing-slides.md) to detach the
selected slide into a project-owned local template. Inspect the resulting nodes;
use the actual typed cycle's stable node ID, not its visible title.

Inspect returns the current project source hash, keyed stages, keyed feedback
loops, current marker, center, layout controls and measured diagram. It reports
bound arguments through `mutation_blocked`. A renderer failure can still return
source semantics, accompanied by `render_error` and no measurement qualification.

A template made from unrelated shapes is not automatically a typed cycle. If
`inspect` refuses the node, retain its source and use an explicit reviewed local
derivative or the applicable diagram workflow.

## Preview and apply one guarded transaction

Create a YAML or JSON patch using the exact `source_sha256` from inspection. This
example adds a stage, connects its feedback path and makes it current. Replace
all sample IDs with keys from your own inspection; the `source-001` identities
below are illustrative.

```yaml
schema: pptxgengo.cycle-patch.v1
expected_source_sha256: REPLACE_WITH_INSPECTION_HASH
actor: Presentation operator
reason: Add monitoring to the illustrative lifecycle
node_id: cycle
operations:
  - action: materialize
    entity: source
  - action: set
    entity: step
    key: monitor
    step:
      key: monitor
      label: Monitor
  - action: set
    entity: loop
    key: monitoring-feedback
    loop:
      key: monitoring-feedback
      from: monitor
      to: source-002
  - action: reorder
    entity: step
    order: [source-001, source-002, source-003, monitor, source-004]
  - action: set
    entity: active
    key: monitor
```

```sh
pptxgengo design project cycle patch --project PATH --slide lifecycle --patch cycle.yaml
pptxgengo design project cycle patch --project PATH --slide lifecycle --patch cycle.yaml --apply
```

The first command measures the candidate and leaves authored files unchanged.
Inspect its warnings, paths and frame allocations; a measured candidate is not a
PowerPoint visual review. `--apply` uses the same source hash guard, writes a
review decision and retains predecessors. Reinspect after applying before making
another patch; the old hash is stale.

`materialize/source` must be first when arguments contain bindings. It explicitly
retains this component's resolved copy in its local template; bindings used by
other nodes remain intact. Omit it when editing an already materialized component.
To keep bound ownership, edit the authored values instead of materializing.
Pinned/shared local templates require explicit revision/forking, and native layout
overrides require preservation and review before an explicit reset. The command
refuses to silently overwrite them.

## Supported operations

| Action/entity | Required values and effect |
| --- | --- |
| `set/step` | `key`, complete `step` with matching key and label; append a new stage or replace the existing stage |
| `remove/step` | `key`; incident feedback loops/current marker must be removed or reassigned first, or explicitly acknowledge `cascade: true` |
| `reorder/step` | `order` lists every current stage key exactly once |
| `set/loop` | `key`, complete `loop` with matching key and explicit `from`/`to`; append or replace |
| `remove/loop` | Existing `key` |
| `reorder/loop` | `order` lists every loop key exactly once; controls emitted loop order |
| `set/active` | Existing stage `key` |
| `remove/active` | Clears the current stage; explicit stage surfaces remain |
| `set/center` | Complete `center` with optional `label`, `title`, `text`, `w` in points |
| `remove/center` | Removes central context |
| `set/layout` | Complete `layout`; optional `rect`, `node_width_pt`, `node_height_pt`, `start_degrees`, `closed`, `ring`, `align`, and `numbered` |

`set` replaces the complete record; omitted fields do not preserve previous copy,
number overrides or styling. For layout, omitted `rect` retains the component's
existing allocation; other omitted controls return to renderer defaults. A rect
uses `{x, y, w, h}` in points relative to the current placement zone. Layout changes
move/reflow the whole cycle, not an arbitrary set of independent stage positions.
Keep the frame/rail/footer and unrelated components unchanged.

Optional step fields are `text`, `n`, `surface`. Optional loop fields are `label`,
`side` (-1 or 1), `bend` (0–2), and `ink` (a design-system color). Self-loops and
unknown endpoints are refused. Unknown fields and irrelevant fields are refused,
including authored `false`, zero, empty and null fields on another operation.
The patch limit is 1 MiB, 1–500 operations, one document, with required actor and
reason. Relationships are validated against the final candidate, allowing explicit
multi-operation changes in one patch.

## Counts, fit and native editing

For a four-stage catalog cycle, first preview adding a fifth stage and a keyed
feedback loop. Reorder and confirm the loop endpoints/current stage keep their
identities. For a three-stage variant, remove a stage with explicit treatment of
its incident loops and marker. Also try longer labels before selecting a density
or smaller allocation: text may not fit even though the stage count is allowed.

If a candidate overflows, preserve the rejection and choose a larger allocation,
a different form, shorter approved copy or another slide. Do not silently shrink
individual fonts or remove a relationship. Cycles with many labels can be less
readable than a short sequence plus a separate feedback explanation.

Build and render the accepted source following [Slide composition](slide-composition.md),
then review text, arrows, emphasis and frame clearance. Keep builds immutable and
make a separate PowerPoint working copy. Receipt-backed supported geometry/text
changes follow [Editing slides](editing-slides.md). Moving a stage in PowerPoint
is a native layout edit; it does not reorder stages, retarget loops or change the
current marker. Adopt those semantics explicitly through the cycle patch.
