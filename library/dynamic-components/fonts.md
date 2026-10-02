# Font selection and native measurement

Measured composition supports explicit installed font families, including IBM
Plex Sans. PowerPoint lays out the actual text; no Arial width approximation is
used for other fonts. Font family names must be explicit, with no surrounding
whitespace or theme aliases such as `+mn-lt`.

An opt-in [Go layout prototype](go-layout.md) also supports explicit families
and styles. `--engine go` measures and builds without PowerPoint or separate
scripts, using font metrics and shaping. Its fit predictions have separate
provenance and are not native measurement evidence.

## Select a font

For adaptive input, set the slide style:

```json
"style": {"font_face": "IBM Plex Sans"}
```

The selected family applies to headings, body text, labels, pods, matrix cells,
legends, page numbers and footers across all five adaptive builders. Omitting
the adaptive setting retains Arial.

For compose input, set `title_font_face` on the slide and `font_face` on each
text element, layout block, pod style, role, phase, legend and process path.
Rich paragraphs set `font_face` on each run and may mix families. Cards accept
optional `font_face`; omitting it retains Arial. Point sizes and `bold` remain
explicit; rich runs also support `italic` and `underline`.

## Availability and cache behavior

Install the family on the Mac running PowerPoint before measurement. The
environment inspector resolves each requested family/style through CoreText,
checks that the family and bold/italic traits match, and requires an accessible
font file. Missing families or requested styles stop measurement with an error.
Only styles actually used by the text are required. Font embedding is separate
from this workflow; generated compose decks reference installed fonts.

Environment fingerprints contain the family, style, PostScript name, font file
path and SHA-256, and any resolved variable font coordinates, along with the OS,
PowerPoint version/build and measurement tooling. A changed family, font file,
variable font instance, text, width or typography requires new measurements.
The cache is conservatively scoped to the full requested font/style set;
changing that set can also invalidate otherwise unchanged text contracts.
Partial probe decks retain the full set, including fonts used by cached hits.

Native verification checks PowerPoint's reported character fonts and styles.
Character bounds are layout advances, rather than raster ink bounds. Changed
fonts can change wrapping and required heights: measure, build, verify and
visually review the resulting deck. Existing Arial fit evidence does not
establish capacity for IBM Plex Sans or another family.

Variable font files are supported for the regular/bold/italic combinations
resolved by the OS. Arbitrary weight, width and optical-size axis settings are
not authoring controls in this API.

## Example

[`fonts.json`](fonts.json) exercises IBM Plex Sans regular, bold, italic and
bold italic, wrapping compared with Arial, and cards in both families:

```sh
go build -o /tmp/pptxcompose-fonts ./cmd/pptxcompose
/tmp/pptxcompose-fonts probe --spec library/dynamic-components/fonts.json --cache /tmp/font-cache --out /tmp/font-probes
/tmp/pptxcompose-fonts measure --bundle /tmp/font-probes --cache /tmp/font-cache --out /tmp/font-evidence.json
/tmp/pptxcompose-fonts build --spec library/dynamic-components/fonts.json --cache /tmp/font-cache --out /tmp/font-output
/tmp/pptxcompose-fonts verify --bundle /tmp/font-output --out /tmp/font-verification.json
```

The bounded example's [native evidence and review record](fonts-proof.json)
records the exact measured font instances and artifacts.
Its required evidence is versioned in `font-support-evidence/`; hashes and
original capture paths are retained. Release packaging does not depend on those
font artifacts remaining in the ignored `samples/` directory.

Use the same binary throughout the run and new output paths. Frozen installed
releases contain their own binaries and scripts; repository changes require a
new release installation before the global launcher uses them.
