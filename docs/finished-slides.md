# Reusable finished slides

Implemented in the `pptx-slide-reuse` source branch. Stable v4.1.0 does not
contain these commands. A finished slide is authored, content-complete material,
with a distinct `finished-slide` kind and `curated/slide/<key>` identity.
Production curation and desktop qualification remain pending.

## Publish a closed revision

```sh
pptxdesign project slide publish --project /path/to/source-deck \
  --id maintained-message --library-id curated/slide/delivery-message \
  --revision 1 --name 'Delivery message' --purpose 'Explain the delivery plan' \
  --owner 'Named steward' --out /private/library/delivery-message/1
```

Publication validates the project's exact toolchain and shared template pins.
It preserves the expanded slide values, notes, density and review annotations
in `slide.yaml`, and records source identity/hash and required asset metadata in
`dependencies.json`. Every local asset is copied with its exact bytes and hash;
registry references retain verified identities. Aliased source is published as
canonical values; comments and alias spelling remain in the original project.
The source project is never modified.

The first scope supports declared shared slot/array bindings and typed cards/3
and cards/4. Source must already be `supplied_content`; synthetic specimens and
local templates are rejected. Claims/evidence references, derived-asset chains,
navigation and other typed identity families currently return explicit errors.
These dependencies must be supported explicitly before those slides can publish.

`--preview FILE.png` and `--review FILE.json` optionally retain supplied artifacts.
They do not grant approval. CLI publication always creates a **draft**. The
manifest API supports explicit approval actor/date/reuse scope and requires
preview/review files for approved revisions; reviewed approval maintenance in
the CLI remains pending. Neither hashing nor compilation establishes human
approval, native acceptance or rights to publish supplied material.

Each revision has a sorted closed inventory with file roles, sizes and SHA256,
plus a digest covering identity, revision, metadata and inventory. Reads reject
missing, changed, unlisted and linked files. Paths reject traversal and portable
name collisions. Existing revisions cannot be overwritten; create a new revision
when metadata or content changes. Approval/freshness dates are explicit metadata,
and expiry is evaluated by UTC date. Content stewardship and the initial 10–20
approved useful pages still require operator selection.

## Find and preview

```sh
pptxdesign library-index --bundle /path/to/pinned/bundle \
  --slide-library /private/library --out /tmp/library-with-slides.sqlite
pptxdesign library-find --index /tmp/library-with-slides.sqlite \
  --kinds finished-slide --retrieval keyword --query 'delivery plan' --summary
pptxdesign library-inspect --index /tmp/library-with-slides.sqlite \
  --id curated/slide/delivery-message
pptxdesign library-preview --index /tmp/library-with-slides.sqlite \
  --id curated/slide/delivery-message
```

The library root contains closed revision directories and grouping directories;
loose files and links are rejected. One current entity per stable identity uses
the highest revision, including its explicit draft/approved/deprecated lifecycle.
Use lifecycle filters deliberately; a newer draft does not become approved
because an older revision was approved. Duplicate identity/revision pairs fail.
All revisions contribute to the root fingerprint. Adding or changing a revision
requires rebuilding the index. `--slide-library` can relocate the exact library
when opening an index; bytes and fingerprint must still match.

Search returns kind, revision, purpose, structural hints, owner, approval/review
metadata, dates, package path and current `reuse_status`. Exact template source
and definition pins are verified at indexing. Indexing validates authored source
identity/kind; insertion additionally validates its values, assets and compiler.
Preview files are checked against pinned hashes when opened. Retrieval and source
shape agreement do not establish measured fit or native acceptance for a new deck.

## Insert an independent editable copy

Author composition entries for the existing deck first. Then:

```sh
pptxdesign project slide insert --project /path/to/destination-deck \
  --package /private/library/delivery-message/1 --id delivery-message-copy \
  --after opening --rationale 'This plan explains the proposed delivery sequence' \
  --allow-draft
```

`--allow-draft` is an explicit opt-in for unapproved draft work. Deprecated or
expired revisions are refused. Choose a new stable slide ID; `--before`, `--after`
and `--into-section` follow the project's existing position/section rules.
Toolchain lock bytes, compiler, template and source pins must match exactly.
Year-bound content must match the destination year. Cross-platform re-pinning
and migration need reviewed publication; compatibility is not inferred.

Insertion creates fresh item and asset IDs, remaps only declared media slots,
and preserves business strings. The slide, owned assets, composition rationale,
lineage and decision receipt use one guarded source commit. The validator reads
exact pending bytes; observed composition/lock predecessors are checked under
the guard. Existing authored comments and unrelated slide files are preserved.
Preimages are retained under `decisions/sources/`; errors roll back source writes.
A process crash during a multi-file commit still needs recovery from these
preimages: this is an I/O rollback contract, not a crash-proof filesystem transaction.

The composition entry records library ID/revision/digest, source project/slide/hash,
asset and item remaps, and the preserved-copy/review policy. Later library
revisions do not alter inserted copies. Copy adaptations require the destination
project's evidence, build and review process. Existing maintainer/client review
annotation rules continue to apply; insertion grants no deck approval.

## Evidence and remaining work

Bounded fixtures insert the same authored revision into two projects, including
split source and a real declared image slot. Both headless PowerPoints build.
Fresh slide/item/asset identity, byte independence, later-revision independence,
business-string preservation, collisions, missing/changed files, source guard,
editorial drift, index relocation and preview drift have targeted checks.

Claims migration, more typed/local dependencies, approval maintenance, actual
curation, and Mac/Windows interactive PowerPoint review remain incomplete.
These fixtures do not establish production content or native qualification.
