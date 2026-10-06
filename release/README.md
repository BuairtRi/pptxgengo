# Current West Monroe presentation release

The current source package is **0.1.0-local.20/v11**. Its production gallery and
SQLite discovery index cover all 649 source templates (648 active, one deprecated).
The gallery includes 352 freshly reviewed PowerPoint specimens and 297 inherited
previews verified by exact composition and visible dependency comparison.
The package contains one production bundle, selected by
`release/default-bundle.txt`.

The asset index includes **1,204 variants** and all **521 original photos** in
West Monroe Photos, with their existing descriptive sidecars. The visual asset
gallery has 760 concepts. Photo search uses pinned metadata; selected original
bytes are verified for preview, rendering and packaging. See
[photography discovery](../docs/semantic-template-discovery.md#photography-discovery).

The local.20 release adds slide body density, independent header density, automatic
body fitting with warnings, source-owned density limits, `measure-style`, and
file-based template selection. It imports the designer's 215 template revisions,
later contrast repairs and updated pillar treatments. All 1,927 supported
template/density combinations pass contrast checks; 20 combinations are explicitly
prohibited by ten Comfortable-only template limits. Typography uses 312 exact
native controls for this source; historical source behavior remains isolated.

The preceding local.19 release added 18 templates: five pillar layouts, seven branching roadmaps
and six narrative roadmaps. The workshop category remains at 29 templates. The package includes the complete
design-system documentation board and `pptxgengo docs` server, with exact source
and asset validation during installation. The CLI, docs, gallery, SQLite index
and linked presentation skill are published together. Released docs are selected by
`release/default-docs.txt` so a newer upstream documentation publication can remain
in the source checkout without changing the qualified native release.

Slide-owned Draft Review Notes support independent status text and color and
are removed from client exports. See the
[Draft Review Notes reference](../skills/west-monroe-presentations/references/draft-review-notes.md).

The maintained Go commands support content-first template matching, source-pinned
authoring aliases, advisory capacity, grouped asset discovery, slide operations,
safe swap proposals, native rendering diagnostics, immutable native review
attachments, HTML reviewer packets and compact source inventories. See the
[follow-up disposition](../docs/engineering-followups.md) and
[testing guide](../docs/testing.md) for qualification and remaining limits.

See the [engineering command reference](../docs/engineering-cli.md) and
[qualification record](../docs/engineering-waves.md). Metadata inference is
distinct from semantic review. Native export requires an eligible macOS GUI caller.

The production source library is **v11**, pinned to
`3c56d842ba3abb5f24eb33cf0082be7a7a67f116`. Every source specimen has an accepted
native PowerPoint preview. Changed content requires its own fit and native review.
The bundle and inventory preserve their original source metadata; the gallery
qualification records acceptance of the published specimens. Existing projects
continue to use their explicit bundle and source pins.

The Go compiler supports maintained `deck.yaml` descriptors and ordered slide
YAML files, shared and local
compositions, real brief files, notes, hidden slides, native sections, editable
charts, source provenance and client/maintainer/offline packages. Typography uses
`wmds-go-foundation.v2`; source revision and engine identity are separate.

`pptxgengo design render --pptx deck.pptx --out review --pdf --png --include-hidden`
exports through installed PowerPoint and renders PNGs with macOS PDFKit. This
requires a normal macOS GUI session with PowerPoint automation access. The command
reports inaccessible GUI sessions or native-export failures and never substitutes
another renderer. The v8, v9 and v11 intake reviews were successfully exported through this command
in the current GUI session; signed export receipts are retained beside the
accepted full-size PowerPoint review images.

The inherited library and photography package checks are summarized in
[local.16 qualification](qualification-local16.json). Draft Review Notes checks
and native visual review are recorded in the engineering command reference.

## Source and discovery

Run from the repository root:

```sh
go build -o /tmp/pptxdesign ./cmd/pptxdesign
/tmp/pptxdesign library-find --index library/wm-design-system/v11/library.sqlite \
  --query 'modernization roadmap' --kinds template --summary
/tmp/pptxdesign library-inspect --index library/wm-design-system/v11/library.sqlite \
  --id lifecycle/three-phases --summary
```

The catalog is `library/wm-design-system/v11/catalog/design-system.html`; SQLite is
`library/wm-design-system/v11/library.sqlite`. Artifact links are relative to the
catalog root. The index records resolved roots in its metadata and verifies hashes.
See [semantic discovery](../docs/semantic-template-discovery.md) and the executable
[project starter](../examples/deck-project/README.md).

## Package or install

```sh
scripts/install-local-release.sh --stage-only /absolute/path/new-v11-stage
```

The installer packages only v11, the current Go authoring commands, skill, design documentation, schema,
examples, registered artwork and discovery artifacts. It derives template and
lifecycle counts from the source catalog, validates every accepted source preview
and the exact gallery artifact links, writes a release manifest, and
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

The [local.20 qualification](qualification-local20.json) records density checks, source/bound sweeps, native review, documentation integrity, package verification and the global installation.
