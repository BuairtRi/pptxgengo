# Deck source and portable project contract

**October 3, 2026 — authored YAML runtime implemented.** This document retains
broader Wave 1 design intent alongside the bounded Wave 2 implementation.
`design project` now loads/builds `pptxgengo.deck-document.v1` YAML; existing
JSON template/foundation routes remain available. The authoritative implemented
bounds are in [the runtime README](../../internal/deckproject/README.md), with a
[runnable starter](../../examples/deck-project/README.md). Native acceptance is
specific to each generated deck; package staging and installation are separate.

Related artifact: [document schema](../../schemas/deck-document-v1.schema.json).

## 1. Authority and ownership

One `deck.yaml` is the authoritative authored deck document. It contains ordered
slides, their actual visible content and data, asset declarations, template
selections, and deck-local template definitions. Agents and people edit this file.
The compiler produces a resolved scene JSON and editable PPTX. Those outputs do
not become a second silently competing source of content.

The shared library retains persistent, versioned definitions for tokens, grids,
primitives, components, composites, frames, and templates. A deck references these
definitions and pins the resolved versions/hashes in `toolchain.lock.json`.
SQLite indexes the definitions and their metadata; it is a reproducible projection,
not the source of truth. A project need not copy the full development repository.

| Artifact | Authority | Editing rule |
| --- | --- | --- |
| `deck.yaml` | Visible slide content, order, selected templates, local compositions | Human/agent authored; build preserves bytes; explicit fork/detach rewrites formatting with predecessor retained |
| Shared definitions | Reusable styles, geometry, vocabulary, template contracts | Library ownership; explicit version migration |
| `toolchain.lock.json` | Exact resolved runtime, skill, library, font and asset dependencies | Updated explicitly with a reviewed migration |
| Evidence and context | Source observations, audience, narrative rationale, claims and caveats | Updated as new material arrives; link affected slide IDs |
| Project state and decisions | Stage, next action, dependency state, approvals | Machine-readable state plus readable explanations |
| Resolved scene and reports | Build interpretation and diagnostics | Derived; regenerate from the authored source |
| PPTX and PDF | Editable delivery and rendered review/export artifacts | Manual edits are detected divergence requiring reconciliation |

Reader-facing slide drafts and outline packets are exports of the source/context
at a named version. Do not maintain independent copies of slide copy in Markdown,
YAML, and resolved JSON. Internal page briefs may add reasoning and evidence links
but must not silently change the visible content.

## 2. Document envelope

The schema identifier is `pptxgengo.deck-document.v1`. The JSON Schema uses draft
2020-12 and describes the JSON-compatible value tree after YAML parsing. It is a
structural contract, not a substitute for catalog-dependent validation.

Required root fields are `schema`, `id`, `title`, `year`, `toolchain`, and `slides`.
Optional fields are `context`, `assets`, `local_templates`, and `media_optimization`.
Unknown fields fail.
`year` is explicit. `toolchain.lockfile` is a package-relative path. A source can be
drafted with an unresolved lockfile, but a production build must resolve and verify
every dependency; the compiler cannot silently choose whichever release is current.

`slides` is an ordered array. Each slide has an immutable authored `id`, a
`content_kind`, one template reference, and a `values` object. Array order determines
presentation order; the ID never incorporates the page number. Titles, eyebrow,
source lines, labels, and diagram data belong in `values` under the selected
template's content schema. `brief` and `evidence_refs` connect to internal context
without becoming additional on-slide content.

Template references are `{scope: shared|local, id: ...}`. Shared IDs are opaque
catalog identifiers; local IDs name entries in `local_templates`. An optional
revision expresses an authored constraint; the lockfile records the exact resolved
revision and hashes regardless. There is no implicit fallback from an unknown
shared ID to an example, legacy template, or custom slide.

## 3. Content zones and composition vocabulary

A template exposes named content zones with JSON Schema value contracts. A zone
describes its communication role, required status, value type/cardinality, and
capacity information. Content binding is by zone name and stable item key, not a
PowerPoint object number or mutable array index. `values` uses the catalog's
declared zone names. Existing closed slot contracts remain valid as adapters; they
must not become unrestricted geometry mutation APIs.

Local templates declare `zones`, each containing a role, required flag, and value
schema. Their composition uses the same typed vocabulary as shared definitions:

Frame content uses reserved zone names `title`, `eyebrow`, `source`, and `nav`
when the selected frame requires/exposes them. A local template declares these
zones in its value contract, and frame binding renders them into the frame-owned
locations. Required header/source/nav content cannot be omitted by hiding it in a
body node. A no-header frame excludes the corresponding header zones. The pinned
frame schema determines which zones and value types are valid.

| Node kind | Authored fields | Binding behavior |
| --- | --- | --- |
| `text` | Stable ID, placement, shared style, ink/alignment, `text` | Literal string or `{binding: zone_name}`; binding must resolve to text |
| `box` | Stable ID, placement, surface, optional border token | No arbitrary drawing code |
| `image` | Stable ID, placement, `asset`, fit, optional rotation | Literal declared asset ID or asset-valued binding |
| `rule` | Stable ID, placement, ink, positive width in points | Native rule; respects grid and zone validation |
| `group` | Stable ID, ordered child nodes | Paint order is explicit; child placements use frame-zone coordinates |
| `component` / `composite` | Stable ID, placement, pinned definition reference, `arguments` | Arguments validated against that definition's executable typed schema |

Placement names a frame zone (`body`, `rail`, `panel`, `short_body`, `tall_body`,
`whiteboard`, or `source`) and either a grid span plus vertical geometry or an
explicit point rectangle. Grid columns are one-based. Coordinates are relative
to the named zone's origin; the adapter translates to the current foundation
renderer's slide-space points. Units are explicit; no mixing inches, CSS pixels,
and points. A group does not change this coordinate origin.

Node array order specifies back-to-front paint order. Nesting preserves group
ownership. Binding objects may appear recursively in component/composite arguments
but are permitted only where the referenced argument schema declares a bindable
value. There is no expression language, interpolation, network fetch, script,
or arbitrary raw OOXML escape hatch in this contract.

The Wave 1 envelope intentionally does not reproduce every table/chart/diagram
schema. Their persistent definition IDs point to executable typed schemas. The
implemented compiler validates the node kind, style, ink/surface, frame/grid,
arguments, values, and binding paths against the pinned definitions before
generating a scene. Syntactic JSON Schema acceptance alone is insufficient.

Frames contain static chrome and declare content zones. A reference cannot make
content outside an allowed zone legal. Frame and grid compatibility, exact item
counts, supported transforms, measured overflow, and asset restrictions remain
enforced. Changing a layout does not require making the semantic content contract
unbounded; it requires choosing or creating a supported local composition.

## 4. Local customization, instantiate, detach and fork

Selecting a shared template supplies its fixed executable content contract.
Selecting a local template supplies a deck-owned composition and zone contract.
Both compile through the same vocabulary and scene compiler.

| Operation | Proposed semantics |
| --- | --- |
| Instantiate | Bind actual content to a shared definition while retaining the shared reference |
| Detach | Copy a resolved shared template's editable composition/zone contract into `local_templates`; change that slide to a local reference |
| Fork | Create a new local template from a shared or local template and use it for one or more slides |
| Revise local | Edit the local composition; revalidate every slide using it and invalidate affected visual reviews |
| Promote | Separate library process to propose a reusable shared definition; never automatic |

Detach/fork preserves provenance: operation, parent scope/ID/revision, definition
hash, and optional reason. It copies the source structure; it does not copy a
rendered screenshot. Assets retain their original identity. A local architecture
starter can be widened, re-spaced, or given different component groupings while
preserving its ancestry. Unchanged shared components within that local template
remain shared references. No implicit synchronization from later shared revisions.

Local zone contracts must use supported JSON Schema constructs and bindable types.
The runtime rejects unsupported schema features rather than accepting a validation
contract it cannot execute. The root schema permits a JSON Schema object as a zone
schema; supported dialect/subset validation is a separate compiler responsibility.
The initial local-zone subset is `type`, `properties`, `required`,
`additionalProperties: false`, `items`, `minItems`, `maxItems`, `minLength`,
`maxLength`, `minimum`, `maximum`, `enum`, and `const`, plus nonexecutable
`description`. Complex unions use an existing registered typed contract until the
validator explicitly supports them. Remote `$ref`, custom keywords, schema code,
and unimplemented validation constructs fail. This is a bounded local-value
validator, not a second unrestricted JSON Schema runtime.

## 5. Stable identity and provenance

| Identity | Scope | Rule |
| --- | --- | --- |
| Deck ID | Project | Stable across rebuilds and sessions |
| Slide ID | Deck | Unique; reorder does not rename it |
| Template ID | Shared registry or local map | Opaque; scope is explicit |
| Node ID | Template tree | Unique throughout that template, including grouped children |
| Content item key | Declared array binding | Unique within the declared zone; preserves item identity when reordered |
| Asset ID | Deck declaration or pinned registry | Original and derivative identities remain distinct |
| Claim/source ID | Project evidence | Locators and hashes identify the actual supporting version |
| Native object identity | A generated slide | Derived from slide ID, node ID, item key, and emitted part role |

Runtime mapping uses a logical key such as
`slide_id/node_id/item_key/part_role`. Numeric PowerPoint shape IDs, slide numbers,
and ZIP part names are recorded as physical addresses, never durable source keys.
One node may emit several native objects. Shared frame objects map to their shared
definition/field, not to a fake editable slide content field. Native names may
carry stable keys; a sidecar object map remains authoritative when names are
changed, lost, or ambiguous.

Duplicate slide/node/item IDs, missing zone bindings, illegal reference scope,
unused required fields, cycles in definition references, and unresolved assets
are semantic errors. JSON Schema cannot enforce all cross-record constraints.

## 6. Assets, derivatives and paths

`assets` maps authored IDs to either a registered asset ID or a package-relative
custom source path. All actual payload hashes are recorded in the lock/build
manifest. Registry assets resolve from the pinned release; custom assets resolve
from the deck package root. Absolute paths, URLs, `..`, symlink escapes, and missing
payloads fail in portable production builds. Evidence URLs are permissible in
the source index; they are not executable image imports.

Original assets are preserved byte-for-byte. Compression/resizing creates a
derivative with its own ID/hash and records original ID/hash, recipe, dimensions,
format, color/quality settings and transformer version. Deck output may use the
derivative while the project retains the original. Lossless ZIP compression is a
packaging setting and does not modify asset pixels. Existing source-deck artwork
that must be retained remains explicitly disclosed.

The `media_optimization` object maps directly to the repository's
`pptx.MediaOptimizationOptions`: `deduplicate`, `resize_jpeg`, `compression`,
`pixels_per_inch`, `jpeg_quality`, and `min_savings_percent`. If omitted, resolve
`DeliveryMediaOptions()` (true/true/true, 220 PPI, JPEG quality 90, minimum saving
10%). An override supplies every field; do not infer enabled flags from a lone
quality setting. A preservation policy can disable JPEG resizing while retaining
lossless ZIP compression and exact media deduplication. The resolved policy and
`pptx.MediaOptimizationReport` belong in the build receipt/layout report.
The current implementation optimizes JPEG photos, preserves vector and non-JPEG
payloads, and records original/output hashes and dimensions. This API and its authored YAML policy mapping are implemented. Its report does
not itself create an external derivative file. The project exporter retains authored originals and declared derivatives.
Output-embedded optimization derivatives are recorded by the layout/media report;
they do not become separately authored derivative files automatically.

## 7. Project package and disclosure

```text
deck-project/
  deck.yaml
  README.md
  toolchain.lock.json
  context/                 # audience, framing, narrative, internal page briefs
  evidence/                # source index, permitted originals, extracts, claim ledger
  assets/originals/
  assets/derivatives/
  decisions/               # decisions and artifact-bound approvals
  state.json
  reviews/                 # immutable restricted reviewer packets and findings
  builds/<build-id>/        # derived scene, maps, reports, PPTX, optional PDF, receipts
```

Relative references preserve portability. Missing/restricted evidence is explicitly
recorded with origin, locator and access limits. An optional offline runtime export
can accompany the project; the normal package references a pinned installed
release. No full developer sample/evidence archive is required in every client
project. Retention policy identifies current outputs and selected historical
baselines while keeping source evidence and originals independently.

The maintainer package may contain internal strategy and private evidence. A
client-facing PPTX/PDF export and an independent reviewer packet are narrower
deliverables. The skill must not blindly share the whole working project.

## 8. State and invalidation contract

These tables define proposed auxiliary records. They are not current CLI formats.

| State field | Meaning |
| --- | --- |
| `schema` / `project_id` | `pptxgengo.deck-project-state.v1` and deck identity |
| `stage` | intake, framing, outline, content, selection, build, review, delivered |
| `next_action` / `open_items` | Concrete resumption point and unresolved choices |
| `source_snapshot` | Authored YAML byte hash and canonical semantic hash |
| `artifact_versions` | Current context/content/composition/review artifacts and hashes |
| `dependencies` | Source/claim/page/definition relationships used for impact analysis |
| `approvals` | Links to version/hash-bound approval records and their scope |
| `invalidations` | Affected stage/page, trigger, status and required next action |
| `baseline` / `current_build` | Explicit build IDs; not inferred from modification time |

Approval records contain an ID, stage, approver, time, decision, exact artifact
version/hash, affected stable slide IDs or whole-deck scope, and any limits.
Formatting-only YAML changes change its byte hash but not canonical semantic hash;
they do not automatically invalidate narrative approvals. New material evidence,
changed claims, changed argument/order, changed visible copy, changed composition,
and changed runtime have different impacts:

| Change | Minimum invalidation |
| --- | --- |
| Evidence supporting a claim changes | Reverify claim and affected visible qualifications; content review if meaning changes |
| Core argument or required premise/order changes | Outline approval and downstream affected content/composition reviews |
| Visible copy/data changes | Affected content approval, fit result, source verification and rendered QA |
| Geometry, frame, template or visual asset changes | Affected fit/rendered QA; content review only if meaning/visibility changes |
| Font, engine, bundle or compiler migration | Rebuild; affected layout/native review; no invented transfer of prior acceptance |
| Unsupported manual PPTX change | Divergence flag and reconciliation before regeneration/delivery |

An approval can remain valid for unaffected pages. Invalidation does not authorize
discarding prior decisions. Resume reads state, checks changed dependencies, reports
the current stage and useful next action, and continues only the affected work.

## 9. Build receipt and baseline contract

| Receipt field | Meaning |
| --- | --- |
| `schema` / `build_id` / `deck_id` | `pptxgengo.deck-build-receipt.v1` plus stable identities |
| `input` | Authored file hash, canonical document hash and supported schema version |
| `toolchain` | Lockfile hash, runtime/skill versions and manifest hashes, engine/bundle/calibration/fonts |
| `definitions` / `assets` | Exact used IDs, revisions, source/derivative hashes and provenance |
| `outputs` | Scene, PPTX, reports, object map and optional PDF hashes/relative paths |
| `build_epoch` / `receipt_time` | Fixed reproducibility clock versus actual receipt creation time |
| `fit_status` / `native_status` / `visual_status` | Separate results with evidence scope and pending issues |
| `pdf_export` | Exporter/environment and exact PPTX/PDF hash association; absent if no export |
| `baseline` | Optional prior generated build ID/hash used for divergence comparison |

A baseline stores the generated deck and authored canonical source at the same
build, plus resolved scene and object map. A receipt must not infer acceptance
from successful compilation, a catalog hit, or the existence of a PDF. Reviewed
specimens remain evidence for those specimens; actual replacement content needs
its own fit/visual evidence.

| Baseline/object-map field | Meaning |
| --- | --- |
| `baseline_id` / `build_receipt` | Stable baseline identity and receipt path/hash |
| `authored_snapshot` / `canonical_snapshot` | Exact YAML bytes and canonical value tree, each stored with hash |
| `generated_pptx` / `resolved_scene` | Exact baseline native deck and derived scene path/hash |
| `logical_object_key` | Stable slide/node/item/part identity independent of physical ordering |
| `source_location` | Named zone, item key, field role and source pointer in the baseline canonical tree |
| `native_location` | Presentation part, slide identity, object ID/name, group chain and table/chart subaddress |
| `editable_fields` | Bounded text/data/style/geometry fields actually owned by this source mapping |
| `baseline_values` | Canonical authored and corresponding emitted native values used for comparison |
| `provenance` | Shared/local definition and hashes, frame role, original/derivative asset lineage |

An array-index pointer is an address within a snapshot, not persistent identity.
Reconciliation resolves the authored item key before applying an edit to the
current source. Objects emitted from shared chrome or fixed illustration parts
may have no deck-owned content field; the map explicitly records that limitation.

## 10. Manual editing and future reconciliation

Native PowerPoint editing is supported as an operator activity. There is no current
general reverse compiler that turns a changed PPTX into maintained semantic YAML.
The first implementation must at least detect divergence from the last generated
PPTX hash before replacing or claiming source/output alignment.

Future reconciliation compares three inputs: the canonical source and generated
native state at the recorded baseline, the current source, and the manually edited
PPTX. Map native changes through stable object keys; do not infer identity only
from shape index, slide order, bounding box, or visible text. Supported edits become
reviewable proposals against named source fields.

| Source versus baseline | PPTX versus baseline | Proposed handling |
| --- | --- | --- |
| Unchanged | Supported text/data changed | Propose promotion into source, then rebuild/review |
| Changed | Unchanged | Build current source; preserve existing baseline/history |
| Same supported semantic change on both | Changed | Confirm matching intent; rebuild and record synchronization |
| Both differ on same source field | Changed | Explicit conflict; show all three values, obtain resolution |
| Any state | Unsupported geometry/art/topology change | Preserve edited artifact; propose local-template change or retained native variant |

No silent overwrite or automatic general merge. Unmatched/deleted/duplicated objects
are ambiguous, not proof that the author deleted a source field. Table cell/item
mapping needs stable row keys and cell roles; slides added in PowerPoint need an
explicit import/adoption decision. A native retained variant is an explicitly
separate route with editable-native capabilities and limits, not a fake shared
template. Full reconciliation implementation is after the source/identity adapter.

## 11. Implemented runtime boundary

`design project` implements strict YAML decoding, canonicalization, zone/ref
validation, shared/local resolution, original/derived asset validation, stable
object mapping, exact runtime/bundle/font/asset locks, immutable scene/native
builds, hash-bound scoped approvals and invalidation, resumption, local fork,
bounded shared detach, and maintainer/client/offline/reviewer ZIP exports.
Existing JSON APIs remain available. The current source bundle is v5
(`wmds-library.v5`, 587 templates at
`d83bd58a9f9de68ebd8d6b3c9b0272c16ed516cf`); the calibrated Go engine remains
`wmds-go-foundation.v2`.

The loader rejects duplicate keys, nonstring mapping keys, unknown fields,
multiple documents, custom tags, nonfinite numbers, merges and aliases. It
preserves scalar text and block-scalar hard/trailing breaks, with filename,
line/column and JSON pointer diagnostics. Source size is bounded to 16 MiB and
nesting to 100. Canonical JSON preserves array order and meaningful text;
native output uses fixed timestamp `2000-01-01T00:00:00Z` and canonical-source
identity seed. Build IDs and receipt clocks identify execution events separately.
No environment interpolation or executable include syntax is accepted.

Context keys are closed to `project`, `audience`, `outline`, `sources`, `claims`,
`decisions` and `state`. Paths must be safe project-relative files/directories.
A file hashes only its bytes: `sources: sources/index.md` does **not** follow
links to evidence. A directory hashes its full safe tree: use `sources: sources`
(or an evidence directory) when added/changed meeting materials must invalidate
approvals. `state: state.json` refers to generated state and is excluded from
authored dependency hashing; other context keys cannot point to that file.
`evidence_refs` describe claim/source identity but do not fetch or validate linked
external content. Put maintained source indexes and material local evidence under
a tracked context directory, respecting access/disclosure constraints.

Local nodes/zone schemas, shared frames/grid, source-scene components, assets and
mutations are bounded by the runtime README. Fork only rebinds slides already
using the parent local template. Detach retains actual supplied content and
verified ancestry for generic source scenes; legacy typed card/metric IR and
source-note/stamp chrome detach remain explicitly unsupported. Build preserves
raw YAML; fork/detach serialize it and preserve exact predecessor snapshots.

The current reviewer ZIP contains technical build evidence and visible draft
bindings; it does not automatically assemble isolated outline/content/audience
packets or run fresh reviewers. Automatic PDF import/export and arbitrary edited
PowerPoint reconciliation are not implemented. Offline execution requires the
pinned compiler's original OS/architecture and does not install PowerPoint or
system fonts. Technical fit/build success leaves native appearance, visual review
and arbitrary-content qualification separate.

The original planning YAML remains a historical illustration
with unresolved example provenance/lock/assets. Use the executable starter for
current commands and supported fields. Focused runtime/candidate smoke evidence
is synthetic; client-content workflow acceptance remains separate.
