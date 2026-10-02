# Native source primitives and media

Implementation: `internal/wmdesign/scene_primitives.go`, `scene_media.go`, and
`scene_primitive_reference.go`. Contracts are
`pptxgengo.wmds-source-primitives.v1` and `pptxgengo.wmds-source-media.v1`.

The source handlers support text, textblock, bullets, numbered lists (`ol`),
index lists, schedules, group labels, numbered and column headings, strong
number lists, pullquotes, stat/photo squares, image frames, logos, art, marks,
and deliverable thumbnails. The strict per-kind field lists reject fields
belonging to another primitive as well as unknown fields. Nested row and bullet
objects are strictly decoded. Empty required copy, invalid enum values, invalid
geometry, unresolved note references, and measured overflow return named errors.

Text remains native editable text. Square bullets, decorative rules, thumbnail
pages, thumbnail diagrams, and thumbnail chart/table placeholders remain native
editable shapes. Combined bullet leads use actual semibold font faces and the
secondary text channel. The source small list size is an explicit authored choice.
No handler truncates copy or switches density to rescue a fit failure.

Index rows honor their fixed row height and center text vertically. Schedule rows
honor fixed key columns and align measured baselines. Numbered headings align
number and heading baselines. Text height and middle/bottom anchoring use the
measured allocation and occupied envelope. Pullquote marks use the source
60-point face and 30-point leading, with conservative occupied bounds.

Inline `[[...]]` marks are parsed into explicit rune ranges and placed across
measured Go line fragments. Highlight, underscore, circle, and spark artwork use
pinned canonical brand bytes; the highlight variant may be 1–4 or automatic.
Highlight on a disallowed dark or callout surface fails. Footnote `[^N]` markup
requires an existing nonempty note at N. It emits a native run at 60% base size,
weight 600, emphasis ink, and an explicit 35%-of-base-size upward baseline shift.
Native mixed-face and superscript parity remains independently unqualified.
Measured native spacing runs preserve the source highlight/circle net padding,
spark margin, and footnote margin. Circle/spark spans use nonbreaking spaces and
reject a line split, preserving source inline-block semantics.

Array identity is obtained from `SceneContext.Keys` at the exact array path.
Frozen reference fixtures without an overlay retain explicit `source-NNN`
ordinals. Content adapters must supply keys for every caller-owned array; they
must not rely on those fixture ordinals for mutable content.

## Canonical media registry

[primitive-media-assets.json](primitive-media-assets.json) records the exact SHA256
for 696 local payloads: photographs, logos, tagline, hand-drawn marks, four
highlights, and every navy/magenta/white icon variant. Runtime resolves these
same entries through the compiled registry. `WMDS_BRANDING_ROOT` may relocate the
canonical folder; every payload must still match its pinned SHA256. Headshots
are the frozen owner-supplied `../career/profile_image_1.png` sibling asset, with
its own pinned bytes and explicit source face crop. Arbitrary paths and URLs are
not accepted as registered asset IDs.

Photo focus is a pair of percentages in [0,100]. Cover cropping is explicit
native picture crop geometry, preserving the source data. Declared grayscale
and the frozen headshot face crop are deterministic Go pixel transformations.
Original SHA256 provenance remains in the picture description. SVG pictures
retain native SVG data and an explicit PNG fallback generated in Go. The
fallback rasterizer handles the closed official filled-path vocabulary, including
transformed rectangles/circles, cubic/quadratic curves, and elliptical arcs;
unsupported visible SVG elements or stroke-only shapes fail explicitly.

Marks extract the official artwork group, omit its hidden editing strokes,
apply the resolved source ink, and use measured artwork bounds with the frozen
0.4-point expansion. The original canonical payload hash is retained in the
picture description alongside the resolved ink. SVG and fallback are separate
picture parts, never an image of a whole slide.

Logo/art height preserves intrinsic aspect. Photo squares and stat squares use
separate content unions. Placeholder thumbnails intentionally represent source
thumbnail artwork; supplied thumbnail photos instead resolve registered media.
Caption text remains separately editable.

## Integration helpers

- `primitiveMediaImage(id, asset string, box Rect, focus string, grayscale bool)`
  returns a native `*pptx.ImageProps` with pinned asset provenance.
- `planIconScene(id, name string, size float64, box Rect, surface, ink string)`
  returns an ordered icon scene plan. Source aliases are preserved; unsupported
  brand icon inks use the source's luminance-based canonical navy/white selection.
- `sceneIcon(plan, id, name string, box Rect, surface, ink string)` appends the
  icon child plan with ownership.
- `primitiveArtworkImage(id, asset string, box Rect, surface, ink string)`
  returns recolored canonical hand-drawn artwork and its fallback.
- `primitiveRichText(plan, id, text string, style Style, box Rect, surface, ink,
  align, emphasis, markInk string, context SceneContext)` emits measured editable
  rich text and its separately owned marks.

`PrimitiveSceneReference(year)` supplies a 15-slide synthetic generation document
covering every owned node type and all four inline mark modes.

These executable modules do not establish native visual qualification. A
reference specimen and changed-content evidence must be recorded per adapter.
