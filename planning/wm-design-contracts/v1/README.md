# WMDS → native PowerPoint contracts, version 1

Status: normative implementation specification, established 2026-10-01. The
[foundation slice](../../../cmd/pptxdesign/README.md) now implements typography,
grids and frames with an independently reviewed reference deck. Remaining
component/template contracts and full native text qualification are future work.
The v2 [text-block/basic-card contract](components.md) is now implemented, with a
reviewed 12-slide reference. The later [card rows](card-rows.md), [standalone metrics/formatting](data-metrics.md)
and [mixed runs](rich-text.md) are implemented with a reviewed 19-slide reference.
Four [bound templates](template-execution.md) now execute with an eight-slide
reviewed reference. Remaining presets and template adapters are future work.
The [typography control slice](../../../samples/wmds-typography-20261001/README.md)
now supplies 75 generated controls and native PDF comparison. Seven wrapping
boundaries and the provisional first-baseline rule fail. The complete native
character capture also exposes occupied-height differences and trailing-break
serialization loss. The foundation has not advanced to native typography qualification.
The specification as a whole is not a runtime capability claim. It applies to the source snapshot in
[the inventory](../../wm-design-system-inventory-2026-10-01.json).

The [policy file](contract.json) captures key machine-readable decisions. The
[binding schema](template-binding.schema.json), [binding example](key-message-stat.binding.json)
and [content example](key-message-stat.content.json) make the content API concrete.

## 1. Authority, precedence and compatibility

**Decision:** WMDS remains the design authority. A versioned adapter contract
resolves execution details and names the specific source differences it resolves.

Resolution order is:

1. These explicitly listed adapter resolutions and supported context exceptions.
2. Canonical tokens for values, frame JSON for geometry/chrome, component JSON
   for anatomy and behavior, and template JSON for authored component composition.
3. Specific current authoring-reference rules for details absent from JSON.
4. Reference-renderer behavior for remaining defaults or internal geometry.
5. Older planning documents and generated previews as historical/reference material.

This is domain ownership, not a rule that a template may override any foundation.
An explicit card surface is valid; an arbitrary font size or off-grid outer
rectangle is not automatically valid because it appears in a template.
Unhandled contradictions produce `source.conflict` with both locations. Unknown
render-affecting fields produce `source.unsupported_field`; they are never dropped.
An extension namespace may hold non-rendering metadata without implying support.

### Specific resolutions

| Difference | Binding rule |
|---|---|
| Five-up width in old gap memo | Use 154.8 pt from tokens, never 151.2. |
| `outline` absent from token surface-role dictionary | White fill and light-surface inks; 1 pt Medium Gray outline inside the container bounds. |
| Old placeholder icon description vs current icon documentation | Official icon catalog, original viewBox, sizes 36/54/72; legacy names map through an explicit alias table. Other sizes require a named component exception. |
| General whiteboard-on-White rule vs dark frame fields | Frame fields support light/dark surfaces with prescribed dot ink. The standalone whiteboard primitive keeps its light-only restriction. |
| Card/list small body text vs general minimum | Standard body remains 14/21. Explicit small bodies/lists are allowed only in designated dense component zones, rail content, roles, table cells or appendix. Record the context; never infer density from overflow. |
| Card minimum width vs narrow rail/five-up examples | General card minimum is 198 pt. Explicit rail/sidebar cards and sanctioned five-up cards may be narrower with small padding; qualify those as separate contextual variants. |
| Page style vs footer token | Legal uses footer 7/9 at weight 400; page uses the same metrics with weight 600, as frames specify. |
| Browser clipping or ellipsis | Report overflow; preserve the supplied content. A clipped preview does not confer fit approval. |
| Preview fallback page `12` | Use actual generated slide sequence. Authored numeric `page` is fixture/reference metadata; it does not hard-code the native field. Covers may set `noPage`. |
| Preview nav labels/active tab | Treat the four labels and second active tab as fixture defaults. Bound slides must provide explicit nav labels/active ID when using a nav rail. |
| `uses` names differ from component IDs | `uses` is descriptive provenance. Actual renderable nodes and their typed fields determine dependencies; do not dispatch by fuzzy component name. |

Source snapshots are immutable build inputs. A change to their hashes requires a
new snapshot and explicit binding migration. Installed packages include the source
snapshot; a local source path is configurable. Existing compose v1 and engine v4
remain independently versioned contracts; new WMDS defaults must not reinterpret
their saved evidence.

## 2. Typography and font identity

### Authored style

Each resolved run has an explicit family, numeric weight, italic flag, size in
points, tracking in points, kerning policy, color, and baseline shift. Each
paragraph has explicit alignment, exact leading in points, before/after spacing,
bullet/indent settings and paragraph-end style.

- Load the 14 token styles exactly, including case transforms. Retain the original
  content and the displayed text in the manifest; uppercase changes are explicit.
- Numeric weights in the initial WMDS profile are 400/500/600/700. Resolve to real
  font faces. Unavailable weights are errors, never synthesized or rounded.
- Legacy `bold` converts to 400/700 only at an explicit legacy-input boundary.
  Conflicting `bold` and numeric weight declarations are errors.
- No family substitution, glyph fallback or hidden condensed face. Default width
  axis is 100. Non-Latin/bidi/tab support remains outside the initial Go profile
  until separately implemented and qualified.

### Resolved font

The font manifest stores authored family/weight/italic separately from actual
typographic family, style, legacy family/subfamily, PostScript name, file hash,
face index and any axes. The writer's typeface and bold/italic flags are an explicit
mapping to that resolved face, not a guess from the authored family string.

**Native WMDS v1 uses static faces.** A variable file can be used for Go exploration
but cannot establish native weight identity merely because the Go shaper selected
`wght=600`. A future variable-native route needs evidence that the PPTX selection
and PowerPoint renderer use the same instance. Embedding a file alone is not that
evidence. Font embedding is optional and must respect the file's embedding rights.

Local inspection found static Plex Mono 400/500/600/700, including:

| Authored style | Actual legacy family / subfamily | PostScript name |
|---|---|---|
| Mono 400 | IBM Plex Mono / Regular | IBMPlexMono-Regular |
| Mono 500 | IBM Plex Mono Medium / Regular | IBMPlexMono-Medium |
| Mono 600 | IBM Plex Mono SemiBold / Regular | IBMPlexMono-SemiBold |
| Mono 700 | IBM Plex Mono / Bold | IBMPlexMono-Bold |

The inspected Plex Sans files are variable, with named 500/600 instances and
`wght` 100–700, `wdth` 75–100. Static Plex Sans faces for this native profile must
be provisioned before qualification. These observations are local prerequisites,
not portable font-availability guarantees. Record exact installed bytes per run.

Native audit checks actual selected face, size, bold/italic and resolved font
metadata; accepted native font-name aliases must be backed by inspected face
identity. A returned family string alone cannot prove weight 600.

### Leading, tracking and kerning

- Exact token leading becomes paragraph `spcPts`, not an inferred spacing multiple.
  Reject simultaneous exact and multiple modes. Body 14/21 means 14 pt glyph size
  and 21 pt requested line advance; it does not mean a 21 pt occupied glyph box.
- Paragraph before/after default to zero. Component gaps are layout gaps, not
  implicit paragraph spacing. Line breaks and empty paragraphs remain explicit.
- Convert em tracking with `tracking_pt = tracking_em * font_size_pt`, then round
  once to the writer's 0.01 pt precision. E.g. title −0.02 × 32 = −0.64 pt;
  eyebrow 0.08 × 10 = 0.80 pt. Measurement uses the serialized value.
- WMDS v1 requests font-metric kerning at all sizes, independently of tracking.
  Write `kern=0` explicitly even when tracking is zero; shape with metric kerning
  enabled. `kern=0` is a minimum-size setting, not a request to disable kerning.
  See Microsoft's [DrawingML primer](https://download.microsoft.com/download/e/1/4/e14fb96f-83b8-4a2a-84db-7fa8acbe061a/Office%20Open%20XML%20Part%203%20-%20Primer.pdf).
- Treat line breaking, glyph clusters, tracking at run boundaries, trailing space,
  ligatures and first-baseline placement as part of the versioned shaping profile.
  Record the selected OpenType feature settings. Do not invent a new per-font
  correction to make one reference pass.
- Write explicit paragraph-end font/size/style/color/tracking/kerning defaults.
  Footnote runs do not become the paragraph-end base style. Empty paragraphs carry
  their authored paragraph base style. This removes dependence on theme font order.
- First-baseline and occupied-height prediction remain engineering calibration
  work. Exact leading and face selection are already fixed requirements; the
  first implementation must measure their native results before claiming parity.

The [writer API](https://learn.microsoft.com/en-us/dotnet/api/documentformat.openxml.drawing.textcharacterpropertiestype?view=openxml-3.0.1)
exposes separate spacing, kerning and baseline properties; implementation must
preserve these separately through the Go/native evidence boundary.

### Footnotes and emphasis

`[^n]` refers to the nth source note. References must exist and be contiguous in
first-use reading order; repeated references may reuse a note. Unreferenced notes
are errors in bound mode. Never renumber silently. Emit 0.6 × base size at weight
600, emphasis ink and +30% base-font baseline shift, explicitly in the resolved
run. With the 0.6 size scale this is `baseline=50000` relative to the marker's
own size; preserve the point offset in the IR. This shift is an adapter default
requiring native visual qualification.
Use displayed digits, not a pasted Unicode superscript substitute.

`[[phrase]]` is parsed into a single emphasized range; markers do not enter visible
text or character budgets. Reject unmatched/nested markers. Highlight/underscore
may produce one artwork fragment per wrapped line; circle/spark require one line
in v1. Allow at most one hand-drawn emphasis per slide. Explicit variant wins;
otherwise highlights cycle 1→2→3→4 across highlight uses in deck order.

## 3. Geometry, surfaces and overflow

- All authored geometry is in points. Preserve floats; convert to EMUs only at
  serialization. Do not repeatedly round layouts to pixels or 1/8 pt.
- Enforce outer box x-lattice and y-module, plus declared five-up/bleed/context
  exceptions. Text baselines/internal padding may use 3 pt rhythm or explicit
  component offsets. A valid node with `y=243` is not rejected merely because
  an outer container would require an 18 pt boundary.
- Explicit width and height are hard capacities. Absent text height is measured
  after width resolution and capped by its assigned zone or next flow element.
- Frame chrome reservations and component sibling envelopes are real constraints.
  A slide without a standard header may author body nodes in the unused header
  area. Cover/display compositions have explicit content/media zones. Do not
  force every body node below y=126 when the chosen template has no header.
- Two-line title allocation is explicit; no automatic third line or movement of
  fixed body coordinates. An adaptive template may select one/two-line header
  only if its body is zone-relative and that behavior is declared in its binding.
- Resolve ink from the nearest explicitly selected surface (`on` for transparent
  text; component surface for contained text). Bands own their surface. Reject
  ambiguous background inheritance; no pixel sampling of artwork for contrast.
- Contrast is 4.5:1 below 18 pt, except real weight ≥700 at ≥14 pt may use 3:1.
  Weight 600 is not treated as 700. Other text ≥18 pt and non-text marks use 3:1.
  Source-prescribed low-contrast combinations are reported, never auto-recolored.
  Decorative background whiteboard fields are exempt; meaningful indicators are not.
- Explicit component overlap (badge, accent, surface, media bleed) is allowed via
  ownership and layer rules. Unrelated overlap remains a collision.
- The default overflow outcome is a named error with required/available dimensions,
  source node, slot and suggested split/variant. Never shrink type, hide text,
  truncate, replace content or switch to appendix to make it fit.
- Continuation occurs only for a template declaring `overflow: continue` with a
  continuation template and split boundaries. Break between whole rows/items;
  repeat declared chrome/headers and keep IDs and totals consistent. If one item
  cannot fit, fail. V1 bindings default to `overflow: error`.

## 4. Template content API and identity

### Binding sidecar

Keep binding/identity sidecars in pptxgengo while the design repository is read-only.
Each sidecar pins a template key, source file hash and contract version. A named
slot binds one or more explicit content fields on persistent node IDs. Targets
use relative JSON Pointers into that node. The reserved ID `$slide` names the root.

An identity overlay assigns IDs using source pointers only during import of that
exact source hash. Store those assigned IDs; never recompute them after edits or
guess a new pointer after source drift. A later source update requires migration
that preserves IDs where identity truly persists.

The key-message example makes a real ambiguity concrete: its `body` slot spans a
claim and support paragraph. It is an object `{claim, support}`, with a shared
180-character guidance budget. `stat` binds to the square's `stat` property, not
to a metric node inferred from the slot's name.

### Input and limits

- Bound mode requires a binding definition. Unknown slots/fields are errors.
  Required values must be supplied; fictional sample copy is not a missing-value
  fallback. Fixed labels are explicitly classified as fixed copy.
- Fixture mode preserves the exact authored sample and is labeled fictional.
  It is allowed for reference import without complete bindings.
- JSON-schema type/item limits are hard semantic limits. Existing `maxChars` and
  word budgets become advisory content guidance, not fit certificates. Count
  displayed Unicode code points after case/format transforms, excluding markup.
  Measured/native fit is always required for the advertised qualification state.
- Binding IDs and node IDs are unique; content-only slots may not mutate geometry,
  typography or chrome. Style/layout customization is a separate explicit API.
- Missing targets, changed target types, nonexistent properties or drifted hashes
  are errors. No fuzzy matching by content or array index after import.

### Repeating content

V1 array bindings select one named component data field, never an arbitrary list
of unrelated slide nodes. Every bound item
has a caller-provided immutable `key`. Input order controls visual order; keys
control identity. Duplicate/missing keys are errors.

For a component field, projection is explicit, e.g. `{key, text}` input becomes
legacy `roles: [text]` plus keyed role identities in the IR. The binding names
the key pointer and value pointer. V1 does not clone disconnected template nodes
as a repeating group. Express such a group as a composite with a declared array
field first. Prototype-based repetition needs a later contract naming owned nodes,
local bounds, placement and min/max count. No runtime inference of prototypes
from visually similar neighboring objects.

Binding transforms have fixed meanings:

- `identity`: replace the target with the validated value of the same semantic
  type; strings remain content and may contain the explicitly supported markup.
- `format_number`: consume a raw number-format specification from the source
  `numberFormats` contract and write its displayed string. Bindings cannot execute
  arbitrary formatting code or invent currency/scale defaults beyond that contract.
- `asset`: resolve a registered asset ID, preserving byte hash, variant and focal
  metadata. An unknown ID is an error; no fetching or photo substitution at build.
- `keyed_array`: project the declared item value into the component's data field
  and retain the caller keys alongside it in the IR. Bounds/count policy comes
  from the component and binding content schema.

The embedded content schema uses object/array/string/number/integer/boolean types,
properties/required/additionalProperties/items, enum/const, numeric bounds,
min/maxItems and min/maxLength. Reject other keywords until implemented, except
annotations and `$schema`. No external references, coercion or guessed defaults.
Binding-schema validity alone cannot check source pointers, target compatibility,
duplicate writes, array keys, slot coverage or physical capacity; the semantic
loader must check those before creating a plan.

Generated child identity is `(slide instance ID, component ID, item key, part ID)`.
Moving/reordering items preserves identity. Removing an item removes exactly its
owned children. Reusing identical labels does not merge objects.

Fixed templates retain a bounded count/layout. Composites may grow measured rows
inside their reserved region. For example `team/pods` cannot simply grow its pods
past the client-role band at y=318; supporting more roles requires a declared
envelope, another variant or continuation, even though the underlying pod renderer
can compute an arbitrary role count.

## 5. Native editability and regeneration

**Decision:** editable native output means meaningful constituent objects remain
editable; component movement and semantic regeneration have explicit boundaries.

| Object | Required native representation / behavior |
|---|---|
| Text and labeled shapes | Native text; preserve runs and paragraph properties. A labeled native shape may own its text rather than duplicate it in a floating box. |
| Fixed component such as card, pod, thumbnail | Native group for its shapes/text/pictures, with named parts. Move the group as a unit. |
| Composite | Retain component subgroups and data identities. Do not group unrelated diagram nodes with connectors in a way that prevents independent node movement. |
| Semantic diagram relation | Native connector attached to source/target connection sites. Ports explicitly identify the owning connectable shape. |
| Coordinate-only authored connector | Native editable line/path with fixed endpoint coordinates; report `attachment: coordinate`. No guessed attachment. |
| Native table/chart | Editable table/chart and embedded workbook for chart data; separately editable overlays for unsupported cell/label adornments. |
| Brand icons/marks/grid | Pinned SVG picture with matching PNG fallback by default. Artwork is not promised to be path-editable. An editable-dot option is a separately declared variant. |
| Frame | Static chrome on native layouts/master where feasible; variable titles/sources/nav and content on the slide. Dynamic page field and explicit layout identity. |
| Review note | Named review-only group and metadata, partly outside the slide; explicit export inclusion policy. |

Output intent is `review` by default: include explicitly authored review groups.
`presentation` omits those groups while retaining review metadata in the manifest.
Neither mode may omit required legal text or whiteboard chrome.

PowerPoint-native connectors may reroute according to PowerPoint. Manual movement
does not run the Go collision planner. Scaling a group does not promise preservation
of type sizes, grid compliance or semantic reflow. Semantic resizing/reflow is
performed by regeneration from data.

Each generated object has an opaque native shape ID and a name carrying the
logical identity; the sidecar manifest maps both. Native connection references use
actual package IDs, not shape-name text. Names alone are insufficient proof after
manual duplication or deletion.

Regeneration always writes a new deck. It never silently incorporates or discards
manual changes in an existing deck. A later recovery route may import unchanged
formatting/content edits for explicitly supported components; unsupported geometry,
grouping, font, chart or table changes produce a reconciliation report. No broad
round-trip editing guarantee is part of v1.

## 6. Acceptance and advertised status

**Decision:** native rendering is the authority for PowerPoint behavior. Browser
previews establish design intent. Whole-image pixel identity is not an acceptance
requirement because browser and PowerPoint text rasterization differ.

| State | Required evidence | Allowed claim |
|---|---|---|
| inventoried | Exact source hashes and dependencies | Design exists in the source library. |
| implemented | All required fields represented, editable package emitted, structural inspection | This variant can be generated. |
| native_verified | Exact deck, font/profile and native text/geometry/data checks | This exact deck passed native checks. |
| visually_reviewed | Native render of every slide reviewed with artifacts/findings | This exact rendered specimen is accepted. |
| qualified_envelope | Implemented/native/visual evidence plus declared content/count/density variants and negative overflow cases | Reuse is supported within the recorded envelope. |

Keep design tier, implementation status and qualification status separate.
Required unsupported fields or unresolved source conflicts prevent an implemented
claim. A Go report must remain a prediction and may not set `powerpoint_verified`.

### Numeric and visual criteria

- Intended point geometry survives XML serialization within 0.01 pt. Native frames
  match declared geometry within 0.15 pt. Tolerances cover measurement precision;
  they are not design padding or permission to erase overflow.
- Native visible text and run styles match resolved content exactly. Line counts
  and wrap boundaries match the Go prediction in qualification controls. Go/native
  baseline or occupied-bound differences over 0.5 pt require investigation or an
  explicit conservative-review limitation; they cannot be called exact parity.
- Native occupied content must fit its safe frame. A positive overflow up to
  0.15 pt is numerically indeterminate and requires visual review plus a recorded
  result; anything larger fails. A verified fit envelope includes interior ink
  where relevant, not only character advances that exclude bullets/overhangs.
- All chrome, source lines, labels, chart data and table values are present.
  Geometry, assets/crops, layer ordering, required groups/attachments and editable
  package parts match the declared representation. No full-slide raster replacement.
- Native visual review accepts hierarchy, alignment, dot fields, marks, crop,
  readable density, and absence of clipping/collisions. Every material deviation
  from the source reference is named; reviewed exceptions stay tied to exact input.

These are initial engineering thresholds and must be recorded with the profile.
Changing them changes the qualification contract and cannot retroactively promote
historical evidence.

## 7. Implementation sequence and remaining experiments

The five semantic ambiguities now have explicit decisions. Remaining uncertainty
is implementation feasibility/measurement, with clear completion conditions:

1. Load the frozen source + these policies; diagnose unknown fields and conflicts.
2. Provision static Sans and resolve real 400/500/600/700 faces; emit every one of
   the 14 type styles without substituting fonts or weights.
3. Implement exact leading/tracking/kerning/defaults and Go range geometry. Native
   typography controls must cover all styles, mixed runs, wrapping boundaries,
   footnotes and paragraph ends before qualifying the new profile.
4. Resolve grid/frame/chrome and native groups/connectors; inspect movement behavior
   for semantic attachments before advertising it.
5. Build the representative source deck and inspect native output. Add bindings and
   measured count/flow envelopes to those variants, then extend by dependencies.

No new font installation, source-repository edit, engine change, deck build,
test suite or native capture was performed while tightening these contracts.


The additive [metric-card contract](metrics.md) implements primary/comparison,
change/status and one/two-secondary grouped cards in the opt-in v2 engine.

The [full-library completion plan](library-completion.md) and
[97-template feature ledger](template-coverage.json) assign the remaining work to
parallel capability modules and template adapters. They are planning artifacts;
inventory does not confer renderer support or native qualification.
