Updated October 4: final decks and visual reference PDFs are in the [samples root](<../../../README.md>). Temporary decks, builds and archives were deleted; historical paths in receipts are retired.

# Patterson Eaglesoft — qualified semantic remap

Maintained source: `deck.yaml`. The 39 source pages, original order, six hidden flags, eight section groups, source assets and original notes are retained. Complete predecessor wording is appended to each page's notes. [PLAN.md](PLAN.md) records the slide-purpose and template choices.

This revision uses eight shared templates directly and twelve pages made from actual Go-scaffold parents. It changes visible wording and layout to communicate each page's argument. Draft pages are brief content-pending layouts; editorial details, imported client prices, roadmaps and backlog counts remain in notes.

## Review artifacts

- Preferred: `deliverables/Patterson-Eaglesoft-WMDS-semantic-preferred-39.pptx` — all 39 pages accepted in native PowerPoint; original six hidden flags retained.
- Inspection only: `deliverables/Patterson-Eaglesoft-WMDS-semantic-INSPECTION-all-39.pptx` and `.pdf` — all 39 pages visible. The PDF is for inspection and includes originally hidden pages.
- Maintainer source and offline runtime: `deliverables/Patterson-Eaglesoft-WMDS-semantic-preferred-maintainer.zip` and `deliverables/Patterson-Eaglesoft-WMDS-semantic-preferred-qualified-offline.zip` — current source, required assets/context/contracts, current build and qualification receipts. Offline rebuild is byte-identical to the preferred PPTX.
- Source-to-output mapping: `deliverables/Patterson-Eaglesoft-WMDS-semantic-page-map.tsv` and [PLAN.md](PLAN.md).
- Source checks: `planning/provenance-audit.json`.
- Compiler layout measurements: `planning/measure-review-v3.json` — advisory, not native acceptance.
- Native review chain: `native-review-v3/receipt.json` — all 39 accepted through 23 v1, 14 v2 and two v3 individual native reviews. Exact payload comparison carries unchanged pages; preferred/inspection package equality normalizes only hidden metadata and section IDs.
- Meaningful layout choices: `deliverables/Patterson-Eaglesoft-WMDS-4-alternative-layouts.pptx` — four alternative pages for sources 3, 11, 22 and 29, all accepted in native PowerPoint. Separate maintainer/offline ZIPs reproduce the accepted PPTX byte-for-byte; the alternative PDF is inspection only.

Both artifacts build with immutable `/tmp/pptxdesign-semantic-remap-v4`, using frozen `library/wm-design-system/v5`. The original deck and earlier qualified conversion remain unchanged.
