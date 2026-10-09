# Commercial and value models

Choose the business question first: option pricing, phase fees, capacity/run-rate, milestone payments, investment/value, ROI, payback, NPV, or an assumption/scope boundary. Preserve the template's visual treatment while changing supplied facts and counts. Source examples are illustrative; they are not approved commercial inputs.

## Calculation and presentation are explicit

The composer registers a decimal calculation model against one or more existing local source components. Each component retains its own presentation (table, card, metric, text, chart, fee summary, allocation cell, or other supported source primitive). Declared outputs replace **only** exact presentation paths. All nodes registered in the same calculation group receive the same updated model in one guarded source transaction.

```sh
pptxgengo design project diagram inspect --project PROJECT --slide SLIDE
pptxgengo design project commercial inspect --project PROJECT --slide SLIDE --node NODE
pptxgengo design project commercial patch --project PROJECT --slide SLIDE --patch patch.yaml
pptxgengo design project commercial patch --project PROJECT --slide SLIDE --patch patch.yaml --apply
```

Before initialization, inspect the actual source node arguments, keys, and geometry with `diagram inspect`; `commercial inspect` requires an already initialized model. Shared templates need deliberate local scaffold/fork first. Select actual stock components and explicitly map each displayed financial claim. Do not automatically parse a formatted `$4.2M`, duration, percentage, or narrative sentence into approved facts.

## Audit the whole slide before mapping

Create a claim inventory covering the title, subtitle, hero numbers, every table cell and total, chart/legend labels, option cards, side panels, narrative sentences, footnotes, units, time horizons and source statements. For **each** quantitative claim, identify either its keyed model row and exact target, or an explicitly reviewed fixed assumption. Verify percentages, cumulative totals, phase totals, ROI denominators, currencies, scaling and date/period conventions agree. Label a demonstration as illustrative; retain no claim that a sample is an approved estimate.

Targets write only declared paths inside registered source components. A frame title or other prose outside those components is not implicitly linked. Update it explicitly when its facts change, or use a qualitative heading that remains valid. A single mapped hero number does not qualify the full slide: replacing `$4.2M` with a sample value while leaving an investment/ROI story unchanged is an incomplete example. After every input or target change, repeat the inventory and review the rendered deck. Remove obsolete unused bindings/values deliberately so the source does not retain conflicting historical claims.

## Initialize a model

Patch envelope: `pptxgengo.commercial-patch.v1`, current `expected_source_sha256`, `actor`, `reason`, `node_id`, and `operations`. First `initialize/source` operation requires `model`, optional additional `nodes`, and `cascade: true` to acknowledge replacing hardcoded displays with declared calculations. This retains the source presentation, frame, other components, comments and provenance. It does not mutate the shared catalog.

```yaml
schema: pptxgengo.commercial-patch.v1
expected_source_sha256: CURRENT_SOURCE_HASH
actor: Operator
reason: Price the supplied staffing estimate using the approved rate assumption
node_id: fee
operations:
  - action: initialize
    entity: source
    cascade: true
    nodes: [fee-caption]
    model:
      precision: 2
      rounding: half_even
      assumptions: [Excludes taxes and expenses; illustrative example]
      rows:
        - {key: hours, label: Estimated hours, unit: hours, value: '12.5', assumption: Supplied estimate}
        - {key: rate, label: Hourly rate, unit: USD/hour, value: '100.10', source: Approved rate card revision 3}
        - key: fee
          label: Extended fee
          unit: USD
          formula: {operation: convert, inputs: [hours, rate]}
          assumption: hours multiplied by USD/hour produces USD
      targets:
        - {node_id: fee, row: fee, path: /value, prefix: '$'}
        - {node_id: fee-caption, row: fee, path: /text, prefix: 'Illustrative fee is $', suffix: '; excludes taxes'}
```

Replace node IDs and paths with the actual selected source. An omitted target `node_id` at initialization means the primary selected node. A target cannot write geometry or change the source type. Numerical chart/table fields may require `numeric: true` without prefixes/suffixes. For a `$M` display use `scale: '1000000'`, `prefix: '$'`, `suffix: M`; the model continues storing base USD. Target-specific `precision` may override global text display precision. A `numeric: true` target retains the scaled exact result until the bounded float64 renderer/workbook boundary; global or target display precision does not round its source fact. Style numerical chart labels through the chart formatting contract. If a business rule requires a rounded numerical input, author that reviewed rounded fact with provenance.

## Facts, formulas, rounding and assumptions

Rows have stable `key`, `label`, `unit`, and exactly one decimal-string `value` or keyed `formula`. Literal input rows require `source` or an explicit `assumption`. Zero is valid input; missing/null is not silently zero. Store unresolved facts outside the complete calculation until supplied. Decimal strings allow up to 18 integer and 12 fractional digits, with no exponent notation; values/calculated results above 10^18 are refused.

| Formula operation | Meaning and guards |
|---|---|
| `sum` | One or more inputs; all input/output units identical |
| `subtract` | Exactly two same-unit inputs |
| `multiply` | One dimensionless `ratio` and one output-unit operand |
| `convert` | Two inputs; explicit dimensional/conversion assumption required |
| `divide` | Nonzero denominator; equal units produce `ratio`; other derived units require explicit assumption |
| `roi` | `(benefit − investment) / investment`; positive investment, same monetary units, `ratio` output |
| `payback` | Investment / positive net annual benefit; explicit period/unit assumption required; simple undiscounted payback |
| `npv` | Discount `ratio` > −1 followed by cash flows at periods **0..N** in one output unit |

NPV therefore includes initial expenditure as a negative period-zero cash flow; discount periodicity must match the cash-flow spacing. For irregular dates, taxes, IRR, fractional break-even interpolation, FX timing, or a specialized accounting policy, explicitly prepare supported period/source facts and assumptions; do not claim the generic algebra implements those policies.

Calculations use rational arithmetic. Intermediate formulas use unrounded values. `precision` 0..8 and `rounding: half_even|half_away` apply to final displays, including negative ties. Inspect returns exact rational results and rounded decimal outputs. If the contract requires per-line rounding **before** summation, author the approved rounded line values as explicit input facts with provenance; never imply final-only rounding is equivalent.

## Change counts and structure

| Action/entity | Fields |
|---|---|
| `materialize/source` | First only; resolves all registered calculation-group bindings explicitly |
| `set/row` | `key`, complete matching `row`; adds or replaces fact/formula |
| `remove/row` | Existing `key`, `cascade: true`; remove dependent formulas and mapped targets explicitly first |
| `reorder/row` | Complete `order`; formulas retain keyed references |
| `set/targets` | Complete `targets`; maps financial displays across registered nodes |
| `set/assumptions` | Complete `assumptions` list |
| `set/rounding` | `precision`, `rounding` |
| `set/presentation` | Exact `path`, `value`; change selected component's existing non-geometric source field |
| `set/collection` | Exact array `path`, stable `key`, complete item `value` |
| `remove/collection` | Exact array `path`, existing `key`, `cascade: true` |
| `reorder/collection` | Exact array `path`, complete `order` |

Presentation collection keys preserve row/phase/card identities. Reordering a selected collection remaps that node's declared output paths to the original item key. A mapped item cannot be deleted until its target is explicitly removed/reassigned. Newly added rows/cards need complete source data and explicit targets; changing model row order alone does not reorder the visual table. Table column widths, allocated card count, labels, row heights and footer bounds still require measured fit. Geometry changes use the applicable diagram/layout runbook.

A calculation group detects missing nodes or diverged mirrored models before updates. Fork the local template for one slide before changing it if shared by other slides. Pinned references, native layout overrides, stale hashes, invalid formulas, cycles, incompatible units, incomplete observations and overflow refuse source mutation.

## Review and deliver

Preview, inspect exact/rounded outputs, apply, and rebuild. Check every financial display and surrounding claim, including headline totals and chart centers. A registered group synchronizes its declared mappings; **unmapped prose, dates, option membership, taxes and assumptions do not change automatically**.

Native geometry/text changes do not authorize changing monetary inputs or formulas. In particular, a calculated label cannot become a new source fact merely because a person typed over it in PowerPoint. Retain that edited package and explicitly interpret whether the input, formula, display rounding, or narrative should change. Numeric chart reconciliation handles literal chart source facts, not computed commercial outputs.

Qualify the changed topology and visual treatment in PowerPoint, retain the decision receipt and predecessor, save a numbered snapshot, and share the portable project. The skill's [project structure](project-structure.md) and [PowerPoint recovery](powerpoint-recovery.md) references cover those workflows.
