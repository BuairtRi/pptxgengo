# Brand extras ingestion report

Read-only discovery run; source files and the existing asset manifest were not modified. No network requests or downloads were made.

## Runtime findings

- Configured branding root `/Users/rscott/Documents/branding`: local font files found: 58 (58 unique SHA-256 records); extensions: {'.ttf': 58}.
- Configured branding root `/Users/rscott/Documents/branding`: explicit local illustration candidates: 0 files (0 unique SHA-256 records). Candidate paths: none. Candidate rule: image/vector file inside an `illustrations` path or filename explicitly containing `illustration`; general icons, arrows, graphic accents, photos, and templates are not labeled illustrations.
- Configured hosted inventory `/Users/rscott/.codex/skills/wm-brand-assets/references/asset-inventory.json`: 1331 records; illustration-named paths: 0. Paths: none.
- Fonts have `readiness: discovered`; SFNT family/full name, subfamily/style, and OS/2 `fsType` were extracted when safely available. `embedding_rights` remains `UNKNOWN`; parsed flags report bitfield indicators, not a permission opinion. WOFF/WOFF2, if encountered, remain identifiable by file/hash but metadata is unsupported unless safely parseable.
- Every record includes a content SHA-256, byte size, local provenance path, and aliases for same-hash files. IDs use `sha256:<digest>` like physical assets; this supplementary manifest is a separate source and should be merged without overwriting existing rows.

## Illustration gap and static context

For this run, `/Users/rscott/Documents/branding/assets/illustrations` exists and contains no candidate files. The runtime count above describes the configured sources only and should not be read as a permanent claim about other roots or future inventory versions. Hosted hand-drawn graphics remain categorized separately.

Static local context checked: the current branding workspace includes an Illustration guide describing Adobe Illustrator artwork and linking to an external S3 sample image. That remote image was not downloaded and is not a local source asset. Likely places to inspect manually include the linked Illustration Suite/source library and supplied PowerPoint templates/decks for embedded vector artwork; likely source formats include `.ai`, `.eps`, `.svg`, and high-resolution raster. These are leads, not discovered catalog assets.

## Excluded file types

- Scan target types: TTF, OTF, WOFF/WOFF2 font files; SVG, AI, EPS, PDF, PNG, JPG/JPEG, WEBP only when explicitly illustration-path identified.
- Existing presentation/document sources, ordinary icons, logos, accent graphics, and photographs remain outside this supplementary manifest.

## Read errors

- None.

## Current font families

The 58 records are IBM Plex Sans (16), IBM Plex Sans Condensed (14), IBM Plex Sans
SemiCondensed (14), and IBM Plex Mono (14). This source-folder scan found no Arial
font binaries. It does not inspect installed system/application fonts and does
not change the corporate PowerPoint typography guidance.
