# Project structure and repairing drift

Use one portable project root for a deck. `deck.yaml` identifies the active
source and slide order; the paths it references are authoritative. Stable slide
and template IDs survive file moves and reordering. Never choose the current
deck by filename, filesystem modification time or the highest build directory.

## Intended layout

```text
client-deck/
  deck.yaml                         # slide order, source paths, asset IDs
  toolchain.lock.json               # exact executable/engine/library pins
  composition-log.yaml              # authored layout choices by slide ID
  state.json                        # CLI state, current build and approvals
  slides/
    situation.yaml                  # one YAML file per stable slide ID
    architecture.yaml
    templates/
      client-architecture.yaml      # project-owned local definitions
  assets/
    objects/sha256/<content-hash>    # immutable bytes shared by every version
  context/
    project.md                      # stage, decisions and next action
    audience-context.md
    outline.md
    claims.md
    win-strategy.md                  # internal; link as win_strategy
    decisions.md                    # authored decision log
  sources/
    index.md                        # source IDs, authority and limits
    <source-material>
  briefs/<slide-id>.md
  notes/<slide-id>.md
  reviews/<review-id>/               # findings and actual review decisions
  working/<build-id>-edited.pptx     # editable copy, never a build baseline
  renders/<render-id>/               # generated native review evidence
  builds/<build-id>/                 # immutable CLI-generated baseline
    deck.pptx
    layout-report.json
    object-map.json
    receipt.json
  decisions/                        # CLI receipts and retained predecessors
    sources/<source-hash>/
    reconciliations/<adoption-id>/
  versions/
    current.json                    # CLI-maintained latest snapshot pointer
    000001/
      source/                       # captured source/context/pins/receipts
      deck.pptx
      manifest.json                 # exact shared asset/build references
```

`project create` makes the initial deck, slide, context/project.md, toolchain
lock and the slides/templates, asset-object and versions directories. Other
files appear when authored or produced by their command; the tree is a
convention, not a claim that an empty project contains everything above.
Working-copy and render directory names are operator conventions; record which
build each copy belongs to. Put outgoing ZIPs outside the project root.
Complete shares and snapshots include additional retained project files, so
keep unrelated downloads and experiments elsewhere.

Create supporting files as their stage needs them. Link them explicitly in
`deck.yaml`, for example:

```yaml
context:
  project: context/project.md
  audience: context/audience-context.md
  outline: context/outline.md
  claims: context/claims.md
  win_strategy: context/win-strategy.md
  decisions: context/decisions.md
  sources: sources
  composition_log: composition-log.yaml
```

`context.decisions` points to an authored file, not the CLI-owned `decisions/`
folder. Slide `brief` and `notes_file` paths are relative to the project root.
Use lowercase stable filenames without Windows-reserved names or filesystem
links. Do not renumber slide filenames merely because slide order changes.

## Assets, versions and edits

Register assets with `project asset add`; change selected bytes through
`project asset revise --expect-sha256 OLD_HASH`. All versions share one
**deck-owned** asset store. Each snapshot records exact hashes; never copy all
asset bytes into every numbered version or overwrite old objects.

Create, inspect and restore versions through `project version save`, `list`,
`verify` and `materialize`. Do not hand-create numbered folders or edit
`versions/current.json`. A snapshot is a committed source/deck state, while
`builds/` holds generated baselines. A PowerPoint working copy is a separate
edited artifact until its reviewed changes are adopted into source.

## Repair an existing project's organization

1. **Inventory.** Read `deck.yaml`, its referenced slides/local templates/assets,
   the lock and state. Run `project status` and `project check`. Preserve every
   edited PowerPoint copy and identify its receipt-backed baseline. If paths are
   missing, recover the exact referenced files from retained copies or a
   verified version before applying layout migration; it must load the project.
   Do not guess mappings from similar names or copy an arbitrary deck.yaml.
2. **Preserve.** Work on a complete project copy when doing substantial repair.
   If the current source builds, save a numbered version before reorganizing.
   Keep exact old files and the matching runtime; do not fabricate state or
   rewrite historical manifests to make them match the new paths.
3. **Preview slide/template normalization.**

   ```sh
   pptxgengo design project layout --project ./client-deck --dry-run
   ```

   Inspect the planned destinations. This command changes active slide
   references to `slides/<stable-id>.yaml` and local template references to
   `slides/templates/<stable-id>.yaml`, including inline definitions. It
   verifies unchanged expanded source and refuses occupied destinations with
   different bytes. It does not normalize assets, context, notes, working copies
   or every folder in the project. Identical expanded copy is not proof that
   prior source-file approvals remain current.
4. **Apply the reviewed layout.**

   ```sh
   pptxgengo design project layout --project ./client-deck --apply
   ```

   Old source paths and source preimages remain available to history. Treat
   unreferenced old files as retained predecessors, not a second active source;
   do not delete them as routine cleanup. The CLI does not promise removal of
   every old folder.
5. **Repair other authored paths deliberately.** For context, briefs and notes,
   preserve the original, copy to the intended location, and update the exact
   `context`, `brief` or `notes_file` reference. For media outside the root,
   register a project-owned copy and update the slide's asset ID after verifying
   its bytes and provenance. Existing assets that older builds/versions reference
   must remain at their old paths or in their immutable store. Do not mass-rename
   `builds/`, `decisions/`, `versions/`, asset objects, locks or state.
6. **Validate and resume.** Run `project check`, `project status` and
   `project build`. Review invalidated approvals, measure and render the new
   baseline, and record the reorganization in context/project.md. Verify
   retained versions with `project version verify`; save a new numbered version
   after the current build and source are ready. Use `project share` for a
   complete colleague ZIP, then `share-extract` and `share-verify` to resume it.

A layout change does not migrate the deck schema, executable, template bundle
or native editing profile. Follow [project upgrades](upgrading-projects.md) for
those changes. OneDrive conflicts, placeholders and active mutation locks must
be resolved without discarding either collaborator's work; do not force a
migration by deleting locks. If a command reports a broken historical receipt,
retain that evidence and repair the source copy rather than rewriting history.
