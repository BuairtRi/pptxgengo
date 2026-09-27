# Full-shortlist template rollout

This checkpoint supersedes five-at-a-time implementation scheduling. All 65
candidates were assigned across three disjoint workstreams: architecture/product
(23), approach/narrative (21), and evidence/people (21). Luna handled most initial
binding work; Sol and root repaired difficult rich-text examples and reviewed
shared code. Native PowerPoint export remained serialized.

## Final review tally

58 of 65 examples are visually reviewed; seven need revision. The 65 contracts
expose 895 named text slots covering 1,570 text bindings. These are editing
contracts, not 65 templates qualified for arbitrary new content.

## Delivered

- 65 source-bound contracts with named text slots, realistic illustrative values,
  pinned source identities, preview hashes and retained-content disclosures.
- `pptxtemplate list`, `inspect`, and `build-review`, including one-template custom
  values and lane/category filters. The tool snapshots inputs, clones native
  scenes, records actual changes, preserves dependencies and builds review decks.
- Shared component engine and a no-write `pptxcomponent check` operation.
- Native render manifests with PDF page mapping and SHA-256 evidence; all native
  exports reused the already-approved `samples/visual-wave3` directory.
- Source/adaptation gallery with per-template QA states, open findings, immutable
  input links, deck/PDF links and stale authoring-input detection.
- Individual native visual review across the shortlist, with iterative corrections.
  Exact unchanged PNG hashes reuse earlier individual review; changed pages are
  inspected again. Historical failures remain in `library/templates/rollout/reviews`.
- Two bounded text-color variants demonstrated natively (t009 and t027). Their
  scope is selected text colors, not global semantic styling.
- Repository skill guidance for selecting, editing and reviewing these designs.
  The prior West Monroe slide skill is treated as a legacy reference asset.

## Review artifacts

Local gallery: `samples/template-expansion/rollout/gallery-v2/index.html`.
Its `index.json` is the machine-readable view. The committed
`library/templates/rollout/checkpoint.json` records the final counts, hashes,
latest evidence and open findings; individual ledgers retain the detailed history.

The latest full workstream bundles are architecture-v3, narrative-v2 and
evidence-v2 under `samples/template-expansion/rollout/review-*`. Single-template
follow-ups t007-v4 and t018-v3 supersede those entries. Two optional color proofs
are `review-style-t009-v1` and `review-style-t027-v1`.

Narrative t009/t027 gained profile definitions after the baseline content render.
The gallery consequently marks those baseline authoring contracts as changed;
the exact baseline snapshots and separately reviewed color variants remain linked
through their manifests. This is disclosed, not treated as a new render.

## Supported boundary

A reviewed example demonstrates the particular changed content and source
arrangement shown. These are fixed-geometry editing contracts. They do not yet
establish measured capacity envelopes, variable item counts, automatic reflow,
editable replacement of every retained diagram, or arbitrary-content qualification.
The `pptxlib` qualification gate remains intact; this review workflow does not
promote the 65 candidates into that qualified collection.

Some slides intentionally retain logos, source diagrams, chart data, photography,
contacts or product claims. The review decks are editing examples and require
content/asset review before use with a client. A retained complex architecture
image is not evidence of a newly composed editable architecture model.

## Open work, in priority order

1. **Measured accents in the template build path.** Five examples expose fixed
   accent failures: t002 underline, t004 highlight width, t014 title highlight,
   t048 phrase underline, t054 title highlight. `pptxanchor` already calculates
   placements from measured phrase bounds, but the contract path does not apply
   them. Bind accent intent to a phrase, measure after content replacement,
   transform the source asset with preserved stacking/aspect, and re-render.
   Use the user's annotated-placement fallback when the measurement is ambiguous.
2. **Theme-aware semantic styling.** t008's current-state legend is blue while the
   retained chart series is magenta. Existing roles accept explicit RGB bindings;
   these swatches use scheme colors. Add theme-aware role resolution without
   changing unrelated brand or series colors, then prove it in native output.
3. **Complete partial compositions.** t030 has only one editable populated speech
   bubble and two empty source bubbles. Add supported text-zone creation before
   describing it as a complete three-quote layout.
4. **Capacity and adaptation.** Measure content zones for representative dense
   templates, preserve paragraph/run style boundaries, and qualify bounded longer
   and shorter copy. Separate overflow handling from copywriting decisions.
5. **Reusable composition and narrative.** Connect shared component style roles
   across templates; expand existing arrow-family proofs; implement the proposed
   overview/detail sequences with stable phase/layer IDs and navigation.

The inventory milestone remains 369 source occurrences, 280 provisional families,
89 grouped repeats, and 65 shortlisted arrangements. Color variants are not new
templates. This rollout advances that shortlist into executable editing examples;
it does not finish the entire library/composition product.

## Checks performed

Go CLI builds, real contract inspections/value checks, real changed-content deck
builds, native PowerPoint PDF exports, and individual PNG visual inspection.
No test suite was added or run. Browser UI inspection remains unavailable in this
session; the gallery's artifact/hash validation was executed successfully.
