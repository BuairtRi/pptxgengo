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

The `pptx-local-vectors` slot implements pinned MiniLM queries through the pure Go
backend, explicit optional package maintenance, complete source-bound persisted
vectors, exact cosine scan and reciprocal rank fusion (`k=60`). The tokenizer and
mean-mask/L2 pooling passed independent golden comparisons. Model bytes stay out
of Git and CLI archives. Normal queries are offline. Missing hybrid resources
produce visible keyword fallback; stale/corrupt resources fail with rebuild
guidance. Metadata remains the default, with source shape and measured-fit status
separate from relevance.

The V11 engineering set covered 6/6 baseline queries and 4/6 synonyms in hybrid's
first ten results, versus 6/6 and 2/6 for keyword. Two synonym queries still missed;
no universal improvement or native/operator acceptance is claimed. A complete
1,969-entity snapshot and query/startup measurements are recorded in
[the runtime evaluation](local-search-runtime-evaluation-20261006.md). New model
dependencies have refreshed notices and zero locally reachable vulnerabilities;
all six no-CGO targets compiled. Exact-head Linux/macOS/Windows model and
regression gates passed before PR #6 merged; optional signed model package
publication remains separate.

Keyword PR #5 merged as `b6e933ca` after exact-head Linux security/developer,
Windows portable and macOS normal/race checks passed. GitLab main and the primary
checkout were aligned through `slot merge`.

Next: finish the managed finished-slide implementation, then establish native
field mappings for direct editing and reconciliation.

## 2b. Reusable finished slides — core in managed slot

The `pptx-slide-reuse` slot implements closed immutable revisions, project
publication/insertion and the existing SQLite discovery/preview interface.
`project slide publish` starts draft revisions from supplied-content shared
slot/array templates or typed cards. `project slide insert` validates exact
source/template/toolchain/year pins, assigns fresh slide/item/asset identities,
and commits the slide, assets, composition lineage and receipt together.
Only declared media slots are remapped; business strings remain copy. Existing
source/comments and observed editorial predecessors use the mutation guard and
recoverable preimages. This provides I/O rollback, not crash-proof transactions.

The index keeps the highest explicit revision per identity, distinct
`finished-slide` kinds, owner/lifecycle/date metadata and current freshness
status. The complete library fingerprint and selected file pins detect library
changes; preview hashes and relocated roots are verified. Keyword search shares
the template index. Model snapshots must be regenerated when entities change.

Targeted tests published explicitly authored fixtures, inserted each into two
projects including split files and local images, and built both headless decks.
They checked fresh identity, exact byte independence, unchanged earlier copies
across revisions, ordinary-string preservation, collisions, unsupported
sources, file/preview drift, relocation and concurrent-source/editorial guards.
Targeted source/reuse race checks passed in 27.32 seconds on this M5 Max.
CLI publish/insert also passed. Full slot preflight and exact-head GitLab
pipeline 21147 (all eight gates), hosted Windows portable/model and macOS
normal/race/model checks passed before PR #7 merged as `2dcde4fc`.

Remaining: closed claim/evidence migration, navigation/derived/local-template
and wider typed identity contracts, reviewed CLI approval maintenance, initial
operator-selected 10–20 approved slides, export/handoff qualification and actual
Mac/Windows native review. No curated production content or native qualification
is claimed. See [commands and scope](../finished-slides.md).

Local vector PR #6 merged as `220fba93` after exact-head GitLab pipeline 21141
(all eight gates), Windows runtime/portable and macOS runtime/normal/race checks
passed. CI snapshot/query uses 26 original pinned entities spanning every kind;
full V5/V11 catalog evidence is recorded separately in the runtime evaluation.
Stable v4.1.0 remains unchanged by these source merges.

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

Hosted lexical CI follow-up: head `8567e908` passed GitLab pipeline 21125,
but one hosted Mac normal job hit the original 150-second package ceiling.
The short-suite package ceiling is now five minutes; selected race workloads
retain their 150-second ceilings. Updated head `42c7c2a0` passed GitLab pipeline
21130 and both Mac normal/race and Windows portable runs before PR #5 merged.
The slowest hosted race job took 9m58s; the next source build raises its job
ceiling to 15 minutes while retaining the per-package bounds.

Installer activation/recovery core is merged; desktop qualification follows when
the runner and private resources arrive. Finish reusable slide integration,
then the shared native editing/reconciliation identity contract.
Agree native field-address mappings before the editing pilot, then implement
bounded reconciliation. Model choice, curated content and native acceptance
remain explicit decisions rather than silently selected defaults.

## Native field contract implementation follow-up

The `pptx-native-fields` slot adds the shared paragraph/run/cell address and
source-field baseline contract before editing or reconciliation changes. Run
boundaries no longer invent native paragraph breaks. Typed card title/body
objects map to their specific keyed field; frame mappings respect actual raw
slots. Exact unique plain fields retain stable source-slot identity and hashes;
rich, dynamic, bullet, cell and ambiguous fields remain explicit manual review.
Native structure hashes separate text-leaf changes from geometry/format/shape
changes conservatively. Targeted XML and generated-card tests passed locally.
Full slot preflight and exact-head GitLab pipeline 21153 (all eight gates),
hosted Mac normal/race/model and Windows portable/model checks passed before
PR #8 merged as `be2ed696`. Native identity survival, editing pilot and three-way
adoption remain pending. See [the maintained mapping contract](../native-field-mapping.md).


## Native generation/shape lineage implementation follow-up

The `pptx-native-lineage` slot adds standard PowerPoint presentation/slide/shape
Tags, deterministic source/lock/native generation pins, and baseline group
ownership. Project builds retain identical native bytes across repeated builds
of one generation. Edited-file inspection follows live relationships and tag
identity, with explicit missing/duplicate/unmatched/ownership-change reports.
Names, numbers, order, current text and geometry are not matching heuristics.
Bounded package/XML validation and headless mutation fixtures passed locally.

A separate owned synthetic Mac PowerPoint 16.113.4 Save As trial timed out with
AppleEvent `-1712` and produced no saved output; its copy was closed without
saving. Desktop survival remains unqualified. Full preflight, exact-head GitLab pipeline
21163 (all eight gates), hosted Mac normal/race/model and Windows portable/model
checks passed before PR #9 merged as `38e2b5fa`. Three-way proposals, guarded
adoption and native editing family pilots follow. See [the lineage contract](../native-lineage.md).

## Three-way text proposals and reviewed adoption — managed source implementation

The `pptx-text-reconcile` slot exposes `project reconcile propose` and `adopt`.
Receipt-pinned immutable native/source baselines are compared with current YAML
and an edited copy through generation/slide/shape tags and keyed source slots.
Reports distinguish no-op, YAML-only, matching changes, native-only changes,
conflicts and manual review. Supported proposals retain exact plain text;
ambiguous fields, rich/bullet/dynamic/cell content, object changes, geometry,
format and other package payloads remain visible for review.

Closed review packets retain all comparison inputs. Explicit named review
choices are bound to the report hash and replayed against verified inputs before
source mutation. Adoption updates only reviewed authored scalars, preserves
comments/split files and parallel-copy predecessors, validates source and changed
slide fit under the existing mutation guard, and retains edited deck/report/raw
choices plus a receipt. Repeating an adoption does not duplicate mutations;
empty/no-op choices write nothing. Unsupported changes remain in receipts.

Headless fixtures and CLI workflows cover three selected fields and rebuild,
conflicts, repeat/no-op, split readable copy, packet tampering, immutable guards
and stale source refusal. Full slot preflight and exact-head hosted CI are pending.
Actual Mac/Windows PowerPoint Save As and colleague editing acceptance remain
unqualified. This capability is not included in released v4.1.0. See
[commands, decision format and scope](../text-reconciliation.md).

## Native editing inventory — managed pilot slot

`pptx-editability-pilot` adds `project editability` over a receipt-verified build.
It inventories actual Selection Pane names, native kinds, token/group ownership,
parent/top-level selection units, nesting, paragraph/cell counts and source-field
coverage. Card-row fixtures verify the actual nested card group and exact title/
body fields; representative list, native table and diagram fixtures establish
structural coverage. Reports retain source/receipt/native pins and explicitly
state desktop qualification is not recorded.

Inventory alone does not establish whole-library/native editing acceptance.
The family task matrix covers copy, selection, move/align/resize, table cells,
diagram adjustment and Save As/identity tasks on both desktop platforms. Full
preflight and hosted checks are pending. See [the maintained pilot](../native-editing-pilot.md).

Reconciliation hosted CI follow-up: local full preflight and GitLab pipeline
21169 passed at `7c424009`. The hosted Mac race batch reached its retained
150-second package ceiling while adding all reconciliation fixtures to the
previous selected workload. Reconciliation now runs as a separate bounded race
command with the same ceiling. Exact head `c78eea29` passed full preflight,
GitLab pipeline 21174 and all ten applicable GitHub gates, then PR #10 merged
as `d464f158`. No tests or race checking were removed.

### Explicit native diagram authoring pilot

The editing slot adds `editable-block` and `attached-connector` as explicit local
authoring choices. A combined block keeps the original measured text envelope,
font/style/alignment/fill in one native rectangle, without an additional surface
and text group. A straight native connector references declared rectangle node
IDs/sites, permits forward endpoints, and rejects missing/ambiguous/unsupported
targets. Existing stock blocks/static connector serialization is retained.

Structural fixtures verify native endpoints, source-field/generation continuity,
inner text insets and three-field reviewed adoption/rebuild; unsupported connector
changes remain manual. Plan comparisons preserve three style roles across all
three V11 density tiers. A separately retained synthetic desktop fixture enables
future move/resize/Save As tasks. Mac/Windows desktop visual and behavior evidence,
actual density acceptance and shared-library rollout are still pending.


Native editing hosted CI follow-up: exact head `96cbbac9` passed the full slot
preflight, all eight GitLab pipeline 21180 jobs and all ten applicable GitHub
checks. PR #11 merged as `96a5676b`. Source and headless geometry/mapping checks
passed; desktop visual/editing behavior and wider rollout remain unqualified.

## Finished-slide stewardship and evidence closure — managed source work

The `pptx-slide-stewardship` slot adds `project slide review-reuse`. Explicit
JSON decisions bind a higher immutable revision to predecessor/source hashes,
actor/date/reason and, for approval, the exact nonempty preview/review artifacts,
reuse scope and expiry policy. Approval/deprecation/draft transitions preserve
all payload and pins plus the raw decision, prior manifest and receipt. These
are operator declarations, not authenticated human identities or native verdicts.

Selected structured claims retain literal copy/source and declared hashed file
artifacts. Markdown evidence retains exact whole-registry bytes and an explicit
selector, with no inferred prose conversion; private packages can therefore
contain material beyond the selected claim. Insertion remaps fresh claim IDs,
copies owned evidence, records lineage, preserves existing registry comments or
Markdown bytes, and validates the entire graph under the existing source guard.
Evidence dependency changes invalidate stage approvals. Bounded verified payload
reads protect approval/integration from changed or oversized package files.

Focused tests passed for six source/destination claim-format combinations and
real headless builds, preserved bytes/copy, lifecycle history, CLI approval and
insertion, false resealed graphs, dependency drift, approval invalidation,
split/reorder, clean client export and relocated maintainer handoff/rebuild.
PR #12 passed full slot preflight, all ten GitHub checks and all eight GitLab pipeline 21187 jobs at exact head `72e64638`; it merged as `0882e4c3`. The initial
10–20 approved real pages, broader typed/local dependencies, branding assets and
Mac/Windows desktop qualification are still outstanding. This work is not in
released v4.1.0; no new release tag has been cut.


Stewardship hosted CI follow-up: exact head `72e64638` passed full local
submission preflight, all eight GitLab pipeline 21187 jobs and all ten
applicable GitHub macOS/Windows checks. PR #12 merged as `0882e4c3` and
slotctl aligned private GitLab/main and the clean primary. Its approved content
set and actual desktop qualification remain outstanding.


## Separate offline model release automation — managed source work

The `pptx-offline-model-release` slot shares dependency-free compiled model pins
and complete package verification with the inference maintenance command.
Release tooling creates a deterministic, relocatable platform-independent ZIP
from exact weights/tokenizer/license/attribution; real pinned local tests verify
reproducibility, extraction, copied evidence, wrong tag and tamper refusal.

Branch CI prepares/scans an unpublished package fixture after pinned inference
checks. Tag CI uses the actual release identity, copies the separate model ZIP,
rechecks/rescans final bytes and includes model evidence/SBOM/vulnerabilities in
the signed private release manifest. Linux/macOS/Windows model lanes exercise
real package regressions. The model remains optional and separate from CLI
archives; data file scans/signature coverage do not establish model safety or
relevance. No new release tag has been cut. The final integration evidence is
recorded below.


Model package validation follow-up: GitLab pipeline 21190 at `57fbcf3c`
passed the Kubernetes `security:offline-model` job. Its actual extracted archive
received a CycloneDX file inventory and Grype report with zero matches or
suppressions and a valid database built 2026-10-07T06:31:48Z. An independent
private download passed complete package/identity/file-hash verification and
matched the locally prepared ZIP byte-for-byte (`51034ad7…`). A simulation of
the complete local tag CI configuration validated all 25 jobs with no warnings.

Hosted Windows model checks exposed an existing archive regression test's Unix
filesystem permission assumption when first enabling release-tool tests there.
The test now checks encoded TAR/ZIP executable modes on every OS, extracted
payloads everywhere and Unix filesystem permissions on Unix. Model package
regressions themselves passed on Windows. PR #13 passed full slot preflight, all
ten GitHub checks and all nine GitLab pipeline 21194 jobs at exact head
`838fbcc5`; it merged as `0ca5f0cf`. No new release publication was performed.


## Native round-trip qualification harness — managed source work

Slot `pptx-native-roundtrip` adds a retained synthetic two-slide fixture for
initial Save As, three prescribed supported text edits and native slide reorder.
The independent verifier binds plans to immutable receipt outputs, reads actual
saved tag/text/order graphs, retains package/format manual items, refuses wrong
text/order/tags or baseline drift, and adopts/rebuilds only exact prescribed
changes in the disposable fixture. Repeated adoption/reconciliation and retained
original edited bytes/baselines are checked. Decisions explicitly identify
fixture automation and do not invent human acceptance.

An opt-in Windows COM helper creates only new owned paths, validates lineage
before actions, records app version/timestamps, bounds helper execution and
exact-path cleanup, and never quits/kills PowerPoint. Both desktop CI lanes run
it after native export smoke; portable Windows CI parses it and tests simulated
failure behavior. Mac uses the same preparation/supplied-file verification
workflow while automated Save As remains unqualified after the earlier timeout.

Focused hermetic and selected race checks passed locally, including tampered
plans/baselines/tags, wrong text/order, cleanup failures, repeated output refusal,
Unicode/spaced paths and exact rebuilt copy. PR #14 passed full slot preflight,
all ten GitHub checks and all nine GitLab pipeline 21205 jobs at exact head
`b8bbf04d`; it merged as `69f64cd3`. One GitHub race job reached its overall
15-minute deadline during the last keyword-concurrency test; a same-head rerun
passed with every assertion retained. No live Windows desktop pass, new Mac Save As evidence, human editing/
visual acceptance or family-wide rollout is claimed. This source work is not in
stable v4.1.0, and no further release tag has been cut.

## Offline search timing and memory — managed source work

Slot `pptx-search-performance` adds a source-only benchmark using production
verified index/open/find APIs. Reports retain fresh-process first-query and
same-index repeated durations, exact dataset/source/model/file pins and sizes,
first/final OS peak resident memory, Go heap reservations and GC settings.
Invalid/stale/fallback/drifting inputs or results and overwritten reports fail.
OS cache is uncontrolled and hashing/preparation/CLI startup are not timed;
model loading still happens on each find. No performance budget is invented.

Pinned model CI adds separate child-process measurements on Kubernetes Linux,
macOS and Windows. Preliminary exact head `54851875` passed the model lanes
and all nine GitLab pipeline 21215 jobs; actual three-platform JSON was
downloaded and checked. Linux's initial getrusage evidence exposed inherited
pre-exec parent memory. The final source uses current-address-space VmHWM and
a touched-parent/subprocess regression instead. Portable Windows tests also
exercise its actual memory API. The Go hosted job budget becomes 20 minutes
while preserving every 150-second package race ceiling and assertion.

The current full 1,969-entity V11 Mac sample measured hybrid first 475.4 ms,
repeated median 468.7 ms, first peak 368.3 MiB and final peak 772.3 MiB, with
91,102,969 bytes of pinned inference artifacts. Complete scope, input hashes
and keyword/semantic comparisons are in `docs/search-performance.md`.
PR #15 passed full five-step slot preflight, all ten GitHub checks and all nine
GitLab pipeline 21225 jobs at exact head `8cec1d7b`; it merged as `bd63c193`.
Independently downloaded nine actual platform/mode reports passed count/model/
source/raw memory/GC checks. Both hosted race runs passed in 17m27s/16m24s.
GitLab's race run at `334da317` exposed asynchronous Linux RSS accounting:
the raw last VmHWM sample was 241,664 bytes below the first sample. The report
now retains both raw samples and their observed maximum, with explicit kernel
approximation documented; no assertion or qualified memory precision is hidden.
Actual full-library Windows/other architecture performance, agreed budgets,
operator relevance/native fit and desktop editing acceptance remain open.
No further release tag was cut.

## Actual installer processes and native architecture CI — managed source work

Slot `pptx-installer-process` builds two synthetic CLI fixture versions of
the actual three source tools, then runs stage-only, install/upgrade, repeat
installation, immutable dispatch, rollback, missing-dispatcher repair,
false-version startup refusal, recovery and uninstall through new processes.
Owned paths contain spaces; controlled child PATH resolves the intended tools,
registry PATH is unchanged and the original user skill/releases are preserved.
Unsigned source fixtures do not establish publisher authenticity.

Windows also executes the actual PowerShell installer in owned app data with
StageOnly, then repeated NoPath/SkipSkill/SkipFonts activation against synthetic
full-format resources. Initial hosted execution exposed PowerShell 7 module
paths inherited through Go, breaking Windows PowerShell's Utility autoloading.
A shared child environment filter now resets only that variable for native
PowerShell export/round-trip helpers and the installer harness. Actual Windows
hashing/JSON regression calls use an unusable inherited module path, with no COM.

Mac local actual-tool runs passed; preliminary `b3adbc57` passed all ten
private GitLab pipeline 21227 jobs and hosted Windows run 37607310782. Downloaded
Linux and Windows qualification JSON confirms exact source hashes, actual native
architectures and all eight/eleven declared scenarios. Earlier hosted Mac
execution also passed. Full tag-configuration simulation passed without creating
a tag; source work remains outside released v4.1.0.

CI now adds Intel Mac and ARM64 Windows native model/performance/core installer
checks, plus ARM64 model/installer jobs on the verified online home-lab pool.
Every new lane requires its declared Go host architecture. Generic diagnostics
expire after 14 days; binary fixtures/weights/customer content are not uploaded
to GitHub. Both Linux architecture results gate future protected release builds.
These additions and final full preflight/hosted gates are pending.
Real font/skill registration, inherited desktop PATH, signed Windows package
execution, private full resources, Office/visual review and operator curation
remain outstanding; no new release tag was cut.

Native architecture evidence follow-up: exact preliminary `c302b474` passed all
four GitHub model/performance lanes in run 37610005988 and all twelve GitLab
pipeline 21233 jobs, including native ARM64 installer/model execution. Eighteen
actual platform/mode reports were downloaded and verified; timings and scope
are retained in `docs/search-performance.md`. Windows ARM64 subsequently passed
all eleven actual CLI/script scenarios at `a8e117b3` in run 37611230623. Its
parallel same-head run timed out during the Windows PowerShell cmdlet regression, so bounded
startup diagnostics were added instead of discarding that result. Intel Mac's
missing output-parent setup was corrected. Exact `d9641533` passed the full
five-step slot submission preflight; final exact-head hosted qualification and
integration remain required. No further release tag was cut.

Windows module follow-up: both `d9641533` ARM64 failures reached the script,
with stages before hashing/before JSON serialization; a separate native Windows
PowerShell 5.1 file-entry probe completed in 235/242 ms. Explicit imports now
select native PSHOME Management/Utility manifests in export, round-trip,
installer and smoke scripts. Both helper launchers are checked through command
and file entry points, with native module-path evidence and unchanged 30-second
limits. Focused portable checks and Windows ARM64 compilation passed; hosted
execution remains required. Intel Mac's corrected actual-tool lane passed at
`d9641533`, with its downloaded eight-scenario qualification independently
verified. The source adds no desktop Office or operator acceptance claim.

## Derived-image closure and additional product scope

Managed slot `pptx-slide-assets` extends finished-slide publication/insertion
to selected image ancestry. It preserves original/intermediate/result bytes,
verifies exact derivation receipts and remaps every asset identity. Destination
receipts change only parent identity; the original bytes and both receipt hashes
remain in lineage and portable maintainer exports. Candidate compilation reads
bounded pending receipt bytes through the existing source guard.

Focused tests pass for two independent projects, native headless builds,
unchanged image/focus/operation/parameters, exact original receipts, maintainer
export, cycles/missing parents/unpaired metadata and false resealed packages.
Selected race checks pass. Full submission and hosted gates remain required;
real operator content and broad typed/local dependencies remain outstanding.

Ri added packaged browsing libraries and consistent portable projects on
2026-10-07. The linked portable-projects/browsing-decks scope retains the
requirements for both private CI-generated PowerPoints, full supported
template/frame/rail coverage, split source/assets/local templates, numbered
versions and colleague ZIP handoff. These are additional requirements, not
implemented commands.

Mac native evidence follow-up: a bounded Documents-folder attempt exposed a
visible PowerPoint Grant File Access prompt for the owned synthetic folder.
Access was granted to that folder through the native UI. A fresh fixture from
the current lineage contract then opened and saved through PowerPoint
16.113.4 (16.113.26100421). Independent on-disk verification passed all forty
native objects with zero lineage issues after Save As. Baseline SHA256
`8739f94533e397e4a4e44b66daf5e7bb293f7a942cfc6ca3370ee535a187a34c`;
saved SHA256
`6d5954171b63fa56891c2ace70c20e12e62540755cd8e38bab92485bd0e96dbc`.
Evidence remains under
`~/Documents/pptxgengo-qualification/save-as-_tqvwxdy/`; both owned task
presentations were closed. The older retained fixture is incompatible with the
new baseline generation contract and its failed verification is retained.
This proves this Mac Save As identity case only; three-field editing/reorder,
human ergonomics/visual review and Windows desktop qualification remain open.
