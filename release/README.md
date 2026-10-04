# Current West Monroe presentation release

The current source package is **0.1.0-local.13/v5**. It adds content-first template
matching, source-pinned authoring aliases and advisory capacity, grouped asset
discovery and a browsable asset gallery, slide operations and safe swap proposals,
native rendering diagnostics/subsets/contact sheets, immutable native review
attachments, staged HTML reviewer packets, and compact source inventories.

The local.13 follow-up adds audience-only review packets, ranked incomplete
matcher drafts and bounded lifecycle/comparison adapters, more flexible swaps,
earlier fit checks, readable detach, preserved author comments, section-aware
insertion, signed local render receipts, improved asset search and six additional
photos, and renderer-derived card/table/stepper capacity. Catalog caching and
separate test lanes shorten the normal development loop. See the
[follow-up disposition](../docs/engineering-followups.md) and
[testing guide](../docs/testing.md) for qualification and remaining limits.

See the [engineering command reference](../docs/engineering-cli.md) and
[qualification record](../docs/engineering-waves.md). Metadata inference is
distinct from semantic review. Native export still requires an eligible macOS
GUI caller; this agent session returns Apple Event errors −10827/−600.

The retained source library is **v5**: 587 templates (586 active), pinned to
`d83bd58a9f9de68ebd8d6b3c9b0272c16ed516cf`. Every source specimen has an accepted
native PowerPoint preview. Changed content requires its own fit and native review.

The Go compiler supports maintained `deck.yaml` descriptors and ordered slide
YAML files, shared and local
compositions, real brief files, notes, hidden slides, native sections, editable
charts, source provenance and client/maintainer/offline packages. Typography uses
`wmds-go-foundation.v2`; source revision and engine identity are separate.

`pptxgengo design render --pptx deck.pptx --out review --pdf --png --include-hidden`
exports through installed PowerPoint and renders PNGs with macOS PDFKit. This
requires a normal macOS GUI session with PowerPoint automation access. The command
reports inaccessible GUI sessions or native-export failures and never substitutes
another renderer. Actual PowerPoint export is not qualified from an isolated
automation process; the embedded PDFKit PNG output is checked against native
PowerPoint review images.

## Source and discovery

Run from the repository root:

```sh
go build -o /tmp/pptxdesign ./cmd/pptxdesign
/tmp/pptxdesign library-find --index library/wm-design-system/v5/library.sqlite \
  --query 'modernization roadmap' --kinds template --summary
/tmp/pptxdesign library-inspect --index library/wm-design-system/v5/library.sqlite \
  --id lifecycle/three-phases --summary
```

The catalog is `library/wm-design-system/v5/catalog/design-system.html`; SQLite is
`library/wm-design-system/v5/library.sqlite`. Artifact links are relative to the
catalog root. The index records resolved roots in its metadata and verifies hashes.
See [semantic discovery](../docs/semantic-template-discovery.md) and the executable
[project starter](../examples/deck-project/README.md).

## Package or install

```sh
scripts/install-local-release.sh --stage-only /absolute/path/new-v5-stage
```

The installer packages only v5, the current Go authoring commands, skill, schema,
examples, registered artwork and discovery artifacts. It validates 587 accepted
source previews and all 1,760 linked artifacts, writes a release manifest, and
records final installation paths in SQLite before moving the stage into place.
It does not reconstruct old sample proofs or package previous library revisions.

Staging does not change command or skill links. Running without `--stage-only`
installs `release/VERSION` and switches those links. Existing release directories
are never overwritten; publishing another version requires a new version value.
Repository cleanup does not update an already installed release.

Use `scripts/install-local-release.sh --cli-only` to update the globally installed
CLI while preserving the operator's existing presentation skill installation.

Required: Go 1.27.1+, Python 3 and registered branding files (set
`WMDS_BRANDING_ROOT` if they are outside `~/Documents/branding`). Native review
requires local PowerPoint. Generation itself runs in Go.
