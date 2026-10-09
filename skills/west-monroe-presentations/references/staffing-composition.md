# Compose staffing and team curves

The nine team-curve templates are starting points for team transitions, not a
fixed staffing forecast. A filled band is cumulative staffing. Choose `line_basis: independent` for an
independent line observation, or explicitly retain legacy `above_stack` for an
increment above the current filled stack. Decide which quantities are people, FTE, proportions, activity,
or capacity before changing the values. Never silently reinterpret one as another.

## Inspect the authored model

Read installed `project staffing` help and the pinned toolchain. Detach the catalog
slide and inspect its typed `wmds/component/teamcurve` node:

```sh
pptxgengo design project diagram inspect --project PROJECT --slide SLIDE
pptxgengo design project staffing inspect --project PROJECT --slide SLIDE --node CURVE
```

Inspection exposes keyed points, series, phase boundaries, the explicit maximum,
curve/smoothing options and frame allocation. Legacy examples have unspecified
units: their values cannot establish FTE or dates. A `set/scale` declaration of
`unit`, `time_unit`, and `source` is required before semantic mutation. The renderer
shows this declaration in a reserved native bottom caption. Use concise source labels;
keep detailed assumptions in project context/notes.

Preview refuses a scale caption that intersects unrelated measured contextual
text. Explicitly relocate that retained companion with a diagram patch first;
do not omit its copy to make the curve fit. After changing sample spacing,
series values or scale, inspect fixed source annotations as well as the curve:
white labels must remain over their intended dark band. Their placement does
not follow the curve automatically. Preserve heading/bullet identities and
record deliberate source geometry adjustments before rebuilding.
Keep contextual paragraphs inside their declared phase region. Widening one
across a phase divider can change its visual scope or put a rule through the
words. Use an explicit component patch for concise copy or supported note
typography when appropriate, preserving source facts; review native output after
the change. Geometric fit alone does not verify that narrative/phase association.

## Model and compose

- Keep point and series identities stable. Every point has a nonnegative finite
  value for every series; missing values are not silently treated as zero.
- Points have increasing normalized `at` coordinates in `0..1`. Unequal spacing
  can illustrate unequal phase durations, but remains an explicit author decision.
- A series is `area` (or omitted style, also area) or `line`. Area
  values stack; lines do not change the stack. With `above_stack`, their displayed
  height includes the preceding cumulative bands; with `independent`, it does not. A maximum must accommodate each line and the
  complete cumulative stack.
- A phase boundary can name an existing point with `point`, or author its numeric
  `at`. Named-point boundaries follow that point when it is moved/reordered.
- Direct labels name a stable `label_point`, or set `hide_label: true`. Do not
  move a label into an adjacent band's interior.
- Prefer monotone interpolation for bounded plans. Gaussian smoothing is visual
  only and does not modify retained source observations. `smooth` and `tension`
  are in `0..1`; zero smoothing retains the selected native curve representation.

## Patch and measure

Schema `pptxgengo.staffing-patch.v1` requires `expected_source_sha256`, `actor`,
`reason`, `node_id`, and `operations`:

```sh
pptxgengo design project staffing patch --project PROJECT --slide SLIDE --patch change.yaml
pptxgengo design project staffing patch --project PROJECT --slide SLIDE --patch change.yaml --apply
```

```yaml
schema: pptxgengo.staffing-patch.v1
expected_source_sha256: COPY_FROM_INSPECTION
actor: presentation-operator
reason: Explicitly model the revised illustrative FTE transition
node_id: teamcurve
operations:
  - action: materialize
    entity: source
  - action: set
    entity: scale
    scale: {unit: FTE, time_unit: delivery fraction, source: Illustrative assumption, maximum: 12}
  - action: set
    entity: phase
    key: transition
    phase: {key: transition, label: Transition, text: Begin handover, at: 0.75}
  - action: set
    entity: layout
    layout: {curve: monotone, line_basis: independent, smooth: 0, phase_height_pt: 54}
```

Use the inspected node and keys. Operations:

| Operation | Payload and meaning |
|---|---|
| `set/point` | `key`, complete `point: {key,at,label,values: {SERIES_KEY: value}}`; include every series |
| `set/series` | `key`, complete series; a new series needs `values` keyed by every existing point; an existing series retains observations when `values` is omitted |
| `set/phase` | `key`, `phase: {key,label,text,at,point}`; if `point` is supplied it owns the boundary |
| `remove/point` | `key`; require `cascade: true` for phase/direct-label references, or reassign first |
| `remove/series`, `remove/phase` | `key`; related point values are removed with a series |
| `reorder/point` | Every key exactly once, `spacing: even` or `explicit`; explicit needs `at` for every reordered point |
| `reorder/series`, `reorder/phase` | Every key exactly once; phase positions must remain increasing |
| `set/scale` | Complete `scale`; omitted maximum explicitly selects automatic scale |
| `set/layout` | Partial `rect`, `curve`, `line_basis`, `smooth`, `tension`, `phase_height_pt`, `phase_labels` |

No points are interpolated when adding a series/sample. Deleting a referenced
point with cascade removes its phase and hides its direct labels rather than
inventing new relationships. Prefix `materialize/source` only to explicitly turn
resolved component bindings into local constants. Shared/pinned templates and
active native geometry/order are refused; preserve/review native edits first.

## Apply to the actual catalog

`build-together`, `build-together-bands`, `build-together-roles`, `agents`,
`agents-bands`, `agents-roles`, `agents-nav`, and `agents-split` share this typed
contract. `before-after` has two distinct curve nodes: inspect and patch each by
its stable node ID, declaring comparable units/scales explicitly. Role labels,
phase copy, and bullets outside the selected curve stay unchanged; update them
when their claims would become stale.

Preview measured fit, inspect overlap and phase-label capacity, build, and check
native PowerPoint exports. Native band geometry alone does not establish changed
staffing quantities or dates. Reconcile geometry as layout or use an explicit
reviewed semantic decision supported by installed help. Preserve the raw edited
package and accepted source facts; verify version snapshots and sharing archives.
