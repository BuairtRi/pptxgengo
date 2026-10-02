# Coverage: openers and argument

## Scope and evidence

This read-only planning pass covers **36 canonical variants**: 18 openers and
18 argument slides. Three already have closed content adapters: `cards/3`,
`cards/4` and `takeaway-rail/metrics-rail`. The other 33 remain inventory only.
`implemented_bound` describes adapter coverage and saved specimen review; it
does not qualify arbitrary copy or every effect in the original source examples.

[coverage-primitives.json](coverage-primitives.json) records all 36 variant keys,
tiers, recursive rendering nodes, feature locations, source hashes, binding work,
dependencies and source resolutions. It excludes table column `type: icon` from
body-node counts and records it as a typed cell feature instead.

| Frozen source | SHA-256 |
|---|---|
| `templates/library/openers.json` | `54c11bef39b933ead57bf994d1c2da4c06c453d6a9e43761b253557d639ef4b6` |
| `templates/library/argument.json` | `21b54f50ce5d8d13bda451aa94923d5e84cc9532f77ac3e13739cfcc31f5eba2` |
| `explorations/components.src.html` | `017e308d7ad107ef8faca4996f52af2496554404c76f7ab2935fdea6d8f039ca` |

The corresponding live source files in `~/Projects/wm-design-system` were byte
identical during this pass. The renderer and source are authoritative semantic
inputs, not evidence that browser overflow or every fixed-height example is safe
in PowerPoint. No Go implementation, source edits, tests, builds, deck generation
or native capture were performed in this pass.

## Work packages

The dependency IDs below appear in the JSON ledger. They map to the root
[library completion plan](library-completion.md), rather than defining a competing
renderer or binding API.

| ID | Complete scope | Principal source examples |
|---|---|---|
| P01 | Frame/art zones: inferred `noHeader`, authored whiteboard fields/fades/above order, local painted surfaces, bleed geometry, vertical text alignment | Covers, agenda/index, all dividers, key-message statements |
| P02 | Pinned media/logo/icon/tagline registry; intrinsic aspect; cover crop/focus; declared grayscale; asset identity and editable text separation | Cover photos, quote photos, closing/tagline, pillar/hub icons |
| P03 | Inline marks and standalone mark artwork; Go glyph-fragment measurement; rich bullet leads; explicit footnote IR when enabled | Header/body highlight and underscore, from-to arrow, narrative leads |
| P04 | Full source card anatomy and compact contextual variants; body labels, icons, real body-weight titles, local padding/gaps, multiple paragraphs | Method, pillars, objective, goal-steps, hub spokes |
| P05 | Measured keyed lists, schedule/index rows, strong numbers, stat squares, pullquotes and callouts | Openers, objective/numbered-and-panel, comparison |
| P06 | Native blocks, broad/vertical rules, framed tabs, chevrons/text arrows and source point connectors | Dividers, pain-to-theme, framework table, hub/flow |
| P07 | Native compound tables/matrices and group labels; icon cells, fixed row capacity, source microgeometry | Method, capability-table, capability-map, frameworks |
| P08 | Before/after pairs, vertical stepper and keyed argument compositions | From-to rows, transformation, session, hub spokes |
| P09 | Closed family binding registry, strict recursive source decoding/identities and semantic counts | All 33 remaining variants; preserve existing three APIs |
| P10 | Recorded decisions for genuine source conflicts, immutable-source corrections and visual overlap | Capability map duplication/tiny cells; metadata slot mismatches |

Implement frame/asset/primitive foundations in parallel with full cards and
tables. Sequence before/after, stepper and argument adapters after their native
primitive dependencies. Bind template batches only when every required feature
has an executable contract; a familiar top-level `card` name is insufficient.

## Existing code worth reusing

| Capability | Existing implementation | Reuse boundary |
|---|---|---|
| Typography, real font identities, tracking, exact leading | `internal/wmdesign/typography.go`, `candidate.go`, `rich_text.go` | Keep v2 measurement and pinned calibration; unknown custom size/weight combinations remain conservative and unqualified |
| Cards, metric cards, keyed equal-height rows | `components.go`, `metrics.go`, `card_rows.go` | Preserve reviewed old inputs; add explicit contextual source variants instead of globally weakening fit guards |
| Native metrics and number formatting | `data_metrics.go` | Reuse explicit value/format union; no numeric inference from text |
| Frame/chrome/logo and thin rules | `geometry.go`, `render.go`, `rules.go` | Frame needs source art contexts; current rule is only quiet 0.75-point line ink |
| Image crop/focal math | `internal/compose/image.go`, `canvas.go` | Existing contract lists cover/contain/source crop, but WMDS needs its own strict asset and coordinate adapter |
| Native picture cropping and SVG fallback | `pptx.ImageProps`, `ImageSizing`, `SVGFallbackData` | Use native pictures for assets; keep text, tables and diagrams separately editable |
| Phrase selection and measured accent fragments | `internal/compose/accents_phrase.go`, `accents.go` | Reuse selection/geometry ideas; existing accents cannot be assumed qualified for WMDS or middle/bottom-aligned text |
| Routed connectors/process/architecture | `internal/compose/routes.go`, `diagram_path.go`; `internal/adapt/process.go`, `architecture.go` | Extract/adapt useful algorithms; do not run WMDS through the legacy compose font/layout profile implicitly |
| Native shape presets and editable groups | `pptx` shapes; WMDS `groups.go`, `card_rows.go` | Source polygon geometry can differ from presets; grouping must own pictures and graphic frames as well as shapes |
| Comparison semantics | `internal/adapt/comparison.go` | Semantic inputs are reusable; source before/after rows, metric chips, icon implications and fixed anatomy still require a WMDS adapter |

Native shapes and text can implement the source compounds. The Go adapter must
measure them before serialization. Replacing a whole slide with a screenshot
would not complete its content bindings or collaboration/editing requirements.

## Contracts to freeze before expanding the compiler

### Frame, local geometry and paint order

Source slides omit `title`/`eyebrow` when headings are authored in the body.
Covers, several agendas/dividers and quotes therefore require an explicit
headerless compile policy. The source slide renderer always emits footer chrome
and only emits header text/rule when source content exists. `NoHeader` must be
derived from this structure; an empty default header cannot reserve a phantom
body boundary at y126.

Custom whiteboard arrays replace the default patch. Preserve `fade`, `on` and
`above`: an above field is painted after the body, including over a photo/panel.
The Go foundation currently represents one radial header field. Array rectangles,
directional fades and source-layer order require a named extension.

Top-level `on: inverse` often describes an inverse body panel on an otherwise
light slide. Frame surface alone cannot determine contrast. Compile a local
painted-surface context or explicit panel membership, then validate foreground
against that actual surface. Body artwork needs declared full-slide/bleed/cross-
zone bounds: `agenda/index` deliberately crosses x273 with its photo; full-photo
dividers begin at x0,y0; cover/photo reaches x960.

The 18-point outer module remains a root composition rule. Source compounds have
legitimate local coordinates and dimensions: y147/173 inset text, 48-point rows,
78-point cards, 15-point cells, 3-point gaps. Applying `Grid.OuterBox` recursively
will reject much of the source. Introduce explicit contextual compound geometry
with owned bounds and measured fit. Do not round coordinates or silently increase
shape sizes. Distinguish local source geometry from a genuinely unsafe source
height or duplicated object.

### Content and identity

Keep the current `compileTemplateSource` / `BindTemplates` boundary while
splitting family/node compilation into owned modules. A definition should retain
the pinned family-file SHA and recursive identities with source-relative paths,
expected node/cell type, parent context and stable semantic key. No arbitrary
JSON Pointer mutation is needed.

Named binding values supply content, media references and keyed repeated data.
They do not supply template coordinates, base font sizes or visual styles.
Exact template cardinality and semantic states are enforced independently of
advisory character budgets. Every required slot is assigned; sample client names,
analysis claims and contact details cannot survive as silent fallback copy.

Source slot metadata is incomplete in several cases:

- `cover/photo` declares `date` without drawing a date object.
- Quote variants draw a photo but expose only quote/attribution slots.
- `closing/tagline` declares contact name and role while drawing one combined
  contact-role line.
- Rich repeated structures use positional arrays without stable keys.

Resolve these through explicit adapter contracts and source-resolution records.
Fixed design labels, fixed artwork and supplied content need distinct provenance.
Persist caller keys across row reordering; generated ordinals may follow source
order where the design specifies numbering.

### Text, marks and custom styles

The existing rich-run adapter handles weight/ink changes at one base token's
family/size/leading. It does not implement the source `[[...]]` artwork or footnote
superscripts. Header and body emphasis need a range/occurrence binding and Go
cluster/line-fragment placement; underscore/highlight cannot be inferred as native
underline/bold. Circle/spark and standalone dashed arrows are separate assets.

Text `h`/`valign` values require measured top/middle/bottom allocation. Stat
squares bottom-align their stat and label with source padding. Pullquotes add a
60-point curly quote with 30-point leading, and agenda indexes use 11-point Mono
numbers. These custom combinations need explicit resolved styles and conservative
vertical estimates; current token anchors do not qualify them.

No footnote markers or `source.notes` arrays were observed in these 36 variants.
If supplied content enables them later, strict contiguous note references,
superscript face/size/baseline and source-note allocation require a declared IR.
The browser's source-note `nowrap`/ellipsis behavior must become measured Go
rejection or an explicit alternative layout, not unnoticed truncation.

### Full card and compound anatomy

Source cards require `pad: 6/9`, `titleStyle: body`, widths below198, icon
stack/inline/side layouts, label-only body blocks and narrow multiparagraph body
layouts. Source `bodySize: small` is an authored density choice. Map it explicitly
to a named dense variant; never activate it to rescue overflow.

Card inner content and nested source panels need one measured flow model. Tabs,
labels, icons, paragraphs/bullets, metrics and separators must contribute actual
allocated and occupied height. Slot-specific exceptions must not weaken existing
card fit constraints globally. No `card.media` was found in these two families;
other-family media cards should consume the same asset/flow contract rather than
introducing another unmeasured image pipeline.

Before/after rows need keyed before/after/chip content, source row54/gap9 and
authored arrows. Strong numbers need their separate number channel. The vertical
stepper has fixed66-point pitch, 12/18-point state squares and semantic current
state. Schedule/index lists and icon implication nodes need their own compact
measurement rather than pretending to be ordinary cards.

## Source issues that need decisions

1. **Capability-map duplicates content.** Plain value-driver labels/bullets are
   drawn and then later cards repeat the same content over them. Select canonical
   editable objects explicitly and record the decision against the source hash.
2. **Capability-map microcells exceed normal line allocation.** A 15-point cell
   is shorter than the source small12/18 token allocation; narrow cells can also
   wrap. Define a dense bounded typography/geometry resolution, preserving original
   evidence and avoiding silent shrink.
3. **Agenda outcome label without title.** The source `textblock` supports this,
   while the existing bounded adapter rejects it. Add a declared label/body
   variant; do not drop the label.
4. **Vertical divider edge.** A source `rule` has width6 and weight396 with
   callout ink. It is a tall decorative bar, not the existing quiet horizontal
   separator. Compile the actual shape and ink.
5. **Intentional overlaps versus clipping.** Pillar rationale panels overlap
   card ends; photos bleed across panels; matrix cards paint over earlier labels.
   Intentional source paint order needs explicit overlap contexts. It does not
   establish that overflowing text is acceptable.
6. **Compact cards need independent fit.** Goal-step three-paragraph cards at
   width270, pad9/h78 outcome cards and pad6/h36 callout strips require source-
   specific measured limits. A metadata word/character budget is advisory.

## Completion evidence

For each adapter batch, record coverage independently from qualification. Deliver
synthetic bound inputs, editable PPTX and PowerPoint-rendered PDF, structural/text
inspection and actual visual review. Qualification requires a recorded envelope
for replacement content, counts and overflow behavior. Runtime
generation and measurement remain Go-only; native capture/review is development
evidence, not a required user operation for every deck.

No adapter or generic compose capability should be promoted to a qualified
content envelope solely because one reference slide looks correct. Existing
font-file identity and arbitrary-content limitations remain unchanged.
