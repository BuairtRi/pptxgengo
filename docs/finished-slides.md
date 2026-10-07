# Finished-slide contract — initial implementation

Status: implementation contract in managed slot `pptx-slide-reuse`. This is not
a curated content library or evidence of native qualification. It implements
the next independent part of the product-enhancements handoff.

Implemented foundation: `internal/finishedslide` defines and seals manifests,
verifies complete file closure, detects missing/tampered/unlisted/linked files,
checks portable paths and source/compiler pins, reports UTC date freshness and
creates new directories with kernel-enforced non-replacement. Normal/race tests
cover immutable revisions and caller-byte independence. Project publication,
insertion, SQLite discovery and the acceptance work below are still pending;
the command names below are proposals, not available CLI commands.

## Identity and revision

A finished slide is a distinct `finished-slide` discovery entity, with stable
`curated/slide/<key>` identity and a positive immutable revision. It contains
authored content, never a renamed synthetic specimen. Each revision is a closed
directory with a manifest, maintained slide YAML, assets, claim/evidence context,
preview and review evidence. Every file has a portable relative path, size and
SHA256. No symlink, traversal, case-insensitive collision or missing dependency
is accepted. Manifest identity and discovery metadata are hashed with the file
inventory into a revision digest. Existing revisions cannot be overwritten.

The manifest records owner, purpose, approved reuse scope, review status and
approval/freshness dates. `approved` requires explicit approval metadata; neither
compilation nor publication creates approval. Expired content remains visibly
stale and requires reviewed refresh. Initial 10–20 approved pages still need
operator selection and ownership.

## Dependencies and insertion

The first implementation supports one content-complete shared-template slide
with exact template revision/source/definition hashes and project lock/compiler
pins. Unsupported deck-local dependencies fail before writes. The package
contains every required local asset's bytes and registration facts; shared
registry assets need verified identities and hashes. Evidence is copied with
explicit origin and collision-safe references. Internal review annotations
retain existing maintainer/client export behavior.

Insertion is an independent copy with fresh slide and item identities. The
maintained YAML references project assets and shared templates without embedding
another template tree. Library identity, revision/digest, original identity,
asset remaps and content/adaptation decisions remain in provenance and the
composition log. Updating a library revision never changes inserted copies.

Validation completes before writes: project/source/lock compatibility, IDs,
dependency closure, assets, evidence references, position/section rules and
compilation. Use the existing source-mutation guard and validated atomic source
change machinery; retain preimages and a receipt. A failed insertion restores
the previous authored tree and reports any retained staging files.

## CLI and discovery

Proposed commands are `project slide-publish` (explicit revision maintenance)
and `project slide-insert` (package revision, fresh ID, before/after position and
authored rationale). `library-index --slide-library DIR` adds closed revisions
to the existing SQLite projection. `library-find --kinds finished-slide` and
inspect/preview expose their identity, revision, content purpose, reuse/freshness
status, hashes and artifacts. Search and preview never modify a project.

## Acceptance work

Publish a synthetic, explicitly unapproved test revision; insert it into two
temporary projects and build both. Modify either copy and verify the other and
library bytes remain unchanged. Publish a new revision and verify prior copies
and lineage remain unchanged. Cover collisions, missing/tampered files, pin
mismatches, unsupported dependencies, concurrent mutations and write failures.
Verify split/reorder/export/provenance and client review-group removal. Actual
content curation and Mac/Windows native review remain separate acceptance work.
