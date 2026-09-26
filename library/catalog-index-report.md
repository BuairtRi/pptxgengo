# Catalog index build report

The rebuildable SQLite catalog was built from occurrence, asset, and dedup manifests plus any supplied decision, geometry, component, and extras manifests. Full records are preserved in the `items.json` column; FTS5 indexes item titles and searchable text. Slide occurrence search bodies include text aggregated from all shape and group descendants on that slide. Geometry is attached to matching slide shape/group occurrence records and adds no item rows. Components and optional layout-family decisions are separate indexed records; decisions do not replace native dedup review statuses. Source readiness/review fields are preserved without inferring approval.

| Input | Path | SHA-256 | Records ingested |
|---|---|---|---:|
| occurrences | `samples/catalog/occurrences.jsonl` | `d0884111f9be47e6772f542e407a76ee7253888d82345f8a5a6c59f5f7273b69` | 13831 |
| assets | `samples/catalog/assets-v2.jsonl` | `5206b880d787cf4877cb91e7b8d81fc9013133c5eb5f92fc42d3824ad9590afe` | 2169 |
| dedup | `samples/catalog/layout-dedup-v4.json` | `b66fe131c4d9683e81203156f96599bf451d0009650df3c2428a265781f1df43` | 771 |
| decisions | `library/layout-decisions.json` | `e36373b3a12258918cc59d09f649b3b6f6c7fd129d679afeb836de0923096763` | 26 |
| geometry | `samples/catalog-next/geometry.jsonl` | `d42dc9b6c2b6a79bc797ad6cb7d03f79e6c4e5eec57961a8f80db06a7f262af6` | 10102 |
| components | `samples/catalog-next/components-v2.jsonl` | `299f15c7ed2f3423106664a7e024413f58b19c11b78a0beb355df4d64d3d2358` | 336 |
| extras | `samples/catalog-next/brand-extras-v2.jsonl` | `90609a05e9db622c71d2396b0ca0df055f8c04488f773273f0f63b59316ccd08` | 58 |

Total indexed items: **17191**.

| Kind | Items |
|---|---:|
| asset | 2169 |
| component | 336 |
| dedup_slide | 369 |
| font | 58 |
| group | 401 |
| layout_family | 26 |
| master | 11 |
| native_group | 369 |
| native_layout | 313 |
| review_queue | 33 |
| shape | 12737 |
| slide | 369 |

Dedup slide references were checked against source slide occurrences; component member and slot references were checked against slide occurrence IDs, source hashes, and slide boundaries. Geometry IDs were checked for unique exact coverage of slide shape/group occurrences. Every indexed ID is unique.
