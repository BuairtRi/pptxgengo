# Measured canvas and artwork

`SlideSpec.canvas[]` extends the role, pod and card compositor with explicit
editable text, surfaces, straight rules and pinned PNG/JPEG images. Coordinates
are points. The [showcase spec](../showcase/deck.json) is an executable example.
The [spec generator](../../scripts/build-showcase-spec.py) emits the declarative
JSON; the Go CLI authors every slide.

## Elements and overlap

Every item has a stable `id`, `kind`, and `bounds`. Text requires explicit Arial
font size/weight, `foreground`, `align` (`left`, `center`, `right`) and `valign`
(`top`, `middle`). Optional `inset_x`/`inset_y` reserve internal space. Native
probe width exactly matches the final inner width. No character-count estimate,
autofit or automatic font shrinking substitutes for measurement.

Surfaces have `background`. Lines have `foreground` and `line_width_pt`, and one
zero frame dimension. They remain native editable objects. The showcase's
comparison table and roadmap are composed from editable text and shapes; they
are not native PowerPoint table/chart objects with spreadsheet semantics.

An intentional overlap names the other component in `allow_overlap`. This covers
background surfaces, arrow artwork surrounding a label and line endpoints touching
a diagram block. Unspecified text/shape overlaps fail planning. Naming an overlap
permits geometry; it does not establish readability, contrast or correct stacking.
Surfaces render behind artwork, then text and semantic components. Native and
visual review remain required.

Images require a local `asset_path`, its exact `asset_sha256`, and `alt_text`.
PNG/JPEG dimensions must match the requested frame aspect ratio within 0.5%.
The renderer embeds verified bytes, preventing a later path read from switching
assets. Image verification checks native placement; the package hash binds the
embedded content. Paths resolve from the working directory, normally repo root.
Brand asset descriptions/provenance are in [the selected asset catalog](../showcase/assets.json).
Asset files are local/ignored and must be hydrated before rebuilding elsewhere.

## Measured accents

`accents[]` addresses a whole canvas text block using `target`. Modes are
`underline` and `highlight`. `alpha_bounds` identifies the source artwork's
normalized visible bounds, with explicit optical padding/offsets and underline
stroke height. The planner uses v4/v5 native character-bound unions and offsets to
position the artwork behind the target text. The output contains the target,
visible artwork bounds and full image frame for audit.

These calibrated accent assets intentionally scale independently in width and
height. Ordinary images preserve aspect ratio. Transparent image padding may
extend beyond a slide, but visible accent ink must remain inside it. Visible ink
cannot collide with unrelated content, semantic components or reporting lines.

Only unrotated, top-aligned, unfilled text blocks are supported. Height above
1.5×font size is conservatively treated as multiline and returns
`manual_required`. This is a bounded single-line guard, not line-fragment analysis.
For a phrase inside a larger title, use `pptxanchor` with the native phrase adapter.
General arrow selection, aesthetic routing and arbitrary multi-line emphasis remain
manual design work. The showcase's hand-drawn arrows have explicit positions.

## Reusing unchanged text measurements

A build normally requires exact spec, probe-manifest and deck hashes. The explicit
`build --reuse-measurements` option allows a geometry/style/asset revision when
**every complete probe request is unchanged**, including text, width, font, weight,
alignment and measured color contract. Source evidence and manifest/deck hashes
must still match. The CLI reconstructs measurements from raw native observations,
not the editable summary. Changed text widths or typography require new probes.

The new manifest records `reused_from_spec_sha256`. Every resulting deck still
requires final native verification and visual review, because changed placement
can create overlaps or a changed context can affect rendering. This option does
not rebind source component contracts or approve a modified source scene.

## Narrative metadata

`role`, `takeaway` and `notes` describe what each slide contributes. The renderer
writes supplied notes; the narrative is authored, not automatically inferred or
fact checked. The showcase identifies its scenario and numerical inputs as
illustrative and makes no client-result claim.

## Recover colleague text edits from a generated deck

```sh
/tmp/pptxcompose recover-text --spec library/showcase/deck.json \
  --bundle samples/showcase/deck --returned samples/colleague-return.pptx \
  --text-only --out samples/showcase/recovered
```

The output is `spec.json` plus a before/after `recovery.json` report. Recovery
requires the original spec and built bundle, stable slide order and object names,
unchanged object counts/frames and unchanged image assignments/bytes. Missing,
added, moved or resized objects stop recovery and direct the agent to `pptxscene`.
Derived legend labels cannot be independently rebound.

`--text-only` explicitly restores original spec styling. It does not import
colleague formatting or notes, nor reconcile general layout changes. The original
PPTX is untouched. Recovered content needs new probes, a new build and final native
verification. For decks without a known original spec, extract their raw native
scenes; do not infer a semantic composition or YAML contract from visual similarity.

The current v5 adapter reads native character/font property snapshots in batches
and excludes spaces/tabs from visible bounds. Whitespace formatting is still
checked. This avoids end-of-line space advances extending past a wrapped text
frame. Bounds remain character advances rather than a raster-ink measurement.

## Diagnose all fixed text zones

`fit-report` uses the same spec, bundle, native evidence and optional explicit
reuse checks as `build`. It reports available and measured dimensions plus
overflow in points for every fixed title/canvas/role/card text zone. It also runs
the full planner and records its result for dynamic pods, phases, legends,
collisions and routes. It does not render a deck or imply visual approval.

```sh
/tmp/pptxcompose fit-report --spec library/showcase/dense-deck.json \
  --bundle samples/component-adaptation/dynamic-pods/dense-probes-v1 \
  --evidence samples/component-adaptation/dynamic-pods/dense-evidence-v1.json \
  --reuse-measurements --out samples/showcase/dense-fit-report.json
```

The command returns a report even when text does not fit; automation must inspect
`overflow_count` and `planner_passed`. Reflow the layout or deliberately revise
content, then measure any changed text/width/font contract. Do not shrink fonts
automatically to hide excess content.
