Updated October 4: final decks and visual reference PDFs are in the [samples root](<../../../README.md>). Temporary decks, builds and archives were deleted; historical paths in receipts are retired.

# DentalXChange semantic remap — final57

## Decks

| File | Purpose | SHA256 |
| --- | --- | --- |
| DentalXChange-WMDS-semantic-preferred-57.pptx | Preferred semantic layout mapping across all57 source pages | 96cead81d7d2878177724f9898022930cc09a307cbb8099d953e8d3f0ce15e7a |
| DentalXChange-WMDS-semantic-alternative-57.pptx | Same deck, with meaningful alternate context-and-stacked-card layouts on10/13/17 | de98bd120e3ef2b742bb77e3001b667d14b9d678538410384b2ac05ac4a6b96e |

Both decks preserve57 original slide identities,15 hidden slides42–56, visible57, seven original native section names/memberships, full original notes and source wording, and original assets. Visible copy is reshaped for the chosen purpose and geometry. No continuation slides were added. Source45 keeps the exact native editable chart, original category text,12 authored values and six numeric-zero workbook cells.

The original source PPTX remains unchanged, SHA256 `4079af03947081cf446dce38025d441aa02c813c085b997b9464ce0de202d69c`.

## Maintained sources

- Preferred: `../project-v2/deck.yaml`
- Alternative: `../alternative-v1/deck.yaml`
- Mapping: `../authored-mapping-v2.json`
- Authoring and CLI findings: `../authoring-friction.md`

Use the preserved frozen Go executable `/tmp/pptxdesign-semantic-remap-v4`, SHA256 `370c08c471eb0dc13f6bcb9e57ca6c9d199a1aa467bd831ec9eaf849ff3b2e20`, and the pinned `library/wm-design-system/v5` bundle. The offline packages include this executable, bundle, fonts and calibration; no local development checkout or new Python helper is required to rebuild them.

## Native qualification

Root individually reviewed all57 initial PowerPoint native pages. Seven issues were repaired in a new version: three legend mismatches, one clipped PHI qualification and three visible author-editing comments. All seven repaired pages passed individual native recheck. Preferred acceptance is50 exact native-part carries plus seven repaired-page native accepts.

The alternative's three changed pages10/13/17 were individually opened by root and independently by the child reviewer and passed. The remaining54 pages carry exact accepted native parts from the preferred chain. Receipts make this scope explicit; they do not claim every final PNG was reopened.

Original native review evidence is in `../qualification-v2/`: per-PNG/build/PDF hashes, exact XML/rels/notes carry, original source preservation, preferred/inspection equivalence and layout measurements. Those inspection PDFs are the actual PowerPoint exports of reviewed predecessor builds (`643d67a8…` preferred and `7a967153…` alternative); they are not new exports of the current source-cleaned artifacts.

`../qualification-v3-loader-cleanup/` qualifies the current artifacts by closed part comparison: each contains269 parts,268 byte-identical to its native-accepted predecessor. Only seven section UUID attribute values differ in `ppt/presentation.xml`; section names and full membership remain exact. Every slide, notes, media, chart and workbook payload is byte-identical. The cleanup removes seven unused empty content zones and stores all57 exact original brief strings in real `context/*.md` files. Business copy and the frozen runtime lock remain unchanged. Both projects now load with the current repository validation. Fresh relocated builds without Go reproduce each entire current PPTX and five additional maintained artifacts byte-for-byte.

## Packages

`../packages-v2/` contains six refreshed minimal packages: separate preferred and alternative client, maintainer and offline ZIPs with Go export manifests and SHA-bound receipts. Client ZIPs contain the PPTX; maintainer ZIPs contain current editable sources/context/assets/ancestors/build and native carry receipts; offline ZIPs also contain the frozen runtime and dependencies. Historical builds and one-off authoring helpers are excluded. Native inspection PDFs show all57 pages and belong to the reviewed predecessor builds.

Earlier sources/builds/review material are preserved for archive. There are no one-off authoring scripts in either maintained source project.
