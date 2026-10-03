# Integrated Proof, Evidence and Argument reference

## Coverage

The integrated packet covers all 69 frozen v2 designs in these families:

| Family | Designs | Reused slice 3 pairs | New slice 4 pairs |
|---|---:|---:|---:|
| Proof | 25 | 12 | 13 |
| Evidence | 18 | 8 | 10 |
| Argument | 26 | 8 | 18 |
| Total | 69 | 28 | 41 |

`from-to/rows` is included once as an explicit retained deprecated compatibility
contract. Its recommended replacement remains `transformation/before-after`;
both keys receive separate source and caller specimens. This packet does not
restore the deprecated key to active status.

Each key has exactly two specimens: the preserved source copy and a meaningful
changed-content caller example. All remaining 41 caller examples use explicit
per-key content decisions. Existing reviewed slice 3 pairs are reused with their
accepted native illustrations and caller content. No generic string replacement,
source-example fallback, font reduction or geometry shrinking selects new copy.

## Content and native structure

All 41 new alternates change at least two substantive body slots. Of the 28
reused reviewed pairs, 27 change at least two; `image-text/frame-right-split`
retains its reviewed change to the accountable-review paragraph. Every key has
at least one substantive body change. Changes cover
narrative explanation, close-process steps, capability cells, comparison bullets,
case evidence, table requirements, chart values and facilitator guidance. Caller
content keeps exact scalar slot names, scalar types, array counts and stable key
counts; navigation retains explicit keyed labels and active selection.

Registry asset identifiers and source typography remain intact. Source examples,
quotes and numbers are preserved as frozen design illustrations, not endorsed
client evidence. New alternates use an illustrative eyebrow and, where the
source has one, a short explicit synthetic source note. Reused alternates retain
their reviewed synthetic indicators.

Thumbnail silhouettes become explicit editable miniature exhibits in composed
specimens, using the reviewed preview authoring functions. The annotated
scorecard reuses its reviewed native composition. These illustrative overlays
remain separate from the frozen binding contracts and never supply fixed example
fallback content to later callers.

## Fit decisions and limits

Three new content choices initially exceeded existing measured capacities:

- `capability-table/icon-rows`: the AI interview row now says, “Structure
  interviews and surface evidence gaps for expert validation.” It preserves the
  interview/evidence/validation relationship within the fixed table cell.
- `flow/challenges-to-outcomes`: the denial-prevention caption says, “Flags
  posting errors,” preserving the short card's single-line caption allocation.
- `pillars/four-why-matters`: the bullets say, “Work follows business value and
  delivery risk,” and “Rebuild only with purpose.” The fixed source cards retain
  their original fonts, outer boxes and bullet structure.

These are authored caller-copy decisions, not renderer fit fallbacks. Existing
closed table/card overflow checks reject insufficient allocations. The reviewed
57/75 pt row textblock capacities in both `what-we-did` variants remain enforced.
Native review subsequently identified two allocations that needed explicit guards; these are described below. Source fonts, outer geometry and calibration remain intact.

All 69 source fixtures and all 69 alternate specimens generated successfully:
138 slides total, with one retained deprecated compatibility key. The generated
PPTX and foundation document contain exactly two slides per catalog key.

Generation and native visual review are separate. A successful pair qualifies
only the exact sample for generation; it does not establish an arbitrary-copy
contract envelope. Native review viewed all 138 paired pages (113–250) in the integrated PowerPoint
PDF, including unchanged and deprecated compatibility examples. The first pass
accepted 136 specimens and found two alternate-copy allocation issues. After
regenerating all 69 pairs with the new guards, final native pages 119–120 and
243–244 were reviewed full size and accepted. Both issues are resolved. The
other 136 family page PNGs are byte-identical to the initial reviewed export.
All 138 exact paired specimens are accepted. No tests or negative controls were
added or run.

### Native findings and bounded allocations

- `case-studies/cards-quotes`, alternate page 120: a four-line paragraph pushed
  the primary metric label into its separately positioned platform caption. The
  revised caller paragraph preserves the report count, model migration,
  automation and eight-month result in three lines. Named amendment
  `wmds.cards-quotes-platform-clearance.v2` bounds flowing content in both
  narrative cards at absolute y315, reserving 9pt before platform captions at
  y324. The outer card remains y126/h234, with original typography and widths.
- `transformation/pain-to-theme`, alternate page 244: the second caption wrapped
  below its fixed shaded row. The revised caller caption says, “Daily matching
  routes exceptions to an accountable owner.” All four row captions receive h18
  measured allocations, and their headings h24, preserving the authored source
  positions and fonts. Named amendment
  `wmds.pain-to-theme-fixed-text-capacity.v2` applies these allocations. The
  existing scene text height guard rejects text beyond
  these allocations.

`native-review.json` records per-key source/alternate dispositions, exact native
PNG hashes, preserved initial contact-sheet evidence and findings, and resolved
final rechecks. The final PDF SHA256 is
`27395b8b6dedf46e3505f1067afb32cc6a0884047c9fe29e79ed5ed591b9d153`. This
review applies to the paired specimens only; future callers must satisfy the
measured capacities and still receive review appropriate to their use.

## Reproduction

Use a new output directory with the current repository CLI:

```sh
python3 scripts/wmds-integrated-proof-evidence-argument.py \
  --cli /tmp/wmds-slice3-root-pptxdesign \
  --out /tmp/wmds-integrated-proof-evidence-argument-new
```

Canonical artifacts:
`samples/wmds-refresh-slice4-20261002/work/proof-evidence-argument/`.

- `combined.foundation.json`: 138 paired specimens, two per key.
- `bound-content.json`: all 138 exact source and alternate caller inputs, in the
  same order and with the same IDs as the foundation document, including both
  deprecated compatibility specimens.
- `contracts.json`: frozen v2 definitions used to project caller content.
- `generation-receipts.json`: exact changed slots and body slots, exact key-array
  counts, source/alternate kinds, lifecycle, reuse origin and generation status.
- `source-generation-receipts.json`: all 69 original source generation results.
- `reference.pptx` and `layout-report.json`: editable family reference packet.
- Per-key folders retain two-slide documents, source/alternate caller inputs and generation
  receipts for reproducible review and corrections.
