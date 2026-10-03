---
name: west-monroe-presentations
description: Create or adapt West Monroe PowerPoint presentations with pptxgengo, using its packaged source templates, native scene edits, component library, or measured composition workflows. Use for West Monroe proposal, pursuit, and client presentation decks; it does not imply arbitrary-content template qualification.
---

# West Monroe presentations

Use the installed `pptxgengo` command. Do not assume repository-local Go binaries, mutable checkout paths, or the legacy `generate-west-monroe-slides` skill are installed. Use the packaged library and assets exposed by the command.

Choose the narrowest route:

- `pptxgengo design` to use the modern WM design system: fill a closed template content contract or build an editable foundation composition. This packaged route defaults to the bundled v2 library and Go engine `wmds-go-foundation.v2`; select `--bundle v1` explicitly for the earlier design library. Browse with `pptxgengo catalog --design-system --open` and read the [design-system authoring reference](references/design-system-authoring.md) for this route.
- `pptxgengo template` to adapt supported text/style bindings in an available source-bound template contract and produce a review deck.
- `pptxgengo adapt` to generate editable roadmap, architecture, process, team, or comparison compositions from semantic content and variable counts. Read the [adaptive family reference](references/adaptive-authoring.md) only for this route.
- `pptxgengo compose` to author new editable slides from supported components, saved library contracts, or explicit composition specs. This supports genuinely new slides assembled from native editable components; it is bounded composition, not an unrestricted slide designer.
- `pptxgengo scene` to extract and rebuild source slides while preserving supported native structure/resources.
- `pptxgengo component` to inspect and apply reviewed source-bound text/color contracts in an extracted scene.
- `pptxgengo lib` to find, inspect, preview, instantiate, or assemble saved semantic library contracts. Qualification state matters; candidates are not approved templates.
- `pptxgengo anchor` to calculate measured placement for supported phrase accents. It emits placement data and does not independently author slide content.
- `pptxgengo diff` to compare two same-sized rendered PNGs; it writes a pixel-difference report and overlay, not a structural PPTX diff.
- Use `pptxgengo catalog --templates` or `pptxgengo catalog --components` to print the matching packaged gallery path; add `--open` to open it. `pptxgengo paths` reports the installed gallery/resource paths.

If needed, consult `pptxgengo --help` and `pptxgengo <route> --help` for installed syntax. Read only the relevant route reference:

- [Design-system authoring](references/design-system-authoring.md), for the packaged modern WM library
- [Workflow routes and limits](references/workflows.md)
- [Fixed-source template authoring](references/template-authoring.md)
- [Component and template customization](references/component-customization.md), when a request includes palette, gauge, or accent choices
- [New slide composition](references/compose-authoring.md), when selecting between supported editable recipes
- [West Monroe content and narrative guidance](references/content.md)
- [Component composition and visual review](references/composition-and-review.md)

## Essential constraints

- Keep evidence-backed claims, hypotheses, assumptions, and synthetic examples distinct. Do not invent client facts, metrics, commitments, or quotations.
- Write a decision-led story for the stated audience. Preserve substantive details, qualifications, ownership, dependencies, acceptance conditions, and evidence; resolve overflow by improving the argument, choosing a suitable composition, or splitting a slide.
- Source-template contracts support bounded changes while preserving fixed source geometry and many source details. Their successful build does not show that arbitrary replacement content fits or that the original layout structurally adapts.
- Packaged library contracts and component examples have distinct qualification states. Do not imply general template approval from a catalog hit, a successful build, or a passing fixture.
- Normal `design` builds measure and lay out text in Go using bundled font metrics. Output retains normal IBM Plex Sans and IBM Plex Mono family names; no per-deck native measurement capture or font alias is required. The 167-design catalog has 166 active designs and one deprecated design; paired specimen review is not arbitrary-copy qualification.
- Measure and inspect final decks in PowerPoint at presentation size. A structural check or measurement alone does not establish visual quality.
