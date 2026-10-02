# Pure Go font layout prototype

`pptxcompose measure`, `fit-report`, and `build` accept `--engine go`. This
opt-in path measures text and generates editable PowerPoint objects entirely in
Go. It needs font files, but no PowerPoint, AppleScript, Swift, or font service.
The default engine remains `native`.

The engine is general: the authored `font_face` selects the family. IBM Plex
Sans is an example, with no family-specific measurement code.

## Run it

From the repository root, with IBM Plex Sans, IBM Plex Mono and Arial available:

```sh
go build -o /tmp/pptxcompose-go ./cmd/pptxcompose
/tmp/pptxcompose-go measure --engine go \
  --spec library/dynamic-components/go-font-layout.json \
  --out /tmp/go-measurements.json
/tmp/pptxcompose-go fit-report --engine go \
  --spec library/dynamic-components/go-font-layout.json \
  --out /tmp/go-fit.json
/tmp/pptxcompose-go build --engine go \
  --spec library/dynamic-components/go-font-layout.json \
  --out /tmp/go-font-deck
```

Every output must be new. The build directory contains its PPTX,
`manifest.json`, and `go-layout.json`. Build measures the spec directly;
there is no intermediate probe deck or evidence file to supply.

To use a controlled font set instead of system discovery, add repeatable
`--font-dir` flags to each command:

```sh
/tmp/pptxcompose-go build --engine go \
  --font-dir /path/to/brand-fonts --font-dir /path/to/supporting-fonts \
  --spec your-spec.json --out /tmp/your-go-deck
```

These directories replace the system search paths. The resolver recursively
indexes `.ttf`, `.otf`, and `.ttc` files, including collection faces. It matches
family names without case differences; it does not expand aliases or remove
spaces from the authored family name.
Missing families, styles, or glyphs produce errors. Fonts are referenced in the
PPTX and are not embedded; recipients need the selected fonts for rendering.

Repository changes require rebuilding the binary. A previously installed
frozen `pptxgengo` release does not acquire this prototype automatically.

## How measurement works

1. Resolve each run to an exact font family and a real style. Static regular
   faces use weight 400; bold faces use weight 700 or an explicit font bold
   flag. Variable fonts select `wght=400/700`, `wdth=100`, and `ital=0/1` when
   those axes exist; other axes retain their defaults. A separate italic face
   is also supported. Styles are never synthesized.
2. Shape runs using the pure Go HarfBuzz implementation in
   `github.com/go-text/typesetting`, pinned to v0.3.5. Font kerning and default
   ligatures contribute to advances. Fractional point sizes are preserved by
   shaping at a larger scale and returning metrics in points. Each shaped
   glyph advance is rounded to the nearest 1/8 pt before wrapping, following
   the [wrapping calibration](font-wrap-calibration.md).
3. Wrap the shaped paragraph at Unicode line break opportunities, breaking
   inside an overlong word when necessary. Hard breaks and blank lines remain
   explicit. Trailing line whitespace is trimmed by the wrapper. Mixed font
   runs share a predicted baseline; the largest point size determines the
   character line box.
4. Add paragraph spacing, line spacing multiples and bullet indents. Height
   uses a 1.2 em line allocation, following the
   [fresh native calibration](font-native-calibration.md). Character boxes
   include leading and spacing between paragraphs; glyph baselines retain
   75% of extra leading above. Expanded final allocations are conservative
   estimates and generate review warnings. Blank lines move following text;
   trailing blanks and final paragraph spacing do not enlarge the occupied
   union. Glyph ink stays a separate height diagnostic. Textbox insets stay
   outside the measurement, as in the planner contracts.
5. Pass the measurements into the existing composition planner. The PPTX
   writer emits editable text, fonts, styles and frames. PowerPoint will apply
   its own text layout when it renders those objects.

Reports identify `go-text-prototype.v4` and record the height and advance policies, font paths, collection
indices, SHA-256 hashes, styles, weights and variable axes. Per-request details
include line text, rune ranges, advances, baselines, line heights, layout advances and glyph ink
bounds. Ink bounds exclude underline decoration. `powerpoint_verified` is
always false in a Go layout report, even if a separate native verification is
later performed. Go results never enter the native measurement cache.
Per-request `boundary_warnings` and top-level `warnings` flag lines close to
a width boundary. `fit-report` includes `wrap_boundary_warning_count`.
These warnings are a review heuristic and do not change predicted fit.
`height_warnings` flag cases needing further native character-bound review;
`fit-report` includes `height_review_warning_count`. The fresh 393-case native reference
covers Arial, IBM Plex Sans and IBM Plex Mono, including fractional and mixed
sizes and blank lines. Nondefault line spacing still needs terminal-box review;
other families remain unqualified.

## Scope and fidelity

Supported: horizontal Latin text and covered common symbols, plain text hard
breaks, mixed fonts/sizes, regular/bold/italic/bold italic, underlining in the
editable output, wrapping, left/center/right alignment, paragraph spacing,
line spacing multiples and simple hanging bullets.

Explicitly rejected: tab stops, bidirectional or non-Latin layout, format
controls, glyph fallback and measured phrase bounds/phrase accents. Arbitrary
variable-axis authoring, hyphenation and uncalibrated PowerPoint layout behaviors
are outside this prototype. Font glyph coverage still limits supported symbols
and combining marks.

**Go fit is a prediction of this engine's layout.** PowerPoint can produce
different spacing, bounds and line breaks. In the original v1 ten-request `fonts.json`
control, using the same font files as previously captured native evidence,
all line breaks matched. Maximum absolute width difference was 0.996 pt;
maximum absolute height difference was 4.838 pt. That bounded comparison does
not establish general parity. At 16 pt, the three-line IBM Plex Sans paragraph
measured 62.4375 pt high in Go versus 57.6000 pt in the native evidence; Arial
measured 55.21875 pt versus 57.6000 pt. Height differences can affect capacity
in either direction. The subsequent v3 height rule matches those same ten
native controls within 0.000003 pt. These controls were used for calibration;
they do not establish independent validation or arbitrary-font parity.

Native final verification remains available separately on macOS:

```sh
/tmp/pptxcompose-go verify --bundle /tmp/go-font-deck \
  --out /tmp/go-font-deck-native-verification.json
```

This optional command uses the existing PowerPoint adapter and checks the
actual final deck. It is the path to qualify its native fit. Generation itself
uses no adapter. `--engine go` deliberately does not implement `verify`.

## Implementation and checks

The expanded [three-family reference set](font-layout-reference.md) captures
276 cases for IBM Plex Sans, Arial and IBM Plex Mono. The subsequent
[wrapping calibration](font-wrap-calibration.md) improves matching line breaks
from 252 to 275, with 116/117 matches in an additional corpus and no regressions.
Two narrow-boundary differences remain across the sets and are flagged in the
reports. PDF selection geometry has a separate contract from native TextRange
measurements. The [v3 height calibration](font-height-calibration.md) separately
compares ten native character-bound controls and 441 PDF baseline gaps.
Average baseline-gap error falls from 1.5195 to 0.3535 pt; the largest remaining
gap error is 1.2 pt. The subsequent [fresh native capture](font-native-calibration.md) covers all
393 cases. v4 matches 385 case heights and every visible line top where wrapping
agrees. Six expanded-spacing terminal heights and two wrapping boundaries
remain under review. The [focused 45-control capture](font-native-focused-results.md)
completed, and the exact two-slide v4 example passed native final verification:
all 14 text shapes match native wrapping and fit their safe zones. Go reports
remain predictions; future decks need their own final verification.

- Engine: `internal/textlayout`; CLI integration: `cmd/pptxcompose/go_layout.go`.
- Portable tests use the Go font fixtures bundled with `golang.org/x/image`.
  They cover wrapping, hard breaks, fractional sizes, mixed styles, alignment,
  bullets, spacing, explicit font/style/glyph failures and variable-axis
  selection. CLI tests build with an empty executable search path and enforce
  new-only output and separate provenance.
- An optional macOS integration test exercises installed Arial kerning, Plex
  ligatures and combining marks, variable regular/bold/italic styles, and
  Helvetica Neue collection loading.
- Run source package checks with `go test ./cmd/... ./internal/... ./pptx/...`.
  The ignored `samples/` tree may contain incomplete archived release source
  copies, which are not part of those package checks.
