# Component families: consolidation proposal

Propose **21 building-block families**, preserving the **31 historical retrieval labels** and source designs. Larger arrangements remain composition recipes. These boundaries are inferred from curated source purpose, slots and current authoring behavior; they do not establish geometric interchangeability or portable implementation readiness.

The 244 historical compositions and 132 newer occurrences overlap and use different boundaries. Do not add their counts. A source group can contain several families; table counts below are memberships.

| Family | Content zones | Variant / consolidation boundary | Requested memberships | Examples |
|---|---|---|---:|---|
| Heading block | Eyebrow, title, optional section number | Hierarchy, title width and line count; divider pages are recipes | 20 | t077-uhg-005, t078-uhg-007 |
| Text/list/callout panel | Heading, paragraphs or ordered bullet groups | Border, surface, emphasis and density; retain paragraph/list structure | 53 | t077-uhg-005, t078-uhg-007 |
| Icon with content | Icon, label and supporting text | Icon position and optical size; keep icon-to-copy association | 3 | t077-uhg-005, t078-uhg-007 |
| Numbered item/card | Number, heading and description | Horizontal/vertical arrangements and badge placement; keep numbering semantic | 4 | t005-uhg-012, t091-uhg-033 |
| Metric display | Value, unit, label and optional source | Single/multiple metric arrangement; preserve value/label/source association | 5 | t079-uhg-008, t069-ai-accelerator-006 |
| Quotation | Quotation, attribution and optional portrait | Quote length and attribution position; source claims remain explicit | 2 | t079-uhg-008, t069-ai-accelerator-006 |
| Person profile | Portrait, name, role and optional biography | Compact identity versus detailed biography variants | 9 | t100-uhg-044, t068-ai-accelerator-004 |
| Role tile | Role label, owner and optional allocation | Staffing/ownership colors and explicit legend; distinct from a person biography | 2 | t028-uhg-043, t070-ai-accelerator-013 |
| Pod container | Pod title and ordered role tiles | Role count, columns and spacing; parent owns inter-pod relationships | 1 | t028-uhg-043 |
| Aligned table/evidence row | Named columns, row heading and cell content | Column widths, row count and row/cell styling; headers remain aligned | 22 | t078-uhg-007, t086-uhg-024 |
| Process step | Step label, activities, outputs and connection ports | Step count/order belong to parent flow; local slot hierarchy stays explicit | 21 | t085-uhg-022, t090-uhg-032 |
| Architecture node/tile | Capability/product/system label and optional detail | Node role, width and ports; do not infer arbitrary graph topology | 6 | t010-uhg-015, t051-uhg-025 |
| Architecture layer | Layer label plus ordered capability tiles | Label and tiles share row geometry; repeatable layer owns its segmentation | 3 | t020-graphics-and-layouts-062, t010-uhg-015 |
| Cross-cutting rail | Shared-control labels and declared scope | Separate from layer labels; divider and all-layer scope must remain clear | 3 | t060-uhg-026, t010-uhg-015 |
| Time axis | Ordered period labels and ticks | Period count/width; shared axis coordinates used by all lanes | 3 | t015-graphics-and-layouts-083, t051-uhg-025 |
| Gantt workstream lane | Workstream label, activity intervals and status | Label/bar vertical alignment, period endpoints and explicit status colors | 1 | t015-graphics-and-layouts-083, t039-uhg-038 |
| Milestone/annotation | Marker, date/period and caption | Anchor and caption placement; belongs to a timeline or event sequence | 1 | t039-uhg-038 |
| Gauge or discrete rating | Selected cell/range, pointer and scale/legend | Renderer geometry stays locked; do not substitute progress bars for source gauges | 0 | t045-graphics-and-layouts-045 |
| Legend/key item | Swatch/mark and meaning label | Actor/status/series meaning stays stable across presentation profiles | 1 | t028-uhg-043 |
| Media frame | Image/artifact plus optional caption | Separate photo, portrait, icon, logo and thumbnail aspect/crop policies | 7 | t083-uhg-018, t086-uhg-024 |
| Anchored design accent | Target phrase or source/target objects and approved artwork | Underline, marker and arrow have different placement rules; no standalone slide template | 4 | t095-uhg-037, t096-uhg-039 |

## Recipes and boundaries

architecture-comparison, architecture-map, artifact-gallery, case-study, comparison, evidence-narrative, icon-content-panel, organization-chart, phase-detail, process-flow, radial-benefits, scope-options, section-divider, table-matrix, team-roster, timeline, workshop.

Keep person profiles, role tiles and pod containers distinct. Keep architecture tiles, layer rows and cross-cutting rails distinct. A takeaway band is a text panel with emphasis styling; it is not a hand-drawn accent. A governance callout is not a gauge. Gauges retain their selected renderer and separate selection/pointer semantics.

Dividers use a heading and a slide-level background arrangement. Comparison/pricing/case-study/phase-detail pages assemble shared components while retaining their narrative and topology. A content-panel classification alone does not prove that a complex source group can be transplanted or resized.

## Exact crosswalk

Every historical family and all 132 requested occurrences are addressable in [families.json](families.json), with source/template IDs, object IDs, slot names, family memberships, composition granularity and executable/reference status. Source aliases should survive any eventual consolidation.

## What consolidation changes

Unify content schemas, semantic style roles, sizing metadata and compatible layout behavior. Retain separate source geometry variants when font hierarchy, slot structure, reading order, dependency topology or asset ownership differs. Color differences become style options only when those structural conditions match. No library contracts are migrated by this survey.
