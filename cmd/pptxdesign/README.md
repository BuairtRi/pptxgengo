# WMDS Go presentation authoring

`pptxdesign` loads the retained v5 source library, resolves pinned fonts, grids,
frames and artwork, and writes editable PowerPoint objects. Normal generation
runs in Go. PowerPoint rendering is the native review step.

The library contains **587 templates (586 active)** at source commit
`d83bd58a9f9de68ebd8d6b3c9b0272c16ed516cf`. The accepted gallery qualifies each
illustrated source specimen. New text, diagrams and images require a separate
fit check and native review.

## Build and discover

Run from the repository root:

```sh
go build -o /tmp/pptxdesign ./cmd/pptxdesign
/tmp/pptxdesign library-find --index library/wm-design-system/v5/library.sqlite \
  --query 'buy build modernization economics' --kinds template --limit 5 --summary
/tmp/pptxdesign library-inspect --index library/wm-design-system/v5/library.sqlite \
  --id decision/buy-build-economics --summary
/tmp/pptxdesign library-preview --index library/wm-design-system/v5/library.sqlite \
  --id decision/buy-build-economics
```

`--summary` returns purpose, topology, zone bindings and verified screenshot paths.
Open those screenshots when choosing a layout. Search ranking is a discovery
signal, not a fit guarantee. See [semantic template discovery](../../docs/semantic-template-discovery.md).

To make a new index after deliberate relocation:

```sh
/tmp/pptxdesign library-index --bundle library/wm-design-system/v5 \
  --gallery library/wm-design-system/v5/catalog --out /tmp/NEW-library.sqlite
```

The SQLite index pins the matching source, bundle and gallery. Its artifact paths
are relative to `meta.report.options.gallery`; explicit `--gallery` and `--bundle`
overrides support relocation while preserving all hash checks.

## Maintain a deck

Copy [the project starter](../../examples/deck-project/README.md) into a new project,
then initialize, check and build with the same compiled executable:

```sh
cp -R examples/deck-project /tmp/my-deck-project
/tmp/pptxdesign project init --project /tmp/my-deck-project --bundle library/wm-design-system/v5
/tmp/pptxdesign project check --project /tmp/my-deck-project --bundle library/wm-design-system/v5
/tmp/pptxdesign project build --project /tmp/my-deck-project --bundle library/wm-design-system/v5
```

Source context, slide brief files, notes, hidden states, sections and custom assets
remain maintained inputs. Builds and receipts are immutable derived artifacts.
The toolchain lock pins the executable, engine, library, fonts and asset hashes.
Save PowerPoint edits separately; baseline changes block regeneration.

## Derive and edit a composition

```sh
/tmp/pptxdesign project scaffold --bundle library/wm-design-system/v5 \
  --template lifecycle/three-phases --reason 'Preserve the authored phase structure' \
  --out /tmp/NEW-local-scaffold.json
/tmp/pptxdesign project edit --project /tmp/my-deck-project --patch slide-edits.json
/tmp/pptxdesign project measure --report MY-BUILD/layout-report.json
```

Scaffolding emits a local definition with actual shared ancestry and separately
labelled synthetic values. Replace those values with real content. It does not
provide visual acceptance. Typed legacy bindings and unsupported topology fail
explicitly rather than borrowing source copy.

An edit patch is keyed by stable slide ID. Each entry may replace `values`,
`template: {scope, id}`, or a `brief` project-relative Markdown path. The command
checks the source and bindings before atomic replacement, saves its predecessor,
and preserves notes, hidden state and section membership. Measurement findings
are advisory; chart/table internals and rotated text need visual inspection.

The [local composition example](../../examples/local-composition/README.md)
demonstrates Venn, maturity and road components. See the
[project contract](../../internal/deckproject/README.md) for strict source validation,
assets, source lineage, approvals and portable exports. The retained
[typography calculation contract](../../library/wm-design-system/v5/typography/README.md)
describes calibrated measurements and their limits; the engine identifier remains
`wmds-go-foundation.v2`.
