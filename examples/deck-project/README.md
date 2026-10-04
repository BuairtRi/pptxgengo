# Editable deck project

The authoritative source is `deck.yaml`. It contains a shared `cards/3` slide and a
local frame/grid composition with block text, a paint-ordered group and an original
PNG asset. All content is synthetic. Build output, state and a compiler-specific
lock are generated after copying this starter into your own project directory.

Build the executable once, then use that same executable for the project:

```sh
go build -o /tmp/pptxdesign ./cmd/pptxdesign
cp -R examples/deck-project /tmp/my-deck-project
/tmp/pptxdesign project init --project /tmp/my-deck-project --bundle library/wm-design-system/v5
/tmp/pptxdesign project check --project /tmp/my-deck-project --bundle library/wm-design-system/v5
/tmp/pptxdesign project build --project /tmp/my-deck-project --bundle library/wm-design-system/v5
/tmp/pptxdesign project status --project /tmp/my-deck-project
/tmp/pptxdesign project approve --project /tmp/my-deck-project --stage content --actor maintainer
/tmp/pptxdesign project review --project /tmp/my-deck-project --out /tmp/reviewer.zip
/tmp/pptxdesign project export --project /tmp/my-deck-project --mode client --out /tmp/client.zip
/tmp/pptxdesign project export --project /tmp/my-deck-project --mode maintainer --out /tmp/maintainer.zip
/tmp/pptxdesign project export --project /tmp/my-deck-project --bundle library/wm-design-system/v5 --mode offline --out /tmp/offline.zip
```

Run from the repository root, or provide an absolute v5 `--bundle` path. `go run` may produce a different executable hash;
the project deliberately pins the compiler executable, Go version, OS/architecture,
engine and every bundle file, including fonts/calibration/source/assets. A changed
compiler requires an explicit reviewed re-pin with the prior lock preserved.

Each build writes a new `builds/build-.../` directory containing a PowerPoint,
source snapshot, canonical source, compiled scene, measured layout report, native
object map, lock snapshot and receipt. Native bytes are deterministic under the
same pinned inputs; build IDs and receipt times identify separate build events.

Edit `deck.yaml`, source context and original assets. Generated build artifacts are
read-only baselines. Save PowerPoint edits as separate files: changing a baseline
blocks regeneration and export until the divergence is resolved. Automatic edited
PowerPoint reconciliation and PDF import/export are not implemented by this CLI.
Client ZIPs contain the recorded PPTX and export manifest. Reviewer ZIPs also
contain layout/map/receipt and visible draft bindings. Maintainer/offline ZIPs
contain private context and evidence; offline packages add the original compiler,
pinned library/fonts and used registry assets, with OS-specific execution guidance.

A deck-specific variant of the local template can be created without copying its
slide text into geometry:

```sh
/tmp/pptxdesign project fork --project /tmp/my-deck-project --template editorial-photo --as editorial-client --slides local-composition --reason 'Adjust this deck composition'
```

Generic source-scene shared templates can be detached with `project detach --slide
STABLE-ID --as NEW-LOCAL-ID --bundle library/wm-design-system/v5 --reason '...'`. The command retains actual
caller-provided content, native frame/nav/chrome geometry and frozen ancestry.
Legacy typed card-row/metric IR, source-note/stamp chrome and geometry outside
supported local zones return explicit errors and leave the YAML unchanged. Fork
and detach rewrite YAML formatting/comments; byte-exact predecessor snapshots and
immutable definition snapshots are retained under `decisions/`.

Full contracts and supported composition vocabulary: `internal/deckproject/README.md`.
