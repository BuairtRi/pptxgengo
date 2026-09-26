# Slide geometry catalog

Generated from source-hash-verified PPTX originals in `planning/source-registry.json`; source files were read only. One JSONL record represents one slide object node, including native group containers. Occurrence IDs match `catalog-occurrences.py`.

| Source | Objects | Resolved frames | Unresolved frames | Groups | Inherited transforms | Groups without child map |
|---|---:|---:|---:|---:|---:|---:|
| graphics-and-layouts | 5467 | 5294 | 173 | 203 | 313 | 0 |
| uhg | 1419 | 1254 | 165 | 54 | 184 | 0 |
| enablecomp | 712 | 679 | 33 | 33 | 39 | 0 |
| software-modernization | 2504 | 2442 | 62 | 25 | 43 | 0 |
| **Total** | **10102** | **9669** | **433** | **315** | **579** | **0** |

## Meaning and limits

- `polygon_emu` gives the four transformed corners of the object frame, ordered top-left, top-right, bottom-right, bottom-left in local coordinates. `bounds_emu` is its axis-aligned enclosing box. Inch values divide EMU by 914400. Rotated frames and fills are not treated as exact painted regions.
- Group child frames compose `off/ext/chOff/chExt`, horizontal and vertical flips, and rotation through every ancestor. Descendant IDs and ancestry retain native grouping.
- A missing slide transform can inherit a unique matching placeholder transform from its linked layout, then master. Ambiguous or incomplete transforms remain null. Provenance records the source of each effective transform.
- Text and selected explicit text properties are observations. Typography inherited from layouts, masters, and themes remains unresolved. Shapes may also paint outside their rectangular frame; no render or clipping calculation was performed.
