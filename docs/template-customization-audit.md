# Template customization audit

Audit date: 2026-10-08. Source basis: `1bee6ca58abc9ce348e3d1f4e53f0b4ac1be87ee`,
bundle `wmds-library.v11`. This is a capability assessment and implementation
backlog, **not operator documentation for commands that already exist**.

## Intent and evidence

A catalog template is a visual starting example. Its example people, layer count,
phase count, branches and coordinates are not the required structure of a new
project. A team may omit a program manager, place an executive sponsor outside
the delivery pods, and have a different reporting graph. A layered architecture
may have three or six layers. A composed architecture may have a completely
different topology while retaining its visual language and framing device.

Content-complete reusable slides are a separate library and remain deferred.
This audit does not turn illustrative template copy into approved reusable facts.
The `runbooks` catalog family contains operational runbook **slide templates**;
an agent's composition runbook is a different artifact.

The [machine-readable inventory](template-customization-inventory.json) covers
**all 649 V11 catalog variants in 29 families: 648 active and one deprecated**.
The deprecated `from-to/rows` record points to `transformation/before-after`;
agents should choose its replacement for new work. Each record includes
purpose, source location/hash, typed source records, stock collection counts,
composition-runbook assignments and variant review notes. Assignments combine
catalog intent with source structure: a primitive-built pod diagram still needs
the team composition runbook even if its AST contains no `pod` node.

Review depth is explicit:

- Every variant has an inventory mapping from its current purpose and AST.
- Priority families have renderer/CLI source review described below.
- Other families have an initial semantic assessment and require deeper source
  and composition exercises before qualification.
- This audit ran no new rendering, GUI editing or test suite. No variant is
  declared qualified merely because it maps to a runbook or renders natively.

The inventory does not enumerate the frame/rail Cartesian product. Frame and
footer variants in the catalog remain separate records because they change
available layout allocations. Catalog counts/slots are specimen budgets, not
proof that another content count fits.

## Existing surface and its limits

Current supported project routes are in `cmd/pptxdesign`, exposed through
`pptxgengo design`. Historical `pptxcompose`, adaptive and dynamic-component
examples are not evidence that an equivalent supported project command exists.

| Existing capability | What it establishes | What remains |
| --- | --- | --- |
| `project edit`, explicit bindings and `--check-fit` | Change source copy/values and measure changed slides | Complete family composition transactions, relationship-aware add/remove/reorder, useful layout proposals |
| Local scene components and project-owned templates | Use current typed renderers; keep installed catalog immutable | Domain models that keep related components and identities consistent when topology changes |
| `project template detach` and `project diagram inspect/patch/arrange` | Local node additions/deletions, placement changes and align/distribute | These operations manipulate geometry; they do not define staffing, dates, decision rights or process meaning |
| Attached straight/horizontal/vertical connectors | Known block endpoints, calculated paths, native elbow guides | Obstacle routing, crossings/clearance, connector labels and broader endpoint classes |
| Logical containment | Explicit memberships, padding and allocation checks | General composition fit and exact ink/stroke bounds; good layout is not inferred from containment alone |
| Eligible native lists/cards/tables | Better human text/cell editing, with documented retention reasons | Native editability does not establish domain item insertion or semantic reconciliation |
| Receipt-backed native reconciliation | Known mapped text/transforms/order; supported whole local deletions and explicitly mapped block copies | Family semantic adoption, arbitrary inserted objects, partial deletions, reparenting and real GUI Save As qualification |

The geometry development work is on this main basis. The promoted v4.2.1 binary
does not include that newer geometry surface. Operator guidance must state the
required CLI version rather than presenting main-only capabilities as released.

An unedited macOS GUI Save As succeeded on 2026-10-08. The retained packet reported
42 geometry no-ops and 10 text no-ops, plus 63 manual package/format findings
(including relationship ID rewrites and text/default serialization changes).
This establishes saved-file identity/geometry retention evidence only. It does
not qualify deliberate GUI manipulation, adoption or rebuilt visual comparison.
Office serialization normalization still needs classification fixtures/tests
while preserving unknown changes. Evidence is retained privately under
`~/Documents/pptxgengo-qualification/gui-save-as-20261008/attempt-01` in
`receipt.json`, `gui-phase-evidence.json`, `reconcile-save-as.stdout` and
`save-as-review/`. The source was the horizontal elbow fixture build
`build-20261008T155717-f4ccfb3cc5507957`; exact files/hashes are pinned in the receipt.

Generic known-object reconciliation evidence should be inherited as a foundation,
not reported as a passed team/Gantt/maturity customization demo. A moved Gantt
bar does not necessarily mean its task date changed. A role moved next to another
pod does not necessarily mean reporting or membership changed. Preserve the native
edit and request explicit semantic interpretation where needed.

## Composition runbooks, with template-specific exceptions

Use reusable runbooks for composition types, with a short exception record for
each affected template variant. There is no need for 649 copies of the same
instructions. A slide can require more than one runbook: site rollout needs Gantt
semantics and transition-state colors; an org chart beside a role table needs
team topology and table consistency.

Each runbook should specify:

1. **Communication job:** why this visual form is appropriate and when to choose
   another one; which visual relationships carry actual meaning.
2. **Authored entities:** stable IDs, optional entities, count freedoms,
   relationships, source facts and classification/color meanings.
3. **Composition decisions:** supported topology/layout patterns, explicit
   geometry overrides, reserved frame zones and label rules.
4. **Operations:** inspect, add/remove, reorder/reassign, change relationships,
   measure, propose layout and apply one guarded source transaction.
5. **Fit choices:** show the actual capacity issue and alternatives such as
   wider allocation, another arrangement or another slide; do not silently shrink
   text, delete content or change meaning.
6. **Native editing:** meaningful selection/grouping behavior and the exact
   adoption boundary; ambiguous semantic changes stay visible for review.
7. **Worked examples:** fewer/more elements, longer labels, different frame,
   deliberately unsuitable input, source rebuild and a real PowerPoint round trip.

Per-template exceptions include fixed visual motifs, coupled tables/legends,
special source evidence, supported frame zones, and places where the stock
example is encoded as loose shapes instead of a domain renderer. The inventory
maps every variant to these review assignments; it does not yet certify a finished
runbook or implementation.

## Deep review: priority composition types

### Team topology, pods and governance

Relevant family: `team` (21 variants), plus role matrices and handoff diagrams
elsewhere. Keep people, roles, organizations, pods, reporting, collaboration and
escalation as separate concepts. A pod is not necessarily an org-chart branch.

Current source in `internal/wmdesign/scene_people.go`:

- `pod` accepts a title/band and `roles: string[]`. It uses 42 pt rows and a
  height of `48 + 42 × role count`; adding a role changes the envelope but does
  not automatically reflow neighboring pods or a client-role row.
- `orgchart` accepts a variable recursive root/children tree with optional keys
  and dotted reporting. Nodes are 162 × 54 pt with fixed sibling/depth spacing;
  a wide tree produces a capacity error. Organization encoding is limited to
  `wm`, `client` and `tbd`.
- `governance` accepts variable tiers, cadence, members and decision bullets.
  It uses 90 pt tier spacing and fixed member/decision regions; extra tiers or
  longer member/decision copy need an explicit fit resolution. Current shading
  distinguishes the first two tiers, not a general arbitrary-tier palette.
- Several working pod/org variants are assembled from blocks/connectors rather
  than these semantic renderers. A command must handle that mapping deliberately
  or produce a reviewed local derivative; never pretend every variant has the
  same internal domain model.

Needed operations: create/remove pods or roles; assign roles/people to pods;
reparent reporting; specify dotted collaboration; position shared sponsor and
program leadership; vary organizations and legend; choose rows/columns/tree/
paired leadership layout; measure every label; synchronize responsibility tables.
Stable role identity must survive reordering. Pod role strings currently need
external stable array keys or a richer keyed model.

First acceptance cases: two unequal pods with no program manager; four pods
with one sponsor outside them; a sponsor on the side rather than above; a dotted
client liaison; reordered roles; long two-line labels; removal of a leader with
explicit treatment of dependent relationships; governance with two and four tiers.
Native movement is a layout override unless an operator explicitly confirms a
membership/reporting change.

### Layered versus freely composed architecture

Relevant family: `architecture` (29 variants), with architecture patterns in
software, solution and modernization too.

Treat at least three composition types separately:

- **Layer stack:** variable ordered layers, consistent shared edges, cross-cutting
  controls and a foundation. `layerrow` and `plane` can render individual pieces,
  but a general layer-list layout/palette operation is not exposed. Specify
  ordered tones for 3/5/6 layers with sufficient text contrast; keep categorical
  system colors distinct from ordinal layer tones. Adding a layer must update
  numbering, control span, foundation and relevant connectors together. Do not
  simply repeat the lightest tint until differences become invisible.
- **Composed graph:** stock application/landing-zone/product/reference nodes are
  examples of kinds of things to show. Permit different counts, boundaries,
  hierarchy and topology using shared component styling and project-owned layout.
  Keep semantic relationships independent of point-path connector coordinates.
- **Transition state:** stable component identity across current/interim/target
  views, sites, cohorts and waves; preserve state-color meaning, traffic-share
  meaning and comparisons. A filmstrip intentionally uses consistent layout
  across states, but its component count and state transitions remain variable.

The current diagram commands already provide significant local geometry control.
The next gap is composition-aware layer and graph assistance, then broader
routing. Existing stock `connector` points are not automatically attached edges.

Acceptance cases: three and six layers; one cross-cutting control; changed
foundation; graph with added/removed services and rerouted relationships; nested
container move; failed obstacle route retained as a proposal; same components
across a different number of transition states. See
[geometry scope](geometry-editing-scope.md) for remaining routing/import work.

### Gantt schedules and phase gates

Instances are spread across architecture, approach, roadmaps and offers. The
inventory finds the `gantt` source components independently of catalog family.

`internal/wmdesign/scene_sequences.go` already models variable groups, lanes,
items, phases, gates and a today marker. Items have `from`/`to` or `at`, progress,
soft starts/ends, milestones, event kinds and label-side choices. Period labels
and sublabels define the timeline; positions are numeric period coordinates,
not a date/calendar engine. Layout packs overlapping intervals into tracks,
reserves native label width and reports bounded label/lane/legend failures.
Track pitch is constrained to 24–60 pt.

Needed operations: add/remove/reorder lanes/tasks; shift/resize an interval;
add/remove phases and gates; position a gate by period/date or an explicit task
relationship; change today; choose granularity; repack and preview the affected
timeline, phase bands, label positions and legend. Date support needs an explicit
calendar/time-zone/business-day contract before dates become accepted source.
Do not infer dependencies from temporal proximity: the current item model has
no dependency graph or critical-path scheduling contract.

Acceptance cases: 3 versus 8 lanes; overlapping tasks; a gate at a fractional
period; start/end milestones; changed phase boundaries; longer label near the
right edge; schedule exceeding one slide; native moved bar classified as geometry
until a reviewed semantic date/interval choice is made.

### Roads, forks and portfolio roadmaps

Relevant family: `roadmaps` (36 variants). Distinguish a metaphorical journey,
a decision fork, parallel tracks, a dated timeline and an ordinal portfolio.

- `road` supports variable milestones, normalized positions, current/active
  state, side labels, direction, road width, amplitude and wave count.
- `roadfork` supports variable trunk milestones and branch milestone lists,
  fork position, decision/parallel mode, chosen branch and per-milestone
  “here” markers. These are current source-model capabilities, not dedicated
  project mutation commands. Branch count/label fit constraints still need an
  exercised operator workflow.
- Now/Next/Later variants are predominantly explicit blocks/tables/cards. Their
  horizon, initiative, owner and dependency semantics are not uniformly encoded
  in one roadmap model. Native cells or coordinate arrows do not supply that model.

Needed operations: insert/remove steps before/after fork independently; add/remove
branch; set current step or chosen option; keep current/chosen distinct; change
horizons; move an initiative between horizons; maintain owners/dependencies and
confidence/status encoding; coordinate related criteria/detail cards. “Later”
must not silently acquire a promised delivery date.

Acceptance cases: one and four trunk steps; different milestone counts on
branches; two versus three options; fork relocated; current marker before and
after fork; changed chosen branch; five rather than three portfolio horizons;
initiative transfer with incoming/outgoing dependencies retained.

### Process flows, sequences and cycles

These cross solution, approach, argument, lifecycle and operational-runbook
slides. A semantic process graph requires steps, decision outcomes, actors,
joins, loops and explicit start/end/current markers. Chevron order alone is not
a branching process model.

Current `swimlane` supports lane strings, a column grid, keyed steps with IDs,
process/decision/start/end kinds and links by step ID. It uses 72 pt lanes and
fixed-size step allocations. Source links need explicit routing review when a
lane/column changes. Current `cycle` supports variable items, loop-back indices,
active stage, node dimensions, ring/closed/start choices and a center message.
Cycles and road forks are separate structures, not universal process graphs.

Needed operations: insert/remove/reorder a step; move it to another actor lane;
add a decision with named outcomes; join paths; choose start/current/end markers;
add/remove loop-backs; calculate attached routes and validate missing/duplicate
references. Use stable IDs instead of array indices for long-lived relationships
or rewrite all dependent indices atomically. Reordering a cycle must update
loop-back endpoints and active state.

Acceptance cases: two versus five pre-decision steps; branches with unequal
lengths; joining after fork; cross-lane handoff; current-step marker; moved
decision; deleted referenced step with explicit incident-link treatment;
cycle with three and six items and retained loop-back meaning.

### Maturity curves and staffing/team curves

Relevant families: `maturity` (19 variants) and `team-curves` (9 variants).
Several maturity variants are assessment tables/coaching plans without a curve;
their runbooks also need axis consistency and evidence rules.

`scene_intake_maturity.go` already provides a variable stage list, normalized
stage positions, exponential `shape`, headroom, label width, active/inflection
stage, optional branch and a current-stage label. The curve is a native custom
path with editable labels/markers; changing its path in PowerPoint is not
currently a qualified source adoption of `shape` or stage semantics.

`scene_intake_curves.go` already provides party series/values, point positions,
scale, curve/smoothing/tension options, phases and label anchors. Curves can be
lines or areas and can encode different quantities. Phase descriptions rendered
as separate text need coordinated updates. Before/after variants must retain a
shared scale and units; an attractive rescaling can otherwise mislead.

Needed operations: add/remove/reorder stages or points; set a point's source
value/position; select valid curve intensity/shape; set current/target/inflection;
change series, scale and units; align phase boundaries/labels; preview label fit
and branch clearance. Runbooks must distinguish an illustrative maturity shape
from measured scores, and staffing/FTE from proportions, capacity or activity.

Acceptance cases: 3/5/6 maturity stages; flatter/steeper shape; current and target
markers; a branch after a different stage; unequal staffing point spacing;
added third party; additional handoff phase; shared-scale before/after;
PowerPoint path edits retained as unresolved when source semantics cannot be inferred.

## All-family assessment

“Inventory” means mapped purpose and actual source nodes; it is not a passed
customization exercise. Priority renderer review is detailed above. Every
remaining family requires representative composition and desktop qualification.

| Family | Variants | Semantic intent / variable structure | Runbook or operations gap |
| --- | ---: | --- | --- |
| covers | 28 | Deck entry, section/navigation and closing; contacts, sections, progress and photo framing vary | Navigation consistency and variable agenda/contact count; inventory |
| understanding | 19 | Ask/objectives/scope/traceability; claims, drivers and evidence links vary | Preserve objective-measure and requirement-response identities; inventory |
| core | 53 | Everyday message, card, quote and narrative layouts; variable item count and copy | Content-aware column/card reflow, quote/source coupling; inventory |
| argument | 43 | Reasons, options, capabilities and comparisons; variable branches/grouping | Classify graph versus comparison versus narrative; maintain premise/conclusion and option mappings; inventory |
| evidence | 15 | Metrics, data and source explanation; variable categories/series | Units/base/source and chart/table synchronization; inventory |
| diagrams | 13 | Funnels and pyramids with variable stages/bands | Ordered versus weighted levels, stage insertion and label capacity; inventory |
| heatmaps | 37 | Assessment dimensions and comparison tables; variable axes/rows | Stable axes, explicit score domains, shared scale/legend, missing-versus-zero meaning; inventory |
| status | 7 | Delivery state, risks/issues/dependencies and next actions | Structured owner/date/status updates, no arbitrary color-only status; inventory |
| venn | 12 | Sets and intersections; set membership and region meaning vary | Supported set counts, intersection identity, label/region fit; inventory |
| architecture | 29 | Layer stacks, composed graphs and transition views | Layer count/palette, arbitrary project topology, containment/routing and state identity; source review |
| solution | 15 | End-to-end flow, capabilities and service mappings | Variable process steps/lanes/relationships and coupled narrative; source review for composition primitives |
| lifecycle | 17 | Cycles, subsets, loop-backs, phases and role/health matrices | Variable cycle steps, explicit loop references, current stage and coupled tables; source review |
| software | 7 | System/module/platform structures and maps | Semantic grouping/topology and variable category count behind primitive layouts; inventory |
| modernization | 20 | Technology transition, paths and capability comparisons | Stable current/target identity, alternative path semantics and migration choices; inventory |
| maturity | 19 | Curve stages, capability assessment, current/target and coaching | Stage/intensity controls, branch and marker semantics, coordinated tables; source review |
| decisions | 12 | Choices, evidence, decision ownership and next actions | Option/criterion identity, decision state and responsible person coupling; inventory |
| approach | 16 | Phase sequence, Gantt, gates and handoff | Variable phases/tasks/gates, duration/proportion meaning, evidence/owner links; source review |
| roadmaps | 36 | Roads/forks, timelines and horizon portfolios | Variable before/after-fork counts, current/chosen markers, horizons and dependencies; source review |
| team-curves | 9 | Staffing/capacity handoff across phases | Point count, values, party/phase count, intensity and shared scale; source review |
| interviews | 29 | Research scope, schedules, quotes, findings and synthesis | Respondent/session identity, scheduling and traceability with variable item counts; inventory |
| workshops | 29 | Agenda, facilitation, activities and outputs | Variable agenda blocks, durations, audiences, action ownership and continuation; inventory |
| runbooks | 32 | Executable steps, gates, escalation, rollback and qualification | Variable step/decision graphs, actor/time/check references, multi-slide consistency; source review for composition primitives |
| change | 22 | Stakeholders, readiness, communications and adoption | Audience/channel/wave variables and comparable scores/status/ownership; inventory |
| commercials | 33 | Fees, work/phase economics, assumptions and allocations | Numeric source/calculation contracts, variable rows/phases and units; inventory |
| offers | 17 | Offering scope, work packages, examples and plan | Variable package/deliverable/task structure; retain claims/evidence boundaries; inventory |
| value | 22 | Value drivers, outcomes, measures and causal structure | Source-backed assumptions/formulas, variable graph/series and consistent units; inventory |
| team | 21 | People/role topology, pods, bios and governance | Variable roster/pods/reporting/sponsor placement and decision rights; source review |
| about | 11 | Firm profile, credentials, logos and geographic presence | Asset licensing/availability, variable locations/logos and sourced facts; inventory |
| proof | 26 | Case evidence, examples, results and reference material | Variable case structure/media/metrics with traceability and no invented proof; inventory |

## Proposed implementation order

1. **PowerPoint access and actual Save As qualification.** Establish repeatable
   application lifecycle diagnostics and real saved-file evidence. This is a
   separate workstream, not a requirement that authoring operations invoke Office.
2. **Broader routing on the shared geometry foundation.** Obstacle/label
   avoidance and clearance/crossing diagnostics, explicit manual-route retention,
   then labels/additional endpoints. A failed route must retain a useful preview.
3. **Teams/pods/governance and Gantt composition.** First complete domain mutation
   contracts/runbooks. Reuse guarded project transactions and measured previews;
   choose actual command names only after inspecting source/model needs.
4. **Layer stacks, road forks/processes, maturity and staffing curves.** Build
   on existing renderers, adding variable structure and specialized guidance.
5. **Remaining family exercises.** Use the full inventory to select representative
   composition types and record per-variant exceptions instead of rushing through
   a visual-only catalog sweep.

Do not create a separate CLI for each specimen. Favor a small set of domain
operations plus a strict declarative patch for batches, useful inspection of
accepted fields/IDs, deterministic layout previews, and one atomic apply. The
semantic model and relationships must remain distinct from geometry overrides.
Designing those contracts is the next implementation step; this document does
not promise names such as `team add` or `gantt gate` are available today.

## Completion evidence for each composition type

Close a type only with source manipulation, measured fit, visible native editing
and persistent reconciliation:

- Representative catalog example plus different element counts/topology and
  long labels; preserve its visual language and native profile.
- Inspectable semantic entities and relationships, stable IDs after reorder,
  explicit frame allocation, and a retained failed-layout proposal.
- Add/remove/move/reassign/relationship changes applied as one guarded transaction;
  removal handles every dependent reference and supporting table/legend.
- PowerPoint selection/text editing, actual Save/Save As, and reconciliation of
  supported source/geometry changes; ambiguous domain meaning remains explicit.
- Rebuild, numbered source-and-deck snapshot, and ZIP/share extraction preserve
  semantics, layout, pins and asset references.
- macOS and Windows evidence recorded separately. XML fixtures and unit coverage
  are useful evidence but do not substitute for real Office serialization.

Use nightly/on-demand CI for broad catalog/desktop coverage; keep normal PR/main
checks light as already agreed. All CI and artifacts stay in private GitLab.

## Source references

- Catalog and specimen sources: `library/wm-design-system/v11/source/templates/catalog.json`
  and `templates/library/*.json`.
- Current local scene entry: `internal/wmdesign/scene_composition.go`.
- Team/governance: `internal/wmdesign/scene_people.go`.
- Gantt/swimlanes/sequences: `internal/wmdesign/scene_sequences.go`.
- Layer rows and coordinate connectors: `internal/wmdesign/scene_diagrams.go`.
- Maturity: `internal/wmdesign/scene_intake_maturity.go`.
- Staffing curves: `internal/wmdesign/scene_intake_curves.go`.
- Roads/forks: `internal/wmdesign/scene_intake_road.go`, `scene_roadfork.go`.
- Cycles: `internal/wmdesign/scene_intake_cycle.go`.
- Project command boundary: `cmd/pptxdesign/project_edit.go`, `project_diagram.go`.
- Current operator references: [architecture/geometry](../skills/west-monroe-presentations/references/architecture-geometry.md),
  [source format](../skills/west-monroe-presentations/references/source-format.md),
  [custom slide design](../skills/west-monroe-presentations/references/custom-slide-design.md).

## Inventory maintenance

The JSON inventory is a static audit snapshot. Re-inventory when the catalog or
source hashes change, retain review assignments deliberately, and update the
basis commit. Its template hash is SHA-256 over sorted-key compact JSON for the
complete specimen; family and catalog hashes use exact file bytes. Keys are
`id/variant` and must match the catalog one for one.

Source type counts include nested typed values such as table cell indicators;
they are not counts of native objects or a list of guaranteed project node kinds.
Stock collection counts describe example source, never supported maximums.
`explicit_keys` counts inline `key`/`id` values only; zero does not mean the
renderer cannot receive stable keys through its separate key bindings. Frame
exceptions identify additional review allocation, not a claim that a variant
currently fails. Composition profiles are many-to-many semantic assignments,
not additional templates or completed feature implementations.
