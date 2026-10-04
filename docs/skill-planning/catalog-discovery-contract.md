# Modern library discovery contract

Status: Wave 2/3 native unified discovery implemented, October 3, 2026. The current
v3 source freezes all 248 committed definitions at
`e91e0d7771000b7386f1ea52f51252f0f0a134fd`; source revision is `wmds-library.v3`.
Final native reference acceptance and coherent release installation remain pending.
Discovery reads pinned
modern template sources without changing their geometry, hashes, IDs or content
bindings. No template content envelope or new native review is established by
this work.

## Available in the repository CLI

```sh
pptxgengo design library-catalog
pptxgengo design library-catalog --template-keys architecture/layers-nav
pptxgengo design library-search --items 4 --item-role point --roles point,key-message --limit 8
pptxgengo design library-search --items 4 --item-role point --roles point,key-message \
  --visual-forms icon,image --query "goals and outcomes" --limit 8
```

These commands require a release containing this implementation. An older
installed CLI or skill may lack `library-search` and the modern `design` route.
Check the installed release's paths and help before using it. Direct repository
binary equivalents are `pptxdesign library-catalog` and `pptxdesign
library-search`; provide `--bundle library/wm-design-system/v3` explicitly when
using the repository catalog command.

`library-search` also accepts `--structures`, `--include-deprecated`,
`--bundle`, and a source override that must match the pinned snapshot. It accepts
the installed wrapper's `--engine` argument as a passthrough only; search does
not evaluate engine/bundle compatibility, run layout or render slides. The
result records the requested engine and `compatibility:
not_evaluated_search_only`. Limit is 1–100. Item count is nonnegative; zero
means no count hint.

The existing modern catalog continues to support closed content binding and
foundation builds. Search does not modify those APIs. The two retained card-row
templates keep their typed values contract; all other existing fixed array,
required slot and source-hash rules remain authoritative.

Search results publish the supported vocabulary. Available structure hints are
`comparison`, `gantt`, `hierarchy`, `layers`, `matrix`, `network`, `process`,
`repeated-items`, `sequence`, `swimlanes`, `table`, and `timeline`. These come
from explicit known source node types; source labels alone do not infer a
timeline or hierarchy. Content roles are `content`, `headline`, `icon`, `image`,
`key-message`, `metric`, `navigation`, `numeric-data`, `person`, `point`, `quotation`,
`relationship`, `repeated-content`, `sequence-item`, and `tabular-data`.
`column` identifies a declared column collection rather than point units.
Visual forms are `cards`, `chart`, `diagram`, `icon`, `image`, `table`, and `text`.
Unknown hints remain unmatched ranking hints rather than hidden hard exclusions.

## Scenario and structure are separate signals

A slide with four goals, four problems or four recommendations may share the
same useful arrangement. Original scenario labels do not exclude it. Search
therefore accepts independent dimensions:

| Dimension | Meaning | Example |
| --- | --- | --- |
| Free text | Original purpose, name, family, ID or declared uses | goals and outcomes |
| Content roles | Structural information affordances | point, key-message, image |
| Structures | Arrangements derived from actual source nodes | repeated-items, sequence, comparison, table |
| Visual forms | Present visual affordances | text, cards, image, icon, diagram, chart, table |
| Item count | Desired count within one repeated content group; optional `--item-role` associates it with a role | four points |

All these fields are ranking hints. They never gate on words such as “problem”
or “goal.” Deprecated lifecycle is the only default exclusion. Requests with
missing hints retain candidates and explain what did not match, including
zero-score candidates if needed. Callers should inspect reasons rather than
interpret a returned candidate as a recommendation already proven to fit.

Optional image/icon hints add modest preference to arrangements already
containing those forms. Their absence does not reject another good arrangement.
Adding an image to an arrangement without an image zone remains a foundation
composition edit, not an implicit new template content slot.

## Portable catalog projection

`library-catalog` emits existing contract fields plus:

| Field | Source and authority |
| --- | --- |
| `purpose`, `uses` | Source-authored annotations; advisory suitability |
| `source_advisory_budget` | Original word/mark budget object; not measured capacity |
| `source_advisory_slots` | Original named notes, `maxChars`, counts and other guidance; preserved without pretending they map one-to-one to executable slots |
| `discovery` | `pptxgengo.wmds-discovery.v1`, derived from pinned source structure |

The discovery projection contains sorted `content_roles`, `structures`,
`visual_forms`, and `component_types`, along with:

- `content_groups`: source pointer, component type, role, exact source count,
  derivation basis, and primary/supporting/nested scope. Groups come from
  meaningful same-type-and-role scene siblings, semantic binding arrays, or
  typed adapter exact-count declarations. Cells, row tuples, rich-text runs,
  coordinate pairs and body paragraph blocks are not semantic groups.
- `zones`: actual slide body scene-node pointer, type, derived role, and supplied x/y/w/h bounds.
  Missing source dimensions remain missing. Discovery does not calculate final
  layout, padding, measured text bounds, or intersections.
- `relationships`: structural relationship kind, supporting source pointers and
  derivation basis. Comparison is recognized from an explicit beforeafter node,
  or equal-length parallel lists in aligned cards with a connector joining the
  card edges. Scenario names and labels do not establish this relationship.
- `frame`: source rail, footer, surface, split and navigation fields. This is
  slide chrome metadata, not a new canonical frame ID or a master definition.
- `capability`: separate adapter, queried-content build, specimen-review,
  content-envelope, and capacity-basis statuses.

All current item group counts describe the fixed source arrangement. A count
match alone does not establish that the group carries the requested meaning or
that the supplied copy fits. Decorative box/rule siblings do not create content
groups. Binding-array groups retain the nearest source component's role where
available. Nested groups can legitimately have multiple counts; search shows
the matched pointer and component so an agent can inspect the correct level.
Use `--items 4 --item-role point` to prefer four point units rather than any
four-element group such as table columns. Without `--item-role`, the count hint
can match any group, and the result explains the role. Even a point-role match
can represent four labels rather than four full paragraphs: inspect the actual
contract before binding supplied content.

Primary groups have more ranking weight than supporting or nested groups. A
standalone point list narrower than 60% of the source body span, or without a
known width, is treated as supporting; a list inside a card or bio is nested.
This is an explicit source-geometry heuristic, not a text measurement or proof
of rhetorical importance. `scope_basis` reports its basis. Bare scene text
siblings are not counted as interchangeable points; diagram labels and blank
blocks do not become generic point groups. Metric, person and quote cards retain
their specific roles rather than being pooled with generic idea cards.

The current structural vocabulary is deliberately independent of client/domain
language. `point` covers text/list/card units; `key-message` includes the slide
headline as an available main-message location. This is an affordance, not a
claim that every template has a separate takeaway box. Unknown source types
remain visible in `component_types` and `zones` with role `content`; discovery
does not invent a special meaning for them. Nested table row fields named
`type` are data, not scene component discriminators.

## Capability and evidence boundaries

| Status | Current meaning |
| --- | --- |
| `content_adapter: defined_closed_binding` | A source-pinned content adapter is defined; no build for the query has run |
| `content_adapter: capabilities_pending` | Definition declares unresolved implementation capabilities |
| `build_for_query: not_executed` | Actual user content was not compiled or rendered by search |
| `specimen_review: not_loaded_by_discovery` | Discovery has not imported reviewed specimen evidence; this does not claim specimens were never reviewed |
| `content_envelope: not_established_by_discovery` | No reusable arbitrary-copy envelope is established by this result |
| `capacity_basis: source_advisory_not_measured` | Budget and character annotations are source guidance |

Existing source/template identities and provenance remain attached to every
result: canonical opaque template key, source revision, file hash, template
revision, lifecycle and replacement link. Review of a particular paired
specimen remains scoped to that specimen. A renderer's implementation support,
layout fit for actual supplied copy, native appearance and broad qualification
are different facts.

## Deterministic search output

Search emits `pptxgengo.wmds-library-search.v1` with the query, policy, and
`matches`. Each hit includes its catalog definition, score, reasons, unmatched
hints, and `fit_status: not_measured_for_query`.

Scoring is additive:

- Matching requested content role: 12 each.
- Matching requested structural feature: 16 each.
- Matching requested visual form: 6 each.
- At least one source group with the requested exact count and, when supplied,
  the requested item role: 35 for a primary group, 15 for supporting, or 10 for
  nested. The strongest group adds its weight once; multiple matches do not
  accumulate the count bonus. `count_match` exposes the selected group and weight.
- Each distinct scenario/label word matching the key, name or family: 8.
  Otherwise a word matching only purpose/uses adds 3. Take the strongest field
  match per word and cap the combined text contribution at 12. A direct status
  arrangement therefore ranks above incidental status language in another
  template's purpose, while structural item/role hints retain their independent
  weight. Scenario matches remain preferences rather than exclusions.

Duplicate dimension hints do not add points. Scenario tokenization is case
insensitive and splits at non-letter/non-digit characters. Scores tie-break by
canonical template key. No wall clock, random seed, preference history or
external model affects ranking. Search is intentionally a transparent first
pass, not semantic embedding retrieval or an assertion of design quality.

Recommended skill workflow: retrieve candidates; inspect topology, actual
slots and guidance; bind supplied content; review fit reports and renders;
offer two or three meaningful alternatives when the layout choice is material.
Log the chosen design and any composition amendment against stable slide IDs.

## New template intake and versioning

Canonical keys are opaque authoring identities, such as `architecture/layers-nav`.
Do not derive meaning by parsing those keys. Do not rename them because a new
client uses another scenario. The authoring document references the key;
the document lock pins the bundle revision, source hash and template revision.

Adding or revising a template requires:

1. A stable unique key, lifecycle, template revision and declared source snapshot.
2. Purpose/use annotations and original slot/capacity guidance.
3. Implementable native node types and a complete explicit content adapter.
4. Derived groups, zones and frame metadata, reviewed for semantic suitability.
5. Source and changed-content specimens with separately recorded native review.
6. Any claimed content envelope supported by evidence beyond those specimens.

New templates enter a new additive pinned bundle snapshot. Existing snapshots
and their hashes remain reproducible. For an incompatible API/topology change,
use a new contract revision or template identity with an explicit migration;
deprecate the old identity with a replacement link where appropriate. Retain
legacy typed adapter distinctions. The existing approved snapshot inventory
count checks remain; new snapshots need an intentional inventory migration,
not a changed count disguised as part of search.

## Unified SQLite projection and legacy compatibility

The Wave 1 portable JSON projection is now indexed by the implemented native Go
adapter. It does not mutate the original legacy `catalog-library.sqlite`.
Existing `lib find` remains legacy inventory/contract search with its own
qualification defaults; unified modern discovery uses `design library-find`.

The actual `pptxgengo.unified-library-index.v1` schema is:

| Table | Stored projection |
| --- | --- |
| `meta` | Schema, pinned inputs, options, source revision, registry and projection hashes |
| `entities` | Canonical namespace/kind/key/filter columns plus portable entity JSON |
| `dimensions` | Independent role, structure and visual-form values |
| `content_groups` | Semantic cardinality, role/scope and source pointer with derived JSON |
| `dependencies` | Typed resource links with derivation basis |
| `entity_fts` | FTS5 name/purpose/body text; not a CLI ranking gate |

Zones, capacities, capabilities, evidence states and artifact links remain in the
entity JSON rather than separate hypothetical tables. The database is rebuildable
from pinned portable inputs. Modern input definitions are complete; legacy
inventory uses compact hash-bound original-row references with full lazy inspect.
The measured combined-index experiment decreased from approximately **113 MB to
60 MB**; size depends on selected inventory/gallery inputs. Full inspection still
requires the exact original pinned legacy SQLite and available contract resources.

Future retrieval can add richer semantic descriptions and user preference
overlays. Preserve the plain ranking explanation and distinguish source-advisory
values from derived, measured and reviewer-authored assertions. Search never runs
actual content fit; explicit `library-fit` experiments do.

## Unified native SQLite implementation

Available commands added after Wave 1:

```sh
pptxgengo design library-index --bundle v3 --out ./library.sqlite
pptxgengo design library-find --index ./library.sqlite --query 'weekly status' --kinds template
pptxgengo design library-inspect --index ./library.sqlite --id cards/3
pptxgengo design library-preview --index ./library.sqlite --id cards/3
pptxgengo design library-fit --bundle v3 --spec alternatives.json --out ./candidate-review
```

The native Go `modernc.org/sqlite` driver builds schema
`pptxgengo.unified-library-index.v1`. Entities contain full template contract and
source slide, primitive/component/composite definitions, resolved frame variants
and registered originals. Tables project dimensions, content groups, dependencies
and an FTS5 text index alongside the authoritative entity JSON. The current CLI
uses SQL kind/namespace filters and transparent in-Go soft ranking; FTS5 is
available for downstream SQL consumers and is not a hidden search gate.

Modern IDs are `wmds/template/<opaque-key>`, `wmds/<primitive|component|composite>/<source-id>`,
`wmds/frame/<rail>-<footer>[/split/<split>]`, and `wmds/asset/<registry-key>`.
Shared YAML template references keep the original opaque key. Local frame
references use the base `wmds/frame/<rail>-<footer>` with `frame_options.split`;
split projection IDs are catalog variants, not additional accepted YAML frame IDs.

Optional `--legacy-index` (and `--legacy-root` for nonstandard original database
locations) projects searchable summaries and hash-bound references
under `legacy/inventory/<id>`; inspect lazily reads the exact original JSON row and projects contract metadata/available verified definitions under
`legacy/contract/<id>`. A missing original contract remains explicitly an inspect-
original-contract capability; index presence does not establish execution support.
Raw legacy inventory JSON stays in the original pinned SQLite instead of being
duplicated into the unified projection. Each referenced row has its own hash and
is verified on inspect with a parameterized read-only lookup. This reduces catalog
size while retaining full inspect metadata. Modern and legacy IDs do not overwrite
one another. Qualification states are
preserved rather than promoted.

`--gallery` imports source/alternate preview and value/foundation links only from
a matching source revision, exact canonical template key/revision and pinned
contract file hash. A shared family-file hash cannot authorize a different
template's preview or review state. The index pins its bundle,
source, original legacy SQLite, available legacy definitions and gallery manifest.
Open is read-only, validates pins/asset-registry/projection integrity, and checks
SQL identity/filter columns against the authoritative JSON. Indexed source,
bundle, gallery and legacy-contract pin paths are bounded relative resources;
parent traversal and symlink steps fail. Explicit trusted asset-registry paths
retain their approved resolution behavior and hashes. Preview verifies the
artifact bytes each time. A missing matching render yields a definition and
explicit preparation note, not a fabricated preview. `--bundle`, `--source`,
`--legacy-index`, `--legacy-root` and `--gallery` overrides support deliberate relocation while
retaining hashes.

`library-fit` is a concrete actual-content experiment. It consumes a strict
`pptxgengo.wmds-template-document.v1` with unique supplied-content candidate IDs;
each candidate carries its own complete closed binding. It writes independent
candidate decks/foundation/layout reports and a summary preserving failures.
Success means Go binding/layout succeeded and native/visual review remains
pending. It neither guesses cross-template meanings nor substitutes source copy.
An experiment can complete with failed candidates: inspect `passed`, `failed` and
per-candidate results. YAML projects remain the maintained authoring route; this
JSON is a derived candidate input.

## Implementation evidence and remaining gates

Focused unified catalog/discovery race tests passed in 38.646 seconds, and focused
CLI catalog race tests passed in 4.350 seconds after independent review fixes.
Regressions cover projection/SQL tampering, original legacy input/resource drift,
read-only access, exact gallery identity, preview drift, relocation, semantic
ranking and ambiguous candidate JSON. Actual supplied-content experiments built
two alternative decks and retained a deliberately invalid third candidate with
its binding failure. These are technical evidence, not native appearance or an
arbitrary-content envelope.

The maintained authoring route is implemented `design project` using
`pptxgengo.deck-document.v1` YAML; see [runtime bounds](../../internal/deckproject/README.md)
and the [executable starter](../../examples/deck-project/README.md). The repository
skill pack routes agents into progressive content/discovery/review references.
Current v3 gallery/native work and final coherent installer/package checks are
pending integration gates. Historical v2 paired evidence remains separate.
