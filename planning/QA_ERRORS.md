# Reconstruction QA findings

This log records observed problems and limits. Pixel equality is always scoped
to the named renderer, resolution, and reference. It is not a claim about every
PowerPoint version or about generating new designs from prose.

| ID | Finding | Resolution / remaining work |
| --- | --- | --- |
| QA-01 | PowerPoint AppleScript export returned without creating a PDF when given a plain path string. | Pass a `POSIX file` to `save … in …`; require the PDF to exist before comparison. |
| QA-02 | Extracting/re-serializing source XML alone would overstate independent reconstruction. | Added v2 typed text, geometry, style, table and crop bindings, sentinel validation, mutation tests. Source object topology and master/layout resources are still retained. This is source-derived reconstruction, not inferred design. |
| QA-03 | A shortened deck changes automatic slide-number fields. | Explicit `--freeze-slide-numbers` converts them into editable text with source numbers. Report this intentional transformation. |
| QA-04 | A source slide rendered in a small deck can differ slightly from the same slide rendered in the full source deck. | Keep literal full-deck differences. Separately render an untouched source-slide control in the same small-deck context. Slide 11 control is exactly equal to its rebuilt candidate. This isolates the context effect without silently relaxing tolerance. Exact internal cause remains unproven. |
| QA-05 | Team slides link to omitted appendix slides. | Extract dependencies transitively and append destinations as hidden slides. Do not strip the links or silently redirect them. |
| QA-06 | The slide 5 underline SVG has about 46% transparent padding above and below its visible stroke. | Measure alpha bounds at high resolution; map the desired visible stroke to the full image rectangle. Image-box bounds are insufficient. |
| QA-07 | The slide 5 arrow has a source rotation of 345.7749167 degrees; its transformed visible extent is wider than its image canvas. | Include rotation and transform actual nontransparent raster-cell corners when measuring visible bounds. Unrotated bounding boxes are insufficient. |
| QA-08 | Earlier structural-only notes misidentified the slide 5 right panel and slide 8 imagery. | Native source rendering is mandatory. Metadata is supporting evidence, not a substitute for inspecting the rendered slide. |
| QA-09 | Detailed source slides include small text and dense placement; exact reproduction can preserve existing readability compromises. | Preserve source fidelity in this experiment; evaluate readability separately when adapting the layout to new content. |

## Review discipline

Each candidate is built, exported by native PowerPoint, rasterized at
1920×1080, compared, and visually inspected before moving to the next target.
Difference images exaggerate RGB differences four times; the JSON report
records literal differences, including very small background changes.
Source media remains source media: an existing architecture illustration is
retained as an image, while source-native text, tables, roles, connectors and
shapes remain native objects. No slide is replaced by a full-slide screenshot.

## Additional findings during final review

- **QA-10 — Coordinate schema mismatch, fixed:** native phrase bounds use
  `left/top`; asset alpha bounds use `x/y`. The initial anchor prototype decoded
  both as `x/y`. Root review caught this before applying the solver. Separate
  required-field types and a regression with the real measurement schema now
  prevent silent zero-origin placement.
- **QA-11 — Failed output cleanup, fixed:** extraction and build formerly could
  leave partial outputs on an ordinary error. Both now stage writes; tests cover
  broken relationships, missing bindings and pre-existing reports. The two-file
  PPTX/report commit is not guaranteed atomic across a process crash.
- **QA-12 — Unreliable preview claim, retracted:** one secondary visual review
  claimed missing content. Direct RGBA decoding and zero-difference controls
  contradicted it; the reviewer retracted the claim. Review conclusions must
  identify exact files and be reconciled with pixel/structural evidence.
- **QA-13 — Export differences persist in full-source context:** a second
  85-slide control made some targets exactly equal and changed faint background
  pattern rasterization on others. The underlying OOXML/resources are audited
  separately; no blanket full-reference pixel-perfect claim is made.
- **QA-14 — Experimental wording needed clearer labeling:** the slide 5
  underline-reflow fixture intentionally changes the title to force the phrase
  onto line two. The user noticed this during review. Original wording remains
  in `UHG-5-annotations-baseline.pptx`; the report labels the changed-title file
  as an experiment, not an exact recreation.

## Inventory review findings

- **QA-15 — Hidden-slide metadata location:** the modernization summary initially
  checked presentation slide IDs for `show` and reported zero hidden slides. Root
  review inspected each slide's `p:sld` root and found two hidden parts. Hidden
  status must come from the slide part and then be mapped into presentation order.
  The underlying raw inventory already retains these root attributes.

## Catalog execution findings

- **QA-16 — Hidden slides shift exported page numbers, fixed for review:** the
  first modernization PDF had 81 pages for 83 source slides. Added a task-copy
  preparer that unhides slides and records a source/page map; exported copies
  now have all 83 modernization and 35 EnableComp pages. The initial modernization
  `samples/catalog/render/software-modernization/png/` files are excluded from
  the reviewed manifest; `all-slides-png/` is the corrected source-indexed set.
- **QA-17 — Table fingerprints omitted internal structure, fixed:** outer table
  frames alone produced misleading candidate matches. Added column widths, row
  heights and cell merge attributes. Root also corrected the graphic-frame
  transform lookup to include `p:xfrm`. Geometry remains candidate evidence.
- **QA-18 — SVG viewBox is not pixel size, fixed:** the asset importer now keeps
  coordinate bounds separate from explicit pixel dimensions. Unknown dimensions
  remain null.
- **QA-19 — Repeated unreliable delegated visual descriptions, corrected:**
  draft reports invented missing table text and a nonexistent unboxed guide
  variant. Root opened the exact previews and corrected the reports before
  accepting family decisions. File references or confident descriptions alone
  are not visual evidence. Root review covered all 51 candidate previews.
- **QA-20 — Matching geometry can hide different roles, retained separately:**
  modernization 42 is blank while 77/80 contain titles. Their candidate geometry
  matches, but the blank slide stays separate and is not an approved template.
- **QA-21 — Native export timeout is not an output verdict:** PowerPoint waited
  for task-folder access and the AppleEvent timed out. After folder access was
  granted, export completed. Review checked actual PDF existence and page count
  before using previews, rather than treating the timeout as success or failure.

- **QA-22 — Aggregated slide search text was a list, fixed:** the first rebuild
  after adding descendant-text aggregation raised a type error before publishing
  a database. Joining text blocks fixed the build; a body-only COBOL search then
  returned source slides with readiness labels. Layout IDs are also searchable
  so useful names such as timeline are not lost when rationale wording differs.

## Resolved geometry and component findings

- **QA-23 — Nested placeholder inheritance can mix coordinate systems, guarded:**
  the geometry resolver accepts raw inherited transforms only for top-level
  placeholders. A nested placeholder without a transform remains unresolved;
  its layout/master ancestor matrix is not guessed. Per-slide transform state
  is cleared so matrices cannot leak across source slides.
- **QA-24 — Source biography reaches the footer:** native UHG 73 has a final
  biography line at the footer boundary. Retain the source and its bio-family
  membership while recording this as a source fit issue, not library approval.
- **QA-25 — Source underline detached from its phrase:** modernization 45 keeps
  the short underline below the introduction even though the corresponding
  emphasized wording in 44 has changed. Preserve it as a source issue and a
  future phrase-anchor regression example; no original was edited.
- **QA-26 — Frame bounds do not establish text capacity:** component slots retain
  observed lengths and explicit paragraph properties with null measured capacity.
  Effective typography, ink bounds and replacement fit remain unproven.
- **QA-27 — Pod role naming over-specified meaning, corrected:** the third role
  is Data Scientist in one source pod and Data Engineer in another. Renamed the
  proposed slot to `third_role`; retained each source value and excluded the
  shared phase backdrop from the individual pod compositions.


## Component expansion findings

- **QA-28 — Whole layouts mistaken for component boundaries, corrected:** draft
  grids and whole-slide regions were split into individual cards, phase columns,
  legend units and other coherent compounds. Uncertain broad diagrams remain a
  backlog rather than inflating the accepted component count.
- **QA-29 — XML order does not establish visual pairing, corrected:** stock 101
  bars were paired with unrelated labels, stock 114 bubbles with neighboring text,
  and stock 106 cells with explanations on the opposite side. Root geometry and
  native-preview review corrected or narrowed those selections.
- **QA-30 — Valid object paths can have the wrong semantic role, corrected:**
  proposal drafts called empty-text chevrons content panels and a phase ribbon an
  operating handoff. Those records were removed. The ingester now rejects empty
  text slots unless they have an explicit empty-role rationale. Tables cannot be
  passed as ordinary text-shape slots.
- **QA-31 — Visual descriptions need direct confirmation, corrected:** draft
  descriptions called quotation marks portraits, rectangular headers triangles,
  and white evidence cards blue-topped cards. Root inspected native images and
  corrected the accepted descriptions. Structural validity alone is insufficient.
- **QA-32 — Shared context changes component appearance, recorded:** UHG26 has a
  diagram-wide translucent overlay; stock157 metric pairs share a navy backdrop;
  connected diagrams share arrows and dividers. Composition boundaries record
  exclusions and dependencies; isolated reuse remains unapproved.
- **QA-33 — Duplicate layouts can contain identical subcomponents, consolidated:**
  the five title/body rows on stock24/25 match in geometry, style and text XML after
  excluding only `dirty` spellcheck-cache attributes. Five aliases replace repeat
  enrichment; slide25's added arrows remain distinct layout content.
- **QA-34 — Compiler/index consistency gaps, fixed after code review:** style IDs
  now use consistent namespaces and are validated on indexing; input/alias IDs
  are globally unique; exact selections with conflicting family assignments fail;
  three compiler outputs use staged installation with rollback on failure.

Final corpus ingestion: 249 input examples, 244 retained examples, 74 source
patterns, 31 semantic families, 572 slot candidates and zero unresolved component
bounds. SQLite has 17,451 items. The gallery verifies 61 source previews; the full
preview manifest has 144 entries. Native source review, ingestion and search were
performed; no unit suite was run. Gallery HTML UI remains unverified because the
previous browser security review blocked opening local HTML; no bypass was used.


## Reference selection and contract application — 2026-09-26

- **QA-35 — Component bounds need ownership review, corrected:** UHG43's individual
  pale pod container is included in the selected label shape. Only the shared
  Phase 2 backdrop and reporting connectors are excluded. The two- and three-role
  structures are distinct; a different third-role label is a content example.
- **QA-36 — A visible number was absent from catalog slots, corrected in contract:**
  stock109 path14 (native shape ID22) holds `01`. N1 explicitly binds it alongside
  title and three body paragraphs; the legacy catalog has not been silently changed.
- **QA-37 — Source links can drift from rendered references, fixed:** the shortlist
  gallery now hashes linked source PPTX files as well as PNGs, validates source
  identities and pins both component and geometry inputs.
- **QA-38 — Output containment through a symlink, fixed:** contract application
  resolves existing parent aliases before rejecting destinations inside the source
  project. The negative fixture failed without writing output.
- **QA-39 — Successful native text replacement can overflow, detected:** the metric
  long-label fixture spans 12 lines and exceeds its outer text frame by 114.55pt
  vertically, overlapping surrounding content. Four normal fixtures/14 text segments
  stayed inside their frames. The application command still requires post-build
  native measurement; it does not automatically reject this long-copy case yet.
- **QA-40 — Export readiness must be checked, recovered:** the first native PDF
  export waited on PowerPoint's task-folder access prompt and timed out. The prompt
  was resolved for the task folder; the resulting PDF and all subsequent exports
  were successfully rasterized and visually reviewed. A premature raster attempt
  before PDF availability failed; it was rerun after export completion.

Evidence: [contract checks and native proof](../library/component-contracts/proof-report.json).
Five negative input/output-boundary checks passed. Four edited scene trees and
all retained resources were unchanged except authorized text-binding values.
The native render review covers those particular fixtures, not arbitrary content
or full pixel-diff certification. No unit suite was run. Two Luna agents assisted
with reference/semantic/code review; the primary agent independently inspected
all nine shortlist source slides and all five generated fixture images.

## Dynamic roles and pods — 2026-09-26

- **QA-41 — Explicit zero insets disappeared in writer XML, fixed:** the writer
  omitted zero-valued margins, allowing PowerPoint's default padding to reappear.
  Explicit margin arrays now emit all four values, including zeros. Measurement
  probes use zero margins; final roles use declared horizontal/vertical insets.
- **QA-42 — Native presentation filters gave misleading results, fixed:**
  `every presentation whose name is ...` did not reliably filter open decks;
  iterating presentation references also stalled. The adapter reads the name
  list once, matches literal strings, and then obtains the named presentation.
- **QA-43 — Paragraph defaults contaminated font observations, fixed:** the
  complete text range could return no font name for multiline text or black
  font color for visible white text. Uniform font/style properties are checked across every non-break character,
  while text bounds use one whole-range measurement.
- **QA-44 — Geometry-only QA could miss styling defects, strengthened:** final
  verification now also compares native fills, foreground colors and text-frame
  margins to the generated plan. Planned contrast alone cannot prove the writer
  emitted the intended native colors.
- **QA-45 — Planner review caught row and color bugs before native acceptance:**
  initial row placement reused the same Y coordinate; row Y now accumulates row
  heights and gaps. RGB decoding now parses integer hex components; automatic
  foreground selection compares numeric ratios before selecting a color string.
- **QA-46 — Bundle boundary and freshness checks strengthened:** manifest deck
  names must be local PPTX basenames, deck symlinks are rejected, new-only output
  directory checks use `Lstat`, and native measurement checks that deck bytes did
  not change during the observation. Already-open same-name decks are rejected
  to avoid measuring unsaved state.

The generated engineering examples have their own dimensions and content. They
are not a claim of pixel-perfect reconstruction or general catalog adaptation
approval. Native proof results and artifact hashes are recorded in
[dynamic component proof](../library/dynamic-components/proof-report.json).

## Team composition and reporting routes — 2026-09-26

- **QA-47 — Shared-source exemption allowed branch crossings, fixed in review:**
  allowing every intersection between routes from one manager was too broad.
  Routes may now share one continuous prefix, then diverge without crossing or
  rejoining. Shared collinear strokes are emitted once with their relationship IDs.
- **QA-48 — A shared endpoint could hide overlapping approaches, fixed:** a check
  accepted any intersection if both segments included a shared endpoint. It now
  requires that the intersection collapse to that single point. The converging
  approach fixture is rejected; unrelated strokes also receive an ink-separation check.
- **QA-49 — Legend frames must keep their measured width, corrected:** shrinking
  label frames to the measured glyph width could introduce new wrapping. Final
  legend label frames now retain their exact probe widths and use equal slots.
- **QA-50 — Whole-range bounds produced a false overflow warning, corrected:**
  PowerPoint returned width 93.5pt at x37.625 for “Quality engineer” in a probe
  beginning at x36 with width 93.5pt. Its individual characters occupy about
  90.255pt and actually fit. The v4 adapter unions native character bounds and
  retains whole-range bounds as diagnostics. No copy, font size or container
  position was changed to suppress the warning.
- **QA-51 — Straight lines legitimately have one zero dimension, supported:** the
  adapter accepts horizontal/vertical lines but still rejects point-sized shapes.
  Native line color, width, transparency and absence of arrowheads are verified;
  unintended borders on text/surface shapes are rejected.
- **QA-52 — Phase membership and staffing legends need semantic checks:** phases
  cannot cover undeclared components or the legend. Team roles require canonical
  staffing tokens, and legend entries derive from tokens actually used. Missing
  legends, unknown phase members and literal colors impersonating semantic roles
  fail before native measurement.

See [team proof](../library/dynamic-components/team-proof.json). The examples use
new geometry and content informed by UHG43; the source-sized reconstruction gate
and PowerPoint-glued connector behavior remain separate work.

## Autonomous showcase wave — 2026-09-26

- **QA-53 — Accent collision coverage was incomplete, fixed in review:** initial
  checks covered canvas text/images only. They now include slide title, cards,
  pods, standalone roles, phase surfaces, legend and reporting strokes. Accents
  require an unfilled target so a text-box fill cannot hide the artwork.
- **QA-54 — Surface stacking could hide a highlight, fixed:** canvas backgrounds
  now precede accents; text follows them. Surfaces have background semantics in
  this bounded compositor. Render review still decides whether emphasis works.
- **QA-55 — Alpha crop threshold could discard faint edge pixels, fixed:** tight
  derived assets now include every nonzero alpha pixel. Original PNGs, dimensions,
  source hashes and derivation records are retained in the selected asset catalog.
- **QA-56 — Connected diagram strokes need explicit endpoint relationships:** the
  canvas collision checker includes stroke width, so line ends touching a box are
  intentional overlaps and must name the connected boxes in `allow_overlap`.
- **QA-57 — Reusing text evidence needs a full measurement-contract comparison:**
  the explicit reuse flag still requires source manifest/deck/evidence hashes and
  exact equality of all text probe requests. It records the old spec hash and
  never removes final native verification. Changed text or inner width needs new
  evidence; geometry-only edits do not require measuring identical text again.
- **QA-58 — Wrapped trailing spaces can exceed a text frame without visible
  overflow, corrected:** native probe `ask` reported width 370.565pt in a 370pt
  frame. A character-level inspection found only character 92, a space, outside
  the frame (left 401.810pt, width 4.755pt, right edge 406.565pt versus 406pt).
  The v5 adapter excludes whitespace from visible bounds while still checking
  its font/color. No copy, font size or layout was changed to hide this warning.
  Character/font property snapshots reduce native automation round trips.
- **QA-59 — Text recovery ignored the slide root coordinate transform, fixed:**
  a task-copy mutation shifted the root frame while leaving child coordinates
  unchanged. Recovery now compares root coordinates, rejects nonzero root
  rotation/reflection and checks source deck/bundle slide counts. The shifted-root
  negative fixture is retained; arbitrary imported layout reconciliation is still
  outside text-only recovery.
- **QA-60 — The first showcase did not exercise proposal complexity:** user review
  found it too sparse. It is retained as a mechanics baseline. The new five-page
  benchmark follows EnableComp 3/4 and UHG 28/38/43, with full activity/output rows,
  workflow relationships, dense phase detail, client ownership and shared roles.
  Word/object counts are diagnostics; native and visual review remain separate.
- **QA-61 — Filled labels lacked deliberate padding:** showcase-v1's comparison
  headers and architecture bars have zero internal margins. The dense generator
  defines padded cells and separately inset panel text, avoiding double padding.
  The v1 artifact is preserved as reviewed, not silently overwritten.
- **QA-62 — First-error fit diagnostics made dense-page iteration inefficient:**
  `fit-report` now records every fixed zone's measured and available width/height,
  overflow in points, and the full planner outcome. A two-overflow fixture reports
  both independently; 119 fixed zones from the existing measured baseline fit.
- **QA-63 — Dense layout native fit rejected nine zones:** the first 258-probe
  pass found one activity bullet, four matrix cells, the team assumptions copy
  and three roadmap labels outside their allocated heights. Reallocated bullet
  heights, adjusted matrix row height/padding and enlarged the team copy frame
  without shrinking fonts. The roadmap correction replaced redundant long bar
  labels with explicit month labels. Original failure report is retained.
- **QA-64 — Schedule pages contradicted one another:** the phase narrative targeted
  a first release in weeks 1–12 while the first roadmap placed release in months
  5–7. Revised the month-level roadmap to a first release by month 3 and expansion
  afterwards. Changed labels and widths require fresh native measurements.
- **QA-65 — Dense team routing was legal but visually crowded:** the program
  analyst branch ran 4pt from the pod trunk. Exchanged the horizontal positions
  of Program manager and Change lead so the analyst reports vertically. Copy,
  fonts, widths, role IDs and measurement request order remain unchanged.

## QA66 — Recovery trusted manifest fields without checking the original scene

The code audit found that a valid deck hash alone did not validate the manifest's object frames/text. Recovery now checks the original shape tree against the manifest, then checks returned object types, slide size, frames and reflection state. A real text edit saved by PowerPoint still recovers as exactly one change. Focused negative fixtures record the rejected structural edits.

## QA67 — Fit-report tolerances differed from planner contracts

The fixed-zone report previously used one tolerance for different component contracts and handled whitespace-only titles differently from probe generation. It now uses each contract's tolerance, trims title presence consistently, and reports the tolerance per zone. All 244 fixed zones in the final dense sample fit; the complete planner also covers dynamic zones, collisions and routes.

## QA68 — Reusable layout expansion mutated its input

A shallow slice copy let probe expansion remove layouts from the caller's original
spec. Layout lowering now copies the slide and canvas slices; a permanent
regression checks source bytes before and after probing and planning.

## QA69 — Generated overlap permissions were too broad

Initial lowering allowed every generated layout object to overlap every other
one, which could hide text collisions across nested panels. Automatic exemptions
now cover owning/ancestor backgrounds only. A regression rejects nested text
collisions. Descendants must stay inside padded parent content.

## QA70 — Contrast used the wrong background

Transparent blue output labels on the dense control's gray panel had only 4.33:1
contrast. The old canvas validator assumed white. Layouts now inherit the actual
cell/ancestor surface through `contrast_background`; opaque fills take precedence.
The five output labels are navy in the new controls; copy and font sizes are
preserved. This intentional visual correction is excluded from pixel-identity claims.

## QA71 — Dynamic failure reporting hid later failures

A row resolver stopped at the first overfull row, while a fallback could treat
1pt probe placeholders as dynamic capacity. Each bounded cell now reports its own
overflow. Failed expansion reports original fixed zones separately from layout
failures. Deliberately overfull multi-cell fixtures must identify both cells.

## QA72 — Layer controls could hide descendants

Team phase surfaces previously rendered after canvas labels. Final ordering now
places phase backgrounds first and honors canvas layers; generated layout layer
offsets are bounded and nonnegative. Regression coverage checks phase/background
ordering and rejects negative child offsets that would put text behind its surface.

## QA73 — Roster inventory missed the partial final row

The first Wave 2 audit counted nine core profiles on UHG 44. Root visual review
found the tenth in a partial fourth row. The corrected record lists all 10 core
and 9 specialist portrait/card pairs with source object IDs and asset hashes.

## QA74 — Partial review preview mistaken for slide clipping

The independent reviewer initially rejected the six-row fixture after seeing only
the lower portion of its PNG preview. The primary review contradicted that claim.
Reopening the exact absolute file at original detail, checking 1920×1080 dimensions
and SHA-256, confirmed intact title, headers and all six rows. The reviewer retracted
the rejection. Resolve display discrepancies against the actual full artifact;
never edit a deck to fix a cropped tool preview.

## QA75 — Layer and fit-report limits found in final code review

No finding invalidates the six tested layout fixtures. Before broader API release,
validate direct canvas/row-rule layers and accumulated nested layers, not only each
container/cell/block offset. Row-rule layers currently remain explicit unchecked
values. Cache-backed fit reports have empty legacy top-level probe/evidence hashes;
use their per-contract `cache_uses` evidence paths and SHA-256 values. A follow-up
should omit or replace the empty legacy fields. Neither limitation is hidden by
claiming general layout or provenance approval.

## QA76 — Source screenshot distortion must be explicit

UHG28 picture 19 has a source aspect ratio of approximately 2.417 but a frame
ratio of 1.744. Reproducing it with a generic aspect-preserving placement changes
the source appearance. The control now records explicit `stretch`; normal image
placement rejects unintended distortion. This source behavior is not a default
for replacement imagery.

## QA77 — Native export blocked by file access

The Wave 2 smoke deck passed native measurement and final object verification,
but PDF export timed out (-1712); subsequent opens failed (-9074), including a
known-good Wave 1 deck. Desktop inspection also failed to connect. The user
reported granting access; the next probe deck opened successfully and native
measurement resumed. Preserve unsaved user decks during this failure mode.

## QA78 — Structural shape checks accepted ambiguous input

Code review found a duplicate-name map that could accept a third identical shape
name after rejecting the second, plus adjustment parsing that tolerated duplicate
guides and trailing formula text. Monotonic name counts, unique guide names and
exact canonical formula parsing now reject these cases; regression tests cover
all three. This hardening precedes final Wave 2 qualification.

## QA79 — Picture frames alone do not prove crop fidelity

Native object checks inspect picture bounds but do not expose the full source
crop contract. Independent review identified the need to bind the generated
picture relationship, media hash and source rectangle to the spec. A picture
with the right frame and wrong crop must fail structural verification; rendered
review is still required to assess the visible result.

## QA80 — Source line spacing was lost during caption reconstruction

The first full Wave 2 native probe measured two caption overflows: 0.2983pt in
the source control and 0.2485pt in the scaled instance. The source XML specifies
90% line spacing; the renderer used 100%. A source native read confirmed
`line rule within=true`, `space within=0.899999976158` and a 9.720000267pt
first-character height, versus 10.800000191pt in the generated caption. Preserve
the source frame and font size; add and verify paragraph line spacing instead.

## QA81 — Source roster text was vertically approximated

The source card uses middle vertical anchoring and one paragraph with a soft
line break. Its native name/role glyph tops are 155.2109375pt/166.010940551758pt;
the initial fixture had 151.2pt/165.3pt. Source-control boxes now use the measured
tops, retaining the explicit limitation that they are separate text shapes.

## QA82 — Crop comparison did not exercise a crop

A square source portrait in two square frames gave contain and cover the same
result. The comparison now uses equal 76×52pt frames and a declared cover focal
point so the render can actually demonstrate full-source versus cropped behavior.

## QA83 — Visible icon placement needs paint-order review

All five EnableComp response icons existed at the correct frames and retained
their media hashes, but their default layer 0 put them behind the opaque row
surfaces at layer 5. Full-slide visual review caught the missing icons. The
fixture now explicitly places them at layer 10. The corrected render shows all
five. Object existence, frame checks and `visible=true` do not prove that later
objects have not covered an asset.

## QA84 — Bounding boxes require explicit decorative relationships

Planning found the source milestone stars' boxes touching the corner regions of
the adjacent patterned phase tails. Those source-derived decorative relationships
are now explicit and visually reviewed; label boxes remain outside the artwork.
The response arrow also requires explicit overlap with both the foreground need
tile and its label. A separate accidental overlap between comparison labels was
fixed by narrowing the oversized source-bio label box, then remeasuring its width.

## QA85 — Bulk character queries collapse to an aggregate range

The PowerPoint AppleScript `every character` request returned a single combined
range rather than a list of individual observations. It cannot replace the
per-character checks. Retained exact character inspection and batched paragraph
properties only. The six-object smoke retained exact historical native values.
See `WAVE2_VERIFIER_SPIKE.md`; no large performance improvement is claimed.

## QA86 — Unequal biography bullet pitch and sparse responsibilities

Independent Wave 2 review identified uneven sidebar spacing and a lower panel
with only three responsibilities. The generator now uses a consistent 23pt
minimum for each sidebar item and five proposed responsibility items. Source
portrait/card controls are unchanged. Native follow-up fit and visual review
remain the acceptance gate.

## QA87 — SVG media needs a genuine fallback and a separate media check

The underlying image writer's legacy SVG path could place broken placeholder or
SVG bytes in the PNG fallback relationship. Native SVG support now requires an
explicit pinned PNG fallback in the compose contract, preserving both media
hashes. Structural validation must check the SVG extension relationship and the
PNG relationship independently; successful picture-frame measurement alone does
not prove visible artwork.

## QA88 — Four independent border lines leave corner residuals

UHG28 source picture borders are native picture outlines. The v7 approximation
used four separate line objects, with measurable residuals near the edges.
The follow-up introduces native picture outlines with explicit color and width;
source-region comparisons determine the remaining difference without declaring
whole-slide identity.


## QA89 — Successful automation return does not prove an export exists

The follow-up PDF export timed out. A subsequent save command returned success
but created no file; a file-existence check caught this before any visual proof
was accepted. Later PowerPoint opens returned -9074. Preserve unsaved decks and
require the actual PDF plus successful rasterization before reporting an export.
The user granted repository file access, after which native export and rasterization
succeeded. The exporter now rejects an existing destination and requires a PDF
header, preventing stale or missing files from being accepted as new evidence.

## QA90 — Native bullet metadata does not prove a visible glyph

The v8b native inspector reported a visible 100% Arial `▪` bullet, but the PDF
render showed no marker. Both primary and independent visual reviews rejected
slide 1. Round `•` and dash `–` rendered visibly. The bounded compose contract now
rejects the square glyph, and the fixture uses a round marker. A negative test
prevents accidental reintroduction. Retain the initial render/review and native
verification to demonstrate why metadata checks alone cannot approve a bullet.

## QA91 — Reused explanatory copy referenced a slide absent from the deck

The four-slide follow-up inherited a statement about a “following phase-detail
page” from the eight-slide checkpoint. Visual/content review caught the stale
cross-reference. The follow-up generator now describes the native picture-outline
change instead. Reused components require narrative review as well as geometry QA.
The corrected v2 render removes the stale reference and retains source thumbnail
and caption positions.

## QA92 — Path declarations disappeared during layout expansion

The first path primitive expanded into layouts, but the layout copier then
replaced expanded slides with the original slide array. A process-path test
caught an empty plan. Expansion now retains the lowered slides, and tests verify
node ports and arrow counts before and after cardinality changes.

## QA93 — Connector obstacles must include the whole arrow

Root review found that direct flow arrows checked labels but not prior routed
strokes, and later routes only saw the arrow's semantic centerline. Full arrow
ink frames now participate in obstacle checks with declared clearance; source
IDs and contrast on traversed surfaces are checked. Focused tests cover both
connection orders and generated-ID collisions.

## QA94 — Large source typography was rejected by a blanket contrast threshold

UHG36 uses original #F900D3 numerals at 16pt bold. The engine's blanket 4.5:1
check rejected them. The proposed color darkening was rejected during integration.
Uniform canvas text now applies 3:1 for >=18pt regular or >=14pt bold, retaining
4.5:1 for smaller text. Source colors remain unchanged. This is a typography
rule, not a claim of whole-slide accessibility certification.

## QA95 — Repeated phrase selection can still wrap

A fixture explicitly selected the second occurrence of “pivotal moment,” but
native character measurement showed the phrase split across two lines. The
planner correctly rejected automatic union-box placement. The experiment now
explicitly requests one underline per measured line, without rewriting the text.
Native visual acceptance remains pending.

## QA96 — Arrow catalog geometry and identity require independent inspection

Root inspection rejected the initial connecting-arrow tip: its annotated dot
sat on the separate return stroke, visibly beyond the actual arrowhead. Root
also found the claimed UHG byte identity contradicted the local/source SHA values
in the same draft record. The catalog audit is being corrected; no candidate
was promoted or accepted for native use on the basis of that draft.

## QA97 — Wave 3 export is blocked again after successful measurement

The seven-case phrase probe measured successfully in 64.14 seconds. PDF export
of the six regular cases then timed out (-1712); a subsequent open returned
-9074. Desktop inspection failed to start. The user was asked to inspect/grant
any repository file-access prompt. No PDF was produced, no native visual result
is claimed, and unsaved presentations remain open. File-only implementation and
qualification preparation continue while access is unresolved.

## QA98 — Accent and artwork checks omitted other decorative ink

Root review found that ordinary text/image collisions were checked but overlapping
accents and prior staged arrows could be missed. Both now participate in the
shared collision check. Final native verification also rechecks arrow phrase
endpoints against final character bounds. Regression tests cover shifted/missing
phrase geometry and decorative overlaps. Rich-text phrase resolution now uses
its actual paragraph/run content rather than the empty plain-text field.

## QA99 — Rotated curved-arrow rectangles reject clear empty corners

The initial ±15° arrow experiments failed on source/target labels because the
rotated alpha bounding box included empty corners. Simply increasing endpoint
clearance would have required roughly 45–63pt source gaps in two cases. Candidate
arrows now carry conservative occupied alpha tiles derived from pinned source
SVG rasters; the solver rotates these with the artwork for collision checks.
Nine candidates pass structural planning. Native optical acceptance is pending.

## QA100 — Stock SVG/PNG names do not guarantee identical image canvases

Package tests and independent review found that the single-arrow SVG is about
30.14:1 but its stock PNG is square; the right-angle PNG is 150×186 while its SVG
canvas is square. These are not interchangeable fallbacks. The candidate builder
now rasterizes the exact SVG at the original aspect and pins those derived PNG
bytes. The original SVG is preserved. The narrow static-SVG lane also now permits
plain title/description metadata while rejecting nested content and active SVG.

## QA101 — Narrative and style metadata must bind to the produced slide

Library review found that a narrative could be attached without checking its
assertion title/role/takeaway against output, and token replacement could rewrite
arbitrary strings. Contracts now own explicit narrative bindings; required detail
and qualifications must be present in supplied copy. Style replacements affect
only declared color fields. Tests cover mismatched narrative and unchanged
content/IDs/asset paths. A deterministic assembler validates each instance and
sets only declared slide IDs/page labels, publishing atomically after all pass.

## QA102 — A single accent study is too sparse for the proposal opening

Narrative review found that the opening was mapped to an accent-only mechanics
fixture. A new dense-argument candidate now keeps an assertion/highlight, synthetic
baseline, three detailed mechanism/boundary/evidence panels, and a visible
qualification. It remains unqualified until native measurement and visual review.
The roster and full biography are split into separate pages, giving a 13-page
proposal narrative rather than discarding their required detail.

## QA103 — Source-length caps were applied to nonvisual narrative metadata

The independent authoring check rejected an architecture takeaway and later team
and decision-page role metadata because caps had been scaled from terse source
labels. These fields are slide notes/narrative metadata, not rendered text zones.
They now have an explicit 2048-character metadata bound. Top-level assertion title
input caps have a 128-character editorial floor; for example the needs template
has an 888×52pt, 23pt title frame but its old 65-character guard came only from
source copy length. No native maximum was measured. New title copy remains
unqualified until actual native fit; fonts and geometry are unchanged.

## QA104 — Full proposal content exceeds the first semantic templates

The independent agent authored 13 value files from the narrative. Six pages
assemble in isolation; seven expose exact required-detail/capacity failures.
See library/proposal/authoring-notes.md for the specific sentences and slots.
The source process mechanics fixture is especially sparse: 264 characters of
non-binding string capacity versus 807 characters of required explanation and
qualifications. Roadmap labels and roster role tiles also cannot serve as long
qualification paragraphs. These need designed explanatory/qualification zones,
not silent text cuts or relaxed claims of fit. Full proposal assembly, native fit,
and visual acceptance remain incomplete.

## QA105 — Native open failure is not limited to the new accent deck

After confirming its saved state, root closed only the owned accent-review-spike
copy. Verification then failed opening that file with -9074. To isolate repository
file access, saved task copies were placed in PowerPoint's own document container
under `pptxgengo-wave3-export-probe`: accent-container-check.pptx and
followup-known-good.pptx (the latter from the verified Wave2 follow-up). Both
AppleScript opens failed with -9074. Launch Services accepted an open request for
the known-good copy, but the subsequent presentation inventory showed no new
presentation. No PDF was created. No app restart, source save, or unsaved-deck
closure occurred. The user was asked for the current dialog text/screenshot;
independent code/content work continues. A pending Launch Services request could
open the owned copy after the dialog is resolved, so inspect names before the
next native operation.
