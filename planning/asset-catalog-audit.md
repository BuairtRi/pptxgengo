# Asset catalog implementation audit

Scope: read-only audit of the bundled WM image inventory and `/Users/rscott/Documents/branding`; this is a proposed implementation plan, not a catalog build.

## Observed sources and coverage

- **Skill inventory (authoritative for currently hosted URLs):** `wm-brand-assets/references/asset-inventory.json`, generated 2026-06-04, host `https://assets.westmonroe-cloud.com`, 1,331 records. Each record has only `path`, `directory`, `fileName`, `extension`, `assetType`, `sizeBytes`, and `modifiedAt`; no dimensions, visual description, alt text, tags, usage rules, or explicit subject/industry metadata. The companion `asset-index.md` repeats this inventory and derives inferred usage notes from filenames/folders.
- **Hosted inventory coverage:** 1,103 icons (551 PNG, 552 SVG; grounded-blue 270 each, magenta 281 PNG/282 SVG), 26 hand-drawn assets (13 PNG + 13 SVG), 6 logos (PNG/SVG pairs of three lockups), and 196 photos (53 abstract, 62 industry, 81 provocative; 101 JPEG + 95 JPG). These are actual observed counts.
- **Local photo descriptions:** 521 `.md` sidecars, each paired by identical basename with a local `.jpeg`/`.jpg` photo: 294 JPEG + 227 JPG. Example sidecar includes a descriptive title, proposed filename, narrative description, use guidance, title, alt text, keywords, style, recommended placement, and trending topic. The 521 sidecars are 521 distinct files; their descriptive filenames do not directly match hosted inventory basenames (zero basename matches observed). Treat this as a rich local source, not proof of hosted-photo identity.
- **Other local material:** brand guidance includes dedicated Logo, Iconography, Illustration, Hand-drawn Graphics, Highlight Graphic, and Photography docs. Local `assets/graphics/arrows` has 20 files: 4 variants (connecting, dashed, double, right-angle) across PNG/SVG/EPS, with white variants for several PNG/EPS forms. `assets/illustrations` and `assets/icons` currently contain no files. Separate local logo library includes multiple formats/colorways and WM Capital, ERG, initiative, and product marks; it is broader than the six hosted logo entries. No JSON/CSV/YAML catalog manifest was found in `/Users/rscott/Documents/branding`.

## Proposed data design and source joins

Use source-scoped asset records for both hosted and local material and preserve their source fields verbatim; a local asset need not have a hosted counterpart. Add optional normalized fields: `asset_id`, `source_system`, `source_path`, `canonical_url`, `content_sha256`, `width_px`, `height_px`, `mime_type`, `visual_description`, `alt_text`, `keywords[]`, `style[]`, `recommended_placement[]`, `industry_or_topic[]`, `brand_family`, `variant_group_id`, `variant_role` (colorway/format/lockup), `license_or_usage_note`, `description_status`, and provenance records with source path, author/source, creation date, and whether text is original or generated.

1. **Exact identity first:** key hosted records by normalized full `path` (case-preserving) and canonical URL; never join by basename alone. For local items, record absolute-root-relative path and compute SHA-256 when processing binaries. Hash equality is the strongest cross-root dedup signal; retain all source paths as aliases instead of discarding duplicate rows.
2. **Descriptions before generation:** for each local photo, ingest its paired sidecar fields exactly and mark provenance `branding-sidecar`; do not rewrite its prose or overwrite explicit alt text. Match sidecars to hosted records only with exact source filename/path, known source identifier (e.g. Adobe/Getty ID), or matching content hash. If no defensible match exists, retain the sidecar as a local-photo record, unjoined, and queue an identity review. Use title/keywords and image inspection to add missing hosted descriptions only after reuse matching.
3. **Variant joins:** pair icon and logo renditions by shared semantic stem plus documented family/colorway and format; store one conceptual family with rendition rows. Do not infer that two same-stem files are byte-identical. Compare hashes; SVG/PNG rendition pairs will normally differ by design. Preserve separate lockups and colorways.
4. **Human-approved semantics:** filename/folder-derived terms stay labeled inferred. Separate `observed` facts (extension, dimensions, hash, path) from `editorial` description and `inferred` category/tags. Store source text and any generated enrichment in different fields with timestamps/model/tool metadata.

## Gaps and sequencing proposal

- Add technical metadata for every hosted/local asset: pixel dimensions, MIME, SHA-256, alpha/transparency where relevant, SVG viewBox, and a perceptual hash for near-duplicate review. This supports fit and dedup without pretending similar assets are identical.
- Prioritize actual visual descriptions and accessible alt text for photos and ambiguous icon/illustration files. For icons, existing semantic filename fragments are useful candidate keywords but should be visually checked for generic names such as `Asset 12.png`. For logos, catalog lockup, orientation, color/reversal, minimum background/use guidance from the local Logo guide.
- Ingest existing sidecar text with provenance before any new description pass. The high-value controlled fields already populated locally (subject description, placement, keywords, style) should be reused; identify unmapped photos rather than regenerate wholesale.
- Fold local graphic and identity libraries into a broader catalog source namespace instead of assuming the hosted inventory is exhaustive. Include illustrations, photos, icons, logos, and accents as asset families, with source/availability and format variants explicit.

## Arrow placement fixture plan

The hosted handdrawn family currently offers 5 arrow concepts (single, connecting, dashed, double, right-angle) in PNG and SVG. Local arrows add 4 named styles in several color/format versions, including white. Build a placement fixture matrix from those real variants: straight progression; dashed/secondary path; two-way comparison; right-angle callout around a chart/card; connecting path between nodes; each in short/medium/long visual span, horizontal/vertical/diagonal rotation, dark/light background, and with/without nearby text or image edges. Capture crop bounds, transparency, line visibility at presentation scale, direction/orientation, and collision with slide content. Pair SVG and PNG fixtures to assess rendering differences; include a white-on-dark case and compare EPS only in workflows that support EPS. Use synthetic layout content and record the exact source asset path and placement transform per fixture. This expands useful placement evidence beyond a single generic arrow example without generating new artwork prematurely.

## Recommended implementation order

1. Freeze a schema and provenance policy; import inventory/local files as source records without changing source prose.
2. Compute metadata/hashes and establish exact/path/identifier/hash joins; emit unresolved identity queue for the 521 rich sidecars.
3. Ingest matching existing descriptions verbatim; review unmatched sidecars and duplicate/variant candidates.
4. Visually enrich only records still missing required fields, prioritizing hosted photos, ambiguous icons, logo usage, and arrow/illustration assets.
5. Add fixture references and editorial usage guidance, keeping inferred notes separate from approved brand rules.

Illustration coverage remains an explicit gap: the observed local illustrations directory is empty. Inspect registered decks and other supplied brand sources for reusable illustrations, preserving original provenance and native editability where available; do not report an illustration database as complete based on directory creation alone.
