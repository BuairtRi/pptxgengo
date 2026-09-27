# Adaptive families milestone

This milestone adds a semantic authoring layer to the existing source-template and measured-component routes. The frozen installed `0.1.0-local.3` binary and skill remain unchanged.

## Implemented

- `pptxadapt compile`: strict JSON inputs, deterministic editable composition, immutable input/spec/report/hash bundle.
- Development dispatcher route `pptxgengo adapt`; next-install packaging includes the compiler.
- Five builders: roadmap, architecture, process, team, comparison.
- Shared body bounds, type-size controls, three style profiles, named color roles and contrast checks.
- A capability inventory covering all 65 source editing contracts, with original slots, styles, constraints and provenance. Category matches are discovery hints requiring structural review.
- A local comparison gallery generator with source/reference images, actual variants, input/spec links, evidence, controls and caveats.
- Progressive-disclosure authoring guidance in the repository skill. The installed frozen skill is not changed.

## Controls demonstrated by the review set

| Family | What changes | Examples |
| --- | --- | --- |
| Roadmap | Period and workstream counts, interval endpoints, states, milestones, group rows, caption spans, style | 6 |
| Architecture | Layer/component counts, relative cell widths, component fills, cross-cutting concerns, relationship routes | 6 |
| Process | Stage count, activities/outputs, stage state, panel gap, type/style | 5 |
| Team | Pod/role counts, staffing meaning, connections, matrix rows/columns, right/bottom matrix placement | 6 |
| Comparison | Option/criterion counts, findings/statuses, numeric values and target ranges, linear/dial gauges | 9 |
| Artwork arrows | Three exact artworks at three selected scale/rotation pairs; one annotated manual fallback | 10 |

All 32 family examples have completed native fit, final PowerPoint verification, and visual review. The checkpoint records the exact accepted files and hashes. Supported parser ranges are broader than the observed examples. All 10 arrow cases also completed native and optical review, with nine exact automatic placements and one manual fallback recorded separately.

## Review artifact

Open `samples/adaptive/gallery-v3/index.html` for the 42 reviewed cases and all 65 source-template previews. It includes local copies of inputs, decks and evidence. All 252 recorded artifact hashes matched; 359 local gallery links exist and the inline JavaScript syntax check passed. Browser UI inspection was unavailable, so the HTML interface itself was not visually verified. The durable manifest is `planning/adaptive/review.json`.

## Evidence and acceptance

`library/adaptive/checkpoint.json` is the acceptance record for exact reviewed examples. `planning/adaptive/qa-ledger.jsonl` records errors found during fit, routing and visual review. Neither compilation nor a native text-fit pass qualifies an entire family or arbitrary content.

The review set contains 32 family examples (5 process, 6 roadmap, 6 architecture, 6 team, 9 comparison) plus 10 arrow placement cases. Nine arrow cases propose automatic placement; one deliberately demonstrates annotated manual placement. All 32 family examples are accepted. Consult the separate arrow checkpoint for each placement case; its manual fallback remains manual even after review.

Recompiling all 32 canonical family inputs reproduces the review geometry and content (ignoring only standalone page numbering); see `reproducibility.json`.

User review exposed two associations that the initial optical pass missed: Gantt workstream labels were vertically offset from activity bars, and architecture cross-cutting controls appeared to label individual rows. The revised shared activity lane aligns all 20 measured Gantt row centers; the architecture builder now separates the compact control panel with a vertical rule, attaches each layer label to its component band, and gives layers distinct gaps. Both corrections passed native verification and renewed visual review; see `roadmap-alignment.json` and QA items ADAPT-016/017.

Nine manual CLI rejection probes cover unknown fields/roles, unreadable primary colors, unsupported counts, invalid references and invalid numeric ranges. These are bounded checks, not an automated regression suite. No Go test suite was added or run in this milestone.

## Scope limits

- A family builder creates a new composition. It does not structurally rewrite an arbitrary imported PowerPoint slide or retain its exact pixel appearance.
- The 65 source editing contracts keep their existing geometry limits. The five family implementations do not turn them into 65 structurally adaptable templates.
- Counts accepted by a parser are input bounds; every content/style combination still needs native fit and final visual review.
- Arial is the current native measurement font. The compiler does not silently shrink type or remove content.
- Roadmap states share some default colors; interval-level state distinctions may need explicit wording or separate styling work.
- One supplied comparison gauge belongs to a criterion; it does not separately score every option. Option findings carry independent qualitative statuses.
- Complex relationship routing remains bounded and may reject a composition that needs a different arrangement.
- Role and takeaway fields guide agent authoring and become notes; the engine does not research, substantiate or write the narrative.
- Rich photo/illustration descriptions and general automatic accent placement remain separate work. Existing asset metadata is preserved.

## Next development boundary

Next, map high-value source patterns to proven structural controls through explicit review. Expand source-specific adaptations, independently address complete/risk/blocked state styling, and add richer content-capacity guidance based on measured examples. Keep teammate round-trip editing and arbitrary complex diagrams behind the user's current inventory and adaptation priorities.

A subsequent release needs a new version and synchronized version declarations in the installer, catalog index/HTML, and release README. The installer is still guarded at `0.1.0-local.3`, which already exists and must not be overwritten. Its next-release staging logic copies only named evidence from both `families.*.examples` and the separate `accents.*.examples` checkpoint sections, plus the capability catalog's linked schema source files, checking recorded evidence hashes before staging. Arrow evidence stays separate from the 32 family examples; the arrow spec may live under `library/diagram-components/`. Refresh the release catalog and activate a new version before running the installer; installation is not part of this development checkpoint.
