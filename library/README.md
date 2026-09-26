# Presentation library — expanded component checkpoint

2026-09-25. The catalog retains all **369 source slides**. Reviewed layout
families reduce classification to **306 work units**, avoiding 63 repeated
classifications. This is not the final unique-layout count or an approved
reusable library.

## Available now

| Inventory | Current evidence |
|---|---|
| Source occurrences | 13,831 records across slides, native layouts, masters and their objects; 10,102 slide object nodes include 315 native groups |
| Layout review | 89 source occurrences reviewed into 24 shared families and 2 distinct items; 280 singleton slides still require review |
| Geometry | 9,669 resolved frames, 433 unresolved; 579 transforms recovered from layout/master placeholders |
| Component candidates | 315 native-group occurrences plus 244 selected source examples in 74 source patterns, consolidated into 31 semantic families; 572 named slot candidates |
| Assets | 2,169 image/vector asset records, including 521 existing photo descriptions preserved verbatim |
| Fonts | 58 local font files with family/style/hash metadata; these are IBM Plex variants, not a system-wide font availability audit |
| Search | SQLite FTS5 over 17,451 items (including component families and six proposed style profiles), with source identity and readiness retained |
| Visual evidence | 144 source-indexed native previews with hashes |

The local current database is `samples/component-expansion/catalog-v2.sqlite`.
[Catalog manifest](catalog-manifest.json) records the exact current input and
output paths/hashes. Large generated records, previews and databases stay local.
Only the modernization source deck is tracked under `samples/`.

## Reference selection and first text contracts

A [10-reference visual shortlist](reference-variants.md) now covers metric panels,
numbered cards and delivery pods. The local gallery enlarges each source region,
shows its full-slide context and supports exporting design preferences.
[`pptxcomponent`](../cmd/pptxcomponent/README.md) exposes four source-bound text
contracts. Source styling and geometry are preserved; semantic style application,
measured capacity and adaptation approval remain pending.

## Review and reports

- [Expanded component findings](component-expansion-report.md), [canonical families](component-families.jsonl),
  [semantic styling](component-styling.md), and [taxonomy](component-taxonomy.json).
  Three Luna agents reviewed independent corpus slices; the primary agent reviewed
  selected source pages, corrected boundaries/descriptions and integrated the results.
  A Sol agent reviewed compiler/index code. Of 249 input examples, five reviewed
  duplicates were consolidated; their source references remain as aliases.
  The 244 retained examples are not 244 unique or adaptation-approved designs.


- [Second-pass findings](second-pass-layout-review.md), [layout decisions](layout-decisions.json),
  and [complete worklist](layout-worklist.json).
- [Component candidates](component-report.md), [proposed seed definitions](component-seeds-expanded.json),
  and [local component gallery](../samples/component-expansion/component-review-v2.html).
  The gallery overlays selected frames on unchanged source previews; automated
  browser policy blocked opening local HTML, so its UI is not visually verified.
- [Geometry](geometry-report.md), [source occurrences](occurrence-report.md),
  [asset ingestion](asset-ingestion-report.md), [fonts/illustration gap](brand-extras-report.md),
  [index build](catalog-index-report.md), and [render provenance](render-manifest.json).
- [First structural dedup pass](layout-dedup-report.md) is historical retrieval
  evidence. Current reviewed decisions supersede its pending classification labels.
- [Further metadata candidates](broader-dedup-candidates.md) and the local
  `samples/catalog-next/layout-candidates-after-review.json` are pending review.

The first families to develop into editable components are delivery pods,
numbered cards and metric panels. Their source bounds and named slots are
available; typography, replacement fit, supported resizing, dependencies and
attachment rules still need proof. No current component is adaptation-approved.

## Search the current catalog

Run from the repository root. Python needs SQLite FTS5 support.

```sh
python3 scripts/catalog-index.py search \
  --db samples/component-expansion/catalog-v2.sqlite --kind component_family --query metric --limit 5
python3 scripts/catalog-index.py search \
  --db samples/component-expansion/catalog-v2.sqlite --kind style_profile --query neutral --limit 5
python3 scripts/catalog-index.py search \
  --db samples/component-expansion/catalog-v2.sqlite --kind component --query pod --limit 5
python3 scripts/catalog-index.py search \
  --db samples/component-expansion/catalog-v2.sqlite --kind font --query 'IBM Plex Mono' --limit 5
python3 scripts/catalog-index.py search \
  --db samples/component-expansion/catalog-v2.sqlite --kind layout_family --query bio --limit 5
python3 scripts/catalog-index.py search \
  --db samples/component-expansion/catalog-v2.sqlite --kind asset --query 'glass atrium' --limit 5
python3 scripts/catalog-index.py inspect \
  --db samples/component-expansion/catalog-v2.sqlite \
  --id component:delivery-pod-three-roles:instance-001
```

`inspect` returns complete source records, geometry, slots and constraints.
Geometry enriches existing object rows; it does not inflate the item count.
Raw candidate queue status is retrieval history; consult reviewed layout-family
records and the worklist for current decisions. Queries support FTS operators.

## Rebuild inspection artifacts

These are Python standard-library commands alongside the Go reconstruction tools.
They are not yet integrated into a single product CLI or distributed agent skill.
Registry paths must resolve to the exact source binaries and structural inventories.
Scripts refuse existing outputs; the following uses a fresh run directory.

```sh
python3 scripts/catalog-occurrences.py --registry planning/source-registry.json \
  --out samples/catalog-run/occurrences.jsonl --report samples/catalog-run/occurrences.md
python3 scripts/deduplicate-layouts.py --registry planning/source-registry.json \
  --out samples/catalog-run/dedup.json --report samples/catalog-run/dedup.md
python3 scripts/catalog-geometry.py --registry planning/source-registry.json \
  --out samples/catalog-run/geometry.jsonl --report samples/catalog-run/geometry.md
python3 scripts/compile-component-library.py \
  --inputs library/component-seeds.json library/component-slices/modernization.json \
  library/component-slices/stock-001-083.json library/component-slices/stock-084-166.json \
  library/component-slices/proposals.json --taxonomy library/component-taxonomy.json \
  --out samples/catalog-run/seeds.json --families samples/catalog-run/families.jsonl \
  --report samples/catalog-run/component-expansion.md
python3 scripts/catalog-components.py --registry planning/source-registry.json \
  --occurrences samples/catalog-run/occurrences.jsonl --geometry samples/catalog-run/geometry.jsonl \
  --seeds samples/catalog-run/seeds.json --out samples/catalog-run/components.jsonl \
  --report samples/catalog-run/components.md
python3 scripts/catalog-assets.py ingest --brand-root /Users/rscott/Documents/branding \
  --hosted-inventory /Users/rscott/.codex/skills/wm-brand-assets/references/asset-inventory.json \
  --out samples/catalog-run/assets.jsonl --report samples/catalog-run/assets.md
python3 scripts/catalog-brand-extras.py ingest --brand-root /Users/rscott/Documents/branding \
  --hosted-inventory /Users/rscott/.codex/skills/wm-brand-assets/references/asset-inventory.json \
  --out samples/catalog-run/brand-extras.jsonl --report samples/catalog-run/extras.md
cat samples/catalog-run/brand-extras.jsonl samples/catalog-run/families.jsonl > samples/catalog-run/extras.jsonl
python3 scripts/catalog-index.py build --occurrences samples/catalog-run/occurrences.jsonl \
  --assets samples/catalog-run/assets.jsonl --dedup samples/catalog-run/dedup.json \
  --decisions library/layout-decisions.json --geometry samples/catalog-run/geometry.jsonl \
  --components samples/catalog-run/components.jsonl --extras samples/catalog-run/extras.jsonl \
  --out samples/catalog-run/catalog.sqlite --report samples/catalog-run/index.md
python3 scripts/catalog-worklist.py --occurrences samples/catalog-run/occurrences.jsonl \
  --decisions library/layout-decisions.json --out samples/catalog-run/worklist.json
python3 scripts/catalog-layout-candidates.py --geometry samples/catalog-run/geometry.jsonl \
  --decisions library/layout-decisions.json --out samples/catalog-run/candidates.json \
  --report samples/catalog-run/candidates.md
python3 scripts/catalog-component-review.py --components samples/catalog-run/components.jsonl \
  --geometry samples/catalog-run/geometry.jsonl --previews library/render-manifest.json \
  --out samples/catalog-run/component-review.html
```

The database is disposable. JSON/JSONL, source hashes and reviewed decisions are
the durable record. Changed source bytes require new ingestion and renewed review.
No unit test suite was run in this wave; actual corpus ingestion, native preview
review, reference validation and catalog queries provide the current evidence.

## Preview preparation

PowerPoint PDF export omits hidden slides. Prepare a task copy with
`scripts/prepare-catalog-render.py --source SOURCE --out NEW_COPY`, then use the
native `scripts/export-powerpoint.applescript` and `scripts/render-pdf.swift`
adapters with absolute paths. Verify the PDF page count against the emitted
source map. Footer numbers can differ from source indices. Originals remain
unchanged. Current previews are pinned in `render-manifest.json`.

## Remaining gates

1. Review the remaining 17 geometry pairs and 10 metadata leads; classify and rate
   one representative per family while retaining style/content variants.
2. Resolve inherited fonts, paragraph styles, text insets and measured fit for
   the selected components. Observed character counts are not capacity limits.
3. Connect source dependencies and define anchors, allowed resizing/reflow and
   intentional overlaps; then exercise editable replacements and native QA.
4. Complete illustration coverage and guidance ingestion. The configured local
   illustration folder is empty; this catalog does not establish that no WM
   illustrations exist elsewhere. Hosted binaries have not been downloaded or
   joined to local assets by guessed filenames.
5. Integrate the catalog with authoring commands and the skill pack, then prove
   colleague-edit reconciliation on a small set of supported designs.
