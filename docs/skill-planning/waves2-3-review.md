# Waves 2–3 integration and review

October 3, 2026. The operator authorized parallel implementation, smoke checks and independent review. All feasible implementation work in the plan is complete; live brand examples and operator voice refinement remain external acceptance items. PowerPoint-to-YAML reconciliation is a separate deferred workstream.

## Implemented

- Strict single-file `deck.yaml` loading and shared-template/local-composition compilation. Shared design resources remain versioned library definitions. Source locations, hard breaks, custom assets and original files are retained.
- Pinned compiler executable, Go/platform, typography engine, source, fonts, calibration and assets; stable slide/item IDs, deterministic native identifiers and timestamps. Repeated pinned project builds produce identical PPTX bytes.
- Build receipts, source/object maps, protected generated baselines, scoped approval invalidation, status/resume, bounded fork/detach and client/reviewer/maintainer/offline ZIPs. Generic supported source scenes can be detached; unsupported chrome or legacy typed scenes fail explicitly.
- Read-only native SQLite discovery across modern and legacy definitions. Scenario and content-shape hints rank independently; exact inspect/preview pins and actual-content alternatives keep source examples separate from new content fit. Legacy original rows are hydrated lazily from their pinned database. The measured unified projection fell from 113 MB to 60 MB.
- Progressive skill references for intake, narrative, source maintenance, selection, independent review and provisional personal structure/copy blended with local WM voice guidance.
- Immutable v3 design snapshot at `e91e0d7771000b7386f1ea52f51252f0f0a134fd`: 248 definitions, 247 active, 81 additions and revised `about/glance` against v2; no removals. Stable identities survive family-file moves. Five adapters cover logo slots, geography, team curves, devices and planes. V2 remains available for pinned projects.
- Native automatic image deduplication, conservative JPEG derivatives and DEFLATE. Shared frame layouts cache resolved chrome; per-slide copy and typography ownership remain distinct.

## Independent findings and fixes

Reviewers inspected code, ran targeted/race checks and built an isolated source-backed pilot. Material findings were repaired and rechecked:

| Finding | Repair / evidence |
| --- | --- |
| SVG MIME selected an unusable extension | MIME-aware project asset resolution and real compile regression |
| Rotated contained artwork used the requested box instead of displayed bounds | Contain geometry resolved before rotation; bounds regression |
| Tiny curve data produced degenerate/nonfinite control points | Relative tolerance and derived-coordinate finiteness validation |
| Custom SVG fallback accepted foreign namespaces or unsupported visible styling | Closed namespace, attribute, geometry and style validation; explicit unsupported errors |
| Offline archive omitted sibling calibration | Exact lock `runtime_files`, dependency copy and relocated v3 check/build regression |
| Dotted stable IDs disagreed across source and scene validation | Unified identity validation and exact authored-field pointer regression |
| Approval, fork target or ancestor freshness gaps | Dependency-bound approval checks and immutable ancestry/target checks |
| Same-family gallery could lend another template's preview contract | Exact canonical-key/revision/hash validation |
| Catalog path pins or duplicate candidate JSON were ambiguous | Bounded resource paths and duplicate-input rejection |
| Gallery could attach stale/duplicate bound content | Exact active keyset, source/bound revision/file/hash, output SHA checks before current-gallery mutation |
| Cleaned historical sample directories prevented release staging | Explicit prior-release recovery of only manifest-listed, hash-verified files |

The HTML JSON payload escaping has a preserving `<&></script>` round-trip check. The final form uses an explicit backslash character to avoid ambiguous Python literal interpretation. The narrow-label regression checks wave badges, API gateway and process chevrons.

## Native reference review

The full 248-slide source reference and 247-slide active binding reference compile successfully. The source deck opens in desktop PowerPoint and exports locally using **Best for printing**, preserving normal IBM Plex names. No Microsoft online conversion service was used.

Every source preview was observed: pages 1–83 and 84–167 individually at full resolution; pages 168–248 in four-page contact sheets with dense examples enlarged. Eleven pages required fixes:

- 18: API gateway label uses an explicit compact typography preset.
- 31–32 and 38–43: compact wave badges receive appropriate horizontal padding.
- 156: narrow process chevrons use an explicit small-label preset.
- 206: status table count column widened while retaining total table width.

All eleven repaired pages were re-exported and individually inspected at 1920×1080. The other 237 final PNGs are byte-identical to their reviewed predecessors. No obvious text/object overlap or clipping remains in these source specimens. This is bounded visual review, not an arbitrary-content capacity guarantee or pixel-perfect browser parity claim.

`release/verification-wmds-v3.json` records per-page template identity, preview hashes, reviewer, evidence basis, native PDF hash and package counts. The current local gallery has 248 hash-tied reviewed source previews; historical v2 paired evidence keeps its v2 labels.

## Size, package and performance

| Output | Slides | Size | Layouts | Media |
| --- | ---: | ---: | ---: | ---: |
| Final source reference | 248 | 31,762,716 bytes | 62 | 172 unique payloads |
| Active binding reference | 247 | 31,806,891 bytes | 62 | 172 unique payloads |
| Source PowerPoint PDF | 248 | 8,638,847 bytes | — | — |

The older 167-slide reference was 196,841,070 bytes. The new larger library is about 84% smaller. All 1,329 source-package entries use DEFLATE; all XML parses and all 1,874 internal relationships resolve. The active binding package likewise passes XML/relationship checks. Both use one physical slide master; resolved frames share layouts.

A full source build took 32.17 seconds with 1,845,477,376-byte maximum RSS and 2,096,023,520-byte peak footprint while a bound build and race checks ran concurrently. This measures the entire builder, unlike Wave 1's optimizer-only measurements. Full-build streaming and memory budgets remain future work; smaller output does not imply low peak build memory.

## Project/content pilot

The pilot used the two supplied operating-model/review documents as evidence for a two-slide internal recommendation. It includes actual-content `cards/3` versus `cards/4` alternatives, a deck-local custom composition and original SVG. The pilot distinguishes proposed workflow controls from verified implementation behavior.

Init/check/build, scoped approval, status/resume, changed meeting evidence invalidation, real-content alternatives and project exports were exercised. Maintainer and offline archives were extracted outside the original project and rebuilt; both yielded the exact original PPTX hash `47c0348d80c7a3cb8a6ed431891f2994f0b856ce587d4046e379b335241ee89a`.

An audience-only packet excluded author rationale and source-verification material. Four comprehension findings were preserved with author responses and repaired source; the same independent reviewer accepted the revised copy. The reviewer had earlier implementation context, and the session thread limit prevented a completely fresh agent session. The packet boundaries and this limitation are explicit. No operator content approval was fabricated. The first revised four-card alternative correctly failed measured fit (201pt needed, 198pt available); the failure is retained. A revised four-group arrangement separates audience review from source verification without shrinking fonts. Both final actual-content alternatives pass Go fit; only the selected two-slide pilot has native visual review.

Both final pilot pages were opened read-only and individually observed in desktop PowerPoint at 122%: supplied copy, local composition and custom artwork were contained without visible overlap/clipping. The reference library was restored to slide 80 afterward. Technical, source and audience findings remain separate artifacts.

## Checks

- Full focused package checks pass: `internal/wmdesign`, `internal/deckproject`, `cmd/pptxdesign`, `cmd/pptxgengo`.
- Race checks pass: wmdesign 52.736s; deckproject 52.683s; pptxdesign 4.702s.
- Narrow-label, intake adapter, contained-rotation, custom-SVG, offline-relocation, shared-frame and deterministic-build regressions pass.
- Skill-creator `quick_validate.py` passes with cached PyYAML. Installer shell syntax passes.
- Broad `pptx` tests still have the twelve pre-existing failing functions reproduced at committed HEAD and documented in `wave1-review.md`. The whole repository suite is not claimed green.

## Sharing and retention

Only the current reference deliverables live in `samples/wmds-production-20261003/latest/`. Intermediate experiments remain temporary. Project builds are immutable reconciliation baselines: preserve current state/lock, original assets, ancestry/decisions, reviewer findings and the build baseline with any externally edited deck. Do not prune project history automatically. A maintainer archive is the handoff artifact; a client archive excludes internal evidence. Offline archives retain their exact platform/compiler dependencies.

PDF export is currently a separate native review/delivery step, not an integrated project build artifact. Automated edited-PPTX reconciliation, PNG photo derivatives, color-managed image conversion and full-build streaming are deferred. Live brand main page/three example tabs remain unavailable under browser policy; provisional voice examples await operator refinement.

## Release

Local.7 packages the aligned CLI, skill, YAML starter/schema, v3 bundle, fonts/calibration, original artwork, unified database and current gallery. Installation follows staged integrity and outside-repository smoke checks; activation evidence is added below when complete. Older project locks are not silently migrated: use an explicit reviewed re-pin and preserve the prior lock/baseline, or continue using the pinned release.

### Activation evidence

Local.7 was staged and its release manifest verified before activation. Outside-repository modern/legacy find, inspect and preview; YAML init/check/build/status/resume; and client/maintainer/offline export checks pass. The packaged starter rebuilds byte-identically after both maintainer and offline relocation (`129db26dd6a12e8e5e4279dfccf19873a648944db506b80a7f63f2a30156e03c`). The reviewed staged directory moved to `~/.local/share/pptxgengo/releases/0.1.0-local.7`; launcher and skill symlinks now align. No commit or push was performed. Newer upstream additions observed during this run remain a separate intake, recorded in the plan.
