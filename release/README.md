# Local release 0.1.0-local.7

The current release combines the native Go presentation compiler, the resumable YAML project runtime, the West Monroe skill pack, and a hash-pinned unified SQLite catalog.

## Current capabilities

- One human-editable `deck.yaml` references persistent shared templates or deck-local compositions.
- Strict source validation, custom/shared assets, reproducible builds, provenance reports, output drift protection, scoped approvals, and portable project exports.
- Native SQLite discovery across modern definitions and legacy inventory, with detailed inspection and actual-content alternatives.
- 248 source-pinned design definitions (247 active), including 81 additions and revised `about/glance`; native logo, map, team-curve, device, and plane adapters.
- Automatic image deduplication, conservative photo resizing, ZIP compression, and shared frame layouts.
- A progressively disclosed skill with content, source, selection, voice, and independent review references.

Run `pptxgengo paths` to locate the installed example, skill, catalog, and library. Start with `pptxgengo design project init --help` or the packaged `examples/deck-project/README.md`. Search with `pptxgengo design library-find --query 'weekly status'`; inspect and fit supplied content before delivery.

## Evidence and limits

The current gallery records native preview and review state for each illustrated source specimen. These states qualify those specific specimens; supplied content requires its own fit and review. Historical v2 paired specimens remain labeled v2. The frozen v3 source is commit `e91e0d7771000b7386f1ea52f51252f0f0a134fd`.

Projects preserve ordinary IBM Plex font names. Native capture is calibration/review evidence; normal compilation uses the Go engine. Baseline drift protection is implemented; importing arbitrary PowerPoint edits back into YAML is a later workstream. Voice guidance uses local source snapshots until the live brand examples can be reviewed. Raster optimization conservatively retains unsupported or color-sensitive originals.

See `release/verification-wmds-v3.json` and the tracked implementation review for exact evidence and remaining limits.

## Historical release notes

# Local West Monroe presentation release

Version: **0.1.0-local.6**. This is a frozen local authoring release, not a published package.

## Changes from local.5 to local.6

The modern WMDS grid library is available through `pptxgengo design`, with 167
registered designs (166 active and one retained deprecated compatibility layout).
The separate `catalog --design-system` gallery shows native source/alternate
specimens, caller fields and editable composition examples. This route uses the
packaged v2 source and candidate Go typography engine by default. Source revision
and typography engine remain independently selectable; the standalone developer
`pptxdesign` default remains v1.

```sh
pptxgengo catalog --design-system --open
pptxgengo design library-catalog
pptxgengo design library-catalog --include-deprecated
pptxgengo design template --spec CONTENT.json --out NEW_DIRECTORY
pptxgengo design build --spec COMPOSITION.json --out NEW_DIRECTORY
pptxgengo design library-source-reference --bundle v1 --out OLD_REFERENCE
```

Generation uses Go font layout and packaged, hash-pinned artwork. No PowerPoint
capture or separate AppleScript runs during this route. Native review applies to
the recorded source/alternate specimens; arbitrary new copy still needs fit and
visual review. Original IBM Plex font names are retained. Exact native selection
among competing installed font files remains separately qualified.

The staged package includes its own fonts, calibration, frozen source, complete
artwork registry, content/composition examples, gallery and review receipts. See
`verification-wmds-v2.json` for this release's bounded evidence; the historical
`verification.json` remains evidence for its original version. No tests were
added or run in this WMDS refresh.

## Start

```sh
pptxgengo --version
pptxgengo catalog --open
pptxgengo catalog --templates --open
pptxgengo catalog --components --open
pptxgengo paths
pptxgengo template list
pptxgengo template components --id t068-ai-accelerator-004
```

The landing page points to the **166 active modern WMDS designs**,
**101 legacy source templates** and **132 source component occurrences**. The component gallery identifies 124
executable text/style subcontracts and eight visual-only groups. Use the
`west-monroe-presentations` Codex skill to choose a source layout or supported
composition route for a new deck.

The original 65 template implementations retain their prior review evidence.
The 44 requested exact source layouts (36 additions and eight existing contracts)
are registered from the UHG and AI Accelerator presentations with source-preserving native scenes, source values,
previews, named slots and explicit limits. These are fixed source layouts:
new copy, photos, rows, connectors and gauge states still require native fit and
visual review. The T045 semicircular gauge controls operate on its five source
freeform rows and discrete source pointer positions; the documented example
was natively rendered and visually reviewed. Neither the template nor
component counts imply arbitrary-content or variable-count qualification.

Template builds preserve source geometry and produce review decks grouped by
original presentation. For cross-source decks, use PowerPoint slide import or
copy after reviewing the generated slides. The installed release contains its
own library, source scene projects, selected original source PPTX files,
reviewed previews, exact gauge evidence, scripts and skill references.

## Changes from local.4 to local.5

- Measured composition accepts explicit installed font families, including IBM
  Plex Sans, with regular/bold/italic style resolution and variable font instance
  fingerprints. Missing fonts/styles fail before native measurement.
- All five adaptive builders apply the chosen `font_face` throughout their text,
  including footers, page numbers, pods, matrices and comparison labels.
- Cards expose optional `font_face`; existing cards default to Arial.
- The packaged `library/dynamic-components/fonts.json` example covers IBM Plex
  Sans styles and wrapping alongside Arial. See `fonts.md` for the workflow.

The Go test suite passes for this change.
`library/dynamic-components/fonts-proof.json` records the new bounded native
font example and its packaged evidence. Historical template and adaptive
example evidence retains its original environment and qualification scope.

## Changes from local.3 to local.4

- Separate searchable template and component galleries, with previews, controls,
  source identity and per-item limitations.
- Original source values, component-group discovery and T045 gauge controls in
  the installed CLI; the gauge build matches the reviewed native example.
- Five installed adaptive composition builders: process, roadmap, architecture,
  team and comparison. Their 32 reviewed examples have individual evidence;
  changing the content still requires native measurement and review.
- Compatibility restored for two historical template examples whose plain-text
  replacements omit embedded source line breaks. New or reordered control
  characters remain rejected.

## Paths and reproducibility

`pptxgengo paths` prints the installed release root and the library, scripts,
skill, landing page and all three gallery paths. The command
launcher resolves these paths from its own frozen release directory, so later
repository edits do not affect this version.

To create a reviewable package without changing global links:

```sh
python3 scripts/build-release-catalog.py
scripts/install-local-release.sh --stage-only /absolute/path/new-local-6-stage
```

The stage command checks all 101 template records, the 124 component contracts,
local gallery assets and skill references, then writes a SHA-256
`release-manifest.json`. Running `scripts/install-local-release.sh` without
arguments installs this version and switches the global command and skill
symlinks. The installer refuses to overwrite an existing version directory;
`0.1.0-local.4` remains available for rollback.

Required: Go 1.27.1 and Python 3. Native PowerPoint measurement additionally
requires macOS, PowerPoint and the fonts used by the source decks.

If earlier accepted sample decks have been removed by cleanup, explicitly set
`PPTXGENGO_PRIOR_RELEASE_ROOT` to a frozen installed release when staging. Only
missing legacy evidence is recovered, each file is checked against that release's
manifest and the original proof hashes, and `packaging-inputs.json` records it.
No decks are restored into the working samples folder. This is a package assembly
input; normal authoring uses the new package alone.
