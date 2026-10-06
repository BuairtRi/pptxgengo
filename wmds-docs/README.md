# WM design system docs

The browsable documentation layer for the West Monroe slide design system: foundations, primitives, components,
composites, frames and the template library (with family, section, frame, tier and date filters), plus the catalog
and changelogs.

```sh
pptxgengo docs                      # installed package, http://localhost:8787
pptxgengo paths                     # design_docs and design_docs_source

# From a source checkout:
go run ./cmd/wmdsdocs                 # http://localhost:8787
go run ./cmd/wmdsdocs -addr :8787     # share on your network
```

## Where it comes from

`site/` is a static build published from the upstream design repo (`wm-design-system`) by
`python3 tools/build_docs.py --publish <this repo>`. `site/SOURCE.json` pins the exact upstream commit, its build
time and counts; `/api/source` serves it. Publishing refuses to run from an uncommitted upstream tree. Assets are
content-hashed (`site/assets/<hash>.<ext>`), so unchanged photos and logos keep the same files across publishes.

| File | What it is |
|---|---|
| `site/index.html` | The reference board (all tabs). |
| `site/catalog.json` | Every template: family, section, frame, tier, slots, budget, added/revised dates. |
| `site/changes.json`, `site/CHANGELOG.md` | Template and renderer changes since the pptxgengo snapshot, with PowerPoint implementation notes. |
| `site/web-CHANGELOG.md` | Web-contract changes (tokens, content vocabulary) for the web registry. |

## System of record (staged)

Template authoring still happens upstream; this repo becomes the system of record for the template library in
stages. Today the library reaches pptxgengo through the pinned, qualified intake bundles in
`planning/wm-design-contracts/` and the published `library/wm-design-system/` version. Publish these docs together
with each intake so the docs and the native library describe the same upstream commit (compare
`site/SOURCE.json` with the intake README's pinned commit).

## Package integrity

The global package includes the docs server and the static documentation. Installation verifies the clean source commit, embedded tokens, components, frames and catalog against the selected frozen native bundle, as well as catalog counts, changelogs and every content-hashed asset. A documentation publish alone does not qualify a new template for native PowerPoint; intake and native specimen review remain required.

Upstream documentation can advance while an intake is being qualified. `release/default-docs.txt` selects the matching frozen documentation snapshot for the installed release; without that file the installer uses `wmds-docs/site`. The source checkout keeps the latest upstream board here, while the global package serves the selected release snapshot at its usual `wmds-docs/site` path. The installer still requires an exact match to the native bundle and does not overwrite the latest upstream publication.

The documentation board may render many examples at once. During browser review, Chrome reported high memory use; lazy rendering is a follow-up performance improvement.
