# Dense layout capability review

Read-only review of the current `compose` canvas and the UHG / EnableComp-style
layout target. This distinguishes authoring effort from features that are
actually missing from the current renderer.

## Finding

The canvas can render a dense page made from editable text boxes, filled
rectangles, rules, and pinned images. A nested evidence table or five-phase
activities / outputs matrix can be authored today as explicitly positioned
`canvas[]` primitives. Dense content by itself is not blocked by a missing
PowerPoint primitive.

The current sample does not yet demonstrate that ability. Showcase slide 2 uses
ten canvas objects for a two-column decision page; slide 9 uses 22 canvas objects
for three evidence gates. A five-phase table, nested evidence rows, or a complex
team diagram would require substantially more authored JSON and native fit
measurements. That is first an insufficient example/spec, not proof the renderer
cannot draw the page.

## Capability versus gap

| Requirement | Current path | Assessment |
|---|---|---|
| Nested evidence table | Use one `surface` or `line` plus separately positioned `text` per cell. Each text cell gets its own native probe. | Feasible today; no table model, row/column tracks, cell padding, border styles, or repeated-row generator. Geometry and every relationship are hand-authored. |
| Five phases with activities and outputs | Draw phase headers, activity/output lane backgrounds, cells, and dividers as canvas primitives. | Feasible today as a manual grid. `PhaseSpec` is not the needed layout primitive: it only groups existing role/pod roots, paints one phase surface, and places one bottom label. It cannot own canvas cells. |
| Complex team composition | Existing team API provides role tiles, variable-role pods, a legend, phases around role/pod roots, and routed reporting connections. | Useful for a conventional org/team tree. It cannot route to canvas/table elements or nested groups, connect pod-internal roles as root endpoints, or give different roles independent sub-fields. Team roles in team mode also require the three staffing tokens. Canvas can draw other structures manually but loses team semantics and routing. |
| Layered dense page | `CanvasSpec` supports text, surface, line, and image. Surface and text objects can be overlapped when one side names the relationship. | Basic layering works, but z-order is fixed by category. All canvas surfaces render first, accent artwork next, then all non-surface canvas items; cards and phase surfaces follow. A permitted overlap does not select draw order. A canvas label declared over a `PhaseSpec` surface will be covered because the phase is emitted later. |
| Dense formatting | A text object has one Arial font size, bold flag, foreground, alignment and plain string. Newlines are allowed. | No rich text runs, real list levels, per-paragraph styles, italics, table cell borders, dashed rules, transparency, or outline-only surface. Several adjacent text shapes/rules can imitate these styles, with extra authored objects. |

The relevant implementations are [canvas.go](../internal/compose/canvas.go),
[team.go](../internal/compose/team.go), [routes.go](../internal/compose/routes.go),
[canvas_render.go](../cmd/pptxcompose/canvas_render.go), and the slide assembly
in [main.go](../cmd/pptxcompose/main.go). `canvas.go` restricts the primitive
types to text/surface/line/image and validates overlap declarations. The renderer
preserves images by checking their content hash and, for ordinary canvas images,
rejects a frame whose aspect ratio differs from the source; the canvas schema
does not expose a contain/crop/stretch choice. The accent path is separately
calibrated and intentionally stretches artwork around a measured single-line
target.

## Concrete constraints for a reusable dense authoring path

These are the gaps that justify new structure after one fully authored dense
example proves the content and visual pattern:

1. **Repeated grids have no parent model.** Every cell, row band, gutter, and
   divider repeats absolute coordinates. Add an explicit table/grid layout with
   row and column tracks, reusable cell style, optional header/footer rows, and
   measured cell text. Keep final planned cells as independent editable shapes.
2. **Panels do not own children.** A background surface cannot define an inner
   content zone or contain a set of IDs. An intentional parent surface must list
   each overlapping canvas peer; children still carry absolute slide bounds.
   Add parent bounds/padding plus relative children and make planned coordinates
   absolute only after composition. Then validate containment and sibling
   collisions once at the group boundary.
3. **Phase/activities/output structure is absent.** Add a five-phase swimlane
   component with explicit lanes (at minimum activity and output), per-phase
   entries, measured labels, and a layout rule for headers and dividers. Do not
   make the current team `PhaseSpec` silently stand in for this distinct visual.
4. **Team graph endpoints stop at roots.** `ConnectionSpec` accepts reporting
   relationships only and the planner exposes standalone roles and complete pods
   as endpoints. Add explicit group/component anchors if the intended dense team
   page needs nested role-to-role reporting, advisory links, or callouts to
   evidence rows. Keep relationship meaning explicit rather than routing arbitrary
   canvas lines as if they were org reporting.
5. **Drawing order is categorical.** `finalSlides` emits canvas surfaces, then
   accents, then the rest of canvas, followed by cards, phases, connectors, title,
   pods and roles. A z-order index or ordered layer/group model is needed when
   intended overlaps cannot fit that sequence. Until then, author canvases that
   respect the current back-to-front groups and avoid overlaps across them.
6. **Cell typography has one style per object.** For emphasis inside a dense
   sentence or table value, use multiple text objects only where the visual
   separation is real; add paragraph/rich-text support only if the selected
   source page requires inline emphasis, bullets, or typographic hierarchy that
   cannot be preserved as separate fields.

The next useful proof is one manually authored, source-informed dense page, not a
generic auto-layout promise. Record the source IDs, visible content hierarchy,
and exact cell/phase relationships. Build it with current primitives if the
visual layout is representable; let native measurement and rendered inspection
identify which of the structured extensions above are actually necessary.

## Insets, padding, and text-bound offsets

There are two separate kinds of spacing in the current API:

- A `surface`'s `bounds` are only its painted rectangle. It does not create a
  child coordinate system or inherit padding.
- A `text` object's `bounds` are its outer PowerPoint text frame. `inset_x` and
  `inset_y` become the four native text-frame margins. The safe text area is
  therefore `bounds.width - 2*inset_x` by `bounds.height - 2*inset_y`.

`canvasProbes` measures exactly that safe width in an isolated probe frame with
zero margins. `render.go` then puts the same measured width inside the full text
frame by applying the margins. This is not a double subtraction in the planner;
the common authoring error is to move/size a text frame as though it were already
the padded content box, then also apply the same inset. For example, positioning
text 16 pt inside a surface and setting `inset_x: 16` leaves 32 pt from the
surface edge to the text's safe area. Choose one convention: a coextensive text
frame with inset margins, or an already inset text frame with zero margins.

The measured glyph rectangle is not the text-frame rectangle. Native evidence
records non-whitespace character bounds; `checkNative` stores their dimensions
and offsets from the frame plus margins. `planAccents` uses those offsets after
adding the current `bounds.x/y` and `inset_x/y`, so an accent can follow actual
glyph placement. Do not use glyph bounds as a substitute for a panel/surface
frame or add the inset twice to an accent target. Canvas collision checks use
the full outer frame (not just the inset safe area), which is conservative for
text-to-text spacing and intentional for panel boundaries.

Nested panels are currently absolute sibling objects. Their fills have to be
listed as intentional overlap peers, and the text frame must be positioned
independently. A relative child/padding model would remove the most common source
of coordinate errors while also making dense grids easier to maintain.

## Follow-up from the dense experiment

The five-slide dense draft uses 258 independently measured text objects. The
calendar correction changed only a small portion of those contracts, but the
current CLI requires a complete new probe set whenever any request changes.
A future cache should reuse individual complete text contracts while retaining
origin bundle/evidence hashes and a renderer/font-environment fingerprint. It
must validate every current binding against its original observation, measure
changed or missing contracts, and still perform final native verification and
visual review. Do not merge editable measurement summaries or silently relax the
current full-contract equality check to save time.
