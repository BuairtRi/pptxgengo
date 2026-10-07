# pptxgengo CLI reference

Deck work uses one route: `pptxgengo design`. Check the version with `pptxgengo --version`. The dispatcher fills in the packaged library, index, gallery and engine for discovery commands; existing `project` commands use the project's lock, except an explicit `project migrate` target. New projects use V11. Build/export/review output destinations must be new; source mutations operate on the existing project with retained predecessors.

Redirect `library-match` output to a file; it prints the full report. Use `--summary` with `library-find` for compact output.

## Release paths and galleries

```sh
pptxgengo paths
pptxgengo catalog --design-system --open
pptxgengo catalog --assets --open
```

- `paths` returns JSON with release locations, including `design_system_default`, `design_index`, `design_docs` and `design_docs_source`. Resolve packaged files from these; never hardcode a release directory.
- `paths.skill` identifies the skill snapshot bundled with the CLI release.
  The active installed skill can receive newer documentation independently.
  A documentation-only skill update does not change the executable or source
  pin and does not itself require deck migration.
- `catalog` takes one gallery (`--design-system`, `--templates`, which is the same page, or `--assets`) and `--print` (default) or `--open`.
- The old routes (`template`, `compose`, `scene`, `component`, `lib`, `anchor`, `diff`, `adapt`) have been removed and print a replacement hint.

### Browse upstream design-system documentation

```sh
pptxgengo docs
pptxgengo paths
pptxgengo catalog --design-system --open
```

- `pptxgengo docs [--addr localhost:8787]` serves the packaged static reference board; pass `--dir PATH` only to serve another built docs site. It covers foundations, primitives, components, composites, frames, template filters and a template catalog. See [design-system documentation](design-system-documentation.md).
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
| `project split --bundle v11` | Move each slide into its own file; notes into `notes/`; local templates into `templates/`; use the matching locked bundle for historical projects |
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
- **First macOS render:** PowerPoint may show a "Grant File Access" dialog for the staging folder. The operator must click Grant. Until then, renders fail within seconds with `file_access_denied … PowerPoint is showing Grant File Access`. Windows has its own setup, policy and Protected View diagnostics. Every failed render writes `render-error.txt` in `--out`.
- If a render fails, read `render-error.txt`, run `render-doctor`, ask the operator to clear what they report, then retry once.

## Inspect an existing PowerPoint

```sh
pptxgengo design source-inventory --in existing.pptx --out ./inventory [--project ./client-deck]
```

Writes `source_inventory.json` and `.md`: each slide's title, hidden state, text with positions, tables, charts, images and speaker notes. With `--project` it adds `mapping_report.json` listing source slides and project slides side by side, all `unmapped`; it never matches them for you. See [editing slides](editing-slides.md#rebuild-or-update-an-existing-deck).

## Move a project to a new CLI version

A project pins its executable and source. A changed release can report
`project.toolchain_drift`. When the operator has requested an upgrade:

```sh
pptxgengo design project migrate --project ./client-deck --dry-run
pptxgengo design project migrate --project ./client-deck
pptxgengo design project build --project ./client-deck
```

Migration defaults to V11 and keeps the existing engine unless `--engine` is
specified. It validates source/template compatibility and compiler fit before
changing the lock. A dry run or failed validation leaves the lock unchanged.
Success preserves exact old bytes as `<configured-lock>.pre-migrate-<hash>` and
atomically replaces the pin; repeated migration to the same toolchain creates
no extra backup. It does not rewrite copy, custom templates, revisions or PPTX.
Resolve reported custom ancestry/revision conflicts before retrying. Rebuild,
render and review the new result; prior approvals/native coverage become stale.
This operates on YAML projects; it does not convert a standalone PPTX to YAML.

## One-slide routes and maintenance commands

- Single slides outside a project: [design-system one-slide routes](design-system-authoring.md).
- Library-maintenance commands (`library-index`, `library-search`, `library-sweep`, `library-bound-sweep`, `inspect`, `templates`, the `*-reference` commands, `typography-probes`) are not for deck work.

## Explicit native card pilot in later source builds

`wmds/component/editable-card` places two plain title/body paragraphs in one
filled native rectangle. Use explicit distinct bindings in a local template;
the source `subhead` and `body` roles remain separate. Stock cards are unchanged.
`project editability` inventories both paragraph fields, and bounded text
reconciliation can review/adopt each role independently. Markup, explicit line
breaks, changed role formatting/topology and ambiguous bindings require review.
Geometry remains manual. The component is an opt-in measured design with pending
Mac/Windows selection, move/resize and visual qualification; stable v4.1.0
does not include it. Authoring and limits are in `docs/native-editable-card.md`.

## Reuse authored finished slides in the next source build

`project slide publish --id SOURCE --library-id curated/slide/KEY --revision N
--name NAME --purpose PURPOSE --owner OWNER --out NEW-DIR` creates a closed draft
revision from supported supplied-content shared source. `project slide insert
--package REVISION-DIR --id FRESH-ID --rationale REASON` inserts an independent
copy with composition lineage; draft work requires `--allow-draft`. Both accept
`--project`, `--bundle` and `--engine`; insertion accepts `--before`, `--after`
and `--into-section`. Exact pins and existing authored composition are required.

`library-index --slide-library ROOT` adds closed revisions; `library-find
--kinds finished-slide` distinguishes them from templates. Read lifecycle,
owner, approval/freshness metadata and previews before reuse. Claims and local
or unsupported typed dependencies currently fail explicitly. Stable v4.1.0
lacks these commands; full scope is documented in `docs/finished-slides.md`.

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
