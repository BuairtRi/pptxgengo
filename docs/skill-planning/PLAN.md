# Presentation production implementation plan

**Owner:** primary integration agent and the operator  
**Created:** October 3, 2026  
**Current stage:** Skill/workflow and bounded composition shipped in local.7/v3; frozen587 v5 implementation is in native repair and release qualification  
**Next milestone:** Native recheck repaired 587 gallery, complete full tests and portable qualification, then publish coherent v5 release and skill defaults

This is the working plan and progress record. Update its task states, decisions, evidence, limitations, and next action as work completes. An implemented feature, a successful compilation, and a reviewed PowerPoint specimen are separate states.

## Current status by track

| Track | Implemented and evidenced | Remaining |
| --- | --- | --- |
| Skill and content workflow | Seven-stage entrypoint; progressive framing/narrative/copy/selection/composition/review references; installed local.7 pack | Live brand examples and operator voice refinements; review ZIP is technical, stage-specific audience packets are manually constructed |
| Source and bounded composition | Strict deck.yaml; shared/local designs; frames/grid/assets/components; fork/detach/provenance; alternatives; resume/approvals/maps/portable exports; earlier source-backed local-composition pilot | New v4 adapter qualification through maintained projects; unrestricted automatic composition and PPTX reconciliation are not implemented |
| Frozen587 library | Pinned v5 bundle, source587/bound586 production references; 7033-node audit; all587 native pages individually reviewed | Repair findings and individually recheck every changed render; matching native previews |
| Release | local.7/v3/248 installed; candidate explicit v4 selection works; relocated candidate pilot builds | Preview-backed unified catalog, candidate native pilot, coherent stage/default/skill migration and activation |
| Continuing intake | Frozen587 ingested as separate v5;65 additions and3 revisions, no removals; frozen522 preserved | Future source changes enter another pinned intake |

## 1. Intended outcome

Deliver one versioned West Monroe presentation skill and application that helps an agent develop evidence-backed content, select or compose suitable slides, and generate an editable deck from maintained source. Work must be stoppable, resumable, and shareable across collaborators and sessions.

The operator maintains one authored `deck.yaml` per presentation. Shared design definitions remain persistent application/library resources. PowerPoint remains an ordinary editing and collaboration surface. The initial architecture preserves enough identity and provenance for later reconciliation of PowerPoint edits.

Native image deduplication, image sizing, and package compression happen automatically in production generation. The skill should not need separate image-processing scripts for normal builds.

### Inputs to this plan

- Operator discussion and corrections on October 3, 2026.
- Local draft: `11-integrated-presentation-skill-operating-model.md`.
- Local draft: `12-collaboration-and-review-package.md`.
- Local personal voice profile: `ri.md`.
- Local brand synthesis: `DESIGN.md` and the branding workspace guidance.
- Research audits of packaging, authoring, installed releases, and discovery.

The four supplied input documents are local planning inputs, currently ignored by Git. This plan and implementation contracts are tracked separately through narrow ignore exceptions. The live brand site and its three example tabs remain an outstanding source review because browser admin-policy verification prevented access; the local snapshot is the current research basis.

## 2. Accepted architecture decisions

| ID | Decision | Consequence |
| --- | --- | --- |
| D01 | Shared design foundation is persistent and versioned | Tokens, grids, primitives, components, composites, frames, and assets are library resources |
| D02 | Shared templates are independent of a deck | Decks reference template identity and revision; updating a deck does not rewrite its shared template |
| D03 | One authored `deck.yaml` per deck | Ordered slide definitions and deck-local templates live in one human-editable source document |
| D04 | Compiled scene documents are derived | The current `wmds-foundation` schema is an internal deck-specific representation, not the shared design foundation |
| D05 | Custom slides use deck-local template/composition definitions | Use the common design vocabulary; preserve starter provenance; promotion to the shared library is explicit |
| D06 | Assets have stable identity and provenance | Shared asset IDs and project-relative custom paths are supported; originals survive output optimization |
| D07 | Production media optimization is native and automatic | Deduplicate identical embedded payloads, size appropriate raster derivatives, and compress the package |
| D08 | Search supports scenario and content structure | Purpose labels aid search; compatible zone structure can match across scenario families |
| D09 | Template selection may explore alternatives | Build and compare real-content alternatives where a meaningful choice exists; record selection reasons |
| D10 | Review and qualification are scoped | Implementation support, specimen review, and content-envelope evidence are distinct |
| D11 | PowerPoint collaboration shapes the source contract | Keep stable IDs, authored-field mappings, generated baselines, and asset lineage for future three-way reconciliation |
| D12 | Reconciliation implementation is deferred | This pass establishes contracts and baseline protection; it does not promise arbitrary PowerPoint-to-YAML roundtripping |
| D13 | Approvals identify a snapshot | Changed evidence or meaning invalidates only the affected work and dependent approvals/reviews |
| D14 | Skill, compiler, metadata, and resources ship together | Projects pin a coherent release; an offline package is optional |
| D15 | New templates enter through a continuing intake track | Publish versioned additions/revisions; existing decks retain pinned definitions until intentionally migrated |

## 3. Historical Wave 1 baseline and current implementation

The measurements and gaps below describe the starting Wave 1 baseline. They are
retained as research history, not the current implementation state. Current
Wave 2/3 capabilities and outstanding gates follow this baseline.

### Media

The latest 167-slide reference library is 196.84 MB, including 180.65 MB of media. It has 560 media parts but 111 unique payloads. All 1,599 ZIP entries use STORE. Exact media deduplication plus DEFLATE was estimated at 78.53 MB without changing image quality. This is a research estimate, not an accepted optimized specimen.

Eleven unique JPEGs alone account for 72.73 MB. One 6000 x 4002 image is shown at approximately 5.25 x 3.5 inches. This motivates placement-aware raster derivatives.

The 17.6 GB development archive contains repeated JSON, PPTX, PNG, SQLite, and other artifacts. Smaller final decks and development artifact retention are separate tasks.

### Authoring

- Modern template content and explicit scene authoring currently use strict JSON.
- Templates bind content into a deck-specific foundation scene; the Go renderer emits native editable objects.
- YAML authoring and the unified deck/project contract are proposed additions.
- PowerPoint edits currently do not synchronize back to semantic source.
- Fixed inputs largely determine geometry and typography. Byte-identical output is not yet a public guarantee because metadata and some identifiers vary.

### Catalog and distribution

- The modern 167-design JSON catalog and legacy SQLite library are separate.
- Source purpose/use/budget/slot annotations exist but are incompletely exposed by modern discovery.
- Current legacy contract search defaults to a restrictive qualification state.
- The repository/staged release is local.6; the installed command and skill still point to local.5.
- Existing modern templates mostly expose fixed topology/content bindings. A customizable starter requires an explicit deck-local composition route.

### Frames

The reference library has one physical master and 168 layouts. Shared frame layouts can improve maintainability. Frame sharing must account for resolved chrome/geometry and typography record ownership; it is separate from the primary media size reduction.

### Current Wave 2/3 state

- Strict authored YAML and project operations are implemented in `internal/deckproject` and `design project`: init, check, build, status, resume, approve, fork, detach, review and export. The executable starter is `examples/deck-project`; the earlier planning YAML remains historical/illustrative.
- Unified native Go SQLite discovery covers modern templates, primitives, components, composites, resolved frames and assets, plus optional legacy inventory/contracts. Read-only inspect, hash-checked previews, explained shape/scenario ranking and actual-content candidate fit builds are implemented.
- The actual combined legacy/modern index experiment shrank from approximately 113 MB to 60 MB by retaining searchable summaries and lazily reading hash-pinned original legacy JSON rows. The original SQLite remains a required resource for full legacy inspection.
- `wmds-library.v3` freezes all **248 committed source definitions** at `e91e0d7771000b7386f1ea52f51252f0f0a134fd`. Five new native source adapters and the expanded family variants compile; immutable v2 remains available. The Go typography engine remains `wmds-go-foundation.v2`; a new source bundle does not imply a new typography calibration.
- Shared frame/chrome signatures reuse layouts with independent typography records. Native media optimization remains automatic in modern builds; authored originals remain intact.
- The repository skill pack now provides progressive references for source/resumption, framing, narrative/copy, discovery, actual-content alternatives, review boundaries and provisional personal/brand voice.
- Focused tests, race checks and CLI experiments have passed. Native export of the 248-slide reference candidate has succeeded; every-page review and fixes are in progress. Final native acceptance, coherent packaging/installer validation and installed activation remain pending. No successful build establishes an arbitrary-content envelope.

## 4. Source and project responsibilities

| Artifact | Authority and purpose |
| --- | --- |
| Versioned shared design/library resources | Shared definitions, metadata, fonts, and assets; JSON may remain the published representation |
| `deck.yaml` | Authoritative visible content/data, slide order, template selections, local templates, asset references, and stable IDs |
| Project state and context | Audience, communication job, reading mode, working argument, evidence, open issues, next action, and decisions |
| Source/claim index | Original locations, source IDs, locators, authority, permitted use, and material caveats |
| Toolchain lock | Skill/compiler/library/font/calibration/asset identity and versions |
| Generated content/reviewer packets | Derived snapshots; never independently maintained copies of slide copy |
| Compiled scene and binding/layout reports | Derived build artifacts and diagnostics |
| Baseline and object mapping | Source/build/output hashes, stable identities, and object-to-authored-field provenance |
| PPTX/PDF | Deliverables with generation, native, and visual status recorded separately |

Create stage-specific project artifacts only when useful. Avoid empty scaffolding and duplicated authoritative content. A maintainer package may include internal strategy and evidence. A client export includes the intended shareable deliverables. An offline package includes a pinned runtime and required dependencies where needed.

## 5. Workstreams and ownership

Four concurrency slots permit three implementation agents plus the primary integration agent. Use disjoint file ownership, coordinate shared interfaces before edits, and send small integration updates.

| Workstream | Owner | First deliverables | Dependencies |
| --- | --- | --- | --- |
| W1-A Native media | `wave1_media` | Go optimizer, production integration, provenance report, delivery profile contract | Existing package/media geometry |
| W1-B Deck/source contracts | `authoring_contract_research` | Proposed schema, single-file YAML example, state/baseline/receipt contract | Accepted decisions D01-D06 and D11-D14 |
| W1-C Catalog/discovery | `skill_catalog_research` | Existing metadata exposure, structural projection/search, SQLite adapter contract | Pinned design source and D08-D10 |
| Integration and plan | Primary agent | API coordination, tracked plan, compile integration, status/evidence | All Wave 1 outputs |
| Template intake | Primary agent initially; dedicated agent in a later wave | Delta inventory and intake/migration contract | Updated design source plus catalog identity |
| Skill and voice | Later agent | Progressive references, structure/copy profiles, reviewer exports | Deck/state/discovery contracts |

## 6. Wave 1: media implementation and source/discovery contracts

### W1-A — Native media optimization

- [x] Add a reusable native Go package optimization API with explicit options and a machine-readable report.
- [x] Deduplicate byte-identical compatible media across slides, layouts, and masters.
- [x] Rewrite relevant internal relationships while retaining valid SVG/PNG fallback associations and content types.
- [x] Aggregate displayed usage, group transforms, and crop requirements before selecting derivative dimensions.
- [x] Resize photographs conservatively; preserve transparent/lossless graphics and vector artwork.
- [x] Preserve original source assets and record source/derivative hashes, dimensions, placement requirements, and reuse counts.
- [x] Enable automatic delivery optimization and ZIP compression in the modern production build path.
- [x] Define predictable behavior for unknown/unsupported placements or image formats, retaining originals where sizing cannot be established.
- [x] Document opt-outs/quality profiles and the distinction between source assets and embedded output.
- [x] Compile integrated application changes and record what was and was not checked.

Initial implementation direction: `pptx.OptimizeMedia` returns package bytes and a report; a delivery profile starts with JPEG resizing at 220 pixels per inch and quality 90, plus exact deduplication and compression. Profile defaults are recorded here and in the native API contract. Existing low-level writer defaults remain compatible; the production route explicitly opts into delivery behavior.

Acceptance: implemented native API; no external image tool; original assets unchanged; explicit supported-format/geometry behavior; report records optimization decisions; production route invokes it. Package integrity, actual size reduction, PowerPoint opening, and visual quality remain evidence gates for a reviewed specimen, not claims inferred from compilation.

### W1-B — Deck and project contract (historical Wave 1 delivery)

- [x] Define `pptxgengo.deck-document.v1` and a JSON Schema for YAML-compatible authored values.
- [x] Define one ordered `slides` array with stable identities and shared/local template references.
- [x] Define typed content binding and asset reference unions, including custom project-relative assets.
- [x] Define deck-local templates/compositions using the common authoring vocabulary and starter provenance.
- [x] Document shared design resources versus authored document versus compiled representation.
- [x] Define project state, approvals, dependency invalidation, and resumption responsibilities.
- [x] Define generated baseline and object-to-source mapping needed for later PowerPoint reconciliation.
- [x] Define portable maintainer/client/offline export expectations and toolchain locks.
- [x] Provide one illustrative multi-slide `deck.yaml`, clearly marked as proposed until its runtime adapter exists.
- [x] Record cross-field/semantic checks that a later Go loader must enforce beyond JSON Schema.

Acceptance: concrete schema and readable example; no independently edited duplicate slide copy; no per-slide source file requirement; custom starter adaptation is explicit; reconciliation and proposed commands are not presented as existing capabilities.

### W1-C — Catalog and discovery contracts plus initial implementation (historical Wave 1 delivery)

- [x] Expose existing source purpose, use cases, word budgets, and advisory slot guidance.
- [x] Project structural zones, content roles, cardinality, visual forms, and relationships with explicit provenance.
- [x] Separate structural compatibility from scenario terminology and source-example content.
- [x] Provide explainable modern search across scenario/free text and structural hints.
- [x] Separate discovered/render-supported/reviewed/envelope-qualified state.
- [x] Keep exact binding constraints visible and retain supported custom-composition fallback.
- [x] Specify the unified SQLite projection, source hashes, freshness, and detail lookup contract.
- [x] Identify the boundary between Wave 1 modern search and later SQLite unification.
- [x] Define continuing template intake identity/revision/replacement handling.
- [x] Compile CLI/API additions and record remaining capability gaps.

Acceptance: agent can retrieve rich modern metadata and search by structure across scenario labels; reasons and constraints are returned; no successful-search or successful-build claim implies content fit; published metadata remains rebuildable and pinned. Full SQLite integration is a later adapter unless a bounded Wave 1 implementation is completed and explicitly recorded.

### Wave 1 integration gate

- [x] Review all agent changes against the accepted architecture.
- [x] Confirm ownership boundaries and source/asset/catalog identity contracts agree.
- [x] Record build status and remaining native/visual/evidence checks.
- [x] Update this plan with actual files, available commands, deferred items, and next tasks.
- [x] Keep pre-existing slide feedback changes intact.
- [x] Keep generated experiment output outside the cleaned sample set.

## 7. Wave 2: authored document runtime and production workflow

Implemented after Wave 1 contracts were reviewed. The checked items below describe the bounded runtime, not every broader workflow aspiration.

### Authoring CLI and project package

- [x] Load strict human-authored YAML into the canonical typed deck representation.
- [x] Preserve scalar text, block breaks, and source locations; diagnose duplicate keys, unknown fields, unsupported values, and invalid references.
- [x] Compile shared-template slides and deck-local custom compositions through a common native renderer.
- [x] Resolve shared/custom assets from portable project references.
- [x] Emit source snapshots, compiled scene, binding/layout/media reports, and build receipts.
- [x] Persist stable object mapping and generated baseline; detect external output changes before replacement.
- [x] Provide stage/status/resume and maintainer/client/offline package operations.
- [x] Specify deterministic IDs and metadata policy; pin fonts/calibration/assets/library/engine.

### Unified catalog and template selection

- [x] Index modern entities and legacy inventory through a documented SQLite discovery surface.
- [x] Add detailed inspect/preview and shape queries for roles, relationships, visual forms, capacities, and supported adaptations.
- [x] Rank candidates with explained tradeoffs and real-content fit information.
- [x] Generate two or three materially different alternatives for uncertain/representative slides.
- [x] Record selections and custom-slide rationale without a reuse quota.

### Frames and image policy follow-up

- [x] Share layouts by resolved frame/chrome signature.
- [x] Refactor layout typography ownership to avoid last-slide-wins records.
- [x] Document which elements remain editable per slide and which belong to shared layouts.
- [x] Apply placement-aware derivatives across modern foundation, bound-template, reference and YAML project production builds.
- [ ] Future media expansion: PNG photographs, color-managed conversion and other legacy production routes; conservative source preservation remains supported.
- [x] Add artifact retention/current-build rules independently from output image optimization.

## 8. Wave 3: integrated skill, template intake, and release

### Updated design-library implementation track (parallel with Wave 2)

The v3 bundle freezes **248 committed definitions: 81 additions, one changed definition (`about/glance`), and no removals** against the pinned 167-definition v2 bundle. Source commit is `e91e0d7771000b7386f1ea52f51252f0f0a134fd`; the eight formerly uncommitted core additions are now committed and included. This supersedes the intermediate 193-definition and 240-plus-eight-working-tree inventories. Forty-six retained identities moved family files; identity remains independent of pathname.

The changelog's 143-addition summary uses the older v1 pin and cannot be used as the v2 implementation delta. Tokens, frame JSON, and component JSON are unchanged against v2; template coverage of existing frame combinations has expanded. Five new scene node types now have native adapters: `logoslot`, `dotmap`, `teamcurve`, `device`, and `plane`.

Start with shared capabilities, then compile independent template families. Keep this work parallel to authored YAML/project implementation so new templates do not block source-maintenance work.

| Lane | Assignment / exclusive edit ownership | Dependencies and evidence |
| --- | --- | --- |
| Integration | Source-pinned candidate bundle, source loader/manifest, scene dispatcher, shared binding/discovery registration | Keep v2 intact; integrate only reviewed adapters; no installed activation implied |
| Logo and architecture agent | New logo/device/plane adapter files and tests | Contain-fit logos and editable placeholders; editable device groups and isometric planes; connection anchors, surfaces, text fit and z-order |
| Geography agent | New map adapter, geographic data/provenance, and tests | Matching public-domain outlines and geographic panels; native office/hub markers; reusable map base; cap excessive shapes |
| Team-curve agent | New curve adapter files and tests | Stacked smooth bands, phase-label controls, value validation and curve geometry; review native freeform support before choosing representation |
| Family compilation agents (after adapters settle) | Disjoint family fixture/refinement files | Compile added variants using existing frames; update `about/glance`; preserve stable template/node identities |
| Independent review | Read-only code review plus focused tests and representative generated specimens | Bugs, fit boundaries, namespace/package integrity, CPU/memory, visual/source parity; findings go to owning agent |

With four active slots, integration coordinates three capability agents first; recycle slots for family compilation and independent review. Shared dispatcher/catalog edits stay with integration to avoid conflicting edits.

- [x] Audit latest source against the actual v2 pin and distinguish the changelog baseline.
- [x] Freeze a new candidate snapshot with template, renderer/geography, asset, and source hashes; preserve published v2.
- [x] Specify accepted field/value contracts and explicit rejection cases for all five adapters.
- [x] Implement logo slots and map panels/markers with provenance and resource bounds.
- [x] Implement stacked team curves and phase options with validated data and editable geometry.
- [x] Implement device and plane scene nodes with connection anchors and safe text layout.
- [x] Extend bindings/discovery roles and visual forms for the new nodes.
- [x] Compile all 81 additions and revised `about/glance`, then run actual-content binding smoke cases.
- [x] Validate all supported frame combinations; do not reimplement unchanged frames solely because the changelog lists them.
- [x] Review code independently and fix findings before integration.
- [x] Generate one current reference library with media receipts and perform native visual review.
- [x] Publish migration notes and catalog evidence with the coherent release.

Exact additions, source locators, capability requirements, and family inventory are maintained in [the intake contract](template-intake-contract.md).

### Skill and content workflow

The skill describes stage-specific audience review packets and logging. The current
`project review` ZIP is a technical build/draft packet; it does not automatically
assemble the isolated outline/content/audience packets or run fresh reviewers.
Those packets are constructed according to the skill references when useful.


- [x] Update one `west-monroe-presentations` entrypoint with focused trigger conditions.
- [x] Route progressively into project/resume, intake/frame, narrative/copy, selection/composition, source/build, and review references.
- [x] Operationalize the seven-stage draft workflow with meaningful approvals and resumable state.
- [x] Generate isolated reviewer packets from an explicit authored snapshot; record technical decisions without fabricating operator content approval.
- [x] Keep source verification separate from audience comprehension review.
- [x] Preserve original findings, author responses, operator decisions, and recheck status.

### Voice

- [x] Derive separate personal slide structure and slide copy references from `ri.md`.
- [x] Keep technical precision, useful explanation, and evidence-backed differentiation.
- [x] Omit email signoffs and unrelated interpersonal registers from slide defaults.
- [x] Consolidate stale/contradictory profile annotations and retain provisional status.
- [ ] Review the live brand voice main page and all three example tabs, or supplied exports.
- [x] Blend brand expression and personal style with audience/readability needs; use versioned WMDS tokens for slide implementation.
- [x] Include illustrative paired slide-structure and copy examples.
- [ ] Validate the provisional voice examples with operator refinements.

### Continuing template intake

- [x] Inventory additions/revisions against the last published design snapshot.
- [x] Extract scenario and structural metadata without mutating existing deck pins.
- [x] Implement missing native features and expose pending capabilities honestly.
- [x] Generate focused reference specimens and record review observations.
- [x] Publish replacement/revision information and explicit migration guidance.

### Release and pilot

The integrated two-slide pilot uses the supplied operating-model/review documents,
changed meeting evidence, independent audience copy review, actual-content alternatives
and a local composition. Maintainer and offline packages rebuild byte-identically
after relocation. Both pages passed native review; all four comprehension findings
were repaired and rechecked. This remains an internal maintainer pilot, not a client
acceptance study. A fresh agent session was unavailable under the session thread limit;
the independent copy reviewer had prior implementation context.


- [x] Package aligned compiler, skill, catalog, font/calibration, asset, and gallery resources.
- [x] Correct the installed-versus-repository mismatch through reviewed local.7 activation.
- [x] Run an integrated source-backed pilot when implementation verification is requested.
- [x] Include interruption/resumption, changed meeting evidence, a local custom slide, and real-content design alternatives.
- [x] Establish optimized size and native/visual review evidence for representative and full-library output.
- [x] Publish remaining limits and reproducibility conditions.

## 9. Deferred reconciliation workstream

This is design input for this pass, not an automatic expansion of implementation scope.

- [ ] Compare authored source and edited PPTX against the exact generated baseline.
- [ ] Propose supported text/data/image changes for promotion into YAML.
- [ ] Identify conflicts when both sources changed.
- [ ] Preserve unsupported geometry/art edits as explicit native/manual variants.
- [ ] Carry accepted changes into source and regenerate only through a reviewable operation.
- [ ] Keep full arbitrary semantic reconstruction a distinct capability with explicit limits.

## 10. Execution evidence and check policy

The operator authorized this plan, Wave 1, and then execution of the remaining plan. Existing slide corrections are unrelated working-tree changes and must be preserved. No commit, push, installed-release activation, or sample replacement is implied by this task.

Repository instructions prohibit adding/running tests unless the operator requests testing or implementation verification. On October 3, after implementation delivery, the operator explicitly requested smoke tests and dedicated code review for bugs/performance/memory. Focused tests, real-library smoke runs, and performance/memory measurements are now authorized. Native/visual evidence is recorded separately from package and compile checks.

Keep implementation experiments in temporary output directories. Do not repopulate `samples/` with intermediate decks. Keep long-lived metadata/contracts in tracked source locations rather than disposable capture folders.

## 11. Progress log

Earlier rows record the state at that point in the work. Uncommitted-draft and
missing-runtime statements in historical rows are superseded by the current rows.


| Date | Work | State / evidence | Next action |
| --- | --- | --- | --- |
| 2026-10-03 | Research and operator feedback | Architecture decisions D01-D15 recorded; size/catalog/source gaps identified | Start Wave 1 |
| 2026-10-03 | Wave 1 delegation | Three agents assigned native media, deck/source contract, and catalog/discovery; source schema remains proposed | Implement, integrate, and record outcomes |
| 2026-10-03 | Durable plan | This file created; narrow Git ignore exceptions expose implementation planning files | Update task states after agent delivery |
| 2026-10-03 | Continuing source intake | Clean design repo at `f0c8a5a`: 193 definitions, 26 additions and one revised `about/glance` versus pinned 167; no identities removed | Review new primitives/adapters in a subsequent intake slice |
| 2026-10-03 | W1-A delivered | Native `pptx.OptimizeMedia`; automatic modern WMDS policy; template/foundation policy propagation; per-part receipt; conservative JPEG derivatives | Actual size/package/native/visual evidence pending; other authoring routes and PNG photo derivatives deferred |
| 2026-10-03 | W1-B delivered | Proposed deck JSON Schema, three-slide YAML example, source/project/state/approval/baseline contracts | Implement YAML loader, shared definition references, and baseline/object maps in Wave 2 |
| 2026-10-03 | W1-C delivered | Source advisory annotations and derived shape metadata; `design library-search` with item-role hints, vocabulary and explained ranking | Unified SQLite projection, fit-aware selection, and specimen evidence import remain next-wave work |
| 2026-10-03 | Compilation | All `./cmd/...` binaries compiled successfully using cached Go 1.27.1 dependencies to `/tmp/pptxgengo-wave1-bin/`; no tests or behavior/native runs performed | Preserve evidence gates and current sample/installed release state |
| 2026-10-03 | Review requested | Operator authorized smoke testing and independent bug/performance/memory review; two reviewers assigned media and catalog code | Run focused checks, profile real-library optimization, and fix findings |
| 2026-10-03 | Review fixes and smoke evidence | Relationship/content-type namespace and quoted-attribute bugs fixed; semantic-group/comparison/scenario-ranking fixes covered by regressions; all command binaries compile; automatic/default and explicit build policies verified | Native optimized-image appearance remains a separate acceptance gate |
| 2026-10-03 | Full-library optimization | 196,841,070 → 31,550,932 bytes; 560 → 111 media parts; eight unique JPEG derivatives; 12.58–13.11 seconds, 650–746 MB peak RSS after memory changes; all 1,756 internal relationships resolve | Continue streaming/resource work as scale requires; keep source assets |
| 2026-10-03 | Existing test baseline | Broad `pptx` run has the same 12 failing functions as isolated committed HEAD; focused media/discovery checks pass, including media race check | Repair existing baseline separately; do not claim the full suite is green |
| 2026-10-03 | Updated design-system steering | Operator reports newer changelog, capabilities, and templates; independent intake audit assigned against our actual v2 pin rather than the changelog's older v1 baseline | Record latest delta and dependency-first parallel ownership |
| 2026-10-03 | Refreshed intake audit | 240 committed + eight working-tree draft definitions = 248; 81 additions/one revision versus pinned v2; five missing node adapters; exact family lists/source locators and prerequisite lanes recorded | Freeze reviewed source and run logo/device/plane, geography, and curve lanes before family compilation |
| 2026-10-03 | Frozen v3 intake | All 248 definitions committed and vendored at `e91e0d7771000b7386f1ea52f51252f0f0a134fd`; five adapters and frame variants implemented | Final native findings/review gate |
| 2026-10-03 | Wave 2 project runtime | Strict YAML, shared/local composition, asset/ancestry checks, hash-bound approvals, immutable builds/maps, resume and maintainer/client/offline/reviewer exports implemented | Finish integration/pilot evidence; arbitrary PPTX reconciliation deferred |
| 2026-10-03 | Unified catalog and alternatives | Native Go SQLite, modern/legacy namespaces, read-only pins, lazy full legacy inspect, previews and actual-content candidate builds; actual index 113 → 60 MB | Rebuild coherent packaged catalog with final gallery evidence |
| 2026-10-03 | Skill and voice pack | Focused entrypoint and progressive references implemented; provisional structure/copy and local brand synthesis included | Live brand main page/three tabs or supplied exports still needed |
| 2026-10-03 | Final focused catalog review | Gallery key/revision and path bounds fixed; catalog/discovery race tests pass (38.646 s), CLI catalog race tests pass (4.350 s); real-content 2 successful alternatives/1 retained failure | Final native review and installer checks remain pending |
| 2026-10-03 | Native candidate review | 248-slide native PDF exported; assigned page ranges being reviewed; pages168–248 reviewed and slide206 finding handed to integration | Re-export changed specimens and finish all-page review before native acceptance |

## 12. Open items

| Item | Owner | Effect / default |
| --- | --- | --- |
| Live brand examples unavailable | Primary agent/operator source provision | Local voice guidance is provisional until all examples are reviewed |
| Delivery image profile | Media agent/integration | Start conservatively at JPEG 220 PPI, quality 90; preserve originals and record decisions |
| Template expansion timing | Intake track | New definitions enter as versioned revisions; published deck builds do not silently migrate; dated 369-definition observation is active |
| Final native and release acceptance | Integration/review agents | Completed for frozen v3/local.7; each incoming revision has its own acceptance gate |
| Future PowerPoint reconciliation | Deferred workstream | Stable identity/baseline design included now; importing manual changes later |

## 13. Implementation references

- [Deck/source contract](deck-source-contract.md) — original design plus implemented runtime boundary.
- [Deck schema](../../schemas/deck-document-v1.schema.json) — authored-document structure; runtime applies bounded semantic checks.
- [Historical planning YAML](examples/deck.yaml) — illustrative original proposal; [executable starter](../../examples/deck-project/README.md) is the current runtime example.
- [Catalog/discovery contract](catalog-discovery-contract.md) — metadata, structural search, and SQLite adapter boundaries.
- [Media optimization contract](media-optimization-contract.md) — native defaults and provenance behavior.
- [Template intake contract](template-intake-contract.md) — latest incoming 248-definition snapshot, versioning, and parallel intake responsibilities.
- [Wave 1 review evidence](wave1-review.md) — independent findings, fixes, smoke checks, size/memory measurements, and existing baseline failures.

Historical Wave 1 runtime additions:

- `pptx/media_optimization.go`: reusable native optimizer and delivery policy.
- `internal/wmdesign/render.go`: automatic final-package optimization and receipt.
- `internal/wmdesign/template_bindings.go`: bound source policy propagation.
- `internal/wmdesign/library_templates.go`: source advisory annotations and discovery projection.
- `internal/wmdesign/library_discovery.go`: structural vocabulary, groups/zones, and explained ranking.
- `cmd/pptxdesign/library_search.go`: repository `library-search` command.
- `schemas/deck-document-v1.schema.json`: authored-document schema, now consumed through strict project loader validation.

The repository CLI additions require a coherent release containing these changes.
`release/VERSION` is local.7; installation and activation passed for the frozen
248-definition revision. Incoming capabilities remain unreleased until reviewed. The Wave 1 local.5 mismatch is historical. No commit or
push is implied by implementation.

Wave 2/3 implementation references:

- [Project runtime](../../internal/deckproject/README.md) — actual source/node/context/export bounds.
- [Executable project](../../examples/deck-project/README.md) — runnable YAML starter and CLI sequence.
- `internal/wmdesign/library_index*.go` and `cmd/pptxdesign/library_index.go` — unified SQLite and candidate fit operations.
- [Packaged skill](../../skills/west-monroe-presentations/SKILL.md) — progressive content-to-deck workflow.
- `internal/wmdesign/scene_intake_*.go` — new v3 source adapters.
- `library/wm-design-system/v3/bundle.json` — frozen source identity and pending candidate acceptance status.

### Historical Wave 1 limits and their current disposition

- Wave 1 full-library size and package integrity were measured. Native v3 candidate review is now in progress; final acceptance has not yet been recorded.
- JPEG ICC/Adobe/orientation-sensitive and unmodeled placements retain their source payloads; PNG photo resizing/color-managed conversion remain future work.
- Optimization rebuilds the package in memory. STORE payload views, cached hashes, bounded resampling scratch, 256 MiB part/2 GiB package limits, and a 64-million-pixel decode guard reduce exposure; streaming and whole-build memory budgets remain future work.
- Registered shared frame/grid/component references and bounded local adapters now exist. The old planning example remains illustrative; use the executable project starter for current syntax/resources.
- Structural search still derives affordances rather than measured fit. Unified SQLite is implemented; `library-fit` tests explicitly supplied candidate content and leaves native/visual review pending.
- YAML parsing/building, project exports and new-template import are implemented. Arbitrary runtime reconciliation remains deferred; installed activation and final native/visual acceptance remain separate pending gates.

### Wave 1 smoke/review follow-up

- [x] Add/run targeted media relationship, deduplication, resizing, and preservation tests.
- [x] Smoke modern search against the actual pinned catalog, including cross-scenario four-point queries and invalid inputs.
- [x] Exercise the automatic build path and media policy propagation.
- [x] Measure optimizer time, output size, and memory on the full retained template library using temporary outputs.
- [x] Review media code independently for bugs, performance, and memory issues.
- [x] Review catalog/search code independently for bugs and metadata correctness.
- [x] Fix material findings and rerun affected checks.
- [x] Record measured results, remaining visual/native limits, and the next action.

## 14. Final integration evidence and newly arriving source

See [Waves 2–3 review](waves2-3-review.md) and `release/verification-wmds-v3.json` for checks, native page hashes, repaired specimens, source-backed pilot and exact limitations. The current 248-slide reference is 31.8 MB and native PDF 8.64 MB. Focused/race checks pass; the broad pptx baseline still has its twelve pre-existing failing functions.

A final intake check observed additional upstream work after v3 was frozen: current HEAD `c7678b0efd48e96a2be290fbdee6ed8914a604a1` has 301 committed definitions; the working tree has 337. Against v3, the working inventory adds 89 and revises 13, with no removals. Thirty-six definitions are not yet committed. These evolving additions are a **separate incoming batch**, not relabeled as implemented in local.7. The [delta observation](next-intake-delta-20261003.json) records canonical keys and working-tree state. The initial node-type scan read the wrong path; the corrected observation below supersedes that statement. Nested fields/value contracts and revised team curves require capability review. Source snapshots remain immutable; reviewed projects do not silently migrate.

Next intake: freeze an accepted committed snapshot, audit the 53 committed additions and changed contracts, implement/compile/reference-review independent families in parallel, then qualify the 36 working additions once ready. Live brand examples and operator voice refinements remain open. Arbitrary PowerPoint reconciliation and full-build streaming remain explicitly deferred.

| 2026-10-03 | Final native and release gate | 248 source specimens reviewed, eleven repairs rechecked, source-backed pilot native check passed; local.7 stage integrity and outside-repository commands/relocated builds passed, aligned launcher and skill activated | Next upstream intake remains separately tracked; live voice examples/operator refinements pending |


## 15. Continuing intake: capabilities before template expansion

**Active as of 2026-10-03 17:50 UTC.** The design-system source is still being
edited by several agents. This is an observation of available work, not a claim
that upstream authoring is complete. The installed local.7 release remains pinned
to the reviewed 248-definition v3 library.

The corrected [intake observation](next-intake-observation-20261003.json) reads
`template.slide.body`, rather than the incorrect top-level `template.body` used
in the previous audit. It records **301 committed definitions and 369 working
ones** at HEAD `c7678b0efd48e96a2be290fbdee6ed8914a604a1`. Against v3, the
working tree adds 121 and revises 13, with no removals. The 68 additions beyond
HEAD are working drafts. These counts are dated and will grow as authoring
continues. A double-read hash guard checked the observed source files before
copying them into `planning/wm-design-contracts/v4/intake-20261003/`; this snapshot
is not a published v4 bundle.

### Capability lanes and ownership

| Lane | Source contract / implementation | Owner | State |
| --- | --- | --- | --- |
| Venn | Editable two-, three-, four-circle diagrams; set and intersection labels, bullets, numbered points, lens anchors, opacity | Authoring contract agent | Implemented; twelve frozen node fixtures and editable XML checks pass; full-template qualification pending |
| Maturity | Exponential curve, numbered stages, active stage, axis/inflection, branch, headroom and “you are here” | Proof/evidence agent | Implemented; eight frozen scene fixtures, geometry and editable XML checks pass |
| Revised team curves | Shape-preserving monotone cubics; explicit Catmull option, area and dashed line series | Proof/evidence agent | Implemented and checked; frozen v1/v2/v3 default remains Catmull; later revision adoption is separately gated |
| Tables and heat maps | Plain cells with subtitle; dots with optional text, ink precedence; per-cell heat ramp/value/label; block heat and legend swatches | Catalog agent | Implemented; saturation, paragraph leading, gutters, scalar footprints and linear assembly fixes reviewed and tested |
| Straight arrow | Existing pinned hand-drawn asset, rotation and horizontal flip; compare source registry and emitted transforms | Integration | Source registry matches; intrinsic-height/rotation bounds and multi-turn native angle safety fixed and checked |
| Integration | Scene dispatch/local composition, content bindings, structural discovery, source-backed smoke deck and independent review | Primary agent | Implemented; normal/race checks pass, existing 248 slides rebuild; incoming specimen/native qualification remains pending |
| Slim footer | New source footer and slim source-zone bottom | Next intake integration | Observed in frame delta; requires footer/source placement and project reference review before adoption |

Heat-map backgrounds are now described by the source contract. A separate
arbitrary background field on every ordinary table cell is not yet defined in
the observed source; adopt that when the authoring contract arrives rather than
invent an incompatible field.

### Gates for this capability slice

- [x] Observe committed and draft inventories separately; preserve exact source hashes.
- [x] Correct node-type inventory: new body types are `venn` and `maturity`.
- [x] Implement the ready capability planners and source field contracts.
- [x] Connect local composition, typed content binding, and structural catalog metadata.
- [x] Smoke build representative frozen source nodes and verify editable native XML output. Full-template diagnostic: 43 compiled, 20 rejected from 63 selected definitions; see evidence below.
- [x] Independently review new code for bugs, resource bounds and geometry errors; fix material findings.
- [x] Run targeted and affected package/race checks; record exact evidence and limits.
- [ ] Native review of the new capability deck before specimen acceptance. Operator is actively reviewing the reference deck; no application interruption performed.

### Subsequent template import

After capability acceptance, import the ready committed definitions in family
lanes, retaining canonical keys and source hashes. Compile and review each source
specimen, update the searchable catalog and previews, then create a coherent new
release. Working drafts can enter a later dated snapshot when ready. Repeat the
inventory at each intake boundary; do not overwrite published revisions or
silently migrate existing project pins. This separates continuous authoring from
accepting a reproducible library snapshot.


### Capability implementation evidence

See [capability slice and remaining fit work](intake-capabilities-20261003.md),
`planning/wm-design-contracts/v4/intake-20261003/receipt.json`,
`smoke-results.json`, and `independent-review.json`. All observed source files
still match the frozen manifest. The authoritative 17:50 snapshot renderer hash
is `a7f1f046d38723cb86d6e0dbacd97886c183784df1e4949192322e74d6a33cf6`;
maturity family hash is `63eca0027e5797a19be2b1236986772960889f2bb4ad1bb18a93a08416145c5b`.
Earlier agent observations used older hashes; these do not override the actual
snapshot manifest. Tests use embedded/frozen fixtures, never the live working tree.

The source already had headroom and “you are here” in that frozen renderer.
Table columns with semantic `type: maturity` are data definitions, not scene
nodes; the snapshot contains exactly eight maturity scene nodes. The frame delta
also introduces `slim` footer (rule 495, row 504–516, body bottom 486) and slim
source bottom 489. Tokens are unchanged from v3.

Normal checks passed for `internal/wmdesign`, `internal/deckproject`,
`cmd/pptxdesign`, and `cmd/pptxgengo`. Full affected race checks passed (62.648 s,
61.650 s, and 6.400 s respectively for the three tested packages). The frozen
248-slide v3 source reference regenerated successfully after all fixes.
The incoming 43-slide prototype is 395,211 bytes, with 300 valid internal
relationships and editable native tables/text/geometry. These are implementation
and package checks, not native visual acceptance. No new release was installed
and no commits or push were made in this intake slice.


## 16. Frozen catalog intake, parallel repairs and expanded capabilities

Upstream authoring stopped for now. Target the clean immutable October 3 source
commit `8f9f16ab8a7e2a6fff45a97627308668f4a12753` (522 templates, 24 families).
Against accepted v3: 274 additions, 16 revisions, no removals; 16 family moves.
This supersedes the current target counts in Section 15, preserving its dated
observations. The accepted/installed v3 library remains 248 templates.

### Execution and gates

- [x] Capture and hash-verify all 35 observed source files; retain earlier snapshots.
- [x] Implement funnel, revised pyramid, bracket, road, cycle and standalone gauge.
- [x] Update Gaussian team curves, Venn/maturity placement, slim footer/title/tint,
  table row headers/numbers, statuses, callouts and dense/signed value line charts.
- [x] Wire scene dispatch/local composition, typed bindings and structural discovery.
- [x] Fix all 27 known incoming repair cases with copy-preserving, atomic,
  idempotent, future-revision amendments; frozen v3 behavior unchanged.
- [x] Independently review resource bounds, geometry, footnotes and native XML;
  fix cache identity, tint validation, cycle copy measurement and road joins.
- [x] Full-slide diagnostic: 522 attempted, 495 compiled, 27 newly exposed failures.
- [x] All-node diagnostic: 6,307 nodes, 63 fit/data failures in 26 templates,
  zero unsupported scene types. Standalone gauge passes its own planning/XML checks.
- [x] ZIP/XML/internal relationship checks on the 495-slide diagnostic deck.
- [x] Affected normal tests, focused race/XML checks and 248-slide v3 rebuild.
- [x] Close the new 27-template fit/data queue and all 63 affected nodes;
  preserve native missing observations without changing source data (see Section 17).
- [ ] Native PowerPoint review and visual acceptance of incoming capabilities,
  repair specimens and feedback-loop dash treatment.
- [ ] Complete versioned incoming bundle and matching typed bindings/catalog/previews.
- [ ] Regenerate/qualify reference decks and coherent installed release/skill update.

See [frozen intake capabilities, failure queue and evidence](intake-capabilities-20261003-frozen.md).
Machine evidence lives in `planning/wm-design-contracts/v4/intake-20261003-frozen/`:
`observation.json`, `smoke-results.json`, `node-results.json`, `node-receipt.json`,
`receipt.json`, `package-integrity.json`, `independent-review.json` and
`gauge-review.json` and `checks.json`. Temporary prototype decks do not replace qualified samples.
No commit, push or release installation occurred in this slice.


## 17. Expanded-catalog repair wave and native review packet

The frozen target remains 522 definitions at source commit
`8f9f16ab8a7e2a6fff45a97627308668f4a12753`. All 27 newly exposed template
failures and all 63 node failures from Section 16 are closed.

- [x] Card lane: 43 nodes across 14 templates; retain fonts/copy, correct named
  padding, and place tinted icon cards in two252pt columns with an18pt gutter.
- [x] Table lane: five complete templates; correct column/row allocations with
  original type styles, preserve all data and validate source topology atomically.
- [x] Text/frame lane: seven templates; badge padding, Hypercare full-grid width,
  two-line timeline title and bounded48pt rows, caption/footer and chart/table gaps.
- [x] Native missing line-chart observations: blank workbook/cache positions,
  original category indices, observed zeroes and source-authoritative span policy.
- [x] Independent cross-reviews: fix sparse series alignment, combo indices,
  unsupported workbook topology, bounded timeline copy and Seven-R source widths.
- [x] Full-catalog build: **522 attempted, 522 compiled, zero rejected**.
- [x] Final node audit: **6,307 attempted, zero failures/unsupported types**.
- [x] Normal affected tests, full affected race checks, native XML/ZIP/relationship
  checks and accepted248-template source-reference regeneration.
- [x] Prepare full522 deck and focused50-slide review packet; final subset rebuild
  is byte-identical after guard fixes. Run capture script Bash/AppleScript syntax checks.
- [ ] Native PowerPoint visual inspection of all50 selected pages. Operator is
  running capture from Terminal; tool process launch/bootstrap access is broken.
- [ ] Repair any native visual findings and re-run affected evidence.
- [ ] Complete/accept versioned incoming source+assets bundle, content projection
  (including nullable chart values), SQLite metadata/previews and installed release.

Reference packet: `samples/wmds-incoming-20261003/latest/` contains
`WMDS-template-library.pptx` (522 slides,33,763,375bytes),
`WMDS-intake-review.pptx` (50 slides), reproducible diagnostic IR/indexes,
`review-packet.json` and `capture.sh`. These are unreleased review artifacts;
installed source pin remains v3/248. The current qualified samples are retained.

Evidence: `planning/wm-design-contracts/v4/intake-20261003-frozen/repair-wave/`
contains source/build/node results, package integrity, checks and independent
reviews. Source observations remain immutable. Full affected race passed
95.066s(wmdesign),48.703s(deckproject),4.906s(pptxdesign); focused sparse-chart
and card race passed1.302s(pptx),12.893s(wmdesign). Existing broad pptx golden
failures remain separate; entire repository suite is not claimed green.

Native-access diagnosis found a dead Mach bootstrap endpoint inherited by
command processes. LaunchServices returns−10827, direct AppleEvents return−600,
launchctl cannot reach the user/GUI domain, and computer-use startup fails.
User Terminal successfully returned4PowerPoint presentations. This is evidence
of different process access contexts, not a reason to reset TCC or close the
user’s documents. Native export scripts skip presentations with no saved path
and retain exact saved-path validation; operator capture is the current bridge.
No commit, push or installed-release migration occurred in this wave.


## 18. Resumed native feedback and v4 production integration

The522-template frozen intake remains immutable. Native PowerPoint access now
works through computer use; shell automation remains unavailable and Terminal
computer use remains explicitly blocked. No processes or permissions changed.

- [x] Maturity numbers use native middle alignment in their24pt marker boxes;
  stage label stacks reserve the actual frame body top. Frozen v3 stays unchanged.
- [x] Review all19 maturity and12 Venn variants in an expanded77-slide packet.
- [x] Complete native inspection of all77 pages and repair six findings:
  Seven-R compact headings, evidence-model follow-up spacing, two Here tags,
  maturity table row-header width, and a narrow maturity label/tick intersection.
- [x] Native v2 recheck: six changed pages accepted; other71 PNGs byte-identical.
- [x] Refresh full diagnostic:522/522 compile;6,307/6,307 node audit.
- [x] Create separately pinned v4 source/assets/fonts bundle from all35 source
  files; preserve historical revisions and installed local.7/v3.
- [x] Production catalog exposes522 definitions. Carry v3 amendments only for
 232 semantically unchanged frozen compositions. Explicitly recognize nine
  pinned family files without schema declarations in v4 only.
- [x] Closed nullable numeric content slots for v4 line-chart observations,
  including required missing source positions and observed zeroes. Focused
  catalog/retention/binding checks pass.
- [x] Production source reference522 slides and active bound reference521 slides
  generated through real pinned bundle and binding paths.
- [ ] Native qualification of production reference specimens and supplied-content
  pilot; matching SQLite metadata/previews and portable package qualification.
- [ ] Coherent new installed release and skill update after all release gates.

Native77-page evidence: `planning/wm-design-contracts/v4/
intake-20261003-frozen/repair-wave/native-feedback-review.json`. Reviewable deck:
`samples/wmds-incoming-20261003/feedback-20261003-v2/review/WMDS-feedback-v2.pptx`.
Production references: `samples/wmds-incoming-20261003/v4-production-v1/`.
Subset acceptance does not qualify all522 specimens. Full repository suite is
not claimed green; existing unrelated pptx golden failures remain. No commit,
push or installed release migration has occurred.

### Latest user review reopens visual acceptance

The earlier77-page review is retained as historical evidence. User feedback
on the earlier capability-reference packet requests AI Agents positioning,
Venn label vertical centering and larger three/four circle geometry with lens
text fitting, and roadmap numeral centering. Native acceptance is reopened
until a uniquely named fresh packet is checked. Here tag wrapping was already
fixed and natively checked in the v2 packet. Current upstream587 is being
observed separately from immutable522. User authorized parallel agent work.

### Composition and packaging follow-through

- [x] Audit installed skill/compiler: local.7 skill matched all18 repository
  files before candidate documentation update; original bounded composition
  and source-backed pilot were already shipped.
- [x] Create executable incoming local Venn/maturity/road examples with visible
  copy/data in maintained YAML slide values; no retained example build state.
- [x] Fix local maturity endpoint stroke allocation and component planner label
  headroom; focused regressions preserve honest overflow rejection.
- [x] Build all3 local examples and verify69 offline package hashes; relocated
  packaged compiler produces byte-identical PPTX. Native review remains pending.
- [x] Build uniquely named real-v4-source77-page feedback v3 packet with slide
  index; rebuild production source522 and active bound521.
- [x] Latest affected normal and race checks pass across4 packages.
- [ ] Recheck latest native packet, production references and composition pilot.
- [x] Complete isolated candidate staging;101legacytemplates,132component
  occurrences and10,183releasefiles validated. Installed release unchanged.
- [x] Outside-checkout candidate paths/find/inspect/preview and local composition
  init/check/build pass; packaged compiler produces byte-identical PPTX.
- [ ] Publish matching native preview/review metadata, then coherent release
  defaults/skill activation.

Evidence: repair-wave/feedback-v3-progress.json. Native review remains reopened.

Fresh node audit uses production frame contexts:6,307nodes, zero rejected.
The earlier single three-bullets rejection was the manual harness fake split
zone, corrected without weakening production geometry checks.

##19. Full587gallery completion authorized

Operator explicitly requests persistent implementation, tests, review and
remediation across all587definitions, including native visual review. This
supersedes the earlier next-intake-only status. Freeze d83bd58 as v5, preserve
v4 source/history and installedlocal.7 until qualification.65additions and
3revisions useexistingtypes but expose sourceoptions and footer/geometry gaps.
Initial68delta diagnostic26pass/42reject; defects are being implemented/repaired
in threeparallel capability/composition lanes. Root owns nativeexport/integration.

- [x] Immutable587source snapshot and delta/provenance observation.
- [x] Complete v5 sourcebundle/runtime/CLI support and all587source/boundbuilds.
- [ ] Fullnode/binding/affectednormal/race checks and independent code review.
- [x] NativePowerPoint export and individual visual review of all587gallerypages.
- [ ] Remediate everyfinding and natively recheck affected pages.
- [ ] Qualify maintained incoming local composition pilot.
- [ ] Matching587nativepreviews/catalog/unifiedSQLite and isolatedreleasepackage.
- [ ] Outside-checkout/relocation/portablechecks and coherentinstalledmigration.

### Native review checkpoint

All587 baseline production pages individually viewed; hashed range receipts in v5 repair-wave. Root reinspection supersedes false blank-bar/header findings on373/400. Generic v5 padding/legend/decimal/badge repairs and exact306/311/313 offering spacing fixes are under full test and independent review. Native donut566 positioning repair remains in progress. Full normal repository suite passes after genuine pinned-catalog/font-discovery and baseline-golden remediation; final race and latest migration changes still under test. Installedlocal.7 remains unchanged. Final native acceptance is pending fresh render/recheck of every changed page.
