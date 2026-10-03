# WMDS library refresh — 2026-10-02

## Inspected source

Live design repository: `~/Projects/wm-design-system`, clean commit
`7bdcaee5030a12275a1f881a8542f4d302d207df` (owner review round 4).
Baseline: the actual frozen files in `library/wm-design-system/v1/source`, rather
than an assumed Git base. The source repository also provides explicit migration
notes in `templates/CHANGELOG.md`, `templates/changes.json` and
`templates/change-notes.json`.

[Machine-readable inventory](source-update-2026-10-02.json) contains every key,
old/new template and family-file hashes, changed fields, design notes and feature
dependencies. Reproduce with `python3 scripts/inventory-wmds-update.py`.
The initial inventory was read-only. Slices 1–3 are now implemented; see
[slice 1 review](slice1-completion.md), [slice 2 contracts/review](components-and-revised-bindings.md)
and [added-family implementation/review](slice3-completion.md).
Full-catalog regression review and integrated packaging remain planned. No tests were added
or run.

## Exact delta

| Family | Frozen | Current | Added |
|---|---:|---:|---:|
| Approach | 11 | 27 | 16 |
| Argument | 18 | 26 | 8 |
| Commercials | 4 | 7 | 3 |
| Evidence | 10 | 18 | 8 |
| Openers | 18 | 22 | 4 |
| Proof | 13 | 25 | 12 |
| Solution | 12 | 21 | 9 |
| Team | 11 | 21 | 10 |
| **Total** | **97** | **167** | **70** |

The old 97 comprise 81 unchanged entries, 15 revised designs and one newly
deprecated entry. The current catalog has 166 active designs plus the retained
deprecated `from-to/rows`, replaced by `transformation/before-after`.
No template was removed. Tokens and component definitions are byte-identical
to the frozen bundle; frame definitions add the split family. New semantics also
exist in the reference renderer and documentation, so unchanged component JSON
does not mean that the renderer needs no work.

### Fifteen revised designs

- Openers: `agenda/schedule`, `divider/inverse`, `key-message/stat`,
  `takeaway-rail/metrics-rail`.
- Argument: `comparison/us-versus-others`, `goal-steps/goal-band-three-steps`,
  `pillars/four-why-matters`, `session/stepper-and-guidance`.
- Evidence: `quadrant/subtle`, `quadrant/strong`, `stats/circled-headline`,
  `stats/four-metrics`.
- Solution: `architecture/layers`, `workstreams/five-with-risks`.
- Approach: `phase-detail/team-handoff`.

Significant additions include runbooks and cutover checklists, escalation table /
flow / ladder variants, operational status and RAID reporting, SIPOC and process
views, case-study and annotated-deliverable layouts, and split/nav variants across
the existing families. Full keys and source instructions are in the inventory.

## Shared changes before template batches

| Capability | Designs using it now | Existing Go state / required work |
|---|---:|---|
| Split frame | 19 | Implemented in slice 1: four column allocations, short-column chrome, separate tall-body geometry and zone validation. |
| Source nav binding | 23 | Implemented in slice 1: 2–6 caller labels, stable keys, active-index conversion and native tabs. |
| Three-line split title | 4 | Implemented in slice 1: authored narrow/wide rule positions and font styles. |
| Table checkbox | 2 | Implemented in slice 2: empty/checked outlined boxes and required booleans. |
| Quadrant numbered markers without key | 1 | Implemented in slice 2: closed false / true / `"markers"` union and square v2 geometry. |
| Named-target annotation | 1 | Implemented in slice 2: preserved IDs, measured targets/callouts, canonical arrows and explicit ID errors. |

V2 quadrant geometry follows the square source algorithm and reviewed coordinates.
The previous rectangular plot remains on v1.

Existing renderers already handle ordered lists, small numbered lists, connector
labels/start dots/two-way arrows, rotated marks and compound RACI values such as
`A/R`. These still need inspection in their new layouts but are not new renderer
modules. Table row values such as `Risk`, `Issue` and `Dependency` are data; do not
miscount their `type` column as a new scene-node discriminator.

## Contract decisions for the migration

1. **Version the source revision.** Keep the frozen v1 specimen reproducible and
   add an explicitly selected refreshed bundle/adapter revision. Do not replace
   files under the v1 hash pin and then alter the constant to hide drift. The new
   source revision can reuse the existing Go font engine and unchanged font
   assets; bundle version and font-engine version are separate concerns.
2. **Consume lifecycle metadata.** Extend strict template decoding for `rev`,
   `added`, `revised`, `status` and `replacedBy`. Preserve these in catalog output.
   Hide deprecated entries from normal new-slide selection but retain explicit
   access for old content and expose the suggested replacement.
3. **Represent both split zones.** Keep header/short-body and tall-body rectangles
   distinct, with tall top at 36. Header title/rule, stamp, source and default dots
   follow the short column. Footer/legal geometry stays at full content width.
   A narrow short column uses heading 24/30 with rule 108/126/162; a wide short
   column uses title 32/36 with rule 108/144/180. Support either footer and nav,
   reject panel rails with split, and reserve declared source lines explicitly.
4. **Convert nav semantics deliberately.** Source `active` is a zero-based index;
   native request `Active` is a stable tab ID. Use caller keys for identity and
   preserve source label/selection content. Match the source's footer-based tab
   height; its calculation is independent of source-line reservation. No default
   fictional labels in bound mode.
5. **Refresh content contracts, not just examples.** The revised stats and rail
   templates add explanation labels/paragraphs; their old typed APIs cannot express
   the new anatomy. Keep old APIs on the old revision and declare new required
   fields on the new revision. The revised comparison drops the callout slot for
   handover metric fields; the goal-steps design removes the expected-outcome slot.
   Changed node ordinals require an explicit identity/binding migration.
6. **Keep references structural.** Annotation `target` and frame `id` are fixed
   structural identities, not freely editable text slots. Resolve target bounds
   from named shapes; measure the supplied callout without guessing a target from
   position. Missing targets fail. Registered arrow assets remain editable picture
   objects; slides are not flattened.
7. **Use final JSON and renderer semantics for geometry.** Change-note prose
   includes several earlier review-round values. For example, checklist notes
   mention 18 pt overlays, while the final checkbox cell and current renderer use
   a 12 pt outlined native cell adornment. Implement the final typed cell rather
   than an obsolete overlay workaround. Preserve all rounds as provenance.
8. **Scope old refinements by revision.** Existing `agenda/schedule` refinement
   appends a photo; the new source already contains an offset block and photo.
   Running the old refinement duplicates it. The pillar refinement expects
   connector nodes 17–20; the new source contains hand-drawn marks and changed
   card geometry. Retire these two refinements for the new revision. Audit every
   remaining amendment, including unchanged entries, before carrying it forward.

## Implementation sequence

### Slice 1 — source revision, split frames and navigation

Completed 2026-10-02. The 18-slide native reference was reviewed; this does not
qualify the remaining new or revised templates. See the
[frame and navigation contract](frames-and-navigation.md) and
[review receipt](reference-review/slice1.json).

Freeze the refreshed source and loader/catalog metadata; implement split geometry,
title allocation, source/nav compilation and caller nav content. Update chrome to
use header coordinates for header content and full-width footer coordinates for
footer content. Bind representative `agenda/schedule-split`,
`key-message/stat-split`, `team/roster-split`, `case-study/exhibit-split` and
`bio-full/portrait-nav` specimens. Include all four split directions/widths,
three-line titles, both footers and split + nav in a compact frame reference.

Completion: exact source revision reported, old revision still selectable, every
frame mode generated with named zones and caller-supplied nav, and native renders
reviewed for title fit, header dots, source placement, legal line and tab alignment.

### Slice 2 — new semantics and revised contracts

Completed 2026-10-02. All 15 revised designs and four capability-dependent templates
have paired native specimens. The six initial capability blockers are removed.
See [contracts and review](components-and-revised-bindings.md) and the
[receipt](reference-review/slice2.json). Negative error branches were code-reviewed,
not executed as controls.

Implement checkbox cells, quadrant marker/square behavior and annotations. Migrate
all 15 revised bindings and revision-specific refinements. Build paired source and
changed-content specimens, including empty/checked boxes, external quadrant key,
new explanatory copy and unknown annotation targets. Keep the deprecated key
explicitly available while default catalog selection points to its replacement.

### Slice 3 — added template families

Completed 2026-10-02. All 70 additions have paired source/meaningful-content
specimens in the 140-slide native reference. See [completion](slice3-completion.md)
and [review receipt](reference-review/slice3.json).

With shared files integrated, family adapters can be developed independently:

- Approach/runbooks: 16 additions; depends on frames, nav and checkboxes.
- Proof: 12 additions; depends on frames, nav and named annotations.
- Team: 10 additions; depends on frames/nav and existing people/table renderers.
- Solution: 9 additions; depends on frames/nav and existing diagram renderers.
- Argument + openers: 12 additions; depends on frames/nav and revised contracts.
- Evidence + commercials: 11 additions; depends on frames/nav and quadrant modes.

Use the inventory to own disjoint family files, while one integrator owns the
loader, frame request, shared schemas and dispatch. Parallel delegation is an
execution choice, not a requirement to start agents during this inspection.

### Slice 4 — integrated native review and packaging

Generate the 167-design fixture catalog and meaningful bound-content examples.
Inspect all updated/additional layouts in native PowerPoint, including images,
marks, table headers, source lines and dense wrapping. Recheck unchanged designs
for shared-renderer regressions. Record qualification separately from inventory
and successful generation. Use focused typography evidence only if a newly used
text behavior exposes a discrepancy; a full font recalibration is not implied by
these source changes. Update catalog/CLI documentation and release packaging after
review. A rebuilt client deck is a later consumer of this refreshed library.

## Assessment

The design direction and geometry are well specified. The work is substantial
because there are 70 additional variants and changed binding contracts, but most
of the native rendering foundation is reusable. The main uncertainty is measured
fit and native visual review in narrow split columns and dense working slides.
The first slice should establish the frame/source contract before distributing
template implementation work.
