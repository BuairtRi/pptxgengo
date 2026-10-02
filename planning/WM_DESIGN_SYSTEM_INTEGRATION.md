# Implementing the modern West Monroe design system in pptxgengo

Reviewed 2026-10-01. This is an implementation plan based on the current local
sources, not a claim that the design system is implemented or visually qualified.
The accompanying [inventory](wm-design-system-inventory-2026-10-01.json) records
source hashes, library entries, template dependencies and authored type styles.

The [version 1 implementation contracts](wm-design-contracts/v1/README.md) now
resolve source precedence, typography, bindings/identity, editing behavior and
acceptance. They include machine-readable policies and a concrete binding example.
Those decisions supersede this review's open-ended contract descriptions;
component implementation and native qualification remain pending. The foundation
is implemented, and the [typography qualification controls](../samples/wmds-typography-20261001/README.md)
now expose native wrapping/first-baseline failures. The complete character capture
also exposes occupied-height and paragraph-serialization defects; a corrected
candidate model is needed before proceeding to component qualification.

## Design authority and scope

`/Users/rscott/Projects/wm-design-system` defines the design. pptxgengo should load
and compile its versioned JSON to native PowerPoint. Keep the two responsibilities
as the design repository's README describes them: design definitions there,
PowerPoint generation here.

Current source:

- `tokens/v0/tokens.json`: `wmds.tokens.v0`, point units, 14 type styles.
- `components/v0/components.json`: `wmds.components.v0`, content version 1.6,
  70 definitions: 12 primitives, 36 components, 22 composites.
- `frames/v0/frames.json`: `wmds.frames.v0`, rail/footer choices, chrome and zones.
- Eight `templates/library/<family>.json` files: `wmds.templates.v2`, 97 templates,
  comprising 51 core and 46 working designs.
- `docs/authoring-reference.md`, `docs/icons.md`, and the hand-written
  `explorations/components.src.html` renderer supply field meanings and rendering
  behavior. The generated catalog, previews and React package derive from these.

Core/working are design-library tiers. They do not imply PowerPoint qualification.
Older planning gap reviews describe earlier versions; current canonical JSON and
renderer take precedence when determining what exists today.

## Architecture

Build a Go design-system compiler with this flow:

```text
Versioned WMDS source + slide/template input + asset manifest
    ↓
Load and validate source; retain source hashes and template identity
    ↓
Resolve tokens, grid, frame, density, typography and surface ink roles
    ↓
Expand components/composites into a hierarchical native scene
    ↓
Measure text → compute flow, row heights, ports and accent placement
    ↓
Check zones, fit, contrast, collisions and required chrome
    ↓
Write PPTX, source-to-shape map and layout/fit report
```

Proposed package: `internal/wmdesign` for source loading, semantic validation,
tokens, frames and component expansion. Reuse `internal/textlayout` for Go text
measurement and applicable `internal/compose` layout algorithms. Extend the scene
and writer boundary where native groups, connectors, tables, charts or layouts
cannot be represented by the existing compose plan.

Preserve the logical hierarchy in the intermediate representation: a card,
its title band and its text remain identifiable parts; composites retain their
children, data and layout policy. Flatten only where the writer requires it, and
retain the mapping. Real PowerPoint groups are a separate writer capability.

The existing flat compose-spec v1 is a useful backend for an initial subset.
Making every WMDS feature fit that schema would lose capabilities: it has no
native chart/table plan, frame contract or native component groups. Use a shared
native scene representation as coverage grows, rather than making every template
a separate hand-authored coordinate recipe.

Initial input should support a complete WMDS `slide` node and selection of an
unchanged `id/variant` template. Bound content uses the v1 sidecar contract: named
slots, explicit node/property targets, pinned source hashes, caller-keyed component
arrays and measured envelopes. Current source `slots` are descriptive; each
template needs its own binding sidecar before advertising bound instantiation.

Routine generation and measurement should run in Go. PowerPoint automation stays
in the separate native comparison and final-review workflow.

## Mapping the layers

| Design layer | Go implementation | Existing support / work needed |
|---|---|---|
| Foundations | Versioned tokens, style and grid resolver | Points and RGB exist; runtime WMDS token inheritance is new. |
| Primitives | Native text, fill, lines, preset shapes and pinned pictures | Writer supports much of this; compose exposes a smaller vocabulary and typography contract. |
| Components | Reusable expansion and measured internal flow | Reuse canvas, cards, pods and layout machinery; implement modern WMDS anatomy, spacing and states. |
| Composites | Data-dependent layout with measurable children and ports | Reuse grid, roadmap, team and routing algorithms where their contracts agree; add WMDS tables, charts and other composites. |
| Frames | Resolved zones and native layouts/chrome | Public writer has master/layout and placeholder APIs; compose has no WMDS frame compiler. |
| Slide templates | Frame + components + bindings + content policies | Import 97 designs; qualify each supported variant after compiling its dependencies. |

## Grid and frame rules to implement exactly

- Slide 960 × 540 pt; horizontal content 57–903 (846 pt); 18 pt module and
  3 pt typographic baseline.
- Twelve 54 pt columns with 18 pt gutters. Column `i` starts at
  `57 + 72*(i-1)`; an `n`-column span is `54*n + 18*(n-1)`.
- Outer container lattice: `x = 3 + 18k`, `y = 18k`. Component padding and
  internal glyph placement have their own rules; do not snap every native child
  to outer container edges.
- Five-up is an explicit exception: 5 × 154.8 pt, 18 pt gutters, starts
  57/229.8/402.6/575.4/748.2. Full-width, no rail; separate from bands using
  twelve-column inner edges. The older gap memo's 151.2 pt is superseded by tokens.
- Rails: none/nav/left/right. Left main 345–903, right main 57–615; panel and
  rail content boundaries come from the frame source, not inferred proportions.
- Header: one-line body top 126; two-line 162; appendix 108. No third title line.
- Compact/tall body bottoms 468/450; source reservation reduces body bottom to
  450 or 432 for one/two lines. Rail chrome adapts to the surface beneath it.
- Legal line and whiteboard field required on every slide, including covers and
  dividers. Logo, page, source, stamp and optional footer metadata follow frames.
- A native slide-number placeholder should stay dynamic. Materialize legal text
  with an explicit generation year recorded in the build manifest.
- Frame zones are layout constraints. Native placeholders should be emitted
  where editable content semantics fit; a body containing multiple independent
  components need not become one placeholder.

## Capability gaps

| Area | Evidence in pptxgengo | Required addition |
|---|---|---|
| Numeric font weight | Compose runs expose `bold`; Go resolver selects 400/700. | Resolve 500/600 to real static faces for WMDS native v1; map authored family/weight to the actual PowerPoint typeface and audit face identity. Provision static Plex Sans; no substitution of bold for semibold. |
| Exact leading | Writer has `LineSpacing` → `spcPts`; compose and Go use multiples. | Add exact point leading throughout measurement, planning, emission and verification. WMDS body is 14/21, title 32/36 and display 54/60. |
| Tracking | Writer has `CharSpacing` → run `spc`; its nonzero path also writes `kern=0`. | Convert em tracking to serialized 0.01 pt values. WMDS v1 explicitly requests metric kerning at all sizes, including zero-tracking runs. Measure actual resulting wraps. |
| Footnotes | Writer supports superscript/baseline; compose runs lack these fields. | Parse `[^n]` into real raised runs and source references, preserving reading order and visible text identity. |
| Paragraph defaults | Previous calibration identified implicit paragraph-end font inheritance. | Make new typography's paragraph-end defaults explicit and compare the resulting native allocation. |
| Surface/contrast | Compose has explicit colors and effective backgrounds; text threshold is broadly 4.5:1. | Resolve WMDS ink roles per surface and implement its size/weight-aware 4.5:1/3:1 policy. Do not reuse older staffing colors as WMDS meanings. |
| Backgrounds and outlines | Current compose writer starts slides white; shape canvas lacks general outlines. | Add explicit slide surfaces and outlined/no-fill shapes, including dashed containers and highlighted matrix rows. |
| Preset geometry | Public writer has general presets; compose allows rect/homePlate/star5/rightArrow/blockArc/triangle. | Expose needed chevron, ellipse, diamond, trapezoid, cylinder/can, plane and indicator geometries with audited adjustments. |
| Connectors | Compose routes emit editable segments; endpoint behavior explicitly says `editable_segment`. | Add dotted/both-headed/routed lines and native connection-site references for diagrams intended to follow moved nodes. Coordinate-only source connectors cannot safely infer semantic attachment. |
| Native groups and identity | Object names and scene extraction exist; ordinary compose output is flat. | Add native component groups where required and source-to-shape identity. Preserve authored IDs; path IDs are provisional for unchanged imported fixtures and are not stable after reordering. |
| Native tables | `Slide.AddTable` exists; compose grid cells emit shapes/text. | Add a measured table plan, native table emission, rich cells, total/group rows, indicators and continuation policy. Avoid relying on character-count auto-pagination. |
| Native charts | Writer has `AddChart`, `AddMultiChart`, embedded XLSX and chart options. | Add WMDS data schemas, formatting, chart style contracts and verification. Preserve editable data and provide a data-table alternative. |
| Assets | Compose supports pinned PNG/JPEG and bounded SVG + PNG fallback. | Resolve WMDS photo/icon/logo/mark names into a versioned asset manifest; preserve focal crops and provenance. Handle requested grayscale explicitly. |
| Phrase marks | Native phrase bounds and artwork placement exist; Go phrase requests are unsupported. | Supply Go range geometry and layout bindings for per-line marks; qualify it before claiming Go-only emphasized-title generation. |
| Review note | Source requests a group partly on the pasteboard at x=942. | Introduce an explicit non-presented review-object category; ordinary slide bounds/collision checks cannot accept it unchanged. |
| Notes/accessibility/links | Writer already has speaker notes, image/chart alt text and hyperlinks. | Expose appropriate semantics and preserve source metadata; add reading-order and text/shape accessibility handling where absent. |

The previous 393 + 45 font probes and two-slide example are useful baseline
evidence. They do **not** qualify WMDS's new weights, tracking, exact leading,
footnotes, or frame typography. New controls are needed for those changes.

## Component implementation families

1. **Text/headings/lists:** styled text, group/number/column headings, text blocks,
   bullets, numbered/strong-number/index lists and schedule rows. Support explicit
   mixed runs and measured vertical flow without browser clipping.
2. **Card anatomy:** all seven surfaces including outline; edges inside bounds,
   surface-specific title bands, shared row chrome, numbers, icons, badges, media,
   paragraph/list/column/checklist body, metrics, person/bio/case/quote/fee states.
3. **Data indicators:** number/date formatting, metric additions, legends,
   callouts/quotes, statuses, ratings, Harvey balls, maturity and gauges.
4. **Media/marks:** thumbnails and stacks, square/focal images, logos, icons,
   highlight variants 1–4, underscoring/circles/sparks, arrows and whiteboard fields.
5. **Sequence/structure:** steppers, vertical steps, phase headers/columns,
   timeline/Gantt/horizons, people/roles/pods/org/governance.
6. **Diagrams/frameworks:** matrix, layered/nested architecture, swimlane,
   cycle/hub, pyramid/funnel, before/after and annotation.
7. **Tables/charts/commercials:** typed cells, fee summaries, pricing cards,
   editable charts and matching workbook number formats.
8. **Collaboration:** review-note geometry plus metadata, isolated from final
   slide content constraints.

Similar names in the old pptxgengo component catalog do not establish matching
contracts. Reuse algorithms and writer machinery; resolve modern geometry and
style from WMDS.

## Build sequence and concrete completion criteria

### 1. Source loader and coverage report

Load source via a configurable path, record schemas/content version/hashes, detect
unknown nodes and fields, and report dependencies by template key. Separate
renderable node types from typed table-column/cell descriptors and reference-board
demo helpers. The inventory's recursive `type` counts deliberately include these
descriptors and are not a renderer coverage percentage.

Package a reproducible source snapshot for installed releases so generation does
not require this user's sibling checkout. Keep local source override available.

### 2. Typography, tokens and geometry

Implement real weights 400/500/600/700, exact leading, tracking, paragraph defaults,
surface roles and grid/frame resolution. Pass the complete style contract through
probe keys, font provenance, Go layout, writer and native auditing.

Completion: each of the 14 type styles can be represented without substitution;
regular/bold behavior remains explicit; grid and frame zones resolve exactly.

### 3. Frames and first representative slides

Compile all rail/footer choices and mandatory chrome; build a small representative
deck using the actual source templates: `cover/photo`, `key-message/stat`,
`phase-detail/rail-table`, `team/pods`, and `pricing/capacity`. Include one/two-line
headers, source reservation, dark surfaces and a five-up control. Some of these
require a thin table/card slice before the full families are finished.

Completion: a reviewable native deck with editable text/shapes/data, retained
template identity and explicit unsupported-feature errors. Compare against source
previews, then inspect native PowerPoint rendering. Building alone is not approval.

### 4. Components and composites

Implement the families above in dependency order driven by the template inventory.
Cards, lists and text blocks unlock the greatest reuse. Add native tables/charts
and attached connectors as supported scene types rather than rasterizing them.

Completion: component parameters, data cardinality and overflow behavior are
explicit; fixture success and changed-content qualification are recorded separately.

### 5. Full library and content bindings

Import all 97 designs, instantiate durable v1 slot bindings and repeated-content policies,
and expose template selection/instantiation. Report template availability by actual
dependency support, native review and content envelope, separately from core/working.

Completion: every intended template has implementation status and provenance;
unsupported variants are named; changed content gets fresh measurement and a fit
report. Splitting/continuation is explicit, never silent font shrinking.

### 6. Release and authoring workflow

Ship the source snapshot, asset manifest, generator, documentation and reviewed
examples. Preserve editable identity/notes and record versioned output manifests.
Update installed CLI and skills only after this implementation has a concrete
reviewed output. No release/install changes were made for this review.

## Source differences to handle deliberately

- `outline` is a supported surface in component rules and the renderer but is
  not a named surface-role object in tokens. Resolve it as the explicit outline
  treatment with light-surface inks; record this semantic rule.
- The older icon primitive lists placeholder sizes/names; current `docs/icons.md`
  and renderer use the official 222-icon library and support legacy aliases.
- General text rules reserve small type for specific uses; current cards/lists
  explicitly permit `bodySize/size: small` in dense zones. Encode those contextual
  exceptions instead of silently changing the whole slide to appendix mode.
- The authoring reference says whiteboard only on White in its general rules,
  while current frame definitions and slide-field renderer specify dark-surface
  fields. Follow the explicit frame-field contract for frames; retain the separate
  restriction on the standalone whiteboard demo primitive.
- The reference renderer uses browser clipping/ellipsis for some source lines and
  fixed containers. Native generation must report overflow instead of treating
  clipped content as successful fit.
- `slots` and the React props are not complete validation schemas: generated
  TypeScript interfaces are inferred from examples. The Go loader needs explicit
  node contracts, defaults, enum handling and validation.
- `_gaps.json` includes intentional approximations and historical template keys.
  Implement current authored nodes faithfully; keep speculative extensions,
  deferred charts and old source fidelity work separate from current coverage.

## Typography v2 candidate update (2026-10-01)

The opt-in `wmds-go-foundation.v2` now implements corrected cluster advances,
terminal tracking, standard ligatures at nonzero tracking, fitted vertical
estimates and explicit trailing/empty paragraphs. Normal IBM names are preserved.
Generation and measurement execute in Go without native automation.

The [139-control v2 PDF](../samples/wmds-typography-v2-20261001/README.md)
matches all predicted line breaks and visible baselines within 0.12 pt. The
[18-slide foundation candidate](../samples/wmds-foundation-v2-20261001/README.md)
is reviewed. Original native development replay matches all 75 wraps and 74
comparable heights, but those height anchors are fitted evidence. Independent
[v2 character capture](../samples/wmds-typography-v2-20261001/NATIVE_RESULTS.md)
now passes all 139 controls: content, native wrapping, effective observed styles,
frames, occupied bounds and fit. Maximum occupied-height error is 0.000041 pt;
native line-advance error is 0.132 pt. Four terminal empty entries are omitted by
the scripting collection while exact content and saved paragraph/default structure
are preserved. Exact native file/instance access remains unobserved, so a general
engine envelope remains unqualified. Measured text blocks and basic cards are
the next implementation slice; composites and template bindings follow.

Native capture is a development/reference operation, not a per-deck generation
requirement. Additional fonts, features or material renderer changes can require
new reference measurements.

## Review boundary

Implementation update, 2026-10-01: the
[first foundation slice](../cmd/pptxdesign/README.md) is implemented as the opt-in
`pptxdesign` command. An 18-slide native reference, Go layout report and separate
PowerPoint PDF review receipt are in `samples/wmds-foundation-20261001/final`.
This establishes the source loader, numeric-weight static fonts, grid and frame
geometry/chrome; component/template execution and native text-bound qualification
remain subsequent work. The historical review boundary below describes the
initial inspection, before this implementation.

This review read canonical JSON, field documentation, the reference renderer,
React adapter, tooling and relevant Go measurement/compose/writer code. The design
repository was left unchanged. No deck was generated, no test suite was run and
no native or visual qualification was performed. The browser surfaces were
unavailable when attempting to open the existing local reference board.


## Measured text-block/basic-card slice (2026-10-01)

The opt-in v2 command now renders `text.block` and the basic paragraph, bullet,
inline-number, band and accent-edge card anatomy as real native groups. Typed
Go/JSON contracts specify required content, geometry, measured height, contrast,
explicit density and overflow errors. See [the contract](wm-design-contracts/v1/components.md)
and [the 12-slide reference](../samples/wmds-components-20261001/README.md).

The reviewed reference has 34 groups, 153 preserved text objects, 160 paragraphs
and 195 PDF line matches. Twenty-two negative inputs return explicit diagnostics.
One earlier exact-width wrap mismatch remains in the evidence; component input
now rejects uncertain boundaries within 0.25pt of width. This exclusion does not
establish general native parity. Filled native shapes now serialize an explicit
no-line setting. The navy rail fills the slide edges in PDF and Slide Show;
PowerPoint's editing canvas border is application UI.

The runtime remains Go-only. Metric cards, mixed text runs, remaining component
presets, composites and template binding are the next implementation work.


## Metric-card slice (2026-10-02)

The opt-in v2 command now accepts typed `MetricSpec` and `MetricGroupSpec` card
content. Single values/labels, comparisons, authored changes, KPI status marks
and grouped metrics use Go measurements, native text/shapes and one editable
card group. Existing surfaces, labels/titles, bands, edges and frame zones apply.
The [metric contract](wm-design-contracts/v1/metrics.md) fixes footer anchoring,
comparison baselines, column proportions and status-outline token precedence.

The [nine-slide reference](../samples/wmds-metrics-20261002/README.md) has 28 native
groups, 158 text objects, 160 paragraphs and 163 matching visible PDF lines.
Every page was visually reviewed after local PowerPoint export. XML geometry,
styles and colors agree with the report. The left rail reaches its edges in PDF
and Slide Show. Two source examples require taller/wider geometry under these
measurements; the contract records their fit conflicts without changing the source
snapshot or typography calibration.

Runtime remains Go-only. General font/content qualification and exact native file
identity remain unqualified. No new character capture or test suite occurred.
Next: keyed card rows/composites, remaining card presets, and slide bindings.
Standalone data metrics, numeric formatting and mixed runs are still deferred.


## Parallel component slices (2026-10-02)

Three independently implemented slices now share the opt-in Go v2 renderer:

- [Keyed card rows](wm-design-contracts/v1/card-rows.md): 2–4 equal cards with
  18pt gutters, common measured height, inline/band numbering and nested native
  groups. Keys persist through reordering; ordinals follow input order.
- [Standalone metrics and formatting](wm-design-contracts/v1/data-metrics.md):
  values/labels, native change triangles, comparisons, authored target/status/source
  and exact decimal currency/percent/number/delta/range formatting.
- [Mixed text runs](wm-design-contracts/v1/rich-text.md): one editable text box
  per rich node, real static weights and surface inks, combined paragraph shaping
  and wrapping, explicit base defaults for empty/trailing paragraphs.

The [19-slide combined reference](../samples/wmds-parallel-slices-20261002/README.md)
opened in PowerPoint without repair and was exported locally. All pages were
reviewed individually at 1920×1080. Package/report checks passed for 55 groups,
236 text objects, 242 paragraphs and 47 rich runs. All 255 predicted visible lines
matched the native PDF. Nine row groups preserve 28 nested cards in keyed order;
18 groups are standalone metrics. The slide-13 navy rail has 3,252 opaque edge
pixels and no white perimeter. Rebuilding the published input reproduces every
PPTX part except the core creation/modification timestamps.

This is bounded reference evidence. Mixed-run vertical estimates use existing
uniform anchors where available and conservative bounds otherwise; 14 fixture
runs lack a vertical anchor. Their heights, baseline alignment and ligatures are
not generally qualified by PDF line matching. Rich layouts retain qualification
flags of false. Exact native font-file selection and per-character bounds remain
unobserved for this new reference. The prior 139 uniform controls remain separate.

Runtime remains Go-only; capture/export is development review. Normal IBM Plex
names, frozen inputs, calibration and the installed release remain unchanged.
No tests were added or run. Next integration work is slide-template binding over
these supported nodes, with rich card/bullet content and remaining presets handled
as explicitly scoped additions.


## First bound-template slice (2026-10-02)

Four exact frozen variants now compile into the supported Go components:
`cards/3`, `cards/4`, `stats/four-metrics`, and `takeaway-rail/metrics-rail`.
The [execution contract](wm-design-contracts/v1/template-execution.md) resolves
closed slot names, content classification, array cardinality/identity, source
geometry and fixed capacities. Source hashes and node pointers are verified;
supplied values never fall back to sample content. Cards bind one composite;
metrics bind separate named slots. Eligible body/rail text accepts explicit rich
runs at the source token size/family. A bounded native decorative rule supports
the two metric layouts.

The [eight-slide reference](../samples/wmds-templates-20261002/README.md) exercises
all four adapters twice. PowerPoint opened it without repair. All pages were
reviewed after local PDF export; 133 visible lines, 112 text objects/paragraphs
and 30 native groups passed package/PDF inspection. Source-to-output inspection
confirms 70 slot assignments and four pairs of identical source geometry/styles.
Both right navy rails meet their outer edges. Rebuilding the bound content yields
identical non-timestamp PPTX parts. Three rich runs retain uncalibrated vertical
estimates; general content/character bounds and exact native font-file identity
remain unqualified.

The commands are `templates`, `template-reference`, and `template`; all require
explicit v2 selection. Their typed binding contract is separate from the historical
v1 general JSON-pointer binding proposal. That proposal is still not an arbitrary
source-pointer execution engine. Runtime remains Go-only; frozen sources, fonts,
calibration and installed release are unchanged. No tests were added or run.

Next work can expand opener templates (cover/divider/key message), add rich card
content, or qualify hand-drawn emphasis and footnotes before enabling templates
that require those features. Media, tables, charts and complex diagram templates
remain outside this binding slice.

## Full-library parallel completion plan (2026-10-02)

The [completion plan](wm-design-contracts/v1/library-completion.md) replaces the
open-ended next-slice list with a finite backlog covering all 97 variants. The
eight live template-family files match the frozen source. Four variants have
executable bound adapters; 93 remain. Core/working design tiers do not change
that implementation count.

Three parallel agents audit disjoint families and nested features, with one
integrator responsible for shared compiler/binding/report interfaces. Their
coverage records are merged into
[the full ledger](wm-design-contracts/v1/template-coverage.json), retaining source
hashes, exact features, dependencies, source conflicts and binding work.
`scripts/build-wmds-template-coverage.py` reproduces the merge and rejects missing,
duplicate, unknown or source-mismatched template records.

The proposed implementation tracks are text/primitives/media, cards/native
tables, and diagrams/sequences; native charts start as table capacity frees.
Template adapter batches and native review follow dependency completion. Shared
source compilation, token palettes, native text/group postprocessing and release
integration are centrally owned to avoid concurrent edits to shared files.

This is a planning/coverage pass. No new renderer support, template binding,
font/calibration change, deck/native capture, installed release change or test
execution is implied by the audit. Implementation and specimen/reuse qualification
remain separate states.
