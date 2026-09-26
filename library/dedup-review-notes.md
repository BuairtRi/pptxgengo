# Duplicate fingerprint review notes

## Inventory risks

- `scripts/inventory-pptx.py` is structural evidence, not a visual or lossless fingerprint. It records source/package hashes, direct shape trees, text, local transforms, explicit font attributes and resolved relationship targets. It does **not** resolve inherited style, flatten group transforms, retain all shape XML, nor record `p:presentation` default text style/theme semantics as effective values.
- Raw IDs/names and relationship IDs are not identity proof. `slide_id`, `cNvPr/@id`, shape names, `rId`s, part filenames, and relationship ordering can change during copy/save. Preserve them as provenance; canonicalize only after references are remapped and validated.
- Fingerprint each candidate with slide size, canonical slide XML, relationship type/mode and the recursively resolved part hashes for layout, master, theme, media, chart/workbook, SmartArt, OLE, fonts and other dependencies. Resolve the style cascade through placeholder → layout → master → theme/color map and presentation `defaultTextStyle`; include effective run/paragraph properties.
- For group geometry include nested `off/ext/chOff/chExt`, rotations/flips and child order. Local child bounds alone can make unrelated groups look alike or identical groups look different.
- Unknown/unresolved relationships, `a:graphicData`, `p:oleObj`, `p:graphicFrame`, alternate content, external links and embedded workbooks should set an `opaque_or_uncertain` flag. Do not call these exact visual duplicates from their placeholders, preview images or ProgIDs alone.

## Distinct matching outcomes

1. **Exact package/source duplicate:** same canonical slide XML plus same dependency graph/content hashes and slide size; report original source hashes/locators.
2. **Same designed arrangement, different copy/data:** same canonicalized geometry, ordering, effective styles and asset references, with text/table payload differences listed. This is a layout-family candidate, not a content duplicate.
3. **Same content, different arrangement:** equivalent normalized text/data with distinct geometry or styling. Link as alternate compositions; do not merge designs.
4. **Near match:** bounded direct similarity evidence only; queue for contact-sheet/human review. Similarity links must not union transitively (A≈B and B≈C does not establish A≈C).

Ignore IDs only in a second-stage canonical form after replacing IDs with ordered tree paths and rewriting all internal references to those paths. Keep z-order, group membership, transforms, relationship type/target/mode, table cell structure, text/run boundaries, crop, rotation, flips, hyperlink destinations, animation/connector endpoints and hidden/off-canvas state. Retain a separate raw hash for exact package evidence.

## Quick duplicate/layout candidates

These are inventory-only review candidates, not confirmed duplicates. Listed pairs have repeated titles/object counts and matching top-level kind/transform signatures; text payloads differ, and explicit styles/dependencies are not part of that quick signature.

| Source | Slides | Initial read |
|---|---:|---|
| Graphics and Layouts | 33–34 | “Standard Table”; same six-object geometry, likely content/table-data variants. |
| Graphics and Layouts | 55–56 | “Venn Diagram – 3 Circles”; same 18-object geometry, text differs; compare labels/colors and relationships. |
| Graphics and Layouts | 74–75 | “Timeline”; same 16-object geometry, likely alternate timeline copy. |
| Graphics and Layouts | 111–112 | “Major Points – Value Drivers”; same 24-object geometry, compare text and effective style. |
| EnableComp | 3–8 | Same 30-object geometry and title candidate; review as repeated composition/content variant. |

Do not promote these pairs until resolved-style/dependency fingerprints and native previews agree on the relevant duplicate class. The inventory also shows UHG 29/32 with the same title candidate and 32 objects, but different geometry; that is an alternate-design candidate, not an exact-layout match.

## Review of `deduplicate-layouts.py` first pass

- The exact pass is intentionally high-recall-safe/low-recall: dependency digests hash raw XML/binary bytes. Benign `r:id`, shape IDs, metadata, relationship order, or unrelated inherited objects anywhere in the closure can split genuinely equivalent slides. The 0 exact groups therefore means “no byte-canonical matches,” not “no duplicates.” Keep it as a strong positive signal only.
- `canonical_slide()` drops `cNvPr/@descr` and `@title` with `@name`; alt text/title can differ while slides join as `native_content_equivalent_unrendered`. Keep those fields in the fingerprint or store/report an accessibility-metadata diff before using “content equivalent.”
- The normalized geometry hash rounds coordinates to 4 decimals. Tiny deltas may yield a `same_geometry` score of 1.0; state its tolerance (about 0.0001 slide dimension) and reserve exact/equivalent wording for an unrounded compare.
- More seriously, `similarity()` sorts objects by kind/box and scores only their boxes. When geometry hashes differ, this can still return 1.0 while rotation/flip, preset geometry, placeholder, style or z-order differs; rename the metric `box_similarity` and add explicit mismatch penalties/filters, or make such cases review warnings.
- Candidate matching reorders same-kind shapes by position, so overlapping/repeated objects can be paired differently from their actual z-order. Keep this output as a pair queue only; previews/structural diffs must compare ordered tree paths and overlap/layer changes before curation.
- Exact groups currently pick one representative and compare candidates only against representatives. That is safe only while exact groups remain verified; preserve every occurrence's style, accessibility and dependency deltas and do not hide them behind one representative.
- Group ancestry skipping is a good guard for near-geometry comparisons. Exact XML can still miss equal rendered group layouts encoded with different local coordinate systems; transform flattening belongs in the visual/layout review pass, not in an unsafe fingerprint relaxation.

## Resolution after root review — algorithm version 4

The first-pass comments above are historical review findings. The implementation
now retains accessibility title/description, includes table column widths, row
heights and merge attributes, reads `p:xfrm` for graphic frames, and filters near
matches on preset/custom geometry, crop, transform attributes and placeholders.
The geometry queue uses one direct representative/member edge per bucket; it does
not automatically merge families. A score of 1 still means quantized geometry,
not pixel identity. Effective styles and group flattening remain unresolved.

The native-content pass still uses conservative dependency bytes; zero matches
must not be interpreted as zero reusable layouts. See `layout-decisions.json`
for the separately reviewed visual families and preserved variants.
