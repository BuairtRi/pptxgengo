# Source occurrence catalog

Generated from the registered source binaries and raw structural inventories. Every binary SHA-256 was checked against both the registry and its raw inventory before export. Occurrences preserve source identity and local hierarchy; they do not assert semantic deduplication or design approval.

| Source | Slides (hidden) | Slide object nodes / groups | Native layouts (object nodes / groups) | Masters (object nodes / groups) | Emitted occurrence records | SHA-256 prefix |
|---|---:|---:|---:|---:|---:|---|
| graphics-and-layouts | 166 (0) | 5467 / 203 | 33 (260 / 0) | 1 (4 / 0) | 5931 | `9180d0d74750` |
| uhg | 85 (0) | 1419 / 54 | 185 (1891 / 80) | 7 (28 / 0) | 3615 | `b0f254ed7739` |
| enablecomp | 35 (12) | 712 / 33 | 23 (216 / 0) | 1 (4 / 0) | 991 | `22c39b27bd97` |
| software-modernization | 83 (2) | 2504 / 25 | 72 (625 / 6) | 2 (8 / 0) | 3294 | `c732e693a637` |

Total emitted occurrence records: **13831**. Output JSONL: `samples/catalog/occurrences.jsonl`.

## Counting and identity definitions

- A slide occurrence is one record for each slide part referenced by the presentation, including hidden slides. Hidden status comes from the slide part root `p:sld/@show`, after resolving presentation relationships; absent `show` is visible.
- A shape/group occurrence is one record per inventoried object node in each part. The node set is `sp`, `pic`, `graphicFrame`, `cxnSp`, `grpSp`, and `contentPart`, recursively including descendants inside groups. Group containers receive their own record.
- Native layout and master occurrence records are emitted once per inventoried native layout/master part, along with one record per object node in that part. Counts are package-part counts and are not deduplicated across inheritance.
- Object paths are one-based sibling positions through the inventoried hierarchy, independent of source shape IDs. Source IDs and duplicate-within-part flags are retained as evidence; IDs are not assumed globally unique or semantically stable.
- `observed_characters` and extracted text are source observations, not text-capacity claims. Explicit font face/size/body properties are retained as found in the object. Inherited style resolution and group-transform flattening remain unresolved. Roles are `unclassified`; readiness is `structural`.
