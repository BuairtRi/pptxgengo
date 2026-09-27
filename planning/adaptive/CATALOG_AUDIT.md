# Adaptive catalog and brand asset audit

## Sources and coverage

The capability compiler reads:

- `library/templates/rollout/assignments.json` for the complete 65-item shortlist, source category, lane, representative slide, and source-scene path.
- Each item's `contract.json`, `example-values.json`, and `implementation.json` for source-hash-bound edit slots, binding order/cardinality, RGB/theme color roles and profiles, geometry/retained-content disclosures, and illustrative content.
- Each registered source scene for embedded image relationship targets, picture object names, and real `descr` alternate text when present.
- `library/templates/rollout/checkpoint.json` for current example-review method/status and top-level limitations.

The generator does not interpret a text label or image name as proof that an imported object is a semantic component. It preserves unknown source imagery as linked retained content. Review manifests and individual PNGs remain separate evidence artifacts; their success applies only to those exact examples.

## Semantic family taxonomy hints

The current category-only routing maps these taxonomy groups:

| Proposed family | Source categories | Count |
|---|---|---:|
| Roadmap | `roadmap` | 5 |
| Architecture | `business_architecture`, `technical_architecture`, `layers_components` | 14 |
| Process | `approach_delivery` | 6 |
| Team | `team_credentials` | 5 |
| Comparison | `comparison_evidence` | 6 |
| Unmapped | All other categories | 29 |

These counts are discovery hints only; the 36 category candidates are not structural adaptation recommendations. For example, T007's radial taxonomy chart is classified `business_architecture`, T008's paired radar charts are `comparison_evidence`, and T059's staffing curve is `team_credentials`, but the current architecture, comparison, and team builders do not implement those structures. A source must be inspected and structurally reviewed before selecting a builder. The five reference patterns are illustrative and do not imply that all taxonomy matches—or all 65 source templates—are adaptable. Product, outcomes/commercial, narrative, visual storytelling, and framing/navigation items remain unmapped by these taxonomy hints; that does not rule out a future redesign.

## Existing asset description sources

### Curated local assets

`library/showcase/assets.json` contains selected showcase assets with local path, SHA-256, dimensions, source path, role, use restrictions, description/source description, source sidecar and source-provided keywords. The compiler links to this record set and retains `retained_source_content` from each template. It does not copy a showcase description onto an unrelated image by visual or filename similarity.

### Hosted West Monroe inventory

The bundled WM inventory under `wm-brand-assets/references/asset-inventory.json` reports 1,331 entries with path, directory, filename, extension, byte size and modified date. Its schema has no rich visual description, alt text, dimensions, subject tags or per-asset industry metadata. The companion asset index records inferred folder-level usage guidance; those are catalog-level conventions, not verified descriptions of a particular image.

Use the actual helper `wm-brand-assets/scripts/find_asset.py` to query filenames and folder metadata, for example:

```sh
python3 <wm-brand-assets-skill-dir>/scripts/find_asset.py "handdrawn arrow" --directory handdrawn-animations/svg --limit 5
python3 <wm-brand-assets-skill-dir>/scripts/find_asset.py "commercial banking" --directory icons --limit 8
python3 <wm-brand-assets-skill-dir>/scripts/find_asset.py "AdobeStock" --directory photos/industry --limit 10
```

Inspect selected assets before assigning semantic meaning. Folder/filename matches are candidate retrieval evidence only. Do not manufacture “architecture”, “data”, “roadmap”, or client/industry tags from an icon's filename; do not describe or repurpose a retained source image without a real description or visual inspection.

## Family asset guidance (candidate use only)

- Roadmap and process slides can use native editable bars, labels, gates, and relationships already supported by the bounded shape/layout engine. Hand-drawn arrows/highlights can accent a conclusion only after inspecting the asset and verifying phrase/geometry placement; they do not encode schedule data.
- Architecture diagrams should express actual components and relationships with native shapes/lines when available. Icons from `icons/png` or `icons/svgs` are optional visual labels, not a substitute for a component definition or architecture evidence.
- Team slides should represent sourced roles and reporting relationships as editable role/line objects. Do not infer people, credentials, client staffing, or employment type from photos or stock imagery. Use staffing color tokens only with their explicit meanings and a legend.
- Comparison slides should make dimensions, alternatives, evidence, and qualifications explicit. Icons and accents may support scanning, but do not substitute for measured data or create unsupported ratings.
- Abstract, industry, and provocative photography may set context or storytelling tone where the actual subject fits. Folder placement narrows the search; inspect the specific photo before making a client/industry claim. Never treat imagery as evidence of a client's environment or work completed.

All such suggestions are optional. The adaptive family builders are executable. The checkpoint records exact examples that have completed native verification and visual review; evidence applies only to those inputs, and broader family qualification remains pending. Comparison supports linear or dial gauges; all criterion rows need at least 45pt and dial rows need at least 58pt. Review status comes from the exact adaptive checkpoint, not taxonomy category.

### Taxonomy versus motif borrowing

`t027-software-modernization-035` carries the `roadmap` taxonomy hint because that is its registered source category. A process-family fixture borrows only its three-stage cutover header as a narrow visual/text motif; it does not implement the lower cutover matrix. This is not a suitability or routing decision. The semantic family is selected explicitly in the adaptive spec, and source structure needs review before reuse.
