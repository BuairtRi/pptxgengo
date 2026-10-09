# Layer, maturity and staffing composition contracts

This is the maintained implementation scope, not a desktop qualification claim.
Operator instructions live in the West Monroe skill references:

- [Layer composition](../skills/west-monroe-presentations/references/layer-composition.md)
- [Maturity composition](../skills/west-monroe-presentations/references/maturity-composition.md)
- [Staffing composition](../skills/west-monroe-presentations/references/staffing-composition.md)

## Actual source coverage

| Domain | Actual typed/source assignments | Supported customization |
|---|---|---|
| Layer stacks | Eight `architecture-layers` profile assignments: layers, layers-left, layers-split, layers-nav, layers-icons, layers-3d, layers-3d-systems, today-vs-target. Layer-map is an additional compound case outside this profile | Native row/plane/icon/compound selection; add/clone, update, remove, reorder; allocated packing; ordinal/preserved surfaces; numbered rows; explicit foundation and cross-cutting controls |
| Maturity | Sixteen maturity profile assignments: fifteen typed curves plus ai-assessment. The entire maturity family has nineteen templates: those fifteen curves plus four noncurve companions | Keyed stage count/order/spacing for actual curves; bend intensity; observed current, here, intended target, inflection, alternate branch; labels/headroom/axis/frame allocation. Noncurve companions use their actual assessment/chart/table/card model |
| Staffing | Nine team-curve profiles, ten actual curve nodes because before-after has two | Keyed series/point/phase CRUD and order; explicit sample spacing, references, unit/time/source/max; independent versus legacy above-stack line basis; smoothing/interpolation/layout |

Four maturity companions (`ai-two-axis`, `ai-stage-detail`, `ai-assessment`,
`ai-coaching-plan`) contain no maturity curve. They need chart/assessment/table/card
semantics. `readiness/adoption-curve` is a quantitative plan/actual chart with
missing values and target, not a maturity curve. Any earlier inventory grouping
that treats it as a typed maturity curve must not be used as implementation proof.

The today-vs-target source has four numbered target-side `layerrow` objects;
select only those rows and keep its current-state objects, both containing frames,
raw connector paths, transition arrow and surrounding copy separate. Moving rows
does not rewrite the current-state topology or imply a different transition.
Layer-map membership is explicit, including its separate number, label and
explanation objects. An adjacent matrix is a separate relational model, not an
inferred result of moving the row. Three-dimensional main planes are explicitly
selected independently of external/system planes; paint order is bottom-first.
Icon sources preserve their registered glyphs. Missing machine branding can use
hash-identical installed catalog originals through the bounded registered-original fallback.
The same exact-byte contract permits only the five registered `arrow-*` graphics;
it excludes logos, photos, highlights, circle graphics and unregistered IDs.

## Persistence and validation

All patches require exact inspected source hash, actor, reason and strict schema.
Unknown or action-inapplicable authored fields are rejected even when false/zero/
null. Binding materialization is explicit and first. Candidate construction clones
source before mutation, measures fit, preserves untouched nodes/frame/grid/pins,
and uses the shared guarded source transaction. Shared local templates must be
forked; local revision pins and active native layout overrides are refused.

Maturity/staffing semantic references use stable keys, lowered to the existing
renderer source contract. Source arrays retain explicit key overlays. The composer
does not invent scores, FTE units, calendar dates, missing observations or new
relationships. Staffing unit/time/source declarations reserve a visible native
bottom caption; legacy sources with no declaration keep their old appearance.
Measured previews refuse a new scale caption crossing unrelated contextual text,
before guarded source apply. Fixed source annotations require explicit relocation
if changed observations or spacing move the dark band underneath them. Maturity
target labels use the rendered61-segment curve envelope and measured text height,
with six-point clearance; a target that cannot fit above the axis is refused.
Existing
legacy line sources retain `above_stack`; newly composed band-only models default
to independent lines, and an explicit `line_basis` can select either interpretation.

Layer single-member additions use their semantic key as the native node ID.
Compound additions use `KEY-part-NN` members. Keep the explicit selection mapping
with the project when the semantic group differs from its individual native IDs;
the decision receipt retains the full operation/selections. Array/node ID/comment
preservation is best-effort for unchanged uniquely identified authored elements;
source predecessor receipts preserve comments that cannot be safely matched.

Native movement is geometry until a reviewed semantic decision establishes a fact.
Maturity current/target stages and staffing observations are never inferred from
freeform native curve movements. Existing receipt-backed geometry reconciliation
retains original packages and unknown formatting. The family commands provide
explicit semantic source adoption; native family-specific inference must name its
supported policy and evidence separately.

## Focused acceptance and private fixtures

`layers_curves_composition_test.go` covers:

- 3/5/6 maturity/layer counts and stable current/target/branch references.
- Marker cascade/refusal and regeneration, strict authored-field rejection.
- Binding acknowledgment, stale-source rejection and byte-preserving preview.
- Staffing series/point/phase changes, full phase removal, zero/nonnegative values,
  cumulative maxima, nonuniform spacing and phase/direct-label references.
- Three-dimensional painter order and native compound layer-map objects.
- Invalid selection membership, long-label fit refusal and atomic failed apply.
- Actual V11 source inspection/measured patching for all fifteen typed maturity
  curves and nine staffing variants. The layer test's eight source cases are
  seven architecture-layers profile assignments plus the additional layer-map;
  its historical test name is not evidence that today-vs-target was included.

`TestLayerCompositionActualTodayTargetPreservesCurrentState` separately uses the
complete genuine 24-node V11 source and explicit target membership `node18`–`node21`.
It measures preview, adds two keyed target rows, reorders six within a reviewed
216pt allocation inside the unchanged Target frame, then removes/reorders to
five. All twenty nonmember objects remain exact, including both frames, current
objects, raw connector paths, transition arrow and explanatory copy. Original
parent pins, frame/grid and unrelated source values remain exact. Selected row
binding materialization is acknowledged and its obsolete values are pruned;
the test does not claim that moving target rows recalculates current-state links.
Duplicate members/order, stale hashes and oversized layouts refuse atomically
with unchanged source. These focused tests passed in managed run
`86e8bbbec956ced6bca1d5965c51f506`, completing source customization evidence for
all eight architecture-layers assignments plus the additional layer-map case.

Automatic numbered-row inspection now follows resolved authored unique positive
integer `n`, rather than the incidental node paint array. After source reorder,
auto-inspection and a future layout-only apply retain that authored order.
Duplicate, zero, negative, noninteger and unresolved numbering refuse automatic
selection; an explicit reviewed selection array is required for ambiguous or
unnumbered models. This interprets authored numbering, not native movement.
The final all-layer focused suite passed
`7517763fca4762d5b58f3d5811a4db2f`, including actual today/target count edits,
future-layout order persistence, byte-preserving refusals and ordinal bindings.
The exporter now passes its actual slide values when resolving scaffold numbers;
this introspection fix does not alter the accepted eleven-case native artifacts.

`scene_staffing_semantics_test.go` exercises independent line rendering after a
filled band with/without Gaussian smoothing, including later bands.
`bundled_icon_fallback_test.go` checks exact registered original bytes, missing
catalog, derived-preview refusal, drift, traversal, duplicates and no masking of
local original corruption, all five registered arrow originals and exclusion of
non-approved graphics. `diagram_route_review_test.go` independently checks
explicit containing allocation exclusion while its separate heading is avoided.

Opt-in `TestLayersCurvesQualificationExamples` exports fresh V11-pinned projects
to `PPTXGENGO_FAMILY_QUALIFICATION_DIR`: 3/5/6-stage maturity, unequal-spaced staffing,
both actual before/after source nodes, 3/5/6-layer stacks with controls/foundation,
and catalog-derived 3-D, compound and icon stacks. A caller must perform actual
macOS open/edit/save/render, reconciliation, snapshot and sharing qualification;
fixture construction itself makes no desktop or Windows claims. The separate
[full-source eleven-case ledger](layers-curves-full-source-qualification-20261009.md)
records final macOS native AppleScript open/Save As/PDF export and independent
visual acceptance for three/five/six-stage maturity, three/five/six-layer stacks,
unequal-spaced staffing, before/after curves, and 3D/map/icon companions. That
bounded acceptance does not qualify every profile, native semantic/geometry
adoption or Windows PowerPoint.

## Scaffold geometry fidelity

Measured native ink bounds do not establish the source anchor or source width.
Scaffold/detach capture `_source_geometry` before stripping source coordinates;
this strict internal adapter retains normalized source origin and extent against
its measured allocation. Composition consumes it before strict source planning.
Operators move/resize the allocated component; they do not edit this metadata.
Before/after heading ink, quadrant strokes and axis labels may have bounds that
omit or extend beyond the source anchor. Card-row source `w` means each item width;
the local allocation owns the full row and recomputes each item width using its
current count and unchanged authored gap. Captured source curves retain their
original stroke envelope; new curves reserve their own ink. New current-phase
allocations reserve 24pt marker headroom and two-set Venn allocations reserve the
half-point outline before the renderer measures them.

`scene_source_geometry_test.go` compares real V11 native shape/text geometry before
and after adapting complete app-ecosystem, subtle quadrant and sprint capacity
specimens plus the actual runbook time-axis node (its unrelated undistributed
underscore is excluded). It checks card-row 1/2/3/5/6-item totals, moving/resizing,
strict unknown/null metadata refusal and phase headroom without double padding.
`diagram_keyed_comments_test.go` checks stable nested record/ancestor identities,
reordered staffing point/series observation evidence and ambiguity refusal. No
point identity is guessed for legacy source arrays without authored keys.
