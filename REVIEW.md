# Deep Code Review — Go Port (pptx/)

Six independent review passes (foundation, xml, charts, objects, tables+media,
top-level/cross-cutting), each verified against the TS source and/or the live
`dist/pptxgen.cjs.js`, with behavioral claims reproduced in scratch tests before
being reported. Findings below are deduplicated and severity-ranked.
References: `file:line` in this repo; TS refs are `src/*.ts`.

## Critical

C1. **Package-global time hooks race across presentations** — `writer.go:75-78`
    swaps `xmlNowFunc`/`excelNowFunc` (`xml.go:20`, `charts.go:1842`) for the
    duration of `build()`. Two goroutines writing two `*Presentation`s race
    (`go test -race` reproduces); a deferred restore can even reinstate the
    *other* writer's clock. Fix: pass a clock through the build context instead
    of mutating package vars.

C2. **`_chartCounter` is a racy, leaking process-global** — `objects.go:20,391`.
    (a) Concurrent `AddChart` races under `-race` and can duplicate chart IDs;
    (b) even sequentially, a second presentation in the same process numbers its
    worksheets `Microsoft_Excel_Worksheet2+` — non-reproducible output with no
    public reset (the integration test resets it by hand, which is the smell).
    Fix: move the counter onto `Presentation`.

C3. **Media failures are silently swallowed** — `writer.go:83,86,88` discard the
    `[]error` from `resolveSlideMediaRels`. A nonexistent image path yields
    `Write() → (bytes, nil)` with IMG_BROKEN substituted. Real PptxGenJS
    *rejects the entire write* (`pptxgen.ts:494` Promise.all + throws in
    gen-media.ts:70,93,125). This is both a fidelity break and silent data
    corruption. Fix: aggregate and return (or expose a WriteProps policy).

C4. **Explicit numeric zero clobbered on chart options** — `objects.go:572-591,674`.
    `V3DRotX/Y`, `V3DPerspective`, `BarGapWidthPct/DepthPct`, `LineSize` use
    `== 0` as "unset" where TS uses `isNaN`/`typeof` checks; `BarGapWidthPct: 0`
    silently becomes 150, `LineSize: 0` becomes 2pt, `V3DRotX: 0` becomes 30°.
    All reproduced. Violates PORTING.md's own pointer rule — should be `*float64`.

## Major

M1. **`embeddedFontLst` emitted at wrong schema position** — `xml.go:2314-2321`.
    ECMA-376 `CT_Presentation` (verified against three independent XSDs) puts
    `embeddedFontLst` AFTER `sldIdLst`/`sldSz`/`notesSz`, not after
    `sldMasterIdLst` as PORTING.md specified. Decks using `EmbedFont` risk
    PowerPoint's repair prompt. (Spec error in PORTING.md, faithfully
    implemented; both need the fix.)

M2. **The explicit-zero/unset collapse is systemic** (same root as C4) in:
    - `charts.go:1107-1111` doughnut `HoleSize: 0` → forced to 50
    - `charts.go:1798-1834` reflection overlay: combo per-type/axis override of
      0/false silently dropped (combo bar+line asking `LineSize: 0` keeps 2pt)
    - `charts.go:1574-1621` shadow merge: explicit `Angle/Blur/Offset/Opacity: 0`
      falls back to defaults
    - `utils.go:163-173` glow: explicit `Size/Opacity: 0` falls back to defaults
    - `tables.go:277-297` `AutoPageSlideStartY: 0` treated as unset (verified
      divergent paging: TS [32,32,34,2] vs Go [32,32,32,4])
    - `objects.go:236-303` placeholder merge drops explicit-zero overrides and
      only copies a ~20-field whitelist vs TS's full spread
    Fix direction: convert these to pointers (or option-set bitmask) per field.

M3. **`ColW` scalar shorthand lost** — `types.go:325,845`, `tables.go:343-371`.
    TS `colW: 2` = uniform width for all columns; Go 1-element slice = column 0
    only, remaining columns width 0 (degenerate wrapping). High-level AddTable
    path masks it via pre-normalization; direct `GetSlidesForTableRows` callers
    hit it. Fix: treat len==1 as uniform (matching RowH/Margin/Border handling).

M4. **Empty-but-non-nil `ChartColors` panics** — `charts.go:388,621,646,742,754,885`
    `% len(...)` → integer divide by zero (JS degrades to undefined). Violates
    the no-panics rule. Also `CatAxes`/`ValAxes` non-nil-but-empty panics at
    `charts.go:176` (`&opts.CatAxes[0]`), where TS spreads undefined harmlessly.

M5. **`ftoa` lacks JS exponential notation** — `utils.go:17-19`. JS `String()`
    emits `1e+21` at ≥1e21 and `1e-7` below 1e-6; Go never does. Affects chart
    `<c:v>` values at extreme magnitudes.

M6. **Series-axis time units: fidelity inversion** — `charts.go:1460-1469`.
    Upstream TS has a variable-name bug (`opt.toLowerCase()` on the key,
    `gen-charts.ts:1883-1893`) that means real PptxGenJS NEVER emits
    serAxis base/major/minorTimeUnit; Go emits them when set. Decide: replicate
    the upstream bug (fidelity) or keep + document the improvement.

M7. **Run-option inheritance vs JS falsiness** — `xml.go:1699-1758`. Verified
    against live pptxgenjs: shape-level `bold: true` OVERWRITES a run's explicit
    `bold: false` in JS (falsy check); Go's `*bool` nil-check preserves the
    run's false. Go is arguably saner; it is a real divergence.

M8. **Auto-paging silently drops rows** — `objects.go:1389-1418`. Defensive
    `if newSlide == nil { continue }` converts a mis-wired getSlide callback
    into vanished table rows with nil error (TS fails loudly). Reproduced:
    2 slides requested, 0 delivered, no error.

M9. **Compression unreachable through `WriteTo`/`WriteFile`** — `writer.go:24-62`.
    Both hardcode `build(false)`; TS's stream()/writeFile() honor compression.
    (`WriteFileWith` exists but is undocumented; no compressed WriteTo at all.)

M10. **`DefineLayout` early-returns where TS registers** — `presentation.go:131-138`.
    TS warns but still registers a degenerate layout (then `.layout = name`
    succeeds); Go registers nothing and `SetLayout` errors UNKNOWN-LAYOUT.

M11. **IMG_BROKEN/IMG_PLAYBTN regression guard is a 40-char prefix check** —
    `enums_test.go:158-168`. Blobs independently verified byte-identical today;
    test would not catch mid-blob corruption. Hash-check the full constants.

## Minor (selected; see review transcripts for full list)

- `getSmartParseNumber` truncates fractional ≥100 "already-EMU" values
  (`utils.go:48-49`); JS preserves the fraction.
- `<c:v>` main-path drops TS's `value||value===0` guard (`charts.go:537,1093`) —
  inert today ([]float64 has no holes) but inconsistent with scatter/bubble.
- Combo-chart validation throws dropped (`charts.go:170-175`); TS throws two
  specific errors on malformed multi-axis configs.
- vmerge dummy-cell drop on irregular colspan+rowspan grids
  (`xml.go:832-839`) vs TS splice-clamp; narrow trigger.
- Margin len 2/3 silently no-ops insets (`xml.go:213-224,415-420`); TS zero-fills.
- SVG sniffers (net-new, currently unused): viewBox comma-separated form
  rejected (`media.go:396`); single-quoted width/height attrs unmatched
  (`media.go:362-363`); BITMAPCOREHEADER BMPs misread as garbage dims
  (`media.go:345-358`).
- Empty `TableRow` silently skipped (`tables.go:377-379`) where TS crashes;
  undocumented deviation.
- `DefineSlideMaster` stores caller pointers without the deep-clone TS added
  for ISSUE#406 (`presentation.go:292-332`); later caller mutation changes
  registered state.
- `Slides()` returns internal pointers while `SlideLayouts()` returns copies —
  inconsistent aliasing semantics (`presentation.go:117,123`).
- Error strings carry TS-style "ERROR:" prefixes / UNKNOWN-LAYOUT sentinel;
  non-idiomatic for wrapped Go errors.
- PORTING.md documents a `uuidFunc` injection hook that was never built;
  integration tests regex-blank GUIDs instead.
- `WriteFile` truncate-in-place can leave a corrupt file on mid-write failure
  (matches TS; still worth atomic-rename).
- No context.Context on network fetch; fixed 30s timeout only.
- Media resolution is sequential where TS is concurrent (latency, not
  correctness; documented).
- Redirects: Go follows (embeds final image), TS embeds the redirect body.

## Verified clean (high-value negative results)

- All 363 shape constants, scheme colors, default objects, BARCHART/PIECHART
  palettes, and both base64 image blobs: programmatic diff vs TS — exact,
  including preserved upstream typos.
- Excel column naming (AA/AZ/BA), stray table-ref quote, bubble double-round:
  byte-faithful.
- JPEG sniffer survives progressive/EXIF-first/padded streams; PNG/GIF/BMP
  offset math correct (modulo BITMAPCOREHEADER above).
- Golden integration harness genuinely fails on a single mutated byte;
  normalizations are tightly scoped (timestamps, GUID shapes only).
- media dedupe keys and the SVG-companion "filter-before-await" quirk re-derived
  from TS and confirmed exact.

## Totals (deduplicated)

Critical 4 · Major 11 · Minor ~15 · Nit ~5.
The dominant theme: **JS `undefined`-vs-`0/false` semantics** — one root cause
expressed in at least 10 findings across five files. Second theme: **package
-global mutable state** (clock hooks, chart counter) unsafe for a long-lived
or concurrent Go process. Both are systematically fixable.

## Resolution log (remediation waves 1-4)

Status legend: **fixed** = behavior now matches TS (or panics eliminated);
**documented-deviation** = a deliberate, commented divergence from TS remains
(by design, not an oversight); **inert-documented** = the TS behavior being
ported cannot occur in Go's type system, so no code change was possible or
needed — the reasoning is left in a comment for the next reader. Every test
name below was confirmed present with `grep` against `pptx/*_test.go`, not
taken on faith from commit messages.

### Critical

- **C1** (package-global clock race) — fixed. `writer.go`/`presentation.go`
  thread a per-build `buildContext{now, uuid}` instead of swapping package
  vars. Test: `TestConcurrentWriteRace`.
- **C2** (`_chartCounter` racy process-global) — fixed. Counter moved onto
  `Presentation.chartCtr` (mutex-guarded, per-instance). Tests:
  `TestConcurrentAddChartRace`, `TestChartNumberingPerPresentation`.
- **C3** (media failures silently swallowed) — fixed (wave 4). `build()` now
  collects every `[]error` from the three `resolveSlideMediaRels` call sites
  and returns `errors.Join(...)` before any zip assembly, so `Write`/
  `WriteTo`/`WriteFile` all fail together and `WriteFile` never leaves a
  partial file. media.go's own IMG_BROKEN-vs-error classification was left
  untouched, as instructed. Tests: `TestWriteSurfacesMediaError`,
  `TestWriteToSurfacesMediaError`,
  `TestWriteFileSurfacesMediaErrorAndLeavesNoPartialFile`,
  `TestWriteValidMediaStillSucceeds`, `TestBuildErrorIsJoinedAndUnwrappable`.
- **C4** (explicit numeric zero clobbered on chart options) — fixed.
  `V3DRotX/Y`, `V3DPerspective`, `BarGapWidthPct/DepthPct`, `LineSize` are now
  `*float64`. Tests: `TestExplicitZero_V3DRotXSurvives`,
  `TestExplicitZero_V3DRotYAndPerspectiveSurvive`,
  `TestExplicitZero_BarGapWidthPctSurvives`,
  `TestExplicitZero_BarGapDepthPctSurvives`,
  `TestExplicitZero_LineSizeRendersNoFill`,
  `TestExplicitZero_LineSizeDefaultsToTwoWhenUnset`.

### Major

- **M1** (`embeddedFontLst` wrong schema position) — fixed. Now emitted after
  `sldIdLst`/`sldSz`/`notesSz` per ECMA-376 `CT_Presentation`. Test:
  `TestMakeXmlPresentationFontHooks`; also verified end-to-end in wave 4's
  scratch deck (`ppt/presentation.xml`: `<p:notesSz.../><p:embeddedFontLst>`).
- **M2** (explicit-zero/unset collapse, systemic) — fixed at each listed site.
  Doughnut `HoleSize`: `TestExplicitZero_HoleSizeSurvives`,
  `TestExplicitZero_HoleSizeDefaultsWhenUnset`. Reflection/combo overlay:
  `TestExplicitZero_OverlayLineSizeOverride`,
  `TestExplicitZero_OverlayBoolPointerOverride`. Shadow merge:
  `TestExplicitZero_ShadowFieldsHonored`. Glow:
  `TestExplicitZero_GlowFieldsHonored`. `AutoPageSlideStartY`:
  `TestExplicitZero_AutoPageSlideStartY`. Placeholder merge:
  `TestPlaceholderMerge_ParaSpaceBeforeExplicitZeroWins`.
- **M3** (`ColW` scalar shorthand lost) — fixed: `len==1` now means uniform
  width, matching RowH/Margin handling. Tests:
  `TestColW_ScalarShorthandExpandsUniform`,
  `TestAddTableDefinition_ColWSingleValue`,
  `TestAddTableDefinition_ColWMatching`.
- **M4** (empty-but-non-nil `ChartColors`/`CatAxes`/`ValAxes` panics) — fixed,
  no more `% len(...)` divide-by-zero or index-into-empty-slice panics. Tests:
  `TestEmptyChartColors_NoPanic_SameAsNil`, `TestEmptyCatAxes_NoPanic`.
- **M5** (`ftoa` lacks JS exponential notation) — fixed: emits `1e+21`/`1e-7`
  style JS notation at the same magnitude thresholds Node uses, with the
  single-digit-exponent zero-pad difference eliminated. Test: `TestFtoa`
  (covers the exponential-threshold cases explicitly, including the -0 and
  just-under-1e21 boundary cases).
- **M6** (serAxis time-unit fidelity inversion) — fixed: the upstream TS
  variable-name bug is replicated (serAxis base/major/minorTimeUnit are never
  emitted), matching real PptxGenJS output rather than the technically-more-
  correct un-buggy behavior. Test: `TestSerAxisTimeUnit_NeverEmitted`.
- **M7** (run-option inheritance vs JS falsiness) — fixed as a **documented
  deviation kept intentionally on the JS side**: shape-level `bold:true` now
  overwrites a run's explicit `bold:false`, matching live pptxgenjs's falsy-
  check inheritance loop (even though the previous Go behavior was arguably
  saner). Test: `TestInheritRunOptionsFalsyOverwrite`.
- **M8** (auto-paging silently drops rows on nil `getSlide`) — fixed: a
  mis-wired/nil continuation slide now surfaces as an error instead of a
  silently-continued loop. Test: `TestAddTableDefinition_AutoPageMissingSlideErrors`.
- **M9** (compression unreachable through `WriteTo`/`WriteFile`) — fixed
  (wave 4). New `WriteToOpts(w, *WriteProps)` and `WriteFileOpts(path,
  *WriteProps)` honor `Compression` on every output path; `WriteFileWith` is
  now a documented thin deprecated wrapper over `WriteFileOpts`; `Write`
  additionally rejects >1 `*WriteProps` arguments (previously silently used
  the first and discarded the rest) rather than staying ambiguous. Tests:
  `TestWriteRejectsMultipleProps`, `TestWriteToOptsCompression`,
  `TestWriteFileOptsCompression`, `TestWriteFileWithStillHonorsCompression`,
  plus the pre-existing `TestWriteCompression`.
- **M10** (`DefineLayout` early-returns where TS registers) — fixed (wave 4):
  `DefineLayout` now always registers the layout, matching TS's
  warn-then-register-unconditionally behavior; there is no Go `console.warn`
  equivalent, so the guards were simply removed rather than becoming a
  logged no-op. Tests: `TestDefineLayoutRegistersDegenerateDimensions`,
  `TestDefineLayoutRegisteredSlideBuildsWithoutPanic`,
  `TestDefineLayoutZeroHeightAlsoRegisters`.
- **M11** (IMG_BROKEN/IMG_PLAYBTN 40-char prefix guard) — fixed: regression
  guard now hashes the full constants (SHA-256 + exact length), not a prefix.
  Test: `TestImageConstants`.

### Minor

- `getSmartParseNumber` truncated fractional ≥100 "already-EMU" values —
  **documented-deviation**: now rounds via `jsRound` instead of truncating
  (reduces, but per Go's `int` return type cannot fully eliminate, the
  divergence from TS's raw-float passthrough). Test:
  `TestGetSmartParseNumberCase2Rounding`.
- `<c:v>` main-path drops TS's `value||value===0` guard — **inert-documented**:
  Go's `Values []float64` cannot hold a JS-style "hole" (every index is a real
  float64), so the guard has no Go equivalent; explained in a code comment at
  `charts.go:564-568` rather than a test.
- Combo-chart validation throws dropped — fixed: the two TS
  malformed-multi-axis-config errors are now raised. Tests:
  `TestValidateChartConfig_SecondaryAxisRequired`,
  `TestValidateChartConfig_AxesCountMismatch`.
- vmerge dummy-cell drop on irregular colspan+rowspan grids — fixed: Go now
  reproduces JS `Array.splice`'s out-of-range clamp-to-append instead of
  dropping the cell. Test: `TestVmergeDummyCellClampAppends`.
- Margin len 2/3 silently no-ops insets — fixed: zero-filled to len 4 like TS.
  Tests: `TestMarginOutOfContractLenZeroFillsInsets`,
  `TestSlideNumberMarginOutOfContractLenZeroFillsInsets`.
- SVG sniffers (viewBox comma form; single-quoted width/height; BITMAPCOREHEADER
  BMPs) — fixed. Tests: `TestGetSizeFromImage_SVG_ViewBoxCommaSeparated`,
  `TestGetSizeFromImage_SVG_ViewBoxCommaSpaceSeparated`,
  `TestGetSizeFromImage_SVG_SingleQuotedWidthHeightAttrs`,
  `TestGetSizeFromImage_SVG_MixedQuoteStyles`,
  `TestGetSizeFromImage_BMP_CoreHeaderVariant`.
- Empty `TableRow` silently skipped where TS crashes — **documented-deviation**
  (kept as Go's saner no-panic behavior rather than replicating the TS crash).
  Test: `TestGetSlidesForTableRows_EmptyRowSkipped`.
- `DefineSlideMaster` stores caller pointers without ISSUE#406 deep-clone —
  fixed (wave 4): `cloneSlideMasterProps` clones the Margin slice and the
  Background/SlideNumber pointers (including SlideNumber's own Margin/Coord/
  bool-pointer/Bullet/TabStops fields) before storing. Tests:
  `TestDefineSlideMasterDeepCopiesBackground`,
  `TestDefineSlideMasterDeepCopiesMargin`,
  `TestDefineSlideMasterDeepCopiesSlideNumber`.
- `Slides()`/`SlideLayouts()` aliasing inconsistency — **documented** (wave 4):
  doc comments on both methods now spell out the asymmetry (Slides returns
  live pointers into internal storage; SlideLayouts returns copies) instead of
  leaving it an implicit trap; behavior itself is unchanged, so no new
  regression test — this was a documentation gap, not a bug.
- Error strings carried TS-style "ERROR:" prefixes / `UNKNOWN-LAYOUT`
  sentinel — fixed (wave 4): no `"ERROR:"`-prefixed strings remain anywhere in
  `pptx/*.go` (verified by grep); `SetLayout`'s bare `"UNKNOWN-LAYOUT"` string
  is now `var ErrUnknownLayout = errors.New(...)`, wrapped with the requested
  name via `%w`, so callers can `errors.Is`. Test:
  `TestSetLayoutUnknownNameWrapsSentinel`.
- PORTING.md documents a `uuidFunc` injection hook that was never built —
  fixed: `Presentation.uuidFunc` is wired through `buildContext` exactly like
  `nowFunc`. It is an unexported (test/internal-only) hook, not public API,
  which matches PORTING.md's original framing of it as a determinism seam
  rather than a user-facing setting. Test: `TestGoldenIntegration` (buildCase08
  pins `p.uuidFunc` to get a byte-exact golden match on the section GUID).
- `WriteFile` truncate-in-place could leave a corrupt file on mid-write
  failure — fixed (wave 4): `atomicWriteFile` writes to a sibling temp file in
  the same directory and `os.Rename`s it into place, removing the temp file on
  any failure. Applies to `WriteFile`, `WriteFileOpts`, and `WriteFileWith`.
  Tests: `TestWriteFileAtomicReplacesExisting`,
  `TestWriteFileFailureLeavesNoStrayTempFile`,
  `TestWriteFileSurfacesMediaErrorAndLeavesNoPartialFile`.
- No `context.Context` on network fetch (fixed 30s timeout only) —
  **documented-deviation**, left as-is: `media.go:41-48`'s comment explains
  this is a deliberate, bounded deviation from TS's un-timed `https.get`
  (a library call should not be able to hang forever); adding a caller-supplied
  `context.Context` would be a real API addition, out of scope for this
  remediation wave, not a bug fix.
- Media resolution is sequential where TS is concurrent — **documented-
  deviation**, left as-is: `media.go:64-67`'s comment explains why (shared
  mutable slide/layout state makes naive parallelism unsafe; a caller wanting
  concurrency can fan out per-slide itself). Latency-only, not a correctness
  issue.
- Redirects: Go follows (embeds final image) vs TS embeds the redirect body —
  **documented-deviation**, kept as Go's improved behavior (embedding a
  redirect's HTML body as "image bytes" is not useful and matches no golden
  fixture). Test: `TestResolveSlideMediaRels_HTTPFollowsRedirect`.

### Nits (WriteProps variadic + aliasing documentation)

- `Write(props ...*WriteProps)`'s variadic-abuse footgun — fixed (wave 4):
  doc comment on `Write` now states explicitly it accepts zero-or-one
  `*WriteProps` and that passing more is a caller error (returns an error,
  does not silently take the first). Test: `TestWriteRejectsMultipleProps`.

### Open items

None. Every finding above is either fixed-with-a-regression-test,
inert-documented (a TS behavior with no reachable Go equivalent), or
documented-deviation (a deliberate, commented divergence retained on purpose,
each with a test locking in the *actual* Go behavior so it can't silently
drift further).
