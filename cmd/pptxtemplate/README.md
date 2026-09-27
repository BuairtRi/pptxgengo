# Template implementation and review CLI

`pptxtemplate` operates the complete 65-item implementation queue. It uses the same
component validation/application engine as `pptxcomponent`, followed by the native
scene builder. It does not grant adaptation qualification or bypass `pptxlib`'s
qualified-contract gate.

```sh
go build -o /tmp/pptxscene-rollout ./cmd/pptxscene
go build -o /tmp/pptxtemplate-rollout ./cmd/pptxtemplate
python3 scripts/prepare-template-rollout.py
/tmp/pptxtemplate-rollout list
/tmp/pptxtemplate-rollout list --lane architecture_product
/tmp/pptxtemplate-rollout inspect --id t020-graphics-and-layouts-062
/tmp/pptxtemplate-rollout build-review --out samples/template-rollout-review
```

Run from the repository root, or use `--root`. Outputs must be new paths.
By default, every selected implementation must be complete and valid. `--available`
skips items with missing files for continuous review while other workers continue;
it does not skip malformed contracts or values. `--category` and `--lane` narrow
the selection. `--id` selects one template. `--values path.json` with `--id` supplies
new content instead of the illustrative example.

The bundle contains:

- A source-preserving PowerPoint for each represented source deck.
- Edited native scene projects with complete original dependencies.
- A change log per template, exact input hashes and an expected PDF page mapping.
- Explicit pending native fit/render/review status and retained-source disclosures.

Selected slides originally hidden are shown in the review copy; linked dependency
slides remain hidden. Footer numbers are frozen to their source reference numbers.
The source decks and extracted source projects are unchanged. Review decks may
retain original client/product artwork or facts and are not client-ready proposals.

Stage PowerPoint-facing copies directly in the already-approved
`samples/visual-wave3` folder. Use a distinct filename per immutable review version.
Export through the established native PowerPoint workflow and compare every PNG
with its source. No native renderer is invoked automatically by this CLI.

## Supported boundary

These implementations preserve fixed source geometry, rich-text segmentation and
item counts. Named content slots and explicit RGB style roles are editable. Native
charts, embedded diagrams, images and inherited/theme styles require separate
support where disclosed. Character counts are not measured fit limits.

`list` reports technical binding/value readiness, not editorial or visual acceptance.
`build-review` requires at least one actual binding change per selected template.
New content can overflow or conflict with retained art even when application succeeds.

## Component values check

`pptxcomponent check --project DIR --contract FILE --values FILE` performs the
same contract/value/binding checks as application without creating a project. It
reports requested and actually changed bindings and always reports fit unmeasured.
Values accept only `slots` and optional `profile`; descriptive metadata belongs in
the implementation record.

## Native review and gallery

Keep immutable versions so a visual review always points to its exact inputs.
PowerPoint exports run sequentially and reuse the approved working folder:

```sh
python3 scripts/render-template-review.py samples/template-rollout-review \
  --name template-review-unique-version
python3 scripts/build-template-review-gallery.py \
  --native samples/template-rollout-review/native-render.json \
  --out samples/template-review-gallery
```

Repeat `--native` for multiple workstreams. Later manifests supersede earlier
ones for the same template. The gallery checks source images, frozen inputs,
native decks/PDFs, PNGs and review hashes. It flags current authoring inputs that
have changed since rendering. Review ledgers live in
`library/templates/rollout/reviews/`; unreviewed renders remain pending.

The render script preserves open decks and never closes unsaved user documents.
It rejects output collisions; create a new bundle/version after a failed export.
The gallery includes source/adapted images, PowerPoint/PDF links, exact example
values, retained-asset disclosures and open findings.

The legacy West Monroe slide skill supplies historical reference only. Contracts,
CLI code and native evidence in this repository define this implementation.
