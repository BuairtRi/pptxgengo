# Modern West Monroe design-system authoring

Use the installed packaged route `pptxgengo design`. It defaults to the frozen
v2 design library and the Go typography engine `wmds-go-foundation.v2`. It uses
the installed release's library, fonts and registered assets. No repository
checkout, mutable source path or separate AppleScript is required for a normal
build. Earlier design contracts are available with explicit `--bundle v1`.
This route does not alter the legacy `template`, `compose`, `adapt`, `scene`,
`component` or `lib` workflows.

## Find a design and inspect its contract

```sh
pptxgengo catalog --design-system --open
pptxgengo design library-catalog
pptxgengo design library-catalog --include-deprecated
pptxgengo design library-catalog --template-keys architecture/layers-nav
pptxgengo design asset-catalog
```

The library contains **167 designs: 166 active and one deprecated**. The gallery
hides the deprecated design by default and identifies its replacement. Choose a
family/layout that supports the intended argument and content density. Compare
its source and changed-content specimens, then download the chosen content or
illustrated slide. The catalog exposes exact named slots, kinds, array
cardinalities, source identity and revision metadata. The two sample slides were
reviewed in PowerPoint; their acceptance does not prove that any replacement copy
will fit.

## Fill a template from one-slide content

Download `source-values.json` or `alternate-values.json` from a gallery entry.
The file is a `pptxgengo.wmds-template-document.v1` document containing one bound
slide. Change its explicit content and retain its template key. Keep
`content_kind:synthetic_example` for illustrative copy; set
`content_kind:supplied_content` when using the user's supported content. Use a new
output directory:

```sh
pptxgengo design template --spec alternate-values.json --out /tmp/wm-slide-new
```

This route fills the named content contract. Supply every required slot and all
fixed-array keys; absent content is not taken from the source example. Keep value
kinds correct: strings/rich text, numbers and booleans remain distinct. Arrays
have exact source item counts and order; stable caller keys preserve native
identities. Table-group column indices, diagram relationships, placement, base
styles, frame geometry and fixed artwork are structural. Content values do not
implicitly move or resize these features.

Two retained card-row designs, `cards/3` and `cards/4`, use typed values instead
of the generic `slots`/`keys` projection. Their downloaded content supplies
`eyebrow`, `title` and `cards` directly under `values`. Keep exactly three or four
cards respectively; each card requires a unique authored `key`, `title` and
`body`, and array order controls the card order. The gallery shows these field
descriptions and exact counts. Preserve the downloaded typed structure.

Navigation variants additionally require 2–6 keyed labels and an active key
naming one item. Select registered photo/icon assets when a contract exposes
those fields. Use `design asset-catalog` to inspect available IDs. Qualitative
quadrant coordinates, table metrics and checkbox booleans follow their catalog
kinds and renderer bounds. Geometry, colors and font sizes are not general
content slots.

Keep synthetic claims, metrics and sample identities marked as illustrative until
replaced by supported content. Source-authored client names, quotations and
metrics may be synthetic specimen copy; do not present them as client facts.

## Keep an illustrated composition editable

Download `source.foundation.json` or `alternate.foundation.json` to retain the
exact native composition behind a preview. These are one-slide
`pptxgengo.wmds-foundation.v1` documents. They include explicit editable scene
nodes, illustrated thumbnails/scorecards and any documented composition
amendments, rather than a flattened screenshot.

```sh
pptxgengo design build --spec alternate.foundation.json --out /tmp/wm-composition-new
```

Use this route when the composition itself is the starting point. Change only
supported scene properties and preserve source provenance. Filling a template's
content contract and building its illustrated composition are separate actions;
a thumbnail's explicit example diagram is present in the foundation download.
A content binding alone does not synthesize that illustration from prose.

For the earlier bundle, select it explicitly and use values or a foundation
compatible with its earlier contracts:

```sh
pptxgengo design library-catalog --bundle v1
pptxgengo design template --bundle v1 --spec v1-values.json --out /tmp/wm-v1-new
```

Do not pass a v2 revision-specific contract to v1. Reusing an output directory is
rejected; choose a new directory for each build.

## Typography and final review

Normal generation and layout run in Go from bundled font files. Output uses the
normal **IBM Plex Sans** and **IBM Plex Mono** family names, including applicable
faces; it does not require measurement aliases. No live PowerPoint character
capture is required for each deck. Native captures support engine calibration and
reference qualification; they are not a permanent normal-build dependency.

The candidate Go engine measures wrapping and fixed content capacity. A
successful build is useful evidence, but it does not guarantee native font
selection or arbitrary replacement-copy fit. Open and inspect the final deck in
PowerPoint at presentation size, and review a native local-printing PDF export
before sharing when visual fidelity matters. Resolve overflow by editing the
argument, selecting a roomier template or splitting the slide. Do not silently
shrink fonts, truncate copy or hide measured failures.

The release's paired specimens qualify those reviewed samples. They do not add
variable row counts, automatic continuation, unrestricted layout or a general
content envelope for every design. Preserve that distinction in delivery notes.
