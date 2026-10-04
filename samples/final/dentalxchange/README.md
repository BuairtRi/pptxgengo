# DentalXChange — editable presentation source

This project contains 75 slides: the original 57 operating-model slides followed by 18 engineering findings/action-plan slides. 71 slides use unchanged shared v5 stock templates; four slides use three local definitions for the source's stage-coverage diagram, role-ownership matrices and editable release-history chart. The original 15 hidden states and authored notes are retained. The 18 appended slides are visible. Original PowerPoint files are unchanged.

[Template mapping and designer gaps](TEMPLATE-MAPPING.md) explains the four exceptions.

## What to edit

- `deck.yaml`: slide order, sections, asset declarations and relative file references.
- `slides/*.yaml`: one slide per file. Edit its `content` for visible wording, labels, tables, images or chart data. `bindings` connects those fields to the chosen layout; normally leave it unchanged. `template.scope: shared` and the template ID identify the actual stock layout.
- `notes/*.md`: speaker notes, original wording, source facts and qualifications. The matching slide's `notes_file` points here. Editing notes does not change visible copy.
- `context/*.md`: the slide's purpose and source/design brief, referenced by `brief`.
- `assets/`: original client images. Image content fields may use the declared asset ID; `project:ID` explicitly selects a project asset when an ID also exists in the library.
- `templates/*.yaml`: only the three local exception definitions. These hold placement and styling; business copy belongs in the slide's `content` fields.

The four exceptions are slide 8 (`stage-coverage`), 36 and 54 (`role-ownership`), and 45 (`activation-chart`). Their named `content` fields hold local copy. Slide 45's chart data is now in `content.release_history.categories` and `.series`; edit those arrays together. Preserve numeric zero as `0`, keep each series aligned with the categories, and keep category labels as strings. Chart placement, palette and workbook-preservation flags remain in its template definition.

For example, change a headline under `content.headline`, then update its Markdown notes if the source explanation changes. Do not change the binding target merely to change the words. A content group may have a component-based name such as `card_4`; its binding points to that stock template's slot. Slide IDs, reference paths and note filenames should remain stable when changing copy.

## Check and build

Run from this maintained project directory with the installed CLI:

```sh
pptxgengo design project check --project .
pptxgengo design project build --project .
pptxgengo design project measure --report builds/<new-build-id>/layout-report.json
```

The build reports its new immutable output directory. Keep the pinned `toolchain.lock.json`; a different compiler requires an explicit new pin. Measurement estimates fit; review actual rendered pages after changing content.

## Local PowerPoint PDF and PNG review

```sh
pptxgengo design render --pptx builds/<new-build-id>/deck.pptx \
  --out /path/to/new-review --pdf --png --include-hidden --timeout 10m
```

`--include-hidden` produces all 75 pages from a temporary inspection copy, leaving the source's hidden states intact. Existing output directories are refused. The command uses local Microsoft PowerPoint for PDF and macOS PDFKit for PNGs, with output hashes in `render-manifest.json`. It requires macOS, PowerPoint at `/Applications/Microsoft PowerPoint.app`, Swift tools and a logged-in GUI session with automation access. The isolated agent shell cannot reach that session and returns error −10827; actual native PowerPoint export from that shell is unqualified. The command reports failure rather than using another renderer. Copy closure on a failed or blocked Apple Event is best-effort.

## Final artifacts

- [PowerPoint](deck.pptx): 75 slides, original 15 hidden states retained.
- [Visual reference](visual-reference.pdf): all 75 slides visible for review.
- [Offline package](offline.zip): authored source and frozen runtime for reproducible builds.
- [Qualification](QUALIFICATION.md): native review hashes and checks.

Pinned CLI release: `0.1.0-local.11`. The maintained folder contains final artifacts only; working builds and review images are excluded.
