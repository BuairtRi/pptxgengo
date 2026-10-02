# Font layout reference set

276 cases. PowerPoint PDF references captured for all three families.

The PDF selections measure exported glyph geometry. Their height differs in meaning from PowerPoint TextRange bounds and Go line boxes.

| Family | Cases | Matching line breaks | Font warnings | Largest width delta | Largest PDF height delta |
| --- | ---: | ---: | ---: | ---: | ---: |
| IBM Plex Sans | 92 | 87 | 0 | 55.347 pt | 29.053 pt |
| IBM Plex Mono | 92 | 80 | 0 | 71.360 pt | 43.112 pt |
| Arial | 92 | 85 | 0 | 51.920 pt | 28.239 pt |

## Cases needing calibration

- `plex-sans-long-word`: Go 4 visible lines, PowerPoint 4. Font warnings: none.
- `plex-mono-wrap-24-360-regular`: Go 7 visible lines, PowerPoint 6. Font warnings: none.
- `plex-mono-wrap-24-360-bold`: Go 7 visible lines, PowerPoint 6. Font warnings: none.
- `plex-mono-alignment-left`: Go 3 visible lines, PowerPoint 3. Font warnings: none.
- `plex-mono-alignment-center`: Go 3 visible lines, PowerPoint 3. Font warnings: none.
- `plex-mono-alignment-right`: Go 3 visible lines, PowerPoint 3. Font warnings: none.
- `arial-wrap-12-180-bold`: Go 5 visible lines, PowerPoint 5. Font warnings: none.
- `arial-wrap-24-180-bold`: Go 11 visible lines, PowerPoint 11. Font warnings: none.
- `arial-long-word`: Go 4 visible lines, PowerPoint 4. Font warnings: none.
- `plex-sans-boundary-16-regular-minus-0.15`: Go 2 visible lines, PowerPoint 1. Font warnings: none.
- `plex-sans-boundary-16-bold-plus-0`: Go 1 visible lines, PowerPoint 2. Font warnings: none.
- `plex-sans-boundary-16.5-regular-plus-0`: Go 1 visible lines, PowerPoint 2. Font warnings: none.
- `plex-sans-boundary-24-bold-plus-0`: Go 1 visible lines, PowerPoint 2. Font warnings: none.
- `plex-mono-boundary-16-regular-plus-0`: Go 1 visible lines, PowerPoint 2. Font warnings: none.
- `plex-mono-boundary-16-regular-plus-0.15`: Go 1 visible lines, PowerPoint 2. Font warnings: none.
- `plex-mono-boundary-16-bold-plus-0`: Go 1 visible lines, PowerPoint 2. Font warnings: none.
- `plex-mono-boundary-16-bold-plus-0.15`: Go 1 visible lines, PowerPoint 2. Font warnings: none.
- `plex-mono-boundary-16.5-regular-minus-0.15`: Go 2 visible lines, PowerPoint 1. Font warnings: none.
- `plex-mono-boundary-24-regular-minus-0.15`: Go 2 visible lines, PowerPoint 1. Font warnings: none.
- `plex-mono-boundary-24-bold-minus-0.15`: Go 2 visible lines, PowerPoint 1. Font warnings: none.
- `arial-boundary-16-regular-minus-0.15`: Go 2 visible lines, PowerPoint 1. Font warnings: none.
- `arial-boundary-16.5-regular-minus-0.15`: Go 2 visible lines, PowerPoint 1. Font warnings: none.
- `arial-boundary-24-regular-plus-0`: Go 1 visible lines, PowerPoint 2. Font warnings: none.
- `arial-boundary-24-regular-plus-0.15`: Go 1 visible lines, PowerPoint 2. Font warnings: none.

Width differences include cases with different wrapping; they are not single-line advance errors. Blank lines are excluded from the visible-line comparison, but remain in the corpus and geometry. Bullet markers are excluded from line text comparisons. Font names come from PDF BaseFont resources; PDFKit selection font attributes can report fallback names in this process.

Full measurements and line text are in `archived-comparison.json`. Exact source contracts are in `corpus.json` and `cases.json`. No engine calibration was applied to this reference run.
