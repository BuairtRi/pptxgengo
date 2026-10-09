# Portable projects and packaged browsing decks

Requested by Ri on 2026-10-07. This extends the
[product enhancements handoff](skill-planning/product-enhancements-handoff-20261006.md).
The v4.2.1 release ships the private template browsing deck and its coverage
manifest in all six platform archives. The deck has 717 slides covering the 649
retained v11 templates, and uses synthetic placeholders where external artwork
is unavailable. The separate content-complete reusable-slide deck remains
deferred: tooling exists, but production curation and its approved inventory do
not ship. See [release details](../release/README.md#signed-release-scope-v421).

The sections below retain the implementation contract and original requirements
for context; historical requested scope is not a claim that every requested
artifact shipped.

## Packaged template library and deferred reusable library

Private GitLab release CI regenerates the template browsing deck and includes it
in installation archives. The separate reusable content-complete library remains
deferred; its intended contract is retained below for the later release:

1. A complete template library: native editable placeholder slides for every
   retained template variant in the pinned library, organized with family dividers and a
   “How to use this deck” component. The agreed packaged default contains every
   canonical template plus the declared supported frame/rail gallery. The current
   pinned catalog yields 649 retained template variants with their canonical pinned frames,
   36 gallery entries and 32
   divider/instruction slides: 717 slides. Counts are derived from the exact
   bundle, not fixed invariants. Complete exhaustive frame/rail combinations
   remain an explicit generation option (currently 9,072 frame specimens),
   rather than making the default browse deck thousands of slides. Coverage
   identifies the selected mode and each entry.
2. A separate library of versioned reusable authored slides, organized for
   browsing and copy/paste. Preserve content, assets, source/revision provenance
   and visible approval/freshness information. Do not present placeholder
   specimens as authored reusable slides.

Colleagues must be able to copy/paste slides using PowerPoint without invoking
the CLI. Preserve native editability, theme/font requirements and useful names.
Dividers/instructions must explain inherited formatting, placeholders,
approved reuse scope and the provenance limits of a copied slide.

Use exact bundle/content pins rather than fixed catalog counts. Produce a
coverage manifest and hashes alongside the packaged template deck. Release CI
regenerates it when pinned inputs change and scans/attests final archived bytes.
All artifacts remain private in GitLab. Declared schematic placeholders are
identified in coverage and do not claim original artwork fidelity. Unavailable
reusable content and excluded external branding/photos do not block the current
template-only release. An incomplete deck cannot claim complete template coverage.

The existing library-reference/frame-reference and finished-slide compilation
are starting points. Derive valid frame/rail combinations from contracts and
reject incompatible combinations. Default coverage must close every retained
template variant and every declared gallery specimen; it must not claim exhaustive
variant coverage. In explicit exhaustive mode, include every compatible variant
without silent omissions.
The browse command uses `browsing-library --kind templates --frames catalog|exhaustive`;
`catalog` is the packaged default. Coverage's `frame_mode` labels the mode,
`entries` identify template/frame specimens, and `frame_aliases`/`frame_exclusions`
explain request deduplication and structural incompatibilities in exhaustive
mode. The current exhaustive output has 9,753 slides total (649 template variants,
9,072 resolved frame specimens and 32 organizational slides); those derived
counts can change with pinned inputs.

The planned reusable-slide deck includes only the latest approved revisions. Drafts,
deprecated revisions and expired approvals are excluded. Preserve explicit
withdrawal history so a newer withdrawal cannot resurrect an older approval.

## Consistent portable project layout

The operator skill includes the complete [canonical folder tree and repair workflow](../skills/west-monroe-presentations/references/project-structure.md), including the limits of current layout migration.

Explicit schematic adaptations retain `placeholder_for` asset descriptors and
their exact content-addressed bytes with the project. A share or snapshot must
include these objects and preserve descriptions that identify them as schematic.
Use [scaffold media](../skills/west-monroe-presentations/references/scaffold-media.md)
to install the emitted payloads; do not silently replace a missing original or
change source copy that happens to match an asset ID. Registered original icons
and arrows are distinct from these declared placeholders.

Source-derived geometry metadata and stable collection keys belong to authored
template state and travel with numbered snapshots. They retain source geometry
when measured ink differs, while family commands support intentional topology
changes. Retain them alongside facts, units and semantic relationships; native
movement does not establish a new date, membership or source value.

Projects should have a predictable layout:

    deck.yaml
    slides/
      <stable-slide-id>.yaml
      templates/
        <local-template-id>.yaml
    assets/
    versions/
      <numbered-version>/
    composition-log.yaml
    toolchain.lock.json

Use one authoritative slide order in deck.yaml; filenames and numbered deck
versions must not replace stable semantic slide identities. Shared template
definitions remain pinned references. Keep local template files under
slides/templates and owned original/derived assets under assets.

The numbered-version contract retains authored source, exact asset-revision
references, template pins and generated deck together, with a visible current
version. All versions of one deck share its own assets folder; this is not a
machine-wide or cross-client store. Unchanged image bytes are stored once.
Version assets individually when their bytes change, using immutable hash-bound
objects and exact per-version mappings. Never overwrite an asset object, retained
deck version or immutable build baseline. Old versions must keep their original
assets after newer source selects changed images. Use ordinary relative files
and manifests that survive OneDrive and ZIP; do not depend on filesystem links. Specify the exact schema,
number allocation, latest pointer, concurrent publication behavior and recovery
before implementing the command. Existing build receipts/history remain valid.

Provide a CLI workflow to create a complete ZIP for colleagues and OneDrive
collaboration. Extend existing maintainer export where useful. Archives need
relative paths, clear open/resume instructions and an inventory/hash manifest.
Separate client-only deck delivery from private source/evidence handoff.
Document runtime/font/branding requirements and explicit toolchain migration;
portability must not be confused with changing a project's executable pin.

OneDrive is file synchronization, not a source transaction or lock service.
Detect conflicting copies, interrupted sync and changed predecessors. Preserve
both collaborators' source and native edits; never silently pick one. Define
existing-project migration with a dry run and retain preimages/receipts.

Update CLI help, presentation skill references and offline colleague guides
together. An installed version must expose commands that its documentation
describes; proposed commands must stay clearly labeled until implemented.

## PowerPoint qualification location

Use an owned stable folder under ~/Documents/pptxgengo-qualification for Mac
desktop fixtures rather than global temporary folders. Permission state,
including Full Disk Access, is not a qualification result: record
actual native open/Save As/render output, app version and file hashes.
Bound waits, close only exact task copies and preserve failed evidence.

## Current acceptance evidence and deferred desktop scope

- Default coverage proves every pinned retained template variant and every declared
  frame/rail gallery variant is present, and the packaged template deck opens in
  the named qualified PowerPoint environment. Explicit exhaustive coverage separately closes all compatible
  frame/rail variants without silently labeling the default as exhaustive.
- A colleague copies a template into another deck with expected native objects
  and assets. Windows desktop and content-complete reusable slide qualification
  remain deferred; they are not current v4.3.0 acceptance claims.
- Release archives contain the template deck, matching coverage pins, hashes and signed
  release inventory.
- New projects use the declared layout; migration preserves older project
  semantics, approvals, lineage and immutable history.
- Two numbered source/deck versions remain independent and inspectable.
- The complete ZIP stores the deck-owned shared assets once across versions.
- A colleague extracts a complete ZIP into a different path and can resume,
  inspect provenance and rebuild with the stated runtime requirements.
- Concurrent/conflicting OneDrive changes are detected without losing either
  source history or edited PowerPoint files.

## Implemented portable project workflow

The commands in this section are implemented. Browsing-deck release generation
above is separate scope and is not implied by this project workflow.

```sh
pptxgengo design project create --out ./new-deck --id modernization --title 'Modernization options' --year 2026 --bundle v11 --template cards/3
pptxgengo design project layout --project ./existing-deck --dry-run
pptxgengo design project layout --project ./existing-deck --apply
pptxgengo design project check --project ./new-deck
pptxgengo design project build --project ./new-deck
pptxgengo design project version save --project ./new-deck --actor 'Named author' --message 'First complete source and deck'
pptxgengo design project version list --project ./new-deck
pptxgengo design project version verify --project ./new-deck --number 000001
pptxgengo design project share --project ./new-deck --out ./colleague-project.zip
pptxgengo design project share-extract --archive ./colleague-project.zip --out ./colleague-project
pptxgengo design project share-verify --project ./colleague-project
pptxgengo design project version materialize --project ./colleague-project --number 000001 --out ./restored-version
```

`create` requires a new directory and an actual shared template. The initial
slide is explicitly scaffold/example content, not an authored reusable slide.
The new project uses `slides/first-slide.yaml`, `slides/templates/`, `assets/`
and `versions/`. Existing `project split` now emits stable slide filenames and
local definitions under `slides/templates/`; already referenced paths are
retained until the explicit layout migration.

`layout --dry-run` is the default. It validates every destination before writing;
`--apply` changes only YAML file references and presentation, preserving the
expanded canonical source, stable IDs and authored slide order. Source comments,
notes references and scalar content are retained. Old files and source preimages
remain available for previous receipts and history. Historical approvals are
retained, but source-file dependency changes are evaluated by `project status`;
layout migration is not a new operator approval. Build again before saving a
version. `project migrate` remains the distinct, explicit toolchain migration.

### Immutable asset revisions

New asset registrations use a full SHA256 path:
`assets/objects/sha256/<64-character-hash>`. An asset identity in deck.yaml
selects an exact object, with its own description, focus and content hash. Different
identities can share unchanged bytes. Object files are ordinary immutable files,
not links or a machine-wide cache.

```sh
pptxgengo design project asset add --project ./new-deck --id hero --file ./approved.png --description 'Approved photograph'
pptxgengo design project asset revise --project ./new-deck --id hero --file ./updated.png --description 'Updated approved photograph' --expect-sha256 OLD_FULL_SHA256
```

Revision checks the exact predecessor bytes and source tree under the mutation
guard. The old object remains. Reusing the same bytes reuses the same object;
there is no per-deck-version asset copy. The decision receipt records both old
and new hashes. Asset changes invalidate relevant approvals through existing
project dependency tracking; author names on snapshots do not approve content.

Legacy asset paths are preserved. Numbered snapshots record a mapping from
these paths to immutable objects, so an old snapshot still reconstructs its
original bytes after a later working source chooses a different asset.

### Version schema, allocation and current pointer

Versions are contiguous six-digit numbers, starting with `000001` and capped at
`999999`. Each directory has:

```text
versions/000001/
  manifest.json
  deck.pptx
  source/
    deck.yaml
    slides/...
    toolchain.lock.json
    state.json
    context/...
    decisions/...
```

`manifest.json` uses `pptxgengo.deck-version.v1`. It records project ID, author,
message, timestamp, source/canonical hashes, build ID, generated deck hash,
every retained source/context/evidence file hash, original asset-path-to-object
mappings, and exact shared build-history hashes. It binds the prior version
number and manifest SHA256. Assets are shared at the deck root. Existing immutable
builds/receipts/native baselines remain under `builds/` and are referenced by
hash, rather than copied into every version. The snapshot's `deck.pptx` is the
browsable generated output; independent native working copies in the project
are captured as private source/evidence files too.

`versions/current.json` uses `pptxgengo.deck-version-pointer.v1`, containing the
latest number and exact manifest SHA256. The pointer is published only after
all snapshot files are written and their captured predecessors are checked.
Source snapshots are not working roots: their asset and build mappings are
resolved by `version materialize` into a **new directory outside the source
project**, where ordinary project-relative paths, original pins and receipts
are restored. No `../../../` source references or relaxed SafePath rules are
introduced.

Saving requires a current generated build with matching authored source,
semantic content, dependencies and unmodified immutable baseline. Every retained
build receipt/output is checked. A local exclusive `.project-version.lock`
prevents simultaneous local publishers; source/build locks are never taken
over. Changed file inventories, changed pointers, conflicting copies, unresolved
OneDrive placeholders, case-colliding names and nonportable Windows paths are
rejected. A snapshot left by an interrupted publication is preserved and
reported; it is not silently adopted or deleted.

For a complete verified snapshot whose pointer publication was interrupted,
explicit recovery changes only the pointer:

```sh
pptxgengo design project version recover --project ./new-deck --number 000002 --expect-current-sha256 EXACT_HASH_OF_CURRENT_JSON
```

Use `absent` only if the initial pointer was never created. Recovery verifies
contiguous immutable history and the immediate predecessor, refuses another
writer's lock and requires the exact pointer preimage. Incomplete/conflicting
snapshots need explicit preservation and resolution; recovery cannot invent
missing source or choose between colleagues' edits. OneDrive synchronizes files
and does not participate in the local lock. Offline simultaneous changes can
still create conflicts; retain both copies and reconcile before publication.

### Complete private colleague ZIP

`share` includes working source and native copies, complete numbered snapshots,
all build receipts/history, pins, contexts, approvals, decisions and evidence.
It is explicitly private material. Use existing `project export --mode client`
for a client-only delivery.

The ZIP contains one object per asset SHA256 and a logical file/hash inventory
in `share-manifest.json`. Its `asset_aliases` map retains old paths without
storing identical legacy original bytes twice. `share-extract` validates every
entry, hash and alias, then expands legacy aliases to normal relative files in
a **new** directory. This can require physical legacy path copies after
extraction; asset bytes remain stored once in the transport ZIP and all numbered
versions use the shared object store. No filesystem links are required.
Manual unzip is sufficient to browse `versions/<number>/deck.pptx`; use
`share-extract` for a complete working project. `share-verify` rejects changed,
missing or extra files against the extraction inventory and verifies all
numbered snapshots. It is an integrity baseline, not a restriction against
making deliberate later edits; create a new share after collaboration.

Sharing does not change the lock's executable, OS/architecture, engine, bundle,
font or calibration pins. Rebuild with the original runtime or perform explicit
toolchain migration. Shared registry branding resources and installed fonts are
still required; existing `project export --mode offline` supplies the separate
current-build offline runtime closure when those private resources are present.
The source ZIP does not promise native rendering, cross-platform executable
migration or an operator-approved deck.

### Evidence covered by automated tests

Generic real-build tests cover layout semantic preservation/preimages, new
project pins, three complete numbered snapshots, a changed asset revision,
unchanged-object reuse, exactly two asset objects in the complete ZIP,
validated extraction to a different path, materialization and recompilation of
each snapshot, predecessor/pointer drift, busy locks, sync conflicts/placeholders,
asset tampering, occupied output paths, reserved Windows paths and unsafe ZIP
traversal. Human OneDrive collaboration and PowerPoint copy/paste remain separate
qualification.

Portable producer and extractor limits are explicit: at most 100,000 logical
files/aliases, 512 MiB per ordinary file, 16 MiB per JSON manifest, and 8 GiB
for a complete logical inventory including every expanded legacy alias. Reads
are bounded before allocation/accumulation, and ZIP transport limits do not
replace expansion limits. All logical case and file/directory prefix collisions
are rejected before creating an extraction directory. Immutable version folders
are closed inventories; extra files, including unlisted synchronized copies,
are drift rather than silently ignored content. Large handoffs that exceed a
limit are refused with evidence retained; no files/versions are dropped to make
them fit.

New-directory create, extraction, materialization and failed ZIP outputs are retained if an
operation fails after creation. A concurrent synchronization may already have
introduced another colleague's files; the CLI never recursively deletes that
folder to roll back. Inspect and preserve failed evidence, resolve conflicts or
choose a new output directory explicitly before retrying. Existing output
folders are always refused.
