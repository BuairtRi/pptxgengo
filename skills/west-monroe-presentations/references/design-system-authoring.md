# Design-system one-slide routes

For a full deck, use the [project workflow](project-and-resume.md) and [template selection](template-selection.md). Use the direct routes below for single-slide work.

`pptxgengo design` uses the installed release's V11 design library, the Go typography engine `wmds-go-foundation.v2`, bundled fonts and the asset registry. Selected private media originals must be available separately ([assets](assets.md)). It needs no repository checkout. Resolve the bundle and gallery from `design_system_default` in `pptxgengo paths`.

## Find a design and inspect its contract

```sh
pptxgengo catalog --design-system --open
pptxgengo design library-catalog
pptxgengo design library-catalog --include-deprecated
pptxgengo design library-catalog --template-keys architecture/layers-nav
pptxgengo design asset-catalog
```

- The catalog lists each design's named slots, kinds, array counts, source identity and revision.
- The gallery hides deprecated designs and names their replacements.
- Choose a design that suits the argument and content density. Look at its preview before copying its files.
- A reviewed specimen does not show that your copy will fit. Build and render it.

## Fill a template

Copy `source-values.json` from a template's gallery folder (`catalog/design-system/<template-key>/` under `design_system_default`) (a one-slide `pptxgengo.wmds-template-document.v1` document). Edit the content, keep the template key, and build to a new directory:

```sh
pptxgengo design template --spec source-values.json --out ./wm-slide-new
```

- Set `content_kind: supplied_content` for real content; keep `synthetic_example` for illustrative copy.
- Supply every required slot and every fixed-array key. Missing content is not filled from the example.
- Keep value kinds exact: strings or rich text, numbers and booleans.
- Arrays have exact item counts and order. Stable keys preserve native identities.
- Content values never move or resize structure: table columns, diagram relationships, placement, base styles, frame geometry and artwork are fixed. Geometry, colors and font sizes are not content slots.
- Use `density`, `header_density` and `auto_density` on the slide object for
  supported typography changes, outside its `values`. See [typography density](typography-density.md).
- `cards/3` and `cards/4` take typed values: `eyebrow`, `title` and `cards` under `values`, with exactly three or four cards. Each card needs a unique `key`, `title` and `body`; array order sets card order.
- Navigation variants need 2–6 keyed labels and an `active` key.
- Photo and icon fields take registered asset IDs from `design asset-catalog`; verify their original bytes are available before rendering.
- Quadrant coordinates, table metrics and checkbox booleans follow their catalog kinds and bounds.
- Names, quotes and metrics in gallery example files are synthetic. Never present them as client facts.

## Build an illustrated composition

When the composition itself is the starting point, copy `source.foundation.json` from the same gallery folder (a one-slide `pptxgengo.wmds-foundation.v1` document with editable scene nodes) and build it:

```sh
pptxgengo design build --spec source.foundation.json --out ./wm-composition-new
```

Change only supported scene properties and keep source provenance. Filling a content contract does not generate a template's illustration from prose; the illustration lives in `source.foundation.json`.

Every build needs a new output directory. Template generation defaults to `native-v1`; explicit `--editing-profile stock` retains original object structure. The low-level foundation `build` route uses the profile declared in its document, so set `editing_profile: native-v1` there when desired. Review converted and retained components in the layout report.

## Typography and final review

- Layout runs in Go from bundled font files and outputs normal **IBM Plex Sans** and **IBM Plex Mono** names. No PowerPoint capture is needed per deck.
- A successful build does not guarantee native font selection or that the copy fits. Open the deck in PowerPoint at presentation size, and review a native PDF export before sharing when visual fidelity matters.
- Review automatic density changes and use supported whole-slide presets when
  they preserve reading quality. If overflow remains, edit the argument, choose
  a roomier template or split the slide. Do not shrink individual elements,
  truncate material copy or hide measured failures.
- Designs do not support variable row counts, automatic continuation or unrestricted layout beyond their contracts. Say so in delivery notes when it matters.

## Recent source capabilities

The gallery includes the workshop layouts and new pillar, branching-roadmap and
narrative-roadmap variants. Branching diagrams use editable native `roadfork`
shapes; use their stock content contracts rather than reconstructing their pins
or routes. Normal `road` diagrams remain available independently.

The two six-pillar Magenta variants use `numTile: "callout"`: a fixed 27 × 27 pt
Magenta square with a 14 pt IBM Plex Mono Semibold Grounded numeral and a 9 pt
gap to the title. The tile and numeral remain fixed across density tiers;
`numInk` is ignored when the tile is set. If deriving a scene card, retain its
inline numeral and title, with no title band, and preserve its allocation.
Inspect the current stock source and docs board for component arguments and
color rules; don't copy historical Magenta/White small-text treatments.
