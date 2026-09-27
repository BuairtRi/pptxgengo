# Complex proposal native review — 2026-09-27

## Deliverable and status

- [13-slide PowerPoint review candidate](../samples/visual-wave3/complex-proposal-review-v2.pptx)
- [Native PowerPoint PDF](../samples/visual-wave3/complex-proposal-review-v2.pdf)
- [Rendered pages](../samples/visual-wave3/complex-proposal-review-v2-render)
- [Current assembly trace](../samples/proposal-authoring/full-capacity-v13/assembly-trace.json)

This is a review candidate, not a fully qualified library release. The user asked
for a wrap within 40 minutes, ending 12:22:37 UTC. Native measurement, fit,
rendering, corrections and visual review were completed within that window.
Final native `verify` and expanded/stress native qualification remain unfinished.
No contract was promoted to adaptation-qualified. Wave 5 remains deferred.

## Complexity exercised

13 pages cover a dense argument with phrase highlight, situation matrix,
five paired needs/responses, journey matrix, five-phase plan, detailed discovery
activities and outputs, nine-node architecture with six dependencies, two parallel
process paths, 20-week Gantt, 21-role organization with three pods, 19-tile roster,
rich biography and release-evidence matrix. Copy is fictional and explicitly
qualified; the biography is illustrative. These are adapted proposals, not
pixel-perfect reproductions of the original UHG/EnableComp decks.

## Evidence

- Initial native measurement: 451 requests, 23,902 characters, 24m15s.
- A subsequent changed-text request was measured separately and cached.
- Current fit: **437 fixed zones, zero overflows, zero layout failures**; planner passes.
- PowerPoint exported all 13 pages to PDF; root viewed every initial page.
- Final pages 3, 7 and 10 were re-rendered and reviewed by root and two Luna reviewers.
- The other ten final PNGs are byte-identical to the reviewed first render.
- Full `go test ./... -count=1` passes.
- Current stress configuration assembles twice identically (13 individual specs
  and combined spec); four rejection cases pass. Assembly times: 1.686s, 1.684s.
- Final asset-metadata refresh yields an identical composition spec:
  `3a1fbd007edc4aa296d6b41ed3b608fab3a6668c5ed68e256927ba64ad3ebefa`.

Artifact hashes and native environment are retained in
[the implementation checkpoint](WAVES_3_4_IMPLEMENTATION_CHECKPOINT.json).
The native text cache is tied to the frozen executable
`/tmp/pptxcompose-native-fixed` and its adapter/font/PowerPoint environment.
Do not rebuild that executable and assume its measurements remain reusable.

## Corrections made

1. Enlarged two team text frames by 2pt after native measurement caught 0.4pt
   and 0.2pt overflows. No font reduction or copy deletion.
2. Enlarged response row backgrounds to contain the actual text, preserving
   gutters. Text-box fit alone had missed this visual defect.
3. Moved the architecture parent surface behind its six connectors. All six
   are visible between cards in the final render.
4. Replaced inherited mismatched icons with ownership, documentation and
   organizational-change artwork. Source URLs and hashes are in
   `library/proposal/asset-bindings.json`.
5. Clarified that staffing colors are illustrative. Role counts do not imply
   21 full-time people or a named staffing commitment.

QA115–QA118 in [the QA log](QA_ERRORS.md) capture these findings, including
initial independent reviews missing the panel/connector defects. Future optical
reviews must count expected relationships and inspect actual panel boundaries.
The team legend fits but remains close to the footer; the first two response
rows repeat a people icon. Neither was judged a blocking defect in this review.

## Recovered PowerPoint workflow

The user had already granted access. Staging directly inside the approved
`samples/visual-wave3` directory restored native open/export. New nested working
folders had caused repeated permission prompts. `measure` and `verify` now
support `--native-workspace`; staging preserves byte identity and rejects unsafe
collisions. No app restart, user-deck closure or unsaved-deck overwrite was needed.

## Remaining gates

1. Run final native verification of the saved candidate using the same approved
   workspace, preserving the immutable bundle and evidence hashes.
2. Native-measure/render/review the refreshed 15-page expanded deck, 29 accent
   and arrow cases, and source/art controls. Updated inputs are prepared but
   their earlier probes are historical and need regeneration.
3. Native-measure/render the increased-density/cardinality stress deck; complete
   the independent agent's forward-check render.
4. Promote individual contracts only when their complete evidence supports it.
   The eight changed capacity contracts are version 0.2.3 candidates.

Generated PowerPoint/PDF/PNG files and local downloaded artwork remain ignored.
Code, portable templates, semantic values, binding metadata and these records are
committed. A fresh clone needs its ignored source asset inventory hydrated; this
review does not establish a self-contained asset distribution package.
