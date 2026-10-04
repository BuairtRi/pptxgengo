# Software revision history before cleanup

Retained successful build receipts show **12 preferred-main builds, 10 distinct semantic/PPTX versions, and 9 content revisions after the initial version**. Ignoring only the top-level `toolchain` field in the twelve compiled canonical source documents also gives **10 unique states**. The source byte hashes differ in11 states because the final JSON-to-readable-YAML formatting changed bytes while preserving semantics and the complete PPTX.

There are **32 successful build receipts overall**:12 main,4 case-merge fragment,5 dense-reflow fragment,1 source12 fragment,4 isolated diagnostics and6 builds of the two alternative layouts. Copies of exports, PowerPoint lock files and diagnostic PPTXs are not additional builds or errors.

## Native review and repairs

- Three full-deck native export rounds have PNG evidence:136-page v1,172-page v2a, and183-page final. The intermediate v2 folder contains queued diagnostic decks and no native PNGs.
- Five scoped repair review cycles are documented: dense25 initial and refined3-page recheck; root6-page fixes; topology/hierarchy7-page fixes; source12 two-page grid consolidation.
- Initial native review flagged **47 pages**:46 root findings and1 case quote/adjacent-delivery placement finding. Thirteen were internal guide prose overlap. Evidence: `review-v1/root-individual-native-review.json` and `case-study-fragment/review/native-review-merged-v1-072-120.json`.
- Follow-up172-page receipts contain11 semantic/hierarchy issue rows and12 root pending rows. These include repeated source34 chain findings, six pages for one source12 consolidation, and deferred dense replacements. **70 historical flag/pending observations across51 original source pages are not70 unique errors.** There is no central issue-ID ledger or complete record of failed transient edits, so an exact total defect count is unavailable.

Concrete corrections included number/guidance alignment on9/10, factor/icon placement on16, lifecycle phases and four-thread copy on22, source20/23 phase associations, source20 original Intellio logo, complete source25 icon constituents, source26 colored curve legend, source31 bidirectional expert feedback, source34 relationship overview and linked chain, source44/45 pipeline links and nested text hierarchy, dense10 originals expanded into25 measured continuations, source12 six-page overexpansion consolidated into2, and repeated authored notes added to source44/45 feature continuations. Each repair can affect several objects or pages.

Source16’s complete misalignment factor and source22’s full four-thread set are present in **actual XML of all12 retained successful main PPTXs**. Native v1 records layout/visibility defects; working-state copy gaps cannot be reconstructed as durable emitted-copy omissions. The JSON records the individual phrase checks.

## Items excluded from the error count

- Fourteen final decoded-pixel differences were individually reviewed and accepted; matching normalized XML/target parts and native inspection found no actual layout defect. Evidence: `review-final183/final-native-acceptance-183.json`.
- The formatter’s49 leading-newline strings were one serialization correction, not49 layout errors. That field count is root-reported; rejected working output was not retained. Final parsed semantics and entire PPTX bytes are identical: `review-final183/readable-yaml-semantic-proof.json`.
- Original editorial annotations, internal guides and `NEEDS CONTENT` placeholders were intentionally retained under the earlier source-faithful brief; they are not generated errors. The new design pass may reshape their treatment.
- Sixty graphics with custom semantic replacements are not60 missing assets. Their concrete dispositions are in `review/graphic-semantic-dispositions/semantic-dispositions.json`.

## Minimum retention recommendation

Preserve the immutable original/source inventory and external history archive; current `maintained/deck.yaml`, toolchain lock, context, ancestor snapshots and every declared/retention asset dependency; the latest final build receipt/source snapshots; final preferred PPTX and self-contained offline source ZIP; and the two separately requested alternative PPTXs/offline ZIPs with their acceptance receipts. Keep concise revision, copy/media, native, notes/hidden/sections and offline byte-proof receipts. Historical working decks, obsolete builds, duplicate PDFs/PNGs, fragment projects and one-off scripts can be archived outside the checkout before removal.

No source, delivery package or runtime was changed by this inventory. Counts and paths reflect the pre-cleanup retained evidence; the subsequently requested full Software redesign is a new pass.
