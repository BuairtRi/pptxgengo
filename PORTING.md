# Go Port Conventions (pptxgengo)

This repo is a port of PptxGenJS (TypeScript, `src/`) to Go. The Go library lives in
`pptx/` (package `pptx`, import path `github.com/buairtri/pptxgengo/pptx`).
`go.mod` is at the repo root. **Stdlib only** — no third-party deps.

## Ground rules for all porting agents

1. **Fidelity first**: the goal is byte-identical XML output to the JS library wherever
   feasible (same element/attribute order, same number formatting). Read the TS source
   before writing Go; do not redesign the XML generation logic.
2. **File ownership**: each task owns specific Go files. NEVER edit a file owned by another
   task. If you need a helper that doesn't exist yet, define it **unexported in your own
   file** with a distinctive name and note it in your final report so it can be reconciled.
3. **TDD**: write table-driven tests first (colocated `*_test.go`), then implement.
   `go build ./...` and `go test ./...` must pass before you finish.
4. **Skip browser-only paths**: DOM/`window.getComputedStyle` table scraping, Blob outputs,
   `writeFileToBrowser`. Node `fs`/`https` paths become `os` / `net/http`.

## File map (TS → Go)

| TS source              | Go file             | Owner task |
|------------------------|---------------------|------------|
| src/core-enums.ts      | pptx/enums.go       | foundation |
| src/core-interfaces.ts | pptx/types.go       | foundation |
| src/gen-utils.ts       | pptx/utils.go       | foundation |
| src/gen-xml.ts         | pptx/xml.go         | gen-xml    |
| src/gen-charts.ts      | pptx/charts.go      | gen-charts |
| src/gen-objects.ts     | pptx/objects.go     | gen-objects|
| src/gen-tables.ts      | pptx/tables.go      | gen-tables |
| src/gen-media.ts       | pptx/media.go       | gen-media  |
| src/slide.ts           | pptx/slide.go       | top-level  |
| src/pptxgen.ts         | pptx/presentation.go, pptx/writer.go | top-level |

## Type mappings

- **`number|string` coordinates** (`Coord`): use
  ```go
  // Coord is a position/size: inches by default, or a percentage of the slide dimension.
  type Coord struct { Val float64; IsPct bool }
  func Inches(v float64) Coord  { return Coord{Val: v} }
  func Percent(v float64) Coord { return Coord{Val: v, IsPct: true} }
  ```
  `getSmartParseNumber` takes a `Coord` + layout dimension and returns EMU (int).
- **Colors**: `type Color = string` semantics — hex `"FF0000"` or scheme name (`"tx1"` etc.).
  Keep as plain `string` field; translation happens in `createColorElement`.
- **Optional fields**: use a **pointer** (`*bool`, `*float64`, `*int`, `*string`) ONLY when
  the TS code distinguishes `undefined` from the zero value (e.g. `bold?: boolean` where
  absent means "inherit" but `false` means "off"). When the zero value is never valid
  (e.g. `fontSize`, where 0 is meaningless), use a plain field with zero = unset.
  Provide/use the shared helper in utils.go:
  ```go
  func ptr[T any](v T) *T { return &v }
  ```
- **TS union slide objects** (`ISlideObject`): single `SlideObject` struct with a `Type`
  field (`SLIDE_OBJECT_TYPES` enum) plus pointer sub-structs per kind (`Text *TextProps`, …).
- **Naming**: TS `camelCase` → Go `PascalCase` for exported, keep TS name otherwise.
  Go initialism style: `XML`, `ID`, `URL`, `RGB`, `EMU` (e.g. `slideObjectToXML`).
  Internal generator functions stay unexported.
- **Errors**: TS `throw` → return `error`. No panics in library code.

## JS numeric semantics (critical for byte-identical output)

Use these shared helpers from `utils.go` everywhere a number is interpolated into XML:

- `ftoa(f float64) string` — JS `String(number)` behavior:
  `strconv.FormatFloat(f, 'f', -1, 64)` (integers print without decimal point).
- `jsRound(f float64) float64` — JS `Math.round` rounds .5 toward +Inf (Go's `math.Round`
  rounds half away from zero; they differ for negatives): `math.Floor(f + 0.5)`.
- EMU values are `int` after conversion; `inch2Emu` etc. return ints exactly as the JS
  code produces them.

## XML building

- Use `strings.Builder`; mirror the TS concatenation order exactly.
- `CRLF` constant = `"\r\n"` (matches JS).
- `encodeXmlEntities` must match the TS implementation exactly (order of replacements).

## Testing

- Table-driven unit tests colocated in `pptx/`.
- Golden reference files (produced by the JS library) live in `pptx/testdata/golden/<case>/`
  as extracted XML parts. When comparing, normalize the `dcterms:created`/`modified`
  timestamps in `docProps/core.xml` and any random GUIDs (glow/chart UUIDs).
- Timestamp/UUID injection: `Presentation` has unexported `nowFunc func() time.Time` and
  `uuidFunc func(string) string` hooks; tests override them.

## Font embedding (net-new feature, not in PptxGenJS)

Owned by the fonts task in `pptx/fonts.go` (+ `fonts_test.go`). API sketch:
`(*Presentation).EmbedFont(FontEmbedProps) error` where FontEmbedProps carries
Typeface string plus Regular/Bold/Italic/BoldItalic []byte (or file paths) TTF data.
Writer emits `ppt/fonts/fontN.fntdata` parts. Hooks other tasks MUST provide:

- **xml.go**: `makeXmlContTypes` adds `<Default Extension="fntdata" ContentType="application/x-fontdata"/>`
  when the presentation has embedded fonts; `makeXmlPresentation` emits `embedTrueTypeFonts="1"`
  attribute and `<p:embeddedFontLst>` (one `<p:embeddedFont>` per typeface with
  `<p:font typeface="..."/>` + `<p:regular r:id="..."/>` etc.) immediately after `</p:notesSz>`,
  ordering per the ECMA-376 `CT_Presentation` schema sequence: sldMasterIdLst, notesMasterIdLst,
  handoutMasterIdLst, sldIdLst, sldSz, notesSz, smartTags, embeddedFontLst, custShowLst, ...
  (embeddedFontLst comes AFTER sldIdLst/sldSz/notesSz, not immediately after sldMasterIdLst);
  `makeXmlPresentationRels` adds relationships of type
  `http://schemas.openxmlformats.org/officeDocument/2006/relationships/font` targeting `fonts/fontN.fntdata`.
- **types.go consumers**: presentation-level state lives on the Presentation struct
  (`EmbeddedFonts []EmbeddedFont`), defined in fonts.go, not types.go.

## Output API (replaces JSZip output types)

- `(*Presentation).Write() ([]byte, error)`
- `(*Presentation).WriteTo(w io.Writer) (int64, error)`
- `(*Presentation).WriteFile(path string) error`
- Compression: `archive/zip` with `Deflate` (JS default is STORE unless
  `compression: true`; expose `Compression bool` in WriteProps — golden files are
  generated with default settings, so match the JS default when comparing).
