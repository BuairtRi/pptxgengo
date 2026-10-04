# Continuing design-template intake

**October 3, 2026 — frozen v3 reviewed and installed as local.7; continuing intake remains active.**
The earlier read-only audits below explain the incoming delta. They do not describe
missing capabilities in the current application.

## Current delta — refreshed October 3, 2026

The packaged v2 source is pinned to commit
`7bdcaee5030a12275a1f881a8542f4d302d207df` with 167 definitions. The frozen v3 source in
`library/wm-design-system/v3` contains **248 committed definitions: 81 additions,
one revised definition (`about/glance`), no removed identities, and 46 retained
identities moved between family files.** There are 247 active definitions and one
retained deprecated definition. Source commit is
`e91e0d7771000b7386f1ea52f51252f0f0a134fd`; bundle revision is `wmds-library.v3`.
The eight formerly working-tree-only core additions are committed at that pin.
Identity comparison uses `id/variant` and complete definition values, not generated
gallery counts, family filenames, or the old changelog baseline. The published v2
snapshot remains immutable.

The earlier audit at **2026-10-03 15:04:31 UTC** observed HEAD
`4733162c06f9b28933bf6028d2bf39519ee4bda1` with 240 committed definitions and eight
uncommitted core additions. That historical working-tree state is superseded by
the frozen v3 commit above. Its recorded hashes below retain audit provenance;
the authoritative current hashes are in the v3 inventory/bundle manifests.

Tokens, frame definitions, and component inventory JSON remain semantically
identical to the pinned v2 versions. The expanded examples use existing nav,
left/right, compact/tall, and split-frame formats; they do not introduce a new
frame geometry contract. Five additional scene node types now have native adapters:
`logoslot`, `dotmap`, `teamcurve`, `device`, and `plane`.

## Historical audit — superseded `f0c8a5a` snapshot

The earlier clean `f0c8a5a` audit found 193 definitions, 26 additions, one revision,
and no removed identities. The list below preserves that earlier audit only.
Rounds 7, the frame-coverage update and the now-committed core additions
supersede its counts. A path change still must not imply a new template identity.

| Family | Added definitions |
| --- | --- |
| About | `about/alliances`, `about/client-logos`, `about/industries`, `about/locations`, `about/locations-international` |
| Architecture | `architecture/app-ecosystem`, `architecture/cloud-landing-zone`, `architecture/cloud-landing-zone-azure`, `architecture/layers-3d`, `architecture/layers-3d-systems`, `architecture/layers-icons`, `architecture/migration-waves`, `architecture/site-rollout`, `architecture/strangler`, `architecture/transition-cloud-to-cloud`, `architecture/transition-filmstrip`, `architecture/transition-hybrid` |
| Divider | `divider/inverse-square` |
| Team curve | `team-curve/agents`, `team-curve/agents-bands`, `team-curve/agents-nav`, `team-curve/agents-roles`, `team-curve/before-after`, `team-curve/build-together`, `team-curve/build-together-bands`, `team-curve/build-together-roles` |

`about/glance` is revised. Existing IDs moved between family files include the
rename from `openers.json` to `covers.json` and the added architecture/about/core/
runbooks families. The source history also introduces `logoslot`, `dotmap`, and
`teamcurve` primitives. These received explicit native adapter implementations and focused regression
checks; successful adapter checks remain distinct from native specimen acceptance.

## Published identity and revision

1. Template ID plus variant is stable across family/file changes.
2. Each published catalog records source commit, bundle revision, definition
   revision, file hash, definition hash, lifecycle, and replacement identity.
3. A source draft is discoverable as an intake record without implying build
   support or a reviewed content envelope.
4. Runtime indexes resolve the pinned release snapshot. A maintainer may inspect
   an incoming snapshot separately; it does not replace the active build index.
5. Existing decks lock the definition they used. A migration identifies affected
   slides and requires explicit adoption of changed definitions.

## Intake sequence and deliverables

- Inventory additions, changes, removals, lifecycle updates, moved source files,
  and added primitives/components/frames/tokens.
- Extract scenario purpose plus structural zones, content roles, counts, visual
  forms, declared relationships, and supported transformations.
- Compare required source vocabulary to implemented native adapters. Record
  unsupported features as pending capabilities with source locators.
- Assign independent families/adapters to agents with disjoint file ownership;
  shared primitive implementation is coordinated before family compilation.
- Create a new source-pinned bundle revision without modifying prior releases.
- Compile reference and real-content alternative specimens after implementation.
- Record source/content/renderer/native/visual evidence separately. Successful
  rendering alone does not establish arbitrary-content qualification.
- Publish catalog metadata, previews, editable examples, compiler compatibility,
  and migration notes in the same release as the supporting code/skill.

## Metadata and maintained decks

The catalog's scenario and structural search both use versioned portable
metadata. SQLite is a rebuildable projection. Existing source advisory word and
character budgets remain advisory until a measured/reviewed basis is supplied.

Custom slides may retain shared component references while using deck-local
composition geometry. Promoting a local template is an explicit intake action
with ancestry and content contract preserved. A new library version never
silently changes a local composition or the definitions locked by an existing
deck.

## Current implementation and next intake action

- Source loading, source-hash allowlists and catalog counts now intentionally accept
  the immutable v3 snapshot while retaining v2.
- `scene_intake_architecture.go` implements `logoslot`, `device` and `plane`.
  Logo slots contain-fit registered assets or produce native placeholders;
  devices/planes remain editable native groups with validated labels/anchors.
- `scene_intake_geography.go` uses frozen, attributed geographic data and a shared
  SVG base with PNG fallback, plus native editable markers/rings/labels. Runtime
  builds do not fetch or approximate geography.
- `scene_intake_curves.go` implements validated native cubic band/line geometry,
  phase controls and label placement for `teamcurve`.
- Closed bindings and structural discovery cover the new vocabulary. All 248 source
  specimens compile, including the 81 additions and revised `about/glance`; real
  supplied-content smoke checks and independent focused reviews have run.
- The unified SQLite projection, authored YAML runtime and progressive skill are
  implemented. Actual-content experiments remain caller-mapped closed bindings;
  source capacity notes do not grant arbitrary shared-template reflow.

A native 248-slide reference export has succeeded and every-page review/fixes are
in progress. **Final native acceptance and coherent release/installer validation
remain pending.** Gallery review is tied to exact template/revision/source hashes;
historical paired v2 evidence is retained separately and is not relabeled as v3
acceptance. Next action is to resolve/re-export current findings, complete final
review and package the matching compiler, catalog, skill, assets and gallery.
Subsequent source changes use the continuing additive intake procedure above.

## Latest audit: additions by source family

Each row names the actual file under `wm-design-system/templates/library/`.
The count includes frame/treatment variants as currently authored source identities.
Whether future variable-frame/treatment/count contracts collapse some variants
is an authoring API decision, not permission to discard their present definitions.

| Source file | Added count | Exact added identities |
| --- | ---: | --- |
| `about.json` | 9 | `about/alliances`, `about/alliances-right`, `about/alliances-tall`, `about/client-logos`, `about/industries`, `about/industries-left`, `about/industries-nav`, `about/locations`, `about/locations-international` |
| `approach.json` | 10 | `phases/four-right`, `team-curve/agents`, `team-curve/agents-bands`, `team-curve/agents-nav`, `team-curve/agents-roles`, `team-curve/agents-split`, `team-curve/before-after`, `team-curve/build-together`, `team-curve/build-together-bands`, `team-curve/build-together-roles` |
| `architecture.json` | 19 | `architecture/app-ecosystem`, `architecture/cloud-landing-zone`, `architecture/cloud-landing-zone-azure`, `architecture/cloud-landing-zone-gcp`, `architecture/cloud-landing-zone-nav`, `architecture/layers-3d`, `architecture/layers-3d-systems`, `architecture/layers-icons`, `architecture/layers-left`, `architecture/migration-waves`, `architecture/migration-waves-tall`, `architecture/nested-right`, `architecture/site-rollout`, `architecture/strangler`, `architecture/strangler-split`, `architecture/transition-cloud-to-cloud`, `architecture/transition-filmstrip`, `architecture/transition-hybrid`, `architecture/transition-hybrid-nav` |
| `argument.json` | 1 | `comparison/us-versus-others-four-left` |
| `commercials.json` | 2 | `pricing/fixed-fee-left`, `pricing/options-right` |
| `core.json` | 26 | `cards/narrative-2x2`, `cards/narrative-2x2-band`, `cards/narrative-2x2-icon`, `cards/narrative-2x2-nav`, `cards/narrative-2x2-split`, `cards/narrative-2x2-states`, `cards/narrative-2x3`, `cards/narrative-2x3-badge`, `cards/narrative-2x3-tall`, `cards/narrative-bullets-2x2`, `cards/narrative-bullets-2x2-split`, `key-message/stat-left`, `narrative/lead-in-right`, `quote/cards-left`, `quote/cards-right`, `quote/with-context`, `text/four-column`, `text/four-column-tall`, `text/icon-columns`, `text/lead-two-column`, `text/photo-columns`, `text/single-column`, `text/statement-and-support`, `text/three-column`, `text/three-column-nav`, `text/two-column` |
| `covers.json` | 3 | `agenda/schedule-right`, `agenda/schedule-tall`, `divider/inverse-square` |
| `evidence.json` | 4 | `chart/column-left`, `stats/four-metrics-right`, `status/four-panel-right`, `status/four-panel-tall` |
| `proof.json` | 3 | `case-studies/two-nav`, `case-study/exhibit-left`, `case-study/quote-outcome-right` |
| `runbooks.json` | 3 | `runbook/go-no-go-right`, `runbook/overview-nav`, `runbook/step-list-left` |
| `solution.json` | 1 | `process/current-future-right` |

The eight formerly working-tree-only definitions (now committed in v3) are `cards/narrative-2x2-band`,
`cards/narrative-2x2-icon`, `cards/narrative-2x2-states`,
`cards/narrative-2x3-badge`, `quote/cards-left`, `quote/cards-right`,
`text/icon-columns`, and `text/photo-columns`. Their source file hash at
the historical working-tree audit time was:

`templates/library/core.json` —
`302894382b5a1c1997b9c572c505576a7f098347ec32ab27a7bd5990f837cfd2`.

The audited `templates/library/*.json` file set (including `_gaps.json`) has
manifest fingerprint
`b396fa7f5efd0a89a343bf2ba2b710949c89956ed89fd2ffd8356535e36684d0`:
SHA-256 of UTF-8 JSON mapping package-relative file paths to their file SHA-256,
with keys sorted and separators `,`/`:` and no trailing newline. This identifies
the observed working-tree library independently of later source changes; it does
not pin the renderer/geographic data, which must be included in the final bundle
manifest separately.

No existing core definition was changed relative to HEAD by these eight additions.
`about/glance` is the sole existing-template revision relative to the Go bundle:
rev 2 removes icon cards/oversized quote, replaces them with plain lists, a compact
quote, and a three-stat proof panel. Its original pinned specimen is retained.

## Changelog and source authority

Recent source commits are `2fd6009` (family reorganization/new primitives),
`f0c8a5a` (international map, team curves, transition architecture), `37439f7`
(GCP landing zone/core layouts/frame examples), `2730d4d` (actual geography/finer
map dots), and `4733162` (frame coverage). The generated changelog's headline
**143 added/16 revised/1 deprecated/20 moved** compares against the old
`pptxgengo-wmds-v1` tag at `d8f45d4`; it is not the delta against the Go v2 pin.
It also did not encompass the eight uncommitted additions at the historical audit time. Use the direct
identity/value inventory above for implementation scope.

The now-committed source authoring-reference policy explicitly allows card treatments,
frame variation, and count changes within stated slot ranges. That is incoming
authoring policy, not an already executable Go contract. The current Go content
binding API freezes geometry and exact array counts. Supported custom/local
composition can express these variations; general shared-template reflow requires
an explicit typed adaptation contract and implementation. Do not advertise flexible
counts solely because source advisory slots give a minimum/maximum.

No applicable `AGENTS.md` was found in either project or their checked ancestor
directories. No source repository files were modified by this audit.

## Prerequisite capability inventory

Instance counts below refer to the 81 added slide bodies, not independent
capabilities or reusable template counts. The current implementation column supersedes the earlier audit of missing
adapters. Existing writer geometry alone was not treated as implementation proof;
explicit adapters and focused checks now support those entries.

| Capability | Added instances / dependent definitions | Source contract and required native behavior | Current Go implementation |
| --- | --- | --- | --- |
| `logoslot` | 97 nodes / 8 definitions | `explorations/components.src.html:1434`: contain image inset by 6 pt; otherwise bordered hatch/name placeholder; optional caption. Used by three alliance variants, client logo wall, AWS/Azure/GCP/nav landing zones. | Implemented in `scene_intake_architecture.go`: contain-fit original assets, bounded native placeholders/captions and explicit field validation. Native acceptance remains specimen-scoped. |
| `dotmap` | 2 nodes / 2 definitions | `components.src.html:1445`: US or Americas/Great Britain panels, public-domain outlines and Great Lakes holes; projected lon/lat dots/office markers/hub rings, directional labels, explicit label suppression and framed inset. `about/locations` and `about/locations-international`. | Implemented in `scene_intake_geography.go`: exact frozen geographic data and shared SVG/PNG base; markers/labels native editable; provenance and geometry/resource validation recorded. |
| `teamcurve` | 10 nodes / 9 definitions | `components.src.html:1482`: per-series values at explicit/default sample positions, cumulative area bands and independent dashed/solid line series, explicit max, Catmull–Rom-to-cubic conversion, phase dividers/labels, `phaseLabels:false`, `phaseH:0`, direct labels/false suppression. | Implemented in `scene_intake_curves.go`: validated native cubic bands/lines, phase controls and direct labels; focused geometry and rejection tests. |
| `device` | 3 nodes / 1 definition | `components.src.html:1411`: 36 pt icon plus bold small label and optional label-style subtitle, centred; used by `architecture/app-ecosystem`. | Implemented in `scene_intake_architecture.go`: native icon/text group with registry-backed icons, label bounds and source field validation. |
| `plane` | 8 nodes / 2 definitions | `components.src.html:1549`: diamond plane with token surface, label and title centred in middle 60%; used by `architecture/layers-3d` and `architecture/layers-3d-systems`. | Implemented in `scene_intake_architecture.go`: native diamond plane, surface/typography validation and connector anchors. |

The `logoslot` row's dependency set is eight definitions: `about/alliances`,
`about/alliances-right`, `about/alliances-tall`, `about/client-logos`,
`architecture/cloud-landing-zone`, `architecture/cloud-landing-zone-azure`,
`architecture/cloud-landing-zone-gcp`, and `architecture/cloud-landing-zone-nav`.

Existing core card fields newly exercised by the incoming definitions are `band`,
`bandNumber`, `padTop`, and `placeholder`. The current `sceneCardSource` already
declares them, together with `badge`, `state`, `tag`, icons and edges. Those source
shapes need compilation/visual qualification, not an automatic new adapter.
Tables, charts, cylinders, containers, connectors, layer rows, matrices, chevrons,
photos, text and all incoming frame formats likewise have existing source handlers.
That establishes an implementation starting point, not acceptance of the new
geometries or arbitrary replacement copy.

`screen` appears in source primitive documentation/renderer but in none of these
81 new template bodies. It is a separate future capability, not a blocker for
this inventory. Tokens/frames/components equality does not prove primitive
completeness: the browser's executable vocabulary lives in the renderer source.

## Dependency-first parallel implementation slices — historical execution plan

With four total agent slots, use three adapter lanes plus an integration owner.
Avoid concurrent edits to central dispatch, source pins, catalog counts, binding
projection, or shared typography files.

| Lane | Ownership / deliverable | Dependencies and handoff |
| --- | --- | --- |
| A: logo and architecture primitives | New adapter files for `logoslot`, `device`, and `plane`; typed fields, asset resolution, original/derivative provenance, native groups and diagnostics | Existing image/icon/shape/text APIs. Handoff includes field kinds and count/key requirements; integration owner wires dispatch and content projection. |
| B: geography | New dotmap adapter plus frozen geographic data/provenance; US and Americas/GB base generation, markers, rings, labels and source-faithful panel projection | Source data extraction/hash decision and reusable base representation. Do not fetch maps at runtime or replace them with approximate shapes. |
| C: team curves | New teamcurve adapter; native cubic areas/line series, sample/phase validation, direct labels and phase rows; no screenshot flattening | Existing writer cubic paths; explicit coordinate/value semantics. Integration owner supplies source pins and central dispatch. |
| Integration owner | Freeze source snapshot; new immutable bundle revision; moved-family loading/count updates; wire adapters, exact content/key schemas and discovery metadata | Review working-tree additions before pinning. Keep v2 immutable; expected-count/source allowlists must be revision-specific. |

After these shared adapters land, fan out independent family specimen work:

1. **Core/covers**: 29 additions (cards/text/quotes/key-message plus agenda/divider)
   using existing node vocabulary, including incoming card treatments.
2. **About/architecture**: 28 additions plus revised `about/glance`; logo/map/device/
   plane dependencies and landing-zone/transition diagrams.
3. **Approach/frame variants and remaining families**: 24 additions; nine team-curve
   definitions depend on lane C, and the rest use existing components/frames.

Each family returns complete source and alternate-content scenes, content contracts,
used assets, overflow diagnostics, and a pending-native-review receipt. Final deck
assembly/native exports are sequential integration work. Review panels must include
left/right/tall/split variants, curve label/phase alignment, fine dot geography,
logo placeholders and replacements, and source-to-content mapping. Preserve old
reference evidence and publish the new library/catalog/runtime/skill together.

The original audit performed filesystem/JSON/Git reads and edited this planning
contract only. Subsequent Wave 2/3 work implemented and checked the adapters,
vendored the committed v3 snapshot and generated the candidate reference output.
Frozen v3 native/release acceptance subsequently passed, and local.7 was activated.
Continuing intake has a separate snapshot and acceptance gate; no commit/push is
implied by these implementation tasks.


## Active continuing intake

The source remains under active authoring. The corrected 17:50 UTC observation
contains 301 committed definitions and 369 working definitions (68 additions not
yet in HEAD). It adds 121 and revises 13 against v3, with no removed identities.
New body scene types are Venn and maturity; heat maps, subtitle cells, dots and
revised curves add nested field contracts. The frame delta adds a slim footer.

[PLAN section 15](PLAN.md#15-continuing-intake-capabilities-before-template-expansion)
and [capability evidence](intake-capabilities-20261003.md) describe ownership,
implemented prototypes, exact checks, and remaining template/frame qualification.
The dated snapshot is not a registered v4 library or a publication of upstream
drafts. Import ready committed definitions in subsequent family lanes, then
reobserve newer drafts when ready. Published builds retain their own pins.
