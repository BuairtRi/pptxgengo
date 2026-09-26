# Wave 2 roadmap contract — UHG slide 38

## Scope and provenance

Read-only audit of UHG 38. Source deck: `samples/UHG Fabric Platforming RFP Response - July 2026.pptx`; SHA-256 `b0f254ed7739768d0f345257689264393049d06cdd3849a177fc764f3d348d99`; part `ppt/slides/slide38.xml`. Reviewed preview: `samples/reconstruction/reference/png/slide-038.png`; SHA-256 `50f7e91855a235e9c443d54b93e446a752e812d9091d6511feb3994417567e33`. Candidate: `visual-candidate:uhg-roadmap-bars-milestones-extension-tails` in `library/component-contracts/visual-wave2-candidates.json`; existing phase seed: `component:slice-proposal-uhg-roadmap-phase-bar` in `library/component-seeds-expanded.json`.

## Observed source facts

### Month grid

Month grid is native editable table object ID 59 at slide path `3`, spanning x=0.503472, y=1.904402, w=12.322433, h=4.325312 in. It has 13 equal 866,741 EMU columns (0.947958 in), labeled Month 1–12 and Beyond. Header runs are 12 pt bold centered; Latin uses theme minor, East Asian and complex script specify Open Sans. Header fills: Months 1–9 tx1/dk1 (`#070154`), Months 10–12 explicit `#E8EFF8`, Beyond accent4 (`#E8EEF8`); blank body cells use bg1 white. The frame height is 3,955,065 EMU while raw XML row heights total 3,850,771 EMU. A native PowerPoint/AppleScript read of the opened source table reports rows 36.0 pt and 275.422454833984 pt, with all 13 column widths 68.24732208252 pt. The header expands from raw XML 27.788 pt to 36 pt to fit wrapped 12 pt headings. For a future composition, set native header height explicitly to 36 pt; do not infer native row height from frame scaling.

### Phase ribbons and tails

Each phase is a native `p:grpSp` containing two overlapping editable shapes: a solid `homePlate` base bar carrying text and a second `homePlate` with `a:pattFill prst="wdUpDiag"`. These are neither pictures nor freeforms. Both child shapes have `adj=39542/100000` and no line. The patterned tail has a white bg1 background. Group `off/ext` and `chOff/chExt` are distinct and must be retained. Theme `theme3.xml` resolves accent1 `#F900D3`, accent2 `#50658E`, accent3 `#CED7E6`, accent4 `#E8EEF8`, dk1/tx1 `#070154`, dk2/tx2 `#0047FF`, and bg1 white.

| Phase | Group ID/path | Effective group box x,y,w,h (in) | Solid child | Patterned child |
|---|---|---|---|---|
| 01 | 13 / `11` | .503472, 2.513358, 3.324710, .55 | ID 80 `11/2`, width 3.120737, accent3 | ID 79 `11/1`, width 1.017394, wdUpDiag accent3 |
| 02 | 12 / `8` | 3.371795, 3.304948, 4.879094, .55 | ID 70 `8/2`, width 3.987180, accent2 | ID 61 `8/1`, width 2.865227, wdUpDiag accent2 |
| 03 | 11 / `7` | 6.203118, 4.096537, 2.822440, .55 | ID 69 `7/2`, width 2.457548, tx1/dk1 | ID 63 `7/1`, width 1.615244, wdUpDiag tx1/dk1 |

All bar heights are 502,920 EMU (.55 in). Solid labels are centered, bold 11 pt. Phase 02 has a second centered 11 pt bold italic run, “3 Early Adopters.” The base/tail overlap and protrusion change across these three source examples, so the tail is not a fixed-width decorative rectangle.

The GA Onboarding bar is separate object ID 65/path `5`, a `homePlate` filled tx2/dk2 (`#0047FF`), x=9.053, y=4.888, w=3.78, h=.55 in. Ongoing project/product management is separate ID 67/path `6`, a `homePlate` filled accent4 (`#E8EEF8`), x=.503, y=5.68, w=8.528, h=.55 in.

### Milestones

Each marker is an ungrouped `star5` native shape, solid accent1 (`#F900D3`), no outline, 207,336 × 212,513 EMU (.226745 × .232407 in). Every caption is a separate ungrouped no-fill/no-line rectangle text box. Explicit caption runs are 10 pt italic; East Asian face is Roboto and complex-script face Segoe UI. Latin face and run color are not explicitly set. Captions have zero margins, square wrapping, top anchor and `spAutoFit`.

| Star ID/path | Label ID/path | Caption | Star x,y (in) | Label x,y,w,h (in) |
|---|---|---|---|---|
| 4 / `13` | 7 / `14` | Architecture Approved and Foundation Ready | 3.791620, 2.687399 | 4.084590, 2.719455, 2.979168, .168294 |
| 72 / `9` | 73 / `10` | Targeted Production Release for Early Adopters | 7.123695, 3.002373 | 7.413, 3.034, 2.979168, .168294 |
| 8 / `15` | 10 / `16` | GA Launch | 8.245234, 3.827148 | 8.538204, 3.875077, 2.979168, .168294 |
| 16 / `17` | 17 / `18` | UHG Ownership | 9.025558, 4.604649 | 9.318528, 4.660514, 2.979168, .168294 |

Marker-to-caption and marker-to-month associations are inferred from visual proximity and sequence. OOXML contains no semantic anchor or connector linking these sibling objects to one another or to grid columns.

## Proposed contract — not observed behavior

Represent `month_grid`, `phase_ribbon_group`, `extension_tail`, `milestone_marker`, `milestone_label`, `onboarding_arrow_bar` and `ongoing_arrow_bar` as typed, separately editable elements. Proposed slots: grid labels/cell edges; phase start/end/lane/rich label runs; tail enabled/end/pattern; milestone month/label/marker; ongoing start/end/label. Resolve placements from actual table cell edges rather than fixed slide inches.

A preserve-only build keeps each entire source group transform and both children unchanged. Future fixture candidates: rigidly translate a phase group with child offsets/lane/order preserved; compute start/end from month-grid positions; attach an editable wdUpDiag tail to the bar while preserving point adjustment and overlap; move each star/caption pair together with collision and fit checks. Keep star aspect ratio. These changed-content transforms are not yet approved.

## Writer gaps and risks

`src/core-enums.ts` lists homePlate and star5 presets, but `ShapeFillProps` in `src/core-interfaces.ts` permits only none/solid fills; there is no `wdUpDiag` pattern-fill API. The current Go composition renderer (`cmd/pptxcompose/render.go`) emits text, image, line and rectangular surface elements, not native groups, tables, preset ribbons or grid anchors. It supports rich text face/size/bold/italic inputs, but source paragraph/run identity preservation is unproven.

The main risks are nonuniform bar/tail overlap under duration edits, month positions inferred from raw coordinates, independent star/caption objects separating or colliding after edits, label fit, and losing editable theme-aware hatching. No evidence here qualifies different durations, variable label lengths, changed cardinality, reordering or PowerPoint roundtrip behavior through the project writer.

All numeric source facts and proposed constraints are also recorded in `library/component-contracts/wave2-roadmap-contract.json`.
