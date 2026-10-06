# Assets: photos, icons, graphics and logos

Use only registered library assets, or images the operator supplies and you register in the project. Never pull images from the web or generate them.

## What the library has

| Kind | Count | Notes |
| --- | --- | --- |
| Icons | 222 concepts, each in navy, magenta and white | Line icons on a 42 × 42 grid |
| Photos | 523 catalog records | Includes all 521 originals under West Monroe Photos, with existing descriptive sidecars |
| Graphics | 12 | Brand marks: hand-drawn arrows, highlights, circles, underscores, sparks |
| Logos | 3 | West Monroe logo variants |

The installed asset gallery has 760 concepts and 1,204 variants. The SQLite index
includes the full photography inventory. Metadata comes from existing sidecars;
inspect thumbnails and selected originals before treating a description as proof
of visual suitability. When no photo fits, use approved operator-supplied imagery.

## Search

```sh
pptxgengo design library-find --kinds asset --asset-kind photo --query 'working session' --summary
pptxgengo design library-find --kinds asset --asset-kind icon --query risk --summary
```

- `--kinds asset` and `--summary` are both required for asset search. `--asset-kind` is `photo`, `icon`, `graphic`, `logo` or `all`.
- Results match any query word (with stems and a few synonyms) and rank by how many words matched; ties sort alphabetically. No match returns an empty list. Some concepts aren't tagged ("roadmap" finds nothing; try "strategy", "planning" or "route").
- Each result has `id`, `description`, `tags`, `people`, `industry`, `setting`, `orientation` and `variants`. Icon variants add `color` and `recommended_surfaces`; every variant has a `thumbnail_path`. Open the thumbnails before choosing.
- Icon descriptions are generic and tags come from filenames, so judge icons by their thumbnails.
- A library photo being registered doesn't record its license or approval. For client-facing decks, confirm photo use with the operator.

## Browse

```sh
pptxgengo catalog --assets --open
pptxgengo design asset-gallery --out ./gallery --kind icon --query people
```

`catalog --assets` opens the full packaged gallery. `asset-gallery` builds a filtered gallery in a new folder (`index.html`, `assets.json`, preview files). Use the registered originals in decks, never the gallery previews.

## Use an asset on a slide

- **Photos:** put the photo ID in the slide's photo field (`photo: photo-working-session`).
- **Icons:** icon fields take the concept name, not a color variant ID:

  ```yaml
  icon:
    name: risk-alert-arrow
  ```

  The template sets the icon's color for its surface. In local templates, use navy on light surfaces, white on dark surfaces, and magenta only where contrast holds (see each variant's `recommended_surfaces`).
- **Marks:** the title highlight comes from `[[…]]` in the title, not from placing the highlight graphic. Use at most one mark per slide.
- **Dividers:** `section add --divider-photo ID` takes a photo ID.

## Register a client or operator image

```sh
pptxgengo design project asset add --project ./client-deck --id client-logo \
  --file ./approved/client-logo.png --description "Client logo, supplied by operator" --focus 0.5,0.5
```

- PNG, JPEG or simple self-contained SVG, up to 64 MB. The original is copied into `assets/originals/` and recorded with its hash.
- `--description` is required. Describe what the image shows and where it came from.
- `--focus x,y` (0–1) records the point to keep in view when cropped; it is advisory.
- Builds shrink JPEGs to their placed size (220 ppi) and keep the original. PNGs aren't resized, so shrink large PNGs before registering.
- Use the ID in slide image fields. Use `project:client-logo` if it collides with a library asset ID.
- Record the source and permission for client images in `project.md`. Never use a client logo or photo without the operator's confirmation.

## Choosing imagery

- Photos show real people doing meaningful work, in natural light, relevant to the client's industry. Avoid staged handshakes and isolated devices.
- Abstract cube images suit covers, dividers and section openers, not evidence pages.
- An icon must clarify the item it sits beside. If no icon fits the idea, use a numbered or text-only treatment instead of a vaguely related icon.
- Record the asset chosen for each slide in its composition-log entry.
