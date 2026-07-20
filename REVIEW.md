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
