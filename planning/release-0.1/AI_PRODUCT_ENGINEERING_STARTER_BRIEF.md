# Starter brief: AI Product and Engineering Executive Accelerator Lab

**Source:** `samples/AI Product and Engineering Executive Accelerator Lab.pptx` (66 slides; SHA-256 and slide text are in `deck-outlines.json`). This is a working accelerator-lab/workshop deck, not clean source copy for a new client.

## Existing narrative to preserve

The deck frames how AI changes Product and R&D work, then walks through a hands-on lab. Slides 13–28 define the agenda and six stages: joint framing, parallel preparation, product–engineering alignment, parallel build/planning, acceptance iteration, and scale-up/reflection. Slides 30–35 discuss enablers, maturity, and an illustrative agentic workflow. Slides 39–66 contain workshop design principles, readiness gates, checkpoint/skill architecture, detailed phase views, evaluation prompts, and portfolio-company readiness material.

The intended day is participant-led: select a valuable, feasible opportunity; Product defines intent and acceptance evidence; Engineering establishes a safe delivery path; both authorize a bounded slice; then validate against criteria and capture company actions. Preserve that shared-decision logic and the distinction between example workflow and participant-owned work.

## Useful first authoring pass

Start with a concise executive/workshop version assembled from existing claims and diagrams:

1. Workshop purpose and stated objectives (source slides 2, 40, 50).
2. Day agenda and facilitation format (13–14, 41–43).
3. Six-stage process map and exit decisions (16–28, 51–56).
4. Product and Engineering parallel tracks and handoff evidence (18–24, 52–54).
5. Readiness gates, maturity and operating enablers (30–35, 44–49).
6. Close with participant-owned actions and readiness questions (28, 56, 65–66).

Relevant shortlist patterns to inspect as layout candidates (implemented, source-bound fixed-layout contracts with reviewed examples; none is arbitrary-content capacity-qualified):

- `t034-enablecomp-001` — opening narrative with context/challenge/takeaway.
- `t043-graphics-and-layouts-044` — phase detail with objective, activities, deliverables and value strip.
- `t006-graphics-and-layouts-021` or `t019-graphics-and-layouts-029` — lifecycle/process sequence.
- `t031-graphics-and-layouts-065` — three phase cards; useful only where the source material truly reduces to three phases.
- `t058-graphics-and-layouts-095` — milestones/roadmap.
- `t005-uhg-012` or `t008-graphics-and-layouts-032` — capability/maturity comparison if evidence and scoring remain source-backed.
- `t029-software-modernization-034` or `t041-uhg-014` — evidence/context or component architecture.

## Authoring guardrails

Use no invented company, participant, repository, success, adoption, ROI, or delivery-result claims. Slide 7 contains prep-discovery themes; later slides include named company deep dives and portfolio-company readiness material. Keep those source boundaries explicit and confirm what is approved for the target audience. Slides 10–11 and multiple later slides are topic-specific or incomplete; decide their role from source content before inclusion. Do not turn an illustrative MFA/test workflow into a claimed client implementation. Source names, people, photos, metrics, and footer/legal content must be reviewed for the intended distribution.

These are usable fixed-layout starting points, not promises of fit for arbitrary copy. The 65 implementations expose 895 named slots covering 1,570 source text bindings; 60 examples have native visual review and five have native replay review. Current metadata keeps `adaptation_qualified: false`; short/typical/dense capacity envelopes and dynamic reflow are future work. Thirteen portable library contracts are a separate set, not the same as these 65 reviewed source-pattern implementations.

## Quick start in the repository

```sh
pptxgengo template inspect --id t034-enablecomp-001
pptxgengo template build-review --id t034-enablecomp-001 --values path/to/approved-values.json --out samples/visual-wave3/release-0.1-lab-opening
pptxgengo component check --project path/to/source-project --contract path/to/contract.json --values path/to/approved-values.json
```

Use an unused output directory for every review version. Inspect the returned contract and source artwork before writing values; keep existing paragraph/run cardinality and check the native render before sharing. For a new layout that requires different row/column/item counts, prototype a separate `pptxgengo compose probe --spec path/to/spec.json --out path/to/new-probe` and confirm the needed primitive exists before committing to that design.
