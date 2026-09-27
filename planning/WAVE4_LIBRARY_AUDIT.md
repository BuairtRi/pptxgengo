# Wave 4 library audit and build recommendations

Read-only audit of the Wave 4/Wave 6 plan, existing library/catalog and authoring
commands. This document records current behavior and a concrete proposed handoff;
it does not promote inventory candidates or change native PowerPoint/code.

## Current state and important counts

The authoritative overview is [library/README.md](../library/README.md), with
input hashes and artifacts in [catalog-manifest.json](../library/catalog-manifest.json).
The manifest says the current SQLite database is
`samples/showcase/catalog-v3.sqlite` and pins its SHA-256. It records 13,831 source occurrence rows, 2,169
asset records, 559 component occurrence rows, 31 semantic component families,
26 reviewed layout-family decisions, 6 proposed style profiles, 5 dynamic
recipes, and 17,456 total indexed rows. There are 369 source slides, 306 layout
work units, 24 shared visual families and 2 distinct reviewed items; 280
singleton slides still await review. Geometry covers 9,669 frames and has 433
unresolved frames. The manifest explicitly reports **zero approved adaptable
components**.

The live `catalog-v3.sqlite` has these useful item kinds/counts (queried from the
database): asset 2,169; component 559; component_family 31; dedup_slide 369;
dynamic_recipe 5; font 58; group 401; layout_family 26; master 11;
native_group 369; native_layout 313; review_queue 33; shape 12,737; slide 369;
style_profile 6. The older [catalog-index-report.md](../library/catalog-index-report.md)
reports 17,451 rows and predates the five recipes; use the current database and
manifest for present counts. The catalog is a rebuildable 74.6 MB local index,
not the durable contract store.

### Exact SQLite schema and search

`scripts/catalog-index.py` creates a SQLite user_version 1 database:

```sql
CREATE TABLE items (
  id TEXT PRIMARY KEY, kind TEXT NOT NULL, source_id TEXT,
  category TEXT, title TEXT, body TEXT NOT NULL, json TEXT NOT NULL
);
CREATE VIRTUAL TABLE items_fts USING fts5(
  item_id UNINDEXED, title, body, tokenize='unicode61'
);
CREATE TABLE ingestion (
  input_name TEXT PRIMARY KEY, input_path TEXT NOT NULL,
  sha256 TEXT NOT NULL, record_count INTEGER NOT NULL, detail_json TEXT NOT NULL
);
```

FTS5 `search` takes a required MATCH query, optional exact `--kind`, and limit
1–100; it orders with `bm25`, returns snippets and a compact set of status fields.
Full source JSON is fetched by exact-ID `inspect`. Occurrence slide search bodies
include aggregated descendant text. FTS has no structured filters for readiness,
approval, preference, supported operations, fit envelope, provenance date, or
preview availability. `catalog-index.py` is Python-only. The reviewed
`catalog-reference-review.py find` overlay honors preferences, but it is also
Python-only and searches reviewed short descriptions, not the SQLite corpus.
The SQLite copy of layout families/component occurrences remains the earlier
unreviewed snapshot; do not use it to override the preferences overlay.

### Deduplication, preferences, and what “reviewed” means

`library/layout-decisions.json` is source-hash pinned and has 26 reviewed
classifications. It represents reviewed family membership/disposition, variants,
preview evidence, design preference, rationale and content approval; it does not
mean those compositions are editable or fit-approved. `library/layout-worklist.json`
is the remaining review queue. `layout-dedup-report.md` is an earlier structural
pass and is superseded where later decisions exist. Five reviewed duplicate
component examples were consolidated to aliases; the 244 retained selected
examples are not 244 unique or adaptation-approved patterns.

The ten-entry shortlist in `library/reference-shortlist.json` and its preference
overlay `library/reference-preferences.json` should guide selection. The overlay
hash-pins the shortlist and records two preferred (N2 and P1), six alternate,
one avoided (N4) and one unreviewed (P2). These are user design preferences,
not technical approval. N2 is preferred for explanation rows. P1 conveys a
dynamic subcomponent model: role tiles inside pods, arbitrary role count,
potential multi-column arrangement and variable pod height. That should inform
reusable component composition; source P1/P2 remain two-/three-role references,
not fixed supported cardinalities. Color meanings in UHG43 encode staffing
ownership/commitment and cannot be recast as decoration.

## Qualified evidence: retain gates and classifications

Keep the following distinction explicit in selection records: **inventory** is
searchable, **reviewed** means a human inspected that source/structure,
**measured fixture** means a bounded authored instance passed specified native
checks, and **adaptation-qualified** requires the changed-content and stress
variants described by the plan. No source-bound inventory item should imply the
last state without explicit evidence.

| Proof set | What is genuinely established | What it does not establish |
| --- | --- | --- |
| Five dense proposal recipes, `library/showcase/recipes.jsonl` and `dense-proof.json` | Five-phase approach, phase detail, workflow matrix, 21-role delivery organization, 12-month/seven-workstream roadmap; 318 editable objects across five synthetic proposal slides; native measurement/review and exact spec/proof hashes. Recipe readiness is `native_fixture_verified`. | Each recipe has `adaptation_approved:false`; these are measured fixture specs with explicit geometry, not general library templates or text/cardinality envelopes. |
| Wave1 measured layouts, `library/layout-components/` | Phase-column, phase-detail and workflow-matrix mechanics have copy/row-count/parent-move, fit, native-copy and visual evidence; overflow is explicit. | `proof.json` is `native_fixture_verified_not_general_layout_approval`. Only those bounded patterns/spec envelope are proven. |
| Dynamic pods/team, `library/dynamic-components/` | Variable role tiles/pods have native measurement, fit, parent composition and explicit synthetic cases. Team fixtures cover 3 slides/27 roles; team fixture covers 2 slides/18 roles, five relationships; reports record proof and limitations. | Proof status says `native_fixture_proof_not_general_catalog_approval`. Do not infer arbitrary nested multi-column pod support beyond the tested spec and dimensions. |
| Wave2 visual components | Eight bounded pages plus four follow-up pages have native verification and visual reviews. The follow-up has 101 objects, 63 text objects (all zones fit), SVG pictures, native bullets/outlines, and negative overflow case. | Proof status is `bounded_fixtures_verified_not_full_wave2_or_pixel_identity_approval`; source reconstruction is not cross-slide layout adaptation. |
| Source-bound `pptxcomponent` contracts | M1 metric card, N1 numbered card, P1 and P2 bind text/run IDs in their exact extracted source scenes and can replace text without corrupting topology. | Contracts preserve geometry/cardinality; no measured fit, effective typography, general movement, cross-slide instantiation or adaptation approval. The four expose no profiles. |
| 244 component candidates / 31 canonical semantic families / 26 layout decisions | Stable IDs, candidate slots, hashes and selected preview/source links make a useful retrieval base. | Component rows are readiness `discovered` (320) or `slot_candidates` (239), all preference `unrated` and content `unreviewed`; 31 families are all `semantic_candidate`; 26 layout families are `classification_only` with `visual_family_only_slots_and_adaptation_unproven`; 6 styles are `proposed_contract`. |

So Wave4 should promote contract records that point to proof, but cannot claim
the overall catalog contains adaptation-approved components yet. Promote only
bounded sub-contracts on a measured envelope, with the exact changed-content
and negative evidence tied to that version.

## Existing contract/API surface and gaps

There are several useful but nonuniform records:

- `library/component-contracts/{metric-card,numbered-card,pod-two,pod-three}.json`
  use `pptxgengo.component-contract.v1`, pin source deck SHA, slide and serialized
  scene SHA, object IDs, binding-ID text slots, RGB roles, optional profiles and
  string constraints. The schema does not encode measured fit, transforms,
  preview/proof artifacts or approval state. Current four contracts have no
  semantic profiles.
- `library/component-contracts/wave2-fixture-contracts.json` is a source-bound
  extraction/fixture detail record, explicitly unapproved. `visual-wave2` proof
  JSONs and the showcase recipes use other schemas.
- The five `dynamic_recipe` catalog items expose spec/proof path and hashes, but
  deliberately say `adaptation_approved:false`.
- `cmd/pptxcomponent` Go CLI supports `inspect` and `apply`. Inspect outputs the
  contract and explicit bindings with `fit_status:not_measured` and
  `typography_status:explicit values; inheritance unresolved`. Apply copies a
  scene project to a new directory, edits exact-bound text slots and RGB roles,
  validates cardinalities/sentinels/hashes, and emits an application report
  requiring native measurement/visual review with `adaptation_approved:false`.
  It does not select/search contracts, preview source examples, compose across
  slides, instantiate a pattern as a new `pptxcompose` spec, evaluate evidence,
  or manage a library approval transition.
- `cmd/pptxcompose` supplies Go `probe`, `measure`, `fit-report`, `build`,
  `verify`, cache operations and limited text recovery; `cmd/pptxscene` extracts
  source-native scenes. These are separate authoring operations. Neither offers
  a library search/inspect/preview/instantiate workflow. Both need explicit
  files and geometry today.

## Proposed portable contract model

Add a versioned JSON Schema (YAML may serialize the same model later) under
`library/contracts/`, with each component contract a portable source of truth;
SQLite should index contract IDs/summary/status and not own their contents.
Contract payload should have the following top-level fields:

```json
{
  "schema": "pptxgengo.library-component.v1",
  "id": "...",
  "version": "1.0.0",
  "kind": "layout|component|recipe|asset",
  "name": "...",
  "purpose": "...",
  "content_roles": [],
  "source": {"source_id": "...", "source_sha256": "...", "slide": 1,
              "source_part": "...", "source_object_ids": [], "scene_sha256": "..."},
  "composition": {"spec_path": "...", "spec_sha256": "...", "children": [], "slots": []},
  "assets": [],
  "cardinality": {},
  "fit_envelope": {"measured": [], "unsupported": [], "font_policy": "no_silent_shrink"},
  "transforms": {"translation": "tested", "resize": "unsupported", "rotation": "unsupported"},
  "style_variants": [],
  "preference": {"value": "preferred|alternate|avoid|unreviewed", "source": "...", "note": ""},
  "qualification": {"state": "inventory|reviewed|measured_fixture|adaptation_qualified",
                    "evidence": [], "failures": [], "reviewed_at": "..."},
  "previews": [],
  "provenance": {"created_from": [], "review_history": []}
}
```

Treat the example as a field checklist, not a claim all values are currently
known. Every unknown field should be absent/null or explicitly unresolved,
never guessed. A qualification transition must require source/style hashes,
measured native fit, changed-content fixture, stress/negative fixture and
visual-review hashes. Design preference stays independent of approval/readiness.
Style profiles should name semantic roles/tokens; record exact RGB/type after
resolution and do not present unproven profile colors as applied variants.

## Go library operations to add

Keep SQLite as a versioned, rebuildable index. Add a read-only Go package and
CLI (`cmd/pptxlibrary`) that can open the catalog URI in SQLite read-only mode,
verify index/schema/ingestion hashes and resolve contract paths safely inside the
repository. Suggested operations and behavior:

1. `find --query ... [--kind component|layout|recipe|asset] [--readiness ...]
   [--preference ...] [--source ...] [--limit ...]`: FTS relevance plus
   structured status/preference filters, default exclusion of `avoid` and
   unqualified records unless caller explicitly requests exploratory inventory.
   Rank preferred reviewed references first, but never call them technically
   qualified merely due to preference. Explain why candidates matched and show
   readiness/confidence/source.
2. `inspect --id ...`: load portable contract by exact ID and report source
   hashes, semantic purpose/roles, slots/cardinality, measured envelopes,
   allowed transforms, asset identities, variants, caveats, preference, proof
   state and links. Reject missing hash or stale contract/index correspondence.
3. `preview --id ... [--variant ...]`: serve/open hash-pinned preview path and
   return source full slide plus crop, caption and source link; if absent, provide
   a reproducible read-only render preparation command. Never render/modify
   source PowerPoint invisibly from a finder call.
4. `instantiate --id ... --values ... --out NEW`: resolve a qualified or
   explicitly allowed measured fixture contract into an authoring spec by
   semantic slot, component ID and variant; include exact source/contract hashes,
   narrative trace IDs, selected rationale and measurement/QA state. Reject
   unsupported transforms, out-of-envelope copy/cardinality and missing
   evidence. Do not allow raw component candidates to instantiate by default.

Commands should emit stable JSON results as well as concise text. `instantiate`
must produce a reviewable new spec only; native probe/measure/build/export/visual
review remains the existing `pptxcompose` pipeline and serialized PowerPoint
step. It must not use fixed coordinates hidden in the selector; those belong in
the versioned recipe/contract and authorship reasoning is separately recorded.

### SQLite index changes (rebuildable)

Extend the indexed `items` projection or add normalized read-only tables for
portable contract data: `contract_id/version`, kind, qualification state,
reviewed preference, source family, capacity range, supported transforms,
variant tags, preview URIs, evidence IDs, and contract SHA. Add FTS text for
purpose, content roles, constraints, known failures, evidence summaries and
existing verbatim descriptions. Preserve existing `items.json` full rows,
source identity and hashes. Structured filters should be parameterized, and
result JSON should retain why excluded candidates were omitted when requested.
Do not flatten user preference, review status and technical qualification into
one `readiness` string.

## Narrative/evidence contract before retrieval

Create a separate `pptxgengo.proposal-narrative.v1` brief to precede visual
selection. Minimum objects:

- `brief`: client/project name or explicit synthetic identity, authoring date,
  audience/seniority, decision sought, time horizon, confidentiality and
  source packet hashes.
- `slides[]`: stable ID/order, audience purpose, slide role, one-sentence
  takeaway, assertion-style title, required detail, emphasis targets, preferred
  visual relationship, and `evidence_refs[]`.
- `claims[]`: stable claim ID, exact assertion, claim type, evidence references,
  qualification/limitation, confidence, applicable date/range, status
  (`supported|hypothesis|synthetic|unverified`) and allowed wording.
- `evidence[]`: stable ID, source name/URI/hash, author/publisher, publication
  date/access date, page/slide/paragraph locator, exact quote or typed extract,
  whether quoted/paraphrased, claim support relationship, permission/context,
  and caveats. Quotes require exact attribution and a source locator; invented
  or unverified material is not made stronger by layout selection.
- `selection[]`: slide ID, candidate contract IDs and versions, selected one,
  match reasons to roles/cardinality, preference consulted, fit envelope,
  rejected alternatives with reason, human decision and timestamp.
- `review[]`: factual review, source-date check, native fit result, visual QA,
  corrections and remaining manual interventions.

Draft argument/evidence first. Then retrieval checks role structure and capacity;
when no approved component fits, search a second pattern, split the slide, or
record an explicitly manual new composition. Never delete a source claim or
qualification to fit the chosen layout. Synthetic examples stay labeled as
illustrative; the quote/evidence page is included only if authentic attributed
material exists, matching the Wave6 plan.

## Fresh ~12-slide Wave6 brief and selection trial

Use a fresh, clearly synthetic case to test process without implying client
facts: **Northstar Health Cooperative** (fictional payer), evaluating a governed
enterprise data platform modernization. Audience: COO, CIO, CDO, security and
data-product leaders. Decision: authorize a bounded foundation-and-first-use-case
assessment with stage gates, subject to baseline validation. Horizon: first
release followed by later adoption waves. All numeric benefits, timeline,
investment and staffing assumptions are labeled illustrative until validated;
no real organization/person/customer claims. A source packet should contain a
fictional RFP excerpt, synthetic current-state inventory, risk register, target
outcomes and stakeholder interview notes, each with immutable hash/locator.

Recommended sequence follows the Wave6 hard-pattern set; it is a brief outline,
not final authored copy:

| # | Narrative slide role | Needed content/evidence before pattern selection | Candidate family (only if qualified) |
|---:|---|---|---|
| 1 | Decision framing | Decision requested, one-sentence case for action, scope and audience | Cover/editorial; current dense-cover composition is mechanics-only |
| 2 | Executive decision | Recommendation, decision, three proof points, caveats | Assertion plus evidence cards; use measured layout if slots fit |
| 3 | Needs and response | Client needs aligned to evidence-backed response and qualification | EnableComp5 response rows; Wave2 fixture evidence, no whole-layout promotion |
| 4 | Five-phase approach | Five stages, activity, output and timing | Dense five-phase recipe, measured fixture only |
| 5 | Phase detail | Objective, four activities/work products, acceptance and scope | Dense phase-detail recipe |
| 6 | Workflow matrix | Five workflows across constraint, operating model and measure | Dense matrix recipe |
| 7 | Architecture | Outcomes/users, lifecycle, services, governance/security, placement, ecosystem | Wave3 layered architecture only after port/relationship qualification; until then authored/manual with caveats |
| 8 | Parallel process paths | Two processes, five steps each, approvals and illustrative result | Wave3 process-path contract only after source-control/translation/stress acceptance |
| 9 | Roadmap | First release and later waves over 12 months, overlaps, milestones and accountable owners | Dense release-roadmap recipe; assumption-checked, not general date-driven scheduling |
| 10 | Delivery team | Roles, pods, client counterparts, allocations, shared specialists and transition | Dynamic pods/team measured fixtures; explicitly synthetic, check cardinality envelope |
| 11 | Roster/biographies | Representative role profiles, needed skills and responsibility; avoid fabricated credentials | Wave2 roster/bio bounded fixture; use synthetic role profiles, no false biographies |
| 12 | Evidence and decision gates | Authentic attributed evidence only; otherwise risks, validation plan, gates and open questions | Evidence/quote composition only when packet contains authentic source; else use needs/response or numbered explanation pattern |

For all 12, the narrative spec records title/takeaway and each assertion's
claim/evidence before candidate search. Include one normal-content run, one
increased-text/cardinality run and one colleague-edited return. The second agent
should execute only from saved brief, chosen contract IDs, portable specs,
documented CLI commands and proof paths, with no private coordinates. Save
selection rationale, alternative/rejection notes, timestamps, native measurement,
renders/reviews, and exact final source/spec/contract/artifact hashes.

## Immediate Wave4 implementation order

1. Freeze an inventory snapshot/hash and fix count documentation to distinguish
   `catalog-index-report.md` (17,451, historical) from v3 (17,456).
2. Create portable schema and migrate the five recipe records plus the four
   source-bound text contracts without changing their existing qualification
   state. Link existing dense/layout/dynamic/visual proof manifests rather than
   copying evidence without hashes.
3. Add stable structured selection metadata and preference overlay joining by
   shortlist hash; rebuild SQLite only after contract/index validation.
4. Implement Go read-only `find` and `inspect`; then preview path resolution.
   Exercise inventory-versus-qualified filters and stale/missing artifact errors.
5. Add `instantiate` for the already measured dense recipes first, preserving
   contract and narrative traceability; send result into existing probe/measure/
   fit/build/verify workflow.
6. Author the synthetic Northstar brief/evidence packet and ~12-slide narrative
   specification. Then perform the Wave6 workflow and independent handoff.

The key release test is workflow reproducibility and evidence/fit correctness,
not the size of the inventory or retrieval rank alone.
