# Engineering follow-up qualification

Scope: the 18 observations in `docs/skill-planning/engineering-followups.md`,
following the local.12 survey. Import/reconciliation, a new preview renderer,
wireframes and presentation-skill authoring remain outside this work.

## Implementation disposition

| Ask | Implemented | Remaining limit |
| --- | --- | --- |
| 1. Independent reviewer packet | `review` / `view --audience`; displayed shape/table copy in HTML; audience context only; hidden slides, notes, briefs, rationale and earlier verdicts excluded. Internal `context.win_strategy` supported. Audience deck stage requires current visible-page native output. Copy-only stages do not require trusted historical render evidence. | Native chart text has an explicit rendered-page review gap; nondisplayed workbook data is excluded. Native pages still require an eligible desktop caller. |
| 2. Content matching | Ranked `needs_copy` drafts separated from ready slides; first-N group fills; nested lifecycle activities; rectangular scorecard comparison; sequence owner/duration/state fields when declared editable. Ready output requires a successful Go build. | Bounded source-topology adapters; unknown/intersection relationships and arbitrary complex nesting remain explicit gaps. |
| 3. Native export failure handling | Grant File Access monitoring, failure diagnostics including `render-error.txt`, operational doctor file-access probe, deadline cleanup. | Doctor cannot repair a broken GUI caller. No new live export qualified in this session. |
| 4. Template swap | Different item counts, headline/statement bridge, identical-template paragraphs/panels, explicit leftovers/missing copy; apply receipt identifies `swap`. | Unmapped content requires explicit authorization; no arbitrary semantic rewrite. |
| 5. Native receipt integrity | Local Ed25519 issuance signature; strict receipt and decisions JSON; source/build/output hash checks retained. | Local issuance is not OS attestation. Same-user code can access the key; cross-machine trust transfer has no command. Older unsigned receipts require rerendering. |
| 6. Re-add removed source | Reuses a retained source only when its supplied bytes match; verifies drift under mutation guard. | Divergent destination is rejected rather than overwritten. |
| 7. Slide identity | `slide add --as NEW-ID` preserves the input file. | Standalone rename is unnecessary for this insertion workflow and is not implemented. |
| 8. Earlier fit check | Add/edit `--check-fit`; scaffold rejects invalid stable IDs immediately. Failed fit checks leave sources unchanged. | Go layout checks; native appearance is a separate step. |
| 9. Section placement | Add/move `--into-section`; inserting at the beginning updates the section anchor and reports it. | Moving an existing section anchor still needs explicit `--reanchor`. |
| 10. Readable detach | Friendly nested content/bindings and local-zone aliases; stale stock header removed. | A detached composition remains local and must retain its justification. |
| 11. Capacity comments | Regenerated on shared-slide edit/swap; human head/line/foot comments preserved on retained paths. | Comments for deleted content do not survive its removal. |
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
| `make test`, empty branding root | Pass, 24.50–28.76 seconds wall time; latest run includes final CLI parsing and fixture-isolation regressions |
| `make test-integration INTEGRATION_TEST_TIMEOUT=3m` | Pass, 30.06 seconds; final rerun 46.81 seconds while the exhaustive race suite ran concurrently; includes full catalog and final relocated projects |
| Cache race regressions | Pass, 53.20 seconds; includes concurrent callers, clone isolation, eviction, source/font/calibration drift and source override |
| `make test-race`, empty branding root | Pass, 77.84 seconds wall time; all four focused developer checks |
| Full race lane, final 5-minute per-package ceiling | Incomplete: design-library package timed out at 301.78 seconds in `TestV5DeltaSourceQualification`; no reported data race. All other packages passed, including project 174.02 seconds. Remains out-of-band. |
| CI YAML syntax / installer shell syntax | Pass; remote CI execution and action-specific lint not run |

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

An isolated checkout keeps unrelated in-progress skill and DentalXChange edits
out of the engineering qualification. Those working-tree changes are preserved.
No new Python conversion scripts were introduced.
