# Product enhancements execution record

Started 2026-10-06 in slot `pptx-installer-hardening`, targeting `main` at
`fe1331e36db4eb4bf5438ecad11d8d4ceacee28e`.

Source: [product enhancements handoff](product-enhancements-handoff-20261006.md).
This record describes the next work slices and their acceptance criteria. It
separates implemented drafts from qualification; it does not declare the five
handoff enhancements complete.

## Release baseline

The operator selected `v4.1.0` as the stable tag after the signed CLI prerelease
`v4.1.0-rc.2`. Both tags use the same source commit. The stable pipeline rebuilds
and signs binaries with the stable version embedded; it does not relabel the RC
archives. Downloads and release evidence remain private in GitLab project 17.

Stable release published privately:
[GitLab v4.1.0](https://gitlab.samcott.com/riscott/pptxgengo/-/releases/v4.1.0).
[Pipeline 21092](https://gitlab.samcott.com/riscott/pptxgengo/-/pipelines/21092)
passed all 22 jobs. Apple accepted notarization submission
`648adb9f-74d9-45bf-a7c3-f6a4fbda50ce`. Windows Authenticode verification and
private Sigstore manifest attestation passed. All 30 package files were downloaded
outside CI; all 27 manifest file hashes matched, and the downloaded manifest was
byte-identical to the CI-attested manifest. The three downloaded macOS arm64 tools
passed strict signature verification and ran `--version` as `v4.1.0` locally.
The GitLab description was corrected from the publisher's hardcoded prerelease
wording; a source fix for future stable tags is committed in this slot.

These releases are CLI-only. The handoff's `0.1.0-local.21` presentation installation
is a different package baseline, with V11 resources and registered private
branding originals. Those originals are absent on this Mac. Windows runtime and
native PowerPoint qualification remain pending; signing does not establish them.
Read template counts and source pins from bundle metadata when resources arrive.

## 1. Installer hardening — core merged, desktop qualification pending

### First slice: verify before activation

Implemented in `internal/releasepackage/install-windows.ps1`:

- Require the package architecture to match the native Windows architecture.
- Reject package reparse points, unsafe paths, duplicate hash entries, missing
  files and files absent from the manifest; validate the complete inventory.
- Require the three tools, both version files, library bundle/index/gallery and
  documentation source metadata. Skill and fonts retain their explicit opt-outs.
- Copy into a unique sibling stage, validate copied bytes against the original
  manifest, then check all three executable versions with a 30-second limit per tool before
  promoting by rename.
- Reuse an identical verified destination for repeated installation; refuse
  differing bytes and concurrent destination replacement.
- Clean a failed stage during normal exception handling. A killed process may
  leave an unactivated stage; no prior release is overwritten.
- Implement `-StageOnly` for verified installation without PATH/skill/font changes.

### Shared activation implementation

Implemented `pptxgengo installation install|rollback|recover|doctor|uninstall` in
Go. Both source installers call this engine. Packages live in immutable release
directories; one recorded selection drives all three stable dispatchers. Windows
uses a stable user PATH entry, including exact registry type restoration. Unix
uses owned symlinks. Every activation has a durable predecessor receipt and
normal failures recover it. Recovery refuses conflicting user edits. Original
skills and skill symlinks are backed up without changing linked targets.

Repeated identical installation is idempotent; repair retains the prior-version
rollback target. Uninstall restores unchanged original skills, preserves edits,
removes owned activation bindings and retains releases/backups and all authored
projects. Fonts remain a separate step. Doctor reports resource/skill drift,
selection, competing executable locations and pending recovery.

Verification: bounded Go regression checks and race checks on macOS; the actual
signed stable/RC CLI downloads installed into temporary paths with spaces,
all three dispatchers selected each version correctly, rollback restored the
stable version and read-only diagnostics passed. That check changed no user's
active installation. Windows/Linux cross-compilation passed during development.
The source installer now embeds the release version in all three tools.

Merged through [PR #4](https://github.com/BuairtRi/pptxgengo/pull/4), main commit
`0c62fcab66ecddfbc6ebfbf070baed343562d5a0`, synchronized to GitLab and the primary
checkout. Exact head `de51f4c0f8c8d5e4910da29c9a8438b3eca05483` passed Mac normal/race
checks, hosted Windows portable checks and installer staging/rollback/recovery
regressions (16.5 seconds in the Windows step). The Windows fixtures use temporary
roots and no user PATH mutation. [GitLab pipeline 21114](https://gitlab.samcott.com/riscott/pptxgengo/-/pipelines/21114)
passed all six developer/security jobs.

Remaining qualification: full private resources, real package activation in Windows
user PATH, fonts and interactive Office qualification on the desktop runner.
These remain incomplete despite core implementation and hosted core execution. Detailed operator
workflow: [installation](../installation.md). Search and the remaining product
features below are still independent work to undertake.

## 2a. Hybrid template search — lexical baseline implemented

Implemented in slot `pptx-library-retrieval`, [PR #5](https://github.com/BuairtRi/pptxgengo/pull/5):
`library-find --retrieval keyword` uses SQLite FTS5 BM25 with Porter English
stemming. Existing metadata ranking remains the default. The rich text recipe
indexes names, purposes, exact identities, relationships, groups, component terms
and authoring aliases/descriptions; synthetic example copy is excluded. Text
recipe/corpus hashes and FTS schema/row checks reject stale or altered snapshots.
Older indexes retain metadata discovery and need a new index for keyword mode.

Kinds, namespace, lifecycle and adapter capabilities are filters. `--require-shape`
requires all supplied structural/count hints. Ranking shows a separate BM25/rank,
source structure/count agreement and fit-not-measured status; lexical relevance
cannot hide a count caveat. Exact eligible IDs/unique keys rank first. Preview
links, revisions and source pins stay in the existing discovery interface.

Bounded tests cover real V11 engineering relevance judgments, literal-query
safety, repeated/concurrent read-only searches, corpus drift, old indexes and
compact discovery equivalence. Warm query microbenchmarks improved from about
214 ms/201 MB allocations to 113 ms/16 MB after omitting full source definitions
from discovery reads. Verified open remains about 511 ms. These are local
cached-filesystem samples; source definitions remain available for inspect/fit.
See [discovery interface and measurement limits](../semantic-template-discovery.md).

Model-backed semantic queries, persisted vectors and reciprocal-rank fusion remain
pending. An isolated pinned MiniLM experiment executed through the pure Go GoMLX
backend on Mac ARM64 and cross-built without CGO for all six release targets.
Its hashes, footprint, measurements and follow-up contract are recorded in
[the runtime evaluation](local-search-runtime-evaluation-20261006.md). No new model
dependencies or model bytes are shipped in this lexical change. Prefer a separate
optional offline model package; normal queries must not download files.

Next: validate the Go tokenizer and mask pooling, bind embeddings to model/source/
text hashes and dimensions, compare keyword/vector/hybrid on judged synonyms,
implement explicit missing-model fallback, and collect supported-platform latency.

## 2b. Reusable finished slides — identity contract next

Add a distinct finished-slide entity kind to the same discovery system. Reuse the
index identity/revision/source hash/artifact contracts above; do not expose a
finished slide as an empty template. Initial curation of 10–20 approved slides
requires operator selection, an owner, reuse scope and freshness policy.

Each revision closes over maintained YAML, template/compiler/source pins, assets,
preview and evidence. Insertion copies the revision independently, assigns fresh
slide/item identities, registers assets and writes lineage to the composition
record. Validate collisions and dependencies before any project write. Later
library revisions never mutate inserted copies. Proposed insertion commands and
YAML representation require a small contract before implementation.

Acceptance: insert one curated revision into two projects; both build; editing
either copy leaves the other and the library unchanged; provenance survives
split, reorder, export and handoff. Curation and implementation are separate.

## 3. Native editing — mapping contract before serialization changes

Keep deck/slide/item/logical identities stable. Extend the existing object map
rather than inventing a second mapping system. Where one native object contains
multiple source fields, mappings need paragraph/run or table-cell addresses and
immutable baseline text with explicit supported/ambiguous classifications.

Inventory cards/pillars, lists, tables and one diagram. Pilot meaningful grouping,
selection names, native text roles and predictable movement without changing
visual design, source-field ownership or z-order. Review actual Mac and Windows
editing tasks and before/after renderings at representative density tiers.
Existing native groups do not establish that this work is complete.

## 4. Reconciliation — bounded text proposals after mapping contract

Compare immutable native/source baseline, current YAML and edited PPTX using
proven lineage and stable identities. A supported field with only a PPTX change
produces a reviewable proposal; different YAML and PPTX changes produce a
three-way conflict; matching changes and no-ops produce no duplicate mutation.

Missing/duplicated objects and unsupported geometry, shape structure or formatting
remain explicit manual-review items. Preserve edited files and baselines; apply
approved proposals with backups/receipts, then rebuild and review. Never infer
deletion from a missing native shape or rewrite shared template geometry.
Save As, reordering, duplication, deletion and ungrouping need native identity
survival evidence on both platforms before expanding supported mappings.

## Working order

Installer activation/recovery core is merged; desktop qualification follows when
the runner and private resources arrive. Finish lexical CI integration and build
the local embedding/RRF path, then the shared finished-slide identity contract.
Agree native field-address mappings before the editing pilot, then implement
bounded reconciliation. Model choice, curated content and native acceptance
remain explicit decisions rather than silently selected defaults.
