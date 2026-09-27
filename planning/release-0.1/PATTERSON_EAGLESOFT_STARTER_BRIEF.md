# Starter brief: Patterson / Eaglesoft modernization RFP response

**Source:** `samples/Patterson - Eaglesoft Modernization RFP Deck.pptx` (39 slides; SHA-256 and full slide text are in `deck-outlines.json`). It is explicitly a proposal discussion draft with source-specific facts and unresolved placeholders.

## Existing narrative to preserve

Slides 2–9 set out current-state understanding and target architecture. Slides 11–20 cover transition architecture, modernization lifecycle, Intellio® Evolve, performance assurance, initial-phase planning and shared journey inventory. Slides 22–27 define the Phase 0 workplan, pod/staffing concepts, scope guardrails, fees/team flexibility and decisions/next steps. Slides 28–31 provide case studies and relevant capabilities; slides 32–36 are named representative biographies. Slides 37–39 describe intended outcomes, identified backlog and a conceptual future-state architecture.

## Useful first authoring pass

A focused response can be organized around source-supported material already in the deck:

1. Eaglesoft understanding, constraints and current architecture (slides 2–6).
2. Target architecture and modernization approach (7–14).
3. Performance/operability and shared journey inventory (15–20).
4. Phase 0, pilot, workplan, team model and governance (22–27).
5. Evidence/proof, capabilities and relevant case studies (28–31).
6. Outcomes and future-state architecture (37–39).

Layout candidates from the 65-item shortlist (implemented, source-bound fixed-layout contracts with reviewed examples; none is arbitrary-content capacity-qualified):

- `t001-uhg-013` or `t041-uhg-014` — layered/component architecture overview or detailed component map.
- `t032-software-modernization-009` — source/target architecture and integration landscape.
- `t002-software-modernization-044` — product/platform/integration explanation, only if source wording still accurately describes the offer.
- `t027-software-modernization-035`, `t039-uhg-038` or `t049-enablecomp-035` — transition roadmap/workplan patterns.
- `t016-graphics-and-layouts-041` or `t028-uhg-043` — team/responsibility layouts.
- `t064-uhg-050` or `t052-uhg-056` — case-study evidence, using only approved, source-attributed facts.
- `t054-uhg-049` — healthcare offerings/capability matrix, if it is relevant to Patterson and approved for use.

## Open source issues before external use

The source contains explicit “Tyler to add slide” placeholders (slides 10 and 21), an apparent “Traumasoft” reference on slide 26 within an Eaglesoft proposal, named biographies/roles (slides 23–24 and 32–36), and consequential numerical/result statements (including backlog, team, timeline and case-study claims). Treat these as source text requiring owner confirmation; do not silently repair or repeat them as verified facts. Resolve the pricing/scope/team assumptions against approved commercial material. Review inherited vendor/customer logos, bios, metrics and legal/footer text before circulation.

Use source-derived facts only. No new delivery commitments, staffing, savings, timelines or client outcomes are supplied by this brief. The 65 implementations are usable fixed-layout starting points with reviewed examples (60 native exports and five replay reviews), but arbitrary-content capacity and dynamic reflow are not qualified. Their 895 named slots cover 1,570 source text bindings. Thirteen portable library contracts are a separate set.

## Quick start in the repository

```sh
pptxgengo template inspect --id t001-uhg-013
pptxgengo template build-review --id t001-uhg-013 --values path/to/approved-values.json --out samples/visual-wave3/release-0.1-patterson-architecture
pptxgengo component check --project path/to/source-project --contract path/to/contract.json --values path/to/approved-values.json
```

Starter prompt: “Using only the approved source deck, draft a concise Eaglesoft current-state and target-architecture slide. Preserve the proposal’s existing terminology, separate observed facts from proposed design, and do not add components, vendors, commitments, or performance claims. Map the content into the editable slots of `t001-uhg-013`; keep the original component count and validate the exported review image.”

Use an unused output directory for every review version. Inspect the returned contract and source artwork before writing values; keep existing paragraph/run cardinality and check the native render before sharing. For a new layout that requires different row/column/item counts, prototype a separate `pptxgengo compose probe --spec path/to/spec.json --out path/to/new-probe` and confirm the needed primitive exists before committing to that design.
