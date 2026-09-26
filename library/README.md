# Presentation library — ingestion checkpoint

2026-09-25. This first implementation wave provides a searchable inspection catalog
and a reviewed layout classification worklist. The first pass reduces 369 source
slides to 336 work units, including 17 shared visual families. This is not yet the
final count of unique layouts. It does not yet provide approved
reusable components, content-zone contracts, or semantic slide generation.

## What is available

- [Occurrence inventory](occurrence-report.md): all 369 slides, including hidden
  slides, and every inventoried object in slide, native-layout and master parts.
  The 13,831 records preserve hierarchy, source identity, explicit properties and
  observed text. The 10,102 slide object nodes include 315 native groups; these
  are occurrence counts, not unique reusable shape counts.
- [Layout deduplication](layout-dedup-report.md): conservative native fingerprints,
  geometry candidates and explicit visual decisions before semantic enrichment.
  [Decisions](layout-decisions.json) retain variants and source references.
  [Worklist](layout-worklist.json) accounts for every source slide.
- [Asset ingestion](asset-ingestion-report.md): 2,169 source-scoped/local-hash asset
  records, with 521 existing photo sidecars preserved verbatim. Hosted binaries
  have not been downloaded or joined to local files by filename guesswork.
- [SQLite index](catalog-index-report.md): full-text search over slide text,
  objects, assets/descriptions and reviewed layout families. The local database
  is `samples/catalog/catalog.sqlite`.
- [Render provenance](render-manifest.json): 51 source-indexed previews and their
  hashes. The local [side-by-side gallery](../samples/catalog/layout-review.html)
  shows the candidate pairs; previews and source binaries remain local.

These are Python standard-library inspection commands alongside the existing Go
reconstruction tools. Integration into the product CLI and agent skill remains
part of the implementation plan. No unit test suite was run in this wave; actual
corpus ingestion, native preview review and catalog queries are the evidence.

## Run and search

Run from the repository root. Registry paths must resolve to the four exact source
binaries and their structural inventories. See [source policy](../planning/README.md).
The scripts refuse to overwrite outputs; choose a new directory for another run.
Python needs SQLite FTS5 support.

```sh
python3 scripts/catalog-occurrences.py \
  --registry planning/source-registry.json \
  --out samples/catalog-next/occurrences.jsonl \
  --report samples/catalog-next/occurrences.md

python3 scripts/deduplicate-layouts.py \
  --registry planning/source-registry.json \
  --out samples/catalog-next/layout-dedup.json \
  --report samples/catalog-next/layout-dedup.md

python3 scripts/catalog-assets.py ingest \
  --brand-root /Users/rscott/Documents/branding \
  --hosted-inventory /Users/rscott/.codex/skills/wm-brand-assets/references/asset-inventory.json \
  --out samples/catalog-next/assets.jsonl \
  --report samples/catalog-next/assets.md

python3 scripts/catalog-index.py build \
  --occurrences samples/catalog-next/occurrences.jsonl \
  --assets samples/catalog-next/assets.jsonl \
  --dedup samples/catalog-next/layout-dedup.json \
  --decisions library/layout-decisions.json \
  --out samples/catalog-next/index.sqlite \
  --report samples/catalog-next/index.md

python3 scripts/catalog-index.py search \
  --db samples/catalog/catalog.sqlite --kind slide --query modernization --limit 5
python3 scripts/catalog-index.py search \
  --db samples/catalog/catalog.sqlite --kind asset --query 'glass atrium' --limit 5
python3 scripts/catalog-index.py search \
  --db samples/catalog/catalog.sqlite --kind layout_family --query timeline --limit 5
python3 scripts/catalog-index.py inspect \
  --db samples/catalog/catalog.sqlite --id layout:stock-standard-table
```

Search returns source/review status; `inspect` returns the complete source record.
FTS query syntax supports phrases and operators. All limits are bounded to 1–100.
The index is disposable; JSON/JSONL and reviewed decisions are the source records.
A source change requires reingestion and renewed review, not reuse of stale decisions.

## Preview preparation

PowerPoint PDF export omits hidden slides. Prepare a task copy before export so
source slide indices and PDF pages remain aligned:

```sh
python3 scripts/prepare-catalog-render.py \
  --source 'samples/software modernization campaign pick deck v1 - Repaired.pptx' \
  --out samples/catalog-next/modernization-render.pptx
```

Use the native `scripts/export-powerpoint.applescript` and `scripts/render-pdf.swift`
adapters with absolute paths. Verify the PDF page count against the emitted render
map before labeling previews by source slide index. The displayed footer number
may differ from the source index. Source originals are never unhidden or edited.

```sh
python3 scripts/catalog-review.py \
  --dedup samples/catalog/layout-dedup-v4.json \
  --previews library/render-manifest.json \
  --decisions library/layout-decisions.json \
  --out samples/catalog-next/layout-review.html
```

## Readiness boundaries and next work

1. Broaden candidate retrieval for layouts encoded with different object counts,
   inherited geometry or group transforms. Current families reduce classification
   work; they do not establish the final unique-layout count.
2. Classify one representative per reviewed family, preserving each variant.
   Rate design quality separately from technical readiness and content approval.
3. Extract meaningful components from native groups and ungrouped compositions;
   define named slots, dependencies, local origins, anchor points and fit rules.
4. Resolve missing asset taxonomy/coverage: this scan found zero path-classified
   illustrations. Inspect unclassified files and other approved asset sources;
   do not report the illustration library as complete. Fonts are also outside
   this image scan.
5. Prove slot edits, supported reflow and colleague-edit reconciliation on a small
   set of useful layouts before expanding the approved reusable library.

Native-content hashes are conservative and may miss equivalent slides. Geometry
matches do not prove shared content roles. Character observations are not capacity
limits. Current families have `classification_only` reuse scope and remain unrated
for design preference. Existing source styles and all source occurrences survive.
