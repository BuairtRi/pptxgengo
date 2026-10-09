# Commercial and value model scope and evidence

## Implemented contract

`project commercial inspect|patch` provides explicit decimal source models without replacing the visual language of each example. The assigned V11 scope is all 33 `commercials` variants plus 22 `value-model` profile variants. Existing tables, cards, metrics, plain text, charts, fee summaries and other actual source primitives may be registered as one calculation group. Each keeps its original presentation and allocation. Computed output changes only an explicit nongeometric presentation path; all group members update atomically and reject missing nodes or model drift.

Rows have stable keys, explicit units, decimal-string inputs with provenance or assumptions, and keyed formulas. Supported arithmetic includes sums, subtraction, ratio multiplication, declared dimensional conversion, division, ROI, simple payback and periodic NPV (period zero included). Rational intermediate calculation precedes final decimal display rounding (`half_even` or `half_away`, precision 0..8). Null differs from zero; division by zero, unknown inputs, cycles, invalid units/assumptions and bounded overflow refuse mutation. Target scaling explicitly divides base-unit values for displays such as USD millions. Target-specific precision is a text display override, not silently rounded business input. Numeric targets retain the scaled exact rational until the bounded float64 renderer/workbook boundary; the +/-1e9 renderer bound is checked against the rational before conversion, so a value just outside that bound cannot float-round into acceptance; display precision never changes raw chart facts (for example 10.8 does not become 11 when precision is 0). Nonterminating rationals use the nearest representable float64 only at that renderer boundary. Explicit rounded business inputs require authored reviewed facts.

Models and visual collections may add/remove/reorder keyed entries. Mapped collection targets follow original item identities; deletion requires explicit removal/reassignment of mappings. Nested presentation keys and model-row identity survive changes. New mappings are deliberate; formatted example fees, percentages, dates or prose never automatically become financial source facts. Bound source initialization/materialization is explicit and preserves comments, frame/provenance, native overrides and other nodes.

Scope/assumption-only variants use their actual component content/collection contracts. An optional calculation model can be registered explicitly when the operator has supplied numerical facts and exact mappings. No inferred fee is installed into an assumption-only slide. Computed commercial labels are excluded from literal-chart numeric adoption: native overtyping requires explicit interpretation of the business input, formula or display policy.

## Exact V11 applicability

| Catalog variant | Scope | Actual source types |
|---|---|---|
| `pricing/options` | commercials family | card (3) |
| `pricing/fixed-fee` | commercials family | feesummary (1), num (3), table (1) |
| `pricing/capacity` | commercials family | allocation (1), feesummary (1), table (1) |
| `pricing/capacity-split` | commercials family | allocation (1), feesummary (1), table (1) |
| `scope/service-value` | commercials family | bullets (2), table (1) |
| `scope/service-value-nav` | commercials family | bullets (2), table (1) |
| `pricing/options-nav` | commercials family | card (3) |
| `pricing/fixed-fee-left` | commercials family | num (3), rule (2), table (1), text (6) |
| `pricing/options-right` | commercials family | rule (1), table (1), text (6) |
| `pricing-options/tiers` | commercials family | card (3) |
| `pricing-options/two` | commercials family | card (2), rule (1), strongnum (1), text (2) |
| `pricing-options/two-left` | commercials family | card (2), rule (1), text (7) |
| `pricing-options/matrix-split` | commercials family | block (4), check (3), rule (1), table (1), text (5) |
| `fixed-fee/by-phase` | commercials family | block (1), bullets (3), phasehead (3), text (6) |
| `fixed-fee/with-assumptions` | commercials family | bullets (1), num (2), rule (1), table (1), text (7) |
| `fixed-fee/add-ons` | commercials family | feesummary (1), num (1), table (1), text (2) |
| `capacity/run-rate-inverse` | commercials family | metric (2), text (3), textblock (3) |
| `capacity/ramp` | commercials family | chart (1), metric (1), rule (1), strongnum (1) |
| `capacity/tiers` | commercials family | table (1) |
| `payments/milestone-schedule` | commercials family | card (4), timeaxis (1) |
| `payments/milestone-schedule-tall-left` | commercials family | num (2), table (1), text (8) |
| `payments/milestone-schedule-tall-right` | commercials family | num (2), table (1), text (8) |
| `payments/invoicing-cadence` | commercials family | cardrow (1), num (6), table (1) |
| `assumptions/dependencies` | commercials family | strongnum (1), table (1), text (2) |
| `assumptions/scope-boundary` | commercials family | card (1), rule (1), text (6) |
| `assumptions/change-control` | commercials family | block (1), cardrow (1) |
| `assumptions/detailed-categories` | commercials family | table (6), text (6) |
| `assumptions/detailed-numbered` | commercials family | table (2), text (2) |
| `assumptions/detailed-impact` | commercials family | table (1) |
| `investment/roi-payback` | commercials family | bullets (1), chart (1), metric (4), text (1) |
| `investment/cost-vs-value` | commercials family | chart (1), rule (1), text (5) |
| `investment/phase-fee-table` | commercials family | num (1), table (1) |
| `investment/total-inverse` | commercials family | metric (2), rule (5), text (14) |
| `value-summary/big-number` | value-model profile | mark (1), metric (4), rule (1), text (9), textblock (2) |
| `value-summary/left-panel` | value-model profile | chart (1), metric (3), rule (1), text (6) |
| `value-summary/inverse` | value-model profile | metric (2), rule (5), text (15) |
| `value-summary/five-year-table` | value-model profile | metric (4), num (6), table (1) |
| `value-bridge/levers` | value-model profile | block (7), bullets (1), connector (6), metric (2), rule (2), text (9) |
| `value-bridge/levers-split` | value-model profile | block (10), bullets (1), connector (6), rule (5), text (16) |
| `value-curve/break-even` | value-model profile | chart (1), connector (1), text (8) |
| `value-curve/break-even-split` | value-model profile | bullets (1), chart (1), connector (1), rule (4), text (11) |
| `value-curve/scenarios` | value-model profile | chart (1), connector (1), rule (3), text (10) |
| `value-levers/tree` | value-model profile | block (21), connector (15), text (4) |
| `value-levers/table` | value-model profile | num (3), rating (1), table (1) |
| `value-scenarios/three-case` | value-model profile | block (3), rule (18), text (39) |
| `value-scenarios/table-nav` | value-model profile | num (3), table (1) |
| `value-phases/self-funding` | value-model profile | block (9), rule (3), text (21) |
| `value-phases/wave-table` | value-model profile | num (4), table (1), tag (1), text (3) |
| `value-types/hard-soft` | value-model profile | block (7), bullets (4), rule (4), text (12) |
| `value-types/hard-soft-right` | value-model profile | bullets (1), chart (1), metric (2), rule (2), text (10) |
| `inaction/cost-of-waiting` | value-model profile | bullets (1), chart (1), rule (3), text (9) |
| `inaction/inverse` | value-model profile | block (5), metric (2), rule (6), text (15) |
| `value-tracking/planned-vs-realized` | value-model profile | chart (1), metric (4), num (2), status (1), table (1) |
| `value-tracking/owners-nav` | value-model profile | num (3), status (1), table (1), text (3) |
| `allocation/spend-vs-value` | value-model profile | allocation (2), num (3), table (1), text (3) |

## Focused evidence and runnable fixtures

`internal/deckproject/commercial_composition_test.go` exercises native-source registration, decimal mapping, preview/apply/stale guards, preserved comments and unrelated copy, keyed collection target remapping, repeated row/target count changes and atomically synchronized multi-node groups. `internal/wmdesign/scene_commercial_test.go` covers arithmetic, units, formulas, rounding ties, missing/zero, cycles and overflow. `quantitative_catalog_test.go` scaffolds all 55 actual stock examples: examples with financial displays receive an explicit illustrative source mapping; scope-only examples exercise materialization and their real keyed source collections. Catalog tests retain all other source nodes and frame requirements.

```sh
slotctl --repository pptxgengo --slot pptx-v430-composition exec -- go test ./internal/deckproject ./internal/wmdesign ./cmd/pptxdesign -run 'TestCommercial|TestProjectCommercial' -count=1
```

Set `PPTXGENGO_COMPOSITION_FIXTURES` for catalog exports. Private fixture directories include `pricing-fixed-fee`, `pricing-options`, and `value-summary-big-number`, with `deck.yaml` and current-hash `patch.json`. Pin against V11, preview/apply the commercial patch on `catalog-slide`, inspect the explicit calculated output, and build. These narrow contract fixtures deliberately preserve unrelated catalog copy; they are not complete financial case studies. Before presenting a complete financial slide, audit and synchronize every claim as required by the operator runbook, including titles and prose outside the registered component group. Retain the original frame. Desktop qualification should move a registered component in a separate native copy, reconcile explicitly, rebuild, and verify its numeric source group before snapshot/ZIP. The integration ledger records final stock and desktop run outcomes; these source/package tests do not imply Windows Office qualification.

Operator instructions live in [the commercial runbook](../skills/west-monroe-presentations/references/commercial-models.md).

## CI cadence for source inventory evidence

The exhaustive source catalog audits run without `-short` in the private GitLab
`composition-catalog` nightly/on-demand lane. Normal main and release-tag unit
checks retain representative variable-count, source-preservation, geometry and
semantic regressions, while `testing.Short()` skips the two full inventory sweeps.
A retained complete source audit remains required before cutting the v4.3.0 tag;
the protected release resource job independently regenerates the full template
browsing deck. No race detector or desktop runner is added to release tags.

## Complete six-slide financial and quantitative qualification

The private `v430-six-coherent-r5-20261009` fixture is a complete authored
customization, separately from the narrow stock catalog mutation exercises.
Every visible numerical claim is mapped to a reviewed calculation target or
recorded as a fixed scope/schedule/discount assumption. Arbitrary prose is not
parsed into business facts. The operator must re-audit all claims when changing
those assumptions or inputs.

| Example | Reviewed source meaning |
|---|---|
| Close-duration chart | Three explicitly illustrative Today/After pairs in days; narrative quotes Wave 1 12→5 and makes no unsupported funding or entity-count claim. |
| Count-growth chart | Three waves improve; a fourth long-label wave remains 4→4, explicitly described as unchanged. |
| Scenario curve | Six chronological annual points Y0..Y5, USD 4.2M upfront, annual net benefits 2M/2.6M/3M. All 18 modeled raw chart/cache/workbook values remain exact at the float64 boundary; Y5 endpoints 5.8M/8.8M/10.8M and simple payback 2.1/1.6/1.4 years agree. The obsolete month 31 marker is removed; the zero break-even series remains explicit. |
| Pricing options | USD 400K diagnostic; USD 1M plus 210K/month × 4 months = 1.84M hybrid; 240K/month × 3-month minimum = 720K capacity. Duration and scope are fixed documented assumptions. |
| Five-year value | USD 4.2M upfront; 2.6M/year × 5 = 13M benefit; 8.8M net; 210% ROI; 3.1× value per dollar; 19 months simple payback; 6.2M NPV at 8%. Eight registered displays update atomically; headline and narrative no longer contain unrelated sample 123 or stale 6M. |
| Fixed fee | Milestone fees 100/300/400/200K total 1000K; phases 400/600K; cumulative100/400/800/1000K; proportions 10/30/40/20%; 8% expense cap. Fixed approved example weeks 1/4/7/8 remain documented schedule assumptions. |

Genuine CLI initialization/build used the frozen
`.cache/qualification/pptxdesign-commercial-exact`, not synthetic receipts or
changed toolchain pins. Combined baseline is
`build-20261009T160132-97af26441bcc4def`; exact source SHA256 is
`4b604b96c895c0ee9248797463f51db93f2a528326be20e8099830382abfb52e`.
`claims-audit.json` retains per-slide source hashes, reviewed claim inventories
and exact initialization/build commands. Generated baselines remain immutable;
all GUI working copies have mode 0600.

Full commercial/catalog focused tests passed slotctl run
`ba1183969409f4ea9218fa0211132145`. Fast complete-model and private exact-source /
cache / embedded-workbook tests passed
`7a8a68d537e4c1024930050c6ab7e179`. Exact positive/negative numeric bound, zero and tiny-value regressions also passed final focused run `5f7fae24cf78bc56256fd9098cb6ff0e`:

```sh
slotctl --repository pptxgengo --slot pptx-v430-composition exec -- env \
  PPTXGENGO_COHERENT_COMMERCIAL_FIXTURES=/Users/rscott/Documents/pptxgengo-qualification/v430-six-coherent-r5-20261009 \
  go test ./internal/deckproject ./internal/wmdesign \
  -run 'TestCommercialCoherentSixSourceClaims|TestCommercialCompleteFiveYearValueDisplays|TestCommercialNumericTarget' -count=1
```

An earlier native PDF review identified rounded raw scenario facts caused by
incorrect numeric-target display rounding; the implementation now preserves
raw numerical facts and the regression independently checks all 18 scenario
chart/cache/workbook points. Earlier failed sources and native reviews are
retained; they do not qualify this final baseline. The final native review is recorded below and in the integration ledger. Neither this scoped macOS work
nor the source tests establish Windows PowerPoint qualification.

The value-summary ROI and payback allocations are x57..255 and x273..471 pt,
respectively (18 pt allocation gap). Measured ROI text advance is 92.7998 pt,
leaving 123.2002 pt to the next value's origin; the actual preceding native PNG
also showed separated ink. Final baseline retains those unchanged allocations.

### Actual final macOS Office review

The final r5 working copy opened and completed GUI Save As without repair.
Actual retained native package:
`/Users/rscott/Documents/pptxgengo-qualification/v430-six-coherent-r5-20261009/combined-six/working/combined-six-exact-native-saved.pptx`. Its SHA256 is
`c78577807d3f55b8394097f867f2dbc416720697161239b84db03b70ced699d2`.
The independent authenticated chart reader verified all 18 scenario facts plus
six zero break-even points against both chart caches and the embedded workbook,
with original ownership, keyed series/categories and binding ranges retained.
This passed run `66eecabc5d8abd9d9555472eb410f49e`, together with complete-model,
exact-bound, and private source/build tests.

The native PDF
`/Users/rscott/Documents/pptxgengo-qualification/v430-native-final-review-20261009/financial-six-exact-native-final.pdf`
has SHA256
`c21bd59852dcb06c0e82a51f0e65651aacc3019bbc1163c23a19fab4d5839f76`.
All six 1920×1080 rasters in sibling `financial-six-exact/png` were independently
viewed: no clipping or overlapping content observed; all reviewed headline,
legend, prose, fee and metric claims agree with the declared model/assumptions.
Scenario end labels now read 5.8/8.8/10.8, matching their monetary callouts.
The value-summary ROI and payback have separated native text rectangles as
well as separated visible ink; actual Save As retains x57/w198 and x273/w198 pt.
This is visual acceptance for these exact illustrative fixtures, not a claim
of visual equality for arbitrary source changes or Office versions.

To repeat the exact native fact read, add
`PPTXGENGO_COHERENT_COMMERCIAL_NATIVE` with the native package path above and
include `TestCommercialNativeSixScenarioFacts` in the focused command. The test
reads immutable baseline facts and authentic native Save As; it does not edit
pins, invent receipts, rewrite the evidence package or adopt business changes
from formatting.
