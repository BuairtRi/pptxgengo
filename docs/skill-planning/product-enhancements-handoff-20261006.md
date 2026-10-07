# Product enhancements handoff

## Purpose and status

Date: 2026-10-06. Prepared against repository commit `86a0a7a4`.

Ri identified five priorities for the next product phase:

1. Local vector search for templates, combined with BM25 / keyword search.
2. A library of content-complete slides that can be reused across decks.
3. Hardened installation and upgrade workflows.
4. More intuitive direct editing of generated PowerPoint shapes and geometry.
5. Reconciliation of collaborative PowerPoint edits into maintained YAML.

This document captures the requested direction, recommended initial scope,
dependencies, estimates and completion criteria. It does not mark these new
enhancements implemented. Existing foundations are described separately below.
The estimates and implementation choices remain planning recommendations.

## Current baseline

- Production Mac release: `0.1.0-local.21`, V11 design library. The pinned bundle
  has 649 templates, including 29 workshop templates. Future work must read counts
  and source revisions from bundle metadata rather than assume these counts persist.
- Windows preview: `0.1.0-local.21-windows-preview.1`. Includes x64 binaries,
  user-level installation, fonts, skill, SQLite, gallery, documentation and a
  smoke-test script. Windows x64 and ARM64 cross-compilation passed; actual
  Windows execution and native PowerPoint qualification remain pending.
- Both GitHub and GitLab have Windows native CI configuration for an interactive
  desktop runner. Ri intends to add a Windows node within approximately 48 hours;
  its availability and first results must be confirmed before claiming coverage.
- SQLite already contains an FTS5 table, `entity_fts`. The current unified finder
  primarily uses metadata and structural scoring; combined BM25/vector retrieval
  is not implemented.
- Searchable metadata, template contracts, authoring aliases, specimen previews,
  content matching, split slide YAML, density, sections/dividers, draft review
  notes and project migration are existing capabilities to build upon.
- Generated PowerPoints already contain native groups. Their existence alone
  does not establish an intuitive editing experience.
- Builds retain baseline artifacts and an object map with native identities,
  baseline text, source pointers and mapping classifications. This is a foundation,
  not a complete reverse mapping or reconciliation engine.
- Standalone PPTX inventory is available. Automated PPTX-to-YAML conversion and
  general round-trip reconciliation are not available.

The Windows packages also include four offline HTML guides under `guides/`.
Open `00-start.html`; installation instructions explain both
`pptxgengo catalog --design-system --open` and the documentation server:
`pptxgengo docs`, then browse `http://localhost:8787/` while PowerShell remains open.
The gallery and docs board explain source designs; supplied content still needs
its own build and visual review.

## Recommended sequence and estimates

| Order | Enhancement | First useful release | Focused engineering estimate |
| --- | --- | --- | --- |
| 1 | Installer hardening | Reliable install/upgrade/rollback and repair diagnostics, with real Windows coverage | 3–5 days initially |
| 2a | Hybrid template search | BM25 + local vectors + structural filtering and explainable results | 1–2 weeks |
| 2b | Reusable finished slides | Versioned content-complete slides, search, preview and independent insertion | 1–2 weeks, plus content curation |
| 3 | Native editing improvements | A representative family pilot with preserved semantic identities | 1–2 weeks for the pilot; wider rollout separately |
| 4 | Bounded reconciliation | Three-way comparison and supported text-change proposals | 2–3 weeks for the first version |

These are engineering estimates, not elapsed delivery commitments. They exclude
full-library curation, a complete geometry sweep, signing/procurement lead time,
general reverse compilation and unknown defects found by Windows qualification.
Parallel work can overlap; do not add these ranges into a firm calendar promise.

Search and slide-library development can proceed in parallel after agreeing on
entity identity, revision and discovery contracts. Native editing and reconciliation
must share an identity/mapping contract from the beginning. Implement editing
changes before relying on that contract to reconcile user edits.

## A. Hybrid local template search

### Product outcome

An operator or agent can describe what a slide needs to accomplish and receive
stock layouts that preserve its content relationships, even when the query does
not use the template's exact terminology.

### Initial scope

- Extend the existing discovery interface rather than create another catalog.
- Offer keyword, semantic and hybrid modes. Proposed flag names and command
  extensions must be specified before implementation; they are not current CLI promises.
- Use the existing SQLite FTS5 foundation for BM25 retrieval.
- Precompute template embeddings and perform query embedding locally using a
  pinned model. Record model identity, dimensions, text preparation and source hashes.
- Index names, purposes, relationships, content groups, aliases and useful
  terminology. Include deliberate retrieval descriptions rather than depend on
  synthetic example wording alone.
- Preserve filters for entity kind, lifecycle, structure, item count and supported
  capabilities. Clearly distinguish filters, ranking hints and actual fit checks.
- Combine ranked results using an explicit fusion policy; do not directly add
  incompatible raw BM25 and vector scores.
- Return reasons, identifiers, revisions, preview paths and structural caveats.
- Begin with a straightforward vector scan at this catalog size. Benchmark before
  adding approximate-nearest-neighbor infrastructure or a separate vector database.
- Keep lexical search usable when the optional model is absent or incompatible;
  report the fallback instead of presenting it as semantic retrieval.

### Completion criteria

- Known queries such as interview lists, practices heat maps, modernization
  economics, roadmaps and pillars surface reviewed appropriate candidates.
- Maintain a judged query set with expected acceptable candidates. Compare
  keyword-only, vector-only and hybrid ranking, including exact IDs and synonyms.
- A plausible semantic match cannot hide an item-count or structural incompatibility.
- Search returns actionable previews and explanations; retrieval does not claim
  successful content fit or native acceptance.
- Record cold/warm latency, model footprint and memory on supported Mac and Windows
  architectures. Agree performance targets after the initial benchmark.
- Stale source/model embeddings are detected and regenerated or rejected explicitly.
- Normal use requires no hosted embedding service or Python installation.

### Decisions still required

Choose a model/runtime after evaluating relevance, CPU latency, licensing,
distribution size and cross-platform integration. Decide whether model files ship
in the standard package or in a separate optional offline search package.

## B. Content-complete reusable slide library

### Product outcome

An agent can find and insert an already-developed page, then adapt its copy for
the current presentation. A reusable slide is a distinct entity from an empty
template or a synthetic specimen.

### Initial scope

- Start with 10–20 approved useful slides, selected with the operator.
- Store maintained YAML, template/source pins, assets, evidence, previews, owner,
  revision and approved reuse scope. Include dates where claims can become stale.
- Support the same discovery infrastructure as templates, with visibly distinct
  entity kinds and result descriptions.
- Insert an independent editable copy with fresh slide/item identities. Register
  required assets and record library lineage and composition decisions.
- Detect ID collisions, incompatible pins, missing assets and unsupported local
  dependencies before writing changes.
- Preserve exact content or disclose intentional adaptations. Apply the deck's
  evidence and copy-review process to adapted claims.
- Library updates must not silently change slides already inserted into a deck.
- Retain existing client/maintainer export behavior, including removal of internal
  draft review groups from the client PowerPoint.

### Completion criteria

- Search, preview and insert a curated slide into two different projects.
- Both projects build with all required assets and correct fresh identities.
- Editing one inserted copy affects neither the library nor the other deck.
- Updating a library revision leaves earlier deck copies unchanged and their
  provenance inspectable.
- New pages appear in the composition log and survive split, reorder, export and
  handoff workflows.

### Decisions still required

Agree the initial content set, stewardship, reuse scope and freshness policy.
Specify insertion commands and YAML representation; avoid adding another large
embedded template section to human-editable slide files.

## C. Installer hardening

### Product outcome

A colleague can install, upgrade, diagnose and recover the toolkit without
guessing which executable, library, fonts or skill their session is using.

### Initial scope

- Obtain real Windows core/native smoke results from colleagues and the new runner.
- Preserve package hash verification; validate architecture and required resources.
- Stage installs and verify them before activating the release. Interrupted or
  failed upgrades must leave the previous working release usable.
- Specify upgrade/rollback semantics and keep CLI, library and skill version
  selection visible. Avoid accumulating ambiguous PATH entries for old releases.
- Detect common PATH, font, skill, resource and PowerPoint setup problems, with
  bounded repair operations and clear diagnostics.
- Preserve existing user skill backups and font opt-outs. Uninstall must distinguish
  installed toolkit resources from user projects and authored decks.
- Cover paths with spaces, user-level permissions, repeated installation,
  interrupted copying and missing fonts in the relevant test lanes.
- Treat signing and enterprise distribution as a separate implementation scope
  with certificate/policy dependencies.

### Completion criteria

- Clean install, upgrade and rollback are exercised on supported Mac and Windows
  environments, with real Windows execution distinguished from cross-compilation.
- Failed activation preserves the prior version and provides an actionable report.
- A newly opened terminal resolves the intended executable; diagnostics identify
  stale sessions and conflicting earlier installations.
- Skill and font outcomes are reported separately from CLI installation success.
- Core CI remains headless; Office qualification runs from an interactive desktop
  and includes inspection of actual exported slides.

## D. More intuitive direct PowerPoint editing

### Product outcome

Cards, lists and diagrams behave like the visual units a person sees, with easy
text editing, predictable selection and useful move/resize behavior.

### Initial scope

- Inventory editability by component family before changing all 649 templates.
- Pilot cards/pillars, lists, tables and a representative diagram.
- Combine containing shape and text where the design permits it; use native
  paragraphs/runs for separate title/body roles where suitable.
- Group meaningful units and avoid excessive nesting. Existing groups must be
  evaluated through actual editing tasks rather than counted as completion.
- Provide useful object names in the Selection Pane and predictable ownership.
- Use native tables, charts and connectors where practical. Evaluate masters for
  recurring background/frame decoration without changing visible design.
- Preserve style roles, density behavior, z-order, alignment and logical identities.
- Extend object mappings to include paragraph/run or cell addresses when combining
  several source fields into a single native object.

### Completion criteria

- On representative Mac and Windows slides, a person can edit card copy, select
  and move a whole card, align units, edit a table and adjust a diagram predictably.
- Before/after builds receive native visual review at relevant density tiers.
- Editing improvements retain source-field mapping and baseline lineage needed
  by reconciliation.
- Define family-specific rollout coverage and exceptions after the pilot; do not
  claim an entire-gallery pass from a handful of component examples.

### Dependency and scope boundary

Agree stable identities with reconciliation before changing serialization.
Grouping without correct text/cell mapping is insufficient. Arbitrary manually
resized geometry does not automatically become valid shared-template YAML.

## E. Collaborative PowerPoint-to-YAML reconciliation

### Product outcome

Authors can recover supported edits from a colleague's PowerPoint without losing
parallel YAML changes or falsely claiming that an edited deck matches its source.

### Initial scope: supported text proposals

- Compare the recorded source/native baseline, current YAML and edited PPTX.
- Require recognizable deck/build lineage. Map through stable identities, not
  slide order, current text or bounding boxes alone.
- Propose bounded updates to supported named source fields with enough context
  for a person or agent to review them.
- If YAML and PPTX changed the same field differently, report an explicit conflict
  showing baseline and both current values. Matching changes can be recognized.
- Preserve paragraph/bullet meaning where supported. Report unsupported rich-text
  formatting rather than silently flatten or infer design settings.
- Report missing, duplicated, unmatched or ambiguous objects. A missing shape is
  not proof that its source content should be deleted.
- Preserve the hand-edited deck and immutable baseline. Apply reviewed proposals
  to source with receipts/backups, then create a new build and review it.
- Detect geometry, shape additions/deletions and unsupported formatting as manual
  review items in this release. Do not silently ignore them or claim full synchronization.
- Keep shared design definitions immutable. Later intentional geometry adoption
  can use a local derivative with explicit rationale and visual review.

### Completion criteria

- A colleague changes three supported text fields in PowerPoint; reconciliation
  proposes the correct YAML changes, and applying/rebuilding reproduces their text.
- Parallel conflicting YAML/PPTX edits remain explicit and neither side is lost.
- Supported no-op and repeated reconciliation do not create duplicate mutations.
- Save As, slide reordering and normal PowerPoint text editing are evaluated on
  both platforms for identity survival.
- Duplicate/delete/ungroup operations either preserve proven mappings or produce
  bounded ambiguity reports; they do not guess a source update.
- Unsupported geometry and formatting remain listed even when text proposals succeed.
- Baselines, edited files and source history remain available after adoption.

### Later phases, separately scoped

Stable table-cell and chart-data reconciliation; speaker notes and structural
edits; importing newly added slides; and controlled geometry/local-template
adoption. General lossless PPTX reverse compilation is outside the first release.

## Execution record

Additional scope requested on 2026-10-07: two CI-generated packaged PowerPoint
browsing libraries (all templates/valid frame and rail variants, and versioned
reusable authored slides), consistent portable project folders, numbered
source/deck versions and complete colleague ZIP handoff. See
[portable projects and browsing decks](portable-projects-and-browsing-decks-20261007.md)
for the requirements and acceptance evidence.

Implementation work begun after this handoff is tracked in the
[execution record](product-enhancements-execution-20261006.md). The original
planning scope and completion criteria below remain the baseline.

## Execution guidance for the next agent

1. Read this handoff and confirm the installed release/source, available Windows
   runner and colleague smoke results. Do not assume the preview is qualified.
2. Convert installer/search work into small implementation contracts with explicit
   diagnostics, compatibility and acceptance criteria.
3. Agree common catalog identity/revision contracts for templates and finished
   slides, and native source-field mapping contracts for editing/reconciliation.
4. If parallel agents are authorized, separate installer hardening, discovery,
   slide curation/insertion and the native editing pilot. Keep shared contracts
   under one owner and integrate sequentially.
5. Use bounded tasks for inventory, fixtures, documentation and curated query
   cases; use an integration owner for ranking, runtime packaging and three-way merge.
6. Follow the existing short development lane. Keep exhaustive sweeps, full race
   checks and desktop PowerPoint qualification in their separate lanes. Obtain
   authorization before adding/running tests when required by the active instructions.
7. Update CLI and skill documentation plus the offline colleague guides as each
   capability becomes real. Clearly label proposed commands until implemented.
8. Report core implementation, template/content inventory and native qualification
   separately. Every completion claim should identify actual evidence and limits.

## Relevant source and documentation

- [Discovery behavior](../semantic-template-discovery.md)
- [Index construction and FTS5](../../internal/wmdesign/library_index.go)
- [Current unified finder](../../internal/wmdesign/library_index_query.go)
- [Source/identity and three-way reconciliation contract](deck-source-contract.md)
- [Native object mappings](../../internal/deckproject/objects.go)
- [Component grouping](../../internal/wmdesign/groups.go)
- [Scene grouping](../../internal/wmdesign/scene_groups.go)
- [Mac release installation](../../scripts/install-local-release.sh)
- [Windows installation](../../internal/releasepackage/install-windows.ps1)
- [Windows package producer](../../internal/releasepackage/package.go)
- [Test and Windows runner guidance](../testing.md)
- [Windows preview qualification record](../../release/qualification-windows-preview1.json)
- [Presentation skill](../../skills/west-monroe-presentations/SKILL.md)
- [Offline colleague guides](../../internal/releasepackage/guides/00-start.html)

## Repository handoff

At preparation time, both remotes used `master` as their default branch and had
no `main`. Ri requested this handoff committed to `main` and pushed to both
GitHub and GitLab. Publish `main` with the full history and synchronize `master`
by fast-forward so existing default-branch consumers receive the same handoff.
Changing hosting defaults or removing `master` is a separate branch-management
decision, not necessary for publishing this document.

Track source, documentation and qualification records in Git. Colleague release
ZIPs and compiled distribution artifacts remain outside the repository.
