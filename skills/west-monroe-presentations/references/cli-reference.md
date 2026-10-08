# pptxgengo CLI reference

Deck work uses one route: `pptxgengo design`. Check the version with `pptxgengo --version`. The dispatcher fills in the packaged library, index, gallery and engine for discovery commands; existing `project` commands use the project's lock, except an explicit `project migrate` target. New projects use V11. Build/export/review output destinations must be new; source mutations operate on the existing project with retained predecessors.

Redirect `library-match` output to a file; it prints the full report. Use `--summary` with `library-find` for compact output.

## Release paths and galleries

```sh
pptxgengo paths
pptxgengo catalog --design-system --open
pptxgengo catalog --assets --open
```

- `paths` returns JSON with release locations, including `root`, `design_system_default` and `design_index`. Resolve files from these paths. Optional `design_docs`, `project_example` and `skill` paths can be reported even when that archive omits those resources; check that they exist before use.
- The template archives include the source bundle, fonts, SQLite index, specimen gallery and `browsing/template-library.pptx` under `paths.root`. The skill and optional upstream docs site are installed separately. Documentation updates do not change a project's executable pin.
- `catalog` takes one gallery (`--design-system`, `--templates`, which is the same page, or `--assets`) and `--print` (default) or `--open`.
- The old routes (`template`, `compose`, `scene`, `component`, `lib`, `anchor`, `diff`, `adapt`) have been removed and print a replacement hint.

### Browse upstream design-system documentation

```sh
pptxgengo docs
pptxgengo paths
pptxgengo catalog --design-system --open
```

- If `paths.design_docs` exists, `pptxgengo docs [--addr localhost:8787]` serves that static reference board. Otherwise use the packaged specimen gallery or pass `--dir PATH` for an available built site. It covers foundations, primitives, components, composites, frames, template filters and a template catalog. See [design-system documentation](design-system-documentation.md).
- `paths` reports `design_docs` and `design_docs_source`. The latter points to the frozen `SOURCE.json` receipt; compare its source commit with the current bundle metadata before assuming both artifacts describe the same upstream revision. While running, the docs server also provides `/api/source`.
- `docs` explains upstream concepts and examples. `catalog --design-system` opens the separate native-reviewed specimen gallery for the installed bundle. Docs examples do not qualify replacement copy; fit and review supplied content in the actual build.

## Find and understand templates

```sh
pptxgengo design library-find --query 'phase roadmap' --kinds template --limit 10 --summary
pptxgengo design library-find --structures sequence --items 4 --kinds template --summary
pptxgengo design library-inspect --id lifecycle/three-phases --summary
pptxgengo design library-authoring --template lifecycle/three-phases
pptxgengo design library-preview --id lifecycle/three-phases
```

- Always pass `--kinds template` when looking for layouts; otherwise icons and components mix in.
- Hints: `--query`, `--roles`, `--structures`, `--visual-forms`, `--items` with `--item-role`. `--limit` 1–100. `--include-deprecated`. `--namespace wmds`.
- Results with no match are dropped; `--include-weak` lists everything for browsing. Words aren't split ("heat map" works, "heatmap" doesn't).
- `--summary` returns purpose, content groups and verified `screenshot_paths`. Open the screenshots.
- `library-authoring --template KEY` lists every slot with a readable alias, description, whether it is decorative, and an approximate capacity (characters and lines) where supported. Estimates follow the template's authored typography; unsupported internals are explicit. Whole-slide density and actual copy can change the result. Without `--template` it reports library-wide coverage and templates whose names remain ambiguous.

### Select several specimen templates from a file

```sh
pptxgengo design library-catalog --template-keys-file chosen-templates.txt
pptxgengo design library-source-reference --template-keys-file chosen-templates.txt --out ./stock-reference
```

`--template-keys-file` supports one key per line, comma-separated keys, or a JSON
string array. Catalog/reference/sweep routes share this selection; references
preserve the requested key order. Empty input, duplicates and unknown keys fail.
Use either this flag or `--template-keys`; explicit keys cannot be combined with
`--family`. It does not apply to `library-match`, which uses `--templates`.

## Match page content to templates

```sh
pptxgengo design library-match --page page.yaml --limit 4 --out ./candidates > ./candidates.log
```

See [template selection](template-selection.md#match-page-content-automatically) for the page format and what it can match today. Flags: `--page`, `--out` (new), `--templates k1,k2`, `--limit`, `--render`, `--timeout`. Writes ready slides and `needs_copy` drafts. Exits 1 when nothing is ready; the report and drafts stay in `--out`.

Ready V11 candidates can use automatic body fitting. Review warnings on stderr
and candidate layout reports; stdout remains the machine-readable match report.

## Find assets

```sh
pptxgengo design library-find --kinds asset --asset-kind photo --query 'working session' --summary
pptxgengo design asset-gallery --out ./gallery --kind icon --query risk
```

See [assets](assets.md).

## Compare layouts with real content

```sh
pptxgengo design library-fit --spec alternatives.json --out ./candidate-review
```

## Deck projects

| Command | Does |
| --- | --- |
| `project init` | Pin the executable, engine, library and fonts in `toolchain.lock.json` (only when no lock exists) |
| `project migrate [--bundle v11\|PATH] [--engine ENGINE] [--dry-run]` | Validate target compatibility and fit; preserve old lock and atomically update the pin |
| `project split --bundle v11` | Move each slide into its own file; notes into `notes/`; local templates into `slides/templates/`; use the matching locked bundle for historical projects |
| `project check` | Validate source, bindings, pins, composition log and claim references |
| `project build` | Build a new immutable directory under `builds/`; enforces text fit |
| `project titles [--format json]` | List slide titles in order with ID, hidden state and template |
| `project measure --report FILE [--slides IDS,RANGES]` | Estimated text collisions and panel spills from a build's layout report |
| `project status` | Stage, invalidated approvals and per-slide native review coverage |
| `project resume` | Save the recomputed state |
| `project approve --stage S --actor NAME [--slides IDS]` | Record an operator approval |
| `project scaffold --stock --template KEY --id ID --out FILE` | New slide file on a shared template, with alias and capacity comments |
| `project scaffold --bundle v11 --template KEY --reason R [--omit-nodes IDS] --out FILE.json` | Local derivative of a source-scene template; match the project's lock |
| `project slide add\|move\|remove\|hide\|show --id ID [--as NEW-ID] [--before\|--after ID] [--into-section S] [--reanchor] [--check-fit]` | Slide operations; see [editing slides](editing-slides.md) |
| `project slide draft-review set\|show\|clear --id ID` | Slide-owned status, label, color, owner and internal notes; see [Draft Review Notes](draft-review-notes.md) |
| `project swap --slide ID --template KEY [--apply [--allow-unmapped]]` | Propose a template change; `--apply` only when nothing is missing; `--allow-unmapped` deletes leftover copy |
| `project edit --patch FILE [--check-fit]` | Replace slide fields by stable slide ID |
| `project asset add --id ID --file IMG --description TEXT [--focus x,y]` | Register an image in the project |
| `project section list\|add\|rename\|remove` | PowerPoint sections and divider slides |
| `project detach --slide ID --as NEW --reason R` | Turn a slide's shared template into a local template |
| `project fork --template ID --as NEW --slides IDS --reason R` | Copy a local template for some slides |
| `project attach-render --render DIR [--decisions FILE]` | Attach a signed PowerPoint render (and review decisions) to the current build |
| `project review --stage outline\|content\|deck [--audience] --out DIR` / `project view [--stage S] [--audience] --out DIR` | HTML review packet: `--audience` for independent reviewers, without it for the author and operator (see [review packets](review-packets.md)) |
| `project export --mode client\|maintainer\|offline\|reviewer --out NEW.zip` | Package the current build |

Project commands take `--project PATH`, except `scaffold` (writes `--out` or stdout) and `measure` (reads `--report`). `project review` without `--stage` produces the older technical reviewer ZIP.

## Density and text measurement

V11 slide YAML supports `density`, `header_density` and `auto_density` outside
content. Read [typography density](typography-density.md) for role scales, the
independent Comfortable header default, automatic changes and source ceilings.
Density fields are edited directly in YAML; `project edit` patches do not accept
them. `project build` and `library-match` report automatic body changes on stderr
and in layout reports.

```sh
pptxgengo design measure-style --style small --scope cell --density dense \
  --text 'A longer table value' --width 180
pptxgengo design measure-style --style title --scope header --density compact \
  --text 'A title that needs two lines' --width 700
```

Required flags are `--style`, `--text`, and positive `--width` in points.
`--scope` is `body` (default), `header` or `cell`; `--density` accepts the three
tiers. Omitted density in the default body scope uses the source's base role;
header/cell scope without density uses Comfortable. For an older source, pass
its `--bundle`; `--engine` and a hash-matching `--source` override are available.
The result includes resolved fonts, sizes, leading, line advances and measurement
provenance. Its status is `measurement_only`, with native fit `not_evaluated`.
It neither proves template fit nor expands a template's title/body allocation.

## Render through PowerPoint

```sh
pptxgengo design render-doctor --json
pptxgengo design render --pptx ./client-deck/builds/<build-id>/deck.pptx --out ./review-1 \
  --png --pdf --slides 3,5-7 --contact-sheet
```

- On macOS, `render-doctor` checks PowerPoint, the GUI session, the staging folder, automation, PDFKit and file access. The Windows preview uses PowerShell COM to probe actual PDF/PNG export and font availability; read [Windows workflows](windows.md). A pass doesn't guarantee the next render succeeds.
- `render` flags: `--pptx`, `--out` (new), `--pdf` and/or `--png`, `--slides` (pages or ranges), `--contact-sheet`, `--include-hidden`, `--staging-dir` (or `PPTXGENGO_NATIVE_STAGING`), `--timeout` (default 5m).
- Output: a signed `render-manifest.json`, `native-pages/slide-NNN.png` (numbered by source slide), the PDF and the contact sheet. Never edit the manifest; `attach-render` rejects anything not signed by a render on this machine.
- It works on a temporary copy and never changes the source deck.
- On macOS, reuse one staging folder that PowerPoint can access (for example a dedicated folder under Documents) through `--staging-dir` or `PPTXGENGO_NATIVE_STAGING`. Pass the same location to `render-doctor`; files are staged directly there with unique task names.
- **First macOS render:** PowerPoint may show a "Grant File Access" dialog for the staging folder. The operator must click Grant. Until then, renders fail within seconds with `file_access_denied … PowerPoint is showing Grant File Access`. Windows has its own setup, policy and Protected View diagnostics. Every failed render writes `render-error.txt` in `--out`.
- If a render fails, read `render-error.txt`, run `render-doctor`, ask the operator to clear what they report, then retry once.

## Inspect an existing PowerPoint

```sh
pptxgengo design source-inventory --in existing.pptx --out ./inventory [--project ./client-deck]
```

Writes `source_inventory.json` and `.md`: each slide's title, hidden state, text with positions, tables, charts, images and speaker notes. With `--project` it adds `mapping_report.json` listing source slides and project slides side by side, all `unmapped`; it never matches them for you. See [editing slides](editing-slides.md#rebuild-or-update-an-existing-deck).

## Move a project to a new CLI version

A project pins its executable, engine and source. Follow
[project upgrades](upgrading-projects.md) to preserve the predecessor, choose an
explicit bundle target, inspect template revisions and review the new baseline.
`project migrate` validates and updates the toolchain lock; it does not convert
schemas, rebase local templates or enable a native editing profile. Omitting
`--bundle` selects V11, even for older projects.

## One-slide routes and maintenance commands

- Single slides outside a project: [design-system one-slide routes](design-system-authoring.md).
- Library-maintenance commands (`library-index`, `library-search`, `library-sweep`, `library-bound-sweep`, `inspect`, `templates`, the `*-reference` commands, `typography-probes`) are not for deck work.

## Native editing and maintained slides

New v2 projects persist `editing_profile: native-v1`. Direct template and
browsing-library generation also default to native-v1; use `--editing-profile
stock` on those routes when the original object structure is required. Existing
project profiles stay pinned. In a project, change `deck.yaml` deliberately and
build a new baseline before reviewing the new structure.

Eligible flat lists, including supported bold lead/body runs, become native
bullet paragraphs in one box. Eligible simple cards combine title/body content
and background in one editable shape. Native tables retain rows, columns and
cell editing while redundant outer wrappers are removed. Decorated cards,
nested/ordered lists and unsupported artwork retain their source structure;
inspect the layout report's retention reasons.

For a custom local template, `wmds/component/editable-card` and
`wmds/component/editable-list` expose explicit bounded native components. Inspect
the component's schema before use: their restrictions differ from automatic
profile conversions. Added text can require resizing a box. Combined card/list
source adoption is manual; keep baseline and edited copies, review and rebuild.

Use `project editability` and the [editing workflow](editing-slides.md) for
object ownership and supported text reconciliation. Actual native rendering and
visual review remain required for the deck you are authoring.

For operator-provided authored slide revisions, use `project slide publish`,
`project slide insert` and `project slide review-reuse`; inspect each command's
help and exact revision metadata. Use `library-find --kinds finished-slide` only
against an index built with that supplied slide library. No reusable-slide
inventory or content-complete browsing deck is shipped. See
[maintained authored slides](editing-slides.md#reuse-maintained-authored-slides).

## Portable projects, numbered versions and complete shares

| Command | Purpose |
| --- | --- |
| `project create --out NEW_DIR --id ID --title TITLE --bundle v11 --template KEY` | New pinned project using stable slide filenames; replace scaffold examples. |
| `project layout --project PATH --dry-run` / `--apply` | Preserve expanded source while migrating referenced paths to slides/ and slides/templates/. |
| `project asset revise --project PATH --id ID --file IMAGE --description TEXT --expect-sha256 HASH` | Select a new immutable deck-owned asset revision; retain old objects. |
| `project version save --project PATH --actor NAME --message TEXT` | Commit complete source/generated deck snapshot, exact shared asset revisions and receipts. |
| `project version list --project PATH` | Show verified contiguous history and current version. |
| `project version verify --project PATH --number 000001` | Check snapshot, asset objects and shared build history hashes. |
| `project version materialize --project PATH --number 000001 --out NEW_DIR` | Restore a standalone working root without changing the original or its pins. |
| `project version recover --project PATH --number 000002 --expect-current-sha256 HASH` | Explicitly publish a verified interrupted child snapshot; never force a lock. |
| `project share --project PATH --out NEW_ZIP` | Complete private colleague handoff with all versions/history and one stored object per asset hash. |
| `project share-extract --archive ZIP --out NEW_DIR` | Validate ZIP and safely expand legacy asset aliases to ordinary relative paths. |
| `project share-verify --project PATH` | Verify extraction inventory and numbered version history before deliberate edits. |

Complete shares contain private context/evidence, unlike client exports.
Keep generated baselines immutable; reconcile native working copies separately.
Numbered source snapshots use a shared deck-owned asset store, not filesystem
links or a machine-wide cache. `versions/current.json` binds the latest immutable
manifest. Locks, conflicting OneDrive copies, placeholders, changed predecessors
and nonportable names are diagnosed rather than silently resolved. Sharing
retains executable/OS/architecture pins and branding/font requirements; explicit
toolchain migration and offline runtime export remain separate commands.

## Diagram editing (geometry development build)

`project diagram inspect|patch|connect|arrange` uses `--project` and `--slide`.
Mutations preview by default; `--apply` commits validated source changes.
`project reconcile propose --geometry` also proposes native transforms and
existing-object paint order. See [architecture and geometry](architecture-geometry.md)
for source fields, examples, supported targets, reset behavior and exclusions.
These commands are not in the promoted v4.2.1 binary.
