# Capability survey

Historical survey baseline: 2026-09-25, local `master` at `82b44057`, plus the Go version change described below. The original capability table records that snapshot and is not a current inventory. Neither the survey nor the updates below constitute visual certification of generated presentations.

## Current repository state — 2026-10-08

The repository is now at v4.2.1 (`66229bd1`). `pptxgengo design` provides
template search and inspection, project creation/build/edit, density controls,
content matching, native rendering/review workflows, and portable packaging.
The published default is WMDS v11: 649 templates, 648 active, with template
sources, fonts, gallery and SQLite index included in template-only archives.
New projects default to `native-v1` where a template's measured records support
safe native text emission. This is an implementation default; v11 native
rendering and visual qualification remain pending. See [release status](docs/release-status.md)
and the [documentation index](docs/README.md).

The Go authoring system, skill pack, SQLite catalog and project format described
as absent in the original survey are now implemented to bounded scopes. The
original rows below retain the 2026-09-25 baseline; read them as historical
findings, not current product claims. See [testing lanes](docs/testing.md) for
current CI ownership and qualification boundaries.

## Historical authoring checkpoint — 2026-09-26

The latest `pptxcompose` path includes measured numbered/metric cards, native
canvas text/surfaces/rules, pinned PNG/JPEG imports, narrative notes, single-line
highlight/underline anchors, a complete fixed-zone `fit-report`, and bounded
`recover-text --text-only` for known generated decks. Explicit measurement reuse
requires the entire probe contract to remain unchanged. The repository now has
an [agent skill pack](skills/pptxgengo/SKILL.md).

The user found the first 11-slide showcase too sparse. The follow-up
[dense proposal benchmark](library/showcase/README.md) uses EnableComp 3/4 and
UHG 28/38/43 as structural references. All five final slides passed native verification
and visual review: 318 objects, 258 text blocks, and zero fixed-zone fit failures.
The [proof](library/showcase/dense-proof.json) also records a five-slide scene
roundtrip and recovery of one actual PowerPoint text edit. Dense content can be authored with the
current native primitives. Measured grid/panel contracts are now implemented;
rich diagram geometry remains a gap. See the [reference benchmark](planning/COMPLEXITY_BENCHMARK.md)
and [implementation review](planning/DENSE_LAYOUT_GAPS.md). Historical survey
entries below describe their original checkpoint, not current absence of those tools.

## Measured layout update — Wave 1

`pptxcompose` now expands named grids and nested panels using native measured
text heights, shared row boundaries, explicit gutters/padding and stable slot IDs.
Layer ordering keeps surfaces behind their content. `fit-report` identifies all
cells that exceed bounded row heights. Known-deck text recovery maps generated
layout text back to the named block.

The local measurement cache revalidates original native evidence and binds text,
width, styling and the recorded PowerPoint/font/tool environment. One changed
contract creates one new probe; unchanged contracts can be reused across slides.
It does not replace final native verification or visual review. See the
[layout contract and fixtures](library/layout-components/README.md) and
[checkpoint](planning/WAVE1_CHECKPOINT.md) for acceptance status and limitations.
The [six-slide proof](library/layout-components/proof.json) covers 423 objects and
298 text blocks; all 174 control glyph bounds exactly match the prior dense deck.
Uniform-style Arial blocks and explicit bounded tracks remain the supported scope.

## Reconstruction experiment update

The subsequent UHG experiment added `cmd/pptxscene`, `cmd/pptxdiff` and
`cmd/pptxanchor`, source-derived native JSON bindings, native PowerPoint phrase
measurements, image alpha bounds, and rendered QA controls. See the
[results and limitations](planning/RECONSTRUCTION_CHECKPOINT.md). At this
checkpoint, the complete authoring CLI, semantic deck schema, library database
and skill pack remained planned. The current-state note at the top supersedes
that status. The table below records the initial survey before this spike.

## Catalog execution update

The catalog waves add Python inspection commands for hash-verified source
occurrences, layout deduplication, local/hosted asset and font manifests, resolved
group/placeholder geometry, component candidates, SQLite FTS5 search, and explicit
visual-family decisions. The current 369-slide corpus has 306 classification work
units, 24 shared layout families, 244 selected component examples in 74 source patterns,
31 semantic component families, six proposed style profiles and 17,451 indexed items. See the
[library checkpoint](library/README.md). These helpers do not yet provide approved
semantic layout contracts or an integrated Go authoring command. The historical
survey below remains a record of the pre-experiment baseline.

## Reference and component contract update — 2026-09-26

A [10-reference shortlist](library/reference-variants.md) adds visual comparison,
use-case descriptions, full-slide context and preference export. The new
`cmd/pptxcomponent inspect|apply` edits explicitly bound text slots in four pinned
source scenes. Four normal fixtures were rendered in native PowerPoint; 14 text
segments stayed within their outer frames. A long-copy fixture visibly overflowed
and native measurement detected it. See the
[proof report](library/component-contracts/proof-report.json). No component has
been promoted to general adaptation approval; automatic fit gating, effective
style resolution and cross-slide composition remain pending.

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

At the initial survey checkpoint, outputs included `scripts/inventory-pptx.py`,
286 per-slide sidecars, three JSON inventories and proposal/layout ledgers.
Those were structural discovery artifacts. The CLI, SQLite catalog and skill
design were proposed then; they now exist in bounded forms described above.

## Dynamic role and pod composition (2026-09-26)

`cmd/pptxcompose` adds a bounded new-content path: `probe`, `measure`, `build`
and `verify`. It composes native editable role rectangles into variable-role pods,
selects fitting columns deterministically from native text measurements, checks
semantic colors/contrast, and returns geometry anchors. Final native QA compares
text, frames, typography, fills, foregrounds and margins.

See [workflow and limits](library/dynamic-components/README.md) and the
[fixture proof report](library/dynamic-components/proof-report.json). Explicit
Arial styles are supported; arbitrary source typography inheritance and broad
component adaptation remain future work.

The [team extension](library/dynamic-components/team.md) adds standalone roles,
phase membership and padded containment, a semantic staffing legend, and routed
orthogonal reporting lines. Native QA also checks stroke color, weight, opacity
and arrowhead absence. Lines are editable segments, not PowerPoint-glued
connectors. The two-slide fixture and negative cases are recorded in the
[team proof](library/dynamic-components/team-proof.json).
