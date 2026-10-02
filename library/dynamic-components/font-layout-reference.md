# Three-family font layout reference

The frozen reference set covers **IBM Plex Sans, Arial and IBM Plex Mono**.
It contains 276 cases, 92 per family, with actual PowerPoint PDF output and
the corresponding original v1 Go predictions. Subsequent v2 results are in
the separate [wrapping calibration](font-wrap-calibration.md). The engine
remains general and uses the authored font family.

## Captured results

| Family | Cases | Matching visible line breaks | Unexpected PDF font families |
| --- | ---: | ---: | ---: |
| IBM Plex Sans | 92 | 87 | 0 |
| Arial | 92 | 85 | 0 |
| IBM Plex Mono | 92 | 80 | 0 |

All 276 cases preserve the source text after whitespace normalization. The
24 line-break differences comprise 15 boundary cases, four wrapping cases,
three alignment cases and two long-word cases. Excluding the deliberate
boundary stress cases, 192 of 201 match. All 96 size/style cases match.
These counts describe this corpus, rather than an estimated success rate
for arbitrary presentation content.

The reference was exported locally by Microsoft PowerPoint 16.113.3 on
macOS using **File > Export > PDF > Best for printing**. Each diagnostic
page contains one text sample. The probe PPTX was unchanged after capture.

**Measurement contract:** `powerpoint-pdf-selection-bounds.v1`. PDFKit reads
line selections and their rectangles in the PowerPoint-exported PDF. This
does not expose PowerPoint TextRange bounds, baseline coordinates or a native
fit certificate. Go heights use font line boxes, so subtracting these PDF
heights does not by itself establish a line-height correction. Large width
differences include changed wrapping and cannot be treated as advance errors.

Font identities come from the PDF's BaseFont resources. PDFKit's attributed
selection text reported substitute font names in this process, so those names
are diagnostic only. The Go report records 12 font resolutions with file
hashes, styles, collection indices and variable axes. PDF resource names do
not attest which installed font-file bytes PowerPoint used.

The existing TextRange adapter could not capture references in this session:
its shell process lacked native app/font-service access, and the connected
PowerPoint execution tool required approval unavailable under the session
policy. Actual PowerPoint PDF rendering was captured through the UI instead.

## Corpus coverage

| Category | Cases | Coverage |
| --- | ---: | --- |
| Size and style | 96 | 10, 12, 14, 16, 18, 24, 32, 40 pt; regular, bold, italic, bold italic |
| Fractional sizes | 18 | 12.5, 16.5, 27.25 pt; regular and bold |
| Wrapping | 36 | 12, 16, 24 pt; widths 180/360 pt; regular and bold |
| Width boundaries | 75 | Go single-line advance plus −1, −0.15, 0, +0.15, +1 pt |
| Paragraph spacing | 9 | Line multiples 1, 1.15, 1.5; explicit before/after spacing |
| Hard breaks | 6 | Blank lines and trailing hard break |
| Alignment | 9 | Left, center, right |
| Bullets | 12 | Two markers, two widths, hanging indents |
| Mixed styles | 3 | Mixed sizes/styles and underlining |
| Mixed families | 3 | All three families within one textbox |
| Insets | 3 | Explicit text margins |
| Long words | 3 | Forced breaks within overlong words |
| Punctuation | 3 | Smart punctuation, euro, accents, underscore |

Visible-line comparisons normalize Unicode compatibility forms, strip
whitespace at line edges, and omit blank lines and authored bullet markers.
The original text, blank-line geometry and bullet geometry remain in the
source and measurement files.

## Files

The durable baseline is in [font-reference](font-reference/):

- [Corpus](font-reference/corpus.json) and [case definitions](font-reference/cases.json).
- [Exact probe PPTX](font-reference/font-reference-probes.pptx),
  [probe manifest](font-reference/probe-manifest.json) and
  [PowerPoint reference PDF](font-reference/native-reference.pdf).
- [Go measurements](font-reference/go-measurements.json),
  [PDF measurements](font-reference/native-measurements.json),
  [comparison](font-reference/comparison-final.json) and
  [CSV comparison](font-reference/comparison-final.csv).
- [Capture and artifact hashes](font-reference/proof.json).

Two editable review decks contain 11 slides each, with matching samples in
16:9 and 4:3. Their paths and hashes are in the proof. All 22 slides were
inspected individually from PowerPoint PDF renders. Review frames include
clearance informed by the captured PDF reference; these frames do not change
the frozen Go predictions. The 276-page diagnostic probe is the complete
case set; the review decks show a subset.

The initial `reference-set.json` is a preparation receipt and retains its
pending status. `proof.json` is the final capture receipt. Preliminary
comparison files under `samples/` are superseded by `comparison-final.*`.

## Reproduce or compare a candidate

Build the current source binary and create a new output directory:

```sh
go build -o /tmp/pptxcompose-font-reference ./cmd/pptxcompose
python3 scripts/font-layout-reference.py prepare \
  --binary /tmp/pptxcompose-font-reference --out /tmp/font-reference-new
```

Open `/tmp/font-reference-new/font-reference-probes/font-reference-probes.pptx`
in PowerPoint and export a local PDF. Then extract and compare:

```sh
swift scripts/read-font-reference-pdf.swift \
  /tmp/font-reference-new/font-reference-probes/native-reference.pdf \
  /tmp/font-reference-new/native-measurements.json 2>/tmp/font-pdf-extraction.log
python3 scripts/font-layout-reference.py compare \
  --out /tmp/font-reference-new \
  --native /tmp/font-reference-new/native-measurements.json --name comparison-final
python3 scripts/font-layout-reference.py review \
  --binary /tmp/pptxcompose-font-reference --out /tmp/font-reference-new \
  --native /tmp/font-reference-new/native-measurements.json
```

To compare a changed Go engine against the **frozen** PowerPoint reference:

```sh
/tmp/pptxcompose-font-reference measure --engine go \
  --spec library/dynamic-components/font-reference/corpus.json \
  --out /tmp/font-candidate-go.json
mkdir /tmp/font-candidate-comparison
python3 scripts/font-layout-reference.py compare \
  --out library/dynamic-components/font-reference \
  --native library/dynamic-components/font-reference/native-measurements.json \
  --go /tmp/font-candidate-go.json --reports /tmp/font-candidate-comparison \
  --name comparison-candidate
```

Use new output filenames for every run. The comparator checks the source,
probe and PDF hashes, case/page counts, request order and expected font-family
resources. Boundary widths are frozen against the initial Go engine; reusing
them avoids moving the benchmark when the engine changes. Font hashes in a
candidate report should be compared with the baseline before attributing a
change to the engine.

## Next engineering work

The first wrapping calibration is implemented in v2; see its separate report
for the improved counts, additional corpus and remaining boundary cases.

Investigate line wrapping first, especially the nine boundary cases where Go
predicts one visible line and PowerPoint produces two. Include Mono's wider
paragraph cases and the long-word break policy. Keep fixes grounded in font
metrics and shared layout rules. Then capture native TextRange measurements
to establish a height policy, and expand to realistic template components.
This reference does not enable or qualify Go as the default engine.
