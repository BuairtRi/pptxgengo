# Wave 2 visual component candidates

2026-09-26. Read-only reference and asset audit for Wave 2. These are inventory
candidates, not approved reusable components. The allowed scope was UHG 24/28,
38, 44, 67 and EnableComp 5.

The detailed record is [visual-wave2-candidates.json](../library/component-contracts/visual-wave2-candidates.json).
It records source deck SHA-256 values, slide XML object IDs/paths, source component
seed IDs where they exist, local geometry, slot roles, transform recommendations,
source media part hashes and missing capabilities. Media remains referenced inside
the source PPTX; this audit did not extract, convert, rasterize, or alter assets.

## Inspected evidence

- UHG phase detail: [slide 24 preview](../samples/reconstruction/reference/png/slide-024.png),
  [slide 28 preview](../samples/reconstruction/reference/png/slide-028.png), and their
  extracted text/structure in [slide 24 sidecar](../samples/inspection/uhg/sidecars/slide-024.md)
  and [slide 28 sidecar](../samples/inspection/uhg/sidecars/slide-028.md).
- UHG roadmap, roster and bio: [slide 38](../samples/reconstruction/reference/png/slide-038.png),
  [slide 44](../samples/reconstruction/reference/png/slide-044.png),
  [slide 67](../samples/reconstruction/reference/png/slide-067.png), with
  [slide 38 sidecar](../samples/inspection/uhg/sidecars/slide-038.md),
  [slide 44 sidecar](../samples/inspection/uhg/sidecars/slide-044.md), and
  [slide 67 sidecar](../samples/inspection/uhg/sidecars/slide-067.md).
- EnableComp response rows: [slide 5 preview](../samples/catalog/render/enablecomp/png/slide-005.png)
  and [slide 5 sidecar](../samples/inspection/enablecomp/sidecars/slide-005.md).
- The source records already name [portrait with name and role](../library/component-seeds-expanded.json)
  (`component:roster-person-card`), [bio portrait and identity](../library/component-seeds-expanded.json)
  (`component:bio-identity`), and [UHG roadmap phase bar](../library/component-seeds-expanded.json)
  (`component:slice-proposal-uhg-roadmap-phase-bar`). The named seed entries are candidates;
  their adaptation constraints and review status remain in force.

## Pattern observations and contract implications

### UHG phase 01/02 deliverable preview stacks

Both slides place two labeled thumbnail groups in the right third, with one or two
small wide screenshots in each group. Slide 24 puts each pair of images inside a
PowerPoint group; slide 28 places its four images as separate pictures. Labels sit
outside those groups. Thus “thumbnail stack” is the repeated visual family, not an
identical object structure. The contract must keep each caption paired with its
preview images, retain the exact image source hash, and declare crop versus contain
for every replacement. Evidence source thumbnails are already embedded raster
artwork; a slot is not automatically editable as the represented document.

Geometry is narrow and irregular: on 13.333 × 7.5 in slides the upper group sits in
the approximate x=9.57–12.80 in band and y=1.01–2.22 in; the lower group sits at
x=9.57–12.95 in and y=2.60–3.66 in. The screenshot frames differ in dimensions and
some overlap visually. The initial safe recommendation is fixed source geometry;
whole-group translation is plausible only if captions and images move together.
Resizing and automatic crop are unqualified.

### UHG phase 38 roadmap

The month header is a 13-column table over x=0.503–12.825 in. Three colored phase
bars occupy separate tracks at y≈2.51, 3.01 and 4.12 in. Bars terminate in patterned
hatched extensions; magenta stars and italic milestone labels sit at month-relative
positions. A separate GA onboarding arrow and ongoing work arrow span later periods.

This pattern is a page-level roadmap composition; the existing
`component:slice-proposal-uhg-roadmap-phase-bar` seed only captures individual
labeled grouped bars. A future contract needs both layers: per-bar label/interval
slots and a parent month-grid with milestone labels, arrow rows, ongoing work and
legend. Dates and bar lengths are content. Preserve bar height, point shape, hatch
pattern and star aspect ratio. Change durations using month-grid parameters, with
all related tails and milestones anchored to that grid. Current sources do not
prove such edits, label fit, or collision-safe relocation.

### UHG 44 roster tiles

Slide 44 has two sections with dark navy full-width headers. The core team has ten
portrait/name-role tiles in a three-column arrangement, with Adama Sando as the
left-aligned fourth-row tile, plus a neighboring responsibility list. The
specialists section has nine tiles in the same three-column arrangement and
another responsibility list. A representative tile uses a 2.4 ×
0.5 in pale-blue text card overlapped by a 0.5 × 0.5 in square portrait at its left
edge. Name and role share formatted text; these identities are linked to portraits.

The existing roster seed is the closest match but it contains only three illustrative
pairings; the XML confirms 19. The JSON candidate lists all 19 person/card/picture
source paths and embedded image hashes. Avoid splitting name and role into unrelated
text boxes or letting the portrait become detached from the person.
Portrait replacement needs an explicit crop rule; grid/card placement could become
parametric only inside tested row/column and long-role limits. The whole team page
also includes section headings and responsibility bullets, which the tile seed does
not capture.

### UHG 67 biography page

This is a three-zone biography composition: a left portrait/identity column, a wide
central narrative column with two blue subheads, and two stacked right-side pale-blue
expertise panels containing headings and bullets. Text contains multiple formats
(bold, blue, underlined, italic, bullets) and is materially denser than a simple
profile card. The portrait/identity seed represents only the left identity pairing;
it is not an alias for this complete page.

Use fixed zones until the source layout’s inherited text frames are exposed and
measured. Proposed slots must preserve paragraph and run formatting, bullet and
indent behavior, section-heading associations, and portrait crop. The visible source
copy is specific to its named person and should not be carried into a synthetic
sample. No profile facts or credentials should be inferred from this illustration.

### EnableComp 5 needs-to-response rows

Five dark navy need tiles on the left align conceptually with five pale-gray
response rows on the right. A large pale-gray arrow sits behind the relationship.
Each response row has a vector icon, thin vertical separator, numbered bold heading,
and explanatory sentence; the row-to-need association is conveyed through the
ordered alignment. The SVG icons are original embedded PPTX parts, each retained in
the inventory by package path and SHA-256.

A useful contract needs stable row identity plus explicit needs/response pairing,
not just five visually similar cards. It also needs rich text runs, vector-icon
import or pinned raster preview support, and measured two-line copy behavior. Treat
the left tiles and shared arrow as slide-level structure. The row pitch is about
0.857 in and the row frame is 8.036 × 0.650 in, so added text cannot be expected to
fit through font shrinking.

## Cross-check of candidate cardinalities

The following counts were rechecked against the rendered slide and/or source slide
XML and extracted text, rather than inferred from the sample component seed alone:

| Source | Verified visible/source structure |
| --- | --- |
| UHG 24 | Two deliverable captions; four thumbnail pictures in two pairs; five source pictures total when the logo is included. |
| UHG 28 | Two deliverable captions; four thumbnail pictures in two pairs; five source pictures total when the logo is included. |
| UHG 38 | Thirteen month-grid headers (months 1–12 and Beyond), three grouped phase bars, four star milestones, plus the ongoing work and GA onboarding bars. |
| UHG 44 | Ten core-team portrait/name-role tiles, including Adama Sando, and nine specialist tiles: 19 portrait pictures in total. The logo is inherited and is not a portrait. |
| UHG 67 | One portrait; two expertise list panels with five industry entries and five specialties in the extracted slide text. |
| EnableComp 5 | Five needs, five response rows and five embedded SVG icons. |

## Deduplication and provenance

The JSON distinguishes repeated family instances from duplicates. UHG 24/28 share a
thumbnail-stack layout but have differing image grouping and content. Roadmap bars
11/12/13 share the named phase-bar seed but have different lengths and positions.
Roster tiles repeat a structure; there are 10 core-team and 9 specialist entries, each with a separate portrait media part. The candidate JSON lists all 19 object/media pairings, including Adama Sando (slide44 shape 23, picture 29, `ppt/media/image179.png`). UHG bio
identity and roster tiles share broad `component-family:person-card` lineage, while
the full biography page is separate. EnableComp rows repeat structure while their
icons and explanations differ. None of these observations establish exact duplicate
content or authorize component reuse.

The JSON stores the two source deck hashes and original embedded media part hashes.
Where sidecars have no image descriptions, the candidate record explicitly says so
and uses a neutral source-artwork role. It does not guess thumbnail subject matter,
portrait attributes, or metadata missing from catalog records.

## Review boundary

Visual inspection and existing scene/component inventory inform the proposed slots
and geometry. No component implementation, transform validation, source asset
extraction, or reuse approval occurred in this task. Wave 2 qualification still
requires the source-faithful control, changed-content instance, cardinality/copy
stress cases, native PowerPoint render review, object/text measurements, and a
record of unsupported cases as specified in
[Fidelity implementation waves](FIDELITY_IMPLEMENTATION_WAVES.md).

## Subsequent extraction checkpoint

After the read-only inventory, `scripts/extract-wave2-assets.py` verified both
source deck hashes and each enumerated embedded media hash, then extracted 33
unique binaries with 33 source associations into the ignored local directory
`samples/visual-wave2/assets`. Exact-byte deduplication found no duplicates in this
selected set. The manifest preserves all source slide/part/family associations;
no artwork was converted or substituted. Native shapes such as roadmap bars are
not rasterized by this extractor. All extracted bytes were independently rehashed.

The tracked [extraction checkpoint](../library/component-contracts/visual-wave2-extraction.json)
records source and output hashes. Reproduce with
`python3 scripts/extract-wave2-assets.py --out samples/visual-wave2/new-assets`.
Output must be a new directory. This prepares source artwork only; rich typography,
placement contracts, changed-content adaptation and native visual qualification
remain Wave 2 implementation work.
