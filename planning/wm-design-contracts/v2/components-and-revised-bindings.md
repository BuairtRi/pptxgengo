# Slice 2 contracts and native review

Source: `wmds-library.v2`, round-4 commit
`7bdcaee5030a12275a1f881a8542f4d302d207df`.

## Checkbox cells

Table column type `checkbox` binds a required boolean. False draws an empty
12 pt box; true adds a tick in the same strong ink. The border is inset 0.5 pt,
11 pt square, with a 1 pt stroke. Both states are centered. Tables/text remain
native; adornments are editable shapes. Raw source blank/null cells draw empty
boxes; bound content rejects missing/null booleans. The existing `check` cell
keeps its filled-square/tick versus dash behavior.

Each runbook exposes four checkbox slots plus row text and stable row keys.
New-ledger cells permit explicit empty strings for completing a walk-through.

## Quadrants

`key` is a closed absent/false, true, or `"markers"` union. False draws dots and
labels. True draws numbered points and an internal numbered legend. Markers mode
draws numbered points without labels or an internal legend. The numbered-legend
template supplies a six-item external list; point/list order preserves ordinal
correspondence. Caller array keys supply stable native names.

V2 exposes point x/y as finite number slots; the renderer enforces [0,1]. Diagram
placement remains source structure. The source square is `min(h-36,w-36)`, or
`min(h-36,w-300)` with an internal legend. Plot origin is x+24,y and legend gap
is 36 pt. V1 keeps its rectangular plot, legend reservation and 9 pt top inset.
V2 quadrants omit that inset and report actual drawn bounds, excluding unused
allocated padding. The authored split quadrant at y27 permits a specific 9 pt
border bleed above tall-column y36; footer/source reservation still applies.
Qualitative points remain native shapes. Numerical data mode still rejects
numbered keys; there is no new numerical scatter/workbook claim.

## Annotations

Frame/container IDs and annotation target, placement and arrow are fixed
structure, excluded from caller text slots. Explicit IDs survive compilation,
binding records and native groups. Targets are measured before paint, supporting
forward references. Frames/containers resolve to the actual drawn rectangle,
excluding labels above the border. Unknown/duplicate IDs fail explicitly.

The closed arrow set is registered `arrow-right-angle` and `arrow-double`.
Right-angle supports below-left/right, above-left/right, left and right. Double
supports left/right. Canonical path extrema determine orientation and scale;
the tip sits 4 pt outside the target. Above placements clear any label above
the border; side placements remain centered on the border. Size controls span, default
66 pt. SVG pictures retain a local PNG fallback and canonical asset hash.

Callout label/text are bound content; width/surface remain source structure.
Text is measured with 12 pt padding, 3 pt internal gap and a 6 pt tail gap.
Rich text/footnotes use the existing planner. Text and any marks are emitted at
the resolved position together. Callout overflow fails without shrinking;
callout boxes must fit the zone and arrow pictures must remain on the canvas.

## Revised bindings and explicit amendment

All 15 revised designs have revision-specific slots, identities and hash
provenance. V2 stats/four-metrics requires four value/label pairs, an explanation
label and two paragraphs. V2 takeaway-rail/metrics-rail requires two value/label
pairs, an explanation label/paragraph and rail eyebrow/heading/body. Exact closed
slots/keys are exposed by library-catalog. These use the existing library binding
contract. Earlier typed stats/rail structs stay on v1; typed cards/3 and cards/4
remain available in both revisions. Source example copy never fills absent input.

Revised comparison binds handover metric fields and drops the previous callout;
goal-steps drops the expected-outcome slot. Changed ordinals appear in source
pointers/native identities. Old agenda/pillar amendments remain on v1 because v2
authors the photo/arrows directly. Other existing amendments were inspected and
do not overlap other revised keys.

V2 stats/circled-headline has one explicit amendment: explanation card height
grows from 90 to 108 pt, ending at reserved body bottom 432. Authored text needed
6 pt more internal capacity; the next 18 pt allocation preserves typography.
Reports record `wmds.circled-headline-explanation-capacity.v2`.

The user review identified overlapping photo/title allocations in
`divider/inverse`. The latest user revision of
`wmds.inverse-divider-separated-columns.v2` keeps the 270 pt photo, places it at
x507, and places the whiteboard at x561 with width342 (one 18 pt step narrower).
Its original y54/height396 remain. The photo extends three grid steps (54 pt) to
the left of the lattice. The text column is 432 pt wide, ending 18 pt before the
photo, with the eyebrow at y216 and title at y234 (36 pt higher). Three display
lines at 60 pt leading end at y414, within the reserved body. The current copy
wraps to two lines: “Who you will” / “work with”. V1 and pinned source stay intact.

`thumbnail` also supports `kind:"blank"`: a native white page/border without
schematic bars. It is a composition surface, with the existing caption/stack
behavior. Existing placeholder kinds remain schematic. The annotated reference
specimens use a blank thumbnail plus native text and row fills to illustrate
readiness, close-step targets, owners and dates. Those illustrative details are
explicit specimen content, never fallback content in bound templates. Their
frames grow to 72/90/90 pt to enclose the complete sections. Reapply with
`scripts/refine-wmds-slice2-reference.py`; arbitrary caller previews remain
explicit composition work.

## Reference review

`samples/wmds-refresh-slice2-20261002/final/` contains the editable 46-slide
WMDS-revised-components-reference.pptx and its native PowerPoint PDF. Slides
1–38 pair 19 source/alternate-content examples. Slides 39–44 cover six right-angle
placements, 45 covers a left double arrow, and 46 covers an internal quadrant
legend. Source annotated examples cover right double arrows. Alternates change
metrics, explanation copy, point coordinates, array identities and checkbox
states; several revised-layout pairs vary only their header and keys.

Every page was visually reviewed through PowerPoint's local printing PDF output.
User feedback subsequently corrected slides 7–8 and 31–32. Those four pages were
re-exported and reviewed; the other 42 slide XML documents are unchanged. The
current reference supersedes the earlier review images for those four slides.
Compilation succeeded. No tests were added or run. Unknown-target and invalid
union branches are implemented and code-reviewed; negative controls were not
executed. [Review receipt](reference-review/slice2.json) records artifact hashes,
native structure and the limited specimen scope. Generator qualification flags
remain unchanged. This is not an arbitrary-content or full-catalog qualification.

Next: family-wide implementation/review of additional variants, integrated
catalog review and release packaging.
