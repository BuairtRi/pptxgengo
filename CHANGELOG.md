# pptxgengo changelog

Project release artifacts and CI are private in GitLab. Dates below use the
release pipeline's UTC date; release tags and signed artifacts are immutable.

## [v4.2.1](https://gitlab.samcott.com/riscott/pptxgengo/-/releases/v4.2.1) — 2026-10-08

- Native-v1 becomes the default for new v2 projects and template/reference/browsing generation. Existing project profiles remain pinned.
- Eligible lists become native bullet paragraphs, simple cards combine into one editable shape, and redundant table wrappers are removed while native table cells remain intact. Unsupported decoration/geometry has explicit retention reasons.
- All six platform archives include a 717-slide template browsing deck, coverage report, pinned V11 source/catalog/fonts and rebuilt relocatable SQLite index with 649 templates.
- Browsing media use explicit illustrative placeholders. Curated reusable content and private branding/photo originals remain deferred.
- macOS Developer ID/notarization, Azure Windows signing, artifact scans and private Sigstore manifest verification passed. Windows desktop and full-catalog native qualification remain pending.

## [v4.2.0](https://gitlab.samcott.com/riscott/pptxgengo/-/releases/v4.2.0) — 2026-10-07

- Promoted the portable-project, numbered version/share, native-editing component and reviewed text-reconciliation feature work.
- Published signed/notarized binaries for macOS, Linux and Windows, amd64 and arm64, plus the optional offline model archive and signed security evidence.
- These archives were binary-only: template sources, SQLite, fonts and browsing decks were not included. v4.2.1 adds those template resources.

See [release status and qualification](docs/release-status.md) for exact source,
pipeline and verification scope. The [upstream PptxGenJS changelog](docs/upstream-pptxgenjs-changelog.md)
is retained separately as inherited project history.
