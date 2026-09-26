# Capability survey

Survey date: 2026-09-25. Baseline: local `master`, commit `82b44057`, plus the Go version change described below. This is a code and workflow survey, not a visual certification of generated presentations. Remote branch observations use locally available refs.

## Reconstruction experiment update

The subsequent UHG experiment added `cmd/pptxscene`, `cmd/pptxdiff` and
`cmd/pptxanchor`, source-derived native JSON bindings, native PowerPoint phrase
measurements, image alpha bounds, and rendered QA controls. See the
[results and limitations](planning/RECONSTRUCTION_CHECKPOINT.md). The complete
authoring CLI, semantic deck schema, library database and skill pack remain
planned. The table below records the initial survey before this spike.

## Catalog execution update

The first catalog wave adds Python inspection commands for hash-verified source
occurrences, conservative layout deduplication, local/hosted asset manifests,
SQLite FTS5 search, explicit visual-family decisions and review previews. See the
[library checkpoint](library/README.md). These helpers do not yet provide approved
semantic layout contracts or an integrated Go authoring command. The historical
survey below remains a record of the pre-experiment baseline.

## Initial findings

This repository has a substantial Go presentation writer. It does not yet have the CLI or agent skill pack described in the product vision. Its historical objective was fidelity to PptxGenJS 4.0.1; the next objective should be dependable authoring and revision of branded, editable presentations.

The first product benchmark is a **detailed proposal or document-style deck**, per the user's choice. The user subsequently supplied EnableComp (35 slides), UHG (85 slides), and Graphics and Layouts (166 slides): see [the corpus inventory](planning/README.md). Such slides need to communicate without a presenter. Sparse live-presentation slides should be a separate profile.

| Need | Current state | Evidence and practical implication |
|---|---|---|
| Go generation engine | Implemented | `pptx/presentation.go`, `slide.go`, `writer.go`: presentation/slide construction and ZIP/OOXML output |
| Text and native objects | Implemented | `Slide.AddText`, `AddShape`, `AddTable`, `AddChart`, `AddMultiChart`; rich text, shapes, tables, editable charts with embedded workbooks |
| Images and media | Implemented, with compatibility gaps | `media.go`, `objects.go`, `xml.go`: local/remote media, SVG handling, sizing. SVG preview behavior intentionally preserves an upstream Node limitation; actual output needs renderer testing |
| Masters and theme | Partial design-system foundation | `DefineSlideMaster`, `DefineLayout`, theme heading/body fonts, backgrounds and slide numbers. Current `master` has no configurable theme color scheme; there is code for it on a newer branch |
| Font embedding | Implemented, limited validation | `fonts.go` accepts TTF/OTF data, checks basic SFNT headers, and writes font parts. This does not establish actual font availability, embedding permission, or application fidelity |
| Notes and sections | Implemented | `AddNotes`, `AddSection`; useful for source provenance and deck organization |
| CLI | Absent | No Go `main` package or `cmd/` tree found. Node demo commands are upstream demos |
| Declarative source | Absent | No deck YAML/JSON schema, parser, migrations, content bindings, or project format |
| Repo skill pack | Absent | No repository `SKILL.md` found; existing WM skills live separately in the user's skill directory |
| Narrative and brand voice | External guidance only | Existing WM skill calls for action titles, a claim plus evidence, and varied layouts. No code-backed narrative/evidence model or evaluated document-deck workflow here |
| Layout contracts and shape library | Absent | Generic geometry and shape presets exist; reusable components with named content zones, role-based fonts, capacity, and supported variants do not |
| Overflow and alignment | Partial primitives, no QA system | `TextPropsOptions.Fit`/`AutoFit` serialize rendering instructions. `tables.go:parseTextToLines` uses a character-width heuristic for paging; it does not measure actual rendered text. No general containment, collision, alignment, or typography checks |
| Vetted slide/layout library | Source corpus and survey inventory now present; agent-ready contracts absent | Supplied Graphics and Layouts has 166 slides; external skill has a distinct 179-slide reference. Content approval, capacity and reusable bindings require curation |
| Asset library integration | Absent in repo, reusable external inventory | The WM asset skill indexes 1,331 assets. No pinned asset lockfile, local content cache, visual metadata, or placement contracts here |
| Read/import existing PPTX | Absent | Production code writes packages; ZIP-reading helpers occur in tests. `GetTextContent` reads the in-memory slide model, not an imported file |
| Copy slides between decks | Absent | No dependency-aware package import for slides, layouts, masters, themes, charts, media, notes, etc. |
| Return colleague edits to YAML | Absent | No persistent semantic IDs, import classification, source mapping, reconciliation, or three-way merge |
| Automated visual QA | Absent | Existing tests compare XML and behavior; no checked-in rendered baseline or PowerPoint acceptance harness |
| Distribution | Library only | `go.mod` is present; README, TESTING, RELEASING and npm metadata are largely inherited from PptxGenJS. No Go CLI release workflow or versioned skill package |

## Useful work outside the current branch

Local `origin/fix/powerpoint-integration-review` and corresponding GitLab ref point to a follow-up commit, `a26d4712` (`fix: harden branded presentation rendering`). The inspected diff adds:

- Configurable theme colors and XML escaping of theme font names
- Intrinsic image aspect-ratio inputs for image sizing
- Public clock and section UUID providers
- Bounded media reads and rejection of nonregular local media files
- Tests for parts of those changes

These changes were inspected, not merged or tested in this survey. Reconcile them in the first implementation phase before duplicating the work. They do not add a CLI, importer, skill, or rendering QA. Public UUID injection also does not automatically establish complete reproducibility: `charts.go` still calls `getUuid` directly for scatter labels.

`REVIEW.md` contains an older issue list followed by a resolution log. Its original critical findings should not be reported as outstanding without checking the resolution log and current code. Its final state reports all findings resolved or documented as deliberate deviations. That review addressed port correctness, not the product capabilities above.

## Existing West Monroe resources

Inspected local resources:

- `/Users/rscott/.codex/skills/generate-west-monroe-slides/SKILL.md`
- Its `references/brand-foundation.md`, `references/content-spec.md`, `references/presentations-integration.md`, and Python generator/inspector
- `/Users/rscott/.codex/skills/wm-brand-assets/SKILL.md` and asset inventory/index

The slide skill has a two-slide blank/footer template and a 179-slide reference template. Both are 13⅓ × 7½ inches. Its quick generator offers eight types: title/body, two columns, cards, table, timeline, process, matrix, and section. Its polished workflow depends on the separate Presentations skill and subsequent WM frame assembly.

The installed skill's foundation specifies Arial; navy `070154`; blue `0047FF`; magenta `F900D3` as an accent; required logo, reproduction disclaimer, footer and page markers. It recommends 24–32pt titles, 12–16pt body text, 9–11pt captions, and approximately 0.55-inch side margins. Its content boundary is described as both 6.45 and 6.35 inches; exact frame geometry should settle it. The generator also uses 8.5pt for footer text despite a general 9pt floor; model approved frame text separately from body minimums. Subsequently inspected corporate guidance in `/Users/rscott/Documents/branding` confirms Arial for PowerPoint and adds precise voice, color, image and highlighter rules. Use those fuller sources to build the new brand pack; do not assume the older skill's abbreviated numeric guidance is authoritative.

The inspector checks dimensions, slide presence, certain placeholder strings, disclaimer text, and a crude title-length heuristic. It does not measure text fit, inspect pixels, validate alignment, or recursively account for every inherited/grouped text object.

The asset inventory was generated on 2026-06-04. Its 1,331 entries include 6 logos, 26 hand-drawn assets, 1,083 icon variants, and 196 photos. Names and paths support discovery, but the inventory lacks rich visual descriptions, dimensions, explicit image-specific usage rules and curated semantic tags. Counts include PNG/SVG and color variants, not 1,331 distinct concepts. No assets were bulk downloaded during this survey.

The existing slide skill prefers new body designs and permits cloning only for specifically requested slides. The requested product explicitly includes approved slide and layout reuse. The new pack should make reuse a supported authoring mode and migrate the older guidance deliberately; it should not inherit that restriction as a blanket product rule.

## Baseline verification and changes

- Updated `go.mod` from Go 1.24.7 to **Go 1.27.1**. This raises the minimum required Go version for library consumers as well as repository work.
- Confirmed Go 1.27.1 on the [official download page](https://go.dev/dl/). The local launcher originally selected 1.27.0 and downloaded 1.27.1 automatically after the module change, consistent with [Go toolchain selection](https://go.dev/doc/toolchain).
- `go version`: `go1.27.1 darwin/arm64`
- `go build ./...`: passed
- `go test -race -cover ./...`: passed; **78.1% statement coverage**
- Static inventory: 209 top-level `Test...` functions; eight golden presentation cases, including recursively compared embedded workbooks
- PowerPoint and LibreOffice application bundles are installed locally. Their rendering and automation paths were not exercised in this survey.

These checks establish a usable library baseline under the new toolchain. They do not prove attractive design, no PowerPoint repair prompts, correct SVG previews, font fidelity, or safe import/re-export. Those are explicit experiments in [PRODUCT_PLAN.md](PRODUCT_PLAN.md).

Survey outputs now also include `scripts/inventory-pptx.py`, 286 per-slide sidecars, three JSON inventories, and complete proposal/layout ledgers. The script was exercised on all three supplied decks. These are structural discovery artifacts, not an importer or renderer. The CLI, SQLite catalog and skill design remain proposed work.
