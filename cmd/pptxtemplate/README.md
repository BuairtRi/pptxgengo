# Template implementation and review CLI

`pptxtemplate` operates the registered source-template library (currently 101 contracts). It uses the same
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
the selection. `--id` selects one template; `--ids id1,id2` selects an exact set. `--values path.json` with `--id` supplies
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
item counts. Named content slots, structured paragraphs/runs, explicit RGB and theme-token
style roles are editable. Contracts can target exact, hash-pinned color nodes in
retained chart artwork and declare bounded text zones in empty shapes. Upright
text overlays support rotated freeform artwork without rotating its text. Native
chart data, arbitrary diagrams, images and inherited styles still require separate
support where disclosed. Character counts are not measured fit limits.

`list` reports technical binding/value readiness, not editorial or visual acceptance.
`build-review` normally requires an actual binding change. Explicit `--source-values` and newly imported source-reference examples permit unchanged content so the exact original layout can be reproduced.
New content can overflow or conflict with retained art even when application succeeds.

## Component values check

`pptxcomponent check --project DIR --contract FILE --values FILE` performs the
same contract/value/binding checks as application without creating a project. It
reports requested and actually changed bindings and always reports fit unmeasured.
Values accept `slots`, optional `zones`, and optional `profile`; descriptive metadata belongs in
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

## Measured accent pass

After content replacement, `adapt-accents` creates a new subset review bundle.
It measures the exact intended phrase in native PowerPoint, reads the embedded
artwork's visible alpha bounds, calls `pptxanchor`, and applies picture geometry.
It preserves text content and text-frame placement. The existing picture stays
embedded; highlights are ordered behind their target text. Underline image aspect
and source rotation are preserved; highlights fit their rotated visible bounds.

```sh
go build -o /tmp/pptxanchor ./cmd/pptxanchor
/tmp/pptxtemplate-rollout adapt-accents --bundle /path/to/review-bundle \
  --intents library/templates/rollout/accent-intents.json \
  --out /path/to/new-accent-review --name unique-accent-version \
  --anchor-bin /tmp/pptxanchor --scene-bin /tmp/pptxscene-rollout
python3 scripts/render-template-review.py /path/to/new-accent-review \
  --name unique-final-render
```

Intent rows identify template, text object, picture object, exact phrase, mode and
explicit optical offsets/padding. The shipped phrases target the illustrative
values; authors must select phrases for new content. The tool does not select
emphasis semantically. Native operations are serial and reuse an already-open,
hash-verified input deck when possible; new copies use `samples/visual-wave3`.

The adapter rebuilds the input project and checks it against the pinned deck,
then checks measured text/name/order/frame against scene and native OOXML. It
pins measurements, original artwork, solver output and resulting scene. Native
rotation missing from AppleScript is resolved only against matching OOXML.
Supported artwork: SVG, PNG and a narrow full-frame EMF+ bitmap extraction.
Cropped/flipped images, nested objects, rotated text and multiline phrases require
manual placement. Ambiguity produces an operator instruction and retains the
original artwork; visible on-slide parking/annotation is not implemented yet.
Every applied result still requires native rendering and visual review.

`apply-accent --evidence proof.json --out new-scene.json` is the lower-level Go
application step. It rejects changed evidence and requires the asset to be linked
by the measured picture in both scene and native deck.

## Native geometric fit

```sh
osascript scripts/measure-template-frames.applescript open-review.pptx > frames.json
python3 scripts/check-template-fit.py frames.json \
  --native-deck samples/visual-wave3/open-review.pptx --out fit.json
```

The checker reports `fits`, `overflow`, or `inconclusive` per measured text shape.
It verifies top-level native identity/text/geometry before using OOXML rotation,
compares text bounds with inner frames and records evidence hashes. Group/table
coordinate chains and inherited geometry remain inconclusive where unverified.
`--strict` returns nonzero for definite overflow. This check does not establish
copy capacity, detect every collision, or replace visual inspection.

For the limited case where PowerPoint can export an open task review but rejects
new opens, `scripts/render-open-accent-review.py` records a separate native geometry
replay. It checks picture-only changes against pinned scenes, uses unique names
through z-order changes, exports a PDF, and restores the task session. It does not
save or close the source presentation. It is restricted to established review
copies; unsaved sessions are refused by default. The explicit restored-task-session
option is only for this helper's own prior, verified restoration. This fallback
is visual geometry evidence, **not** proof that the new PPTX opens without repair.

## Explicit source selections and components

The 44 user-selected Lab/UHG layouts are indexed at
`library/templates/requested-templates.json`; source component occurrences and
executable subcontracts are at `library/templates/requested-components.json`.
Eight reuse existing IDs, 36 are new. These are source variants, not 101
independently qualified dynamic layouts.

```sh
go run ./cmd/pptxtemplate values --id t066-ai-accelerator-002 --source-values
go run ./cmd/pptxtemplate components --id t066-ai-accelerator-002
go run ./cmd/pptxtemplate build-review --ids t066-ai-accelerator-002,t077-uhg-005 --source-values --out /absolute/path/new-reference
```

`values --source-values` emits exact native run content. It preserves whitespace,
existing tabs and embedded line breaks. Replacements retain run cardinality. Within a run they can use plain text
without tabs/line breaks, or preserve the original control sequence; new or
reordered controls are rejected. Removing line breaks changes visual wrapping
and still requires native fit review. These are not measured capacity claims. Download
original values from the requested-layout gallery, edit only intended fields,
and supply them with `--values` for a single template.

For original T045 discrete gauges use `apply-gauge`, described in
`library/templates/rollout/evidence_people/t045-graphics-and-layouts-045/gauge-authoring.md`.
This changes the existing cells/pointer, preserving surrounding design. Generic
`adapt comparison` designs are separate, optional new compositions.
