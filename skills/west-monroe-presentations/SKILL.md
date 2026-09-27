---
name: west-monroe-presentations
description: Create or adapt West Monroe PowerPoint presentations with pptxgengo, using its packaged source templates, native scene edits, component library, or measured composition workflows. Use for West Monroe proposal, pursuit, and client presentation decks; it does not imply arbitrary-content template qualification.
---

# West Monroe presentations

Use the installed `pptxgengo` command. Do not assume repository-local Go binaries, mutable checkout paths, or the legacy `generate-west-monroe-slides` skill are installed. Use the packaged library and assets exposed by the command.

Choose the narrowest route:

- `pptxgengo template` to adapt supported text/style bindings in one of the 65 source-bound template contracts and produce a review deck.
- `pptxgengo compose` to author new editable slides from supported components, saved library contracts, or explicit composition specs. This supports genuinely new slides assembled from native editable components; it is bounded composition, not an unrestricted slide designer.
- `pptxgengo scene` to extract and rebuild source slides while preserving supported native structure/resources.
- `pptxgengo component` to inspect and apply reviewed source-bound text/color contracts in an extracted scene.
- `pptxgengo lib` to find, inspect, preview, instantiate, or assemble saved semantic library contracts. Qualification state matters; candidates are not approved templates.
- `pptxgengo anchor` to calculate measured placement for supported phrase accents. It emits placement data and does not independently author slide content.
- `pptxgengo diff` to compare two same-sized rendered PNGs; it writes a pixel-difference report and overlay, not a structural PPTX diff.
- Use `pptxgengo catalog` to inspect available packaged resources; `--open` and `--print` can expose package paths/content as supported.

If needed, consult `pptxgengo --help` and `pptxgengo <route> --help` for installed syntax. Read only the relevant route reference:

- [Workflow routes and limits](references/workflows.md)
- [Fixed-source template authoring](references/template-authoring.md)
- [New slide composition](references/compose-authoring.md)
- [West Monroe content and narrative guidance](references/content.md)
- [Component composition and visual review](references/composition-and-review.md)

## Essential constraints

- Keep evidence-backed claims, hypotheses, assumptions, and synthetic examples distinct. Do not invent client facts, metrics, commitments, or quotations.
- Write a decision-led story for the stated audience. Preserve substantive details, qualifications, ownership, dependencies, acceptance conditions, and evidence; resolve overflow by improving the argument, choosing a suitable composition, or splitting a slide.
- The 65 source-template contracts support bounded changes while preserving fixed source geometry and many source details. Their successful build does not show that arbitrary replacement content fits or that the original layout structurally adapts.
- Packaged library contracts and component examples have distinct qualification states. Do not imply general template approval from a catalog hit, a successful build, or a passing fixture.
- Measure and inspect final decks in PowerPoint at presentation size. A structural check or measurement alone does not establish visual quality.
