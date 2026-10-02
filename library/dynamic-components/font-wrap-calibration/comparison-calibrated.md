# Font layout reference set

276 cases. PowerPoint PDF references captured for all three families.

The PDF selections measure exported glyph geometry. Their height differs in meaning from PowerPoint TextRange bounds and Go line boxes.

| Family | Cases | Matching line breaks | Font warnings | Largest width delta | Largest PDF height delta |
| --- | ---: | ---: | ---: | ---: | ---: |
| IBM Plex Sans | 92 | 91 | 0 | 35.780 pt | 25.798 pt |
| IBM Plex Mono | 92 | 92 | 0 | 1.593 pt | 31.505 pt |
| Arial | 92 | 92 | 0 | 1.419 pt | 21.507 pt |

## Cases needing calibration

- `plex-sans-boundary-16.5-regular-plus-0`: Go 1 visible lines, PowerPoint 2. Font warnings: none.

Width differences include cases with different wrapping; they are not single-line advance errors. Blank lines are excluded from the visible-line comparison, but remain in the corpus and geometry. Bullet markers are excluded from line text comparisons. Font names come from PDF BaseFont resources; PDFKit selection font attributes can report fallback names in this process.

Full measurements and line text are in `comparison-calibrated.json`. Exact source contracts are in `corpus.json` and `cases.json`. No engine calibration was applied to this reference run.
