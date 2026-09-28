# Local West Monroe presentation release

Version: **0.1.0-local.4**. This is a frozen local authoring release, not a published package.

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

The landing page points to two galleries: **101 supported source templates** and
**132 source component occurrences**. The component gallery identifies 124
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

## Changes from local.3

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
skill, landing page, template gallery and component gallery paths. The command
launcher resolves these paths from its own frozen release directory, so later
repository edits do not affect this version.

To create a reviewable package without changing global links:

```sh
python3 scripts/build-release-catalog.py
scripts/install-local-release.sh --stage-only /absolute/path/new-local-4-stage
```

The stage command checks all 101 template records, the 124 component contracts,
local gallery assets and skill references, then writes a SHA-256
`release-manifest.json`. Running `scripts/install-local-release.sh` without
arguments installs this version and switches the global command and skill
symlinks. The installer refuses to overwrite an existing version directory;
`0.1.0-local.3` remains available for rollback.

Required: Go 1.27.1 and Python 3. Native PowerPoint measurement additionally
requires macOS, PowerPoint and the fonts used by the source decks.
