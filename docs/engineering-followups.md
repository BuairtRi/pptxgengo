# Engineering follow-up qualification

Scope: the 18 observations in `docs/skill-planning/engineering-followups.md`,
following the local.12 survey. Import/reconciliation, a new preview renderer,
wireframes and presentation-skill authoring remain outside this work.

## October 5 publication and photography extension

The repository default and single production gallery/index are now v7: **616
templates** and **1,204 asset variants**. Photo registration covers **all 521**
originals and matching metadata sidecars under West Monroe Photos, preserving the
20 earlier IDs. Search uses pinned sidecar semantics and dimensions; selected
originals are verified for preview, rendering and packaging. See
[discovery and maintenance commands](semantic-template-discovery.md#photography-discovery).
The Go `photo-register` command replaces manual registry additions. Originals and
sidecars are unchanged; no fresh human visual review is claimed for those descriptions.

Publication verifies 30 new/changed native specimens and 586 inherited renderer
outputs. The global CLI is now **0.1.0-local.16 / v7**, installed with `--cli-only`
to preserve the presentation skill link. Outside-checkout photo discovery and
all installed package hashes were verified. The release commit excludes the
other agent's presentation skill and DentalXChange slide edits.
The earlier qualification tables below retain their original version/count scope.

## Implementation disposition

| Ask | Implemented | Remaining limit |
| --- | --- | --- |
| 1. Independent reviewer packet | `review` / `view --audience`; displayed shape/table copy in HTML; audience context only; hidden slides, notes, briefs, rationale and earlier verdicts excluded. Internal `context.win_strategy` supported. Audience deck stage requires current visible-page native output. Copy-only stages do not require trusted historical render evidence. | Native chart text has an explicit rendered-page review gap; nondisplayed workbook data is excluded. Native pages still require an eligible desktop caller. |
| 2. Content matching | Ranked `needs_copy` drafts separated from ready slides; first-N group fills; nested lifecycle activities; rectangular scorecard comparison; sequence owner/duration/state fields when declared editable. Ready output requires a successful Go build. | Bounded source-topology adapters; unknown/intersection relationships and arbitrary complex nesting remain explicit gaps. |
| 3. Native export failure handling | Grant File Access monitoring, failure diagnostics including `render-error.txt`, operational doctor read/PDF-write probe in the renderer's stable staging folder, deadline cleanup. | Doctor records observed access for that folder; it cannot repair a broken GUI caller or guarantee future access. See the regression qualification below for live-export status. |
| 4. Template swap | Different item counts, headline/statement bridge, identical-template paragraphs/panels, explicit leftovers/missing copy; apply receipt identifies `swap`. | Unmapped content requires explicit authorization; no arbitrary semantic rewrite. |
| 5. Native receipt integrity | Local Ed25519 issuance signature; strict receipt and decisions JSON; source/build/output hash checks retained. | Local issuance is not OS attestation. Same-user code can access the key; cross-machine trust transfer has no command. Older unsigned receipts require rerendering. |
| 6. Re-add removed source | Reuses a retained source only when its supplied bytes match; verifies drift under mutation guard. | Divergent destination is rejected rather than overwritten. |
| 7. Slide identity | `slide add --as NEW-ID` preserves the input file. | Standalone rename is unnecessary for this insertion workflow and is not implemented. |
| 8. Earlier fit check | Add/edit `--check-fit`; scaffold rejects invalid stable IDs immediately. Failed fit checks leave sources unchanged. | Go layout checks; native appearance is a separate step. |
| 9. Section placement | Add/move `--into-section`; inserting at the beginning updates the section anchor and reports it. | Moving an existing section anchor still needs explicit `--reanchor`. |
| 10. Readable detach | Friendly nested content/bindings and local-zone aliases; stale stock header removed. | A detached composition remains local and must retain its justification. |
| 11. Capacity comments | Regenerated once on shared-slide edit/swap after stripping prior generated blocks from every YAML comment position; human head/line/foot comments preserved on retained paths. | Comments for deleted content do not survive its removal. |
| 12. Image delivery size | Existing project policy confirmed with registered large-image regression: delivery derivatives and receipt, unchanged original; explicit resize-disable respected. | No new compression default; unsupported profiles/types retain originals with reported reason. |
| 13. Flow YAML diagnostics | Detects likely unquoted commas with source line/column and quoting example. Explicit null remains valid. | Heuristic handles phrase-like accidental keys; ordinary YAML parser owns syntax errors. |
| 14. Search result quality | Drops queried zero-score rows by default; `--include-weak` opt-in; groups icon variants before limiting, retaining registered variant IDs. | Existing semantic ranking remains metadata based. |
| 15. Asset discovery | Ranked OR/partial matches and synonyms; filename tags across 222 icon concepts / 666 color variants; six visually inspected additions bring the photo registry to 22. | Local WM sidecars do not map renamed photos to stock source IDs/URLs or license records; approval traceability remains unresolved. |
| 16. Common aliases | Source-derived recipes for statement, status columns and dense interview fields; lifecycle and scorecard topology. 48 recipe templates; 429 still ambiguous. | Zero templates have a recorded human semantic review. This is a bounded engineering improvement. |
| 17. Component capacity | Actual renderer/font plans add 14 typed-card slots, 773 plain table/header slots and 25 vertical-stepper description slots. | 14,580 slots remain unsupported; rich/status tables and variable component internals need further work. Estimates are advisory and do not establish native fit. |
| 18. Documentation | Re-pin procedure, dispatcher defaults, ZIP versus staged review, dependency staleness, claims behavior, section removal, and new operations documented; machine-specific recovery moved to its own appendix. | Recovery procedure and remote CI execution remain unverified. |

## Test-cycle repair

Repeated full-library decoding was the main race-run bottleneck. The catalog now
uses a bounded synchronized cache keyed by actual verified source dependencies,
including fonts. Callers receive isolated deep copies. Load and drift validation
remain in place; failed catalog builds are not cached.
Library-index tests also build one immutable database seed and copy it into each
test's own SQLite file. Per-copy metadata and report containers remain isolated;
corruption in one fixture does not affect another. Independent legacy and
gallery-build fixtures continue to exercise their real builders.

`make test` is the everyday short lane. `make test-race` covers mutation guards,
native worker cleanup, concurrent PPTX serialization and the catalog cache.
Full fixture and race lanes remain explicit and are configured for out-of-band
CI. See [testing lanes](testing.md). Repository CI files are present; enabling
full jobs requires a dedicated runner and `PPTXGENGO_FULL_TESTS=true`.

## Qualification

Consolidated qualification on macOS arm64 with Go 1.27.1:

| Check | Result |
| --- | --- |
| `go build ./...` | Pass |
| `make test`, empty branding root | Pass, latest clean-checkout run 23.96 seconds; includes final CLI parsing, fixture isolation and all 587 frozen-preview dependencies |
| `make test-integration INTEGRATION_TEST_TIMEOUT=3m` | Pass, 30.06 seconds; final rerun 46.81 seconds while the exhaustive race suite ran concurrently; includes full catalog and final relocated projects |
| Cache race regressions | Pass, 53.20 seconds; includes concurrent callers, clone isolation, eviction, source/font/calibration drift and source override |
| `make test-race`, empty branding root | Pass, 77.84 seconds wall time; all four focused developer checks |
| Full race lane, final 5-minute per-package ceiling | Incomplete: design-library package timed out at 301.78 seconds in `TestV5DeltaSourceQualification`; no reported data race. All other packages passed, including project 174.02 seconds. Remains out-of-band. |
| CI YAML syntax / installer shell syntax | Pass; remote CI execution and action-specific lint not run |
| Packaging closure / private-task dialog monitor, focused race | Pass; native 1.31 seconds, gallery 2.39 seconds |

An earlier three-minute full-race attempt timed out in the design-library package
while rebuilding SQLite fixtures; the project package passed in 155.63 seconds.
The immutable fixture seed removes repeated database construction without dropping
assertions. The later five-minute attempt reached the gallery/source qualification
sweeps but still did not complete. The configured exhaustive CI lane has a
10-minute per-package ceiling. Neither timeout is recorded as a successful race
qualification. The normal and focused developer lanes do not depend on this
exhaustive run.

The first consolidated run detected a stale maintained SQLite asset-registry
fingerprint after the photo additions. The unified index was regenerated;
subsequent qualification uses the current index (587 templates, 703 asset
variants/registry entries). Frozen source definitions and native specimens did
not change.

Cache benchmarks after qualification: full uncached derivation 448 ms/op, warm
isolated clone 5.6 ms/op, fully hash-validated load 19.3 ms/op. These local timings
are indicative, not performance guarantees.

Go renderer/serialization and synthetic native receipt tests do not constitute
a live PowerPoint export or visual acceptance. The actual doctor completes in
2.89 seconds with the existing −10827 dispatch failure and unknown automation/
file access; it does not issue a file-open probe when operational dispatch fails.
Source preflight and malformed-duration failures were checked with the actual
CLI and leave `render-error.txt`. No permission or daemon changes were made.

Clean-checkout packaging caught one dependency gap: the generic `build` ignore
rule had excluded the `phase/build` specimen's four gallery files. Those existing
hash-pinned files are now tracked, with a full 587-specimen closure regression.
Native command monitoring is also restricted to absolute private task scripts;
inline doctor AppleScript no longer writes `dialog.swift` into the caller's
working directory. The stray diagnostic file was removed.

## Prior local.13 qualification

Global **0.1.0-local.13** is installed at
`~/.local/share/pptxgengo/releases/0.1.0-local.13`, through
`~/.local/bin/pptxgengo`. Binaries were built from clean committed runtime code
`92d2d66661ce4913fd8a89f0ad575616ee5362d1`, with `vcs.modified=false`.
The subsequent correction permits the deprecated specimen's absent optional
source-values link in the new closure test; it changes no CLI runtime code.

The installer validates 587 source specimens, 1,760 linked artifacts and 3,864
package files. Outside-repository checks pass init/build, audience content review,
new software-photo search, section-aware add with `--as` / `--check-fit`, split
source and retained-file re-add, and malformed render-flag diagnostics. The real doctor
still reports dispatch failure and unknown permission, and creates no stray
working-directory script. The existing presentation-skill link is preserved.

An isolated checkout keeps unrelated in-progress skill and DentalXChange edits
out of the engineering qualification. Those working-tree changes are preserved.
No new Python conversion scripts were introduced.

## Round 2 comment and staging regression repairs

The capacity-comment regression came from YAML rereads moving generated blocks
onto adjacent keys and map footers, and retaining `#` prefixes in parsed comment
text. Refresh now strips complete generated blocks from every head/line/foot
position before adding current template metadata once. It repairs accumulated
blocks on the next shared-slide edit or swap and preserves human comments on
retained paths. Tests exercise three identical disk edits, cards/4 to cards/3
(body estimate changes from ~104 to ~148 characters), removed-slot cleanup,
already-duplicated comments, and byte-stable repeated same-template swaps.

Doctor formerly opened a copy in one random task subfolder while render used a
different subfolder, and never tested PDF writing. Both now place unique task
PPTX/PDF files directly in the same stable staging root and use the same native
export script/argument factory. Doctor opens, writes a one-slide PDF, closes the
exact copy, and requires a PDF header before reporting observed file access.
Scripts and PNG intermediates stay private. Exclusive name acquisition and
retained inode identity protect task cleanup; confirmed tasks remove only their
own files. Regression tests cover folder parity, failed PDF writes, collisions,
symlinks, missing metadata, retained tasks and legacy nested cleanup.

Independent peer review found no remaining known code defects. The clean-checkout
fast suite passes in 31.20 seconds; comment regressions pass under the race detector (15.157s),
the full native package passes normally (3.833s) and under race (7.591s), and the
AppleScript compiles. Live doctor and render attempts still fail with this
caller's operational Apple-event `-10827` before opening the task copy. No Grant
dialog was observed. The computer-use tool rejected control of Warp for safety
reasons, so that desktop-terminal fallback was not used. Native PDF reservation
overwrite and actual visual export remain unqualified; no new native acceptance
is recorded. Task scratch artifacts are removed; open original decks are preserved.

Global **0.1.0-local.15** is installed from clean runtime commit
`962c03a335099578d1742046012639c70fd069c0` (`vcs.modified=false`). The installer
verified all 587 previews, 1,760 gallery links and 3,864 package files. The installed
CLI kept three identical cards/4 edits at 59 lines with identical SHA-256 hashes;
cards/4 to cards/3 changed the body estimate from ~104 to ~148, and three repeated
same-template swaps remained byte-identical at 51 lines. Installed doctor and
review-deck render still report operational dispatch `-10827`; the failed render
wrote `render-error.txt` and issued no success manifest. The presentation-skill
link is unchanged.

## Restart, native review and workshop intake — 2026-10-05

After reboot, the installed local.15 CLI exported all 16 heat-map review pages
successfully, producing PDF, PNGs and a signed receipt. PowerPoint overwrote the
CLI's reserved task PDF without a Replace prompt. That initial receipt applies
to the pre-repair deck, not the final accepted deck.

Native inspection exposed A1–A9 badges and LATER chips splitting despite passing
Go measurement, plus narrow single-word heat headers wrapping. The v6/v7 width
repair adds measured native allowance to reference/priority labels and reduces
heat-header padding, with a bounded 9→8 pt reduction when the original score
column cannot fit the full header. v5 geometry and frozen source bytes remain
unchanged. All **16/16 repaired heat-map specimens** pass native visual review.

The initial live doctor dispatched to PowerPoint but returned `-9074` for its
minimal probe. The revised doctor supplies explicit slide/text geometry, Arial
fonts, text sizing and a white background. Package/XML and focused race tests
pass. Fresh automated doctor/render checks under the now restricted caller fail
Apple-event dispatch `-10827`; no revised live doctor pass is claimed. No Grant
File Access dialog was observed. Final visual checks used PowerPoint's local
**Export PDF → Best for printing** and the existing Swift PDF rasterizer, with
unsigned review manifests distinct from automated CLI receipts.

The new pinned v7 adds **14 workshop templates**, bringing the source total to
**616**. All 14 pass editable source/bound round trips and native visual review.
Five initial allocation failures required geometry amendments. The first native
pass found three defects (Open questions clearance and two status column wraps);
all were repaired and independently rereviewed. All 602 earlier definitions and
bindings are retained. Full Go builds pass for 616 source and 615 active bound
slides in **89.106s**. The final everyday repository suite passes in **120.61s**;
exhaustive catalog builds remain excluded from that short lane.

SQLite creation and workshop search pass for 616 templates / 703 asset entries.
Discovery previously rejected the candidate's legitimate inherited source links;
a source-only resolver now verifies registered bundle/inventory pins, the source
relative path and actual file hash. Generic resource path restrictions remain in
place. Focused regressions pass normally (16.954s) and under race (80.035s).

Final evidence is consolidated under the
[602-template intake](../planning/wm-design-contracts/v6/intake-20261004-602-frozen/README.md)
and [616-template workshop intake](../planning/wm-design-contracts/v7/intake-20261005-616-frozen/README.md).
Working candidate decks and generated full catalog projections were deleted.
Temporary review windows were closed; the user's two DentalXChange decks were
preserved. No Python conversion helper was added.

Production remains **v5 / 587** and global **local.15**. Candidate gallery/index
promotion is pending. Current workspace permissions exclude `.git` writes and
the global install directory, so this turn's changes are uncommitted and not
installed globally. The earlier clean local.15 installation remains intact.
The existing repository `./pptxdesign` was rebuilt from the final source; its
explicit-v7 discovery smoke check passes. Both root-owned scratch trees were
removed after final evidence hash checks and confirming the PowerPoint Window
menu listed only the user's two DentalXChange decks.
