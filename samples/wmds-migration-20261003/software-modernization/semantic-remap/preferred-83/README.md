Updated October 4: final decks and visual reference PDFs are in the [samples root](<../../../../README.md>). Temporary decks, builds and archives were deleted; historical paths in receipts are retired.

# Software Modernization — preferred source

This maintained project maps the original 83 slides to 83 preferred slides. It uses the frozen 587-template WMDS v5 library: 42 shared slides and 41 local slides, with 35 reusable local definitions. The source remains an internal pick deck; draft placeholders and original commercial qualifications are preserved.

Start with `deck.yaml`. Visible copy and data are in slide values; geometry is in reusable local definitions. The 83 files in `context/` explain each page’s purpose. Original authored notes and complete original visible copy remain in native slide notes. Original hidden slides 7 and 20 and all 12 native sections are preserved.

`source-page-map.json` identifies the source page, purpose, selected recipe or actual parent, bound content zones, and verified source previews. `native-qualification.json` records the individual PowerPoint review of every preferred page and the strict artifact comparison used to carry acceptance into this cleaned source.

The separately delivered **Software Modernization - Visual Reference (83 pages).pdf** is the reviewed all-visible inspection v3 PDF, including the two hidden source pages. It was rendered by Microsoft PowerPoint from inspection PPTX SHA256 `cd9aba4934bf44f1434d1e0cd96e0ac3e932c740443ebcff50eefad4fcaa0d65`. It is retained as a visual reference, rather than described as a fresh PDF export of the metadata-cleaned preferred file. Every slide, layout, master, notes, and media part is byte-identical to the accepted preferred predecessor; only native section UUIDs follow the new canonical source identity.

The maintainer package includes the 38 registered asset dependencies and their declared derivation receipt, plus the source25 icon derivation evidence. The original PPTX remains untouched at `samples/software modernization campaign pick deck v1 - Repaired.pptx` in the workspace, with SHA256 `c732e693a6370a0abdf3b54f34ff456304a4a25025303c508a48559f1bee73e6`. Its 121 original media parts were byte verified. It is not duplicated inside this minimal package; the external migration archive contains generated history.

## Offline use

Extract the offline maintainer ZIP into a new directory. From that directory:

```sh
chmod +x runtime/pptxdesign
WMDS_BRANDING_ROOT="$PWD/runtime/branding" runtime/pptxdesign project check --project . --bundle runtime/library/wm-design-system/pinned
WMDS_BRANDING_ROOT="$PWD/runtime/branding" runtime/pptxdesign project build --project . --bundle runtime/library/wm-design-system/pinned
```

The runtime is the frozen Darwin arm64 v4 binary with SHA256 `370c08c471eb0dc13f6bcb9e57ca6c9d199a1aa467bd831ec9eaf849ff3b2e20`. The expected rebuilt PPTX SHA256 is `fb4b51e5e433a8cb574fd6f03167dccbc8ac193805ada5cc00fd2975141cbc3f`. The delivered ZIP is accompanied by `offline-rebuild-proof.json` and `delivery-manifest.json`. This package does not install or activate a skill. Native PowerPoint rendering uses the supplied fonts.
