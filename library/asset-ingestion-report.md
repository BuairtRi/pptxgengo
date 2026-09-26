# Asset ingestion report

Run mode: bounded local ingestion; no network requests or asset downloads.

## Counts

- Hosted inventory records read: 1331
- Canonical JSONL asset records written: 2169 (local exact-byte duplicates collapse to SHA-256 IDs; hosted entries remain source-scoped)
- Existing paired photo sidecars preserved verbatim: 521
- Local duplicate-hash groups: 22
- Exact hosted-path/local-file candidates: 0

| Category | Catalog records |
|---|---:|
| photo | 715 |
| icon | 1103 |
| illustration | 0 |
| logo | 207 |
| accent | 73 |
| other | 71 |

## Provenance and coverage

- Local records point to paths relative to the configured branding root and carry SHA-256 when readable. PNG/JPEG dimensions are parsed with the Python standard library; SVG pixel dimensions are used only when explicit numeric px or unitless width/height are present. SVG viewBox is retained separately as coordinate metadata, never converted to pixels. Unsupported/unreadable formats have null dimensions; no values are guessed.
- Sidecar text is stored as decoded UTF-8 exactly as read and JSON-escaped in JSONL; it is not parsed, summarized, or rewritten. Provenance identifies the relative sidecar path and `existing-branding-sidecar` authority.
- Hosted rows retain source inventory metadata and canonical hosted URL. Since binaries are not downloaded, their hash and dimensions are null; remote visual descriptions, alt text, and keywords are marked unavailable. Filename-derived descriptions are not invented.
- Distinct source files sharing a local hash are represented by one canonical hash ID with every path in `aliases` and `source_refs`. Hosted/local matches are not claimed from basename similarity. Similar filename variants are retained as separate rows with `variant_status: unconfirmed` and no inferred family link.
- Categories are path-inferred and retained as a list when duplicate hashes have aliases from multiple categories. Illustration coverage is measured from files actually found; an empty category is reported as zero. `other` counts unclassified image files; fonts and templates are outside the scan.

## Errors

- None.
