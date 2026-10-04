Updated October 4: final decks and visual reference PDFs are in the [samples root](<../../README.md>). Temporary decks, builds and archives were deleted; historical paths in receipts are retired.

# Source deck inventory

Read-only preparation for later sections and conversions. No conversion, rendering, visual acceptance or section implementation was performed. The three original PPTX files were read and hashed before and after extraction; hashes remained unchanged.

## Inputs

| Deck | Slides | Native sections | Hidden slides | Inventory |
| --- | ---: | ---: | ---: | --- |
| software-modernization | 83 | 12 | 2 | [JSON](software-modernization/inventory.json) · [Sidecars](software-modernization/index.md) |
| patterson-eaglesoft | 39 | 0 | 6 | [JSON](patterson-eaglesoft/inventory.json) · [Sidecars](patterson-eaglesoft/index.md) |
| dentalxchange | 57 | 7 | 15 | [JSON](dentalxchange/inventory.json) · [Sidecars](dentalxchange/index.md) |

## Preserved evidence

Each deck folder contains:

- `inventory.json`: exact OOXML text nodes and paragraph XML, speaker notes, layout names/types/text, shape records, relationships and structural review flags.
- `native-sections.json`: original names, IDs, ordered native slide IDs and mapped slide numbers.
- `ooxml-evidence/`: original bytes for presentation/section metadata, slides, slide relationships, notes, layouts and directly related chart/diagram parts.
- `sidecars/` and `index.md`: metadata-only output from the PowerPoint inspection skill.

OOXML ordering does not establish visual reading order. Exact speaker-notes extraction retains placeholder text as well as authored notes. Existing structural flags identify review candidates; no diagram semantics or text visibility were inferred. Images were not OCRed.

## Native sections

### software-modernization

- Intro: 1
- Modernization Context: 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12
- WM Modernization Perspective: 13, 14, 15, 16, 17, 18, 19, 20
- Lifecycle: 21, 22, 23, 24, 25, 26, 27
- Rationalize: 28, 29, 30, 31
- Build: 32, 33, 34, 35
- Sustain: 36, 37, 38
- Modernization Assets Portfolio: 39, 40, 41, 42
- Evolve: 43, 44, 45, 46, 47, 48, 49, 50
- Case Studies: 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75
- Off the Shelf: 76, 77, 78, 79, 80, 81, 82
- First Call: 83

### Structural review queue

| Slide | First authored paragraph (candidate label) | Shapes | Connectors | Groups | Tables | Charts |
| ---: | --- | ---: | ---: | ---: | ---: | ---: |
| 82 | © 2026 West Monroe Partners \| Reproduction and/or distribution without West Monroe Partners’ prior  | 232 | 0 | 0 | 0 | 0 |
| 79 | 79 | 104 | 0 | 0 | 0 | 0 |
| 81 | © 2026 West Monroe Partners \| Reproduction and/or distribution without West Monroe Partners’ prior  | 91 | 0 | 0 | 0 | 0 |
| 7 | CALLOUT HEADLINE | 26 | 5 | 6 | 0 | 0 |
| 20 | 	Modernization Execution & Sustainment	Timeline: 2-Week Sprints  	(duration depends on scope) | 32 | 2 | 6 | 0 | 0 |
| 34 | Persistent, agent-ready context from rationalization can compress the path from prototype to product | 59 | 0 | 0 | 0 | 0 |
| 25 | AI turns operational evidence into durable capability. | 56 | 0 | 0 | 0 | 0 |
| 22 | Modernization follows a standard lifecycle from strategy through sustained value. | 55 | 0 | 0 | 0 | 0 |
| 35 | AI can orchestrate phased cutover by sequencing capability slices, preparing each wave, and monitori | 52 | 0 | 0 | 0 | 0 |
| 6 | WEST MONROE MODERNIZATION IMPACT | 51 | 0 | 0 | 0 | 0 |
| 27 | © 2026 West Monroe Partners \| Reproduction and/or distribution without West Monroe Partners’ prior  | 51 | 0 | 0 | 1 | 0 |
| 12 | Example: quote-to-cash modernization requires untangling the full ecosystem | 49 | 0 | 0 | 0 | 0 |

### patterson-eaglesoft

No native PowerPoint sections stored.

### Structural review queue

| Slide | First authored paragraph (candidate label) | Shapes | Connectors | Groups | Tables | Charts |
| ---: | --- | ---: | ---: | ---: | ---: | ---: |
| 37 | © 2026 West Monroe Partners \| Reproduction and/or distribution without West Monroe Partners’ prior  | 60 | 40 | 12 | 1 | 0 |
| 19 | © 2026 West Monroe Partners \| Reproduction and/or distribution without West Monroe Partners’ prior  | 62 | 16 | 0 | 1 | 0 |
| 11 | TRANSITION ARCHITECTURE | 71 | 0 | 0 | 1 | 0 |
| 9 | © 2026 West Monroe Partners \| Reproduction and/or distribution without West Monroe Partners’ prior  | 30 | 9 | 2 | 1 | 0 |
| 8 | © 2026 West Monroe Partners \| Reproduction and/or distribution without West Monroe Partners’ prior  | 66 | 0 | 0 | 1 | 0 |
| 23 | West Monroe will scale delivery pods by workstream, working side by side with Eaglesoft and transiti | 65 | 0 | 0 | 0 | 0 |
| 5 | © 2026 West Monroe Partners \| Reproduction and/or distribution without West Monroe Partners’ prior  | 31 | 7 | 2 | 1 | 0 |
| 13 | Modernization follows a standard lifecycle from strategy through sustained value. | 55 | 0 | 0 | 0 | 0 |
| 24 | West Monroe will scale delivery pods by workstream, working side by side with Eaglesoft and transiti | 53 | 0 | 0 | 0 | 0 |
| 12 | MODERNIZATION METHOD | 44 | 0 | 0 | 1 | 0 |
| 14 | © 2026 West Monroe Partners \| Reproduction and/or distribution without West Monroe Partners’ prior  | 21 | 2 | 1 | 0 | 0 |
| 17 | Output: | 31 | 0 | 0 | 1 | 0 |

### dentalxchange

- Default Section: 1, 2, 3, 4
- Executive Summary: 5, 6, 7, 8
- Deliverables: 9
- Practices map: 10, 11, 12
- Activation summary: 13, 14, 15, 16
- Activation roadmap: 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41
- Graveyard: 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57

### Structural review queue

| Slide | First authored paragraph (candidate label) | Shapes | Connectors | Groups | Tables | Charts |
| ---: | --- | ---: | ---: | ---: | ---: | ---: |
| 3 | © 2026 West Monroe Partners \| Reproduction and/or distribution without West Monroe Partners’ prior  | 55 | 0 | 0 | 0 | 0 |
| 35 | © 2026 West Monroe Partners \| Reproduction and/or distribution without West Monroe Partners’ prior  | 54 | 0 | 0 | 0 | 0 |
| 34 | © 2026 West Monroe Partners \| Reproduction and/or distribution without West Monroe Partners’ prior  | 42 | 0 | 0 | 0 | 0 |
| 8 | © 2026 West Monroe Partners \| Reproduction and/or distribution without West Monroe Partners’ prior  | 40 | 0 | 0 | 0 | 0 |
| 6 | Executive summary | 38 | 0 | 0 | 0 | 0 |
| 14 | Activation Summary: from signal to implementation plan | 37 | 0 | 0 | 0 | 0 |
| 19 | First experiment: PRD to tech spec to release | 30 | 0 | 0 | 0 | 0 |
| 47 | Activation Summary: from signal to implementation plan | 30 | 0 | 0 | 0 | 0 |
| 40 | Strengthen Core testing and release controls | 27 | 0 | 0 | 0 | 0 |
| 38 | Product strategy spike | 25 | 0 | 0 | 0 | 0 |
| 39 | Scale Reconcile AI engineering across Core | 25 | 0 | 0 | 0 | 0 |
| 41 | Three parallel workstreams, one combined investment | 25 | 0 | 0 | 1 | 0 |

## Source SHA-256

| Deck | SHA-256 |
| --- | --- |
| software-modernization | `c732e693a6370a0abdf3b54f34ff456304a4a25025303c508a48559f1bee73e6` |
| patterson-eaglesoft | `a32b4e7d4715a7a7bfa420fa7c2c53092c58c71f97c0efa98d96adf8f90ac0e3` |
| dentalxchange | `4079af03947081cf446dce38025d441aa02c813c085b997b9464ce0de202d69c` |
