# Remaining implementation and library plan

2026-09-25 · Revised for exhaustive catalogs, dynamic composition and delegated execution

This is the current implementation sequence. It supersedes the earlier sequence
in [PRODUCT_PLAN.md](PRODUCT_PLAN.md); that document and
[DESIGN_WORKFLOW.md](DESIGN_WORKFLOW.md) retain the broader architecture. Commands,
schemas, library counts and acceptance gates below are proposed unless explicitly
described as completed.

## Current slice — detailed proposal composition

The first mechanics showcase was insufficiently complex in user review. The
immediate acceptance gate is now the five-page dense proposal benchmark:
EnableComp 3/4 and UHG 28/38/43 patterns with substantial text, aligned evidence
relationships, shared specialists and client ownership. Keep the sparse showcase
as a regression/example baseline. Dense layouts must pass native measurement,
native final verification and per-slide visual inspection before catalog indexing.
The five-slide fixture has now passed those gates and is indexed with exact
spec/proof hashes; that acceptance applies to the authored examples only.

Implemented during this wave: explicit canvas primitives, measured numbered and
metric cards, pinned artwork, single-line accents, narrative notes, diagnostic
`fit-report`, full-contract measurement reuse, and bounded known-deck text recovery.
The repository skill documents their limits. The next implementation priorities
are reusable grid/panel contracts, source-shaped diagram/ribbon geometry, rich-text
paragraph semantics, and broader revision reconciliation. These remain pending;
explicitly authored geometry is not a general automatic visual designer.

### Next fidelity gates

The detailed [six-wave implementation sequence](planning/FIDELITY_IMPLEMENTATION_WAVES.md)
defines reference slides, deliverables, dependencies and acceptance fixtures. It is
the execution order for the gaps below.

1. **Measured grids and panels:** turn the five dense examples into reusable row/column and nested-panel contracts. Preserve aligned activity/evidence rows, explicit gutters and safe zones; report all overflowing cells before build. Re-run changed-copy and added-row variants against native PowerPoint.
2. **Visual evidence and source geometry:** build one deliverable-thumbnail montage and one phase ribbon/hatched roadmap extension using inventoried source graphics. Preserve editable text and pinned asset provenance, then compare against the relevant reference crop.
3. **Rich text and architecture:** add run-level emphasis with native phrase anchors, then a nested architecture diagram with captions, icons and routed relationships. Verify attachment positions after line-wrap changes.
4. **Revision scope:** extend the proven known-deck text-only recovery fixture to agreed formatting/layout changes, retaining a visible conflict report. General PPTX-to-semantic-YAML reconciliation remains a separate feature.

For each gate, delegate reference/asset review and bounded fixtures to Luna; keep compiler changes, integration and final visual acceptance with the lead or a coding specialist. Promote only the exact passing variants into retrieval.

## Completed foundation — dynamic roles, pods and teams

2026-09-26: [Ten reference variants](library/reference-variants.md) now have a
visual shortlist and exact IDs. Four source-bound text contracts are executable
with `pptxcomponent`; native rendering and frame measurement demonstrate normal
edits and a deliberate overflow failure. No styles or components are broadly
adaptation-approved yet.

### User review incorporated — 2026-09-26

The [review export](library/reference-preferences.json) contains two preferred,
six alternate, one avoided and one unreviewed reference. Preferred N2 becomes the
numbered-card priority; N4 is excluded from default reference recommendations.
M1 remains the first metric candidate based on the user's note while preserving
its selected “alternate” status. N5 retains its segmented layout with a typography
and density refinement backlog.

The central architectural change is **role tile → dynamic pod container → team
composition**. P1/P2's fixed cardinalities describe source edit fixtures. The
product must accept a variable list of roles, support several columns and compute
height from measured content. See the [decisions, proposed API and acceptance
fixtures](planning/dynamic-components.md).

Implemented in this slice: `pptxcompose probe|measure|build|verify`, native
measurement evidence tied to exact input hashes, variable-role pods, deterministic
column selection, semantic colors, contrast checks, and connector anchor geometry.
See [workflow and limits](library/dynamic-components/README.md) and the
[native proof report](library/dynamic-components/proof-report.json). This first
implementation uses explicit Arial styles; arbitrary source-theme inheritance is
still pending.

The enclosing [team composition](library/dynamic-components/team.md) now adds
standalone roles, a derived staffing legend, shared phase surfaces and routed
reporting relationships. Native character-bound measurement also avoids a
whole-range bounding-box anomaly without altering copy or typography. See the
[team fixture proof](library/dynamic-components/team-proof.json); lines remain
editable segments and do not follow shapes moved manually in PowerPoint.

Earlier gates and current status:

1. Prove source-sized P1 variants alongside the new dynamic examples; establish
   supported style profiles from native measurements and visual review.
2. N2-inspired numbered cards and M1/M3-inspired metric compositions are
   implemented with native measurement; N5 refinement remains pending.
3. The distributable skill is present. Index only native-verified, visually
   reviewed recipes, retaining their precise spec/proof hashes. Keep design
   preference and technical approval separate.

## 1. Current position

Completed foundations:

- Go 1.27.1 update and existing writer baseline.
- Structural inventories of 166 Graphics and Layouts slides, 85 UHG slides and
  35 EnableComp slides: 286 slide records. Graphics and Layouts also has 33
  native layout parts; these are not equivalent to 33 designed slide templates.
- Modernization inventory: 83 slides, 72 native layout parts, 2 masters and
  2,504 slide object nodes. Total registered corpus: **369 slides**. The tracked
  [source registry](planning/source-registry.json) records source identities; it
  is an initial manifest, not yet a CLI ingestion implementation. Across all
  slide parts there are **10,102 raw object nodes and 315 native group
  containers**; these counts precede deduplication and semantic grouping.
- Source-derived native reconstruction of the ten selected UHG slides, plus
  bio 67 and accent fixtures 5 and 6. The ten-slide delivery matches source
  controls in the same rendering context. Full-source export differences remain
  documented in the [experiment report](planning/RECONSTRUCTION_CHECKPOINT.md).
- Native phrase measurement, visible-art bounds, calibrated underline and
  highlight placement, native rendering and pixel/structural QA tools.

The [expanded catalog checkpoint](library/README.md) now includes occurrence, asset
and font ingestion, resolved geometry, component candidates and SQLite search.
Review of 89 source occurrences reduces 369 slides to 306 classification work units
(24 shared families, 2 distinct items, 280 unreviewed singletons). The component expansion supplies 244 source examples in 74 source patterns,
31 semantic families, six proposed color profiles and 572 named slot candidates;
capacity and adaptation remain unproven. The next work uses the representative
worklist, reviews the remaining dedup candidates, and develops measured contracts
for pods, identity cards and metric pairs.

Not yet completed: a curated broadly approved library, general semantic slide
authoring, narrative generation, arbitrary layout fit/reflow or colleague-edit
reconciliation. Bounded fit/collision checks and a repository skill now exist;
they do not prove independent visual design from a brief or arbitrary PPTX-to-YAML
recovery.

**Next product milestone:** an agent can find a suitable approved design, explain
the choice, supply new content through named slots, build and inspect the slide,
and recover a colleague's text edit without losing the source design.

## 2. What belongs in the library

Use a unified catalog with linked item types. A polished slide can exemplify a
layout that contains components and uses assets. Keep those relationships so
agents can search by either business purpose or visual structure.

| Item type | Examples | What makes it usable |
| --- | --- | --- |
| Assets | Photography, icons, logos, illustrations, fonts, hand-drawn accents | Preview, description, variants, source/hash, dimensions, crop/focal-point information, visible-art bounds where relevant, brand usage guidance |
| Shapes and components | Card, metric tile, quote block, portrait/name block, phase bar, pod, diagram node, legend, repeated table row | Native editable objects, named slots, font roles, padding, size rules, allowed transforms, attachment points and measured content limits |
| Slide templates/layouts | Phase-detail page, timeline, comparison, team composition, dense bio | Semantic zones, component cardinality, exact frame/theme dependencies, alignment rules, capacity, variants and explicit overflow/fallback behavior |
| Polished content slides | A vetted case study, capability story, methodology or reusable proposal section | Finished source slide, preview, takeaway, factual provenance, applicable audience/context, permitted edits and content review state |
| Guidance and examples | WM voice, design rules, strong titles, emphasis examples | Searchable excerpts with source authority, applicable profile, examples and counterexamples |

Distinguish raw object inventory from reusable components. Every source object
should be addressable; meaningful compositions such as a pod or portrait block
become curated items. A native group is a candidate boundary, not proof that it
is a coherent reusable component.

Distinguish native PowerPoint layouts from designed slide templates. A useful
stock design may contain dozens of ordinary shapes over one shared native
layout. Preserve the master/layout dependency and curate the full arrangement.

### Scope and source priority

1. **Graphics and Layouts:** visually inventory all 166 slides and inspect their
   33 native layouts. Classify every slide, including instructions, covers,
   duplicate variants and reference-only examples.
2. **UHG and EnableComp:** render and classify all 120 proposal slides, including
   hidden slides. Seed curation with the 13 UHG fixtures already exercised and
   the user-selected ten as preferred visual examples. Evaluate complementary
   EnableComp designs, particularly team slides 15–16.
3. **Software modernization campaign pick deck:** include the newly supplied
   source in the slide, object, group and native-layout inventories. Preserve its
   original filename and bytes; this is the one tracked source under `samples/`.
   See [its inventory](planning/modernization-inventory.md) for observed counts.
   Visual ratings remain pending until rendered review.
4. **Branding folder:** inventory `/Users/rscott/Documents/branding`, including
   its other PowerPoint templates, graphics, logos, photos, fonts and guidance.
   Existing file counts are not unique asset counts. Deduplicate identical bytes
   and connect color/format/crop variants without assuming visual similarity
   means interchangeability.
5. **Existing WM skill inventories:** reconcile their assets and references with
   the current branding folder. Keep the older 179-slide deck distinct from the
   supplied 166-slide Graphics and Layouts deck.

Use an explicit source registry. Exclude experiment outputs, negative controls
and changed-title fixtures from library ingestion unless explicitly registered
as test examples. They must not be discovered as approved presentation content.

### Item record

Each curated item needs:

- Stable item ID, version, kind, family and brand/profile.
- Source file hash and slide/object locators; derivative lineage.
- Thumbnail plus full-size preview, renderer identity and dependency manifest.
- Purpose, suitable content relationships, density, use cases and unsuitable uses.
- Semantic slot names and mapping to native objects or generated components.
- Effective typography, fills, lines, margins, crops, groups, layering and anchors.
- Observed content density, tested limits and permitted adaptations.
- Supported operations: inspect, preserve, replace content, resize, reflow or
  regenerate. Explicitly mark images/embedded objects that remain opaque.
- Separate design preference, technical readiness and content reuse states.
- Evaluation fixtures, known limitations, owner/reviewer and review date.

Readiness progresses through discovered → rendered → classified → bound →
adaptation-tested → agent-ready. Design preference and content approval remain
separate fields. For example, UHG 43 can be a preferred team layout while the
particular staffing assumptions remain specific to that proposal.

## 3. Inventory and curation pipeline

### Pass A0 — Discover and deduplicate before expensive processing

Register every source slide and its objects, then deduplicate across **all four
sources** before assigning visual classification, component extraction or layout
binding work. Keep all 369 occurrence records; report the measured number of
unique layouts separately. A reduction to 250–300 is a hypothesis, not a quota.

Use a staged comparison:

1. **Exact duplicates:** canonicalize slide objects and resolved dependencies,
   ignoring incidental package IDs/paths while retaining text, effective styles,
   geometry, crops, layer order and media hashes. Matching slide XML or native
   layout names alone cannot establish equivalence across decks. Record notes,
   links and source-specific metadata separately even when visible designs match.
2. **Shared layouts with different content:** derive structural fingerprints from
   object roles, normalized slide-space geometry, groups, text zones, typography,
   margins, table/diagram structure and image frames. Abstract literal copy and
   replaceable image identity only for layout comparison. Keep content-specific
   records and distinguish actual differences in style, crop or composition.
3. **Near-duplicate candidates:** bucket similar structures, compare within those
   buckets and review ambiguous pairs using representative/variant previews.
   Small decorative changes can be variants; materially different zones,
   cardinality, hierarchy or fit behavior remain distinct layouts. Similarity
   only proposes a merge. Each member must match the canonical contract; chains
   of pairwise similarity must not silently collapse different end members.

Use deterministic package/geometry processing first. The existing inventories
are a starting point; incomplete style inheritance or unresolved group transforms
must mark a comparison uncertain rather than produce an automatic merge. Render
ambiguous candidates before merging and inspect a sample of accepted matches.

**Deliverable / gate:** a deduplication manifest mapping every source slide to a
canonical layout, explicit variant or unresolved singleton, with representative
selection, comparison method/confidence and merge/split rationale. Show counts
for input occurrences, exact duplicates, confirmed shared layouts, variants and
unresolved candidates; distinguish overlapping categories. Preserve reversible
membership decisions. Cache by source/dependency hash and algorithm version.

Only then dispatch expensive layout-level work, once per confirmed canonical
layout/variant. Prefer the cleanest editable representative, respecting the
user's preferred exemplars. Reuse component extraction, descriptions and bindings
across confirmed members. Unique content, factual reuse and rendered fit still
need occurrence-level checks when wording or assets differ. A shared layout does
not make two polished content slides interchangeable.

### Pass A1 — Render representatives and inspect variants

Generate native previews/contact sheets for canonical representatives, meaningful
variants and uncertain candidates first. Reuse already valid renders where
possible. Every source slide retains an explicit preview/review state; do not
label unrendered members visually verified through cluster membership. Render
additional occurrences on demand for content/fit review and broader fidelity
coverage. Brand asset inventory and description reuse proceed independently.

**Deliverable:** a browsable deduplicated design catalog with expandable source
occurrences, variants and explicit missing/failed previews. Native export may
process a whole deck cheaply in one batch; expensive agent review and enrichment
still operate on the deduplicated worklist.

### Pass B — Classify and identify reusable parts

For each canonical layout and meaningful variant, identify its family, density
and major components once. Record slide role/takeaway against each distinct
content example and retain occurrence-level exceptions.
Produce component previews in context and isolated where possible. Link repeated
patterns across slides; keep meaningful variants. Resolve inherited typography
and group coordinates before treating extracted dimensions as reusable geometry.

**Deliverable:** every stock slide has a classification/disposition; each
proposed component or layout links to the source that demonstrates it. Automated
labels carry confidence/review status. Untagged photos remain discoverable by
metadata but are not represented as fully semantically indexed.

### Pass C — Curate the first useful release

Inventory coverage is exhaustive across registered sources: all slide objects,
native groups, native layouts and designed arrangements receive locators. We
expect hundreds of useful shape/group candidates, but will report actual counts
and deduplication rules rather than manufacture a target count. A text rectangle
is an inventoried object; a coherent reusable pod is a component; both are useful
at different levels.

Proposed first **validated release** targets, not limits on catalog coverage,
refined after visual classification:

- About **20 component families** with explicit supported variants.
- About **12 layout families**, emphasizing document-style proposals.
- About **10 polished slide candidates**, with reuse state recorded. The number
  actually approved for general content reuse depends on their facts and context.
- Reviewed assets needed by those items, plus the core logo/icon/accent variants.

Prepare preview galleries and populated examples for taste review. Capture
preferences in manifests. Complete extraction, descriptions and working examples
before requesting review; review should be of concrete items.

Starting families:

| Family | Seed evidence | Important variation |
| --- | --- | --- |
| Client proof/testimonials | UHG 8 | Quote length, proof metrics and source attribution |
| Current/future comparison | UHG 11 | Balanced and uneven content |
| Capability progression | UHG 12 | Step count, headings and dense supporting text |
| Architecture explanation | UHG 14 | Existing illustration versus editable diagram capability |
| Phase/scope detail | UHG 24, 28 | Shared family with content-density variants |
| Governance/process flow | UHG 36 | Step labels, paths and relationships |
| Roadmap/Gantt | UHG 38 | Dates, durations, milestones and extensions |
| Pods/team composition | UHG 43; EnableComp 15–16 | Pod counts, shared specialists, long role names and fractional roles |
| Named team roster | UHG 44 | Person/headshot pairing, count changes and pagination |
| Dense bio | UHG 67 | Profile length, headings, portrait crop and selected experience |
| Executive framing/emphasis | UHG 5–6 | Phrase choice, typography, underline/highlight movement |
| General stock compositions | Graphics and Layouts | Columns, cards, tables, cycles, matrices and section frames |

These are curation candidates, not declarations that all families are already
implemented or approved.

### Pass D — Prove adaptation

For each promoted item, test typical, short, long and boundary content. Include
wide/narrow characters, bullets, mixed emphasis, long names and changed item
counts. Measure rendered fit; character budgets are advisory.

Test examples include 1/3/6 pods, short/long bios, roster count changes, alternate
timeline scales, changed table rows, and emphasis phrases at different positions
and sizes. Restore the original content and compare the baseline again.

**Promotion gate:** named slots work; content meaning is preserved; no unexplained
clipping or overlap; native editability and dependencies survive; intended
decorative overlap is declared; manual fallbacks are precise. Keep technical
QA distinct from human design/content review.

## 4. Architecture to implement

### Library storage and retrieval

Keep original artifacts and versioned manifests as durable authority. Build a
local SQLite full-text index across assets, components, layouts, polished slides
and guidance. Store large binaries/previews in files, referenced by hashes.
Rebuilding the database must preserve curation because that state lives in the
manifests. Separate physical assets from concepts/variants and source occurrences:
one image can occur on many slides, and identical file bytes can have several
existing descriptions. Preserve each description and its authority/provenance;
add enrichment separately rather than replacing existing copy. See the
[asset audit](planning/asset-catalog-audit.md) for observed sources and gaps. Pin item and brand-pack versions in each deck's lockfile.

Proposed CLI surface: `library scan`, `library index`, `library search`,
`library inspect`, `library preview`, and `library evaluate`. Search returns a
bounded candidate list with preview, purpose, readiness, fit hints and compatible
variants. The agent inspects candidates and selects against actual content.
Choose a SQLite driver after checking required deployment and full-text support;
add semantic retrieval only when retrieval evaluations demonstrate a need.

### Semantic authoring and two build routes

Introduce a versioned semantic source model, with slide role, takeaway, evidence,
layout ID, named content slots, asset choices and emphasis intent. Native XML
paths remain implementation details behind versioned bindings.

Use the preservation compiler for imported/vetted slide designs and the existing
Go writer for newly composed objects/layouts. Both need common source maps,
capability reporting and QA. Prove cross-deck copying, mixed-master/font behavior,
internal links and resource remapping before broad library reuse.

### Dynamic composition and the grid

Use exact integer EMU geometry as the canonical representation. Offer a logical
layout grid (columns, gutters, spacing increments, safe regions and optional
snap) for newly composed slides; do not quantize imported geometry. The grid is a
placement aid, not a guarantee of attractive or collision-free composition.

A component exposes local bounds, visible-art bounds, text insets, connection
anchors, baseline/alignment anchors, minimum/maximum size and permitted
stretching. Resolve group child offsets/scales, rotations/flips and parent
transforms into slide coordinates while preserving editable local coordinates.
Measure transformed corners and visible ink; axis-aligned bounding boxes alone
can overstate collisions. Layer order and permitted overlaps are explicit.

Compose a custom slide by selecting a frame and zones, selecting components,
assigning content, solving alignment/equal spacing/containment constraints, then
measuring text and adjusting layout. Record which constraints changed and why.
Never silently shrink text below its role minimum or distort an icon/logo to fit.
If content cannot fit, recommend another variant, splitting content or a precise
manual fallback. Character limits are observed/tested hints, not hard promises.

Prove this on one new arrangement made from catalog components, alongside the
first semantic template. It should include changed card counts and different
text lengths; reusing an unchanged source arrangement alone does not pass this
gate. See [shape/layout execution detail](planning/shape-layout-execution.md).

### Fit and accent placement

Generalize the native phrase adapter to effective text styles, line bounds,
container insets, tables and transformed groups. Include font availability and
renderer information in diagnostics. Changes invalidate affected measurements;
an accent is recalculated after its target text is laid out.

Store underline/highlight style parameters as versioned contracts, with explicit
calibration provenance. Develop arrow contracts for endpoint meaning, curve
variant, tangents, whitespace and collisions. Preserve the off-slide asset plus
placement-note fallback. General aesthetic arrow routing remains unproven.

### Narrative and voice

Build the source-to-story workflow: brief → argument/evidence map → slide roles
and takeaways → slide copy → composition → assets/accents → editorial/rendered
review. Curate WM title, evidence and density examples alongside layout entries.
Use a fresh source packet for the new-content evaluation; an exact reconstruction
is a separate evaluation. Synthetic fixtures can exercise the pipeline before
the next real proposal supplies that packet.

### Colleague edits

Introduce stable semantic identities and a handoff baseline early. Prove a text
edit round trip before scaling the library. Then add three-way comparison of
handed-off source, current source and returned PPTX: text, rich formatting,
geometry, crops, tables, slide order, additions and deletions. Preserve opaque
objects; present ambiguous matches/conflicts rather than silently rewriting them.

### Skill pack

Write a thin working skill as the commands become usable. Include library
selection, document-proposal narrative, named-slot binding, fit repair, accent
placement, QA and import/reconciliation. Load references on demand. Release the
CLI and skill with compatibility versions; install the private WM brand/library
pack separately. Evaluate from a clean checkout without hidden session knowledge.

## 5. Delivery sequence and acceptance gates

| Milestone | Work | Exit evidence |
| --- | --- | --- |
| M0: Harden the successful experiment | Reproducible renderer/measurement adapters, stable IDs/source maps, immutable inputs, machine-readable diagnostics and environment checks | Existing reconstruction/accent fixtures remain reproducible; source preserved; supported/opaque operations explicit |
| M1: Make the corpus discoverable | Source registry, cross-source deduplication, representative/variant rendering, SQLite/manifests, taxonomy and initial component candidates | All 369 source occurrences mapped to confirmed canonical layouts/variants or explicit unresolved singletons; dedup counts reported; 166/166 stock slides have a disposition; all native layouts tracked; missing/failed items explicit; useful search across current sources |
| M2: Deliver the first agent-ready library | Semantic schema, initial component/layout/slide candidates, slot contracts, fit fixtures, dependency-safe reuse and thin skill | New content builds through named slots and one custom component composition; each promoted item has a rendered adaptation example; simple colleague text edit survives import and rebuild |
| M3: Build a new proposal end to end | Narrative/evidence/voice workflow, candidate selection, fit repair, assets and dedicated accent pass | A 10–15-slide document proposal from a fresh source packet; per-slide evidence and fit checks; review records human edits and remaining issues |
| M4: Make revision dependable | Wider change detection, three-way reconciliation, identity-loss and conflict fixtures | Supported colleague edits survive; ambiguity reported; repeated import is idempotent; original handoff and returned PPTX retained |
| M5: Expand coverage and distribute | Broader stock contracts, remaining proposal fixtures, special objects, packages and clean-environment evaluation | All 166 stock items have an honest reuse disposition; all registered proposal/campaign slides accounted for in inventory and declared evaluation coverage; declared supported operations tested; another agent succeeds with installed CLI/skill/brand pack |

M0 and M1 can proceed together. Start the narrow re-entry test during M2 so that
identity or preservation problems are found before extensive curation. Expand
the corpus continuously; a complete inventory does not require waiting for every
item to become a flexible template.

Record effort and intervention counts during M1/M2 before assigning calendar
estimates to full curation. Inherited formatting, opaque artwork and embedded
objects will affect the cost per family substantially.

## 6. First implementation work package

1. Create the source registry and item-manifest schema, including separate
   readiness/design/content states and explicit exclusion of experiment outputs.
2. Deduplicate layouts across all 369 slides using exact and structural
   fingerprints plus targeted review; publish canonical/variant membership and
   measured counts before expensive classification.
3. Render and classify canonical representatives and meaningful variants; record
   each source occurrence and its shapes/groups. Generate a gallery with duplicate
   occurrences collapsed. Track native layouts/dependencies separately.
4. Join the proposal, modernization and branding asset inventories into one rebuildable
   SQLite index; inspect remaining occurrences as needed for distinct content,
   fit and unresolved variant decisions.
5. Present a shortlist for the first library release with actual previews and
   source links. Prioritize phase detail, timeline, pod/team, roster, bio and
   executive framing because these have both proven fixtures and clear demand.
6. Convert one phase-detail design and one team/pod design into semantic layouts;
   fill them with different content, exercise capacity, and round-trip a native
   PowerPoint text edit. Use these to validate the schema before scaling.

7. Compose a new slide from approved shape components using grid/constraint rules;
   keep its generation/fit evidence separate from template adaptation.

The immediate product deliverable is a usable visual catalog plus the first
complete reuse-and-revision workflow and a bounded custom-composition example.

## 7. Delegation, dependencies and execution waves

Use three workers plus the coordinating architect, matching the four available
concurrent slots. Target roughly **80% of bounded worker assignments to GPT-6
Luna**, with Sol/Terra for substantial coding and Astra only for a specifically
identified difficult architecture or fidelity problem. This is an assignment
target, not a guaranteed token/cost ratio. Escalate after a concrete failed case,
not merely because a task has a large corpus. The architect owns schemas,
integration, readiness claims, final visual QA and commits.

| Wave | Parallel worker assignments | Dependency and integration gate |
| --- | --- | --- |
| 0 — Checkpoint and audit (current) | Luna: modernization inventory; Luna: existing asset/description audit; Luna: shape/layout and grid plan | Architect commits code, one authorized source deck and durable findings; all other samples ignored |
| 1 — Inventory and deduplicate | Sol/Terra: schema, dependency resolution and fingerprints; Luna: candidate-cluster/variant review; Luna: asset source/description joins | Publish cross-source canonical membership before layout classification; uncertain matches remain separate; native previews use one queue |
| 2 — Curate and make searchable | Luna: canonical component candidates; Luna: unique layout/variant ratings and slots; Sol/Terra: SQLite/FTS import and search | Workers emit separate manifests; a single integrator validates/indexes them, avoiding shared DB write conflicts |
| 3 — Prove reuse and composition | Sol/Terra: slot/fit/grid implementation; Luna: adaptation content and edge cases; Luna: accent-family fixtures and graphics matching | Components/layouts/assets must have stable IDs; architect reviews native renders and explicit failure cases |
| 4 — Prove revision and narrative | Sol/Terra: narrow import/reconcile support; Luna: narrative/evidence and voice cases; Luna: skill instructions and clean-session evaluation | Preserve handoff baseline and source map before colleague edits; independent agent must use documented commands |
| 5 — Scale approved coverage | Mostly Luna batches for inventory enrichment, curation, variants and QA; targeted coding workers fix demonstrated gaps | Promote each item independently; unresolved items remain searchable with honest limitations |

The initial waves temporarily use a higher coding proportion. The much larger
classification/enrichment/fixture workload supplies the intended overall Luna
share. Reuse workers for bounded follow-ups instead of spawning an agent per
slide. A typical curation batch is 10–15 canonical layouts/variants or 25–40 component candidates;
reduce it for dense diagrams. Each handoff supplies exact input paths, schema,
owned output directory, expected counts, evidence requirements and uncertainties.

Dependency chain:

`checkpoint → source registry/schema → raw inventory → cross-source deduplication + targeted previews → canonical layout/variant classification`

`asset audit → preserved metadata + verified file mapping → asset index`

`classification + asset index → searchable catalogs → slots/constraints → adaptation/custom composition`

`stable identities + handoff snapshot → text-edit re-entry → broader reconciliation`

Asset auditing/enrichment can proceed alongside shape/layout work throughout.
Text/style/group resolution and package import receive coding review before
hundreds of contracts are generated. Native rendering and final artifact
integration are serialized; descriptions and manifests can be prepared in parallel.

## 8. Concrete completion checks and review deliverables

### Catalog coverage and design choices

- Source registry: exact SHA-256, local/configurable location, source role and
  allowed ingestion scope; modernization included; experiment outputs excluded.
- Raw inventory: slide, shape, nested group, layout and master counts reconcile
  against each source. Mark missing previews, opaque objects and unresolved styles.
- Deduplication: every source occurrence has a reversible cluster assignment or
  unresolved status; report canonical counts and avoided repeated work. Preserve
  source-specific content/fit checks and never equate shared masters with shared
  layouts. Cache layout-level results against canonical versions.
- Component candidates: preserve native group membership and support inferred
  compositions of ungrouped objects. Record grouping rationale; do not merge
  shapes solely because their bounding boxes touch.
- Layout catalog: one entry per designed arrangement with variants and native
  layout dependencies, plus semantic zones and layout-selection cues.
- Polished-slide catalog: keep design preference separate from factual reuse.
  Proposed design dispositions: preferred, usable, reference-only, avoid,
  unreviewed. Record reasons such as hierarchy, density, clarity and editability.
- Review gallery: filtered views for components, slide templates, polished slides
  and assets; show populated previews, source links, limitations and proposed
  ratings. User feedback updates versioned manifests.

### Assets and accents

- Inventory illustrations, icons, logos, photography and decorative marks with
  explicit category coverage. Report physical files, deduplicated assets, variants
  and description coverage separately. Never count inferred tags as vetted copy.
- Preserve existing descriptions verbatim with provenance. Join by exact path
  and content hash when available; uncertain name/visual matches stay candidates.
- Enrich missing semantics only after examining the image. Add focal points,
  safe crop, allowed color changes, alpha bounds and brand usage rules where useful.
- Review asset/content relevance and deck-wide consistency as a dedicated pass.
- Broaden arrow fixtures across actual catalog families: straight/curved/elbow,
  dashed/solid, double-headed or loop where present. Define tail/tip and tangents
  in asset coordinates; do not guess endpoints from the image rectangle alone.
- Exercise direction changes, scale, short/long spans, endpoint movement,
  obstructed paths, crowded text and intentional overlap. Record flip/stretch
  permissions per asset. Check curve choice and visual spacing after geometry.
- An accent passes only with the original phrase/content intact unless the
  experiment explicitly declares a change. Unsupported placement returns
  `manual_required` with the asset staged off-slide and a precise placement note.

### First product proof and QA log

Produce a phase-detail adaptation, a pod/team adaptation, one newly arranged
component-based slide and a colleague text-edit round trip. Each carries its
semantic source, source map, locked dependencies, preview and QA record. These
four proofs precede mass promotion of inventory entries into agent-ready designs.

Continue the [QA defect log](planning/QA_ERRORS.md), adding source/item IDs,
expected versus observed behavior, renderer/resolution, severity, diagnostic
images, fix, remaining limitation and recurrence fixture. Compare literal pixels
for unchanged reconstructions; evaluate intentional adaptations using fit,
hierarchy, typography, asset relevance and human design review. Pixel equality
alone cannot judge whether newly written content makes a persuasive proposal.

Track completion by coverage, promoted items, successful adaptations, unresolved
QA defects and human intervention count. Set calendar estimates after the first
classification and adaptation batches establish actual throughput.
