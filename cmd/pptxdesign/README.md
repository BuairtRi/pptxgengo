# WMDS Go presentation authoring

`pptxdesign` loads the current v7 source library, resolves pinned fonts, grids,
frames and artwork, and writes editable PowerPoint objects. Normal generation
runs in Go. PowerPoint rendering is the native review step.

The [engineering command reference](../../docs/engineering-cli.md) covers
content-first matching, aliases/capacity, asset discovery/registration, slide
mutations/swaps, staged review packets and source inventories.

### Local native rendering (macOS)

```sh
pptxgengo design render --pptx /path/deck.pptx --out /path/new-review --pdf --png
pptxgengo design render-doctor
pptxgengo design render --pptx /path/deck.pptx --out /path/new-selected-review --png --slides 3,5-7 --contact-sheet
# Include hidden slides in the inspection copy:
pptxgengo design render --pptx /path/deck.pptx --out /path/new-all-slides --pdf --png --include-hidden --timeout 10m
```

This command uses installed Microsoft PowerPoint's native PDF save API and
macOS PDFKit for PNGs. It opens a uniquely named temporary copy, closes that
copy without saving, and never opens or writes the original. Existing output
directories are rejected. Hidden slides retain their original state; the
optional flag makes them visible only in the temporary copy. PNGs fit within
1920 × 1080 while preserving the PDF aspect ratio. `render-manifest.json`
records source/output hashes and the exported page count. PNG-only output still
uses a native PDF internally, then removes that PDF.

Run in a logged-in macOS GUI session with Microsoft PowerPoint installed at
`/Applications/Microsoft PowerPoint.app`, Swift tools available, and automation
permission for the invoking terminal. A remote or isolated agent shell can lack
access to the GUI session (Apple Event errors −10827/−600); such an export fails
explicitly and does not publish a successful manifest. There is no cloud or
LibreOffice fallback. The command's timeout terminates its subprocess; it does
not terminate PowerPoint or write other open presentations. If an app dialog
blocks export, dismiss it and inspect `render-error.txt` before retrying in a
new output directory.

The library contains **616 templates (615 active)** at source commit
`c788cefeb5bb409118ac217adb53216d8156eec3`. The accepted gallery qualifies each
illustrated source specimen. New text, diagrams and images require a separate
fit check and native review.

## Build and discover

Run from the repository root:

```sh
go build -o /tmp/pptxdesign ./cmd/pptxdesign
/tmp/pptxdesign library-find --index library/wm-design-system/v7/library.sqlite \
  --query 'buy build modernization economics' --kinds template --limit 5 --summary
/tmp/pptxdesign library-inspect --index library/wm-design-system/v7/library.sqlite \
  --id decision/buy-build-economics --summary
/tmp/pptxdesign library-preview --index library/wm-design-system/v7/library.sqlite \
  --id decision/buy-build-economics
```

`--summary` returns purpose, topology, zone bindings and verified screenshot paths.
Open those screenshots when choosing a layout. Search ranking is a discovery
signal, not a fit guarantee. See [semantic template discovery](../../docs/semantic-template-discovery.md).

To make a new index after deliberate relocation:

```sh
/tmp/pptxdesign library-index --bundle library/wm-design-system/v7 \
  --gallery library/wm-design-system/v7/catalog --out /tmp/NEW-library.sqlite
```

The SQLite index pins the matching source, bundle and gallery. Its artifact paths
are relative to `meta.report.options.gallery`; explicit `--gallery` and `--bundle`
overrides support relocation while preserving all hash checks.

## Maintain a deck

Copy [the project starter](../../examples/deck-project/README.md) into a new project,
then initialize, check and build with the same compiled executable:

```sh
cp -R examples/deck-project /tmp/my-deck-project
/tmp/pptxdesign project init --project /tmp/my-deck-project --bundle library/wm-design-system/v7
/tmp/pptxdesign project check --project /tmp/my-deck-project --bundle library/wm-design-system/v7
/tmp/pptxdesign project build --project /tmp/my-deck-project --bundle library/wm-design-system/v7
```

Source context, slide brief files, notes, hidden states, sections and custom assets
remain maintained inputs. Builds and receipts are immutable derived artifacts.
The toolchain lock pins the executable, engine, library, fonts and asset hashes.
Save PowerPoint edits separately; baseline changes block regeneration.

## Start with a shared stock slide

```sh
pptxgengo design project scaffold --stock --template cards/3 \
  --id three-actions --out slides/076-three-actions.yaml
```

The stock scaffold keeps the actual shared layout and emits editable YAML with
readable `content`. Its initial copy is explicitly marked `synthetic_example`.
Replace that copy, set `content_kind: supplied_content`, and add the file to the
ordered `slides` list in `deck.yaml`. Required fields and exact item counts still
come from the selected library contract. Shared stock slides need no local
template definition or adaptation reason.

For a declared stock image field, use the project's asset ID as its content
value, for example `photo: client-logo` when `assets.client-logo.path` names the
image file in `deck.yaml`. Compilation resolves that ID only in source-declared
media slots; ordinary copy is unchanged. Existing library registry keys remain
valid. Use `project:client-logo` to explicitly select a project asset if its ID
collides with a library registry key. The scaffold's image field name depends on
the selected template.

## Derive and edit a composition

Create a local derivative when the actual layout needs a deliberate change:

```sh
/tmp/pptxdesign project scaffold --bundle library/wm-design-system/v7 \
  --template lifecycle/three-phases --reason 'Preserve the authored phase structure' \
  --out /tmp/NEW-local-scaffold.json
/tmp/pptxdesign project edit --project /tmp/my-deck-project --patch slide-edits.json
/tmp/pptxdesign project measure --report MY-BUILD/layout-report.json
```

Scaffolding emits a local definition with actual shared ancestry and separately
labelled synthetic values. Replace those values with real content. It does not
provide visual acceptance. Typed legacy bindings and unsupported topology fail
explicitly rather than borrowing source copy.

An edit patch is keyed by stable slide ID. Each entry may replace `values`, `content`, `bindings`,
`template: {scope, id}`, or a `brief` project-relative Markdown path. The command
checks the source and bindings before atomic replacement, saves its predecessor,
and preserves notes, hidden state and section membership. Measurement findings
are advisory; chart/table internals and rotated text need visual inspection.

### Split a deck into editable slide files

```sh
pptxdesign project split --project /path/to/project --bundle v7
```

`deck.yaml` becomes the ordered index. Individual slide files contain human copy;
speaker notes move to `notes/*.md`; rare local layouts move to `templates/*.yaml`.
Stock slides keep their shared template and use readable `content` with explicit
`bindings` to the original contract. All references are project-relative.

The converter checks that expanded canonical content is unchanged, retains the
existing lock and a recoverable source predecessor, and refuses conflicting
destination files. `check`, `build`, `edit`, `fork`, `detach`, sections and exports
accept both inline and multi-file sources. Slide edits preserve unselected file
bytes and comments on untouched nodes. `content` patches replace the complete
content map; they do not merge arrays or fill missing copy from examples.

The [local composition example](../../examples/local-composition/README.md)
demonstrates Venn, maturity and road components. See the
[project contract](../../internal/deckproject/README.md) for strict source validation,
assets, source lineage, approvals and portable exports. The retained
[typography calculation contract](../../library/wm-design-system/v7/typography/README.md)
describes calibrated measurements and their limits; the engine identifier remains
`wmds-go-foundation.v2`.
