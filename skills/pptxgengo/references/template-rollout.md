# Adapting a shortlisted source design

Use `pptxtemplate` when an existing arrangement suits the content and fixed item
counts are acceptable. This is the repository's implementation workflow. The old
West Monroe slide skill is a legacy reference asset.

From the repository root:

```sh
go build -o /tmp/pptxtemplate ./cmd/pptxtemplate
/tmp/pptxtemplate list --category technical_architecture
/tmp/pptxtemplate inspect --id t001-uhg-013
/tmp/pptxtemplate build-review --id t001-uhg-013 \
  --values /path/to/values.json --out /path/to/new-review-bundle
```

Inspect the contract, example values, source preview, retained-content disclosures
and latest native review before selecting a design. Values contain `slots`, optional `zones`, and optional `profile`. Keep each slot's exact run count and paragraph membership.
A rich-text paragraph can span several runs; preserve word boundaries and spaces
at style transitions. Superscript trademark runs must not receive ordinary prose.

The CLI snapshots inputs, applies changed values to cloned source scenes, builds
PowerPoints grouped by source deck and writes an expected PDF page map. Sources
and source dependencies remain unchanged. Outputs preserve source geometry, text
styles and fixed item counts. Native charts, diagrams, photography and client
identifiers may remain source content; these are disclosed editing examples.

Technical validity is reported separately from visual acceptance. A passed values
check does not establish that text fits. The reviewed example covers its exact
copy and selected style, not every possible value. Character counts are descriptive,
not measured limits. The broader `pptxlib instantiate` qualification gate is unchanged.

For native export and the source/adaptation gallery, follow
`cmd/pptxtemplate/README.md`. On this workstation reuse `samples/visual-wave3` for
PowerPoint-facing files, with distinct versioned filenames. Serialize exports.
Preserve open unsaved decks. Review ledgers pin the exact PNG; the gallery flags
changed authoring inputs and displays unresolved findings.

Fixed accents require a separate check after editing text. A highlight may cover
part of the next word and an underline may detach from its phrase even when every
text box fits. `pptxtemplate adapt-accents` measures the exact phrase, calls `pptxanchor`, and
applies the existing picture geometry in a new review bundle. See the CLI README
for intent rows, supported assets and explicit optical offsets. Phrases must match
the new copy; the illustrative intent file is not a semantic emphasis policy.
Keep unmeasured or visually unresolved examples marked as needing revision. Do not rewrite the user's approved copy just
to make a fixed accent appear correct.

For mixed-style text, prefer structured paragraph/run values with binding IDs.
`value_format: paragraphs` can require that format. Inspect exposes paragraph
membership and run style attributes; preserve emphasis and superscript semantics.
Theme roles declare `color_kind: scheme`; pin exact theme tokens instead of
replacing unrelated theme colors. Empty-shape zones may insert text or add an
upright native textbox with an explicit, source-bounded interior frame. These
capabilities are opt-in contract features, not automatic layout inference.

Run the native frame checker after changed copy, then inspect the rendered slide.
An inconclusive grouped/inherited frame is not a pass. Specific reviewed examples
do not qualify arbitrary lengths or variable item counts. Preserve failed review
versions and record the correction in an evidence-bound QA ledger.
