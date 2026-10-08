# Design-system documentation

Use the upstream design-system docs board to learn the system's vocabulary and see how its parts fit together. The board covers foundations, primitives, components, composites, frames and the template catalog. It is a browsing and explanation tool; use the CLI library commands and the specimen gallery when making or validating deck choices.

## Open the board

```sh
pptxgengo paths
pptxgengo docs
# Optional: bind a different interface/address.
pptxgengo docs --addr localhost:8790
```

When the optional site exists at `paths.design_docs`, `pptxgengo docs` serves it at `localhost:8787` by default. The normal process stays in the foreground; stop it with Ctrl-C. Use `pptxgengo paths` to find `design_docs` and `design_docs_source`. The current template archives omit this optional site; use the packaged gallery and template browsing deck, or pass `--dir PATH` for an available built site.

The board's sections serve different questions:

- **Foundations:** colors, type styles, grid, spacing and number formats.
- **Primitives:** the basic drawing and text elements.
- **Components:** reusable card, chart, table and other slide elements, with variants and usage rules.
- **Composites:** arrangements of components that solve a recurring content pattern.
- **Frames:** slide-level rails, footers, title/body zones and split layouts.
- **Templates:** the complete one-slide patterns, searchable and filterable by family and metadata.
- **Catalog:** a compact inventory of each template's purpose, frame, tier, nodes, slots and content budget.

Use the board to understand terminology, compare options and see examples. For an actual deck, search the installed library and inspect the selected template:

```sh
pptxgengo design library-find --query 'phase roadmap' --kinds template --limit 10 --summary
pptxgengo design library-inspect --id lifecycle/three-phases --summary
pptxgengo design library-find --kinds component --query card --summary
```

Use `pptxgengo catalog --design-system --open` for the separate native specimen gallery. It contains the accepted PowerPoint previews for the packaged source templates. That gallery and the docs board have different jobs: docs explain the upstream system; the gallery shows source specimens that passed native review. Neither the docs examples nor an accepted source specimen qualifies arbitrary replacement copy. Inspect the template contract and fit and review your supplied content in its built deck.

## Check the source pin

The docs site is a frozen publication, not a live view of upstream. Its `SOURCE.json` records the exact upstream commit and the counts used to build the board. Read the file at the `design_docs_source` path from `pptxgengo paths` and compare its source commit with the current bundle metadata before treating the board as synchronized with the native library. The docs server also exposes the same receipt at `/api/source` while it is running.

The current bundle's source catalog has 649 templates. Use the installed bundle
and project lock to resolve the source revision; retain those pins when resuming
an older project. The optional docs publication must match that source before it
is used as a reference. Documentation examples do not establish fit or native
acceptance for a supplied-content deck.

V11 supports slide-level Comfortable, Compact and Dense typography. Read
[Typography density](typography-density.md) when adjusting copy capacity or
interpreting fitting warnings. A template named “dense” is an authored layout;
its name does not override the slide density setting.
