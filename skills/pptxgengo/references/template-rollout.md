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
and latest native review before selecting a design. Values contain only `slots`
and optional `profile`. Keep each slot's exact run count and paragraph membership.
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
text box fits. `pptxanchor` computes placements from native phrase measurements;
the template contract path does not yet apply those geometry changes. Keep such
examples marked as needing revision. Do not rewrite the user's approved copy just
to make a fixed accent appear correct.
