# Font layout reference set

117 cases. PowerPoint PDF references captured for all three families.

Go engine: `go-text-prototype.v2`. Font provenance matches the dataset's baseline: True.

The PDF selections measure exported glyph geometry. Their height differs in meaning from PowerPoint TextRange bounds and Go line boxes.

| Family | Cases | Matching line breaks | Font warnings | Largest width delta | Largest PDF height delta |
| --- | ---: | ---: | ---: | ---: | ---: |
| IBM Plex Sans | 39 | 39 | 0 | 0.839 pt | 11.183 pt |
| IBM Plex Mono | 39 | 39 | 0 | 0.972 pt | 11.183 pt |
| Arial | 39 | 38 | 0 | 65.127 pt | 30.111 pt |

## Cases needing calibration

- `arial-heldout-26-regular-boundary-0.2`: Go 1 visible lines, PowerPoint 2. Font warnings: none.

Width differences include cases with different wrapping; they are not single-line advance errors. Blank lines are excluded from the visible-line comparison, but remain in the corpus and geometry. Bullet markers are excluded from line text comparisons. Font names come from PDF BaseFont resources; PDFKit selection font attributes can report fallback names in this process.

Full measurements and line text are in `comparison-final.json`. Exact source contracts are in the dataset's `corpus.json` and `cases.json`. The comparator does not modify measurements or calibrate the engine.
