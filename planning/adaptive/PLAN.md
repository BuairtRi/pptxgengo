# Semantic adaptive capabilities plan

## Objective

Build a bounded semantic slide authoring surface alongside the existing source-bound editing contracts. The five initial family builders are roadmap, architecture, process, team, and comparison. Each will take asserted content and family controls and lower them to the existing editable/measurable composition engine. They do not transform an imported slide into an automatically resized or source-faithful template.

## Current state

- 65 source-bound editing contracts exist. They expose 895 named text slots across 1,570 text bindings, with declared source styles, constraints, retained source content, and fixed geometry.
- Their example decks have native render/visual review evidence (60 direct native open/export and five picture-geometry replay cases). That proves only the reviewed examples. Five generated accent PPTX files still need a direct open/repair check. None of the 65 has arbitrary-content capacity qualification.
- `scripts/build-adaptive-capabilities.py` compiles `library/adaptive/capabilities.json` from rollout assignments, each contract/values/implementation, source scenes, and review checkpoints. It records 36 taxonomy discovery hints requiring structural review and 29 source items with no current category hint. Neither count says whether a particular slide fits a builder. Re-run it after source metadata or checkpoint changes.
- `internal/adapt/types.go` and the five family builders define the semantic input/result envelope and bounded controls. The compilation command emits compose specs/reports. The adaptive checkpoint records exact reviewed examples with inputs, controls, render/evidence hashes, and limitations. This evidence applies to those examples only; every family remains pending broader qualification.
- Existing `pptxcompose` primitives have bounded native examples: cards, team/pods, named grids/panels, canvas shapes/images/text, and measured accent paths. These do not transfer adaptation approval to the new family builders.

## Qualification gates

For each family, keep records keyed to immutable spec/deck/evidence hashes and move through these states explicitly:

1. **Declared**: input controls, supported values, output objects, fixed parameters, and error/limitation cases are in the catalog.
2. **Builder implemented**: deterministic semantic input validation and lowering to compose specs; no qualification claim.
3. **Native fit measured**: representative and changed-content examples pass the current native text/geometry measurement and verification rules.
4. **Negative cases reviewed**: unsupported cardinalities, overflow, missing evidence, invalid relationships, and style limits reject or are reported accurately.
5. **Visual review complete**: every page is reviewed at presentation size for hierarchy, evidence, density, asset usage, and collisions.
6. **Qualified scope recorded**: only the tested content/geometry envelope is approved; all broader combinations stay unqualified.

Every family remains pending broader qualification, and every source-template adaptation remains unqualified. Accepted examples have exact native and visual review records, but do not qualify other count/content combinations or automatic adaptation of any source template. Passing code validation or compiling a generated fixture is not by itself native qualification. Source fidelity is not a goal or implication of these family builders.

## Initial semantic control areas

Current builder inputs are narrower than these broader semantic concepts:

- **Roadmap**: 3–12 ordered periods; 1–8 workstreams; optional groups; explicit period-ID intervals and milestones; planned, active, complete, at-risk, or blocked states. It does not infer dates or dependencies.
- **Architecture**: 2–8 ordered layers, each with 1–8 component cells; optional details, fills, width weights, up to five left-rail items, and explicit relations between known component IDs. It does not build arbitrary graph layouts, nested diagrams, or free-form groupings.
- **Process**: 2–6 ordered stages with summary, 1–8 activities, 1–5 outputs, optional planned/active/complete state, and 12–36pt panel gap. Owners, entry/exit gates, and dependencies are not represented.
- **Team**: 1–6 pods with 1–8 roles each; optional staffing tokens, up to 12 explicit pod relationships (`reporting`, `dependency`, `advisory`, `annotation`), and an optional responsibility matrix. The matrix defaults to `right` (2–4 data columns) or can be `bottom` (2–6 data columns); either position supports 1–6 rows. Credentials, team phases, and free-form matrix topology are not represented.
- **Comparison**: 2–4 options and 1–7 criteria; each criterion has a supplied numeric gauge with scale/target, one finding per option, and a strong/mixed/limited/unknown status. Linear and dial modes are supported; scores and recommendations are not inferred.

Family schemas must define which controls can vary, their cardinality and bounds, which geometry/style is fixed, and how required facts/qualifications are represented. Use native editable shapes/text and existing measured layout primitives when suitable. Do not silently discard detail, reduce font size, or shrink copy to force a fit.

## Maintenance

When a fixed-source contract changes, regenerate the capability JSON and retain exact source hashes. When a family implementation changes, update its schema/evidence separately and avoid overwriting fixed-source provenance. Treat taxonomy matches as discovery hints only; inspect structure before choosing a builder. Keep other source contracts discoverable under `unmapped` for future review, without treating that status as a permanent exclusion.
