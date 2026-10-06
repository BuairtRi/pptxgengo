# Design-system documentation

Use the upstream design-system docs board to learn the system's vocabulary and see how its parts fit together. The board covers foundations, primitives, components, composites, frames and the template catalog. It is a browsing and explanation tool; use the CLI library commands and the qualified specimen gallery when making or validating deck choices.

## Open the board

```sh
pptxgengo paths
pptxgengo docs
# Optional: bind a different interface/address.
pptxgengo docs --addr localhost:8790
```

`pptxgengo docs` serves the packaged static site at `localhost:8787` by default. The normal process stays in the foreground; stop it with Ctrl-C. Use `pptxgengo paths` to find `design_docs` and `design_docs_source`. Pass `--dir PATH` only when intentionally serving another built site.

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

The v10 stock bundle contains 649 templates in 29 families, including 29 workshop layouts. The released docs and native library share upstream commit `c14fb286fb38e15800a6fd476a1ed67956f1165f`. Installation rejects a docs publication whose source or assets differ from the selected bundle. Do not change the frozen source library based on a docs-only example.

The latest upstream documentation in a source checkout may describe work that is still being qualified. The installed package uses the frozen documentation selected for its native bundle. Three typography density presets remain outside the v10 template intake; a stock template whose name includes “dense” is a separate authored layout, not a slide-level density setting.
