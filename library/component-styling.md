# Component families, variants and semantic styles

## Retrieval model

The expanded library distinguishes three levels:

1. **Family** — a common purpose and content relationship: metric panel, numbered
   card, process step, person identity, delivery pod, comparison row, and so on.
2. **Source pattern / structural variant** — a specific arrangement, slot structure
   and cardinality. A two-role pod and a three-role pod share a family but require
   different bindings. A circle above a caption differs from a side-tab card.
3. **Source composition** — exact object membership within a hash-pinned slide.
   Its original text, paragraph XML, geometry and explicit colors remain inspectable.

Repeated source examples do not increase the canonical family count. Exact
same-source selections are removed before enrichment; aliases retain alternative
annotations. Similar designs from different slides remain evidence variants until
we prove that one editable implementation can reproduce both. Native PowerPoint
groups are retained separately, because a group is not necessarily a reusable unit.

The taxonomy is in [component-taxonomy.json](component-taxonomy.json). It defines
semantic retrieval families, proposed color profiles and promotion requirements.
The compiler emits the populated family inventory and pins its inputs.

## Color roles

| Role | Meaning | Default |
|---|---|---|
| `surface` | Component background | White |
| `text.primary` | Main content | Grounded Blue `#070154` |
| `text.secondary` | Supporting copy | Dark Gray `#50658E` |
| `accent` | Selected emphasis such as a heading or rule | Grounded Blue |
| `border` | Quiet component boundary | Medium Gray `#CED7E6` |

Six proposed profiles cover neutral, subtle, inverse, blue accent, magenta accent
and headline highlight. They derive from the local branding `DESIGN.md`; its hash
is recorded with the contracts. These are proposed library APIs, not a claim that
all source examples follow current brand guidance or that WM has approved them.

Yellow is restricted to the marker behind one to four headline words. Highlight
Blue is a text/rule accent, not a general surface. Large magenta emphasis needs
contrast review. Source-specific data-series, actor, phase, status and intensity
colors require separate semantic bindings; recoloring them can change meaning.
Office typography defaults to Arial. Exact reconstruction retains source fonts.
Digital type sizes and spacing in the guidance are not slide measurements.

## Application boundary

The catalog records explicit `srgbClr`, `schemeClr`, preset/system and other color
expressions with their OOXML location and transforms. A scheme expression is not
a resolved RGB color. Pictures may contain baked-in colors that this metadata
cannot recolor. Inherited styles and effective typography remain unresolved.

Applying a profile will require an explicit object or text-run binding:

```json
{
  "component_family": "component-family:metric-panel",
  "profile": "subtle",
  "role_bindings": [
    {"object_path": "<verified background>", "property": "fill", "role": "surface"},
    {"object_path": "<verified metric>", "property": "text", "role": "text.primary"}
  ]
}
```

This is an illustrative contract, not an executable authoring command. Global hex
replacement is unsafe: a source color may appear in body text, a chart legend,
a logo, and an accent. Mixed rich-text runs also require finer targeting than the
text shape as a whole. Preserve alpha/tint transforms and intentional layering.

## Next implementation gate

Start with metric panels, numbered cards and delivery pods. For each:

- Choose a clean source variant and bind every semantic slot and presentation role.
- Resolve dependencies, effective typography, text insets and meaningful anchors.
- Prove typical and boundary content edits with native PowerPoint rendering.
- Exercise neutral, subtle and inverse styling without changing data encodings.
- Record supported operations and measured capacity; fail with a useful fallback
  for unsupported cardinality, overflow or ambiguous object roles.

Then compose a new slide from approved components on the slide grid. The current
catalog is broader and searchable, but no candidate is yet adaptation-approved.

## 2026-09-26 implementation update

The first [reference shortlist](reference-variants.md) and four
[source-bound text contracts](component-contracts/) are available through
`pptxcomponent inspect|apply`. These fixtures preserve source styling. No style
profile is enabled in the four contracts: the source objects include inherited
fills, theme expressions and typography that need explicit resolution before
neutral/subtle/inverse variants can be safely applied.
