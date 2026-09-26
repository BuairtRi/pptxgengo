# Implementation waves for proposal fidelity

2026-09-26. Proposed execution sequence following the five-slide dense benchmark.
This expands the next fidelity gates in IMPLEMENTATION_PLAN.md. Future capabilities
and acceptance criteria below are planned, not implemented or approved.

## Outcome and working method

An agent should take a brief plus supporting material, find an appropriate WM
pattern, populate its semantic slots, compose editable objects, review native
PowerPoint output, and safely ingest supported colleague edits. The current dense
sample proves explicit composition; it does not yet prove this whole workflow.

Each pattern gets three fixtures:

1. A source-faithful control with original copy, fonts, artwork, geometry and layers.
2. A newly authored instance using the component contract and changed content.
3. Stress variants that change copy length, item count or supported style, including
   a deliberate unsupported case that must fail with an actionable explanation.

Raw source copying supplies a preservation control. Reuse approval requires the
changed-content fixtures. Save native renders, source/target crops, object and text
measurements, contract versions, source hashes and a QA error log for every fixture.
Do not claim pixel identity for intentionally changed-content variants.

## Wave 1 — Layout ownership, spacing and layering

Checkpoint: the first three dense patterns and three stress variants have passed
native verification and visual review. See [implementation and limits](WAVE1_CHECKPOINT.md).
Named fixture tokens are resolved by the authoring helper; runtime token inheritance
and broader validation/variant coverage remain future work.

**References:** EnableComp 3/4, UHG 24/28, current dense slides 1–3.

**Build:**

- Parent panels with relative child coordinates and four explicit padding values.
- Measured row/column tracks, gutters, shared row heights, aligned column edges,
  and named slots; initially support the exact three benchmark patterns.
- Row height from the tallest measured cell plus padding. Fixed slide-safe areas
  and explicit minimum/maximum tracks; no silent font reduction.
- Ordered layers with parent/child stacking and explicit decorative overlap.
  Fix the existing canvas-label/PhaseSpec background ordering defect.
- Semantic spacing/type/color tokens resolved into exact values in the manifest.
- Incremental text measurement keyed by the full text/style/width contract plus
  PowerPoint/font environment, with provenance. Keep final native verification.

**Deliver:** phase-column, phase-detail and workflow-matrix components. Rebuild the
three existing pages from these components without hand-positioning each cell.

**Gate:** original fixture remains visually stable; 4/5/6-row or phase variants,
longer copy and multiline headings either reflow inside declared limits or report
all failing zones. Moving a parent moves its children. Backgrounds cannot cover
labels. Changing one text contract cannot reuse its old measurement.

## Wave 2 — Rich typography and faithful visual components

Checkpoint: eight bounded examples now pass native verification and independent
visual review. See [the Wave 2 checkpoint](WAVE2_CHECKPOINT.md) for scope and proof.
Rich native bullets, native SVG pictures, exact picture-border joins and broader
source-slide identity remain open; the original full gate below is not yet closed.

**References:** UHG 24/28 deliverable panels, UHG 38 phase bars/extensions,
UHG 44 portraits, UHG 67 biography, EnableComp 5 response rows.

**Build:**

- Paragraphs and rich text runs: font family, size, weight, italics, color,
  paragraph spacing, bullet indentation and hanging indentation. Resolve inherited
  typography explicitly; report unavailable fonts rather than substitute silently.
- Run-aware native measurement and fit checks; preserve paragraph/run identity.
- Named stock compositions with local coordinates, text slots, asset slots and
  allowed transforms. Start with deliverable stacks, phase ribbons, milestone
  markers, hatched extension tails, portrait/name/role tiles and biography zones.
- Source geometry preservation for unsupported freeform artwork. Explicitly mark
  geometry as fixed, uniformly scalable or parametrically resizable; reject other
  transforms until proven. Do not convert an entire editable slide into an image.
- Image contain/crop modes, focal point, clipping frame and original asset hashes.
  Produce thumbnail artwork from pinned document/deck pages with captions and
  source links; keep surrounding text editable.

**Deliver:** refreshed phase-detail and roadmap pages, plus roster and biography
pages. Deliverable previews depict actual pinned sample content; illustrative
people/credentials and claims must be clearly labeled.

**Gate:** inline emphasis survives wrapping; bullet indentation matches the
reference; crop does not distort portraits; captions and thumbnails remain paired;
ribbon/tail geometry stays intact under every declared supported size. Native
rendered region comparisons cover all of these details.

## Wave 3 — Complex diagrams and anchored WM accents

**References:** UHG 14 architecture, UHG 36 two process paths, UHG 43 team,
UHG 5 underline/arrow and UHG 6 highlight; UHG 11/12 supporting hierarchy patterns.

**Build:**

- Nested containers with named edge/port anchors, sibling alignment and cross-cutting
  governance/security bands. Explicitly typed relationships: reporting, flow,
  dependency, advisory and annotation.
- Routes around measured text/image bounds, with minimum clearances, bend penalties,
  preferred directions and endpoint offsets. Keep visual tuning parameters explicit.
- Native connector attachment where verified; otherwise record editable-segment
  behavior and reject claims that moving a shape will reroute the connector.
- Phrase anchors identified by text range and occurrence, measured per wrapped line.
  Highlight uses the phrase region behind text; underline uses a calibrated baseline
  offset; arrows use declared source/target anchors and visible artwork bounds.
- Curated hand-drawn arrow variants with direction, visible endpoints, bend shape,
  aspect-ratio limits and clearance envelopes. Do not stretch every arrow into one
  generic routing path.
- A low-confidence/manual-placement state: stage the chosen artwork in a reserved
  review area with a precise target note, and mark the slide unfinished. Never
  silently accept a guessed position as a polished result.

**Deliver:** layered architecture, parallel process paths, richer team relationships,
and an accent stress page. Publish the first expanded visual review deck after
this wave, including the improved existing five pages and the new hard patterns.

**Gate:** routes avoid labels at normal viewing scale; parent moves preserve nested
alignment; tested arrows remain pleasant in short/long and alternate-direction
cases. Phrase emphasis follows moved text, changed widths, line wrapping and repeat
phrases without manual coordinate edits. Unsupported multiline/ambiguous placements
must surface clearly.

## Wave 4 — Library contracts, narrative and agent selection

**References:** all accepted patterns above; UHG 8 layered proof, EnableComp 5
needs/response and 27 deliverables/decisions for evidence structure.

**Build:**

- Promote proven instances into versioned component and layout contracts. Link
  slide pattern → slots/components → artwork → source evidence and style variants.
- Deduplicate by geometry, slot structure and behavior; represent equivalent colors
  as semantic style options. Retain source-specific differences and all provenance.
- Add previews, business purpose, content roles, cardinalities, measured fit envelope,
  approved transforms, preferred variants and known failures to SQLite retrieval.
- Agent-facing find/inspect/preview/instantiate operations through the Go CLI over
  the existing catalog. SQLite remains a rebuildable index; portable specs/assets/
  contracts remain the durable source of truth.
- Narrative specification: audience, slide role, one-sentence takeaway, assertion
  title, evidence references, qualifications, required detail, emphasis targets and
  suggested visual relationship. Draft this before selecting the visual pattern.
- Layout selection checks content structure/capacity and user preferences. Select
  another layout or split content when it cannot fit; do not trim evidence merely
  to fit a favorite design.
- Skill instructions for narrative, retrieval, composition, brand accents, native
  QA and revision. Keep deterministic geometry/fit in the CLI and editorial/design
  judgments in the agent with recorded reasons.

**Deliver:** an agent-generated proposal section from a fresh brief, with recorded
pattern/asset choices and evidence traceability. Include an evidence/quote page only
when authentic, attributed material is available; synthetic examples stay labeled.

**Gate:** a second agent can reproduce the supported workflow from the skill and
CLI without private implementation knowledge or hand-coded cell coordinates.
Search ranks suitable approved variants; forbidden transforms and unsupported
claims are rejected. Quality is assessed by argument/evidence structure and visual
review, not word count.

## Wave 5 — Colleague-edit reconciliation

**Starting point:** one real PowerPoint text edit already recovers into a new JSON
spec when original identity, geometry and media are retained. Styling is restored.

**Build:**

- Stable semantic IDs for slides, panels, cells, paragraphs and components, persisted
  through supported PowerPoint saves and tied to the build manifest.
- Three-way comparison of original spec, generated deck and returned deck. Report
  text, formatting, geometry, asset and structural changes separately.
- Import supported text/rich-text changes first. Next support defined image swaps
  and component moves/resizes with explicit contract checks and conflict handling.
- Preserve unsupported objects/edits as source scenes or explicit unresolved changes;
  never silently discard them. Reordering, deleted slides and duplicated IDs need
  dedicated reconciliation rules and fixtures.
- YAML serialization of the same versioned semantic model, with JSON/YAML equivalence
  checks. Serialization alone does not make arbitrary PPTX decomposition semantic.

**Deliver:** returned-deck diff, proposed recovered configuration and a rebuilt deck
for supported changes. Unsupported arbitrary PPTX remains on the scene-preservation
path with review requirements.

**Gate:** edit text, emphasis, image, position and slide order in native PowerPoint;
verify each supported change or a precise conflict. Rebuilt supported changes retain
untouched content/design. All new text/layout contracts are measured again.

## Wave 6 — End-to-end qualification and release

**Build/deliver:** a fresh approximately 12-slide proposal from research/brief inputs,
selected through the library and authored through the skill. Suggested pattern set:
five-phase approach, phase detail, workflow matrix, delivery team, roadmap,
architecture, parallel process paths, roster, biography, evidence/quotes,
needs/response and an accent-heavy argument slide. Exact narrative can change;
the difficult visual coverage must remain.

Run three variants: normal content, increased text/cardinality, and a returned deck
with real colleague-style edits. Reuse supported contracts; do not rescue failures
with unrecorded one-off coordinates or silent content cuts. Record build time,
measurement reuse, automatic failures, visual corrections and manual interventions.

**Release gate:** all slides native-rendered and individually reviewed; no unhandled
text overflow or accidental collision; approved typography/assets; consistent
margins and grid; correct evidence, dates and allocations; traceable recovery;
repeatable build from saved configuration. Report exact supported envelopes and
remaining manual steps. Library inventory size and pixel similarity alone are
insufficient release criteria.

## Parallel execution and ownership

- Start with three bounded lanes: Luna reference/slot audit; coding specialist
  layout/compiler work; Luna asset/geometry catalog. Root owns integration and
  acceptance. Aim for roughly 80% of delegated review/catalog tasks on Luna;
  use Sol/Terra for implementation and Astra only for a demonstrated hard design
  or reconciliation issue.
- Wave 2 asset extraction can run alongside Wave 1. Wave 3 anchor/arrow inventory
  can start early; diagram implementation waits for panels/layers. Wave 4 metadata
  and voice work runs alongside each wave, but promotion waits for proof. Wave 5
  identity requirements shape the schema from Wave 1; reconciliation follows stable
  rich-text/components. Wave 6 follows the required qualified capabilities.
- Keep native PowerPoint automation in one serialized queue. Parallelize file-only
  extraction, code and review. Use different output paths and file ownership for
  every agent task.
- Finish one representative slide at a time within each family: control, changed
  content, stress, fix, then promote. Retain every QA finding and correction.
- Deliver a small visible comparison at every wave, with the first expanded deck
  after Wave 3 and the independent end-to-end qualification after Wave 6.
