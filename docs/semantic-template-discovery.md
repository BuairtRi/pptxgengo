# Semantic template discovery

An agent should identify the original slide's purpose and argument before
searching: the decision it supports, comparison or relationship it shows,
material facts, and which details can move to notes. Counts help describe that
structure; they do not justify splitting a coherent argument or filling unused
slots with unsupported claims.

## Agent workflow

1. Search purpose and content structure separately through `library-find`.
   Try several precise descriptions rather than only the original title.
2. Inspect two or three suitable candidates with `library-inspect --summary`.
   Open their verified `screenshot_paths` individually and inspect the
   zone-to-binding map, capacities, fixed counts and required data.
3. Select a layout that expresses the original argument. Use actual shared
   bindings when appropriate. When topology differs, derive a local definition
   with an explicit reason and actual shared ancestry. Do not invent metrics,
   phases, connections or claims to satisfy a contract.
4. Map and reshape the real content into those zones. Keep source detail and
   editorial annotations in notes/provenance. Record the candidates considered,
   choice, binding map and meaningful content tradeoffs per original slide.
5. Build actual-content alternatives when consequential. Review measurements
   and every native PowerPoint page; repair both geometry and meaning.

Discovery offers two ranking methods in the next source build:

- `--retrieval metadata` is the existing default: scenario and structural hints
  contribute separate deterministic metadata scores.
- `--retrieval keyword` uses SQLite FTS5 BM25 with English stemming. It indexes
  names, purposes, canonical identities, component terminology, relationships,
  content groups and authoring aliases/descriptions. It excludes synthetic
  example copy. Exact eligible IDs or unique canonical keys rank first; other
  matches use ascending BM25 with field weights name/purpose/discovery 5/2/1.

This is a lexical baseline. Local model-backed semantic and hybrid ranking are
not implemented in this build; requesting them fails explicitly. Existing stable
`v4.1.0` binaries do not yet have these new flags.

Build a **new** index with the matching bundle and optional gallery; existing
indexes remain usable for metadata discovery. Keyword mode rejects old indexes
with an actionable rebuild message. No installed immutable package is modified:

```sh
pptxdesign library-index --bundle library/wm-design-system/v11 \
  --gallery library/wm-design-system/v11/catalog --out /tmp/library-keyword.sqlite
pptxdesign library-find --index /tmp/library-keyword.sqlite --retrieval keyword \
  --query 'modernization economics' --kinds template --summary
pptxdesign library-find --index /tmp/library-keyword.sqlite --retrieval keyword \
  --query 'pillars' --kinds template --roles point --items 3 --item-role point \
  --require-shape --summary
```

Kinds, namespace, lifecycle (`--lifecycles`) and declared adapter capability
(`--content-adapter`) are explicit filters. Deprecated entries additionally need
`--include-deprecated`. Roles, structures, visual forms and exact source group
counts remain hints unless `--require-shape` is supplied, which requires **all**
the supplied hints. Count evidence retains its source scope; an exact nested
count does not establish primary-page capacity. A count/structure mismatch is
visible even for a highly relevant keyword result.

The result's `retrieval` report states the actual ranking method and text recipe;
each hit has a separate raw BM25/rank and `structural_status`. The existing
`score` remains metadata/shape scoring; it is never added to BM25. Empty text
queries explicitly report metadata browsing. `--include-weak` may add unmatched
rows after lexical matches. Neither retrieval nor source agreement measures fit
for supplied copy or native acceptance. Use `library-fit` and native review.

Input is bounded literal natural language, not the FTS expression language.
Source/entity and retrieval-text hashes detect stale projections, while FTS
schema and corpus checks detect altered rows. Rebuild instead of silently using
stale metadata. Asset registry `--kinds asset --summary` keeps its separate metadata
path and rejects the keyword flag; omit `--summary` for indexed asset search.

## Screenshots and SQLite

The accepted 616 gallery already has a native screenshot for every template.
The retained catalog is `library/wm-design-system/v7/catalog`; its unified index
is `library/wm-design-system/v7/library.sqlite`. Run from the repository root:

```sh
pptxdesign library-find --index library/wm-design-system/v7/library.sqlite \
  --query 'phased delivery roadmap' --kinds template --summary
pptxdesign library-inspect --index library/wm-design-system/v7/library.sqlite \
  --id lifecycle/three-phases --summary
```

Build the index with its matching `--gallery` to link these images and preserve
their SHA256 identities. The compact CLI search returns verified absolute
screenshot paths; selection cards relate editable slots to source content zones.
Source zone coordinates are points and describe the original layout. They do
not measure fit for new content.

New indexes expose read-only SQLite views derived from their pinned entity JSON:

```sql
SELECT e.key, e.purpose, a.path, a.sha256
FROM entities e JOIN artifacts a ON a.entity_id=e.id
WHERE e.kind='template' AND a.role='source_preview';

SELECT pointer, component_type, role, bounds_json
FROM content_zones WHERE entity_id='wmds/template/decision/buy-build-economics';

SELECT name, pointer, kind, allow_empty
FROM content_slots WHERE entity_id='wmds/template/decision/buy-build-economics';
```

Artifact paths are relative to the gallery root, recorded in `meta` report
options. Use `library-preview` or compact discovery for resolved paths and hash
verification. Explicit gallery relocation overrides retain those checks.
Existing indexes remain readable; rebuild a new index to obtain these views.

## Reusable Go operations

`project scaffold`, `project edit` and `project measure` replace recurring
source-layout cloning, stable-ID content replacement and geometry diagnostics.
They belong in the CLI. Slide purpose, narrative rewriting and candidate choice
belong to the agent workflow. Native PowerPoint rendering remains the visual
review step. No per-deck Python conversion program is required.

The presentation skill is maintained by another agent. This document records
the implemented discovery/CLI interface and evidence for that agent to adopt;
it does not replace or modify the skill.

## Photography discovery

The production asset index includes all **521** JPEG originals in
`~/Documents/branding/West Monroe Photos/`: 115 Abstract, 337 Industry and
69 Provocative. The 20 previously registered photo IDs are retained. There are
1,204 registered asset variants overall, including 523 photo variants (two
other photo variants are outside this folder).

Photo descriptions come from the existing matching Markdown sidecars. Each
record preserves the full sidecar text and SHA256, title, alt text, keywords,
topics, people description, location, style and placement guidance. Dimensions
and original hashes are measured during registration. This is metadata based
search, not a new human visual review or a vector embedding search.

```sh
pptxgengo library-find --kinds asset --asset-kind photo --summary \
  --query "healthcare clinicians" --limit 5
pptxgengo library-inspect --id photo/library/0f49a52d4f5ee542bd47476e
pptxgengo library-preview --id photo/library/0f49a52d4f5ee542bd47476e
```

Search uses pinned metadata without rehashing the entire 4 GB collection.
Photo search results explicitly report
`registered_original_not_verified_in_query`; preview and deck use verify the
selected original. The visual asset gallery contains derived thumbnails from
verified originals at `library/wm-design-system/v7/catalog/assets/index.html`.
Inspect the photo before choosing a crop or deciding where slide copy can sit.

The SQLite `entities` table stores each photo as `kind='asset'`, `family='photo'`.
Its `source_file` is a relative branding path in the entity JSON; `definition`
contains the registered asset facts and `source_metadata`. Both descriptions and
keywords contribute to search. Asset and sidecar changes invalidate the maintained
index fingerprint, requiring a new snapshot and index.

For library maintenance, use the reusable Go command and rebuild the binary
with the resulting snapshot; originals and sidecars are read only:

```sh
./pptxdesign photo-register --branding-root "$HOME/Documents/branding" \
  --out /tmp/wm-photos-new.json
# Replace internal/wmdesign/photo_registry.json with the reviewed snapshot,
# rebuild, then regenerate the production SQLite index and asset gallery.
```

## Lexical baseline measurements (2026-10-06)

On this Apple M5 Max Mac, Go 1.27.1, the newly built V11 index was approximately
72.1 MB and contained counts read from its report: 649 templates plus indexed
assets/components/frames. A three-iteration microbenchmark measured about
113 ms per warm keyword query (modernization economics, templates only) and
16.2 MB allocated per query. The full verification/open path measured about
511 ms and 608 MB cumulative allocations per open. These are cached-filesystem
microbenchmarks, not uncached startup, peak resident memory, Windows measurements
or performance targets. Full entity/source verification remains at open.

The finder now reads compact discovery projections rather than full source
scene definitions. Before that change the same warm sample took about 214 ms
and allocated 201 MB. Inspection and fit retain the original source definitions.
Engineering judgments cover exact cards/3, interview lists, practices heat maps,
modernization economics, roadmaps and pillars, with an acceptable candidate in
its first ten results. These judgments derive from source purposes and still
need operator and supplied-content/native review.

Reproduce the bounded measurements with:

```sh
go test -run '^$' -bench '^BenchmarkKeyword' -benchtime=3x -benchmem \
  -timeout=3m ./internal/wmdesign
```
