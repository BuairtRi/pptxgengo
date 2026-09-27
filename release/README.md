# Local West Monroe presentation release

Version: **0.1.0-local.3**. This is a local authoring release, not a published package.

## Start today

```sh
pptxgengo --version
pptxgengo catalog --open
pptxgengo paths
pptxgengo template list
pptxgengo template inspect --id t020-graphics-and-layouts-062
```

In a new Codex conversation, invoke **`$west-monroe-presentations`** and give the
agent the source deck, the slides to change, and the intended content. The skill
has a short entry point and separate references for content, template editing,
composition, assets and native review. Existing Codex conversations may need to
be restarted to discover a newly installed skill.

Starter prompt:

> Use $west-monroe-presentations and the installed pptxgengo release. Read the
> current deck at [path]. Preserve its source. Help me develop [section/argument],
> select suitable templates or supported components, and generate new review
> slides. First establish each slide's role, takeaway, evidence and density.
> Retain the substantive detail of the proposal. Tell me which requested
> structural changes are supported and which require manual editing.

The Patterson and KKR source decks remain unchanged in `samples/`. The planning
briefs under `planning/release-0.1/` map their existing content to candidate
layouts. They are starting suggestions, not approved new copy.

## What is available

- One global compiled dispatcher with seven packaged tool binaries.
- A frozen local data snapshot, separate from future repository changes.
- 65 source-template contracts with named content slots, source previews,
  reviewed adaptations, and explicit limits.
- New slide composition using supported pods/teams, cards, grids/panels,
  canvas objects, images, rich text, and measured accents. Component schemas
  and maturity differ; see the skill's deeper references.
- An installed Codex skill and a local searchable HTML catalog.

Imported template geometry and item counts generally remain fixed. Changing the
number of rows, boxes, phases, connections or indicators requires a supported
component composition or additional adapter work. Current text replacement does
not certify fit for arbitrary new content.

The 65 reviewed examples comprise 60 direct native PowerPoint open/exports and
five accent geometry replays. Reopening the five generated accent PPTX files
remains pending. Newly authored slides still need native fit and visual review.

Template builds produce source-preserving review decks grouped by original
source. General cross-source merging into an existing proposal is not provided
by the scene tool; use PowerPoint's slide import/copy workflow when needed.

## Reproduce the local installation

Required: Go 1.27.1, Python 3, local source projects and review PNGs. Native
measurement additionally requires macOS, PowerPoint and the specified fonts.

```sh
python3 scripts/build-release-catalog.py
scripts/install-local-release.sh
```

The installer refuses to overwrite an existing version directory. Increment the
release version for later builds. Generated catalog assets and verification decks
are local artifacts; they are not committed as new binaries.

## Cleanup

PowerPoint reported zero open presentations at cleanup time. Forty obsolete
verification PPTX/PDF files were moved to
`samples/_archive/release-cleanup-20260927/`. `cleanup.json` records their original
paths and hashes. Original sources and current template review images were kept.
