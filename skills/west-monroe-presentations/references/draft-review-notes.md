# Draft Review Notes

Use Draft Review Notes to track who is drafting or reviewing a slide, its status,
dates and iteration notes. Attach them to the slide source with `draft_review`;
they work with shared and local templates. No template fork or component
reconstruction is needed. Keep this mapping outside `content`, `values` and
`bindings`. It is separate from speaker `notes` / `notes_file`.

## YAML

Add this mapping to the slide's YAML file, or its inline object in `deck.yaml`:

```yaml
draft_review:
  status: wip
  status_text: Work in progress
  status_color: brand.blue
  owner: Ri
  due: 'Oct 20'
  updated: '2026-10-05'
  notes: |
    Confirm the source for the cost estimate.
    Ready for partner review after that change.
  placement: edge
```

Only `status` is required. Omit `status_text` and `status_color` to use the
selected status's default label and color:

| `status` | Default tab text | Default color reference |
| --- | --- | --- |
| `notstarted` | Not started | `kpi.off` (red) |
| `wip` | In progress | `kpi.risk` (yellow) |
| `complete` | Complete | `kpi.on` (green) |
| `qa` | QA'd | `brand.blue` (blue) |

- `status_text` changes the displayed tab label independently of the workflow
  status. Use a short single line, at most 128 bytes, that fits the tab.
- `status_color` changes the tab fill independently of status and text. It accepts
  the four palette references above. Text contrast is chosen automatically.
- `owner` appears in the header; `due` beside it; `updated` at the bottom. Dates
  are authored display strings, not calculated timestamps. Quote date values.
- `notes` is literal text, with line breaks supported; presentation `[[…]]`
  highlight marks are not interpreted.
- `placement: edge` is the default: an 18 pt status tab sits at the top-right
  slide edge, with the remaining card outside the slide on the pasteboard.
  `placement: pasteboard` puts the entire card outside the slide.

The card stays 234 × 144 pt. Builds reject copy that exceeds its fixed allocation;
shorten the field and keep longer context in the project's working documents.
Inspect the card in PowerPoint's editing view: slide PNGs and PDFs crop away the
off-slide body. The output is a native editable group named `wm-review/<slide-id>`
with matching slide metadata. Make durable edits in YAML and rebuild.

## CLI

Use the stable slide ID, not the slide number:

```sh
pptxgengo design project slide draft-review set --project ./client-deck --id findings \
  --status wip --status-text 'Work in progress' --status-color brand.blue \
  --owner Ri --due 'Oct 20' --updated '2026-10-05' --notes 'Confirm the source.'
pptxgengo design project slide draft-review show --project ./client-deck --id findings
```

`set` updates only supplied fields. A new note defaults to `notstarted` when no
status is supplied. Existing custom text and color survive status changes; reset
both overrides explicitly to display the new status's defaults:

```sh
pptxgengo design project slide draft-review set --project ./client-deck --id findings \
  --status complete --status-text '' --status-color ''
pptxgengo design project slide draft-review clear --project ./client-deck --id findings
```

Empty optional fields clear their value; `clear` removes the whole mapping.
`set`, `show` and `clear` return JSON receipts. Identical updates preserve source
bytes without a new decision. Notes survive project splitting and template swaps.
After changing them, run `project check` and `project build`; existing build and
review state becomes stale as for other source edits. If the installed CLI lacks
these commands, use the updated release and the agreed [toolchain migration](cli-reference.md#move-a-project-to-a-new-cli-version).

## Delivery

Normal builds and reviewer, maintainer and offline exports retain the card and
its metadata. Draft Review Notes are excluded from audience copy review.

For client delivery, use `project export --mode client --out NEW.zip` with
`--project PATH`. This removes the whole native review group, including the
on-slide status tab, and its private metadata from the packaged PPTX. The
immutable draft build remains intact. Cropping the card out of a PNG or PDF does
not remove private data from a draft PPTX; deliver the cleaned client export.
