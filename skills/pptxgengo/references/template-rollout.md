# Adapt a template

Inspect the actual template key, content slots, source revision and preview before
changing copy. Keep fixed array counts and stable item keys. For a source-managed
deck, use readable slide YAML and the project's pinned bundle.

```sh
pptxgengo design project scaffold --stock --template cards/3 --id situation --out ./new-slide.yaml
pptxgengo design project slide add --project ./deck --file ./new-slide.yaml --as situation --check-fit
```

Update the composition log and mark real copy as supplied content. Review fit,
native output and visual hierarchy. New v2 projects use native-v1 where supported;
existing projects retain their profile and need deliberate changes/new baselines.

For a layout change, propose `project swap`, inspect mappings and missing fields,
then apply only a reviewed complete patch. Use `project detach` or a scaffolded
local derivative for supported source-scene templates; keep local definitions
in `slides/templates/`. Typed templates cannot be detached through that route.
Preserve required evidence/qualifications when changing form or splitting slides.

For fixed-geometry adaptations of older source scenes, use the repository-only
`pptxtemplate` route described in `docs/legacy-authoring/template-rollout.md`.
Inspect retained source content and exact qualification boundaries; a historical
specimen does not qualify arbitrary replacement copy or changed geometry.
