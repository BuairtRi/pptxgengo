# Reusable finished slides

Implemented in maintained source after v4.1.0. Stable v4.1.0 does not
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
local templates are rejected. Selected claims/evidence references are preserved
through the closed dependency contract below. Derived-asset chains,
navigation and other typed identity families currently return explicit errors.
These dependencies must be supported explicitly before those slides can publish.

`--preview FILE.png` and `--review FILE.json` optionally retain supplied artifacts.
They do not grant approval. CLI publication always creates a **draft**. The
`review-reuse` command records an explicit operator decision as a new revision
and requires exact source/preview/review pins for approval. Neither hashing nor
compilation establishes human approval, native acceptance or rights to publish supplied material.

Each revision has a sorted closed inventory with file roles, sizes and SHA256,
plus a digest covering identity, revision, metadata and inventory. Reads reject
missing, changed, unlisted and linked files. Paths reject traversal and portable
name collisions. Existing revisions cannot be overwritten; create a new revision
when metadata or content changes. Approval/freshness dates are explicit metadata,
and expiry is evaluated by UTC date. Content stewardship and the initial 10–20
approved useful pages still require operator selection.

## Preserve selected claims and evidence

`evidence_refs` requires a registered `context.claims` file. Structured YAML/JSON
claims preserve selected literal `text` and human-readable `source` strings.
Unselected structured claims are omitted from the revision. File dependencies
must be declared explicitly, for example:

```yaml
schema: pptxgengo.claims.v1
claims:
  - id: finding-1
    text: Supplied finding for review.
    source: Interview notes from the named source.
    artifacts:
      - path: evidence/interview-notes.txt
        sha256: SHA256_OF_THE_EXACT_FILE
```

The publisher verifies declared SHA256 values and copies exact evidence bytes
into the closed revision. `source` is descriptive copy, not an inferred file
path or a URL to fetch. Publication/insertion never establishes factual support
or approves an adapted claim.

Markdown registries retain their entire exact bytes and their explicit heading/
anchor selector. They are not rewritten into inferred claim text. A structured
registry can preserve such a reference with `registry: {path: claims.md,
sha256: ..., claim_id: finding-1, format: markdown}` and no `text`. Fenced,
indented and inline code examples do not define claim IDs. Full Markdown files
can contain material beyond the selected claim; keep these packages private.

Insertion creates fresh claim IDs, remaps the inserted slide's `evidence_refs`,
copies evidence into owned files, and records `evidence_remaps` in library lineage.
Existing YAML registry nodes/comments remain; an existing Markdown registry is
retained byte-for-byte and referenced from a new structured registry. The new
slide, registry, evidence files and composition decision share one source guard.
Missing selectors, changed bytes and incompatible graphs refuse insertion.
Registry, evidence and slide reference changes participate in stage dependency
pins, so prior approvals do not silently cover changed evidence. Maintainer
exports retain the private registry/evidence files; client exports retain the
deck-only delivery contract.

## Record reviewed reuse as a new revision

After reviewing a draft's authored source, preview and review file, write an
explicit JSON decision using the exact hashes in its manifest:

```json
{
  "schema": "pptxgengo.finished-slide-review.v1",
  "revision_sha256": "SHA256_FROM_THE_DRAFT_MANIFEST",
  "new_revision": 2,
  "action": "approve",
  "actor": "Named reviewer",
  "date": "2026-10-07",
  "reason": "The reviewed content is suitable for the stated reuse context.",
  "reuse_scope": "Explicitly approved audience and context",
  "valid_until": "2026-11-07",
  "source_sha256": "SHA256_FROM_THE_SOURCE_FILE_ENTRY",
  "preview": {"path": "preview.png", "sha256": "SHA256_FROM_THE_PREVIEW_ENTRY"},
  "report": {"path": "review.json", "sha256": "SHA256_FROM_THE_REVIEW_ENTRY"}
}
```

Replace the hash placeholders and review facts with the actual values, then:

```sh
pptxdesign project slide review-reuse --package /private/library/message/1 \
  --decision ./reuse-decision.json --out /private/library/message/2
```

The decision must name a higher revision and the exact predecessor digest.
Approval requires `supplied_content`, a shared template, matching source and
nonempty preview/review artifacts, actor/date/reason/scope, and an explicit
expiry date or `valid_until: "none"`. Future decisions and already-expired new
approvals are refused. The command verifies the entire predecessor closure,
preserves payload and template/toolchain pins, and retains the exact decision,
prior manifest and receipt. It creates a new directory without overwriting
either revision. Identity metadata is an operator declaration, not authentication
of the named person or a cryptographic signature.

Use `action: "deprecate"` or `action: "draft"` with actor/date/reason, source hash
and revision pins to withdraw reuse or request more review. Omit approval-only
fields for these actions. The new revision has no carried approval or expiry;
the predecessor and decision history remain inspectable. Rebuild the library
index to discover any new revision. Existing inserted decks retain their original
revision and still need their own copy, evidence, fit and native review.

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
asset, item and evidence remaps, and the preserved-copy/review policy. Later library
revisions do not alter inserted copies. Copy adaptations require the destination
project's evidence, build and review process. Existing maintainer/client review
annotation rules continue to apply; insertion grants no deck approval.

## Evidence and remaining work

Bounded fixtures insert the same authored revision into two projects, including
split source and a real declared image slot. Both headless PowerPoints build.
Fresh slide/item/asset identity, byte independence, later-revision independence,
business-string preservation, collisions, missing/changed files, source guard,
editorial drift, index relocation and preview drift have targeted checks.

Focused tests cover selected structured/Markdown claims, exact evidence closure,
fresh claim identities in two independent projects, preserved original registries,
false dependency graphs, approved insertion, lifecycle history and hash-bound
review decisions, evidence approval invalidation and concurrent dependency drift.
Additional fixtures split/reorder/build, verify clean client export, relocate a
maintainer handoff with exact evidence/source bytes and rebuild it. Payload reads
are bounded to 64 MiB per file and 512 MiB per integration, with verified closure
and predecessor digest. More typed/local dependencies, actual curation, and Mac/Windows
interactive PowerPoint review remain incomplete.
These fixtures do not establish production content or native qualification.
