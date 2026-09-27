# Template discovery and release queue

This catalog expands coverage across the full 369-slide source corpus. It records
source designs, their business purpose, altitude, density and reusable arrangements.
It does **not** turn every inspected slide into an executable template.

- `taxonomy.json`: independent category, altitude and density definitions.
- `reviews/`: isolated source review slices with image hashes and review status.
- `catalog.json`: compiled families and complete source occurrence coverage.
- `release-shortlist.json`: 65 initial implementation candidates, not qualified designs.
- `sequence-patterns.json`: proposed overview/detail narrative patterns; automatic
  sequence expansion is not implemented.
- `REVIEW_PROTOCOL.md`: visual inspection and grouping rules.
- `curation-decisions.json`: root visual-review corrections with preview hashes.
- `rollout/`: all 65 source-bound editing contracts, illustrative values, parallel
  assignments, inspection reports and hash-pinned native review ledgers.
- `batch-01/`: five inspected source-bound editing contracts; native adaptation
  and semantic styling remain pending.

`business_architecture`, `technical_architecture`, `product_overview`,
`product_detail`, `layers_components` and `framing_navigation` are explicit categories.
A family's categories/altitudes/densities preserve the observed uses of its members.
A color or text change does not create another template.

Compile only after workers have completed their records:

```sh
python3 scripts/compile-template-catalog.py
```

The compiler validates source coverage, preview hashes, known duplicate decisions,
and every family definition. Cross-slice mergers are explicit decisions, not fuzzy
text matching. Metadata-only reviews remain visible and cannot qualify adaptations.
The generated gallery is `samples/template-expansion/gallery.html`.

Rebuild the Go library index to include the current catalog and sequence patterns:

```sh
go build -o /tmp/pptxlib-template-expansion ./cmd/pptxlib
/tmp/pptxlib-template-expansion index --out samples/template-expansion/library.sqlite
/tmp/pptxlib-template-expansion find --index samples/template-expansion/library.sqlite --inventory --kind template_family --category technical_architecture --altitude overview
/tmp/pptxlib-template-expansion find --index samples/template-expansion/library.sqlite --inventory --kind template_family --density dense
/tmp/pptxlib-template-expansion find --index samples/template-expansion/library.sqlite --inventory --kind narrative_sequence --query phase
/tmp/pptxlib-template-expansion inspect --index samples/template-expansion/library.sqlite --inventory --id sequence:phase-expansion
```

Use returned family IDs with `inspect --inventory` and `preview --inventory`.
Preview resolves and validates the pinned source PNG. Family and sequence records
remain outside the executable-contract table. `instantiate` cannot use them or
silently bypass native adaptation requirements. Changing review inputs invalidates
the compiled catalog; changing the catalog invalidates its SQLite projection.

## Current milestone

All 369 source occurrences have visual inspection records. They currently form
280 provisional classification families, with 89 repeated occurrences grouped.
Further cross-family deduplication remains possible. The implementation shortlist
contains 65 arrangements. All now have source-bound editing implementations and
native rendered examples; their individual acceptance states and open findings
are recorded separately. This is not 65 arbitrary-content qualified templates.
The separate existing library still has 13 candidate executable contracts.

The gallery uses local source PNGs. Compilation validates their exact hashes, so
missing local render artifacts must be restored from the matching native exports
before regenerating it. The browser UI could not be visually inspected in the
current session because the computer-use browser connection was unavailable.


## Changed-content review workflow

Use [`pptxtemplate`](../../cmd/pptxtemplate/README.md) to inspect a contract and
build a source-preserving review bundle from illustrative or custom values.
See the [rollout checkpoint](../../planning/TEMPLATE_ROLLOUT_CHECKPOINT.md) for
latest artifacts, support boundaries, corrected QA failures and remaining work.
The local source/adaptation gallery is
`samples/template-expansion/rollout/gallery-v2/index.html`.
