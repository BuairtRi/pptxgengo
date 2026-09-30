# Component consolidation and sizing survey

Status: proposal for discussion. This survey does not change the installed local.4 release or migrate any source contracts.

## Survey findings

- **101 registered templates** form the template survey scope.
- **21 proposed building-block families** organize reuse, while retaining **31 historical retrieval labels** and the individual source variants. This is not a reduction to 21 components or templates.
- **132 requested source-group occurrences** have an explicit family/recipe crosswalk. Larger groups often contain several families. The historical 244 curated compositions are a separate, overlapping evidence set.
- **Sizing evidence:** 1,357 text-bearing source objects, 185 low-confidence drafting estimates, and 278 picture/media candidates with frame bounds. Table cells, effective font inheritance and optical artwork bounds remain explicit gaps.
- **Semantic styling is uneven:** 11/101 main contracts and 20/124 requested executable subcontracts expose color roles/profiles. The remaining contracts preserve source styling; common role names and behavior need consolidation.

Read the [family proposal and representative templates](families.md), [semantic styling proposal](semantic-colors.md), and [text/media sizing findings](sizing.md). Their adjacent JSON files preserve the detailed crosswalk and per-template evidence.

## What should be consolidated

Keep five levels separate:

1. **Primitives:** text, native shapes, images, rules and connectors.
2. **Component families:** independently meaningful, reusable parts with named content zones and controls—for example a metric, person profile, role tile, gauge or workstream lane.
3. **Variants:** compatible arrangements, density and appearance within a family. Color alone does not create a family. Portrait and landscape arrangements may require separate geometry contracts even when they share a content schema.
4. **Composition recipes:** arrangements of components, such as a case study, team hierarchy, phase-detail page or architecture map. Reusing a card does not make its entire parent slide a portable card.
5. **Slide templates and narrative sequences:** complete source designs and related overview/detail slides. Preserve their source identity and useful narrative tags.

Consolidate common schemas, style roles and layout rules first. Preserve source-specific geometry as named reference variants. Do not destroy a source variant merely because another design has similar text or color.

The historical 244 curated compositions and the newer 132 source component occurrences overlap and use different boundaries. They are separate evidence sets, not 376 unique components. Family membership is a proposal for reuse, not proof of interchangeable implementations.

## What an agent should be able to customize

These are proposed shared contracts, not claims that every source variant already supports each operation.

| Family | Content and structure controls | Styling and sizing controls |
|---|---|---|
| Text/list panel | Optional heading, paragraphs, bullet count and nesting | Density variant, padding, font hierarchy, declared surface/border/ink roles |
| Person profile | Portrait, name, role, short or detailed biography | Portrait frame/crop, separate name/role/body budgets, compact versus biography geometry |
| Gauge | Highlighted cell(s), independently chosen pointer cell; pointer defaults to sole selected cell | Existing gauge geometry, gray/navy/blue/pink selected color, distinct inactive color |
| Gantt lane + time axis | Workstream count, periods, activity start/end, status | One shared row center for label and bars, status legend, spacing and period width |
| Architecture layer + tile | Layer count, tile count per layer, layer labels | Shared row bounds and segmentation; cross-cutting rail has a separate divider and explicit scope |
| Table/evidence row | Row count, defined column schema, cell copy | Column widths, row heights, cell padding and budgets, header/body roles |

Structural controls belong at the smallest level that owns the relationships: a process step owns its copy and local shape; its parent flow owns step count, arrangement and connectors. A person profile does not own the spacing of an entire team roster. This makes shared behavior practical without erasing the source layout.

## Semantic styling

Resolve styles through **brand tokens → semantic roles → explicit component bindings**. Keep four choices independent:

- **Brand:** West Monroe Office first; additional approved brand packs can supply their own tokens.
- **Presentation:** neutral, subtle, inverse or restrained emphasis.
- **Information meaning:** status, ownership, data series, phase and selected rating. These mappings need their own labels/legends and must survive presentation restyling.
- **Source policy:** preserve a reference design, or deliberately adapt it to the selected brand rules.

A blue text rule and a blue completed-work bar can share a color without sharing a semantic role. A single global replacement must not change both. Client co-branding should be an explicit scope (e.g. client logo or data series), not an automatic replacement of West Monroe's entire presentation theme.

Current executable coverage is narrower than the palette inventory: 11 of 101 main template contracts and 20 of 124 requested component subcontracts declare color roles/profiles. The remaining contracts still preserve source styling. The adaptive builders and measured components have their own implemented style controls; a unified brand-token interface is proposed, not currently installed.

See [semantic color findings](semantic-colors.md) and [machine-readable styling proposal](semantic-colors.json).

## Text zones: guidance rather than a false hard limit

For each zone, keep the source text count separate from an estimated drafting budget. A mostly empty source box does not establish its capacity; a dense source box does not guarantee that different wording will fit.

A useful record contains:

- Template, source slide, object/zone ID and associated semantic slots.
- Width/height and usable area after insets; points and inches.
- Font family, size range, paragraph/list structure and inheritance confidence.
- Observed characters (including spaces) and paragraph count.
- An approximate drafting range where geometry and typography support one, with assumptions and confidence.
- Native fit evidence, when it exists, tied to the exact text/font/layout. Estimates should guide an agent before drafting, not replace the native fit pass.

Keep short label, heading, paragraph and bullet-list zones distinct. Reserve room for bullet indents and paragraph spacing. An unusually long word, changed font, added paragraph or extra line can invalidate a character estimate. Use warnings and suggestions to split or reframe content; do not silently shrink fonts or delete qualifications to satisfy a budget.

See [sizing findings](sizing.md) and [per-template sizing inventory](sizing.json).

## Image and icon slots

Treat the placed frame, the asset's aspect ratio and the visible artwork as separate measurements. Record crop/fit behavior, frame dimensions, rotation, source resource and whether an asset is a photograph, portrait, logo, illustration, icon or unresolved artwork. Native vector icons can be shape groups, not picture objects.

Use a small number of **family-specific size variants** after comparing the observed distribution. Keep square icons, portraits, wide photography, logos and deliverable thumbnails separate. A universal square icon size is inappropriate for tall or wide artwork. Preserve logo clear space and intrinsic proportions; use contain for logos/icons and explicit cover/crop for photography where appropriate. For raster assets, a future authoring hint can specify `recommended_pixels = placed_inches × chosen_output_ppi`; do not confuse source DPI metadata with the resolution available at the chosen placement size.

## Discussion and migration sequence

1. Review family boundaries and representative variants, including which larger groups are recipes.
2. Agree the role names and the default source-preservation policy; resolve current style drift without recoloring accepted source references automatically.
3. Attach sizing records to template/component discovery as advisory guidance. Keep unresolved typography, table cells and vector groups visible as gaps.
4. Implement shared families incrementally, retaining aliases and provenance for old IDs. Start with repeated panels/metrics, person profiles/role tiles, table rows and legends before larger team/architecture/roadmap compositions.
5. Qualify ordinary, dense, fewer-element and more-element examples. Only then promote a component from source editing to portable structural composition.

**Recommended discussion:** agree the reuse boundaries and brand-role model before migrating the library. The source designs remain available throughout consolidation.
