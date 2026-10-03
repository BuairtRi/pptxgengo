# Slice 4: integrated Team and Solution contracts

Frozen source `wmds-library.v2`, commit
`7bdcaee5030a12275a1f881a8542f4d302d207df`.

The complete families contain **42 active designs**: every design gets a source
specimen and a meaningful alternate, producing **84 slides**. All 42 source
fixtures and all 42 source/alternate pairs generate. The 19 added designs reuse
exact reviewed slice 3 body content, navigation, identities and registered assets.
The remaining 23 designs have deliberate per-template body substitutions, or
corresponding reviewed body bindings from their nav/split sibling. No specimen
uses only header changes. Every alternate has at least two changed body slots.

All 84 integrated native pages have been reviewed and accepted for these paired
specimens. Five alternate-only identity headers were corrected and rechecked at
full size in the final native export. Earlier slice 3
acceptance covers its prior 19-pair packet; generation does not qualify arbitrary copy.
No tests, negative controls or native UI automation were run by this family agent.

## Remaining designs: closed content contract

| Template | Required slots | Fixed array key sets |
| --- | ---: | ---: |
| architecture/layer-map | 56 | 7 |
| architecture/layers | 26 | 2 |
| architecture/nested | 18 | 1 |
| architecture/product | 45 | 2 |
| architecture/reference | 46 | 3 |
| bio-full/portrait-list | 13 | 1 |
| bio-full/portrait-quote | 16 | 1 |
| bios/leadership-specialists | 52 | 8 |
| bios/three | 26 | 6 |
| context/build-sustain | 25 | 2 |
| context/three-zones | 46 | 0 |
| governance/stack | 29 | 16 |
| pathways/two-lanes | 27 | 2 |
| patterns/three-rows | 24 | 3 |
| process/swimlane | 18 | 2 |
| readiness/six-criteria | 30 | 6 |
| roles/by-phase | 43 | 3 |
| team/org-chart | 20 | 4 |
| team/org-roles | 40 | 9 |
| team/pods | 23 | 3 |
| team/pods-pairs | 36 | 5 |
| team/roster | 48 | 4 |
| workstreams/five-with-risks | 44 | 10 |

The exact slots, source pointers, kinds and fixed cardinalities come from the
catalog generated alongside this packet. Array keys remain explicit caller input.
Structural table-group column indices remain outside v2 content slots. Chart,
process, org-chart and architecture relationships remain source geometry; this
packet does not add count-driven diagram adaptation. Photo/icon asset IDs stay
valid registered keys. Navigation retains the separate caller labels/keys/active
contract on the 19 reviewed additions.

The corresponding reviewed body bindings are reused explicitly for layers,
leadership-specialists, three bios, six readiness criteria, phased roles, org
charts, org roles, pods, paired pods and roster. Slots absent from a sibling's
specific contract are never supplied. Each design's source header and source copy
remain unchanged. Initials remain consistent with the alternate fictional names.

Other templates have explicit per-key body changes: platform layer labels,
analysis inputs and synthetic duration/counts; reviewed proof-point wording;
portrait names and experience; governance decision scope; process routing copy;
platform patterns, pathways and workstream activity/risk statements. The script
contains these exact substitutions in a closed dictionary. No blind replacement,
font shrinking or missing-content fallback is used.

All claims, metrics and identities in both specimens are synthetic, as recorded
by their `content_kind:synthetic_example` and generation receipts. They are not
client findings, team assignments or commitments. Source-authored qualifications
and claim text are preserved in the source specimen.

## Geometry and review history

The existing named refinements and v2 outline-surface parity are retained. No new
Go, source JSON or v1 changes were needed to generate these 42 pairs. The initial
architecture/layer-map alternate used a two-word ingestion label that exceeded
its one-line cell capacity. Its alternate-only copy is now `Intake`; source copy,
source geometry and fonts remain intact. The corrected packet generates cleanly.

The integrated native review found no clipping, overlapping text, bad badge wraps
or missing outline margins across the 42 source/alternate pairs. All pages were
inspected in contact sheets; dense diagrams, bios, tables and bottom risk bands
were also inspected at full size. Alternate header identity mismatches on pages
260, 266, 270, 302 and 318 were corrected explicitly in `HEADER_COPY`: Lakeview
tenant/readiness/pods and Jamie Chen/Alex Morgan profiles now have matching
titles. Every alternate still has at least two substantive body changes. All 42
pairs regenerate; only those five compiled slides differ from the initial packet.
The initial native PDF/PNG evidence remains separate and unchanged. The durable
`native-review.json` records initial findings, final PDF/PNG hashes, resolution
history and accepted per-key source/alternate results. The final five corrected
pages are clean; SHA256 comparison confirms the other 79 family PNGs are identical
to their initially reviewed versions. The prior family packet is retained in
`work/archive/`. This acceptance covers the 84 explicit specimens, not arbitrary
replacement copy or array counts.

## Reproduction and artifacts

```sh
python3 scripts/wmds-integrated-team-solution.py \
  --cli /tmp/wmds-slice3-root-pptxdesign \
  --out /tmp/wmds-integrated-team-solution-new
```

The script selects every active Team/Solution entry from the current v2 catalog,
requires 42 designs and at least two changed body fields per alternate, and binds
source and changed specimens for every key. Each output directory must be new.
The reviewed slice 3 packet is an explicit provenance dependency, not an implicit
source-content fallback.

Stable family packet:
`samples/wmds-refresh-slice4-20261002/work/team-solution/`.

- `combined.foundation.json`: all 84 compiled specimens for integration.
- `bound-content.json`: exact source/alternate caller slots, keys and navigation.
- `generation-receipts.json`: every changed body slot with original/new values,
  key sets, content kinds, specimen origin, generation result and artifact hash.
- `catalog.json`: current closed contracts for the complete family.
- Per-template directories: paired foundation, bound content, binding/layout
  reports and editable paired deck.

Root owns the integrated native PowerPoint/PDF review and final publication.
