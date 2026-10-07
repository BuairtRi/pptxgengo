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

## 1. Installer hardening — active

### First slice: verify before activation

Draft implemented in `internal/releasepackage/install-windows.ps1`:

- Require the package architecture to match the native Windows architecture.
- Reject package reparse points, unsafe paths, duplicate hash entries, missing
  files and files absent from the manifest; validate the complete inventory.
- Require the three tools, both version files, library bundle/index/gallery and
  documentation source metadata. Skill and fonts retain their explicit opt-outs.
- Copy into a unique sibling stage, validate copied bytes against the original
  manifest, then check all three executable versions with a 30-second limit per tool before
  promoting by rename.
- Refuse destination replacement, including a destination created concurrently.
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

Remaining qualification: full private resources, Windows PATH/PowerShell execution,
actual Windows upgrade/rollback, and Office/font qualification on the desktop
runner. These remain incomplete despite core implementation. Detailed operator
workflow: [installation](../installation.md). Search and the remaining product
features below are still independent work to undertake.

## 2a. Hybrid template search — implementation contract next

Extend `LibraryIndex.Find` and `library-find`; keep existing calls compatible.
Proposed `--search-mode keyword|semantic|hybrid` is not implemented. Keyword uses
FTS5 BM25; semantic uses an optional pinned local model and a direct vector scan.
Fuse ranks with an explicit reciprocal-rank policy, deterministic identity ties
and recorded policy parameters; do not add BM25 and cosine scores directly.

Preserve entity ID, namespace, kind, revision, source SHA and artifact hash
contracts already represented by the unified index. Apply explicit filters before
ranking and report structural fit separately from retrieval relevance. Keep the
current structural hints visible rather than treating them as proven content fit.

An embedding record must bind entity/revision, retrieval text hash, source hash,
model digest, runtime version, dimensions and normalization. Missing/stale model
inputs must produce an explicit lexical fallback or semantic-mode error. Select
the runtime/model only after licensing, size, CPU relevance and cross-platform
integration evaluation. No Python or hosted service in normal operation.

First evaluation set: interview lists, practices heat maps, modernization
economics, roadmaps, pillars, exact IDs and synonyms. Judge acceptable candidates
before comparing rank modes. Record cold/warm latency and memory before setting
performance targets. A model package decision is still open.

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

Finish the installer activation/recovery contract and first Windows execution
when the runner arrives. Next implement the lexical search mode and shared
finished-slide identity contract while evaluating the optional local model.
Agree native field-address mappings before the editing pilot, then implement
bounded reconciliation. Model choice, curated content and native acceptance
remain explicit decisions rather than silently selected defaults.
