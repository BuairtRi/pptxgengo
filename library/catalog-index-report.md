# Catalog index build report

The rebuildable SQLite catalog was built from occurrence, asset, and dedup manifests plus any supplied decision, geometry, component, and extras manifests. Full records are preserved in the `items.json` column; FTS5 indexes item titles and searchable text. Slide occurrence search bodies include text aggregated from all shape and group descendants on that slide. Geometry is attached to matching slide shape/group occurrence records and adds no item rows. Components and optional layout-family decisions are separate indexed records; decisions do not replace native dedup review statuses. Source readiness/review fields are preserved without inferring approval.

| Input | Path | SHA-256 | Records ingested |
|---|---|---|---:|
| occurrences | `samples/catalog/occurrences.jsonl` | `d0884111f9be47e6772f542e407a76ee7253888d82345f8a5a6c59f5f7273b69` | 13831 |
| assets | `samples/catalog/assets-v2.jsonl` | `5206b880d787cf4877cb91e7b8d81fc9013133c5eb5f92fc42d3824ad9590afe` | 2169 |
| dedup | `samples/catalog/layout-dedup-v4.json` | `b66fe131c4d9683e81203156f96599bf451d0009650df3c2428a265781f1df43` | 771 |
| decisions | `library/layout-decisions.json` | `e36373b3a12258918cc59d09f649b3b6f6c7fd129d679afeb836de0923096763` | 26 |
| geometry | `samples/catalog-next/geometry.jsonl` | `d42dc9b6c2b6a79bc797ad6cb7d03f79e6c4e5eec57961a8f80db06a7f262af6` | 10102 |
| components | `samples/component-expansion/components-v2.jsonl` | `2caf63572b638f635063dd6e7f8b40c6997d38e23a3d552e411658fbbe9eee68` | 559 |
| extras | `samples/component-expansion/extras-v2.jsonl` | `23cdd10070541bcc1a3cf31212eb975efa73541c5ae573ed24e2ca78553eb2af` | 95 |

Total indexed items: **17451**.

| Kind | Items |
|---|---:|
| asset | 2169 |
| component | 559 |
| component_family | 31 |
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
| style_profile | 6 |

Dedup slide references were checked against source slide occurrences; component member and slot references were checked against slide occurrence IDs, source hashes, and slide boundaries. Geometry IDs were checked for unique exact coverage of slide shape/group occurrences. Every indexed ID is unique.
