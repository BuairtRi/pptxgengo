# WMDS library source revision v2

Frozen round-4 source from `wm-design-system` commit
`7bdcaee5030a12275a1f881a8542f4d302d207df`.

The standalone developer CLI selects it explicitly with
`--bundle library/wm-design-system/v2`; its default remains frozen v1. The packaged
`pptxgengo design` route defaults to this v2 library. Source revision is independent of the typography engine;
use `--engine wmds-go-foundation.v2` for scene/template rendering. Fonts and logo
assets are copied byte-for-byte from v1 and use their original IBM/WM names.

The source contains 167 designs (166 active, one retained deprecated entry).
Catalog presence does not imply generation or native qualification. Slice 1 adds four split frames, three-line split titles, configurable 2–6-tab source
navigation, lifecycle metadata and revision-specific refinement handling.
Slice 2 adds checkbox cells, square quadrant plots/numbered markers, named-target
annotations and revised metric bindings. The six initial capability blockers are
removed. Slice 3 implements all 70 additions with source and meaningful-content
pairs in a 140-slide reference. Slice 4 covers every design with source and
meaningful alternate specimens: all 167 pairs / 334 slides generate and were
reviewed through a native PowerPoint PDF. The final library deck has one source
example per design. The local.6 release packages fonts, calibration, artwork and
editable specimens behind a separate searchable gallery.

See [integrated completion](../../../planning/wm-design-contracts/v2/slice4-completion.md)
and [review receipt](../../../planning/wm-design-contracts/v2/reference-review/slice4.json).
Acceptance applies to those specimens; arbitrary copy, automatic count changes
and exact native font-file selection remain unqualified. New copy requires fit
and visual review. Normal layout/generation runs in Go without native capture.

The frame reference and five focused template bindings have separate native
review evidence under `planning/wm-design-contracts/v2/`. A separate 46-slide
reference covers all 15 revised designs and new component semantics. The [added-family review](../../../planning/wm-design-contracts/v2/slice3-completion.md)
records all 70 added pairs and exact amendments. The previous 97-design review
does not automatically qualify new or revised layouts.

`scripts/freeze-wmds-v2.py` documents the byte-for-byte freeze and refuses to
overwrite an existing bundle. Any future source change requires a new pin and an
explicit migration. Generation never runs source Python/JavaScript code.
