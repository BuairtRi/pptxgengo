# Catalog index build report

The rebuildable SQLite catalog was built from the occurrence, asset, and dedup JSON/JSONL manifests. Full records are preserved in the `items.json` column; FTS5 indexes item titles and searchable text. Slide occurrence search bodies include text aggregated from all shape and group descendants on that slide. This inspection index carries source readiness/review fields as supplied and does not infer approval. Optional layout-family decisions are stored as separate records and do not replace native dedup review statuses.

| Input | Path | SHA-256 | Records ingested |
|---|---|---|---:|
| occurrences | `samples/catalog/occurrences.jsonl` | `d0884111f9be47e6772f542e407a76ee7253888d82345f8a5a6c59f5f7273b69` | 13831 |
| assets | `samples/catalog/assets-v2.jsonl` | `5206b880d787cf4877cb91e7b8d81fc9013133c5eb5f92fc42d3824ad9590afe` | 2169 |
| dedup | `samples/catalog/layout-dedup-v4.json` | `b66fe131c4d9683e81203156f96599bf451d0009650df3c2428a265781f1df43` | 771 |
| decisions | `library/layout-decisions.json` | `59a416a80eab0b7180e5b333c1fb7234b185f8b7248fab4f3896125dc93a4e70` | 18 |

Total indexed items: **16789**.

| Kind | Items |
|---|---:|
| asset | 2169 |
| dedup_slide | 369 |
| group | 401 |
| layout_family | 18 |
| master | 11 |
| native_group | 369 |
| native_layout | 313 |
| review_queue | 33 |
| shape | 12737 |
| slide | 369 |

Dedup slide references were checked against source slide occurrences using `(source_id, slide_number)` mapping; all native-group members/representatives and both review-queue endpoints must resolve. Every indexed ID is unique.

## Execution evidence

- Final database: `samples/catalog/catalog.sqlite` (local, ignored).
- Slide search includes descendant copy: `COBOL` retrieved modernization source
  slides whose titles do not contain the term.
- Asset search `glass atrium` retrieved preserved photo sidecar descriptions.
- Layout-family search `timeline` retrieved both stock timeline arrangements
  with classification-only readiness displayed.
- The 18 decision records comprise 17 shared families and one distinct blank
  reference item. No approved semantic template is implied by this record kind.
- No Go or Python unit test suite was run during this catalog wave.
