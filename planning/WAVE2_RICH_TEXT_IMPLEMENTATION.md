# Wave 2 — bounded rich text implementation

Status: the native rich-text smoke probe and final smoke-deck verification passed
the v6 character and paragraph checks. That result is historical evidence for
runs and point spacing. The current contract is v7 because it also binds native
paragraph line spacing. Full fixture qualification is tracked in
[the Wave 2 checkpoint](WAVE2_CHECKPOINT.md). Keep one frozen CLI binary per
qualification run; the measurement cache intentionally rejects changed binaries.

## Supported schema

Rich text is available on a text `CanvasSpec` and on a layout `BlockSpec` through
explicit paragraphs and runs:

```json
{
  "id": "summary",
  "kind": "text",
  "bounds": {"x": 36, "y": 160, "width": 480, "height": 120},
  "paragraphs": [
    {
      "id": "lead",
      "align": "left",
      "space_after_pt": 6,
      "runs": [
        {
          "id": "label",
          "text": "Measured ",
          "font_face": "Arial",
          "font_size_pt": 14,
          "foreground": "#070154"
        },
        {
          "id": "emphasis",
          "text": "bold and italic",
          "font_face": "Arial",
          "font_size_pt": 14,
          "bold": true,
          "italic": true,
          "foreground": "#070154"
        }
      ]
    },
    {
      "id": "detail",
      "align": "left",
      "space_before_pt": 2,
      "line_spacing_multiple": 0.9,
      "runs": [
        {
          "id": "link-style",
          "text": "Underlined color",
          "font_face": "Arial",
          "font_size_pt": 12,
          "underline": true,
          "foreground": "#2355FF"
        }
      ]
    }
  ],
  "align": "left",
  "valign": "top",
  "inset_x": 0,
  "inset_y": 0
}
```

Paragraph and run IDs are stable within their containing text element. Paragraph
alignment supports `left`, `center`, and `right`. Runs support Arial, point font
size, bold, italic, single underline, and concrete RGB colors. Paragraph spacing
before and after is expressed in points. `line_spacing_multiple` is a unitless
PowerPoint within-line multiplier from 0.5 through 4; omission means 1. The
legacy uniform `text` schema remains supported; one element cannot mix legacy
text fields with `paragraphs`.

## Native and cache contract

Probe generation carries the complete paragraph and run structure. The cache key
binds the text, paragraph IDs and spacing, run IDs and all run styles, frame width,
insets, alignment, foreground/background context, and the legacy uniform fields.
A change to any bound value creates a different text contract.

The environment fingerprint binds the executable, adapter and inspector hashes;
the PowerPoint version/build; the operating system; and the file hashes for Arial
regular, bold, italic, and bold italic. A missing face or a substituted family
fails environment inspection.

Cache import requires `pptxgengo.compose-text-measurement.v7`. The v7 native
adapter records each character's text, font name, point size, bold, italic,
underline enum, and RGB value, plus each paragraph's alignment and point spacing.
It also records PowerPoint's within-line rule and unitless spacing value. The
checker requires the within-line rule and compares its multiplier exactly apart
from floating-point representation noise. It compares all records to the spec
and reconstructs measured bounds from native observations. There is no
uniform-style fallback for rich text.

## Commands

Build the CLI once before qualification and use the same binary for every step:

```sh
go build -o /tmp/pptxcompose-wave2 ./cmd/pptxcompose

/tmp/pptxcompose-wave2 probe \
  --spec samples/visual-wave2/smoke-spec.json \
  --cache /tmp/pptxcompose-wave2-cache \
  --out /tmp/wave2-smoke-probe

/tmp/pptxcompose-wave2 measure \
  --bundle /tmp/wave2-smoke-probe \
  --cache /tmp/pptxcompose-wave2-cache \
  --out /tmp/wave2-smoke-evidence.json

/tmp/pptxcompose-wave2 build \
  --spec samples/visual-wave2/smoke-spec.json \
  --cache /tmp/pptxcompose-wave2-cache \
  --out /tmp/wave2-smoke-final

/tmp/pptxcompose-wave2 verify \
  --bundle /tmp/wave2-smoke-final \
  --out /tmp/wave2-smoke-final-evidence.json
```

The native `measure` and `verify` commands open PowerPoint through the configured
AppleScript adapter. The other commands do not provide substitute native
measurements when a cache entry is absent.

Static checks used before freezing the binary:

```sh
go test ./...
osacompile -o /tmp/measure-compose-text.scpt scripts/measure-compose-text.applescript
swiftc -typecheck scripts/compose-environment.swift
git diff --check
```

## Qualification findings

- The historical v6 smoke probe passed exact rich character-style and paragraph-style
  checks for two widths, two paragraphs, Arial bold italic, single underline,
  RGB color, and 2pt/6pt paragraph spacing.
- Source-deck QA found 0.9 within-line spacing on compact captions. This property
  materially changes native measured height, so v7 binds the rule and multiplier;
  older rich evidence cannot enter the current cache.
- The generated OOXML contains distinct native paragraphs and editable runs.
- Layout blocks retain stable measurement IDs while lowering rich paragraphs to
  canvas text.
- Unit coverage confirms cache invalidation on run-style changes, native mismatch
  rejection, stable layout lowering, source non-aliasing, and invalid-schema
  rejection.
- `go test ./...`, AppleScript compilation, Swift type checking, and
  `git diff --check` passed before the smoke run.

## Explicit limits

- Rich runs support Arial only. Other font families fail validation.
- Rich native bullets and hanging indentation are deferred. Layouts continue to
  support the existing measured marker/text block pair.
- Runs cannot contain carriage returns or line feeds; use another paragraph.
- Non-BMP characters are rejected because the native character index contract is
  bounded to matching Go and PowerPoint character positions.
- Rich-text content recovery is unsupported and fails explicitly instead of
  dropping run styles. Regenerate from the source spec or use `pptxscene` for a
  broader edit workflow.
- The supported underline style is single underline. Per-run hyperlinks, baseline
  shifts, letter spacing, and additional paragraph properties are outside this
  contract.
- Native final verification and visual review remain required for deliverables.

## Image placement review

The adjacent Wave 2 image implementation was reviewed without code changes.
Within its declared PNG/JPEG scope, the geometry is internally consistent:

- `preserve` rejects a frame whose aspect ratio differs from the source by more
  than 0.5%; `stretch` is explicit.
- `contain` emits symmetric negative source edges, preserving the source ratio
  while adding transparent extent inside the exact requested frame.
- `cover` computes a ratio-preserving visible source window and clamps the focal
  point so the window stays inside the source.
- `source_crop` validates the visible source area and rejects crops whose visible
  aspect ratio would distort in the requested frame.
- Rendering checks the pinned asset hash and keeps the PowerPoint picture frame
  at the requested point bounds. Package tests confirm the expected `srcRect`
  and frame XML for a cover crop.

No must-fix arithmetic issue was found in these paths. Actual PowerPoint evidence
for contain, cover, and source-crop rendering is still required before claiming
native crop qualification. The renderer currently registers PNG and JPEG
decoders; other image formats are outside the reviewed scope.
