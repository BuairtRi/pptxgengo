# WMDS Go typography v2 candidate

This candidate is opt-in: `--engine wmds-go-foundation.v2`. The default remains
`wmds-go-foundation.v1`. Normal generation loads font bytes and computes layout
in Go; it does not call PowerPoint, AppleScript, Swift, or an external service.

## Calculation contract

- Uniform horizontal Latin text, exact leading, zero margins and paragraph
  spacing, no autofit. Unsupported scripts, tabs, missing glyphs and mixed-run
  authoring remain explicit errors.
- HarfBuzz shaping uses `kern=0`, `liga=1`, `clig=1`, including when tracking is
  nonzero. Captured Plex text uses the `fi` cluster at nonzero tracking.
- Glyph advances round to 1/8 pt before serialized 0.01 pt cluster tracking is
  applied. Terminal tracking participates in fitting. Tracking is retained in
  64× layout units to avoid accumulating a 1/64 pt rounding error per character.
- First baseline and terminal character height use 17 pinned estimates keyed by
  source font SHA, point size and leading. They were fitted from the preserved
  v1 measurements, and require independent v2 validation. No family-wide height
  formula or arbitrary-size qualification is claimed.
- Blank paragraphs reserve exact leading. Visible occupancy is separate from
  allocated height and records a top offset for leading blank paragraphs.
- Unknown font/size/leading combinations use a provisional baseline and
  conservative terminal allocation, named explicitly in reports. The reference
  deck's derived Mono Semibold 7 pt/9 pt page-number style is such a combination.
- Each authored hard break, including a trailing break, becomes an explicit
  DrawingML paragraph with complete font, size, tracking, color, bold and italic
  end defaults. This pass is scoped to WMDS v2; the shared writer is unchanged.

## Fonts and evidence

PPTX typefaces and PDF PostScript names retain IBM's normal names. No renamed
aliases are created and no installed font files are changed. The Go model uses
pinned static source files. Independent CoreText inspections inventory installed
files, variable axes/instances, family selection and PostScript selection. These
fingerprints are not an observation of which font file PowerPoint accessed.

[Development replay](../../../samples/wmds-typography-v2-20261001/final/development-replay.json)
now predicts all 75 historical native line breaks and all 74 comparable heights
within 0.5 pt. The old trailing-break failure is retained and excluded from those
height comparisons. This is a fit to development evidence, not a fresh native pass.

The [new controls](../../../samples/wmds-typography-v2-20261001/README.md) retain
those 75 original geometries and add 64 controls: tighter boundaries, heldout
accent/descender strings, exact-height frames, leading/trailing/multiple blanks,
punctuation, trailing spaces and long tracking. The new native PDF matches all
139 line predictions; first-baseline/pitch errors are no greater than 0.12 pt.
All 47 control pages and the 18-slide foundation candidate have been reviewed.

[Independent native character capture](../../../samples/wmds-typography-v2-20261001/NATIVE_RESULTS.md)
now passes all 139 controls on the captured environment. Content, native wrapping,
observed character/paragraph styles, frames, occupied height/top and fit agree.
Maximum occupied-height error is 0.000041 pt; native line-advance error is 0.132 pt.
The collection omits four final zero-length paragraphs; exact native text plus
saved paragraph structure/defaults verify their semantics. No raw data is changed.

The fitted calibration JSON is retained as its original development record;
the independent native review is separate validation evidence. This validates
the bounded matrix, not arbitrary strings/widths/sizes. Exact native font
file/instance access remains unobserved and `native_qualified` remains false.
PDF selection widths are diagnostic and are not native character bounds.

Native capture is development/qualification evidence, not a runtime dependency.
After validating a bounded font/style/content envelope, normal generation should
use Go alone. New fonts/features and material PowerPoint changes may require new
reference capture; every authored deck does not require character capture.
