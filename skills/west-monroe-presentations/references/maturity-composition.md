# Compose maturity models

Use maturity templates as examples of an ordered qualitative progression. A curve
is illustrative unless the author supplies measured observations. Its steepness
is not an assessment score, date forecast, or probability.

## Choose and inspect

Inspect installed `project maturity` help and the project lock before using these
commands. Detach a catalog example, fork a shared local template for this slide,
then inspect the typed `wmds/component/maturity` node:

```sh
pptxgengo design project diagram inspect --project PROJECT --slide SLIDE
pptxgengo design project maturity inspect --project PROJECT --slide SLIDE --node CURVE
```

Inspection supplies the exact source hash, stable stage keys, normalized horizontal
positions, current stage, here/target/inflection markers, optional branch, shape
intensity, and allocated rectangle. It reports bound arguments and render failures;
inspection success alone is not a passed render.

## Design the progression

- Decide the meaning and count of stages before selecting a stock example.
- Keep stage identity stable when wording or order changes. Position `at` is a
  normalized coordinate in `0..1`, not a measured score. Positions must increase.
- Use 3/5/6 stages where they suit the message; 1..12 are representable, but the
  measured label fit determines whether a particular allocation is usable.
- `shape` is an exponential bend intensity, greater than zero and at most 100.
  A smaller number is flatter; increasing it concentrates ascent at the right.
- Current is an observed authored stage. Target is explicitly intended future
  state. Here and inflection have their own stage references and labels. A branch
  names its departure stage and describes a possible alternative; it is not a
  probabilistic forecast.
- Target copy is measured and placed below the actual curve across its label
  width, with six-point clearance. If it cannot fit above the axis inside the
  allocation, preview refuses it: shorten the label or explicitly enlarge the
  component rather than accepting an intersecting target.
- Preserve qualitative navy/blue/magenta marker roles. Do not encode an unapproved
  assessment scale through geometry or color.

## Guarded patch

Use schema `pptxgengo.maturity-patch.v1`, `expected_source_sha256`, `actor`,
`reason`, `node_id`, and `operations`. Preview before applying:

```sh
pptxgengo design project maturity patch --project PROJECT --slide SLIDE --patch change.yaml
pptxgengo design project maturity patch --project PROJECT --slide SLIDE --patch change.yaml --apply
```

```yaml
schema: pptxgengo.maturity-patch.v1
expected_source_sha256: COPY_FROM_INSPECTION
actor: presentation-operator
reason: Add a future stage and preserve the current observation
node_id: maturity
operations:
  - action: materialize
    entity: source
  - action: set
    entity: stage
    key: future
    stage: {key: future, label: Future state, text: Shared practice scales, at: 0.98}
  - action: reorder
    entity: stage
    order: [stage-one, stage-two, stage-three, future]
    spacing: even
  - action: set
    entity: target
    marker: {stage: future, label: Target state}
  - action: set
    entity: layout
    layout: {shape: 2.0, label_width_pt: 140}
```

Replace example keys with inspected keys. `set/stage` supplies a complete stage,
including `at`; adding a stage does not interpolate source meaning. Reordering
requires `spacing: even` or `curve`; both redistribute normalized positions in
`.04.. .90` and preserve marker references by key. `set/spacing` explicitly
redistributes existing stages. `remove/stage` requires `cascade: true` when a
current/marker/branch references that stage, or reassign those references first.

Other operations: `set/remove current`, `set/remove here`, `set/remove target`,
`set/remove inflection`, `set/remove branch`, and `set/layout`. Current uses `key`;
markers use `{stage,label}`; branch uses `{from,label,text,n}`. Layout supports
`rect` (points relative to its frame zone), `shape`, `headroom_pt`, `label_width_pt`,
`axis`, and `axis_label`. Layout updates preserve omitted properties.

Only explicitly begin with `materialize/source` when converting resolved bindings
to constants is intended. Otherwise edit bound authored values. The command refuses
shared/pinned local templates, stale hashes and active native layout overrides.
Use `project reconcile` to retain/review native changes before resetting overrides.
A moved marker does not by itself establish a different assessed stage.

## Catalog companions and verification

All existing typed maturity curves use this source contract: full-width, split,
nav, 3/4/5/6-stage, table-companion, AI and insight layouts. Preserve surrounding
copy, insight cards, tables and provenance. Changing a stage count does not rewrite
an assessment rubric or claims in those companions. Inspect and explicitly edit
them together using the applicable table/assessment/chart runbook.

The `ai-two-axis`, `ai-stage-detail`, `ai-assessment`, and `ai-coaching-plan`
variants contain no maturity curve. Use explicit chart/assessment/table/card
operations for their actual semantics; do not manufacture a curve to fit this
command. The readiness adoption-curve is a quantitative plan/actual chart and follows the
quantitative runbook, including its missing observations and target.

Build after accepted source changes, render in PowerPoint, and check label/branch
clearance, marker meaning, frame bounds, and companion consistency. Keep generated
builds immutable and edit a copy for a native round trip. Retain the reviewed native
package, decisions and source predecessor. Save numbered source/deck snapshots;
use project sharing commands to verify portability.
